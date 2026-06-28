package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero/api"
	"google.golang.org/protobuf/proto"
)

// packPtrLen 将指针和长度打包为一个 int64
// 高 32 位: 指针, 低 32 位: 长度
func packPtrLen(ptr uint32, length uint32) int64 {
	return (int64(ptr) << 32) | int64(length)
}

// unpackPtrLen 从 int64 中解包指针和长度
func unpackPtrLen(packed int64) (ptr uint32, length uint32) {
	return uint32(packed >> 32), uint32(packed & 0xFFFFFFFF)
}

// readWasmMemory 从 WASM 线性内存读取数据
func readWasmMemory(mod api.Module, ptr uint32, length uint32) ([]byte, error) {
	if length == 0 {
		return nil, nil
	}
	mem := mod.Memory()
	if mem == nil {
		return nil, errors.New("no memory available")
	}
	data, ok := mem.Read(ptr, length)
	if !ok {
		return nil, fmt.Errorf("failed to read %d bytes from WASM memory at offset %d", length, ptr)
	}
	return data, nil
}

// writeWasmMemory 将数据写入 WASM 线性内存（调用 guest allocate）
// 返回打包的 (ptr, len)
func writeWasmMemory(ctx context.Context, mod api.Module, data []byte) int64 {
	if len(data) == 0 {
		return packPtrLen(0, 0)
	}

	alloc := mod.ExportedFunction("allocate")
	if alloc == nil {
		return packPtrLen(0, 0)
	}

	results, err := alloc.Call(ctx, uint64(len(data)))
	if err != nil {
		return packPtrLen(0, 0)
	}

	ptr := uint32(results[0])
	mod.Memory().Write(ptr, data)
	return packPtrLen(ptr, uint32(len(data)))
}

// HostFunc 是宿主函数的业务逻辑签名
type HostFunc[T proto.Message, R proto.Message] func(
	ctx context.Context,
	pluginName string,
	req T,
) (R, error)

// MakeHostFunc 创建一个 wazero host function（GoModuleFunc）。
//
// WASM 函数签名: (i32, i32) -> i64
//   - param 0 (i32): 指向 protobuf 请求数据的指针
//   - param 1 (i32): 请求数据长度
//   - return (i64): 打包的 (result_ptr << 32) | result_len
//
// 内部自动完成：
//
//	读内存 → proto.Unmarshal → 业务逻辑 → proto.Marshal → allocate + 写内存 → 返回打包指针
func MakeHostFunc[T, R proto.Message](
	pluginName string,
	fn HostFunc[T, R],
	newT func() T,
) api.GoModuleFunc {
	return api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
		// 从栈读取参数 (i32, i32)
		paramPtr := api.DecodeI32(stack[0])
		paramLen := api.DecodeI32(stack[1])

		// 1. 从 WASM 内存读取请求字节
		data, err := readWasmMemory(mod, uint32(paramPtr), uint32(paramLen))
		if err != nil {
			stack[0] = uint64(packPtrLen(0, 0))
			return
		}

		// 2. 解码 protobuf 请求
		req := newT()
		if len(data) > 0 {
			if err := proto.Unmarshal(data, req); err != nil {
				stack[0] = uint64(packPtrLen(0, 0))
				return
			}
		}

		// 3. 执行业务逻辑
		resp, err := fn(ctx, pluginName, req)
		if err != nil {
			stack[0] = uint64(packPtrLen(0, 0))
			return
		}

		// 4. 编码 protobuf 响应
		resultData, err := proto.Marshal(resp)
		if err != nil {
			stack[0] = uint64(packPtrLen(0, 0))
			return
		}

		// 5. 写入 WASM 内存并返回打包指针
		stack[0] = uint64(writeWasmMemory(ctx, mod, resultData))
	})
}

// CallGuestStart 调用 guest 的 start 函数（可选）
func CallGuestStart(ctx context.Context, mod api.Module) error {
	fn := mod.ExportedFunction("start")
	if fn == nil {
		return nil // start 是可选的
	}
	_, err := fn.Call(ctx)
	return err
}

// CallGuestStop 调用 guest 的 stop 函数（可选）
func CallGuestStop(ctx context.Context, mod api.Module) error {
	fn := mod.ExportedFunction("stop")
	if fn == nil {
		return nil // stop 是可选的
	}
	_, err := fn.Call(ctx)
	return err
}

// CallGuestFn 调用 guest WASM 模块的导出函数
//
// 自动完成: proto.Marshal → allocate + 写入 WASM 内存 → 调用 guest fn → 读内存 → proto.Unmarshal
//
// 参数:
//   - exportName: guest 导出的函数名（如 "init", "handle_get_stats"）
//   - req: 请求 protobuf 消息
//   - newR: 创建响应类型的工厂函数
func CallGuestFn[T, R proto.Message](
	ctx context.Context,
	mod api.Module,
	exportName string,
	req T,
	newR func() R,
) (R, error) {
	var zero R

	// 1. 查找导出函数
	fn := mod.ExportedFunction(exportName)
	if fn == nil {
		return zero, fmt.Errorf("exported function %q not found", exportName)
	}

	// 2. 编码请求并写入 WASM 内存
	reqData, err := proto.Marshal(req)
	if err != nil {
		return zero, fmt.Errorf("marshal request for %s: %w", exportName, err)
	}

	packed := writeWasmMemory(ctx, mod, reqData)
	reqPtr, reqLen := unpackPtrLen(packed)

	// 3. 调用 guest 函数 (i32, i32) -> i64
	results, err := fn.Call(ctx, uint64(reqPtr), uint64(reqLen))
	if err != nil {
		return zero, fmt.Errorf("call %q: %w", exportName, err)
	}
	if len(results) == 0 {
		return zero, nil // 无返回值
	}

	// 4. 读取响应
	respPtr, respLen := unpackPtrLen(int64(results[0]))
	if respLen == 0 {
		return zero, nil // 空响应
	}
	defer freeWasmMemory(ctx, mod, respPtr, respLen)

	respData, err := readWasmMemory(mod, respPtr, respLen)
	if err != nil {
		return zero, fmt.Errorf("read response from %q: %w", exportName, err)
	}

	// 5. 解码响应
	resp := newR()
	if len(respData) > 0 {
		if err := proto.Unmarshal(respData, resp); err != nil {
			return zero, fmt.Errorf("unmarshal response from %q: %w", exportName, err)
		}
	}

	return resp, nil
}

// freeWasmMemory 释放 WASM 内存
func freeWasmMemory(ctx context.Context, mod api.Module, ptr, length uint32) {
	if length == 0 {
		return
	}
	deallocFn := mod.ExportedFunction("deallocate")
	if deallocFn == nil {
		return
	}
	_, _ = deallocFn.Call(ctx, uint64(ptr), uint64(length))
}
