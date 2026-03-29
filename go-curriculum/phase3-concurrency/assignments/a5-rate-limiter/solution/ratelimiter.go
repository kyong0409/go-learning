// 패키지 선언
// 참고 솔루션 - 풀기 전에 보지 마세요!
package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// TokenBucket은 토큰 버킷 알고리즘 기반의 속도 제한기입니다.
type TokenBucket struct {
	mu         sync.Mutex
	tokens     float64   // 현재 토큰 수
	rate       float64   // 초당 토큰 보충 속도
	burst      int       // 최대 토큰 수
	lastRefill time.Time // 마지막 보충 시각
}

// NewTokenBucket은 새 토큰 버킷을 생성합니다.
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		tokens:     float64(burst), // 가득 찬 상태로 시작
		rate:       rate,
		burst:      burst,
		lastRefill: time.Now(),
	}
}

// refill은 경과 시간에 비례해 토큰을 보충합니다.
// mu를 잠근 상태에서 호출해야 합니다.
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}
	tb.lastRefill = now
}

// Allow는 토큰 1개 소비를 시도합니다.
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN은 토큰 n개 소비를 시도합니다.
func (tb *TokenBucket) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		return true
	}
	return false
}

// Wait는 토큰 1개가 생길 때까지 블로킹 대기합니다.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	// 즉시 소비 가능한지 먼저 확인
	if tb.Allow() {
		return nil
	}

	for {
		// ctx 이미 취소됐는지 확인
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 토큰 1개를 얻기까지 필요한 대기 시간 계산
		tb.mu.Lock()
		tb.refill()
		needed := 1.0 - tb.tokens
		var waitDur time.Duration
		if needed > 0 {
			waitDur = time.Duration(needed / tb.rate * float64(time.Second))
		}
		tb.mu.Unlock()

		if waitDur <= 0 {
			// 이미 토큰이 보충됐을 수 있음
			if tb.Allow() {
				return nil
			}
			waitDur = time.Millisecond // 최소 대기
		}

		select {
		case <-time.After(waitDur):
			if tb.Allow() {
				return nil
			}
			// 아직 부족하면 루프 재시도
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// SetRate는 속도와 버스트 크기를 동적으로 변경합니다.
func (tb *TokenBucket) SetRate(rate float64, burst int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill() // 변경 전 현재까지 보충

	tb.rate = rate
	tb.burst = burst
	if tb.tokens > float64(burst) {
		tb.tokens = float64(burst)
	}
}

// KeyedRateLimiter는 키별로 독립적인 속도 제한을 적용합니다.
type KeyedRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*TokenBucket
	rate    float64
	burst   int
}

// NewKeyedRateLimiter는 새 키별 속도 제한기를 생성합니다.
func NewKeyedRateLimiter(rate float64, burst int) *KeyedRateLimiter {
	return &KeyedRateLimiter{
		buckets: make(map[string]*TokenBucket),
		rate:    rate,
		burst:   burst,
	}
}

// getBucket은 키에 해당하는 버킷을 반환하며, 없으면 생성합니다.
func (krl *KeyedRateLimiter) getBucket(key string) *TokenBucket {
	krl.mu.Lock()
	defer krl.mu.Unlock()

	if _, ok := krl.buckets[key]; !ok {
		krl.buckets[key] = NewTokenBucket(krl.rate, krl.burst)
	}
	return krl.buckets[key]
}

// Allow는 지정한 키에 대해 토큰 1개 소비를 시도합니다.
func (krl *KeyedRateLimiter) Allow(key string) bool {
	return krl.getBucket(key).Allow()
}

// Wait는 지정한 키에 대해 토큰이 생길 때까지 대기합니다.
func (krl *KeyedRateLimiter) Wait(ctx context.Context, key string) error {
	return krl.getBucket(key).Wait(ctx)
}
