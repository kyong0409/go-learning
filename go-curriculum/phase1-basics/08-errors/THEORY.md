# 08-errors: 에러 처리

## Go의 에러 처리 철학: "에러는 값이다"

Rob Pike의 유명한 말: **"Errors are values."**

Go에는 예외(Exception)가 없습니다. 에러는 그냥 함수의 반환값입니다. 이는 처음에는 장황하게 느껴지지만, 에러가 코드 흐름에서 명시적으로 처리되므로 에러를 무시하기 어렵고 예측 가능합니다.

```go
// Java/Python 방식 (Go에 없음):
// try {
//     result = riskyOperation();
// } catch (IOException e) {
//     handleError(e);
// }

// Go 방식:
result, err := riskyOperation()
if err != nil {
    handleError(err)
    return  // 또는 에러 전파
}
useResult(result)
```

---

## error 인터페이스

`error`는 Go 내장 인터페이스입니다.

```go
type error interface {
    Error() string
}
```

`Error() string` 메서드를 구현하면 어떤 타입이든 `error`입니다. `nil` 에러는 "에러 없음"을 의미합니다.

```go
var err error = nil
fmt.Println(err == nil)  // true (에러 없음)

err = errors.New("무언가 잘못됨")
fmt.Println(err)         // "무언가 잘못됨"
fmt.Println(err.Error()) // "무언가 잘못됨"
fmt.Printf("%T\n", err)  // *errors.errorString
```

---

## errors.New와 fmt.Errorf

### errors.New — 단순한 에러 생성
```go
import "errors"

err := errors.New("파일을 찾을 수 없습니다")
fmt.Println(err)  // "파일을 찾을 수 없습니다"
```

### fmt.Errorf — 형식화된 에러 생성
```go
filename := "config.json"
err := fmt.Errorf("파일 %q 열기 실패: 권한 거부", filename)
// "파일 \"config.json\" 열기 실패: 권한 거부"

userID := 42
err2 := fmt.Errorf("사용자 ID %d를 찾을 수 없습니다", userID)
```

### fmt.Errorf + %w — 에러 래핑 (Go 1.13+)
```go
baseErr := errors.New("연결 시간 초과")
wrapped := fmt.Errorf("DB 조회 실패: %w", baseErr)

// %w 로 래핑된 에러는 errors.Is/As로 탐색 가능
fmt.Println(errors.Is(wrapped, baseErr))  // true
fmt.Println(wrapped)  // "DB 조회 실패: 연결 시간 초과"
```

---

## if err != nil 패턴

Go에서 가장 많이 보게 될 패턴입니다.

```go
// 기본 패턴
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething 실패: %w", err)
}

// 에러만 반환하는 함수
if err := writeFile(path, data); err != nil {
    return err
}

// 에러 전파 (상위 레이어에서 처리)
func processUser(id int) error {
    user, err := fetchUser(id)
    if err != nil {
        return fmt.Errorf("processUser(%d): %w", id, err)
    }

    if err := validateUser(user); err != nil {
        return fmt.Errorf("사용자 유효성 검사 실패: %w", err)
    }

    return saveUser(user)
}
```

### 왜 반복적이어도 좋은가?

"장황하다"는 비판이 있지만, Go 커뮤니티는 이를 **장점**으로 봅니다:
1. 코드 리뷰에서 에러 처리 누락을 즉시 알 수 있음
2. 에러 경로가 정상 경로와 동일 수준으로 명시됨
3. 어떤 함수가 에러를 낼 수 있는지 항상 명확함

---

## 센티널 에러 (Sentinel Errors)

패키지 수준에서 미리 선언된 잘 알려진 에러 값입니다. `==` 비교나 `errors.Is`로 확인합니다.

```go
// 선언 (패키지 수준, 대문자로 exported)
var (
    ErrNotFound     = errors.New("찾을 수 없음")
    ErrUnauthorized = errors.New("권한 없음")
    ErrInvalidInput = errors.New("잘못된 입력")
)

// 반환
func findUser(id int) (*User, error) {
    if id <= 0 {
        return nil, ErrInvalidInput
    }
    user, exists := db[id]
    if !exists {
        return nil, fmt.Errorf("findUser(%d): %w", id, ErrNotFound)
    }
    return user, nil
}

// 사용 (errors.Is로 래핑된 에러도 확인)
user, err := findUser(99)
if errors.Is(err, ErrNotFound) {
    // 404 응답
} else if errors.Is(err, ErrInvalidInput) {
    // 400 응답
} else if err != nil {
    // 500 응답
}
```

표준 라이브러리의 센티널 에러 예:
```go
io.EOF              // 파일/스트림 끝
os.ErrNotExist      // 파일 없음
os.ErrPermission    // 권한 없음
sql.ErrNoRows       // 쿼리 결과 없음
context.DeadlineExceeded
context.Canceled
```

---

## 에러 래핑과 errors.Is / errors.As

### errors.Is — 에러 체인에서 특정 값 찾기
```go
// 에러 체인 구성
original := errors.New("원본 에러")
layer1 := fmt.Errorf("레이어1: %w", original)
layer2 := fmt.Errorf("레이어2: %w", layer1)

// 체인의 어느 위치에 있든 찾을 수 있음
fmt.Println(errors.Is(layer2, original))  // true
fmt.Println(errors.Is(layer2, layer1))    // true

// 직접 비교는 최상위만
fmt.Println(layer2 == original)           // false
```

### 커스텀 에러 타입과 errors.As
```go
// 커스텀 에러 타입 정의
type ValidationError struct {
    Field   string
    Message string
    Value   any
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("유효성 검사 실패 - %s: %s (값: %v)",
        e.Field, e.Message, e.Value)
}

// 반환
func validateAge(age int) error {
    if age < 0 {
        return &ValidationError{
            Field:   "age",
            Message: "0 이상이어야 합니다",
            Value:   age,
        }
    }
    return nil
}

// errors.As: 특정 타입의 에러를 추출
err := validateAge(-5)
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("필드: %s\n", valErr.Field)
    fmt.Printf("메시지: %s\n", valErr.Message)
    fmt.Printf("잘못된 값: %v\n", valErr.Value)
}
```

### Unwrap() 메서드 — errors.As/Is가 동작하는 원리
```go
type DatabaseError struct {
    Op  string
    Err error  // 래핑된 원인 에러
}

func (e *DatabaseError) Error() string {
    return fmt.Sprintf("DB [%s]: %v", e.Op, e.Err)
}

// Unwrap()을 구현하면 errors.Is/As가 내부 에러까지 탐색
func (e *DatabaseError) Unwrap() error {
    return e.Err
}

dbErr := &DatabaseError{Op: "SELECT", Err: ErrNotFound}
fmt.Println(errors.Is(dbErr, ErrNotFound))  // true (Unwrap 덕분)
```

### errors.Join — 여러 에러 합치기 (Go 1.20+)
```go
err1 := errors.New("필드 A 유효성 오류")
err2 := errors.New("필드 B 유효성 오류")
err3 := errors.New("필드 C 유효성 오류")

combined := errors.Join(err1, err2, err3)
fmt.Println(combined)
// "필드 A 유효성 오류\n필드 B 유효성 오류\n필드 C 유효성 오류"
```

---

## panic과 recover

### panic — 복구 불가능한 상황
```go
// panic: 즉시 현재 함수 실행을 중단하고 스택을 역방향으로 풀어냄
// 모든 defer가 실행된 후 프로그램 종료

// 언제 panic을 써야 하는가:
// 1. 프로그래밍 오류 (인덱스 범위 초과, nil 역참조)
// 2. 불변 조건 위반 (이래서는 안 됨이 명확한 경우)
// 3. 초기화 실패 (startup에서 필수 리소스 로드 실패)

func mustOpenDB(dsn string) *sql.DB {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        panic(fmt.Sprintf("DB 연결 실패: %v", err))
    }
    return db
}

// regexp.MustCompile도 내부적으로 panic 사용
var re = regexp.MustCompile(`\d+`)  // 잘못된 패턴이면 startup에서 panic
```

### recover — panic 잡기
```go
// recover는 defer 함수 내에서만 동작

func safeOperation(fn func() int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("패닉 복구: %v", r)
        }
    }()
    result = fn()
    return
}

// 사용
result, err := safeOperation(func() int {
    return 10 / 0  // 패닉!
})
// result=0, err="패닉 복구: runtime error: integer divide by zero"
```

**recover의 사용 범위**: HTTP 서버에서 고루틴 하나가 panic으로 전체 서버가 죽지 않도록 최상단에서 recover하는 패턴이 일반적입니다. 일반 비즈니스 로직에서는 거의 사용하지 않습니다.

---

## 에러 처리 모범 사례

### 에러는 한 번만 처리
```go
// 나쁜 예: 같은 에러를 두 번 처리
func bad() error {
    err := doSomething()
    if err != nil {
        log.Printf("에러: %v", err)  // 로그
        return err                    // 그리고 반환 (호출자도 로그할 것)
    }
    return nil
}

// 좋은 예: 한 번만 처리 (로그하거나 반환하거나)
func good() error {
    if err := doSomething(); err != nil {
        return fmt.Errorf("good: %w", err)  // 컨텍스트 추가 후 반환
    }
    return nil
}
// 최상위(main, HTTP handler)에서 한 번 로그
```

### 에러에 컨텍스트 추가
```go
// 에러를 전파할 때 어디서 발생했는지 컨텍스트 추가
func getUser(id int) (*User, error) {
    user, err := db.QueryUser(id)
    if err != nil {
        return nil, fmt.Errorf("getUser(%d): %w", id, err)
        // "getUser(42): sql: no rows in result set"
    }
    return user, nil
}
```

---

## Python/Java와의 비교

### try/catch vs if err != nil
```python
# Python: 예외로 제어 흐름 변경
try:
    result = divide(a, b)
    save_result(result)
    notify_user(result)
except ZeroDivisionError as e:
    handle_error(e)
except IOError as e:
    handle_io_error(e)
```
```go
// Go: 각 단계마다 명시적 확인
result, err := divide(a, b)
if err != nil {
    return handleError(err)
}

if err := saveResult(result); err != nil {
    return handleError(err)
}

if err := notifyUser(result); err != nil {
    return handleError(err)
}
```

### 예외 계층 vs 에러 값
```java
// Java: 예외 계층 구조
class AppException extends RuntimeException { }
class NotFoundException extends AppException { }
class ValidationException extends AppException { }

try {
    findUser(id);
} catch (NotFoundException e) {
    // 404
} catch (AppException e) {
    // 500
}
```
```go
// Go: 에러 값 + errors.Is/As
var ErrNotFound = errors.New("not found")

err := findUser(id)
if errors.Is(err, ErrNotFound) {
    // 404
} else if err != nil {
    // 500
}
```

### finally vs defer
```python
try:
    f = open("file.txt")
    process(f)
finally:
    f.close()  # 예외 발생 여부와 무관하게 실행
```
```go
f, err := os.Open("file.txt")
if err != nil { return err }
defer f.Close()  // 함수 반환 시 항상 실행
process(f)
```

---

## 핵심 정리

1. `error`는 `Error() string` 메서드를 가진 인터페이스 — 어떤 타입도 구현 가능
2. 에러는 반환값 — `if err != nil` 패턴이 Go의 표준
3. 센티널 에러(`var ErrXxx = errors.New(...)`)는 특정 에러 종류를 식별할 때
4. 에러 래핑(`%w`)으로 컨텍스트를 추가하면서 원인을 보존
5. `errors.Is` = 에러 값 비교 (체인 탐색), `errors.As` = 에러 타입 추출
6. `panic`은 복구 불가능한 상황에만 — 일반 에러는 `error` 반환으로
