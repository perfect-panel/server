package result

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/pkg/xerr"
)

// HttpResult HTTP Result
func HttpResult(ctx *gin.Context, resp interface{}, err error) {

	if err == nil {
		// Success Result
		ctx.JSON(http.StatusOK, Success(resp))
		return
	}

	// Init Error Code and Message
	code := xerr.ERROR
	msg := "Internal Server Error"

	// Get Error Type
	var e *xerr.CodeError
	if errors.As(errors.Cause(err), &e) {
		// Custom Code Error
		code = e.GetErrCode()
		msg = e.GetErrMsg()
	}
	ctx.JSON(http.StatusOK, Error(code, msg))
}

// ParamErrorResult Param Error Result
func ParamErrorResult(ctx *gin.Context, err error) {
	errMsg := err.Error()
	_ = ctx.Error(errors.New(errMsg))
	ctx.JSON(http.StatusOK, Error(xerr.InvalidParams, errMsg))
}
