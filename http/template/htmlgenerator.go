package html

// HTML页面生成模块，负责填充HTML模版，
// 包括完整页面与 AJAX 返回的 HTML 片段

import (
	htmltpl "html/template"
	"net/http"

	"LabSystem/http/view"
	"LabSystem/internal/domain"
)

func NewLoginPageGenerator() *LoginPageGenerator {
	return &LoginPageGenerator{}
}

func NewStuHomeGenerator() *StuHomeGenerator {
	return &StuHomeGenerator{}
}

func NewTecHomeGenerator() *TecHomeGenerator {
	return &TecHomeGenerator{}
}

var templates = htmltpl.Must(htmltpl.ParseGlob("html/*.html"))

// LoginPageGenerator 返回登录页面。该页面无需额外构造，直接返回。
type LoginPageGenerator struct{}

func (lg *LoginPageGenerator) Page(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return templates.ExecuteTemplate(w, "loginpage.html", nil)
}

// StuHomeGenerator 返回学生端页面。
// 接受上层传递的http.ResponseWriter接口、视图结构体
// 根据学生项目视图结构体，填写html模版并将其写入接口
type StuHomeGenerator struct{}

func (sg *StuHomeGenerator) Page(w http.ResponseWriter, v *view.StuProjectViewWithUrl) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return templates.ExecuteTemplate(w, "studenthome.html", v)
}

// ProjectCard 返回单张学生端项目卡片片段（供 AJAX 回写）
func (sg *StuHomeGenerator) ProjectCard(w http.ResponseWriter, v *view.ProjectStuItemWithUrl) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return templates.ExecuteTemplate(w, "project-stu", v)
}

// TecHomeGenerator 返回教师端页面。
// 接受上层传递的http.ResponseWriter接口、视图结构体
// 根据教师项目视图结构体，填写html模版并将其写入接口
type TecHomeGenerator struct{}

func (tg *TecHomeGenerator) Page(w http.ResponseWriter, v *view.TecProjectViewWithUrl) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return templates.ExecuteTemplate(w, "teacherhome.html", v)
}

// ProjectCard 返回单张教师端项目卡片片段（供 AJAX 回写）
func (tg *TecHomeGenerator) ProjectCard(w http.ResponseWriter, v *view.ProjectTecItemWithUrl) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return templates.ExecuteTemplate(w, "project-tec", v)
}

// SubmissionList 返回学生完成情况列表片段（供模态框 AJAX 回写）
func (tg *TecHomeGenerator) SubmissionList(w http.ResponseWriter, statuses []domain.StuReportStatus) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return templates.ExecuteTemplate(w, "submission-list", statuses)
}
