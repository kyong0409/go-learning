# 03. select 문

## select란?

`select`는 여러 채널 연산을 **동시에 기다리다가** 준비된 것을 실행하는 제어 구조입니다. `switch`와 문법이 비슷하지만 채널 연산 전용입니다.

```go
select {
case v := <-ch1:
    // ch1에서 값을 수신했을 때
    fmt.Println("ch1:", v)
case ch2 <- 42:
    // ch2에 값을 송신했을 때
    fmt.Println("ch2에 송신 완료")
case <-time.After(1 * time.Second):
    // 1초가 지났을 때
    fmt.Println("타임아웃")
default:
    // 준비된 채널이 없을 때 즉시 실행
    fmt.Println("현재 준비된 채널 없음")
}
```

---

## select 동작 원리

1. 모든 case의 채널 표현식을 동시에 평가
2. 준비된 case가 없으면: `default`가 있으면 실행, 없으면 블로킹
3. 준비된 case가 하나: 해당 case 실행
4. 준비된 case가 여러 개: **균일한 무작위 선택** (특정 case가 굶지 않도록)

```go
ch1 := make(chan int, 1)
ch2 := make(chan int, 1)
ch1 <- 1
ch2 <- 2

// 두 case 모두 준비됨 → 매 실행마다 무작위 선택
select {
case v := <-ch1:
    fmt.Println("ch1:", v)
case v := <-ch2:
    fmt.Println("ch2:", v)
}
```

---

## 타임아웃 패턴

외부 API 호출, DB 쿼리 등 응답 시간이 불확실한 작업에 필수입니다.

```go
// time.After: 지정 시간 후 현재 시간을 전송하는 채널 반환
func fetchWithTimeout(url string) (string, error) {
    resultCh := make(chan string, 1)

    go func() {
        result := callAPI(url) // 느릴 수 있음
        resultCh <- result
    }()

    select {
    case result := <-resultCh:
        return result, nil
    case <-time.After(2 * time.Second):
        return "", errors.New("타임아웃: 2초 초과")
    }
}
```

### time.After vs time.NewTimer

```go
// time.After: 간단하지만 타이머를 재사용/중단 불가 → 메모리 누수 주의
select {
case <-time.After(5 * time.Second):
    fmt.Println("5초 경과")
}

// time.NewTimer: 재사용 가능, Stop()으로 명시적 중단
timer := time.NewTimer(5 * time.Second)
defer timer.Stop() // 중요: 타이머를 항상 Stop해야 GC됨

select {
case result := <-workCh:
    timer.Stop() // 작업 완료 시 타이머 중단
    process(result)
case <-timer.C:
    fmt.Println("타임아웃")
}
```

`time.After`는 루프 안에서 반복 호출 시 매번 새 타이머가 생성되어 완료 전까지 GC 되지 않습니다. 루프 안에서는 `time.NewTimer`를 루프 밖에서 한 번 생성하거나 `time.NewTicker`를 사용하세요.

---

## 논블로킹 채널 연산 (default case)

`default` case를 추가하면 모든 채널이 준비되지 않았을 때 블로킹 없이 즉시 실행됩니다.

```go
ch := make(chan int, 1)

// 논블로킹 수신
select {
case v := <-ch:
    fmt.Println("수신:", v)
default:
    fmt.Println("데이터 없음, 계속 진행") // 블로킹 없이 즉시
}

// 논블로킹 송신
select {
case ch <- 42:
    fmt.Println("송신 성공")
default:
    fmt.Println("버퍼 가득 참, 송신 건너뜀")
}
```

### 논블로킹의 활용: 폴링(Polling)

```go
for {
    select {
    case job := <-jobQueue:
        process(job)
    default:
        // 작업 없으면 다른 일 하거나 잠시 대기
        doLowPriorityWork()
        time.Sleep(10 * time.Millisecond)
    }
}
```

---

## for-select 루프: 고루틴의 메인 루프

고루틴의 이벤트 루프 패턴입니다.

```go
func worker(jobs <-chan Job, results chan<- Result, done <-chan struct{}) {
    for {
        select {
        case <-done:
            fmt.Println("워커 종료")
            return

        case job, ok := <-jobs:
            if !ok {
                // jobs 채널이 닫힘 → 정상 종료
                return
            }
            result := process(job)
            select {
            case results <- result:
            case <-done: // 결과 전송 중에도 취소 확인
                return
            }
        }
    }
}
```

### for-select에서 break 주의사항

```go
// 잘못된 코드: break가 select를 탈출하지만 for는 계속됨
for {
    select {
    case v := <-ch:
        if v == 0 {
            break // for가 아닌 select를 탈출!
        }
    }
}

// 올바른 방법 1: 레이블 사용
loop:
for {
    select {
    case v := <-ch:
        if v == 0 {
            break loop // for 루프 탈출
        }
    }
}

// 올바른 방법 2: 플래그 변수
done := false
for !done {
    select {
    case v := <-ch:
        if v == 0 {
            done = true
        }
    }
}

// 올바른 방법 3: return (함수 안에서)
func process(ch <-chan int) {
    for {
        select {
        case v := <-ch:
            if v == 0 {
                return // 함수 종료
            }
        }
    }
}
```

---

## 주기적 작업 (time.Ticker)

```go
func periodicTask() {
    // time.Tick: 프로그램 전체 생명주기 동안 유지 (Stop 불가)
    // 장기 실행 프로그램에서는 메모리 누수 가능 → time.NewTicker 사용 권장
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop() // 반드시 Stop 호출

    done := make(chan struct{})

    go func() {
        time.Sleep(2 * time.Second)
        close(done)
    }()

    for {
        select {
        case t := <-ticker.C:
            fmt.Println("tick:", t.Format("15:04:05"))
        case <-done:
            fmt.Println("주기적 작업 종료")
            return
        }
    }
}
```

---

## nil 채널을 select에서 활용

닫힌 채널을 nil로 설정해서 해당 case를 동적으로 비활성화할 수 있습니다.

```go
// 두 채널이 모두 닫힐 때까지 팬인
func fanIn(ch1, ch2 <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for ch1 != nil || ch2 != nil {
            select {
            case v, ok := <-ch1:
                if !ok {
                    ch1 = nil // 이 case 비활성화
                    continue
                }
                out <- v
            case v, ok := <-ch2:
                if !ok {
                    ch2 = nil // 이 case 비활성화
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

## 우선순위 select

Go의 select는 기본적으로 공정한 무작위 선택을 합니다. 우선순위가 필요하면 중첩 select로 구현합니다.

```go
// 우선순위: highPriority > lowPriority
func prioritySelect(highPriority, lowPriority <-chan int) {
    for {
        // 먼저 high priority 확인 (논블로킹)
        select {
        case v := <-highPriority:
            fmt.Println("고우선순위 처리:", v)
            continue
        default:
        }

        // high priority가 없으면 둘 다 기다림
        select {
        case v := <-highPriority:
            fmt.Println("고우선순위 처리:", v)
        case v := <-lowPriority:
            fmt.Println("저우선순위 처리:", v)
        }
    }
}
```

---

## select와 데드락

```go
// 모든 case가 블로킹이고 default도 없으면 데드락
ch1 := make(chan int)
ch2 := make(chan int)

select {
case v := <-ch1: // 아무도 ch1에 보내지 않음
    fmt.Println(v)
case v := <-ch2: // 아무도 ch2에 보내지 않음
    fmt.Println(v)
}
// fatal error: all goroutines are asleep - deadlock!
```

---

## 패턴 요약

| 패턴 | 코드 |
|------|------|
| 타임아웃 | `case <-time.After(d):` |
| 논블로킹 수신 | `select { case v := <-ch: ... default: }` |
| 취소 신호 | `case <-ctx.Done():` 또는 `case <-done:` |
| 주기적 작업 | `case <-ticker.C:` |
| 동적 case 비활성화 | `ch = nil` |
| for-select 루프 탈출 | `break loop` 또는 `return` |

---

## Python/Java 개발자를 위한 비교

### Python asyncio.wait vs select

```python
import asyncio

# Python: asyncio.wait로 여러 코루틴 중 완료된 것 처리
done, pending = await asyncio.wait(
    [task1, task2],
    return_when=asyncio.FIRST_COMPLETED,
    timeout=2.0
)
```

```go
// Go: select로 여러 채널 중 준비된 것 처리
select {
case result1 := <-ch1:
    process(result1)
case result2 := <-ch2:
    process(result2)
case <-time.After(2 * time.Second):
    fmt.Println("타임아웃")
}
```

### Java CompletableFuture vs select

Java에서 여러 비동기 작업 중 첫 번째 완료를 기다리려면 `CompletableFuture.anyOf()`를 쓰지만 타입 안전성이 없습니다. Go의 `select`는 타입 안전하고 타임아웃, 취소를 자연스럽게 통합합니다.
