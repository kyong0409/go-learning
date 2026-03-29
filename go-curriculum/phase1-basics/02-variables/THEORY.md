# 02-variables: 변수, 상수, 타입 시스템

## 변수 선언 4가지 방법

Go에서 변수를 선언하는 방법은 4가지입니다. 각각 용도가 다릅니다.

### 방법 1: `var` 명시적 타입 선언
```go
var age int         // 타입 지정, 제로값으로 초기화
var name string
var price float64
var active bool
```

### 방법 2: `var` + 초기값 (타입 추론)
```go
var city = "서울"     // string으로 추론
var version = 1.21   // float64로 추론
var count = 0        // int로 추론 (주의: 0.0이면 float64)
```

### 방법 3: `:=` 단축 선언 (가장 많이 사용)
```go
score := 95           // int
temperature := 36.5   // float64
greeting := "안녕!"   // string
running := true       // bool
```

### 방법 4: `var` 블록 선언 (여러 변수를 묶어서)
```go
var (
    firstName = "길동"
    lastName  = "홍"
    birthYear = 1990
    height    = 175.5
)
```

---

## `:=` 단축 선언의 규칙과 제한

`:=`는 편리하지만 규칙이 있습니다.

### 규칙 1: 함수 내부에서만 사용 가능
```go
package main

x := 10  // 컴파일 에러! 패키지 수준에서는 var만 사용 가능

func main() {
    y := 20  // 정상
}
```

### 규칙 2: 좌변에 최소 하나의 새 변수가 있어야 함
```go
x := 10
x := 20  // 에러! x는 이미 선언됨

// 다중 할당에서 하나라도 새 변수면 OK
x, y := 10, 20  // x는 재선언, y는 신규 선언이므로 OK
```

### 규칙 3: 타입이 고정됨
```go
x := 42
x = "hello"  // 컴파일 에러! int 변수에 string 할당 불가
```

### 다중 할당과 swap
```go
a, b := 10, 20
fmt.Println(a, b)  // 10 20

// 우아한 값 교환 (임시 변수 불필요)
a, b = b, a
fmt.Println(a, b)  // 20 10

// 함수의 다중 반환값 받기
min, max := findMinMax([]int{3, 1, 4, 1, 5})
```

---

## 기본 타입 상세

### 정수 타입
```go
var i int       // 플랫폼에 따라 32 또는 64비트 (64비트 시스템에서 int64)
var i8  int8    // -128 ~ 127
var i16 int16   // -32,768 ~ 32,767
var i32 int32   // -2,147,483,648 ~ 2,147,483,647
var i64 int64   // -9,223,372,036,854,775,808 ~ 9,223,372,036,854,775,807

var u   uint    // 0 ~ 플랫폼 최대값
var u8  uint8   // 0 ~ 255  (= byte)
var u16 uint16  // 0 ~ 65,535
var u32 uint32  // 0 ~ 4,294,967,295
var u64 uint64  // 0 ~ 18,446,744,073,709,551,615

// 가독성을 위한 숫자 리터럴 구분자 (Go 1.13+)
population := 9_700_000   // 9700000
hexVal := 0xFF            // 255
binVal := 0b1010          // 10
octVal := 0o17            // 15
```

**언제 어떤 타입?**
- 일반적인 경우: `int` (플랫폼 최적화)
- 배열 인덱스, 크기: `int`
- 바이트 처리: `byte` (= `uint8`)
- 유니코드: `rune` (= `int32`)
- 명시적 크기 필요 시 (프로토콜, 파일 포맷): `int32`, `uint64` 등

### 부동소수점 타입
```go
var f32 float32  // IEEE 754, 약 7자리 정밀도
var f64 float64  // IEEE 754, 약 15자리 정밀도 (기본값)

pi := 3.14       // float64로 추론
pi32 := float32(3.14)
```

**`float64`를 기본으로 사용하세요.** `float32`는 GPU 연산, 특수 목적에만 사용합니다.

### 복소수 타입
```go
var c64  complex64   // real: float32, imag: float32
var c128 complex128  // real: float64, imag: float64

z := 3 + 4i          // complex128으로 추론
fmt.Println(real(z)) // 3.0
fmt.Println(imag(z)) // 4.0
```

### bool 타입
```go
var flag bool   // false (제로값)
active := true

// Go의 bool은 정수로 변환 불가 (C와 다름)
// if 1 { }        // 컴파일 에러! 정수를 bool로 암시적 변환 없음
// if flag == 1 { } // 컴파일 에러!
if flag { }         // 올바른 Go 코드
```

### string 타입
```go
s := "Hello, 세계"   // UTF-8 인코딩된 불변 바이트 시퀀스
empty := ""
raw := `줄바꿈\n이
그대로 들어갑니다`   // 원시 문자열 리터럴

fmt.Println(len(s))  // 바이트 수 (한국어 한 글자 = 3바이트)
// s[0] = 'h'        // 컴파일 에러! 문자열은 불변
```

### byte와 rune
```go
var b byte = 'A'   // uint8의 별칭, ASCII 문자
var r rune = '한'  // int32의 별칭, 유니코드 코드포인트

fmt.Println(b)     // 65 (ASCII 코드)
fmt.Printf("%c\n", b)  // A (문자로 출력)
fmt.Println(r)     // 54620 (유니코드 코드포인트)
fmt.Printf("%c\n", r)  // 한
```

---

## 제로값(Zero Value) — Go의 안전망

Go의 모든 변수는 선언 시 자동으로 **제로값**으로 초기화됩니다. C/C++처럼 초기화되지 않은 변수로 인한 쓰레기 값 버그가 없습니다.

| 타입 | 제로값 |
|------|--------|
| `int`, `int8`, ..., `uint64` | `0` |
| `float32`, `float64` | `0.0` |
| `complex64`, `complex128` | `0+0i` |
| `bool` | `false` |
| `string` | `""` (빈 문자열) |
| `byte`, `rune` | `0` |
| 포인터 (`*T`) | `nil` |
| 슬라이스 (`[]T`) | `nil` |
| 맵 (`map[K]V`) | `nil` |
| 채널 (`chan T`) | `nil` |
| 함수 (`func(...)`) | `nil` |
| 인터페이스 | `nil` |
| 구조체 | 모든 필드가 각자의 제로값 |

```go
type Config struct {
    Host    string
    Port    int
    TLS     bool
}
var cfg Config
// cfg.Host == "", cfg.Port == 0, cfg.TLS == false
// 별도 초기화 없이 바로 사용 가능
```

---

## 상수(const)와 iota

### 기본 상수
```go
const Pi = 3.14159265358979  // 타입 없는 상수 (untyped constant)
const MaxRetry = 3
const AppName = "MyApp"

// 타입 있는 상수
const timeout time.Duration = 30 * time.Second
```

**상수는 컴파일 타임에 결정**되므로 함수 호출 결과를 할당할 수 없습니다.
```go
const t = time.Now()  // 컴파일 에러! 런타임 값
```

### iota — 열거형 상수
`iota`는 `const` 블록 안에서 0부터 시작하는 자동 증가 정수입니다.

```go
// 기본 iota: 요일 열거
const (
    Sunday    = iota  // 0
    Monday            // 1
    Tuesday           // 2
    Wednesday         // 3
    Thursday          // 4
    Friday            // 5
    Saturday          // 6
)
```

### iota 패턴 1: 비트 플래그
```go
const (
    ReadPermission  = 1 << iota  // 1 << 0 = 1  (001)
    WritePermission               // 1 << 1 = 2  (010)
    ExecPermission                // 1 << 2 = 4  (100)
)

// 비트 OR로 권한 조합
perm := ReadPermission | WritePermission  // 3 (011)
// 비트 AND로 권한 확인
hasRead := perm&ReadPermission != 0       // true
```

### iota 패턴 2: 크기 단위
```go
const (
    _  = iota               // 0 버리기
    KB = 1 << (10 * iota)   // 1 << 10 = 1,024
    MB                       // 1 << 20 = 1,048,576
    GB                       // 1 << 30
    TB                       // 1 << 40
)
```

### iota 패턴 3: 값 건너뛰기
```go
const (
    StatusUnknown = iota * 10  // 0
    StatusPending               // 10
    StatusActive                // 20
    StatusInactive              // 30
)
```

---

## 타입 변환 (명시적만 허용)

Go는 **암시적 타입 변환이 없습니다**. 항상 명시적으로 변환해야 합니다.

```go
var i int = 42
var f float64 = float64(i)    // int -> float64
var u uint = uint(f)           // float64 -> uint (소수점 잘림!)
var b byte = byte(i)           // int -> byte (오버플로 주의!)
var i32 int32 = int32(i)       // int -> int32

// 이런 코드는 컴파일 에러:
// var f float64 = i            // 에러!
// var sum = i + f              // 에러! int + float64 직접 불가
var sum = f + float64(i)       // 정상
```

### float → int 변환 주의사항
```go
pi := 3.99
n := int(pi)  // 3 (반올림이 아니라 소수점 버림!)
```

### string ↔ []byte ↔ []rune 변환
```go
s := "Hello, 한국"

// string -> []byte (바이트 수준 조작)
b := []byte(s)
b[0] = 'h'
s2 := string(b)  // "hello, 한국"

// string -> []rune (문자 수준 조작)
r := []rune(s)
fmt.Println(len(b))  // 바이트 수: 13 (한국어 2글자 = 6바이트)
fmt.Println(len(r))  // 문자 수: 9
```

### 정수 → string 함정
```go
n := 65
// string(n)은 ASCII 65번 문자 = 'A' 로 변환됨!
wrong := string(rune(n))    // "A" (아스키 문자)

// 숫자를 문자열로 변환하려면:
correct := fmt.Sprintf("%d", n)    // "65"
correct2 := strconv.Itoa(n)         // "65" (더 효율적)
```

---

## 타입 추론 규칙

컴파일러가 오른쪽 값을 보고 타입을 결정합니다.

```go
x := 42          // int (정수 리터럴의 기본 타입)
y := 3.14        // float64 (부동소수점 리터럴의 기본 타입)
z := true        // bool
s := "hello"     // string
c := 'A'         // rune (int32)
b := []byte{...} // []byte

// 주의: 작은 정수도 int로 추론됨
small := 1   // int, int8이 아님!
```

---

## Python/Java와의 비교

### 동적 vs 정적 타이핑
```python
# Python: 런타임에 타입 결정
x = 42
x = "hello"  # 아무 문제없음
```
```go
// Go: 컴파일 타임에 타입 고정
x := 42
x = "hello"  // 컴파일 에러!
```

### 암시적 변환
```python
# Python: 자동 변환
result = 1 + 1.5   # 2.5 (int를 float으로 자동 변환)
```
```java
// Java: 일부 자동 변환 (widening)
int i = 42;
double d = i;  // 자동 변환 (int -> double)
```
```go
// Go: 항상 명시적 변환
var i int = 42
var d float64 = float64(i)  // 반드시 명시적으로
```

### 선언 방식
```python
x = 10       # 타입 없음, 선언과 할당 동일
```
```java
int x = 10;  // 타입 명시 필수
var x = 10;  // Java 10+, 타입 추론
```
```go
var x int = 10  // 명시적
var x = 10      // 타입 추론
x := 10         // 단축 (함수 내부)
```

---

## 핵심 정리

1. `:=`는 함수 내부에서만, 좌변에 새 변수가 하나 이상 있어야 함
2. 모든 변수는 제로값으로 자동 초기화됨 → 초기화 버그 없음
3. 암시적 타입 변환 없음 → `float64(x)` 처럼 명시적으로
4. `const` + `iota`로 열거형을 표현 (enum 키워드 없음)
5. `byte` = `uint8`, `rune` = `int32` — 자주 쓰이므로 외워두세요
