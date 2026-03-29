# 06-strings: 문자열과 UTF-8

## Go 문자열의 본질

Go의 문자열은 **불변(immutable)의 바이트 슬라이스**입니다. 내부적으로는 `[]byte`와 유사하지만 변경할 수 없다는 점이 다릅니다.

```go
s := "Hello, 세계"

// 문자열은 []byte처럼 인덱싱 가능하지만 결과는 byte(uint8)
fmt.Println(s[0])         // 72 ('H'의 ASCII 코드)
fmt.Printf("%c\n", s[0]) // H

// 직접 수정은 불가
// s[0] = 'h'  // 컴파일 에러: cannot assign to s[0]

// 수정하려면 []byte로 변환 후 다시 string으로
b := []byte(s)
b[0] = 'h'
s2 := string(b)  // "hello, 세계"
```

**왜 불변인가?** 문자열이 불변이면 여러 변수가 같은 문자열 데이터를 안전하게 공유할 수 있습니다. 복사 없이 포인터만 전달해도 됩니다.

---

## UTF-8 인코딩 — Go의 핵심 설계

Rob Pike와 Ken Thompson은 Go 개발 중 UTF-8 인코딩 자체를 공동 발명했습니다(1992). Go 언어는 소스 코드 파일 자체가 UTF-8이며, 문자열도 기본적으로 UTF-8입니다.

### UTF-8의 특성
- ASCII 문자(0-127)는 1바이트 (하위 호환성 완벽)
- 한국어, 중국어, 일본어 등은 일반적으로 3바이트
- 이모지 등은 4바이트
- 코드 단위와 코드 포인트가 분리됨

```
'H'  = 0x48         = 1바이트
'é'  = 0xC3 0xA9    = 2바이트
'한' = 0xED 0x95 0x9C = 3바이트
'😀' = 0xF0 0x9F 0x98 0x80 = 4바이트
```

### len()은 바이트 수를 반환
```go
s := "Hello, 세계"
fmt.Println(len(s))                        // 13 (바이트 수)
// 'H','e','l','l','o',',',' '  = 7바이트
// '세' = 3바이트, '계' = 3바이트 → 총 13바이트

fmt.Println(utf8.RuneCountInString(s))     // 9 (문자 수)
```

---

## byte vs rune

| 타입 | 별칭 | 크기 | 용도 |
|------|------|------|------|
| `byte` | `uint8` | 1바이트 | ASCII 문자, 원시 바이트 처리 |
| `rune` | `int32` | 4바이트 | 유니코드 코드포인트 |

```go
var b byte = 'A'   // ASCII 65
var r rune = '한'  // 유니코드 U+D55C = 54620

fmt.Printf("byte: %d = %c\n", b, b)  // 65 = A
fmt.Printf("rune: %d = %c\n", r, r)  // 54620 = 한
```

---

## 문자열 순회: byte vs rune

### 인덱스 접근 — byte 단위
```go
s := "Hello, 한국"
for i := 0; i < len(s); i++ {
    fmt.Printf("s[%2d] = 0x%02X\n", i, s[i])
}
// s[ 0] = 0x48  ('H')
// s[ 1] = 0x65  ('e')
// ...
// s[ 7] = 0xED  ('한'의 첫 번째 바이트)
// s[ 8] = 0x95
// s[ 9] = 0x9C
// s[10] = 0xEA  ('국'의 첫 번째 바이트)
```

### for-range — rune 단위
```go
s := "Hello, 한국"
for i, r := range s {
    fmt.Printf("인덱스[%2d]: U+%04X = %q\n", i, r, r)
}
// 인덱스[ 0]: U+0048 = 'H'
// 인덱스[ 1]: U+0065 = 'e'
// ...
// 인덱스[ 7]: U+D55C = '한'  ← 인덱스 7 (3바이트 차지)
// 인덱스[10]: U+AD6D = '국'  ← 인덱스 10 (7+3)
```

`for-range`는 UTF-8을 디코딩하여 한 번에 하나의 rune을 반환합니다.

---

## string, []byte, []rune 변환

```go
s := "Hello, 한국"

// string -> []byte: 바이트 수준 조작, 네트워크 전송, 파일 I/O
b := []byte(s)
fmt.Println(len(b))  // 13 (바이트 수)

// string -> []rune: 문자 단위 인덱싱, 문자 수 계산
r := []rune(s)
fmt.Println(len(r))   // 9 (문자 수)
fmt.Printf("%c\n", r[7])  // 한 (7번째 문자에 직접 접근 가능)

// 주의: s[7]은 바이트, r[7]은 rune
fmt.Println(s[7])     // 237 (한의 첫 바이트, 숫자)
fmt.Printf("%c\n", r[7])  // 한 (실제 문자)

// 변환 비용: 새 슬라이스를 만들고 데이터를 복사
// 반복적인 변환은 성능에 영향
```

---

## strings 패키지 주요 함수

```go
import "strings"
text := "  Hello, Go 세계!  "

// ─── 검색 ───
strings.Contains(text, "Go")         // true
strings.ContainsAny(text, "aeiou")   // true (어느 하나라도 포함)
strings.ContainsRune(text, '한')     // 코드포인트로 검색
strings.HasPrefix(text, "  Hello")   // true
strings.HasSuffix(text, "!  ")       // true
strings.Index(text, "Go")            // 8 (바이트 인덱스)
strings.LastIndex(text, "l")         // 마지막 'l'의 위치
strings.Count(text, "l")             // 'l' 개수

// ─── 변환 ───
strings.ToUpper(text)                // "  HELLO, GO 세계!  "
strings.ToLower(text)                // "  hello, go 세계!  "
strings.TrimSpace(text)              // "Hello, Go 세계!"
strings.Trim("!Hello!", "!")         // "Hello" (양쪽의 '!' 제거)
strings.TrimLeft("   hello", " ")   // "hello"
strings.TrimRight("hello   ", " ")  // "hello"
strings.TrimPrefix("Hello, Go!", "Hello, ")  // "Go!"
strings.TrimSuffix("Hello, Go!", ", Go!")    // "Hello"
strings.Replace(text, "l", "L", 2)  // 처음 2개만 교체
strings.ReplaceAll(text, "l", "L")  // 모두 교체

// ─── 분리 / 결합 ───
parts := strings.Split("a,b,c", ",")   // ["a" "b" "c"]
parts2 := strings.SplitN("a,b,c", ",", 2)  // ["a" "b,c"] (최대 2개)
fields := strings.Fields("  a  b  c  ") // ["a" "b" "c"] (공백 기준)
strings.Join([]string{"a", "b", "c"}, "-")  // "a-b-c"
strings.Repeat("Go! ", 3)               // "Go! Go! Go! "

// ─── 비교 ───
strings.EqualFold("Go", "go")   // true (대소문자 무시 비교)
strings.Compare("a", "b")       // -1, 0, 1
```

---

## strings.Builder — 효율적인 문자열 연결

```go
// 비효율: + 연산자는 매번 새 문자열을 할당
result := ""
for i := 0; i < 1000; i++ {
    result += fmt.Sprintf("item%d", i)  // O(n²) 시간복잡도!
}

// 효율: strings.Builder 사용
var sb strings.Builder
sb.Grow(10000)  // 미리 용량 예약 (선택사항)
for i := 0; i < 1000; i++ {
    fmt.Fprintf(&sb, "item%d", i)
    // sb.WriteString("text")   // 문자열 추가
    // sb.WriteByte('\n')        // 바이트 추가
    // sb.WriteRune('한')        // rune 추가
}
result := sb.String()
sb.Reset()  // 내용 초기화 (재사용 가능)
```

**성능 비교 (1000개 연결 기준):**
- `+` 연산: 매 반복마다 새 메모리 할당 → O(n²)
- `strings.Builder`: 내부 버퍼에 추가 → O(n)
- `strings.Join`: 슬라이스가 이미 있을 때 가장 간결

---

## strconv 패키지 — 타입 ↔ 문자열 변환

```go
import "strconv"

// ─── 정수 변환 ───
s := strconv.Itoa(42)          // "42"  (int -> string, 가장 빠름)
n, err := strconv.Atoi("123") // 123, nil

// 더 범용적: ParseInt(s, 진법, 비트크기)
n64, err := strconv.ParseInt("FF", 16, 64)    // 255 (16진수)
n64, err = strconv.ParseInt("0b1010", 0, 64)  // 10 (자동 감지)
n64, err = strconv.ParseInt("42", 10, 64)     // 42

// ─── 실수 변환 ───
f, err := strconv.ParseFloat("3.14159", 64)   // 3.14159
s2 := strconv.FormatFloat(f, 'f', 3, 64)      // "3.142"
// 형식: 'f'=소수점, 'e'=지수, 'g'=짧은 쪽

// ─── 불리언 변환 ───
b, err := strconv.ParseBool("true")   // true
b, err = strconv.ParseBool("1")       // true
b, err = strconv.ParseBool("false")   // false
b, err = strconv.ParseBool("0")       // false
s3 := strconv.FormatBool(true)        // "true"

// ─── 진법 변환 ───
s4 := strconv.FormatInt(255, 2)   // "11111111" (2진수)
s5 := strconv.FormatInt(255, 16)  // "ff" (16진수)
```

### 에러 처리
```go
n, err := strconv.Atoi("abc")
if err != nil {
    var numErr *strconv.NumError
    if errors.As(err, &numErr) {
        fmt.Printf("변환 실패: 함수=%s, 입력=%s\n", numErr.Func, numErr.Num)
    }
}
```

---

## unicode/utf8 패키지

```go
import (
    "unicode"
    "unicode/utf8"
)

// 문자 분류
unicode.IsLetter('A')    // true
unicode.IsLetter('한')   // true
unicode.IsDigit('5')     // true
unicode.IsSpace(' ')     // true
unicode.IsUpper('A')     // true
unicode.IsLower('a')     // true
unicode.IsPunct('!')     // true

// 대소문자 변환
unicode.ToUpper('a')     // 'A'
unicode.ToLower('A')     // 'a'

// UTF-8 유틸리티
utf8.RuneCountInString("Hello, 한국")  // 9 (문자 수)
utf8.RuneLen('한')                      // 3 (바이트 수)
utf8.ValidString("Hello")              // true
utf8.ValidString("\xFF\xFE")           // false (잘못된 UTF-8)
```

---

## 정규표현식 — regexp 패키지 기초

```go
import "regexp"

// 패턴 컴파일 (프로그램 시작 시 한 번만)
re := regexp.MustCompile(`\b\w+@\w+\.\w+\b`)  // 이메일 패턴

// 패키지 수준 변수로 선언하는 게 관용적
var emailRe = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

// 사용
fmt.Println(emailRe.MatchString("user@example.com"))  // true
fmt.Println(emailRe.FindString("연락처: user@example.com 입니다"))
// "user@example.com"

matches := emailRe.FindAllString(text, -1)  // 모든 매칭

// 치환
result := re.ReplaceAllString(text, "[이메일]")

// Compile vs MustCompile
re2, err := regexp.Compile(`[invalid`)  // 에러 반환
re3 := regexp.MustCompile(`\d+`)        // 에러 시 panic (프로그램 초기화에 적합)
_ = re2
_ = re3
```

---

## Python/Java와의 비교

### str vs string
```python
# Python str: 유니코드 문자열 (Python 3)
s = "Hello, 한국"
len(s)     # 9 (문자 수)
s[7]       # '한' (문자 직접 접근)
s.upper()  # 메서드 방식
```
```go
// Go string: 바이트 슬라이스
s := "Hello, 한국"
len(s)                          // 13 (바이트 수!)
utf8.RuneCountInString(s)       // 9 (문자 수)
[]rune(s)[7]                    // '한' ([]rune 변환 후 접근)
strings.ToUpper(s)              // 패키지 함수 방식
```

### 인코딩/디코딩
```python
# Python: 명시적 인코딩 필요
s = "Hello"
b = s.encode("utf-8")   # bytes
s2 = b.decode("utf-8")  # str
```
```go
// Go: string은 이미 UTF-8
s := "Hello"
b := []byte(s)    // string -> []byte (내부적으로 UTF-8 바이트)
s2 := string(b)  // []byte -> string
// 다른 인코딩이 필요하면 golang.org/x/text 패키지 사용
```

### 문자열 포맷팅
```python
name, age = "Alice", 30
s = f"이름: {name}, 나이: {age}"  # f-string
```
```java
String s = String.format("이름: %s, 나이: %d", name, age);
```
```go
s := fmt.Sprintf("이름: %s, 나이: %d", name, age)
```

---

## 핵심 정리

1. Go 문자열 = 불변 UTF-8 바이트 슬라이스
2. `len(s)` = 바이트 수, `utf8.RuneCountInString(s)` = 문자 수
3. `for-range`로 문자열 순회 = rune 단위, `s[i]`로 접근 = byte 단위
4. 문자 단위 처리가 필요하면 `[]rune(s)`로 변환
5. 문자열 연결이 많으면 `strings.Builder` 사용 (O(n) vs `+`의 O(n²))
6. `strconv.Atoi`/`Itoa`는 int↔string 변환의 가장 빠른 방법
