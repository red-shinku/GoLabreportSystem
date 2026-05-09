package api

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// 存放请求/响应的DTO

// LoginJWT 自定义JWT的结构
// 该JWT在通过登录时发放
// 除登录与忘记密码，其他接口均需要验证该JWT
type LoginJWT struct {
	UserID string `json:"user_id"`
	Role   string `json:"identity"`
	jwt.RegisteredClaims
}

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

// roleMap 身份码到身份的映射表，用于设置cookie
var roleMap = map[uint8]string{
	1: "student",
	2: "teacher",
	3: "operator",
}

// Role 根据身份码获取身份字符串
func Role(code uint8) (string, bool) {
	role, ok := roleMap[code]
	return role, ok
}

var identityMap = map[string]uint8{
	"student":  1,
	"teacher":  2,
	"operator": 3,
}

// Identity 根据身份字符串获取身份码
func Identity(role string) (uint8, bool) {
	code, ok := identityMap[role]
	return code, ok
}
