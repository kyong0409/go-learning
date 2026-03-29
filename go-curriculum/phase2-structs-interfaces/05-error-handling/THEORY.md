# 05-error-handling: 에러 처리 심화

> Phase 1에서 `error` 반환과 `if err != nil` 패턴을 배웠습니다. 여기서는 에러에 구조와 맥락을 부여하는 방법을 다룹니다.

---

## 1. 커스텀 에러 타입 만들기

`error` 인터페이스는 단 하나의 메서드로 구성됩니다.

```go
type error interface {
    Error() string
}
```

이를 구현한 구조체가 커스텀 에러 타입입니다.

```go
// 필드가 있는 커스텀 에러
type ValidationError struct {
    Field   string  // 에러가 발생한 필드
    Message string  // 에러 메시지
    Value   any     // 잘못된 값 (진단용)
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("유효성 검사 실패 - 필드 '%s': %s (값: %v)",
        e.Field, e.Message, e.Value)
}

// 사용
err := &ValidationError{Field: "Age", Message: "음수 불가", Value: -5}
fmt.Println(err)  // Error() 자동 호출
```

커스텀 에러는 일반 `errors.New` 에러보다 더 많은 정보를 담을 수 있어 호출자가 에러 종류에 따른 처리를 할 수 있습니다.

---

## 2. Unwrap() error 메서드

`Unwrap()` 메서드를 구현하면 `errors.Is()`와 `errors.As()`가 에러 체인을 탐색할 수 있습니다.

```go
type DatabaseError struct {
    Operation string
    Table     string
    Err       error  // 원인 에러 (내부에 래핑)
}

func (e *DatabaseError) Error() string {
    return fmt.Sprintf("DB 에러 [%s on %s]: %v", e.Operation, e.Table, e.Err)
}

// Unwrap을 구현해야 errors.Is/As가 Err 필드를 탐색할 수 있다
func (e *DatabaseError) Unwrap() error {
    return e.Err
}
```

---

## 3. 에러 래핑: fmt.Errorf + %w

```go
func getUser(id int) (*User, error) {
    user, err := findUser(id)
    if err != nil {
        // %w: 에러를 래핑하면서 컨텍스트 메시지 추가
        // errors.Is()와 errors.As()로 내부 에러를 꺼낼 수 있음
        return nil, fmt.Errorf("getUser(%d): %w", id, err)
    }
    return user, nil
}
```

**%w vs %v 차이**

```go
err := errors.New("원본 에러")

// %v: 에러 문자열만 포함 — errors.Is/As로 내부 에러 접근 불가
wrapped1 := fmt.Errorf("컨텍스트: %v", err)
fmt.Println(errors.Is(wrapped1, err))  // false

// %w: 에러를 래핑 — errors.Is/As로 내부 에러 접근 가능
wrapped2 := fmt.Errorf("컨텍스트: %w", err)
fmt.Println(errors.Is(wrapped2, err))  // true
```

**왜 래핑하는가**

Go에는 Java/Python의 스택 트레이스가 없습니다. 에러를 래핑하면서 각 레이어에서 컨텍스트를 추가하는 것이 Go의 방식입니다.

```
에러 체인 예시:
DatabaseError("SELECT on users")
  → fmt.Errorf("getUser(999): ...")
    → NotFoundError{Resource:"User", ID:999}
```

---

## 4. 에러 래핑 체인 시각화

```go
// 레이어 1: 가장 낮은 레벨
func findUser(id int) (*User, error) {
    // ...
    return nil, &NotFoundError{Resource: "User", ID: id}
}

// 레이어 2: 서비스 레이어
func getUser(id int) (*User, error) {
    user, err := findUser(id)
    if err != nil {
        return nil, fmt.Errorf("getUser(%d): %w", id, err)
    }
    return user, nil
}

// 레이어 3: 저장소 레이어
func dbQueryUser(id int) (*User, error) {
    user, err := getUser(id)
    if err != nil {
        return nil, &DatabaseError{
            Operation: "SELECT",
            Table:     "users",
            Err:       err,  // 레이어 2의 에러를 래핑
        }
    }
    return user, nil
}

// 최상위에서 받는 에러:
// DatabaseError → fmt.Errorf → NotFoundError
```

---

## 5. errors.Is(): 에러 체인에서 특정 에러 찾기

`errors.Is(err, target)`는 에러 체인을 재귀적으로 탐색하며 `target`과 같은 에러를 찾습니다.

```go
var ErrNotFound = errors.New("항목을 찾을 수 없습니다")

// 아주 깊이 래핑된 에러에서도 ErrNotFound를 찾을 수 있다
_, err := dbQueryUser(999)
fmt.Println(errors.Is(err, ErrNotFound))       // true — 체인 어딘가에 있음
fmt.Println(errors.Is(err, ErrUnauthorized))   // false — 체인에 없음
```

**커스텀 Is() 구현**

기본 동작은 `==` 비교이지만, `Is()` 메서드를 구현하면 비교 방식을 커스터마이징할 수 있습니다.

```go
type NotFoundError struct {
    Resource string
    ID       int
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s ID=%d를 찾을 수 없습니다", e.Resource, e.ID)
}

// ErrNotFound 센티넬과 비교될 때 true를 반환
func (e *NotFoundError) Is(target error) bool {
    return target == ErrNotFound
}

// 이제 errors.Is(anyNotFoundError, ErrNotFound) == true
```

---

## 6. errors.As(): 에러 체인에서 특정 타입 찾기

`errors.As(err, &target)`는 에러 체인을 탐색하며 `target` 타입의 에러를 찾아 `target`에 설정합니다.

```go
_, err := dbQueryUser(999)

// DatabaseError 타입 꺼내기
var dbErr *DatabaseError
if errors.As(err, &dbErr) {
    fmt.Printf("DB 작업: %s, 테이블: %s\n", dbErr.Operation, dbErr.Table)
}

// NotFoundError 타입 꺼내기 (더 깊은 체인)
var notFoundErr *NotFoundError
if errors.As(err, &notFoundErr) {
    fmt.Printf("리소스: %s, ID: %d\n", notFoundErr.Resource, notFoundErr.ID)
}
```

**errors.Is vs errors.As 선택**

| 상황 | 사용 |
|------|------|
| 특정 에러 값인지 확인 (센티넬 에러 비교) | `errors.Is` |
| 특정 에러 타입인지 확인하고 필드를 꺼내고 싶다 | `errors.As` |

---

## 7. errors.Join(): 여러 에러 합치기 (Go 1.20+)

```go
import "errors"

func validateAll(u User) error {
    var errs []error

    if u.Name == "" {
        errs = append(errs, errors.New("이름 필수"))
    }
    if u.Age < 0 {
        errs = append(errs, fmt.Errorf("나이 오류: %d", u.Age))
    }
    if u.Email == "" {
        errs = append(errs, errors.New("이메일 필수"))
    }

    return errors.Join(errs...)  // nil 에러는 자동으로 제외됨
}

err := validateAll(User{})
fmt.Println(err)
// 이름 필수
// 나이 오류: 0
// 이메일 필수
```

`errors.Join`이 반환한 에러도 `errors.Is`와 `errors.As`로 탐색할 수 있습니다.

---

## 8. 센티넬 에러 vs 커스텀 타입 vs 에러 문자열: 언제 무엇을

### 센티넬 에러 (Sentinel Error)

```go
var ErrNotFound = errors.New("not found")
var ErrTimeout   = errors.New("timeout")
```

- **언제**: 호출자가 특정 에러 조건을 `errors.Is`로 구분해야 할 때
- **예**: `io.EOF`, `sql.ErrNoRows`, `os.ErrNotExist`
- **주의**: 패키지의 공개 API가 되므로 신중하게 정의

### 커스텀 에러 타입

```go
type ValidationError struct { Field, Message string }
```

- **언제**: 호출자가 에러의 구조화된 정보(필드, 코드 등)를 꺼내야 할 때
- **예**: HTTP 400 응답에 어떤 필드가 잘못됐는지 반환

### 단순 에러 문자열

```go
return fmt.Errorf("파싱 실패: 줄 %d", lineNum)
```

- **언제**: 에러가 로그에 기록되거나 사람이 읽는 용도이고, 프로그래밍적 처리가 불필요할 때
- **패키지 내부 에러**에 주로 사용

---

## 9. 에러 처리 안티패턴

### 안티패턴 1: 에러 무시

```go
// 절대 하지 마세요
result, _ := os.Open("file.txt")  // 에러를 버림

// 대신
f, err := os.Open("file.txt")
if err != nil {
    return fmt.Errorf("파일 열기 실패: %w", err)
}
```

### 안티패턴 2: 로그 + 반환 동시

```go
// 중복 로그가 쌓임 — 로그 아니면 반환, 둘 중 하나만
func processUser(id int) error {
    user, err := getUser(id)
    if err != nil {
        log.Printf("에러: %v", err)  // 로그 기록하고
        return err                    // 또 반환 — 상위에서 또 로그 찍힐 수 있음
    }
    _ = user
    return nil
}

// 대신: 에러는 래핑해서 반환만 하고, 최상위 레이어에서 한 번만 로그
func processUser(id int) error {
    user, err := getUser(id)
    if err != nil {
        return fmt.Errorf("processUser(%d): %w", id, err)
    }
    _ = user
    return nil
}
```

### 안티패턴 3: 과도한 래핑

```go
// 모든 레이어에서 래핑하면 에러 메시지가 너무 길어짐
// "handler: service: repository: db: getUser(1): not found"

// 의미 있는 컨텍스트만 추가하세요
return fmt.Errorf("사용자 조회 실패 (id=%d): %w", id, err)
```

---

## 10. 실전 에러 전략

### 패키지 경계에서 래핑

```go
// 외부 패키지의 에러를 내 패키지의 에러로 변환
func (r *UserRepo) FindByID(id int) (*User, error) {
    row := r.db.QueryRow("SELECT ...", id)
    if err := row.Scan(&user); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound  // 내 패키지의 에러로 변환
        }
        return nil, fmt.Errorf("UserRepo.FindByID: %w", err)
    }
    return &user, nil
}
```

### 최상위(HTTP 핸들러, main)에서 로깅

```go
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    user, err := userService.GetUser(id)
    if err != nil {
        // 에러 종류에 따른 HTTP 응답 결정
        switch {
        case errors.Is(err, ErrUserNotFound):
            http.Error(w, "사용자를 찾을 수 없습니다", http.StatusNotFound)
        case errors.Is(err, ErrUnauthorized):
            http.Error(w, "권한이 없습니다", http.StatusUnauthorized)
        default:
            log.Printf("내부 오류: %v", err)  // 최상위에서 한 번만 로그
            http.Error(w, "내부 서버 오류", http.StatusInternalServerError)
        }
        return
    }
    // 정상 응답 처리
}
```

---

## 11. Python / Java 비교

### Python 예외 vs Go 에러

```python
# Python: 예외 기반
def find_user(id: int) -> User:
    user = db.query(id)
    if user is None:
        raise UserNotFoundError(f"User {id} not found")
    return user

try:
    user = find_user(999)
except UserNotFoundError as e:
    print(f"처리: {e}")
```

```go
// Go: 에러 반환 기반
func FindUser(id int) (*User, error) {
    user, ok := db[id]
    if !ok {
        return nil, &NotFoundError{Resource: "User", ID: id}
    }
    return &user, nil
}

user, err := FindUser(999)
if err != nil {
    var notFound *NotFoundError
    if errors.As(err, &notFound) {
        fmt.Printf("처리: %v\n", notFound)
    }
}
```

**핵심 차이**
- Python 예외는 호출 스택을 타고 자동으로 전파됩니다. Go 에러는 명시적으로 반환하고 처리해야 합니다.
- Python의 `try/except`는 여러 레이어를 건너뛸 수 있지만, Go는 각 레이어에서 에러를 처리하거나 래핑해야 합니다.
- Go의 명시적 에러 처리는 장황하지만, 어디서 에러가 발생하고 어떻게 전파되는지 코드를 읽는 것만으로 알 수 있습니다.

### Java checked exception vs Go 에러

```java
// Java: checked exception — 반드시 처리하거나 선언해야 함
public User findUser(int id) throws UserNotFoundException {
    // ...
}

try {
    User user = findUser(999);
} catch (UserNotFoundException e) {
    System.out.println(e.getMessage());
}
```

Go 에러는 Java checked exception과 유사하게 "반드시 처리해야 한다"는 압박을 줍니다. 다만 `_`로 무시할 수 있어 강제성은 Java보다 약합니다. 린터(`errcheck`)로 누락된 에러 처리를 잡는 것이 권장됩니다.
