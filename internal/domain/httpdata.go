package domain

// 该文件定义 HTTP 层最终交给业务层的所有数据

import "time"

//FIXME: 移至独立的api层（请求与响应的json结构）

// ProjectData 当新建项目时，从表单或请求中获取的相关数据
// 填充完毕后会转交给业务层，用于下一步的DTO构建
type ProjectData struct {
	OfferingID  uint
	ProjectName string
	CloseTime   time.Time
}

// NewProjectData 创建ProjectForm实例
func NewProjectData(offeringID uint, projectName string, closeTime time.Time) *ProjectData {
	return &ProjectData{
		OfferingID:  offeringID,
		ProjectName: projectName,
		CloseTime:   closeTime,
	}
}

// ProjectFileData 当重传项目要求文件时，从表单或请求中获取的数据
type ProjectFileData struct {
	ProjectID uint
	FileName  string
}

func NewProjectFileData(projectID uint, fileName string) *ProjectFileData {
	return &ProjectFileData{
		ProjectID: projectID,
		FileName:  fileName,
	}
}

type StuReportData struct {
	StudentID string
	ProjectID uint
	Format    string
}

// NewStuReportData 创建StuReportForm实例
func NewStuReportData(studentID string, projectID uint, format string) *StuReportData {
	return &StuReportData{
		// 控制层获取token中的学生ID
		StudentID: studentID,
		// 控制层从URL获取项目ID
		ProjectID: projectID,
		// 控制层获取文件格式并填充
		Format: format,
	}
}
