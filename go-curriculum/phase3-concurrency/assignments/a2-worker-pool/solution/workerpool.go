// 패키지 선언
// 참고 솔루션 - 풀기 전에 보지 마세요!
package workerpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
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
	workers   int
	jobs      chan Job
	results   chan Result
	progress  chan ProgressEvent
	wg        sync.WaitGroup
	shutdown  atomic.Bool // 종료 여부 (atomic으로 스레드 안전)
	completed atomic.Int64
	submitted atomic.Int64
}

// NewPool은 지정한 수의 워커를 가진 새 Pool을 생성합니다.
func NewPool(workers int) *Pool {
	if workers <= 0 {
		workers = 1
	}
	return &Pool{
		workers:  workers,
		jobs:     make(chan Job, workers*2),
		results:  make(chan Result, workers*10),
		progress: make(chan ProgressEvent, workers*10),
	}
}

// Start는 워커 고루틴들을 시작합니다.
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Context 취소: 남은 작업 취소 처리
					for job := range p.jobs {
						r := Result{Job: job, Err: ctx.Err(), Duration: 0}
						select {
						case p.results <- r:
						default:
						}
					}
					return
				case job, ok := <-p.jobs:
					if !ok {
						return // 채널 닫힘 = 정상 종료
					}
					result := processJob(ctx, job)
					completed := int(p.completed.Add(1))
					total := int(p.submitted.Load())

					// 결과 전송
					select {
					case p.results <- result:
					case <-ctx.Done():
					}

					// 진행률 이벤트 전송 (블로킹 방지를 위해 non-blocking)
					event := ProgressEvent{
						Completed: completed,
						Total:     total,
						Current:   job,
					}
					select {
					case p.progress <- event:
					default:
					}
				}
			}
		}(i + 1)
	}
}

// Submit은 작업을 풀에 제출합니다.
func (p *Pool) Submit(job Job) error {
	if p.shutdown.Load() {
		return ErrPoolShutdown
	}
	p.submitted.Add(1)
	select {
	case p.jobs <- job:
		return nil
	default:
		// 버퍼가 가득 찬 경우 블로킹 전송
		p.jobs <- job
		return nil
	}
}

// Progress는 진행률 이벤트 채널을 반환합니다.
func (p *Pool) Progress() <-chan ProgressEvent {
	return p.progress
}

// Results는 결과 채널을 반환합니다.
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Shutdown은 새 작업 수신을 중단하고 진행 중인 작업이 끝나길 기다립니다.
func (p *Pool) Shutdown() {
	if p.shutdown.CompareAndSwap(false, true) {
		close(p.jobs) // 워커들이 range 루프를 종료하게 함
		p.wg.Wait()   // 모든 워커 완료 대기
		close(p.results)
		close(p.progress)
	}
}

// Wait는 모든 작업이 완료될 때까지 기다리고 결과를 반환합니다.
func (p *Pool) Wait() []Result {
	p.Shutdown()

	var results []Result
	for r := range p.results {
		results = append(results, r)
	}
	return results
}

// processJob은 단일 작업을 처리합니다.
func processJob(ctx context.Context, job Job) Result {
	start := time.Now()

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
