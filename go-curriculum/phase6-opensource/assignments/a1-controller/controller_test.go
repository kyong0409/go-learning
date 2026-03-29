// controller_test.go
// 리소스 컨트롤러 테스트 및 채점
//
// 실행:
//
//	go test -v
//	go test -v -run TestGrade   (채점만)
package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
// 테스트 헬퍼
// ============================================================

// mockReconciler는 테스트용 Reconciler입니다.
type mockReconciler struct {
	mu       sync.Mutex
	calls    []string          // 호출된 키 목록 (순서대로)
	failKeys map[string]int    // key → 실패시킬 횟수
	callCount map[string]int   // key → 호출 횟수
}

func newMockReconciler() *mockReconciler {
	return &mockReconciler{
		failKeys:  make(map[string]int),
		callCount: make(map[string]int),
	}
}

func (m *mockReconciler) Reconcile(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, key)
	m.callCount[key]++
	if fails := m.failKeys[key]; fails > 0 {
		m.failKeys[key]--
		return fmt.Errorf("reconcile 실패: %s", key)
	}
	return nil
}

func (m *mockReconciler) getCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.calls))
	copy(result, m.calls)
	return result
}

func (m *mockReconciler) getCallCount(key string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount[key]
}

// ============================================================
// WorkQueue 테스트
// ============================================================

func TestWorkQueue_AddGet(t *testing.T) {
	q := NewWorkQueue()
	defer q.ShutDown()

	q.Add("key1")
	q.Add("key2")

	if q.Len() != 2 {
		t.Errorf("Len() = %d, 원하는 값: 2", q.Len())
	}

	key, shutdown := q.Get()
	if shutdown {
		t.Fatal("shutdown이 true여서는 안 됩니다")
	}
	if key != "key1" && key != "key2" {
		t.Errorf("예상치 못한 키: %s", key)
	}
}

func TestWorkQueue_DuplicateAdd(t *testing.T) {
	q := NewWorkQueue()
	defer q.ShutDown()

	q.Add("key1")
	q.Add("key1") // 중복
	q.Add("key1") // 중복

	if q.Len() != 1 {
		t.Errorf("중복 Add 후 Len() = %d, 원하는 값: 1", q.Len())
	}
}

func TestWorkQueue_DoneRequeues(t *testing.T) {
	q := NewWorkQueue()
	defer q.ShutDown()

	q.Add("key1")
	key, _ := q.Get()

	// 처리 중에 같은 키 추가 → Done 후 재추가되어야 함
	q.Add(key)

	q.Done(key)

	// 재추가되었는지 확인
	if q.Len() != 1 {
		t.Errorf("Done 후 Len() = %d, 원하는 값: 1 (재추가됨)", q.Len())
	}
}

func TestWorkQueue_Shutdown(t *testing.T) {
	q := NewWorkQueue()

	done := make(chan struct{})
	go func() {
		_, shutdown := q.Get()
		if !shutdown {
			t.Error("shutdown 후 Get은 true를 반환해야 합니다")
		}
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	q.ShutDown()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("ShutDown 후 Get이 1초 내에 반환되지 않음")
	}
}

func TestWorkQueue_AddRateLimited(t *testing.T) {
	q := NewWorkQueue()
	defer q.ShutDown()

	start := time.Now()
	q.AddRateLimited("key1") // 첫 번째 = baseDelay * 2^0 = baseDelay

	key, _ := q.Get()
	elapsed := time.Since(start)

	if key != "key1" {
		t.Errorf("예상 키: key1, 실제: %s", key)
	}
	// baseDelay(5ms)보다는 지연되어야 함
	if elapsed < 1*time.Millisecond {
		t.Errorf("AddRateLimited 후 딜레이가 너무 짧음: %v", elapsed)
	}
}

func TestWorkQueue_Forget(t *testing.T) {
	q := NewWorkQueue()
	defer q.ShutDown()

	// 재시도 횟수를 쌓은 뒤 Forget
	q.AddRateLimited("key1")
	q.AddRateLimited("key1")
	q.Forget("key1")

	// Forget 후 AddRateLimited는 다시 baseDelay부터 시작해야 함
	start := time.Now()
	q.AddRateLimited("key1")

	// 비어있지 않을 때까지 잠시 대기
	time.Sleep(50 * time.Millisecond)
	elapsed := time.Since(start)

	// Forget 후이므로 딜레이가 크지 않아야 함 (2^0 * base)
	if elapsed > 200*time.Millisecond {
		t.Errorf("Forget 후 딜레이가 너무 큼: %v", elapsed)
	}
}

// ============================================================
// Controller 테스트
// ============================================================

func TestController_BasicReconcile(t *testing.T) {
	r := newMockReconciler()
	c := NewController("test", r)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventCh := make(chan Event, 10)
	go c.Watch(ctx, eventCh)

	eventCh <- Event{Type: EventCreate, Resource: Resource{Name: "res-1"}}

	go c.Run(ctx, 1)

	// Reconcile이 호출될 때까지 대기
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if r.getCallCount("res-1") > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if r.getCallCount("res-1") == 0 {
		t.Error("res-1에 대한 Reconcile이 호출되지 않았습니다")
	}
}

func TestController_MultipleResources(t *testing.T) {
	r := newMockReconciler()
	c := NewController("test", r)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventCh := make(chan Event, 10)
	go c.Watch(ctx, eventCh)
	go c.Run(ctx, 2)

	resources := []string{"res-1", "res-2", "res-3"}
	for _, name := range resources {
		eventCh <- Event{Type: EventCreate, Resource: Resource{Name: name}}
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		allDone := true
		for _, name := range resources {
			if r.getCallCount(name) == 0 {
				allDone = false
				break
			}
		}
		if allDone {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	for _, name := range resources {
		if r.getCallCount(name) == 0 {
			t.Errorf("리소스 %s에 대한 Reconcile이 호출되지 않았습니다", name)
		}
	}
}

func TestController_DuplicateEvents(t *testing.T) {
	var callCount int64
	r := &countingReconciler{count: &callCount}
	c := NewController("test", r)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventCh := make(chan Event, 100)
	go c.Watch(ctx, eventCh)

	// 같은 리소스에 대한 이벤트 10개 연속 전송
	for i := 0; i < 10; i++ {
		eventCh <- Event{Type: EventUpdate, Resource: Resource{Name: "res-1"}}
	}

	go c.Run(ctx, 1)

	time.Sleep(300 * time.Millisecond)

	count := atomic.LoadInt64(&callCount)
	// 중복 제거로 10번보다 훨씬 적게 호출되어야 함
	if count > 5 {
		t.Errorf("중복 이벤트 10개에 대해 %d번 Reconcile 호출됨 (5번 이하여야 함)", count)
	}
}

// countingReconciler는 호출 횟수만 세는 Reconciler입니다.
type countingReconciler struct {
	count *int64
}

func (r *countingReconciler) Reconcile(_ context.Context, _ string) error {
	atomic.AddInt64(r.count, 1)
	return nil
}

func TestController_RetryOnError(t *testing.T) {
	r := newMockReconciler()
	r.failKeys["res-1"] = 2 // 처음 2번 실패

	c := NewController("test", r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh := make(chan Event, 10)
	go c.Watch(ctx, eventCh)
	go c.Run(ctx, 1)

	eventCh <- Event{Type: EventCreate, Resource: Resource{Name: "res-1"}}

	// 실패 + 재시도 + 성공까지 기다림
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if r.getCallCount("res-1") >= 3 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	count := r.getCallCount("res-1")
	if count < 3 {
		t.Errorf("재시도 포함 최소 3번 Reconcile 필요, 실제: %d번", count)
	}
}

func TestController_MaxRetries(t *testing.T) {
	r := newMockReconciler()
	r.failKeys["res-1"] = 100 // 항상 실패

	c := NewController("test", r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh := make(chan Event, 10)
	go c.Watch(ctx, eventCh)
	go c.Run(ctx, 1)

	eventCh <- Event{Type: EventCreate, Resource: Resource{Name: "res-1"}}

	time.Sleep(2 * time.Second)

	count := r.getCallCount("res-1")
	// maxRetries(5) + 초기 1번 = 최대 6번
	if count > 7 {
		t.Errorf("최대 재시도 초과: %d번 호출됨 (최대 7번)", count)
	}
	if count < 1 {
		t.Error("최소 1번은 호출되어야 합니다")
	}
}

func TestController_ContextCancel(t *testing.T) {
	r := newMockReconciler()
	c := NewController("test", r)

	ctx, cancel := context.WithCancel(context.Background())

	eventCh := make(chan Event, 10)
	go c.Watch(ctx, eventCh)

	done := make(chan struct{})
	go func() {
		c.Run(ctx, 2)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("ctx 취소 후 Run이 1초 내에 반환되지 않음")
	}
}

func TestController_DeleteEvent(t *testing.T) {
	r := newMockReconciler()
	c := NewController("test", r)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventCh := make(chan Event, 10)
	go c.Watch(ctx, eventCh)
	go c.Run(ctx, 1)

	// DELETE 이벤트도 Reconcile되어야 함
	eventCh <- Event{Type: EventDelete, Resource: Resource{Name: "res-del"}}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if r.getCallCount("res-del") > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if r.getCallCount("res-del") == 0 {
		t.Error("DELETE 이벤트에 대한 Reconcile이 호출되지 않았습니다")
	}
}

func TestController_ConcurrentWorkers(t *testing.T) {
	var mu sync.Mutex
	maxConcurrent := 0
	current := 0

	r := &concurrentReconciler{
		mu:            &mu,
		maxConcurrent: &maxConcurrent,
		current:       &current,
		delay:         20 * time.Millisecond,
	}

	c := NewController("test", r)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh := make(chan Event, 20)
	go c.Watch(ctx, eventCh)
	go c.Run(ctx, 4) // 4개 워커

	for i := 0; i < 8; i++ {
		eventCh <- Event{Type: EventCreate, Resource: Resource{Name: fmt.Sprintf("res-%d", i)}}
	}

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	mc := maxConcurrent
	mu.Unlock()

	if mc < 2 {
		t.Errorf("동시 실행 최대값 = %d, 2 이상이어야 합니다 (워커 4개)", mc)
	}
}

// concurrentReconciler는 동시 실행 수를 측정하는 Reconciler입니다.
type concurrentReconciler struct {
	mu            *sync.Mutex
	maxConcurrent *int
	current       *int
	delay         time.Duration
}

func (r *concurrentReconciler) Reconcile(_ context.Context, _ string) error {
	r.mu.Lock()
	*r.current++
	if *r.current > *r.maxConcurrent {
		*r.maxConcurrent = *r.current
	}
	r.mu.Unlock()

	time.Sleep(r.delay)

	r.mu.Lock()
	*r.current--
	r.mu.Unlock()
	return nil
}

func TestExponentialBackoff(t *testing.T) {
	base := 5 * time.Millisecond
	max := time.Second

	cases := []struct {
		retry    int
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{0, 4 * time.Millisecond, 20 * time.Millisecond},
		{1, 8 * time.Millisecond, 40 * time.Millisecond},
		{2, 16 * time.Millisecond, 80 * time.Millisecond},
		{10, 900 * time.Millisecond, max + time.Millisecond},
	}

	for _, tc := range cases {
		d := exponentialBackoff(base, tc.retry, max)
		if d < tc.minDelay || d > tc.maxDelay {
			t.Errorf("retry=%d: exponentialBackoff=%v, 예상 범위 [%v, %v]",
				tc.retry, d, tc.minDelay, tc.maxDelay)
		}
	}
}

// ============================================================
// 채점 함수 (TestGrade)
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════════════")
	fmt.Println("  과제 A1: 리소스 컨트롤러 채점 결과")
	fmt.Println("  패턴: Kubernetes Controller + WorkQueue")
	fmt.Println("═══════════════════════════════════════════════════")

	// WorkQueue 기본 동작 (20점)
	t.Run("WorkQueue_기본동작", func(t *testing.T) {
		q := NewWorkQueue()
		defer q.ShutDown()
		q.Add("a")
		q.Add("b")
		key, ok := q.Get()
		if !ok && (key == "a" || key == "b") && q.Len() == 1 {
			score += 20
			fmt.Printf("  ✓ WorkQueue 기본 동작 (Add/Get/Len)   20/20점\n")
		} else if key == "a" || key == "b" {
			score += 10
			fmt.Printf("  △ WorkQueue 부분 동작                 10/20점\n")
		} else {
			fmt.Printf("  ✗ WorkQueue 기본 동작                  0/20점\n")
		}
	})

	// 중복 키 제거 (15점)
	t.Run("중복키_제거", func(t *testing.T) {
		q := NewWorkQueue()
		defer q.ShutDown()
		q.Add("dup")
		q.Add("dup")
		q.Add("dup")
		if q.Len() == 1 {
			score += 15
			fmt.Printf("  ✓ 중복 키 제거                        15/15점\n")
		} else {
			fmt.Printf("  ✗ 중복 키 제거 (Len=%d, 원하는 값:1)   0/15점\n", q.Len())
		}
	})

	// Controller Watch → Enqueue (15점)
	t.Run("Watch_Enqueue", func(t *testing.T) {
		r := newMockReconciler()
		c := NewController("grade", r)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		eventCh := make(chan Event, 5)
		go c.Watch(ctx, eventCh)
		go c.Run(ctx, 1)

		eventCh <- Event{Type: EventCreate, Resource: Resource{Name: "grade-res"}}

		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			if r.getCallCount("grade-res") > 0 {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}

		if r.getCallCount("grade-res") > 0 {
			score += 15
			fmt.Printf("  ✓ Watch → Enqueue → Reconcile         15/15점\n")
		} else {
			fmt.Printf("  ✗ Watch → Enqueue → Reconcile          0/15점\n")
		}
	})

	// Reconcile 호출 (15점)
	t.Run("Reconcile_호출", func(t *testing.T) {
		r := newMockReconciler()
		c := NewController("grade", r)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		eventCh := make(chan Event, 10)
		go c.Watch(ctx, eventCh)
		go c.Run(ctx, 2)

		for i := 0; i < 3; i++ {
			eventCh <- Event{Type: EventCreate, Resource: Resource{Name: fmt.Sprintf("r%d", i)}}
		}

		time.Sleep(500 * time.Millisecond)
		allDone := r.getCallCount("r0") > 0 && r.getCallCount("r1") > 0 && r.getCallCount("r2") > 0
		if allDone {
			score += 15
			fmt.Printf("  ✓ 다중 리소스 Reconcile              15/15점\n")
		} else {
			fmt.Printf("  ✗ 다중 리소스 Reconcile               0/15점\n")
		}
	})

	// 지수 백오프 재시도 (20점)
	t.Run("지수_백오프_재시도", func(t *testing.T) {
		r := newMockReconciler()
		r.failKeys["retry-res"] = 2
		c := NewController("grade", r)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		eventCh := make(chan Event, 5)
		go c.Watch(ctx, eventCh)
		go c.Run(ctx, 1)
		eventCh <- Event{Type: EventCreate, Resource: Resource{Name: "retry-res"}}

		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if r.getCallCount("retry-res") >= 3 {
				break
			}
			time.Sleep(30 * time.Millisecond)
		}

		if r.getCallCount("retry-res") >= 3 {
			score += 20
			fmt.Printf("  ✓ 지수 백오프 재시도                  20/20점\n")
		} else {
			fmt.Printf("  ✗ 지수 백오프 재시도 (호출=%d, 필요:3+)  0/20점\n", r.getCallCount("retry-res"))
		}
	})

	// 컨텍스트 취소 시 정상 종료 (15점)
	t.Run("컨텍스트_취소", func(t *testing.T) {
		r := newMockReconciler()
		c := NewController("grade", r)
		ctx, cancel := context.WithCancel(context.Background())

		eventCh := make(chan Event, 5)
		go c.Watch(ctx, eventCh)

		done := make(chan struct{})
		go func() {
			c.Run(ctx, 2)
			close(done)
		}()

		time.Sleep(30 * time.Millisecond)
		cancel()

		select {
		case <-done:
			score += 15
			fmt.Printf("  ✓ 컨텍스트 취소 시 정상 종료          15/15점\n")
		case <-time.After(time.Second):
			fmt.Printf("  ✗ 컨텍스트 취소 시 종료 안 됨           0/15점\n")
		}
	})

	fmt.Println("───────────────────────────────────────────────────")
	fmt.Printf("  최종 점수: %d / %d점\n", score, total)

	grade := "F"
	switch {
	case score >= 95:
		grade = "A+"
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	}
	fmt.Printf("  등급: %s\n", grade)
	fmt.Print("═══════════════════════════════════════════════════\n\n")
}
