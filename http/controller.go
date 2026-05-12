package controller

// HTTP控制层，该文件定义最终的HTTP处理器
//
// 负责解析 HTTP 的请求头、请求体等，获取请求内容，并调用业务逻辑
// 根据 RESTful API ，以资源划分处理器，命名格式为 XxxHandler
// XxxHandler 会使用 service 中对应的业务类型

import (
	"LabSystem/api"
	"LabSystem/http/middleware"
	"LabSystem/http/route"
	html "LabSystem/http/template"
	"LabSystem/http/view"
	"LabSystem/internal/domain"
	"LabSystem/internal/httperr"
	"LabSystem/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewSessions(authService *service.AuthService, jwtSecret string) *Sessions {
	return &Sessions{
		auth:      authService,
		jwtSecret: []byte(jwtSecret),
	}
}

func NewHome(
	tecPjService *service.TeacherProjectService,
	stuPjService *service.StudentProjectService,
	lgPageGen *html.LoginPageGenerator,
	stuHomeGen *html.StuHomeGenerator,
	tecHomeGen *html.TecHomeGenerator) *Home {
	return &Home{
		tecProjects: tecPjService,
		stuProjects: stuPjService,
		lgPage:      lgPageGen,
		stuHome:     stuHomeGen,
		tecHome:     tecHomeGen,
	}
}

func NewOfferingClass(tecPjService *service.TeacherProjectService, tecHome *html.TecHomeGenerator) *OfferingClass {
	return &OfferingClass{
		tecProjects: tecPjService,
		tecHome:     tecHome,
	}
}

func NewProjects(
	tecPjService *service.TeacherProjectService,
	stuPjService *service.StudentProjectService,
	tecRpService *service.TeacherReportService,
	stuRpService *service.StudentReportService,
	tecHomeGen *html.TecHomeGenerator,
	stuHomeGen *html.StuHomeGenerator) *Projects {
	return &Projects{
		tecProjects: tecPjService,
		stuProjects: stuPjService,
		tecReports:  tecRpService,
		stuReports:  stuRpService,
		tecHome:     tecHomeGen,
		stuHome:     stuHomeGen,
	}
}

func NewSubmissions(tecRpService *service.TeacherReportService) *Submissions {
	return &Submissions{
		tecReports: tecRpService,
	}
}

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
	auth      *service.AuthService
	jwtSecret []byte
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

	// TODO: 加入配置选项
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
		s.jwtSecret,
	)
	authJWTCk := http.Cookie{
		Name:     "auth_token",
		Value:    authJwt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &authJWTCk)

	// 使用 HTMX 重定向到用户界面
	//http.Redirect(w, r, "/", http.StatusMovedPermanently)
	w.Header().Set("HX-Redirect", "/")
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
	userID, okU := ctx.Value(middleware.CtxKeyUserID).(string)
	role, okR := ctx.Value(middleware.CtxKeyRole).(string)
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
	tecHome     *html.TecHomeGenerator
}

// CreateProject 在url中解析到的offeringId下，新建项目，返回新项目卡片HTML片段
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
	item, err := o.tecProjects.CreateProject(form)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return o.tecHome.ProjectCard(w, view.BuildProjectTecItemWithUrl(item))
}

// Projects 与项目资源相关的
type Projects struct {
	tecProjects *service.TeacherProjectService
	stuProjects *service.StudentProjectService
	tecReports  *service.TeacherReportService
	stuReports  *service.StudentReportService
	tecHome     *html.TecHomeGenerator
	stuHome     *html.StuHomeGenerator
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

// WatchStuSubmissions 查看学生完成情况，返回HTML片段给模态框
func (p *Projects) WatchStuSubmissions(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	statuses, err := p.tecReports.CheckStuReportStatus(projectID)
	if err != nil {
		return err
	}

	return p.tecHome.SubmissionList(w, statuses)
}

// SwiftProjectStatus 根据请求体的status值，开启或关闭项目，返回刷新后的项目卡片片段
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

	item, err := p.tecProjects.ChangeProjectStatus(projectID)
	if err != nil {
		return err
	}

	return p.tecHome.ProjectCard(w, view.BuildProjectTecItemWithUrl(item))
}

// UploadRequirement form-data格式，上传要求文件；返回空片段以收起上传表单
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return nil
}

// DeleteProject 删除项目，返回空片段让 HTMX 移除对应卡片
func (p *Projects) DeleteProject(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	if err := p.tecProjects.DeleteProject(projectID); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return nil
}

// UploadReport 上传实验报告，返回刷新后的学生端项目卡片片段
func (p *Projects) UploadReport(w http.ResponseWriter, r *http.Request) error {
	projectID, err := parseUintPath(r, "projectId")
	if err != nil {
		return err
	}

	studentID, ok := r.Context().Value(middleware.CtxKeyUserID).(string)
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
	item, err := p.stuReports.UploadStuReport(file, form)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return p.stuHome.ProjectCard(w, view.BuildProjectStuItemWithUrl(item))
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

// Courses 课程资源
type Courses struct {
	importer *service.CourseImportService
}

func NewCourses(importer *service.CourseImportService) *Courses {
	return &Courses{importer: importer}
}

// ImportCourse 教师上传课程信息表（Excel）
// multipart 表单字段：filename(file), courseName, className, term, closeTime
// closeTime 使用 HTML datetime-local 格式（2006-01-02T15:04），按服务器本地时区解析；
// 兼容带秒格式 2006-01-02T15:04:05。className 缺省时由 service 兜底为 '-'
func (c *Courses) ImportCourse(w http.ResponseWriter, r *http.Request) error {
	teacherID, ok := r.Context().Value(middleware.CtxKeyUserID).(string)
	if !ok || teacherID == "" {
		return httperr.WithStatus(
			fmt.Errorf("Courses.ImportCourse(): missing teacher id in context"),
			http.StatusUnauthorized)
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return httperr.WithStatus(err, http.StatusBadRequest)
	}

	file, _, err := r.FormFile("filename")
	if err != nil {
		return httperr.WithStatus(err, http.StatusBadRequest)
	}
	defer file.Close()

	closeTimeStr := r.FormValue("closeTime")
	closeTime, err := time.ParseInLocation("2006-01-02T15:04", closeTimeStr, time.Local)
	if err != nil {
		closeTime, err = time.ParseInLocation("2006-01-02T15:04:05", closeTimeStr, time.Local)
	}
	if err != nil {
		closeTime, err = time.Parse(time.RFC3339, closeTimeStr)
	}
	if err != nil {
		return httperr.WithStatus(
			fmt.Errorf("Courses.ImportCourse(): closeTime: %v", err),
			http.StatusBadRequest)
	}

	data := domain.NewImportCourseData(
		teacherID,
		strings.TrimSpace(r.FormValue("courseName")),
		strings.TrimSpace(r.FormValue("className")),
		strings.TrimSpace(r.FormValue("term")),
		closeTime,
	)
	if data.CourseName == "" || data.Term == "" {
		return httperr.WithStatus(
			fmt.Errorf("Courses.ImportCourse(): courseName and term are required"),
			http.StatusBadRequest)
	}

	if err := c.importer.Import(file, data); err != nil {
		if errors.Is(err, domain.ErrSheetFormat) {
			return httperr.WithStatus(err, http.StatusBadRequest)
		}
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}
