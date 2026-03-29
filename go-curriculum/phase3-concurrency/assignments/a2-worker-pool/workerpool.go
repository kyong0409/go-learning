// 패키지 선언
package workerpool

import (
	"context"
	"errors"
	"time"
)

// Job은 워커 풀에서 처리할 작업을 나타냅니다.
type Job struct {
	ID       int
	Filename string
	SizeKB   int
}

// Result는 작업 처리 결과를 나타냅니다.
type Result struct {
	Job      Job
	Output   string
	Duration time.Duration
	Err      error
}

// ProgressEvent는 진행률 이벤트를 나타냅니다.
type ProgressEvent struct {
	Completed int
	Total     int
	Current   Job
}

// ErrPoolShutdown은 종료된 풀에 작업을 제출하려 할 때 반환됩니다.
var ErrPoolShutdown = errors.New("워커 풀이 종료됨")

// Pool은 워커 풀입니다.
type Pool struct {
	// TODO: 필요한 필드를 추가하세요
	// 힌트:
	// - workers int: 워커 수
	// - jobs chan Job: 작업 채널
	// - results chan Result: 결과 채널
	// - progress chan ProgressEvent: 진행률 채널
	// - wg sync.WaitGroup: 워커 완료 추적
	// - shutdown 여부를 나타내는 필드
}

// NewPool은 지정한 수의 워커를 가진 새 Pool을 생성합니다.
func NewPool(workers int) *Pool {
	// TODO: 구현하세요
	// - 채널과 필드 초기화
	panic("구현 필요")
}

// Start는 워커 고루틴들을 시작합니다.
// ctx가 취소되면 모든 워커가 즉시 종료됩니다.
func (p *Pool) Start(ctx context.Context) {
	// TODO: 구현하세요
	// - workers 수만큼 고루틴 시작
	// - 각 워커는 jobs 채널에서 작업을 꺼내 처리
	// - ctx.Done() 감지 시 즉시 종료
	panic("구현 필요")
}

// Submit은 작업을 풀에 제출합니다.
// 풀이 종료됐으면 ErrPoolShutdown을 반환합니다.
func (p *Pool) Submit(job Job) error {
	// TODO: 구현하세요
	// - shutdown 상태면 ErrPoolShutdown 반환
	// - 아니면 jobs 채널에 job 전송
	panic("구현 필요")
}

// Progress는 진행률 이벤트 채널을 반환합니다.
// 각 작업이 완료될 때마다 ProgressEvent가 전송됩니다.
func (p *Pool) Progress() <-chan ProgressEvent {
	// TODO: 구현하세요
	panic("구현 필요")
}

// Results는 결과 채널을 반환합니다.
func (p *Pool) Results() <-chan Result {
	// TODO: 구현하세요
	panic("구현 필요")
}

// Shutdown은 새 작업 수신을 중단하고 진행 중인 작업이 끝나길 기다립니다.
// 이후 Submit 호출은 ErrPoolShutdown을 반환합니다.
func (p *Pool) Shutdown() {
	// TODO: 구현하세요
	// - shutdown 상태로 표시
	// - jobs 채널 닫기 (워커들이 range로 종료)
	// - 모든 워커 완료 대기 (wg.Wait)
	// - results, progress 채널 닫기
	panic("구현 필요")
}

// Wait는 모든 작업이 완료될 때까지 기다리고 결과를 반환합니다.
// Shutdown()을 내부적으로 호출합니다.
func (p *Pool) Wait() []Result {
	// TODO: 구현하세요
	// - Shutdown() 호출
	// - results 채널에서 모든 결과 수집
	// - 결과 슬라이스 반환
	panic("구현 필요")
}

// processJob은 단일 작업을 처리합니다 (내부 함수).
// 실제 파일 처리를 시뮬레이션합니다.
func processJob(ctx context.Context, job Job) Result {
	start := time.Now()

	// SizeKB에 비례한 처리 시간 시뮬레이션 (최대 50ms)
	delay := time.Duration(job.SizeKB) * time.Millisecond
	if delay > 50*time.Millisecond {
		delay = 50 * time.Millisecond
	}

	select {
	case <-time.After(delay):
		return Result{
			Job:      job,
			Output:   "처리 완료: " + job.Filename,
			Duration: time.Since(start),
		}
	case <-ctx.Done():
		return Result{
			Job:      job,
			Duration: time.Since(start),
			Err:      ctx.Err(),
		}
	}
}
