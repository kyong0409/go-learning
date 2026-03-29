// 패키지 선언
package main

// Go 동시성: sync 패키지
//
// 채널이 "통신을 통한 메모리 공유"라면,
// sync 패키지는 "전통적인 뮤텍스 기반 메모리 공유"를 제공합니다.
//
// 언제 sync를 사용하나?
// - 단순한 공유 상태 보호 (채널보다 간단할 때)
// - 캐시, 레지스트리 등의 공유 자료구조
// - 성능이 중요한 저수준 동기화

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ─────────────────────────────────────────
// 1. 레이스 컨디션(Race Condition) 데모
// ─────────────────────────────────────────

// UnsafeCounter는 동기화 없는 위험한 카운터입니다.
type UnsafeCounter struct {
	value int
}

func (c *UnsafeCounter) Increment() {
	c.value++ // 비원자적 연산: 읽기-증가-쓰기 세 단계
	// 여러 고루틴이 동시에 이 세 단계를 실행하면:
	// 고루틴 A: value 읽음(0) → 고루틴 B: value 읽음(0) →
	// 고루틴 A: 1로 씀 → 고루틴 B: 1로 씀 → 결과: 1 (2가 되어야 함!)
}

func raceConditionDemo() {
	fmt.Println("\n--- 1. 레이스 컨디션 데모 ---")
	fmt.Println("  (실제 레이스는 -race 플래그로 감지: go run -race main.go)")

	counter := &UnsafeCounter{}
	var wg sync.WaitGroup

	const goroutines = 1000
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()

	fmt.Printf("  기대값: %d, 실제값: %d\n", goroutines, counter.value)
	fmt.Println("  (실제값이 기대값보다 작을 수 있음 → 레이스 컨디션!)")
}

// ─────────────────────────────────────────
// 2. sync.Mutex: 기본 뮤텍스
// ─────────────────────────────────────────

// SafeCounter는 Mutex로 보호되는 안전한 카운터입니다.
type SafeCounter struct {
	mu    sync.Mutex // 뮤텍스: zero value로 바로 사용 가능, 복사 금지!
	value int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()   // 임계 구역(critical section) 시작: 다른 고루틴 대기
	defer c.mu.Unlock() // 임계 구역 종료: 항상 Unlock 보장 (defer 권장)
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// SafeMap은 Mutex로 보호되는 안전한 맵입니다.
type SafeMap struct {
	mu   sync.Mutex
	data map[string]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{data: make(map[string]int)}
}

func (m *SafeMap) Set(key string, value int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *SafeMap) Get(key string) (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	return v, ok
}

func mutexDemo() {
	fmt.Println("\n--- 2. sync.Mutex ---")

	// 수정된 카운터
	counter := &SafeCounter{}
	var wg sync.WaitGroup

	const goroutines = 1000
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()

	fmt.Printf("  뮤텍스 보호 카운터: 기대값=%d, 실제값=%d ✓\n",
		goroutines, counter.Value())

	// 안전한 맵
	safeMap := NewSafeMap()
	var wg2 sync.WaitGroup

	// 동시에 맵에 쓰기
	for i := 0; i < 5; i++ {
		wg2.Add(1)
		go func(n int) {
			defer wg2.Done()
			key := fmt.Sprintf("key%d", n)
			safeMap.Set(key, n*10)
		}(i)
	}
	wg2.Wait()

	fmt.Println("  안전한 맵 내용:")
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("key%d", i)
		if v, ok := safeMap.Get(key); ok {
			fmt.Printf("    %s = %d\n", key, v)
		}
	}
}

// ─────────────────────────────────────────
// 3. sync.RWMutex: 읽기/쓰기 분리 뮤텍스
// ─────────────────────────────────────────

// Cache는 RWMutex로 보호되는 읽기 많은 캐시입니다.
// RWMutex 규칙:
// - 읽기: 여러 고루틴 동시 가능 (RLock/RUnlock)
// - 쓰기: 하나만 가능, 읽기도 불가 (Lock/Unlock)
type Cache struct {
	mu    sync.RWMutex
	items map[string]string
}

func NewCache() *Cache {
	return &Cache{items: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()         // 읽기 잠금: 다른 읽기와 동시 실행 가능
	defer c.mu.RUnlock() // 읽기 잠금 해제
	v, ok := c.items[key]
	return v, ok
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()         // 쓰기 잠금: 독점적 접근
	defer c.mu.Unlock() // 쓰기 잠금 해제
	c.items[key] = value
}

func rwMutexDemo() {
	fmt.Println("\n--- 3. sync.RWMutex ---")
	fmt.Println("  읽기 > 쓰기 비율이 높은 경우 RWMutex가 더 효율적")

	cache := NewCache()
	var wg sync.WaitGroup

	// 쓰기 고루틴 2개
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("item%d", n)
			cache.Set(key, fmt.Sprintf("value%d", n))
			fmt.Printf("  쓰기: %s 저장\n", key)
		}(i)
	}

	// 읽기 고루틴 6개 (동시에 읽기 가능)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond) // 쓰기 후 읽기
			key := fmt.Sprintf("item%d", n%2)
			if v, ok := cache.Get(key); ok {
				fmt.Printf("  읽기: %s = %s\n", key, v)
			}
		}(i)
	}

	wg.Wait()
}

// ─────────────────────────────────────────
// 4. sync.Once: 한 번만 실행
// ─────────────────────────────────────────

// Singleton은 한 번만 초기화되는 싱글톤 패턴입니다.
type Singleton struct {
	value string
}

var (
	instance *Singleton
	once     sync.Once
)

// GetInstance는 싱글톤 인스턴스를 반환합니다.
// Once.Do는 프로그램 전체에서 딱 한 번만 실행됩니다.
func GetInstance() *Singleton {
	once.Do(func() {
		// 이 함수는 처음 호출될 때만 실행됨
		fmt.Println("    싱글톤 초기화 중... (딱 한 번만 출력)")
		time.Sleep(50 * time.Millisecond) // 무거운 초기화 시뮬레이션
		instance = &Singleton{value: "초기화된 값"}
	})
	return instance
}

func onceDemo() {
	fmt.Println("\n--- 4. sync.Once (싱글톤 패턴) ---")

	var wg sync.WaitGroup

	// 10개 고루틴이 동시에 인스턴스 요청
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			inst := GetInstance()
			fmt.Printf("  고루틴 #%d: 인스턴스 주소=%p, 값=%s\n",
				n, inst, inst.value)
		}(i)
	}
	wg.Wait()

	fmt.Println("  (모든 고루틴이 동일한 인스턴스를 공유)")
}

// ─────────────────────────────────────────
// 5. sync.Pool: 객체 재사용
// ─────────────────────────────────────────

// Buffer는 재사용할 버퍼입니다.
type Buffer struct {
	data []byte
	id   int
}

var bufferID int32 // atomic으로 증가

var bufferPool = sync.Pool{
	// New: 풀에 객체가 없을 때 새 객체 생성
	New: func() interface{} {
		id := int(atomic.AddInt32(&bufferID, 1))
		fmt.Printf("    새 Buffer #%d 생성\n", id)
		return &Buffer{
			data: make([]byte, 1024),
			id:   id,
		}
	},
}

func poolDemo() {
	fmt.Println("\n--- 5. sync.Pool (객체 재사용) ---")
	fmt.Println("  Pool은 GC 압력을 줄이기 위해 객체를 재사용합니다.")
	fmt.Println("  주의: GC 시 Pool은 비워질 수 있습니다 (캐시 역할).")

	var wg sync.WaitGroup

	// 5개 고루틴이 버퍼를 빌려 쓰고 반납
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			// Pool에서 객체 획득
			buf := bufferPool.Get().(*Buffer)
			fmt.Printf("  고루틴 #%d: Buffer #%d 획득\n", n, buf.id)

			// 작업 수행
			copy(buf.data, fmt.Sprintf("고루틴 #%d의 데이터", n))
			time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)

			// Pool에 객체 반납 (재사용 가능하게)
			// 반납 전에 상태를 초기화하는 것이 좋습니다
			buf.data[0] = 0
			bufferPool.Put(buf)
			fmt.Printf("  고루틴 #%d: Buffer #%d 반납\n", n, buf.id)
		}(i)
	}
	wg.Wait()

	// fmt.Scanner 등 표준 라이브러리도 sync.Pool을 사용합니다.
	fmt.Println("  (실제 사용: bytes.Buffer, fmt 내부 등)")
}

// ─────────────────────────────────────────
// 6. sync/atomic: 원자적 연산
// ─────────────────────────────────────────

func atomicDemo() {
	fmt.Println("\n--- 6. sync/atomic ---")
	fmt.Println("  atomic은 단일 메모리 위치에 대한 원자적 연산 제공")
	fmt.Println("  Mutex보다 가볍지만 복잡한 임계 구역에는 사용 불가")

	// atomic 카운터
	var counter int64
	var wg sync.WaitGroup

	const goroutines = 1000
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1) // 원자적 증가
		}()
	}
	wg.Wait()

	fmt.Printf("  atomic 카운터: 기대=%d, 실제=%d ✓\n",
		goroutines, atomic.LoadInt64(&counter))

	// atomic.Value: 임의 타입의 원자적 저장/로드
	var config atomic.Value

	// 초기 설정 저장
	type Config struct {
		Debug   bool
		MaxConn int
	}

	config.Store(Config{Debug: false, MaxConn: 100})

	// 고루틴에서 읽기
	var wg2 sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg2.Add(1)
		go func(n int) {
			defer wg2.Done()
			cfg := config.Load().(Config) // 타입 어서션
			fmt.Printf("  고루틴 #%d: Config = %+v\n", n, cfg)
		}(i)
	}

	// 설정 업데이트 (원자적으로 교체)
	config.Store(Config{Debug: true, MaxConn: 200})
	wg2.Wait()

	// CAS (Compare-And-Swap): 낙관적 잠금
	var state int64 = 0
	// state가 0이면 1로 변경 (성공하면 true 반환)
	swapped := atomic.CompareAndSwapInt64(&state, 0, 1)
	fmt.Printf("  CAS (0→1): swapped=%v, state=%d\n", swapped, state)

	// 이미 1이므로 실패
	swapped = atomic.CompareAndSwapInt64(&state, 0, 2)
	fmt.Printf("  CAS (0→2): swapped=%v, state=%d (변경 안 됨)\n", swapped, state)
}

// ─────────────────────────────────────────
// 7. sync.Map: 동시성 안전한 맵
// ─────────────────────────────────────────

func syncMapDemo() {
	fmt.Println("\n--- 7. sync.Map ---")
	fmt.Println("  sync.Map은 특정 패턴에서 Mutex+map보다 효율적:")
	fmt.Println("  - 키가 한 번 쓰이고 여러 번 읽힐 때")
	fmt.Println("  - 여러 고루틴이 서로 다른 키를 사용할 때")

	var sm sync.Map

	var wg sync.WaitGroup

	// 동시 쓰기
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			sm.Store(fmt.Sprintf("key%d", n), n*100)
		}(i)
	}
	wg.Wait()

	// 읽기
	fmt.Println("  저장된 값:")
	sm.Range(func(key, value interface{}) bool {
		fmt.Printf("    %v = %v\n", key, value)
		return true // false를 반환하면 순회 중단
	})

	// LoadOrStore: 없으면 저장, 있으면 기존 값 반환
	actual, loaded := sm.LoadOrStore("key0", 9999)
	fmt.Printf("  LoadOrStore key0: actual=%v, loaded=%v\n", actual, loaded)

	actual, loaded = sm.LoadOrStore("new_key", 42)
	fmt.Printf("  LoadOrStore new_key: actual=%v, loaded=%v\n", actual, loaded)
}

// ─────────────────────────────────────────
// 8. sync.Cond: 조건 변수
// ─────────────────────────────────────────

// Queue는 조건 변수를 사용하는 생산자-소비자 큐입니다.
type Queue struct {
	mu    sync.Mutex
	cond  *sync.Cond
	items []int
	maxSize int
}

func NewQueue(maxSize int) *Queue {
	q := &Queue{maxSize: maxSize}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Push(item int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 큐가 가득 차면 대기
	for len(q.items) >= q.maxSize {
		q.cond.Wait() // 잠금 해제 후 대기, 신호 받으면 잠금 재획득
	}
	q.items = append(q.items, item)
	q.cond.Broadcast() // 대기 중인 모든 고루틴에게 신호
}

func (q *Queue) Pop() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 큐가 비어있으면 대기
	for len(q.items) == 0 {
		q.cond.Wait()
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.cond.Broadcast() // 대기 중인 생산자에게 신호
	return item
}

func condDemo() {
	fmt.Println("\n--- 8. sync.Cond (조건 변수) ---")

	queue := NewQueue(3) // 최대 3개
	var wg sync.WaitGroup

	// 생산자
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 6; i++ {
			queue.Push(i)
			fmt.Printf("  생산자: %d 추가\n", i)
		}
	}()

	// 소비자
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 6; i++ {
			time.Sleep(30 * time.Millisecond)
			item := queue.Pop()
			fmt.Printf("  소비자: %d 소비\n", item)
		}
	}()

	wg.Wait()
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== Go 동시성: sync 패키지 ===")
	fmt.Printf("CPU 코어 수: %d\n", runtime.NumCPU())

	raceConditionDemo()
	mutexDemo()
	rwMutexDemo()
	onceDemo()
	poolDemo()
	atomicDemo()
	syncMapDemo()
	condDemo()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
