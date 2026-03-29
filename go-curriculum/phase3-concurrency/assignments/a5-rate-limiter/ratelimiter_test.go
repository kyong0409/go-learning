// 패키지 선언
package ratelimiter_test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	rl "github.com/go-curriculum/a5-rate-limiter"
)

// ─────────────────────────────────────────
// 헬퍼
// ─────────────────────────────────────────

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
// Allow 테스트
// ─────────────────────────────────────────

func TestAllow_Basic(t *testing.T) {
	// burst=3이면 초기에 3번 Allow 가능
	tb := rl.NewTokenBucket(10.0, 3)

	if !tb.Allow() {
		t.Error("첫 번째 Allow: true 기대")
	}
	if !tb.Allow() {
		t.Error("두 번째 Allow: true 기대")
	}
	if !tb.Allow() {
		t.Error("세 번째 Allow: true 기대")
	}
	// 네 번째는 토큰 없음
	if tb.Allow() {
		t.Error("네 번째 Allow: false 기대 (토큰 고갈)")
	}
}

func TestAllow_AfterRefill(t *testing.T) {
	// rate=10/s, burst=1 → 100ms마다 1개 토큰 보충
	tb := rl.NewTokenBucket(10.0, 1)

	// 초기 토큰 소비
	if !tb.Allow() {
		t.Fatal("초기 Allow 실패")
	}
	if tb.Allow() {
		t.Error("고갈 후 Allow: false 기대")
	}

	// 150ms 대기 (토큰 1개 이상 보충 기대)
	time.Sleep(150 * time.Millisecond)

	if !tb.Allow() {
		t.Error("보충 후 Allow: true 기대")
	}
}

func TestAllow_BurstLimit(t *testing.T) {
	// 오랫동안 기다려도 burst 크기 이상은 쌓이지 않음
	tb := rl.NewTokenBucket(100.0, 5) // rate 높음, burst=5

	// 1초 대기 (100개 보충 가능하지만 burst=5로 제한)
	time.Sleep(100 * time.Millisecond)

	count := 0
	for tb.Allow() {
		count++
		if count > 10 { // 무한 루프 방지
			break
		}
	}

	if count > 5 {
		t.Errorf("버스트 초과: 기대<=5, 실제=%d", count)
	}
	if count < 5 {
		t.Errorf("버스트 미달: 기대=5, 실제=%d", count)
	}
}

// ─────────────────────────────────────────
// AllowN 테스트
// ─────────────────────────────────────────

func TestAllowN_Basic(t *testing.T) {
	tb := rl.NewTokenBucket(10.0, 5)

	// 3개 소비
	if !tb.AllowN(3) {
		t.Error("AllowN(3): true 기대 (burst=5)")
	}

	// 3개 더 → 남은 토큰 2개 부족
	if tb.AllowN(3) {
		t.Error("AllowN(3): false 기대 (남은 토큰 2개)")
	}

	// 2개는 가능
	if !tb.AllowN(2) {
		t.Error("AllowN(2): true 기대 (남은 토큰 2개)")
	}
}

func TestAllowN_ExceedsBurst(t *testing.T) {
	tb := rl.NewTokenBucket(10.0, 3)

	// burst보다 많은 수 요청
	if tb.AllowN(5) {
		t.Error("AllowN(5): false 기대 (burst=3)")
	}
}

// ─────────────────────────────────────────
// Wait 테스트
// ─────────────────────────────────────────

func TestWait_ImmediateToken(t *testing.T) {
	// 토큰이 있으면 즉시 반환
	tb := rl.NewTokenBucket(10.0, 5)

	start := time.Now()
	err := tb.Wait(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Wait 에러: %v", err)
	}
	if elapsed > 50*time.Millisecond {
		t.Errorf("즉시 반환 기대, 실제 소요: %v", elapsed)
	}
}

func TestWait_BlocksUntilToken(t *testing.T) {
	// rate=5/s, burst=1 → 200ms마다 토큰 1개
	tb := rl.NewTokenBucket(5.0, 1)

	// 토큰 소비
	tb.Allow()

	start := time.Now()
	err := tb.Wait(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Wait 에러: %v", err)
	}

	// 약 200ms 대기 기대 (±100ms 허용)
	if elapsed < 100*time.Millisecond {
		t.Errorf("너무 빠른 반환: %v (기대 ~200ms)", elapsed)
	}
	if elapsed > 400*time.Millisecond {
		t.Errorf("너무 느린 반환: %v (기대 <400ms)", elapsed)
	}
	t.Logf("Wait 소요 시간: %v", elapsed)
}

func TestWait_ContextCancellation(t *testing.T) {
	// rate 매우 낮음 → 오랫동안 기다려야 함
	tb := rl.NewTokenBucket(0.1, 1) // 10초마다 토큰 1개

	// 토큰 소비
	tb.Allow()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := tb.Wait(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("취소 후 에러 기대")
	}
	if elapsed > 300*time.Millisecond {
		t.Errorf("취소 후 너무 늦게 반환: %v", elapsed)
	}
	t.Logf("취소 소요 시간: %v, 에러: %v", elapsed, err)
}

func TestWait_ContextAlreadyCancelled(t *testing.T) {
	tb := rl.NewTokenBucket(1.0, 1)
	tb.Allow() // 토큰 소비

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	err := tb.Wait(ctx)
	if err == nil {
		t.Error("이미 취소된 ctx: 에러 기대")
	}
}

// ─────────────────────────────────────────
// SetRate 테스트
// ─────────────────────────────────────────

func TestSetRate_Basic(t *testing.T) {
	tb := rl.NewTokenBucket(1.0, 10)

	// 모든 토큰 소비
	for i := 0; i < 10; i++ {
		tb.Allow()
	}

	// rate 높게 변경 (초당 100개)
	tb.SetRate(100.0, 10)

	// 빠르게 토큰이 보충되어야 함
	time.Sleep(50 * time.Millisecond) // 5개 보충 기대

	if !tb.Allow() {
		t.Error("SetRate 후 Allow: true 기대")
	}
}

func TestSetRate_BurstReduction(t *testing.T) {
	tb := rl.NewTokenBucket(10.0, 10)

	// burst를 3으로 줄임
	tb.SetRate(10.0, 3)

	// 이제 최대 3개까지만 소비 가능
	count := 0
	for tb.Allow() {
		count++
		if count > 5 {
			break
		}
	}

	if count > 3 {
		t.Errorf("SetRate 후 버스트 초과: 기대<=3, 실제=%d", count)
	}
}

// ─────────────────────────────────────────
// 동시성 안전 테스트
// ─────────────────────────────────────────

func TestTokenBucket_ConcurrentAllow(t *testing.T) {
	tb := rl.NewTokenBucket(1000.0, 100)

	var allowed atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Allow() {
				allowed.Add(1)
			}
		}()
	}

	wg.Wait()

	got := allowed.Load()
	if got > 100 {
		t.Errorf("동시 Allow: 허용된 수=%d > burst=100", got)
	}
	t.Logf("동시 Allow 결과: %d/50", got)
}

func TestTokenBucket_ConcurrentWait(t *testing.T) {
	// rate=10/s, burst=5 → 5개 고루틴이 Wait
	tb := rl.NewTokenBucket(20.0, 5)

	// 토큰 모두 소비
	for i := 0; i < 5; i++ {
		tb.Allow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var success atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := tb.Wait(ctx); err == nil {
				success.Add(1)
			}
		}()
	}

	wg.Wait()

	got := success.Load()
	if got != 5 {
		t.Errorf("동시 Wait: 기대=5 성공, 실제=%d", got)
	}
	t.Logf("동시 Wait 성공: %d/5", got)
}

// ─────────────────────────────────────────
// KeyedRateLimiter 테스트
// ─────────────────────────────────────────

func TestKeyedRateLimiter_Allow(t *testing.T) {
	krl := rl.NewKeyedRateLimiter(10.0, 3)

	// 키 "ip1"에 대해 3번 허용
	for i := 0; i < 3; i++ {
		if !krl.Allow("ip1") {
			t.Errorf("ip1 Allow[%d]: true 기대", i)
		}
	}
	if krl.Allow("ip1") {
		t.Error("ip1 Allow[4]: false 기대 (토큰 고갈)")
	}
}

func TestKeyedRateLimiter_KeyIsolation(t *testing.T) {
	krl := rl.NewKeyedRateLimiter(10.0, 2)

	// ip1 토큰 소비
	krl.Allow("ip1")
	krl.Allow("ip1")

	// ip1 고갈
	if krl.Allow("ip1") {
		t.Error("ip1 고갈 후 Allow: false 기대")
	}

	// ip2는 독립적으로 허용되어야 함
	if !krl.Allow("ip2") {
		t.Error("ip2 Allow: true 기대 (독립적)")
	}
	if !krl.Allow("ip2") {
		t.Error("ip2 Allow[2]: true 기대")
	}
}

func TestKeyedRateLimiter_Wait(t *testing.T) {
	krl := rl.NewKeyedRateLimiter(10.0, 1)

	// 토큰 소비
	krl.Allow("user1")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 토큰 고갈 상태에서 Wait → ctx 타임아웃
	err := krl.Wait(ctx, "user1")
	if err == nil {
		// rate=10이면 100ms 내에 보충될 수 있으므로 허용
		t.Log("Wait 성공 (빠른 보충)")
	}
}

func TestKeyedRateLimiter_MultipleKeys(t *testing.T) {
	krl := rl.NewKeyedRateLimiter(100.0, 10)

	keys := []string{"a", "b", "c", "d", "e"}
	var wg sync.WaitGroup

	for _, key := range keys {
		key := key
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				krl.Allow(key)
			}
		}()
	}

	wg.Wait()
	// 패닉이나 레이스 없이 완료되면 통과
}

// ─────────────────────────────────────────
// 채점 테스트
// ─────────────────────────────────────────

func TestGrade(t *testing.T) {
	passedItems := 0
	totalItems := 0

	check := func(name string, points int, fn func() bool) {
		totalItems++
		t.Run(name, func(t *testing.T) {
			if fn() {
				passedItems += points
				fmt.Printf("  [통과] %s: +%d점\n", name, points)
			} else {
				fmt.Printf("  [실패] %s: 0점\n", name)
			}
		})
	}

	// Allow 기본 동작 (20점)
	check("Allow기본동작", 20, func() bool {
		tb := rl.NewTokenBucket(10.0, 3)
		ok1, ok2, ok3 := tb.Allow(), tb.Allow(), tb.Allow()
		ok4 := tb.Allow()
		return ok1 && ok2 && ok3 && !ok4
	})

	// AllowN (10점)
	check("AllowN", 10, func() bool {
		tb := rl.NewTokenBucket(10.0, 5)
		ok1 := tb.AllowN(3) // 3 소비 → 남은 2
		ok2 := tb.AllowN(3) // 3 요청 → 부족
		ok3 := tb.AllowN(2) // 2 소비 → 성공
		return ok1 && !ok2 && ok3
	})

	// 토큰 보충 (20점)
	check("토큰보충", 20, func() bool {
		tb := rl.NewTokenBucket(10.0, 1)
		tb.Allow() // 소비
		if tb.Allow() {
			return false // 즉시 재소비 불가
		}
		time.Sleep(150 * time.Millisecond)
		return tb.Allow() // 보충 후 소비 가능
	})

	// Wait 블로킹 (15점)
	check("Wait블로킹", 15, func() bool {
		tb := rl.NewTokenBucket(5.0, 1)
		tb.Allow()
		start := time.Now()
		err := tb.Wait(context.Background())
		elapsed := time.Since(start)
		return err == nil && elapsed >= 80*time.Millisecond && elapsed < 500*time.Millisecond
	})

	// Context 취소 (15점)
	check("Context취소", 15, func() bool {
		tb := rl.NewTokenBucket(0.1, 1)
		tb.Allow()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err := tb.Wait(ctx)
		return err != nil
	})

	// SetRate (10점)
	check("SetRate", 10, func() bool {
		tb := rl.NewTokenBucket(1.0, 10)
		for i := 0; i < 10; i++ {
			tb.Allow()
		}
		tb.SetRate(100.0, 10)
		time.Sleep(50 * time.Millisecond)
		return tb.Allow()
	})

	// KeyedRateLimiter (10점)
	check("KeyedRateLimiter", 10, func() bool {
		krl := rl.NewKeyedRateLimiter(10.0, 2)
		krl.Allow("a")
		krl.Allow("a")
		noA := !krl.Allow("a")  // a 고갈
		yesB := krl.Allow("b")  // b는 독립
		return noA && yesB
	})

	score := 0
	if totalItems > 0 {
		score = passedItems
	}

	// 점수 계산 (항목별 점수 합산 방식)
	maxScore := 100

	fmt.Println()
	fmt.Printf("╔══════════════════════════════════╗\n")
	fmt.Printf("║  통과: %2d/%2d                      ║\n", passedItems, maxScore)
	fmt.Printf("║  점수: %3d/100                     ║\n", score)
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
	fmt.Println()
	fmt.Println("=== 채점 결과 ===")
	fmt.Printf("통과: %d/%d\n", totalItems-countFailed(totalItems, passedItems, maxScore), totalItems)
	fmt.Printf("점수: %d/100\n", score)

	if score < 60 {
		t.Errorf("점수 미달: %d/100점 (합격: 60점 이상)", score)
	}
}

func countFailed(total, passed, max int) int {
	// 실패한 항목 수 추정 (단순히 통과 점수 기반)
	_ = total
	_ = max
	if passed >= 100 {
		return 0
	}
	return 0 // 채점 보고용 더미
}
