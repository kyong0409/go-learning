// solution/stats.go - 참고 풀이
package main

import (
	"math"
	"sort"
	"sync"
	"time"
)

// Stats는 부하 테스트 통계를 수집합니다.
type Stats struct {
	mu        sync.Mutex
	latencies []time.Duration
	successes int
	failures  int
}

// NewStats는 Stats 생성자입니다.
func NewStats() *Stats {
	return &Stats{
		latencies: make([]time.Duration, 0, 1000),
	}
}

// Record는 단일 요청 결과를 기록합니다.
func (s *Stats) Record(latency time.Duration, statusCode int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.latencies = append(s.latencies, latency)

	if err != nil || statusCode < 200 || statusCode >= 300 {
		s.failures++
	} else {
		s.successes++
	}
}

// Percentile은 p번째 백분위수를 반환합니다.
func (s *Stats) Percentile(p float64) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.latencies) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(s.latencies))
	copy(sorted, s.latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	idx := int(math.Ceil(p/100.0*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}

	return sorted[idx]
}

// Summary는 최종 결과를 생성합니다.
func (s *Stats) Summary(totalDuration time.Duration) Result {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := len(s.latencies)
	if total == 0 {
		return Result{}
	}

	// 평균 지연시간
	var sum time.Duration
	minL := s.latencies[0]
	maxL := s.latencies[0]
	for _, l := range s.latencies {
		sum += l
		if l < minL {
			minL = l
		}
		if l > maxL {
			maxL = l
		}
	}
	avg := sum / time.Duration(total)

	rps := 0.0
	if totalDuration.Seconds() > 0 {
		rps = float64(total) / totalDuration.Seconds()
	}

	// 정렬 (백분위수용)
	sorted := make([]time.Duration, total)
	copy(sorted, s.latencies)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	pct := func(p float64) time.Duration {
		idx := int(math.Ceil(p/100.0*float64(len(sorted)))) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sorted) {
			idx = len(sorted) - 1
		}
		return sorted[idx]
	}

	return Result{
		TotalRequests: total,
		Successes:     s.successes,
		Failures:      s.failures,
		TotalDuration: totalDuration,
		RPS:           rps,
		AvgLatency:    avg,
		P50:           pct(50),
		P90:           pct(90),
		P95:           pct(95),
		P99:           pct(99),
		MinLatency:    minL,
		MaxLatency:    maxL,
	}
}

// Count는 총 기록된 요청 수를 반환합니다.
func (s *Stats) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.latencies)
}

// ErrorCount는 실패 수를 반환합니다.
func (s *Stats) ErrorCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.failures
}
