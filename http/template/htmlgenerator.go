package html

// HTML页面生成模块，只负责填充HTML模版

import (
	htmltpl "html/template"
	"net/http"

	"LabSystem/http/view"
)

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

// TecHomeGenerator 返回教师端页面。
// 接受上层传递的http.ResponseWriter接口、视图结构体
// 根据教师项目视图结构体，填写html模版并将其写入接口
type TecHomeGenerator struct{}

func (tg *TecHomeGenerator) Page(w http.ResponseWriter, v *view.TecProjectViewWithUrl) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return templates.ExecuteTemplate(w, "teacherhome.html", v)
}
