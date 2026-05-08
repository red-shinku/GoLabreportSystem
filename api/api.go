package api

import "time"

// 存放请求/响应的DTO

// LoginRequest 登录POST方法的请求体
type LoginRequest struct {
	UserNumber string `json:"userNumber"`
	Password   string `json:"password"`
}

// SwiftProjectStatusRequest 开放/关闭项目PATCH方法的请求体
type SwiftProjectStatusRequest struct {
	Status string `json:"status"`
}

// CreateProjectRequest 新建项目 POST 方法的请求体
type CreateProjectRequest struct {
	ProjectName string    `json:"projectname"`
	CloseTime   time.Time `json:"closeTime"`
}

// ProjectFileFormRequest 上传项目要求文件的表单数据
type ProjectFileFormRequest struct {
	Filename string `json:"filename"`
}

// ReportFileFormRequest 上传实验报告文件的表单数据
type ReportFileFormRequest struct {
	Filename string `json:"filename"`
}

// identityMap 身份码到身份的映射表，用于设置cookie
var identityMap = map[uint8]string{
	1: "student",
	2: "teacher",
	3: "operator",
}

// Identity 根据身份码获取身份字符串
func Identity(code uint8) (string, bool) {
	identity, ok := identityMap[code]
	return identity, ok
}

var identityCodeMap = map[string]uint8{
	"student":  1,
	"teacher":  2,
	"operator": 3,
}

// IdentityCode 根据身份字符串获取身份码
func IdentityCode(identity string) (uint8, bool) {
	code, ok := identityCodeMap[identity]
	return code, ok
}
