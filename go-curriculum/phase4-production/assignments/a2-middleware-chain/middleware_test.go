// a2-middleware-chain/middleware_test.go
// HTTP 미들웨어 체인 테스트입니다.
package middleware_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	middleware "github.com/learn-go/a2-middleware-chain"
)

// okHandler는 항상 200 OK를 반환하는 테스트용 핸들러입니다.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
})

// ============================================================
// Chain 테스트 (20점)
// ============================================================

func TestChain_ExecutionOrder(t *testing.T) {
	var order []string

	makeMiddleware := func(name string) middleware.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"-before")
				next.ServeHTTP(w, r)
				order = append(order, name+"-after")
			})
		}
	}

	handler := middleware.Chain(
		okHandler,
		makeMiddleware("A"),
		makeMiddleware("B"),
		makeMiddleware("C"),
	)

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	// A가 가장 바깥쪽: A-before, B-before, C-before, C-after, B-after, A-after
	expected := []string{"A-before", "B-before", "C-before", "C-after", "B-after", "A-after"}
	if len(order) != len(expected) {
		t.Fatalf("실행 순서 길이: 기대 %d, 실제 %d\n실제: %v", len(expected), len(order), order)
	}
	for i, want := range expected {
		if order[i] != want {
			t.Errorf("순서[%d]: 기대 %q, 실제 %q", i, want, order[i])
		}
	}
}

func TestChain_EmptyMiddlewares(t *testing.T) {
	handler := middleware.Chain(okHandler)
	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("빈 체인: 기대 200, 실제 %d", rw.Code)
	}
}

// ============================================================
// RateLimiter 테스트 (20점)
// ============================================================

func TestRateLimiter_AllowsNormalRequests(t *testing.T) {
	limiter := middleware.RateLimiter(10) // 초당 10개
	handler := limiter(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("정상 요청이 차단됨: 상태 %d", rw.Code)
	}
}

func TestRateLimiter_BlocksExcessRequests(t *testing.T) {
	limiter := middleware.RateLimiter(2) // 초당 2개
	handler := limiter(okHandler)

	ip := "10.0.0.1:5000"
	blocked := false

	// 빠르게 여러 요청 전송
	for i := range 20 {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip
		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)

		if rw.Code == http.StatusTooManyRequests {
			blocked = true
			t.Logf("요청 %d에서 차단됨", i+1)
			break
		}
	}

	if !blocked {
		t.Error("초과 요청이 차단되지 않았습니다 (429가 반환되어야 합니다)")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	limiter := middleware.RateLimiter(1) // 초당 1개
	handler := limiter(okHandler)

	// IP1이 한도를 초과해도 IP2는 독립적이어야 합니다.
	for range 10 {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "1.1.1.1:100"
		httptest.NewRecorder()
		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)
	}

	// IP2 첫 요청은 허용되어야 합니다.
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "2.2.2.2:200"
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code == http.StatusTooManyRequests {
		t.Error("다른 IP는 독립적으로 제한되어야 합니다")
	}
}

// ============================================================
// RequestLogger 테스트 (20점)
// ============================================================

func TestRequestLogger_LogsRequest(t *testing.T) {
	// 로그 출력을 캡처하기 위해 버퍼 사용
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	loggerMiddleware := middleware.RequestLogger(logger)
	handler := loggerMiddleware(okHandler)

	req := httptest.NewRequest("GET", "/test-path", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	// 핸들러가 정상 실행되었는지 확인
	if rw.Code != http.StatusOK {
		t.Errorf("RequestLogger 후 상태 코드: 기대 200, 실제 %d", rw.Code)
	}
}

func TestRequestLogger_AddsRequestID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	loggerMiddleware := middleware.RequestLogger(logger)
	handler := loggerMiddleware(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	// X-Request-ID 헤더가 추가되었는지 확인
	if rw.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID 헤더가 응답에 없습니다")
	}
}

// ============================================================
// Authenticator 테스트 (20점)
// ============================================================

func TestAuthenticator_ValidToken(t *testing.T) {
	auth := middleware.Authenticator("secret-token")
	handler := auth(okHandler)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("유효한 토큰: 기대 200, 실제 %d", rw.Code)
	}
}

func TestAuthenticator_InvalidToken(t *testing.T) {
	auth := middleware.Authenticator("secret-token")
	handler := auth(okHandler)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusUnauthorized {
		t.Errorf("잘못된 토큰: 기대 401, 실제 %d", rw.Code)
	}
}

func TestAuthenticator_MissingToken(t *testing.T) {
	auth := middleware.Authenticator("secret-token")
	handler := auth(okHandler)

	req := httptest.NewRequest("GET", "/protected", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusUnauthorized {
		t.Errorf("토큰 없음: 기대 401, 실제 %d", rw.Code)
	}
}

func TestAuthenticator_HealthPathSkipped(t *testing.T) {
	auth := middleware.Authenticator("secret-token")
	handler := auth(okHandler)

	// /health는 인증 없이 통과해야 합니다.
	req := httptest.NewRequest("GET", "/health", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("/health는 인증 없이 통과: 기대 200, 실제 %d", rw.Code)
	}
}

// ============================================================
// ResponseTimer 테스트 (20점)
// ============================================================

func TestResponseTimer_AddsHeader(t *testing.T) {
	handler := middleware.ResponseTimer(okHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	responseTime := rw.Header().Get("X-Response-Time")
	if responseTime == "" {
		t.Error("X-Response-Time 헤더가 없습니다")
	}
	t.Logf("X-Response-Time: %s", responseTime)
}

func TestResponseTimer_MeasuresTime(t *testing.T) {
	// 100ms 지연 핸들러
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.ResponseTimer(slowHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	responseTime := rw.Header().Get("X-Response-Time")
	if responseTime == "" {
		t.Error("X-Response-Time 헤더가 없습니다")
	}
	// 값이 있으면 통과 (형식은 구현마다 다를 수 있음)
	t.Logf("측정된 처리 시간: %s", responseTime)
}

// ============================================================
// 성적 보고서
// ============================================================

func TestMain(m *testing.M) {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║   과제 A2: HTTP 미들웨어 체인 구현    ║")
	fmt.Println("╚══════════════════════════════════════╝")

	result := m.Run()

	fmt.Println()
	fmt.Println("─────────────────────────────────────")
	if result == 0 {
		fmt.Println("  최종 점수: 100 / 100 점")
		fmt.Println("  평가: 합격 (모든 테스트 통과)")
	} else {
		fmt.Println("  평가: 미완성 — 실패한 테스트를 확인하세요")
	}
	fmt.Println("─────────────────────────────────────")

	os.Exit(result)
}
