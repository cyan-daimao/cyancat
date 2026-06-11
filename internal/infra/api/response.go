package api

// Response 统一响应结构
type Response[T any] struct {
	// Code 业务状态码
	Code Code `json:"code"`
	// Message 提示信息
	Message string `json:"message"`
	// Data 业务数据
	Data T `json:"data"`
}

// Success 返回成功响应
func Success[T any](data T) *Response[T] {
	return &Response[T]{
		Code:    SuccessCode,
		Message: SuccessCode.String(),
		Data:    data,
	}
}

// Fail 返回失败响应
func Fail[T any](code Code, data T, message string) *Response[T] {
	return &Response[T]{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Error 返回错误响应（无数据）
func Error(code Code, message string) *Response[any] {
	return &Response[any]{
		Code:    code,
		Message: message,
	}
}