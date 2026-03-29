# 과제 A2: HTTP 미들웨어 체인 구현

## 목표

프로덕션 수준의 HTTP 미들웨어 컴포넌트를 구현합니다.

## 구현할 컴포넌트

### 1. `Chain(handler, middlewares...)` — 미들웨어 체이닝
여러 미들웨어를 순서대로 체이닝합니다.
```go
handler = Chain(myHandler, RequestLogger, RateLimiter(10), Authenticator("secret"))
// 실행 순서: RequestLogger -> RateLimiter -> Authenticator -> myHandler
```

### 2. `RateLimiter(rps int)` — 속도 제한 (토큰 버킷)
초당 `rps`개 요청만 허용합니다. 초과 시 `429 Too Many Requests`.
- IP별 독립 제한
- `X-Forwarded-For` 헤더 지원

### 3. `RequestLogger(logger *slog.Logger)` — 요청 로거
모든 요청을 구조화된 형식으로 기록합니다.
- 메서드, 경로, 상태 코드, 소요 시간 기록
- `X-Request-ID` 헤더 포함

### 4. `Authenticator(token string)` — Bearer 토큰 인증
`Authorization: Bearer <token>` 헤더를 검증합니다.
- 없거나 잘못된 경우 `401 Unauthorized`
- `/health` 경로는 인증 제외

### 5. `ResponseTimer` — 응답 시간 측정
`X-Response-Time` 헤더에 처리 시간을 추가합니다 (예: `"15ms"`).

## 채점 기준

| 항목 | 배점 |
|------|------|
| Chain 구현 | 20점 |
| RateLimiter 구현 | 20점 |
| RequestLogger 구현 | 20점 |
| Authenticator 구현 | 20점 |
| ResponseTimer 구현 | 20점 |
| **합계** | **100점** |

## 실행 방법

```bash
cd assignments/a2-middleware-chain
go test ./... -v
```

## 힌트

- `RateLimiter`는 `sync.Mutex`와 맵으로 IP별 상태를 관리하세요.
- `Authenticator`는 `strings.TrimPrefix`로 Bearer 토큰을 추출하세요.
- `ResponseTimer`는 핸들러 호출 전후 `time.Now()`와 `time.Since()`를 사용하세요.
- 응답 상태 코드를 캡처하려면 `http.ResponseWriter`를 래핑하세요.
