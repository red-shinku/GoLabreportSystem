package service

import (
	"LabSystem/internal/domain"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

type stubSheetParser struct {
	parseFn func(r io.Reader) (*domain.SheetData, error)
}

func (s stubSheetParser) Parse(r io.Reader) (*domain.SheetData, error) { return s.parseFn(r) }

type stubUserRegisterRepo struct {
	insertNewUserBatchIgnoreFn func(users *[]domain.UserInfo) error
}

func (s stubUserRegisterRepo) InsertNewUserBatchIgnore(users *[]domain.UserInfo) error {
	return s.insertNewUserBatchIgnoreFn(users)
}

type stubCourseRegisterRepo struct {
	findOrInsertCourseFn         func(name string) (uint, error)
	findOrInsertCourseOfferingFn func(courseID uint, className, term string) (uint, error)
	findOrInsertTeacherCourseFn  func(teacherID string, offeringID uint) error
	findOrInsertStudentCourseFn  func(studentID string, offeringID uint) error
}

func (s stubCourseRegisterRepo) FindOrInsertCourse(name string) (uint, error) {
	return s.findOrInsertCourseFn(name)
}
func (s stubCourseRegisterRepo) FindOrInsertCourseOffering(courseID uint, className, term string) (uint, error) {
	return s.findOrInsertCourseOfferingFn(courseID, className, term)
}
func (s stubCourseRegisterRepo) FindOrInsertTeacherCourse(teacherID string, offeringID uint) error {
	return s.findOrInsertTeacherCourseFn(teacherID, offeringID)
}
func (s stubCourseRegisterRepo) FindOrInsertStudentCourse(studentID string, offeringID uint) error {
	return s.findOrInsertStudentCourseFn(studentID, offeringID)
}

type stubProjectRegisterRepo struct {
	findOrInsertProjectFn func(offeringID uint, projectName string, startTime, deadline time.Time) (uint, error)
}

func (s stubProjectRegisterRepo) FindOrInsertProject(offeringID uint, projectName string, startTime, deadline time.Time) (uint, error) {
	return s.findOrInsertProjectFn(offeringID, projectName, startTime, deadline)
}

func newSheet() *domain.SheetData {
	return domain.NewSheetData(
		[]domain.StudentRow{
			{Number: "20260001", Name: "Alice"},
			{Number: "20260002", Name: "Bob"},
		},
		[]string{"Lab1", "Lab2"},
	)
}

func TestCourseImportService_Import_Orchestration(t *testing.T) {
	deadline := fixedTime().Add(48 * time.Hour)

	var (
		registeredUsers []domain.UserInfo
		gotCourseName   string
		gotOffArgs      [3]string
		gotTeacherArgs  struct {
			id    string
			offID uint
		}
		stuArgs []struct {
			id    string
			offID uint
		}
		projectArgs []struct {
			offID uint
			name  string
		}
	)

	svc := NewCourseImportService(
		stubSheetParser{parseFn: func(io.Reader) (*domain.SheetData, error) { return newSheet(), nil }},
		stubUserRegisterRepo{insertNewUserBatchIgnoreFn: func(users *[]domain.UserInfo) error {
			registeredUsers = append([]domain.UserInfo{}, (*users)...)
			return nil
		}},
		stubCourseRegisterRepo{
			findOrInsertCourseFn: func(name string) (uint, error) {
				gotCourseName = name
				return 3, nil
			},
			findOrInsertCourseOfferingFn: func(courseID uint, className, term string) (uint, error) {
				gotOffArgs = [3]string{
					uintStr(courseID), className, term,
				}
				return 11, nil
			},
			findOrInsertTeacherCourseFn: func(teacherID string, offeringID uint) error {
				gotTeacherArgs.id = teacherID
				gotTeacherArgs.offID = offeringID
				return nil
			},
			findOrInsertStudentCourseFn: func(studentID string, offeringID uint) error {
				stuArgs = append(stuArgs, struct {
					id    string
					offID uint
				}{studentID, offeringID})
				return nil
			},
		},
		stubProjectRegisterRepo{
			findOrInsertProjectFn: func(offeringID uint, projectName string, startTime, dl time.Time) (uint, error) {
				if !dl.Equal(deadline) {
					t.Fatalf("unexpected deadline: %v", dl)
				}
				projectArgs = append(projectArgs, struct {
					offID uint
					name  string
				}{offeringID, projectName})
				return 100, nil
			},
		},
	)

	data := domain.NewImportCourseData("T001", "Database", "", "2025-2026-1", deadline)
	if err := svc.Import(strings.NewReader("ignored"), data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.ClassName != "-" {
		t.Fatalf("empty class should default to '-', got %q", data.ClassName)
	}
	if len(registeredUsers) != 2 {
		t.Fatalf("expected 2 users registered, got %d", len(registeredUsers))
	}
	for _, u := range registeredUsers {
		if u.Identity != 1 {
			t.Fatalf("expected identity=1, got %d", u.Identity)
		}
		if u.Passwd != u.Number {
			t.Fatalf("initial password must equal student number, got passwd=%q num=%q", u.Passwd, u.Number)
		}
	}
	if registeredUsers[0].Name != "Alice" || registeredUsers[1].Name != "Bob" {
		t.Fatalf("expected imported student names to be preserved, got %+v", registeredUsers)
	}
	if gotCourseName != "Database" {
		t.Fatalf("unexpected course name: %q", gotCourseName)
	}
	if gotOffArgs != [3]string{"3", "-", "2025-2026-1"} {
		t.Fatalf("unexpected offering args: %v", gotOffArgs)
	}
	if gotTeacherArgs.id != "T001" || gotTeacherArgs.offID != 11 {
		t.Fatalf("unexpected teacher args: %+v", gotTeacherArgs)
	}
	if len(stuArgs) != 2 || stuArgs[0].id != "20260001" || stuArgs[0].offID != 11 {
		t.Fatalf("unexpected student args: %+v", stuArgs)
	}
	if len(projectArgs) != 2 || projectArgs[0].name != "Lab1" || projectArgs[1].name != "Lab2" {
		t.Fatalf("unexpected project args: %+v", projectArgs)
	}
}

func TestCourseImportService_Import_ParserError(t *testing.T) {
	wantErr := errors.New("parse failed")
	svc := NewCourseImportService(
		stubSheetParser{parseFn: func(io.Reader) (*domain.SheetData, error) { return nil, wantErr }},
		stubUserRegisterRepo{insertNewUserBatchIgnoreFn: func(*[]domain.UserInfo) error {
			t.Fatal("user repo should not be called")
			return nil
		}},
		stubCourseRegisterRepo{},
		stubProjectRegisterRepo{},
	)

	err := svc.Import(bytes.NewReader(nil), domain.NewImportCourseData("T001", "Db", "-", "T", time.Now()))
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected parser error, got %v", err)
	}
}

func TestCourseImportService_Import_StopsOnFirstError(t *testing.T) {
	wantErr := errors.New("course insert failed")
	teacherCalled := false

	svc := NewCourseImportService(
		stubSheetParser{parseFn: func(io.Reader) (*domain.SheetData, error) { return newSheet(), nil }},
		stubUserRegisterRepo{insertNewUserBatchIgnoreFn: func(*[]domain.UserInfo) error { return nil }},
		stubCourseRegisterRepo{
			findOrInsertCourseFn: func(string) (uint, error) { return 0, wantErr },
			findOrInsertTeacherCourseFn: func(string, uint) error {
				teacherCalled = true
				return nil
			},
		},
		stubProjectRegisterRepo{},
	)

	err := svc.Import(strings.NewReader("ignored"),
		domain.NewImportCourseData("T001", "Db", "C", "T", time.Now()))
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected course err, got %v", err)
	}
	if teacherCalled {
		t.Fatal("teacher binding should not run after course-insert failure")
	}
}

func TestCourseImportService_Import_StopsOnStudentCourseError(t *testing.T) {
	wantErr := errors.New("bind student failed")
	projectCalled := false

	svc := NewCourseImportService(
		stubSheetParser{parseFn: func(io.Reader) (*domain.SheetData, error) { return newSheet(), nil }},
		stubUserRegisterRepo{insertNewUserBatchIgnoreFn: func(*[]domain.UserInfo) error { return nil }},
		stubCourseRegisterRepo{
			findOrInsertCourseFn:         func(string) (uint, error) { return 3, nil },
			findOrInsertCourseOfferingFn: func(uint, string, string) (uint, error) { return 11, nil },
			findOrInsertTeacherCourseFn:  func(string, uint) error { return nil },
			findOrInsertStudentCourseFn: func(studentID string, offeringID uint) error {
				if studentID == "20260002" && offeringID == 11 {
					return wantErr
				}
				return nil
			},
		},
		stubProjectRegisterRepo{
			findOrInsertProjectFn: func(uint, string, time.Time, time.Time) (uint, error) {
				projectCalled = true
				return 0, nil
			},
		},
	)

	err := svc.Import(strings.NewReader("ignored"),
		domain.NewImportCourseData("T001", "Db", "C", "T", time.Now()))
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if err == nil || !strings.Contains(err.Error(), `student "20260002"`) {
		t.Fatalf("expected student context in error, got %v", err)
	}
	if projectCalled {
		t.Fatal("project registration should not run after student-course binding failure")
	}
}

func TestCourseImportService_Import_RejectsOversizedStudentFields(t *testing.T) {
	t.Run("student number too long", func(t *testing.T) {
		svc := NewCourseImportService(
			stubSheetParser{parseFn: func(io.Reader) (*domain.SheetData, error) {
				return domain.NewSheetData(
					[]domain.StudentRow{{Number: "12345678901234567", Name: "Alice"}},
					[]string{"Lab1"},
				), nil
			}},
			stubUserRegisterRepo{insertNewUserBatchIgnoreFn: func(*[]domain.UserInfo) error {
				t.Fatal("user repo should not be called")
				return nil
			}},
			stubCourseRegisterRepo{},
			stubProjectRegisterRepo{},
		)

		err := svc.Import(strings.NewReader("ignored"),
			domain.NewImportCourseData("T001", "Db", "C", "T", time.Now()))
		if !errors.Is(err, domain.ErrSheetFormat) {
			t.Fatalf("expected ErrSheetFormat, got %v", err)
		}
	})

	t.Run("student name too long", func(t *testing.T) {
		svc := NewCourseImportService(
			stubSheetParser{parseFn: func(io.Reader) (*domain.SheetData, error) {
				return domain.NewSheetData(
					[]domain.StudentRow{{Number: "20260001", Name: "abcdefghijklmnopq"}},
					[]string{"Lab1"},
				), nil
			}},
			stubUserRegisterRepo{insertNewUserBatchIgnoreFn: func(*[]domain.UserInfo) error {
				t.Fatal("user repo should not be called")
				return nil
			}},
			stubCourseRegisterRepo{},
			stubProjectRegisterRepo{},
		)

		err := svc.Import(strings.NewReader("ignored"),
			domain.NewImportCourseData("T001", "Db", "C", "T", time.Now()))
		if !errors.Is(err, domain.ErrSheetFormat) {
			t.Fatalf("expected ErrSheetFormat, got %v", err)
		}
	})
}

// strconv 局部替代，避免新增 import；测试中只用于把 uint 拼成断言字符串
func uintStr(v uint) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func TestExcelSheetParser_Parse(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		buf := buildXLSX(t,
			[]string{"ID", "SNO", "Sname", "Lab1", "Lab2"},
			[][]string{
				{"1", "20260001", "Alice"},
				{"2", "20260002", "Bob"},
				{"", "", ""},
			})

		got, err := NewExcelSheetParser().Parse(buf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got.Students) != 2 {
			t.Fatalf("expected 2 students, got %d", len(got.Students))
		}
		if got.Students[0].Number != "20260001" || got.Students[0].Name != "Alice" {
			t.Fatalf("unexpected first student: %+v", got.Students[0])
		}
		if len(got.ProjectNames) != 2 || got.ProjectNames[0] != "Lab1" || got.ProjectNames[1] != "Lab2" {
			t.Fatalf("unexpected project names: %v", got.ProjectNames)
		}
	})

	t.Run("invalid header", func(t *testing.T) {
		buf := buildXLSX(t,
			[]string{"foo", "bar", "baz", "Lab1"},
			[][]string{{"1", "20260001", "Alice"}})

		_, err := NewExcelSheetParser().Parse(buf)
		if !errors.Is(err, domain.ErrSheetFormat) {
			t.Fatalf("expected ErrSheetFormat, got %v", err)
		}
	})

	t.Run("missing project columns", func(t *testing.T) {
		buf := buildXLSX(t,
			[]string{"ID", "SNO", "Sname", "   "},
			[][]string{{"1", "20260001", "Alice"}})

		_, err := NewExcelSheetParser().Parse(buf)
		if !errors.Is(err, domain.ErrSheetFormat) {
			t.Fatalf("expected ErrSheetFormat, got %v", err)
		}
	})

	t.Run("not an xlsx", func(t *testing.T) {
		_, err := NewExcelSheetParser().Parse(strings.NewReader("not a real xlsx"))
		if !errors.Is(err, domain.ErrSheetFormat) {
			t.Fatalf("expected ErrSheetFormat, got %v", err)
		}
	})
}

// buildXLSX 在内存里构造一个 xlsx，便于解析器测试
func buildXLSX(t *testing.T, header []string, rows [][]string) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()

	sheet := f.GetSheetName(0)

	for col, v := range header {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			t.Fatalf("CoordinatesToCellName: %v", err)
		}
		if err := f.SetCellValue(sheet, cell, v); err != nil {
			t.Fatalf("SetCellValue: %v", err)
		}
	}
	for r, row := range rows {
		for col, v := range row {
			cell, err := excelize.CoordinatesToCellName(col+1, r+2)
			if err != nil {
				t.Fatalf("CoordinatesToCellName: %v", err)
			}
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				t.Fatalf("SetCellValue: %v", err)
			}
		}
	}
	buf := &bytes.Buffer{}
	if err := f.Write(buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return buf
}
