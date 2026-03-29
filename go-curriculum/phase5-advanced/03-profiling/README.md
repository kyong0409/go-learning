# 03-profiling: Go 성능 프로파일링

Go 내장 `pprof` 도구를 사용해 CPU, 메모리, 고루틴 병목 지점을 찾고 최적화하는 방법을 학습합니다.

## 개요

**pprof**는 Go 표준 라이브러리에 내장된 프로파일링 도구입니다.
별도 설치 없이 사용할 수 있으며 다음을 측정합니다:
- **CPU 프로파일**: 어떤 함수가 CPU 시간을 가장 많이 사용하는가
- **힙 프로파일**: 어디서 메모리를 가장 많이 할당하는가
- **고루틴 프로파일**: 고루틴이 무엇을 하고 있는가
- **블록 프로파일**: 어디서 블로킹이 일어나는가

## 프로젝트 구조

```
03-profiling/
├── main.go           # HTTP 서버 + pprof 엔드포인트
├── heavy/
│   ├── heavy.go      # 의도적으로 비효율적인 함수들
│   └── heavy_test.go # 벤치마크 테스트
└── go.mod
```

## 실행 방법

```bash
cd 03-profiling
go run main.go
```

서버 2개가 시작됩니다:
- `http://localhost:8080` - 작업 트리거 엔드포인트
- `http://localhost:6060/debug/pprof/` - pprof 프로파일 엔드포인트

## pprof 사용법

### 1. 웹 UI로 즉시 확인

브라우저에서 접속:
```
http://localhost:6060/debug/pprof/
```

### 2. CPU 프로파일 수집

```bash
# 서버 실행 상태에서 부하 발생
curl http://localhost:8080/work/all &

# 10초간 CPU 프로파일 수집
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10
```

pprof 대화형 셸에서:
```
(pprof) top10          # 상위 10개 함수
(pprof) list heavy     # heavy 패키지 함수의 소스별 시간
(pprof) web            # 브라우저에서 플레임 그래프 열기
(pprof) quit
```

### 3. 힙 메모리 프로파일 수집

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

```
(pprof) top10 -cum     # 누적 할당량 기준 상위 10개
(pprof) list MemoryIntensive
```

### 4. 고루틴 프로파일 수집

```bash
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### 5. 실행 트레이스 수집

```bash
# 5초간 트레이스 수집
curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5

# 트레이스 분석 (브라우저 열림)
go tool trace trace.out
```

## 벤치마크 테스트

```bash
cd heavy

# 모든 벤치마크 실행
go test -bench=. -benchmem

# 특정 벤치마크
go test -bench=BenchmarkCPU -benchtime=5s

# CPU 프로파일 생성
go test -bench=. -benchmem -cpuprofile=cpu.prof
go tool pprof cpu.prof

# 메모리 프로파일 생성
go test -bench=. -benchmem -memprofile=mem.prof
go tool pprof mem.prof
```

### 벤치마크 출력 읽는 법

```
BenchmarkMemoryIntensive-8        200    5432198 ns/op    2048123 B/op    1024 allocs/op
                         ^         ^          ^                ^               ^
                         코어수    반복횟수   작업당 나노초    작업당 바이트  작업당 할당 횟수
```

## pprof 플레임 그래프 읽는 법

```
┌─────────────────────────────────────────┐
│           main.handleAll                 │  ← 상단이 호출자
├────────────────────┬────────────────────┤
│  heavy.CPUIntensive│ heavy.MemoryIntens.│  ← 너비 = CPU 사용 비율
├──────────┬─────────┤                    │
│ math.Sqrt│math.Sin │                    │
└──────────┴─────────┴────────────────────┘
                                            ← 하단이 실제 실행 함수
```

- **너비가 넓을수록** CPU를 많이 사용
- **탑(top)**: 전체 시간의 몇 %를 이 함수가 직접 사용하는가 (flat)
- **cum**: 이 함수 + 호출한 모든 하위 함수의 누적 시간

## 최적화 전/후 비교

| 함수 | 비효율 버전 | 최적화 버전 | 개선 |
|------|-------------|-------------|------|
| CPU | `CPUIntensive` | `CPUIntensiveOptimized` | ~10x |
| 메모리 | `MemoryIntensive` | `MemoryIntensiveOptimized` | ~50x 할당 감소 |
| 고루틴 | `GoroutineIntensive` | `GoroutineIntensiveOptimized` | ~3x |

## 학습 포인트

1. `import _ "net/http/pprof"` 한 줄로 pprof 활성화
2. pprof는 **별도 포트**에서 실행하는 것이 권장 (보안)
3. CPU 프로파일 수집 중 부하를 발생시켜야 의미있는 결과
4. `b.ReportAllocs()`로 벤치마크에서 메모리 통계 활성화
5. **먼저 측정하고 나서 최적화** (추측이 아닌 데이터 기반)
