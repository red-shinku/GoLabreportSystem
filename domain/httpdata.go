package domain

import "time"

// ProjectForm 当上传项目要求文件时，从表单获取的相关数据
// 填充完毕后会转交给业务层，用于下一步的DTO构建
type ProjectForm struct {
	OfferingID  uint
	ProjectID   uint
	ProjectName string
	FileName    string
	CloseTime   time.Time
}

// NewProjectForm 创建ProjectForm实例
func NewProjectForm(offeringID, projectID uint, projectName, fileName string, closeTime time.Time) *ProjectForm {
	return &ProjectForm{
		OfferingID:  offeringID,
		ProjectID:   projectID,
		ProjectName: projectName,
		FileName:    fileName,
		CloseTime:   closeTime,
	}
}

type StuReportForm struct {
	StudentID  string
	ProjectID  uint
	Format     string
	SubmitTime time.Time
}

// NewStuReportForm 创建StuReportForm实例
func NewStuReportForm(studentID string, projectID uint, format string, submitTime time.Time) *StuReportForm {
	return &StuReportForm{
		StudentID:  studentID,
		ProjectID:  projectID,
		Format:     format,
		SubmitTime: submitTime,
	}
}
