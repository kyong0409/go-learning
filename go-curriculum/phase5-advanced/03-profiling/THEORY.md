# Go 성능 최적화 - pprof, 벤치마크, 이스케이프 분석

## Go의 성능 철학

> "측정 먼저, 최적화 나중" - Rob Pike

최적화의 황금 규칙:
```
1. 먼저 올바르게 동작하게 만들어라
2. 그 다음 테스트가 통과하는지 확인하라
3. 성능이 실제 문제인지 측정으로 확인하라
4. 병목 구간만 최적화하라 (전체 코드의 3%가 90%의 시간을 차지)
5. 최적화 후 다시 측정하여 효과를 검증하라
```

---

## pprof 프로파일링

### 1. net/http/pprof: HTTP 엔드포인트 (라이브 프로파일)

```go
import (
    "net/http"
    _ "net/http/pprof" // 사이드 이펙트: /debug/pprof/ 경로 등록
)

func main() {
    // 기존 HTTP 서버에 추가하거나 별도 포트로 실행
    go func() {
        log.Println("pprof 서버 시작: http://localhost:6060/debug/pprof/")
        log.Fatal(http.ListenAndServe("localhost:6060", nil))
    }()

    // ... 메인 애플리케이션 로직
}
```

```bash
# 실행 중인 서버에서 30초 CPU 프로파일 수집
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 메모리 프로파일 수집
go tool pprof http://localhost:6060/debug/pprof/heap

# 고루틴 프로파일 (현재 고루틴 스택 트레이스)
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 뮤텍스 프로파일 (잠금 경합)
go tool pprof http://localhost:6060/debug/pprof/mutex
```

### 2. runtime/pprof: 프로그래밍 방식 수집

```go
import "runtime/pprof"

func main() {
    // CPU 프로파일 파일로 저장
    cpuFile, _ := os.Create("cpu.prof")
    defer cpuFile.Close()
    pprof.StartCPUProfile(cpuFile)
    defer pprof.StopCPUProfile()

    // ... 프로파일할 코드 실행 ...

    // 메모리 프로파일
    memFile, _ := os.Create("mem.prof")
    defer memFile.Close()
    runtime.GC() // GC 후 현재 상태 측정
    pprof.WriteHeapProfile(memFile)
}
```

### go tool pprof 사용법

```bash
# 프로파일 파일 분석 (인터랙티브 모드)
go tool pprof cpu.prof

# 주요 명령어 (pprof 인터랙티브 쉘 내부)
(pprof) top10          # CPU 시간 기준 상위 10개 함수
(pprof) top20 -cum     # 누적 시간 기준 (하위 호출 포함)
(pprof) list myFunc    # 특정 함수의 소스 코드 수준 분석
(pprof) web            # 브라우저에서 SVG 호출 그래프 열기
(pprof) svg > graph.svg # SVG 파일로 저장

# 플레임 그래프 (go tool pprof -http 모드)
go tool pprof -http=:8080 cpu.prof
# → 브라우저에서 http://localhost:8080 접속
# → "Flame Graph" 뷰 선택
```

### 플레임 그래프 읽는 법

```
넓이 = CPU 시간 (넓을수록 더 많은 시간 소비)
높이 = 호출 스택 깊이 (위가 호출된 함수)
색상 = 의미 없음 (구분을 위한 임의 색상)

    ┌──────────────────────────────┐
    │     json.Marshal (30%)       │  ← 가장 넓은 부분이 병목
    ├────────────┬─────────────────┤
    │ reflect    │ bytes.Buffer    │
    ├────┬───────┤                 │
    │fmt │encode │                 │
    └────┴───────┴─────────────────┘
    ↑ 바닥이 main() 또는 goroutine entry
```

### 프로파일 종류별 언제 사용하는가

```
CPU 프로파일    → "왜 이렇게 느려?" - 어디서 연산하는지
메모리 프로파일 → "왜 메모리를 많이 써?" - 어디서 할당하는지
고루틴 프로파일 → "왜 멈춰있어?" - 어디서 블로킹하는지
뮤텍스 프로파일 → "왜 잠금이 많아?" - 어디서 경합하는지
블록 프로파일   → "왜 고루틴이 대기해?" - 채널/뮤텍스 대기
```

---

## 벤치마크

### 기본 벤치마크 작성

```go
// bookmark_test.go
package bookmark

import "testing"

// 함수 이름은 반드시 BenchmarkXxx 형식
func BenchmarkMarshal(b *testing.B) {
    bm := Bookmark{
        ID:    1,
        URL:   "https://golang.org",
        Title: "Go 공식 사이트",
        Tags:  []string{"go", "programming"},
    }

    // b.N: Go 런타임이 신뢰할 수 있는 결과를 위해 자동 조정
    for i := 0; i < b.N; i++ {
        _, err := json.Marshal(bm)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### testing.B.Loop() — Go 1.24 추가

Go 1.24에서 `b.Loop()`가 추가되었습니다. `for i := 0; i < b.N; i++` 패턴을 대체하는 더 간결하고 안전한 방식입니다. 루프 변수 선언 실수(`i` 누락 등)를 방지하고, 향후 런타임이 벤치마크 측정 방식을 개선할 수 있는 확장 포인트가 됩니다.

```go
// Go 1.24+ — b.Loop() 사용 (권장)
func BenchmarkMarshalLoop(b *testing.B) {
    bm := Bookmark{
        ID:    1,
        URL:   "https://golang.org",
        Title: "Go 공식 사이트",
        Tags:  []string{"go", "programming"},
    }

    for b.Loop() {  // b.N 루프를 대체
        _, err := json.Marshal(bm)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// b.ResetTimer()와 함께 사용 — 셋업 시간 제외
func BenchmarkWithSetupLoop(b *testing.B) {
    data := generateLargeData()
    b.ResetTimer()
    for b.Loop() {
        process(data)
    }
}
```

```go

// b.ResetTimer(): 셋업 시간 제외
func BenchmarkWithSetup(b *testing.B) {
    // 셋업 코드 (시간 측정에서 제외됨)
    data := generateLargeData()

    b.ResetTimer() // 여기부터 측정 시작
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

// b.ReportAllocs(): 할당 횟수와 바이트 수 출력
func BenchmarkAllocs(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        _ = fmt.Sprintf("bookmark-%d", i) // 힙 할당 발생
    }
}

// 서브벤치마크: 여러 케이스 비교
func BenchmarkSerialize(b *testing.B) {
    bm := createTestBookmark()
    b.Run("JSON", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            json.Marshal(bm)
        }
    })
    b.Run("Protobuf", func(b *testing.B) {
        pb := toProto(bm)
        for i := 0; i < b.N; i++ {
            proto.Marshal(pb)
        }
    })
}
```

### 벤치마크 실행

```bash
# 벤치마크 실행 (테스트는 건너뜀)
go test -bench=. -benchmem ./...

# 특정 벤치마크만
go test -bench=BenchmarkMarshal -benchmem

# 더 오래 실행 (더 안정적인 결과)
go test -bench=. -benchtime=5s

# 결과 저장 후 비교
go test -bench=. -benchmem > before.txt
# ... 코드 변경 ...
go test -bench=. -benchmem > after.txt
benchstat before.txt after.txt

# 벤치마크 결과 해석
# BenchmarkMarshal-8   1234567   987 ns/op   256 B/op   3 allocs/op
# 이름-GOMAXPROCS      반복횟수  1회당 나노초  1회당 바이트  1회당 할당
```

### benchstat으로 통계 비교

```bash
go install golang.org/x/perf/cmd/benchstat@latest

# 여러 번 실행하여 통계적 신뢰도 확보
go test -bench=. -count=10 > before.txt
# 코드 최적화 후
go test -bench=. -count=10 > after.txt

benchstat before.txt after.txt
# 출력 예시:
# name         old time/op  new time/op  delta
# Marshal-8    987ns ± 2%   423ns ± 1%   -57.15%  (p=0.000 n=10+10)
# p < 0.05 이면 통계적으로 유의미한 차이
```

---

## 이스케이프 분석

Go 컴파일러는 변수를 스택 또는 힙에 할당합니다. 힙 할당은 GC 부담을 늘립니다.

### 이스케이프 분석 확인

```bash
# -m 플래그: 이스케이프 분석 결과 출력
go build -gcflags="-m" ./...

# 더 자세한 출력
go build -gcflags="-m=2" ./...
```

### 이스케이프 발생 예시

```go
// 1. 포인터를 반환하면 힙으로 이스케이프
func newBookmark() *Bookmark {
    b := Bookmark{} // ← 힙에 할당됨 (함수 종료 후에도 사용)
    return &b
}

// 2. 인터페이스에 저장하면 힙으로 이스케이프
func store(v interface{}) {
    // v는 힙에 저장됨
}

func example() {
    b := Bookmark{} // ← 힙으로 이스케이프
    store(b)
}

// 3. 클로저가 외부 변수를 캡처하면 이스케이프
func makeAdder(x int) func(int) int {
    // x는 힙으로 이스케이프 (클로저가 참조)
    return func(y int) int { return x + y }
}

// 4. 슬라이스 용량을 알 수 없으면 힙 할당
func makeSlice(n int) []int {
    return make([]int, n) // n이 컴파일 타임에 알 수 없으면 힙
}
```

### 이스케이프 줄이기

```go
// 나쁜 패턴: 불필요한 힙 할당
func processBookmarks(urls []string) []*Bookmark {
    result := make([]*Bookmark, len(urls))
    for i, url := range urls {
        b := &Bookmark{URL: url} // 매번 힙 할당
        result[i] = b
    }
    return result
}

// 좋은 패턴: 슬라이스 값으로 저장
func processBookmarks(urls []string) []Bookmark {
    result := make([]Bookmark, len(urls)) // 한 번의 연속 할당
    for i, url := range urls {
        result[i] = Bookmark{URL: url} // 스택 또는 슬라이스 내부
    }
    return result
}
```

---

## sync.Pool: 객체 재사용

```go
var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func marshalBookmark(b Bookmark) ([]byte, error) {
    buf := bufPool.Get().(*bytes.Buffer) // 풀에서 가져옴
    buf.Reset()
    defer bufPool.Put(buf) // 사용 후 풀에 반환

    if err := json.NewEncoder(buf).Encode(b); err != nil {
        return nil, err
    }
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes())
    return result, nil
}
```

---

## runtime/trace FlightRecorder — Go 1.25

Go 1.25에서 `runtime/trace` 패키지에 `FlightRecorder`가 추가되었습니다. 프로덕션 서버에서 항상 트레이스를 수집하되, 문제가 발생했을 때만 덤프하는 "블랙박스" 패턴입니다.

```go
import "runtime/trace"

func main() {
    // FlightRecorder: 최근 N초 분량의 트레이스를 항상 메모리에 유지
    fr := &trace.FlightRecorder{}
    fr.Start()
    defer fr.Stop()

    // ... 서버 로직 실행 ...

    // 장애나 이상 감지 시 트레이스 덤프
    http.HandleFunc("/debug/trace-dump", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/octet-stream")
        if err := fr.WriteTo(w); err != nil {
            http.Error(w, err.Error(), 500)
        }
    })
}
```

**기존 방식과의 차이**:
- 기존 `go tool trace`: 시작 시점부터 명시적으로 수집 — 프로덕션에서 상시 사용 부담
- `FlightRecorder`: 링 버퍼로 최근 데이터만 유지 — 오버헤드 최소, 항상 켜두기 적합

---

## GC: Green Tea GC — Go 1.26 기본값

Go의 가비지 컬렉터가 Go 1.26에서 **Green Tea GC**로 전환되어 기본값이 되었습니다.

```
GC 발전 역사:
  Go 1.0~1.4:  Stop-the-world GC (전체 중단)
  Go 1.5:      삼색 마크-스윕 (STW 크게 감소)
  Go 1.14:     비선점 goroutine → GC 협업 개선
  Go 1.24:     Swiss Table 맵 기본값 (맵 성능 25~40% 향상)
  Go 1.25:     Green Tea GC 실험적 도입 (GOEXPERIMENT=greenteagc)
  Go 1.26:     Green Tea GC 기본값 (별도 설정 불필요)
```

**Green Tea GC 주요 개선**:
- 힙 스캔 방식 개선으로 GC 일시 중단(pause) 시간 단축
- 대규모 힙(수 GB)에서 특히 효과적
- 기존 코드 변경 없이 자동 적용

```go
// Go 1.26: 별도 설정 없이 Green Tea GC 자동 적용
// Go 1.25에서는 GOEXPERIMENT=greenteagc 환경 변수가 필요했음

// GC 튜닝 (여전히 유효)
import "runtime/debug"

// GOGC: GC 트리거 비율 (기본값 100 = 힙이 2배가 되면 GC)
debug.SetGCPercent(200)  // GC 빈도 줄이기 (메모리 더 사용)
debug.SetGCPercent(50)   // GC 빈도 늘리기 (메모리 절약)

// GOMEMLIMIT: 메모리 사용 상한 (Go 1.19+)
debug.SetMemoryLimit(512 * 1024 * 1024) // 512MB 상한
```

---

## PGO (Profile-Guided Optimization)

Go 1.21+에서 지원하는 프로파일 기반 최적화입니다.

```bash
# 1단계: 프로덕션에서 CPU 프로파일 수집
curl http://prod-server:6060/debug/pprof/profile?seconds=30 > default.pgo

# 2단계: default.pgo를 소스 루트에 배치
# go build가 자동으로 PGO 적용
go build -o myapp ./cmd/myapp

# PGO 없이 빌드 (비교용)
go build -pgo=off -o myapp-nopgo ./cmd/myapp

# 성능 향상: 일반적으로 2~14% (함수 인라이닝 개선이 주요 원인)
```

---

## 최적화 체크리스트

```go
// 1. 프리할당: 슬라이스/맵 크기 미리 지정
result := make([]Bookmark, 0, expectedSize) // 용량 힌트 제공

// 2. 문자열 빌더 사용
var sb strings.Builder
for _, tag := range tags {
    sb.WriteString(tag)
    sb.WriteString(",")
}

// 3. 반복문에서 불필요한 변환 피하기
// 나쁨: 매 반복마다 문자열 변환
for _, b := range bookmarks {
    if strings.ToLower(b.URL) == query { ... } // 매번 변환
}
// 좋음: 한 번만 변환
lowerQuery := strings.ToLower(query)
for _, b := range bookmarks {
    if strings.ToLower(b.URL) == lowerQuery { ... }
}

// 4. 큰 구조체는 포인터로 전달 (복사 비용 절약)
func process(b *Bookmark) { ... } // 값 복사 없음

// 5. 인터페이스 사용 최소화 (동적 디스패치 비용)
// 핫 패스(자주 실행되는 코드)에서는 구체 타입 사용
```
