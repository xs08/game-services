package utils

import (
	"errors"
	"net/http"
)

// AppError 应用错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// 错误码定义
const (
	ErrCodeInternal     = 1000
	ErrCodeInvalidInput = 1001
	ErrCodeNotFound     = 1002
	ErrCodeUnauthorized = 1003
	ErrCodeForbidden    = 1004
	ErrCodeConflict     = 1005
	ErrCodeTooManyRequests = 1006
)

// 错误构造函数
func NewError(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewErrorWithErr(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// HTTP 状态码映射
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeInvalidInput:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// 常用错误
var (
	ErrInternal     = errors.New("内部服务器错误")
	ErrInvalidInput = errors.New("无效的输入参数")
	ErrNotFound     = errors.New("资源未找到")
	ErrUnauthorized = errors.New("未授权")
	ErrForbidden    = errors.New("禁止访问")
	ErrConflict     = errors.New("资源冲突")
)

