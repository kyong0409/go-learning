// a2-middleware-chain/middleware.go
// HTTP 미들웨어 체인 과제입니다.
// TODO 주석이 있는 모든 함수를 구현하세요.
package middleware

import (
	"log/slog"
	"net/http"
)

// Middleware는 미들웨어 함수의 타입입니다.
type Middleware func(http.Handler) http.Handler

// Chain은 여러 미들웨어를 순서대로 체이닝합니다.
// 첫 번째 미들웨어가 가장 바깥쪽에서 실행됩니다.
// 예: Chain(h, A, B, C) → A(B(C(h)))
// TODO: 구현하세요.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	return h
}

// RateLimiter는 IP별 초당 요청 수를 제한하는 미들웨어를 반환합니다.
// rps 초과 시 429 Too Many Requests를 반환합니다.
// TODO: 토큰 버킷 알고리즘으로 구현하세요.
func RateLimiter(rps int) Middleware {
	// TODO: IP별 버킷 관리 구조체를 초기화하세요.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: IP를 추출하고 속도 제한을 적용하세요.
			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger는 slog를 사용하여 HTTP 요청을 로깅하는 미들웨어를 반환합니다.
// 로그에는 method, path, status, duration, request_id가 포함되어야 합니다.
// TODO: 구현하세요.
func RequestLogger(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: 요청 전후에 로깅하세요.
			next.ServeHTTP(w, r)
		})
	}
}

// Authenticator는 Bearer 토큰 인증 미들웨어를 반환합니다.
// Authorization: Bearer <token> 헤더를 검증합니다.
// /health 경로는 인증 없이 통과합니다.
// 실패 시 401 Unauthorized를 반환합니다.
// TODO: 구현하세요.
func Authenticator(token string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: 인증을 구현하세요.
			next.ServeHTTP(w, r)
		})
	}
}

// ResponseTimer는 X-Response-Time 헤더에 처리 시간을 추가하는 미들웨어입니다.
// 형식: "15ms" 또는 "1.234s"
// TODO: 구현하세요.
func ResponseTimer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: 처리 시간을 측정하고 헤더에 추가하세요.
		next.ServeHTTP(w, r)
	})
}
