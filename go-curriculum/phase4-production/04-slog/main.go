// 04-slog/main.go
// Go 1.21+ log/slog 구조화된 로깅을 학습합니다.
// TextHandler, JSONHandler, 로그 레벨, 커스텀 핸들러, HTTP 미들웨어를 다룹니다.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// ============================================================
// 컨텍스트 키
// ============================================================

type ctxKey string

const loggerKey ctxKey = "logger"

// ============================================================
// 1. 기본 slog 사용법
// ============================================================

func demonstrateBasicSlog() {
	fmt.Println("=== 1. 기본 slog 사용법 ===")

	// 기본 로거 (텍스트 형식, os.Stderr 출력)
	slog.Info("기본 로거 사용")
	slog.Debug("이 메시지는 기본적으로 보이지 않습니다")

	// 구조화된 속성 추가
	slog.Info("사용자 로그인",
		slog.String("username", "김철수"),
		slog.Int("user_id", 42),
		slog.Bool("admin", false),
	)

	// Any를 사용하여 임의의 값 전달
	slog.Info("요청 처리",
		slog.Any("duration", 150*time.Millisecond),
		slog.Any("tags", []string{"web", "api"}),
	)
}

// ============================================================
// 2. TextHandler vs JSONHandler
// ============================================================

func demonstrateHandlers() {
	fmt.Println("\n=== 2. TextHandler vs JSONHandler ===")

	// TextHandler: 사람이 읽기 쉬운 텍스트 형식
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Debug 레벨부터 출력
	})
	textLogger := slog.New(textHandler)
	fmt.Println("--- TextHandler ---")
	textLogger.Info("텍스트 형식 로그",
		slog.String("format", "text"),
		slog.Int("version", 1),
	)
	textLogger.Debug("디버그 메시지", slog.String("detail", "세부 정보"))

	// JSONHandler: 기계가 처리하기 쉬운 JSON 형식 (프로덕션 환경)
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	jsonLogger := slog.New(jsonHandler)
	fmt.Println("\n--- JSONHandler ---")
	jsonLogger.Info("JSON 형식 로그",
		slog.String("format", "json"),
		slog.Int("version", 1),
	)
	jsonLogger.Warn("경고 메시지", slog.String("reason", "임계값 초과"))
}

// ============================================================
// 3. 로그 레벨
// ============================================================

func demonstrateLogLevels() {
	fmt.Println("\n=== 3. 로그 레벨 ===")

	// 동적으로 로그 레벨 변경 가능
	var level slog.LevelVar
	level.Set(slog.LevelDebug)

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &level,
	})
	logger := slog.New(handler)

	logger.Debug("DEBUG: 상세 디버그 정보")
	logger.Info("INFO: 일반 정보")
	logger.Warn("WARN: 경고 메시지")
	logger.Error("ERROR: 오류 발생", slog.String("error", "연결 실패"))

	// 런타임에 로그 레벨 변경
	fmt.Println("\n--- 레벨을 Warn으로 변경 ---")
	level.Set(slog.LevelWarn)
	logger.Debug("이 메시지는 출력되지 않습니다")
	logger.Info("이 메시지도 출력되지 않습니다")
	logger.Warn("이 경고는 출력됩니다")
}

// ============================================================
// 4. 그룹과 With를 사용한 구조화
// ============================================================

func demonstrateGroupsAndWith() {
	fmt.Println("\n=== 4. 그룹과 With ===")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// With: 모든 로그에 공통 필드 추가
	requestLogger := logger.With(
		slog.String("service", "user-service"),
		slog.String("version", "1.0.0"),
	)

	requestLogger.Info("서비스 시작")
	requestLogger.Info("요청 처리 완료", slog.Int("count", 100))

	// Group: 관련 필드를 그룹으로 묶기
	logger.Info("HTTP 요청",
		slog.Group("request",
			slog.String("method", "GET"),
			slog.String("path", "/api/users"),
			slog.Int("status", 200),
		),
		slog.Group("client",
			slog.String("ip", "192.168.1.1"),
			slog.String("user_agent", "Go-http-client/1.1"),
		),
	)
}

// ============================================================
// 5. 커스텀 핸들러
// 특수한 형식이나 동작이 필요할 때 커스텀 핸들러를 구현합니다.
// ============================================================

// PrettyHandler는 개발 환경에서 컬러 출력을 제공하는 핸들러입니다.
type PrettyHandler struct {
	slog.Handler
	out io.Writer
}

// NewPrettyHandler는 새 PrettyHandler를 생성합니다.
func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	return &PrettyHandler{
		Handler: slog.NewTextHandler(out, opts),
		out:     out,
	}
}

// Handle은 로그 레코드를 처리합니다.
func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	// ANSI 색상 코드
	const (
		reset  = "\033[0m"
		red    = "\033[31m"
		yellow = "\033[33m"
		blue   = "\033[34m"
		gray   = "\033[90m"
	)

	// 레벨에 따른 색상 선택
	levelColor := blue
	switch r.Level {
	case slog.LevelDebug:
		levelColor = gray
	case slog.LevelWarn:
		levelColor = yellow
	case slog.LevelError:
		levelColor = red
	}

	// 시간과 레벨 출력
	fmt.Fprintf(h.out, "%s[%s]%s %s%s%s: %s",
		gray, r.Time.Format("15:04:05"), reset,
		levelColor, r.Level.String(), reset,
		r.Message,
	)

	// 속성 출력
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(h.out, " %s%s=%v%s", gray, a.Key, a.Value, reset)
		return true
	})
	fmt.Fprintln(h.out)

	return nil
}

func demonstrateCustomHandler() {
	fmt.Println("\n=== 5. 커스텀 핸들러 ===")

	prettyLogger := slog.New(NewPrettyHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	prettyLogger.Debug("디버그", slog.String("key", "value"))
	prettyLogger.Info("정보", slog.Int("count", 42))
	prettyLogger.Warn("경고", slog.String("reason", "임계값 초과"))
	prettyLogger.Error("오류", slog.String("error", "연결 실패"))
}

// ============================================================
// 6. 컨텍스트를 통한 로거 전달
// ============================================================

// LoggerFromContext는 컨텍스트에서 로거를 추출합니다.
// 없으면 기본 로거를 반환합니다.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithLogger는 컨텍스트에 로거를 저장합니다.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// processOrder는 컨텍스트에서 로거를 꺼내 사용하는 비즈니스 로직 예제입니다.
func processOrder(ctx context.Context, orderID string) error {
	logger := LoggerFromContext(ctx)
	logger.Info("주문 처리 시작", slog.String("order_id", orderID))

	// 비즈니스 로직 시뮬레이션
	time.Sleep(10 * time.Millisecond)

	logger.Info("주문 처리 완료",
		slog.String("order_id", orderID),
		slog.String("status", "completed"),
	)
	return nil
}

func demonstrateContextLogger() {
	fmt.Println("\n=== 6. 컨텍스트를 통한 로거 전달 ===")

	// 요청별 고유 속성을 가진 로거 생성
	requestLogger := slog.Default().With(
		slog.String("request_id", "req-12345"),
		slog.String("user_id", "user-67890"),
	)

	ctx := WithLogger(context.Background(), requestLogger)

	// 여러 함수에 걸쳐 동일한 요청 컨텍스트 정보가 포함된 로그 출력
	processOrder(ctx, "order-111")
}

// ============================================================
// 7. HTTP 미들웨어에서의 구조화된 로깅
// ============================================================

// SlogMiddleware는 slog를 사용하는 HTTP 로깅 미들웨어입니다.
func SlogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 요청별 로거에 공통 필드 추가
			reqLogger := logger.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			// 컨텍스트에 로거 저장 (핸들러에서 사용 가능)
			ctx := WithLogger(r.Context(), reqLogger)

			// 응답 상태 코드 캡처
			rw := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

			reqLogger.Debug("요청 수신")
			next.ServeHTTP(rw, r.WithContext(ctx))

			// 요청 완료 로그
			reqLogger.Info("요청 처리 완료",
				slog.Int("status", rw.statusCode),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func demonstrateHTTPLogging() {
	fmt.Println("\n=== 7. HTTP 미들웨어 로깅 (시뮬레이션) ===")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// 핸들러
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 핸들러에서 컨텍스트 로거 사용
		log := LoggerFromContext(r.Context())
		log.Info("핸들러 실행 중")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// 미들웨어 적용
	middleware := SlogMiddleware(logger)
	wrappedHandler := middleware(handler)

	// 가상의 요청 생성 및 처리 시뮬레이션
	req, _ := http.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:54321"

	rw := &mockResponseWriter{header: make(http.Header)}
	wrappedHandler.ServeHTTP(rw, req)
}

// mockResponseWriter는 테스트용 간단한 ResponseWriter입니다.
type mockResponseWriter struct {
	header     http.Header
	statusCode int
	body       []byte
}

func (m *mockResponseWriter) Header() http.Header        { return m.header }
func (m *mockResponseWriter) Write(b []byte) (int, error) { m.body = append(m.body, b...); return len(b), nil }
func (m *mockResponseWriter) WriteHeader(code int)        { m.statusCode = code }

// ============================================================
// 메인 함수
// ============================================================

func main() {
	demonstrateBasicSlog()
	demonstrateHandlers()
	demonstrateLogLevels()
	demonstrateGroupsAndWith()
	demonstrateCustomHandler()
	demonstrateContextLogger()
	demonstrateHTTPLogging()

	fmt.Println("\n=== slog 학습 완료 ===")
}
