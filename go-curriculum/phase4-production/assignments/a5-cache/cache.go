// a5-cache/cache.go
// TTL과 LRU 퇴거를 가진 동시성 안전 캐시 과제입니다.
// TODO 주석이 있는 모든 함수/메서드를 구현하세요.
package cache

import (
	"time"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// EvictionReason은 항목이 캐시에서 제거되는 이유입니다.
type EvictionReason string

const (
	EvictionReasonExpired  EvictionReason = "expired"  // TTL 만료
	EvictionReasonCapacity EvictionReason = "capacity" // 최대 크기 초과 (LRU)
	EvictionReasonDeleted  EvictionReason = "deleted"  // 명시적 삭제
	EvictionReasonFlushed  EvictionReason = "flushed"  // Flush 호출
)

// Stats는 캐시 사용 통계입니다.
type Stats struct {
	Hits        uint64 // Get 히트 횟수
	Misses      uint64 // Get 미스 횟수
	Evictions   uint64 // LRU 퇴거 횟수
	Expirations uint64 // TTL 만료 횟수
}

// Options는 캐시 생성 옵션입니다.
type Options[K comparable, V any] struct {
	MaxSize    int                                    // 최대 항목 수 (0 = 무제한)
	DefaultTTL time.Duration                          // 기본 TTL (0 = 만료 없음)
	OnEviction func(key K, value V, reason EvictionReason) // 퇴거 콜백 (nil 가능)
	// 백그라운드 정리 주기 (0이면 기본값 1분 사용)
	CleanupInterval time.Duration
}

// ============================================================
// Cache
// ============================================================

// Cache는 TTL과 LRU 퇴거를 지원하는 제네릭 인메모리 캐시입니다.
// 고루틴 안전해야 합니다.
type Cache[K comparable, V any] struct {
	// TODO: 필드를 추가하세요.
	// 힌트: container/list를 사용한 LRU 관리, map으로 O(1) 조회
}

// New는 새 Cache를 생성합니다.
// TODO: 구현하세요. 백그라운드 정리 고루틴을 시작하세요.
func New[K comparable, V any](opts Options[K, V]) *Cache[K, V] {
	panic("New: 아직 구현되지 않았습니다")
}

// Set은 키-값 쌍을 기본 TTL로 저장합니다.
// 키가 이미 존재하면 값과 TTL을 갱신하고 LRU 순서를 최신으로 이동합니다.
// TODO: 구현하세요.
func (c *Cache[K, V]) Set(key K, value V) {
}

// SetWithTTL은 키-값 쌍을 지정된 TTL로 저장합니다.
// ttl이 0이면 만료되지 않습니다.
// TODO: 구현하세요.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
}

// Get은 키에 해당하는 값을 반환합니다.
// 만료된 항목은 삭제 후 false를 반환합니다.
// 히트 시 LRU 순서를 최신으로 갱신합니다.
// TODO: 구현하세요.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zero V
	return zero, false
}

// Delete는 키를 명시적으로 삭제합니다.
// OnEviction 콜백이 설정된 경우 EvictionReasonDeleted로 호출합니다.
// TODO: 구현하세요.
func (c *Cache[K, V]) Delete(key K) {
}

// Len은 현재 캐시에 저장된 항목 수를 반환합니다.
// 만료된 항목은 포함하지 않습니다 (지연 만료이므로 근사값일 수 있음).
// TODO: 구현하세요.
func (c *Cache[K, V]) Len() int {
	return 0
}

// Stats는 현재 캐시 통계를 반환합니다.
// TODO: 구현하세요.
func (c *Cache[K, V]) Stats() Stats {
	return Stats{}
}

// Flush는 모든 항목을 삭제합니다.
// OnEviction 콜백이 설정된 경우 각 항목에 대해 EvictionReasonFlushed로 호출합니다.
// TODO: 구현하세요.
func (c *Cache[K, V]) Flush() {
}

// Close는 백그라운드 정리 고루틴을 중지합니다.
// TODO: 구현하세요.
func (c *Cache[K, V]) Close() {
}
