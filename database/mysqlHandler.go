package database

import (
	"LabSystem/domain"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

// error flag
var (
	ErrQuery    = errors.New("query failed")
	ErrModify   = errors.New("modify database failed")
	ErrNotFound = errors.New("entry not found")
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

// UsersRepo 关于用户表的操作
type UsersRepo struct {
	db *sql.DB
}

// QueryPasswd 获取密码
func (u *UsersRepo) QueryPasswd(number string) (string, error) {
	var passwd string
	err := u.db.QueryRow("select passwd from %s where number = ?", tabUsers, number).Scan(&passwd)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("QueryPasswd(): %w", ErrNotFound)
		}
		return "", fmt.Errorf("QueryPasswd(): %w, %v", ErrQuery, err)
	}
	return passwd, nil
}

// WhoAmI 返回该号码对应的身份信息
func (u *UsersRepo) WhoAmI(number string) uint8 {
	var identity uint8
	err := u.db.QueryRow("select identity from %s where number = ?", tabUsers, number).Scan(&identity)
	if err != nil {
		return 0
	}
	return identity
}

// InsertNewUser 插入新用户
func (u *UsersRepo) InsertNewUser(user *domain.UserInfo) error {
	if u.db == nil || user == nil {
		return fmt.Errorf("InsertNewUser: invalid input parameters")
	}
	res, err := u.db.Exec(
		"insert into %s (identity, number, passwd) values (?, ?, ?)",
		tabUsers,
		user.Identity,
		user.Number,
		user.Passwd,
	)
	if err != nil {
		return fmt.Errorf("InsertNewUser() insert failed: %w, %v", ErrModify, err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("InsertNewUser(): %w", ErrNotFound)
	}
	return nil
}

// InsertNewUserBatch 批量插入新用户
func (u *UsersRepo) InsertNewUserBatch(users *[]domain.UserInfo) error {
	if u.db == nil || users == nil {
		return fmt.Errorf("InsertNewUserBatch: invalid input parameters")
	}

	stmt, err := u.db.Prepare(fmt.Sprintf("insert into %s (identity, number, passwd) values (?, ?, ?)", tabUsers))
	if err != nil {
		return fmt.Errorf("InsertNewUserBatch() prepare failed: %w, %v", ErrModify, err)
	}
	defer stmt.Close()

	for _, user := range *users {
		res, err := stmt.Exec(user.Identity, user.Number, user.Passwd)
		if err != nil {
			return fmt.Errorf("InsertNewUserBatch() execute failed: %w, %v", ErrModify, err)
		}
		count, _ := res.RowsAffected()
		if count == 0 {
			return fmt.Errorf("InsertNewUserBatch(): %w", ErrNotFound)
		}
	}
	return nil
}

// ChangePassword 修改User表的密码
func (u *UsersRepo) ChangePassword(userNumber string, newPassword string) error {
	res, err := u.db.Exec("update %s set passwd = ? where number = ?",
		tabUsers, newPassword, userNumber)
	if err != nil {
		return fmt.Errorf("ChangePassword(): %w, %v", ErrModify, err)
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("ChangePassword(): %w", ErrNotFound)
	}
	return nil
}

// ProjectRepo 关于项目表的操作
type ProjectRepo struct {
	db *sql.DB
}

// QueryStuProject 查询一个学生的所有项目及其信息
func (p *ProjectRepo) QueryStuProject(studentID string) ([]domain.StudentProjectInfo, error) {
	query := fmt.Sprintf("select c.courseName, p.projectID, p.projectName, p.startTime, p.deadline, p.isActive, srp.stuReportID "+
		"from %s stuc "+
		"join %s coff on stuc.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"join %s srp on p.projectID = srp.projectID and stuc.studentID = srp.studentID "+
		"where stuc.studentID = ?",
		tabStudentCourse, tabCourseOffering, tabCourse, tabProject, tabStuReport)
	scanFunc := func(rows *sql.Rows, row *domain.StudentProjectInfo) error {
		return rows.Scan(&row.CourseName, &row.ProjectID, &row.ProjectName,
			&row.StartTime, &row.CloseTime, &row.IsActive, &row.StuReportID)
	}
	args := []any{studentID}

	var result []domain.StudentProjectInfo
	if err := queryTemplate(p.db, query, args, scanFunc, &result); err != nil {
		return nil, fmt.Errorf("QueryStuProject failed: %w", err)
	}
	return result, nil
}

// QueryTeacherProject 查询一个教师所管理的项目，包含项目所属班级、课程等信息
func (p *ProjectRepo) QueryTeacherProject(teacherID string) ([]domain.TeacherProjectInfo, error) {
	query := fmt.Sprintf("select c.courseName, coff.className, p.projectID, p.projectName, p.deadline, p.isActive "+
		"from %s tec "+
		"join %s coff on tec.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"where teacherID = ?",
		tabTeacherCourse, tabCourseOffering, tabCourse, tabProject)
	scanFunc := func(rows *sql.Rows, row *domain.TeacherProjectInfo) error {
		return rows.Scan(&row.CourseName, &row.ClassName, &row.ProjectID, &row.ProjectName,
			&row.CloseTime, &row.IsActive)
	}
	args := []any{teacherID}

	var result []domain.TeacherProjectInfo
	if err := queryTemplate(p.db, query, args, scanFunc, &result); err != nil {
		return nil, fmt.Errorf("QueryTeacherProject() failed: %w", err)
	}
	return result, nil
}

// UpdateProjectFlag 教师开启/关闭项目
func (p *ProjectRepo) UpdateProjectFlag(projectID uint, flag bool) error {
	res, err := p.db.Exec("update %s set isActive = ? where projectID = ?",
		tabProject, flag, projectID)
	if err != nil {
		return fmt.Errorf("UpdateProjectFlag() failed: %w, %v", ErrModify, err)
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return ErrModify
	}
	return nil
}

// AddProject 教师新建项目
func (p *ProjectRepo) AddProject(project *domain.ProjectInfo) error {
	if p.db == nil || project == nil {
		return fmt.Errorf("AddProject: invalid input parameters")
	}

	res, err := p.db.Exec(
		"insert into %s (offeringID, projectName, projectFilePath, startTime, deadline) values (?, ?, ?, ?, ?)",
		tabProject,
		project.OfferingID,
		project.ProjectName,
		project.ProjectFilePath,
		project.StartTime,
		project.CloseTime,
	)
	if err != nil {
		return fmt.Errorf("AddProject() insert failed: %w, %v", ErrModify, err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("AddProject(): %w", ErrNotFound)
	}
	return nil
}

// DelProject 教师删除项目
func (p *ProjectRepo) DelProject(projectID uint) error {
	if p.db == nil {
		return fmt.Errorf("DelProject: invalid database connection")
	}

	res, err := p.db.Exec("delete from %s where projectID = ?", tabProject, projectID)
	if err != nil {
		return fmt.Errorf("DelProject(): %w, %v", ErrModify, err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("DelProject() failed: %w", ErrNotFound)
	}
	return nil
}

// UpdateProjectFile 教师更新项目文件路径
func (p *ProjectRepo) UpdateProjectFile(projectID uint, projectFilePath string) error {
	if p.db == nil {
		return fmt.Errorf("UpdateProjectFile: invalid database connection")
	}

	res, err := p.db.Exec("update %s set projectFilePath = ? where projectID = ?",
		tabProject, projectFilePath, projectID)
	if err != nil {
		return fmt.Errorf("UpdateProjectFile() %w, %v", ErrModify, err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("UpdateProjectFile() failed: %w", ErrNotFound)
	}
	return nil
}

// QueryProjectFile 获取项目文件路径
func (p *ProjectRepo) QueryProjectFile(projectID uint) (string, error) {
	if p.db == nil {
		return "", fmt.Errorf("QueryProjectFile: invalid database connection")
	}

	var filePath string
	err := p.db.QueryRow("select projectFilePath from %s where projectID = ?",
		tabProject, projectID).Scan(&filePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("QueryProjectFile() failed: %w", ErrNotFound)
		}
		return "", fmt.Errorf("QueryProjectFile(): %w, %v", ErrQuery, err)
	}
	return filePath, nil
}

// ReportRepo 关于报告表的操作
type ReportRepo struct {
	db *sql.DB
}

// QueryStuReportStatus 教师查询某项目下学生的完成情况
func (r *ReportRepo) QueryStuReportStatus(projectID uint) ([]domain.StuReportStatus, error) {
	query := fmt.Sprintf("select srp.studentID, srp.stuReportID "+
		"from %s p "+
		"join %s srp on p.projectID = srp.projectID "+
		"where projectID = ?",
		tabProject, tabStuReport)
	scanFunc := func(rows *sql.Rows, row *domain.StuReportStatus) error {
		return rows.Scan(&row.StudentID, &row.StuReportID)
	}
	args := []any{projectID}

	var result []domain.StuReportStatus
	if err := queryTemplate(r.db, query, args, scanFunc, &result); err != nil {
		return nil, fmt.Errorf("QueryStuProjectStatus() failed: %w", err)
	}
	return result, nil
}

// InsertStuReport 学生提交报告，报告表中插入新行
func (r *ReportRepo) InsertStuReport(stuRp *domain.StuReportInfo) error {
	if r.db == nil {
		return fmt.Errorf("InsertStuReport: invalid database connection")
	}
	res, err := r.db.Exec(
		"insert into %s (studentID, projectID, reportFilePath, submitTime) values (?, ?, ?, ?)",
		tabStuReport,
		stuRp.StudentID,
		stuRp.ProjectID,
		stuRp.ReportFilePath,
		stuRp.SubmitTime)
	if err != nil {
		return fmt.Errorf("InsertStuReport(): %w, %v", ErrModify, err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("InsertStuReport() invailed")
	}
	return nil
}

// QueryStuReportFileAll 获取某项目下，所有学生的报告文件路径
func (r *ReportRepo) QueryStuReportFileAll(projectID uint) ([]string, error) {
	if r.db == nil {
		return nil, fmt.Errorf("QueryStuReportFileAll: invalid database connection")
	}
	query := fmt.Sprintf("select reportFilePath from %s where projectID = ?", tabStuReport)
	scanFunc := func(rows *sql.Rows, row *string) error {
		return rows.Scan(row)
	}
	args := []any{projectID}

	var result []string
	if err := queryTemplate(r.db, query, args, scanFunc, &result); err != nil {
		return nil, fmt.Errorf("QueryStuReportFileAll(): %w", err)
	}
	return result, nil
}

// ManCourseOfferRepo 学生/教师选课表相关操作
type ManCourseOfferRepo struct {
	db *sql.DB
}

// InsertStuCourseOffer 批量添加学生选课，插入学生选课表
func (mco *ManCourseOfferRepo) InsertStuCourseOffer(courses *[]domain.StudentCourseInfo) error {
	if mco.db == nil || courses == nil {
		return fmt.Errorf("InsertStuCourseOffer: invalid input parameters")
	}
	for _, course := range *courses {
		res, err := mco.db.Exec(
			"insert into %s (studentID, offeringID) values (?, ?)",
			tabStudentCourse,
			course.StudentID,
			course.OfferingID,
		)
		if err != nil {
			return fmt.Errorf("InsertStuCourseOffer(): %w, %v", ErrModify, err)
		}

		count, _ := res.RowsAffected()
		if count == 0 {
			return fmt.Errorf("InsertStuCourseOffer(): %w", ErrNotFound)
		}
	}
	return nil
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
