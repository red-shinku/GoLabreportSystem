// 该文件下的代码负责具体业务的逻辑

package server

import "LabSystem/domain"

// AuthService 负责认证相关业务，包括登录等
type AuthService struct {
	repoAuth *authOp
}

// auth 用户认证接口
type authOp interface {
	QueryPasswd(number string) (string, error)
	WhoAmI(number string) uint8
}

// UserService 负责用户信息相关业务，包括注册、修改资料等
type UserService struct {
	repoRegister *registerOp
}

type registerOp interface {
	InsertNewUserBatch(users *[]domain.UserInfo) error
}

// TeacherProjectService 负责教师端项目相关的业务，如增删、开启关闭
type TeacherProjectService struct {
	repoTecProject *teacherProjectOp
	repoPubProject *publicProjectOp
}

// StudentProjectService 负责学生端项目相关的业务，如查询、下载要求
type StudentProjectService struct {
	repoStuProject *studentProjectOp
	repoPubProject *publicProjectOp
}

// teacherProjectOp 教师操作项目的接口
type teacherProjectOp interface {
	QueryTeacherProject(teacherID string) ([]domain.TeacherProjectInfo, error)
	UpdateProjectFlag(projectID uint, flag bool) error
	AddProject(project *domain.ProjectInfo) error
	DelProject(projectID uint) error
	UpdateProjectFile(projectID uint, projectFilePath string) error
}

// studentProjectOp 学生操作项目的接口
type studentProjectOp interface {
	QueryStuProject(studentID string) ([]domain.StudentProjectInfo, error)
}

// publicProjectOp 操作项目的公共接口
type publicProjectOp interface {
	QueryProjectFile(projectID uint) (string, error)
}

// TeacherReportService 教师端实验报告相关业务
type TeacherReportService struct {
	repoTecReport *teacherReportOp
}

// StudentReportService 学生端实验报告相关的业务、如上传、下载、检查等
type StudentReportService struct {
	repoStuReport *studentReportOp
}

type teacherReportOp interface {
	QueryStuReportStatus(projectID uint) ([]domain.StuReportStatus, error)
	QueryStuReportFileAll(projectID uint) ([]string, error)
}

type studentReportOp interface {
	InsertStuReport(stuRp *domain.StuReportInfo) error
}

// CourseService 课程相关业务，如批量导入选课信息
type CourseService struct {
	repoCourse *courseOp
}

type courseOp interface {
	InsertStuCourseOffer(courses *[]domain.StudentCourseInfo) error
}
