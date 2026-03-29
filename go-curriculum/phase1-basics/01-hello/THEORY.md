# 01-hello: Go 프로그램의 구조와 fmt 패키지

## Go 언어 소개와 역사

Go는 2007년 Google 내부에서 설계가 시작되어 2009년 공개된 오픈소스 언어입니다.

**설계자:**
- **Rob Pike** — UTF-8 공동 발명자, Plan 9 OS 개발자, Bell Labs 출신
- **Ken Thompson** — Unix 운영체제와 C 언어의 아버지, Bell Labs 출신
- **Robert Griesemer** — V8 JavaScript 엔진, Java HotSpot JVM 개발 참여

이들은 Google 내부에서 C++로 대규모 프로젝트를 빌드할 때의 고통(느린 컴파일, 복잡한 의존성, 언어 복잡성)을 해결하고자 Go를 만들었습니다.

**Go 1.0** 출시: 2012년 3월 — 이후 하위 호환성을 공식 보장
**현재**: Go 1.21+ (2023), 6개월마다 새 버전 출시

---

## Go의 설계 철학

### 단순성 (Simplicity)
"적을수록 좋다(Less is more)." Go는 의도적으로 기능을 제한합니다. 제네릭은 Go 1.18에 추가될 때까지 없었고, 예외(Exception) 대신 에러 반환값을 사용합니다.

### 명시성 (Explicitness)
코드를 읽는 사람이 코드의 동작을 즉시 이해할 수 있어야 합니다. 마법(magic)이 없습니다.

### 컴포지션 (Composition over Inheritance)
클래스 상속 대신 인터페이스와 구조체 임베딩으로 코드를 재사용합니다.

---

## Go 프로그램의 구조

```go
package main          // (1) 패키지 선언

import (              // (2) 임포트
    "fmt"
    "os"
)

func main() {         // (3) 진입점
    fmt.Println("Hello, World!")
}
```

### (1) 패키지 선언 — `package`
모든 Go 파일의 첫 번째 줄은 패키지 선언입니다.
- `package main` — 실행 가능한 프로그램의 패키지. `main()` 함수를 포함해야 합니다.
- `package mathutil` — 라이브러리 패키지. 다른 패키지에서 임포트해서 사용합니다.

### (2) 임포트 — `import`
```go
// 단일 임포트
import "fmt"

// 그룹 임포트 (권장)
import (
    "fmt"
    "os"
    "strings"
)

// 별칭 임포트
import (
    f "fmt"              // f.Println()으로 사용
    _ "database/driver"  // 사이드 이펙트만 (init() 실행)
)
```

**규칙**: 임포트한 패키지를 사용하지 않으면 **컴파일 에러**입니다. Python/Java와 다릅니다.

### (3) main 함수
`func main()` 은 인자와 반환값이 없습니다. 프로세스 종료 코드는 `os.Exit(n)`으로 제어합니다.

```go
func main() {
    // 프로그램 시작
    os.Exit(1)  // 오류 종료 (0 = 정상, 1 이상 = 오류)
}
```

---

## fmt 패키지 상세

`fmt`(format)는 Go의 표준 출력/입력 패키지입니다.

### fmt.Println — 줄바꿈 포함 출력
```go
fmt.Println("안녕하세요")           // "안녕하세요\n"
fmt.Println("값:", 42, true, 3.14) // 인자 사이에 공백 자동 추가
fmt.Println()                       // 빈 줄 출력
```

### fmt.Print — 줄바꿈 없는 출력
```go
fmt.Print("Hello ")
fmt.Print("World\n")  // 수동으로 줄바꿈
// 주의: 두 인자가 모두 문자열이 아닌 경우에만 공백 추가
fmt.Print(1, 2)   // "12" (공백 없음, 둘 다 숫자)
fmt.Print(1, "a") // "1 a" (공백 있음, 타입이 다름)
```

### fmt.Printf — 형식 지정 출력
```go
name := "Go"
version := 1.21
fmt.Printf("언어: %s, 버전: %.2f\n", name, version)
// 출력: 언어: Go, 버전: 1.21
```

### fmt.Sprintf — 문자열로 반환
```go
// 출력하지 않고 형식화된 문자열을 반환
s := fmt.Sprintf("사용자: %s, 나이: %d", "홍길동", 30)
fmt.Println(s)  // 변수에 저장하거나 다른 함수에 전달
```

### fmt.Fprintf — 지정된 Writer에 출력
```go
fmt.Fprintf(os.Stdout, "표준 출력: %s\n", "hello")
fmt.Fprintf(os.Stderr, "표준 에러: %v\n", err)

// 파일에 쓰기
f, _ := os.Create("output.txt")
fmt.Fprintf(f, "파일에 씀: %d\n", 42)
```

### fmt.Errorf — 형식화된 에러 생성
```go
id := 404
err := fmt.Errorf("사용자 ID %d를 찾을 수 없습니다", id)
// err.Error() == "사용자 ID 404를 찾을 수 없습니다"

// %w로 에러 래핑 (Go 1.13+)
wrapped := fmt.Errorf("조회 실패: %w", err)
```

---

## Printf 형식 동사(Format Verbs) 전체 정리

### 범용
| 동사 | 설명 | 예시 |
|------|------|------|
| `%v` | 기본 형식 (모든 타입) | `42`, `true`, `[1 2 3]` |
| `%+v` | 구조체: 필드명 포함 | `{Name:Alice Age:30}` |
| `%#v` | Go 문법 표현 | `main.Person{Name:"Alice"}` |
| `%T` | 타입 출력 | `int`, `string`, `[]int` |

### 정수
| 동사 | 설명 | 예시 (`42`) |
|------|------|------|
| `%d` | 10진수 | `42` |
| `%b` | 2진수 | `101010` |
| `%o` | 8진수 | `52` |
| `%x` | 16진수 (소문자) | `2a` |
| `%X` | 16진수 (대문자) | `2A` |
| `%c` | 문자 (유니코드) | `*` (42='*') |
| `%q` | 따옴표 포함 문자 | `'*'` |

### 부동소수점
| 동사 | 설명 | 예시 (`3.14159`) |
|------|------|------|
| `%f` | 소수점 표기 | `3.141590` |
| `%.2f` | 소수점 2자리 | `3.14` |
| `%e` | 지수 표기 (소문자) | `3.141590e+00` |
| `%E` | 지수 표기 (대문자) | `3.141590E+00` |
| `%g` | 짧은 쪽 선택 | `3.14159` |

### 문자열/바이트
| 동사 | 설명 | 예시 (`"Go"`) |
|------|------|------|
| `%s` | 문자열 그대로 | `Go` |
| `%q` | 쌍따옴표 포함 | `"Go"` |
| `%x` | 16진수 인코딩 | `476f` |

### 기타
| 동사 | 설명 |
|------|------|
| `%t` | 불리언 (`true`/`false`) |
| `%p` | 포인터 주소 (`0xc0000b4010`) |
| `%%` | 리터럴 `%` 문자 |

### 너비와 정렬
```go
fmt.Printf("%10s",  "Go")  // "        Go"  (오른쪽 정렬, 너비 10)
fmt.Printf("%-10s", "Go")  // "Go        "  (왼쪽 정렬)
fmt.Printf("%010d", 42)    // "0000000042"  (0으로 패딩)
fmt.Printf("%+d",   42)    // "+42"         (양수에도 부호 표시)
fmt.Printf("% d",   42)    // " 42"         (양수에 공백)
```

---

## Go 프로그램 빌드와 실행

```bash
# 직접 실행 (컴파일 + 실행, 임시 바이너리)
go run main.go
go run .          # 현재 디렉토리의 main 패키지 실행

# 바이너리 빌드
go build          # 현재 디렉토리, 디렉토리명으로 파일 생성
go build -o myapp # 이름 지정
go build ./...    # 모든 패키지 빌드 (오류 확인용)

# 설치 (GOPATH/bin 또는 GOBIN에 설치)
go install        # 시스템 어디서나 실행 가능하게 설치

# 크로스 컴파일
GOOS=linux GOARCH=amd64 go build -o myapp-linux
GOOS=windows GOARCH=amd64 go build -o myapp.exe
```

---

## Python/Java와의 비교

### main 함수
```python
# Python: 관례적 패턴 (강제 아님)
if __name__ == "__main__":
    print("Hello")
```
```java
// Java: 클래스 안에 정적 메서드
public class Main {
    public static void main(String[] args) { }
}
```
```go
// Go: 최상위 함수, 인자 없음
func main() {
    fmt.Println("Hello")
}
```

### 세미콜론
Go에는 세미콜론이 없습니다(정확히는 렉서가 자동으로 삽입). 그래서 여는 중괄호 `{`는 반드시 같은 줄에 있어야 합니다.
```go
// 올바른 Go 코드:
func main() {
    if true {
    }
}

// 컴파일 에러:
func main()
{           // 에러! { 가 다음 줄에 있으면 안 됨
}
```

### 출력
```python
print("Hello")                    # Python
print(f"이름: {name}")            # f-string
```
```java
System.out.println("Hello");      // Java
System.out.printf("이름: %s%n", name);
```
```go
fmt.Println("Hello")              // Go
fmt.Printf("이름: %s\n", name)
```

### 특수 문자 이스케이프와 Raw String
```go
// 이스케이프 시퀀스
fmt.Println("탭:\t, 줄바꿈:\n, 따옴표:\"")

// Raw string (백틱): 이스케이프 처리 안 됨
rawStr := `이것은 "원시 문자열"
줄바꿈도 그대로, \t도 그대로`
fmt.Println(rawStr)
```

---

## 핵심 정리

1. 모든 Go 파일은 `package` 선언으로 시작
2. `package main` + `func main()` = 실행 가능한 프로그램
3. 임포트하고 사용하지 않으면 컴파일 에러
4. `fmt.Printf`의 형식 동사를 익히면 디버깅이 편해짐
5. `go run`으로 빠르게 실험, `go build`로 배포용 바이너리 생성
