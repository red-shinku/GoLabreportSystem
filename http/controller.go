package server

// HTTP控制层，该文件定义最终的HTTP处理器
//
// 负责解析 HTTP 的请求头、请求体等，获取请求内容，并调用业务逻辑
// 根据 RESTful API ，以资源划分处理器，命名格式为 XxxHandler
// XxxHandler 会使用 service 中对应的业务类型

import (
	"LabSystem/api"
	"LabSystem/internal/httperr"
	"LabSystem/service"
	"encoding/json"
	"fmt"
	"net/http"
)

// ServeError 统一的 HTTP 层错误处理
func ServeError(w http.ResponseWriter, err error) {
	status := httperr.HTTPStatus(err)
	msg := "Internal Server Error"
	// 不暴露服务器内部错误
	if status < 500 {
		msg = err.Error()
	}

	// 统一以json格式返回
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// HandlerFunc 自定义的 Handler 类型，添加了对错误的返回
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Adapt http.HandlerFunc的适配器，适配自定义Handler
// 让Handler统一将错误返回，在一处处理
func Adapt(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			ServeError(w, err)
		}
	}
}

// Sessions 会话资源，即登录
type Sessions struct {
	auth *service.AuthService
}

// Login 登录。验证密码后，重定向URL，指向对应的用户界面
// 其后置中间件会负责处理凭证
func (s *Sessions) Login(w http.ResponseWriter, r *http.Request) error {
	var body api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}
	userInfo, err := s.auth.LoginAuth(body.UserNumber, body.Password)
	if err != nil {
		return err
	}
	userIdCk := http.Cookie{
		Name:     "user_id",
		Value:    userInfo.Number,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	identity, ok := api.Identity(userInfo.Identity)
	if !ok {
		return httperr.WithStatus(
			fmt.Errorf("identity code not exist"),
			http.StatusInternalServerError)
	}
	identityCk := http.Cookie{
		Name:     "identity",
		Value:    identity,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	//TODO: 身份验证如何做？
	w.Header().Add("Set-Cookie", userIdCk.String())
	w.Header().Add("Set-Cookie", identityCk.String())
	// 重定向到用户界面
	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
	return nil
}

type Home struct{}

func (h *Home) LoginPage(w http.ResponseWriter, r *http.Request) error {}

// HomePage 根据解析到的cookie中的身份信息，返回对应的用户界面
func (h *Home) HomePage(w http.ResponseWriter, r *http.Request) error {}

// Projects 与项目资源相关的
type Projects struct {
	tecProjects *service.TeacherProjectService
	stuProjects *service.StudentProjectService
	tecReports  *service.TeacherReportService
	stuReports  *service.StudentReportService
}

// Submissions 实验报告资源
// 在下载/预览单个报告时作为主资源
// 其他时候仅作为项目的子资源
type Submissions struct {
	tecReports *service.TeacherReportService
}

// Users 用户资源，包括用户信息操作
type Users struct{}

// PasswordResets 忘记密码
type PasswordResets struct{}
