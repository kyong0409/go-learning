# 과제 A3: 메트릭 레지스트리 시스템

**난이도**: ★★★★☆
**예상 소요 시간**: 4~6시간
**참고 패턴**: Prometheus Go Client (prometheus/client_golang)

## 배경

Prometheus는 Go 생태계에서 표준 메트릭 시스템입니다.
이 과제에서는 Prometheus와 호환되는 메트릭 레지스트리를 직접 구현합니다.

## 요구사항

### 구현할 컴포넌트

#### 1. Collector 인터페이스

```go
type Collector interface {
    Describe(ch chan<- *Desc)
    Collect(ch chan<- Metric)
}
```

#### 2. Registry

```go
type Registry struct {
    // Register(Collector) error
    // Unregister(Collector) bool
    // Gather() ([]*MetricFamily, error)  // 모든 메트릭 수집
}
```

#### 3. 메트릭 타입

**CounterVec**: 레이블이 있는 Counter (증가만 가능)
```go
cv := NewCounterVec("http_requests_total", "HTTP 요청 수", []string{"method", "code"})
cv.WithLabelValues("GET", "200").Inc()
cv.WithLabelValues("GET", "200").Add(5)
```

**GaugeVec**: 레이블이 있는 Gauge (증감 가능)
```go
gv := NewGaugeVec("active_connections", "활성 연결 수", []string{"service"})
gv.WithLabelValues("api").Set(42)
gv.WithLabelValues("api").Inc()
gv.WithLabelValues("api").Dec()
```

**HistogramVec**: 레이블이 있는 Histogram (버킷 분포)
```go
hv := NewHistogramVec("request_duration_seconds", "요청 시간",
    []string{"handler"}, []float64{0.01, 0.1, 0.5, 1.0, 5.0})
hv.WithLabelValues("/api").Observe(0.042)
```

#### 4. 텍스트 노출 형식

`Registry.Gather()` 결과를 Prometheus 텍스트 형식으로 직렬화합니다:

```
# HELP http_requests_total HTTP 요청 수
# TYPE http_requests_total counter
http_requests_total{code="200",method="GET"} 6
http_requests_total{code="500",method="POST"} 1
# HELP active_connections 활성 연결 수
# TYPE active_connections gauge
active_connections{service="api"} 42
# HELP request_duration_seconds 요청 시간
# TYPE request_duration_seconds histogram
request_duration_seconds_bucket{handler="/api",le="0.01"} 0
request_duration_seconds_bucket{handler="/api",le="0.1"} 1
request_duration_seconds_bucket{handler="/api",le="+Inf"} 1
request_duration_seconds_sum{handler="/api"} 0.042
request_duration_seconds_count{handler="/api"} 1
```

**형식 규칙**:
- 레이블은 알파벳 순으로 정렬
- `# HELP`와 `# TYPE` 줄은 메트릭 이름당 한 번
- Histogram은 `_bucket`, `_sum`, `_count` 접미어 사용
- `+Inf` 버킷은 항상 포함

### MetricFamily 구조

```go
type MetricFamily struct {
    Name    string
    Help    string
    Type    string // "counter", "gauge", "histogram"
    Metrics []MetricPoint
}

type MetricPoint struct {
    Labels map[string]string
    Value  float64
    // Histogram 전용:
    Buckets map[float64]uint64 // le → count
    Sum     float64
    Count   uint64
}
```

## 채점 기준 (100점)

| 항목 | 점수 |
|------|------|
| CounterVec (Inc/Add, 레이블) | 20점 |
| GaugeVec (Set/Inc/Dec, 레이블) | 15점 |
| HistogramVec (Observe, 버킷) | 20점 |
| Registry Register/Gather | 15점 |
| 텍스트 형식 직렬화 | 20점 |
| 동시성 안전 | 10점 |

## 실행 방법

```bash
cd a3-metrics-registry
go mod tidy
go test ./... -v -race
go test -v -run TestGrade
```

## 참고 자료

- `github.com/prometheus/client_golang/prometheus/registry.go`
- `github.com/prometheus/client_golang/prometheus/counter.go`
- `github.com/prometheus/client_golang/prometheus/histogram.go`
- `../02-prometheus-patterns/README.md`
