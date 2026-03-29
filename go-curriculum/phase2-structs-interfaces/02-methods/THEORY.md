# 02-methods: 메서드 (Methods)

> Go에서 메서드는 특정 타입에 연결된 함수입니다. 클래스의 `self`/`this` 대신 명시적 리시버(receiver)를 사용합니다.

---

## 1. 메서드 정의

```go
// 문법: func (리시버변수 리시버타입) 메서드명(파라미터) 반환타입
func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}
```

리시버는 함수 이름 앞에 괄호로 묶어 씁니다. 리시버 변수명은 관례적으로 타입 이름의 첫 글자 소문자를 씁니다 (`r`, `c`, `a` 등).

---

## 2. 값 리시버 vs 포인터 리시버

### 값 리시버 (Value Receiver)

```go
func (r Rectangle) Area() float64 {
    return r.Width * r.Height  // r은 복사본
}
```

- 구조체의 **복사본**을 받는다
- 메서드 내에서 필드를 수정해도 원본에 영향 없음
- **읽기 전용** 연산에 적합
- 작은 구조체에서 복사 비용이 낮을 때 적합

### 포인터 리시버 (Pointer Receiver)

```go
func (r *Rectangle) Scale(factor float64) {
    r.Width *= factor   // 원본 수정
    r.Height *= factor  // 원본 수정
}
```

- 구조체의 **포인터(원본 참조)**를 받는다
- 메서드 내에서 원본을 수정할 수 있다
- 큰 구조체에서 복사 비용을 절약할 수 있다
- **상태를 변경**하는 메서드에 사용

### 값/포인터 리시버 선택 가이드라인

| 상황 | 권장 리시버 |
|------|------------|
| 구조체 필드를 수정해야 함 | 포인터 리시버 |
| 구조체가 크다 (복사 비용 고려) | 포인터 리시버 |
| 동일 타입에 포인터 리시버 메서드가 이미 있다 | 포인터 리시버로 통일 |
| 순수 읽기 연산, 작은 구조체 | 값 리시버 |

> **실용적 규칙**: 한 타입의 메서드 중 하나라도 포인터 리시버를 쓴다면, 나머지도 포인터 리시버로 통일하는 것이 권장됩니다.

### Go의 자동 변환

```go
r := Rectangle{Width: 2, Height: 3}

// r이 값 타입이어도 포인터 리시버 메서드 호출 가능
// Go가 (&r).Scale(3) 으로 자동 변환
r.Scale(3)

// *r이 포인터 타입이어도 값 리시버 메서드 호출 가능
// Go가 (*rPtr).Area() 로 자동 변환
rPtr := &Rectangle{Width: 5, Height: 5}
fmt.Println(rPtr.Area())
```

---

## 3. 메서드 세트 (Method Set) 규칙

| 타입 | 포함하는 메서드 세트 |
|------|-------------------|
| `T` (값 타입) | 값 리시버 `(T)` 메서드만 |
| `*T` (포인터 타입) | 값 리시버 `(T)` + 포인터 리시버 `(*T)` 메서드 모두 |

이 규칙은 인터페이스 충족 여부에 영향을 줍니다.

```go
type Mover interface {
    Move()
}

type Car struct{ Speed int }

func (c *Car) Move() { c.Speed++ }  // 포인터 리시버

var m Mover

car := Car{}
// m = car    // 컴파일 에러! Car의 메서드 세트에 Move()가 없음
m = &car      // OK: *Car의 메서드 세트에 Move() 있음
```

---

## 4. 관례적 생성자: NewXxx() 패턴

Go에는 클래스 생성자가 없지만, `NewXxx()` 함수로 생성자 역할을 합니다.

```go
type BankAccount struct {
    owner   string  // 소문자: 패키지 외부 직접 접근 불가
    balance float64
}

// NewBankAccount는 BankAccount 생성자입니다.
// 유효성 검사를 포함하고 에러를 반환합니다.
func NewBankAccount(owner string, initialBalance float64) (*BankAccount, error) {
    if owner == "" {
        return nil, fmt.Errorf("계좌 주인 이름은 필수입니다")
    }
    if initialBalance < 0 {
        return nil, fmt.Errorf("초기 잔액은 0 이상이어야 합니다")
    }
    return &BankAccount{
        owner:   owner,
        balance: initialBalance,
    }, nil
}
```

**NewXxx() 패턴 특징**
- 포인터를 반환하는 것이 일반적 (비공개 필드 접근, 포인터 리시버 메서드 호출 용이)
- 유효성 검사와 기본값 설정을 한 곳에서 처리
- "인터페이스를 받아들이고, 구조체를 반환하라" 원칙의 반환부 구현

---

## 5. String() 메서드: fmt.Stringer 인터페이스

`fmt.Stringer` 인터페이스는 `String() string` 메서드 하나로 구성됩니다. 이를 구현하면 `fmt.Println`, `fmt.Printf("%v")` 등이 자동으로 이 메서드를 호출합니다.

```go
// fmt.Stringer 인터페이스 (표준 라이브러리)
// type Stringer interface {
//     String() string
// }

func (r Rectangle) String() string {
    return fmt.Sprintf("Rectangle(%.2f x %.2f)", r.Width, r.Height)
}

rect := Rectangle{Width: 5, Height: 3}
fmt.Println(rect)         // "Rectangle(5.00 x 3.00)" 자동 출력
fmt.Printf("%v\n", rect)  // 동일
fmt.Printf("%s\n", rect)  // 동일
```

> Python의 `__str__`, Java의 `toString()`에 해당합니다.

---

## 6. 메서드 값(Method Value)과 메서드 표현식(Method Expression)

```go
rect := Rectangle{Width: 4, Height: 5}

// 메서드 값: 특정 인스턴스에 바인딩된 함수
areaFn := rect.Area  // func() float64 타입
fmt.Println(areaFn())  // rect.Area() 와 동일

// 메서드 표현식: 타입에서 가져오는 함수 (리시버를 첫 번째 인자로 전달)
areaExpr := Rectangle.Area  // func(Rectangle) float64 타입
fmt.Println(areaExpr(rect))  // rect.Area() 와 동일
```

메서드 값은 콜백 함수로 전달할 때 유용합니다.

```go
shapes := []Rectangle{{1, 2}, {3, 4}, {5, 6}}
areas := make([]float64, len(shapes))
for i, s := range shapes {
    areas[i] = s.Area  // 나중에 호출할 함수 저장
}
```

---

## 7. 비구조체 타입에 메서드 붙이기

Go에서는 구조체뿐 아니라 어떤 타입에도 메서드를 붙일 수 있습니다. 단, 같은 패키지에 정의된 타입이어야 합니다.

```go
// 기본 타입에 별칭을 만들어 메서드 추가
type Celsius float64
type Fahrenheit float64

func (c Celsius) ToFahrenheit() Fahrenheit {
    return Fahrenheit(c*9/5 + 32)
}

func (f Fahrenheit) ToCelsius() Celsius {
    return Celsius((f - 32) * 5 / 9)
}

temp := Celsius(100)
fmt.Printf("%.1f°C = %.1f°F\n", temp, temp.ToFahrenheit())  // 100.0°C = 212.0°F
```

슬라이스 타입에도 메서드를 붙일 수 있습니다.

```go
type IntSlice []int

func (s IntSlice) Sum() int {
    total := 0
    for _, v := range s {
        total += v
    }
    return total
}

nums := IntSlice{1, 2, 3, 4, 5}
fmt.Println(nums.Sum())  // 15
```

---

## 8. Python / Java 비교

### Python self vs Go 리시버

```python
# Python
class Rectangle:
    def __init__(self, width, height):
        self.width = width
        self.height = height

    def area(self):          # self는 암묵적 첫 번째 인자
        return self.width * self.height

    def scale(self, factor): # self를 통해 원본 수정
        self.width *= factor
        self.height *= factor
```

```go
// Go
type Rectangle struct {
    Width, Height float64
}

func (r Rectangle) Area() float64 {       // 값 리시버: 복사본
    return r.Width * r.Height
}

func (r *Rectangle) Scale(factor float64) { // 포인터 리시버: 원본 수정
    r.Width *= factor
    r.Height *= factor
}
```

**핵심 차이**
- Python `self`는 항상 원본 참조이지만, Go는 값/포인터 리시버를 명시적으로 선택한다
- Go의 값 리시버는 Python에서 `self`로 받은 후 아무것도 수정하지 않는 것과 의미상 같지만, Go는 진짜 복사본이라 실수로 수정해도 원본에 영향이 없다

### Java this vs Go 리시버

```java
// Java
public class BankAccount {
    private double balance;

    public void deposit(double amount) {
        this.balance += amount;  // this는 암묵적
    }
}
```

```go
// Go
type BankAccount struct {
    balance float64
}

func (a *BankAccount) Deposit(amount float64) {
    a.balance += amount  // 명시적 리시버
}
```

### __str__ vs String()

```python
# Python
class Rectangle:
    def __str__(self):
        return f"Rectangle({self.width} x {self.height})"
```

```go
// Go
func (r Rectangle) String() string {
    return fmt.Sprintf("Rectangle(%.2f x %.2f)", r.Width, r.Height)
}
```

Go의 `String() string`은 `fmt.Stringer` 인터페이스를 충족시킵니다. 명시적 `implements` 선언 없이 메서드만 있으면 자동으로 인터페이스를 만족합니다 (다음 챕터의 핵심 개념).
