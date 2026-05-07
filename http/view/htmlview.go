package view

import (
	"LabSystem/domain"
	"LabSystem/http/route"
	"net/url"
	"time"
)

// HXAction 表示htmx的组件，包含URL与HTTP方法
type HXAction struct {
	URL    string
	Method string
}

// 教师的项目视图
// 树状结构：课程列表 -> 班级列表 -> 项目列表

type TecProjectViewWithUrl struct {
	Courses []CourseTecItemWithUrl
}

type CourseTecItemWithUrl struct {
	CourseName string
	Classes    []ClassTecItemWithUrl
}

type ClassTecItemWithUrl struct {
	// 班级名称
	ClassName string
	// 新建项目
	CreateProject HXAction
	Projects      []ProjectTecItemWithUrl
}

type ProjectTecItemWithUrl struct {
	ProjectName string
	StartTime   time.Time
	CloseTime   time.Time
	// 项目开放状态
	IsActive bool
	// 上传项目要求文件
	Afford HXAction
	// 预览项目要求文件
	Check HXAction
	// 开放实验
	SwiftOpen HXAction
	// 关闭实验
	SwiftClose HXAction
	// 删除项目
	Delete HXAction
	// 检查学生完成情况
	WatchStuSubmission HXAction
	// 打包下载学生报告
	DownloadStuRp HXAction
}

func BuildTecProjectViewWithUrl(serviceViews []domain.TeacherProjectView) *TecProjectViewWithUrl {
	preview := url.Values{route.QueryKeyPreview: {"true"}}

	result := &TecProjectViewWithUrl{}
	for _, sv := range serviceViews {
		courseItem := CourseTecItemWithUrl{CourseName: sv.CourseName}
		for _, cls := range sv.Classes {
			classItem := ClassTecItemWithUrl{
				ClassName:     cls.ClassName,
				CreateProject: HXAction{route.ProjectsURL(), "POST"},
			}
			for _, pj := range cls.Projects {
				classItem.Projects = append(classItem.Projects, ProjectTecItemWithUrl{
					ProjectName:        pj.ProjectName,
					StartTime:          pj.StartTime,
					CloseTime:          pj.CloseTime,
					IsActive:           pj.IsActive,
					Afford:             HXAction{route.ProjectRequirementURL(pj.ProjectID), "PUT"},
					Check:              HXAction{route.WithQuery(route.ProjectRequirementURL(pj.ProjectID), preview), "GET"},
					SwiftOpen:          HXAction{route.ProjectURL(pj.ProjectID), "PATCH"},
					SwiftClose:         HXAction{route.ProjectURL(pj.ProjectID), "PATCH"},
					Delete:             HXAction{route.ProjectURL(pj.ProjectID), "DELETE"},
					WatchStuSubmission: HXAction{route.ProjectSubmissionsURL(pj.ProjectID), "GET"},
					DownloadStuRp:      HXAction{route.ProjectSubmissionsArchiveURL(pj.ProjectID), "GET"},
				})
			}
			courseItem.Classes = append(courseItem.Classes, classItem)
		}
		result.Courses = append(result.Courses, courseItem)
	}
	return result
}

// 学生的项目视图
// 树状结构：课程列表 -> 项目列表

type StuProjectViewWithUrl struct {
	Courses []CourseStuItemWithUrl
}

type CourseStuItemWithUrl struct {
	CourseName string
	Projects   []ProjectStuItemWithUrl
}

type ProjectStuItemWithUrl struct {
	// 项目名称
	ProjectName string
	// 开始时间
	StartTime time.Time
	// 截止时间
	CloseTime time.Time
	// 报告是否提交
	IsSubmit bool
	// 报告开放状态
	IsActive bool
	// 下载项目要求文件
	DownloadReq HXAction
	// 提交实验报告
	Submit HXAction
	// 检查（预览）提交的报告
	Check HXAction
}

func BuildStuProjectViewWithUrl(serviceViews []domain.StudentProjectView) *StuProjectViewWithUrl {
	preview := url.Values{route.QueryKeyPreview: {"true"}}

	result := &StuProjectViewWithUrl{}
	for _, sv := range serviceViews {
		courseItem := CourseStuItemWithUrl{CourseName: sv.CourseName}
		for _, pj := range sv.Projects {
			item := ProjectStuItemWithUrl{
				ProjectName: pj.ProjectName,
				StartTime:   pj.StartTime,
				CloseTime:   pj.CloseTime,
				IsSubmit:    pj.SubmitStatus,
				IsActive:    pj.IsActive,
				DownloadReq: HXAction{route.ProjectRequirementURL(pj.ProjectID), "GET"},
				Submit:      HXAction{route.ProjectSubmissionsURL(pj.ProjectID), "POST"},
			}
			if pj.SubmitStatus {
				item.Check = HXAction{route.WithQuery(route.SubmissionFileURL(pj.StuReportID), preview), "GET"}
			}
			courseItem.Projects = append(courseItem.Projects, item)
		}
		result.Courses = append(result.Courses, courseItem)
	}
	return result
}
