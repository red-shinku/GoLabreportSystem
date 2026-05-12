package service

import (
	"LabSystem/internal/domain"
	"io"
	"time"
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

	users := make([]domain.UserInfo, 0, len(sheet.Students))
	for _, s := range sheet.Students {
		// identity=1 学生；初始密码 = 学号
		users = append(users, *domain.NewUserInfo(1, s.Number, s.Number))
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

	for _, s := range sheet.Students {
		if err := c.repoCrs.FindOrInsertStudentCourse(s.Number, offeringID); err != nil {
			return err
		}
	}

	startTime := time.Now()
	for _, name := range sheet.ProjectNames {
		if _, err := c.repoPrj.FindOrInsertProject(offeringID, name, startTime, data.DefaultCloseTime); err != nil {
			return err
		}
	}
	return nil
}
