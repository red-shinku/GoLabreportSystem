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

// NewUserInfo 创建UserInfo实例
func NewUserInfo(identity uint8, number, passwd string) *UserInfo {
	return &UserInfo{
		Identity: identity,
		Number:   number,
		Passwd:   passwd,
	}
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

// NewStudentProjectInfo 创建StudentProjectInfo实例
func NewStudentProjectInfo(courseName string, projectID uint, projectName string, startTime, closeTime time.Time, isActive bool, stuReportID sql.NullInt32) *StudentProjectInfo {
	return &StudentProjectInfo{
		CourseName:  courseName,
		ProjectID:   projectID,
		ProjectName: projectName,
		StartTime:   startTime,
		CloseTime:   closeTime,
		IsActive:    isActive,
		StuReportID: stuReportID,
	}
}

// TeacherProjectInfo 扁平化的教师管理项目列表，返回给上层
type TeacherProjectInfo struct {
	CourseName  string
	ClassName   string
	OfferingID  uint
	ProjectID   uint
	ProjectName string
	CloseTime   time.Time
	IsActive    bool
}

// NewTeacherProjectInfo 创建TeacherProjectInfo实例
func NewTeacherProjectInfo(courseName, className string, offeringID, projectID uint, projectName string, closeTime time.Time, isActive bool) *TeacherProjectInfo {
	return &TeacherProjectInfo{
		CourseName:  courseName,
		ClassName:   className,
		OfferingID:  offeringID,
		ProjectID:   projectID,
		ProjectName: projectName,
		CloseTime:   closeTime,
		IsActive:    isActive,
	}
}

// ProjectInfo 要插入表中的项目信息单行数据
type ProjectInfo struct {
	OfferingID      uint
	ProjectName     string
	ProjectFilePath string
	StartTime       time.Time
	CloseTime       time.Time
}

// NewProjectInfo 创建ProjectInfo实例
func NewProjectInfo(offeringID uint, projectName, projectFilePath string, startTime, closeTime time.Time) *ProjectInfo {
	return &ProjectInfo{
		OfferingID:      offeringID,
		ProjectName:     projectName,
		ProjectFilePath: projectFilePath,
		StartTime:       startTime,
		CloseTime:       closeTime,
	}
}

type ProjectInfoBuilder struct {
	offeringID      uint
	projectName     string
	projectFilePath string
	startTime       time.Time
	closeTime       time.Time
}

//func NewProjectInfoBuilder() *ProjectInfoBuilder {
//	return &ProjectInfoBuilder{}
//}
//
//func (pb *ProjectInfoBuilder) SetOfferingID(id uint) *ProjectInfoBuilder {
//	pb.offeringID = id
//	return pb
//}
//
//func (pb *ProjectInfoBuilder) SetProjectName(name string) *ProjectInfoBuilder {
//	pb.projectName = name
//	return pb
//}
//
//func (pb *ProjectInfoBuilder) SetProjectFilePath(path string) *ProjectInfoBuilder {
//	pb.projectFilePath = path
//	return pb
//}
//
//func (pb *ProjectInfoBuilder) SetStartTime(startTime time.Time) *ProjectInfoBuilder {
//	pb.startTime = startTime
//	return pb
//}
//
//func (pb *ProjectInfoBuilder) SetCloseTime(closeTime time.Time) *ProjectInfoBuilder {
//	pb.closeTime = closeTime
//	return pb
//}
//
//func (pb *ProjectInfoBuilder) Build() (*ProjectInfo, error) {
//	if pb.projectFilePath == "" {
//		return nil, fmt.Errorf("ProjectInfo Bilder: projectFilePath is empty")
//	}
//	if pb.projectName == "" {
//		return nil, fmt.Errorf("ProjectInfo Bilder: projectName is empty")
//	}
//	if pb.closeTime.IsZero() {
//		return nil, fmt.Errorf("ProjectInfo Bilder: closeTime is empty")
//	}
//	return &ProjectInfo{
//		OfferingID:      pb.offeringID,
//		ProjectName:     pb.projectName,
//		ProjectFilePath: pb.projectFilePath,
//		StartTime:       pb.startTime,
//		CloseTime:       pb.closeTime,
//	}, nil
//}

// StuReportStatus 一个项目中学生的完成情况
type StuReportStatus struct {
	StudentID   string
	StuReportID sql.NullInt32
}

// NewStuReportStatus 创建StuReportStatus实例
func NewStuReportStatus(studentID string, stuReportID sql.NullInt32) *StuReportStatus {
	return &StuReportStatus{
		StudentID:   studentID,
		StuReportID: stuReportID,
	}
}

// StuReportInfo 要插入表中的学生报告信息单行数据
type StuReportInfo struct {
	StudentID      string
	ProjectID      uint
	ReportFilePath string
	SubmitTime     time.Time
}

// NewStuReportInfo 创建StuReportInfo实例
func NewStuReportInfo(studentID string, projectID uint, reportFilePath string, submitTime time.Time) *StuReportInfo {
	return &StuReportInfo{
		StudentID:      studentID,
		ProjectID:      projectID,
		ReportFilePath: reportFilePath,
		SubmitTime:     submitTime,
	}
}

//type StuReportInfoBuilder struct {
//	StudentID      string
//	ProjectID      uint
//	ReportFilePath string
//	SubmitTime     time.Time
//}
//
//func NewStuReportInfoBuilder() *StuReportInfoBuilder {
//	return &StuReportInfoBuilder{}
//}
//
//func (pb *StuReportInfoBuilder) SetStudentID(id string) *StuReportInfoBuilder {
//	pb.StudentID = id
//	return pb
//}
//
//func (pb *StuReportInfoBuilder) SetProjectID(id uint) *StuReportInfoBuilder {
//	pb.ProjectID = id
//	return pb
//}
//
//func (pb *StuReportInfoBuilder) SetReportFilePath(path string) *StuReportInfoBuilder {
//	pb.ReportFilePath = path
//	return pb
//}
//
//func (pb *StuReportInfoBuilder) Build() (*StuReportInfo, error) {
//	if pb.ReportFilePath == "" {
//		return nil, fmt.Errorf("StuReportInfo Builder: ReportFilePath is empty")
//	}
//	return &StuReportInfo{
//		StudentID:      pb.StudentID,
//		ProjectID:      pb.ProjectID,
//		ReportFilePath: pb.ReportFilePath,
//		SubmitTime:     pb.SubmitTime,
//	}, nil
//}

// StudentCourseInfo 扁平化的学生选课信息
type StudentCourseInfo struct {
	StudentID  string
	OfferingID uint
}

// NewStudentCourseInfo 创建StudentCourseInfo实例
func NewStudentCourseInfo(studentID string, offeringID uint) *StudentCourseInfo {
	return &StudentCourseInfo{
		StudentID:  studentID,
		OfferingID: offeringID,
	}
}
