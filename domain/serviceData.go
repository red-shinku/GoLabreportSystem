package domain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrAuth = errors.New("failed to auth")

var (
	ErrNotSafe  = errors.New("not safe")
	ErrNotAllow = errors.New("not allow")
)

const baseDir string = "./files"

var fileFmtSet map[string]struct{}

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

// StuReportMeta 学生报告业务元数据，包含所属课程等信息
type StuReportMeta struct {
	courseName  string
	className   string
	studentID   string
	studentName string
	projectName string
	format      string
}

// FilePath 返回安全的文件路径
func (srm *StuReportMeta) FilePath() (string, error) {
	filename := fmt.Sprintf("%s-%s-%s.%s", srm.studentID, srm.studentName, srm.projectName, srm.format)
	path := filepath.Join(baseDir, srm.courseName, srm.className, srm.projectName, filename)
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, baseDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("StuReportMeta(): path %w", ErrNotSafe)
	}
	return cleanPath, nil
}

// Check 检查文件格式
func (srm *StuReportMeta) Check() error {
	if _, ok := fileFmtSet[srm.format]; !ok {
		return ErrNotAllow
	}
	return nil
}

// ProjectFileMeta 项目要求文件的业务元数据
type ProjectFileMeta struct {
	courseName  string
	className   string
	projectName string
	fileName    string
}

// FilePath 返回安全的文件路径
func (pfm *ProjectFileMeta) FilePath() (string, error) {
	path := filepath.Join(baseDir, pfm.courseName, pfm.className, pfm.projectName, pfm.fileName)
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, baseDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("StuReportMeta(): path not save")
	}
	return cleanPath, nil
}
