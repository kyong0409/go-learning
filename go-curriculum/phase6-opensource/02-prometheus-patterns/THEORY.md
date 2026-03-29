# 02. Prometheus Go 클라이언트 패턴 - 이론 심화

## Prometheus 아키텍처: Pull 모델

Prometheus는 대부분의 모니터링 시스템과 달리 **Pull 모델**을 사용합니다.

```
Push 모델 (일반적):
  애플리케이션 → 모니터링 서버에 메트릭 전송

Pull 모델 (Prometheus):
  Prometheus 서버 → 애플리케이션 /metrics 엔드포인트 scrape
```

Pull 모델의 장점:
- 애플리케이션이 죽으면 자동으로 감지됨 (scrape 실패)
- 메트릭 서버 주소 변경 시 애플리케이션 재배포 불필요
- 수집 간격을 Prometheus에서 중앙 제어

```
┌─────────────────────────────────────────────────────────────┐
│                    Prometheus 아키텍처                        │
│                                                              │
│  ┌──────────┐     HTTP GET /metrics      ┌────────────────┐ │
│  │   앱 A   │ ←────────────────────────── │                │ │
│  │ :8080    │ ──── text/plain ──────────→ │  Prometheus    │ │
│  └──────────┘                             │   Server       │ │
│                                           │                │ │
│  ┌──────────┐     HTTP GET /metrics      │  - TSDB        │ │
│  │   앱 B   │ ←────────────────────────── │  - PromQL      │ │
│  │ :9090    │ ──── text/plain ──────────→ │  - Alerting    │ │
│  └──────────┘                             └────────┬───────┘ │
│                                                    │         │
│  ┌──────────┐     HTTP GET /metrics               │         │
│  │  Node    │ ←───────────────────────────────────┘         │
│  │ Exporter │ ──── text/plain ──────────→                   │
│  └──────────┘                                               │
│                                                              │
│              ┌─────────────┐                                │
│              │   Grafana   │ ← PromQL 쿼리                   │
│              └─────────────┘                                │
└─────────────────────────────────────────────────────────────┘
```

---

## Prometheus Go 클라이언트 코드 구조

```
github.com/prometheus/client_golang/
├── prometheus/
│   ├── collector.go      ← Collector 인터페이스 정의
│   ├── registry.go       ← Registry 구현 (핵심)
│   ├── desc.go           ← Desc: 메트릭 설명자
│   ├── counter.go        ← Counter 타입
│   ├── gauge.go          ← Gauge 타입
│   ├── histogram.go      ← Histogram 타입
│   ├── summary.go        ← Summary 타입
│   ├── labels.go         ← 레이블 관련 유틸리티
│   ├── wrap.go           ← WrapRegisterer (레이블 주입)
│   └── promhttp/
│       └── http.go       ← HTTP 핸들러 (Handler, HandlerFor)
└── examples/             ← 실제 사용 예시
```

---

## Collector 인터페이스 심층 분석

### 인터페이스 정의

```go
// github.com/prometheus/client_golang/prometheus/collector.go
type Collector interface {
    // Describe: 이 Collector가 생성할 메트릭의 설명자(Desc)를 채널로 전송
    // Registry.Register() 시점에 한 번 호출됨 (정적)
    // 목적: 중복 등록 감지, 타입 검증
    Describe(chan<- *Desc)

    // Collect: 현재 메트릭 값을 채널로 전송
    // /metrics HTTP 요청 시마다 호출됨 (동적)
    // 목적: 실시간 값 수집
    Collect(chan<- Metric)
}
```

### Describe와 Collect의 역할 분리

```
등록 시:
  Registry.Register(myCollector)
       ↓
  myCollector.Describe(ch) 호출
       ↓
  Registry가 Desc 수집 → 중복/충돌 검사
       ↓
  통과하면 등록 완료

수집 시:
  HTTP GET /metrics
       ↓
  Registry.Gather()
       ↓
  myCollector.Collect(ch) 호출
       ↓
  실제 값 채널로 전송
       ↓
  텍스트 형식으로 직렬화 → HTTP 응답
```

### Desc: 메트릭 설명자

```go
// 메트릭의 정적 메타데이터 (이름, 도움말, 레이블 이름)
type Desc struct {
    fqName     string    // 전체 이름: "namespace_subsystem_name"
    help       string    // # HELP에 표시될 설명
    constLabelPairs []*dto.LabelPair  // 고정 레이블 (빌드 시 결정)
    variableLabels  []string          // 변수 레이블 이름 (수집 시 값 결정)
    id         uint64    // 해시 (중복 감지용)
}

// Desc 생성
desc := prometheus.NewDesc(
    "http_requests_total",           // 이름
    "HTTP 요청 총 수",                // 도움말
    []string{"method", "status"},    // 변수 레이블 이름 (Collect 시 값 지정)
    prometheus.Labels{"app": "api"}, // 고정 레이블 (항상 이 값)
)
```

### 완전한 커스텀 Collector 구현

```go
// 데이터베이스 연결 풀 메트릭을 수집하는 Collector
type DBPoolCollector struct {
    db          *sql.DB
    activeDesc  *prometheus.Desc
    idleDesc    *prometheus.Desc
    waitDesc    *prometheus.Desc
}

func NewDBPoolCollector(db *sql.DB, dbName string) *DBPoolCollector {
    labels := prometheus.Labels{"db": dbName}
    return &DBPoolCollector{
        db: db,
        activeDesc: prometheus.NewDesc(
            "db_pool_active_connections",
            "현재 사용 중인 DB 연결 수",
            nil, labels,
        ),
        idleDesc: prometheus.NewDesc(
            "db_pool_idle_connections",
            "유휴 DB 연결 수",
            nil, labels,
        ),
        waitDesc: prometheus.NewDesc(
            "db_pool_wait_count_total",
            "연결 대기 총 횟수",
            nil, labels,
        ),
    }
}

func (c *DBPoolCollector) Describe(ch chan<- *prometheus.Desc) {
    // 이 Collector가 생성할 모든 Desc를 채널로 전송
    ch <- c.activeDesc
    ch <- c.idleDesc
    ch <- c.waitDesc
}

func (c *DBPoolCollector) Collect(ch chan<- prometheus.Metric) {
    stats := c.db.Stats()  // 실제 DB 상태 조회

    ch <- prometheus.MustNewConstMetric(
        c.activeDesc,
        prometheus.GaugeValue,
        float64(stats.InUse),
    )
    ch <- prometheus.MustNewConstMetric(
        c.idleDesc,
        prometheus.GaugeValue,
        float64(stats.Idle),
    )
    ch <- prometheus.MustNewConstMetric(
        c.waitDesc,
        prometheus.CounterValue,
        float64(stats.WaitCount),
    )
}
```

---

## 메트릭 타입 상세

### Counter: 단조 증가

Counter는 절대 감소하지 않습니다. 프로세스 재시작 시 0으로 리셋됩니다.

```go
// 내부 구조 (단순화)
type counter struct {
    // 원자적 업데이트를 위해 64비트 정렬 보장
    valInt   uint64  // Inc()에 최적화된 정수 부분
    valFloat float64 // Add(float)에 사용
    selfCollector
    desc *Desc
    labelPairs []*dto.LabelPair
}

func (c *counter) Inc() {
    atomic.AddUint64(&c.valInt, 1)  // lock-free 원자적 증가
}

func (c *counter) Add(v float64) {
    if v < 0 {
        panic("counter cannot decrease in value")  // 감소 시 패닉
    }
    // float 부분에 추가 (CAS 루프)
}
```

사용 시 고려사항:
```go
// PromQL에서 Counter 활용
// rate(): 초당 증가율 계산
// rate(http_requests_total[5m])  → 최근 5분 평균 RPS

// increase(): 특정 기간 총 증가량
// increase(http_requests_total[1h])  → 최근 1시간 총 요청 수
```

### Gauge: 증감 가능한 현재 값

```go
// Gauge는 현재 상태를 나타냄
activeConns := prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "active_connections",
    Help: "현재 활성 연결 수",
})

// Set: 절대값 설정 (외부에서 측정한 값)
activeConns.Set(float64(pool.ActiveCount()))

// Inc/Dec: 상대적 변경 (자신이 제어하는 값)
activeConns.Inc()    // 연결 추가 시
activeConns.Dec()    // 연결 해제 시

// PromQL에서는 현재 값 그대로 사용
// active_connections → 현재 연결 수 그래프
```

### Histogram: 분포 관측 (버킷 기반)

Histogram은 값의 분포를 미리 정의된 버킷으로 측정합니다.

```go
// 버킷 동작 원리
// Buckets: [0.01, 0.05, 0.1, 0.5, 1.0, +Inf]
// Observe(0.042) 호출 시:
//   le=0.01: 해당 안됨
//   le=0.05: 해당 (0.042 < 0.05) → 이 버킷 +1
//   le=0.1:  해당 → +1
//   le=0.5:  해당 → +1
//   le=1.0:  해당 → +1
//   le=+Inf: 항상 해당 → +1 (총 count)
//
// 각 버킷은 누적 카운트 (작거나 같은 값의 개수)

// 텍스트 출력:
// http_request_duration_seconds_bucket{le="0.01"} 0
// http_request_duration_seconds_bucket{le="0.05"} 234
// http_request_duration_seconds_bucket{le="0.1"}  892
// http_request_duration_seconds_bucket{le="0.5"}  1203
// http_request_duration_seconds_bucket{le="+Inf"} 1250
// http_request_duration_seconds_sum 127.3
// http_request_duration_seconds_count 1250

// PromQL로 백분위수 근사:
// histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
// → 최근 5분간 요청의 99번째 백분위수 응답 시간 (근사값)
```

버킷 선택 가이드:
```go
// 너무 적은 버킷: 분포 정보 손실
// 너무 많은 버킷: 메모리/CPU 낭비

// 기본 버킷 (DefBuckets): 일반 응답 시간용
// .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10 (초)

// 커스텀 버킷: 도메인에 맞게 조정
prometheus.ExponentialBuckets(0.001, 2, 10)
// → 0.001, 0.002, 0.004, 0.008, ... (2배씩, 10개)

prometheus.LinearBuckets(0, 10, 11)
// → 0, 10, 20, 30, ..., 100 (10씩, 11개)
```

### Summary: 클라이언트 사이드 분위수

```go
// Summary는 클라이언트에서 직접 분위수를 계산
// Histogram과 달리 서버에서 근사 계산 불필요 (정확한 분위수)
// 단점: 집계 불가 (여러 인스턴스의 분위수를 합산할 수 없음)

summary := prometheus.NewSummary(prometheus.SummaryOpts{
    Name: "rpc_duration_seconds",
    Objectives: map[float64]float64{
        0.5:  0.05,   // 50번째 백분위수, 오차 5%
        0.9:  0.01,   // 90번째 백분위수, 오차 1%
        0.99: 0.001,  // 99번째 백분위수, 오차 0.1%
    },
    MaxAge: 10 * time.Minute,  // 최근 10분 데이터만 유지
})
```

### Counter vs Gauge vs Histogram vs Summary 선택 가이드

```
질문 1: 값이 감소할 수 있는가?
  아니오 → Counter
  예     → Gauge 또는 Histogram/Summary

질문 2: 값의 분포(분위수)가 필요한가?
  아니오 → Gauge
  예     → Histogram 또는 Summary

질문 3: 여러 인스턴스의 데이터를 집계해야 하는가?
  예     → Histogram (집계 가능)
  아니오 → Summary (더 정확한 분위수)

일반 원칙:
  요청 수, 에러 수   → Counter
  메모리, 연결 수    → Gauge
  응답 시간, 크기    → Histogram (거의 항상 이것)
  서버 수가 적고 정확도 중요 → Summary
```

---

## 라벨(Labels)의 철학과 카디널리티

### 라벨이 메트릭을 다차원으로 만드는 방법

```go
// 라벨 없는 단일 카운터
requestsTotal := prometheus.NewCounter(...)
requestsTotal.Inc()  // 모든 요청을 하나로 셈

// 라벨이 있는 Vec 타입
requestsTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "http_requests_total"},
    []string{"method", "path", "status"},
)

// 각 레이블 조합이 독립적인 Counter
requestsTotal.WithLabelValues("GET",  "/api/users", "200").Inc()
requestsTotal.WithLabelValues("POST", "/api/users", "201").Inc()
requestsTotal.WithLabelValues("GET",  "/api/users", "404").Inc()

// PromQL로 특정 차원만 조회
// http_requests_total{method="GET"}          → GET 요청만
// http_requests_total{status=~"5.."}         → 5xx 에러만
// sum by (method) (http_requests_total)      → 메서드별 합계
```

### 카디널리티 폭발 문제

```go
// 위험한 패턴: 고유값이 많은 레이블
requestsTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "http_requests_total"},
    []string{"user_id"},  // ← 사용자 ID: 수백만 개의 고유값 가능
)
// 결과: 수백만 개의 Counter 생성 → 메모리 폭발, Prometheus 크래시

// 안전한 패턴: 제한된 카디널리티
requestsTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "http_requests_total"},
    []string{"method", "status_class"},  // 4개 × 5개 = 최대 20개 조합
)

// 상태 코드를 클래스로 버킷화
func statusClass(code int) string {
    return fmt.Sprintf("%dxx", code/100)  // "2xx", "4xx", "5xx"
}
```

라벨 카디널리티 원칙:
- 레이블 값의 가능한 고유 조합 수 = 카디널리티
- 카디널리티가 10,000을 넘기 시작하면 주의
- 절대 사용 금지: user_id, request_id, IP 주소, 타임스탬프

---

## Registry 패턴 심층 분석

### 전역 레지스트리 vs 커스텀 레지스트리

```go
// 전역 레지스트리 (기본 동작)
// prometheus.DefaultRegisterer / prometheus.DefaultGatherer
prometheus.MustRegister(myCollector)  // 전역에 등록

// 커스텀 레지스트리 (테스트, 격리에 유용)
reg := prometheus.NewRegistry()
reg.MustRegister(myCollector)

// 핸들러도 커스텀 레지스트리 사용
http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
```

### MustRegister와 Register의 차이

```go
// Register: 에러 반환
if err := prometheus.Register(myCollector); err != nil {
    // AlreadyRegisteredError: 이미 등록된 Collector 재사용 가능
    if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
        myCollector = are.ExistingCollector.(*MyCollector)
    } else {
        log.Fatal(err)
    }
}

// MustRegister: 실패 시 패닉 (init() 함수에서 주로 사용)
var (
    requestsTotal = prometheus.NewCounterVec(...)
)
func init() {
    prometheus.MustRegister(requestsTotal)  // 프로그램 시작 시 등록
}
```

### Gather: 모든 메트릭 수집 과정

```go
// registry.go 내부 동작 (단순화)
func (r *Registry) Gather() ([]*dto.MetricFamily, error) {
    var errs MultiError
    metricChan := make(chan Metric, capMetricChan)

    // 모든 Collector에 goroutine으로 Collect 호출
    go func() {
        for _, collector := range r.collectorsByID {
            collector.Collect(metricChan)
        }
        close(metricChan)
    }()

    // 채널에서 Metric 수집 → MetricFamily로 그룹화
    families := map[string]*dto.MetricFamily{}
    for metric := range metricChan {
        desc := metric.Desc()
        family := families[desc.fqName]
        if family == nil {
            family = &dto.MetricFamily{Name: desc.fqName, Help: desc.help}
            families[desc.fqName] = family
        }
        // metric을 dto.Metric으로 변환 후 family에 추가
    }

    // 정렬 후 반환 (일관된 출력을 위해)
    return sortFamilies(families), errs.MaybeUnwrap()
}
```

---

## 텍스트 형식 (Exposition Format)

Prometheus가 사용하는 텍스트 형식은 사람이 읽을 수 있고 파싱이 쉽습니다.

```
# HELP http_requests_total HTTP 요청 총 수
# TYPE http_requests_total counter
http_requests_total{method="GET",status="200"} 1234
http_requests_total{method="GET",status="404"} 56
http_requests_total{method="POST",status="201"} 789

# HELP process_resident_memory_bytes 프로세스 RSS 메모리
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 1.048576e+08

# HELP http_request_duration_seconds 요청 처리 시간
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{handler="/api",le="0.005"} 0
http_request_duration_seconds_bucket{handler="/api",le="0.01"}  234
http_request_duration_seconds_bucket{handler="/api",le="0.025"} 567
http_request_duration_seconds_bucket{handler="/api",le="+Inf"}  892
http_request_duration_seconds_sum{handler="/api"} 8.453
http_request_duration_seconds_count{handler="/api"} 892
```

형식 규칙:
- `# HELP 이름 설명` - 메트릭 설명 (선택)
- `# TYPE 이름 타입` - 타입 선언 (선택, 권장)
- `이름{레이블} 값 [타임스탬프ms]` - 실제 값
- Histogram: `_bucket`, `_sum`, `_count` 접미사 자동 추가
- Summary: `{quantile="0.99"}`, `_sum`, `_count`

---

## 실전 계측 전략: RED 메서드

RED 메서드는 서비스의 건강 상태를 측정하는 세 가지 핵심 메트릭입니다.

```
R - Rate:     초당 요청 수 (처리량)
E - Errors:   초당 에러 수 또는 에러 비율
D - Duration: 요청 처리 시간 (지연)
```

```go
// RED 메서드 완전 구현
type REDMetrics struct {
    requestsTotal  *prometheus.CounterVec  // R + E (status로 구분)
    requestDuration *prometheus.HistogramVec // D
}

func NewREDMetrics(service string) *REDMetrics {
    return &REDMetrics{
        requestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name:        "requests_total",
                Help:        "총 요청 수",
                ConstLabels: prometheus.Labels{"service": service},
            },
            []string{"method", "status_class"},
        ),
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:        "request_duration_seconds",
                Help:        "요청 처리 시간",
                Buckets:     prometheus.DefBuckets,
                ConstLabels: prometheus.Labels{"service": service},
            },
            []string{"method"},
        ),
    }
}

// 미들웨어로 자동 계측
func (m *REDMetrics) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start).Seconds()
        statusClass := fmt.Sprintf("%dxx", wrapped.statusCode/100)

        // Rate와 Errors를 함께 기록 (status_class로 구분)
        m.requestsTotal.WithLabelValues(r.Method, statusClass).Inc()
        // Duration 기록
        m.requestDuration.WithLabelValues(r.Method).Observe(duration)
    })
}

// PromQL 대시보드 쿼리:
// Rate: rate(requests_total[5m])
// Error Rate: rate(requests_total{status_class="5xx"}[5m]) / rate(requests_total[5m])
// P99 Latency: histogram_quantile(0.99, rate(request_duration_seconds_bucket[5m]))
```

---

## 고급 패턴: WrapRegisterer

서브시스템에 공통 레이블을 자동으로 붙이는 패턴:

```go
// 모든 메트릭에 {version="1.2.3", environment="prod"} 자동 추가
wrapped := prometheus.WrapRegistererWith(
    prometheus.Labels{
        "version":     "1.2.3",
        "environment": "prod",
    },
    prometheus.DefaultRegisterer,
)

// 이 wrapped 레지스트리에 등록된 모든 메트릭에 위 레이블이 자동 추가
counter := prometheus.NewCounter(prometheus.CounterOpts{Name: "my_counter"})
wrapped.MustRegister(counter)
// 실제 메트릭: my_counter{version="1.2.3",environment="prod"} 42
```
