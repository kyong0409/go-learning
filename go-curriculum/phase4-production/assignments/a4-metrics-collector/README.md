# 과제 A4: 메트릭 수집 시스템 구현

**난이도**: ★★★★☆ (4.5/5)
**예상 소요 시간**: 4~6시간

## 과제 설명

Prometheus에서 영감을 받은 메트릭 수집 시스템을 구현합니다.
세 가지 메트릭 타입(Counter, Gauge, Histogram)과 이를 관리하는 Registry,
그리고 HTTP로 메트릭을 노출하는 핸들러를 구현하세요.

## 구현할 타입

### 1. `Counter` — 단조 증가 카운터

값이 오직 증가만 하는 메트릭입니다 (예: 총 요청 수, 오류 수).

```go
c := NewCounter("http_requests_total", "총 HTTP 요청 수")
c.Inc()           // 1 증가
c.Add(5)          // 5 증가
c.Value()         // 현재 값 반환
c.Name()          // "http_requests_total"
c.Help()          // "총 HTTP 요청 수"
```

### 2. `Gauge` — 임의 값 게이지

값이 자유롭게 증가/감소하는 메트릭입니다 (예: 현재 연결 수, 메모리 사용량).

```go
g := NewGauge("active_connections", "현재 활성 연결 수")
g.Set(42.0)       // 값 설정
g.Inc()           // 1 증가
g.Dec()           // 1 감소
g.Add(10.5)       // 덧셈
g.Sub(3.2)        // 뺄셈
g.Value()         // 현재 값 반환
```

### 3. `Histogram` — 히스토그램

관찰 값을 미리 정의된 버킷에 분류하여 분포를 측정합니다 (예: 응답 시간).

```go
buckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}
h := NewHistogram("http_duration_seconds", "HTTP 응답 시간", buckets)
h.Observe(0.042)   // 값 관찰
h.Observe(0.150)
h.Count()          // 관찰 횟수
h.Sum()            // 관찰 값 합계
h.Buckets()        // 버킷별 누적 카운트 반환 (map[float64]uint64)
```

**버킷 규칙**: 각 버킷은 해당 상한값 이하의 모든 관찰 수를 누적합니다.
예를 들어 `0.042`를 관찰하면 `0.05`, `0.1`, `0.25`, ..., `5.0` 버킷이 모두 1 증가합니다.

### 4. `Registry` — 메트릭 레지스트리

메트릭을 이름으로 등록하고 관리합니다.

```go
reg := NewRegistry()
reg.Register(counter)     // 메트릭 등록 (이름 중복 시 error)
reg.Get("http_requests_total")  // 이름으로 조회 (없으면 nil, false)
reg.List()                // 등록된 모든 메트릭 반환 (이름 정렬)
reg.Unregister("http_requests_total")  // 등록 해제
```

### 5. `MetricsHandler` — HTTP 노출 핸들러

`/metrics` 엔드포인트에서 텍스트 형식으로 메트릭을 노출합니다.

```
# HELP http_requests_total 총 HTTP 요청 수
# TYPE http_requests_total counter
http_requests_total 42

# HELP active_connections 현재 활성 연결 수
# TYPE active_connections gauge
active_connections 7

# HELP http_duration_seconds HTTP 응답 시간
# TYPE http_duration_seconds histogram
http_duration_seconds_bucket{le="0.005"} 0
http_duration_seconds_bucket{le="0.01"} 0
http_duration_seconds_bucket{le="0.05"} 1
http_duration_seconds_bucket{le="+Inf"} 2
http_duration_seconds_sum 0.192
http_duration_seconds_count 2
```

## 구현 요구사항

- **고루틴 안전성**: 모든 메트릭 타입은 동시 접근에 안전해야 합니다.
- **`Add(-1)` 방지**: Counter의 `Add`는 음수 값을 무시하거나 패닉 없이 처리하세요.
- **버킷 정렬**: Histogram 생성 시 버킷을 오름차순으로 정렬하세요.
- **`+Inf` 버킷**: Histogram은 항상 `+Inf` 버킷을 포함합니다 (모든 관찰 포함).

## 채점 기준

| 항목 | 배점 |
|------|------|
| Counter 구현 (동시성 포함) | 20점 |
| Gauge 구현 (동시성 포함) | 20점 |
| Histogram 구현 (버킷 로직) | 25점 |
| Registry 구현 | 15점 |
| MetricsHandler HTTP 출력 형식 | 20점 |
| **합계** | **100점** |

## 실행 방법

```bash
cd assignments/a4-metrics-collector
go test ./... -v
go test ./... -v -run TestGrade
```

## 힌트

- `sync/atomic` 패키지의 `atomic.Int64`, `atomic.Uint64`를 Counter/Gauge에 활용하세요.
- Gauge의 float64 원자적 연산은 `sync.Mutex`나 `sync/atomic`의 `Uint64`+`math.Float64bits`를 사용하세요.
- Histogram의 버킷 카운트는 `sync.Mutex`로 보호된 `map[float64]uint64`로 관리하세요.
- `math.Inf(1)`로 `+Inf` 버킷을 표현할 수 있습니다.
- HTTP 출력 형식은 `fmt.Fprintf`로 직접 작성하면 됩니다.
