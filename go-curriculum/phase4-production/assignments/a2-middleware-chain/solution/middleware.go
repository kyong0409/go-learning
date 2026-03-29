// a2-middleware-chain/solution/middleware.go
// HTTP 미들웨어 체인 참고 답안입니다.
package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Middleware는 미들웨어 함수의 타입입니다.
type Middleware func(http.Handler) http.Handler

// Chain은 미들웨어를 체이닝합니다. 첫 번째가 가장 바깥쪽에서 실행됩니다.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// ── RateLimiter ──────────────────────────────────────────────

type bucket struct {
	tokens   float64
	lastTime time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64
	capacity float64
}

func RateLimiter(rps int) Middleware {
	rl := &rateLimiter{
		buckets:  make(map[string]*bucket),
		rate:     float64(rps),
		capacity: float64(rps),
	}
	go func() {
		for range time.Tick(time.Minute) {
			rl.mu.Lock()
			cutoff := time.Now().Add(-5 * time.Minute)
			for ip, b := range rl.buckets {
				if b.lastTime.Before(cutoff) {
					delete(rl.buckets, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = strings.Split(fwd, ",")[0]
			}

			rl.mu.Lock()
			now := time.Now()
			b, ok := rl.buckets[ip]
			if !ok {
				b = &bucket{tokens: rl.capacity, lastTime: now}
				rl.buckets[ip] = b
			}
			elapsed := now.Sub(b.lastTime).Seconds()
			b.tokens = min64(b.tokens+elapsed*rl.rate, rl.capacity)
			b.lastTime = now
			allowed := b.tokens >= 1
			if allowed {
				b.tokens--
			}
			rl.mu.Unlock()

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprint(w, `{"error":"요청이 너무 많습니다","code":"RATE_LIMIT_EXCEEDED"}`)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func min64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ── RequestLogger ────────────────────────────────────────────

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func RequestLogger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := fmt.Sprintf("%d", start.UnixNano())
			w.Header().Set("X-Request-ID", reqID)

			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sr, r)

			logger.Info("HTTP 요청",
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sr.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// ── Authenticator ────────────────────────────────────────────

func Authenticator(token string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// /health는 인증 제외
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			got := strings.TrimPrefix(auth, "Bearer ")
			if got == "" || got == auth || got != token {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error":"인증이 필요합니다","code":"UNAUTHORIZED"}`)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ── ResponseTimer ────────────────────────────────────────────

func ResponseTimer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		d := time.Since(start)
		var val string
		if d < time.Second {
			val = fmt.Sprintf("%dms", d.Milliseconds())
		} else {
			val = fmt.Sprintf("%.3fs", d.Seconds())
		}
		w.Header().Set("X-Response-Time", val)
	})
}
