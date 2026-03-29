// internal/handler/middleware.go
// HTTP 미들웨어를 구현합니다.
// 요청 로깅, 속도 제한, 요청 ID 생성을 제공합니다.
package handler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ============================================================
// 컨텍스트 키
// ============================================================

type middlewareCtxKey string

const (
	reqIDKey    middlewareCtxKey = "request_id"
	reqStartKey middlewareCtxKey = "request_start"
)

// ============================================================
// 요청 ID 미들웨어
// ============================================================

// RequestID는 각 요청에 고유 ID를 부여하는 미들웨어입니다.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("%d-%04d", time.Now().UnixNano(), rand.Intn(10000))
		}
		ctx := context.WithValue(r.Context(), reqIDKey, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID는 컨텍스트에서 요청 ID를 반환합니다.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(reqIDKey).(string); ok {
		return id
	}
	return "unknown"
}

// ============================================================
// 구조화된 로깅 미들웨어
// ============================================================

// statusWriter는 HTTP 상태 코드를 캡처하는 ResponseWriter 래퍼입니다.
type statusWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	n, err := sw.ResponseWriter.Write(b)
	sw.size += n
	return n, err
}

// Logger는 slog를 사용하는 구조화된 HTTP 로깅 미들웨어입니다.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			duration := time.Since(start)
			reqID := GetRequestID(r.Context())

			// 상태 코드에 따라 로그 레벨 결정
			logFn := logger.Info
			if sw.status >= 500 {
				logFn = logger.Error
			} else if sw.status >= 400 {
				logFn = logger.Warn
			}

			logFn("HTTP 요청",
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.Int("status", sw.status),
				slog.Int("size", sw.size),
				slog.Duration("duration", duration),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}

// ============================================================
// 속도 제한 미들웨어 (토큰 버킷 알고리즘)
// ============================================================

// tokenBucket은 IP당 토큰 버킷을 관리합니다.
type tokenBucket struct {
	tokens   float64
	lastTime time.Time
}

// RateLimiter는 IP 기반 속도 제한 미들웨어입니다.
// 토큰 버킷 알고리즘을 사용합니다.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     float64 // 초당 요청 수
	capacity float64 // 버킷 최대 용량
}

// NewRateLimiter는 새 속도 제한기를 생성합니다.
// rate: 초당 허용 요청 수
// capacity: 최대 버스트 크기
func NewRateLimiter(rate, capacity float64) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
	}

	// 오래된 버킷 정기 정리 (메모리 누수 방지)
	go rl.cleanup()

	return rl
}

// Allow는 요청을 허용할지 확인합니다.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.buckets[ip]
	if !exists {
		bucket = &tokenBucket{tokens: rl.capacity, lastTime: now}
		rl.buckets[ip] = bucket
	}

	// 경과 시간에 비례해 토큰 보충
	elapsed := now.Sub(bucket.lastTime).Seconds()
	bucket.tokens = min64(bucket.tokens+elapsed*rl.rate, rl.capacity)
	bucket.lastTime = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}
	return false
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// cleanup은 1분마다 오래된 버킷을 삭제합니다.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-5 * time.Minute)
		for ip, bucket := range rl.buckets {
			if bucket.lastTime.Before(cutoff) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware는 속도 제한 미들웨어를 반환합니다.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 클라이언트 IP 추출 (프록시 환경 고려)
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			ip = realIP
		}

		if !rl.Allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error":"요청이 너무 많습니다. 잠시 후 다시 시도해주세요.","code":"RATE_LIMIT_EXCEEDED"}`)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================
// CORS 미들웨어
// ============================================================

// CORS는 Cross-Origin Resource Sharing 헤더를 설정하는 미들웨어입니다.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================
// 복구 미들웨어
// ============================================================

// Recoverer는 패닉을 복구하는 미들웨어입니다.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqID := GetRequestID(r.Context())
					logger.Error("패닉 복구",
						slog.String("request_id", reqID),
						slog.Any("error", err),
					)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error":"내부 서버 오류","code":"INTERNAL_ERROR","request_id":"%s"}`, reqID)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
