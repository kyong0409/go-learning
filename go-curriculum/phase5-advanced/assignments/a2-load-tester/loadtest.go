// loadtest.go
// HTTP 부하 테스트 도구 핵심 로직
//
// TODO: LoadTester 구조체와 Run 메서드를 구현하세요.
package main

import (
	"context"
	"net/http"
	"time"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// Config는 부하 테스트 설정을 담습니다.
type Config struct {
	URL         string        // 대상 URL
	Method      string        // HTTP 메서드 (기본: GET)
	Concurrency int           // 동시 고루틴 수
	Requests    int           // 총 요청 수 (0이면 Duration 사용)
	Duration    time.Duration // 실행 시간 (0이면 Requests 사용)
	Timeout     time.Duration // 단일 요청 타임아웃
	Headers     http.Header   // 추가 헤더
}

// Result는 부하 테스트 최종 결과를 담습니다.
type Result struct {
	TotalRequests int           // 총 요청 수
	Successes     int           // 성공 (2xx) 수
	Failures      int           // 실패 (4xx/5xx + 연결 오류) 수
	TotalDuration time.Duration // 전체 테스트 시간
	RPS           float64       // 초당 요청 수
	AvgLatency    time.Duration // 평균 지연시간
	P50           time.Duration // 50번째 백분위수
	P90           time.Duration // 90번째 백분위수
	P95           time.Duration // 95번째 백분위수
	P99           time.Duration // 99번째 백분위수
	MinLatency    time.Duration // 최소 지연시간
	MaxLatency    time.Duration // 최대 지연시간
}

// ============================================================
// TODO: 아래를 구현하세요
// ============================================================

// LoadTester는 HTTP 부하 테스트를 실행합니다.
type LoadTester struct {
	// TODO: 필드를 추가하세요
}

// NewLoadTester는 LoadTester 생성자입니다.
//
// 구현 요구사항:
//   - config를 저장하세요.
//   - HTTP 클라이언트를 초기화하세요 (타임아웃 설정 포함).
func NewLoadTester(config Config) *LoadTester {
	// TODO: 구현하세요
	panic("NewLoadTester: 아직 구현되지 않았습니다")
}

// Run은 부하 테스트를 실행하고 결과를 반환합니다.
//
// 구현 요구사항:
//   1. config.Concurrency 수만큼 워커 고루틴을 시작합니다.
//   2. 각 워커는 작업 채널에서 요청을 받아 실행합니다.
//   3. config.Requests > 0이면 해당 수만큼 요청 후 종료합니다.
//   4. config.Duration > 0이면 해당 시간 후 종료합니다.
//   5. ctx가 취소되면 즉시 종료합니다.
//   6. 모든 요청 완료 후 Stats.Summary()로 결과를 반환합니다.
//
// 힌트:
//   - 작업 채널: make(chan struct{}, config.Concurrency)
//   - sync.WaitGroup으로 워커 완료 대기
//   - 시간 기반 종료: time.After 또는 context.WithTimeout
func (lt *LoadTester) Run(ctx context.Context) (*Result, error) {
	// TODO: 구현하세요
	panic("Run: 아직 구현되지 않았습니다")
}

// sendRequest는 단일 HTTP 요청을 전송하고 지연시간과 상태 코드를 반환합니다.
//
// 구현 요구사항:
//   - 요청 시작 시각을 기록하고 완료 후 지연시간을 계산합니다.
//   - 응답 Body를 반드시 닫습니다 (defer resp.Body.Close()).
//   - 연결 오류와 HTTP 오류(4xx/5xx)를 구분해 반환합니다.
func (lt *LoadTester) sendRequest() (latency time.Duration, statusCode int, err error) {
	// TODO: 구현하세요
	panic("sendRequest: 아직 구현되지 않았습니다")
}
