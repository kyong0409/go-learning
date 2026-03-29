// solution/registry.go
// 메트릭 레지스트리 참고 풀이
//
// 이 파일은 참고용입니다. 먼저 직접 구현해보세요.
package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

// ============================================================
// 타입 정의
// ============================================================

type Desc struct {
	Name       string
	Help       string
	LabelNames []string
	MetricType string
}

type Metric struct {
	Labels      map[string]string
	Value       float64
	Buckets     map[float64]uint64
	Sum         float64
	Count       uint64
	IsHistogram bool
}

type MetricFamily struct {
	Name    string
	Help    string
	Type    string
	Metrics []Metric
}

type Collector interface {
	Describe(ch chan<- *Desc)
	Collect(ch chan<- Metric)
}

// ============================================================
// Registry
// ============================================================

type Registry struct {
	mu         sync.RWMutex
	collectors []Collector
	names      map[string]struct{}
}

func NewRegistry() *Registry {
	return &Registry{names: make(map[string]struct{})}
}

func (r *Registry) Register(c Collector) error {
	descCh := make(chan *Desc, 10)
	go func() {
		c.Describe(descCh)
		close(descCh)
	}()

	var descs []*Desc
	for d := range descCh {
		descs = append(descs, d)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, d := range descs {
		if _, exists := r.names[d.Name]; exists {
			return fmt.Errorf("메트릭 %q가 이미 등록되어 있습니다", d.Name)
		}
	}
	for _, d := range descs {
		r.names[d.Name] = struct{}{}
	}
	r.collectors = append(r.collectors, c)
	return nil
}

func (r *Registry) Unregister(c Collector) bool {
	descCh := make(chan *Desc, 10)
	go func() {
		c.Describe(descCh)
		close(descCh)
	}()
	var names []string
	for d := range descCh {
		names = append(names, d.Name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existing := range r.collectors {
		if existing == c {
			r.collectors = append(r.collectors[:i], r.collectors[i+1:]...)
			for _, name := range names {
				delete(r.names, name)
			}
			return true
		}
	}
	return false
}

func (r *Registry) Gather() ([]*MetricFamily, error) {
	r.mu.RLock()
	collectors := make([]Collector, len(r.collectors))
	copy(collectors, r.collectors)
	r.mu.RUnlock()

	familyMap := make(map[string]*MetricFamily)

	for _, c := range collectors {
		descCh := make(chan *Desc, 10)
		go func(col Collector) {
			col.Describe(descCh)
			close(descCh)
		}(c)
		for d := range descCh {
			if _, exists := familyMap[d.Name]; !exists {
				familyMap[d.Name] = &MetricFamily{
					Name: d.Name,
					Help: d.Help,
					Type: d.MetricType,
				}
			}
		}

		metricCh := make(chan Metric, 100)
		go func(col Collector) {
			col.Collect(metricCh)
			close(metricCh)
		}(c)
		for m := range metricCh {
			// 메트릭 이름 확인
			for name, fam := range familyMap {
				_ = name
				fam.Metrics = append(fam.Metrics, m)
				break
			}
		}
	}

	// 더 정확한 구현: Collector별로 이름 추적
	familyMap2 := make(map[string]*MetricFamily)
	for _, c := range collectors {
		descCh := make(chan *Desc, 10)
		go func(col Collector) {
			col.Describe(descCh)
			close(descCh)
		}(c)
		var collectorDescs []*Desc
		for d := range descCh {
			collectorDescs = append(collectorDescs, d)
			if _, exists := familyMap2[d.Name]; !exists {
				familyMap2[d.Name] = &MetricFamily{
					Name: d.Name,
					Help: d.Help,
					Type: d.MetricType,
				}
			}
		}

		metricCh := make(chan Metric, 100)
		go func(col Collector) {
			col.Collect(metricCh)
			close(metricCh)
		}(c)
		for m := range metricCh {
			if len(collectorDescs) > 0 {
				name := collectorDescs[0].Name
				familyMap2[name].Metrics = append(familyMap2[name].Metrics, m)
			}
		}
	}

	result := make([]*MetricFamily, 0, len(familyMap2))
	for _, fam := range familyMap2 {
		result = append(result, fam)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// ============================================================
// CounterVec
// ============================================================

type counter struct {
	mu    sync.Mutex
	value float64
}

func (c *counter) Inc() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func (c *counter) Add(delta float64) {
	if delta < 0 {
		panic(fmt.Sprintf("Counter에 음수 값을 추가할 수 없습니다: %g", delta))
	}
	c.mu.Lock()
	c.value += delta
	c.mu.Unlock()
}

func (c *counter) get() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

type CounterVec struct {
	desc       *Desc
	mu         sync.RWMutex
	counters   map[string]*counter
	labelNames []string
}

func NewCounterVec(name, help string, labelNames []string) *CounterVec {
	return &CounterVec{
		desc: &Desc{
			Name:       name,
			Help:       help,
			LabelNames: labelNames,
			MetricType: "counter",
		},
		counters:   make(map[string]*counter),
		labelNames: labelNames,
	}
}

func (cv *CounterVec) WithLabelValues(labelValues ...string) *counter {
	labels := make(map[string]string, len(cv.labelNames))
	for i, name := range cv.labelNames {
		if i < len(labelValues) {
			labels[name] = labelValues[i]
		}
	}
	key := labelsToKey(labels)

	cv.mu.RLock()
	if c, ok := cv.counters[key]; ok {
		cv.mu.RUnlock()
		return c
	}
	cv.mu.RUnlock()

	cv.mu.Lock()
	defer cv.mu.Unlock()
	if c, ok := cv.counters[key]; ok {
		return c
	}
	c := &counter{}
	cv.counters[key] = c
	return c
}

func (cv *CounterVec) Describe(ch chan<- *Desc) {
	ch <- cv.desc
}

func (cv *CounterVec) Collect(ch chan<- Metric) {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	for key, c := range cv.counters {
		labels := keyToLabels(key, cv.labelNames)
		ch <- Metric{Labels: labels, Value: c.get()}
	}
}

// ============================================================
// GaugeVec
// ============================================================

type gauge struct {
	mu    sync.Mutex
	value float64
}

func (g *gauge) Set(v float64) {
	g.mu.Lock()
	g.value = v
	g.mu.Unlock()
}

func (g *gauge) Inc() {
	g.mu.Lock()
	g.value++
	g.mu.Unlock()
}

func (g *gauge) Dec() {
	g.mu.Lock()
	g.value--
	g.mu.Unlock()
}

func (g *gauge) Add(delta float64) {
	g.mu.Lock()
	g.value += delta
	g.mu.Unlock()
}

func (g *gauge) get() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

type GaugeVec struct {
	desc       *Desc
	mu         sync.RWMutex
	gauges     map[string]*gauge
	labelNames []string
}

func NewGaugeVec(name, help string, labelNames []string) *GaugeVec {
	return &GaugeVec{
		desc: &Desc{
			Name:       name,
			Help:       help,
			LabelNames: labelNames,
			MetricType: "gauge",
		},
		gauges:     make(map[string]*gauge),
		labelNames: labelNames,
	}
}

func (gv *GaugeVec) WithLabelValues(labelValues ...string) *gauge {
	labels := make(map[string]string, len(gv.labelNames))
	for i, name := range gv.labelNames {
		if i < len(labelValues) {
			labels[name] = labelValues[i]
		}
	}
	key := labelsToKey(labels)

	gv.mu.RLock()
	if g, ok := gv.gauges[key]; ok {
		gv.mu.RUnlock()
		return g
	}
	gv.mu.RUnlock()

	gv.mu.Lock()
	defer gv.mu.Unlock()
	if g, ok := gv.gauges[key]; ok {
		return g
	}
	g := &gauge{}
	gv.gauges[key] = g
	return g
}

func (gv *GaugeVec) Describe(ch chan<- *Desc) { ch <- gv.desc }

func (gv *GaugeVec) Collect(ch chan<- Metric) {
	gv.mu.RLock()
	defer gv.mu.RUnlock()
	for key, g := range gv.gauges {
		labels := keyToLabels(key, gv.labelNames)
		ch <- Metric{Labels: labels, Value: g.get()}
	}
}

// ============================================================
// HistogramVec
// ============================================================

type histogramSample struct {
	mu      sync.Mutex
	buckets []float64
	counts  map[float64]uint64
	sum     float64
	count   uint64
}

func newHistogramSample(buckets []float64) *histogramSample {
	sorted := make([]float64, len(buckets))
	copy(sorted, buckets)
	sort.Float64s(sorted)

	counts := make(map[float64]uint64, len(sorted))
	for _, b := range sorted {
		counts[b] = 0
	}
	return &histogramSample{buckets: sorted, counts: counts}
}

func (h *histogramSample) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, b := range h.buckets {
		if v <= b {
			h.counts[b]++
		}
	}
	h.sum += v
	h.count++
}

type HistogramVec struct {
	desc       *Desc
	mu         sync.RWMutex
	histograms map[string]*histogramSample
	labelNames []string
	buckets    []float64
}

func NewHistogramVec(name, help string, labelNames []string, buckets []float64) *HistogramVec {
	return &HistogramVec{
		desc: &Desc{
			Name:       name,
			Help:       help,
			LabelNames: labelNames,
			MetricType: "histogram",
		},
		histograms: make(map[string]*histogramSample),
		labelNames: labelNames,
		buckets:    buckets,
	}
}

func (hv *HistogramVec) WithLabelValues(labelValues ...string) *histogramSample {
	labels := make(map[string]string, len(hv.labelNames))
	for i, name := range hv.labelNames {
		if i < len(labelValues) {
			labels[name] = labelValues[i]
		}
	}
	key := labelsToKey(labels)

	hv.mu.RLock()
	if h, ok := hv.histograms[key]; ok {
		hv.mu.RUnlock()
		return h
	}
	hv.mu.RUnlock()

	hv.mu.Lock()
	defer hv.mu.Unlock()
	if h, ok := hv.histograms[key]; ok {
		return h
	}
	h := newHistogramSample(hv.buckets)
	hv.histograms[key] = h
	return h
}

func (hv *HistogramVec) Describe(ch chan<- *Desc) { ch <- hv.desc }

func (hv *HistogramVec) Collect(ch chan<- Metric) {
	hv.mu.RLock()
	defer hv.mu.RUnlock()
	for key, h := range hv.histograms {
		labels := keyToLabels(key, hv.labelNames)
		h.mu.Lock()
		bucketsCopy := make(map[float64]uint64, len(h.counts))
		for k, v := range h.counts {
			bucketsCopy[k] = v
		}
		infCount := h.count
		sum := h.sum
		count := h.count
		h.mu.Unlock()
		// +Inf 버킷
		bucketsCopy[math.Inf(1)] = infCount
		ch <- Metric{
			Labels:      labels,
			IsHistogram: true,
			Buckets:     bucketsCopy,
			Sum:         sum,
			Count:       count,
		}
	}
}

// ============================================================
// 텍스트 형식 직렬화
// ============================================================

func WriteTextFormat(families []*MetricFamily) string {
	var sb strings.Builder
	for _, fam := range families {
		fmt.Fprintf(&sb, "# HELP %s %s\n", fam.Name, fam.Help)
		fmt.Fprintf(&sb, "# TYPE %s %s\n", fam.Name, fam.Type)

		for _, m := range fam.Metrics {
			if m.IsHistogram {
				labelsStr := labelsToPrometheusFormat(m.Labels)
				// 버킷을 le 순으로 정렬
				les := make([]float64, 0, len(m.Buckets))
				for le := range m.Buckets {
					les = append(les, le)
				}
				sort.Float64s(les)
				for _, le := range les {
					leStr := fmt.Sprintf("%g", le)
					if math.IsInf(le, 1) {
						leStr = "+Inf"
					}
					var bucketLabels string
					if labelsStr != "" {
						bucketLabels = labelsStr + fmt.Sprintf(",le=%q", leStr)
					} else {
						bucketLabels = fmt.Sprintf("le=%q", leStr)
					}
					fmt.Fprintf(&sb, "%s_bucket{%s} %d\n",
						fam.Name, bucketLabels, m.Buckets[le])
				}
				// sum, count
				if labelsStr != "" {
					fmt.Fprintf(&sb, "%s_sum{%s} %g\n", fam.Name, labelsStr, m.Sum)
					fmt.Fprintf(&sb, "%s_count{%s} %d\n", fam.Name, labelsStr, m.Count)
				} else {
					fmt.Fprintf(&sb, "%s_sum %g\n", fam.Name, m.Sum)
					fmt.Fprintf(&sb, "%s_count %d\n", fam.Name, m.Count)
				}
			} else {
				labelsStr := labelsToPrometheusFormat(m.Labels)
				if labelsStr != "" {
					fmt.Fprintf(&sb, "%s{%s} %g\n", fam.Name, labelsStr, m.Value)
				} else {
					fmt.Fprintf(&sb, "%s %g\n", fam.Name, m.Value)
				}
			}
		}
	}
	return sb.String()
}

// ============================================================
// 헬퍼
// ============================================================

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

func keyToLabels(key string, labelNames []string) map[string]string {
	labels := make(map[string]string)
	if key == "" {
		return labels
	}
	parts := strings.Split(key, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return labels
}

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

func main() {}
