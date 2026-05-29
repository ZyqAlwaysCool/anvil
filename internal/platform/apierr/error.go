package apierr

import (
	"errors"
	"net/http"
)

const (
	CodeSuccess        = 0
	CodeBadRequest     = 10001
	CodeInternal       = 10002
	CodeNotFound       = 10003
	CodeUnauthorized   = 10004
	CodeTaskConflict   = 10005
	CodeTaskNotAllowed = 10006
)

// Error 平台通用错误；业务域错误码由业务包自行定义，不得写入本包。
type Error struct {
	httpStatus int
	code       int
	msg        string
	data       any
}

func New(httpStatus, code int, msg string) *Error {
	return NewWithData(httpStatus, code, msg, nil)
}

func NewWithData(httpStatus, code int, msg string, data any) *Error {
	return &Error{
		httpStatus: httpStatus,
		code:       code,
		msg:        msg,
		data:       data,
	}
}

func BadRequest(msg string) *Error {
	return New(http.StatusBadRequest, CodeBadRequest, msg)
}

func NotFound(msg string) *Error {
	return New(http.StatusNotFound, CodeNotFound, msg)
}

func Internal(msg string) *Error {
	return New(http.StatusInternalServerError, CodeInternal, msg)
}

func Unauthorized(msg string) *Error {
	return New(http.StatusUnauthorized, CodeUnauthorized, msg)
}

// From 将未知错误统一收敛为平台内部错误，避免向客户端泄漏底层细节。
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err.Error())
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.msg
}

func (e *Error) StatusCode() int {
	if e == nil {
		return http.StatusInternalServerError
	}
	return e.httpStatus
}

func (e *Error) Code() int {
	if e == nil {
		return CodeInternal
	}
	return e.code
}

func (e *Error) Message() string {
	if e == nil {
		return "internal server error"
	}
	return e.msg
}

func (e *Error) Data() any {
	if e == nil {
		return nil
	}
	return e.data
}
