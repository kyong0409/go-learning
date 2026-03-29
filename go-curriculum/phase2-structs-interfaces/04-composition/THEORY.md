# 04-composition: 컴포지션 (Composition)

> Go는 상속 대신 임베딩과 컴포지션으로 코드를 재사용합니다. "상속보다 컴포지션을 선호하라"는 오래된 원칙을 언어 수준에서 강제합니다.

---

## 1. 왜 Go는 상속을 뺐는가

### 깊은 계층의 문제점

```
// Java 스타일 상속 계층
Vehicle
  └── MotorVehicle
        └── Car
              └── ElectricCar
                    └── LuxuryElectricCar
```

- 계층이 깊어질수록 최상위 클래스 변경이 모든 하위 클래스에 영향을 준다
- 어떤 클래스의 메서드가 어디서 정의됐는지 추적하기 어렵다
- 하위 클래스가 상위 클래스의 내부 구현에 강하게 결합된다 (취약한 기반 클래스 문제)

### 다이아몬드 문제

```
// Python의 다중 상속
class A:
    def method(self): return "A"

class B(A):
    def method(self): return "B"

class C(A):
    def method(self): return "C"

class D(B, C):  # D.method()는 B의 것인가, C의 것인가?
    pass
```

Go는 구조체 상속을 없애 이 문제를 원천 차단했습니다.

---

## 2. 구조체 임베딩 (Embedding)

임베딩은 필드명 없이 타입명만 쓰는 것입니다.

```go
type Animal struct {
    Name string
    Age  int
}

func (a Animal) Describe() string {
    return fmt.Sprintf("이름: %s, 나이: %d살", a.Name, a.Age)
}

// Dog는 Animal을 임베딩합니다.
type Dog struct {
    Animal        // 임베딩: 필드명 없이 타입명만
    Breed string  // Dog 고유 필드
}
```

임베딩과 일반 필드의 차이:

```go
// 임베딩
type Dog struct {
    Animal         // Animal의 메서드와 필드가 Dog로 프로모션됨
    Breed string
}

// 일반 필드 (임베딩 아님)
type Dog struct {
    Animal Animal  // 필드명이 있음 — 프로모션 없음
    Breed  string
}
```

---

## 3. 메서드 프로모션 (Promotion)

임베딩된 타입의 메서드와 필드가 외부 타입으로 자동으로 올라옵니다.

```go
dog := Dog{
    Animal: Animal{Name: "바둑이", Age: 3},
    Breed:  "진돗개",
}

// 프로모션된 필드 접근 — 두 가지 방법 모두 가능
fmt.Println(dog.Name)         // 프로모션된 접근 (권장)
fmt.Println(dog.Animal.Name)  // 명시적 접근

// 프로모션된 메서드 호출
fmt.Println(dog.Describe())   // Animal.Describe()가 Dog로 프로모션됨
```

**인터페이스 충족도 프로모션됩니다**

```go
type Describer interface {
    Describe() string
}

var d Describer = dog  // Dog가 Describe()를 가지므로 Describer 충족
```

---

## 4. 섀도잉 (Shadowing): 오버라이드 효과

외부 타입이 임베딩된 타입과 같은 이름의 메서드를 정의하면 외부 타입의 메서드가 우선합니다.

```go
func (a Animal) Speak() string {
    return fmt.Sprintf("%s(이)가 소리를 냅니다.", a.Name)
}

// Dog가 Speak를 재정의 (섀도잉)
func (d Dog) Speak() string {
    return fmt.Sprintf("%s(이)가 '멍멍!' 짖습니다.", d.Name)
}

dog := Dog{Animal: Animal{Name: "바둑이"}}
fmt.Println(dog.Speak())         // "바둑이(이)가 '멍멍!' 짖습니다." — Dog의 것
fmt.Println(dog.Animal.Speak())  // "바둑이(이)가 소리를 냅니다." — Animal의 것 (명시적 접근)
```

섀도잉은 Java의 `@Override`와 비슷하지만, Go에서는 임베딩이 상속이 아니므로 "오버라이드"라는 용어보다 "섀도잉"이 더 정확합니다.

---

## 5. 임베딩 vs 명시적 필드: 언제 무엇을 쓸지

**임베딩을 선택할 때**

```go
// "is-a" 관계이거나 모든 메서드를 프로모션하고 싶을 때
type NamedWriter struct {
    io.Writer  // Writer의 모든 메서드가 NamedWriter로 프로모션됨
    Name string
}
```

**명시적 필드를 선택할 때**

```go
// "has-a" 관계이거나 의도적으로 분리하고 싶을 때
type UserService struct {
    db     Database  // 명시적 필드 — db.Save()처럼 접근
    logger Logger    // 명시적 필드 — logger.Log()처럼 접근
}
```

**실용적 판단 기준**

| 상황 | 선택 |
|------|------|
| 내부 타입의 인터페이스를 외부 타입도 만족시키고 싶다 | 임베딩 |
| 내부 타입의 메서드를 직접 호출하게 하고 싶다 | 임베딩 |
| 내부 타입을 구현 세부사항으로 숨기고 싶다 | 명시적 필드 |
| 같은 이름의 메서드를 가진 두 타입을 함께 쓴다 | 명시적 필드 |

---

## 6. 다중 임베딩과 이름 충돌 해결

두 개 이상의 타입을 임베딩할 수 있습니다.

```go
type Logger struct{ prefix string }
func (l *Logger) Log(msg string) { fmt.Printf("[%s] %s\n", l.prefix, msg) }

type Validator struct{ errors []string }
func (v *Validator) IsValid() bool { return len(v.errors) == 0 }

// 두 타입을 모두 임베딩
type UserService struct {
    Logger     // Log() 메서드 프로모션
    Validator  // IsValid() 메서드 프로모션
    users map[string]string
}

svc := &UserService{Logger: Logger{prefix: "SVC"}}
svc.Log("사용자 생성")   // Logger.Log() 직접 호출
svc.IsValid()            // Validator.IsValid() 직접 호출
```

**이름 충돌 해결**

두 임베딩 타입이 같은 이름의 메서드를 가지면 컴파일 에러가 납니다. 명시적으로 접근해야 합니다.

```go
type A struct{}
func (a A) Hello() string { return "A" }

type B struct{}
func (b B) Hello() string { return "B" }

type C struct {
    A
    B
}

c := C{}
// c.Hello()         // 컴파일 에러: ambiguous selector c.Hello
c.A.Hello()          // OK: 명시적 접근
c.B.Hello()          // OK: 명시적 접근
```

---

## 7. 인터페이스 임베딩

인터페이스도 다른 인터페이스를 임베딩할 수 있습니다.

```go
type Reader interface { Read() string }
type Writer interface { Write(s string) }
type Closer interface { Close() error }

// 합성 인터페이스
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    ReadWriter  // ReadWriter 자체가 Reader + Writer
    Closer
}

// ReadWriteCloser를 충족하려면 Read, Write, Close 세 메서드가 모두 있어야 함
type Buffer struct{ data strings.Builder; closed bool }
func (b *Buffer) Read() string       { return b.data.String() }
func (b *Buffer) Write(s string)     { b.data.WriteString(s) }
func (b *Buffer) Close() error       { b.closed = true; return nil }

var rwc ReadWriteCloser = &Buffer{}  // 세 메서드 모두 있으므로 충족
```

표준 라이브러리에서 이 패턴을 광범위하게 사용합니다: `io.ReadWriter`, `io.ReadWriteCloser`, `io.ReadSeeker`, `io.ReadWriteSeeker` 등.

---

## 8. 실전 컴포지션 패턴

### 데코레이터 패턴

기존 구현을 감싸서 기능을 추가합니다.

```go
type UpperCaseWriter struct {
    w io.Writer  // 내부 Writer에 위임
}

func (u UpperCaseWriter) Write(p []byte) (int, error) {
    return u.w.Write([]byte(strings.ToUpper(string(p))))
}

// 사용: os.Stdout을 감싸서 대문자 변환 기능 추가
ucw := UpperCaseWriter{w: os.Stdout}
fmt.Fprintln(ucw, "hello world")  // "HELLO WORLD" 출력
```

### 위임(Delegation) 패턴

인터페이스를 임베딩해 기본 구현을 위임하고 일부만 오버라이드합니다.

```go
type LoggingWriter struct {
    io.Writer          // 기본 구현 위임
    bytesWritten int
}

func (lw *LoggingWriter) Write(p []byte) (int, error) {
    n, err := lw.Writer.Write(p)  // 원래 Writer에 위임
    lw.bytesWritten += n          // 추가 동작
    return n, err
}
```

### 다중 알림(MultiNotifier) 패턴

인터페이스 슬라이스로 여러 구현체를 조합합니다.

```go
type Notifier interface {
    Notify(message string) error
}

type MultiNotifier struct {
    notifiers []Notifier
}

func (m *MultiNotifier) AddNotifier(n Notifier) {
    m.notifiers = append(m.notifiers, n)
}

func (m *MultiNotifier) Notify(message string) error {
    for _, n := range m.notifiers {
        if err := n.Notify(message); err != nil {
            return err
        }
    }
    return nil
}

// MultiNotifier 자체도 Notifier를 구현 — 재귀적 조합 가능
var _ Notifier = &MultiNotifier{}
```

---

## 9. Python / Java 비교

### Python 다중 상속 vs Go 임베딩

```python
# Python: 다중 상속
class Logger:
    def log(self, msg): print(msg)

class Validator:
    def is_valid(self): return True

class UserService(Logger, Validator):  # 다중 상속
    pass

svc = UserService()
svc.log("hello")    # Logger에서 상속
svc.is_valid()      # Validator에서 상속
```

```go
// Go: 다중 임베딩
type UserService struct {
    Logger     // 임베딩
    Validator  // 임베딩
}

svc := &UserService{}
svc.Log("hello")   // Logger에서 프로모션
svc.IsValid()      // Validator에서 프로모션
```

**핵심 차이**
- Python 다중 상속은 `UserService`가 `Logger`와 `Validator`의 하위 타입이 됩니다 (`isinstance(svc, Logger)` == True)
- Go 임베딩에서 `UserService`는 `Logger`나 `Validator`의 하위 타입이 아닙니다. 인터페이스로만 다형성을 표현합니다
- Python MRO(Method Resolution Order) 같은 복잡한 규칙이 Go에는 없습니다

### Python mixin vs Go 임베딩

```python
# Python mixin
class JSONMixin:
    def to_json(self):
        import json
        return json.dumps(self.__dict__)

class User(JSONMixin):  # mixin 적용
    def __init__(self, name): self.name = name
```

```go
// Go: mixin 개념 없음 — 임베딩 또는 독립 함수로 처리
type JSONMarshaler struct{}
func (j JSONMarshaler) ToJSON(v any) (string, error) {
    data, err := json.Marshal(v)
    return string(data), err
}

type User struct {
    JSONMarshaler  // 임베딩으로 ToJSON 메서드 프로모션
    Name string
}
```

### Java super() vs Go의 명시적 접근

```java
// Java: super()로 부모 메서드 호출
class Dog extends Animal {
    @Override
    public String speak() {
        return super.speak() + " 그리고 멍멍!";
    }
}
```

```go
// Go: 임베딩된 타입에 직접 접근
func (d Dog) Speak() string {
    return d.Animal.Speak() + " 그리고 멍멍!"  // 명시적으로 Animal.Speak() 호출
}
```

Go에는 `super()`가 없습니다. 임베딩된 타입 이름으로 직접 접근합니다. 이것이 오히려 어느 타입의 메서드를 호출하는지 명확하게 보여줍니다.
