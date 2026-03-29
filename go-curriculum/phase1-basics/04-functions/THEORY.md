# 04-functions: 함수

## 함수 선언과 호출

```go
// 기본 형태: func 이름(매개변수) 반환타입 { }
func greet(name string) string {
    return "안녕하세요, " + name + "!"
}

// 같은 타입의 매개변수는 타입을 한 번만 선언 가능
func add(a, b int) int {
    return a + b
}

// 반환값이 없으면 타입 생략
func printHello() {
    fmt.Println("Hello!")
}

// 호출
result := greet("홍길동")
sum := add(3, 4)
printHello()
```

---

## 다중 반환값 (Multiple Return Values)

Go 함수의 가장 특징적인 기능입니다. 에러 처리의 기반이 됩니다.

```go
// 두 값 반환
func minMax(nums []int) (int, int) {
    min, max := nums[0], nums[0]
    for _, n := range nums[1:] {
        if n < min { min = n }
        if n > max { max = n }
    }
    return min, max
}

// 에러와 함께 반환 (Go의 관용 패턴)
func safeDivide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("0으로 나눌 수 없습니다")
    }
    return a / b, nil
}

// 사용
min, max := minMax([]int{3, 1, 4, 1, 5})

result, err := safeDivide(10, 3)
if err != nil {
    fmt.Printf("에러: %v\n", err)
} else {
    fmt.Printf("결과: %.4f\n", result)
}

// 불필요한 반환값은 _ 로 무시
_, onlyMax := minMax([]int{3, 1, 4})
_, err = safeDivide(5, 0)  // result는 필요 없음
```

**관용 패턴**: 마지막 반환값을 `error`로 사용합니다. `nil`은 에러 없음을 의미합니다.

---

## Named Return Values (명명된 반환값)

반환값에 이름을 붙이면 두 가지 장점이 있습니다:
1. 함수 시작 시 자동으로 제로값으로 초기화
2. `return` 키워드만으로 반환 (naked return)

```go
func circleStats(radius float64) (area, circumference float64) {
    const pi = 3.14159265358979
    area = pi * radius * radius
    circumference = 2 * pi * radius
    return  // naked return: area, circumference를 반환
}

// 에러 처리와 함께
func parseFullName(fullName string) (firstName, lastName string, err error) {
    parts := strings.Fields(fullName)
    if len(parts) < 2 {
        err = fmt.Errorf("전체 이름 필요: %q", fullName)
        return  // firstName="", lastName="", err=...
    }
    firstName = parts[0]
    lastName = parts[len(parts)-1]
    return
}
```

### Named Return 사용 주의사항
```go
// 짧은 함수에서는 OK
func divide(a, b float64) (result float64, err error) {
    if b == 0 {
        err = errors.New("division by zero")
        return
    }
    result = a / b
    return
}

// 긴 함수에서는 혼란 야기 가능 — 명시적 return 권장
func complexOperation() (result int, err error) {
    // ... 50줄의 코드 ...
    return result, err  // 명시적이 더 명확
}
```

Naked return은 **짧은 함수**에서 문서화 목적으로 사용할 때 가장 적합합니다. 긴 함수에서는 어떤 값이 반환되는지 불명확해집니다.

---

## 가변 인자 함수 (Variadic Functions)

`...타입`으로 0개 이상의 인자를 슬라이스로 받습니다.

```go
// 기본 가변 인자
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

sum()           // 0 (인자 없음)
sum(1)          // 1
sum(1, 2, 3)    // 6
sum(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)  // 55

// 가변 인자와 일반 인자 혼합 (가변 인자는 항상 마지막)
func log(level string, messages ...string) {
    for _, msg := range messages {
        fmt.Printf("[%s] %s\n", level, msg)
    }
}
log("INFO", "서버 시작", "포트: 8080")

// 슬라이스를 가변 인자로 전달: 슬라이스명...
nums := []int{1, 2, 3, 4, 5}
result := sum(nums...)  // 슬라이스를 펼쳐서 전달
```

**내부적으로**: 가변 인자는 슬라이스로 처리됩니다. `nums ...int`는 함수 내부에서 `[]int`입니다.

---

## 일급 함수 (First-Class Functions)

Go에서 함수는 일급 시민(first-class citizen)입니다. 변수에 할당하고, 인자로 전달하고, 반환값으로 사용할 수 있습니다.

### 함수 타입 선언
```go
type MathFunc func(int, int) int
type Predicate func(int) bool
type Handler func(http.ResponseWriter, *http.Request)
```

### 함수를 변수에 할당
```go
add := func(a, b int) int { return a + b }
result := add(3, 4)  // 7

// 기존 함수도 변수에 할당 가능
f := fmt.Println  // func(...interface{}) (int, error)
f("hello")
```

### 함수를 인자로 전달 (고차 함수)
```go
func applyOperation(a, b int, op MathFunc) int {
    return op(a, b)
}

applyOperation(10, 3, func(a, b int) int { return a + b })  // 13
applyOperation(10, 3, func(a, b int) int { return a * b })  // 30

// filter: 조건에 맞는 요소만
func filter(nums []int, pred Predicate) []int {
    var result []int
    for _, n := range nums {
        if pred(n) {
            result = append(result, n)
        }
    }
    return result
}

evens := filter([]int{1,2,3,4,5,6}, func(n int) bool {
    return n%2 == 0
})
// [2 4 6]
```

### 함수를 반환값으로 사용 (함수 팩토리)
```go
func makeAdder(n int) func(int) int {
    return func(x int) int {
        return x + n
    }
}

add5 := makeAdder(5)
add10 := makeAdder(10)
fmt.Println(add5(3))   // 8
fmt.Println(add10(3))  // 13
```

---

## 클로저 (Closures)

클로저는 자신이 정의된 스코프의 변수를 **캡처(capture)**하는 함수입니다.

### 카운터 패턴
```go
func makeCounter() func() int {
    count := 0  // 이 변수는 반환된 함수에 캡처됨
    return func() int {
        count++
        return count
    }
}

counter1 := makeCounter()
counter2 := makeCounter()  // 독립적인 count 변수

fmt.Println(counter1())  // 1
fmt.Println(counter1())  // 2
fmt.Println(counter1())  // 3
fmt.Println(counter2())  // 1  (독립적!)
fmt.Println(counter1())  // 4  (계속 유지됨)
```

### 누적 합산 패턴
```go
func makeAccumulator() func(int) int {
    total := 0
    return func(n int) int {
        total += n
        return total
    }
}

acc := makeAccumulator()
fmt.Println(acc(10))  // 10
fmt.Println(acc(20))  // 30
fmt.Println(acc(30))  // 60
```

### 클로저의 변수 공유 (주의!)
```go
// 위험한 패턴: 루프 변수 캡처
funcs := make([]func(), 3)
for i := 0; i < 3; i++ {
    funcs[i] = func() {
        fmt.Println(i)  // i를 직접 참조 (캡처)
    }
}
funcs[0]()  // 3 (i의 최종값!)
funcs[1]()  // 3
funcs[2]()  // 3

// 올바른 패턴: 루프마다 새 변수에 복사
for i := 0; i < 3; i++ {
    i := i  // 새 변수 i에 현재 값 복사 (shadowing)
    funcs[i] = func() {
        fmt.Println(i)
    }
}
funcs[0]()  // 0
funcs[1]()  // 1
funcs[2]()  // 2
```

### 메모이제이션 패턴
```go
func makeFibMemo() func(int) int {
    cache := map[int]int{}
    var fib func(int) int
    fib = func(n int) int {
        if n <= 1 { return n }
        if v, ok := cache[n]; ok { return v }
        result := fib(n-1) + fib(n-2)
        cache[n] = result
        return result
    }
    return fib
}

fib := makeFibMemo()
fmt.Println(fib(50))  // 매우 빠름
```

---

## 익명 함수 (Anonymous Functions)

이름 없이 즉시 정의하고 실행하는 함수입니다.

```go
// 즉시 실행 함수 (IIFE)
result := func(a, b int) int {
    return a + b
}(3, 4)
fmt.Println(result)  // 7

// sort.Slice에서 자주 사용
sort.Slice(words, func(i, j int) bool {
    return words[i] < words[j]
})
```

---

## defer

`defer`는 현재 함수가 반환될 때 실행될 함수를 예약합니다.

### 기본 동작
```go
func example() {
    defer fmt.Println("3: 마지막에 실행")
    defer fmt.Println("2: 그다음에")
    defer fmt.Println("1: 먼저")
    fmt.Println("0: 함수 본문")
}
// 출력 순서:
// 0: 함수 본문
// 1: 먼저          (LIFO: 나중에 등록된 것이 먼저 실행)
// 2: 그다음에
// 3: 마지막에 실행
```

**LIFO(Last In, First Out)** 순서입니다. Stack과 같습니다.

### defer 인자 평가 시점
```go
x := 10
defer fmt.Println("defer에서 x:", x)  // x = 10이 지금 평가됨!
x = 20
fmt.Println("함수 본문에서 x:", x)
// 출력:
// 함수 본문에서 x: 20
// defer에서 x: 10  (defer 등록 시점의 값)
```

### 자원 정리 패턴 (가장 일반적인 사용)
```go
func readFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // 함수 종료 시 무조건 닫힘 (에러 반환 시도 포함)

    // 파일 읽기...
    return nil
}

func withDB() error {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return err
    }
    defer db.Close()

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()  // 성공 시 Commit 후에는 Rollback이 no-op

    // ... 트랜잭션 작업 ...
    return tx.Commit()
}
```

### defer와 panic/recover
```go
func safeOperation() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("패닉 복구: %v\n", r)
        }
    }()
    panic("의도적인 패닉!")
}
// panic이 발생해도 프로그램이 종료되지 않음
```

---

## init() 함수

패키지가 로드될 때 자동으로 실행되는 특수 함수입니다.

```go
package mypackage

var config Config

func init() {
    // 패키지 초기화 코드
    config = loadConfig()
    log.Println("mypackage 초기화 완료")
}
```

**init() 규칙:**
- 인자와 반환값 없음
- 하나의 패키지에 여러 개 가능 (같은 파일에도)
- 실행 순서: 변수 초기화 → init() → main()
- 직접 호출 불가 (`init()` 호출 시 컴파일 에러)

---

## 재귀와 Go의 스택 관리

Go는 **증가하는 스택(growable stack)**을 사용합니다. 초기 고루틴 스택 크기는 약 8KB이며, 필요하면 자동으로 확장됩니다. 따라서 재귀가 깊어도 스택 오버플로가 Java처럼 쉽게 발생하지 않습니다.

```go
func factorial(n int) int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n-1)
}

// 단, 진짜 무한 재귀는 메모리가 다 찰 때까지 계속 증가합니다
func infinite() {
    infinite()  // 결국 메모리 부족으로 패닉
}
```

---

## Python/Java와의 비교

### 람다 vs 클로저
```python
# Python: lambda는 단일 표현식만 가능
double = lambda x: x * 2
# 여러 줄이 필요하면 def 사용
def makeAdder(n):
    def adder(x):
        return x + n
    return adder
```
```java
// Java: 람다 (함수형 인터페이스 필요)
Function<Integer, Integer> double = x -> x * 2;
// 클로저는 effectively final 변수만 캡처
int n = 5;
Function<Integer, Integer> adder = x -> x + n;
// n = 10;  // 에러! effectively final이어야 함
```
```go
// Go: 클로저로 어떤 변수도 캡처 가능, 수정도 가능
makeAdder := func(n int) func(int) int {
    return func(x int) int { return x + n }
}
// 또는 패키지 수준 함수로
```

### try-finally vs defer
```python
# Python
try:
    f = open("file.txt")
    process(f)
finally:
    f.close()  # 항상 실행
```
```java
// Java
try (FileInputStream f = new FileInputStream("file.txt")) {
    process(f);
}  // try-with-resources: 자동으로 close() 호출
```
```go
// Go
f, err := os.Open("file.txt")
if err != nil { return err }
defer f.Close()  // 함수 종료 시 자동 실행
process(f)
```

### 다중 반환값
```python
# Python: 튜플 반환
def divide(a, b):
    if b == 0:
        return None, "division by zero"
    return a / b, None
result, err = divide(10, 3)
```
```java
// Java: 다중 반환 없음 → 객체로 래핑
record Result<T>(T value, Exception error) {}
```
```go
// Go: 언어 차원 지원
func divide(a, b float64) (float64, error) {
    if b == 0 { return 0, errors.New("division by zero") }
    return a / b, nil
}
result, err := divide(10, 3)
```

---

## 핵심 정리

1. 다중 반환값은 Go의 에러 처리 패턴의 핵심 — `(result, error)` 패턴을 익혀야 함
2. Named return은 짧은 함수의 문서화에 유용하지만 남용 금지
3. 클로저는 외부 변수를 캡처함 — 루프 변수 캡처 시 의도치 않은 동작 주의
4. `defer`는 자원 정리의 표준 패턴 — open 직후 바로 defer close 작성
5. `defer`는 LIFO 순서, 인자는 등록 시점에 평가됨
