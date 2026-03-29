# 04-slog: 구조화된 로깅 (log/slog)

> Go 1.21에서 표준 라이브러리에 추가된 `log/slog`는 key-value 쌍 기반의 구조화된 로깅을 제공합니다. `logrus`, `zerolog`, `zap` 같은 서드파티 로거를 대체하는 공식 솔루션입니다.

---

## 1. 구조화된 로깅이란

### 비구조화 로그 (전통적 방식)

```
2024-01-15 10:23:45 INFO  사용자 로그인: 김철수 (ID: 42), IP: 192.168.1.1
2024-01-15 10:23:46 ERROR DB 연결 실패: connection refused, 재시도: 3/5
```

문자열로만 이루어져 있어 파싱이 어렵고, 특정 사용자의 로그를 검색하거나 집계하기 힘듭니다.

### 구조화된 로그 (slog 방식)

```json
{"time":"2024-01-15T10:23:45Z","level":"INFO","msg":"사용자 로그인","username":"김철수","user_id":42,"ip":"192.168.1.1"}
{"time":"2024-01-15T10:23:46Z","level":"ERROR","msg":"DB 연결 실패","error":"connection refused","retry":3,"max_retry":5}
```

JSON이므로 Elasticsearch, Loki, CloudWatch Logs 같은 로그 집계 시스템이 자동으로 파싱하고 인덱싱할 수 있습니다.

**구조화 로깅의 장점**:
- 특정 `user_id`의 모든 요청 추적 가능
- 에러 발생 빈도, 평균 응답 시간 등 지표 집계 가능
- 로그 파이프라인(Fluentd, Vector 등)과 통합 용이

---

## 2. log/slog 기본 사용법

### 패키지 레벨 함수

```go
import "log/slog"

// 기본 로거 (TextHandler, os.Stderr 출력)
slog.Info("서버 시작", slog.String("addr", ":8080"))
slog.Debug("디버그 정보")     // 기본 레벨은 Info — Debug는 출력 안 됨
slog.Warn("경고 발생", slog.Int("count", 42))
slog.Error("오류 발생", slog.String("error", err.Error()))
```

### slog.Attr 타입들

```go
slog.String("key", "value")          // 문자열
slog.Int("count", 42)                // 정수
slog.Int64("id", int64(id))          // int64
slog.Float64("ratio", 0.95)          // 부동소수점
slog.Bool("enabled", true)           // 불리언
slog.Duration("elapsed", 150*time.Millisecond)  // 시간 간격
slog.Time("created_at", time.Now())  // 시각
slog.Any("data", someStruct)         // 임의 타입
```

---

## 3. TextHandler vs JSONHandler

### TextHandler: 개발 환경

```go
handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
logger := slog.New(handler)

logger.Info("요청 처리", slog.String("path", "/api/users"), slog.Int("status", 200))
// 출력: time=2024-01-15T10:23:45.000Z level=INFO msg=요청 처리 path=/api/users status=200
```

사람이 읽기 쉬운 형식입니다. 터미널 개발 환경에 적합합니다.

### JSONHandler: 프로덕션 환경

```go
handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
logger := slog.New(handler)

logger.Info("요청 처리", slog.String("path", "/api/users"), slog.Int("status", 200))
// 출력: {"time":"2024-01-15T10:23:45Z","level":"INFO","msg":"요청 처리","path":"/api/users","status":200}
```

로그 집계 시스템이 파싱할 수 있는 형식입니다. 프로덕션 서버에 적합합니다.

### HandlerOptions

```go
opts := &slog.HandlerOptions{
    Level:     slog.LevelDebug,    // 최소 출력 레벨
    AddSource: true,               // 소스 파일 위치 포함 (파일명:줄번호)
    ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
        // 특정 키 이름 변경 또는 값 마스킹
        if a.Key == "password" {
            return slog.String("password", "***")
        }
        return a
    },
}
```

---

## 4. 로그 레벨

```go
// 표준 레벨 (낮을수록 상세)
slog.LevelDebug = -4    // 상세 디버그 정보
slog.LevelInfo  =  0    // 일반 운영 정보
slog.LevelWarn  =  4    // 주의가 필요한 상황
slog.LevelError =  8    // 오류 발생
```

### 동적 레벨 변경: slog.LevelVar

배포 중인 서버의 로그 레벨을 재시작 없이 변경할 수 있습니다.

```go
var logLevel slog.LevelVar
logLevel.Set(slog.LevelInfo)  // 초기값

handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: &logLevel,  // 포인터 전달 — LevelVar 변경이 즉시 반영됨
})
logger := slog.New(handler)

// HTTP 엔드포인트로 런타임 변경
http.HandleFunc("PUT /admin/log-level", func(w http.ResponseWriter, r *http.Request) {
    var req struct{ Level string `json:"level"` }
    json.NewDecoder(r.Body).Decode(&req)

    switch req.Level {
    case "debug":
        logLevel.Set(slog.LevelDebug)
    case "warn":
        logLevel.Set(slog.LevelWarn)
    default:
        logLevel.Set(slog.LevelInfo)
    }
    w.WriteHeader(http.StatusNoContent)
})
```

---

## 5. With와 Group: 로그 구조화

### With: 공통 속성 미리 설정

모든 로그에 반복적으로 포함되는 속성을 미리 설정합니다.

```go
// 서비스 레벨 공통 속성
baseLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
serviceLogger := baseLogger.With(
    slog.String("service", "user-api"),
    slog.String("version", "1.2.3"),
    slog.String("env", "production"),
)

// 이후 모든 로그에 service, version, env가 자동 포함됨
serviceLogger.Info("서버 시작")
// {"service":"user-api","version":"1.2.3","env":"production","msg":"서버 시작",...}

// 요청별 속성 추가
requestLogger := serviceLogger.With(
    slog.String("request_id", "req-abc123"),
    slog.String("user_id", "user-456"),
)
requestLogger.Info("주문 생성")
// {"service":"user-api","request_id":"req-abc123","user_id":"user-456","msg":"주문 생성",...}
```

### Group: 관련 속성 묶기

```go
logger.Info("HTTP 요청 완료",
    slog.Group("request",
        slog.String("method", "POST"),
        slog.String("path", "/api/orders"),
        slog.Int("bytes", 1024),
    ),
    slog.Group("response",
        slog.Int("status", 201),
        slog.Duration("duration", 45*time.Millisecond),
    ),
)
// {"msg":"HTTP 요청 완료","request":{"method":"POST","path":"/api/orders","bytes":1024},"response":{"status":201,"duration":"45ms"}}

// WithGroup: 이후 모든 속성을 그룹 아래에 추가
httpLogger := logger.WithGroup("http")
httpLogger.Info("요청", slog.String("method", "GET"))
// {"http":{"msg":"요청","method":"GET"}}
```

---

## 6. 컨텍스트 통합

요청 ID나 추적 정보를 컨텍스트를 통해 전파합니다.

```go
// 컨텍스트에 로거 저장
type ctxKey string
const loggerKey ctxKey = "logger"

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, logger)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}

// slog.InfoContext — 컨텍스트를 직접 전달 (핸들러가 컨텍스트에서 추가 정보 추출 가능)
slog.InfoContext(ctx, "처리 시작", slog.String("order_id", id))

// 비즈니스 로직에서 컨텍스트 로거 사용
func processPayment(ctx context.Context, amount float64) error {
    logger := LoggerFromContext(ctx)
    logger.Info("결제 시작", slog.Float64("amount", amount))
    // ...
    logger.Info("결제 완료")
    return nil
}
```

---

## 7. 새로운 slog API (Go 1.25~1.26)

### slog.GroupAttrs() — Go 1.25

`slog.GroupAttrs`는 여러 `Attr`을 그룹으로 묶어 반환합니다. `slog.Group`과 달리 `Attr` 타입을 반환하므로 `With`나 다른 `Attr` 슬라이스와 조합할 때 편리합니다.

```go
// Go 1.25+
attrs := slog.GroupAttrs("request",
    slog.String("method", "GET"),
    slog.String("path", "/api/users"),
    slog.Int("status", 200),
)
logger.Info("HTTP 완료", attrs)
// {"msg":"HTTP 완료","request":{"method":"GET","path":"/api/users","status":200}}
```

### Record.Source() — Go 1.25

`slog.Record`에 `Source()` 메서드가 추가되어 커스텀 핸들러에서 호출 위치 정보에 쉽게 접근할 수 있습니다.

```go
// Go 1.25+ — 커스텀 핸들러에서 소스 위치 추출
func (h *MyHandler) Handle(ctx context.Context, r slog.Record) error {
    if src := r.Source(); src != nil {
        fmt.Printf("%s:%d (%s)\n", src.File, src.Line, src.Function)
    }
    // ...
    return nil
}
```

### slog.MultiHandler — Go 1.26

Go 1.26에서 `slog.MultiHandler`가 표준 라이브러리에 추가되었습니다. 여러 핸들러에 동시에 로그를 기록할 수 있습니다.

```go
// Go 1.26+ — 여러 핸들러에 동시 기록
textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
})
jsonHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})

// 두 핸들러에 동시에 로그를 기록
multi := slog.NewMultiHandler(textHandler, jsonHandler)
logger := slog.New(multi)

logger.Info("서버 시작", slog.String("addr", ":8080"))
// stdout: 텍스트 형식으로 출력 (Debug 이상)
// logFile: JSON 형식으로 기록 (Info 이상)
```

Go 1.26 이전에는 `MultiHandler`를 직접 구현해야 했습니다:

```go
// Go 1.25 이하 — 직접 구현 필요
type multiHandler struct{ handlers []slog.Handler }

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
    for _, h := range m.handlers {
        if h.Enabled(ctx, level) { return true }
    }
    return false
}
func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
    for _, h := range m.handlers {
        if h.Enabled(ctx, r.Level) {
            h.Handle(ctx, r.Clone())
        }
    }
    return nil
}
func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    hs := make([]slog.Handler, len(m.handlers))
    for i, h := range m.handlers { hs[i] = h.WithAttrs(attrs) }
    return &multiHandler{hs}
}
func (m *multiHandler) WithGroup(name string) slog.Handler {
    hs := make([]slog.Handler, len(m.handlers))
    for i, h := range m.handlers { hs[i] = h.WithGroup(name) }
    return &multiHandler{hs}
}
```

---

## 8. 커스텀 핸들러

`slog.Handler` 인터페이스를 구현하면 어떤 출력 형식이든 만들 수 있습니다.

```go
type Handler interface {
    Enabled(context.Context, Level) bool    // 해당 레벨 출력 여부
    Handle(context.Context, Record) error   // 로그 레코드 처리
    WithAttrs(attrs []Attr) Handler         // With() 지원
    WithGroup(name string) Handler          // WithGroup() 지원
}
```

### 개발용 컬러 핸들러 예시

```go
type PrettyHandler struct {
    slog.Handler  // 임베딩으로 WithAttrs, WithGroup 위임
    out io.Writer
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
    color := map[slog.Level]string{
        slog.LevelDebug: "\033[90m",   // 회색
        slog.LevelInfo:  "\033[34m",   // 파란색
        slog.LevelWarn:  "\033[33m",   // 노란색
        slog.LevelError: "\033[31m",   // 빨간색
    }[r.Level]

    fmt.Fprintf(h.out, "%s[%s] %s\033[0m",
        color, r.Level, r.Message)

    r.Attrs(func(a slog.Attr) bool {
        fmt.Fprintf(h.out, " %s=%v", a.Key, a.Value)
        return true
    })
    fmt.Fprintln(h.out)
    return nil
}
```

---

## 8. HTTP 미들웨어에서의 구조화된 로깅

```go
func SlogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // 요청별 로거: 공통 필드를 미리 설정
            reqLogger := logger.With(
                slog.String("request_id", r.Header.Get("X-Request-ID")),
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
                slog.String("remote_ip", r.RemoteAddr),
            )

            // 컨텍스트에 로거 저장 → 핸들러에서 꺼내 쓸 수 있음
            ctx := WithLogger(r.Context(), reqLogger)
            rw := &statusRecorder{ResponseWriter: w, status: 200}

            next.ServeHTTP(rw, r.WithContext(ctx))

            // 요청 완료 로그 — 응답 정보 포함
            reqLogger.Info("요청 완료",
                slog.Int("status", rw.status),
                slog.Duration("duration", time.Since(start)),
                slog.Int64("bytes", rw.written),
            )
        })
    }
}
```

---

## 9. logrus / zerolog / zap 비교

| | logrus | zerolog | zap | slog (표준) |
|-|--------|---------|-----|------------|
| 출시 | 2013 | 2019 | 2018 | 2023 (Go 1.21) |
| 성능 | 보통 | 매우 빠름 | 매우 빠름 | 빠름 |
| API | 체이닝 | 체이닝 | 구조체 | 함수/메서드 |
| 외부 의존성 | O | O | O | X |
| 표준 통합 | X | X | X | O |

**왜 slog로 수렴하는가**:
- 표준 라이브러리이므로 외부 의존성 없음
- `slog.Handler` 인터페이스로 백엔드 교체 가능
- zerolog/zap 어댑터가 있어 기존 코드와 통합 가능
- Go 팀이 장기 지원 보장

```go
// slog + zerolog 백엔드 조합 예시 (성능이 중요한 경우)
// github.com/mdobak/go-xerrors 같은 어댑터 패키지 사용
```

---

## 10. 프로덕션 로깅 전략

```go
func setupLogger() *slog.Logger {
    // 환경에 따라 핸들러 선택
    var handler slog.Handler
    if os.Getenv("ENV") == "development" {
        // 개발: 텍스트, Debug 레벨
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    } else {
        // 프로덕션: JSON, Info 레벨
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level:     slog.LevelInfo,
            AddSource: true,  // 소스 위치 포함
        })
    }

    logger := slog.New(handler).With(
        slog.String("service", "my-service"),
        slog.String("version", os.Getenv("APP_VERSION")),
    )

    // 전역 기본 로거로 설정
    slog.SetDefault(logger)
    return logger
}
```

**컨테이너 환경 권장 사항**:
- 로그는 항상 `stdout`/`stderr`로 출력 (파일 X)
- 컨테이너 런타임(Docker, Kubernetes)이 로그를 수집
- JSON 형식으로 출력 → Fluentd/Vector → Elasticsearch/Loki
