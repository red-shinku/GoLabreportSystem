// 该文件定义DTO数据结构：扁平化的数据库表的行内容，
// 它们由数据库操作层查询并返回给上层处理
// 或者由数据库操作层从上层接收并将其写入数据库

package domain

import (
	"database/sql"
	"errors"
	"time"
)

// error flag
var (
	//FIXME: SQL操作错误不返回给上层？
	ErrQuery    = errors.New("query failed")
	ErrModify   = errors.New("modify database failed")
	ErrNotFound = errors.New("entry not found")
)

// UserInfo 用户信息，用于初始化插入用户表
type UserInfo struct {
	Identity uint8
	Number   string
	Passwd   string
}

// StudentProjectInfo 扁平化的学生项目列表查询结果，返回给上层
type StudentProjectInfo struct {
	CourseName  string
	ProjectID   uint
	ProjectName string
	StartTime   time.Time
	CloseTime   time.Time
	IsActive    bool
	StuReportID sql.NullInt32
}

// TeacherProjectInfo 扁平化的教师管理项目列表，返回给上层
type TeacherProjectInfo struct {
	CourseName  string
	ClassName   string
	ProjectID   uint
	ProjectName string
	CloseTime   time.Time
	IsActive    bool
}

// ProjectInfo 要插入表中的项目信息单行数据
type ProjectInfo struct {
	OfferingID      uint
	ProjectName     string
	ProjectFilePath string
	StartTime       time.Time
	CloseTime       time.Time
}

// StuReportStatus 一个项目中学生的完成情况
type StuReportStatus struct {
	StudentID   string
	StuReportID sql.NullInt32
}

// StuReportInfo 要插入表中的学生报告信息单行数据
type StuReportInfo struct {
	StudentID      string
	ProjectID      uint
	ReportFilePath string
	SubmitTime     time.Time
}

// StudentCourseInfo 扁平化的学生选课信息
type StudentCourseInfo struct {
	StudentID  string
	OfferingID uint
}
