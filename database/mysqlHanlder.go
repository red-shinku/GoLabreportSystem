package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

// error flag
var (
	ErrQuery        = errors.New("query error")
	ErrUserNotFound = errors.New("user not found")
	ErrPasswd       = errors.New("password not correct")
	ErrGetPasswd    = errors.New("get password error")
)

// name of mysql table
const (
	tabUsers          string = "Users"
	tabCourse         string = "Course"
	tabCourseOffering string = "CourseOffering"
	tabStudentCourse  string = "StudentCourse"
	tabTeacherCourse  string = "TeacherCourse"
	tabProject        string = "Project"
	tabStuReport      string = "StuReport"
)

// FIXME: 改为纯函数，注意添加必要形参
// User :use for login、register and otherwise
type User struct {
	whoami uint8
	number string
}

//login method, may implement the Login interface
//after checking password successfully,
//other code can use Whoami() to obtain identity,
//then return corresponding website to user

func (u *User) InputAccount(number string) {
	u.number = number
}

func (u *User) CheckLogin(passwd string, db *sql.DB) error {
	var correctPasswd string
	var whoami uint8
	errQ := db.QueryRow(
		"select (passwd, identity) from %s where number = ?", tabUsers, u.number,
	).Scan(&correctPasswd, &whoami)

	if errQ != nil {
		if errors.Is(errQ, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("%w: %v", ErrQuery, errQ)
	}
	//FIXME: 密码的加密存储将如何处理
	if passwd != correctPasswd {
		return ErrPasswd
	}
	u.whoami = whoami
	return nil
}

func (u *User) Whoami() uint8 {
	return u.whoami
}

// queryTemplate 多行查询模版
func queryTemplate[T any](db *sql.DB, query string, args []any,
	scanFunc func(*sql.Rows, *T) error, result *[]T) error {
	rows, errQ := db.Query(query, args...)
	if errQ != nil {
		return fmt.Errorf("%w: %v", ErrQuery, errQ)
	}
	defer rows.Close()

	for rows.Next() {
		var row T
		if err := scanFunc(rows, &row); err != nil {
			return fmt.Errorf("%w: %v", ErrQuery, err)
		}
		*result = append(*result, row)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrQuery, err)
	}
	return nil
}

// StudentProjectRow 扁平化的学生项目列表查询结果，返回给上层
type StudentProjectRow struct {
	CourseName  string
	ProjectID   uint
	ProjectName string
	StartTime   time.Time
	CloseTime   time.Time
	IsActive    bool
	StuReportID sql.NullInt32
}

// QueryStuProject 查询一个学生的所有项目及其信息
func QueryStuProject(db *sql.DB, studentID string) ([]StudentProjectRow, error) {
	query := fmt.Sprintf("select c.courseName, p.projectID, p.projectName, p.startTime, p.deadline, p.isActive, srp.stuReportID "+
		"from %s stuc "+
		"join %s coff on stuc.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"join %s srp on p.projectID = srp.projectID and stuc.studentID = srp.studentID "+
		"where stuc.studentID = ?",
		tabStudentCourse, tabCourseOffering, tabCourse, tabProject, tabStuReport)
	scanFunc := func(rows *sql.Rows, row *StudentProjectRow) error {
		return rows.Scan(&row.CourseName, &row.ProjectID, &row.ProjectName,
			&row.StartTime, &row.CloseTime, &row.IsActive, &row.StuReportID)
	}
	args := []any{studentID}

	var result []StudentProjectRow
	if err := queryTemplate(db, query, args, scanFunc, &result); err != nil {
		return nil, fmt.Errorf("QueryStuProject failed: %w", err)
	}
	return result, nil
}

// TeacherProjectRow 扁平化的教师管理项目列表，返回给上层
type TeacherProjectRow struct {
	CourseName  string
	ClassName   string
	ProjectID   uint
	ProjectName string
	CloseTime   time.Time
	IsActive    bool
}

// QueryTeacherProject 查询一个教师所管理的项目，包含项目所属班级、课程等信息
func QueryTeacherProject(db *sql.DB, teacherID string) ([]TeacherProjectRow, error) {
	query := fmt.Sprintf("select c.courseName, coff.className, p.projectID, p.projectName, p.deadline, p.isActive "+
		"from %s tec "+
		"join %s coff on tec.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"where teacherID = ?",
		tabTeacherCourse, tabCourseOffering, tabCourse, tabProject)
	scanFunc := func(rows *sql.Rows, row *TeacherProjectRow) error {
		return rows.Scan(&row.CourseName, &row.ClassName, &row.ProjectID, &row.ProjectName,
			&row.CloseTime, &row.IsActive)
	}
	args := []any{teacherID}

	var result []TeacherProjectRow
	if err := queryTemplate(db, query, args, scanFunc, &result); err != nil {
		return nil, fmt.Errorf("QueryTeacherProject() failed: %w", err)
	}
	return result, nil
}

// StuProjectStatus 一个项目中学生的完成情况
type StuProjectStatus struct {
	StudentID   string
	StuReportID sql.NullInt32
}

// QueryStuProjectStatus 教师查询某项目下学生的完成情况
func QueryStuProjectStatus(db *sql.DB, projectID uint) ([]StuProjectStatus, error) {
	query := fmt.Sprintf("select srp.studentID, srp.stuReportID "+
		"from %s p "+
		"join %s srp on p.projectID = srp.projectID "+
		"where projectID = ?",
		tabProject, tabStuReport)
	scanFunc := func(rows *sql.Rows, row *StuProjectStatus) error {
		return rows.Scan(&row.StudentID, &row.StuReportID)
	}
	args := []any{projectID}

	var result []StuProjectStatus
	if err := queryTemplate(db, query, args, scanFunc, &result); err != nil {
		return nil, fmt.Errorf("QueryStuProjectStatus() failed: %w", err)
	}
	return result, nil
}

// UpdateProjectFlag 教师开启/关闭项目
func UpdateProjectFlag(db *sql.DB, projectID uint, flag bool) error {

}

/*
	register method, may implement the Register interface
*/

// FIXME: 批量插入?
func (u *User) InsertNewAccount(number, password string) {
	//init passwd = number
	u.number = number
	//TODO: 插入数据库
}

/*
	change password method, when forget password
*/

func (u *User) ChangePassword(password string) {
	//TODO: 修改密码：修改数据库
}

// QueryCourses query courses' name based on u.whoami
// the courses for student is which they study
// or for teacher is which they manage
//func (u *User) QueryCourses(db *sql.DB) ([]string, error) {
//	var tab string
//	switch u.whoami {
//	case 1:
//		tab = tabStudentCourse
//	case 2:
//		tab = tabTeacherCourse
//		//FIXME: 如何解决查询课程的无效身份
//	}
//
//	rows, errQ := db.Query(
//		"select c.courseName from %s st "+
//			"join %s coff on st.offeringID = coff.offeringID "+
//			"join %s c on coff.courseID = c.courseID "+
//			"where number = ?",
//		tab, tabCourseOffering, tabCourse, u.number,
//	)
//	if errQ != nil {
//		return nil, fmt.Errorf("%w: %v", ErrQuery, errQ)
//	}
//	var result []string
//	for rows.Next() {
//		var name string
//		if err := rows.Scan(&name); err != nil {
//			return nil, fmt.Errorf("%w: %v", ErrQuery, err)
//		}
//		result = append(result, name)
//	}
//	if err := rows.Err(); err != nil {
//		return nil, fmt.Errorf("%w: %v", ErrQuery, err)
//	}
//	return result, nil
//}
