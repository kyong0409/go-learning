// 패키지 선언
package main

// Go의 동시성 모델: 채널(Channel)
//
// "메모리를 공유해서 통신하지 말고, 통신해서 메모리를 공유하라."
// - Rob Pike (Go 언어 설계자)
//
// 채널은 고루틴 간 데이터를 안전하게 주고받는 파이프입니다.
// 내부적으로 뮤텍스와 큐로 구현되어 있어 스레드 안전합니다.

import (
	"fmt"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 1. 기본 채널: 생성, 송신, 수신
// ─────────────────────────────────────────

func basicChannel() {
	fmt.Println("\n--- 1. 기본 채널 ---")

	// make(chan T): T 타입의 언버퍼 채널 생성
	// 언버퍼 채널: 송신자와 수신자가 동시에 준비되어야 전달 가능 (동기 방식)
	ch := make(chan int)

	// 고루틴에서 채널로 값 송신
	go func() {
		fmt.Println("  고루틴: 값 42 송신 중...")
		ch <- 42 // 채널에 값 전송 (수신자가 받을 때까지 블로킹)
		fmt.Println("  고루틴: 송신 완료")
	}()

	// 메인 고루틴에서 수신
	val := <-ch // 채널에서 값 수신 (송신자가 보낼 때까지 블로킹)
	fmt.Printf("  메인: 수신한 값 = %d\n", val)

	// 채널로 여러 값 주고받기
	strCh := make(chan string)
	go func() {
		messages := []string{"첫 번째", "두 번째", "세 번째"}
		for _, msg := range messages {
			strCh <- msg
		}
	}()

	for i := 0; i < 3; i++ {
		fmt.Printf("  수신: %s\n", <-strCh)
	}
}

// ─────────────────────────────────────────
// 2. 버퍼 채널 vs 언버퍼 채널
// ─────────────────────────────────────────

func bufferedVsUnbuffered() {
	fmt.Println("\n--- 2. 버퍼 채널 vs 언버퍼 채널 ---")

	// 언버퍼 채널: 동기 방식 - 송신과 수신이 동시에 이루어짐
	fmt.Println("  [언버퍼 채널] - 동기 방식:")
	unbuffered := make(chan int) // 버퍼 없음
	go func() {
		fmt.Println("    송신자: 채널에 전송 시도...")
		unbuffered <- 1 // 수신자가 준비될 때까지 블로킹
		fmt.Println("    송신자: 전송 완료 (수신자가 받았음)")
	}()
	time.Sleep(50 * time.Millisecond) // 송신자가 먼저 블로킹되게 잠시 대기
	fmt.Println("    수신자: 이제 수신합니다...")
	<-unbuffered
	fmt.Println("    수신자: 수신 완료")

	// 버퍼 채널: 비동기 방식 - 버퍼가 찰 때까지 블로킹 없이 송신 가능
	fmt.Println("\n  [버퍼 채널] - 비동기 방식 (버퍼 크기: 3):")
	buffered := make(chan int, 3) // 버퍼 크기 3

	// 수신자 없이도 버퍼가 찰 때까지 즉시 송신 가능
	fmt.Println("    수신자 없이 3개 연속 송신:")
	buffered <- 10 // 즉시 반환 (버퍼에 저장)
	buffered <- 20 // 즉시 반환
	buffered <- 30 // 즉시 반환
	fmt.Printf("    버퍼 길이: %d, 용량: %d\n", len(buffered), cap(buffered))

	// 버퍼가 가득 차면 블로킹 (4번째 송신 시도 시 블로킹)
	// buffered <- 40 // 이 줄의 주석을 풀면 데드락 발생!

	// 수신
	fmt.Println("    버퍼에서 수신:")
	for i := 0; i < 3; i++ {
		fmt.Printf("    수신: %d\n", <-buffered)
	}

	// 버퍼 채널 특성 요약
	fmt.Println("\n  버퍼 채널 특성:")
	fmt.Println("  - 버퍼가 비어있을 때 수신: 블로킹")
	fmt.Println("  - 버퍼가 가득 찰 때 송신: 블로킹")
	fmt.Println("  - 그 외: 즉시 반환")
}

// ─────────────────────────────────────────
// 3. 방향성 채널 (Directional Channels)
// ─────────────────────────────────────────

// sendOnly는 송신 전용 채널만 받습니다.
// chan<- int: 오직 송신만 가능 (← 화살표가 chan 앞에 없음)
func sendOnly(ch chan<- int, values []int) {
	for _, v := range values {
		ch <- v
	}
	close(ch) // 송신 전용 채널도 닫을 수 있음
}

// receiveOnly는 수신 전용 채널만 받습니다.
// <-chan int: 오직 수신만 가능 (← 화살표가 chan 앞에 있음)
func receiveOnly(ch <-chan int) int {
	sum := 0
	for v := range ch { // 채널이 닫힐 때까지 수신
		sum += v
	}
	return sum
}

// producer는 데이터를 생성해서 채널에 보냅니다.
// 반환 타입을 수신 전용으로 제한하면 호출자가 실수로 송신하지 못합니다.
func producer(nums []int) <-chan int {
	ch := make(chan int, len(nums))
	go func() {
		for _, n := range nums {
			ch <- n
		}
		close(ch)
	}()
	return ch // <-chan int로 자동 변환
}

func directionalChannels() {
	fmt.Println("\n--- 3. 방향성 채널 ---")

	// 양방향 채널을 방향성 채널 함수에 전달
	ch := make(chan int, 5)
	go sendOnly(ch, []int{1, 2, 3, 4, 5})
	sum := receiveOnly(ch)
	fmt.Printf("  합계: %d\n", sum)

	// producer 패턴
	resultCh := producer([]int{10, 20, 30, 40, 50})
	total := 0
	for v := range resultCh {
		total += v
	}
	fmt.Printf("  producer 합계: %d\n", total)

	fmt.Println("  방향성 채널의 장점:")
	fmt.Println("  - 컴파일 타임에 잘못된 방향의 연산 방지")
	fmt.Println("  - API 설계 시 의도를 명확히 표현")
}

// ─────────────────────────────────────────
// 4. 채널 닫기와 range over channel
// ─────────────────────────────────────────

func closeAndRange() {
	fmt.Println("\n--- 4. 채널 닫기와 range ---")

	ch := make(chan int, 5)

	// 생성자 고루틴
	go func() {
		for i := 1; i <= 5; i++ {
			ch <- i * i // 제곱수 전송
			time.Sleep(10 * time.Millisecond)
		}
		close(ch) // 더 이상 보낼 데이터 없음 → 채널 닫기
		// 주의: 닫힌 채널에 송신하면 패닉 발생!
		// 채널은 송신 측에서만 닫아야 합니다.
	}()

	// range로 채널에서 수신: 채널이 닫히고 버퍼가 비워지면 자동 종료
	fmt.Println("  range로 수신:")
	for val := range ch {
		fmt.Printf("    수신: %d\n", val)
	}

	// 수동으로 채널 닫힘 확인
	ch2 := make(chan string, 2)
	ch2 <- "hello"
	ch2 <- "world"
	close(ch2)

	// ok 패턴: 두 번째 반환값으로 채널 상태 확인
	for {
		val, ok := <-ch2
		if !ok {
			fmt.Println("  채널이 닫혔고 비어있음 (ok=false)")
			break
		}
		fmt.Printf("  ok 패턴 수신: %s (ok=%v)\n", val, ok)
	}

	// 닫힌 채널에서 수신: 버퍼가 비어있으면 zero value와 ok=false 반환
	closedCh := make(chan int)
	close(closedCh)
	v, ok := <-closedCh
	fmt.Printf("  닫힌 채널 수신: v=%d, ok=%v (zero value!)\n", v, ok)
}

// ─────────────────────────────────────────
// 5. done 채널 패턴
// ─────────────────────────────────────────

// generateNumbers는 숫자를 생성하는 고루틴입니다.
// done 채널로 취소 신호를 받으면 종료합니다.
func generateNumbers(done <-chan struct{}) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		n := 0
		for {
			select {
			case <-done: // 취소 신호 수신
				fmt.Println("    generateNumbers: 취소 신호 받아 종료")
				return
			case ch <- n: // 다음 숫자 전송
				n++
			}
		}
	}()
	return ch
}

func doneChannelPattern() {
	fmt.Println("\n--- 5. done 채널 패턴 ---")

	// done 채널: 취소/종료 신호를 전파하는 관용적 패턴
	// close(done)으로 브로드캐스트: 여러 고루틴이 동시에 신호를 받을 수 있음
	done := make(chan struct{}) // struct{}는 메모리를 사용하지 않음

	numbers := generateNumbers(done)

	// 처음 5개만 수신
	fmt.Println("  처음 5개 숫자:")
	for i := 0; i < 5; i++ {
		fmt.Printf("    받은 숫자: %d\n", <-numbers)
	}

	// done 채널 닫기로 모든 하위 고루틴에게 취소 신호 전파
	close(done)
	time.Sleep(50 * time.Millisecond) // 고루틴이 정리될 시간

	fmt.Println("  done 채널 패턴의 장점:")
	fmt.Println("  - close()는 브로드캐스트: 여러 고루틴 동시 취소")
	fmt.Println("  - struct{}{}는 데이터 없는 순수 신호")
	fmt.Println("  - context 패키지가 이 패턴을 발전시킨 것")
}

// ─────────────────────────────────────────
// 6. 채널 파이프라인 간단 예제
// ─────────────────────────────────────────

// naturals는 자연수를 생성합니다.
func naturals(done <-chan struct{}) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 1; ; i++ {
			select {
			case <-done:
				return
			case ch <- i:
			}
		}
	}()
	return ch
}

// squares는 입력을 제곱합니다.
func squares(done <-chan struct{}, in <-chan int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for v := range in {
			select {
			case <-done:
				return
			case ch <- v * v:
			}
		}
	}()
	return ch
}

func simplePipeline() {
	fmt.Println("\n--- 6. 간단한 채널 파이프라인 ---")

	done := make(chan struct{})
	defer close(done) // main 종료 시 모든 고루틴 정리

	// 파이프라인: naturals → squares
	nums := naturals(done)
	sqrs := squares(done, nums)

	fmt.Println("  처음 5개 자연수의 제곱:")
	for i := 0; i < 5; i++ {
		fmt.Printf("    %d\n", <-sqrs)
	}
}

// ─────────────────────────────────────────
// 7. 채널 nil 주의사항
// ─────────────────────────────────────────

func nilChannelDemo() {
	fmt.Println("\n--- 7. nil 채널 주의사항 ---")

	var ch chan int // nil 채널 (초기화하지 않음)

	fmt.Printf("  nil 채널: %v\n", ch)
	fmt.Println("  nil 채널에 송신: 영원히 블로킹 (데드락)")
	fmt.Println("  nil 채널에서 수신: 영원히 블로킹 (데드락)")
	fmt.Println("  nil 채널 닫기: 패닉!")
	fmt.Println("  select에서 nil 채널: 해당 case는 절대 선택되지 않음 (유용!)")

	// nil 채널의 유용한 사용: select에서 case 비활성화
	ch1 := make(chan int, 1)
	ch1 <- 42

	var ch2 chan int // nil: 이 case는 절대 선택 안 됨

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case v := <-ch1:
			fmt.Printf("  ch1에서 수신: %d (ch2는 nil이라 선택 안 됨)\n", v)
		case v := <-ch2: // ch2가 nil이므로 절대 선택되지 않음
			fmt.Printf("  ch2에서 수신: %d\n", v)
		}
	}()
	wg.Wait()
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== Go 동시성: 채널(Channel) ===")

	basicChannel()
	bufferedVsUnbuffered()
	directionalChannels()
	closeAndRange()
	doneChannelPattern()
	simplePipeline()
	nilChannelDemo()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
