// a4-metrics-collector/metrics_test.go
// 메트릭 수집 시스템 테스트 및 채점
//
// 실행:
//
//	go test ./... -v
//	go test ./... -v -run TestGrade
package metrics_test

import (
	"fmt"
	"math"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	metrics "github.com/learn-go/a4-metrics-collector"
)

// ============================================================
// Counter 테스트 (20점)
// ============================================================

func TestCounter_Inc(t *testing.T) {
	c := metrics.NewCounter("test_total", "테스트 카운터")
	c.Inc()
	c.Inc()
	c.Inc()

	if got := c.Value(); got != 3 {
		t.Errorf("Inc 3회 후 Value: 기대 3, 실제 %v", got)
	}
}

func TestCounter_Add(t *testing.T) {
	c := metrics.NewCounter("test_total", "테스트 카운터")
	c.Add(10)
	c.Add(5)

	if got := c.Value(); got != 15 {
		t.Errorf("Add(10)+Add(5) 후 Value: 기대 15, 실제 %v", got)
	}
}

func TestCounter_AddNegativeIgnored(t *testing.T) {
	c := metrics.NewCounter("test_total", "테스트 카운터")
	c.Add(5)
	c.Add(-3) // 음수는 무시해야 합니다
	c.Add(0)  // 0도 무시

	if got := c.Value(); got != 5 {
		t.Errorf("음수 Add 무시: 기대 5, 실제 %v", got)
	}
}

func TestCounter_Metadata(t *testing.T) {
	c := metrics.NewCounter("http_requests_total", "총 HTTP 요청 수")
	if c.Name() != "http_requests_total" {
		t.Errorf("Name: 기대 %q, 실제 %q", "http_requests_total", c.Name())
	}
	if c.Help() != "총 HTTP 요청 수" {
		t.Errorf("Help: 기대 %q, 실제 %q", "총 HTTP 요청 수", c.Help())
	}
	if c.Type() != "counter" {
		t.Errorf("Type: 기대 %q, 실제 %q", "counter", c.Type())
	}
}

func TestCounter_Concurrent(t *testing.T) {
	c := metrics.NewCounter("concurrent_total", "동시성 테스트")
	const goroutines = 100
	const incsEach = 100

	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range incsEach {
				c.Inc()
			}
		}()
	}
	wg.Wait()

	expected := float64(goroutines * incsEach)
	if got := c.Value(); got != expected {
		t.Errorf("동시 Inc %d회: 기대 %v, 실제 %v", goroutines*incsEach, expected, got)
	}
}

// ============================================================
// Gauge 테스트 (20점)
// ============================================================

func TestGauge_Set(t *testing.T) {
	g := metrics.NewGauge("temperature", "온도")
	g.Set(36.5)
	if got := g.Value(); got != 36.5 {
		t.Errorf("Set(36.5): 기대 36.5, 실제 %v", got)
	}

	g.Set(-10.0)
	if got := g.Value(); got != -10.0 {
		t.Errorf("Set(-10): 기대 -10, 실제 %v", got)
	}
}

func TestGauge_IncDec(t *testing.T) {
	g := metrics.NewGauge("connections", "연결 수")
	g.Set(10)
	g.Inc()
	g.Inc()
	g.Dec()

	if got := g.Value(); got != 11 {
		t.Errorf("Set(10)+Inc()+Inc()+Dec(): 기대 11, 실제 %v", got)
	}
}

func TestGauge_AddSub(t *testing.T) {
	g := metrics.NewGauge("memory", "메모리")
	g.Set(100)
	g.Add(50.5)
	g.Sub(20.0)

	if got := g.Value(); math.Abs(got-130.5) > 1e-9 {
		t.Errorf("Add/Sub 후: 기대 130.5, 실제 %v", got)
	}
}

func TestGauge_Metadata(t *testing.T) {
	g := metrics.NewGauge("active_connections", "현재 활성 연결 수")
	if g.Type() != "gauge" {
		t.Errorf("Type: 기대 %q, 실제 %q", "gauge", g.Type())
	}
}

func TestGauge_Concurrent(t *testing.T) {
	g := metrics.NewGauge("concurrent_gauge", "동시성 게이지")
	const goroutines = 50

	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Inc()
			g.Dec()
			g.Add(2)
			g.Sub(1)
		}()
	}
	wg.Wait()

	// 각 고루틴: +1 -1 +2 -1 = +1, 50개 고루틴이면 최종 50
	if got := g.Value(); got != float64(goroutines) {
		t.Errorf("동시 연산 후: 기대 %v, 실제 %v", float64(goroutines), got)
	}
}

// ============================================================
// Histogram 테스트 (25점)
// ============================================================

func TestHistogram_Observe(t *testing.T) {
	buckets := []float64{0.01, 0.05, 0.1, 0.5, 1.0}
	h := metrics.NewHistogram("latency", "지연시간", buckets)

	h.Observe(0.042) // 0.05, 0.1, 0.5, 1.0 버킷에 해당
	h.Observe(0.150) // 0.5, 1.0 버킷에 해당

	if h.Count() != 2 {
		t.Errorf("Count: 기대 2, 실제 %d", h.Count())
	}
	if math.Abs(h.Sum()-0.192) > 1e-9 {
		t.Errorf("Sum: 기대 0.192, 실제 %v", h.Sum())
	}
}

func TestHistogram_BucketCounts(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0}
	h := metrics.NewHistogram("req_duration", "요청 시간", buckets)

	h.Observe(0.05)  // <= 0.1, <= 0.5, <= 1.0, <= +Inf
	h.Observe(0.3)   // <= 0.5, <= 1.0, <= +Inf (> 0.1)
	h.Observe(2.0)   // <= +Inf 만

	bkts := h.Buckets()
	if bkts[0.1] != 1 {
		t.Errorf("버킷 0.1: 기대 1, 실제 %d", bkts[0.1])
	}
	if bkts[0.5] != 2 {
		t.Errorf("버킷 0.5: 기대 2, 실제 %d", bkts[0.5])
	}
	if bkts[1.0] != 2 {
		t.Errorf("버킷 1.0: 기대 2, 실제 %d", bkts[1.0])
	}
	// +Inf 버킷은 항상 전체 카운트
	infBucket := math.Inf(1)
	if bkts[infBucket] != 3 {
		t.Errorf("+Inf 버킷: 기대 3, 실제 %d", bkts[infBucket])
	}
}

func TestHistogram_InfBucketAlwaysPresent(t *testing.T) {
	h := metrics.NewHistogram("h", "h", []float64{1.0})
	h.Observe(999.0)

	bkts := h.Buckets()
	infBucket := math.Inf(1)
	if _, ok := bkts[infBucket]; !ok {
		t.Error("+Inf 버킷이 반드시 존재해야 합니다")
	}
}

func TestHistogram_Metadata(t *testing.T) {
	h := metrics.NewHistogram("http_duration_seconds", "HTTP 응답 시간", metrics.DefaultBuckets)
	if h.Type() != "histogram" {
		t.Errorf("Type: 기대 %q, 실제 %q", "histogram", h.Type())
	}
}

func TestHistogram_Concurrent(t *testing.T) {
	h := metrics.NewHistogram("concurrent_hist", "동시성 히스토그램", []float64{1.0, 10.0})
	const goroutines = 50

	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.Observe(0.5)
			h.Observe(5.0)
		}()
	}
	wg.Wait()

	if h.Count() != uint64(goroutines*2) {
		t.Errorf("동시 Observe Count: 기대 %d, 실제 %d", goroutines*2, h.Count())
	}
}

// ============================================================
// Registry 테스트 (15점)
// ============================================================

func TestRegistry_RegisterAndGet(t *testing.T) {
	reg := metrics.NewRegistry()
	c := metrics.NewCounter("reqs_total", "총 요청")

	if err := reg.Register(c); err != nil {
		t.Fatalf("Register: 예상치 못한 오류: %v", err)
	}

	got, ok := reg.Get("reqs_total")
	if !ok {
		t.Fatal("Get: 등록된 메트릭을 찾을 수 없습니다")
	}
	if got.Name() != "reqs_total" {
		t.Errorf("Get 결과 이름: 기대 %q, 실제 %q", "reqs_total", got.Name())
	}
}

func TestRegistry_DuplicateReturnsError(t *testing.T) {
	reg := metrics.NewRegistry()
	c := metrics.NewCounter("dup", "중복 테스트")

	if err := reg.Register(c); err != nil {
		t.Fatalf("첫 번째 Register: 예상치 못한 오류: %v", err)
	}
	if err := reg.Register(c); err == nil {
		t.Error("중복 Register: 오류가 반환되어야 합니다")
	}
}

func TestRegistry_ListSorted(t *testing.T) {
	reg := metrics.NewRegistry()
	reg.Register(metrics.NewCounter("z_metric", "z"))
	reg.Register(metrics.NewGauge("a_metric", "a"))
	reg.Register(metrics.NewCounter("m_metric", "m"))

	list := reg.List()
	if len(list) != 3 {
		t.Fatalf("List 개수: 기대 3, 실제 %d", len(list))
	}
	names := []string{list[0].Name(), list[1].Name(), list[2].Name()}
	expected := []string{"a_metric", "m_metric", "z_metric"}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("List[%d]: 기대 %q, 실제 %q", i, want, names[i])
		}
	}
}

func TestRegistry_Unregister(t *testing.T) {
	reg := metrics.NewRegistry()
	c := metrics.NewCounter("temp", "임시")
	reg.Register(c)

	if err := reg.Unregister("temp"); err != nil {
		t.Fatalf("Unregister: 예상치 못한 오류: %v", err)
	}
	if _, ok := reg.Get("temp"); ok {
		t.Error("Unregister 후 Get: 찾으면 안 됩니다")
	}
	if err := reg.Unregister("temp"); err == nil {
		t.Error("존재하지 않는 메트릭 Unregister: 오류가 반환되어야 합니다")
	}
}

// ============================================================
// MetricsHandler 테스트 (20점)
// ============================================================

func TestMetricsHandler_CounterOutput(t *testing.T) {
	reg := metrics.NewRegistry()
	c := metrics.NewCounter("http_requests_total", "총 HTTP 요청 수")
	c.Add(42)
	reg.Register(c)

	handler := metrics.MetricsHandler(reg)
	req := httptest.NewRequest("GET", "/metrics", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	body := rw.Body.String()

	if !strings.Contains(body, "# HELP http_requests_total 총 HTTP 요청 수") {
		t.Errorf("HELP 줄 누락:\n%s", body)
	}
	if !strings.Contains(body, "# TYPE http_requests_total counter") {
		t.Errorf("TYPE 줄 누락:\n%s", body)
	}
	if !strings.Contains(body, "http_requests_total 42") {
		t.Errorf("값 줄 누락 (기대 42):\n%s", body)
	}
}

func TestMetricsHandler_GaugeOutput(t *testing.T) {
	reg := metrics.NewRegistry()
	g := metrics.NewGauge("active_connections", "현재 활성 연결 수")
	g.Set(7)
	reg.Register(g)

	handler := metrics.MetricsHandler(reg)
	req := httptest.NewRequest("GET", "/metrics", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	body := rw.Body.String()
	if !strings.Contains(body, "# TYPE active_connections gauge") {
		t.Errorf("TYPE gauge 줄 누락:\n%s", body)
	}
	if !strings.Contains(body, "active_connections 7") {
		t.Errorf("값 줄 누락 (기대 7):\n%s", body)
	}
}

func TestMetricsHandler_HistogramOutput(t *testing.T) {
	reg := metrics.NewRegistry()
	h := metrics.NewHistogram("http_duration_seconds", "HTTP 응답 시간", []float64{0.05, 0.1, 0.5})
	h.Observe(0.042)
	h.Observe(0.150)
	reg.Register(h)

	handler := metrics.MetricsHandler(reg)
	req := httptest.NewRequest("GET", "/metrics", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	body := rw.Body.String()
	if !strings.Contains(body, "# TYPE http_duration_seconds histogram") {
		t.Errorf("TYPE histogram 줄 누락:\n%s", body)
	}
	if !strings.Contains(body, "http_duration_seconds_bucket{le=") {
		t.Errorf("버킷 줄 누락:\n%s", body)
	}
	if !strings.Contains(body, "http_duration_seconds_sum") {
		t.Errorf("sum 줄 누락:\n%s", body)
	}
	if !strings.Contains(body, "http_duration_seconds_count 2") {
		t.Errorf("count 줄 누락:\n%s", body)
	}
	if !strings.Contains(body, `{le="+Inf"}`) {
		t.Errorf("+Inf 버킷 줄 누락:\n%s", body)
	}
}

func TestMetricsHandler_ContentType(t *testing.T) {
	reg := metrics.NewRegistry()
	handler := metrics.MetricsHandler(reg)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	ct := rw.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("Content-Type: 기대 text/plain 포함, 실제 %q", ct)
	}
}

// ============================================================
// 채점 함수
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	type result struct {
		name  string
		pts   int
		maxPt int
		pass  bool
	}
	var results []result

	check := func(name string, maxPt int, fn func() bool) {
		pass := fn()
		pts := 0
		if pass {
			pts = maxPt
			score += maxPt
		}
		results = append(results, result{name, pts, maxPt, pass})
	}

	// Counter (20점)
	check("Counter Inc/Add/Value", 8, func() bool {
		c := metrics.NewCounter("c1", "h")
		c.Inc()
		c.Add(9)
		c.Add(-1) // 무시
		return c.Value() == 10
	})
	check("Counter 메타데이터", 4, func() bool {
		c := metrics.NewCounter("myname", "myhelp")
		return c.Name() == "myname" && c.Help() == "myhelp" && c.Type() == "counter"
	})
	check("Counter 동시성", 8, func() bool {
		c := metrics.NewCounter("cc", "h")
		var wg sync.WaitGroup
		for range 100 {
			wg.Add(1)
			go func() { defer wg.Done(); c.Inc() }()
		}
		wg.Wait()
		return c.Value() == 100
	})

	// Gauge (20점)
	check("Gauge Set/Inc/Dec/Add/Sub", 12, func() bool {
		g := metrics.NewGauge("g1", "h")
		g.Set(10)
		g.Inc()
		g.Dec()
		g.Add(5)
		g.Sub(2)
		return math.Abs(g.Value()-13) < 1e-9
	})
	check("Gauge 동시성", 8, func() bool {
		g := metrics.NewGauge("gc", "h")
		var wg sync.WaitGroup
		for range 50 {
			wg.Add(1)
			go func() { defer wg.Done(); g.Inc(); g.Dec(); g.Add(2); g.Sub(1) }()
		}
		wg.Wait()
		return math.Abs(g.Value()-50) < 1e-9
	})

	// Histogram (25점)
	check("Histogram Observe Count/Sum", 10, func() bool {
		h := metrics.NewHistogram("h1", "h", []float64{0.1, 1.0})
		h.Observe(0.05)
		h.Observe(0.5)
		h.Observe(2.0)
		return h.Count() == 3 && math.Abs(h.Sum()-2.55) < 1e-9
	})
	check("Histogram 버킷 누적 카운트", 10, func() bool {
		h := metrics.NewHistogram("h2", "h", []float64{1.0, 5.0})
		h.Observe(0.5)  // 1.0, 5.0, +Inf
		h.Observe(3.0)  // 5.0, +Inf
		h.Observe(10.0) // +Inf 만
		bkts := h.Buckets()
		return bkts[1.0] == 1 && bkts[5.0] == 2 && bkts[math.Inf(1)] == 3
	})
	check("Histogram +Inf 버킷 항상 존재", 5, func() bool {
		h := metrics.NewHistogram("h3", "h", []float64{1.0})
		h.Observe(999)
		_, ok := h.Buckets()[math.Inf(1)]
		return ok
	})

	// Registry (15점)
	check("Registry Register/Get/List", 10, func() bool {
		reg := metrics.NewRegistry()
		c1 := metrics.NewCounter("z", "z")
		c2 := metrics.NewCounter("a", "a")
		if reg.Register(c1) != nil || reg.Register(c2) != nil {
			return false
		}
		_, ok := reg.Get("z")
		if !ok {
			return false
		}
		list := reg.List()
		return len(list) == 2 && list[0].Name() == "a"
	})
	check("Registry 중복/Unregister 오류 처리", 5, func() bool {
		reg := metrics.NewRegistry()
		c := metrics.NewCounter("x", "x")
		reg.Register(c)
		dupErr := reg.Register(c)
		reg.Unregister("x")
		missErr := reg.Unregister("x")
		return dupErr != nil && missErr != nil
	})

	// MetricsHandler (20점)
	check("MetricsHandler Counter 출력", 7, func() bool {
		reg := metrics.NewRegistry()
		c := metrics.NewCounter("req_total", "요청")
		c.Add(99)
		reg.Register(c)
		h := metrics.MetricsHandler(reg)
		req := httptest.NewRequest("GET", "/metrics", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		body := rw.Body.String()
		return strings.Contains(body, "# TYPE req_total counter") &&
			strings.Contains(body, "req_total 99")
	})
	check("MetricsHandler Histogram 출력", 8, func() bool {
		reg := metrics.NewRegistry()
		h2 := metrics.NewHistogram("dur", "d", []float64{0.1})
		h2.Observe(0.05)
		reg.Register(h2)
		h := metrics.MetricsHandler(reg)
		req := httptest.NewRequest("GET", "/metrics", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		body := rw.Body.String()
		return strings.Contains(body, "dur_bucket{le=") &&
			strings.Contains(body, `{le="+Inf"}`) &&
			strings.Contains(body, "dur_count 1")
	})
	check("MetricsHandler Content-Type", 5, func() bool {
		reg := metrics.NewRegistry()
		h := metrics.MetricsHandler(reg)
		req := httptest.NewRequest("GET", "/metrics", nil)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		return strings.Contains(rw.Header().Get("Content-Type"), "text/plain")
	})

	// 결과 출력
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║     과제 A4: 메트릭 수집 시스템 채점 결과    ║")
	fmt.Println("╠══════════════════════════════════════════════╣")
	passed := 0
	for _, r := range results {
		mark := "✗"
		if r.pass {
			mark = "✓"
			passed++
		}
		fmt.Printf("║  %s %-32s %3d/%d점  ║\n", mark, r.name, r.pts, r.maxPt)
	}
	fmt.Println("╠══════════════════════════════════════════════╣")
	fmt.Printf("║  통과: %d/%d                                   ║\n", passed, len(results))
	fmt.Printf("║  점수: %d/%d                                  ║\n", score, total)
	grade := "F"
	switch {
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	}
	fmt.Printf("║  등급: %s                                      ║\n", grade)
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("=== 채점 결과 ===")
	fmt.Printf("통과: %d/%d\n", passed, len(results))
	fmt.Printf("점수: %d/%d\n", score, total)
}
