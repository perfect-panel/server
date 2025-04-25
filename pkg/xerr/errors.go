package xerr

import (
	"errors"
	"fmt"
)

/**
General common fixed error
*/

type CodeError struct {
	errCode uint32
	errMsg  string
}

var StatusNotModified = errors.New("304 Not Modified")

// GetErrCode returns the error code displayed to the front end
func (e *CodeError) GetErrCode() uint32 {
	return e.errCode
}

// GetErrMsg returns the error message displayed to the front end
func (e *CodeError) GetErrMsg() string {
	return e.errMsg
}

func (e *CodeError) Error() string {
	return fmt.Sprintf("ErrCode:%d，ErrMsg:%s", e.errCode, e.errMsg)
}

func NewErrCodeMsg(errCode uint32, errMsg string) *CodeError {
	return &CodeError{errCode: errCode, errMsg: errMsg}
}
func NewErrCode(errCode uint32) *CodeError {
	return &CodeError{errCode: errCode, errMsg: MapErrMsg(errCode)}
}

func NewErrMsg(errMsg string) *CodeError {
	return &CodeError{errCode: ERROR, errMsg: errMsg}
}
