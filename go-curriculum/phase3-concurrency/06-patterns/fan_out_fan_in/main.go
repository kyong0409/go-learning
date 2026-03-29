// 패키지 선언
package main

// 동시성 패턴: 팬아웃/팬인 (Fan-Out / Fan-In)
//
// 팬아웃 (Fan-Out): 하나의 입력 채널에서 여러 워커로 작업 분배
// 팬인  (Fan-In):  여러 채널의 결과를 하나의 채널로 합침
//
// 사용 사례:
// - 여러 API를 병렬 호출 후 결과 수집
// - 대용량 데이터를 여러 프로세서로 병렬 처리
// - 검색 엔진의 병렬 인덱스 조회

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 기본 타입
// ─────────────────────────────────────────

// SearchResult는 검색 결과를 나타냅니다.
type SearchResult struct {
	Source  string
	Query   string
	Results []string
	Latency time.Duration
}

// ─────────────────────────────────────────
// 1. 기본 팬아웃/팬인
// ─────────────────────────────────────────

// fanOut은 입력 채널을 n개의 출력 채널로 분배합니다.
func fanOut(input <-chan int, n int) []<-chan int {
	outputs := make([]<-chan int, n)
	for i := 0; i < n; i++ {
		out := make(chan int, 10)
		outputs[i] = out
		go func(ch chan<- int) {
			defer close(ch)
			for v := range input {
				ch <- v * v // 각 워커가 제곱 계산
				time.Sleep(time.Duration(rand.Intn(30)) * time.Millisecond)
			}
		}(out)
	}
	return outputs
}

// fanIn은 여러 채널을 하나로 합칩니다.
func fanIn(channels ...<-chan int) <-chan int {
	merged := make(chan int, 20)
	var wg sync.WaitGroup

	// 각 입력 채널에 대해 고루틴 생성
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				merged <- v
			}
		}(ch)
	}

	// 모든 입력 채널이 닫히면 merged도 닫기
	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func basicFanOutFanIn() {
	fmt.Println("\n--- 1. 기본 팬아웃/팬인 ---")

	// 입력: 1~12
	input := make(chan int, 12)
	for i := 1; i <= 12; i++ {
		input <- i
	}
	close(input)

	// 팬아웃: 4개 워커로 분배
	fmt.Println("  팬아웃: 12개 숫자를 4개 워커로 분배")
	workers := fanOut(input, 4)

	// 팬인: 결과 하나로 합치기
	merged := fanIn(workers...)

	// 결과 수집
	var results []int
	for v := range merged {
		results = append(results, v)
	}

	fmt.Printf("  팬인 결과 (%d개): %v\n", len(results), results)
}

// ─────────────────────────────────────────
// 2. Context를 지원하는 팬인
// ─────────────────────────────────────────

// fanInWithContext는 Context 취소를 지원하는 팬인입니다.
func fanInWithContext(ctx context.Context, channels ...<-chan string) <-chan string {
	merged := make(chan string, len(channels)*2)
	var wg sync.WaitGroup

	output := func(c <-chan string) {
		defer wg.Done()
		for {
			select {
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case merged <- v:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}

	for _, c := range channels {
		wg.Add(1)
		go output(c)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

// ─────────────────────────────────────────
// 3. 병렬 API 호출 패턴 (실제 사용 사례)
// ─────────────────────────────────────────

// searchEngine은 특정 검색 엔진에 쿼리를 보냅니다.
func searchEngine(ctx context.Context, name, query string) <-chan SearchResult {
	ch := make(chan SearchResult, 1)
	go func() {
		defer close(ch)

		// 각 검색 엔진마다 다른 응답 시간
		latencies := map[string]int{
			"Google": 100,
			"Bing":   150,
			"Naver":  80,
			"Kakao":  120,
		}
		baseLatency := latencies[name]
		delay := time.Duration(baseLatency+rand.Intn(50)) * time.Millisecond

		select {
		case <-time.After(delay):
			result := SearchResult{
				Source:  name,
				Query:   query,
				Latency: delay,
				Results: []string{
					fmt.Sprintf("[%s] %s 결과 1", name, query),
					fmt.Sprintf("[%s] %s 결과 2", name, query),
					fmt.Sprintf("[%s] %s 결과 3", name, query),
				},
			}
			select {
			case ch <- result:
			case <-ctx.Done():
			}
		case <-ctx.Done():
			// 취소 시 에러 없이 종료 (채널 닫힘)
		}
	}()
	return ch
}

// parallelSearch는 여러 검색 엔진에 동시에 검색을 보냅니다.
func parallelSearch(ctx context.Context, query string, timeout time.Duration) []SearchResult {
	// 타임아웃 설정
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	engines := []string{"Google", "Bing", "Naver", "Kakao"}

	// 팬아웃: 각 검색 엔진에 동시 요청
	channels := make([]<-chan SearchResult, len(engines))
	for i, engine := range engines {
		channels[i] = searchEngine(ctx, engine, query)
	}

	// 팬인: 결과 수집
	var results []SearchResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan SearchResult) {
			defer wg.Done()
			select {
			case result, ok := <-c:
				if ok {
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
				}
			case <-ctx.Done():
			}
		}(ch)
	}

	wg.Wait()
	return results
}

func parallelAPIDemo() {
	fmt.Println("\n--- 2. 병렬 API 호출 (검색 엔진) ---")

	ctx := context.Background()

	// 200ms 타임아웃으로 검색
	fmt.Println("  쿼리: 'Go 동시성 프로그래밍' (타임아웃: 200ms)")
	start := time.Now()
	results := parallelSearch(ctx, "Go 동시성 프로그래밍", 200*time.Millisecond)
	elapsed := time.Since(start)

	fmt.Printf("  %d개 검색 엔진 응답 (%v 소요):\n", len(results), elapsed.Round(time.Millisecond))
	for _, r := range results {
		fmt.Printf("  [%s] %v 지연: %v\n", r.Source, r.Results[0], r.Latency.Round(time.Millisecond))
	}
}

// ─────────────────────────────────────────
// 4. 첫 번째 응답 사용 패턴 (Race)
// ─────────────────────────────────────────

// getFromReplica는 데이터베이스 복제본에서 데이터를 가져옵니다.
func getFromReplica(ctx context.Context, replicaID int) <-chan string {
	ch := make(chan string, 1)
	go func() {
		delay := time.Duration(50+rand.Intn(200)) * time.Millisecond
		select {
		case <-time.After(delay):
			select {
			case ch <- fmt.Sprintf("복제본 #%d 응답 (%v)", replicaID, delay.Round(time.Millisecond)):
			case <-ctx.Done():
			}
		case <-ctx.Done():
		}
		close(ch)
	}()
	return ch
}

// fastest는 여러 소스 중 가장 빠른 응답을 반환합니다.
// 헤징 요청(Hedged Request) 패턴이라고도 합니다.
func fastest(ctx context.Context, n int) string {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // 첫 번째 응답 후 나머지 요청 취소

	results := make(chan string, n)

	for i := 1; i <= n; i++ {
		go func(id int) {
			ch := getFromReplica(ctx, id)
			select {
			case v, ok := <-ch:
				if ok {
					results <- v
				}
			case <-ctx.Done():
			}
		}(i)
	}

	// 첫 번째 응답 반환
	return <-results
}

func firstResponseDemo() {
	fmt.Println("\n--- 3. 첫 번째 응답 사용 패턴 (헤징 요청) ---")
	fmt.Println("  3개 복제본에 동시 요청, 가장 빠른 응답 사용:")

	for i := 0; i < 3; i++ {
		start := time.Now()
		result := fastest(context.Background(), 3)
		fmt.Printf("  시도 #%d: %s (%.0fms)\n",
			i+1, result, float64(time.Since(start).Milliseconds()))
	}
}

// ─────────────────────────────────────────
// 5. 동적 팬인: 런타임에 채널 추가/제거
// ─────────────────────────────────────────

// Merger는 동적으로 채널을 추가할 수 있는 팬인입니다.
type Merger struct {
	mu      sync.Mutex
	output  chan int
	done    chan struct{}
	wg      sync.WaitGroup
}

// NewMerger는 새 Merger를 생성합니다.
func NewMerger() *Merger {
	m := &Merger{
		output: make(chan int, 100),
		done:   make(chan struct{}),
	}
	return m
}

// Add는 새 입력 채널을 추가합니다.
func (m *Merger) Add(input <-chan int) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case v, ok := <-input:
				if !ok {
					return
				}
				select {
				case m.output <- v:
				case <-m.done:
					return
				}
			case <-m.done:
				return
			}
		}
	}()
}

// Output은 병합된 출력 채널을 반환합니다.
func (m *Merger) Output() <-chan int {
	return m.output
}

// Close는 Merger를 종료합니다.
func (m *Merger) Close() {
	close(m.done)
	m.wg.Wait()
	close(m.output)
}

func dynamicMerger() {
	fmt.Println("\n--- 4. 동적 팬인 ---")

	merger := NewMerger()

	// 3개 채널을 동적으로 추가
	for i := 1; i <= 3; i++ {
		ch := make(chan int, 3)
		for j := 0; j < 3; j++ {
			ch <- i*10 + j
		}
		close(ch)
		merger.Add(ch)
	}

	// 수집
	go func() {
		time.Sleep(100 * time.Millisecond)
		merger.Close()
	}()

	var values []int
	for v := range merger.Output() {
		values = append(values, v)
	}
	fmt.Printf("  병합된 값 (%d개): %v\n", len(values), values)
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== 동시성 패턴: 팬아웃/팬인 (Fan-Out/Fan-In) ===")

	basicFanOutFanIn()
	parallelAPIDemo()
	firstResponseDemo()
	dynamicMerger()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
