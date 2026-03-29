# 01. 고루틴 (Goroutine)

## 고루틴이란?

고루틴은 Go 런타임이 관리하는 **경량 실행 단위**입니다. OS 스레드가 아닙니다.

- OS 스레드: 운영체제가 직접 스케줄링, 초기 스택 1~8MB
- 고루틴: Go 런타임이 스케줄링, 초기 스택 2~8KB (동적 성장/축소)

```go
// 고루틴 시작: go 키워드 하나로 끝
go doWork()          // 기존 함수를 고루틴으로 실행
go func() {          // 익명 함수를 고루틴으로 실행
    fmt.Println("hello")
}()
```

---

## 고루틴의 메모리 특성

| 항목 | OS 스레드 | 고루틴 |
|------|----------|--------|
| 초기 스택 크기 | 1~8MB (고정) | 2~8KB (동적) |
| 최대 스택 크기 | OS 제한 (~수십 MB) | 기본 1GB (변경 가능) |
| 컨텍스트 전환 비용 | 높음 (커널 모드 전환) | 낮음 (사용자 공간) |
| 생성 가능 수 | 수천 개 (메모리 제한) | 수백만 개 |

스택은 필요에 따라 Go 런타임이 자동으로 늘리고 줄입니다. Python/Java 개발자가 스레드 풀을 미리 생성하는 이유가 바로 이 비용 때문인데, Go에서는 필요할 때마다 고루틴을 만들어도 됩니다.

---

## GMP 스케줄러 모델

Go 런타임은 M:N 스케줄러를 사용합니다.

```
G (Goroutine)  - 실행할 작업 단위 (수백만 개 가능)
M (Machine)    - OS 스레드 (GOMAXPROCS 수만큼 활성)
P (Processor)  - 스케줄러 컨텍스트 + 로컬 G 큐
```

```
P0의 로컬 큐: [G1, G2, G3]
P1의 로컬 큐: [G4, G5]
글로벌 큐:    [G6, G7, G8]

P0 → M0 → CPU 코어 0  (G1 실행 중)
P1 → M1 → CPU 코어 1  (G4 실행 중)
```

**Work Stealing**: P0의 큐가 비면 P1의 큐에서 절반을 가져옵니다. 로드 밸런싱이 자동으로 됩니다.

### GOMAXPROCS

```go
import "runtime"

// 현재 값 확인 (기본: CPU 코어 수)
fmt.Println(runtime.GOMAXPROCS(0))

// 변경 (실험/테스트용)
runtime.GOMAXPROCS(1) // 단일 스레드로 제한
```

환경 변수로도 설정 가능: `GOMAXPROCS=4 go run main.go`

### 스케줄링 포인트

고루틴이 CPU를 양보하는 시점:

```go
// 1. 채널 연산 (블로킹)
ch <- val   // 수신자 없으면 양보
<-ch        // 송신자 없으면 양보

// 2. 시스템 콜 (I/O, 파일, 네트워크)
os.ReadFile("data.txt") // 내부적으로 양보

// 3. 명시적 양보
runtime.Gosched() // "나 잠깐 쉴게, 다른 고루틴 실행해"

// 4. 함수 호출 (Go 1.14+ 선점형 스케줄링)
// 긴 루프도 자동으로 선점될 수 있음
```

---

## sync.WaitGroup: 고루틴 완료 대기

고루틴은 비동기로 실행되므로 main이 먼저 종료되면 모든 고루틴이 강제 종료됩니다.

```go
// 나쁜 방법: time.Sleep
go doWork()
time.Sleep(1 * time.Second) // 1초면 충분할까? 모름!

// 좋은 방법: sync.WaitGroup
var wg sync.WaitGroup

wg.Add(1)           // 카운터 +1 (고루틴 시작 전에 호출!)
go func() {
    defer wg.Done() // 카운터 -1 (반드시 호출 보장: defer 사용)
    doWork()
}()

wg.Wait() // 카운터가 0이 될 때까지 블로킹
```

### WaitGroup 사용 규칙

```go
var wg sync.WaitGroup

// 규칙 1: Add는 반드시 고루틴 시작 전에 호출
for i := 0; i < 5; i++ {
    wg.Add(1)  // ← 여기에 (for 루프 안, go 키워드 전)
    go func(n int) {
        defer wg.Done()
        process(n)
    }(i)
}

// 규칙 2: WaitGroup을 값으로 복사하지 않음
// 포인터로 전달하거나 클로저로 캡처
func worker(wg *sync.WaitGroup) { // 포인터로 전달
    defer wg.Done()
}
```

---

## 클로저와 고루틴: 변수 캡처 함정

### 루프 변수 캡처 문제 (Go 1.22 이전)

```go
// 잘못된 코드: 모든 고루틴이 같은 i를 참조 (Go 1.21 이하에서 문제)
for i := 0; i < 5; i++ {
    go func() {
        fmt.Println(i) // i는 루프 변수를 직접 참조
        // 루프 종료 후 실행되면 모두 5를 출력할 수 있음!
    }()
}

// 올바른 방법 1: 인자로 복사 (모든 버전에서 명시적이고 권장됨)
for i := 0; i < 5; i++ {
    go func(n int) { // n은 i의 복사본
        fmt.Println(n)
    }(i) // 현재 i 값을 즉시 전달
}

// 올바른 방법 2: 로컬 변수로 복사 (Go 1.22 이전 관용구)
for i := 0; i < 5; i++ {
    i := i // 루프 변수를 새 변수로 섀도잉
    go func() {
        fmt.Println(i)
    }()
}
```

**Go 1.22(Feb 2024)부터** `for` 루프 변수가 매 반복마다 새로 생성되어 이 문제가 해결됩니다. 하지만 명시적 인자 전달이 더 명확하므로 여전히 권장됩니다.

### sync.WaitGroup.Go() — Go 1.25 신규

Go 1.25(Aug 2025)에서 `sync.WaitGroup`에 `.Go()` 메서드가 추가되었습니다. `Add(1)` + `go func() { defer Done() ... }()` 패턴을 한 줄로 줄여줍니다.

```go
var wg sync.WaitGroup

// 기존 방식 (Go 1.24 이하)
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        process(n)
    }(i)
}

// 새 방식 (Go 1.25+): wg.Go()가 Add(1) + 고루틴 시작 + defer Done()을 처리
for i := 0; i < 5; i++ {
    wg.Go(func() {
        process(i) // Go 1.22+이므로 루프 변수 캡처 문제 없음
    })
}

wg.Wait()
```

---

## 고루틴 누수 (Goroutine Leak)

고루틴이 종료되지 않고 계속 살아있는 상태를 누수라고 합니다. 메모리 증가, 프로그램 성능 저하로 이어집니다.

### 누수 원인 패턴

```go
// 원인 1: 아무도 받지 않는 채널에 보내려고 영원히 블로킹
func leak1() {
    ch := make(chan int)
    go func() {
        ch <- 42 // 수신자 없음 → 영원히 블로킹 → 누수!
    }()
    // ch를 반환하거나 닫지 않음
}

// 원인 2: 아무도 보내지 않는 채널에서 영원히 기다림
func leak2() {
    ch := make(chan int)
    go func() {
        val := <-ch // 송신자 없음 → 영원히 블로킹 → 누수!
        fmt.Println(val)
    }()
    // ch를 닫거나 값을 보내지 않음
}
```

### 누수 방지: done 채널 패턴

```go
func noLeak() {
    done := make(chan struct{}) // 취소 신호용 채널

    go func() {
        select {
        case <-done:
            // 취소 신호 수신 → 정상 종료
            return
        case <-time.After(5 * time.Second):
            // 타임아웃 방어
            return
        }
    }()

    // 작업 완료 후 고루틴 종료 신호
    close(done) // 모든 수신자에게 브로드캐스트
}
```

### 누수 방지: context 사용 (권장)

```go
func withContext(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return // context 취소 시 정상 종료
            default:
                doWork()
            }
        }
    }()
}
```

### 고루틴 수 모니터링

```go
import "runtime"

fmt.Println(runtime.NumGoroutine()) // 현재 실행 중인 고루틴 수
```

---

## 경량성 데모: 10만 개 고루틴

```go
var wg sync.WaitGroup

start := time.Now()
for i := 0; i < 100_000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        time.Sleep(1 * time.Millisecond)
    }()
}
wg.Wait()

fmt.Printf("10만 고루틴 완료: %v\n", time.Since(start))
// 결과: 수백ms (Java 스레드 10만 개? 수GB 메모리, 생성 자체가 불가)
```

---

## Python/Java 개발자를 위한 비교

### Python threading과의 차이

```python
# Python: GIL로 인해 CPU 바운드는 실제 병렬 실행 안 됨
import threading
t = threading.Thread(target=cpu_intensive)
t.start()
t.join()  # join() = Go의 wg.Wait()
```

```go
// Go: 진정한 병렬 실행
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    cpuIntensive()
}()
wg.Wait()
```

### Java Thread와의 차이

```java
// Java: 스레드 생성 비용 큼, 풀 관리 필요
ExecutorService pool = Executors.newFixedThreadPool(10);
pool.submit(() -> doWork());
pool.shutdown();
pool.awaitTermination(1, TimeUnit.MINUTES);
```

```go
// Go: 고루틴은 저렴하므로 필요할 때 바로 생성
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()
wg.Wait()
// 워커 풀이 필요한 경우도 있지만 (06-patterns 참고),
// 기본적으로 고루틴을 아끼지 않아도 됩니다.
```

---

## 자주 발생하는 실수 정리

```go
// 실수 1: WaitGroup을 복사로 전달
func bad(wg sync.WaitGroup) { // 복사본이므로 원본에 영향 없음!
    defer wg.Done()
}
func good(wg *sync.WaitGroup) { // 포인터로 전달
    defer wg.Done()
}

// 실수 2: Add를 고루틴 안에서 호출
go func() {
    wg.Add(1) // 위험! 고루틴이 시작되기 전에 Wait()가 끝날 수 있음
    defer wg.Done()
}()

// 실수 3: Done을 defer 없이 호출 (패닉 시 Done 호출 안 됨)
go func() {
    doRiskyWork() // 패닉 발생 시
    wg.Done()     // 이 줄이 실행 안 됨 → Wait()가 영원히 블로킹
}()
// 올바른 방법:
go func() {
    defer wg.Done() // 패닉, return 등 모든 경우에 보장
    doRiskyWork()
}()
```
