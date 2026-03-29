// 02-middleware/main.go
// HTTP 미들웨어 패턴을 학습합니다.
// 미들웨어 체이닝, 로깅, 복구, CORS, 요청 ID 전파를 다룹니다.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime/debug"
	"time"
)

// ============================================================
// 컨텍스트 키 타입
// 컨텍스트에 값을 저장할 때 충돌을 피하기 위해 전용 타입을 사용합니다.
// ============================================================

type contextKey string

const (
	// requestIDKey는 요청 ID를 컨텍스트에 저장하는 키입니다.
	requestIDKey contextKey = "request_id"
	// startTimeKey는 요청 시작 시간을 저장하는 키입니다.
	startTimeKey contextKey = "start_time"
)

// ============================================================
// 미들웨어 타입 정의
// 미들웨어는 http.Handler를 받아 http.Handler를 반환하는 함수입니다.
// ============================================================

// Middleware는 미들웨어 함수의 타입입니다.
type Middleware func(http.Handler) http.Handler

// Chain은 여러 미들웨어를 순서대로 체이닝합니다.
// 사용법: handler = Chain(handler, logging, recovery, cors)
// 실행 순서: cors -> recovery -> logging -> handler
// (마지막 미들웨어가 가장 바깥쪽에서 실행됩니다)
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// 역순으로 적용하여 첫 번째 미들웨어가 가장 먼저 실행되도록 합니다.
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// ============================================================
// 미들웨어 1: 요청 ID 생성 및 전파
// 각 요청에 고유한 ID를 부여하고 컨텍스트와 응답 헤더에 전달합니다.
// ============================================================

// RequestID는 각 요청에 고유한 ID를 할당하는 미들웨어입니다.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 요청 헤더에서 기존 ID를 확인합니다 (프록시 환경에서 유용).
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// 새 요청 ID를 생성합니다.
			requestID = generateRequestID()
		}

		// 컨텍스트에 요청 ID를 저장합니다.
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		// 응답 헤더에 요청 ID를 추가합니다.
		w.Header().Set("X-Request-ID", requestID)

		// 업데이트된 컨텍스트로 다음 핸들러를 호출합니다.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID는 컨텍스트에서 요청 ID를 추출합니다.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// generateRequestID는 간단한 랜덤 요청 ID를 생성합니다.
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))
}

// ============================================================
// 미들웨어 2: 로깅
// 요청/응답 정보를 기록합니다.
// ============================================================

// responseWriter는 상태 코드를 캡처하는 ResponseWriter 래퍼입니다.
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Logging은 HTTP 요청과 응답을 로깅하는 미들웨어입니다.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 컨텍스트에 시작 시간 저장
		ctx := context.WithValue(r.Context(), startTimeKey, start)
		r = r.WithContext(ctx)

		// 응답 코드를 캡처하기 위해 래퍼 사용
		wrapped := newResponseWriter(w)

		// 요청 처리
		next.ServeHTTP(wrapped, r)

		// 요청 완료 후 로깅
		duration := time.Since(start)
		requestID := GetRequestID(r.Context())

		log.Printf("[%s] %s %s %d %v",
			requestID,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// ============================================================
// 미들웨어 3: 패닉 복구 (Recovery)
// 핸들러에서 발생한 패닉을 잡아 500 오류로 변환합니다.
// ============================================================

// Recovery는 핸들러의 패닉을 복구하는 미들웨어입니다.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())

				// 스택 트레이스와 함께 오류를 로깅합니다.
				log.Printf("[%s] 패닉 복구: %v\n스택 트레이스:\n%s",
					requestID, err, debug.Stack())

				// 클라이언트에게 500 오류를 반환합니다.
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{
					"error":      "내부 서버 오류",
					"request_id": requestID,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ============================================================
// 미들웨어 4: CORS
// Cross-Origin Resource Sharing 헤더를 설정합니다.
// ============================================================

// CORSConfig는 CORS 미들웨어 설정입니다.
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// DefaultCORSConfig는 기본 CORS 설정을 반환합니다.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		MaxAge:         86400, // 24시간
	}
}

// CORS는 CORS 헤더를 설정하는 미들웨어입니다.
func CORS(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 허용된 출처 확인
			allowed := false
			for _, o := range config.AllowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				// CORS 헤더 설정
				w.Header().Set("Access-Control-Allow-Origin",
					joinStrings(config.AllowedOrigins, ", "))
				w.Header().Set("Access-Control-Allow-Methods",
					joinStrings(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers",
					joinStrings(config.AllowedHeaders, ", "))
				w.Header().Set("Access-Control-Max-Age",
					fmt.Sprintf("%d", config.MaxAge))
			}

			// Preflight 요청(OPTIONS) 처리
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// ============================================================
// 미들웨어 5: 요청 타임아웃
// 지정된 시간 내에 처리되지 않은 요청을 중단합니다.
// ============================================================

// Timeout은 요청 처리 시간을 제한하는 미들웨어입니다.
func Timeout(duration time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			// 타임아웃 채널
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				// 정상 완료
			case <-ctx.Done():
				// 타임아웃
				requestID := GetRequestID(r.Context())
				log.Printf("[%s] 요청 타임아웃: %v", requestID, duration)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGatewayTimeout)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "요청 처리 시간 초과",
				})
			}
		})
	}
}

// ============================================================
// 핸들러 예제
// ============================================================

func handleHello(w http.ResponseWriter, r *http.Request) {
	requestID := GetRequestID(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "안녕하세요!",
		"request_id": requestID,
	})
}

func handlePanic(w http.ResponseWriter, r *http.Request) {
	// Recovery 미들웨어가 이 패닉을 처리합니다.
	panic("의도적인 패닉 테스트!")
}

func handleSlow(w http.ResponseWriter, r *http.Request) {
	// 타임아웃 미들웨어 테스트를 위한 느린 핸들러
	select {
	case <-time.After(5 * time.Second):
		json.NewEncoder(w).Encode(map[string]string{"message": "느린 응답"})
	case <-r.Context().Done():
		// 컨텍스트가 취소되면 조기 종료
		log.Println("요청이 취소되었습니다")
	}
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	requestID := GetRequestID(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"request_id": requestID,
		"method":     r.Method,
		"path":       r.URL.Path,
		"headers":    r.Header,
	})
}

// ============================================================
// 메인 함수
// ============================================================

func main() {
	mux := http.NewServeMux()

	// 핸들러 등록
	mux.HandleFunc("GET /hello", handleHello)
	mux.HandleFunc("GET /panic", handlePanic)   // Recovery 테스트
	mux.HandleFunc("GET /slow", handleSlow)     // Timeout 테스트
	mux.HandleFunc("GET /echo", handleEcho)
	mux.HandleFunc("OPTIONS /hello", handleHello) // CORS preflight

	// 미들웨어 체이닝
	// 실행 순서: RequestID -> CORS -> Logging -> Recovery -> 핸들러
	// Chain 함수를 사용하면 코드 가독성이 향상됩니다.
	handler := Chain(
		mux,
		RequestID,
		CORS(DefaultCORSConfig()),
		Logging,
		Recovery,
		Timeout(3*time.Second),
	)

	addr := ":8081"
	fmt.Printf("미들웨어 데모 서버: http://localhost%s\n", addr)
	fmt.Println()
	fmt.Println("테스트 명령어:")
	fmt.Println("  curl -v http://localhost:8081/hello")
	fmt.Println("  curl -v http://localhost:8081/echo")
	fmt.Println("  curl -v http://localhost:8081/panic  # Recovery 테스트")
	fmt.Println("  curl -v http://localhost:8081/slow   # Timeout 테스트")
	fmt.Println("  curl -v -X OPTIONS -H 'Origin: http://example.com' http://localhost:8081/hello")

	log.Fatal(http.ListenAndServe(addr, handler))
}
