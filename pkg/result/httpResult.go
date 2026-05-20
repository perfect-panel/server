package result

import (
	"net/http"

	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/pkg/xerr"
)

type HTTPResult struct {
	StatusCode int
	Body       interface{}
}

func BuildHTTPResult(resp interface{}, err error) HTTPResult {
	if err == nil {
		return HTTPResult{
			StatusCode: http.StatusOK,
			Body:       Success(resp),
		}
	}

	code := xerr.ERROR
	msg := "Internal Server Error"

	var e *xerr.CodeError
	if errors.As(errors.Cause(err), &e) {
		code = e.GetErrCode()
		msg = e.GetErrMsg()
	}

	return HTTPResult{
		StatusCode: http.StatusOK,
		Body:       Error(code, msg),
	}
}

func BuildParamErrorResult(err error) HTTPResult {
	return HTTPResult{
		StatusCode: http.StatusOK,
		Body:       Error(xerr.InvalidParams, err.Error()),
	}
}

// HttpResult HTTP Result
func HttpResult(ctx *hertzx.Context, resp interface{}, err error) {
	result := BuildHTTPResult(resp, err)
	ctx.JSON(result.StatusCode, result.Body)
}

// ParamErrorResult Param Error Result
func ParamErrorResult(ctx *hertzx.Context, err error) {
	errMsg := err.Error()
	_ = ctx.Error(errors.New(errMsg))
	result := BuildParamErrorResult(err)
	ctx.JSON(result.StatusCode, result.Body)
}
