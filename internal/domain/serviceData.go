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

// 用于文件业务（与文件系统操作相关）的标记错误
var (
	ErrNotSafe  = errors.New("not safe")
	ErrNotAllow = errors.New("not allow")
	ErrNotExist = errors.New("not exist")
)

// ErrSheetFormat 表格解析错误：文件打不开、表头缺失、列数不足等
var ErrSheetFormat = errors.New("invalid sheet format")

// TODO: 当基文件夹不存在时创建？
const baseDir string = "./files"

var fileFmtSet map[string]struct{}

type LoginUserInfo struct {
	Identity uint8
	Number   string
}

// NewLoginUserInfo 创建LoginUserInfo实例
func NewLoginUserInfo(identity uint8, number string) *LoginUserInfo {
	return &LoginUserInfo{
		Identity: identity,
		Number:   number,
	}
}

//===========================================
// 结构化的视图，在业务层生成并返回给上层
// 上层使用该类结构生成新的、包含相关URL的结构类型的实例
//===========================================

// StudentProjectView 学生端的树状项目视图
type StudentProjectView struct {
	CourseName string
	Projects   []ProjectStuItem
}

// NewStudentProjectView 创建StudentProjectView实例
func NewStudentProjectView(courseName string) *StudentProjectView {
	return &StudentProjectView{
		CourseName: courseName,
		Projects:   make([]ProjectStuItem, 0),
	}
}

type ProjectStuItem struct {
	*ProjectItem
	SubmitStatus bool
	StuReportID  uint
}

func NewProjectStuItem(projectName string, startTime, closeTime time.Time, isActive bool, projectID uint, sbStatus bool, stuReportID uint) *ProjectStuItem {
	return &ProjectStuItem{
		ProjectItem:  NewProjectItem(projectName, startTime, closeTime, isActive, projectID),
		SubmitStatus: sbStatus,
		StuReportID:  stuReportID,
	}
}

// TeacherProjectView 教师端的树状项目视图
type TeacherProjectView struct {
	CourseName string
	Classes    []ClassItem
}

// NewTeacherProjectView 创建TeacherProjectView实例
func NewTeacherProjectView(courseName string) *TeacherProjectView {
	return &TeacherProjectView{
		CourseName: courseName,
		Classes:    make([]ClassItem, 0),
	}
}

type ClassItem struct {
	ClassName  string
	OfferingID uint
	Projects   []ProjectTecItem
}

// NewClassItem 创建ClassItem实例
func NewClassItem(className string, offeringID uint) *ClassItem {
	return &ClassItem{
		ClassName:  className,
		OfferingID: offeringID,
		Projects:   make([]ProjectTecItem, 0),
	}
}

type ProjectTecItem struct {
	*ProjectItem
}

func NewProjectTecItem(projectName string, startTime, closeTime time.Time, isActive bool, projectID uint) *ProjectTecItem {
	return &ProjectTecItem{
		NewProjectItem(projectName, startTime, closeTime, isActive, projectID),
	}
}

type ProjectItem struct {
	ProjectName string
	StartTime   time.Time
	CloseTime   time.Time
	IsActive    bool
	ProjectID   uint
}

// NewProjectItem 创建ProjectItem实例
func NewProjectItem(projectName string, startTime, closeTime time.Time, isActive bool, projectID uint) *ProjectItem {
	return &ProjectItem{
		ProjectName: projectName,
		StartTime:   startTime,
		CloseTime:   closeTime,
		IsActive:    isActive,
		ProjectID:   projectID,
	}
}

//=============================================
// 用于生成文件路径
// 包含路径生成规则、以及安全检查
//=============================================

// StuReportMeta 学生报告业务元数据，包含所属课程等信息
type StuReportMeta struct {
	// 查数据库获取
	CourseName  string
	ClassName   string
	StudentName string
	ProjectName string
	// 从表单获取
	StudentID string
	Format    string
}

// NewStuReportMeta 创建StuReportMeta实例
func NewStuReportMeta(courseName, className, studentName, projectName, studentID, format string) *StuReportMeta {
	return &StuReportMeta{
		CourseName:  courseName,
		ClassName:   className,
		StudentName: studentName,
		ProjectName: projectName,
		StudentID:   studentID,
		Format:      format,
	}
}

// FilePath 返回安全的文件路径
func (srm *StuReportMeta) FilePath() (string, error) {
	filename := fmt.Sprintf("%s-%s-%s.%s", srm.StudentID, srm.StudentName, srm.ProjectName, srm.Format)
	path := filepath.Join(baseDir, srm.CourseName, srm.ClassName, srm.ProjectName, filename)
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, baseDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("StuReportMeta(): path %w", ErrNotSafe)
	}
	return cleanPath, nil
}

// Check 检查文件格式
func (srm *StuReportMeta) Check() error {
	if _, ok := fileFmtSet[srm.Format]; !ok {
		return ErrNotAllow
	}
	return nil
}

// ProjectFileMeta 项目要求文件的业务元数据
type ProjectFileMeta struct {
	CourseName  string
	ClassName   string
	ProjectName string
	FileName    string
}

// NewProjectFileMeta 创建ProjectFileMeta实例
func NewProjectFileMeta(courseName, className, projectName, fileName string) *ProjectFileMeta {
	return &ProjectFileMeta{
		CourseName:  courseName,
		ClassName:   className,
		ProjectName: projectName,
		FileName:    fileName,
	}
}

// FilePath 返回安全的文件路径
func (pfm *ProjectFileMeta) FilePath() (string, error) {
	path := filepath.Join(baseDir, pfm.CourseName, pfm.ClassName, pfm.ProjectName, pfm.FileName)
	cleanpath, err := pfm.cleanPath(path)
	if err != nil {
		return "", err
	}
	return cleanpath, nil
}

func (pfm *ProjectFileMeta) DirectoryPath() (string, error) {
	path := filepath.Join(baseDir, pfm.CourseName, pfm.ClassName, pfm.ProjectName)
	cleanpath, err := pfm.cleanPath(path)
	if err != nil {
		return "", err
	}
	return cleanpath, nil
}

func (pfm *ProjectFileMeta) cleanPath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	if !strings.HasPrefix(cleanPath, baseDir+string(os.PathSeparator)) {
		return "", fmt.Errorf("StuReportMeta(): path %w", ErrNotSafe)
	}
	return cleanPath, nil
}

type TargetPaths struct {
	paths []string
	count int
	cur   int
}

func NewTargetPaths(paths []string) *TargetPaths {
	return &TargetPaths{
		paths: paths,
		count: len(paths),
		cur:   0,
	}
}

func (t *TargetPaths) Next() string {
	if t.cur < t.count {
		index := t.cur
		t.cur++
		return t.paths[index]
	}
	return ""
}

func (t *TargetPaths) HasNext() bool {
	return t.cur < t.count
}

//=============================================
// 课程信息表导入相关
//=============================================

// StudentRow 表格中一行学生记录
type StudentRow struct {
	Number string
	Name   string
}

// NewStudentRow 创建StudentRow实例
func NewStudentRow(number, name string) *StudentRow {
	return &StudentRow{Number: number, Name: name}
}

// SheetData 表格解析器的标准输出
// 与具体文件格式解耦，由业务层消费
type SheetData struct {
	Students     []StudentRow
	ProjectNames []string
}

// NewSheetData 创建SheetData实例
func NewSheetData(students []StudentRow, projectNames []string) *SheetData {
	return &SheetData{
		Students:     students,
		ProjectNames: projectNames,
	}
}

// ImportCourseData 教师导入课程信息时，从表单与上下文收集到的元数据
type ImportCourseData struct {
	TeacherID        string
	CourseName       string
	ClassName        string
	Term             string
	DefaultCloseTime time.Time
}

// NewImportCourseData 创建ImportCourseData实例
func NewImportCourseData(teacherID, courseName, className, term string, defaultCloseTime time.Time) *ImportCourseData {
	return &ImportCourseData{
		TeacherID:        teacherID,
		CourseName:       courseName,
		ClassName:        className,
		Term:             term,
		DefaultCloseTime: defaultCloseTime,
	}
}
