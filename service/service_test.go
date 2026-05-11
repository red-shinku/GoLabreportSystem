package service

import (
	"LabSystem/internal/domain"
	"archive/zip"
	"bytes"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func fixedTime() time.Time {
	return time.Date(2026, 5, 10, 12, 30, 0, 0, time.UTC)
}

func chdirTemp(t *testing.T) string {
	t.Helper()

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() failed: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("os.Chdir(%q) failed: %v", tempDir, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore workdir failed: %v", err)
		}
	})

	return tempDir
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) failed: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) failed: %v", path, err)
	}
}

func readZipEntries(t *testing.T, payload []byte) map[string]string {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		t.Fatalf("zip.NewReader() failed: %v", err)
	}

	entries := make(map[string]string, len(reader.File))
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("file.Open(%q) failed: %v", file.Name, err)
		}

		body, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("io.ReadAll(%q) failed: %v", file.Name, err)
		}
		entries[file.Name] = string(body)
	}
	return entries
}

type stubAuthRepo struct {
	queryPasswdFn func(number string) (string, error)
	whoAmIFn      func(number string) (uint8, error)
}

func (s stubAuthRepo) QueryPasswd(number string) (string, error) {
	return s.queryPasswdFn(number)
}

func (s stubAuthRepo) WhoAmI(number string) (uint8, error) {
	return s.whoAmIFn(number)
}

type stubRegisterRepo struct {
	insertNewUserBatchFn func(users *[]domain.UserInfo) error
}

func (s stubRegisterRepo) InsertNewUserBatch(users *[]domain.UserInfo) error {
	return s.insertNewUserBatchFn(users)
}

type stubStudentProjectRepo struct {
	queryStuProjectFn func(studentID string) ([]domain.StudentProjectInfo, error)
}

func (s stubStudentProjectRepo) QueryStuProject(studentID string) ([]domain.StudentProjectInfo, error) {
	return s.queryStuProjectFn(studentID)
}

type stubPublicProjectRepo struct {
	queryProjectFileFn  func(projectID uint) (string, error)
	queryProjectFlagFn  func(projectID uint) (bool, error)
	queryProjectInfoFn  func(projectID uint) (courseName, className, projectName string, err error)
	queryOfferingInfoFn func(offeringID uint) (courseName, className string, err error)
}

func (s stubPublicProjectRepo) QueryProjectFile(projectID uint) (string, error) {
	return s.queryProjectFileFn(projectID)
}

func (s stubPublicProjectRepo) QueryProjectFlag(projectID uint) (bool, error) {
	return s.queryProjectFlagFn(projectID)
}

func (s stubPublicProjectRepo) QueryProjectInfo(projectID uint) (courseName, className, projectName string, err error) {
	return s.queryProjectInfoFn(projectID)
}

func (s stubPublicProjectRepo) QueryOfferingInfo(offeringID uint) (courseName, className string, err error) {
	return s.queryOfferingInfoFn(offeringID)
}

type stubTeacherProjectRepo struct {
	queryTeacherProjectFn func(teacherID string) ([]domain.TeacherProjectInfo, error)
	queryProjectByIDFn    func(projectID uint) (*domain.TeacherProjectInfo, error)
	updateProjectFlagFn   func(projectID uint, flag bool) error
	addProjectFn          func(project *domain.ProjectInfo) (uint, error)
	delProjectFn          func(projectID uint) error
	updateProjectFileFn   func(projectID uint, projectFilePath string) error
}

func (s stubTeacherProjectRepo) QueryTeacherProject(teacherID string) ([]domain.TeacherProjectInfo, error) {
	return s.queryTeacherProjectFn(teacherID)
}

func (s stubTeacherProjectRepo) QueryProjectByID(projectID uint) (*domain.TeacherProjectInfo, error) {
	return s.queryProjectByIDFn(projectID)
}

func (s stubTeacherProjectRepo) UpdateProjectFlag(projectID uint, flag bool) error {
	return s.updateProjectFlagFn(projectID, flag)
}

func (s stubTeacherProjectRepo) AddProject(project *domain.ProjectInfo) (uint, error) {
	return s.addProjectFn(project)
}

func (s stubTeacherProjectRepo) DelProject(projectID uint) error {
	return s.delProjectFn(projectID)
}

func (s stubTeacherProjectRepo) UpdateProjectFile(projectID uint, projectFilePath string) error {
	return s.updateProjectFileFn(projectID, projectFilePath)
}

type stubTeacherReportRepo struct {
	queryStuReportStatusFn  func(projectID uint) ([]domain.StuReportStatus, error)
	queryStuReportFileAllFn func(projectID uint) ([]string, error)
}

func (s stubTeacherReportRepo) QueryStuReportStatus(projectID uint) ([]domain.StuReportStatus, error) {
	return s.queryStuReportStatusFn(projectID)
}

func (s stubTeacherReportRepo) QueryStuReportFileAll(projectID uint) ([]string, error) {
	return s.queryStuReportFileAllFn(projectID)
}

type stubStudentReportRepo struct {
	insertStuReportFn func(stuRp *domain.StuReportInfo) (uint, error)
}

func (s stubStudentReportRepo) InsertStuReport(stuRp *domain.StuReportInfo) (uint, error) {
	return s.insertStuReportFn(stuRp)
}

type stubProjectInfoRepo struct {
	queryProjectInfoFn    func(projectID uint) (courseName, className, projectName string, err error)
	queryStuProjectByIDFn func(studentID string, projectID uint) (*domain.StudentProjectInfo, error)
}

func (s stubProjectInfoRepo) QueryProjectInfo(projectID uint) (courseName, className, projectName string, err error) {
	return s.queryProjectInfoFn(projectID)
}

func (s stubProjectInfoRepo) QueryStuProjectByID(studentID string, projectID uint) (*domain.StudentProjectInfo, error) {
	return s.queryStuProjectByIDFn(studentID, projectID)
}

type stubStudentInfoRepo struct {
	queryStudentNameFn func(studentID string) (string, error)
}

func (s stubStudentInfoRepo) QueryStudentName(studentID string) (string, error) {
	return s.queryStudentNameFn(studentID)
}

type stubCourseRepo struct {
	insertStuCourseOfferBatchFn func(courses *[]domain.StudentCourseInfo) error
}

func (s stubCourseRepo) InsertStuCourseOfferBatch(courses *[]domain.StudentCourseInfo) error {
	return s.insertStuCourseOfferBatchFn(courses)
}

func TestAuthService_LoginAuth(t *testing.T) {
	t.Run("returns query error", func(t *testing.T) {
		wantErr := errors.New("query failed")
		svc := &AuthService{
			repoAuth: stubAuthRepo{
				queryPasswdFn: func(number string) (string, error) {
					if number != "20260001" {
						t.Fatalf("unexpected number: %q", number)
					}
					return "", wantErr
				},
				whoAmIFn: func(string) (uint8, error) {
					t.Fatal("WhoAmI should not be called")
					return 0, nil
				},
			},
		}

		_, err := svc.LoginAuth("20260001", "secret")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})

	t.Run("returns ErrAuth on password mismatch", func(t *testing.T) {
		whoCalled := false
		svc := &AuthService{
			repoAuth: stubAuthRepo{
				queryPasswdFn: func(string) (string, error) {
					return "correct-pass", nil
				},
				whoAmIFn: func(string) (uint8, error) {
					whoCalled = true
					return 0, nil
				},
			},
		}

		_, err := svc.LoginAuth("20260001", "wrong-pass")
		if !errors.Is(err, domain.ErrAuth) {
			t.Fatalf("expected ErrAuth, got %v", err)
		}
		if whoCalled {
			t.Fatal("WhoAmI should not be called on password mismatch")
		}
	})

	t.Run("returns identity lookup error", func(t *testing.T) {
		wantErr := errors.New("identity lookup failed")
		svc := &AuthService{
			repoAuth: stubAuthRepo{
				queryPasswdFn: func(string) (string, error) {
					return "secret", nil
				},
				whoAmIFn: func(number string) (uint8, error) {
					if number != "20260001" {
						t.Fatalf("unexpected number: %q", number)
					}
					return 0, wantErr
				},
			},
		}

		_, err := svc.LoginAuth("20260001", "secret")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})

	t.Run("returns login user info", func(t *testing.T) {
		svc := &AuthService{
			repoAuth: stubAuthRepo{
				queryPasswdFn: func(string) (string, error) {
					return "secret", nil
				},
				whoAmIFn: func(number string) (uint8, error) {
					if number != "20260001" {
						t.Fatalf("unexpected number: %q", number)
					}
					return 2, nil
				},
			},
		}

		got, err := svc.LoginAuth("20260001", "secret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Identity != 2 || got.Number != "20260001" {
			t.Fatalf("unexpected login user info: %+v", got)
		}
	})
}

func TestUserService_RegisterUsersBatch(t *testing.T) {
	users := []domain.UserInfo{{Identity: 1, Number: "20260001", Passwd: "pass"}}

	t.Run("passes users to repo", func(t *testing.T) {
		var gotUsers *[]domain.UserInfo
		svc := &UserService{
			repoRegister: stubRegisterRepo{
				insertNewUserBatchFn: func(input *[]domain.UserInfo) error {
					gotUsers = input
					return nil
				},
			},
		}

		if err := svc.RegisterUsersBatch(&users); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotUsers != &users {
			t.Fatal("service should pass the original slice pointer to repo")
		}
	})

	t.Run("returns repo error", func(t *testing.T) {
		wantErr := errors.New("insert failed")
		svc := &UserService{
			repoRegister: stubRegisterRepo{
				insertNewUserBatchFn: func(*[]domain.UserInfo) error {
					return wantErr
				},
			},
		}

		if err := svc.RegisterUsersBatch(&users); !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})
}

func TestStudentProjectService_ListProject(t *testing.T) {
	t.Run("returns repo error", func(t *testing.T) {
		wantErr := errors.New("query failed")
		svc := &StudentProjectService{
			repoStuProject: stubStudentProjectRepo{
				queryStuProjectFn: func(string) ([]domain.StudentProjectInfo, error) {
					return nil, wantErr
				},
			},
		}

		got, err := svc.ListProject("20260001")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty result, got %+v", got)
		}
	})

	t.Run("groups rows by course", func(t *testing.T) {
		start := fixedTime()
		closeTime := start.Add(48 * time.Hour)
		rows := []domain.StudentProjectInfo{
			{
				CourseName:  "Go Basics",
				ProjectID:   1001,
				ProjectName: "Lab1",
				StartTime:   start,
				CloseTime:   closeTime,
				IsActive:    true,
				StuReportID: sql.NullInt32{Int32: 9001, Valid: true},
			},
			{
				CourseName:  "Go Basics",
				ProjectID:   1002,
				ProjectName: "Lab2",
				StartTime:   start,
				CloseTime:   closeTime,
				IsActive:    false,
				StuReportID: sql.NullInt32{},
			},
			{
				CourseName:  "Rust Intro",
				ProjectID:   2001,
				ProjectName: "Ownership",
				StartTime:   start,
				CloseTime:   closeTime,
				IsActive:    true,
				StuReportID: sql.NullInt32{Int32: 9010, Valid: true},
			},
		}

		svc := &StudentProjectService{
			repoStuProject: stubStudentProjectRepo{
				queryStuProjectFn: func(studentID string) ([]domain.StudentProjectInfo, error) {
					if studentID != "20260001" {
						t.Fatalf("unexpected student id: %q", studentID)
					}
					return rows, nil
				},
			},
		}

		got, err := svc.ListProject("20260001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 courses, got %d", len(got))
		}
		if got[0].CourseName != "Go Basics" || len(got[0].Projects) != 2 {
			t.Fatalf("unexpected first course view: %+v", got[0])
		}
		if !got[0].Projects[0].SubmitStatus || got[0].Projects[0].StuReportID != 9001 {
			t.Fatalf("unexpected first project: %+v", got[0].Projects[0])
		}
		if got[0].Projects[1].SubmitStatus || got[0].Projects[1].StuReportID != 0 {
			t.Fatalf("unexpected second project: %+v", got[0].Projects[1])
		}
		if got[1].CourseName != "Rust Intro" || len(got[1].Projects) != 1 {
			t.Fatalf("unexpected second course view: %+v", got[1])
		}
	})
}

func TestStudentProjectService_DownloadProjectFile(t *testing.T) {
	t.Run("returns repo error", func(t *testing.T) {
		wantErr := errors.New("query failed")
		svc := &StudentProjectService{
			repoPubProject: stubPublicProjectRepo{
				queryProjectFileFn: func(uint) (string, error) {
					return "", wantErr
				},
			},
			fs: &FileService{},
		}

		var buf bytes.Buffer
		if err := svc.DownloadProjectFile(&buf, 1001); !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})

	t.Run("loads file content", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "project.txt")
		mustWriteFile(t, filePath, "project-content")

		svc := &StudentProjectService{
			repoPubProject: stubPublicProjectRepo{
				queryProjectFileFn: func(projectID uint) (string, error) {
					if projectID != 1001 {
						t.Fatalf("unexpected project id: %d", projectID)
					}
					return filePath, nil
				},
			},
			fs: &FileService{},
		}

		var buf bytes.Buffer
		if err := svc.DownloadProjectFile(&buf, 1001); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if buf.String() != "project-content" {
			t.Fatalf("unexpected content: %q", buf.String())
		}
	})
}

func TestTeacherProjectService_ListProject(t *testing.T) {
	t.Run("returns repo error", func(t *testing.T) {
		wantErr := errors.New("query failed")
		svc := &TeacherProjectService{
			repoTecProject: stubTeacherProjectRepo{
				queryTeacherProjectFn: func(string) ([]domain.TeacherProjectInfo, error) {
					return nil, wantErr
				},
			},
		}

		got, err := svc.ListProject("T001")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty result, got %+v", got)
		}
	})

	t.Run("groups rows by course and class", func(t *testing.T) {
		closeTime := fixedTime().Add(72 * time.Hour)
		rows := []domain.TeacherProjectInfo{
			{CourseName: "Go Basics", ClassName: "Class A", OfferingID: 11, ProjectID: 1001, ProjectName: "Lab1", CloseTime: closeTime, IsActive: true},
			{CourseName: "Go Basics", ClassName: "Class A", OfferingID: 11, ProjectID: 1002, ProjectName: "Lab2", CloseTime: closeTime, IsActive: false},
			{CourseName: "Go Basics", ClassName: "Class B", OfferingID: 12, ProjectID: 1003, ProjectName: "Lab3", CloseTime: closeTime, IsActive: true},
			{CourseName: "Rust Intro", ClassName: "Class C", OfferingID: 21, ProjectID: 2001, ProjectName: "Ownership", CloseTime: closeTime, IsActive: true},
		}

		svc := &TeacherProjectService{
			repoTecProject: stubTeacherProjectRepo{
				queryTeacherProjectFn: func(teacherID string) ([]domain.TeacherProjectInfo, error) {
					if teacherID != "T001" {
						t.Fatalf("unexpected teacher id: %q", teacherID)
					}
					return rows, nil
				},
			},
		}

		got, err := svc.ListProject("T001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 courses, got %d", len(got))
		}
		if got[0].CourseName != "Go Basics" || len(got[0].Classes) != 2 {
			t.Fatalf("unexpected first course view: %+v", got[0])
		}
		if got[0].Classes[0].ClassName != "Class A" || len(got[0].Classes[0].Projects) != 2 {
			t.Fatalf("unexpected first class view: %+v", got[0].Classes[0])
		}
		if got[0].Classes[1].ClassName != "Class B" || len(got[0].Classes[1].Projects) != 1 {
			t.Fatalf("unexpected second class view: %+v", got[0].Classes[1])
		}
		if got[1].CourseName != "Rust Intro" || len(got[1].Classes) != 1 {
			t.Fatalf("unexpected second course view: %+v", got[1])
		}
	})
}

func TestTeacherProjectService_UploadProjectFile(t *testing.T) {
	root := chdirTemp(t)
	var updatedProjectID uint
	var updatedPath string
	wantPath, pathErr := domain.NewProjectFileMeta("Go Basics", "Class A", "Lab1", "requirements.pdf").FilePath()

	svc := &TeacherProjectService{
		repoPubProject: stubPublicProjectRepo{
			queryProjectInfoFn: func(projectID uint) (string, string, string, error) {
				if projectID != 1001 {
					t.Fatalf("unexpected project id: %d", projectID)
				}
				return "Go Basics", "Class A", "Lab1", nil
			},
		},
		repoTecProject: stubTeacherProjectRepo{
			updateProjectFileFn: func(projectID uint, projectFilePath string) error {
				updatedProjectID = projectID
				updatedPath = projectFilePath
				return nil
			},
		},
		fs: &FileService{},
	}

	form := domain.NewProjectFileData(1001, "requirements.pdf")
	err := svc.UploadProjectFile(bytes.NewBufferString("file-content"), form)
	if pathErr != nil {
		if !errors.Is(err, domain.ErrNotSafe) {
			t.Fatalf("expected ErrNotSafe, got %v", err)
		}
		if updatedProjectID != 0 || updatedPath != "" {
			t.Fatalf("repo should not be updated on path error: id=%d path=%q", updatedProjectID, updatedPath)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updatedProjectID != 1001 {
		t.Fatalf("unexpected project id: %d", updatedProjectID)
	}
	if updatedPath != wantPath {
		t.Fatalf("unexpected update path: got %q want %q", updatedPath, wantPath)
	}

	body, err := os.ReadFile(filepath.Join(root, wantPath))
	if err != nil {
		t.Fatalf("os.ReadFile() failed: %v", err)
	}
	if string(body) != "file-content" {
		t.Fatalf("unexpected file content: %q", string(body))
	}
}

func TestTeacherProjectService_CreateProject(t *testing.T) {
	closeTime := fixedTime().Add(24 * time.Hour)
	var inserted *domain.ProjectInfo
	before := time.Now()

	svc := &TeacherProjectService{
		repoTecProject: stubTeacherProjectRepo{
			addProjectFn: func(project *domain.ProjectInfo) (uint, error) {
				inserted = project
				return 123, nil
			},
		},
	}

	got, err := svc.CreateProject(domain.NewProjectData(11, "Lab1", closeTime))
	after := time.Now()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inserted == nil {
		t.Fatal("repo AddProject should receive project info")
	}
	if inserted.OfferingID != 11 || inserted.ProjectName != "Lab1" || inserted.ProjectFilePath != "" {
		t.Fatalf("unexpected inserted project: %+v", inserted)
	}
	if !inserted.CloseTime.Equal(closeTime) {
		t.Fatalf("unexpected close time: %v", inserted.CloseTime)
	}
	if inserted.StartTime.Before(before) || inserted.StartTime.After(after) {
		t.Fatalf("unexpected start time: %v", inserted.StartTime)
	}
	if got.ProjectID != 123 || got.ProjectName != "Lab1" || got.IsActive {
		t.Fatalf("unexpected created item: %+v", got)
	}
	if !got.StartTime.Equal(inserted.StartTime) || !got.CloseTime.Equal(closeTime) {
		t.Fatalf("unexpected returned time fields: %+v", got)
	}
}

func TestTeacherProjectService_ChangeProjectStatus(t *testing.T) {
	closeTime := fixedTime().Add(24 * time.Hour)
	var updatedProjectID uint
	var updatedFlag bool

	svc := &TeacherProjectService{
		repoPubProject: stubPublicProjectRepo{
			queryProjectFlagFn: func(projectID uint) (bool, error) {
				if projectID != 1001 {
					t.Fatalf("unexpected project id: %d", projectID)
				}
				return false, nil
			},
		},
		repoTecProject: stubTeacherProjectRepo{
			updateProjectFlagFn: func(projectID uint, flag bool) error {
				updatedProjectID = projectID
				updatedFlag = flag
				return nil
			},
			queryProjectByIDFn: func(projectID uint) (*domain.TeacherProjectInfo, error) {
				return &domain.TeacherProjectInfo{
					ProjectID:   projectID,
					ProjectName: "Lab1",
					CloseTime:   closeTime,
					IsActive:    true,
				}, nil
			},
		},
	}

	got, err := svc.ChangeProjectStatus(1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedProjectID != 1001 || !updatedFlag {
		t.Fatalf("unexpected update call: id=%d flag=%v", updatedProjectID, updatedFlag)
	}
	if got.ProjectID != 1001 || got.ProjectName != "Lab1" || !got.IsActive {
		t.Fatalf("unexpected project item: %+v", got)
	}
}

func TestTeacherProjectService_DeleteProject(t *testing.T) {
	root := chdirTemp(t)
	relativeDir, pathErr := domain.NewProjectFileMeta("Go Basics", "Class A", "Lab1", "Lab1").DirectoryPath()
	var targetDir string
	if pathErr == nil {
		targetDir = filepath.Join(root, relativeDir)
		mustWriteFile(t, filepath.Join(targetDir, "requirements.pdf"), "file-content")
	}

	var deletedProjectID uint
	delCalled := false
	svc := &TeacherProjectService{
		repoPubProject: stubPublicProjectRepo{
			queryProjectInfoFn: func(projectID uint) (string, string, string, error) {
				if projectID != 1001 {
					t.Fatalf("unexpected project id: %d", projectID)
				}
				return "Go Basics", "Class A", "Lab1", nil
			},
		},
		repoTecProject: stubTeacherProjectRepo{
			delProjectFn: func(projectID uint) error {
				delCalled = true
				deletedProjectID = projectID
				return nil
			},
		},
		fs: &FileService{},
	}

	err := svc.DeleteProject(1001)
	if pathErr != nil {
		if !errors.Is(err, domain.ErrNotSafe) {
			t.Fatalf("expected ErrNotSafe, got %v", err)
		}
		if delCalled {
			t.Fatal("DelProject should not be called when directory path is unsafe")
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !delCalled || deletedProjectID != 1001 {
		t.Fatalf("unexpected deleted project id: %d", deletedProjectID)
	}
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Fatalf("expected directory to be removed, got err=%v", err)
	}
}

func TestTeacherReportService_CheckStuReportStatus(t *testing.T) {
	t.Run("returns repo error", func(t *testing.T) {
		wantErr := errors.New("query failed")
		svc := &TeacherReportService{
			repoTecReport: stubTeacherReportRepo{
				queryStuReportStatusFn: func(uint) ([]domain.StuReportStatus, error) {
					return nil, wantErr
				},
			},
		}

		got, err := svc.CheckStuReportStatus(1001)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty result, got %+v", got)
		}
	})

	t.Run("returns report status rows", func(t *testing.T) {
		rows := []domain.StuReportStatus{
			{StudentID: "20260001", StuReportID: sql.NullInt32{Int32: 9001, Valid: true}},
			{StudentID: "20260002", StuReportID: sql.NullInt32{}},
		}
		svc := &TeacherReportService{
			repoTecReport: stubTeacherReportRepo{
				queryStuReportStatusFn: func(projectID uint) ([]domain.StuReportStatus, error) {
					if projectID != 1001 {
						t.Fatalf("unexpected project id: %d", projectID)
					}
					return rows, nil
				},
			},
		}

		got, err := svc.CheckStuReportStatus(1001)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 || got[0].StudentID != "20260001" || !got[0].StuReportID.Valid {
			t.Fatalf("unexpected statuses: %+v", got)
		}
	})
}

func TestTeacherReportService_DownloadStuReportBatch(t *testing.T) {
	t.Run("returns repo error", func(t *testing.T) {
		wantErr := errors.New("query failed")
		svc := &TeacherReportService{
			repoTecReport: stubTeacherReportRepo{
				queryStuReportFileAllFn: func(uint) ([]string, error) {
					return nil, wantErr
				},
			},
			fs: &FileService{},
		}

		var buf bytes.Buffer
		if err := svc.DownloadStuReportBatch(&buf, 1001); !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})

	t.Run("writes zip archive", func(t *testing.T) {
		dir := t.TempDir()
		fileA := filepath.Join(dir, "a.txt")
		fileB := filepath.Join(dir, "b.txt")
		mustWriteFile(t, fileA, "alpha")
		mustWriteFile(t, fileB, "beta")

		svc := &TeacherReportService{
			repoTecReport: stubTeacherReportRepo{
				queryStuReportFileAllFn: func(projectID uint) ([]string, error) {
					if projectID != 1001 {
						t.Fatalf("unexpected project id: %d", projectID)
					}
					return []string{fileA, fileB}, nil
				},
			},
			fs: &FileService{},
		}

		var buf bytes.Buffer
		if err := svc.DownloadStuReportBatch(&buf, 1001); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		entries := readZipEntries(t, buf.Bytes())
		if len(entries) != 2 {
			t.Fatalf("expected 2 zip entries, got %d", len(entries))
		}
		if entries["a.txt"] != "alpha" || entries["b.txt"] != "beta" {
			t.Fatalf("unexpected zip entries: %+v", entries)
		}
	})
}

func TestStudentReportService_GenStuReportData(t *testing.T) {
	t.Run("returns project info error", func(t *testing.T) {
		wantErr := errors.New("project query failed")
		svc := &StudentReportService{
			repoProject: stubProjectInfoRepo{
				queryProjectInfoFn: func(uint) (string, string, string, error) {
					return "", "", "", wantErr
				},
				queryStuProjectByIDFn: func(string, uint) (*domain.StudentProjectInfo, error) {
					t.Fatal("QueryStuProjectByID should not be called")
					return nil, nil
				},
			},
			repoUser: stubStudentInfoRepo{
				queryStudentNameFn: func(string) (string, error) {
					t.Fatal("QueryStudentName should not be called")
					return "", nil
				},
			},
		}

		_, _, err := svc.genStuReportData(domain.NewStuReportData("20260001", 1001, "pdf"))
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})

	t.Run("builds report metadata and info", func(t *testing.T) {
		before := time.Now()
		wantPath, pathErr := domain.NewStuReportMeta("Go Basics", "Class A", "Alice", "Lab1", "20260001", "pdf").FilePath()
		svc := &StudentReportService{
			repoProject: stubProjectInfoRepo{
				queryProjectInfoFn: func(projectID uint) (string, string, string, error) {
					if projectID != 1001 {
						t.Fatalf("unexpected project id: %d", projectID)
					}
					return "Go Basics", "Class A", "Lab1", nil
				},
				queryStuProjectByIDFn: func(string, uint) (*domain.StudentProjectInfo, error) {
					t.Fatal("QueryStuProjectByID should not be called by genStuReportData")
					return nil, nil
				},
			},
			repoUser: stubStudentInfoRepo{
				queryStudentNameFn: func(studentID string) (string, error) {
					if studentID != "20260001" {
						t.Fatalf("unexpected student id: %q", studentID)
					}
					return "Alice", nil
				},
			},
		}

		meta, info, err := svc.genStuReportData(domain.NewStuReportData("20260001", 1001, "pdf"))
		if pathErr != nil {
			if !errors.Is(err, domain.ErrNotSafe) {
				t.Fatalf("expected ErrNotSafe, got %v", err)
			}
			if meta != nil || info != nil {
				t.Fatalf("expected nil meta/info on path error, got meta=%+v info=%+v", meta, info)
			}
			return
		}
		after := time.Now()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if meta.CourseName != "Go Basics" || meta.ClassName != "Class A" || meta.StudentName != "Alice" || meta.ProjectName != "Lab1" {
			t.Fatalf("unexpected meta: %+v", meta)
		}
		if info.StudentID != "20260001" || info.ProjectID != 1001 {
			t.Fatalf("unexpected report info: %+v", info)
		}
		if info.SubmitTime.Before(before) || info.SubmitTime.After(after) {
			t.Fatalf("unexpected submit time: %v", info.SubmitTime)
		}
		if info.ReportFilePath != wantPath {
			t.Fatalf("unexpected report file path: got %q want %q", info.ReportFilePath, wantPath)
		}
	})
}

func TestStudentReportService_UploadStuReport(t *testing.T) {
	insertCalled := false
	queryProjectByIDCalled := false
	_, pathErr := domain.NewStuReportMeta("Go Basics", "Class A", "Alice", "Lab1", "20260001", "pdf").FilePath()

	svc := &StudentReportService{
		repoStuReport: stubStudentReportRepo{
			insertStuReportFn: func(*domain.StuReportInfo) (uint, error) {
				insertCalled = true
				return 0, nil
			},
		},
		repoProject: stubProjectInfoRepo{
			queryProjectInfoFn: func(projectID uint) (string, string, string, error) {
				if projectID != 1001 {
					t.Fatalf("unexpected project id: %d", projectID)
				}
				return "Go Basics", "Class A", "Lab1", nil
			},
			queryStuProjectByIDFn: func(string, uint) (*domain.StudentProjectInfo, error) {
				queryProjectByIDCalled = true
				return nil, nil
			},
		},
		repoUser: stubStudentInfoRepo{
			queryStudentNameFn: func(studentID string) (string, error) {
				if studentID != "20260001" {
					t.Fatalf("unexpected student id: %q", studentID)
				}
				return "Alice", nil
			},
		},
		fs: &FileService{},
	}

	_, err := svc.UploadStuReport(bytes.NewBufferString("report-content"), domain.NewStuReportData("20260001", 1001, "pdf"))
	if pathErr != nil {
		if !errors.Is(err, domain.ErrNotSafe) {
			t.Fatalf("expected ErrNotSafe, got %v", err)
		}
	} else if !errors.Is(err, domain.ErrNotAllow) {
		t.Fatalf("expected ErrNotAllow, got %v", err)
	}
	if insertCalled {
		t.Fatal("InsertStuReport should not be called when format check fails")
	}
	if queryProjectByIDCalled {
		t.Fatal("QueryStuProjectByID should not be called when format check fails")
	}
}

func TestCourseService_RegisterStuCourseOfferBatch(t *testing.T) {
	courses := []domain.StudentCourseInfo{{StudentID: "20260001", OfferingID: 11}}

	t.Run("passes courses to repo", func(t *testing.T) {
		var gotCourses *[]domain.StudentCourseInfo
		svc := &CourseService{
			repoCourse: stubCourseRepo{
				insertStuCourseOfferBatchFn: func(input *[]domain.StudentCourseInfo) error {
					gotCourses = input
					return nil
				},
			},
		}

		if err := svc.RegisterStuCourseOfferBatch(&courses); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotCourses != &courses {
			t.Fatal("service should pass the original slice pointer to repo")
		}
	})

	t.Run("returns repo error", func(t *testing.T) {
		wantErr := errors.New("insert failed")
		svc := &CourseService{
			repoCourse: stubCourseRepo{
				insertStuCourseOfferBatchFn: func(*[]domain.StudentCourseInfo) error {
					return wantErr
				},
			},
		}

		if err := svc.RegisterStuCourseOfferBatch(&courses); !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})
}
