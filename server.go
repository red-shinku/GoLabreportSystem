package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	_ "github.com/go-sql-driver/mysql"

	"LabSystem/database"
	controller "LabSystem/http"
	"LabSystem/http/middleware"
	server "LabSystem/http/router"
	html "LabSystem/http/template"
	"LabSystem/service"
)

const (
	schemaName     = "LabSystem"
	sentinelTable  = "Users"
	initScriptPath = "scripts/init.sql"
)

type Config struct {
	Ipaddr       string `json:"ip_addr" env:"IP_ADDR"`
	Port         string `json:"port" env:"PORT"`
	EnableTLS    bool   `json:"enable_tls" env:"ENABLE_TLS"`
	DatabaseAddr string `json:"database_addr" env:"DATABASE_ADDR"`

	// secret
	JWTSecret     string `env:"JWT_SECRET"`
	JWTSecretFile string `json:"jwt_secret_file"`
}

func LoadJSONConfig(path string) Config {
	var cfg Config

	b, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		panic(err)
	}

	return cfg
}

func LoadEnvConfig(cfg *Config) {
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
}

func (c *Config) LoadSecret() error {
	// ENV 优先
	if c.JWTSecret != "" {
		return nil
	}

	if c.JWTSecretFile != "" {
		b, err := os.ReadFile(c.JWTSecretFile)
		if err != nil {
			return err
		}
		c.JWTSecret = strings.TrimSpace(string(b))
		return nil
	}

	return errors.New("missing JWT secret")
}

func applyDefaults(c *Config) {
	if c.Ipaddr == "" {
		c.Ipaddr = "127.0.0.1"
	}
	if c.Port == "" {
		c.Port = "8080"
	}
}

func LoadConfig() Config {
	// 读取JSON配置
	cfg := LoadJSONConfig("config.json")
	// 用环境变量配置覆盖
	LoadEnvConfig(&cfg)
	// 加载JWT密钥
	if err := cfg.LoadSecret(); err != nil {
		panic(err)
	}
	// 未配置选项设置默认值
	applyDefaults(&cfg)

	return cfg
}

// EnsureSchema 检测数据库是否已初始化，若未初始化则执行 scripts/init.sql。
// 以 LabSystem 库下是否存在哨兵表（Users）判断初始化状态。
func EnsureSchema(db *sql.DB) error {
	var count int
	err := db.QueryRow(
		"select count(*) from information_schema.tables where table_schema = ? and table_name = ?",
		schemaName, sentinelTable,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check schema: %w", err)
	}
	if count > 0 {
		return nil
	}

	log.Printf("schema %q not initialized, running %s", schemaName, initScriptPath)
	script, err := os.ReadFile(initScriptPath)
	if err != nil {
		return fmt.Errorf("read init script: %w", err)
	}

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Close()

	for _, stmt := range splitSQLStatements(string(script)) {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec init stmt %q: %w", firstLine(stmt), err)
		}
	}
	return nil
}

// splitSQLStatements 按分号拆分多条 SQL 语句，过滤注释与空行。
func splitSQLStatements(script string) []string {
	var stmts []string
	var buf strings.Builder
	for _, raw := range strings.Split(script, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
		if strings.HasSuffix(line, ";") {
			stmts = append(stmts, strings.TrimSpace(buf.String()))
			buf.Reset()
		}
	}
	if tail := strings.TrimSpace(buf.String()); tail != "" {
		stmts = append(stmts, tail)
	}
	return stmts
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}

func main() {
	cfg := LoadConfig()

	//建立数据库连接
	db, err := sql.Open("mysql", cfg.DatabaseAddr)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	// 初始化数据库
	if err := EnsureSchema(db); err != nil {
		log.Fatalf("ensure schema: %v", err)
	}

	middleware.Secret = []byte(cfg.JWTSecret)

	usersRepo := database.NewUsersRepo(db)
	projectRepo := database.NewProjectRepo(db)
	reportRepo := database.NewReportRepo(db)
	manCourseOfferRepo := database.NewManCourseOfferRepo(db)
	//FIXME: 补充课表导入功能
	_ = manCourseOfferRepo

	fileService := service.NewFileService()
	authService := service.NewAuthService(usersRepo)
	tecProjectService := service.NewTeacherProjectService(projectRepo, projectRepo, fileService)
	stuProjectService := service.NewStudentProjectService(projectRepo, projectRepo, fileService)
	tecReportService := service.NewTeacherReportService(reportRepo, fileService)
	stuReportService := service.NewStudentReportService(reportRepo, projectRepo, usersRepo, fileService)

	lgPageGen := html.NewLoginPageGenerator()
	stuHomeGen := html.NewStuHomeGenerator()
	tecHomeGen := html.NewTecHomeGenerator()

	sessionsCtl := controller.NewSessions(authService, cfg.JWTSecret)
	homeCtl := controller.NewHome(tecProjectService, stuProjectService, lgPageGen, stuHomeGen, tecHomeGen)
	offeringClassCtl := controller.NewOfferingClass(tecProjectService, tecHomeGen)
	projectsCtl := controller.NewProjects(tecProjectService, stuProjectService, tecReportService, stuReportService, tecHomeGen, stuHomeGen)
	submissionsCtl := controller.NewSubmissions(tecReportService)

	// 路由初始化
	mux := http.NewServeMux()
	router := server.NewRouter(mux, homeCtl, sessionsCtl, offeringClassCtl, projectsCtl, submissionsCtl)
	router.Init()

	addr := cfg.Ipaddr + ":" + cfg.Port
	log.Printf("server listening on %s", addr)
	if cfg.EnableTLS {
		// TODO: 证书的管理
		if err := http.ListenAndServeTLS(addr, "server.crt", "server.key", mux); err != nil {
			log.Fatalf("serve tls: %v", err)
		}
		return
	}
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
