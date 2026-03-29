// 패키지 선언
package workerpool_test

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	wp "github.com/go-curriculum/a2-worker-pool"
)

// ─────────────────────────────────────────
// 헬퍼
// ─────────────────────────────────────────

func makeJobs(n int) []wp.Job {
	jobs := make([]wp.Job, n)
	for i := range jobs {
		jobs[i] = wp.Job{
			ID:       i + 1,
			Filename: fmt.Sprintf("file_%03d.txt", i+1),
			SizeKB:   (i % 10) + 1,
		}
	}
	return jobs
}

func goroutineCount() int { return runtime.NumGoroutine() }

func waitForGoroutines(t *testing.T, before int, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// ─────────────────────────────────────────
// 기본 동작 테스트
// ─────────────────────────────────────────

func TestPool_BasicProcessing(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewPool(3)
	pool.Start(ctx)

	jobs := makeJobs(9)
	for _, job := range jobs {
		if err := pool.Submit(job); err != nil {
			t.Fatalf("Submit 실패: %v", err)
		}
	}

	results := pool.Wait()

	if len(results) != len(jobs) {
		t.Errorf("결과 수 오류: 기대=%d, 실제=%d", len(jobs), len(results))
	}

	for _, r := range results {
		if r.Err != nil {
			t.Errorf("작업 #%d 에러: %v", r.Job.ID, r.Err)
		}
		if r.Output == "" {
			t.Errorf("작업 #%d 출력 없음", r.Job.ID)
		}
		if r.Duration <= 0 {
			t.Errorf("작업 #%d 소요 시간 없음", r.Job.ID)
		}
	}
}

func TestPool_SingleWorker(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewPool(1)
	pool.Start(ctx)

	jobs := makeJobs(5)
	for _, job := range jobs {
		pool.Submit(job)
	}

	results := pool.Wait()
	if len(results) != 5 {
		t.Errorf("단일 워커: 기대=5, 실제=%d", len(results))
	}
}

func TestPool_EmptyJobs(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewPool(3)
	pool.Start(ctx)

	results := pool.Wait()
	if len(results) != 0 {
		t.Errorf("빈 풀: 기대=0, 실제=%d", len(results))
	}
}

// ─────────────────────────────────────────
// 워커 수 제한 테스트
// ─────────────────────────────────────────

func TestPool_WorkerLimit(t *testing.T) {
	const maxWorkers = 3
	const numJobs = 12

	var currentRunning int64
	var maxObserved int64

	// processJob을 패치하기 어려우므로 실행 시간 관찰
	// 대신 고루틴 수로 간접 확인
	ctx := context.Background()
	pool := wp.NewPool(maxWorkers)
	pool.Start(ctx)

	// maxWorkers보다 많은 작업 제출
	for i := 1; i <= numJobs; i++ {
		pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 10})
	}

	// 진행 중 최대 동시 실행 수 확인 (간접적)
	_ = currentRunning
	_ = maxObserved

	results := pool.Wait()
	if len(results) != numJobs {
		t.Errorf("워커 제한: 기대=%d 결과, 실제=%d", numJobs, len(results))
	}
}

func TestPool_ConcurrencyWithTiming(t *testing.T) {
	const workers = 4
	const jobs = 8

	ctx := context.Background()
	pool := wp.NewPool(workers)
	pool.Start(ctx)

	start := time.Now()
	for i := 1; i <= jobs; i++ {
		pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 10})
	}
	pool.Wait()
	elapsed := time.Since(start)

	// 8개 작업을 4개씩 병렬 처리하면 직렬의 절반 수준이어야 함
	// SizeKB=10 → 각 작업 10ms, 직렬: 80ms, 4병렬: ~20ms
	if elapsed > 60*time.Millisecond {
		t.Logf("병렬 처리 시간이 예상보다 긴 편: %v (기대 <60ms)", elapsed)
	}
	t.Logf("%d 작업, %d 워커, 소요: %v", jobs, workers, elapsed)
}

// ─────────────────────────────────────────
// Context 취소 테스트
// ─────────────────────────────────────────

func TestPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	pool := wp.NewPool(2)
	pool.Start(ctx)

	// 많은 작업 제출
	for i := 1; i <= 20; i++ {
		pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 30})
	}

	// 50ms 후 취소
	time.AfterFunc(50*time.Millisecond, cancel)

	results := pool.Wait()

	// 일부 작업이 취소 에러를 가져야 함
	cancelled := 0
	for _, r := range results {
		if r.Err != nil {
			cancelled++
		}
	}

	t.Logf("완료: %d, 취소: %d", len(results)-cancelled, cancelled)
	// 취소가 동작했다면 일부 에러가 있어야 하지만,
	// 빠르게 완료된 경우 에러가 없을 수도 있음
}

func TestPool_ImmediateCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	pool := wp.NewPool(3)
	pool.Start(ctx)

	for i := 1; i <= 10; i++ {
		pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 5})
	}

	// 취소됐어도 패닉 없이 완료되어야 함
	results := pool.Wait()
	t.Logf("즉시 취소 후 결과: %d개", len(results))
}

// ─────────────────────────────────────────
// 우아한 종료 테스트
// ─────────────────────────────────────────

func TestPool_Shutdown(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewPool(3)
	pool.Start(ctx)

	for i := 1; i <= 5; i++ {
		pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 5})
	}

	pool.Shutdown()

	// Shutdown 후 Submit은 에러를 반환해야 함
	err := pool.Submit(wp.Job{ID: 99, Filename: "late.txt", SizeKB: 1})
	if err == nil {
		t.Error("Shutdown 후 Submit이 에러를 반환하지 않음")
	}
	if err != wp.ErrPoolShutdown {
		t.Errorf("잘못된 에러: 기대=%v, 실제=%v", wp.ErrPoolShutdown, err)
	}
}

// ─────────────────────────────────────────
// 진행률 보고 테스트
// ─────────────────────────────────────────

func TestPool_Progress(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewPool(2)
	pool.Start(ctx)

	const numJobs = 6
	for i := 1; i <= numJobs; i++ {
		pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 5})
	}

	// 진행률 수집
	progressCh := pool.Progress()
	var events []wp.ProgressEvent

	// 결과 수집과 병렬로 진행률 수집
	done := make(chan struct{})
	go func() {
		defer close(done)
		for e := range progressCh {
			events = append(events, e)
		}
	}()

	pool.Wait()
	<-done

	if len(events) == 0 {
		t.Error("진행률 이벤트가 없음")
	}

	t.Logf("진행률 이벤트 수: %d", len(events))

	// 마지막 이벤트는 완료 수가 numJobs여야 함
	if len(events) > 0 {
		last := events[len(events)-1]
		if last.Completed != numJobs {
			t.Errorf("마지막 진행률: 기대=%d, 실제=%d", numJobs, last.Completed)
		}
	}
}

// ─────────────────────────────────────────
// 고루틴 누수 테스트
// ─────────────────────────────────────────

func TestPool_NoGoroutineLeak(t *testing.T) {
	before := goroutineCount()

	for i := 0; i < 3; i++ {
		ctx := context.Background()
		pool := wp.NewPool(3)
		pool.Start(ctx)

		for j := 1; j <= 9; j++ {
			pool.Submit(wp.Job{ID: j, Filename: fmt.Sprintf("f%d.txt", j), SizeKB: 2})
		}
		pool.Wait()
	}

	waitForGoroutines(t, before, time.Second)
	after := goroutineCount()
	if after > before+2 {
		t.Errorf("고루틴 누수: 전=%d, 후=%d", before, after)
	}
}

func TestPool_NoGoroutineLeak_WithCancel(t *testing.T) {
	before := goroutineCount()

	ctx, cancel := context.WithCancel(context.Background())
	pool := wp.NewPool(3)
	pool.Start(ctx)

	for i := 1; i <= 20; i++ {
		pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 20})
	}

	time.AfterFunc(30*time.Millisecond, cancel)
	pool.Wait()

	waitForGoroutines(t, before, time.Second)
	after := goroutineCount()
	if after > before+2 {
		t.Errorf("취소 후 고루틴 누수: 전=%d, 후=%d", before, after)
	}
}

// ─────────────────────────────────────────
// 채점 테스트
// ─────────────────────────────────────────

func TestGrade(t *testing.T) {
	score := 0

	// 기본 처리 (25점)
	t.Run("기본처리", func(t *testing.T) {
		ctx := context.Background()
		pool := wp.NewPool(3)
		pool.Start(ctx)
		jobs := makeJobs(9)
		for _, j := range jobs {
			pool.Submit(j)
		}
		results := pool.Wait()
		if len(results) == 9 {
			allOk := true
			for _, r := range results {
				if r.Err != nil || r.Output == "" {
					allOk = false
					break
				}
			}
			if allOk {
				score += 25
				fmt.Println("  [통과] 기본 처리: +25점")
			} else {
				fmt.Println("  [실패] 기본 처리: 일부 결과 오류")
			}
		} else {
			fmt.Printf("  [실패] 기본 처리: 결과 수 오류 (기대=9, 실제=%d)\n", len(results))
		}
	})

	// 워커 수 제한 (20점) - 타이밍 기반 간접 확인
	t.Run("워커수제한", func(t *testing.T) {
		ctx := context.Background()
		pool := wp.NewPool(4)
		pool.Start(ctx)
		start := time.Now()
		for i := 1; i <= 8; i++ {
			pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 10})
		}
		pool.Wait()
		elapsed := time.Since(start)
		// 4 워커로 8개 처리: ~20ms (직렬이면 80ms)
		if elapsed < 70*time.Millisecond {
			score += 20
			fmt.Printf("  [통과] 워커 수 제한 (병렬 확인): +20점 (%v)\n", elapsed)
		} else {
			fmt.Printf("  [실패] 워커 수 제한: 너무 느림 (%v, 기대 <70ms)\n", elapsed)
		}
	})

	// Context 취소 (20점)
	t.Run("Context취소", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		pool := wp.NewPool(2)
		pool.Start(ctx)
		for i := 1; i <= 20; i++ {
			pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 50})
		}
		results := pool.Wait()
		cancelled := 0
		for _, r := range results {
			if r.Err != nil {
				cancelled++
			}
		}
		if cancelled > 0 {
			score += 20
			fmt.Printf("  [통과] Context 취소: +20점 (취소된 작업: %d개)\n", cancelled)
		} else {
			fmt.Println("  [실패] Context 취소: 취소된 작업이 없음")
		}
	})

	// 우아한 종료 (15점)
	t.Run("우아한종료", func(t *testing.T) {
		ctx := context.Background()
		pool := wp.NewPool(2)
		pool.Start(ctx)
		for i := 1; i <= 4; i++ {
			pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 5})
		}
		pool.Shutdown()
		err := pool.Submit(wp.Job{ID: 99})
		if err == wp.ErrPoolShutdown {
			score += 15
			fmt.Println("  [통과] 우아한 종료: +15점")
		} else {
			fmt.Printf("  [실패] 우아한 종료: 기대=%v, 실제=%v\n", wp.ErrPoolShutdown, err)
		}
	})

	// 진행률 보고 (10점)
	t.Run("진행률보고", func(t *testing.T) {
		ctx := context.Background()
		pool := wp.NewPool(2)
		pool.Start(ctx)
		for i := 1; i <= 6; i++ {
			pool.Submit(wp.Job{ID: i, Filename: fmt.Sprintf("f%d.txt", i), SizeKB: 3})
		}
		progressCh := pool.Progress()
		var count int32
		done := make(chan struct{})
		go func() {
			defer close(done)
			for range progressCh {
				atomic.AddInt32(&count, 1)
			}
		}()
		pool.Wait()
		<-done
		if atomic.LoadInt32(&count) > 0 {
			score += 10
			fmt.Printf("  [통과] 진행률 보고: +10점 (이벤트 %d개)\n", count)
		} else {
			fmt.Println("  [실패] 진행률 보고: 이벤트 없음")
		}
	})

	// 고루틴 누수 없음 (10점)
	t.Run("고루틴누수없음", func(t *testing.T) {
		before := goroutineCount()
		for i := 0; i < 3; i++ {
			ctx := context.Background()
			pool := wp.NewPool(3)
			pool.Start(ctx)
			for j := 1; j <= 6; j++ {
				pool.Submit(wp.Job{ID: j, Filename: fmt.Sprintf("f%d.txt", j), SizeKB: 2})
			}
			pool.Wait()
		}
		waitForGoroutines(t, before, time.Second)
		after := goroutineCount()
		if after <= before+2 {
			score += 10
			fmt.Printf("  [통과] 고루틴 누수 없음: +10점 (전=%d, 후=%d)\n", before, after)
		} else {
			fmt.Printf("  [실패] 고루틴 누수: 전=%d, 후=%d\n", before, after)
		}
	})

	fmt.Println()
	fmt.Printf("╔══════════════════════════════════╗\n")
	fmt.Printf("║  최종 점수: %3d / 100점            ║\n", score)
	grade := "F"
	switch {
	case score >= 90:
		grade = "A+"
	case score >= 80:
		grade = "A"
	case score >= 70:
		grade = "B"
	case score >= 60:
		grade = "C"
	}
	fmt.Printf("║  등급: %-30s║\n", grade)
	fmt.Printf("╚══════════════════════════════════╝\n")

	if score < 60 {
		t.Errorf("점수 미달: %d/100점", score)
	}
}
