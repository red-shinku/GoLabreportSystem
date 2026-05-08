package httperr

import "net/http"

// 该文件定义 HTTP 的领域错误
// 包括错误类型与http错误码的包装
// 可用于控制层中，根据返回的下层领域错误进行包装，绑定状态码
// 这统一了错误类型和其处理

// statusError 错误类型与http状态码的包装类型，不导出，通过接口访问
type statusError struct {
	error
	status int
}

// Unwrap 提取错误内容
func (err statusError) Unwrap() error {
	return err.error
}

// HTTPStatus 返回被包装的错误码
func (err statusError) HTTPStatus() int {
	return err.status
}

// WithStatus 将错误与http状态码包装
func WithStatus(err error, status int) error {
	return statusError{err, status}
}

// HTTPStatus 提取错误的HTTP状态码，默认返回500
func HTTPStatus(err error) int {
	if err == nil {
		return 0
	}
	type statErrInterface interface{ HTTPStatus() int }
	if s, ok := err.(statErrInterface); ok {
		return s.HTTPStatus()
	}
	return http.StatusInternalServerError
}
