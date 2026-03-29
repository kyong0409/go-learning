# 02. Prometheus Go 클라이언트 패턴

## 개요

Prometheus Go 클라이언트(`prometheus/client_golang`)는 메트릭 수집과 노출의 모범 사례를 보여줍니다.
레지스트리 패턴, Collector 인터페이스, 레이블 기반 메트릭이 핵심입니다.

---

## 1. Registry 패턴

### 개념

Registry는 모든 Collector를 중앙에서 관리하고, `Gather()` 호출 시 모든 메트릭을 수집합니다.

```
Collector 등록 (Register)
        ↓
HTTP /metrics 요청
        ↓
Registry.Gather()
        ↓
각 Collector.Collect(ch) 호출
        ↓
MetricFamily 직렬화 (Prometheus 텍스트 형식)
        ↓
HTTP 응답
```

### 실제 코드 위치

```
github.com/prometheus/client_golang/prometheus/registry.go
  - type Registerer interface        // Register, Unregister, MustRegister
  - type Gatherer interface          // Gather() ([]*dto.MetricFamily, error)
  - type Registry struct             // 기본 구현체
  - var DefaultRegisterer Registerer // 전역 기본 레지스트리
  - var DefaultGatherer Gatherer

github.com/prometheus/client_golang/prometheus/wrap.go
  - func WrapRegistererWith(...) Registerer  // 레이블 주입
```

### Registry 인터페이스

```go
// 실제 Prometheus 코드
type Registerer interface {
    Register(Collector) error
    MustRegister(...Collector)
    Unregister(Collector) bool
}

type Gatherer interface {
    Gather() ([]*dto.MetricFamily, error)
}
```

---

## 2. Collector 인터페이스

### 개념

모든 메트릭 타입은 `Collector` 인터페이스를 구현합니다.
`Describe`는 메트릭 메타데이터를, `Collect`는 실제 값을 채널로 전송합니다.

### 실제 코드 위치

```
github.com/prometheus/client_golang/prometheus/collector.go
  - type Collector interface
    - Describe(chan<- *Desc)
    - Collect(chan<- Metric)

github.com/prometheus/client_golang/prometheus/desc.go
  - type Desc struct            // 메트릭 설명자 (이름, 도움말, 레이블)
  - func NewDesc(...) *Desc
```

### Collector 구현 예시

```go
// 실제 Prometheus 패턴
type myCollector struct {
    desc *prometheus.Desc
}

func (c *myCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.desc
}

func (c *myCollector) Collect(ch chan<- prometheus.Metric) {
    ch <- prometheus.MustNewConstMetric(
        c.desc,
        prometheus.GaugeValue,
        getValue(), // 실제 값 수집
        "label_value",
    )
}
```

---

## 3. 메트릭 타입

### Counter

증가만 하는 누적 값. 요청 수, 에러 수 등에 사용.

```
github.com/prometheus/client_golang/prometheus/counter.go
  - type Counter interface
  - type CounterVec struct    // 레이블이 있는 Counter 모음
  - func NewCounterVec(opts CounterOpts, labelNames []string) *CounterVec
```

```go
// 레이블이 있는 Counter
httpRequests := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "HTTP 요청 총 수",
    },
    []string{"method", "code"},
)
// 사용
httpRequests.WithLabelValues("GET", "200").Inc()
httpRequests.WithLabelValues("POST", "500").Add(3)
```

### Gauge

증감 가능한 현재 값. 현재 연결 수, 메모리 사용량 등.

```
github.com/prometheus/client_golang/prometheus/gauge.go
  - type Gauge interface
  - type GaugeVec struct
```

```go
activeConns := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{Name: "active_connections"},
    []string{"service"},
)
activeConns.WithLabelValues("api").Set(42)
activeConns.WithLabelValues("api").Inc()
activeConns.WithLabelValues("api").Dec()
```

### Histogram

값의 분포를 버킷으로 측정. 응답 시간 분포, 요청 크기 등.

```
github.com/prometheus/client_golang/prometheus/histogram.go
  - type Histogram interface
  - type HistogramVec struct
  - var DefBuckets []float64   // 기본 버킷: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
```

```go
requestDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1.0, 5.0},
    },
    []string{"handler"},
)
requestDuration.WithLabelValues("/api/users").Observe(0.042)
```

---

## 4. 텍스트 노출 형식

Prometheus는 사람이 읽을 수 있는 텍스트 형식을 사용합니다.

### 형식 규칙

```
# HELP metric_name 도움말 설명
# TYPE metric_name counter|gauge|histogram|summary
metric_name{label1="value1",label2="value2"} 숫자값 [타임스탬프ms]
```

### 예시

```
# HELP http_requests_total HTTP 요청 총 수
# TYPE http_requests_total counter
http_requests_total{code="200",method="GET"} 1234
http_requests_total{code="500",method="POST"} 7

# HELP http_request_duration_seconds 요청 처리 시간
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{handler="/api",le="0.1"} 892
http_request_duration_seconds_bucket{handler="/api",le="0.5"} 1203
http_request_duration_seconds_bucket{handler="/api",le="+Inf"} 1250
http_request_duration_seconds_sum{handler="/api"} 127.3
http_request_duration_seconds_count{handler="/api"} 1250
```

### 실제 코드 위치

```
github.com/prometheus/common/expfmt/text_create.go
  - func MetricFamilyToText(out io.Writer, in *dto.MetricFamily) (int, error)

github.com/prometheus/client_golang/prometheus/promhttp/http.go
  - func Handler() http.Handler
  - func HandlerFor(reg prometheus.Gatherer, opts HandlerOpts) http.Handler
```

---

## 5. HTTP 노출 패턴

```go
// 실제 사용 패턴
func main() {
    // 커스텀 레지스트리
    reg := prometheus.NewRegistry()
    reg.MustRegister(myCollector)

    // HTTP 핸들러
    http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
    http.ListenAndServe(":8080", nil)
}
```

---

## 학습 포인트 요약

| 패턴 | 핵심 아이디어 | 적용 과제 |
|------|---------------|-----------|
| Registry | 중앙 Collector 관리, Gather | A3 |
| Collector 인터페이스 | Describe + Collect 분리 | A3 |
| Vec 타입 | 레이블로 메트릭 다차원화 | A3 |
| 버킷 Histogram | 분포 측정, le 레이블 | A3 |
| 텍스트 형식 | Prometheus 호환 출력 | A3 |

## 다음 단계

[A3 - 메트릭 레지스트리](../assignments/a3-metrics-registry/) 과제를 진행하세요.
이 과제는 Phase 6에서 가장 독립적이며 시작하기 좋습니다.
