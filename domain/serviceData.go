package domain

import (
	"errors"
	"time"
)

var ErrAuth = errors.New("failed to auth")

type LoginUserInfo struct {
	Identity uint8
	Number   string
}

// StudentProjectView 学生端的树状项目视图
type StudentProjectView struct {
	CourseName string
	Projects   []ProjectItem
}

// TeacherProjectView 教师端的树状项目视图
type TeacherProjectView struct {
	CourseName string
	Classes    []ClassItem
}

type ClassItem struct {
	ClassName string
	Projects  []ProjectItem
}

type ProjectItem struct {
	ProjectName string
	StartTime   time.Time
	CloseTime   time.Time
	ProjectID   uint
	// 用于学生端显示提交状态
	SubmitStatus bool
}
