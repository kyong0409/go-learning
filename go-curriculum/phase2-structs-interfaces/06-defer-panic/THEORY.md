# 06-defer-panic: defer / panic / recover

> `defer`는 Go의 리소스 정리 메커니즘입니다. `panic`/`recover`는 예외가 아닌 극단적 상황을 위한 도구입니다.

---

## 1. defer 기본 동작

`defer`로 등록된 함수 호출은 **둘러싼 함수가 반환되기 직전**에 실행됩니다.

```go
func example() {
    fmt.Println("시작")
    defer fmt.Println("defer 1")
    defer fmt.Println("defer 2")
    defer fmt.Println("defer 3")
    fmt.Println("끝")
}
// 출력:
// 시작
// 끝
// defer 3   ← 마지막 등록이 먼저 실행 (LIFO)
// defer 2
// defer 1
```

### LIFO (후입선출) 순서

여러 `defer`는 스택처럼 동작합니다. 나중에 등록된 것이 먼저 실행됩니다.

```go
func cleanup() {
    defer fmt.Println("3. DB 연결 해제")  // 세 번째 등록 → 첫 번째 실행
    defer fmt.Println("2. 트랜잭션 롤백") // 두 번째 등록 → 두 번째 실행
    defer fmt.Println("1. 락 해제")       // 첫 번째 등록 → 세 번째 실행
    // ...
}
// 출력:
// 3. DB 연결 해제
// 2. 트랜잭션 롤백
// 1. 락 해제
```

LIFO 순서는 의도적입니다. 리소스는 획득한 역순으로 해제하는 것이 안전합니다.

---

## 2. 인자는 defer 시점에 평가됨

`defer` 구문의 인자는 `defer`가 선언되는 시점에 즉시 평가됩니다. 함수 실행 시점이 아닙니다.

```go
func demoArgs() {
    x := 10
    defer fmt.Println("defer 시점의 x:", x)  // x=10이 지금 평가됨
    x = 20
    fmt.Println("함수 끝의 x:", x)
}
// 출력:
// 함수 끝의 x: 20
// defer 시점의 x: 10  ← defer 등록 시점의 값
```

---

## 3. 클로저로 defer 시 최종 값 참조

클로저를 사용하면 함수 실행 시점의 변수 값을 참조합니다.

```go
func demoClosureDefer() {
    x := 10
    defer func() {
        fmt.Println("클로저 defer의 x:", x)  // 실행 시점(return 직전)의 x 참조
    }()
    x = 20
    fmt.Println("함수 끝의 x:", x)
}
// 출력:
// 함수 끝의 x: 20
// 클로저 defer의 x: 20  ← 클로저는 변수를 캡처하므로 최신 값
```

**인자 평가 vs 클로저 요약**

| 방식 | 값 평가 시점 |
|------|------------|
| `defer f(x)` | `defer` 선언 시점 |
| `defer func() { f(x) }()` | 함수 실행 시점 (return 직전) |

---

## 4. Named Return과 defer의 상호작용

이름 있는 반환값(named return)과 `defer` 클로저를 결합하면 반환 직전에 반환값을 수정할 수 있습니다.

```go
// 반환값을 defer로 2배로 만드는 예
func doubleResult() (result int) {  // named return: result
    defer func() {
        result *= 2  // 반환 직전에 result 수정
    }()
    result = 21
    return  // result=21이지만, defer가 42로 만들고 반환
}
// doubleResult() == 42
```

**실용적인 패턴: defer로 에러에 컨텍스트 추가**

```go
func processFile(path string) (err error) {  // named return: err
    defer func() {
        if err != nil {
            // 에러가 발생했으면 래핑하여 컨텍스트 추가
            err = fmt.Errorf("processFile(%s): %w", path, err)
        }
    }()

    f, err := os.Open(path)
    if err != nil {
        return  // defer가 err를 래핑해서 반환
    }
    defer f.Close()
    // ...
    return
}
```

---

## 5. defer 실전 패턴

### 리소스 정리: 파일 닫기

```go
func readFile(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()  // 함수 어디서 return해도 파일이 닫힌다

    data, err := io.ReadAll(f)
    if err != nil {
        return "", err  // defer가 f.Close() 보장
    }
    return string(data), nil
}
```

### 뮤텍스 해제

```go
func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()  // Lock 바로 다음 줄에 defer — 잊어버릴 일 없음
    c.count++
}
```

### HTTP 응답 바디 닫기

```go
func fetchData(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()  // 항상 닫아야 함 (메모리 누수 방지)

    return io.ReadAll(resp.Body)
}
```

### 타이밍 측정

```go
func timeTrack(name string) func() {
    start := time.Now()
    return func() {
        fmt.Printf("%s 소요 시간: %v\n", name, time.Since(start))
    }
}

func slowOperation() {
    defer timeTrack("slowOperation")()  // 끝의 () 주의 — 클로저를 즉시 호출
    time.Sleep(100 * time.Millisecond)
}
```

---

## 6. panic: 언제 사용하고 언제 사용하지 않는가

`panic`은 현재 함수 실행을 즉시 중단하고, 등록된 `defer`를 실행하며 콜 스택을 역방향으로 풀어 올라갑니다. 아무도 `recover`하지 않으면 프로그램이 종료됩니다.

### 언제 사용: 프로그래머 실수, 복구 불가능한 상태

```go
// 절대 nil이 아니어야 하는 상황
func mustOpen(path string) *os.File {
    f, err := os.Open(path)
    if err != nil {
        panic(fmt.Sprintf("필수 파일을 열 수 없습니다: %v", err))
    }
    return f
}

// 프로그래머 오류: 잘못된 인자
func divide(a, b int) int {
    if b == 0 {
        panic("divide: 제수는 0이 될 수 없습니다")  // 호출자의 버그
    }
    return a / b
}
```

### 언제 사용하지 않음: 일반적인 에러 처리

```go
// 잘못된 예: 일반 에러에 panic 사용
func readConfig(path string) Config {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err)  // 파일이 없는 건 일반적인 상황 — error 반환이 맞다
    }
    // ...
}

// 올바른 예: error 반환
func readConfig(path string) (Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Config{}, fmt.Errorf("설정 파일 읽기 실패: %w", err)
    }
    // ...
}
```

**panic vs error 판단 기준**

| 상황 | 사용 |
|------|------|
| 파일이 없음, 네트워크 실패, 잘못된 사용자 입력 | `error` 반환 |
| nil 포인터 역참조, 슬라이스 범위 초과 (런타임) | 런타임 panic (자동) |
| 초기화 실패 (서버 시작 시 필수 설정 없음) | `panic` 또는 `log.Fatal` |
| 라이브러리 내부의 불가능한 상태 | `panic` |
| 외부에서 호출 가능한 모든 함수의 에러 | `error` 반환 |

---

## 7. panic의 동작: 스택 해제와 defer 실행

```
main()
  ↓ 호출
processData()
  ↓ 호출
parseInput()
  ↓ panic 발생!

실행 흐름:
1. parseInput의 defer 실행
2. processData의 defer 실행
3. main의 defer 실행
4. 아무도 recover하지 않으면 → 프로그램 종료 + 스택 트레이스 출력
```

---

## 8. recover: defer 내에서만 동작

`recover()`는 `panic`을 잡아 프로그램이 계속 실행되도록 합니다. **반드시 `defer` 내에서만 동작**합니다.

```go
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            // panic 값을 에러로 변환
            err = fmt.Errorf("패닉 복구: %v", r)
        }
    }()

    result = a / b  // b=0이면 런타임 panic
    return
}

result, err := safeDiv(10, 0)
// result=0, err="패닉 복구: runtime error: integer divide by zero"
```

`recover()` 반환값:
- panic이 없었으면: `nil`
- panic이 있었으면: `panic()`에 전달된 값

---

## 9. 라이브러리 패턴: 내부 panic → 외부 error

라이브러리에서 내부적으로 `panic`을 쓰고 공개 함수에서 `error`로 변환하는 패턴입니다. 깊은 재귀에서 에러를 매번 반환하지 않아도 되어 코드가 간결해집니다.

```go
type Parser struct{ data string; pos int }

// 내부 함수: panic 사용 (외부에 노출되지 않음)
func (p *Parser) expect(ch byte) {
    if p.pos >= len(p.data) || p.data[p.pos] != ch {
        panic(fmt.Sprintf("위치 %d: '%c' 기대", p.pos, ch))
    }
    p.pos++
}

// 공개 함수: panic을 error로 변환
func (p *Parser) Parse() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("파싱 실패: %v", r)
        }
    }()

    p.expect('[')  // 실패 시 panic — 위의 defer가 잡음
    p.expect(']')
    return nil
}
```

---

## 10. HTTP 서버의 panic 복구 미들웨어

실제 서비스에서 가장 중요한 `recover` 활용 패턴입니다.

```go
// 미들웨어: 핸들러의 panic을 잡아 500 응답으로 변환
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                log.Printf("핸들러 패닉 복구: %v\n%s", rec, debug.Stack())
                http.Error(w, "내부 서버 오류", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// 사용
mux := http.NewServeMux()
mux.HandleFunc("/", myHandler)
http.ListenAndServe(":8080", recoveryMiddleware(mux))
```

한 요청의 panic이 전체 서버를 죽이지 않도록 합니다.

---

## 11. Python / Java 비교

### Python try/finally vs Go defer

```python
# Python
def read_file(path):
    f = open(path)
    try:
        return f.read()
    finally:
        f.close()  # 항상 실행

# 또는 with 문 (권장)
def read_file(path):
    with open(path) as f:
        return f.read()
```

```go
// Go: defer
func readFile(path string) (string, error) {
    f, err := os.Open(path)
    if err != nil {
        return "", err
    }
    defer f.Close()  // Python의 with와 유사
    data, _ := io.ReadAll(f)
    return string(data), nil
}
```

**차이**: Python `with`는 블록을 벗어날 때 실행되지만, Go `defer`는 함수를 벗어날 때 실행됩니다. Go에서 블록 수준의 정리가 필요하면 익명 함수를 사용합니다.

```go
func example() {
    // 블록 수준 defer: 익명 함수로 감싸기
    func() {
        resource := acquire()
        defer release(resource)
        // 이 블록을 벗어날 때 release 실행
    }()
    // 여기서는 이미 release됨
}
```

### Python raise vs Go panic

```python
# Python: 모든 예외 상황에 raise 사용
def find_user(id):
    if id not in db:
        raise UserNotFoundError(f"User {id} not found")  # 일반적인 에러
```

```go
// Go: 일반 에러는 error 반환, 극단적 상황만 panic
func FindUser(id int) (*User, error) {
    user, ok := db[id]
    if !ok {
        return nil, &NotFoundError{ID: id}  // error 반환 (panic 아님)
    }
    return &user, nil
}
```

**핵심 차이**: Python/Java의 예외는 일반적인 제어 흐름에 쓰이지만, Go의 `panic`은 진짜 예외적 상황(프로그래머 실수, 복구 불가능한 상태)에만 씁니다. 일반 에러는 반드시 `error` 반환으로 처리합니다.

### Java try-with-resources vs Go defer

```java
// Java
try (FileInputStream fis = new FileInputStream("file.txt")) {
    // fis 사용
}  // 블록을 벗어날 때 자동으로 close()
```

```go
// Go
f, err := os.Open("file.txt")
if err != nil { return err }
defer f.Close()  // 함수를 벗어날 때 close
```

Java `try-with-resources`는 블록 수준이지만, Go `defer`는 함수 수준입니다. 이 차이로 인해 Go에서 루프 안에서 파일을 열고 닫을 때는 주의가 필요합니다.

```go
// 주의: 루프 안의 defer는 함수 종료 시 모두 실행됨 (루프 반복마다 아님)
for _, path := range files {
    f, _ := os.Open(path)
    defer f.Close()  // 모든 파일이 함수 종료 시 닫힘 — 의도치 않은 동작
}

// 올바른 패턴: 익명 함수로 감싸기
for _, path := range files {
    func() {
        f, _ := os.Open(path)
        defer f.Close()  // 이 익명 함수 종료 시 닫힘
        process(f)
    }()
}
```
