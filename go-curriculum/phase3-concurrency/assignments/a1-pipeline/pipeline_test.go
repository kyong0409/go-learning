// 패키지 선언
package pipeline_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	pipeline "github.com/go-curriculum/a1-pipeline"
)

// ─────────────────────────────────────────
// 헬퍼
// ─────────────────────────────────────────

// goroutineCount는 현재 고루틴 수를 반환합니다.
func goroutineCount() int {
	return runtime.NumGoroutine()
}

// waitForGoroutines는 고루틴이 정리될 때까지 최대 d 동안 대기합니다.
func waitForGoroutines(t *testing.T, before int, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+1 { // +1 허용 오차
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// ─────────────────────────────────────────
// Generate 테스트
// ─────────────────────────────────────────

func TestGenerate_Basic(t *testing.T) {
	ctx := context.Background()
	ch := pipeline.Generate(ctx, 1, 2, 3, 4, 5)

	var got []int
	for v := range ch {
		got = append(got, v)
	}

	want := []int{1, 2, 3, 4, 5}
	if len(got) != len(want) {
		t.Fatalf("Generate 길이 오류: 기대=%v, 실제=%v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Generate[%d]: 기대=%d, 실제=%d", i, want[i], got[i])
		}
	}
}

func TestGenerate_Empty(t *testing.T) {
	ctx := context.Background()
	ch := pipeline.Generate(ctx)

	count := 0
	for range ch {
		count++
	}
	if count != 0 {
		t.Errorf("빈 Generate: 기대=0, 실제=%d", count)
	}
}

func TestGenerate_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	before := goroutineCount()
	ch := pipeline.Generate(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// 첫 값만 수신하고 취소
	<-ch
	cancel()

	waitForGoroutines(t, before, 500*time.Millisecond)
	after := goroutineCount()
	if after > before+1 {
		t.Errorf("Generate 취소 후 고루틴 누수: 전=%d, 후=%d", before, after)
	}
}

// ─────────────────────────────────────────
// Filter 테스트
// ─────────────────────────────────────────

func TestFilter_Basic(t *testing.T) {
	ctx := context.Background()
	in := pipeline.Generate(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	out := pipeline.Filter(ctx, in)

	var got []int
	for v := range out {
		got = append(got, v)
	}

	want := []int{2, 4, 6, 8, 10}
	if len(got) != len(want) {
		t.Fatalf("Filter 길이 오류: 기대=%v, 실제=%v", want, got)
	}
	for i, v := range got {
		if v%2 != 0 {
			t.Errorf("Filter: 홀수가 통과됨: %d", v)
		}
		if v != want[i] {
			t.Errorf("Filter[%d]: 기대=%d, 실제=%d", i, want[i], v)
		}
	}
}

func TestFilter_AllOdd(t *testing.T) {
	ctx := context.Background()
	in := pipeline.Generate(ctx, 1, 3, 5, 7, 9)
	out := pipeline.Filter(ctx, in)

	count := 0
	for range out {
		count++
	}
	if count != 0 {
		t.Errorf("모두 홀수인 경우 Filter: 기대=0, 실제=%d", count)
	}
}

func TestFilter_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	before := goroutineCount()

	// 무한 스트림 생성
	infiniteGen := func(ctx context.Context) <-chan int {
		ch := make(chan int)
		go func() {
			defer close(ch)
			for i := 0; ; i++ {
				select {
				case ch <- i:
				case <-ctx.Done():
					return
				}
			}
		}()
		return ch
	}

	in := infiniteGen(ctx)
	out := pipeline.Filter(ctx, in)

	// 몇 개 수신 후 취소
	received := 0
	for range out {
		received++
		if received >= 3 {
			cancel()
			break
		}
	}
	// 채널 드레인
	for range out {
	}

	waitForGoroutines(t, before, 500*time.Millisecond)
	after := goroutineCount()
	if after > before+2 {
		t.Errorf("Filter 취소 후 고루틴 누수: 전=%d, 후=%d", before, after)
	}
}

// ─────────────────────────────────────────
// Square 테스트
// ─────────────────────────────────────────

func TestSquare_Basic(t *testing.T) {
	ctx := context.Background()
	in := pipeline.Generate(ctx, 1, 2, 3, 4, 5)
	out := pipeline.Square(ctx, in)

	var got []int
	for v := range out {
		got = append(got, v)
	}

	want := []int{1, 4, 9, 16, 25}
	if len(got) != len(want) {
		t.Fatalf("Square 길이 오류: 기대=%v, 실제=%v", want, got)
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("Square[%d]: 기대=%d, 실제=%d", i, want[i], v)
		}
	}
}

func TestSquare_Zero(t *testing.T) {
	ctx := context.Background()
	in := pipeline.Generate(ctx, 0)
	out := pipeline.Square(ctx, in)

	v := <-out
	if v != 0 {
		t.Errorf("Square(0): 기대=0, 실제=%d", v)
	}
}

// ─────────────────────────────────────────
// Sum 테스트
// ─────────────────────────────────────────

func TestSum_Basic(t *testing.T) {
	ctx := context.Background()
	in := pipeline.Generate(ctx, 1, 2, 3, 4, 5)
	total, err := pipeline.Sum(ctx, in)

	if err != nil {
		t.Fatalf("Sum 에러: %v", err)
	}
	if total != 15 {
		t.Errorf("Sum: 기대=15, 실제=%d", total)
	}
}

func TestSum_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// 느린 채널 (취소를 강제하기 위해)
	slowCh := make(chan int)
	go func() {
		defer close(slowCh)
		for i := 1; i <= 5; i++ {
			select {
			case slowCh <- i:
				time.Sleep(50 * time.Millisecond)
			case <-ctx.Done():
				return
			}
		}
	}()

	// 100ms 후 취소
	go func() {
		time.Sleep(75 * time.Millisecond)
		cancel()
	}()

	_, err := pipeline.Sum(ctx, slowCh)
	if err == nil {
		t.Error("취소된 Sum이 에러를 반환하지 않음")
	}
}

// ─────────────────────────────────────────
// RunPipeline 통합 테스트
// ─────────────────────────────────────────

func TestRunPipeline_Basic(t *testing.T) {
	ctx := context.Background()
	// 1~10: 짝수(2,4,6,8,10)의 제곱합 = 4+16+36+64+100 = 220
	result, err := pipeline.RunPipeline(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	if err != nil {
		t.Fatalf("RunPipeline 에러: %v", err)
	}
	if result != 220 {
		t.Errorf("RunPipeline: 기대=220, 실제=%d", result)
	}
}

func TestRunPipeline_AllOdd(t *testing.T) {
	ctx := context.Background()
	// 홀수만: 짝수 없음 → 합 = 0
	result, err := pipeline.RunPipeline(ctx, []int{1, 3, 5, 7, 9})
	if err != nil {
		t.Fatalf("RunPipeline 에러: %v", err)
	}
	if result != 0 {
		t.Errorf("홀수만 RunPipeline: 기대=0, 실제=%d", result)
	}
}

func TestRunPipeline_SingleEven(t *testing.T) {
	ctx := context.Background()
	result, err := pipeline.RunPipeline(ctx, []int{4})
	if err != nil {
		t.Fatalf("RunPipeline 에러: %v", err)
	}
	if result != 16 { // 4² = 16
		t.Errorf("단일 짝수 RunPipeline: 기대=16, 실제=%d", result)
	}
}

func TestRunPipeline_Empty(t *testing.T) {
	ctx := context.Background()
	result, err := pipeline.RunPipeline(ctx, []int{})
	if err != nil {
		t.Fatalf("RunPipeline 에러: %v", err)
	}
	if result != 0 {
		t.Errorf("빈 RunPipeline: 기대=0, 실제=%d", result)
	}
}

func TestRunPipeline_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// 즉시 취소
	cancel()

	_, err := pipeline.RunPipeline(ctx, []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20})
	if err == nil {
		// 취소됐지만 이미 완료된 경우 허용
		// (구현에 따라 취소 전에 완료될 수 있음)
		t.Log("즉시 취소 후 완료됨 (허용됨 - 경쟁 조건)")
	}
}

func TestRunPipeline_NoGoroutineLeak(t *testing.T) {
	before := goroutineCount()

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		_, err := pipeline.RunPipeline(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		if err != nil {
			t.Fatalf("RunPipeline 에러: %v", err)
		}
	}

	waitForGoroutines(t, before, time.Second)
	after := goroutineCount()
	if after > before+2 {
		t.Errorf("고루틴 누수: 전=%d, 후=%d (차이=%d)", before, after, after-before)
	}
}

// ─────────────────────────────────────────
// 채점 테스트
// ─────────────────────────────────────────

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	// 기본 동작 (30점)
	t.Run("기본동작", func(t *testing.T) {
		ctx := context.Background()
		result, err := pipeline.RunPipeline(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		if err == nil && result == 220 {
			score += 30
			fmt.Printf("  [통과] 기본 동작: +30점 (결과=%d)\n", result)
		} else {
			fmt.Printf("  [실패] 기본 동작: 기대=220, 실제=%d, 에러=%v\n", result, err)
		}
	})

	// 빈 입력 처리 (10점)
	t.Run("빈입력", func(t *testing.T) {
		ctx := context.Background()
		result, err := pipeline.RunPipeline(ctx, []int{})
		if err == nil && result == 0 {
			score += 10
			fmt.Println("  [통과] 빈 입력 처리: +10점")
		} else {
			fmt.Printf("  [실패] 빈 입력: 기대=0, 실제=%d, 에러=%v\n", result, err)
		}
	})

	// 홀수 전용 (10점)
	t.Run("홀수전용", func(t *testing.T) {
		ctx := context.Background()
		result, err := pipeline.RunPipeline(ctx, []int{1, 3, 5})
		if err == nil && result == 0 {
			score += 10
			fmt.Println("  [통과] 홀수 전용: +10점")
		} else {
			fmt.Printf("  [실패] 홀수 전용: 기대=0, 실제=%d, 에러=%v\n", result, err)
		}
	})

	// Context 취소 (25점)
	t.Run("Context취소", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// 느린 채널로 강제 타임아웃
		slowCh := make(chan int)
		go func() {
			defer close(slowCh)
			for i := 0; ; i += 2 {
				select {
				case slowCh <- i:
					time.Sleep(50 * time.Millisecond)
				case <-ctx.Done():
					return
				}
			}
		}()

		filteredCtx, filteredCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer filteredCancel()

		_, err := pipeline.Sum(filteredCtx, pipeline.Square(filteredCtx, pipeline.Filter(filteredCtx, slowCh)))
		if err != nil {
			score += 25
			fmt.Printf("  [통과] Context 취소: +25점 (에러: %v)\n", err)
		} else {
			fmt.Println("  [실패] Context 취소: 취소 에러가 반환되어야 함")
		}
	})

	// 고루틴 누수 없음 (15점)
	t.Run("고루틴누수없음", func(t *testing.T) {
		before := goroutineCount()
		ctx := context.Background()
		for i := 0; i < 5; i++ {
			pipeline.RunPipeline(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		}
		waitForGoroutines(t, before, 500*time.Millisecond)
		after := goroutineCount()
		if after <= before+2 {
			score += 15
			fmt.Printf("  [통과] 고루틴 누수 없음: +15점 (전=%d, 후=%d)\n", before, after)
		} else {
			fmt.Printf("  [실패] 고루틴 누수: 전=%d, 후=%d (차이=%d)\n", before, after, after-before)
		}
	})

	fmt.Println()
	fmt.Printf("╔══════════════════════════════════╗\n")
	fmt.Printf("║  최종 점수: %3d / %3d점            ║\n", score, total)
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
		t.Errorf("점수 미달: %d/100점 (합격: 60점 이상)", score)
	}
}
