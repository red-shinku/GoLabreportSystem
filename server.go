package main

import (
	"encoding/json"
	"errors"
	"github.com/caarlos0/env/v11"
	"os"
	"strings"
)

type Config struct {
	Ipaddr    string `json:"ipaddr" env:"IP_ADDR"`
	Port      string `json:"port" env:"PORT"`
	EnableTLS bool   `json:"enable_tls" env:"ENABLE_TLS"`

	// secret
	JWTSecret     string `env:"JWT_SECRET"`
	JWTSecretFile string `json:"jwt_secret_file"`
}

func LoadJSONConfig(path string) Config {
	var cfg Config

	b, err := os.ReadFile(path)
	if err != nil {
		// 文件不存在可以允许（看你需求）
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

func main() {

}
