# Phase 1: Go 기초 - 학습 로드맵

## Go 언어를 배우는 이유

Go(Golang)는 2009년 Google에서 공개한 오픈소스 언어입니다. 현재 클라우드 인프라와 백엔드 개발의 핵심 언어로 자리잡았습니다.

**Go가 지배하는 생태계:**
- **컨테이너/오케스트레이션**: Docker, Kubernetes, containerd
- **클라우드 도구**: Terraform, Prometheus, Grafana, etcd
- **네트워크/프록시**: Caddy, Traefik, Istio, Envoy(일부)
- **데이터베이스 클라이언트/서버**: CockroachDB, InfluxDB, TiDB
- **CLI 도구**: GitHub CLI, kubectl, Helm

Python이나 Java를 알고 있다면 Go를 배우는 이유는 명확합니다. Go는 이 두 언어의 장점을 취하면서 단점을 제거했습니다. Python의 단순한 문법 + Java의 정적 타입 안전성 + C의 성능에 가까운 실행 속도를 제공합니다.

---

## Go의 핵심 원칙 5가지

### 1. 단순성 (Simplicity)
Go의 키워드는 단 25개입니다. Python은 35개, Java는 50개입니다. 언어 자체가 작기 때문에 전체를 머릿속에 담을 수 있습니다.

```go
// 예약어 전체: break, case, chan, const, continue, default, defer,
// else, fallthrough, for, func, go, goto, if, import, interface,
// map, package, range, return, select, struct, switch, type, var
```

### 2. 명시성 (Explicitness)
Go는 숨겨진 동작을 싫어합니다. 타입 변환은 항상 명시적이고, 에러는 반환값으로 처리하며, 상속 없이 컴포지션만 사용합니다.

```go
// Python: x = 1 + 1.5  (암시적 int->float 변환)
// Go:     x := 1 + int(1.5)  (명시적 변환 필요)
```

### 3. 컴포지션 (Composition)
Go에는 클래스 상속이 없습니다. 대신 인터페이스와 구조체 임베딩으로 유연한 설계를 합니다. "has-a" 관계를 선호하고 "is-a" 관계를 피합니다.

### 4. 동시성 (Concurrency)
고루틴(goroutine)과 채널(channel)은 Go의 핵심입니다. Phase 2에서 다루지만, Go를 배우는 가장 큰 이유 중 하나입니다.

```go
go func() { /* 새 고루틴에서 실행 */ }()
```

### 5. 도구 중심 (Tooling)
`go fmt`, `go vet`, `go test`, `go build` — 표준 도구가 언어에 내장되어 있어 팀 전체가 일관된 스타일을 유지합니다.

---

## Phase 1 학습 목표와 순서

| 단계 | 디렉토리 | 핵심 개념 | 전제 지식 |
|------|----------|-----------|-----------|
| 1 | `01-hello` | 프로그램 구조, fmt 패키지 | 없음 |
| 2 | `02-variables` | 타입 시스템, 제로값, const/iota | 01 |
| 3 | `03-control-flow` | for/if/switch, range | 02 |
| 4 | `04-functions` | 다중반환, 클로저, defer | 03 |
| 5 | `05-data-structures` | 배열, 슬라이스, 맵 | 04 |
| 6 | `06-strings` | UTF-8, byte/rune, strings 패키지 | 05 |
| 7 | `07-pointers` | 포인터, 값/참조 전달 | 05 |
| 8 | `08-errors` | error 인터페이스, 에러 패턴 | 04, 07 |
| 9 | `09-packages` | 패키지 시스템, 모듈, 가시성 | 01~08 |

---

## 다른 언어에서 올 때 5가지 마인드셋 전환

### 1. 예외(Exception) 대신 에러 값(Error Value)
```go
// Java/Python 방식 사고 (Go에는 없음):
// try { result = divide(a, b); } catch (Exception e) { ... }

// Go 방식:
result, err := divide(a, b)
if err != nil {
    // 에러 처리
}
```
에러는 특별한 것이 아니라 그냥 함수의 반환값입니다.

### 2. 클래스 대신 구조체 + 메서드 + 인터페이스
```go
// Java 방식 사고: class Animal { ... }, class Dog extends Animal { ... }

// Go 방식:
type Dog struct { Name string }
func (d Dog) Speak() string { return "Woof" }
// 인터페이스는 암시적으로 구현됨 (implements 키워드 없음)
```

### 3. while/do-while 없음 — for 하나로 모든 반복
```go
// Python: while condition: ...
// Go:
for condition {
    // while처럼 동작
}
```

### 4. 암시적 타입 변환 없음
```go
var x int = 42
var y float64 = float64(x)  // 반드시 명시적 변환
// var y float64 = x  // 컴파일 에러!
```

### 5. 사용하지 않는 변수/임포트는 컴파일 에러
```go
import "fmt"  // 사용하지 않으면 컴파일 에러
x := 42       // 사용하지 않으면 컴파일 에러
_ = x         // 빈 식별자로 의도적 무시
```

---

## 환경 설정 가이드

### Go 설치
```bash
# 공식 사이트: https://go.dev/dl/
# Windows: .msi 설치 파일 사용
# macOS:   brew install go
# Linux:   tar 파일 다운로드 후 /usr/local/go에 압축 해제

go version  # 설치 확인
```

### VS Code 설정 (권장)
1. VS Code 설치
2. "Go" 확장 설치 (Google 공식, ID: `golang.go`)
3. VS Code에서 Go 파일 열기 → 우측 하단 "Install All" 클릭
4. gopls(언어 서버), dlv(디버거), staticcheck 자동 설치

### GOPATH vs 모듈 모드
- **GOPATH 모드** (구버전, Go 1.11 이전): 모든 코드를 `~/go/src/`에 위치
- **모듈 모드** (현재 표준, Go 1.16+): `go.mod` 파일이 있는 어느 디렉토리에서나 작업

```bash
# 모듈 초기화
mkdir myproject && cd myproject
go mod init github.com/username/myproject

# go.mod 파일이 생성됨:
# module github.com/username/myproject
# go 1.26
```

### 기본 명령어
```bash
go run main.go        # 컴파일 없이 실행
go build              # 실행 파일 빌드
go build -o myapp     # 이름 지정해서 빌드
go test ./...         # 모든 테스트 실행
go fmt ./...          # 코드 포맷팅
go vet ./...          # 정적 분석
go mod tidy           # go.mod/go.sum 정리
go get package@v1.2.3 # 패키지 추가
```

---

## 추천 학습 자료

### 공식 자료
- **A Tour of Go**: https://go.dev/tour/ — 대화형 튜토리얼, 무조건 완주
- **Effective Go**: https://go.dev/doc/effective_go — Go다운 코드 작성법
- **Go 표준 라이브러리**: https://pkg.go.dev/std

### 책
- **"The Go Programming Language"** (Donovan & Kernighan) — 바이블. C 배경자에게 특히 좋음
- **"Learning Go"** (Jon Bodner) — Python/Java 배경자에게 친절한 현대적 입문서
- **"100 Go Mistakes"** (Teiva Harsanyi) — 중급 이상 필독서

### 온라인
- **Go by Example**: https://gobyexample.com — 예제 중심 레퍼런스
- **Go Playground**: https://play.golang.org — 브라우저에서 바로 실행
- **Gopher Slack**: https://invite.slack.golangbridge.org — 커뮤니티

---

## 이 커리큘럼의 사용법

각 예제 디렉토리에는:
- `THEORY.md` — 이 파일처럼 코드 실행 전에 읽는 이론 가이드
- `main.go` — 실행 가능한 예제 코드

**권장 학습 순서:**
1. `THEORY.md` 읽기 (개념 이해)
2. `main.go` 코드 읽기 (적용 확인)
3. `go run main.go` 실행 (출력 확인)
4. 코드 수정하며 실험

> Go 학습의 핵심은 읽는 것이 아니라 **직접 타이핑하고 컴파일러와 싸우는 것**입니다.
