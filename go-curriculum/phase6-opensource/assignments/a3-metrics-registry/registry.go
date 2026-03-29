// registry.go
// 메트릭 레지스트리 - Prometheus 패턴 구현
//
// TODO: 아래 타입과 함수들을 완성하세요.
// Collector 인터페이스, Registry, CounterVec, GaugeVec, HistogramVec를 구현합니다.
package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// Desc는 메트릭 설명자입니다. 이름, 도움말, 레이블 이름을 담습니다.
type Desc struct {
	Name       string
	Help       string
	LabelNames []string
	MetricType string // "counter", "gauge", "histogram"
}

// Metric은 단일 메트릭 값(레이블 집합 + 값)을 표현합니다.
type Metric struct {
	Labels  map[string]string
	Value   float64
	// Histogram 전용
	Buckets map[float64]uint64 // le → cumulative count
	Sum     float64
	Count   uint64
	IsHistogram bool
}

// MetricFamily는 같은 이름을 가진 메트릭 집합입니다.
// Gather() 결과로 반환됩니다.
type MetricFamily struct {
	Name    string
	Help    string
	Type    string // "counter", "gauge", "histogram"
	Metrics []Metric
}

// Collector는 메트릭을 설명하고 수집하는 인터페이스입니다.
// Prometheus의 Collector 인터페이스와 동일합니다.
type Collector interface {
	Describe(ch chan<- *Desc)
	Collect(ch chan<- Metric)
}

// ============================================================
// Registry 구현
// ============================================================

// Registry는 Collector를 등록하고 수집하는 중앙 레지스트리입니다.
type Registry struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - collectors: []Collector (등록된 컬렉터 목록)
	//   - mu: sync.RWMutex
}

// NewRegistry는 새 Registry를 생성합니다.
func NewRegistry() *Registry {
	// TODO: 구현하세요
	panic("NewRegistry: 아직 구현되지 않았습니다")
}

// Register는 Collector를 레지스트리에 등록합니다.
// 같은 메트릭 이름이 이미 등록되어 있으면 에러를 반환합니다.
func (r *Registry) Register(c Collector) error {
	// TODO: 구현하세요
	panic("Register: 아직 구현되지 않았습니다")
}

// Unregister는 Collector를 레지스트리에서 제거합니다.
// 성공하면 true, 없으면 false를 반환합니다.
func (r *Registry) Unregister(c Collector) bool {
	// TODO: 구현하세요
	panic("Unregister: 아직 구현되지 않았습니다")
}

// Gather는 모든 등록된 Collector에서 메트릭을 수집해 MetricFamily 목록으로 반환합니다.
// 같은 이름의 메트릭은 하나의 MetricFamily로 묶입니다.
func (r *Registry) Gather() ([]*MetricFamily, error) {
	// TODO: 구현하세요
	// 힌트:
	//   1. 각 Collector.Collect(ch) 호출
	//   2. 메트릭 이름별로 MetricFamily 생성
	//   3. Desc에서 Help, Type 가져오기
	//   4. 이름 순으로 정렬해 반환
	panic("Gather: 아직 구현되지 않았습니다")
}

// ============================================================
// CounterVec 구현
// ============================================================

// counter는 단일 레이블 집합에 대한 Counter입니다.
type counter struct {
	mu    sync.Mutex
	value float64
}

func (c *counter) Inc() {
	// TODO: 구현하세요
	panic("Inc: 아직 구현되지 않았습니다")
}

func (c *counter) Add(delta float64) {
	// TODO: 구현하세요
	// 힌트: delta < 0이면 panic (Counter는 감소 불가)
	panic("Add: 아직 구현되지 않았습니다")
}

func (c *counter) get() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// CounterVec는 레이블이 있는 Counter 모음입니다.
// Prometheus의 CounterVec과 동일한 역할입니다.
type CounterVec struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - desc: *Desc
	//   - mu: sync.RWMutex
	//   - counters: map[string]*counter (레이블 키 → counter)
}

// NewCounterVec는 새 CounterVec를 생성합니다.
func NewCounterVec(name, help string, labelNames []string) *CounterVec {
	// TODO: 구현하세요
	panic("NewCounterVec: 아직 구현되지 않았습니다")
}

// WithLabelValues는 레이블 값에 해당하는 counter를 반환합니다.
// 존재하지 않으면 새로 생성합니다.
// labelValues는 NewCounterVec의 labelNames 순서와 일치해야 합니다.
func (cv *CounterVec) WithLabelValues(labelValues ...string) *counter {
	// TODO: 구현하세요
	// 힌트: labelNames와 labelValues를 zip하여 map 생성 후 키 문자열로 변환
	panic("WithLabelValues: 아직 구현되지 않았습니다")
}

// Describe는 이 컬렉터의 Desc를 채널로 전송합니다.
func (cv *CounterVec) Describe(ch chan<- *Desc) {
	// TODO: 구현하세요
	panic("CounterVec.Describe: 아직 구현되지 않았습니다")
}

// Collect는 현재 메트릭 값을 채널로 전송합니다.
func (cv *CounterVec) Collect(ch chan<- Metric) {
	// TODO: 구현하세요
	panic("CounterVec.Collect: 아직 구현되지 않았습니다")
}

// ============================================================
// GaugeVec 구현
// ============================================================

// gauge는 단일 레이블 집합에 대한 Gauge입니다.
type gauge struct {
	mu    sync.Mutex
	value float64
}

func (g *gauge) Set(v float64) {
	// TODO: 구현하세요
	panic("Set: 아직 구현되지 않았습니다")
}

func (g *gauge) Inc() {
	// TODO: 구현하세요
	panic("gauge.Inc: 아직 구현되지 않았습니다")
}

func (g *gauge) Dec() {
	// TODO: 구현하세요
	panic("Dec: 아직 구현되지 않았습니다")
}

func (g *gauge) Add(delta float64) {
	// TODO: 구현하세요
	panic("gauge.Add: 아직 구현되지 않았습니다")
}

func (g *gauge) get() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// GaugeVec는 레이블이 있는 Gauge 모음입니다.
type GaugeVec struct {
	// TODO: 필요한 필드를 추가하세요.
}

func NewGaugeVec(name, help string, labelNames []string) *GaugeVec {
	// TODO: 구현하세요
	panic("NewGaugeVec: 아직 구현되지 않았습니다")
}

func (gv *GaugeVec) WithLabelValues(labelValues ...string) *gauge {
	// TODO: 구현하세요
	panic("GaugeVec.WithLabelValues: 아직 구현되지 않았습니다")
}

func (gv *GaugeVec) Describe(ch chan<- *Desc) {
	// TODO: 구현하세요
	panic("GaugeVec.Describe: 아직 구현되지 않았습니다")
}

func (gv *GaugeVec) Collect(ch chan<- Metric) {
	// TODO: 구현하세요
	panic("GaugeVec.Collect: 아직 구현되지 않았습니다")
}

// ============================================================
// HistogramVec 구현
// ============================================================

// histogramSample은 단일 레이블 집합에 대한 Histogram입니다.
type histogramSample struct {
	mu      sync.Mutex
	buckets []float64         // 상한값 목록 (오름차순)
	counts  map[float64]uint64 // le → 누적 개수
	sum     float64
	count   uint64
}

func newHistogramSample(buckets []float64) *histogramSample {
	// TODO: 구현하세요
	// 힌트: buckets를 정렬하고, counts 맵 초기화
	panic("newHistogramSample: 아직 구현되지 않았습니다")
}

// Observe는 관찰값을 히스토그램에 기록합니다.
func (h *histogramSample) Observe(v float64) {
	// TODO: 구현하세요
	// 힌트: v <= le인 모든 버킷의 counts를 증가, sum과 count도 갱신
	panic("Observe: 아직 구현되지 않았습니다")
}

// HistogramVec는 레이블이 있는 Histogram 모음입니다.
type HistogramVec struct {
	// TODO: 필요한 필드를 추가하세요.
}

func NewHistogramVec(name, help string, labelNames []string, buckets []float64) *HistogramVec {
	// TODO: 구현하세요
	panic("NewHistogramVec: 아직 구현되지 않았습니다")
}

func (hv *HistogramVec) WithLabelValues(labelValues ...string) *histogramSample {
	// TODO: 구현하세요
	panic("HistogramVec.WithLabelValues: 아직 구현되지 않았습니다")
}

func (hv *HistogramVec) Describe(ch chan<- *Desc) {
	// TODO: 구현하세요
	panic("HistogramVec.Describe: 아직 구현되지 않았습니다")
}

func (hv *HistogramVec) Collect(ch chan<- Metric) {
	// TODO: 구현하세요
	panic("HistogramVec.Collect: 아직 구현되지 않았습니다")
}

// ============================================================
// 텍스트 형식 직렬화
// ============================================================

// WriteTextFormat은 MetricFamily 목록을 Prometheus 텍스트 형식으로 변환합니다.
//
// 형식:
//
//	# HELP <name> <help>
//	# TYPE <name> <type>
//	<name>{label1="v1",...} <value>
//
// Histogram의 경우:
//
//	<name>_bucket{...,le="0.1"} <count>
//	<name>_bucket{...,le="+Inf"} <total_count>
//	<name>_sum{...} <sum>
//	<name>_count{...} <count>
func WriteTextFormat(families []*MetricFamily) string {
	// TODO: 구현하세요
	// 힌트:
	//   - strings.Builder 사용
	//   - 레이블은 알파벳 순으로 정렬
	//   - Histogram 버킷도 le 순으로 정렬
	//   - fmt.Sprintf로 float64 포맷: "%g"
	panic("WriteTextFormat: 아직 구현되지 않았습니다")
}

// ============================================================
// 헬퍼 함수
// ============================================================

// labelsToKey는 레이블 맵을 정렬된 문자열 키로 변환합니다.
// 예: {"method":"GET","code":"200"} → "code=200,method=GET"
func labelsToKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}
	return strings.Join(parts, ",")
}

// labelsToPrometheusFormat은 레이블 맵을 Prometheus 형식으로 변환합니다.
// 예: {"method":"GET","code":"200"} → `code="200",method="GET"`
func labelsToPrometheusFormat(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", k, labels[k]))
	}
	return strings.Join(parts, ",")
}

func main() {
	// 테스트를 실행하세요: go test ./... -v -race
}
