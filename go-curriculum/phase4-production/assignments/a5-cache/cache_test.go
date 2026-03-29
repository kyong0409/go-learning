// a5-cache/cache_test.go
// TTL과 LRU 퇴거를 가진 동시성 안전 캐시 테스트 및 채점
//
// 실행:
//
//	go test ./... -v
//	go test -race ./...
//	go test ./... -v -run TestGrade
package cache_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/learn-go/a5-cache"
)

// ============================================================
// 기본 Get/Set/Delete 테스트 (20점)
// ============================================================

func TestCache_SetAndGet(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{})
	defer c.Close()

	c.Set("key1", 42)
	c.Set("key2", 100)

	if val, ok := c.Get("key1"); !ok || val != 42 {
		t.Errorf("Get key1: 기대 (42, true), 실제 (%v, %v)", val, ok)
	}
	if val, ok := c.Get("key2"); !ok || val != 100 {
		t.Errorf("Get key2: 기대 (100, true), 실제 (%v, %v)", val, ok)
	}
}

func TestCache_GetMissing(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{})
	defer c.Close()

	val, ok := c.Get("nonexistent")
	if ok || val != 0 {
		t.Errorf("존재하지 않는 키: 기대 (0, false), 실제 (%v, %v)", val, ok)
	}
}

func TestCache_Delete(t *testing.T) {
	c := cache.New[string, string](cache.Options[string, string]{})
	defer c.Close()

	c.Set("foo", "bar")
	c.Delete("foo")

	if _, ok := c.Get("foo"); ok {
		t.Error("Delete 후 Get: false여야 합니다")
	}
}

func TestCache_Overwrite(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{})
	defer c.Close()

	c.Set("k", 1)
	c.Set("k", 2)

	if val, ok := c.Get("k"); !ok || val != 2 {
		t.Errorf("덮어쓰기 후 Get: 기대 (2, true), 실제 (%v, %v)", val, ok)
	}
}

func TestCache_Len(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{})
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	if l := c.Len(); l != 3 {
		t.Errorf("Len: 기대 3, 실제 %d", l)
	}

	c.Delete("b")
	if l := c.Len(); l != 2 {
		t.Errorf("Delete 후 Len: 기대 2, 실제 %d", l)
	}
}

// ============================================================
// TTL 만료 테스트 (25점)
// ============================================================

func TestCache_TTLExpiry_Lazy(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{
		DefaultTTL: 50 * time.Millisecond,
	})
	defer c.Close()

	c.Set("expiring", 99)

	// 만료 전에는 조회 가능
	if val, ok := c.Get("expiring"); !ok || val != 99 {
		t.Errorf("만료 전 Get: 기대 (99, true), 실제 (%v, %v)", val, ok)
	}

	time.Sleep(80 * time.Millisecond)

	// 만료 후에는 조회 불가
	if _, ok := c.Get("expiring"); ok {
		t.Error("TTL 만료 후 Get: false여야 합니다")
	}
}

func TestCache_SetWithTTL(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{
		DefaultTTL: time.Hour, // 기본은 길게
	})
	defer c.Close()

	c.SetWithTTL("short", 1, 50*time.Millisecond)
	c.Set("long", 2) // 기본 TTL 사용

	time.Sleep(80 * time.Millisecond)

	if _, ok := c.Get("short"); ok {
		t.Error("짧은 TTL 항목: 만료 후 false여야 합니다")
	}
	if _, ok := c.Get("long"); !ok {
		t.Error("긴 TTL 항목: 아직 유효해야 합니다")
	}
}

func TestCache_ZeroTTL_NoExpiry(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{
		DefaultTTL: 0, // 만료 없음
	})
	defer c.Close()

	c.Set("forever", 42)
	time.Sleep(20 * time.Millisecond)

	if val, ok := c.Get("forever"); !ok || val != 42 {
		t.Errorf("TTL=0 항목: 기대 (42, true), 실제 (%v, %v)", val, ok)
	}
}

func TestCache_TTLExpiry_LenUpdate(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{
		DefaultTTL:      30 * time.Millisecond,
		CleanupInterval: 20 * time.Millisecond,
	})
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)

	time.Sleep(80 * time.Millisecond) // 만료 + 정리 기다리기

	// Get으로 지연 만료 트리거
	c.Get("a")
	c.Get("b")

	if l := c.Len(); l != 0 {
		t.Errorf("TTL 만료 후 Len: 기대 0, 실제 %d", l)
	}
}

func TestCache_EvictionCallback_Expiry(t *testing.T) {
	evicted := make(map[string]cache.EvictionReason)
	var mu sync.Mutex

	c := cache.New[string, int](cache.Options[string, int]{
		DefaultTTL: 30 * time.Millisecond,
		OnEviction: func(key string, val int, reason cache.EvictionReason) {
			mu.Lock()
			evicted[key] = reason
			mu.Unlock()
		},
	})
	defer c.Close()

	c.Set("k1", 1)
	time.Sleep(60 * time.Millisecond)
	c.Get("k1") // 지연 만료 트리거

	mu.Lock()
	reason, ok := evicted["k1"]
	mu.Unlock()

	if !ok {
		t.Error("TTL 만료 시 OnEviction 콜백이 호출되어야 합니다")
	}
	if ok && reason != cache.EvictionReasonExpired {
		t.Errorf("퇴거 사유: 기대 %q, 실제 %q", cache.EvictionReasonExpired, reason)
	}
}

// ============================================================
// LRU 퇴거 테스트 (25점)
// ============================================================

func TestCache_LRU_MaxSize(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{
		MaxSize: 3,
	})
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	// LRU 순서: a(가장 오래됨) < b < c(가장 최근)

	c.Set("d", 4) // 최대 크기 초과: a가 퇴거되어야 함

	if _, ok := c.Get("a"); ok {
		t.Error("LRU 퇴거: 가장 오래된 'a'가 제거되어야 합니다")
	}
	if _, ok := c.Get("d"); !ok {
		t.Error("새로 추가된 'd': 존재해야 합니다")
	}
	if c.Len() > 3 {
		t.Errorf("MaxSize 초과: 기대 <= 3, 실제 %d", c.Len())
	}
}

func TestCache_LRU_GetUpdatesOrder(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{
		MaxSize: 3,
	})
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// 'a'를 Get하여 최근 사용으로 갱신
	c.Get("a")
	// LRU 순서: b(가장 오래됨) < c < a(가장 최근)

	c.Set("d", 4) // b가 퇴거되어야 함

	if _, ok := c.Get("b"); ok {
		t.Error("Get으로 갱신 후 LRU: 'b'가 제거되어야 합니다")
	}
	if _, ok := c.Get("a"); !ok {
		t.Error("최근 사용된 'a': 남아 있어야 합니다")
	}
}

func TestCache_LRU_SetUpdatesOrder(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{
		MaxSize: 2,
	})
	defer c.Close()

	c.Set("x", 1)
	c.Set("y", 2)
	c.Set("x", 10) // x를 최근으로 갱신
	// LRU 순서: y(가장 오래됨) < x(가장 최근)

	c.Set("z", 3) // y가 퇴거되어야 함

	if _, ok := c.Get("y"); ok {
		t.Error("Set으로 갱신 후 LRU: 'y'가 제거되어야 합니다")
	}
	if val, ok := c.Get("x"); !ok || val != 10 {
		t.Errorf("갱신된 'x': 기대 (10, true), 실제 (%v, %v)", val, ok)
	}
}

func TestCache_LRU_EvictionCallback(t *testing.T) {
	var evictedKey string
	var evictedReason cache.EvictionReason

	c := cache.New[string, int](cache.Options[string, int]{
		MaxSize: 2,
		OnEviction: func(key string, val int, reason cache.EvictionReason) {
			evictedKey = key
			evictedReason = reason
		},
	})
	defer c.Close()

	c.Set("p", 1)
	c.Set("q", 2)
	c.Set("r", 3) // p 퇴거

	if evictedKey != "p" {
		t.Errorf("LRU 퇴거 키: 기대 %q, 실제 %q", "p", evictedKey)
	}
	if evictedReason != cache.EvictionReasonCapacity {
		t.Errorf("LRU 퇴거 사유: 기대 %q, 실제 %q", cache.EvictionReasonCapacity, evictedReason)
	}
}

// ============================================================
// 통계 및 콜백 테스트 (15점)
// ============================================================

func TestCache_Stats_HitsAndMisses(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{})
	defer c.Close()

	c.Set("k", 1)
	c.Get("k")        // 히트
	c.Get("k")        // 히트
	c.Get("missing")  // 미스

	stats := c.Stats()
	if stats.Hits != 2 {
		t.Errorf("Hits: 기대 2, 실제 %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Misses: 기대 1, 실제 %d", stats.Misses)
	}
}

func TestCache_Stats_Evictions(t *testing.T) {
	c := cache.New[string, int](cache.Options[string, int]{
		MaxSize: 2,
	})
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // 퇴거 1회
	c.Set("d", 4) // 퇴거 1회

	stats := c.Stats()
	if stats.Evictions < 2 {
		t.Errorf("Evictions: 기대 >= 2, 실제 %d", stats.Evictions)
	}
}

func TestCache_Flush(t *testing.T) {
	flushed := 0
	var mu sync.Mutex

	c := cache.New[string, int](cache.Options[string, int]{
		OnEviction: func(key string, val int, reason cache.EvictionReason) {
			if reason == cache.EvictionReasonFlushed {
				mu.Lock()
				flushed++
				mu.Unlock()
			}
		},
	})
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Flush()

	if c.Len() != 0 {
		t.Errorf("Flush 후 Len: 기대 0, 실제 %d", c.Len())
	}
	mu.Lock()
	f := flushed
	mu.Unlock()
	if f != 3 {
		t.Errorf("Flush 콜백 횟수: 기대 3, 실제 %d", f)
	}
}

func TestCache_DeleteCallback(t *testing.T) {
	var reason cache.EvictionReason
	c := cache.New[string, int](cache.Options[string, int]{
		OnEviction: func(key string, val int, r cache.EvictionReason) {
			reason = r
		},
	})
	defer c.Close()

	c.Set("x", 1)
	c.Delete("x")

	if reason != cache.EvictionReasonDeleted {
		t.Errorf("Delete 퇴거 사유: 기대 %q, 실제 %q", cache.EvictionReasonDeleted, reason)
	}
}

// ============================================================
// 동시성 테스트 (15점)
// ============================================================

func TestCache_ConcurrentReadWrite(t *testing.T) {
	c := cache.New[int, int](cache.Options[int, int]{
		MaxSize: 100,
	})
	defer c.Close()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			c.Set(n, n*2)
		}(i)
		go func(n int) {
			defer wg.Done()
			c.Get(n)
		}(i)
	}
	wg.Wait()
	// 레이스 컨디션 없이 완료되면 통과
}

func TestCache_ConcurrentExpiry(t *testing.T) {
	c := cache.New[int, int](cache.Options[int, int]{
		DefaultTTL: 10 * time.Millisecond,
	})
	defer c.Close()

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set(n, n)
			time.Sleep(20 * time.Millisecond)
			c.Get(n) // 만료 트리거
		}(i)
	}
	wg.Wait()
	// 레이스 컨디션 없이 완료되면 통과
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

	// 기본 동작 (20점)
	check("Set/Get 기본 동작", 8, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{})
		defer c.Close()
		c.Set("k", 7)
		val, ok := c.Get("k")
		return ok && val == 7
	})
	check("Delete 및 Len", 7, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{})
		defer c.Close()
		c.Set("a", 1)
		c.Set("b", 2)
		c.Delete("a")
		_, aOk := c.Get("a")
		return !aOk && c.Len() == 1
	})
	check("Set 덮어쓰기", 5, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{})
		defer c.Close()
		c.Set("k", 1)
		c.Set("k", 99)
		val, ok := c.Get("k")
		return ok && val == 99
	})

	// TTL 만료 (25점)
	check("DefaultTTL 만료 (지연)", 10, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{
			DefaultTTL: 30 * time.Millisecond,
		})
		defer c.Close()
		c.Set("k", 1)
		time.Sleep(60 * time.Millisecond)
		_, ok := c.Get("k")
		return !ok
	})
	check("SetWithTTL 개별 TTL", 10, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{
			DefaultTTL: time.Hour,
		})
		defer c.Close()
		c.SetWithTTL("short", 1, 30*time.Millisecond)
		c.Set("long", 2)
		time.Sleep(60 * time.Millisecond)
		_, shortOk := c.Get("short")
		_, longOk := c.Get("long")
		return !shortOk && longOk
	})
	check("TTL=0 만료 없음", 5, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{DefaultTTL: 0})
		defer c.Close()
		c.Set("k", 42)
		time.Sleep(20 * time.Millisecond)
		val, ok := c.Get("k")
		return ok && val == 42
	})

	// LRU 퇴거 (25점)
	check("MaxSize LRU 퇴거", 12, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{MaxSize: 3})
		defer c.Close()
		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)
		c.Set("d", 4)
		_, aOk := c.Get("a")
		_, dOk := c.Get("d")
		return !aOk && dOk && c.Len() <= 3
	})
	check("Get이 LRU 순서 갱신", 8, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{MaxSize: 3})
		defer c.Close()
		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)
		c.Get("a") // a를 최근으로
		c.Set("d", 4) // b가 퇴거
		_, aOk := c.Get("a")
		_, bOk := c.Get("b")
		return aOk && !bOk
	})
	check("LRU 퇴거 콜백", 5, func() bool {
		var reason cache.EvictionReason
		c := cache.New[string, int](cache.Options[string, int]{
			MaxSize: 1,
			OnEviction: func(k string, v int, r cache.EvictionReason) {
				reason = r
			},
		})
		defer c.Close()
		c.Set("a", 1)
		c.Set("b", 2)
		return reason == cache.EvictionReasonCapacity
	})

	// 콜백 및 통계 (15점)
	check("Stats Hits/Misses", 8, func() bool {
		c := cache.New[string, int](cache.Options[string, int]{})
		defer c.Close()
		c.Set("k", 1)
		c.Get("k")
		c.Get("k")
		c.Get("x")
		s := c.Stats()
		return s.Hits == 2 && s.Misses == 1
	})
	check("Flush 콜백", 7, func() bool {
		count := 0
		var mu sync.Mutex
		c := cache.New[string, int](cache.Options[string, int]{
			OnEviction: func(k string, v int, r cache.EvictionReason) {
				if r == cache.EvictionReasonFlushed {
					mu.Lock()
					count++
					mu.Unlock()
				}
			},
		})
		defer c.Close()
		c.Set("a", 1)
		c.Set("b", 2)
		c.Flush()
		mu.Lock()
		n := count
		mu.Unlock()
		return c.Len() == 0 && n == 2
	})

	// 동시성 (15점) — race detector로 검증 (-race 플래그 사용)
	check("동시 읽기/쓰기 안전성", 15, func() bool {
		c := cache.New[int, int](cache.Options[int, int]{MaxSize: 50})
		defer c.Close()
		var wg sync.WaitGroup
		for i := range 30 {
			wg.Add(2)
			go func(n int) { defer wg.Done(); c.Set(n, n) }(i)
			go func(n int) { defer wg.Done(); c.Get(n) }(i)
		}
		wg.Wait()
		return true // 패닉/데드락 없으면 통과
	})

	// 결과 출력
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════╗")
	fmt.Println("║   과제 A5: TTL/LRU 캐시 구현 채점 결과        ║")
	fmt.Println("╠════════════════════════════════════════════════╣")
	passed := 0
	for _, r := range results {
		mark := "✗"
		if r.pass {
			mark = "✓"
			passed++
		}
		fmt.Printf("║  %s %-34s %3d/%d점  ║\n", mark, r.name, r.pts, r.maxPt)
	}
	fmt.Println("╠════════════════════════════════════════════════╣")
	fmt.Printf("║  통과: %d/%d                                     ║\n", passed, len(results))
	fmt.Printf("║  점수: %d/%d                                    ║\n", score, total)
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
	fmt.Printf("║  등급: %s                                        ║\n", grade)
	fmt.Println("╚════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("=== 채점 결과 ===")
	fmt.Printf("통과: %d/%d\n", passed, len(results))
	fmt.Printf("점수: %d/%d\n", score, total)
}
