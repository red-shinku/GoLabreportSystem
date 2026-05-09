package controller

// HTTP控制层，该文件定义最终的HTTP处理器
//
// 负责解析 HTTP 的请求头、请求体等，获取请求内容，并调用业务逻辑
// 根据 RESTful API ，以资源划分处理器，命名格式为 XxxHandler
// XxxHandler 会使用 service 中对应的业务类型

import (
	"LabSystem/api"
	"LabSystem/http/route"
	html "LabSystem/http/template"
	"LabSystem/http/view"
	"LabSystem/internal/domain"
	"LabSystem/internal/httperr"
	"LabSystem/service"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 暂定义的JWT密钥
// TODO: JWT密钥获取
var secret []byte

// contextKey 上层中间件与控制层之间传递身份信息的 context 键类型
type contextKey string

const (
	// CtxKeyUserID 登录用户编号（学号/工号）
	CtxKeyUserID contextKey = "userID"
	// CtxKeyRole 登录用户身份（student / teacher / operator）
	CtxKeyRole contextKey = "role"
)

// parseUintPath 从 URL 路径变量解析 uint ID，失败返回 400
func parseUintPath(r *http.Request, name string) (uint, error) {
	raw := r.PathValue(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, httperr.WithStatus(
			fmt.Errorf("parse path %q: %v", name, err),
			http.StatusBadRequest)
	}
	return uint(id), nil
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
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
	return nil
}

// Home 生成完整主界面
type Home struct {
	tecProjects *service.TeacherProjectService
	stuProjects *service.StudentProjectService
	lgPage      *html.LoginPageGenerator
	stuHome     *html.StuHomeGenerator
	tecHome     *html.TecHomeGenerator
}

// LoginPage 返回登录页面
func (h *Home) LoginPage(w http.ResponseWriter, r *http.Request) error {
	return h.lgPage.Page(w)
}

// HomePage 根据解析到的cookie中的身份信息，返回对应的用户界面
func (h *Home) HomePage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID, okU := ctx.Value(CtxKeyUserID).(string)
	role, okR := ctx.Value(CtxKeyRole).(string)
	if !okU || !okR || userID == "" {
		return httperr.WithStatus(
			fmt.Errorf("Home.HomePage(): missing identity in context"),
			http.StatusUnauthorized)
	}

	switch role {
	case "student":
		views, err := h.stuProjects.ListProject(userID)
		if err != nil {
			return err
		}
		return h.stuHome.Page(w, view.BuildStuProjectViewWithUrl(views))
	case "teacher":
		views, err := h.tecProjects.ListProject(userID)
		if err != nil {
			return err
		}
		return h.tecHome.Page(w, view.BuildTecProjectViewWithUrl(views))
	default:
		return httperr.WithStatus(
			fmt.Errorf("Home.HomePage(): role %q not supported", role),
			http.StatusForbidden)
	}
}

// OfferingClass 授课班级及其子资源
type OfferingClass struct {
	tecProjects *service.TeacherProjectService
}

// CreateProject 在url中解析到的offeringId下，新建项目
func (o *OfferingClass) CreateProject(w http.ResponseWriter, r *http.Request) error {
	offeringID, err := parseUintPath(r, "offeringId")
	if err != nil {
		return err
	}

	var body api.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return httperr.WithStatus(err, http.StatusBadRequest)
	}

	form := domain.NewProjectData(offeringID, body.ProjectName, body.CloseTime)
	if err := o.tecProjects.CreateProject(form); err != nil {
		return err
	}

	//TODO: 需要更新前端页面，html层要生成html片段并返回
	w.WriteHeader(http.StatusCreated)
	return nil
}

// Projects 与项目资源相关的
type Projects struct {
	tecProjects *service.TeacherProjectService
	stuProjects *service.StudentProjectService
	tecReports  *service.TeacherReportService
	stuReports  *service.StudentReportService
}

// DownloadRequirement 解析preview查询参数，下载或预览项目要求文件，
func (p *Projects) DownloadRequirement(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	preview := r.URL.Query().Get(route.QueryKeyPreview) == "true"

	w.Header().Set("Content-Type", "application/octet-stream")
	if preview {
		w.Header().Set("Content-Disposition", "inline")
	} else {
		w.Header().Set("Content-Disposition", "attachment")
	}

	return p.stuProjects.DownloadProjectFile(w, projectID)
}

// WatchStuSubmissions 查看学生完成情况
func (p *Projects) WatchStuSubmissions(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	statuses, err := p.tecReports.CheckStuReportStatus(projectID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(statuses)
}

// SwiftProjectStatus 根据请求体的status值，开启或关闭项目
func (p *Projects) SwiftProjectStatus(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	var body api.SwiftProjectStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return httperr.WithStatus(err, http.StatusBadRequest)
	}
	if body.Status != "open" && body.Status != "closed" {
		return httperr.WithStatus(
			fmt.Errorf("invalid status %q", body.Status),
			http.StatusBadRequest)
	}

	if err := p.tecProjects.ChangeProjectStatus(projectID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// UploadRequirement form-data格式，上传要求文件
func (p *Projects) UploadRequirement(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return httperr.WithStatus(err, http.StatusBadRequest)
	}

	file, header, err := r.FormFile("filename")
	if err != nil {
		return httperr.WithStatus(err, http.StatusBadRequest)
	}
	defer file.Close()

	form := domain.NewProjectFileData(projectID, header.Filename)
	if err := p.tecProjects.UploadProjectFile(file, form); err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// DeleteProject 删除项目
func (p *Projects) DeleteProject(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	if err := p.tecProjects.DeleteProject(projectID); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

// UploadReport 上传实验报告
func (p *Projects) UploadReport(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	studentID, ok := r.Context().Value(CtxKeyUserID).(string)
	if !ok || studentID == "" {
		return httperr.WithStatus(
			fmt.Errorf("Projects.UploadReport(): missing student id in context"),
			http.StatusUnauthorized)
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return httperr.WithStatus(err, http.StatusBadRequest)
	}

	file, header, err := r.FormFile("filename")
	if err != nil {
		return httperr.WithStatus(err, http.StatusBadRequest)
	}
	defer file.Close()

	format := strings.TrimPrefix(filepath.Ext(header.Filename), ".")
	form := domain.NewStuReportData(studentID, projectID, format)
	if err := p.stuReports.UploadStuReport(file, form); err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

// Submissions 实验报告资源
// 在下载/预览单个报告时作为主资源
// 其他时候仅作为项目的子资源
type Submissions struct {
	tecReports *service.TeacherReportService
}

// DownSubmission 根据查询参数preview的值，下载或预览实验报告
func (s *Submissions) DownSubmission(w http.ResponseWriter, r *http.Request) error {
	if _, err := parseUintPath(r, "submissionId"); err != nil {
		return err
	}
	_ = r.URL.Query().Get(route.QueryKeyPreview) == "true"

	// TODO: TeacherReportService 尚未提供按 reportID 下载单份报告的方法，
	// 待补充（需要在 teacherReportOp 上增加 QueryStuReportFile(reportID)）
	return httperr.WithStatus(
		fmt.Errorf("Submissions.DownSubmission(): not implemented"),
		http.StatusNotImplemented)
}

// DownSubmissionBatch 打包下载学生实验报告
func (s *Submissions) DownSubmissionBatch(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="reports.zip"`)

	return s.tecReports.DownloadStuReportBatch(w, projectID)
}

// Users 用户资源，包括用户信息操作
type Users struct{}

// PasswordResets 忘记密码
type PasswordResets struct{}
