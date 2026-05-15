package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	mysql "github.com/go-sql-driver/mysql"

	"LabSystem/database"
	controller "LabSystem/http"
	"LabSystem/http/middleware"
	server "LabSystem/http/router"
	html "LabSystem/http/template"
	"LabSystem/internal/domain"
	"LabSystem/service"
)

const (
	schemaName     = "LabSystem"
	sentinelTable  = "Users"
	initScriptPath = "scripts/init.sql"
)

// Config 所有配置选项
type Config struct {
	Ipaddr    string `json:"ip_addr" env:"IP_ADDR"`
	Port      string `json:"port" env:"PORT"`
	EnableTLS bool   `json:"enable_tls" env:"ENABLE_TLS"`

	// pprof 性能分析
	EnablePprof bool   `json:"enable_pprof" env:"ENABLE_PPROF"`
	PprofAddr   string `json:"pprof_addr" env:"PPROF_ADDR"`

	// JWT secret
	JWTSecret     string `env:"JWT_SECRET"`
	JWTSecretFile string `json:"jwt_secret_file" env:"JWT_SECRET_FILE"`

	//Database
	DatabaseAddr       string `json:"database_addr" env:"DATABASE_ADDR"`
	DatabasePort       string `json:"database_port" env:"DATABASE_PORT"`
	DatabaseUser       string `json:"database_user" env:"DATABASE_USER"`
	DatabasePasswd     string `env:"DATABASE_PASSWD"`
	DatabasePasswdFile string `json:"database_passwd_file" env:"DATABASE_PASSWD_FILE"`

	// Database connection pool（durations 以秒为单位；<=0 表示使用默认值）
	DBMaxOpenConns       int `json:"db_max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	DBMaxIdleConns       int `json:"db_max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
	DBConnMaxLifetimeSec int `json:"db_conn_max_lifetime_sec" env:"DB_CONN_MAX_LIFETIME_SEC"`
	DBConnMaxIdleTimeSec int `json:"db_conn_max_idle_time_sec" env:"DB_CONN_MAX_IDLE_TIME_SEC"`
}

// LoadJSONConfig 读取config.json
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

// LoadEnvConfig 读取环境变量
func LoadEnvConfig(cfg *Config) {
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}
}

func (c *Config) LoadJWTSecret() error {
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

func (c *Config) LoadDBPasswd() error {
	if c.DatabasePasswd != "" {
		return nil
	}
	if c.DatabasePasswdFile != "" {
		b, err := os.ReadFile(c.DatabasePasswdFile)
		if err != nil {
			return err
		}
		c.DatabasePasswd = strings.TrimSpace(string(b))
		return nil
	}
	return errors.New("missing database passwd")
}

// applyDefaults 保底设置部分配置的默认值
func applyDefaults(c *Config) {
	if c.Ipaddr == "" {
		c.Ipaddr = "0.0.0.0"
	}
	if c.Port == "" {
		c.Port = "8080"
	}
	if c.DBMaxOpenConns <= 0 {
		c.DBMaxOpenConns = 50
	}
	if c.DBMaxIdleConns <= 0 {
		c.DBMaxIdleConns = 10
	}
	if c.DBConnMaxLifetimeSec <= 0 {
		c.DBConnMaxLifetimeSec = 1800
	}
	if c.DBConnMaxIdleTimeSec <= 0 {
		c.DBConnMaxIdleTimeSec = 300
	}
	if c.PprofAddr == "" {
		c.PprofAddr = "127.0.0.1:6060"
	}
}

// LoadConfig 从环境变量、配置文件加载配置，并读取可能存在的密钥文件
func LoadConfig() Config {
	// 读取JSON配置
	cfg := LoadJSONConfig("config.json")
	// 用环境变量配置覆盖
	LoadEnvConfig(&cfg)
	// 加载JWT密钥
	if err := cfg.LoadJWTSecret(); err != nil {
		panic(err)
	}
	// 加载数据库密码
	if err := cfg.LoadDBPasswd(); err != nil {
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

	log.Printf("[warnning] schema %q not initialized, running %s", schemaName, initScriptPath)
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

func buildMySQLDSN(cfg *Config, dbName string) string {
	return (&mysql.Config{
		User:      cfg.DatabaseUser,
		Passwd:    cfg.DatabasePasswd,
		Net:       "tcp",
		Addr:      net.JoinHostPort(cfg.DatabaseAddr, cfg.DatabasePort),
		DBName:    dbName,
		ParseTime: true,
		Loc:       time.Local,
	}).FormatDSN()
}

// EnsureFilesDir 确保存放项目或报告文件的基文件夹存在
func EnsureFilesDir() error {
	filesDirPath := domain.FilesBaseDir()
	info, err := os.Stat(filesDirPath)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", filesDirPath)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", filesDirPath, err)
	}
	if err := os.MkdirAll(filesDirPath, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", filesDirPath, err)
	}
	return nil
}

// configureDBPool 应用连接池参数（最大连接数、空闲连接数、连接存活/空闲时长）。
func configureDBPool(db *sql.DB, cfg *Config) {
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeSec) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.DBConnMaxIdleTimeSec) * time.Second)
}

// ConnectDB 生成DSN并建立数据库连接
func ConnectDB(cfg *Config) {
	dsn := buildMySQLDSN(cfg, "")
	tempdb, errt := sql.Open("mysql", dsn)
	if errt != nil {
		log.Fatalf("open database: %v", errt)
	}
	if err := tempdb.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	// 检查并初始化数据库
	if err := EnsureSchema(tempdb); err != nil {
		log.Fatalf("ensure schema: %v", err)
	}
	tempdb.Close()

	dsnFinal := buildMySQLDSN(cfg, schemaName)
	var err error
	db, err = sql.Open("mysql", dsnFinal)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	configureDBPool(db, cfg)
}

var (
	db *sql.DB
)

// startPprofServer 在独立协程上启动 pprof HTTP 端点，仅注册 net/http/pprof 的路由，
// 避免污染业务 mux。监听地址默认 127.0.0.1:6060，可通过 cfg.PprofAddr 覆盖。
func startPprofServer(addr string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))

	srv := &http.Server{Addr: addr, Handler: mux}
	log.Printf("pprof listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("pprof server: %v", err)
	}
}

func main() {
	cfg := LoadConfig()

	if cfg.EnablePprof {
		go startPprofServer(cfg.PprofAddr)
	}

	if err := EnsureFilesDir(); err != nil {
		log.Fatalf("ensure files dir: %v", err)
	}
	//连接数据库
	ConnectDB(&cfg)
	defer db.Close()

	middleware.Secret = []byte(cfg.JWTSecret)

	usersRepo := database.NewUsersRepo(db)
	projectRepo := database.NewProjectRepo(db)
	reportRepo := database.NewReportRepo(db)
	//manCourseOfferRepo := database.NewManCourseOfferRepo(db)
	courseRepo := database.NewCourseRepo(db)
	//_ = manCourseOfferRepo

	fileService := service.NewFileService()
	authService := service.NewAuthService(usersRepo)
	tecProjectService := service.NewTeacherProjectService(projectRepo, projectRepo, fileService)
	stuProjectService := service.NewStudentProjectService(projectRepo, projectRepo, fileService)
	tecReportService := service.NewTeacherReportService(reportRepo, fileService)
	stuReportService := service.NewStudentReportService(reportRepo, projectRepo, usersRepo, fileService)
	courseImportService := service.NewCourseImportService(service.NewExcelSheetParser(), usersRepo, courseRepo, projectRepo)

	lgPageGen := html.NewLoginPageGenerator()
	stuHomeGen := html.NewStuHomeGenerator()
	tecHomeGen := html.NewTecHomeGenerator()

	sessionsCtl := controller.NewSessions(authService, cfg.JWTSecret)
	homeCtl := controller.NewHome(tecProjectService, stuProjectService, lgPageGen, stuHomeGen, tecHomeGen)
	offeringClassCtl := controller.NewOfferingClass(tecProjectService, tecHomeGen)
	projectsCtl := controller.NewProjects(tecProjectService, stuProjectService, tecReportService, stuReportService, tecHomeGen, stuHomeGen)
	submissionsCtl := controller.NewSubmissions(tecReportService)
	coursesCtl := controller.NewCourses(courseImportService)

	// 路由初始化
	mux := http.NewServeMux()
	router := server.NewRouter(mux, homeCtl, sessionsCtl, offeringClassCtl, projectsCtl, submissionsCtl, coursesCtl)
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
