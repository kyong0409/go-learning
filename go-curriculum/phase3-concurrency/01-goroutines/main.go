// 패키지 선언
package main

// Go의 동시성 모델: 고루틴(Goroutine)
//
// 고루틴은 Go 런타임이 관리하는 경량 스레드입니다.
// OS 스레드와 달리 초기 스택 크기가 약 2KB로 매우 작고,
// 필요에 따라 자동으로 크기가 조절됩니다.
// 수천~수백만 개의 고루틴을 동시에 실행할 수 있습니다.

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 헬퍼 함수들
// ─────────────────────────────────────────

// printMessage는 메시지를 출력하고 잠시 대기하는 함수입니다.
func printMessage(id int, msg string) {
	fmt.Printf("  고루틴 #%d: %s\n", id, msg)
	time.Sleep(10 * time.Millisecond)
}

// heavyWork는 계산 집약적인 작업을 시뮬레이션합니다.
func heavyWork(id int, wg *sync.WaitGroup) {
	defer wg.Done() // 함수 종료 시 WaitGroup 카운터 감소

	// 간단한 계산 작업 시뮬레이션
	sum := 0
	for i := 0; i < 1000; i++ {
		sum += i
	}
	fmt.Printf("  워커 #%d 완료: 합계=%d\n", id, sum)
}

// ─────────────────────────────────────────
// 나쁜 방법: time.Sleep으로 고루틴 대기
// ─────────────────────────────────────────

// badWaitWithSleep은 time.Sleep으로 고루틴을 기다리는 안티패턴입니다.
// 문제점:
// 1. 고루틴이 얼마나 걸릴지 모름 → Sleep 시간 결정 어려움
// 2. 너무 짧으면 고루틴이 끝나기 전에 main이 종료됨
// 3. 너무 길면 불필요한 대기 발생
func badWaitWithSleep() {
	fmt.Println("\n--- 나쁜 방법: time.Sleep 사용 ---")

	for i := 1; i <= 3; i++ {
		go printMessage(i, "time.Sleep으로 기다리는 중...") // go 키워드로 고루틴 시작
	}

	// 문제: 모든 고루틴이 끝났는지 보장할 수 없습니다!
	// 고루틴이 늦게 시작되거나 지연되면 출력이 누락될 수 있습니다.
	time.Sleep(200 * time.Millisecond) // 임의의 대기 시간 ← 안티패턴!
	fmt.Println("  (경고: 이 방식은 프로덕션 코드에서 사용하면 안 됩니다)")
}

// ─────────────────────────────────────────
// 좋은 방법: sync.WaitGroup으로 고루틴 대기
// ─────────────────────────────────────────

// goodWaitWithWaitGroup은 WaitGroup을 사용한 올바른 패턴입니다.
// WaitGroup 동작 원리:
// - Add(n): 카운터를 n만큼 증가
// - Done(): 카운터를 1 감소 (고루틴 완료 시 호출)
// - Wait(): 카운터가 0이 될 때까지 블로킹
func goodWaitWithWaitGroup() {
	fmt.Println("\n--- 좋은 방법: sync.WaitGroup 사용 ---")

	var wg sync.WaitGroup // WaitGroup 선언 (zero value로 바로 사용 가능)

	for i := 1; i <= 5; i++ {
		wg.Add(1) // 고루틴 시작 전에 카운터 증가 (고루틴 안에서 하면 레이스 컨디션!)
		go func(id int) {
			defer wg.Done() // 반드시 Done 호출 보장 (defer 사용 권장)
			printMessage(id, "WaitGroup으로 동기화")
		}(i) // 루프 변수 캡처 문제 방지: i를 인자로 전달
	}

	wg.Wait() // 모든 고루틴이 Done을 호출할 때까지 대기
	fmt.Println("  모든 고루틴 완료!")
}

// ─────────────────────────────────────────
// 클로저와 고루틴: 변수 캡처 주의사항
// ─────────────────────────────────────────

// closureGoroutinePitfall은 클로저 변수 캡처 문제를 보여줍니다.
func closureGoroutinePitfall() {
	fmt.Println("\n--- 클로저 캡처 문제 ---")

	var wg sync.WaitGroup

	// 잘못된 예: 루프 변수 i를 직접 참조
	fmt.Println("  [잘못된 예] 모든 고루틴이 같은 i를 참조할 수 있음:")
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// i는 루프가 끝난 후의 값(4)일 수 있음!
			// Go 1.22+부터는 루프 변수가 매 반복마다 새로 생성되어 이 문제가 해결됨
			// 하지만 명시적 전달이 더 안전하고 명확합니다.
			fmt.Printf("    잘못된 예 - i 값: %d\n", i)
		}()
	}
	wg.Wait()

	fmt.Println("  [올바른 예] 인자로 명시적 전달:")
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(num int) { // i를 num으로 복사해서 전달
			defer wg.Done()
			fmt.Printf("    올바른 예 - num 값: %d\n", num)
		}(i) // 현재 i 값을 즉시 전달
	}
	wg.Wait()
}

// ─────────────────────────────────────────
// 고루틴 수천 개 실행: 경량성 데모
// ─────────────────────────────────────────

// massiveGoroutineDemo는 고루틴의 경량성을 보여줍니다.
// OS 스레드는 수천 개 생성 시 메모리/스케줄링 오버헤드가 크지만,
// 고루틴은 수십만 개도 가볍게 실행 가능합니다.
func massiveGoroutineDemo() {
	fmt.Println("\n--- 고루틴 경량성 데모: 10만 개 고루틴 ---")

	const numGoroutines = 100_000

	var wg sync.WaitGroup

	// 메모리 사용량 측정 시작
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	startTime := time.Now()

	// 10만 개의 고루틴 생성
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 아주 간단한 작업: 카운터 증가
			// (실제로는 채널이나 atomic을 써야 하지만 여기서는 생략)
			_ = 1 + 1
		}()
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	// 메모리 사용량 측정 종료
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	fmt.Printf("  실행한 고루틴 수: %d개\n", numGoroutines)
	fmt.Printf("  총 소요 시간: %v\n", elapsed)
	fmt.Printf("  힙 할당 증가: %d KB\n", (memAfter.TotalAlloc-memBefore.TotalAlloc)/1024)
	fmt.Printf("  현재 활성 고루틴 수: %d\n", runtime.NumGoroutine())
	fmt.Printf("  사용 가능한 CPU 코어: %d\n", runtime.NumCPU())
}

// ─────────────────────────────────────────
// GOMAXPROCS: 병렬 실행 스레드 수 제어
// ─────────────────────────────────────────

// gomaxprocsDemo는 GOMAXPROCS 설정을 보여줍니다.
func gomaxprocsDemo() {
	fmt.Println("\n--- GOMAXPROCS 설정 ---")

	// GOMAXPROCS: Go 런타임이 사용할 OS 스레드 수
	// 기본값: runtime.NumCPU() (CPU 코어 수)
	current := runtime.GOMAXPROCS(0) // 0을 전달하면 현재 값만 반환 (변경 안 함)
	fmt.Printf("  현재 GOMAXPROCS: %d\n", current)
	fmt.Printf("  CPU 코어 수: %d\n", runtime.NumCPU())

	// 실제 병렬 실행: GOMAXPROCS가 1이면 동시성(concurrency)만,
	// 2 이상이면 병렬성(parallelism)도 가능합니다.
	fmt.Println("  동시성(Concurrency) vs 병렬성(Parallelism):")
	fmt.Println("  - 동시성: 여러 작업을 번갈아가며 실행 (1개 코어에서도 가능)")
	fmt.Println("  - 병렬성: 여러 작업을 동시에 실행 (여러 코어 필요)")
}

// ─────────────────────────────────────────
// 고루틴 누수(Goroutine Leak) 예방
// ─────────────────────────────────────────

// goroutineLeakExample은 고루틴 누수 문제와 해결책을 보여줍니다.
func goroutineLeakExample() {
	fmt.Println("\n--- 고루틴 누수 예방 ---")

	// 고루틴 누수: 고루틴이 종료되지 않고 계속 살아있는 상태
	// 주요 원인:
	// 1. 아무도 받지 않는 채널에 보내려고 블로킹
	// 2. 아무도 보내지 않는 채널에서 받으려고 블로킹
	// 3. 잠금이 해제되지 않아 영원히 대기

	fmt.Printf("  데모 시작 전 고루틴 수: %d\n", runtime.NumGoroutine())

	// done 채널을 이용한 안전한 종료 패턴
	done := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-done: // done 채널이 닫히면 고루틴 종료
			fmt.Println("  고루틴: done 신호 받아 종료")
		case <-time.After(5 * time.Second): // 타임아웃 방어
			fmt.Println("  고루틴: 타임아웃으로 종료")
		}
	}()

	fmt.Printf("  고루틴 실행 중 고루틴 수: %d\n", runtime.NumGoroutine())

	// done 채널을 닫아 고루틴에게 종료 신호 전송
	close(done)
	wg.Wait()

	fmt.Printf("  고루틴 종료 후 고루틴 수: %d\n", runtime.NumGoroutine())
	fmt.Println("  (done 채널 패턴으로 고루틴 누수 방지!)")
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== Go 동시성: 고루틴(Goroutine) ===")

	// 1. 나쁜 방법
	badWaitWithSleep()

	// 2. 좋은 방법
	goodWaitWithWaitGroup()

	// 3. 클로저 캡처 주의사항
	closureGoroutinePitfall()

	// 4. 경량성 데모
	massiveGoroutineDemo()

	// 5. GOMAXPROCS 정보
	gomaxprocsDemo()

	// 6. 고루틴 누수 예방
	goroutineLeakExample()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
