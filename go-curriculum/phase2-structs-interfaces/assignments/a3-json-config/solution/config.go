// Package main은 JSON 설정 파일 파서 과제의 참고 풀이입니다.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// ─────────────────────────────────────────
// 설정 구조체
// ─────────────────────────────────────────

// Config는 전체 애플리케이션 설정을 나타냅니다.
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	App      AppConfig      `json:"app"`
}

// ServerConfig는 HTTP 서버 설정입니다.
type ServerConfig struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Timeout int    `json:"timeout"`
	TLS     bool   `json:"tls"`
}

// DatabaseConfig는 데이터베이스 연결 설정입니다.
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	MaxConns int    `json:"max_connections"`
}

// AppConfig는 애플리케이션 일반 설정입니다.
type AppConfig struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Debug        bool     `json:"debug"`
	AllowedHosts []string `json:"allowed_hosts,omitempty"`
}

// ─────────────────────────────────────────
// LoadConfig
// ─────────────────────────────────────────

// LoadConfig는 JSON 파일을 읽어 Config 구조체로 파싱합니다.
func LoadConfig(filename string) (*Config, error) {
	// 1. 파일 읽기
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("설정 파일 읽기 실패 %q: %w", filename, err)
	}

	// 2. JSON 파싱
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
	}

	// 3. 기본값 설정
	setDefaults(&cfg)

	// 4. 환경변수 오버라이드
	applyEnvOverrides(&cfg)

	// 5. 유효성 검사
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("설정 유효성 검사 실패: %w", err)
	}

	return &cfg, nil
}

// setDefaults는 비어있는 필드에 기본값을 설정합니다.
func setDefaults(c *Config) {
	// Server 기본값
	if c.Server.Host == "" {
		c.Server.Host = "localhost"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Timeout == 0 {
		c.Server.Timeout = 30
	}

	// Database 기본값
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Database.MaxConns == 0 {
		c.Database.MaxConns = 10
	}

	// App 기본값
	if c.App.Version == "" {
		c.App.Version = "1.0.0"
	}
}

// applyEnvOverrides는 환경변수로 설정을 오버라이드합니다.
func applyEnvOverrides(c *Config) {
	if v := os.Getenv("APP_HOST"); v != "" {
		c.Server.Host = v
	}
	if v := os.Getenv("APP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Server.Port = port
		}
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Database.Port = port
		}
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
}

// validate는 설정값의 유효성을 검사합니다.
func validate(c *Config) error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("Server.Port 범위 오류: %d (1~65535)", c.Server.Port)
	}
	if c.Database.Name == "" {
		return fmt.Errorf("Database.Name은 필수입니다")
	}
	if c.App.Name == "" {
		return fmt.Errorf("App.Name은 필수입니다")
	}
	if c.Database.MaxConns < 1 {
		return fmt.Errorf("Database.MaxConns는 1 이상이어야 합니다: %d", c.Database.MaxConns)
	}
	return nil
}

func main() {
	fmt.Println("=== JSON 설정 파일 파서 참고 풀이 ===")

	cfg, err := LoadConfig("testdata/valid.json")
	if err != nil {
		fmt.Printf("에러: %v\n", err)
		return
	}

	fmt.Printf("서버: %s:%d (TLS=%v, Timeout=%ds)\n",
		cfg.Server.Host, cfg.Server.Port, cfg.Server.TLS, cfg.Server.Timeout)
	fmt.Printf("DB: %s@%s:%d/%s (MaxConns=%d)\n",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port,
		cfg.Database.Name, cfg.Database.MaxConns)
	fmt.Printf("앱: %s v%s (Debug=%v)\n",
		cfg.App.Name, cfg.App.Version, cfg.App.Debug)
}
