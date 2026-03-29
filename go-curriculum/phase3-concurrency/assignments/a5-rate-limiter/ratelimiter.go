// 패키지 선언
package ratelimiter

import "context"

// TokenBucket은 토큰 버킷 알고리즘 기반의 속도 제한기입니다.
// 고루틴 안전(goroutine-safe)해야 합니다.
type TokenBucket struct {
	// TODO: 필요한 필드를 추가하세요
	// 힌트:
	// - mu sync.Mutex: 동시 접근 보호
	// - tokens float64: 현재 토큰 수
	// - rate float64: 초당 토큰 보충 속도
	// - burst int: 최대 토큰 수 (버스트 크기)
	// - lastRefill time.Time: 마지막 토큰 보충 시각
}

// NewTokenBucket은 새 토큰 버킷을 생성합니다.
// 초기 토큰 수는 burst와 동일합니다 (가득 찬 상태로 시작).
//
// 매개변수:
//   - rate: 초당 토큰 보충 속도 (예: 10.0 = 초당 10개)
//   - burst: 버킷 최대 용량 (최대 버스트 크기)
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	// TODO: 구현하세요
	// - 필드 초기화
	// - tokens = float64(burst) (가득 찬 상태로 시작)
	// - lastRefill = time.Now()
	panic("구현 필요")
}

// refill은 마지막 보충 이후 경과 시간에 비례해 토큰을 추가합니다.
// 이 함수는 mu를 이미 잠근 상태에서 호출되어야 합니다 (내부 함수).
func (tb *TokenBucket) refill() {
	// TODO: 구현하세요
	// 1. now := time.Now()
	// 2. elapsed := now.Sub(tb.lastRefill).Seconds()
	// 3. tb.tokens += elapsed * tb.rate
	// 4. if tb.tokens > float64(tb.burst) { tb.tokens = float64(tb.burst) }
	// 5. tb.lastRefill = now
	panic("구현 필요")
}

// Allow는 토큰 1개를 소비를 시도합니다.
// 토큰이 있으면 소비하고 true, 없으면 false를 반환합니다 (비블로킹).
func (tb *TokenBucket) Allow() bool {
	// TODO: 구현하세요
	// 1. mu.Lock() / defer mu.Unlock()
	// 2. refill() 로 토큰 보충
	// 3. tokens >= 1 이면 tokens-- 하고 true 반환
	// 4. 아니면 false 반환
	panic("구현 필요")
}

// AllowN은 토큰 n개 소비를 시도합니다.
// n개 이상의 토큰이 있으면 소비하고 true, 없으면 false를 반환합니다 (비블로킹).
func (tb *TokenBucket) AllowN(n int) bool {
	// TODO: 구현하세요
	// Allow()와 동일하지만 n개를 소비
	panic("구현 필요")
}

// Wait는 토큰 1개가 생길 때까지 블로킹 대기합니다.
// ctx가 취소되면 ctx.Err()를 반환합니다.
// 토큰을 성공적으로 소비하면 nil을 반환합니다.
func (tb *TokenBucket) Wait(ctx context.Context) error {
	// TODO: 구현하세요
	// 1. Allow()로 즉시 소비 가능한지 먼저 확인
	// 2. 불가능하면 대기 시간 계산:
	//    mu.Lock() 후 needed := 1.0 - tb.tokens, waitDur := time.Duration(needed/tb.rate * float64(time.Second)) 계산, mu.Unlock()
	// 3. select { case <-time.After(waitDur): case <-ctx.Done(): return ctx.Err() }
	// 4. 대기 후 Allow() 재시도 (루프)
	panic("구현 필요")
}

// SetRate는 속도와 버스트 크기를 동적으로 변경합니다.
// 현재 토큰 수는 새 버스트 크기를 초과하지 않도록 조정됩니다.
func (tb *TokenBucket) SetRate(rate float64, burst int) {
	// TODO: 구현하세요
	// 1. mu.Lock() / defer mu.Unlock()
	// 2. refill() 로 현재까지 토큰 보충
	// 3. tb.rate = rate, tb.burst = burst
	// 4. if tb.tokens > float64(burst) { tb.tokens = float64(burst) }
	panic("구현 필요")
}

// KeyedRateLimiter는 키(예: IP 주소)별로 독립적인 속도 제한을 적용합니다.
// 고루틴 안전해야 합니다.
type KeyedRateLimiter struct {
	// TODO: 필요한 필드를 추가하세요
	// 힌트:
	// - mu sync.Mutex: 맵 접근 보호
	// - buckets map[string]*TokenBucket: 키별 버킷
	// - rate float64: 각 버킷의 속도
	// - burst int: 각 버킷의 버스트 크기
}

// NewKeyedRateLimiter는 새 키별 속도 제한기를 생성합니다.
func NewKeyedRateLimiter(rate float64, burst int) *KeyedRateLimiter {
	// TODO: 구현하세요
	panic("구현 필요")
}

// getBucket은 키에 해당하는 버킷을 반환합니다.
// 없으면 새로 생성합니다 (내부 함수).
func (krl *KeyedRateLimiter) getBucket(key string) *TokenBucket {
	// TODO: 구현하세요
	// 1. mu.Lock() / defer mu.Unlock()
	// 2. buckets[key]가 없으면 NewTokenBucket(rate, burst) 으로 생성 후 저장
	// 3. 해당 버킷 반환
	panic("구현 필요")
}

// Allow는 지정한 키에 대해 토큰 1개 소비를 시도합니다.
func (krl *KeyedRateLimiter) Allow(key string) bool {
	// TODO: 구현하세요
	// getBucket(key).Allow() 반환
	panic("구현 필요")
}

// Wait는 지정한 키에 대해 토큰이 생길 때까지 대기합니다.
func (krl *KeyedRateLimiter) Wait(ctx context.Context, key string) error {
	// TODO: 구현하세요
	// getBucket(key).Wait(ctx) 반환
	panic("구현 필요")
}
