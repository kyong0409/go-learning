# Phase 3: Go 동시성 (Concurrency) - 학습 가이드

## 학습 목표

이 단계에서는 Go의 가장 강력한 특징인 동시성 모델을 완전히 이해하고 실용적으로 활용하는 능력을 기릅니다.

- 고루틴(goroutine)의 내부 동작 원리와 GMP 스케줄러 이해
- 채널(channel)을 통한 고루틴 간 안전한 통신
- `select` 문으로 여러 채널 연산 다중화
- `sync` 패키지로 공유 상태 보호
- `context` 패키지로 고루틴 생명주기 관리
- 실무에서 자주 쓰이는 동시성 패턴 구현
- 레이스 컨디션 감지 및 수정

---

## 핵심 철학: "동시성은 병렬성이 아니다"

Rob Pike(Go 설계자)의 유명한 말입니다.

- **동시성(Concurrency)**: 여러 작업을 *구조적으로* 독립적으로 다루는 능력. 단일 코어에서도 가능.
- **병렬성(Parallelism)**: 여러 작업을 *물리적으로* 동시에 실행하는 것. 여러 코어가 필요.

Go는 동시성을 언어 수준에서 지원하며, 병렬 실행은 런타임이 알아서 처리합니다.

```
동시성: 여러 대화를 번갈아가며 하는 사람 (한 명)
병렬성: 여러 대화를 동시에 하는 여러 명의 사람
```

---

## CSP (Communicating Sequential Processes) 모델

Go의 동시성 모델은 1978년 Tony Hoare가 제안한 CSP 이론에 기반합니다.

핵심 아이디어:
- 독립적인 프로세스들이 **채널을 통해 메시지를 주고받으며** 협력
- 공유 메모리 직접 접근을 최소화
- "메모리를 공유해서 통신하지 말고, **통신해서 메모리를 공유하라**"

```go
// CSP 스타일: 채널로 통신
ch := make(chan int)
go func() { ch <- 42 }()  // 생산자
val := <-ch               // 소비자

// 전통적 스타일: 공유 메모리 + 뮤텍스
var mu sync.Mutex
var shared int
go func() { mu.Lock(); shared = 42; mu.Unlock() }()
```

두 방식 모두 Go에서 지원되며, 상황에 따라 적합한 방식을 선택합니다.

---

## Go 런타임의 M:N 스케줄러 (GMP 모델)

Go 런타임은 M:N 스레드 모델을 사용합니다. M개의 고루틴을 N개의 OS 스레드에 매핑합니다.

### GMP 구성 요소

| 구성 요소 | 영문 | 역할 |
|----------|------|------|
| **G** | Goroutine | 실행할 코드, 스택, 상태를 담은 경량 실행 단위 |
| **M** | Machine | OS 스레드. 실제 CPU에서 코드를 실행 |
| **P** | Processor | 스케줄러 컨텍스트. G의 로컬 실행 큐를 보유 |

```
[P0] --- [M0] --- CPU 코어 0
 |
[G1, G2, G3, ...]  ← G의 로컬 큐

[P1] --- [M1] --- CPU 코어 1
 |
[G4, G5, G6, ...]

                   [글로벌 큐: G7, G8, ...]
```

### GOMAXPROCS

`GOMAXPROCS`는 동시에 실행될 수 있는 P(프로세서)의 수를 결정합니다.

```go
import "runtime"

// 기본값: runtime.NumCPU() (CPU 코어 수)
runtime.GOMAXPROCS(4) // P를 4개로 설정

// 현재 값 확인 (0 전달 시 변경 없이 현재 값 반환)
current := runtime.GOMAXPROCS(0)
```

> **Go 1.25(Aug 2025)**: 컨테이너 환경에서 `GOMAXPROCS`가 cgroup CPU 쿼터를 자동으로 인식합니다. Kubernetes Pod에서 CPU limit을 설정하면 Go 런타임이 그에 맞춰 GOMAXPROCS를 자동 조정합니다. 기존에는 `uber-go/automaxprocs` 같은 외부 라이브러리가 필요했습니다.

### 스케줄링 포인트

고루틴이 다른 고루틴에게 실행 기회를 양보하는 시점:
- 채널 연산 (블로킹)
- 시스템 콜 (파일 I/O, 네트워크 등)
- `runtime.Gosched()` 명시적 양보
- 함수 호출 (Go 1.14+ 선점형 스케줄링)
- GC 관련 작업

---

## Python GIL vs Java 스레드 vs Go 고루틴 비교

| 항목 | Python (threading) | Java (Thread) | Go (goroutine) |
|------|-------------------|---------------|----------------|
| **초기 스택** | OS 기본 (~8MB) | ~512KB~1MB | ~2KB (동적 성장) |
| **생성 비용** | 높음 (OS 스레드) | 높음 (OS 스레드) | 매우 낮음 |
| **동시 실행 수** | 제한적 (GIL) | 수천 개 | 수백만 개 |
| **병렬 실행** | GIL로 제한 (I/O만) | 진정한 병렬 | 진정한 병렬 |
| **통신 방식** | Queue, Lock | wait/notify | 채널 (CSP) |
| **컨텍스트 전환** | OS 수준 | OS 수준 | 런타임 수준 (빠름) |
| **스케줄링** | OS 선점형 | OS 선점형 | Go 런타임 (협력+선점) |

### Python과의 주요 차이

```python
# Python: GIL로 인해 CPU 바운드 작업은 진정한 병렬성 없음
import threading
threads = [threading.Thread(target=cpu_work) for _ in range(4)]
# 실제로는 하나씩 실행됨 (GIL)
```

```go
// Go: 진정한 병렬 실행 (GOMAXPROCS 개수만큼 동시 실행)
for i := 0; i < 4; i++ {
    go cpuWork() // 실제로 병렬 실행됨
}
```

### Java와의 주요 차이

```java
// Java: 스레드 생성 비용이 크고, 수천 개가 한계
Thread t = new Thread(() -> doWork());
t.start(); // OS 스레드 직접 생성
```

```go
// Go: 수백만 개도 가능한 고루틴
go doWork() // 경량 고루틴 생성
```

---

## Phase 3 학습 순서와 의존 관계

```
01-goroutines  ─────────────────────────────┐
    │                                        │
    ▼                                        │
02-channels ──────────────────────────┐     │
    │                                  │     │
    ▼                                  │     │
03-select ─────────────────────┐      │     │
    │                           │      │     │
    ▼                           ▼      ▼     ▼
04-sync          05-context ──────────────────
    │                │
    └────────────────┘
           │
           ▼
      06-patterns
           │
           ▼
    07-race-detector
```

각 단계별 핵심 개념:

1. **01-goroutines**: `go` 키워드, `sync.WaitGroup`, GMP 모델 이해
2. **02-channels**: `make(chan T)`, 언버퍼/버퍼 채널, 채널 방향성
3. **03-select**: 다중 채널 멀티플렉싱, 타임아웃, 논블로킹 연산
4. **04-sync**: 뮤텍스, RWMutex, Once, Pool, atomic
5. **05-context**: 취소 전파, 타임아웃, 요청 범위 값
6. **06-patterns**: 워커 풀, 팬아웃/팬인, 파이프라인, errgroup
7. **07-race-detector**: 레이스 컨디션 감지 및 수정

---

## 언제 채널을 쓰고 언제 뮤텍스를 쓸까?

Go 팀이 권장하는 기준:

| 상황 | 권장 방식 |
|------|----------|
| 소유권(데이터) 전달 | 채널 |
| 작업 배분 | 채널 |
| 결과 수집 | 채널 |
| 공유 캐시/상태 보호 | 뮤텍스 |
| 세마포어 | 버퍼 채널 |
| 일회성 초기화 | `sync.Once` |

---

## 추천 학습 자료

### 도서
- **Concurrency in Go** - Katherine Cox-Buday (O'Reilly, 2017)
  - Go 동시성 패턴의 바이블. 파이프라인, 팬인/팬아웃, 에러 처리 등 심층 다룸

### Go 공식 블로그
- [Go Concurrency Patterns](https://go.dev/blog/pipelines) - 파이프라인과 취소
- [Advanced Go Concurrency Patterns](https://go.dev/blog/advanced-go-concurrency-patterns)
- [Share Memory By Communicating](https://go.dev/blog/codelab-share)

### 동영상
- [Concurrency Is Not Parallelism - Rob Pike](https://www.youtube.com/watch?v=oV9rvDllKEg)
- [Go Concurrency Patterns - Rob Pike (Google I/O 2012)](https://www.youtube.com/watch?v=f6kdp27TYZs)

### 레퍼런스
- [The Go Memory Model](https://go.dev/ref/mem) - 공식 메모리 모델 명세
- [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync) - errgroup, semaphore 등 확장 패키지
