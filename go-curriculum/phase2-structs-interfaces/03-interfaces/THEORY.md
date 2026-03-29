# 03-interfaces: 인터페이스 (Interfaces)

> Go 인터페이스의 핵심은 암묵적 충족(implicit satisfaction)입니다. `implements` 키워드 없이 메서드만 구현하면 자동으로 인터페이스를 만족합니다.

---

## 1. 인터페이스의 핵심: 암묵적 충족

```go
// 인터페이스 정의
type Shape interface {
    Area() float64
    Perimeter() float64
}

// Circle은 Shape 인터페이스를 "암묵적으로" 충족합니다.
// "implements Shape" 선언이 없어도 됩니다.
type Circle struct{ Radius float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

// Shape 인터페이스 변수에 Circle을 담을 수 있다
var s Shape = Circle{Radius: 5}
fmt.Println(s.Area())
```

**implements 키워드가 없는 이유**

명시적 선언이 없으면 인터페이스를 정의한 패키지와 구현 패키지 사이에 **의존성이 생기지 않습니다**. 나중에 정의된 인터페이스에 기존 타입이 자동으로 맞춰질 수 있습니다. 이것이 Go 생태계에서 인터페이스가 자연스럽게 조합되는 이유입니다.

---

## 2. 덕 타이핑과의 차이

Go의 암묵적 인터페이스 충족은 Python의 덕 타이핑과 비슷해 보이지만 **컴파일 타임에 검증**된다는 점이 다릅니다.

```python
# Python 덕 타이핑 — 런타임에만 오류 발견
def print_area(shape):  # 타입 힌트 없으면 런타임 실패
    print(shape.area())
```

```go
// Go — 컴파일 타임에 Shape 인터페이스 충족 여부를 검증
func printArea(s Shape) {   // Shape를 충족하지 않으면 컴파일 에러
    fmt.Println(s.Area())
}
```

인터페이스 충족 여부를 컴파일 타임에 강제로 확인하고 싶을 때:

```go
// 컴파일 타임 인터페이스 충족 검증 관용구
var _ Shape = (*Circle)(nil)  // Circle이 Shape를 충족하지 않으면 컴파일 에러
```

---

## 3. 인터페이스 설계 원칙: 작게, 1-2개 메서드

Go 표준 라이브러리의 대부분 인터페이스는 메서드가 1-2개입니다.

```go
// 좋은 예: 단일 책임, 작고 명확한 인터페이스
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 나쁜 예: 너무 많은 메서드 — 구현하기 어렵고, 모킹하기도 어렵다
type BigInterface interface {
    Read() string
    Write(s string)
    Close() error
    Flush() error
    Seek(offset int64) error
    Size() int64
}
```

**인터페이스가 작을수록 좋은 이유**
- 구현하기 쉽다
- 테스트 더블(mock)을 만들기 쉽다
- 더 많은 타입이 자연스럽게 충족할 수 있다

---

## 4. 표준 라이브러리의 대표 인터페이스

### io.Reader

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

파일, 네트워크 연결, HTTP 요청 바디, 문자열 등 모든 "읽을 수 있는 것"이 이 인터페이스를 구현합니다.

```go
// strings.NewReader: 문자열을 io.Reader로 변환
r := strings.NewReader("Hello, Go!")
buf := make([]byte, 4)
for {
    n, err := r.Read(buf)
    if n > 0 { fmt.Printf("읽음: %q\n", buf[:n]) }
    if err == io.EOF { break }
}
```

### io.Writer

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

파일, 네트워크 연결, HTTP 응답 바디, `bytes.Buffer` 등이 구현합니다.

```go
// os.Stdout도 io.Writer를 구현
fmt.Fprintln(os.Stdout, "Hello!")

// 커스텀 Writer
type UpperWriter struct{ w io.Writer }
func (u UpperWriter) Write(p []byte) (int, error) {
    return u.w.Write([]byte(strings.ToUpper(string(p))))
}
```

### io.Copy: 인터페이스 조합의 힘

```go
// Reader에서 Writer로 데이터를 복사
// 파일→파일, 파일→HTTP응답, HTTP요청→파일 모두 동일한 코드
func io.Copy(dst Writer, src Reader) (int64, error)
```

### fmt.Stringer

```go
type Stringer interface {
    String() string
}
```

`fmt.Println`, `fmt.Printf("%v")` 등이 자동으로 호출합니다.

### error

```go
type error interface {
    Error() string
}
```

Go의 모든 에러는 이 인터페이스를 충족합니다. 단 하나의 메서드만 있습니다.

### sort.Interface

```go
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}
```

`sort.Sort(data)` 함수는 이 인터페이스를 받습니다. 어떤 슬라이스든 세 메서드를 구현하면 정렬할 수 있습니다.

### http.Handler

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Go HTTP 서버의 핵심 인터페이스. 하나의 메서드로 라우터, 미들웨어, 핸들러 등 모든 HTTP 처리 컴포넌트를 표현합니다.

---

## 5. 인터페이스 합성 (Interface Composition)

작은 인터페이스를 임베딩해 큰 인터페이스를 만들 수 있습니다.

```go
type Reader interface { Read(p []byte) (int, error) }
type Writer interface { Write(p []byte) (int, error) }
type Closer interface { Close() error }

// 합성
type ReadWriter interface {
    Reader  // Reader 임베딩
    Writer  // Writer 임베딩
}

type ReadWriteCloser interface {
    ReadWriter
    Closer
}
```

표준 라이브러리의 `io.ReadWriter`, `io.ReadWriteCloser`, `io.ReadSeeker` 등이 이 패턴으로 정의되어 있습니다.

---

## 6. 빈 인터페이스 (any / interface{})

```go
// any는 interface{}의 별칭 (Go 1.18+)
// 어떤 타입이든 담을 수 있다
var v any = 42
v = "hello"
v = []int{1, 2, 3}
```

**빈 인터페이스를 지양해야 하는 이유**
- 컴파일 타임 타입 안전성을 잃는다
- 사용 시마다 타입 단언이 필요해 코드가 복잡해진다
- `any`를 받는 함수는 실제로 무엇을 기대하는지 알 수 없다

**빈 인터페이스를 사용해도 되는 경우**
- JSON 역직렬화 (`map[string]any`)
- 제네릭이 없던 시절의 범용 컨테이너 (Go 1.18+ 이후에는 제네릭 권장)
- `fmt.Println`, `fmt.Sprintf` 같은 가변 인자 함수

---

## 7. 타입 단언 (Type Assertion)

인터페이스 변수에서 구체 타입의 값을 꺼냅니다.

```go
var i any = "Go 언어"

// 단일 반환값 형태 — 실패 시 panic 발생
str := i.(string)  // 위험: i가 string이 아니면 panic

// 안전한 ok 패턴 — 실패해도 panic 없음
str, ok := i.(string)
if ok {
    fmt.Println("문자열:", str)
} else {
    fmt.Println("string이 아님")
}
```

---

## 8. 타입 스위치 (Type Switch)

여러 타입을 처리할 때 일련의 타입 단언보다 깔끔합니다.

```go
func describe(v any) string {
    switch val := v.(type) {
    case int:
        return fmt.Sprintf("정수: %d", val)
    case float64:
        return fmt.Sprintf("실수: %.2f", val)
    case string:
        return fmt.Sprintf("문자열: %q", val)
    case bool:
        return fmt.Sprintf("불리언: %v", val)
    case Shape:                          // 인터페이스 타입도 가능
        return fmt.Sprintf("도형: 넓이=%.2f", val.Area())
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("알 수 없는 타입: %T", val)
    }
}
```

`val := v.(type)` 에서 `val`은 각 `case` 블록 안에서 해당 타입의 값입니다.

---

## 9. 인터페이스의 nil 주의사항

Go 인터페이스는 내부적으로 `(타입, 값)` 쌍으로 구성됩니다. **타입이 있으면 값이 nil이어도 인터페이스 자체는 nil이 아닙니다.**

```go
// nil 인터페이스: 타입도 nil, 값도 nil
var s Shape
fmt.Println(s == nil)  // true

// 함정: nil 포인터를 인터페이스에 담으면 인터페이스는 nil이 아님
var c *Circle = nil   // nil 포인터
var s2 Shape = c      // (타입=*Circle, 값=nil)
fmt.Println(s2 == nil) // false! — 타입 정보(*Circle)가 있으므로
```

**실수하기 쉬운 패턴**

```go
// 잘못된 패턴: nil 포인터를 error 인터페이스에 담아 반환
func badFunc() error {
    var err *MyError = nil
    return err  // nil이 아닌 error 반환! 호출자가 err != nil로 체크하면 true
}

// 올바른 패턴
func goodFunc() error {
    return nil  // 순수 nil 반환
}
```

---

## 10. "인터페이스를 받아들이고, 구조체를 반환하라" 상세 설명

```go
// 함수 파라미터: 인터페이스로 받는다 — 유연성 극대화
func SaveUser(db Database, user User) error {
    return db.Save(user)
}

// 함수 반환: 구체 타입(포인터)을 반환한다 — 사용 편의성 극대화
func NewUserService(db Database) *UserService {
    return &UserService{db: db}
}

// 이렇게 하지 마세요: 반환 타입을 인터페이스로 하면
// 호출자가 구체 타입에만 있는 메서드에 접근할 수 없다
func NewUserService(db Database) UserServiceInterface { // 피할 것
    return &UserService{db: db}
}
```

**왜 반환은 구체 타입인가**

- 반환 타입이 인터페이스이면 호출자는 인터페이스에 없는 메서드를 쓰려면 타입 단언이 필요하다
- 라이브러리가 반환 타입을 나중에 인터페이스로 넓히는 건 쉽지만, 구체 타입으로 좁히는 건 하위 호환성을 깬다
- 예외: `errors.New()`처럼 불투명한 값을 반환하고 싶을 때는 인터페이스 반환이 적절하다

---

## 11. Python / Java 비교

### Python ABC vs Go 인터페이스

```python
# Python: 명시적 상속이 필요
from abc import ABC, abstractmethod

class Shape(ABC):
    @abstractmethod
    def area(self) -> float: ...

class Circle(Shape):  # 명시적으로 Shape를 상속
    def area(self) -> float:
        return 3.14 * self.radius ** 2
```

```go
// Go: 명시적 선언 없음 — 메서드만 있으면 자동 충족
type Shape interface {
    Area() float64
}

type Circle struct{ Radius float64 }
func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }
// Shape 인터페이스 충족 선언 없음 — 자동으로 충족됨
```

### Python Protocol vs Go 인터페이스

```python
# Python 3.8+ Protocol: 구조적 타이핑 (Go와 가장 유사)
from typing import Protocol

class Shape(Protocol):
    def area(self) -> float: ...

# Circle이 Protocol을 명시적으로 상속하지 않아도 됨
class Circle:
    def area(self) -> float: return 3.14 * self.radius ** 2
```

Python `Protocol`이 Go 인터페이스와 가장 유사하지만, Go는 런타임이 아닌 컴파일 타임에 검증합니다.

### Java interface vs Go interface

```java
// Java: implements 명시 필요
interface Shape {
    double area();
}

class Circle implements Shape {  // 명시적 implements
    public double area() { return Math.PI * radius * radius; }
}
```

```go
// Go: implements 없음
type Shape interface { Area() float64 }
type Circle struct{ Radius float64 }
func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }
```

**핵심 차이**: Java는 구현 클래스가 인터페이스에 의존하지만, Go는 인터페이스가 구현 타입을 모릅니다. 두 패키지가 서로 모르는 상태에서도 인터페이스로 연결할 수 있습니다.
