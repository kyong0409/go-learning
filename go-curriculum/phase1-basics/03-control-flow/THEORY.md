# 03-control-flow: 제어 흐름

## Go의 제어 흐름 철학

Go는 제어 흐름 구조를 의도적으로 최소화했습니다. 반복문은 `for` 하나, 조건문은 `if`/`switch`, 점프문은 `break`/`continue`/`goto`/`fallthrough`가 전부입니다. `while`, `do-while`, `until`, 삼항 연산자(`? :`)는 존재하지 않습니다.

---

## for 루프 — 3가지 형태

Go에서 모든 반복은 `for` 하나로 처리합니다.

### 형태 1: C 스타일 (초기화; 조건; 후처리)
```go
for i := 0; i < 5; i++ {
    fmt.Println(i)
}

// 카운트다운
for i := 10; i > 0; i-- {
    fmt.Printf("%d ", i)
}

// 2씩 증가
for i := 0; i <= 100; i += 2 {
    // 짝수 처리
}
```

### 형태 2: while 스타일 (조건만)
```go
n := 1
for n < 1000 {
    n *= 2
}

// 무한 루프 (조건 생략)
for {
    // 영원히 실행
    if someCondition {
        break
    }
}
```

### 형태 3: for-range (컬렉션 순회)
```go
// 슬라이스/배열: (인덱스, 값)
fruits := []string{"사과", "바나나", "체리"}
for i, fruit := range fruits {
    fmt.Printf("[%d] %s\n", i, fruit)
}

// 인덱스만 필요할 때
for i := range fruits {
    fmt.Println(i)
}

// 값만 필요할 때 (인덱스는 _ 로 무시)
for _, fruit := range fruits {
    fmt.Println(fruit)
}
```

---

## for-range 심화

### 맵 순회
```go
scores := map[string]int{"Alice": 95, "Bob": 87}
for name, score := range scores {
    fmt.Printf("%s: %d\n", name, score)
}
// 주의: 맵 순회 순서는 실행할 때마다 다릅니다 (의도적 설계)
```

### 문자열 순회 — rune 단위
```go
msg := "Hello, 한국!"
for i, r := range msg {
    fmt.Printf("인덱스[%2d]: U+%04X = %q\n", i, r, r)
}
// 인덱스[0]:  U+0048 = 'H'
// 인덱스[1]:  U+0065 = 'e'
// ...
// 인덱스[7]:  U+D55C = '한'  ← 인덱스가 7, 8, 9가 아니라 7, 10으로 건너뜀
// 인덱스[10]: U+AD6D = '국'
//
// 한국어 한 글자는 UTF-8로 3바이트이므로 인덱스가 3씩 건너뜁니다.
// range는 바이트 인덱스를 반환하지만, 값은 rune(유니코드 코드포인트)입니다.
```

### 정수 range (Go 1.22+)
```go
// for i := range N  은  for i := 0; i < N; i++  과 동일
// Go 1.22(Feb 2024)에서 추가됨 — 현재 최신: Go 1.26(Feb 2026)
for i := range 5 {
    fmt.Printf("%d ", i)  // 0 1 2 3 4
}
```

### 채널 range
```go
ch := make(chan int, 3)
ch <- 1; ch <- 2; ch <- 3
close(ch)

for v := range ch {  // 채널이 닫힐 때까지 값 수신
    fmt.Println(v)
}
```

---

## break, continue, 라벨

### break와 continue
```go
// continue: 현재 반복 건너뛰기
for i := 1; i <= 10; i++ {
    if i%2 == 0 {
        continue  // 짝수 건너뜀
    }
    fmt.Printf("%d ", i)  // 홀수만 출력
}

// break: 루프 탈출
for i := 0; ; i++ {
    if i >= 5 {
        break
    }
    fmt.Println(i)
}
```

### 라벨(Label) break/continue — 중첩 루프 탈출
```go
// 라벨 break: 지정한 루프까지 탈출
outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if i == 1 && j == 1 {
            break outer  // 바깥 루프(outer)까지 탈출
        }
        fmt.Printf("(%d,%d) ", i, j)
    }
}
// 출력: (0,0) (0,1) (0,2) (1,0)

// 라벨 continue: 지정한 루프의 다음 반복으로
search:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if j == 1 {
            continue search  // 바깥 루프의 다음 반복
        }
        fmt.Printf("(%d,%d) ", i, j)
    }
}
// 출력: (0,0) (1,0) (2,0)
```

---

## if 문

### 기본 구조
```go
if condition {
    // ...
} else if otherCondition {
    // ...
} else {
    // ...
}

// 주의: 괄호 불필요 (Java/C와 다름)
// if (x > 0) { }  // 동작하지만 Go스럽지 않음
// if x > 0 { }    // Go 스타일
```

### init 구문이 있는 if
Go `if`의 강력한 기능입니다. `if` 블록 범위에서만 유효한 변수를 선언하고 바로 조건을 검사할 수 있습니다.

```go
// 기본 패턴
if val := computeValue(); val > 100 {
    fmt.Printf("큰 값: %d\n", val)
} else {
    fmt.Printf("작은 값: %d\n", val)
}
// val은 if/else 블록 밖에서 접근 불가

// 에러 처리에서 가장 많이 쓰임
if result, err := divide(10, 3); err != nil {
    fmt.Printf("에러: %v\n", err)
} else {
    fmt.Printf("결과: %.4f\n", result)
}
// result와 err는 여기서 접근 불가 → 스코프 오염 방지
```

이 패턴은 Go 코드에서 매우 자주 등장합니다. 에러 처리 변수가 외부 스코프를 오염시키지 않는 깔끔한 방법입니다.

---

## switch 문

### Expression switch (값 비교)
```go
day := "월요일"
switch day {
case "월요일":
    fmt.Println("한 주의 시작!")
case "금요일":
    fmt.Println("주말이 코앞!")
case "토요일", "일요일":  // 여러 값을 콤마로 묶기
    fmt.Println("주말!")
default:
    fmt.Println("평범한 평일")
}
```

**Go switch의 핵심**: 각 case는 자동으로 break됩니다. C/Java처럼 break를 쓸 필요가 없습니다.

### init 구문이 있는 switch
```go
switch grade := getScore(); {  // 세미콜론 뒤 표현식 없음 = tagless switch
case grade >= 90:
    fmt.Printf("A학점: %d\n", grade)
case grade >= 80:
    fmt.Printf("B학점: %d\n", grade)
default:
    fmt.Printf("C학점 이하: %d\n", grade)
}
```

### Tagless switch (if-else if 대체)
```go
hour := 14
switch {  // 표현식 없음
case hour < 12:
    fmt.Println("오전")
case hour < 18:
    fmt.Println("오후")
default:
    fmt.Println("저녁")
}
```

### Type switch (인터페이스 타입 확인)
```go
func describe(v interface{}) {
    switch t := v.(type) {
    case int:
        fmt.Printf("정수: %d\n", t)
    case string:
        fmt.Printf("문자열: %q (길이: %d)\n", t, len(t))
    case float64:
        fmt.Printf("실수: %.2f\n", t)
    case bool:
        fmt.Printf("불리언: %t\n", t)
    case nil:
        fmt.Println("nil")
    default:
        fmt.Printf("알 수 없는 타입: %T\n", t)
    }
}

describe(42)        // 정수: 42
describe("hello")   // 문자열: "hello" (길이: 5)
describe(nil)       // nil
```

### fallthrough — 명시적 통과
```go
n := 3
switch n {
case 1:
    fmt.Println("case 1")
case 3:
    fmt.Println("case 3")
    fallthrough  // 다음 case의 조건을 무시하고 실행
case 4:
    fmt.Println("case 4 (fallthrough로 실행됨)")
case 5:
    fmt.Println("case 5")  // 실행 안 됨
}
// 출력:
// case 3
// case 4 (fallthrough로 실행됨)
```

`fallthrough`는 다음 case의 **조건을 검사하지 않고** 바로 본문을 실행합니다. 실제로 사용할 일은 드뭅니다.

---

## Go switch vs C/Java switch

| | Go | C/Java |
|---|---|---|
| 자동 break | O (각 case 후 자동 break) | X (명시적 break 필요) |
| 다음 case 실행 | `fallthrough` 키워드 | break 없으면 자동 fall-through |
| 여러 값 | `case 1, 2, 3:` | `case 1: case 2: case 3:` |
| 타입 switch | O (`v.(type)`) | X |
| 표현식 없는 switch | O (tagless) | X |
| init 구문 | O | X |

---

## 삼항 연산자가 없는 이유

Go에는 `condition ? a : b` 같은 삼항 연산자가 없습니다. Go 설계자들은 삼항 연산자가 복잡한 표현식에서 가독성을 해친다고 판단했습니다.

```go
// 다른 언어:
// max = (a > b) ? a : b

// Go 방식:
max := b
if a > b {
    max = a
}

// 또는 함수로:
func minInt(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

---

## goto — 존재하지만 거의 안 씀

```go
i := 0
loop:
    if i < 3 {
        fmt.Println(i)
        i++
        goto loop
    }
```

`goto`는 Go에 존재하지만 사용을 권장하지 않습니다. 에러 처리 cleanup 코드에서 드물게 사용되는 경우가 있습니다. 일반적인 비즈니스 로직에서는 쓰지 마세요.

---

## Python/Java와의 비교

### while/do-while 없음
```python
# Python
while n < 100:
    n *= 2

# do-while 없으므로 다르게 표현
while True:
    process()
    if not condition:
        break
```
```go
// Go
for n < 100 {
    n *= 2
}

// do-while 대체
for {
    process()
    if !condition {
        break
    }
}
```

### range 비교
```python
# Python
for i in range(10):
    print(i)
for i, v in enumerate(fruits):
    print(i, v)
```
```java
// Java
for (int i : array) { }
```
```go
// Go
for i := range 10 { }   // Go 1.22+ (현재 최신: Go 1.26)
for i, v := range fruits { }
```

### switch 비교
```java
// Java: break 없으면 fall-through
switch (day) {
    case "Monday":
        System.out.println("Monday");
        break;  // break 필수!
    case "Friday":
        System.out.println("Friday");
        // break 없으면 다음 case로 넘어감
}
```
```go
// Go: 자동 break
switch day {
case "Monday":
    fmt.Println("Monday")
    // break 필요 없음
case "Friday":
    fmt.Println("Friday")
}
```

---

## 핵심 정리

1. 반복문은 `for` 하나 — while, do-while 없음
2. `for-range`는 슬라이스, 맵, 문자열, 채널, 정수(1.22+) 모두 순회 가능
3. `if`/`switch`의 init 구문(`;`)으로 변수 스코프를 블록 안으로 제한
4. Go `switch`는 자동 break — C/Java와 반대
5. 삼항 연산자 없음 — `if`를 쓰세요
6. 라벨 break/continue로 중첩 루프를 깔끔하게 탈출
