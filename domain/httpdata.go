package domain

import "time"

// ProjectForm 当新建项目时，从表单或请求中获取的相关数据
// 填充完毕后会转交给业务层，用于下一步的DTO构建
type ProjectForm struct {
	OfferingID  uint
	ProjectName string
	FileName    string
	CloseTime   time.Time
}

// NewProjectForm 创建ProjectForm实例
func NewProjectForm(offeringID uint, projectName, fileName string, closeTime time.Time) *ProjectForm {
	return &ProjectForm{
		OfferingID:  offeringID,
		ProjectName: projectName,
		FileName:    fileName,
		CloseTime:   closeTime,
	}
}

// ProjectFileForm 当重传项目要求文件时，从表单或请求中获取的数据
type ProjectFileForm struct {
	ProjectID uint
	FileName  string
}

func NewProjectFileForm(projectID uint, fileName string) *ProjectFileForm {
	return &ProjectFileForm{
		ProjectID: projectID,
		FileName:  fileName,
	}
}

type StuReportForm struct {
	StudentID string
	ProjectID uint
	Format    string
}

// NewStuReportForm 创建StuReportForm实例
func NewStuReportForm(studentID string, projectID uint, format string) *StuReportForm {
	return &StuReportForm{
		// 控制层获取token中的学生ID
		StudentID: studentID,
		// 控制层从URL获取项目ID
		ProjectID: projectID,
		// 控制层获取 Content-type并填充
		Format: format,
	}
}
