// registry_test.go
// 메트릭 레지스트리 테스트 및 채점
//
// 실행:
//
//	go test -v -race
//	go test -v -run TestGrade
package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// ============================================================
// CounterVec 테스트 (20점)
// ============================================================

func TestCounterVec_Inc(t *testing.T) {
	cv := NewCounterVec("requests_total", "총 요청 수", []string{"method"})
	c := cv.WithLabelValues("GET")
	c.Inc()
	c.Inc()
	c.Inc()
	if c.get() != 3 {
		t.Errorf("Inc 3번 후 value = %g, 원하는 값: 3", c.get())
	}
}

func TestCounterVec_Add(t *testing.T) {
	cv := NewCounterVec("bytes_total", "총 바이트", []string{"dir"})
	c := cv.WithLabelValues("in")
	c.Add(100)
	c.Add(50.5)
	if c.get() != 150.5 {
		t.Errorf("Add 후 value = %g, 원하는 값: 150.5", c.get())
	}
}

func TestCounterVec_NegativeAddPanic(t *testing.T) {
	cv := NewCounterVec("test", "test", []string{})
	c := cv.WithLabelValues()
	defer func() {
		if r := recover(); r == nil {
			t.Error("음수 Add는 panic해야 합니다")
		}
	}()
	c.Add(-1)
}

func TestCounterVec_MultipleLabels(t *testing.T) {
	cv := NewCounterVec("http_requests", "HTTP 요청", []string{"method", "code"})
	cv.WithLabelValues("GET", "200").Inc()
	cv.WithLabelValues("GET", "200").Add(5)
	cv.WithLabelValues("POST", "500").Inc()

	if cv.WithLabelValues("GET", "200").get() != 6 {
		t.Errorf("GET/200 = %g, 원하는 값: 6", cv.WithLabelValues("GET", "200").get())
	}
	if cv.WithLabelValues("POST", "500").get() != 1 {
		t.Errorf("POST/500 = %g, 원하는 값: 1", cv.WithLabelValues("POST", "500").get())
	}
}

func TestCounterVec_SameLabelSameInstance(t *testing.T) {
	cv := NewCounterVec("test", "test", []string{"k"})
	c1 := cv.WithLabelValues("v")
	c2 := cv.WithLabelValues("v")
	c1.Inc()
	if c2.get() != 1 {
		t.Error("같은 레이블은 같은 인스턴스를 반환해야 합니다")
	}
}

// ============================================================
// GaugeVec 테스트 (15점)
// ============================================================

func TestGaugeVec_SetIncDec(t *testing.T) {
	gv := NewGaugeVec("connections", "연결 수", []string{"svc"})
	g := gv.WithLabelValues("api")

	g.Set(100)
	if g.get() != 100 {
		t.Errorf("Set(100) 후 value = %g", g.get())
	}

	g.Inc()
	if g.get() != 101 {
		t.Errorf("Inc 후 value = %g, 원하는 값: 101", g.get())
	}

	g.Dec()
	g.Dec()
	if g.get() != 99 {
		t.Errorf("Dec 2번 후 value = %g, 원하는 값: 99", g.get())
	}
}

func TestGaugeVec_CanGoNegative(t *testing.T) {
	gv := NewGaugeVec("temp", "온도", []string{})
	g := gv.WithLabelValues()
	g.Set(-5.5)
	if g.get() != -5.5 {
		t.Errorf("Gauge 음수값: %g, 원하는 값: -5.5", g.get())
	}
}

func TestGaugeVec_Add(t *testing.T) {
	gv := NewGaugeVec("mem", "메모리", []string{"node"})
	g := gv.WithLabelValues("node-1")
	g.Set(1000)
	g.Add(-200)
	if g.get() != 800 {
		t.Errorf("Add(-200) 후 value = %g, 원하는 값: 800", g.get())
	}
}

// ============================================================
// HistogramVec 테스트 (20점)
// ============================================================

func TestHistogramVec_Observe(t *testing.T) {
	hv := NewHistogramVec("latency", "지연시간",
		[]string{"handler"}, []float64{0.1, 0.5, 1.0})
	h := hv.WithLabelValues("/api")

	h.Observe(0.05)  // 0.1 버킷에 포함
	h.Observe(0.3)   // 0.5 버킷에 포함
	h.Observe(0.8)   // 1.0 버킷에 포함
	h.Observe(2.0)   // +Inf 버킷에만 포함

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.count != 4 {
		t.Errorf("count = %d, 원하는 값: 4", h.count)
	}
	if h.sum != 0.05+0.3+0.8+2.0 {
		t.Errorf("sum = %g, 원하는 값: %g", h.sum, 0.05+0.3+0.8+2.0)
	}
}

func TestHistogramVec_BucketCumulative(t *testing.T) {
	hv := NewHistogramVec("size", "크기",
		[]string{}, []float64{10, 100, 1000})
	h := hv.WithLabelValues()

	h.Observe(5)
	h.Observe(50)
	h.Observe(500)
	h.Observe(5000)

	h.mu.Lock()
	defer h.mu.Unlock()

	// 누적 카운트 확인
	if h.counts[10] != 1 {
		t.Errorf("le=10 count = %d, 원하는 값: 1 (5만 해당)", h.counts[10])
	}
	if h.counts[100] != 2 {
		t.Errorf("le=100 count = %d, 원하는 값: 2 (5,50 해당)", h.counts[100])
	}
	if h.counts[1000] != 3 {
		t.Errorf("le=1000 count = %d, 원하는 값: 3 (5,50,500 해당)", h.counts[1000])
	}
}

// ============================================================
// Registry 테스트 (15점)
// ============================================================

func TestRegistry_RegisterGather(t *testing.T) {
	reg := NewRegistry()

	cv := NewCounterVec("req_total", "요청 수", []string{"method"})
	cv.WithLabelValues("GET").Inc()
	cv.WithLabelValues("GET").Inc()
	cv.WithLabelValues("POST").Add(3)

	if err := reg.Register(cv); err != nil {
		t.Fatalf("Register 오류: %v", err)
	}

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather 오류: %v", err)
	}

	if len(families) == 0 {
		t.Fatal("Gather 결과가 비어있습니다")
	}

	var found *MetricFamily
	for _, f := range families {
		if f.Name == "req_total" {
			found = f
			break
		}
	}
	if found == nil {
		t.Fatal("req_total MetricFamily를 찾을 수 없습니다")
	}
	if len(found.Metrics) != 2 {
		t.Errorf("Metrics 수 = %d, 원하는 값: 2", len(found.Metrics))
	}
}

func TestRegistry_DuplicateRegister(t *testing.T) {
	reg := NewRegistry()
	cv1 := NewCounterVec("dup", "중복", []string{})
	cv2 := NewCounterVec("dup", "중복2", []string{})

	reg.Register(cv1)
	err := reg.Register(cv2)
	if err == nil {
		t.Error("중복 등록은 에러를 반환해야 합니다")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	reg := NewRegistry()
	cv := NewCounterVec("unregtest", "테스트", []string{})
	reg.Register(cv)

	if !reg.Unregister(cv) {
		t.Error("Unregister가 true를 반환해야 합니다")
	}

	families, _ := reg.Gather()
	for _, f := range families {
		if f.Name == "unregtest" {
			t.Error("Unregister 후 Gather에 메트릭이 남아있습니다")
		}
	}
}

// ============================================================
// 텍스트 형식 테스트 (20점)
// ============================================================

func TestWriteTextFormat_Counter(t *testing.T) {
	cv := NewCounterVec("http_req", "HTTP 요청", []string{"method", "code"})
	cv.WithLabelValues("GET", "200").Add(10)
	cv.WithLabelValues("POST", "500").Add(2)

	reg := NewRegistry()
	reg.Register(cv)
	families, _ := reg.Gather()

	text := WriteTextFormat(families)

	if !strings.Contains(text, "# HELP http_req HTTP 요청") {
		t.Error("HELP 줄이 없습니다")
	}
	if !strings.Contains(text, "# TYPE http_req counter") {
		t.Error("TYPE 줄이 없습니다")
	}
	if !strings.Contains(text, "http_req{") {
		t.Error("메트릭 값 줄이 없습니다")
	}
}

func TestWriteTextFormat_Histogram(t *testing.T) {
	hv := NewHistogramVec("duration", "소요 시간",
		[]string{"op"}, []float64{0.1, 1.0})
	hv.WithLabelValues("read").Observe(0.05)
	hv.WithLabelValues("read").Observe(0.5)

	reg := NewRegistry()
	reg.Register(hv)
	families, _ := reg.Gather()

	text := WriteTextFormat(families)

	if !strings.Contains(text, "duration_bucket{") {
		t.Errorf("histogram _bucket 줄이 없습니다\n출력:\n%s", text)
	}
	if !strings.Contains(text, `le="+Inf"`) {
		t.Errorf("+Inf 버킷이 없습니다\n출력:\n%s", text)
	}
	if !strings.Contains(text, "duration_sum{") {
		t.Errorf("histogram _sum 줄이 없습니다\n출력:\n%s", text)
	}
	if !strings.Contains(text, "duration_count{") {
		t.Errorf("histogram _count 줄이 없습니다\n출력:\n%s", text)
	}
}

func TestWriteTextFormat_LabelsSorted(t *testing.T) {
	cv := NewCounterVec("sorted", "정렬 테스트", []string{"z_last", "a_first"})
	cv.WithLabelValues("zval", "aval").Inc()

	reg := NewRegistry()
	reg.Register(cv)
	families, _ := reg.Gather()
	text := WriteTextFormat(families)

	// a_first가 z_last보다 먼저 나와야 함
	aIdx := strings.Index(text, "a_first")
	zIdx := strings.Index(text, "z_last")
	if aIdx == -1 || zIdx == -1 {
		t.Errorf("레이블이 출력에 없습니다:\n%s", text)
		return
	}
	if aIdx > zIdx {
		t.Errorf("레이블이 알파벳 순으로 정렬되지 않았습니다:\n%s", text)
	}
}

// ============================================================
// 동시성 테스트 (10점)
// ============================================================

func TestConcurrentMetrics(t *testing.T) {
	cv := NewCounterVec("concurrent", "동시성 테스트", []string{"worker"})
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cv.WithLabelValues(fmt.Sprintf("worker-%d", i)).Inc()
			}
		}(i)
	}
	wg.Wait()

	total := 0.0
	for i := 0; i < 10; i++ {
		total += cv.WithLabelValues(fmt.Sprintf("worker-%d", i)).get()
	}
	if total != 1000 {
		t.Errorf("동시 Inc 후 합계 = %g, 원하는 값: 1000", total)
	}
}

// ============================================================
// 채점 함수 (TestGrade)
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════════════")
	fmt.Println("  과제 A3: 메트릭 레지스트리 채점 결과")
	fmt.Println("  패턴: Prometheus Go Client")
	fmt.Println("═══════════════════════════════════════════════════")

	// CounterVec (20점)
	t.Run("CounterVec", func(t *testing.T) {
		cv := NewCounterVec("test_c", "test", []string{"k"})
		cv.WithLabelValues("v").Inc()
		cv.WithLabelValues("v").Add(4)
		cv.WithLabelValues("v2").Add(2)

		ok := cv.WithLabelValues("v").get() == 5 &&
			cv.WithLabelValues("v2").get() == 2

		// 음수 panic 확인
		panicked := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			cv.WithLabelValues("v").Add(-1)
		}()

		if ok && panicked {
			score += 20
			fmt.Printf("  ✓ CounterVec (Inc/Add/레이블/음수패닉)20/20점\n")
		} else if ok {
			score += 15
			fmt.Printf("  △ CounterVec (음수패닉 없음)          15/20점\n")
		} else {
			fmt.Printf("  ✗ CounterVec                           0/20점\n")
		}
	})

	// GaugeVec (15점)
	t.Run("GaugeVec", func(t *testing.T) {
		gv := NewGaugeVec("test_g", "test", []string{"k"})
		gv.WithLabelValues("v").Set(10)
		gv.WithLabelValues("v").Inc()
		gv.WithLabelValues("v").Dec()
		gv.WithLabelValues("v").Add(-5)

		if gv.WithLabelValues("v").get() == 5 {
			score += 15
			fmt.Printf("  ✓ GaugeVec (Set/Inc/Dec/Add)          15/15점\n")
		} else {
			fmt.Printf("  ✗ GaugeVec (value=%g, 원하는:5)         0/15점\n",
				gv.WithLabelValues("v").get())
		}
	})

	// HistogramVec (20점)
	t.Run("HistogramVec", func(t *testing.T) {
		hv := NewHistogramVec("test_h", "test", []string{}, []float64{1, 10, 100})
		h := hv.WithLabelValues()
		h.Observe(0.5)
		h.Observe(5)
		h.Observe(50)
		h.Observe(500)

		h.mu.Lock()
		c1 := h.counts[1]
		c10 := h.counts[10]
		c100 := h.counts[100]
		cnt := h.count
		h.mu.Unlock()

		if c1 == 1 && c10 == 2 && c100 == 3 && cnt == 4 {
			score += 20
			fmt.Printf("  ✓ HistogramVec (Observe/버킷/누적)    20/20점\n")
		} else {
			fmt.Printf("  ✗ HistogramVec (le=1:%d,le=10:%d,le=100:%d,count:%d)  0/20점\n",
				c1, c10, c100, cnt)
		}
	})

	// Registry (15점)
	t.Run("Registry", func(t *testing.T) {
		reg := NewRegistry()
		cv := NewCounterVec("reg_test", "test", []string{"x"})
		cv.WithLabelValues("a").Add(7)
		reg.Register(cv)

		families, err := reg.Gather()
		ok := err == nil && len(families) > 0

		// 중복 등록 에러
		cv2 := NewCounterVec("reg_test", "test2", []string{})
		err2 := reg.Register(cv2)

		if ok && err2 != nil {
			score += 15
			fmt.Printf("  ✓ Registry (Register/Gather/중복에러) 15/15점\n")
		} else {
			fmt.Printf("  ✗ Registry (ok=%v, dup_err=%v)          0/15점\n", ok, err2)
		}
	})

	// 텍스트 형식 (20점)
	t.Run("텍스트_형식", func(t *testing.T) {
		reg := NewRegistry()
		cv := NewCounterVec("txt_c", "카운터", []string{"m"})
		cv.WithLabelValues("GET").Add(3)
		reg.Register(cv)

		families, _ := reg.Gather()
		text := WriteTextFormat(families)

		hasHelp := strings.Contains(text, "# HELP txt_c")
		hasType := strings.Contains(text, "# TYPE txt_c counter")
		hasValue := strings.Contains(text, "txt_c{")

		if hasHelp && hasType && hasValue {
			score += 20
			fmt.Printf("  ✓ 텍스트 형식 (HELP/TYPE/값)          20/20점\n")
		} else if hasValue {
			score += 10
			fmt.Printf("  △ 텍스트 형식 (HELP/TYPE 없음)        10/20점\n")
		} else {
			fmt.Printf("  ✗ 텍스트 형식                          0/20점\n")
		}
	})

	fmt.Println("───────────────────────────────────────────────────")
	fmt.Printf("  최종 점수: %d / %d점\n", score, total)
	fmt.Printf("  (동시성 안전 10점은 go test -race로 확인)\n")

	grade := "F"
	switch {
	case score >= 95:
		grade = "A+"
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	}
	fmt.Printf("  등급: %s\n", grade)
	fmt.Print("═══════════════════════════════════════════════════\n\n")
}
