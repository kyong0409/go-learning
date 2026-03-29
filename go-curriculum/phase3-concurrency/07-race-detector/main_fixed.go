//go:build ignore

// 패키지 선언
package main

// 레이스 컨디션 수정 버전
//
// 각 레이스 컨디션을 올바른 동기화 기법으로 수정했습니다.
//
// 실행 방법:
//   go run main_fixed.go
//   go run -race main_fixed.go   ← 레이스 없음 확인

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ─────────────────────────────────────────
// 수정 1: atomic으로 카운터 보호
// ─────────────────────────────────────────

var fixedCounter int64 // int64로 변경 (atomic 패키지 요구사항)

func incrementFixed(wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 1000; i++ {
		atomic.AddInt64(&fixedCounter, 1) // 원자적 증가
	}
}

// ─────────────────────────────────────────
// 수정 2: Mutex로 슬라이스 보호
// ─────────────────────────────────────────

var (
	fixedSlice []int
	sliceMu    sync.Mutex
)

func appendFixed(wg *sync.WaitGroup, val int) {
	defer wg.Done()
	sliceMu.Lock()
	fixedSlice = append(fixedSlice, val)
	sliceMu.Unlock()
}

// ─────────────────────────────────────────
// 수정 3: 루프 변수 명시적 전달
// ─────────────────────────────────────────

func closureFixed() []int {
	results := make([]int, 5)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) { // i를 idx로 복사해서 전달
			defer wg.Done()
			results[idx] = idx * idx
		}(i)
	}

	wg.Wait()
	return results
}

// ─────────────────────────────────────────
// 수정 4: sync.Map으로 맵 보호
// ─────────────────────────────────────────

var fixedMap sync.Map

func writeMapFixed(wg *sync.WaitGroup, key, val int) {
	defer wg.Done()
	fixedMap.Store(key, val) // sync.Map: 동시 접근 안전
}

// ─────────────────────────────────────────
// 수정 5: 채널로 결과 수집
// ─────────────────────────────────────────

func sumWithChannel(nums []int) int {
	ch := make(chan int, len(nums))
	var wg sync.WaitGroup

	for _, n := range nums {
		wg.Add(1)
		n := n
		go func() {
			defer wg.Done()
			ch <- n * n
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	total := 0
	for v := range ch {
		total += v
	}
	return total
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== 레이스 컨디션 수정 버전 ===")
	fmt.Println("go run -race main_fixed.go 로 확인하세요 (경고 없음)")
	fmt.Println()

	// 수정 1: atomic 카운터
	fmt.Println("--- 수정 1: atomic 카운터 ---")
	fixedCounter = 0
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go incrementFixed(&wg)
	}
	wg.Wait()
	fmt.Printf("기대값: 5000, 실제값: %d ✓\n", atomic.LoadInt64(&fixedCounter))
	fmt.Println()

	// 수정 2: Mutex 슬라이스
	fmt.Println("--- 수정 2: Mutex 슬라이스 ---")
	fixedSlice = nil
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go appendFixed(&wg, i)
	}
	wg.Wait()
	fmt.Printf("추가한 원소 수: 10, 실제 슬라이스 길이: %d ✓\n", len(fixedSlice))
	fmt.Println()

	// 수정 3: 클로저 변수 전달
	fmt.Println("--- 수정 3: 클로저 변수 명시적 전달 ---")
	results := closureFixed()
	fmt.Printf("결과: %v ✓\n", results)
	fmt.Println()

	// 수정 4: sync.Map
	fmt.Println("--- 수정 4: sync.Map ---")
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go writeMapFixed(&wg, i, i*10)
	}
	wg.Wait()
	fmt.Println("맵 내용:")
	fixedMap.Range(func(k, v interface{}) bool {
		fmt.Printf("  %v: %v\n", k, v)
		return true
	})
	fmt.Println()

	// 수정 5: 채널 기반 집계
	fmt.Println("--- 수정 5: 채널로 결과 수집 ---")
	total := sumWithChannel([]int{1, 2, 3, 4, 5})
	fmt.Printf("1²+2²+3²+4²+5² = %d ✓\n", total) // 55
	fmt.Println()

	fmt.Println("=== 모든 레이스 컨디션 수정 완료 ===")
	fmt.Println()
	fmt.Println("동기화 기법 선택 가이드:")
	fmt.Println("  단순 카운터/플래그    → sync/atomic")
	fmt.Println("  공유 자료구조 보호    → sync.Mutex / sync.RWMutex")
	fmt.Println("  한 번만 초기화        → sync.Once")
	fmt.Println("  맵 동시 접근          → sync.Map")
	fmt.Println("  고루틴 간 데이터 전달 → channel")
	fmt.Println("  결과 수집/집계        → channel + WaitGroup")
}
