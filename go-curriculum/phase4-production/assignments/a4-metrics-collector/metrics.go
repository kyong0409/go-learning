// a4-metrics-collector/metrics.go
// Prometheus에서 영감을 받은 메트릭 수집 시스템 과제입니다.
// TODO 주석이 있는 모든 함수/메서드를 구현하세요.
package metrics

import (
	"math"
	"net/http"
)

// ============================================================
// Metric 인터페이스 (수정하지 마세요)
// ============================================================

// Metric은 모든 메트릭 타입이 구현해야 하는 인터페이스입니다.
type Metric interface {
	Name() string
	Help() string
	Type() string // "counter", "gauge", "histogram"
}

// ============================================================
// Counter
// ============================================================

// Counter는 단조 증가하는 메트릭입니다.
// 고루틴 안전해야 합니다.
type Counter struct {
	name string
	help string
	// TODO: 값을 저장할 필드를 추가하세요. (atomic 권장)
}

// NewCounter는 새 Counter를 생성합니다.
func NewCounter(name, help string) *Counter {
	// TODO: 구현하세요.
	return &Counter{name: name, help: help}
}

func (c *Counter) Name() string { return c.name }
func (c *Counter) Help() string { return c.help }
func (c *Counter) Type() string { return "counter" }

// Inc는 카운터를 1 증가시킵니다.
// TODO: 구현하세요.
func (c *Counter) Inc() {
}

// Add는 카운터를 delta만큼 증가시킵니다.
// delta가 0 이하이면 무시합니다.
// TODO: 구현하세요.
func (c *Counter) Add(delta float64) {
}

// Value는 현재 카운터 값을 반환합니다.
// TODO: 구현하세요.
func (c *Counter) Value() float64 {
	return 0
}

// ============================================================
// Gauge
// ============================================================

// Gauge는 임의로 증감할 수 있는 메트릭입니다.
// 고루틴 안전해야 합니다.
type Gauge struct {
	name string
	help string
	// TODO: 값을 저장할 필드를 추가하세요.
}

// NewGauge는 새 Gauge를 생성합니다.
func NewGauge(name, help string) *Gauge {
	// TODO: 구현하세요.
	return &Gauge{name: name, help: help}
}

func (g *Gauge) Name() string { return g.name }
func (g *Gauge) Help() string { return g.help }
func (g *Gauge) Type() string { return "gauge" }

// Set은 게이지 값을 설정합니다.
// TODO: 구현하세요.
func (g *Gauge) Set(value float64) {
}

// Inc는 게이지를 1 증가시킵니다.
// TODO: 구현하세요.
func (g *Gauge) Inc() {
}

// Dec는 게이지를 1 감소시킵니다.
// TODO: 구현하세요.
func (g *Gauge) Dec() {
}

// Add는 게이지를 delta만큼 더합니다.
// TODO: 구현하세요.
func (g *Gauge) Add(delta float64) {
}

// Sub는 게이지를 delta만큼 뺍니다.
// TODO: 구현하세요.
func (g *Gauge) Sub(delta float64) {
}

// Value는 현재 게이지 값을 반환합니다.
// TODO: 구현하세요.
func (g *Gauge) Value() float64 {
	return 0
}

// ============================================================
// Histogram
// ============================================================

// Histogram은 관찰 값을 버킷으로 분류하는 메트릭입니다.
// 고루틴 안전해야 합니다.
type Histogram struct {
	name    string
	help    string
	buckets []float64 // 상한값 목록 (정렬됨, +Inf 포함)
	// TODO: 버킷별 카운트, sum, count 필드를 추가하세요.
}

// NewHistogram은 새 Histogram을 생성합니다.
// buckets는 오름차순으로 정렬되며, +Inf는 자동으로 추가됩니다.
// TODO: 구현하세요.
func NewHistogram(name, help string, buckets []float64) *Histogram {
	return &Histogram{name: name, help: help, buckets: buckets}
}

func (h *Histogram) Name() string { return h.name }
func (h *Histogram) Help() string { return h.help }
func (h *Histogram) Type() string { return "histogram" }

// Observe는 값을 관찰하고 해당 버킷들을 업데이트합니다.
// value <= bucket 상한값인 모든 버킷의 카운트를 증가시킵니다.
// TODO: 구현하세요.
func (h *Histogram) Observe(value float64) {
}

// Count는 총 관찰 횟수를 반환합니다.
// TODO: 구현하세요.
func (h *Histogram) Count() uint64 {
	return 0
}

// Sum은 모든 관찰 값의 합계를 반환합니다.
// TODO: 구현하세요.
func (h *Histogram) Sum() float64 {
	return 0
}

// Buckets는 버킷 상한값 -> 누적 카운트 맵을 반환합니다.
// +Inf 버킷도 포함됩니다.
// TODO: 구현하세요.
func (h *Histogram) Buckets() map[float64]uint64 {
	return nil
}

// DefaultBuckets는 기본 히스토그램 버킷입니다 (초 단위 응답 시간에 적합).
var DefaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}

// infBucket은 +Inf 버킷 상한값입니다.
var infBucket = math.Inf(1)

// ============================================================
// Registry
// ============================================================

// Registry는 메트릭을 이름으로 관리합니다.
// 고루틴 안전해야 합니다.
type Registry struct {
	// TODO: 메트릭 저장 필드를 추가하세요.
}

// NewRegistry는 새 Registry를 생성합니다.
// TODO: 구현하세요.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register는 메트릭을 레지스트리에 등록합니다.
// 같은 이름이 이미 등록된 경우 error를 반환합니다.
// TODO: 구현하세요.
func (r *Registry) Register(m Metric) error {
	return nil
}

// Unregister는 이름으로 메트릭을 제거합니다.
// 존재하지 않으면 error를 반환합니다.
// TODO: 구현하세요.
func (r *Registry) Unregister(name string) error {
	return nil
}

// Get은 이름으로 메트릭을 조회합니다.
// TODO: 구현하세요.
func (r *Registry) Get(name string) (Metric, bool) {
	return nil, false
}

// List는 등록된 모든 메트릭을 이름 오름차순으로 반환합니다.
// TODO: 구현하세요.
func (r *Registry) List() []Metric {
	return nil
}

// ============================================================
// MetricsHandler
// ============================================================

// MetricsHandler는 Registry의 메트릭을 텍스트 형식으로 노출하는
// http.Handler를 반환합니다.
//
// 출력 형식 예시:
//
//	# HELP http_requests_total 총 HTTP 요청 수
//	# TYPE http_requests_total counter
//	http_requests_total 42
//
//	# HELP active_connections 현재 활성 연결 수
//	# TYPE active_connections gauge
//	active_connections 7
//
//	# HELP http_duration_seconds HTTP 응답 시간
//	# TYPE http_duration_seconds histogram
//	http_duration_seconds_bucket{le="0.005"} 0
//	http_duration_seconds_bucket{le="+Inf"} 2
//	http_duration_seconds_sum 0.192
//	http_duration_seconds_count 2
//
// TODO: 구현하세요.
func MetricsHandler(reg *Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: 레지스트리의 메트릭을 텍스트 형식으로 응답하세요.
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	})
}
