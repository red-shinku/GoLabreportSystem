package view

import (
	"LabSystem/domain"
	"time"
)

//TODO: 给html生成模块使用的最终View结果，包含视图

// 教师的项目视图
// 是一个树状结构：课程列表 -> 班级列表 -> 项目列表

type TecProjectViewWithUrl struct {
	Courses []CourseTecItemWithUrl
}

type CourseTecItemWithUrl struct {
	CourseName string
	Classes    []ClassTecItemWithUrl
}

type ClassTecItemWithUrl struct {
	ClassName string
	// URL：新建项目
	CreateProject string
	Projects      []ProjectTecItemWithUrl
}

type ProjectTecItemWithUrl struct {
	ProjectName string
	StartTime   time.Time
	CloseTime   time.Time
	Status      bool
	// URL：提供（上传报告要求文件）
	Afford string
	// URL：检查（预览）报告要求文件
	Check string
	// URL：变更项目开放状态
	SwiftStatus string
	// URL：删除项目
	Delete string
}

func BuildTecProjectViewWithUrl(serviceView *domain.TeacherProjectView) *TecProjectViewWithUrl {

}

// 学生的项目视图
// 是一个树状结构：课程列表 -> 项目列表

type StuProjectViewWithUrl struct {
	Courses []CourseStuItemWithUrl
}

type CourseStuItemWithUrl struct {
	CourseName string
	Projects   []ProjectStuItemWithUrl
}

type ProjectStuItemWithUrl struct {
	ProjectName string
	StartTime   time.Time
	CloseTime   time.Time
	IsSubmit    bool
	Status      bool
	// URL: 提交报告
	Submit string
	// URL: 检查（预览）报告
	Check string
}

func BuildStuProjectViewWithUrl(serviceView *domain.StudentProjectView) *StuProjectViewWithUrl {
	
}
