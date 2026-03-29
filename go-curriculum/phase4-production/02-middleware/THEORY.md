# 02-middleware: HTTP 미들웨어 패턴

> 미들웨어는 HTTP 핸들러를 감싸는 함수입니다. 로깅, 인증, 복구 같은 횡단 관심사(cross-cutting concerns)를 핵심 비즈니스 로직에서 분리합니다.

---

## 1. 미들웨어란

**미들웨어 시그니처**:

```go
type Middleware func(http.Handler) http.Handler
```

HTTP 핸들러를 받아서 HTTP 핸들러를 반환합니다. 반환된 핸들러는 내부에서 원본 핸들러를 호출하거나, 조건에 따라 건너뜁니다.

```go
// 가장 단순한 미들웨어 예시 — 아무것도 하지 않는 통과형
func NoOp(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 핸들러 호출 전에 무언가 할 수 있음
        next.ServeHTTP(w, r)
        // 핸들러 호출 후에 무언가 할 수 있음
    })
}
```

---

## 2. 미들웨어 체이닝과 실행 순서

미들웨어는 양파처럼 겹겹이 쌓입니다.

```go
// 직접 중첩: A(B(C(handler)))
handler = C(handler)
handler = B(handler)
handler = A(handler)

// Chain 헬퍼 함수로 가독성 향상
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

handler := Chain(mux, RequestID, Logging, Recovery)
```

**실행 흐름** (요청 진입 → 응답 반환):

```
요청 →  RequestID  →  Logging  →  Recovery  →  핸들러
응답 ←  RequestID  ←  Logging  ←  Recovery  ←  핸들러
```

코드로 표현하면:

```go
// RequestID 미들웨어 내부
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ① 요청 진입 시 실행
        id := generateID()
        ctx := context.WithValue(r.Context(), requestIDKey, id)
        w.Header().Set("X-Request-ID", id)

        next.ServeHTTP(w, r.WithContext(ctx))  // ② 다음 미들웨어/핸들러 호출

        // ③ 응답 반환 후 실행 (여기서는 없음)
    })
}
```

---

## 3. 미들웨어 1: 요청 ID (Request ID)

모든 요청에 고유 ID를 부여합니다. 분산 추적, 로그 상관관계 분석에 필수입니다.

```go
type contextKey string
const requestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 업스트림 프록시가 이미 ID를 설정했을 수 있음
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = generateID()  // UUID 또는 타임스탬프 기반
        }

        // 컨텍스트에 저장 — 하위 핸들러에서 추출 가능
        ctx := context.WithValue(r.Context(), requestIDKey, requestID)
        // 응답 헤더에 포함 — 클라이언트가 추적 가능
        w.Header().Set("X-Request-ID", requestID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 컨텍스트에서 요청 ID 추출 (핸들러에서 사용)
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return "unknown"
}
```

**컨텍스트 키는 반드시 전용 타입으로 정의하세요.** `string` 타입 키를 그대로 사용하면 다른 패키지와 키가 충돌할 수 있습니다.

---

## 4. 미들웨어 2: 로깅

모든 HTTP 요청과 응답을 기록합니다. 응답 상태 코드를 캡처하려면 `ResponseWriter`를 래핑해야 합니다.

```go
// ResponseWriter 래퍼 — 상태 코드 캡처
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

// 로깅 미들웨어
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := newResponseWriter(w)

        next.ServeHTTP(wrapped, r)

        log.Printf("[%s] %s %s → %d (%v)",
            GetRequestID(r.Context()),
            r.Method,
            r.URL.Path,
            wrapped.statusCode,
            time.Since(start),
        )
    })
}
```

**왜 래퍼가 필요한가**: `http.ResponseWriter`는 `WriteHeader`를 호출한 후에는 상태 코드를 읽을 방법이 없습니다. 래퍼로 가로채서 기록해야 합니다.

---

## 5. 미들웨어 3: 패닉 복구 (Recovery)

핸들러에서 발생한 `panic`을 잡아 500 응답으로 변환합니다. 이 미들웨어가 없으면 패닉 발생 시 서버 프로세스 전체가 죽을 수 있습니다.

```go
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                // 스택 트레이스 로깅
                log.Printf("패닉 복구 [%s]: %v\n%s",
                    GetRequestID(r.Context()),
                    err,
                    debug.Stack(),
                )

                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "내부 서버 오류",
                })
            }
        }()

        next.ServeHTTP(w, r)
    })
}
```

**Recovery는 반드시 Logging 바깥에 위치해야 합니다.** Recovery가 안쪽에 있으면 패닉 복구 후 Logging이 상태 코드 500을 기록하지 못합니다.

```go
// 올바른 순서 — Recovery가 Logging 바깥
handler := Chain(mux,
    RequestID,   // 가장 바깥 (요청 ID 먼저)
    Logging,     // Recovery 전에 시작하고, 후에 로깅
    Recovery,    // 핸들러 바로 바깥에서 패닉 캐치
)
```

---

## 6. 미들웨어 4: CORS

브라우저의 동일 출처 정책(Same-Origin Policy)을 우회하는 CORS 헤더를 설정합니다.

```go
type CORSConfig struct {
    AllowedOrigins []string
    AllowedMethods []string
    AllowedHeaders []string
    MaxAge         int
}

func CORS(cfg CORSConfig) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin",
                strings.Join(cfg.AllowedOrigins, ", "))
            w.Header().Set("Access-Control-Allow-Methods",
                strings.Join(cfg.AllowedMethods, ", "))
            w.Header().Set("Access-Control-Allow-Headers",
                strings.Join(cfg.AllowedHeaders, ", "))

            // Preflight 요청 — 브라우저가 실제 요청 전에 보내는 OPTIONS
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return  // 핸들러 호출 없이 즉시 반환
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

**Preflight 요청**: 브라우저는 `POST`, `PUT` 등의 요청 전에 `OPTIONS` 메서드로 서버가 CORS를 허용하는지 먼저 확인합니다. 이 요청에 빠르게 응답해야 실제 요청이 진행됩니다.

---

## 7. 미들웨어 5: 요청 타임아웃

처리 시간이 너무 긴 요청을 중단합니다. `context.WithTimeout`으로 컨텍스트에 마감 시간을 설정합니다.

```go
func Timeout(duration time.Duration) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), duration)
            defer cancel()

            done := make(chan struct{})
            go func() {
                next.ServeHTTP(w, r.WithContext(ctx))
                close(done)
            }()

            select {
            case <-done:
                // 정상 완료
            case <-ctx.Done():
                w.WriteHeader(http.StatusGatewayTimeout)
                json.NewEncoder(w).Encode(map[string]string{
                    "error": "요청 처리 시간 초과",
                })
            }
        })
    }
}
```

**핸들러에서도 컨텍스트를 확인해야 합니다.** 타임아웃 컨텍스트를 DB 쿼리, HTTP 호출 등에 전달하면 타임아웃 시 자동으로 중단됩니다.

```go
func handleSlow(w http.ResponseWriter, r *http.Request) {
    select {
    case <-time.After(5 * time.Second):
        // 실제 처리
    case <-r.Context().Done():
        // 타임아웃 또는 클라이언트 연결 끊김
        return
    }
}
```

---

## 8. 설정을 받는 미들웨어 패턴

미들웨어가 설정이 필요할 때는 클로저로 감쌉니다.

```go
// 패턴: 설정을 받아 Middleware를 반환
func RateLimit(requestsPerSecond int) Middleware {
    limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                w.WriteHeader(http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// 사용
handler := Chain(mux,
    RequestID,
    RateLimit(100),  // 초당 100 요청 제한
    CORS(DefaultCORSConfig()),
    Logging,
    Recovery,
)
```

---

## 9. Express.js / Django 미들웨어와의 비교

### Express.js (JavaScript)

```javascript
// Express 미들웨어: (req, res, next) => void
app.use((req, res, next) => {
    console.log(`${req.method} ${req.path}`)
    next()  // 다음 미들웨어로 이동
})

// 에러 미들웨어: (err, req, res, next) => void
app.use((err, req, res, next) => {
    res.status(500).json({ error: err.message })
})
```

```go
// Go 미들웨어: func(http.Handler) http.Handler
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)  // next()에 해당
    })
}
```

**차이점**: Express는 `next()` 함수를 호출 규약으로 사용하지만, Go는 `next.ServeHTTP(w, r)`을 호출합니다. Go의 방식은 타입 시스템이 보장하며, `next`를 호출하지 않으면 단순히 요청이 이 미들웨어에서 멈춥니다.

### Django (Python)

```python
# Django 미들웨어 클래스
class RequestIDMiddleware:
    def __init__(self, get_response):
        self.get_response = get_response

    def __call__(self, request):
        request.META['REQUEST_ID'] = str(uuid.uuid4())
        response = self.get_response(request)
        response['X-Request-ID'] = request.META['REQUEST_ID']
        return response
```

Go의 미들웨어는 Django의 `__init__`(설정 단계) + `__call__`(요청 처리)과 구조적으로 동일하지만, 클로저로 더 간결하게 표현합니다.
