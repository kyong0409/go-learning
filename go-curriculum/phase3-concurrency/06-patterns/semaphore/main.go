// 패키지 선언
package main

// 동시성 패턴: 세마포어 (Semaphore)
//
// 세마포어는 동시에 실행할 수 있는 고루틴 수를 제한합니다.
// Go 표준 라이브러리에 세마포어가 없어서 버퍼 채널로 구현합니다.
//
// 버퍼 채널 세마포어 원리:
// - 버퍼 크기 = 최대 동시 실행 수
// - 획득(Acquire): 채널에 값을 넣음 (가득 차면 블로킹)
// - 반환(Release): 채널에서 값을 뺌
//
// 사용 사례:
// - DB 커넥션 풀 제한
// - 파일 동시 접근 제한
// - 외부 API 동시 호출 제한

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 1. 기본 세마포어 구현
// ─────────────────────────────────────────

// Semaphore는 버퍼 채널 기반 세마포어입니다.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore는 n개 동시 실행을 허용하는 세마포어를 생성합니다.
func NewSemaphore(n int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, n)}
}

// Acquire는 세마포어를 획득합니다 (슬롯이 없으면 블로킹).
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{} // 버퍼에 넣기 (가득 찬 경우 대기)
}

// AcquireCtx는 Context를 지원하는 획득입니다.
func (s *Semaphore) AcquireCtx(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release는 세마포어를 반환합니다.
func (s *Semaphore) Release() {
	<-s.ch // 버퍼에서 꺼내기
}

// Available는 현재 사용 가능한 슬롯 수를 반환합니다.
func (s *Semaphore) Available() int {
	return cap(s.ch) - len(s.ch)
}

// Used는 현재 사용 중인 슬롯 수를 반환합니다.
func (s *Semaphore) Used() int {
	return len(s.ch)
}

func basicSemaphoreDemo() {
	fmt.Println("\n--- 1. 기본 세마포어 ---")
	fmt.Println("  최대 3개 동시 실행:")

	sem := NewSemaphore(3) // 최대 3개 동시 실행
	var wg sync.WaitGroup

	start := time.Now()

	for i := 1; i <= 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sem.Acquire() // 슬롯 획득 (최대 3개까지만 동시 진입)
			defer sem.Release()

			fmt.Printf("  작업 #%d 시작 (동시 실행: %d개, %.0fms)\n",
				id, sem.Used(), float64(time.Since(start).Milliseconds()))

			// 작업 시뮬레이션
			time.Sleep(time.Duration(100+rand.Intn(100)) * time.Millisecond)

			fmt.Printf("  작업 #%d 완료 (%.0fms)\n",
				id, float64(time.Since(start).Milliseconds()))
		}(i)
	}

	wg.Wait()
	fmt.Printf("  전체 소요 시간: %v\n", time.Since(start).Round(time.Millisecond))
}

// ─────────────────────────────────────────
// 2. 가중치 세마포어 (Weighted Semaphore)
// ─────────────────────────────────────────

// WeightedSemaphore는 가중치 기반 세마포어입니다.
// 작업마다 다른 양의 리소스를 사용할 수 있습니다.
// 예: 작은 쿼리=1 토큰, 대용량 쿼리=5 토큰
type WeightedSemaphore struct {
	mu      sync.Mutex
	cond    *sync.Cond
	current int64
	max     int64
}

// NewWeightedSemaphore는 최대 n 가중치를 허용하는 세마포어를 생성합니다.
func NewWeightedSemaphore(max int64) *WeightedSemaphore {
	ws := &WeightedSemaphore{max: max}
	ws.cond = sync.NewCond(&ws.mu)
	return ws
}

// Acquire는 n 가중치를 획득합니다.
func (ws *WeightedSemaphore) Acquire(ctx context.Context, n int64) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for ws.current+n > ws.max {
		// 조건 변수로 대기
		waitCh := make(chan struct{})
		go func() {
			ws.cond.Wait()
			close(waitCh)
		}()

		ws.mu.Unlock()
		select {
		case <-waitCh:
			ws.mu.Lock()
		case <-ctx.Done():
			ws.mu.Lock()
			return ctx.Err()
		}
	}

	ws.current += n
	return nil
}

// Release는 n 가중치를 반환합니다.
func (ws *WeightedSemaphore) Release(n int64) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.current -= n
	ws.cond.Broadcast()
}

func weightedSemaphoreDemo() {
	fmt.Println("\n--- 2. 가중치 세마포어 ---")
	fmt.Println("  총 10 토큰: 소형 작업=1 토큰, 대형 작업=3 토큰")

	ws := NewWeightedSemaphore(10)
	ctx := context.Background()
	var wg sync.WaitGroup

	type task struct {
		id     int
		weight int64
		name   string
	}

	tasks := []task{
		{1, 1, "소형"},
		{2, 3, "대형"},
		{3, 1, "소형"},
		{4, 3, "대형"},
		{5, 1, "소형"},
		{6, 3, "대형"},
		{7, 1, "소형"},
		{8, 1, "소형"},
	}

	start := time.Now()
	for _, t := range tasks {
		wg.Add(1)
		go func(task task) {
			defer wg.Done()

			if err := ws.Acquire(ctx, task.weight); err != nil {
				fmt.Printf("  작업 #%d 취소\n", task.id)
				return
			}
			defer ws.Release(task.weight)

			fmt.Printf("  %s 작업 #%d 시작 (%d 토큰, %.0fms)\n",
				task.name, task.id, task.weight,
				float64(time.Since(start).Milliseconds()))

			time.Sleep(time.Duration(task.weight*50) * time.Millisecond)
		}(t)
	}

	wg.Wait()
	fmt.Printf("  완료 (%v)\n", time.Since(start).Round(time.Millisecond))
}

// ─────────────────────────────────────────
// 3. DB 커넥션 풀 시뮬레이션
// ─────────────────────────────────────────

// DBConnection은 데이터베이스 연결을 나타냅니다.
type DBConnection struct {
	ID int
}

// ConnectionPool은 세마포어 기반 커넥션 풀입니다.
type ConnectionPool struct {
	sem         *Semaphore
	connections []*DBConnection
	mu          sync.Mutex
	available   []*DBConnection
}

// NewConnectionPool은 커넥션 풀을 생성합니다.
func NewConnectionPool(size int) *ConnectionPool {
	pool := &ConnectionPool{
		sem:         NewSemaphore(size),
		connections: make([]*DBConnection, size),
		available:   make([]*DBConnection, 0, size),
	}
	for i := 0; i < size; i++ {
		conn := &DBConnection{ID: i + 1}
		pool.connections[i] = conn
		pool.available = append(pool.available, conn)
	}
	return pool
}

// Get은 커넥션을 획득합니다.
func (p *ConnectionPool) Get(ctx context.Context) (*DBConnection, error) {
	if err := p.sem.AcquireCtx(ctx); err != nil {
		return nil, fmt.Errorf("커넥션 획득 실패: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	conn := p.available[len(p.available)-1]
	p.available = p.available[:len(p.available)-1]
	return conn, nil
}

// Put은 커넥션을 반환합니다.
func (p *ConnectionPool) Put(conn *DBConnection) {
	p.mu.Lock()
	p.available = append(p.available, conn)
	p.mu.Unlock()
	p.sem.Release()
}

// Query는 쿼리를 실행합니다.
func (p *ConnectionPool) Query(ctx context.Context, query string) (string, error) {
	conn, err := p.Get(ctx)
	if err != nil {
		return "", err
	}
	defer p.Put(conn)

	// 쿼리 실행 시뮬레이션
	delay := time.Duration(50+rand.Intn(100)) * time.Millisecond
	select {
	case <-time.After(delay):
		return fmt.Sprintf("연결 #%d: '%s' 결과", conn.ID, query), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func connectionPoolDemo() {
	fmt.Println("\n--- 3. DB 커넥션 풀 (세마포어 기반) ---")
	fmt.Println("  풀 크기: 3, 동시 요청: 8개")

	pool := NewConnectionPool(3)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	queries := []string{
		"SELECT * FROM users",
		"SELECT * FROM orders",
		"INSERT INTO logs VALUES(...)",
		"UPDATE products SET ...",
		"SELECT COUNT(*) FROM items",
		"DELETE FROM sessions WHERE ...",
		"SELECT * FROM reports",
		"SELECT * FROM analytics",
	}

	start := time.Now()
	for i, query := range queries {
		wg.Add(1)
		go func(id int, q string) {
			defer wg.Done()
			result, err := pool.Query(ctx, q)
			if err != nil {
				fmt.Printf("  쿼리 #%d 실패: %v\n", id, err)
				return
			}
			fmt.Printf("  쿼리 #%d 완료 (%.0fms): %s\n",
				id, float64(time.Since(start).Milliseconds()), result)
		}(i+1, query)
	}

	wg.Wait()
	fmt.Printf("  전체 소요 시간: %v\n", time.Since(start).Round(time.Millisecond))
}

// ─────────────────────────────────────────
// 4. 세마포어 vs 뮤텍스 비교
// ─────────────────────────────────────────

func semaphoreVsMutex() {
	fmt.Println("\n--- 4. 세마포어 vs 뮤텍스 ---")

	fmt.Println("  뮤텍스 (Mutex):")
	fmt.Println("  - 이진 세마포어 (0 또는 1)")
	fmt.Println("  - 동시에 1개만 임계 구역 진입 가능")
	fmt.Println("  - Lock/Unlock 쌍으로 사용")
	fmt.Println("  - 소유권 개념: Lock한 고루틴이 Unlock해야 함")
	fmt.Println()
	fmt.Println("  세마포어 (Semaphore):")
	fmt.Println("  - 카운팅 세마포어 (0..N)")
	fmt.Println("  - 동시에 N개까지 진입 가능")
	fmt.Println("  - Acquire/Release 쌍으로 사용")
	fmt.Println("  - 소유권 없음: 다른 고루틴이 Release 가능")
	fmt.Println()
	fmt.Println("  선택 기준:")
	fmt.Println("  - 임계 구역 보호 → Mutex")
	fmt.Println("  - 동시 실행 수 제한 → Semaphore")
	fmt.Println("  - 리소스 풀 관리 → Semaphore")

	// 실제 비교 데모
	fmt.Println("\n  동시 실행 제한 비교 (최대 2개):")

	// Mutex: 항상 1개만
	var mu sync.Mutex
	var wg sync.WaitGroup
	executing := 0

	for i := 1; i <= 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mu.Lock()
			executing++
			fmt.Printf("  Mutex 작업 #%d: 동시 실행=%d (항상 1)\n", id, executing)
			time.Sleep(30 * time.Millisecond)
			executing--
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	// Semaphore: 최대 2개
	sem := NewSemaphore(2)
	executing = 0
	var mu2 sync.Mutex

	for i := 1; i <= 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem.Acquire()
			mu2.Lock()
			executing++
			count := executing
			mu2.Unlock()
			fmt.Printf("  Semaphore 작업 #%d: 동시 실행=%d (최대 2)\n", id, count)
			time.Sleep(30 * time.Millisecond)
			mu2.Lock()
			executing--
			mu2.Unlock()
			sem.Release()
		}(i)
	}
	wg.Wait()
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== 동시성 패턴: 세마포어 (Semaphore) ===")

	basicSemaphoreDemo()
	weightedSemaphoreDemo()
	connectionPoolDemo()
	semaphoreVsMutex()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
