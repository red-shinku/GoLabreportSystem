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

type StuReportForm struct {
	StudentID  string
	ProjectID  uint
	Format     string
	SubmitTime time.Time
}
