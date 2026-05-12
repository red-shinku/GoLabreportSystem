package service

import (
	"LabSystem/internal/domain"
	"fmt"
	"io"
	"time"
	"unicode/utf8"
)

const (
	userNumberMaxLen = 16
	userNameMaxLen   = 16
)

// userRegisterOp 学生注册（幂等）
type userRegisterOp interface {
	InsertNewUserBatchIgnore(users *[]domain.UserInfo) error
}

// courseRegisterOp 课程登记：课程、开课、教师授课、学生选课
type courseRegisterOp interface {
	FindOrInsertCourse(name string) (uint, error)
	FindOrInsertCourseOffering(courseID uint, className, term string) (uint, error)
	FindOrInsertTeacherCourse(teacherID string, offeringID uint) error
	FindOrInsertStudentCourse(studentID string, offeringID uint) error
}

// projectRegisterOp 项目登记（幂等）
type projectRegisterOp interface {
	FindOrInsertProject(offeringID uint, projectName string, startTime, deadline time.Time) (uint, error)
}

// CourseImportService 教师"课程信息表"导入业务
// 组合 sheetParser + 三个领域的 Repo 接口，编排整个导入流程
type CourseImportService struct {
	parser  sheetParser
	repoUsr userRegisterOp
	repoCrs courseRegisterOp
	repoPrj projectRegisterOp
}

func NewCourseImportService(
	parser sheetParser,
	repoUsr userRegisterOp,
	repoCrs courseRegisterOp,
	repoPrj projectRegisterOp,
) *CourseImportService {
	return &CourseImportService{
		parser:  parser,
		repoUsr: repoUsr,
		repoCrs: repoCrs,
		repoPrj: repoPrj,
	}
}

// Import 解析Excel并按顺序完成：学生注册 → 课程/开课登记 → 教师绑定 → 学生选课 → 项目登记
// 任一阶段错误即返回；已写入的部分保留（未启用事务，已在数据库层FIXME中说明）
func (c *CourseImportService) Import(r io.Reader, data *domain.ImportCourseData) error {
	sheet, err := c.parser.Parse(r)
	if err != nil {
		return err
	}

	if data.ClassName == "" {
		data.ClassName = "-"
	}

	users, err := buildImportUsers(sheet.Students)
	if err != nil {
		return err
	}
	if err := c.repoUsr.InsertNewUserBatchIgnore(&users); err != nil {
		return err
	}

	courseID, err := c.repoCrs.FindOrInsertCourse(data.CourseName)
	if err != nil {
		return err
	}
	offeringID, err := c.repoCrs.FindOrInsertCourseOffering(courseID, data.ClassName, data.Term)
	if err != nil {
		return err
	}
	if err := c.repoCrs.FindOrInsertTeacherCourse(data.TeacherID, offeringID); err != nil {
		return err
	}

	if err := c.registerStudentCourses(sheet.Students, offeringID); err != nil {
		return err
	}

	startTime := time.Now()
	for _, name := range sheet.ProjectNames {
		if _, err := c.repoPrj.FindOrInsertProject(offeringID, name, startTime, data.DefaultCloseTime); err != nil {
			return err
		}
	}
	return nil
}

func (c *CourseImportService) registerStudentCourses(students []domain.StudentRow, offeringID uint) error {
	for _, student := range students {
		if err := c.repoCrs.FindOrInsertStudentCourse(student.Number, offeringID); err != nil {
			return fmt.Errorf(
				"registerStudentCourses(): bind student %q to offering %d: %w",
				student.Number,
				offeringID,
				err,
			)
		}
	}
	return nil
}

func buildImportUsers(students []domain.StudentRow) ([]domain.UserInfo, error) {
	users := make([]domain.UserInfo, 0, len(students))
	for _, s := range students {
		if utf8.RuneCountInString(s.Number) > userNumberMaxLen {
			return nil, fmt.Errorf(
				"buildImportUsers(): %w: student number %q exceeds %d characters",
				domain.ErrSheetFormat,
				s.Number,
				userNumberMaxLen,
			)
		}
		if utf8.RuneCountInString(s.Name) > userNameMaxLen {
			return nil, fmt.Errorf(
				"buildImportUsers(): %w: student name %q exceeds %d characters",
				domain.ErrSheetFormat,
				s.Name,
				userNameMaxLen,
			)
		}

		user := domain.NewUserInfo(1, s.Number, s.Number)
		user.Name = s.Name
		users = append(users, *user)
	}
	return users, nil
}
