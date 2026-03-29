# 09-packages: 패키지와 모듈 시스템

## 패키지 시스템 기본 개념

Go에서 **패키지(package)**는 코드 구성의 기본 단위입니다.

핵심 규칙:
- **하나의 디렉토리 = 하나의 패키지**
- 같은 디렉토리의 모든 `.go` 파일은 같은 패키지에 속함
- 파일 이름과 패키지 이름은 무관 (관례상 디렉토리명과 일치)
- `package main`만 실행 가능한 프로그램을 만들 수 있음

```
myproject/
├── main.go          (package main)
├── mathutil/
│   ├── math.go      (package mathutil)
│   └── math_test.go (package mathutil)
└── stringutil/
    └── string.go    (package stringutil)
```

---

## 가시성 규칙: 대문자 = exported

Go의 접근 제어는 단 하나의 규칙으로 이루어집니다:

**대문자로 시작하는 이름 = exported (외부 패키지에서 접근 가능)**
**소문자로 시작하는 이름 = unexported (패키지 내부 전용)**

```go
package mathutil

// Exported: 외부에서 사용 가능
const Pi = 3.14159265358979
var ErrDivisionByZero = errors.New("0으로 나누기")

type Calculator struct {
    Result float64  // exported 필드
    history []float64  // unexported 필드 (패키지 내부만)
}

func Add(a, b int) int { return a + b }  // exported

// unexported: 패키지 내부에서만
func validateInput(n int) bool { return n >= 0 }
var maxIterations = 1000
type internalState struct { ... }
```

```go
package main

import "myproject/mathutil"

func main() {
    fmt.Println(mathutil.Pi)        // OK
    fmt.Println(mathutil.Add(1, 2)) // OK
    // mathutil.validateInput(5)    // 컴파일 에러!
    // mathutil.maxIterations       // 컴파일 에러!
}
```

이 규칙은 함수, 변수, 상수, 타입, 구조체 필드, 메서드, 인터페이스 메서드 모두에 적용됩니다.

---

## import 경로와 별칭

```go
// 표준 라이브러리: 경로만 쓰면 됨
import "fmt"
import "strings"
import "encoding/json"   // 패키지명은 json (마지막 부분)
import "net/http"        // 패키지명은 http

// 외부 패키지: 모듈 경로 포함
import "github.com/gin-gonic/gin"      // 패키지명: gin
import "go.uber.org/zap"               // 패키지명: zap

// 로컬 패키지
import "myproject/mathutil"            // go.mod의 module명 기준

// 별칭 (이름 충돌 방지 또는 긴 이름 단축)
import (
    "fmt"
    myfmt "github.com/myorg/myfmt"  // myfmt.Println()으로 사용
    _ "database/sql/driver"          // 사이드 이펙트만 (init() 실행)
    . "math"                         // Sqrt() 처럼 패키지명 없이 사용 (비권장)
)
```

---

## Go 모듈 시스템

Go 1.11에서 도입되고 1.16에서 기본이 된 모듈 시스템입니다.

### go mod init — 모듈 초기화
```bash
mkdir myproject
cd myproject
go mod init github.com/username/myproject
# go.mod 파일 생성됨
```

### go.mod 파일
```
module github.com/username/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    go.uber.org/zap v1.26.0
)
```

### go.sum 파일
```
# 의존성의 암호화 해시 (보안 검증)
# 직접 수정 금지
github.com/gin-gonic/gin v1.9.1 h1:4idEAncQnU5cB7BeUSHh1wRVXRW2zxwjT...
```

### 주요 모듈 명령어
```bash
# 패키지 추가
go get github.com/gin-gonic/gin
go get github.com/gin-gonic/gin@v1.9.1  # 버전 지정
go get github.com/gin-gonic/gin@latest  # 최신 버전

# go.mod/go.sum 정리 (사용하지 않는 의존성 제거)
go mod tidy

# 의존성을 vendor/ 디렉토리에 복사 (오프라인 빌드)
go mod vendor

# 의존성 그래프 확인
go mod graph

# 캐시 확인
go env GOMODCACHE
```

---

## 패키지 구조 설계

### init() 함수와 패키지 초기화 순서
```go
package mypackage

import "fmt"

var globalVar = initVar()  // 1. 패키지 수준 변수 초기화

func initVar() string {
    return "initialized"
}

func init() {  // 2. init() 실행 (변수 초기화 후)
    fmt.Println("mypackage init()")
    // DB 연결, 설정 로드, 등록 등에 사용
}

// 실행 순서:
// 1. import된 패키지의 변수 초기화
// 2. import된 패키지의 init()
// 3. 현재 패키지의 변수 초기화
// 4. 현재 패키지의 init()
// 5. main()
```

```go
// 하나의 파일에 여러 init() 가능
func init() { fmt.Println("init 1") }
func init() { fmt.Println("init 2") }
// 파일 내 선언 순서대로 실행됨
```

### 순환 의존성 금지
```
// 허용 안 됨:
packageA → packageB → packageA  // 순환! 컴파일 에러

// 올바른 구조:
packageC (공통 기반)
    ↑         ↑
packageA    packageB  // 둘 다 C에만 의존
```

Go 컴파일러는 순환 의존성을 컴파일 에러로 처리합니다. 순환이 발생하면 패키지 구조를 재설계해야 합니다.

---

## internal 패키지

`internal` 디렉토리는 부모 디렉토리 이상의 코드만 임포트할 수 있습니다.

```
myproject/
├── main.go
├── internal/
│   └── auth/
│       └── jwt.go    (package auth)
└── web/
    └── handler.go
```

```go
// main.go: OK (myproject의 하위이므로)
import "myproject/internal/auth"

// 외부 패키지: 컴파일 에러!
// import "myproject/internal/auth"  // 접근 불가
```

내부 구현 세부사항을 외부에 노출하지 않으면서도 패키지로 분리하고 싶을 때 사용합니다.

---

## 표준 라이브러리 주요 패키지

```go
// ─── 입출력 ───
"fmt"           // 형식화 출력/입력
"os"            // OS 파일, 환경변수, 프로세스
"io"            // I/O 인터페이스 (Reader, Writer)
"bufio"         // 버퍼드 I/O
"log"           // 기본 로거

// ─── 문자열/텍스트 ───
"strings"       // 문자열 조작
"strconv"       // 타입 ↔ 문자열 변환
"regexp"        // 정규표현식
"unicode"       // 유니코드 분류

// ─── 데이터 인코딩 ───
"encoding/json" // JSON 인코딩/디코딩
"encoding/xml"  // XML
"encoding/csv"  // CSV
"encoding/base64"

// ─── 네트워크/HTTP ───
"net/http"      // HTTP 서버/클라이언트
"net/url"       // URL 파싱
"net"           // TCP/UDP 소켓

// ─── 자료구조 ───
"sort"          // 정렬
"container/heap"
"container/list"

// ─── 수학 ───
"math"          // 수학 함수 (Sqrt, Sin, Cos 등)
"math/rand"     // 난수
"math/big"      // 임의 정밀도 정수/실수

// ─── 동시성 ───
"sync"          // Mutex, WaitGroup, Once
"sync/atomic"   // 원자적 연산
"context"       // 컨텍스트 전파, 취소

// ─── 시간 ───
"time"          // 시간, 타이머, 기간

// ─── 테스트 ───
"testing"       // 유닛 테스트, 벤치마크

// ─── 경로 ───
"path/filepath" // 파일 경로 (OS에 맞는 구분자)
"path"          // URL 경로 (슬래시 기준)
```

---

## 패키지 설계 관례

### 패키지 이름 관례
```go
// 짧고 소문자, 단수형
package math     // O
package strings  // O
package myutil   // O
package MyUtil   // X (대문자 금지)
package my_util  // X (언더스코어 금지)
package utils    // X (너무 범용적, 피하세요)
package helpers  // X (마찬가지)
```

### 패키지명과 사용
```go
// 패키지명이 용도를 설명해야 함
// user.User 보다 user.Profile이 나음
// 반복 피하기: http.HTTPClient -> http.Client

package http
type Client struct { ... }    // http.Client (O)
type HTTPClient struct { ... } // http.HTTPClient (X, 중복)
```

### 문서 주석
```go
// Package mathutil provides basic mathematical utilities.
package mathutil

// Add returns the sum of a and b.
// Both a and b must be non-negative integers.
func Add(a, b int) int {
    return a + b
}
```

`go doc mathutil.Add` 명령으로 문서를 볼 수 있습니다.

---

## 실용 예제: 패키지 만들기

```go
// 파일: mathutil/math.go
package mathutil

import "errors"

// 공개 상수
const (
    Pi = 3.14159265358979
    E  = 2.71828182845905
)

// 공개 에러
var ErrDivisionByZero = errors.New("0으로 나누기 불가")

// 공개 함수
func Add(a, b int) int { return a + b }

func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, ErrDivisionByZero
    }
    return a / b, nil
}

// 비공개 헬퍼 함수
func isValid(n int) bool { return n >= 0 }
```

```go
// 파일: main.go
package main

import (
    "errors"
    "fmt"
    "myproject/mathutil"
)

func main() {
    fmt.Println(mathutil.Add(3, 4))   // 7
    fmt.Println(mathutil.Pi)           // 3.14159...

    result, err := mathutil.Divide(10, 0)
    if errors.Is(err, mathutil.ErrDivisionByZero) {
        fmt.Println("0으로 나눌 수 없습니다")
    }
    _ = result
}
```

---

## Python/Java와의 비교

### 패키지/모듈 시스템
```python
# Python: __init__.py가 패키지 표시
# mypackage/__init__.py 존재 시 패키지
# from mypackage import myfunction
# import mypackage.submodule
```
```java
// Java: package 선언 + 디렉토리 구조 일치
// package com.example.myapp;
// import com.example.myapp.util.Helper;
// Maven/Gradle로 의존성 관리
```
```go
// Go: 디렉토리 = 패키지 (자동)
// go.mod로 모듈 정의
// import "github.com/user/repo/subpkg"
// go get으로 의존성 관리
```

### 접근 제어
```python
# Python: 관례적 (_prefix = 비공개)
# _private_func() (강제 아님, 관례)
# __mangled (name mangling)
```
```java
// Java: public, protected, private, (default) 키워드
public class MyClass {
    public void publicMethod() { }
    private void privateMethod() { }
    protected void protectedMethod() { }
}
```
```go
// Go: 첫 글자 대소문자만으로 결정
func PublicFunc() { }   // exported (public)
func privateFunc() { }  // unexported (private)
// protected 없음 (패키지 내부이거나 public이거나)
```

---

## 핵심 정리

1. 디렉토리 = 패키지, 대문자 = exported (공개), 소문자 = unexported (비공개)
2. `go mod init`으로 모듈 초기화, `go.mod`가 의존성 관리의 기준
3. `go mod tidy`로 불필요한 의존성 정리 — PR 전에 습관적으로 실행
4. `init()`은 패키지 초기화용, 직접 호출 불가
5. 순환 의존성은 컴파일 에러 — 패키지 구조 재설계 필요
6. `internal/`로 패키지를 분리하되 외부 노출을 제한 가능
7. 패키지 이름은 짧고 소문자, 도구(`utils`, `helpers`)보다 도메인 이름 선호
