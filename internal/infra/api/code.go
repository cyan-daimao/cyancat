// Package api 定义统一响应模型和状态码
package api

// Code 业务状态码
type Code int

const (
	// SuccessCode 成功
	SuccessCode Code = 200
	// BadRequestCode 请求参数错误
	BadRequestCode Code = 400
	// NotFoundCode 资源不存在
	NotFoundCode Code = 404
	// AuthErrorCode 认证失败
	AuthErrorCode Code = 401
	// ConflictCode 资源冲突
	ConflictCode Code = 409
	// ErrorCode 内部错误
	ErrorCode Code = 500
)

// String 返回状态码对应的消息
func (c Code) String() string {
	switch c {
	case SuccessCode:
		return "success"
	case BadRequestCode:
		return "bad request"
	case NotFoundCode:
		return "not found"
	case AuthErrorCode:
		return "authentication error"
	case ConflictCode:
		return "conflict"
	case ErrorCode:
		return "internal error"
	default:
		return "unknown"
	}
}