// loadtest_test.go
// 부하 테스터 테스트 및 채점
//
// 실행:
//   go test -v
//   go test -v -run TestGrade
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
// 테스트 서버 헬퍼
// ============================================================

// newTestServer는 테스트용 HTTP 서버를 생성합니다.
func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// newEchoServer는 항상 200을 반환하는 서버를 생성합니다.
func newEchoServer() *httptest.Server {
	return newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond) // 약간의 지연
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

// ============================================================
// LoadTester 기본 테스트 (25점)
// ============================================================

func TestLoadTester_BasicRun(t *testing.T) {
	srv := newEchoServer()
	defer srv.Close()

	lt := NewLoadTester(Config{
		URL:         srv.URL,
		Method:      "GET",
		Concurrency: 2,
		Requests:    10,
		Timeout:     5 * time.Second,
	})

	ctx := context.Background()
	result, err := lt.Run(ctx)

	if err != nil {
		t.Fatalf("Run 오류: %v", err)
	}

	if result.TotalRequests != 10 {
		t.Errorf("총 요청 수: 기대 10, 실제 %d", result.TotalRequests)
	}
	if result.Successes != 10 {
		t.Errorf("성공 수: 기대 10, 실제 %d", result.Successes)
	}
	if result.Failures != 0 {
		t.Errorf("실패 수: 기대 0, 실제 %d", result.Failures)
	}
}

func TestLoadTester_Concurrency(t *testing.T) {
	// 동시성이 실제로 동작하는지 확인
	var concurrent int64
	var maxConcurrent int64

	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		curr := atomic.AddInt64(&concurrent, 1)
		defer atomic.AddInt64(&concurrent, -1)

		// 최대 동시 수 업데이트
		for {
			max := atomic.LoadInt64(&maxConcurrent)
			if curr <= max || atomic.CompareAndSwapInt64(&maxConcurrent, max, curr) {
				break
			}
		}

		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	lt := NewLoadTester(Config{
		URL:         srv.URL,
		Concurrency: 5,
		Requests:    20,
		Timeout:     5 * time.Second,
	})

	_, err := lt.Run(context.Background())
	if err != nil {
		t.Fatalf("Run 오류: %v", err)
	}

	// 최소 2개 이상 동시 실행됐어야 함
	if maxConcurrent < 2 {
		t.Errorf("동시성 부족: 최대 동시 요청 %d개 (기대 >= 2)", maxConcurrent)
	}
}

// ============================================================
// 시간 기반 종료 테스트 (15점)
// ============================================================

func TestLoadTester_DurationBased(t *testing.T) {
	srv := newEchoServer()
	defer srv.Close()

	duration := 200 * time.Millisecond

	lt := NewLoadTester(Config{
		URL:         srv.URL,
		Concurrency: 2,
		Duration:    duration, // 요청 수 대신 시간 기반
		Timeout:     5 * time.Second,
	})

	start := time.Now()
	result, err := lt.Run(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Run 오류: %v", err)
	}

	// 실행 시간이 Duration 이상이어야 함
	if elapsed < duration {
		t.Errorf("실행 시간 부족: %v < %v", elapsed, duration)
	}
	// 너무 오래 실행되면 안 됨 (2배 이내)
	if elapsed > duration*3 {
		t.Errorf("실행 시간 초과: %v > %v", elapsed, duration*3)
	}

	if result.TotalRequests == 0 {
		t.Error("시간 기반 실행에서 요청이 0개")
	}
}

func TestLoadTester_ContextCancel(t *testing.T) {
	// 느린 서버
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	lt := NewLoadTester(Config{
		URL:         srv.URL,
		Concurrency: 2,
		Requests:    1000, // 많은 요청
		Timeout:     5 * time.Second,
	})

	start := time.Now()
	_, err := lt.Run(ctx)
	elapsed := time.Since(start)

	// 컨텍스트 취소로 조기 종료되어야 함
	_ = err // 오류는 허용
	if elapsed > 500*time.Millisecond {
		t.Errorf("컨텍스트 취소 후 너무 오래 실행: %v", elapsed)
	}
}

// ============================================================
// Stats 테스트 (20점 - 카운터 + 25점 - 백분위수)
// ============================================================

func TestStats_Record(t *testing.T) {
	s := NewStats()

	s.Record(10*time.Millisecond, 200, nil)
	s.Record(20*time.Millisecond, 200, nil)
	s.Record(30*time.Millisecond, 500, nil) // 서버 오류

	if s.Count() != 3 {
		t.Errorf("Count: 기대 3, 실제 %d", s.Count())
	}
	if s.ErrorCount() != 1 {
		t.Errorf("ErrorCount: 기대 1, 실제 %d", s.ErrorCount())
	}
}

func TestStats_Percentile(t *testing.T) {
	s := NewStats()

	// 1~100ms 순서대로 기록 (1ms 간격)
	latencies := make([]time.Duration, 100)
	for i := 0; i < 100; i++ {
		d := time.Duration(i+1) * time.Millisecond
		latencies[i] = d
		s.Record(d, 200, nil)
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	p50 := s.Percentile(50)
	p90 := s.Percentile(90)
	p99 := s.Percentile(99)

	// P50은 약 50ms 근처여야 함
	if p50 < 45*time.Millisecond || p50 > 55*time.Millisecond {
		t.Errorf("P50: 기대 ~50ms, 실제 %v", p50)
	}

	// P90은 약 90ms 근처여야 함
	if p90 < 85*time.Millisecond || p90 > 95*time.Millisecond {
		t.Errorf("P90: 기대 ~90ms, 실제 %v", p90)
	}

	// P99는 약 99ms여야 함
	if p99 < 95*time.Millisecond || p99 > 100*time.Millisecond {
		t.Errorf("P99: 기대 ~99ms, 실제 %v", p99)
	}

	// P50 < P90 < P99 순서여야 함
	if !(p50 < p90 && p90 < p99) {
		t.Errorf("백분위수 순서 오류: P50=%v, P90=%v, P99=%v", p50, p90, p99)
	}
}

func TestStats_Empty(t *testing.T) {
	s := NewStats()

	// 데이터 없을 때 패닉 없이 0 반환
	p50 := s.Percentile(50)
	if p50 != 0 {
		t.Errorf("빈 Stats의 Percentile: 기대 0, 실제 %v", p50)
	}

	result := s.Summary(time.Second)
	if result.TotalRequests != 0 {
		t.Errorf("빈 Stats의 Summary.TotalRequests: 기대 0, 실제 %d", result.TotalRequests)
	}
}

// ============================================================
// 오류율 테스트 (15점)
// ============================================================

func TestLoadTester_ErrorRate(t *testing.T) {
	var requestCount int64

	// 3번 중 1번 오류 반환
	srv := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt64(&requestCount, 1)
		if n%3 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	lt := NewLoadTester(Config{
		URL:         srv.URL,
		Concurrency: 1,
		Requests:    30,
		Timeout:     5 * time.Second,
	})

	result, err := lt.Run(context.Background())
	if err != nil {
		t.Fatalf("Run 오류: %v", err)
	}

	// 약 1/3이 실패여야 함 (허용 오차 ±3)
	expectedFailures := 10
	if result.Failures < expectedFailures-3 || result.Failures > expectedFailures+3 {
		t.Errorf("실패 수: 기대 ~%d, 실제 %d", expectedFailures, result.Failures)
	}
}

// ============================================================
// 채점 함수
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════")
	fmt.Println("  과제 A2: HTTP 부하 테스트 도구 채점 결과")
	fmt.Println("═══════════════════════════════════════════")

	// 동시 요청 (25점)
	t.Run("동시_요청", func(t *testing.T) {
		srv := newEchoServer()
		defer srv.Close()

		lt := NewLoadTester(Config{
			URL: srv.URL, Concurrency: 3, Requests: 9, Timeout: 5 * time.Second,
		})
		result, err := lt.Run(context.Background())
		if err == nil && result != nil && result.TotalRequests == 9 && result.Successes == 9 {
			score += 25
			fmt.Printf("  ✓ 동시 요청 실행         25/25점\n")
		} else {
			fmt.Printf("  ✗ 동시 요청 실행         0/25점\n")
		}
	})

	// 시간 기반 종료 (15점)
	t.Run("시간_종료", func(t *testing.T) {
		srv := newEchoServer()
		defer srv.Close()

		d := 150 * time.Millisecond
		lt := NewLoadTester(Config{
			URL: srv.URL, Concurrency: 2, Duration: d, Timeout: 5 * time.Second,
		})
		start := time.Now()
		result, err := lt.Run(context.Background())
		elapsed := time.Since(start)

		if err == nil && result != nil && elapsed >= d && result.TotalRequests > 0 {
			score += 15
			fmt.Printf("  ✓ 시간 기반 종료         15/15점\n")
		} else {
			fmt.Printf("  ✗ 시간 기반 종료         0/15점\n")
		}
	})

	// 통계 수집 (20점)
	t.Run("통계_수집", func(t *testing.T) {
		s := NewStats()
		s.Record(10*time.Millisecond, 200, nil)
		s.Record(20*time.Millisecond, 404, nil)

		if s.Count() == 2 && s.ErrorCount() == 1 {
			score += 20
			fmt.Printf("  ✓ 통계 수집              20/20점\n")
		} else {
			fmt.Printf("  ✗ 통계 수집              0/20점 (Count=%d, Errors=%d)\n", s.Count(), s.ErrorCount())
		}
	})

	// 백분위수 (25점)
	t.Run("백분위수", func(t *testing.T) {
		s := NewStats()
		for i := 1; i <= 100; i++ {
			s.Record(time.Duration(i)*time.Millisecond, 200, nil)
		}
		p50 := s.Percentile(50)
		p99 := s.Percentile(99)

		if p50 >= 45*time.Millisecond && p50 <= 55*time.Millisecond &&
			p99 >= 95*time.Millisecond && p99 <= 100*time.Millisecond {
			score += 25
			fmt.Printf("  ✓ 백분위수 계산          25/25점\n")
		} else {
			fmt.Printf("  ✗ 백분위수 계산          0/25점 (P50=%v, P99=%v)\n", p50, p99)
		}
	})

	// 오류율 (15점)
	t.Run("오류율", func(t *testing.T) {
		s := NewStats()
		for i := 0; i < 10; i++ {
			s.Record(10*time.Millisecond, 200, nil)
		}
		for i := 0; i < 5; i++ {
			s.Record(5*time.Millisecond, 500, nil)
		}
		result := s.Summary(time.Second)

		if result.Failures == 5 && result.Successes == 10 {
			score += 15
			fmt.Printf("  ✓ 오류율 집계            15/15점\n")
		} else {
			fmt.Printf("  ✗ 오류율 집계            0/15점 (실패=%d, 성공=%d)\n", result.Failures, result.Successes)
		}
	})

	fmt.Println("───────────────────────────────────────────")
	fmt.Printf("  최종 점수: %d / %d점\n", score, total)
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
	fmt.Printf("  등급: %s\n", grade)
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println()
}
