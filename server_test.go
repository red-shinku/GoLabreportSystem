package main

import (
	"testing"

	mysql "github.com/go-sql-driver/mysql"
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
