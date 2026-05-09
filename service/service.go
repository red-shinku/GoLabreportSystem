// 该文件下的代码负责具体业务的逻辑

package service

import (
	"LabSystem/internal/domain"
	"io"
	"time"
)

//=========================================================
//	关于认证相关的业务，目前有：
//	登录认证
//=========================================================

// AuthService 负责认证相关业务，包括登录认证等
type AuthService struct {
	repoAuth authOp
}

// auth 用户认证接口
type authOp interface {
	QueryPasswd(number string) (string, error)
	WhoAmI(number string) (uint8, error)
}

func (as *AuthService) LoginAuth(number string, passwd string) (*domain.LoginUserInfo, error) {
	correctPasswd, err := as.repoAuth.QueryPasswd(number)
	if err != nil {
		return nil, err
	}
	if correctPasswd != passwd {
		return nil, domain.ErrAuth
	}
	id, errW := as.repoAuth.WhoAmI(number)
	if errW != nil {
		return nil, errW
	}
	return domain.NewLoginUserInfo(id, number), nil
}

//=========================================================
//	关于用户相关的业务，目前有：
//	批量注册
//=========================================================

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

//=========================================================
//	关于项目相关的业务，分为教师端与学生端。目前有：
//	教师端：
//	1) 生成教师项目视图
//	2) 新建项目（需要同时上传项目文件）
//	3) 重传项目文件
//	4) 修改项目开启状态
//	5) 删除项目（同时删除文件
//	学生端：
//	1) 生成学生项目视图
//	2) 获取项目文件
//=========================================================

// TeacherProjectService 负责教师端项目相关的业务，如增删、开启关闭
type TeacherProjectService struct {
	repoTecProject teacherProjectOp
	repoPubProject publicProjectOp
	fs             *FileService
}

// StudentProjectService 负责学生端项目相关的业务，如查询、下载要求
type StudentProjectService struct {
	repoStuProject studentProjectOp
	repoPubProject publicProjectOp
	fs             *FileService
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
	QueryProjectInfo(projectID uint) (courseName, className, projectName string, err error)
	QueryOfferingInfo(offeringID uint) (courseName, className string, err error)
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
		var reportID uint
		if rowInfo.StuReportID.Valid {
			reportID = uint(rowInfo.StuReportID.Int32)
		}
		pj := domain.NewProjectStuItem(
			rowInfo.ProjectName,
			rowInfo.StartTime,
			rowInfo.CloseTime,
			rowInfo.IsActive,
			rowInfo.ProjectID,
			rowInfo.StuReportID.Valid,
			reportID)
		// 若解析过该课程名，将项目归到其下
		if idx, ok := indexMap[rowInfo.CourseName]; ok {
			result[idx].Projects = append(result[idx].Projects, *pj)
		} else {
			newCourse := domain.NewStudentProjectView(rowInfo.CourseName)
			newCourse.Projects = append(newCourse.Projects, *pj)
			indexMap[rowInfo.CourseName] = len(result)
			result = append(result, *newCourse)
		}
	}
	return result
}

// DownloadProjectFile 客户下载项目要求文件
func (sp *StudentProjectService) DownloadProjectFile(w io.Writer, projectID uint) error {
	path, err := sp.repoPubProject.QueryProjectFile(projectID)
	if err != nil {
		return err
	}

	if err := sp.fs.LoadFile(w, path); err != nil {
		return err
	}
	return nil
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
		pj := domain.NewProjectTecItem(
			rowInfo.ProjectName,
			time.Time{},
			rowInfo.CloseTime,
			rowInfo.IsActive,
			rowInfo.ProjectID)
		if courseIdx, ok := courseIndexMap[rowInfo.CourseName]; ok {
			if classIdxMap, ok := classIndexMap[rowInfo.CourseName]; ok {
				if classIdx, ok := classIdxMap[rowInfo.ClassName]; ok {
					result[courseIdx].Classes[classIdx].Projects = append(result[courseIdx].Classes[classIdx].Projects, *pj)
				} else {
					newClass := domain.NewClassItem(rowInfo.ClassName, rowInfo.OfferingID)
					newClass.Projects = append(newClass.Projects, *pj)
					classIdxMap[rowInfo.ClassName] = len(result[courseIdx].Classes)
					result[courseIdx].Classes = append(result[courseIdx].Classes, *newClass)
				}
			}
		} else {
			newClass := domain.NewClassItem(rowInfo.ClassName, rowInfo.OfferingID)
			newClass.Projects = append(newClass.Projects, *pj)
			newCourse := domain.NewTeacherProjectView(rowInfo.CourseName)
			newCourse.Classes = append(newCourse.Classes, *newClass)
			courseIndexMap[rowInfo.CourseName] = len(result)
			classIndexMap[rowInfo.CourseName] = make(map[string]int)
			classIndexMap[rowInfo.CourseName][rowInfo.ClassName] = 0
			result = append(result, *newCourse)
		}
	}
	return result
}

// UploadProjectFile 教师上传/重传项目文件后更改路径信息
func (tp *TeacherProjectService) UploadProjectFile(r io.Reader, form *domain.ProjectFileData) error {
	courseName, className, projectName, err := tp.repoPubProject.QueryProjectInfo(form.ProjectID)
	if err != nil {
		return err
	}
	meta := domain.NewProjectFileMeta(courseName, className, projectName, form.FileName)
	if err := tp.fs.SaveFile(r, meta); err != nil {
		return err
	}
	path, _ := meta.FilePath()
	if err := tp.repoTecProject.UpdateProjectFile(form.ProjectID, path); err != nil {
		return err
	}
	return nil
}

// CreateProject 教师新建项目
func (tp *TeacherProjectService) CreateProject(form *domain.ProjectData) error {
	info := domain.NewProjectInfo(
		form.OfferingID,
		form.ProjectName,
		"",
		time.Now(),
		form.CloseTime)
	if err := tp.repoTecProject.AddProject(info); err != nil {
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

// DeleteProject 删除项目，并清除相关文件
func (tp *TeacherProjectService) DeleteProject(projectID uint) error {
	//FIXME: 文件删除与数据库事务绑定
	courseName, className, projectName, err := tp.repoPubProject.QueryProjectInfo(projectID)
	if err != nil {
		return err
	}
	if err := tp.fs.DeleteDirectory(domain.NewProjectFileMeta(courseName, className, projectName, projectName)); err != nil {
		return err
	}
	if err := tp.repoTecProject.DelProject(projectID); err != nil {
		return err
	}
	return nil
}

//=========================================================
//	关于实验报告相关的业务，分为学生端与教师端，目前有：
//	教师端：
// 	1) 获取学生完成情况
//	2) 批量下载学生报告
//	学生端：
//	1) 上传报告文件
//=========================================================

// TeacherReportService 教师端实验报告相关业务
type TeacherReportService struct {
	repoTecReport teacherReportOp
	fs            *FileService
}

// StudentReportService 学生端实验报告相关的业务、如上传、下载、检查等
type StudentReportService struct {
	repoStuReport studentReportOp
	repoProject   projectInfoOp
	repoUser      studentInfoOp
	fs            *FileService
}

type teacherReportOp interface {
	QueryStuReportStatus(projectID uint) ([]domain.StuReportStatus, error)
	QueryStuReportFileAll(projectID uint) ([]string, error)
}

type studentReportOp interface {
	InsertStuReport(stuRp *domain.StuReportInfo) error
}

type projectInfoOp interface {
	QueryProjectInfo(projectID uint) (courseName, className, projectName string, err error)
}

type studentInfoOp interface {
	QueryStudentName(studentID string) (string, error)
}

// CheckStuReportStatus 教师检查学生完成情况
func (tr *TeacherReportService) CheckStuReportStatus(projectID uint) ([]domain.StuReportStatus, error) {
	result, err := tr.repoTecReport.QueryStuReportStatus(projectID)
	if err != nil {
		return []domain.StuReportStatus{}, err
	}
	return result, nil
}

// DownloadStuReportBatch 客户打包下载的学生报告文件
func (tr *TeacherReportService) DownloadStuReportBatch(w io.Writer, projectID uint) error {
	result, err := tr.repoTecReport.QueryStuReportFileAll(projectID)
	if err != nil {
		return err
	}
	tr.fs.LoadFileBatch(w, domain.NewTargetPaths(result))
	return nil
}

// UploadStuReport 学生上传报告
func (sr *StudentReportService) UploadStuReport(r io.Reader, form *domain.StuReportData) error {
	//TODO : 启用两个协程完成 ?
	meta, info, err := sr.genStuReportData(form)
	if err != nil {
		return err
	}
	if err := meta.Check(); err != nil {
		return err
	}

	if err := sr.fs.SaveFile(r, meta); err != nil {
		return err
	}
	//FIXME: 开启事务
	if err := sr.repoStuReport.InsertStuReport(info); err != nil {
		return err
	}
	return nil
}

// genStuReportData 生成学生报告业务元数据
func (sr *StudentReportService) genStuReportData(form *domain.StuReportData) (*domain.StuReportMeta, *domain.StuReportInfo, error) {
	courseName, className, projectName, err := sr.repoProject.QueryProjectInfo(form.ProjectID)
	if err != nil {
		return nil, nil, err
	}
	studentName, err := sr.repoUser.QueryStudentName(form.StudentID)
	if err != nil {
		return nil, nil, err
	}
	meta := domain.NewStuReportMeta(courseName, className, studentName, projectName, form.StudentID, form.Format)
	filePath, err := meta.FilePath()
	if err != nil {
		return nil, nil, err
	}
	info := domain.NewStuReportInfo(form.StudentID, form.ProjectID, filePath, time.Now())
	return meta, info, nil
}

//=========================================================
//	关于课程相关的业务，目前有：
//	批量添加学生选课
//=========================================================

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
