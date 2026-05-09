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
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

// 暂定义的JWT密钥
var secret []byte

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

// genShortJWT 生成短期（30分钟）的JWT，
// secret为服务端管理员定义的密钥，需要在配置中获取
func (s *Sessions) genShortJWT(userID string, role string, secret []byte) (string, error) {
	expirationTime := time.Now().Add(30 * time.Minute)

	claims := api.LoginJWT{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "auth", // 可选
		},
	}

	// 创建 token（指定签名算法，HS256）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名并生成字符串
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
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
	role, ok := api.Role(userInfo.Identity)
	if !ok {
		return httperr.WithStatus(
			fmt.Errorf("Sessions.Login() error: unknown identity code"),
			http.StatusInternalServerError)
	}
	authJwt, err := s.genShortJWT(
		userInfo.Number,
		role,
		secret,
	)
	authJWTCk := http.Cookie{
		Name:     "auth_token",
		Value:    authJwt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &authJWTCk)

	// 重定向到用户界面
	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
	return nil
}

type Home struct{}

func (h *Home) LoginPage(w http.ResponseWriter, r *http.Request) error {}

// HomePage 根据解析到的cookie中的身份信息，返回对应的用户界面
func (h *Home) HomePage(w http.ResponseWriter, r *http.Request) error {}

// OfferingClass 授课班级及其子资源
type OfferingClass struct {
	tecProjects *service.TeacherProjectService
}

// CreateProject 在url中解析到的offeringId下，新建项目
func (o *OfferingClass) CreateProject(w http.ResponseWriter, r *http.Request) error {}

// Projects 与项目资源相关的
type Projects struct {
	tecProjects *service.TeacherProjectService
	stuProjects *service.StudentProjectService
	tecReports  *service.TeacherReportService
	stuReports  *service.StudentReportService
}

// DownloadRequirement 解析preview查询参数，下载或预览项目要求文件，
func (p *Projects) DownloadRequirement(w http.ResponseWriter, r *http.Request) error {}

// WatchStuSubmissions 查看学生完成情况
func (p *Projects) WatchStuSubmissions(w http.ResponseWriter, r *http.Request) error {}

// SwiftProjectStatus 根据请求体的status值，开启或关闭项目
func (p *Projects) SwiftProjectStatus(w http.ResponseWriter, r *http.Request) error {}

// UploadRequirement form-data格式，上传要求文件
func (p *Projects) UploadRequirement(w http.ResponseWriter, r *http.Request) error {

}

// DeleteProject 删除项目
func (p *Projects) DeleteProject(w http.ResponseWriter, r *http.Request) error {}

// UploadReport 上传实验报告
func (p *Projects) UploadReport(w http.ResponseWriter, r *http.Request) error {}

// Submissions 实验报告资源
// 在下载/预览单个报告时作为主资源
// 其他时候仅作为项目的子资源
type Submissions struct {
	tecReports *service.TeacherReportService
}

// DownSubmission 根据查询参数preview的值，下载或预览实验报告
func (s *Submissions) DownSubmission(w http.ResponseWriter, r *http.Request) error {}

// DownSubmissionBatch 打包下载学生实验报告
func (s *Submissions) DownSubmissionBatch(w http.ResponseWriter, r *http.Request) error {}

// Users 用户资源，包括用户信息操作
type Users struct{}

// PasswordResets 忘记密码
type PasswordResets struct{}
