package database

import (
	"LabSystem/internal/domain"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
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

func NewUsersRepo(db *sql.DB) *UsersRepo {
	return &UsersRepo{db: db}
}

func NewProjectRepo(db *sql.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func NewReportRepo(db *sql.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

func NewManCourseOfferRepo(db *sql.DB) *ManCourseOfferRepo {
	return &ManCourseOfferRepo{db: db}
}

// UsersRepo 关于用户表的操作
type UsersRepo struct {
	db *sql.DB
}

// QueryPasswd 获取密码
func (u *UsersRepo) QueryPasswd(number string) (string, error) {
	var passwd string
	err := u.db.QueryRow(fmt.Sprintf("select passwd from %s where number = ?", tabUsers), number).Scan(&passwd)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("QueryPasswd(): %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("QueryPasswd(): %w, %v", domain.ErrQuery, err)
	}
	return passwd, nil
}

// WhoAmI 返回该号码对应的身份信息
func (u *UsersRepo) WhoAmI(number string) (uint8, error) {
	var identity uint8
	err := u.db.QueryRow(fmt.Sprintf("select identity from %s where number = ?", tabUsers), number).Scan(&identity)
	if err != nil {
		return 0, fmt.Errorf("WhoAmI(): %w, %v", domain.ErrQuery, err)
	}
	return identity, nil
}

// InsertNewUser 插入新用户
func (u *UsersRepo) InsertNewUser(user *domain.UserInfo) error {
	if u.db == nil || user == nil {
		return fmt.Errorf("InsertNewUser: invalid input parameters")
	}
	_, err := u.db.Exec(
		fmt.Sprintf("insert into %s (identity, number, passwd) values (?, ?, ?)", tabUsers),
		user.Identity,
		user.Number,
		user.Passwd,
	)
	if err != nil {
		return fmt.Errorf("InsertNewUser() insert failed: %w, %v", domain.ErrModify, err)
	}
	return nil
}

// InsertNewUserBatch 批量插入新用户
// FIXME: 使用事务
func (u *UsersRepo) InsertNewUserBatch(users *[]domain.UserInfo) error {
	if u.db == nil || users == nil {
		return fmt.Errorf("InsertNewUserBatch: invalid input parameters")
	}

	stmt, err := u.db.Prepare(fmt.Sprintf("insert into %s (identity, number, passwd) values (?, ?, ?)", tabUsers))
	if err != nil {
		return fmt.Errorf("InsertNewUserBatch() prepare failed: %w, %v", domain.ErrModify, err)
	}
	defer stmt.Close()

	for _, user := range *users {
		_, err := stmt.Exec(user.Identity, user.Number, user.Passwd)
		if err != nil {
			return fmt.Errorf("InsertNewUserBatch() execute failed: %w, %v", domain.ErrModify, err)
		}
	}
	return nil
}

// ChangePassword 修改User表的密码
func (u *UsersRepo) ChangePassword(userNumber string, newPassword string) error {
	_, err := u.db.Exec(fmt.Sprintf("update %s set passwd = ? where number = ?", tabUsers),
		newPassword, userNumber)
	if err != nil {
		return fmt.Errorf("ChangePassword(): %w, %v", domain.ErrModify, err)
	}
	return nil
}

// QueryStudentName 根据学生ID查询学生名
func (u *UsersRepo) QueryStudentName(studentID string) (string, error) {
	var studentName string
	err := u.db.QueryRow(fmt.Sprintf("select name from %s where number = ?", tabUsers), studentID).Scan(&studentName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("QueryStudentName(): %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("QueryStudentName(): %w, %v", domain.ErrQuery, err)
	}
	return studentName, nil
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

// QueryStuProjectByID 查询单个学生项目的当前信息（用于 AJAX 上传报告后刷新单张卡片）
func (p *ProjectRepo) QueryStuProjectByID(studentID string, projectID uint) (*domain.StudentProjectInfo, error) {
	if p.db == nil {
		return nil, fmt.Errorf("QueryStuProjectByID: invalid database connection")
	}
	query := fmt.Sprintf("select c.courseName, p.projectID, p.projectName, p.startTime, p.deadline, p.isActive, srp.stuReportID "+
		"from %s p "+
		"join %s coff on p.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"left join %s srp on p.projectID = srp.projectID and srp.studentID = ? "+
		"where p.projectID = ?",
		tabProject, tabCourseOffering, tabCourse, tabStuReport)
	info := &domain.StudentProjectInfo{}
	err := p.db.QueryRow(query, studentID, projectID).Scan(
		&info.CourseName, &info.ProjectID, &info.ProjectName,
		&info.StartTime, &info.CloseTime, &info.IsActive, &info.StuReportID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("QueryStuProjectByID(): %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("QueryStuProjectByID(): %w, %v", domain.ErrQuery, err)
	}
	return info, nil
}

// QueryTeacherProject 查询一个教师所管理的项目，包含项目所属班级、课程等信息
func (p *ProjectRepo) QueryTeacherProject(teacherID string) ([]domain.TeacherProjectInfo, error) {
	query := fmt.Sprintf("select c.courseName, coff.className, coff.offeringID, p.projectID, p.projectName, p.deadline, p.isActive "+
		"from %s tec "+
		"join %s coff on tec.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"where tec.teacherID = ?",
		tabTeacherCourse, tabCourseOffering, tabCourse, tabProject)
	scanFunc := func(rows *sql.Rows, row *domain.TeacherProjectInfo) error {
		return rows.Scan(&row.CourseName, &row.ClassName, &row.OfferingID, &row.ProjectID, &row.ProjectName,
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
	_, err := p.db.Exec(fmt.Sprintf("update %s set isActive = ? where projectID = ?", tabProject),
		flag, projectID)
	if err != nil {
		return fmt.Errorf("UpdateProjectFlag() failed: %w, %v", domain.ErrModify, err)
	}
	return nil
}

func (p *ProjectRepo) QueryProjectFlag(projectID uint) (bool, error) {
	query := fmt.Sprintf("select isActive from %s where projectID = ?", tabProject)
	var flag bool
	if err := p.db.QueryRow(query, projectID).Scan(&flag); err != nil {
		return false, fmt.Errorf("QueryProjectFlag(): %v", err)
	}
	return flag, nil
}

// QueryOfferingInfo 通过OfferingID查询课程名和班级名，用于新建项目时获取文件路径信息
func (p *ProjectRepo) QueryOfferingInfo(offeringID uint) (courseName, className string, err error) {
	query := fmt.Sprintf("select c.courseName, coff.className "+
		"from %s coff "+
		"join %s c on coff.courseID = c.courseID "+
		"where coff.offeringID = ?",
		tabCourseOffering, tabCourse)
	if err := p.db.QueryRow(query, offeringID).Scan(&courseName, &className); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("QueryOfferingInfo(): %w", domain.ErrNotFound)
		}
		return "", "", fmt.Errorf("QueryOfferingInfo(): %w, %v", domain.ErrQuery, err)
	}
	return courseName, className, nil
}

// AddProject 教师新建项目，返回新项目ID
func (p *ProjectRepo) AddProject(project *domain.ProjectInfo) (uint, error) {
	if p.db == nil || project == nil {
		return 0, fmt.Errorf("AddProject: invalid input parameters")
	}

	res, err := p.db.Exec(
		fmt.Sprintf("insert into %s (offeringID, projectName, projectFilePath, startTime, deadline) values (?, ?, ?, ?, ?)", tabProject),
		project.OfferingID,
		project.ProjectName,
		project.ProjectFilePath,
		project.StartTime,
		project.CloseTime,
	)
	if err != nil {
		return 0, fmt.Errorf("AddProject() insert failed: %w, %v", domain.ErrModify, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("AddProject() LastInsertId failed: %w, %v", domain.ErrModify, err)
	}
	return uint(id), nil
}

// QueryProjectByID 根据项目ID查询项目当前信息（用于切换状态后刷新片段）
func (p *ProjectRepo) QueryProjectByID(projectID uint) (*domain.TeacherProjectInfo, error) {
	if p.db == nil {
		return nil, fmt.Errorf("QueryProjectByID: invalid database connection")
	}
	query := fmt.Sprintf("select projectName, deadline, isActive from %s where projectID = ?", tabProject)
	info := &domain.TeacherProjectInfo{ProjectID: projectID}
	err := p.db.QueryRow(query, projectID).Scan(&info.ProjectName, &info.CloseTime, &info.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("QueryProjectByID(): %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("QueryProjectByID(): %w, %v", domain.ErrQuery, err)
	}
	return info, nil
}

// DelProject 教师删除项目
func (p *ProjectRepo) DelProject(projectID uint) error {
	if p.db == nil {
		return fmt.Errorf("DelProject: invalid database connection")
	}

	res, err := p.db.Exec(fmt.Sprintf("delete from %s where projectID = ?", tabProject), projectID)
	if err != nil {
		return fmt.Errorf("DelProject(): %w, %v", domain.ErrModify, err)
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("DelProject() failed: %w", domain.ErrNotFound)
	}
	return nil
}

// UpdateProjectFile 教师更新项目文件路径
func (p *ProjectRepo) UpdateProjectFile(projectID uint, projectFilePath string) error {
	if p.db == nil {
		return fmt.Errorf("UpdateProjectFile: invalid database connection")
	}

	_, err := p.db.Exec(fmt.Sprintf("update %s set projectFilePath = ? where projectID = ?", tabProject),
		projectFilePath, projectID)
	if err != nil {
		return fmt.Errorf("UpdateProjectFile() %w, %v", domain.ErrModify, err)
	}
	return nil
}

// QueryProjectFile 获取项目文件路径
func (p *ProjectRepo) QueryProjectFile(projectID uint) (string, error) {
	if p.db == nil {
		return "", fmt.Errorf("QueryProjectFile: invalid database connection")
	}

	var filePath string
	err := p.db.QueryRow(fmt.Sprintf("select projectFilePath from %s where projectID = ?", tabProject),
		projectID).Scan(&filePath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("QueryProjectFile() failed: %w", domain.ErrNotFound)
		}
		return "", fmt.Errorf("QueryProjectFile(): %w, %v", domain.ErrQuery, err)
	}
	return filePath, nil
}

// QueryProjectInfo 根据项目ID查询项目所属的课程名、班级名、项目名
func (p *ProjectRepo) QueryProjectInfo(projectID uint) (courseName, className, projectName string, err error) {
	if p.db == nil {
		return "", "", "", fmt.Errorf("QueryProjectInfo: invalid database connection")
	}

	query := fmt.Sprintf(
		"select c.courseName, coff.className, pj.projectName "+
			"from %s pj "+
			"join %s coff on pj.offeringID = coff.offeringID "+
			"join %s c on coff.courseID = c.courseID "+
			"where pj.projectID = ?",
		tabProject, tabCourseOffering, tabCourse)

	err = p.db.QueryRow(query, projectID).Scan(&courseName, &className, &projectName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", fmt.Errorf("QueryProjectInfo(): %w", domain.ErrNotFound)
		}
		return "", "", "", fmt.Errorf("QueryProjectInfo(): %w, %v", domain.ErrQuery, err)
	}
	return courseName, className, projectName, nil
}

// ReportRepo 关于报告表的操作
type ReportRepo struct {
	db *sql.DB
}

// QueryStuReportStatus 教师查询某项目下学生的完成情况
func (r *ReportRepo) QueryStuReportStatus(projectID uint) ([]domain.StuReportStatus, error) {
	query := fmt.Sprintf("select stuc.studentID, srp.stuReportID "+
		"from %s stuc "+
		"join %s coff on stuc.offeringID = coff.offeringID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"left join %s srp on p.projectID = srp.projectID and stuc.studentID = srp.studentID "+
		"where p.projectID = ?",
		tabStudentCourse, tabCourseOffering, tabProject, tabStuReport)
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

// InsertStuReport 学生提交报告，报告表中插入新行，返回新报告ID
func (r *ReportRepo) InsertStuReport(stuRp *domain.StuReportInfo) (uint, error) {
	if r.db == nil {
		return 0, fmt.Errorf("InsertStuReport: invalid database connection")
	}
	res, err := r.db.Exec(
		fmt.Sprintf("insert into %s (studentID, projectID, reportFilePath, submitTime) values (?, ?, ?, ?)", tabStuReport),
		stuRp.StudentID,
		stuRp.ProjectID,
		stuRp.ReportFilePath,
		stuRp.SubmitTime)
	if err != nil {
		return 0, fmt.Errorf("InsertStuReport(): %w, %v", domain.ErrModify, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("InsertStuReport() LastInsertId: %w, %v", domain.ErrModify, err)
	}
	return uint(id), nil
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

// InsertStuCourseOfferBatch 批量添加学生选课，插入学生选课表
// FIXME: 使用事务
func (mco *ManCourseOfferRepo) InsertStuCourseOfferBatch(courses *[]domain.StudentCourseInfo) error {
	if mco.db == nil || courses == nil {
		return fmt.Errorf("InsertStuCourseOfferBatch: invalid input parameters")
	}
	for _, course := range *courses {
		_, err := mco.db.Exec(
			fmt.Sprintf("insert into %s (studentID, offeringID) values (?, ?)", tabStudentCourse),
			course.StudentID,
			course.OfferingID,
		)
		if err != nil {
			return fmt.Errorf("InsertStuCourseOfferBatch(): %w, %v", domain.ErrModify, err)
		}
	}
	return nil
}

// queryTemplate 多行查询模版
func queryTemplate[T any](db *sql.DB, query string, args []any,
	scanFunc func(*sql.Rows, *T) error, result *[]T) error {
	rows, errQ := db.Query(query, args...)
	if errQ != nil {
		return fmt.Errorf("%w: %v", domain.ErrQuery, errQ)
	}
	defer rows.Close()

	for rows.Next() {
		var row T
		if err := scanFunc(rows, &row); err != nil {
			return fmt.Errorf("%w: %v", domain.ErrQuery, err)
		}
		*result = append(*result, row)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrQuery, err)
	}
	return nil
}
