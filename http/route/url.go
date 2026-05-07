package route

import (
	"fmt"
	"net/url"
)

const APIBase = "/api/v1"

const QueryKeyPreview = "preview"

// WithQuery 为 URL 附加查询参数
func WithQuery(rawURL string, params url.Values) string {
	return rawURL + "?" + params.Encode()
}

// SessionsURL 登录认证（POST 创建会话）
func SessionsURL() string {
	return APIBase + "/sessions"
}

// ProjectsURL 获取项目列表（GET）/ 新建项目（POST）
func ProjectsURL() string {
	return APIBase + "/projects"
}

// ProjectURL 操作单个项目：修改项目状态（PATCH）/ 删除项目（DELETE）
func ProjectURL(projectID uint) string {
	return fmt.Sprintf("%s/projects/%d", APIBase, projectID)
}

// ProjectRequirementURL 下载项目要求文件（GET）/ 上传或重传项目要求文件（PUT）
func ProjectRequirementURL(projectID uint) string {
	return fmt.Sprintf("%s/projects/%d/requirement", APIBase, projectID)
}

// ProjectSubmissionsURL 查看学生完成情况（GET）/ 学生提交实验报告（POST）
func ProjectSubmissionsURL(projectID uint) string {
	return fmt.Sprintf("%s/projects/%d/submissions", APIBase, projectID)
}

// ProjectSubmissionsArchiveURL 打包下载该项目下所有学生报告（GET）
func ProjectSubmissionsArchiveURL(projectID uint) string {
	return fmt.Sprintf("%s/projects/%d/submissions/archive", APIBase, projectID)
}

// SubmissionFileURL 下载或预览单份已提交的实验报告文件（GET）
func SubmissionFileURL(submissionID uint) string {
	return fmt.Sprintf("%s/submissions/%d/file", APIBase, submissionID)
}

// UserEmailURL 绑定用户邮箱（POST）
func UserEmailURL() string {
	return APIBase + "/users/me/email"
}

// PasswordResetsURL 请求密码重置（POST）
func PasswordResetsURL() string {
	return APIBase + "/password-resets"
}

// PasswordResetURL 执行密码重置（PUT）
func PasswordResetURL(token string) string {
	return fmt.Sprintf("%s/password-resets/%s", APIBase, token)
}
