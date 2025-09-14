package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/perfect-panel/server/pkg/logger"

	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/perfect-panel/server/internal/svc"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func LoggerMiddleware(svc *svc.ServiceContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		// get response body
		w := &responseBodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w
		// get request body
		var requestBody []byte
		if c.Request.Body != nil {
			// c.Request.Body It can only be read once, and after reading, it needs to be reassigned to c.Request Body
			requestBody, _ = io.ReadAll(c.Request.Body)
			// After reading, reassign c.Request Body ， For subsequent operations
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
		// start time
		start := time.Now()
		c.Next()
		// Start recording logs
		cost := time.Since(start)
		responseStatus := c.Writer.Status()

		host := c.Request.Host

		logs := []logger.LogField{
			{
				Key:   "status",
				Value: responseStatus,
			},
			{
				Key:   "request",
				Value: c.Request.Method + " " + host + c.Request.URL.String(),
			},
			{
				Key:   "query",
				Value: c.Request.URL.RawQuery,
			},
			{
				Key:   "ip",
				Value: c.ClientIP(),
			},
			{
				Key:   "user-agent",
				Value: c.Request.UserAgent(),
			},
		}
		if c.Errors.Last() != nil {
			var e *xerr.CodeError
			var errMessage string
			if errors.As(c.Errors.Last().Err, &e) {
				errMessage = e.GetErrMsg()
			} else {
				errMessage = c.Errors.Last().Error()
			}
			logs = append(logs, logger.Field("error", errMessage))
		}
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			// request content
			logs = append(logs, logger.Field("request_body", string(maskSensitiveFields(requestBody, []string{"password", "old_password", "new_password"}))))
			// response content
			logs = append(logs, logger.Field("response_body", w.body.String()))
		}
		logs = append(logs, logger.Field("duration", cost))
		if responseStatus >= 500 && responseStatus <= 599 {
			logger.WithContext(c.Request.Context()).Errorw("HTTP Error", logs...)
		} else {
			logger.WithContext(c.Request.Context()).Infow("HTTP Request", logs...)
		}

		if responseStatus == 404 {
			logger.WithContext(c.Request.Context()).Debugf("404 Not Found: Host:%s Path:%s IsPanDomain:%v", host, c.Request.URL.Path, svc.Config.Subscribe.PanDomain)
		}
	}
}

func maskSensitiveFields(data []byte, fieldsToMask []string) []byte {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return data
	}

	for _, field := range fieldsToMask {
		if _, exists := jsonData[field]; exists {
			jsonData[field] = "***" // use *** to mask sensitive fields
		}
	}
	maskedData, err := json.Marshal(jsonData)
	if err != nil {
		return data
	}
	return maskedData
}
