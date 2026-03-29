// stats.go
// 통계 수집 및 백분위수 계산
//
// TODO: Stats 구조체와 메서드를 구현하세요.
package main

import (
	"sync"
	"time"
)

// ============================================================
// TODO: 아래를 구현하세요
// ============================================================

// Stats는 부하 테스트 중 통계를 수집합니다.
// 동시에 여러 고루틴에서 접근하므로 스레드 안전해야 합니다.
//
// 구현 요구사항:
//   - 동시성 안전 (sync.Mutex 또는 sync/atomic 사용)
//   - 지연시간 슬라이스 저장 (백분위수 계산용)
//   - 성공/실패 카운터
type Stats struct {
	mu        sync.Mutex    // 동시성 보호
	latencies []time.Duration // 모든 요청의 지연시간
	successes int           // 성공 요청 수
	failures  int           // 실패 요청 수
	// TODO: 필요한 필드를 추가하세요
}

// NewStats는 Stats 생성자입니다.
func NewStats() *Stats {
	// TODO: 구현하세요
	panic("NewStats: 아직 구현되지 않았습니다")
}

// Record는 단일 요청 결과를 기록합니다.
//
// 구현 요구사항:
//   - 스레드 안전하게 기록합니다.
//   - statusCode가 200-299이면 성공, 나머지는 실패로 분류합니다.
//   - err != nil이면 실패로 분류합니다.
//   - latency를 latencies 슬라이스에 추가합니다.
func (s *Stats) Record(latency time.Duration, statusCode int, err error) {
	// TODO: 구현하세요
	panic("Record: 아직 구현되지 않았습니다")
}

// Percentile은 지연시간의 p번째 백분위수를 반환합니다.
//
// p: 0.0 ~ 100.0 (예: 99.0 = P99)
//
// 구현 요구사항:
//   - 데이터가 없으면 0을 반환합니다.
//   - 정렬된 슬라이스에서 인덱스를 계산합니다.
//   - math.Ceil 또는 반올림으로 인덱스를 구합니다.
//
// 힌트:
//   idx = int(math.Ceil(p/100.0*float64(len(sorted)))) - 1
func (s *Stats) Percentile(p float64) time.Duration {
	// TODO: 구현하세요
	panic("Percentile: 아직 구현되지 않았습니다")
}

// Summary는 수집된 통계에서 Result를 생성합니다.
//
// 구현 요구사항:
//   - 평균 지연시간 = 전체 지연시간 합 / 총 요청 수
//   - RPS = 총 요청 수 / totalDuration.Seconds()
//   - P50, P90, P95, P99 계산
//   - 최소/최대 지연시간
func (s *Stats) Summary(totalDuration time.Duration) Result {
	// TODO: 구현하세요
	panic("Summary: 아직 구현되지 않았습니다")
}

// Count는 현재까지 기록된 총 요청 수를 반환합니다. (실시간 진행 표시용)
func (s *Stats) Count() int {
	// TODO: 구현하세요
	panic("Count: 아직 구현되지 않았습니다")
}

// ErrorCount는 현재까지 기록된 실패 수를 반환합니다.
func (s *Stats) ErrorCount() int {
	// TODO: 구현하세요
	panic("ErrorCount: 아직 구현되지 않았습니다")
}
