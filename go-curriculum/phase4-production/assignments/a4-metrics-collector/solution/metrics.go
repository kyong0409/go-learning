// a4-metrics-collector/solution/metrics.go
// 메트릭 수집 시스템 참고 답안입니다.
package metrics

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
)

// ============================================================
// Metric 인터페이스
// ============================================================

type Metric interface {
	Name() string
	Help() string
	Type() string
}

// ============================================================
// Counter
// ============================================================

// Counter는 float64를 uint64 비트로 변환하여 atomic 연산으로 관리합니다.
// 단조 증가만 허용합니다.
type Counter struct {
	name  string
	help  string
	bits  atomic.Uint64 // math.Float64bits로 인코딩
}

func NewCounter(name, help string) *Counter {
	return &Counter{name: name, help: help}
}

func (c *Counter) Name() string { return c.name }
func (c *Counter) Help() string { return c.help }
func (c *Counter) Type() string { return "counter" }

func (c *Counter) Inc() {
	c.Add(1)
}

// Add는 delta가 0보다 클 때만 원자적으로 더합니다.
func (c *Counter) Add(delta float64) {
	if delta <= 0 {
		return
	}
	for {
		old := c.bits.Load()
		newVal := math.Float64frombits(old) + delta
		if c.bits.CompareAndSwap(old, math.Float64bits(newVal)) {
			return
		}
	}
}

func (c *Counter) Value() float64 {
	return math.Float64frombits(c.bits.Load())
}

// ============================================================
// Gauge
// ============================================================

type Gauge struct {
	name string
	help string
	bits atomic.Uint64
}

func NewGauge(name, help string) *Gauge {
	return &Gauge{name: name, help: help}
}

func (g *Gauge) Name() string { return g.name }
func (g *Gauge) Help() string { return g.help }
func (g *Gauge) Type() string { return "gauge" }

func (g *Gauge) Set(value float64) {
	g.bits.Store(math.Float64bits(value))
}

func (g *Gauge) Inc() { g.Add(1) }
func (g *Gauge) Dec() { g.Add(-1) }

func (g *Gauge) Add(delta float64) {
	for {
		old := g.bits.Load()
		newVal := math.Float64frombits(old) + delta
		if g.bits.CompareAndSwap(old, math.Float64bits(newVal)) {
			return
		}
	}
}

func (g *Gauge) Sub(delta float64) {
	g.Add(-delta)
}

func (g *Gauge) Value() float64 {
	return math.Float64frombits(g.bits.Load())
}

// ============================================================
// Histogram
// ============================================================

type Histogram struct {
	name    string
	help    string
	bounds  []float64 // 정렬된 버킷 상한값 (+Inf 포함)
	mu      sync.Mutex
	counts  []uint64 // bounds와 동일 길이
	sum     float64
	count   uint64
}

func NewHistogram(name, help string, buckets []float64) *Histogram {
	// 복사 후 정렬
	b := make([]float64, len(buckets))
	copy(b, buckets)
	sort.Float64s(b)

	// +Inf 추가 (마지막에 없으면)
	if len(b) == 0 || !math.IsInf(b[len(b)-1], 1) {
		b = append(b, math.Inf(1))
	}

	return &Histogram{
		name:   name,
		help:   help,
		bounds: b,
		counts: make([]uint64, len(b)),
	}
}

func (h *Histogram) Name() string { return h.name }
func (h *Histogram) Help() string { return h.help }
func (h *Histogram) Type() string { return "histogram" }

func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.count++
	h.sum += value

	// value <= bound인 모든 버킷 카운트 증가 (누적)
	for i, bound := range h.bounds {
		if value <= bound {
			h.counts[i]++
		}
	}
}

func (h *Histogram) Count() uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

func (h *Histogram) Sum() float64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sum
}

func (h *Histogram) Buckets() map[float64]uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make(map[float64]uint64, len(h.bounds))
	for i, bound := range h.bounds {
		result[bound] = h.counts[i]
	}
	return result
}

var DefaultBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0}

// ============================================================
// Registry
// ============================================================

type Registry struct {
	mu      sync.RWMutex
	metrics map[string]Metric
}

func NewRegistry() *Registry {
	return &Registry{metrics: make(map[string]Metric)}
}

func (r *Registry) Register(m Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.metrics[m.Name()]; exists {
		return fmt.Errorf("메트릭 %q이 이미 등록되어 있습니다", m.Name())
	}
	r.metrics[m.Name()] = m
	return nil
}

func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.metrics[name]; !exists {
		return fmt.Errorf("메트릭 %q을 찾을 수 없습니다", name)
	}
	delete(r.metrics, name)
	return nil
}

func (r *Registry) Get(name string) (Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.metrics[name]
	return m, ok
}

func (r *Registry) List() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.metrics))
	for name := range r.metrics {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]Metric, len(names))
	for i, name := range names {
		result[i] = r.metrics[name]
	}
	return result
}

// ============================================================
// MetricsHandler
// ============================================================

func MetricsHandler(reg *Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		for _, m := range reg.List() {
			fmt.Fprintf(w, "# HELP %s %s\n", m.Name(), m.Help())
			fmt.Fprintf(w, "# TYPE %s %s\n", m.Name(), m.Type())

			switch v := m.(type) {
			case *Counter:
				fmt.Fprintf(w, "%s %g\n", v.Name(), v.Value())

			case *Gauge:
				fmt.Fprintf(w, "%s %g\n", v.Name(), v.Value())

			case *Histogram:
				name := v.Name()
				bkts := v.Buckets()

				// 버킷 상한값을 정렬하여 출력
				bounds := make([]float64, 0, len(bkts))
				for b := range bkts {
					bounds = append(bounds, b)
				}
				sort.Float64s(bounds)

				for _, bound := range bounds {
					cnt := bkts[bound]
					if math.IsInf(bound, 1) {
						fmt.Fprintf(w, "%s_bucket{le=\"+Inf\"} %d\n", name, cnt)
					} else {
						fmt.Fprintf(w, "%s_bucket{le=\"%g\"} %d\n", name, bound, cnt)
					}
				}
				fmt.Fprintf(w, "%s_sum %g\n", name, v.Sum())
				fmt.Fprintf(w, "%s_count %d\n", name, v.Count())
			}
			fmt.Fprintln(w)
		}
	})
}
