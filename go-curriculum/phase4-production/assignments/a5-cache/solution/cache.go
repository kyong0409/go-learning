// a5-cache/solution/cache.go
// TTL과 LRU 퇴거를 가진 동시성 안전 캐시 참고 답안입니다.
package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 타입 정의
// ============================================================

type EvictionReason string

const (
	EvictionReasonExpired  EvictionReason = "expired"
	EvictionReasonCapacity EvictionReason = "capacity"
	EvictionReasonDeleted  EvictionReason = "deleted"
	EvictionReasonFlushed  EvictionReason = "flushed"
)

type Stats struct {
	Hits        uint64
	Misses      uint64
	Evictions   uint64
	Expirations uint64
}

type Options[K comparable, V any] struct {
	MaxSize         int
	DefaultTTL      time.Duration
	OnEviction      func(key K, value V, reason EvictionReason)
	CleanupInterval time.Duration
}

// ============================================================
// 내부 항목
// ============================================================

type entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time // 제로값이면 만료 없음
}

// ============================================================
// Cache
// ============================================================

type Cache[K comparable, V any] struct {
	mu         sync.RWMutex
	opts       Options[K, V]
	items      map[K]*list.Element // 키 -> LRU 리스트 원소
	lru        *list.List          // 앞쪽이 가장 최근, 뒤쪽이 가장 오래됨
	hits       atomic.Uint64
	misses     atomic.Uint64
	evictions  atomic.Uint64
	expirations atomic.Uint64
	stopCh     chan struct{}
}

func New[K comparable, V any](opts Options[K, V]) *Cache[K, V] {
	interval := opts.CleanupInterval
	if interval <= 0 {
		interval = time.Minute
	}

	c := &Cache[K, V]{
		opts:   opts,
		items:  make(map[K]*list.Element),
		lru:    list.New(),
		stopCh: make(chan struct{}),
	}

	// 백그라운드 정리 고루틴
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.deleteExpired()
			case <-c.stopCh:
				return
			}
		}
	}()

	return c
}

// expiredAt은 opts.DefaultTTL을 고려한 만료 시각을 반환합니다.
func (c *Cache[K, V]) expiryTime(ttl time.Duration) time.Time {
	if ttl == 0 {
		return time.Time{} // 만료 없음
	}
	return time.Now().Add(ttl)
}

func (c *Cache[K, V]) isExpired(e *entry[K, V]) bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.expiresAt)
}

// ── 쓰기 연산 ───────────────────────────────────────────────

func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.opts.DefaultTTL)
}

func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	e := &entry[K, V]{
		key:       key,
		value:     value,
		expiresAt: c.expiryTime(ttl),
	}

	c.mu.Lock()
	var evicted *entry[K, V]

	if elem, ok := c.items[key]; ok {
		// 기존 항목 갱신 + LRU 앞으로 이동
		elem.Value = e
		c.lru.MoveToFront(elem)
	} else {
		// 새 항목 추가
		elem := c.lru.PushFront(e)
		c.items[key] = elem

		// MaxSize 초과 시 가장 오래된 항목 퇴거
		if c.opts.MaxSize > 0 && c.lru.Len() > c.opts.MaxSize {
			evicted = c.evictOldest()
		}
	}

	c.mu.Unlock()

	// 콜백은 잠금 해제 후 동기적으로 호출
	if evicted != nil && c.opts.OnEviction != nil {
		c.opts.OnEviction(evicted.key, evicted.value, EvictionReasonCapacity)
	}
}

// evictOldest는 잠금이 보유된 상태에서 LRU 가장 오래된 항목을 제거하고 반환합니다.
func (c *Cache[K, V]) evictOldest() *entry[K, V] {
	back := c.lru.Back()
	if back == nil {
		return nil
	}
	e := back.Value.(*entry[K, V])
	c.lru.Remove(back)
	delete(c.items, e.key)
	c.evictions.Add(1)
	return e
}

// ── 읽기 연산 ───────────────────────────────────────────────

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		c.misses.Add(1)
		var zero V
		return zero, false
	}

	e := elem.Value.(*entry[K, V])

	// 지연 만료 확인
	if c.isExpired(e) {
		c.lru.Remove(elem)
		delete(c.items, key)
		c.expirations.Add(1)
		c.misses.Add(1)

		if c.opts.OnEviction != nil {
			key, val := e.key, e.value
			cb := c.opts.OnEviction
			go cb(key, val, EvictionReasonExpired)
		}

		var zero V
		return zero, false
	}

	// LRU 갱신
	c.lru.MoveToFront(elem)
	c.hits.Add(1)
	return e.value, true
}

// ── 삭제 연산 ───────────────────────────────────────────────

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	elem, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		return
	}
	e := elem.Value.(*entry[K, V])
	c.lru.Remove(elem)
	delete(c.items, key)
	c.mu.Unlock()

	if c.opts.OnEviction != nil {
		c.opts.OnEviction(e.key, e.value, EvictionReasonDeleted)
	}
}

func (c *Cache[K, V]) Flush() {
	c.mu.Lock()
	entries := make([]*entry[K, V], 0, len(c.items))
	for _, elem := range c.items {
		entries = append(entries, elem.Value.(*entry[K, V]))
	}
	c.items = make(map[K]*list.Element)
	c.lru.Init()
	c.mu.Unlock()

	if c.opts.OnEviction != nil {
		for _, e := range entries {
			c.opts.OnEviction(e.key, e.value, EvictionReasonFlushed)
		}
	}
}

// ── 조회 연산 ───────────────────────────────────────────────

func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *Cache[K, V]) Stats() Stats {
	return Stats{
		Hits:        c.hits.Load(),
		Misses:      c.misses.Load(),
		Evictions:   c.evictions.Load(),
		Expirations: c.expirations.Load(),
	}
}

// ── 백그라운드 정리 ─────────────────────────────────────────

func (c *Cache[K, V]) deleteExpired() {
	now := time.Now()
	c.mu.Lock()
	var expired []*entry[K, V]
	for key, elem := range c.items {
		e := elem.Value.(*entry[K, V])
		if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
			expired = append(expired, e)
			c.lru.Remove(elem)
			delete(c.items, key)
			c.expirations.Add(1)
		}
	}
	c.mu.Unlock()

	if c.opts.OnEviction != nil {
		for _, e := range expired {
			c.opts.OnEviction(e.key, e.value, EvictionReasonExpired)
		}
	}
}

func (c *Cache[K, V]) Close() {
	close(c.stopCh)
}
