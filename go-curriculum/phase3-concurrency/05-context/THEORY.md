# 05. context 패키지

## context란?

`context.Context`는 **요청 범위의 취소 신호, 데드라인, 값**을 고루틴 트리 전체에 전파하는 표준 메커니즘입니다.

> **Go 1.24 신규**: `testing.T.Context()`가 추가되어 테스트가 종료될 때 자동으로 취소되는 컨텍스트를 반환합니다. 아래 [테스트에서의 context](#테스트에서의-context) 섹션 참고.

복잡한 서비스에서는 하나의 HTTP 요청이 수십 개의 고루틴(DB 쿼리, 외부 API 호출, 캐시 조회 등)을 생성합니다. 클라이언트가 연결을 끊거나 타임아웃이 발생하면, 이 고루틴들을 모두 정리해야 합니다. `context`는 이런 생명주기 관리를 위한 솔루션입니다.

```
HTTP 요청 컨텍스트
    ├── DB 쿼리 고루틴
    │     └── 서브쿼리 고루틴
    ├── 외부 API 호출 고루틴
    └── 캐시 조회 고루틴

→ 요청 취소 시 모든 하위 고루틴이 연쇄 취소됨
```

---

## Context 인터페이스

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool) // 취소 시각 (설정된 경우)
    Done() <-chan struct{}                    // 취소 시 닫히는 채널
    Err() error                              // 취소 이유 (Canceled or DeadlineExceeded)
    Value(key interface{}) interface{}        // 키-값 조회
}
```

---

## 루트 컨텍스트

```go
// context.Background(): 루트 컨텍스트
// - 취소 없음, 값 없음, 데드라인 없음
// - main(), 서버 초기화, 테스트 등 최상위에서 사용
ctx := context.Background()

// context.TODO(): 아직 어떤 컨텍스트를 써야 할지 불확실할 때
// - 기술적으로 Background와 동일
// - 정적 분석 도구(go vet, staticcheck)가 TODO를 감지해 경고 가능
// - 나중에 채워야 함을 표시하는 플레이스홀더
ctx := context.TODO()
```

---

## context.WithCancel: 수동 취소

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // 중요: 반드시 cancel을 호출해야 리소스 누수 방지

// 고루틴에 컨텍스트 전달
go func(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            // ctx.Err() == context.Canceled
            fmt.Println("취소됨:", ctx.Err())
            return
        default:
            doWork()
        }
    }
}(ctx)

// 어느 시점에 취소
cancel() // Done() 채널을 닫아 모든 수신자에게 브로드캐스트
```

### 연쇄 취소

부모 컨텍스트가 취소되면 모든 자식 컨텍스트도 자동으로 취소됩니다.

```go
parent, parentCancel := context.WithCancel(context.Background())
child1, child1Cancel := context.WithCancel(parent)
child2, _ := context.WithTimeout(parent, 10*time.Second)

defer parentCancel()
defer child1Cancel()

// parent 취소 시 child1, child2도 자동 취소
parentCancel()

// child1.Err() == context.Canceled
// child2.Err() == context.Canceled
```

---

## context.WithTimeout: 지정 시간 후 자동 취소

```go
// 2초 타임아웃
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel() // 타임아웃 전에 완료되면 즉시 리소스 해제

result, err := doSlowOperation(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("타임아웃!")
    }
}
```

실제 사용 예: DB 쿼리에 타임아웃 적용

```go
func queryUser(ctx context.Context, id int) (*User, error) {
    // 쿼리 타임아웃: 부모 컨텍스트 타임아웃의 일부로 설정
    ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
    defer cancel()

    // 대부분의 Go DB 드라이버는 ctx를 받아 쿼리 취소를 지원
    return db.QueryRowContext(ctx, "SELECT * FROM users WHERE id=$1", id)
}
```

---

## context.WithDeadline: 절대 시각에 취소

```go
// 특정 시각에 취소 (WithTimeout과 차이: 절대 시각 vs 상대 시간)
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()

// WithTimeout(ctx, d) == WithDeadline(ctx, time.Now().Add(d))
// 실제로 WithTimeout이 WithDeadline을 래핑

// 마감 시간 확인
if dl, ok := ctx.Deadline(); ok {
    remaining := time.Until(dl)
    fmt.Printf("남은 시간: %v\n", remaining)
}
```

---

## context.WithValue: 요청 범위 값 전달

```go
// 키 타입: string, int 등 기본 타입을 직접 쓰면 다른 패키지와 충돌!
// → 비공개(unexported) 커스텀 타입 사용

type contextKey string // 비공개 타입

const (
    requestIDKey contextKey = "requestID"
    userIDKey    contextKey = "userID"
)

// 값 추가: 래퍼 함수 패턴 (타입 안전성 확보)
func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

// 값 꺼내기: 래퍼 함수 패턴
func GetRequestID(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKey).(string)
    return id, ok
}

// 사용
ctx = WithRequestID(ctx, "req-abc-123")
if id, ok := GetRequestID(ctx); ok {
    fmt.Println("요청 ID:", id)
}
```

### WithValue 올바른 사용

```go
// 올바른 사용: 요청 범위 메타데이터
// - 요청 ID, 추적 ID, 인증 토큰 (미들웨어에서 설정)
// - 사용자 ID, 세션 정보

// 잘못된 사용: 함수 매개변수를 대체하는 데이터
// Bad: DB 연결을 context에 넣음
ctx = context.WithValue(ctx, "db", db)
// Good: DB 연결은 명시적 매개변수로 전달
func getUser(ctx context.Context, db *sql.DB, id int) {}
```

---

## ctx.Done()과 ctx.Err()

```go
func longOperation(ctx context.Context) error {
    for i := 0; i < 100; i++ {
        // 방법 1: select로 취소 확인 (블로킹 연산과 함께)
        select {
        case <-ctx.Done():
            return ctx.Err() // context.Canceled 또는 context.DeadlineExceeded
        default:
            processStep(i)
        }

        // 방법 2: 직접 확인 (논블로킹 간단 확인)
        if ctx.Err() != nil {
            return ctx.Err()
        }
    }
    return nil
}

// 에러 구분
switch ctx.Err() {
case context.Canceled:
    fmt.Println("명시적 취소 (cancel() 호출)")
case context.DeadlineExceeded:
    fmt.Println("타임아웃 또는 데드라인 초과")
case nil:
    fmt.Println("취소 없음")
}
```

---

## context 사용 규칙 (공식 권고)

### 규칙 1: 첫 번째 매개변수로 전달

```go
// 관례: ctx context.Context가 항상 첫 번째
func DoWork(ctx context.Context, arg1 int, arg2 string) error { ... }
func FetchData(ctx context.Context, url string) ([]byte, error) { ... }

// 잘못된 방법: context가 중간에
func BadFunc(arg1 int, ctx context.Context, arg2 string) {} // 비관례적
```

### 규칙 2: 구조체에 저장하지 않음

```go
// Bad: 구조체 필드로 저장
type Handler struct {
    ctx context.Context // 금지!
    db  *sql.DB
}

// Good: 메서드 매개변수로 전달
type Handler struct {
    db *sql.DB
}
func (h *Handler) Handle(ctx context.Context, req Request) {}
```

컨텍스트는 요청 단위로 살아있어야 합니다. 구조체에 저장하면 생명주기를 제어하기 어렵습니다.

### 규칙 3: nil을 전달하지 않음

```go
// Bad
doWork(nil) // 수신 측에서 ctx.Done() 등 호출 시 패닉

// Good: 불확실하면 context.TODO() 사용
doWork(context.TODO())
```

### 규칙 4: WithValue는 요청 범위 데이터만

함수 매개변수로 전달해야 할 것들을 context에 숨기지 않습니다.

---

## HTTP 서버에서의 context

Go의 `net/http` 패키지는 각 요청에 컨텍스트를 자동으로 연결합니다.

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // r.Context(): 요청에 연결된 컨텍스트
    // 클라이언트 연결 끊기면 자동으로 Done() 채널 닫힘
    ctx := r.Context()

    // 요청별 타임아웃 추가
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    // 미들웨어에서 값 추가
    ctx = WithRequestID(ctx, generateID())

    // 모든 하위 호출에 ctx 전달
    user, err := userService.GetUser(ctx, userID)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            http.Error(w, "요청 처리 시간 초과", http.StatusGatewayTimeout)
            return
        }
        http.Error(w, "서버 오류", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

### 미들웨어 패턴

```go
// 요청 ID를 컨텍스트에 주입하는 미들웨어
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := uuid.New().String()
        ctx := WithRequestID(r.Context(), id)
        w.Header().Set("X-Request-ID", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## context 트리 구조 시각화

```
Background (루트)
    └── WithValue(requestID="req-001")
          ├── WithTimeout(3초) ← HTTP 요청 타임아웃
          │     ├── WithValue(userID="user-42") ← 인증 미들웨어
          │     │     └── DB 쿼리 고루틴
          │     └── 캐시 조회 고루틴
          └── WithCancel ← 수동 취소용
                └── 백그라운드 작업 고루틴
```

```go
root := context.Background()
withReq := context.WithValue(root, requestIDKey, "req-001")
timedCtx, cancel := context.WithTimeout(withReq, 3*time.Second)
defer cancel()
withUser := context.WithValue(timedCtx, userIDKey, "user-42")

// withUser에서 부모 값 접근 가능
reqID := withUser.Value(requestIDKey) // "req-001" (부모에서 상속)
userID := withUser.Value(userIDKey)   // "user-42"

// 데드라인은 가장 가까운 것이 적용
dl, _ := withUser.Deadline() // timedCtx의 3초 데드라인 상속
```

---

## Python/Java 비교

### Python asyncio CancelledError vs context.Canceled

```python
import asyncio

async def work(task_group):
    try:
        await asyncio.sleep(10)
    except asyncio.CancelledError:
        print("취소됨")
        raise

# 취소
task.cancel()
```

```go
func work(ctx context.Context) {
    select {
    case <-time.After(10 * time.Second):
        fmt.Println("완료")
    case <-ctx.Done():
        fmt.Println("취소됨:", ctx.Err())
    }
}

cancel() // 취소
```

### Java의 InterruptedException vs context.Canceled

Java에서 스레드 인터럽트와 달리 Go의 context는 계층 구조를 통해 자동으로 하위로 전파됩니다. 수신 측이 `ctx.Done()`을 확인하기만 하면 됩니다.

---

## 테스트에서의 context

### testing.T.Context() — Go 1.24 신규

Go 1.24(Feb 2025)에서 `*testing.T`와 `*testing.B`에 `.Context()` 메서드가 추가되었습니다. 이 메서드는 테스트가 종료될 때(정상 완료, 실패, 패닉 등) 자동으로 취소되는 컨텍스트를 반환합니다.

```go
func TestWithContext(t *testing.T) {
    // 테스트 종료 시 자동 취소되는 컨텍스트
    ctx := t.Context()

    // 이전 방식: context.Background() + cancel + defer
    // ctx, cancel := context.WithCancel(context.Background())
    // defer cancel()

    result, err := fetchData(ctx, "https://example.com")
    if err != nil {
        t.Fatalf("fetchData 실패: %v", err)
    }
    _ = result
}

// 서브테스트에서도 동작: 서브테스트 종료 시 해당 ctx만 취소됨
func TestSubtestContext(t *testing.T) {
    t.Run("서브테스트", func(t *testing.T) {
        ctx := t.Context() // 이 서브테스트가 끝나면 취소
        doWork(ctx)
    })
}
```

`t.Context()`는 `context.Background()`로 시작한 뒤 `defer cancel()`을 수동으로 호출하는 기존 패턴을 대체합니다. 테스트 종료 시 리소스(DB 연결, HTTP 클라이언트 등)가 확실히 정리됩니다.
