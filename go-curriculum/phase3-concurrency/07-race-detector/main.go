// 패키지 선언
package main

// 레이스 컨디션 (Race Condition) 데모
//
// 레이스 컨디션: 두 개 이상의 고루틴이 동기화 없이
// 같은 메모리 위치에 동시에 접근하고, 그 중 하나 이상이 쓰기인 경우
//
// 문제점:
// - 결과가 실행 순서에 따라 달라짐 (비결정적)
// - 디버깅이 매우 어려움 (가끔만 발생)
// - 메모리 손상, 크래시, 잘못된 결과 유발
//
// 탐지 방법:
// go run -race main.go
// go test -race ./...
// go build -race -o myapp

import (
	"fmt"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 레이스 컨디션 예제 1: 카운터
// ─────────────────────────────────────────

// racyCounter는 보호되지 않은 카운터 (레이스 컨디션 발생)
var racyCounter int

func incrementRacy(wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 1000; i++ {
		racyCounter++ // 레이스: 읽기-증가-쓰기가 원자적이지 않음
	}
}

// ─────────────────────────────────────────
// 레이스 컨디션 예제 2: 슬라이스 추가
// ─────────────────────────────────────────

// racySlice는 보호되지 않은 슬라이스 추가
var racySlice []int

func appendRacy(wg *sync.WaitGroup, val int) {
	defer wg.Done()
	racySlice = append(racySlice, val) // 레이스: append는 스레드 안전하지 않음
}

// ─────────────────────────────────────────
// 레이스 컨디션 예제 3: 클로저 루프 캡처
// ─────────────────────────────────────────

func closureRace() []int {
	results := make([]int, 5)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { // i를 직접 캡처 → 레이스!
			defer wg.Done()
			results[i] = i * i // i 값이 불확실
		}()
	}

	wg.Wait()
	return results
}

// ─────────────────────────────────────────
// 레이스 컨디션 예제 4: 맵 동시 접근
// ─────────────────────────────────────────

var racyMap = make(map[int]int)

func writeMapRacy(wg *sync.WaitGroup, key, val int) {
	defer wg.Done()
	racyMap[key] = val // 레이스: 맵은 동시 쓰기에 안전하지 않음
}

// ─────────────────────────────────────────
// main 함수: 레이스 컨디션 유발
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== 레이스 컨디션 데모 (의도적) ===")
	fmt.Println()
	fmt.Println("이 파일은 -race 플래그로 실행하세요:")
	fmt.Println("  go run -race main.go")
	fmt.Println()
	fmt.Println("레이스 디텍터가 활성화되면 레이스 컨디션 감지 시")
	fmt.Println("다음과 같은 출력이 표시됩니다:")
	fmt.Println("  WARNING: DATA RACE")
	fmt.Println("  Read at 0x... by goroutine N:")
	fmt.Println("  Write at 0x... by goroutine M:")
	fmt.Println()

	// 예제 1: 카운터 레이스
	fmt.Println("--- 예제 1: 카운터 레이스 컨디션 ---")
	racyCounter = 0
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go incrementRacy(&wg)
	}
	wg.Wait()

	fmt.Printf("기대값: 5000, 실제값: %d\n", racyCounter)
	fmt.Println("(레이스 컨디션으로 인해 결과가 5000 미만일 수 있음)")
	fmt.Println()

	// 예제 2: 슬라이스 레이스
	fmt.Println("--- 예제 2: 슬라이스 append 레이스 ---")
	racySlice = nil

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go appendRacy(&wg, i)
	}
	wg.Wait()

	fmt.Printf("추가한 원소 수: 10, 실제 슬라이스 길이: %d\n", len(racySlice))
	fmt.Println("(append 레이스로 인해 일부 원소가 유실될 수 있음)")
	fmt.Println()

	// 예제 3: 맵 레이스 (Go 런타임이 즉시 패닉시킴)
	fmt.Println("--- 예제 3: 맵 동시 접근 레이스 ---")
	fmt.Println("(맵 동시 쓰기는 Go 런타임이 감지해 패닉 발생 가능)")
	fmt.Println("(아래 코드는 실제 레이스를 유발하지만 데모를 위해 짧게 실행)")

	// 짧은 시간만 실행 (패닉 방지용으로 recover 사용)
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("패닉 발생 (예상됨): %v\n", r)
			}
		}()

		for i := 0; i < 3; i++ {
			wg.Add(1)
			go writeMapRacy(&wg, i, i*10)
		}
		// 짧은 실행 후 대기
		time.Sleep(5 * time.Millisecond)
	}()

	wg.Wait()
	fmt.Println()
	fmt.Println("=== 수정된 버전은 main_fixed.go를 참고하세요 ===")
}
