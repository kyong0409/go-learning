// 패키지 선언
package main

// 동시성 패턴: 워커 풀 (Worker Pool)
//
// 워커 풀 패턴:
// - 고정된 수의 워커 고루틴이 작업 채널에서 작업을 꺼내 처리
// - 고루틴 수를 제한해 리소스 사용 제어
// - 작업 큐가 있어 작업이 몰려도 처리 가능
//
// 사용 사례: HTTP 요청 처리, DB 쿼리, 파일 처리, 이미지 변환 등

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 기본 Job/Result 타입
// ─────────────────────────────────────────

// Job은 워커가 처리할 작업을 나타냅니다.
type Job struct {
	ID      int
	Payload int // 처리할 데이터
}

// Result는 작업 처리 결과를 나타냅니다.
type Result struct {
	JobID    int
	WorkerID int
	Output   int
	Duration time.Duration
	Err      error
}

// ─────────────────────────────────────────
// 방법 1: 기본 워커 풀
// ─────────────────────────────────────────

// processJob은 단일 작업을 처리합니다.
func processJob(workerID int, job Job) Result {
	start := time.Now()

	// 작업 시뮬레이션: 0~100ms 소요
	delay := time.Duration(rand.Intn(100)) * time.Millisecond
	time.Sleep(delay)

	// 간단한 계산 (제곱)
	output := job.Payload * job.Payload

	return Result{
		JobID:    job.ID,
		WorkerID: workerID,
		Output:   output,
		Duration: time.Since(start),
	}
}

// basicWorker는 기본 워커 함수입니다.
func basicWorker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("  워커 #%d 시작\n", id)

	for job := range jobs { // 채널이 닫힐 때까지 작업 수신
		fmt.Printf("  워커 #%d: 작업 #%d 처리 중 (payload=%d)\n",
			id, job.ID, job.Payload)
		result := processJob(id, job)
		results <- result
	}

	fmt.Printf("  워커 #%d 종료\n", id)
}

func basicWorkerPool() {
	fmt.Println("\n--- 1. 기본 워커 풀 ---")

	const (
		numWorkers = 3
		numJobs    = 9
	)

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// 워커 시작
	var wg sync.WaitGroup
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go basicWorker(i, jobs, results, &wg)
	}

	// 작업 전송
	fmt.Printf("  %d개 작업 전송 중...\n", numJobs)
	for i := 1; i <= numJobs; i++ {
		jobs <- Job{ID: i, Payload: i * 10}
	}
	close(jobs) // 더 이상 작업 없음 → 워커들이 range 루프 종료

	// 워커 완료 대기 후 results 채널 닫기
	go func() {
		wg.Wait()
		close(results)
	}()

	// 결과 수집
	fmt.Println("\n  처리 결과:")
	total := 0
	for result := range results {
		fmt.Printf("  작업 #%d → 워커 #%d: %d² = %d (%v)\n",
			result.JobID, result.WorkerID, result.JobID*10,
			result.Output, result.Duration.Round(time.Millisecond))
		total += result.Output
	}
	fmt.Printf("  총합: %d\n", total)
}

// ─────────────────────────────────────────
// 방법 2: Context를 지원하는 워커 풀
// ─────────────────────────────────────────

// WorkerPool은 Context를 지원하는 워커 풀 구조체입니다.
type WorkerPool struct {
	numWorkers int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
}

// NewWorkerPool은 새 워커 풀을 생성합니다.
func NewWorkerPool(numWorkers, jobBufferSize int) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan Job, jobBufferSize),
		results:    make(chan Result, jobBufferSize),
	}
}

// contextAwareWorker는 Context 취소를 지원하는 워커입니다.
func (p *WorkerPool) contextAwareWorker(ctx context.Context, id int) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("  워커 #%d: Context 취소, 종료\n", id)
			return
		case job, ok := <-p.jobs:
			if !ok {
				fmt.Printf("  워커 #%d: 작업 채널 닫힘, 정상 종료\n", id)
				return
			}

			// 작업 처리 (각 단계에서도 취소 확인)
			result := p.processWithContext(ctx, id, job)
			if result.Err != nil {
				fmt.Printf("  워커 #%d: 작업 #%d 취소됨\n", id, job.ID)
				continue
			}

			// 결과 전송 (취소 시 버림)
			select {
			case p.results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// processWithContext는 Context를 확인하며 작업을 처리합니다.
func (p *WorkerPool) processWithContext(ctx context.Context, workerID int, job Job) Result {
	start := time.Now()

	// 여러 단계로 나뉜 처리 (각 단계에서 취소 확인)
	timer := time.NewTimer(time.Duration(rand.Intn(150)) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return Result{JobID: job.ID, Err: ctx.Err()}
	case <-timer.C:
		return Result{
			JobID:    job.ID,
			WorkerID: workerID,
			Output:   job.Payload * job.Payload,
			Duration: time.Since(start),
		}
	}
}

// Start는 워커 풀을 시작합니다.
func (p *WorkerPool) Start(ctx context.Context) {
	for i := 1; i <= p.numWorkers; i++ {
		p.wg.Add(1)
		go p.contextAwareWorker(ctx, i)
	}
}

// Submit은 작업을 워커 풀에 제출합니다.
func (p *WorkerPool) Submit(job Job) bool {
	select {
	case p.jobs <- job:
		return true
	default:
		return false // 버퍼 가득 참
	}
}

// Stop은 워커 풀을 정상 종료합니다.
func (p *WorkerPool) Stop() <-chan Result {
	close(p.jobs) // 워커들에게 종료 신호
	go func() {
		p.wg.Wait()
		close(p.results)
	}()
	return p.results
}

func contextWorkerPool() {
	fmt.Println("\n--- 2. Context 지원 워커 풀 ---")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	pool := NewWorkerPool(4, 20)
	pool.Start(ctx)

	// 15개 작업 제출
	submitted := 0
	for i := 1; i <= 15; i++ {
		if pool.Submit(Job{ID: i, Payload: i}) {
			submitted++
		}
	}
	fmt.Printf("  %d개 작업 제출됨\n", submitted)

	// 결과 수집
	results := pool.Stop()
	completed := 0
	for result := range results {
		if result.Err == nil {
			completed++
			fmt.Printf("  작업 #%d 완료: %d² = %d\n",
				result.JobID, result.JobID, result.Output)
		}
	}
	fmt.Printf("  완료: %d/%d개\n", completed, submitted)
	fmt.Printf("  종료 이유: %v\n", ctx.Err())
}

// ─────────────────────────────────────────
// 방법 3: 에러 수집 워커 풀
// ─────────────────────────────────────────

// FileJob은 파일 처리 작업을 나타냅니다.
type FileJob struct {
	Path string
	Size int
}

// processFile은 파일 처리를 시뮬레이션합니다.
func processFile(job FileJob) (string, error) {
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	// 10% 확률로 에러 발생
	if rand.Intn(10) == 0 {
		return "", fmt.Errorf("파일 처리 실패: %s (권한 없음)", job.Path)
	}

	return fmt.Sprintf("처리 완료: %s (%d bytes)", job.Path, job.Size), nil
}

func errorCollectingPool() {
	fmt.Println("\n--- 3. 에러 수집 워커 풀 ---")

	files := []FileJob{
		{"/data/file1.txt", 1024},
		{"/data/file2.txt", 2048},
		{"/data/file3.txt", 512},
		{"/data/file4.txt", 4096},
		{"/data/file5.txt", 768},
		{"/data/file6.txt", 1536},
		{"/data/file7.txt", 3072},
		{"/data/file8.txt", 2560},
	}

	type fileResult struct {
		job     FileJob
		message string
		err     error
	}

	jobs := make(chan FileJob, len(files))
	results := make(chan fileResult, len(files))

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				msg, err := processFile(job)
				results <- fileResult{job: job, message: msg, err: err}
			}
		}(i)
	}

	// 작업 전송
	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	// 완료 대기 후 결과 채널 닫기
	go func() {
		wg.Wait()
		close(results)
	}()

	// 결과 및 에러 수집
	var errors []error
	successCount := 0

	for r := range results {
		if r.err != nil {
			errors = append(errors, r.err)
			fmt.Printf("  에러: %v\n", r.err)
		} else {
			successCount++
			fmt.Printf("  성공: %s\n", r.message)
		}
	}

	fmt.Printf("\n  처리 결과: 성공 %d개, 실패 %d개\n",
		successCount, len(errors))
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== 동시성 패턴: 워커 풀 (Worker Pool) ===")

	basicWorkerPool()
	contextWorkerPool()
	errorCollectingPool()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
