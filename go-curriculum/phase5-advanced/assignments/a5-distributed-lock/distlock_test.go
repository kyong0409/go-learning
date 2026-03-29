// a5-distributed-lock/distlock_test.go
// 분산 잠금 시스템 테스트 및 채점
//
// 실행:
//
//	go test ./... -v
//	go test -race ./...
//	go test ./... -v -run TestGrade
package distlock_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/learn-go/a5-distributed-lock"
)

// ============================================================
// Acquire / Release 기본 테스트 (20점)
// ============================================================

func TestAcquire_Basic(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	lock, err := svc.Acquire(ctx, "resource-1", "ownerA", time.Second)
	if err != nil {
		t.Fatalf("Acquire: 예상치 못한 오류: %v", err)
	}
	if lock == nil {
		t.Fatal("Acquire: nil Lock 반환")
	}
	if lock.Key != "resource-1" {
		t.Errorf("Lock.Key: 기대 %q, 실제 %q", "resource-1", lock.Key)
	}
	if lock.Owner != "ownerA" {
		t.Errorf("Lock.Owner: 기대 %q, 실제 %q", "ownerA", lock.Owner)
	}
}

func TestAcquire_IsLocked(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	svc.Acquire(ctx, "key1", "owner1", time.Second)

	if !svc.IsLocked("key1") {
		t.Error("잠금 획득 후 IsLocked: true여야 합니다")
	}
	if svc.IsLocked("key2") {
		t.Error("잠금 없는 키 IsLocked: false여야 합니다")
	}
}

func TestRelease_Basic(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	lock, _ := svc.Acquire(ctx, "res", "owner", time.Second)

	if err := svc.Release(lock); err != nil {
		t.Fatalf("Release: 오류: %v", err)
	}
	if svc.IsLocked("res") {
		t.Error("Release 후 IsLocked: false여야 합니다")
	}
}

func TestRelease_NotFound(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	fakeLock := &distlock.Lock{Key: "ghost", Owner: "nobody", Token: 1}
	err := svc.Release(fakeLock)
	if err == nil {
		t.Error("존재하지 않는 잠금 Release: 오류가 반환되어야 합니다")
	}
}

func TestGetLock(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	svc.Acquire(ctx, "mykey", "me", time.Second)

	lock, ok := svc.GetLock("mykey")
	if !ok || lock == nil {
		t.Fatal("GetLock: 잠금을 찾을 수 없습니다")
	}
	if lock.Owner != "me" {
		t.Errorf("GetLock Owner: 기대 %q, 실제 %q", "me", lock.Owner)
	}

	if _, ok := svc.GetLock("nokey"); ok {
		t.Error("없는 키 GetLock: false여야 합니다")
	}
}

// ============================================================
// 펜싱 토큰 테스트 (20점)
// ============================================================

func TestFencingToken_MonotonicallyIncreasing(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	lock1, _ := svc.Acquire(ctx, "k1", "owner1", time.Second)
	lock2, _ := svc.Acquire(ctx, "k2", "owner2", time.Second)
	lock3, _ := svc.Acquire(ctx, "k3", "owner3", time.Second)

	if lock1.Token >= lock2.Token {
		t.Errorf("토큰 단조 증가: lock1.Token(%d) >= lock2.Token(%d)", lock1.Token, lock2.Token)
	}
	if lock2.Token >= lock3.Token {
		t.Errorf("토큰 단조 증가: lock2.Token(%d) >= lock3.Token(%d)", lock2.Token, lock3.Token)
	}
}

func TestFencingToken_NewTokenOnReacquire(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	lock1, _ := svc.Acquire(ctx, "shared", "owner1", 30*time.Millisecond)

	// 만료 후 재획득
	time.Sleep(60 * time.Millisecond)
	lock2, _ := svc.Acquire(ctx, "shared", "owner2", time.Second)

	if lock2 == nil {
		t.Fatal("만료 후 재획득: nil Lock 반환")
	}
	if lock2.Token <= lock1.Token {
		t.Errorf("재획득 시 새 토큰: 기대 > %d, 실제 %d", lock1.Token, lock2.Token)
	}
}

func TestFencingToken_InvalidTokenOnRelease(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	lock1, _ := svc.Acquire(ctx, "protected", "owner1", 30*time.Millisecond)

	// 만료 후 다른 소유자가 획득
	time.Sleep(60 * time.Millisecond)
	svc.Acquire(ctx, "protected", "owner2", time.Second)

	// 이전 토큰으로 Release 시도 -> 실패해야 함
	err := svc.Release(lock1)
	if err == nil {
		t.Error("오래된 토큰으로 Release: 오류가 반환되어야 합니다")
	}
}

func TestLock_IsValid(t *testing.T) {
	now := time.Now()
	valid := &distlock.Lock{
		Key:       "k",
		Owner:     "o",
		Token:     1,
		ExpiresAt: now.Add(time.Second),
	}
	expired := &distlock.Lock{
		Key:       "k",
		Owner:     "o",
		Token:     2,
		ExpiresAt: now.Add(-time.Second),
	}

	if !valid.IsValid() {
		t.Error("유효한 잠금 IsValid: true여야 합니다")
	}
	if expired.IsValid() {
		t.Error("만료된 잠금 IsValid: false여야 합니다")
	}
}

// ============================================================
// TTL 만료 테스트 (20점)
// ============================================================

func TestTTL_Expiry(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	svc.Acquire(ctx, "ttlkey", "owner", 40*time.Millisecond)

	if !svc.IsLocked("ttlkey") {
		t.Fatal("획득 직후 IsLocked: true여야 합니다")
	}

	time.Sleep(80 * time.Millisecond)

	if svc.IsLocked("ttlkey") {
		t.Error("TTL 만료 후 IsLocked: false여야 합니다")
	}
}

func TestTTL_ExpiredLockAllowsReacquire(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	svc.Acquire(ctx, "res", "owner1", 30*time.Millisecond)
	time.Sleep(60 * time.Millisecond)

	// 만료 후 다른 소유자가 즉시 획득 가능해야 함
	lock, err := svc.Acquire(ctx, "res", "owner2", time.Second)
	if err != nil {
		t.Fatalf("만료 후 재획득: 오류: %v", err)
	}
	if lock.Owner != "owner2" {
		t.Errorf("재획득 Owner: 기대 %q, 실제 %q", "owner2", lock.Owner)
	}
}

func TestRefresh_Basic(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	lock, _ := svc.Acquire(ctx, "refresh-key", "owner", 100*time.Millisecond)

	// 갱신 전 만료 시각 기록
	origExpiry := lock.ExpiresAt

	time.Sleep(30 * time.Millisecond)
	newLock, err := svc.Refresh(lock, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("Refresh: 오류: %v", err)
	}
	if !newLock.ExpiresAt.After(origExpiry) {
		t.Error("Refresh 후 ExpiresAt이 연장되어야 합니다")
	}
}

func TestRefresh_ExpiredReturnsError(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	lock, _ := svc.Acquire(ctx, "expkey", "owner", 30*time.Millisecond)
	time.Sleep(60 * time.Millisecond)

	_, err := svc.Refresh(lock, time.Second)
	if err == nil {
		t.Error("만료된 잠금 Refresh: 오류가 반환되어야 합니다")
	}
}

// ============================================================
// 대기 큐 테스트 (25점)
// ============================================================

func TestAcquire_BlocksWhenLocked(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	lock1, _ := svc.Acquire(ctx, "blocked-res", "owner1", time.Second)

	acquired := make(chan struct{})
	go func() {
		svc.Acquire(ctx, "blocked-res", "owner2", time.Second)
		close(acquired)
	}()

	// 짧은 대기 후 아직 획득 안 됐어야 함
	select {
	case <-acquired:
		t.Error("owner1이 보유 중인데 owner2가 즉시 획득됨")
	case <-time.After(50 * time.Millisecond):
		// 정상: 대기 중
	}

	svc.Release(lock1)

	// Release 후 획득되어야 함
	select {
	case <-acquired:
		// 정상
	case <-time.After(500 * time.Millisecond):
		t.Error("Release 후 대기 중인 Acquire가 완료되지 않음")
	}
}

func TestAcquire_ContextCancellation(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	svc.Acquire(ctx, "cancel-res", "holder", time.Hour)

	cancelCtx, cancel := context.WithTimeout(ctx, 80*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := svc.Acquire(cancelCtx, "cancel-res", "waiter", time.Second)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("컨텍스트 취소 후 Acquire: 오류가 반환되어야 합니다")
	}
	if elapsed > 300*time.Millisecond {
		t.Errorf("컨텍스트 취소 후 너무 오래 대기: %v", elapsed)
	}
}

func TestAcquire_TTLExpiry_Unblocks(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	svc.Acquire(ctx, "expblock", "holder", 60*time.Millisecond)

	start := time.Now()
	lock, err := svc.Acquire(ctx, "expblock", "waiter", time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("TTL 만료 후 Acquire: 오류: %v", err)
	}
	if lock.Owner != "waiter" {
		t.Errorf("획득 소유자: 기대 %q, 실제 %q", "waiter", lock.Owner)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("TTL 만료 후 Acquire 너무 오래 걸림: %v", elapsed)
	}
}

func TestAcquire_MultipleWaiters(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	holder, _ := svc.Acquire(ctx, "multi", "holder", time.Second)

	var acquired int64
	var wg sync.WaitGroup

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lock, err := svc.Acquire(ctx, "multi", fmt.Sprintf("waiter-%d", atomic.AddInt64(&acquired, 1)), 500*time.Millisecond)
			if err == nil && lock != nil {
				svc.Release(lock)
			}
		}()
	}

	// 홀더 해제
	time.Sleep(20 * time.Millisecond)
	svc.Release(holder)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 정상
	case <-time.After(3 * time.Second):
		t.Error("다수 대기자: 모든 Acquire가 완료되지 않음")
	}
}

// ============================================================
// 데드락 감지 테스트 (15점)
// ============================================================

func TestDetectDeadlock_NoDeadlock(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()
	svc.Acquire(ctx, "k1", "ownerA", time.Second)
	svc.Acquire(ctx, "k2", "ownerB", time.Second)

	cycle, err := svc.DetectDeadlock()
	if err != nil {
		t.Fatalf("DetectDeadlock: 오류: %v", err)
	}
	if cycle != nil {
		t.Errorf("데드락 없음인데 사이클 반환: %v", cycle)
	}
}

func TestDetectDeadlock_TwoNodeCycle(t *testing.T) {
	svc := distlock.NewLockService()
	defer svc.Close()

	ctx := context.Background()

	// ownerA가 key1 보유
	svc.Acquire(ctx, "key1", "ownerA", time.Hour)
	// ownerB가 key2 보유
	svc.Acquire(ctx, "key2", "ownerB", time.Hour)

	// ownerA가 key2 대기 (고루틴에서)
	ctxA, cancelA := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancelA()
	go func() {
		svc.Acquire(ctxA, "key2", "ownerA", time.Second)
	}()

	// ownerB가 key1 대기 (고루틴에서)
	ctxB, cancelB := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancelB()
	go func() {
		svc.Acquire(ctxB, "key1", "ownerB", time.Second)
	}()

	// 대기 상태 수립 기다리기
	time.Sleep(50 * time.Millisecond)

	cycle, err := svc.DetectDeadlock()
	if err != nil {
		t.Fatalf("DetectDeadlock: 오류: %v", err)
	}
	if cycle == nil {
		t.Error("데드락 상황에서 사이클이 감지되어야 합니다")
	} else {
		t.Logf("감지된 사이클: %v", cycle)
	}
}

// ============================================================
// 채점 함수
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	type result struct {
		name  string
		pts   int
		maxPt int
		pass  bool
	}
	var results []result

	check := func(name string, maxPt int, fn func() bool) {
		pass := func() (ok bool) {
			defer func() {
				if r := recover(); r != nil {
					ok = false
				}
			}()
			return fn()
		}()
		pts := 0
		if pass {
			pts = maxPt
			score += maxPt
		}
		results = append(results, result{name, pts, maxPt, pass})
	}

	// Acquire / Release 기본 (20점)
	check("Acquire 기본 동작", 8, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		lock, err := svc.Acquire(context.Background(), "k", "owner", time.Second)
		return err == nil && lock != nil && lock.Key == "k" && lock.Owner == "owner"
	})
	check("IsLocked / GetLock", 6, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		svc.Acquire(context.Background(), "k", "o", time.Second)
		lock, ok := svc.GetLock("k")
		return svc.IsLocked("k") && ok && lock != nil
	})
	check("Release 기본 동작", 6, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		lock, _ := svc.Acquire(ctx, "k", "o", time.Second)
		err := svc.Release(lock)
		return err == nil && !svc.IsLocked("k")
	})

	// 펜싱 토큰 (20점)
	check("토큰 단조 증가", 8, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		l1, _ := svc.Acquire(ctx, "k1", "o", time.Second)
		l2, _ := svc.Acquire(ctx, "k2", "o", time.Second)
		return l1 != nil && l2 != nil && l1.Token < l2.Token
	})
	check("오래된 토큰 Release 실패", 7, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		l1, _ := svc.Acquire(ctx, "k", "o1", 30*time.Millisecond)
		time.Sleep(60 * time.Millisecond)
		svc.Acquire(ctx, "k", "o2", time.Second)
		return svc.Release(l1) != nil
	})
	check("IsValid 만료 판단", 5, func() bool {
		valid := &distlock.Lock{ExpiresAt: time.Now().Add(time.Second)}
		expired := &distlock.Lock{ExpiresAt: time.Now().Add(-time.Second)}
		return valid.IsValid() && !expired.IsValid()
	})

	// TTL 만료 (20점)
	check("TTL 자동 만료", 10, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		svc.Acquire(context.Background(), "k", "o", 40*time.Millisecond)
		time.Sleep(80 * time.Millisecond)
		return !svc.IsLocked("k")
	})
	check("만료 후 재획득", 6, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		svc.Acquire(ctx, "k", "o1", 30*time.Millisecond)
		time.Sleep(60 * time.Millisecond)
		lock, err := svc.Acquire(ctx, "k", "o2", time.Second)
		return err == nil && lock != nil && lock.Owner == "o2"
	})
	check("Refresh TTL 연장", 4, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		lock, _ := svc.Acquire(ctx, "k", "o", 200*time.Millisecond)
		orig := lock.ExpiresAt
		newLock, err := svc.Refresh(lock, time.Second)
		return err == nil && newLock != nil && newLock.ExpiresAt.After(orig)
	})

	// 대기 큐 (25점)
	check("잠금 중 Acquire 블로킹", 10, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		lock1, _ := svc.Acquire(ctx, "r", "o1", time.Second)
		acquired := make(chan bool, 1)
		go func() {
			l, err := svc.Acquire(ctx, "r", "o2", time.Second)
			acquired <- (err == nil && l != nil)
		}()
		select {
		case <-acquired:
			return false // 즉시 획득: 잘못된 동작
		case <-time.After(30 * time.Millisecond):
		}
		svc.Release(lock1)
		select {
		case ok := <-acquired:
			return ok
		case <-time.After(500 * time.Millisecond):
			return false
		}
	})
	check("컨텍스트 취소 대기 중단", 10, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		svc.Acquire(ctx, "r", "holder", time.Hour)
		cancelCtx, cancel := context.WithTimeout(ctx, 60*time.Millisecond)
		defer cancel()
		start := time.Now()
		_, err := svc.Acquire(cancelCtx, "r", "waiter", time.Second)
		return err != nil && time.Since(start) < 300*time.Millisecond
	})
	check("TTL 만료로 대기 해제", 5, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		svc.Acquire(ctx, "r", "holder", 50*time.Millisecond)
		start := time.Now()
		lock, err := svc.Acquire(ctx, "r", "waiter", time.Second)
		return err == nil && lock != nil && time.Since(start) < 400*time.Millisecond
	})

	// 데드락 감지 (15점)
	check("데드락 없음 감지", 5, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		svc.Acquire(ctx, "k1", "oA", time.Second)
		svc.Acquire(ctx, "k2", "oB", time.Second)
		cycle, err := svc.DetectDeadlock()
		return err == nil && cycle == nil
	})
	check("2-노드 데드락 감지", 10, func() bool {
		svc := distlock.NewLockService()
		defer svc.Close()
		ctx := context.Background()
		svc.Acquire(ctx, "k1", "oA", time.Hour)
		svc.Acquire(ctx, "k2", "oB", time.Hour)
		ctxA, cancelA := context.WithTimeout(ctx, 400*time.Millisecond)
		defer cancelA()
		ctxB, cancelB := context.WithTimeout(ctx, 400*time.Millisecond)
		defer cancelB()
		go func() { svc.Acquire(ctxA, "k2", "oA", time.Second) }()
		go func() { svc.Acquire(ctxB, "k1", "oB", time.Second) }()
		time.Sleep(60 * time.Millisecond)
		cycle, err := svc.DetectDeadlock()
		return err == nil && cycle != nil && len(cycle) >= 2
	})

	// 결과 출력
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════╗")
	fmt.Println("║   과제 A5: 분산 잠금 시스템 구현 채점 결과       ║")
	fmt.Println("╠═══════════════════════════════════════════════════╣")
	passed := 0
	for _, r := range results {
		mark := "✗"
		if r.pass {
			mark = "✓"
			passed++
		}
		fmt.Printf("║  %s %-36s %3d/%d점  ║\n", mark, r.name, r.pts, r.maxPt)
	}
	fmt.Println("╠═══════════════════════════════════════════════════╣")
	fmt.Printf("║  통과: %d/%d                                         ║\n", passed, len(results))
	fmt.Printf("║  점수: %d/%d                                        ║\n", score, total)
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
	fmt.Printf("║  등급: %s                                           ║\n", grade)
	fmt.Println("╚═══════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("=== 채점 결과 ===")
	fmt.Printf("통과: %d/%d\n", passed, len(results))
	fmt.Printf("점수: %d/%d\n", score, total)
}
