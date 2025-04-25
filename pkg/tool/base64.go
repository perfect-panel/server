package tool

import (
	"encoding/base64"
	"strings"
)

// IsValidImageSize 检查base64图片是否有效且未超出大小限制
// base64Str: base64编码的图片字符串
// maxSizeKB: 最大允许大小（KB），int64类型
// 返回: bool - true表示图片有效且未超限，false表示无效或超限
func IsValidImageSize(base64Str string, maxSizeKB int64) bool {
	// 输入验证
	if base64Str == "" || maxSizeKB < 0 {
		return false
	}

	// 提取纯base64数据
	data := base64Str
	if idx := strings.Index(base64Str, ","); idx != -1 {
		data = base64Str[idx+1:]
	}

	// 快速估算大小（避免完整解码）
	// base64编码后每3字节原始数据变成4字节，解码后大小约为输入长度的3/4
	approxSizeBytes := int64(len(data)) * 3 / 4
	approxSizeKB := approxSizeBytes / 1024

	// 如果估算大小已超限，无需解码
	if approxSizeKB > maxSizeKB {
		return false
	}

	// 验证base64有效性
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false
	}

	// 精确检查大小
	return int64(len(decoded))/1024 <= maxSizeKB
}
