// config_test.go: JSON 설정 파일 파서 과제 채점 테스트
package main

import (
	"fmt"
	"os"
	"testing"
)

// ─────────────────────────────────────────
// 채점 시스템
// ─────────────────────────────────────────

type scorer struct {
	total  int
	passed int
}

func newScorer() *scorer { return &scorer{} }

func (s *scorer) check(t *testing.T, name string, cond bool, msg string) {
	t.Helper()
	s.total++
	if cond {
		s.passed++
	} else {
		t.Errorf("  FAIL [%s]: %s", name, msg)
	}
}

func (s *scorer) report(t *testing.T) {
	score := 0
	if s.total > 0 {
		score = s.passed * 100 / s.total
	}
	fmt.Printf("\n=== 채점 결과 ===\n")
	fmt.Printf("통과: %d/%d\n", s.passed, s.total)
	fmt.Printf("점수: %d/100\n", score)
}

// ─────────────────────────────────────────
// 유효한 JSON 파일 테스트
// ─────────────────────────────────────────

func TestLoadValidConfig(t *testing.T) {
	sc := newScorer()

	cfg, err := LoadConfig("testdata/valid.json")
	sc.check(t, "에러 없음", err == nil, fmt.Sprintf("에러 발생: %v", err))
	sc.check(t, "nil 아님", cfg != nil, "Config가 nil입니다")

	if cfg == nil {
		sc.report(t)
		return
	}

	// Server 설정
	sc.check(t, "Server.Host", cfg.Server.Host == "example.com", fmt.Sprintf("got=%q, want=%q", cfg.Server.Host, "example.com"))
	sc.check(t, "Server.Port", cfg.Server.Port == 9090, fmt.Sprintf("got=%d, want=9090", cfg.Server.Port))
	sc.check(t, "Server.Timeout", cfg.Server.Timeout == 60, fmt.Sprintf("got=%d, want=60", cfg.Server.Timeout))
	sc.check(t, "Server.TLS", cfg.Server.TLS == true, "TLS가 true여야 합니다")

	// Database 설정
	sc.check(t, "DB.Host", cfg.Database.Host == "db.example.com", fmt.Sprintf("got=%q", cfg.Database.Host))
	sc.check(t, "DB.Port", cfg.Database.Port == 5432, fmt.Sprintf("got=%d", cfg.Database.Port))
	sc.check(t, "DB.Name", cfg.Database.Name == "myapp", fmt.Sprintf("got=%q", cfg.Database.Name))
	sc.check(t, "DB.MaxConns", cfg.Database.MaxConns == 20, fmt.Sprintf("got=%d", cfg.Database.MaxConns))

	// App 설정
	sc.check(t, "App.Name", cfg.App.Name == "MyApp", fmt.Sprintf("got=%q", cfg.App.Name))
	sc.check(t, "App.Version", cfg.App.Version == "2.0.0", fmt.Sprintf("got=%q", cfg.App.Version))
	sc.check(t, "App.Debug", cfg.App.Debug == false, "Debug가 false여야 합니다")

	sc.report(t)
}

// ─────────────────────────────────────────
// 기본값 설정 테스트
// ─────────────────────────────────────────

func TestDefaultValues(t *testing.T) {
	sc := newScorer()

	cfg, err := LoadConfig("testdata/partial.json")
	sc.check(t, "에러 없음", err == nil, fmt.Sprintf("에러 발생: %v", err))

	if cfg == nil {
		sc.report(t)
		return
	}

	// 기본값 확인
	sc.check(t, "Server.Host 기본값", cfg.Server.Host == "localhost",
		fmt.Sprintf("got=%q, want=%q", cfg.Server.Host, "localhost"))
	sc.check(t, "Server.Port 기본값", cfg.Server.Port == 8080,
		fmt.Sprintf("got=%d, want=8080", cfg.Server.Port))
	sc.check(t, "Server.Timeout 기본값", cfg.Server.Timeout == 30,
		fmt.Sprintf("got=%d, want=30", cfg.Server.Timeout))
	sc.check(t, "DB.Port 기본값", cfg.Database.Port == 5432,
		fmt.Sprintf("got=%d, want=5432", cfg.Database.Port))
	sc.check(t, "DB.MaxConns 기본값", cfg.Database.MaxConns == 10,
		fmt.Sprintf("got=%d, want=10", cfg.Database.MaxConns))
	sc.check(t, "App.Version 기본값", cfg.App.Version == "1.0.0",
		fmt.Sprintf("got=%q, want=%q", cfg.App.Version, "1.0.0"))

	sc.report(t)
}

// ─────────────────────────────────────────
// 환경변수 오버라이드 테스트
// ─────────────────────────────────────────

func TestEnvOverrides(t *testing.T) {
	sc := newScorer()

	// 환경변수 설정
	os.Setenv("APP_HOST", "env-host.com")
	os.Setenv("APP_PORT", "3000")
	os.Setenv("DB_HOST", "env-db.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_PASSWORD", "secret123")

	// 테스트 후 환경변수 정리
	defer func() {
		os.Unsetenv("APP_HOST")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_PASSWORD")
	}()

	cfg, err := LoadConfig("testdata/partial.json")
	sc.check(t, "에러 없음", err == nil, fmt.Sprintf("에러 발생: %v", err))

	if cfg == nil {
		sc.report(t)
		return
	}

	sc.check(t, "APP_HOST 오버라이드", cfg.Server.Host == "env-host.com",
		fmt.Sprintf("got=%q, want=%q", cfg.Server.Host, "env-host.com"))
	sc.check(t, "APP_PORT 오버라이드", cfg.Server.Port == 3000,
		fmt.Sprintf("got=%d, want=3000", cfg.Server.Port))
	sc.check(t, "DB_HOST 오버라이드", cfg.Database.Host == "env-db.com",
		fmt.Sprintf("got=%q, want=%q", cfg.Database.Host, "env-db.com"))
	sc.check(t, "DB_PORT 오버라이드", cfg.Database.Port == 5433,
		fmt.Sprintf("got=%d, want=5433", cfg.Database.Port))
	sc.check(t, "DB_PASSWORD 오버라이드", cfg.Database.Password == "secret123",
		fmt.Sprintf("got=%q, want=%q", cfg.Database.Password, "secret123"))

	sc.report(t)
}

// ─────────────────────────────────────────
// 유효성 검사 테스트
// ─────────────────────────────────────────

func TestValidation(t *testing.T) {
	sc := newScorer()

	// 잘못된 JSON 파일
	_, err := LoadConfig("testdata/invalid.json")
	sc.check(t, "유효하지 않은 JSON 에러", err != nil, "에러가 없습니다")

	// 존재하지 않는 파일
	_, err = LoadConfig("testdata/nonexistent.json")
	sc.check(t, "파일 없음 에러", err != nil, "에러가 없습니다")

	sc.report(t)
}

// ─────────────────────────────────────────
// 종합 채점
// ─────────────────────────────────────────

func TestFinalScore(t *testing.T) {
	sc := newScorer()

	// 1. 유효한 파일 로드
	cfg, err := LoadConfig("testdata/valid.json")
	sc.check(t, "valid.json 로드", err == nil && cfg != nil,
		fmt.Sprintf("err=%v, cfg=%v", err, cfg))

	if cfg != nil {
		sc.check(t, "Server 설정", cfg.Server.Port == 9090, "Server.Port")
		sc.check(t, "DB 설정", cfg.Database.Name == "myapp", "Database.Name")
		sc.check(t, "App 설정", cfg.App.Name == "MyApp", "App.Name")
	} else {
		sc.total += 3
	}

	// 2. 기본값
	cfg2, _ := LoadConfig("testdata/partial.json")
	if cfg2 != nil {
		sc.check(t, "기본값 Host", cfg2.Server.Host == "localhost", "Server.Host 기본값")
		sc.check(t, "기본값 Port", cfg2.Server.Port == 8080, "Server.Port 기본값")
		sc.check(t, "기본값 Timeout", cfg2.Server.Timeout == 30, "Server.Timeout 기본값")
		sc.check(t, "기본값 DBPort", cfg2.Database.Port == 5432, "DB.Port 기본값")
		sc.check(t, "기본값 MaxConns", cfg2.Database.MaxConns == 10, "DB.MaxConns 기본값")
		sc.check(t, "기본값 Version", cfg2.App.Version == "1.0.0", "App.Version 기본값")
	} else {
		sc.total += 6
	}

	// 3. 환경변수 오버라이드
	os.Setenv("APP_HOST", "override.com")
	os.Setenv("APP_PORT", "9999")
	defer os.Unsetenv("APP_HOST")
	defer os.Unsetenv("APP_PORT")

	cfg3, _ := LoadConfig("testdata/partial.json")
	if cfg3 != nil {
		sc.check(t, "환경변수 Host", cfg3.Server.Host == "override.com", "APP_HOST 오버라이드")
		sc.check(t, "환경변수 Port", cfg3.Server.Port == 9999, "APP_PORT 오버라이드")
	} else {
		sc.total += 2
	}

	// 4. 에러 처리
	_, errInvalid := LoadConfig("testdata/invalid.json")
	sc.check(t, "invalid.json 에러", errInvalid != nil, "유효하지 않은 JSON 에러 없음")

	_, errMissing := LoadConfig("testdata/nonexistent.json")
	sc.check(t, "없는 파일 에러", errMissing != nil, "파일 없음 에러 없음")

	os.Unsetenv("APP_HOST")
	os.Unsetenv("APP_PORT")

	sc.report(t)
}
