// 该文件下的代码负责具体业务的逻辑

package server

import (
	"LabSystem/domain"
	"time"
)

// AuthService 负责认证相关业务，包括登录认证等
type AuthService struct {
	repoAuth authOp
}

// auth 用户认证接口
type authOp interface {
	QueryPasswd(number string) (string, error)
	WhoAmI(number string) (uint8, error)
}

func (as *AuthService) LoginAuth(number string, passwd string) (domain.LoginUserInfo, error) {
	correctPasswd, err := as.repoAuth.QueryPasswd(number)
	if err != nil {
		return domain.LoginUserInfo{}, err
	}
	if correctPasswd != passwd {
		return domain.LoginUserInfo{}, domain.ErrAuth
	}
	id, errW := as.repoAuth.WhoAmI(number)
	if errW != nil {
		return domain.LoginUserInfo{}, errW
	}
	return domain.LoginUserInfo{id, number}, nil
}

// UserService 负责用户信息相关业务，包括注册、修改资料等
type UserService struct {
	repoRegister registerOp
}

type registerOp interface {
	InsertNewUserBatch(users *[]domain.UserInfo) error
}

// RegisterUsersBatch 批量注册用户
func (u *UserService) RegisterUsersBatch(users *[]domain.UserInfo) error {
	if err := u.repoRegister.InsertNewUserBatch(users); err != nil {
		return err
	}
	return nil
}

// TeacherProjectService 负责教师端项目相关的业务，如增删、开启关闭
type TeacherProjectService struct {
	repoTecProject teacherProjectOp
	repoPubProject publicProjectOp
}

// StudentProjectService 负责学生端项目相关的业务，如查询、下载要求
type StudentProjectService struct {
	repoStuProject studentProjectOp
	repoPubProject publicProjectOp
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
	QueryProjectFlag(projectID uint) (bool, error)
}

// ListProject 列出学生的项目列表，返回树状结构，包含课程、项目的层次信息
func (sp *StudentProjectService) ListProject(number string) ([]domain.StudentProjectView, error) {
	resInfo, err := sp.repoStuProject.QueryStuProject(number)
	if err != nil {
		return []domain.StudentProjectView{}, err
	}
	return sp.organizeStuPjView(&resInfo), nil
}

func (sp *StudentProjectService) organizeStuPjView(rowsInfo *[]domain.StudentProjectInfo) []domain.StudentProjectView {
	var result []domain.StudentProjectView
	indexMap := make(map[string]int)
	for _, rowInfo := range *rowsInfo {
		pj := domain.ProjectItem{
			ProjectName:  rowInfo.ProjectName,
			StartTime:    rowInfo.StartTime,
			CloseTime:    rowInfo.CloseTime,
			ProjectID:    rowInfo.ProjectID,
			SubmitStatus: rowInfo.StuReportID.Valid,
		}
		// 若解析过该课程名，将项目归到其下
		if idx, ok := indexMap[rowInfo.CourseName]; ok {
			result[idx].Projects = append(result[idx].Projects, pj)
		} else {
			newCourse := domain.StudentProjectView{CourseName: rowInfo.CourseName}
			newCourse.Projects = append(newCourse.Projects, pj)
			indexMap[rowInfo.CourseName] = len(result)
			result = append(result, newCourse)
		}
	}
	return result
	//FIXME: 处理可能的panic？
}

// GetProjectFilePath 返回要下载的文件路径
func (sp *StudentProjectService) GetProjectFilePath(projectID uint) (string, error) {
	path, err := sp.repoPubProject.QueryProjectFile(projectID)
	if err != nil {
		return "", err
	}
	return path, nil
}

// ListProject 列出教师管理的项目列表，返回树状结构，包含课程、班级、项目的层次信息
func (tp *TeacherProjectService) ListProject(number string) ([]domain.TeacherProjectView, error) {
	resInfo, err := tp.repoTecProject.QueryTeacherProject(number)
	if err != nil {
		return []domain.TeacherProjectView{}, err
	}
	return tp.organizeTecPjView(&resInfo), nil
}

func (tp *TeacherProjectService) organizeTecPjView(rowsInfo *[]domain.TeacherProjectInfo) []domain.TeacherProjectView {
	var result []domain.TeacherProjectView
	courseIndexMap := make(map[string]int)
	classIndexMap := make(map[string]map[string]int)

	for _, rowInfo := range *rowsInfo {
		pj := domain.ProjectItem{
			ProjectName: rowInfo.ProjectName,
			CloseTime:   rowInfo.CloseTime,
			ProjectID:   rowInfo.ProjectID,
		}
		if courseIdx, ok := courseIndexMap[rowInfo.CourseName]; ok {
			if classIdxMap, ok := classIndexMap[rowInfo.CourseName]; ok {
				if classIdx, ok := classIdxMap[rowInfo.ClassName]; ok {
					result[courseIdx].Classes[classIdx].Projects = append(result[courseIdx].Classes[classIdx].Projects, pj)
				} else {
					newClass := domain.ClassItem{ClassName: rowInfo.ClassName}
					newClass.Projects = append(newClass.Projects, pj)
					classIdxMap[rowInfo.ClassName] = len(result[courseIdx].Classes)
					result[courseIdx].Classes = append(result[courseIdx].Classes, newClass)
				}
			}
		} else {
			newClass := domain.ClassItem{ClassName: rowInfo.ClassName}
			newClass.Projects = append(newClass.Projects, pj)
			newCourse := domain.TeacherProjectView{CourseName: rowInfo.CourseName}
			newCourse.Classes = append(newCourse.Classes, newClass)
			courseIndexMap[rowInfo.CourseName] = len(result)
			classIndexMap[rowInfo.CourseName] = make(map[string]int)
			classIndexMap[rowInfo.CourseName][rowInfo.ClassName] = 0
			result = append(result, newCourse)
		}
	}
	return result
}

// ChangeProjectFilePath 教师重传项目文件后更改路径信息
func (tp *TeacherProjectService) ChangeProjectFilePath(projectID uint, path string) error {
	if err := tp.repoTecProject.UpdateProjectFile(projectID, path); err != nil {
		return err
	}
	return nil
}

// CreateProject 教师新建项目
func (tp *TeacherProjectService) CreateProject(pjInfo *domain.ProjectInfo) error {
	if err := tp.repoTecProject.AddProject(pjInfo); err != nil {
		return err
	}
	return nil
}

// ChangeProjectStatus 开启/关闭项目
func (tp *TeacherProjectService) ChangeProjectStatus(projectID uint) error {
	flag, err := tp.repoPubProject.QueryProjectFlag(projectID)
	if err != nil {
		return err
	}
	flag = !flag
	if err := tp.repoTecProject.UpdateProjectFlag(projectID, flag); err != nil {
		return err
	}
	return nil
}

// DeleteProject 删除项目
func (tp *TeacherProjectService) DeleteProject(projectID uint) error {
	if err := tp.repoTecProject.DelProject(projectID); err != nil {
		return err
	}
	return nil
}

// TeacherReportService 教师端实验报告相关业务
type TeacherReportService struct {
	repoTecReport teacherReportOp
}

// StudentReportService 学生端实验报告相关的业务、如上传、下载、检查等
type StudentReportService struct {
	repoStuReport studentReportOp
}

type teacherReportOp interface {
	QueryStuReportStatus(projectID uint) ([]domain.StuReportStatus, error)
	QueryStuReportFileAll(projectID uint) ([]string, error)
}

type studentReportOp interface {
	InsertStuReport(stuRp *domain.StuReportInfo) error
}

// CheckStuReportStatus 教师检查学生完成情况
func (tr *TeacherReportService) CheckStuReportStatus(projectID uint) ([]domain.StuReportStatus, error) {
	result, err := tr.repoTecReport.QueryStuReportStatus(projectID)
	if err != nil {
		return []domain.StuReportStatus{}, err
	}
	return result, nil
}

// GetStuReportFilePath 获取要下载的学生报告文件路径
func (tr *TeacherReportService) GetStuReportFilePath(projectID uint) ([]string, error) {
	result, err := tr.repoTecReport.QueryStuReportFileAll(projectID)
	if err != nil {
		return []string{}, err
	}
	return result, nil
}

// AddStuReport 学生添加报告
func (sr *StudentReportService) AddStuReport(stuRp *domain.StuReportInfo) error {
	stuRp.SubmitTime = time.Now()
	if err := sr.repoStuReport.InsertStuReport(stuRp); err != nil {
		return err
	}
	return nil
}

// CourseService 课程相关业务，如批量导入选课信息
type CourseService struct {
	repoCourse courseOp
}

type courseOp interface {
	InsertStuCourseOfferBatch(courses *[]domain.StudentCourseInfo) error
}

// RegisterStuCourseOfferBatch 批量注册学生选课信息
func (c *CourseService) RegisterStuCourseOfferBatch(stuCourseOffer *[]domain.StudentCourseInfo) error {
	if err := c.repoCourse.InsertStuCourseOfferBatch(stuCourseOffer); err != nil {
		return err
	}
	return nil
}

//TODO: 文件格式、存储服务（文件系统IO）？
