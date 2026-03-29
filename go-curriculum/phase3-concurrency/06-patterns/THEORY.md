# 06. 동시성 패턴 (Concurrency Patterns)

## 왜 패턴이 중요한가?

고루틴과 채널은 강력하지만, 잘못 사용하면 데드락, 누수, 레이스 컨디션이 발생합니다. 검증된 패턴을 재사용하면 안전하고 유지보수하기 쉬운 동시성 코드를 작성할 수 있습니다.

---

## 워커 풀 (Worker Pool)

고정된 수의 워커가 공유 작업 큐에서 작업을 가져와 처리합니다. 무제한 고루틴 생성을 방지하고 리소스를 제어합니다.

```
작업 채널 → [워커1]
           → [워커2]  → 결과 채널
           → [워커3]
```

```go
func workerPool(numWorkers int, jobs <-chan int) <-chan int {
    results := make(chan int, numWorkers)

    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for job := range jobs { // jobs 채널이 닫히면 루프 종료
                result := processJob(job)
                results <- result
            }
        }(i)
    }

    // 모든 워커가 완료되면 결과 채널 닫기
    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}

func main() {
    jobs := make(chan int, 100)
    for i := 0; i < 100; i++ {
        jobs <- i
    }
    close(jobs) // 더 이상 작업 없음

    results := workerPool(5, jobs) // 5개 워커

    for result := range results {
        fmt.Println(result)
    }
}
```

### context를 사용한 취소 가능한 워커 풀

```go
func workerPoolWithCtx(ctx context.Context, numWorkers int, jobs <-chan Job) <-chan Result {
    results := make(chan Result, numWorkers)

    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case job, ok := <-jobs:
                    if !ok {
                        return
                    }
                    select {
                    case results <- process(job):
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    return results
}
```

---

## 팬아웃 / 팬인 (Fan-out / Fan-in)

### 팬아웃: 하나의 입력을 여러 고루틴에 분배

```
           ┌→ [고루틴1]
입력 채널 ─┼→ [고루틴2] ─→ 각자 결과 채널
           └→ [고루틴3]
```

```go
func fanOut(in <-chan int, numWorkers int) []<-chan int {
    channels := make([]<-chan int, numWorkers)
    for i := 0; i < numWorkers; i++ {
        channels[i] = processAsync(in) // 각 워커가 in에서 읽음
    }
    return channels
}

func processAsync(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            out <- v * v
        }
    }()
    return out
}
```

### 팬인: 여러 채널의 결과를 하나로 합침

```
[채널1] ─┐
[채널2] ─┼→ 병합 채널
[채널3] ─┘
```

```go
func fanIn(channels ...<-chan int) <-chan int {
    merged := make(chan int)
    var wg sync.WaitGroup

    // 각 채널을 하나의 merged 채널로 합침
    output := func(ch <-chan int) {
        defer wg.Done()
        for v := range ch {
            merged <- v
        }
    }

    wg.Add(len(channels))
    for _, ch := range channels {
        go output(ch)
    }

    // 모든 입력이 완료되면 merged 닫기
    go func() {
        wg.Wait()
        close(merged)
    }()

    return merged
}

// 사용
ch1 := produce(1, 2, 3)
ch2 := produce(4, 5, 6)
ch3 := produce(7, 8, 9)

for v := range fanIn(ch1, ch2, ch3) {
    fmt.Println(v) // 순서는 불확정
}
```

---

## 파이프라인 (Pipeline)

여러 처리 단계를 채널로 연결합니다. 각 단계는 독립 고루틴으로 실행되어 병렬 처리됩니다.

```
생성 → 변환1 → 변환2 → 필터 → 수집
```

```go
// 각 단계는 <-chan T를 받아 <-chan T를 반환하는 패턴
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            out <- v * v
        }
    }()
    return out
}

func filter(in <-chan int, pred func(int) bool) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            if pred(v) {
                out <- v
            }
        }
    }()
    return out
}

// 파이프라인 조립: generate → square → filter
nums := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
squared := square(nums)
evens := filter(squared, func(n int) bool { return n%2 == 0 })

for v := range evens {
    fmt.Println(v) // 4, 16, 36, 64, 100
}
```

### context를 사용한 취소 가능한 파이프라인

```go
func squareCtx(ctx context.Context, in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            select {
            case out <- v * v:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}
```

---

## 속도 제한 (Rate Limiting)

초당 요청 수를 제한하여 외부 서비스 보호 또는 공정한 리소스 사용을 구현합니다.

### time.Ticker 기반 간단한 속도 제한

```go
// 초당 10회 제한
func rateLimited(requests <-chan Request) {
    limiter := time.NewTicker(100 * time.Millisecond) // 100ms마다 1회 = 초당 10회
    defer limiter.Stop()

    for req := range requests {
        <-limiter.C // 다음 tick까지 대기
        go handle(req)
    }
}
```

### 버스트 허용 속도 제한 (토큰 버킷)

```go
// 평소: 초당 5회 / 버스트: 최대 3회 즉시 처리
func burstRateLimiter() {
    // 버퍼 채널 = 토큰 버킷
    limiter := make(chan time.Time, 3) // 버스트 크기 3

    // 초당 5개 토큰 보충
    go func() {
        ticker := time.NewTicker(200 * time.Millisecond)
        defer ticker.Stop()
        for t := range ticker.C {
            select {
            case limiter <- t:
            default: // 버킷이 가득 차면 버림
            }
        }
    }()

    for req := range requests {
        <-limiter // 토큰 소비
        go handle(req)
    }
}
```

### golang.org/x/time/rate 패키지 (권장)

```go
import "golang.org/x/time/rate"

// 초당 10회, 버스트 최대 30회
limiter := rate.NewLimiter(rate.Limit(10), 30)

for _, req := range requests {
    // Wait: 허용될 때까지 블로킹
    if err := limiter.Wait(ctx); err != nil {
        return err
    }
    go handle(req)
}

// Allow: 즉시 확인 (블로킹 없음)
if limiter.Allow() {
    handle(req)
} else {
    rejectWithTooManyRequests()
}
```

---

## 세마포어 (Semaphore)

동시에 실행 중인 고루틴 수를 제한합니다. 버퍼 채널로 구현합니다.

```go
// N개까지 동시 실행 허용하는 세마포어
type Semaphore chan struct{}

func NewSemaphore(n int) Semaphore {
    return make(chan struct{}, n)
}

func (s Semaphore) Acquire() {
    s <- struct{}{} // 빈 슬롯이 있으면 즉시, 없으면 블로킹
}

func (s Semaphore) Release() {
    <-s // 슬롯 반환
}

// 사용
sem := NewSemaphore(5) // 동시에 최대 5개

var wg sync.WaitGroup
for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        sem.Acquire()         // 슬롯 획득
        defer sem.Release()   // 슬롯 반환
        fetch(u)
    }(url)
}
wg.Wait()
```

### golang.org/x/sync/semaphore (가중치 지원)

```go
import "golang.org/x/sync/semaphore"

// 총 가중치 10, 각 요청은 1~3 가중치 사용 가능
sem := semaphore.NewWeighted(10)

// 가중치 3 획득 (무거운 작업)
if err := sem.Acquire(ctx, 3); err != nil {
    return err
}
defer sem.Release(3)
doHeavyWork()
```

---

## errgroup: 에러 전파와 취소

`sync.WaitGroup` + 에러 처리 + 자동 취소를 통합한 패키지입니다.

```go
import "golang.org/x/sync/errgroup"

func fetchAll(ctx context.Context, urls []string) ([][]byte, error) {
    g, ctx := errgroup.WithContext(ctx)
    // g.Go()로 시작한 고루틴 중 하나가 에러를 반환하면
    // ctx가 취소되고 다른 고루틴들도 종료됩니다.

    results := make([][]byte, len(urls))

    for i, url := range urls {
        i, url := i, url // 루프 변수 캡처
        g.Go(func() error {
            resp, err := httpGetWithContext(ctx, url)
            if err != nil {
                return fmt.Errorf("URL %s 실패: %w", url, err)
            }
            results[i] = resp
            return nil
        })
    }

    // 모든 고루틴 완료 대기 + 첫 번째 에러 반환
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return results, nil
}
```

### errgroup.SetLimit: 동시 실행 수 제한

```go
g := new(errgroup.Group)
g.SetLimit(5) // 동시에 최대 5개 고루틴

for _, job := range jobs {
    job := job
    g.Go(func() error { // 5개가 실행 중이면 이 호출이 블로킹됨
        return process(job)
    })
}

if err := g.Wait(); err != nil {
    return err
}
```

---

## 패턴 비교 및 선택 가이드

| 패턴 | 사용 시점 | 핵심 구성 요소 |
|------|----------|--------------|
| 워커 풀 | CPU/IO 집약 작업 수 제한 | 작업 채널 + N개 고루틴 |
| 팬아웃 | 독립적 작업을 병렬 처리 | 하나의 채널 → N개 고루틴 |
| 팬인 | 여러 결과를 하나로 수집 | N개 채널 → 하나의 채널 |
| 파이프라인 | 단계적 변환, 스트리밍 처리 | 채널로 연결된 단계들 |
| 속도 제한 | API 쿼터, 공정한 처리 | Ticker + 토큰 버킷 |
| 세마포어 | 동시 접근 수 제한 | 버퍼 채널 |
| errgroup | 병렬 작업 + 에러 처리 | errgroup.Group |

---

## 실전 조합 예시: 병렬 웹 크롤러

```go
func crawl(ctx context.Context, urls []string) (map[string]string, error) {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // 최대 10개 동시 요청

    var mu sync.Mutex
    results := make(map[string]string)

    for _, url := range urls {
        url := url
        g.Go(func() error {
            body, err := fetchPage(ctx, url)
            if err != nil {
                return err
            }
            mu.Lock()
            results[url] = body
            mu.Unlock()
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return results, nil
}
```

---

## Python/Java 비교

### Python asyncio.gather vs errgroup

```python
import asyncio

async def fetch_all(urls):
    tasks = [fetch(url) for url in urls]
    results = await asyncio.gather(*tasks, return_exceptions=True)
    return results
```

```go
g, ctx := errgroup.WithContext(ctx)
for _, url := range urls {
    url := url
    g.Go(func() error { return fetch(ctx, url) })
}
err := g.Wait()
```

Go의 errgroup은 첫 번째 에러 발생 시 다른 고루틴도 취소하는 반면, Python의 `gather`는 모든 태스크가 완료(또는 에러)될 때까지 기다립니다.

### Java ExecutorService vs Worker Pool

```java
ExecutorService executor = Executors.newFixedThreadPool(5);
List<Future<Integer>> futures = new ArrayList<>();
for (int job : jobs) {
    futures.add(executor.submit(() -> process(job)));
}
executor.shutdown();
for (Future<Integer> f : futures) {
    results.add(f.get());
}
```

```go
// Go 워커 풀: 채널 기반으로 더 유연
results := workerPool(5, jobs)
for result := range results {
    collect(result)
}
```
