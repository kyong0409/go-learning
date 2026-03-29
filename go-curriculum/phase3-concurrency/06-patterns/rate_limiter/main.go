// 패키지 선언
package main

// 동시성 패턴: 속도 제한 (Rate Limiter)
//
// 속도 제한은 단위 시간당 처리할 수 있는 요청 수를 제한합니다.
// - API 호출 제한 (외부 API 쿼터 초과 방지)
// - 서버 과부하 방지
// - 공정한 리소스 배분
//
// Go에서의 구현 방법:
// 1. time.Ticker 기반 단순 제한
// 2. 토큰 버킷 (Token Bucket) 알고리즘
// 3. 슬라이딩 윈도우

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 1. 기본 속도 제한 (time.Ticker)
// ─────────────────────────────────────────

func basicRateLimiter() {
	fmt.Println("\n--- 1. 기본 속도 제한 (time.Ticker) ---")
	fmt.Println("  초당 3개 요청 제한:")

	// 초당 3개 = 333ms마다 1개 처리
	limiter := time.NewTicker(333 * time.Millisecond)
	defer limiter.Stop()

	requests := []int{1, 2, 3, 4, 5}

	start := time.Now()
	for _, req := range requests {
		<-limiter.C // 다음 tick까지 대기 (속도 제한)
		fmt.Printf("  요청 #%d 처리 (%.0fms 경과)\n",
			req, float64(time.Since(start).Milliseconds()))
	}
}

// ─────────────────────────────────────────
// 2. 버스트 허용 속도 제한
// ─────────────────────────────────────────

func burstRateLimiter() {
	fmt.Println("\n--- 2. 버스트 허용 속도 제한 ---")
	fmt.Println("  일반: 초당 1개, 버스트: 최대 3개 동시 허용")

	// 버퍼 채널로 토큰 버킷 구현
	// 버퍼 크기 = 버스트 허용량
	burstyLimiter := make(chan time.Time, 3)

	// 버스트용 초기 토큰 3개 채움
	for i := 0; i < 3; i++ {
		burstyLimiter <- time.Now()
	}

	// 이후 초당 1개씩 토큰 추가
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for t := range ticker.C {
			burstyLimiter <- t
		}
	}()

	requests := make([]int, 7)
	for i := range requests {
		requests[i] = i + 1
	}

	start := time.Now()
	for _, req := range requests {
		<-burstyLimiter // 토큰 획득
		fmt.Printf("  요청 #%d 처리 (%.0fms 경과)\n",
			req, float64(time.Since(start).Milliseconds()))
	}
}

// ─────────────────────────────────────────
// 3. 토큰 버킷 (Token Bucket) 구현
// ─────────────────────────────────────────

// TokenBucket은 토큰 버킷 알고리즘 구현입니다.
// - 버킷에 토큰이 있으면 즉시 처리 (버스트 허용)
// - 버킷이 비면 토큰 생성까지 대기
// - 토큰은 일정 속도로 보충됨
type TokenBucket struct {
	tokens   float64       // 현재 토큰 수 (소수점 허용)
	maxBurst float64       // 최대 버킷 크기
	rate     float64       // 초당 토큰 생성률
	lastTime time.Time     // 마지막 토큰 보충 시간
	mu       sync.Mutex
}

// NewTokenBucket은 새 토큰 버킷을 생성합니다.
func NewTokenBucket(ratePerSec float64, burst int) *TokenBucket {
	return &TokenBucket{
		tokens:   float64(burst),
		maxBurst: float64(burst),
		rate:     ratePerSec,
		lastTime: time.Now(),
	}
}

// Allow는 토큰 1개를 소비하고 허용 여부를 반환합니다.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()
	tb.lastTime = now

	// 경과 시간만큼 토큰 보충
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.maxBurst {
		tb.tokens = tb.maxBurst
	}

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

// Wait는 토큰이 생길 때까지 대기합니다.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if tb.Allow() {
			return nil
		}
		// 다음 토큰까지 대기 시간 계산
		tb.mu.Lock()
		waitTime := time.Duration((1.0-tb.tokens)/tb.rate*1000) * time.Millisecond
		tb.mu.Unlock()

		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func tokenBucketDemo() {
	fmt.Println("\n--- 3. 토큰 버킷 구현 ---")
	fmt.Println("  속도: 초당 5개, 버스트: 최대 3개")

	tb := NewTokenBucket(5, 3) // 초당 5개, 버스트 3개
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	for i := 1; i <= 10; i++ {
		if err := tb.Wait(ctx); err != nil {
			fmt.Printf("  요청 #%d 취소: %v\n", i, err)
			break
		}
		fmt.Printf("  요청 #%d 처리 (%.0fms 경과)\n",
			i, float64(time.Since(start).Milliseconds()))
	}
}

// ─────────────────────────────────────────
// 4. 동시성 제한 (Semaphore로 최대 동시 요청 수 제한)
// ─────────────────────────────────────────

// ConcurrentLimiter는 동시 실행 수를 제한합니다.
type ConcurrentLimiter struct {
	sem chan struct{}
}

// NewConcurrentLimiter는 동시성 제한기를 생성합니다.
func NewConcurrentLimiter(maxConcurrent int) *ConcurrentLimiter {
	return &ConcurrentLimiter{
		sem: make(chan struct{}, maxConcurrent),
	}
}

// Acquire는 슬롯을 획득합니다 (대기 가능).
func (l *ConcurrentLimiter) Acquire(ctx context.Context) error {
	select {
	case l.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release는 슬롯을 반환합니다.
func (l *ConcurrentLimiter) Release() {
	<-l.sem
}

// simulateAPICall은 API 호출을 시뮬레이션합니다.
func simulateAPICall(ctx context.Context, id int) error {
	select {
	case <-time.After(100 * time.Millisecond):
		fmt.Printf("  API 호출 #%d 완료\n", id)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func concurrentLimiterDemo() {
	fmt.Println("\n--- 4. 동시성 제한 (최대 3개 동시 요청) ---")

	limiter := NewConcurrentLimiter(3) // 최대 3개 동시 실행
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	start := time.Now()

	// 10개 요청을 최대 3개씩 동시 처리
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			if err := limiter.Acquire(ctx); err != nil {
				fmt.Printf("  요청 #%d 취소: %v\n", id, err)
				return
			}
			defer limiter.Release()

			fmt.Printf("  요청 #%d 시작 (%.0fms 경과, 현재 동시 실행: %d)\n",
				id, float64(time.Since(start).Milliseconds()), len(limiter.sem))
			simulateAPICall(ctx, id)
		}(i)
	}

	wg.Wait()
	fmt.Printf("  전체 소요 시간: %v\n", time.Since(start).Round(time.Millisecond))
}

// ─────────────────────────────────────────
// 5. 슬라이딩 윈도우 속도 제한
// ─────────────────────────────────────────

// SlidingWindowLimiter는 슬라이딩 윈도우 방식의 속도 제한기입니다.
// 최근 N초 내에 처리된 요청 수를 추적합니다.
type SlidingWindowLimiter struct {
	mu       sync.Mutex
	window   time.Duration // 윈도우 크기
	maxReqs  int           // 윈도우 내 최대 요청 수
	requests []time.Time   // 요청 타임스탬프
}

// NewSlidingWindowLimiter는 슬라이딩 윈도우 제한기를 생성합니다.
func NewSlidingWindowLimiter(window time.Duration, maxReqs int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		window:  window,
		maxReqs: maxReqs,
	}
}

// Allow는 현재 요청을 허용할지 결정합니다.
func (l *SlidingWindowLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	// 윈도우 밖의 오래된 요청 제거
	valid := l.requests[:0]
	for _, t := range l.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	l.requests = valid

	if len(l.requests) >= l.maxReqs {
		return false
	}

	l.requests = append(l.requests, now)
	return true
}

func slidingWindowDemo() {
	fmt.Println("\n--- 5. 슬라이딩 윈도우 속도 제한 ---")
	fmt.Println("  윈도우: 1초, 최대 요청: 3개")

	limiter := NewSlidingWindowLimiter(time.Second, 3)
	start := time.Now()

	// 빠르게 6개 요청 시도
	for i := 1; i <= 6; i++ {
		allowed := limiter.Allow()
		status := "허용"
		if !allowed {
			status = "거부"
		}
		fmt.Printf("  요청 #%d: %s (%.0fms 경과)\n",
			i, status, float64(time.Since(start).Milliseconds()))
		time.Sleep(200 * time.Millisecond)
	}

	// 1초 후 재시도
	time.Sleep(800 * time.Millisecond)
	fmt.Println("  1초 경과 후 재시도:")
	for i := 7; i <= 9; i++ {
		allowed := limiter.Allow()
		status := "허용"
		if !allowed {
			status = "거부"
		}
		fmt.Printf("  요청 #%d: %s\n", i, status)
	}
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== 동시성 패턴: 속도 제한 (Rate Limiter) ===")

	basicRateLimiter()
	burstRateLimiter()
	tokenBucketDemo()
	concurrentLimiterDemo()
	slidingWindowDemo()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
