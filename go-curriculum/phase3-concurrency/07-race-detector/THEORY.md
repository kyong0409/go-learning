# 07. 레이스 디텍터 (Race Detector)

## 데이터 레이스란?

두 고루틴이 **동기화 없이** 같은 메모리 위치에 동시에 접근하고, 그 중 **하나 이상이 쓰기**인 상황입니다.

```go
var counter int

go func() { counter++ }() // 쓰기: 읽기(load) + 증가(add) + 쓰기(store)
go func() { counter++ }() // 동시 쓰기 → 데이터 레이스!

// 읽기-읽기는 레이스가 아님 (항상 안전)
go func() { fmt.Println(counter) }()
go func() { fmt.Println(counter) }()
```

데이터 레이스의 위험성:
- 결과가 비결정적 (실행마다 다른 결과)
- 가끔만 발생 (재현이 어려움)
- 메모리 손상, 잘못된 결과, 크래시 유발
- 디버깅이 매우 어려움 (스택 트레이스가 없을 수 있음)

---

## Go 레이스 디텍터

Go 런타임에 내장된 동적 분석 도구입니다. ThreadSanitizer(TSan) 알고리즘을 기반으로 합니다.

```bash
# 실행 시 레이스 감지
go run -race main.go

# 테스트 시 레이스 감지 (가장 흔한 사용법)
go test -race ./...

# 빌드 시 포함 (성능 모니터링 서버 등)
go build -race -o myapp ./...
```

CI/CD 파이프라인에서 `go test -race`를 항상 실행하는 것을 권장합니다.

---

## 레이스 디텍터 출력 읽는 법

```
==================
WARNING: DATA RACE
Write at 0x00c000122048 by goroutine 7:         ← 쓰기 위치와 고루틴 번호
  main.incrementRacy()
    /path/to/main.go:35 +0x34                   ← 소스 파일과 라인

Previous read at 0x00c000122048 by goroutine 8: ← 이전 읽기 위치
  main.incrementRacy()
    /path/to/main.go:35 +0x28

Goroutine 7 (running) created at:              ← 고루틴이 생성된 위치
  main.main()
    /path/to/main.go:48 +0x84

Goroutine 8 (running) created at:
  main.main()
    /path/to/main.go:48 +0x84
==================
```

읽는 순서:
1. `WARNING: DATA RACE` - 레이스 감지됨
2. `Write at` - 어디서 쓰기가 발생했는지
3. `Previous read/write at` - 충돌한 이전 접근
4. `Goroutine N created at` - 문제의 고루틴이 어디서 만들어졌는지

---

## 일반적인 레이스 패턴과 해결법

### 패턴 1: 보호되지 않은 카운터

```go
// 레이스 발생
var counter int
go func() { counter++ }()
go func() { counter++ }()

// 해결 1: sync/atomic
var counter int64
go func() { atomic.AddInt64(&counter, 1) }()
go func() { atomic.AddInt64(&counter, 1) }()

// 해결 2: sync.Mutex
var mu sync.Mutex
var counter int
go func() { mu.Lock(); counter++; mu.Unlock() }()
go func() { mu.Lock(); counter++; mu.Unlock() }()

// 해결 3: 채널 (Go 1.19+의 atomic.Int64 권장)
var counter atomic.Int64
go func() { counter.Add(1) }()
go func() { counter.Add(1) }()
```

### 패턴 2: 보호되지 않은 맵 접근

```go
// 레이스 발생: 맵은 동시 접근에 안전하지 않음
m := make(map[string]int)
go func() { m["key"] = 1 }()  // 쓰기
go func() { _ = m["key"] }()  // 읽기 (동시 접근!)

// Go 런타임은 동시 맵 접근을 감지하면 즉시 패닉:
// "concurrent map read and map write"

// 해결 1: sync.Mutex + map
type SafeMap struct {
    mu sync.RWMutex
    m  map[string]int
}
func (s *SafeMap) Get(k string) int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.m[k]
}
func (s *SafeMap) Set(k string, v int) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.m[k] = v
}

// 해결 2: sync.Map (특정 패턴에서 유리)
var sm sync.Map
go func() { sm.Store("key", 1) }()
go func() { sm.Load("key") }()
```

### 패턴 3: 클로저 루프 변수 캡처 (Go 1.21 이전)

```go
// 레이스 발생: results[i] 접근이 동시에 이루어짐
results := make([]int, 5)
var wg sync.WaitGroup
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        results[i] = i * i // i가 레이스, results[i]도 레이스!
    }()
}

// 해결: 인자로 i 전달, 각 고루틴이 독립 인덱스 사용
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(idx int) {
        defer wg.Done()
        results[idx] = idx * idx // idx는 각 고루틴의 복사본
    }(i)
}
```

### 패턴 4: 구조체 필드 동시 접근

```go
type Stats struct {
    Requests int
    Errors   int
}

// 레이스: 여러 고루틴이 같은 구조체 필드에 동시 쓰기
stats := &Stats{}
go func() { stats.Requests++ }()
go func() { stats.Errors++ }()

// 해결 1: 뮤텍스로 구조체 보호
type SafeStats struct {
    mu       sync.Mutex
    Requests int
    Errors   int
}
func (s *SafeStats) AddRequest() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.Requests++
}

// 해결 2: atomic 필드 사용
type AtomicStats struct {
    Requests atomic.Int64
    Errors   atomic.Int64
}
stats.Requests.Add(1)
stats.Errors.Add(1)
```

### 패턴 5: 슬라이스 동시 append

```go
// 레이스: append는 스레드 안전하지 않음
var results []int
go func() { results = append(results, 1) }()
go func() { results = append(results, 2) }()

// 해결 1: 뮤텍스
var mu sync.Mutex
var results []int
go func() {
    mu.Lock()
    results = append(results, 1)
    mu.Unlock()
}()

// 해결 2: 채널로 수집
ch := make(chan int, 2)
go func() { ch <- 1 }()
go func() { ch <- 2 }()
results := []int{<-ch, <-ch}

// 해결 3: 인덱스가 독립적이면 직접 접근
results := make([]int, 5)
for i := 0; i < 5; i++ {
    go func(idx int) {
        results[idx] = idx * idx // 다른 인덱스 → 레이스 없음
    }(i)
}
```

---

## 레이스 디텍터 동작 원리

ThreadSanitizer(TSan) 알고리즘:

1. **계측(Instrumentation)**: 컴파일 시 모든 메모리 읽기/쓰기에 추적 코드 삽입
2. **벡터 시계(Vector Clock)**: 각 고루틴과 메모리 위치의 접근 순서를 추적
3. **happens-before 관계**: 동기화 연산(Mutex, 채널 등)으로 설정된 순서 관계 추적
4. **레이스 감지**: happens-before 관계 없이 동시 접근 시 경고

레이스가 발생해도 즉시 크래시되지 않고 보고서를 출력한 후 계속 실행됩니다.

---

## 성능 영향

| 항목 | 영향 |
|------|------|
| 실행 속도 | 5~10배 느려짐 |
| 메모리 사용 | 5~10배 증가 |
| 이진 파일 크기 | 증가 |

이 때문에 `-race`는 개발/테스트 환경에서만 사용하고, 프로덕션 빌드에는 포함하지 않습니다.

---

## CI/CD에서 레이스 감지 설정

### GitHub Actions 예시

```yaml
- name: Test with race detector
  run: go test -race -timeout 60s ./...
```

### Makefile 예시

```makefile
test:
    go test ./...

test-race:
    go test -race ./...

ci: test-race
```

---

## 레이스 없는 Go 코드 작성 체크리스트

```
□ 공유 변수에 접근하는 모든 고루틴에 동기화 적용
□ 맵: 동시 접근 시 sync.Mutex 또는 sync.Map 사용
□ 슬라이스: 동시 append 시 뮤텍스 또는 채널 수집
□ 구조체 필드: 동시 수정 시 뮤텍스 또는 atomic
□ 루프 변수: Go 1.22+는 자동 수정됨; 명시적 인자 전달이 더 명확하므로 권장
□ WaitGroup.Add(): 고루틴 시작 전에 호출
□ CI에서 go test -race ./... 실행
```

---

## Python/Java 비교

### Python threading 레이스

Python의 GIL은 CPython에서 단순 연산에 대해 레이스를 방지해주지만, 복잡한 연산이나 PyPy에서는 여전히 레이스가 발생합니다.

```python
import threading

counter = 0
def increment():
    global counter
    for _ in range(1000):
        counter += 1  # Python에서는 GIL이 보호해줌 (일반적으로)
                      # 하지만 이에 의존하면 안 됨

threads = [threading.Thread(target=increment) for _ in range(10)]
for t in threads: t.start()
for t in threads: t.join()
```

Go에는 GIL이 없으므로 모든 공유 상태를 명시적으로 보호해야 합니다.

### Java synchronized vs Go sync.Mutex

```java
// Java: synchronized 메서드/블록
public synchronized void increment() {
    counter++;
}

// 또는
synchronized (this) {
    counter++;
}
```

```go
// Go: 명시적 Mutex
func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}
```

Java의 `synchronized`와 Go의 `sync.Mutex`는 역할이 같지만, Go는 `defer`로 해제를 보장하는 패턴이 관용적입니다.

### Java ThreadSanitizer vs Go -race

Java에는 내장 레이스 디텍터가 없고, 외부 도구(FindBugs, SpotBugs, Helgrind 등)를 사용해야 합니다. Go는 `go test -race` 하나로 해결됩니다.
