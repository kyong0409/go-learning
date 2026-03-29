// cmd/server/main.go
// URL 단축 서비스의 진입점입니다.
// 설정 로딩, 서버 초기화, Graceful Shutdown을 처리합니다.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/learn-go/url-shortener/internal/config"
	"github.com/learn-go/url-shortener/internal/handler"
	"github.com/learn-go/url-shortener/internal/store"
)

func main() {
	// 1. 설정 로딩
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "설정 로딩 실패: %v\n", err)
		os.Exit(1)
	}

	// 2. 로거 초기화
	logger := newLogger(cfg)
	slog.SetDefault(logger)

	logger.Info("URL 단축 서비스 시작",
		slog.String("addr", cfg.Server.Addr),
		slog.String("env", cfg.App.Environment),
		slog.String("base_url", cfg.App.BaseURL),
	)

	// 3. 저장소 초기화
	// DATABASE_URL이 설정된 경우 PostgreSQL을 사용하고, 없으면 메모리 저장소를 사용합니다.
	var s store.Store
	if cfg.Database.URL != "" && cfg.Database.URL != "postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable" {
		logger.Info("PostgreSQL 저장소 사용", slog.String("url", maskPassword(cfg.Database.URL)))
		// 실제 pgx 연결은 여기서 초기화합니다.
		// 이 스켈레톤에서는 메모리 저장소로 대체합니다.
		s = store.NewMemoryStore()
	} else {
		logger.Info("인메모리 저장소 사용 (개발/테스트 모드)")
		s = store.NewMemoryStore()
	}
	defer s.Close()

	// 4. 핸들러 생성
	h := handler.New(s, cfg.App.BaseURL, logger, cfg.App.ShortCodeLength)

	// 5. 라우터 설정
	r := chi.NewRouter()

	// 기본 chi 미들웨어
	r.Use(chimiddleware.RealIP)        // 실제 클라이언트 IP 추출
	r.Use(handler.RequestID)           // 요청 ID 생성
	r.Use(handler.Logger(logger))      // 구조화된 로깅
	r.Use(handler.Recoverer(logger))   // 패닉 복구
	r.Use(handler.CORS)                // CORS 헤더

	// 속도 제한 (초당 요청 수, 버스트 크기)
	limiter := handler.NewRateLimiter(float64(cfg.App.RateLimit)/60, float64(cfg.App.RateLimit))
	r.Use(limiter.Middleware)

	// 핸들러 등록
	h.RegisterRoutes(r)

	// 6. HTTP 서버 설정
	server := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 7. Graceful Shutdown 설정
	// OS 시그널(SIGINT, SIGTERM)을 받으면 진행 중인 요청을 완료 후 종료합니다.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 서버를 별도 고루틴에서 시작
	go func() {
		logger.Info("HTTP 서버 리스닝", slog.String("addr", cfg.Server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("서버 오류", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// 시그널 대기
	sig := <-quit
	logger.Info("종료 시그널 수신", slog.String("signal", sig.String()))

	// Graceful Shutdown: 최대 30초 동안 기존 요청 완료를 기다립니다.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Graceful Shutdown 실패", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("서버가 정상적으로 종료되었습니다")
}

// newLogger는 환경에 맞는 로거를 생성합니다.
func newLogger(cfg *config.Config) *slog.Logger {
	// 로그 레벨 파싱
	var level slog.Level
	switch cfg.App.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.IsProduction() {
		// 프로덕션: JSON 형식 (로그 수집 시스템 호환)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// 개발: 텍스트 형식 (사람이 읽기 쉬운)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// maskPassword는 데이터베이스 URL에서 비밀번호를 마스킹합니다.
func maskPassword(url string) string {
	// postgres://user:password@host/db -> postgres://user:***@host/db
	for i, c := range url {
		if c == ':' {
			// 두 번째 ':' 이후부터 '@'까지를 마스킹
			rest := url[i+1:]
			for j, rc := range rest {
				if rc == '@' {
					return url[:i+1] + "***" + rest[j:]
				}
			}
		}
	}
	return url
}
