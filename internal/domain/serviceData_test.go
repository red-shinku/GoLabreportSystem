package domain

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestProjectFileMeta_FilePath_UsesFilesBaseDir(t *testing.T) {
	got, err := NewProjectFileMeta("Database", "Class A", "Lab1", "requirements.docx").FilePath()
	if err != nil {
		t.Fatalf("FilePath() failed: %v", err)
	}

	want := filepath.Join(FilesBaseDir(), "Database", "Class A", "Lab1", "requirements.docx")
	if got != want {
		t.Fatalf("unexpected file path: got %q want %q", got, want)
	}
}

func TestProjectFileMeta_FilePath_RejectsPathEscape(t *testing.T) {
	_, err := NewProjectFileMeta("..", "..", "outside", "requirements.docx").FilePath()
	if !errors.Is(err, ErrNotSafe) {
		t.Fatalf("expected ErrNotSafe, got %v", err)
	}
}

func TestStuReportMeta_FilePath_UsesFilesBaseDir(t *testing.T) {
	got, err := NewStuReportMeta("Database", "Class A", "Alice", "Lab1", "20260001", "pdf").FilePath()
	if err != nil {
		t.Fatalf("FilePath() failed: %v", err)
	}

	want := filepath.Join(FilesBaseDir(), "Database", "Class A", "Lab1", "20260001-Alice-Lab1.pdf")
	if got != want {
		t.Fatalf("unexpected file path: got %q want %q", got, want)
	}
}
