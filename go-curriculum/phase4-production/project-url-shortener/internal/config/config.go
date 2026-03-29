// internal/config/config.go
// 환경 변수 기반 설정 관리 패키지입니다.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config는 애플리케이션 전체 설정을 담습니다.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
}

// ServerConfig는 HTTP 서버 설정입니다.
type ServerConfig struct {
	// 서버가 바인딩할 주소 (기본: :8080)
	Addr string
	// 읽기 타임아웃 (기본: 10s)
	ReadTimeout time.Duration
	// 쓰기 타임아웃 (기본: 10s)
	WriteTimeout time.Duration
	// 유휴 타임아웃 (기본: 60s)
	IdleTimeout time.Duration
}

// DatabaseConfig는 PostgreSQL 연결 설정입니다.
type DatabaseConfig struct {
	// 연결 URL (예: postgres://user:pass@localhost:5432/dbname)
	URL string
	// 최대 연결 수 (기본: 10)
	MaxConns int
	// 최소 연결 수 (기본: 2)
	MinConns int
	// 연결 최대 생존 시간 (기본: 1h)
	MaxConnLifetime time.Duration
}

// AppConfig는 애플리케이션별 설정입니다.
type AppConfig struct {
	// 단축 URL 기본 도메인 (예: http://localhost:8080)
	BaseURL string
	// 단축 코드 길이 (기본: 6)
	ShortCodeLength int
	// 요청당 최대 처리량 (기본: 100)
	RateLimit int
	// 로그 레벨 (debug, info, warn, error)
	LogLevel string
	// 환경 (development, production)
	Environment string
}

// Load는 환경 변수에서 설정을 읽어 Config를 반환합니다.
// 필수 환경 변수가 없는 경우 기본값을 사용합니다.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Addr:         getEnv("SERVER_ADDR", ":8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable"),
			MaxConns:        getIntEnv("DB_MAX_CONNS", 10),
			MinConns:        getIntEnv("DB_MIN_CONNS", 2),
			MaxConnLifetime: getDurationEnv("DB_MAX_CONN_LIFETIME", time.Hour),
		},
		App: AppConfig{
			BaseURL:         getEnv("BASE_URL", "http://localhost:8080"),
			ShortCodeLength: getIntEnv("SHORT_CODE_LENGTH", 6),
			RateLimit:       getIntEnv("RATE_LIMIT", 100),
			LogLevel:        getEnv("LOG_LEVEL", "info"),
			Environment:     getEnv("ENVIRONMENT", "development"),
		},
	}

	// 필수 설정 검증
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("설정 검증 실패: %w", err)
	}

	return cfg, nil
}

// validate는 설정 값의 유효성을 검사합니다.
func (c *Config) validate() error {
	if c.App.ShortCodeLength < 4 || c.App.ShortCodeLength > 20 {
		return fmt.Errorf("SHORT_CODE_LENGTH는 4~20 사이여야 합니다 (현재: %d)", c.App.ShortCodeLength)
	}
	if c.Database.MaxConns < c.Database.MinConns {
		return fmt.Errorf("DB_MAX_CONNS(%d)는 DB_MIN_CONNS(%d)보다 커야 합니다",
			c.Database.MaxConns, c.Database.MinConns)
	}
	return nil
}

// IsDevelopment는 개발 환경인지 확인합니다.
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction은 프로덕션 환경인지 확인합니다.
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// ============================================================
// 헬퍼 함수
// ============================================================

// getEnv는 환경 변수를 읽거나 기본값을 반환합니다.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// getIntEnv는 정수형 환경 변수를 읽거나 기본값을 반환합니다.
func getIntEnv(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}

// getDurationEnv는 시간 환경 변수를 읽거나 기본값을 반환합니다.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
