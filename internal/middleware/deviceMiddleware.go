package middleware

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/perfect-panel/server/internal/svc"
	pkgaes "github.com/perfect-panel/server/pkg/aes"
	"github.com/perfect-panel/server/pkg/constant"
	"github.com/perfect-panel/server/pkg/result"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

func DeviceMiddleware(srvCtx *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {

		if !srvCtx.Config.Device.Enable {
			c.Next()
			return
		}

		if srvCtx.Config.Device.SecuritySecret == "" {
			result.HttpResult(c, nil, errors.Wrapf(xerr.NewErrCode(xerr.SecretIsEmpty), "Secret is empty"))
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		if ctx.Value(constant.CtxKeyUser) == nil && c.GetHeader("Login-Type") != "" {
			ctx = context.WithValue(ctx, constant.LoginType, c.GetHeader("Login-Type"))
			c.Request = c.Request.WithContext(ctx)
		}

		loginType, ok := ctx.Value(constant.LoginType).(string)
		if !ok || loginType != "device" {
			c.Next()
			return
		}

		rw := NewResponseWriter(c, srvCtx)
		if !rw.Decrypt() {
			result.HttpResult(c, nil, errors.Wrapf(xerr.NewErrCode(xerr.InvalidCiphertext), "Invalid ciphertext"))
			c.Abort()
			return
		}
		c.Writer = rw
		c.Next()
		rw.FlushAbort()
	}
}

func NewResponseWriter(c *gin.Context, srvCtx *svc.ServiceContext) (rw *ResponseWriter) {
	rw = &ResponseWriter{
		c:              c,
		body:           new(bytes.Buffer),
		ResponseWriter: c.Writer,
	}
	rw.encryptionKey = srvCtx.Config.Device.SecuritySecret
	rw.encryptionMethod = "AES"
	rw.encryption = true
	return rw
}

func (rw *ResponseWriter) Encrypt() {
	if !rw.encryption {
		return
	}
	buf := rw.body.Bytes()
	params := map[string]interface{}{}
	err := json.Unmarshal(buf, &params)
	if err != nil {
		return
	}
	data := params["data"]
	if data != nil {
		var jsonData []byte
		str, ok := data.(string)
		if ok {
			jsonData = []byte(str)
		} else {
			jsonData, _ = json.Marshal(data)
		}
		encrypt, iv, err := pkgaes.Encrypt(jsonData, rw.encryptionKey)
		if err != nil {
			return
		}
		params["data"] = map[string]interface{}{
			"data": encrypt,
			"time": iv,
		}

	}
	marshal, _ := json.Marshal(params)
	rw.body.Reset()
	rw.body.Write(marshal)
}

func (rw *ResponseWriter) Decrypt() bool {
	if !rw.encryption {
		return true
	}

	//判断url链接中是否存在data和iv数据，存在就进行解密并设置回去
	query := rw.c.Request.URL.Query()
	dataStr := query.Get("data")
	timeStr := query.Get("time")
	if dataStr != "" && timeStr != "" {
		decrypt, err := pkgaes.Decrypt(dataStr, rw.encryptionKey, timeStr)
		if err == nil {
			params := map[string]interface{}{}
			err = json.Unmarshal([]byte(decrypt), &params)
			if err == nil {
				for k, v := range params {
					query.Set(k, fmt.Sprintf("%v", v))
				}
				query.Del("data")
				query.Del("time")
				rw.c.Request.RequestURI = fmt.Sprintf("%s?%s", rw.c.Request.RequestURI[:strings.Index(rw.c.Request.RequestURI, "?")], query.Encode())
				rw.c.Request.URL.RawQuery = query.Encode()
			}
		}
	}

	//判断body是否存在数据，存在就尝试解密，并设置回去
	body, err := io.ReadAll(rw.c.Request.Body)
	if err != nil {
		return true
	}

	if len(body) == 0 {
		return true
	}

	params := map[string]interface{}{}
	err = json.Unmarshal(body, &params)
	data := params["data"]
	nonce := params["time"]
	if err != nil || data == nil {
		return false
	}

	str, ok := data.(string)
	if !ok {
		return false
	}
	iv, ok := nonce.(string)
	if !ok {
		return false
	}

	decrypt, err := pkgaes.Decrypt(str, rw.encryptionKey, iv)
	if err != nil {
		return false
	}
	rw.c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(decrypt)))
	return true
}

func (rw *ResponseWriter) FlushAbort() {
	defer rw.c.Abort()
	responseBody := rw.body.String()
	fmt.Println("Original Response Body:", responseBody)
	rw.flush = true
	if rw.encryption {
		rw.Encrypt()
	}
	_, err := rw.Write(rw.body.Bytes())
	if err != nil {
		return
	}
}

type ResponseWriter struct {
	http.ResponseWriter
	size             int
	status           int
	flush            bool
	body             *bytes.Buffer
	c                *gin.Context
	encryption       bool
	encryptionKey    string
	encryptionMethod string
}

func (rw *ResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

//nolint:unused
func (rw *ResponseWriter) reset(writer http.ResponseWriter) {
	rw.ResponseWriter = writer
	rw.size = noWritten
	rw.status = defaultStatus
}

func (rw *ResponseWriter) WriteHeader(code int) {
	if code > 0 && rw.status != code {
		if rw.Written() {
			return
		}
		rw.status = code
	}
}

func (rw *ResponseWriter) WriteHeaderNow() {
	if !rw.Written() {
		rw.size = 0
		rw.ResponseWriter.WriteHeader(rw.status)
	}
}

func (rw *ResponseWriter) Write(data []byte) (n int, err error) {
	if rw.flush {
		rw.WriteHeaderNow()
		n, err = rw.ResponseWriter.Write(data)
		rw.size += n
	} else {
		rw.body.Write(data)
	}
	return
}

func (rw *ResponseWriter) WriteString(s string) (n int, err error) {
	if rw.flush {
		rw.WriteHeaderNow()
		n, err = rw.ResponseWriter.Write([]byte(s))
		rw.size += n
	} else {
		rw.body.Write([]byte(s))
	}
	return
}

func (rw *ResponseWriter) Status() int {
	return rw.status
}

func (rw *ResponseWriter) Size() int {
	return rw.size
}

func (rw *ResponseWriter) Written() bool {
	return rw.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rw.size < 0 {
		rw.size = 0
	}
	return rw.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotifier interface.
func (rw *ResponseWriter) CloseNotify() <-chan bool {
	// 通过 r.Context().Done() 来监听请求的取消
	done := rw.c.Request.Context().Done()
	closed := make(chan bool)

	// 当上下文被取消时，通过 closed channel 发送通知
	go func() {
		<-done
		closed <- true
	}()

	return closed
}

// Flush implements the http.Flusher interface.
func (rw *ResponseWriter) Flush() {
	rw.WriteHeaderNow()
	rw.ResponseWriter.(http.Flusher).Flush()
}

func (rw *ResponseWriter) Pusher() (pusher http.Pusher) {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}
