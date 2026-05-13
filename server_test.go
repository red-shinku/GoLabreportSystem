package main

import (
	"os"
	"path/filepath"
	"testing"

	mysql "github.com/go-sql-driver/mysql"

	"LabSystem/internal/domain"
)

func TestBuildMySQLDSN_EnablesTimeParsing(t *testing.T) {
	cfg := &Config{
		DatabaseAddr:   "127.0.0.1",
		DatabasePort:   "3306",
		DatabaseUser:   "tester",
		DatabasePasswd: "secret",
	}

	dsn := buildMySQLDSN(cfg, schemaName)
	parsed, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("ParseDSN() failed: %v", err)
	}

	if !parsed.ParseTime {
		t.Fatal("expected ParseTime=true in MySQL DSN")
	}
	if parsed.DBName != schemaName {
		t.Fatalf("unexpected db name: %q", parsed.DBName)
	}
	if parsed.Addr != "127.0.0.1:3306" {
		t.Fatalf("unexpected addr: %q", parsed.Addr)
	}
}

func TestEnsureFilesDir_CreatesFilesDirectoryInWorkingDir(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() failed: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir(%q) failed: %v", tmp, err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory failed: %v", err)
		}
	}()

	if err := EnsureFilesDir(); err != nil {
		t.Fatalf("EnsureFilesDir() failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(tmp, domain.FilesBaseDir()))
	if err != nil {
		t.Fatalf("Stat(files) failed: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected files path to be a directory")
	}
}
