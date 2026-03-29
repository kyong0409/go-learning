// solution/controller.go
// 리소스 컨트롤러 참고 풀이
//
// 이 파일은 참고용입니다. 먼저 직접 구현해보세요.
package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// ============================================================
// 타입 정의
// ============================================================

type EventType string

const (
	EventCreate EventType = "CREATE"
	EventUpdate EventType = "UPDATE"
	EventDelete EventType = "DELETE"
)

type Event struct {
	Type     EventType
	Resource Resource
}

type ResourceSpec struct {
	Replicas int
	Image    string
}

type ResourceStatus struct {
	ReadyReplicas int
	Phase         string
}

type Resource struct {
	Name   string
	Spec   ResourceSpec
	Status ResourceStatus
}

type Reconciler interface {
	Reconcile(ctx context.Context, key string) error
}

// ============================================================
// WorkQueue 구현
// ============================================================

type workQueue struct {
	mu         sync.Mutex
	cond       *sync.Cond
	queue      []string          // 처리 대기 순서
	queued     map[string]bool   // 큐에 있는 키
	processing map[string]bool   // 현재 처리 중인 키
	dirty      map[string]bool   // 처리 중에 재추가된 키
	retries    map[string]int    // 재시도 횟수
	shutdown   bool
	baseDelay  time.Duration
}

func NewWorkQueue() *workQueue {
	q := &workQueue{
		queued:     make(map[string]bool),
		processing: make(map[string]bool),
		dirty:      make(map[string]bool),
		retries:    make(map[string]int),
		baseDelay:  5 * time.Millisecond,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *workQueue) Add(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.shutdown {
		return
	}

	// 처리 중인 키: dirty 표시 (Done 후 재추가됨)
	if q.processing[key] {
		q.dirty[key] = true
		return
	}

	// 이미 큐에 있으면 중복 추가하지 않음
	if q.queued[key] {
		return
	}

	q.queue = append(q.queue, key)
	q.queued[key] = true
	q.cond.Signal()
}

func (q *workQueue) Get() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.queue) == 0 && !q.shutdown {
		q.cond.Wait()
	}

	if q.shutdown && len(q.queue) == 0 {
		return "", true
	}

	key := q.queue[0]
	q.queue = q.queue[1:]
	delete(q.queued, key)
	q.processing[key] = true

	return key, false
}

func (q *workQueue) Done(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.processing, key)

	// dirty 상태였으면 재추가
	if q.dirty[key] {
		delete(q.dirty, key)
		q.queue = append(q.queue, key)
		q.queued[key] = true
		q.cond.Signal()
	}
}

func (q *workQueue) Forget(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.retries, key)
}

func (q *workQueue) AddRateLimited(key string) {
	q.mu.Lock()
	retries := q.retries[key]
	q.retries[key] = retries + 1
	delay := exponentialBackoff(q.baseDelay, retries, time.Second)
	q.mu.Unlock()

	// 별도 goroutine에서 딜레이 후 추가
	go func() {
		time.Sleep(delay)
		q.Add(key)
	}()
}

func (q *workQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.queue)
}

func (q *workQueue) ShutDown() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.shutdown = true
	q.cond.Broadcast()
}

// ============================================================
// Controller 구현
// ============================================================

type Controller struct {
	name       string
	reconciler Reconciler
	queue      *workQueue
	maxRetries int
}

func NewController(name string, r Reconciler) *Controller {
	return &Controller{
		name:       name,
		reconciler: r,
		queue:      NewWorkQueue(),
		maxRetries: 5,
	}
}

func (c *Controller) Watch(ctx context.Context, eventCh <-chan Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-eventCh:
			if !ok {
				return
			}
			c.queue.Add(ev.Resource.Name)
		}
	}
}

func (c *Controller) Run(ctx context.Context, workers int) {
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c.processNextItem(ctx) {
			}
		}()
	}

	<-ctx.Done()
	c.queue.ShutDown()
	wg.Wait()
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)

	err := c.reconciler.Reconcile(ctx, key)
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	// 재시도 횟수 확인
	retries := c.retryCount(key)
	if retries < c.maxRetries {
		c.queue.AddRateLimited(key)
	} else {
		// 최대 재시도 초과 → 포기
		c.queue.Forget(key)
		fmt.Printf("[%s] 키 %s 최대 재시도 초과, 포기합니다\n", c.name, key)
	}

	return true
}

func (c *Controller) retryCount(key string) int {
	c.queue.mu.Lock()
	defer c.queue.mu.Unlock()
	return c.queue.retries[key]
}

// ============================================================
// 지수 백오프
// ============================================================

func exponentialBackoff(base time.Duration, retry int, maxDelay time.Duration) time.Duration {
	if retry < 0 {
		retry = 0
	}
	// base * 2^retry
	multiplier := math.Pow(2, float64(retry))
	delay := time.Duration(float64(base) * multiplier)
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

func main() {}
