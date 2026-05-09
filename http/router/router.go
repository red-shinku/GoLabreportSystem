package server

import (
	controller "LabSystem/http"
	"LabSystem/http/middleware"
	"net/http"
)

// 该文件注册路由
// 根据 HTTP 方法以及 URL 路由到对应的后端接口

type Router struct {
	mux           *http.ServeMux
	home          controller.Home
	sessions      controller.Sessions
	offeringClass controller.OfferingClass
	projects      controller.Projects
	submissions   controller.Submissions
}

func NewRouter(mux *http.ServeMux) *Router {
	return &Router{mux: mux}
}

func (r *Router) Init() {
	// 公开端点：Recovery -> Logger
	pub := func(h middleware.HandlerFunc) http.HandlerFunc {
		return middleware.Adapt(middleware.Recovery(middleware.Logger(h)))
	}
	// 需登录端点：Recovery -> Logger -> JwtValidator
	auth := func(h middleware.HandlerFunc) http.HandlerFunc {
		return middleware.Adapt(middleware.Recovery(middleware.Logger(middleware.JwtValidator(h))))
	}

	// 进入主界面，若未登录则重定向到登录页面；已登录则解析 JWT 身份
	r.mux.HandleFunc("GET /", middleware.Adapt(
		middleware.Recovery(
			middleware.Logger(
				middleware.LoginCheck(
					middleware.JwtValidator(r.home.HomePage))))))
	// 登录页面
	r.mux.HandleFunc("GET /sessions", pub(r.home.LoginPage))

	// 会话（登录）
	r.mux.HandleFunc("POST /api/v1/sessions", pub(r.sessions.Login))

	// 班级资源
	r.mux.HandleFunc("POST /api/v1/offeringclass/{offeringId}", auth(r.offeringClass.CreateProject))

	// 项目资源
	r.mux.HandleFunc("GET /api/v1/projects/{projectId}/requirement", auth(r.projects.DownloadRequirement))
	r.mux.HandleFunc("PUT /api/v1/projects/{projectId}/requirement", auth(r.projects.UploadRequirement))
	r.mux.HandleFunc("GET /api/v1/projects/{projectId}/submissions", auth(r.projects.WatchStuSubmissions))
	r.mux.HandleFunc("POST /api/v1/projects/{projectId}/submissions", auth(r.projects.UploadReport))
	r.mux.HandleFunc("GET /api/v1/projects/{projectId}/submissions/archive", auth(r.submissions.DownSubmissionBatch))
	r.mux.HandleFunc("PATCH /api/v1/projects/{projectId}", auth(r.projects.SwiftProjectStatus))
	r.mux.HandleFunc("DELETE /api/v1/projects/{projectId}", auth(r.projects.DeleteProject))

	// 实验报告资源
	r.mux.HandleFunc("GET /api/v1/submissions/{submissionId}/file", auth(r.submissions.DownSubmission))
}
