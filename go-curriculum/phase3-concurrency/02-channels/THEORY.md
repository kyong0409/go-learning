# 02. 채널 (Channel)

## 채널의 핵심 철학

> "메모리를 공유해서 통신하지 말고, 통신을 통해 메모리를 공유하라."
> — Rob Pike

채널은 고루틴 간에 데이터를 안전하게 주고받는 **타입이 있는 파이프**입니다. 내부적으로 뮤텍스와 큐로 구현되어 있어 별도 동기화 없이도 스레드 안전합니다.

---

## 채널 생성

```go
// 언버퍼 채널 (버퍼 크기 0)
ch := make(chan int)        // int를 전달하는 채널
sch := make(chan string)    // string을 전달하는 채널
errCh := make(chan error)   // error를 전달하는 채널

// 버퍼 채널 (버퍼 크기 N)
bch := make(chan int, 5)    // 버퍼 크기 5짜리 int 채널
```

---

## 언버퍼 채널 vs 버퍼 채널

### 언버퍼 채널: 동기적(Synchronous)

송신자와 수신자가 **동시에 준비**되어야 데이터가 전달됩니다. 랑데부(rendezvous) 방식이라고도 합니다.

```go
ch := make(chan int) // 언버퍼

go func() {
    fmt.Println("송신 준비...")
    ch <- 42  // 수신자가 나타날 때까지 블로킹
    fmt.Println("송신 완료")
}()

time.Sleep(100 * time.Millisecond) // 수신자가 늦게 준비됨
fmt.Println("수신 준비...")
val := <-ch // 이 시점에 송신도 완료됨
fmt.Println("수신 완료:", val)
```

언버퍼 채널은 **동기화 도구**입니다. 두 고루틴이 같은 시점에 만나는 것을 보장합니다.

### 버퍼 채널: 비동기적(Asynchronous)

버퍼가 찰 때까지 수신자 없이도 송신할 수 있습니다.

```go
ch := make(chan int, 3) // 버퍼 크기 3

// 수신자 없이 3개 즉시 송신 가능
ch <- 10 // 즉시 반환
ch <- 20 // 즉시 반환
ch <- 30 // 즉시 반환
// ch <- 40 // 버퍼 가득 참 → 수신자 나타날 때까지 블로킹!

fmt.Println("버퍼 크기:", len(ch))  // 3 (현재 원소 수)
fmt.Println("버퍼 용량:", cap(ch))  // 3 (최대 원소 수)
```

| 상태 | 송신 결과 | 수신 결과 |
|------|----------|----------|
| 버퍼 비어있음 | 즉시 | 블로킹 |
| 버퍼 중간 | 즉시 | 즉시 |
| 버퍼 가득 참 | 블로킹 | 즉시 |

---

## 방향성 채널 (Directional Channels)

채널의 방향을 함수 시그니처에서 제한하여 안전성을 높입니다.

```go
chan T      // 양방향 (송수신 모두 가능)
chan<- T   // 송신 전용 (→ 채널로 들어가는 방향)
<-chan T   // 수신 전용 (← 채널에서 나오는 방향)
```

화살표 방향을 기억하는 팁: `<-ch`로 수신, `ch <-`로 송신. `<-chan`은 채널 왼쪽에 화살표가 있으므로 "꺼내기만 가능".

```go
// 생산자: 수신 전용으로 반환 (호출자가 실수로 송신 못 하게)
func producer(n int) <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 0; i < n; i++ {
            ch <- i
        }
    }()
    return ch // <-chan int로 자동 변환
}

// 소비자: 수신 전용 채널만 받음
func consumer(ch <-chan int) {
    for v := range ch {
        fmt.Println(v)
    }
}

// 파이프 처리기: 수신 전용 입력, 수신 전용 출력 반환
func doubler(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for v := range in {
            out <- v * 2
        }
    }()
    return out
}

func main() {
    nums := producer(5)
    doubled := doubler(nums)
    consumer(doubled)
}
```

---

## 채널 닫기 (close)

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
ch <- 3
close(ch) // 더 이상 보낼 데이터 없음을 알림
```

### 닫기 규칙

```go
// 규칙 1: 닫힌 채널에 송신하면 패닉!
close(ch)
ch <- 42  // panic: send on closed channel

// 규칙 2: 닫힌 채널에서 수신은 가능 (버퍼의 남은 값 먼저, 그 다음 zero value)
ch := make(chan int, 2)
ch <- 10
ch <- 20
close(ch)
fmt.Println(<-ch) // 10 (버퍼 값)
fmt.Println(<-ch) // 20 (버퍼 값)
fmt.Println(<-ch) // 0 (zero value, 채널이 닫혔음)

// 규칙 3: nil 채널 닫기는 패닉!
var nilCh chan int
close(nilCh) // panic: close of nil channel

// 규칙 4: 채널은 송신 측에서만 닫아야 함 (수신 측에서 닫으면 혼란)
```

### range over channel

```go
ch := make(chan int)
go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch) // range가 종료되려면 반드시 close 해야 함!
}()

for v := range ch { // 채널이 닫히고 버퍼가 비워질 때 종료
    fmt.Println(v)
}
```

### comma-ok 패턴

```go
v, ok := <-ch
// ok == true:  정상 수신
// ok == false: 채널이 닫혔고 비어있음, v는 zero value

for {
    v, ok := <-ch
    if !ok {
        break // 채널 닫힘
    }
    process(v)
}
// 위 코드는 for v := range ch { process(v) }와 동일
```

---

## nil 채널

선언만 하고 초기화하지 않은 채널은 nil입니다.

```go
var ch chan int // nil 채널

// nil 채널에 송신: 영원히 블로킹 (데드락)
// ch <- 42

// nil 채널에서 수신: 영원히 블로킹 (데드락)
// <-ch

// nil 채널 닫기: 패닉
// close(ch)

// select에서 nil 채널: 해당 case는 절대 선택되지 않음 ← 유용!
```

### nil 채널의 유용한 활용

```go
// 동적으로 채널 case 비활성화
func merge(ch1, ch2 <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for ch1 != nil || ch2 != nil {
            select {
            case v, ok := <-ch1:
                if !ok {
                    ch1 = nil // 닫힌 채널을 nil로 설정 → 이 case 비활성화
                    continue
                }
                out <- v
            case v, ok := <-ch2:
                if !ok {
                    ch2 = nil // 닫힌 채널을 nil로 설정
                    continue
                }
                out <- v
            }
        }
    }()
    return out
}
```

---

## 채널 용량 선택 가이드

| 버퍼 크기 | 용도 | 예시 |
|----------|------|------|
| `0` (언버퍼) | 동기화, 랑데부 | 두 고루틴 정확한 시점 맞추기 |
| `1` | 시그널, 한 번만 전달 | 완료 알림, 최초 결과 |
| `N` | 배압(backpressure) 제어 | 워커 풀, 요청 큐 |
| `len(data)` | 고루틴 없이 즉시 완료 | 테스트, 간단한 수집 |

```go
// 배압 제어 예시: 최대 100개까지 쌓임, 초과 시 생산자가 느려짐
workQueue := make(chan Job, 100)

// 시그널 예시: 에러 발생 시 한 번만 알림
errOnce := make(chan error, 1) // 버퍼 1: 블로킹 없이 에러 전달
go func() {
    if err := doWork(); err != nil {
        select {
        case errOnce <- err: // 첫 번째 에러만 전달
        default:             // 이미 에러 있으면 무시
        }
    }
}()
```

---

## done 채널 패턴

취소/종료 신호를 여러 고루틴에게 동시에 전파하는 관용적 패턴입니다.

```go
done := make(chan struct{}) // struct{}는 메모리 0바이트 사용

// 여러 고루틴이 동시에 종료 신호를 기다림
for i := 0; i < 3; i++ {
    go func(id int) {
        select {
        case <-done:
            fmt.Printf("고루틴 %d 종료\n", id)
        case result := <-workCh:
            process(result)
        }
    }(i)
}

// close()는 브로드캐스트: 모든 수신자에게 동시 전파
close(done)
// 단순 ch <- struct{}{} 는 하나에게만 전달됨
```

---

## 채널 관련 데드락 패턴

```go
// 데드락 1: 언버퍼 채널에 혼자 송수신
ch := make(chan int)
ch <- 42   // 수신자 없음 → 영원히 블로킹 → 데드락
val := <-ch

// 해결: 고루틴으로 비동기 송신
go func() { ch <- 42 }()
val := <-ch

// 데드락 2: 버퍼 초과
ch := make(chan int, 1)
ch <- 1
ch <- 2 // 버퍼 가득 참 → 수신자 없음 → 데드락

// 해결: 버퍼 크기 늘리거나 별도 고루틴에서 수신

// 데드락 3: 닫지 않은 채널을 range로 순회
ch := make(chan int)
go func() {
    ch <- 1
    ch <- 2
    // close(ch) 빠뜨림!
}()
for v := range ch { // 2개 받고 나서 영원히 블로킹 → 데드락
    fmt.Println(v)
}
```

---

## Python/Java 개발자를 위한 비교

### Python Queue vs Go Channel

```python
import queue
q = queue.Queue(maxsize=5)  # 버퍼 채널과 유사

# 생산자 스레드
q.put(42)           # ch <- 42

# 소비자 스레드
val = q.get()       # val := <-ch
q.task_done()

q.join()            # 모든 항목 처리 대기 (wg.Wait()과 유사)
```

```go
ch := make(chan int, 5)

go func() {
    ch <- 42
    close(ch)
}()

for val := range ch {
    fmt.Println(val)
}
```

Go 채널은 타입 안전하고, 방향성 제한이 가능하며, `select`와 통합됩니다.

### Java BlockingQueue vs Go Channel

```java
BlockingQueue<Integer> queue = new LinkedBlockingQueue<>(5);
queue.put(42);    // ch <- 42 (블로킹)
int val = queue.take(); // <-ch (블로킹)
```

Go 채널이 더 간결하고 언어 레벨에서 지원됩니다.
