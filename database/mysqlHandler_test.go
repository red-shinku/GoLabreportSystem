package database

import (
	"LabSystem/internal/domain"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() failed: %v", err)
	}
	return db, mock
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 10, 12, 30, 0, 0, time.UTC)
}

func mustExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func userPasswdSQL() string {
	return fmt.Sprintf("select passwd from %s where number = ?", tabUsers)
}

func userIdentitySQL() string {
	return fmt.Sprintf("select identity from %s where number = ?", tabUsers)
}

func userStudentNameSQL() string {
	return fmt.Sprintf("select name from %s where number = ?", tabUsers)
}

func insertUserSQL() string {
	return fmt.Sprintf("insert into %s (identity, number, passwd) values (?, ?, ?)", tabUsers)
}

func changePasswordSQL() string {
	return fmt.Sprintf("update %s set passwd = ? where number = ?", tabUsers)
}

func insertStudentCourseOfferSQL() string {
	return fmt.Sprintf("insert into %s (studentID, offeringID) values (?, ?)", tabStudentCourse)
}

func queryStuProjectSQL() string {
	return fmt.Sprintf("select c.courseName, p.projectID, p.projectName, p.startTime, p.deadline, p.isActive, srp.stuReportID "+
		"from %s stuc "+
		"join %s coff on stuc.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"join %s srp on p.projectID = srp.projectID and stuc.studentID = srp.studentID "+
		"where stuc.studentID = ?",
		tabStudentCourse, tabCourseOffering, tabCourse, tabProject, tabStuReport)
}

func queryStuProjectByIDSQL() string {
	return fmt.Sprintf("select c.courseName, p.projectID, p.projectName, p.startTime, p.deadline, p.isActive, srp.stuReportID "+
		"from %s p "+
		"join %s coff on p.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"left join %s srp on p.projectID = srp.projectID and srp.studentID = ? "+
		"where p.projectID = ?",
		tabProject, tabCourseOffering, tabCourse, tabStuReport)
}

func queryTeacherProjectSQL() string {
	return fmt.Sprintf("select c.courseName, coff.className, coff.offeringID, p.projectID, p.projectName, p.deadline, p.isActive "+
		"from %s tec "+
		"join %s coff on tec.offeringID = coff.offeringID "+
		"join %s c on coff.courseID = c.courseID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"where tec.teacherID = ?",
		tabTeacherCourse, tabCourseOffering, tabCourse, tabProject)
}

func updateProjectFlagSQL() string {
	return fmt.Sprintf("update %s set isActive = ? where projectID = ?", tabProject)
}

func queryProjectFlagSQL() string {
	return fmt.Sprintf("select isActive from %s where projectID = ?", tabProject)
}

func queryOfferingInfoSQL() string {
	return fmt.Sprintf("select c.courseName, coff.className "+
		"from %s coff "+
		"join %s c on coff.courseID = c.courseID "+
		"where coff.offeringID = ?",
		tabCourseOffering, tabCourse)
}

func addProjectSQL() string {
	return fmt.Sprintf("insert into %s (offeringID, projectName, projectFilePath, startTime, deadline) values (?, ?, ?, ?, ?)", tabProject)
}

func queryProjectByIDSQL() string {
	return fmt.Sprintf("select projectName, deadline, isActive from %s where projectID = ?", tabProject)
}

func deleteProjectSQL() string {
	return fmt.Sprintf("delete from %s where projectID = ?", tabProject)
}

func updateProjectFileSQL() string {
	return fmt.Sprintf("update %s set projectFilePath = ? where projectID = ?", tabProject)
}

func queryProjectFileSQL() string {
	return fmt.Sprintf("select projectFilePath from %s where projectID = ?", tabProject)
}

func queryProjectInfoSQL() string {
	return fmt.Sprintf(
		"select c.courseName, coff.className, pj.projectName "+
			"from %s pj "+
			"join %s coff on pj.offeringID = coff.offeringID "+
			"join %s c on coff.courseID = c.courseID "+
			"where pj.projectID = ?",
		tabProject, tabCourseOffering, tabCourse)
}

func insertStuReportSQL() string {
	return fmt.Sprintf("insert into %s (studentID, projectID, reportFilePath, submitTime) values (?, ?, ?, ?)", tabStuReport)
}

func queryStuReportFileAllSQL() string {
	return fmt.Sprintf("select reportFilePath from %s where projectID = ?", tabStuReport)
}

func queryStuReportStatusSQL() string {
	return fmt.Sprintf("select stuc.studentID, srp.stuReportID "+
		"from %s stuc "+
		"join %s coff on stuc.offeringID = coff.offeringID "+
		"join %s p on coff.offeringID = p.offeringID "+
		"left join %s srp on p.projectID = srp.projectID and stuc.studentID = srp.studentID "+
		"where p.projectID = ?",
		tabStudentCourse, tabCourseOffering, tabProject, tabStuReport)
}

func TestUsersRepo_QueryPasswd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(userPasswdSQL())).
			WithArgs("20260001").
			WillReturnRows(sqlmock.NewRows([]string{"passwd"}).AddRow("hashed-passwd"))

		repo := &UsersRepo{db: db}
		got, err := repo.QueryPasswd("20260001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hashed-passwd" {
			t.Fatalf("unexpected passwd: got %q", got)
		}

		mustExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(userPasswdSQL())).
			WithArgs("20260002").
			WillReturnError(sql.ErrNoRows)

		repo := &UsersRepo{db: db}
		_, err := repo.QueryPasswd("20260002")
		if err == nil || !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}

		mustExpectations(t, mock)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(userPasswdSQL())).
			WithArgs("20260003").
			WillReturnError(errors.New("db down"))

		repo := &UsersRepo{db: db}
		_, err := repo.QueryPasswd("20260003")
		if err == nil || !errors.Is(err, domain.ErrQuery) {
			t.Fatalf("expected ErrQuery, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestUsersRepo_WhoAmI(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(userIdentitySQL())).
		WithArgs("20260001").
		WillReturnRows(sqlmock.NewRows([]string{"identity"}).AddRow(uint8(2)))

	repo := &UsersRepo{db: db}
	got, err := repo.WhoAmI("20260001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2 {
		t.Fatalf("unexpected identity: got %d", got)
	}

	mustExpectations(t, mock)
}

func TestUsersRepo_InsertNewUser(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		repo := &UsersRepo{}
		if err := repo.InsertNewUser(nil); err == nil {
			t.Fatal("expected error for nil input")
		}
	})

	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		user := &domain.UserInfo{
			Identity: 1,
			Number:   "20260001",
			Passwd:   "pass",
		}

		mock.ExpectExec(regexp.QuoteMeta(insertUserSQL())).
			WithArgs(user.Identity, user.Number, user.Passwd).
			WillReturnResult(sqlmock.NewResult(1, 1))

		repo := &UsersRepo{db: db}
		if err := repo.InsertNewUser(user); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		mustExpectations(t, mock)
	})

	t.Run("exec error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		user := &domain.UserInfo{
			Identity: 1,
			Number:   "20260002",
			Passwd:   "pass",
		}

		mock.ExpectExec(regexp.QuoteMeta(insertUserSQL())).
			WithArgs(user.Identity, user.Number, user.Passwd).
			WillReturnError(errors.New("insert failed"))

		repo := &UsersRepo{db: db}
		err := repo.InsertNewUser(user)
		if err == nil || !errors.Is(err, domain.ErrModify) {
			t.Fatalf("expected ErrModify, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestUsersRepo_InsertNewUserBatch(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		repo := &UsersRepo{}
		if err := repo.InsertNewUserBatch(nil); err == nil {
			t.Fatal("expected error for nil input")
		}
	})

	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		users := []domain.UserInfo{
			{Identity: 1, Number: "20260001", Passwd: "p1"},
			{Identity: 2, Number: "20260002", Passwd: "p2"},
		}

		prep := mock.ExpectPrepare(regexp.QuoteMeta(insertUserSQL()))
		prep.ExpectExec().
			WithArgs(users[0].Identity, users[0].Number, users[0].Passwd).
			WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().
			WithArgs(users[1].Identity, users[1].Number, users[1].Passwd).
			WillReturnResult(sqlmock.NewResult(2, 1))

		repo := &UsersRepo{db: db}
		if err := repo.InsertNewUserBatch(&users); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		mustExpectations(t, mock)
	})

	t.Run("prepare error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		users := []domain.UserInfo{{Identity: 1, Number: "20260001", Passwd: "p1"}}
		mock.ExpectPrepare(regexp.QuoteMeta(insertUserSQL())).
			WillReturnError(errors.New("prepare failed"))

		repo := &UsersRepo{db: db}
		err := repo.InsertNewUserBatch(&users)
		if err == nil || !errors.Is(err, domain.ErrModify) {
			t.Fatalf("expected ErrModify, got: %v", err)
		}

		mustExpectations(t, mock)
	})

	t.Run("exec error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		users := []domain.UserInfo{
			{Identity: 1, Number: "20260001", Passwd: "p1"},
			{Identity: 2, Number: "20260002", Passwd: "p2"},
		}

		prep := mock.ExpectPrepare(regexp.QuoteMeta(insertUserSQL()))
		prep.ExpectExec().
			WithArgs(users[0].Identity, users[0].Number, users[0].Passwd).
			WillReturnResult(sqlmock.NewResult(1, 1))
		prep.ExpectExec().
			WithArgs(users[1].Identity, users[1].Number, users[1].Passwd).
			WillReturnError(errors.New("exec failed"))

		repo := &UsersRepo{db: db}
		err := repo.InsertNewUserBatch(&users)
		if err == nil || !errors.Is(err, domain.ErrModify) {
			t.Fatalf("expected ErrModify, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestUsersRepo_ChangePassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectExec(regexp.QuoteMeta(changePasswordSQL())).
			WithArgs("new-pass", "20260001").
			WillReturnResult(sqlmock.NewResult(0, 1))

		repo := &UsersRepo{db: db}
		if err := repo.ChangePassword("20260001", "new-pass"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		mustExpectations(t, mock)
	})

	t.Run("exec error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectExec(regexp.QuoteMeta(changePasswordSQL())).
			WithArgs("new-pass", "20260001").
			WillReturnError(errors.New("update failed"))

		repo := &UsersRepo{db: db}
		err := repo.ChangePassword("20260001", "new-pass")
		if err == nil || !errors.Is(err, domain.ErrModify) {
			t.Fatalf("expected ErrModify, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestUsersRepo_QueryStudentName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(userStudentNameSQL())).
			WithArgs("20260001").
			WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Alice"))

		repo := &UsersRepo{db: db}
		got, err := repo.QueryStudentName("20260001")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "Alice" {
			t.Fatalf("unexpected name: %q", got)
		}

		mustExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(userStudentNameSQL())).
			WithArgs("20260002").
			WillReturnError(sql.ErrNoRows)

		repo := &UsersRepo{db: db}
		_, err := repo.QueryStudentName("20260002")
		if err == nil || !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestProjectRepo_QueryStuProject(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	now := fixedTime()
	later := now.Add(48 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"courseName", "projectID", "projectName", "startTime", "deadline", "isActive", "stuReportID",
	}).AddRow("Go Basics", int64(1001), "Lab1", now, later, true, int64(9001))

	mock.ExpectQuery(regexp.QuoteMeta(queryStuProjectSQL())).
		WithArgs("20260001").
		WillReturnRows(rows)

	repo := &ProjectRepo{db: db}
	got, err := repo.QueryStuProject("20260001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got))
	}
	if got[0].CourseName != "Go Basics" || got[0].ProjectName != "Lab1" || got[0].IsActive != true {
		t.Fatalf("unexpected row: %+v", got[0])
	}

	mustExpectations(t, mock)
}

func TestProjectRepo_QueryStuProjectByID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	now := fixedTime()
	later := now.Add(24 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta(queryStuProjectByIDSQL())).
		WithArgs("20260001", uint(1001)).
		WillReturnRows(sqlmock.NewRows([]string{
			"courseName", "projectID", "projectName", "startTime", "deadline", "isActive", "stuReportID",
		}).AddRow("Go Basics", int64(1001), "Lab1", now, later, true, int64(9001)))

	repo := &ProjectRepo{db: db}
	got, err := repo.QueryStuProjectByID("20260001", 1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ProjectID != 1001 || got.ProjectName != "Lab1" || !got.StuReportID.Valid || got.StuReportID.Int32 != 9001 {
		t.Fatalf("unexpected result: %+v", got)
	}

	mustExpectations(t, mock)
}

func TestProjectRepo_QueryTeacherProject(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	now := fixedTime()

	rows := sqlmock.NewRows([]string{
		"courseName", "className", "offeringID", "projectID", "projectName", "deadline", "isActive",
	}).AddRow("Go Basics", "Class A", int64(11), int64(1001), "Lab1", now, true)

	mock.ExpectQuery(regexp.QuoteMeta(queryTeacherProjectSQL())).
		WithArgs("T001").
		WillReturnRows(rows)

	repo := &ProjectRepo{db: db}
	got, err := repo.QueryTeacherProject("T001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].ClassName != "Class A" || got[0].ProjectID != 1001 {
		t.Fatalf("unexpected result: %+v", got)
	}

	mustExpectations(t, mock)
}

func TestProjectRepo_UpdateProjectFlag(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(updateProjectFlagSQL())).
		WithArgs(true, uint(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := &ProjectRepo{db: db}
	if err := repo.UpdateProjectFlag(1001, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustExpectations(t, mock)
}

func TestProjectRepo_QueryProjectFlag(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryProjectFlagSQL())).
			WithArgs(uint(1001)).
			WillReturnRows(sqlmock.NewRows([]string{"isActive"}).AddRow(true))

		repo := &ProjectRepo{db: db}
		got, err := repo.QueryProjectFlag(1001)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Fatal("expected true")
		}

		mustExpectations(t, mock)
	})

	t.Run("query error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryProjectFlagSQL())).
			WithArgs(uint(1001)).
			WillReturnError(errors.New("query failed"))

		repo := &ProjectRepo{db: db}
		_, err := repo.QueryProjectFlag(1001)
		if err == nil || !strings.Contains(err.Error(), "QueryProjectFlag():") {
			t.Fatalf("unexpected error: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestProjectRepo_QueryOfferingInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryOfferingInfoSQL())).
			WithArgs(uint(11)).
			WillReturnRows(sqlmock.NewRows([]string{"courseName", "className"}).AddRow("Go Basics", "Class A"))

		repo := &ProjectRepo{db: db}
		courseName, className, err := repo.QueryOfferingInfo(11)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if courseName != "Go Basics" || className != "Class A" {
			t.Fatalf("unexpected result: %q, %q", courseName, className)
		}

		mustExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryOfferingInfoSQL())).
			WithArgs(uint(11)).
			WillReturnError(sql.ErrNoRows)

		repo := &ProjectRepo{db: db}
		_, _, err := repo.QueryOfferingInfo(11)
		if err == nil || !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestProjectRepo_AddProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		now := fixedTime()
		project := &domain.ProjectInfo{
			OfferingID:      11,
			ProjectName:     "Lab1",
			ProjectFilePath: "/tmp/lab1.zip",
			StartTime:       now,
			CloseTime:       now.Add(24 * time.Hour),
		}

		mock.ExpectExec(regexp.QuoteMeta(addProjectSQL())).
			WithArgs(project.OfferingID, project.ProjectName, project.ProjectFilePath, project.StartTime, project.CloseTime).
			WillReturnResult(sqlmock.NewResult(123, 1))

		repo := &ProjectRepo{db: db}
		got, err := repo.AddProject(project)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 123 {
			t.Fatalf("unexpected id: %d", got)
		}

		mustExpectations(t, mock)
	})

	t.Run("last insert id error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		now := fixedTime()
		project := &domain.ProjectInfo{
			OfferingID:      11,
			ProjectName:     "Lab1",
			ProjectFilePath: "/tmp/lab1.zip",
			StartTime:       now,
			CloseTime:       now.Add(24 * time.Hour),
		}

		mock.ExpectExec(regexp.QuoteMeta(addProjectSQL())).
			WithArgs(project.OfferingID, project.ProjectName, project.ProjectFilePath, project.StartTime, project.CloseTime).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("last insert id failed")))

		repo := &ProjectRepo{db: db}
		_, err := repo.AddProject(project)
		if err == nil || !errors.Is(err, domain.ErrModify) {
			t.Fatalf("expected ErrModify, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestProjectRepo_QueryProjectByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		now := fixedTime()

		mock.ExpectQuery(regexp.QuoteMeta(queryProjectByIDSQL())).
			WithArgs(uint(1001)).
			WillReturnRows(sqlmock.NewRows([]string{"projectName", "deadline", "isActive"}).AddRow("Lab1", now, true))

		repo := &ProjectRepo{db: db}
		got, err := repo.QueryProjectByID(1001)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ProjectID != 1001 || got.ProjectName != "Lab1" || !got.IsActive {
			t.Fatalf("unexpected result: %+v", got)
		}

		mustExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryProjectByIDSQL())).
			WithArgs(uint(1001)).
			WillReturnError(sql.ErrNoRows)

		repo := &ProjectRepo{db: db}
		_, err := repo.QueryProjectByID(1001)
		if err == nil || !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestProjectRepo_DelProject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectExec(regexp.QuoteMeta(deleteProjectSQL())).
			WithArgs(uint(1001)).
			WillReturnResult(sqlmock.NewResult(0, 1))

		repo := &ProjectRepo{db: db}
		if err := repo.DelProject(1001); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		mustExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectExec(regexp.QuoteMeta(deleteProjectSQL())).
			WithArgs(uint(1001)).
			WillReturnResult(sqlmock.NewResult(0, 0))

		repo := &ProjectRepo{db: db}
		err := repo.DelProject(1001)
		if err == nil || !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestProjectRepo_UpdateProjectFile(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(updateProjectFileSQL())).
		WithArgs("/new/path/report.zip", uint(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := &ProjectRepo{db: db}
	if err := repo.UpdateProjectFile(1001, "/new/path/report.zip"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustExpectations(t, mock)
}

func TestProjectRepo_QueryProjectFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryProjectFileSQL())).
			WithArgs(uint(1001)).
			WillReturnRows(sqlmock.NewRows([]string{"projectFilePath"}).AddRow("/tmp/lab1.zip"))

		repo := &ProjectRepo{db: db}
		got, err := repo.QueryProjectFile(1001)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "/tmp/lab1.zip" {
			t.Fatalf("unexpected path: %q", got)
		}

		mustExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryProjectFileSQL())).
			WithArgs(uint(1001)).
			WillReturnError(sql.ErrNoRows)

		repo := &ProjectRepo{db: db}
		_, err := repo.QueryProjectFile(1001)
		if err == nil || !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestProjectRepo_QueryProjectInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryProjectInfoSQL())).
			WithArgs(uint(1001)).
			WillReturnRows(sqlmock.NewRows([]string{"courseName", "className", "projectName"}).AddRow("Go Basics", "Class A", "Lab1"))

		repo := &ProjectRepo{db: db}
		courseName, className, projectName, err := repo.QueryProjectInfo(1001)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if courseName != "Go Basics" || className != "Class A" || projectName != "Lab1" {
			t.Fatalf("unexpected result: %q, %q, %q", courseName, className, projectName)
		}

		mustExpectations(t, mock)
	})

	t.Run("not found", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(queryProjectInfoSQL())).
			WithArgs(uint(1001)).
			WillReturnError(sql.ErrNoRows)

		repo := &ProjectRepo{db: db}
		_, _, _, err := repo.QueryProjectInfo(1001)
		if err == nil || !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestReportRepo_QueryStuReportStatus(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"studentID", "stuReportID"}).
		AddRow("20260001", int64(9001)).
		AddRow("20260002", int64(0))

	mock.ExpectQuery(regexp.QuoteMeta(queryStuReportStatusSQL())).
		WithArgs(uint(1001)).
		WillReturnRows(rows)

	repo := &ReportRepo{db: db}
	got, err := repo.QueryStuReportStatus(1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}
	if got[0].StudentID != "20260001" || !got[0].StuReportID.Valid || got[0].StuReportID.Int32 != 9001 {
		t.Fatalf("unexpected first row: %+v", got[0])
	}

	mustExpectations(t, mock)
}

func TestReportRepo_InsertStuReport(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		now := fixedTime()
		stuRp := &domain.StuReportInfo{
			StudentID:      "20260001",
			ProjectID:      1001,
			ReportFilePath: "/tmp/report.pdf",
			SubmitTime:     now,
		}

		mock.ExpectExec(regexp.QuoteMeta(insertStuReportSQL())).
			WithArgs(stuRp.StudentID, stuRp.ProjectID, stuRp.ReportFilePath, stuRp.SubmitTime).
			WillReturnResult(sqlmock.NewResult(888, 1))

		repo := &ReportRepo{db: db}
		got, err := repo.InsertStuReport(stuRp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 888 {
			t.Fatalf("unexpected id: %d", got)
		}

		mustExpectations(t, mock)
	})

	t.Run("last insert id error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		now := fixedTime()
		stuRp := &domain.StuReportInfo{
			StudentID:      "20260001",
			ProjectID:      1001,
			ReportFilePath: "/tmp/report.pdf",
			SubmitTime:     now,
		}

		mock.ExpectExec(regexp.QuoteMeta(insertStuReportSQL())).
			WithArgs(stuRp.StudentID, stuRp.ProjectID, stuRp.ReportFilePath, stuRp.SubmitTime).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("last insert id failed")))

		repo := &ReportRepo{db: db}
		_, err := repo.InsertStuReport(stuRp)
		if err == nil || !errors.Is(err, domain.ErrModify) {
			t.Fatalf("expected ErrModify, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}

func TestReportRepo_QueryStuReportFileAll(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(queryStuReportFileAllSQL())).
		WithArgs(uint(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"reportFilePath"}).
			AddRow("/tmp/a.pdf").
			AddRow("/tmp/b.pdf"))

	repo := &ReportRepo{db: db}
	got, err := repo.QueryStuReportFileAll(1001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != "/tmp/a.pdf" || got[1] != "/tmp/b.pdf" {
		t.Fatalf("unexpected result: %#v", got)
	}

	mustExpectations(t, mock)
}

func TestManCourseOfferRepo_InsertStuCourseOfferBatch(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		repo := &ManCourseOfferRepo{}
		if err := repo.InsertStuCourseOfferBatch(nil); err == nil {
			t.Fatal("expected error for nil input")
		}
	})

	t.Run("success", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		courses := []domain.StudentCourseInfo{
			{StudentID: "20260001", OfferingID: 11},
			{StudentID: "20260002", OfferingID: 12},
		}

		mock.ExpectExec(regexp.QuoteMeta(insertStudentCourseOfferSQL())).
			WithArgs(courses[0].StudentID, courses[0].OfferingID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(insertStudentCourseOfferSQL())).
			WithArgs(courses[1].StudentID, courses[1].OfferingID).
			WillReturnResult(sqlmock.NewResult(2, 1))

		repo := &ManCourseOfferRepo{db: db}
		if err := repo.InsertStuCourseOfferBatch(&courses); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		mustExpectations(t, mock)
	})

	t.Run("exec error", func(t *testing.T) {
		db, mock := newMockDB(t)
		defer db.Close()

		courses := []domain.StudentCourseInfo{
			{StudentID: "20260001", OfferingID: 11},
			{StudentID: "20260002", OfferingID: 12},
		}

		mock.ExpectExec(regexp.QuoteMeta(insertStudentCourseOfferSQL())).
			WithArgs(courses[0].StudentID, courses[0].OfferingID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(insertStudentCourseOfferSQL())).
			WithArgs(courses[1].StudentID, courses[1].OfferingID).
			WillReturnError(errors.New("insert failed"))

		repo := &ManCourseOfferRepo{db: db}
		err := repo.InsertStuCourseOfferBatch(&courses)
		if err == nil || !errors.Is(err, domain.ErrModify) {
			t.Fatalf("expected ErrModify, got: %v", err)
		}

		mustExpectations(t, mock)
	})
}
