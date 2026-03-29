// Package main은 JSON 설정 파일 파서 과제 구현 파일입니다.
// TODO 주석을 찾아 구현을 완성하세요.
package main

import "fmt"

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
	Timeout int    `json:"timeout"` // 초 단위
	TLS     bool   `json:"tls"`
}

// DatabaseConfig는 데이터베이스 연결 설정입니다.
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"` // JSON 출력 시 비어있으면 생략
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
// LoadConfig 함수
// ─────────────────────────────────────────

// LoadConfig는 JSON 파일을 읽어 Config 구조체로 파싱합니다.
// 파싱 후 기본값을 채우고, 환경변수로 오버라이드하고, 유효성을 검사합니다.
// TODO: 구현하세요.
func LoadConfig(filename string) (*Config, error) {
	// TODO 1: os.ReadFile로 파일을 읽으세요.
	//          파일이 없으면 에러를 반환하세요.

	// TODO 2: json.Unmarshal로 JSON을 Config 구조체로 파싱하세요.

	// TODO 3: setDefaults(&config)를 호출하여 기본값을 설정하세요.

	// TODO 4: applyEnvOverrides(&config)를 호출하여 환경변수를 적용하세요.

	// TODO 5: validate(&config)를 호출하여 유효성을 검사하세요.

	return nil, fmt.Errorf("구현되지 않았습니다")
}

// setDefaults는 비어있는 필드에 기본값을 설정합니다.
// TODO: 구현하세요.
func setDefaults(c *Config) {
	// Server 기본값:
	//   Host    → "localhost"
	//   Port    → 8080
	//   Timeout → 30

	// Database 기본값:
	//   Port     → 5432
	//   MaxConns → 10

	// App 기본값:
	//   Version → "1.0.0"
}

// applyEnvOverrides는 환경변수로 설정을 오버라이드합니다.
// TODO: 구현하세요.
func applyEnvOverrides(c *Config) {
	// APP_HOST  → c.Server.Host
	// APP_PORT  → c.Server.Port  (strconv.Atoi로 변환)
	// DB_HOST   → c.Database.Host
	// DB_PORT   → c.Database.Port (strconv.Atoi로 변환)
	// DB_PASSWORD → c.Database.Password
}

// validate는 설정값의 유효성을 검사합니다.
// TODO: 구현하세요.
func validate(c *Config) error {
	// Server.Port: 1~65535 범위 확인
	// Database.Name: 비어있으면 에러
	// App.Name: 비어있으면 에러
	// Database.MaxConns: 1 이상 확인
	return fmt.Errorf("구현되지 않았습니다")
}

func main() {
	fmt.Println("JSON 설정 파일 파서 과제")
	fmt.Println("config.go의 TODO를 구현하고 'go test -v'로 테스트하세요.")
}
