# 04. sync 패키지

## 채널 대신 sync를 쓰는 경우

Go는 채널 기반 통신을 권장하지만, 아래 상황에서는 `sync` 패키지가 더 적합합니다.

| 상황 | 권장 방식 | 이유 |
|------|----------|------|
| 데이터 전달 / 소유권 이전 | 채널 | CSP 철학에 부합 |
| 공유 캐시 / 레지스트리 보호 | sync.Mutex | 단순 상태 보호 |
| 읽기 많은 공유 자료구조 | sync.RWMutex | 읽기 동시성 최적화 |
| 일회성 초기화 (싱글톤) | sync.Once | 초기화 경쟁 방지 |
| 단순 카운터 / 플래그 | sync/atomic | 뮤텍스보다 빠름 |
| 객체 재사용 (GC 최적화) | sync.Pool | GC 부담 감소 |

---

## sync.Mutex: 임계 구역 보호

뮤텍스(Mutual Exclusion)는 한 번에 하나의 고루틴만 임계 구역(critical section)에 들어갈 수 있도록 보장합니다.

```go
type SafeCounter struct {
    mu    sync.Mutex // zero value로 바로 사용 가능
    value int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()         // 잠금: 다른 고루틴 대기
    defer c.mu.Unlock() // 잠금 해제: defer로 반드시 보장
    c.value++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value // 읽기도 보호 필요!
}
```

### Mutex 사용 규칙

```go
// 규칙 1: 항상 defer로 Unlock (패닉 시에도 해제 보장)
c.mu.Lock()
defer c.mu.Unlock() // 이 한 줄이 모든 return, panic 경우를 처리

// 규칙 2: Mutex를 값으로 복사하지 않음 (항상 포인터 사용)
type Bad struct { mu sync.Mutex }
func process(b Bad) {} // 복사됨 → 잠금 상태가 복사될 수 있어 위험

type Good struct { mu sync.Mutex }
func process(g *Good) {} // 포인터로 전달

// 규칙 3: 잠금 상태에서 다른 잠금 획득 주의 (데드락 가능)
mu1.Lock()
mu2.Lock() // 다른 고루틴이 mu2 후 mu1을 잠그면 데드락!

// 규칙 4: 잠금 보유 시간 최소화
mu.Lock()
data := sharedData   // 복사
mu.Unlock()          // 즉시 해제
process(data)        // 잠금 없이 처리
```

### 구조체에 임베딩하는 패턴

```go
// 안티패턴: Mutex를 노출
type BadCache struct {
    sync.Mutex // 임베딩하면 Lock/Unlock이 공개 메서드로 노출됨
    data map[string]int
}

// 권장: Mutex를 필드로 (비공개)
type Cache struct {
    mu   sync.Mutex // 소문자로 비공개
    data map[string]int
}
```

---

## sync.RWMutex: 읽기/쓰기 분리 뮤텍스

읽기가 쓰기보다 훨씬 많은 경우 `RWMutex`로 동시 읽기를 허용해 성능을 높입니다.

```go
type ReadHeavyCache struct {
    mu    sync.RWMutex
    items map[string]string
}

// 읽기: 여러 고루틴이 동시에 RLock 가능
func (c *ReadHeavyCache) Get(key string) (string, bool) {
    c.mu.RLock()         // 읽기 잠금
    defer c.mu.RUnlock() // 읽기 잠금 해제
    v, ok := c.items[key]
    return v, ok
}

// 쓰기: 하나만 가능, 모든 읽기도 불가
func (c *ReadHeavyCache) Set(key, value string) {
    c.mu.Lock()         // 쓰기 잠금 (독점)
    defer c.mu.Unlock()
    c.items[key] = value
}
```

### RWMutex 동작 규칙

| 잠금 상태 | 새 RLock | 새 Lock |
|----------|---------|---------|
| 없음 | 허용 | 허용 |
| RLock 1개 이상 | 허용 | 대기 |
| Lock 1개 | 대기 | 대기 |

읽기 99%, 쓰기 1% 상황에서 Mutex 대신 RWMutex를 쓰면 성능이 크게 향상됩니다.

---

## sync.Once: 일회성 초기화

프로그램 전체에서 딱 한 번만 실행을 보장합니다. 싱글톤, 지연 초기화에 사용합니다.

```go
var (
    instance *Database
    once     sync.Once
)

func GetDB() *Database {
    once.Do(func() {
        // 이 함수는 프로그램 전체에서 딱 한 번만 실행
        // 동시에 여러 고루틴이 호출해도 한 번만 실행됨
        fmt.Println("DB 연결 초기화 중...")
        instance = connectToDB()
    })
    return instance
}

// 10개 고루틴이 동시에 호출해도 "DB 연결 초기화 중..."은 한 번만 출력
for i := 0; i < 10; i++ {
    go func() {
        db := GetDB() // 안전
        _ = db
    }()
}
```

### Once vs init()

```go
// init(): 패키지 로드 시 항상 실행 (지연 불가)
func init() {
    db = connectToDB() // 앱 시작 시 항상 실행
}

// sync.Once: 처음 사용 시 초기화 (지연 가능, 더 유연)
var once sync.Once
func getDB() *Database {
    once.Do(func() { db = connectToDB() }) // 처음 사용 시 초기화
    return db
}
```

---

## sync.WaitGroup 상세

카운터 기반으로 고루틴 완료를 대기합니다.

```go
var wg sync.WaitGroup

// Add(delta): 카운터를 delta만큼 증가
// Done(): 카운터를 1 감소 (= Add(-1))
// Wait(): 카운터가 0이 될 때까지 블로킹

wg.Add(3) // 3개의 고루틴 시작 예정
for i := 0; i < 3; i++ {
    go func(n int) {
        defer wg.Done()
        work(n)
    }(i)
}
wg.Wait() // 3개 모두 Done() 호출 후 반환
```

### sync.WaitGroup.Go() — Go 1.25 신규

Go 1.25(Aug 2025)에서 `WaitGroup`에 `.Go(func())` 메서드가 추가되었습니다. `Add(1)` + `go func() { defer Done() }()` 관용구를 단일 호출로 대체합니다.

```go
var wg sync.WaitGroup

// 기존 방식
wg.Add(1)
go func() {
    defer wg.Done()
    work()
}()

// Go 1.25+ 방식: 동일한 동작을 한 줄로
wg.Go(func() {
    work()
})

wg.Wait()
```

### sync.Map 내부 구현 — Swiss Table (Go 1.24)

Go 1.24(Feb 2025)부터 Go 런타임의 맵 구현이 **Swiss Table** 알고리즘으로 교체되었습니다. 기존 체이닝 해시 테이블 대비 캐시 효율이 높아 조회/삽입 성능이 평균 수십% 향상됩니다. API는 동일하며 코드 변경 없이 자동으로 적용됩니다.

---

## sync.Pool: 객체 재사용

자주 할당/해제되는 임시 객체를 재사용하여 GC 부담을 줄입니다.

```go
var bufPool = sync.Pool{
    New: func() interface{} {
        // 풀에 객체가 없을 때만 호출됨
        return &bytes.Buffer{}
    },
}

func processRequest(data []byte) string {
    // 풀에서 버퍼 획득 (없으면 New 호출)
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()        // 상태 초기화 (중요!)
        bufPool.Put(buf)   // 풀에 반납
    }()

    buf.Write(data)
    // ... 처리 ...
    return buf.String()
}
```

### Pool 주의사항

```go
// 주의 1: GC 시 풀이 비워질 수 있음 → 캐시로만 사용, 영구 저장 금지
pool := sync.Pool{New: func() interface{} { return new(MyObj) }}
obj := pool.Get().(*MyObj)
// GC 후에는 pool이 비워짐 → 다음 Get은 New를 다시 호출

// 주의 2: 반납 전 반드시 상태 초기화
buf.Reset()   // bytes.Buffer
s.Reset()     // scanner 등
*obj = MyObj{} // 구조체 초기화

// 주의 3: 풀 객체를 여러 곳에서 동시 사용하지 않음
obj := pool.Get().(*MyObj)
go use(obj)  // 위험! obj를 다른 곳에서도 Get으로 받을 수 있음
pool.Put(obj) // Put 후에는 손대지 않음
```

표준 라이브러리에서의 Pool 사용: `fmt` 패키지 내부 버퍼, `encoding/json` 등.

---

## sync.Map: 동시성 안전 맵

`sync.Map`은 특정 패턴에서만 `Mutex + map` 조합보다 효율적입니다.

```go
var sm sync.Map

// Store: 저장
sm.Store("key", 42)

// Load: 읽기
if val, ok := sm.Load("key"); ok {
    fmt.Println(val.(int))
}

// LoadOrStore: 없으면 저장, 있으면 기존 값 반환
actual, loaded := sm.LoadOrStore("key", 100)
// loaded=true: 이미 있었음, actual=42 (기존 값)
// loaded=false: 새로 저장됨, actual=100

// Delete: 삭제
sm.Delete("key")

// Range: 순회 (순서 보장 없음)
sm.Range(func(key, value interface{}) bool {
    fmt.Println(key, value)
    return true // false 반환 시 순회 중단
})
```

### sync.Map vs Mutex+map

`sync.Map`이 유리한 경우:
- 키가 한 번 쓰이고 여러 번 읽힐 때 (예: 캐시)
- 여러 고루틴이 서로 다른 키를 읽고 쓸 때

`Mutex+map`이 유리한 경우:
- 쓰기가 잦을 때
- 타입 안전성이 중요할 때 (sync.Map은 `interface{}`)
- 맵 연산 외 로직과 함께 잠금이 필요할 때

---

## sync.Cond: 조건 변수

특정 조건이 만족될 때까지 대기하다가 신호를 받으면 깨어납니다.

```go
type BoundedQueue struct {
    mu      sync.Mutex
    cond    *sync.Cond
    items   []int
    maxSize int
}

func NewBoundedQueue(max int) *BoundedQueue {
    q := &BoundedQueue{maxSize: max}
    q.cond = sync.NewCond(&q.mu)
    return q
}

func (q *BoundedQueue) Push(item int) {
    q.mu.Lock()
    defer q.mu.Unlock()

    for len(q.items) >= q.maxSize {
        q.cond.Wait() // 잠금 해제 후 대기, 신호 오면 잠금 재획득
    }
    q.items = append(q.items, item)
    q.cond.Broadcast() // 대기 중인 모든 고루틴에게 신호
}

func (q *BoundedQueue) Pop() int {
    q.mu.Lock()
    defer q.mu.Unlock()

    for len(q.items) == 0 {
        q.cond.Wait()
    }
    item := q.items[0]
    q.items = q.items[1:]
    q.cond.Broadcast()
    return item
}
```

`Signal()`은 대기 중인 고루틴 하나에게, `Broadcast()`는 전부에게 신호를 보냅니다. 일반적으로 `Broadcast`가 더 안전합니다.

---

## sync/atomic: 원자적 연산

단일 메모리 위치에 대한 원자적 연산입니다. 뮤텍스보다 빠르지만 복잡한 임계 구역에는 사용 불가합니다.

```go
import "sync/atomic"

// Go 1.19+ 타입 기반 atomic (권장)
var counter atomic.Int64
counter.Add(1)
counter.Load()
counter.Store(42)
counter.Swap(100)
counter.CompareAndSwap(100, 200) // 100이면 200으로 교체

var flag atomic.Bool
flag.Store(true)
flag.Load()

// 구버전 함수 기반 (하위 호환)
var n int64
atomic.AddInt64(&n, 1)
atomic.LoadInt64(&n)
atomic.StoreInt64(&n, 42)

// CAS (Compare-And-Swap): 낙관적 잠금 구현에 사용
old := atomic.LoadInt64(&n)
if atomic.CompareAndSwapInt64(&n, old, old+1) {
    // 성공: n이 old였고 old+1로 변경됨
} else {
    // 실패: 다른 고루틴이 먼저 변경함 → 재시도
}
```

### atomic.Value: 임의 타입 원자적 저장

```go
// 설정을 원자적으로 교체 (서버 설정 hot reload에 유용)
var config atomic.Value

type Config struct {
    MaxConn int
    Debug   bool
}

config.Store(Config{MaxConn: 100, Debug: false})

// 읽기 (타입 어서션 필요)
cfg := config.Load().(Config)
fmt.Println(cfg.MaxConn)

// 새 설정으로 원자적 교체
config.Store(Config{MaxConn: 200, Debug: true})
```

---

## 레이스 컨디션

두 고루틴이 동기화 없이 같은 메모리에 동시 접근하고, 하나 이상이 쓰기인 상태.

```go
// 레이스 컨디션 예시
var counter int
go func() { counter++ }() // 읽기-증가-쓰기 (3단계, 비원자적)
go func() { counter++ }() // 동시 실행 시 결과 불확실
```

감지 방법: `go run -race main.go` 또는 `go test -race ./...`

레이스 감지기 출력:
```
WARNING: DATA RACE
Write at 0x... by goroutine 7:
  main.main.func1()
    main.go:5
Read at 0x... by goroutine 8:
  main.main.func2()
    main.go:6
```

---

## Python/Java 비교

### Python threading.Lock vs sync.Mutex

```python
import threading
lock = threading.Lock()

with lock:          # 자동 acquire/release
    counter += 1
```

```go
var mu sync.Mutex
mu.Lock()
defer mu.Unlock()
counter++
```

Python의 `with lock`이 Go의 `defer mu.Unlock()`과 같은 역할입니다.

### Java ConcurrentHashMap vs sync.Map

```java
ConcurrentHashMap<String, Integer> map = new ConcurrentHashMap<>();
map.put("key", 42);                    // sm.Store("key", 42)
map.getOrDefault("key", 0);            // sm.Load("key")
map.putIfAbsent("key", 100);           // sm.LoadOrStore("key", 100)
```

Java의 `ConcurrentHashMap`은 타입 안전하지만 `sync.Map`은 `interface{}`를 사용합니다. 제네릭 기반 타입 안전한 동시성 맵이 필요하면 `Mutex + map`을 사용하세요.
