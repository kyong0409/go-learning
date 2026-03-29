// 패키지 선언
package main

// Go의 동시성 모델: select 문
//
// select는 여러 채널 연산을 동시에 기다리는 제어 구조입니다.
// switch문과 유사하지만 채널 연산에 특화되어 있습니다.
//
// 동작 방식:
// - 준비된 case가 없으면: 준비될 때까지 블로킹
// - 준비된 case가 하나: 해당 case 실행
// - 준비된 case가 여러 개: 무작위로 하나 선택 (공정한 랜덤)
// - default case: 준비된 채널 없을 때 즉시 실행

import (
	"fmt"
	"math/rand"
	"time"
)

// ─────────────────────────────────────────
// 1. 기본 select: 다중 채널 멀티플렉싱
// ─────────────────────────────────────────

func basicSelect() {
	fmt.Println("\n--- 1. 기본 select: 다중 채널 멀티플렉싱 ---")

	ch1 := make(chan string)
	ch2 := make(chan string)

	// 두 채널에 서로 다른 지연 시간으로 값 전송
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "채널1에서 메시지"
	}()
	go func() {
		time.Sleep(50 * time.Millisecond)
		ch2 <- "채널2에서 메시지"
	}()

	// select로 두 채널 중 먼저 준비된 것을 수신
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Printf("  수신: %s\n", msg1)
		case msg2 := <-ch2:
			fmt.Printf("  수신: %s\n", msg2)
		}
	}

	// 여러 case가 동시에 준비된 경우: Go 런타임이 무작위 선택
	fmt.Println("\n  동시에 준비된 경우 (무작위 선택):")
	c1 := make(chan int, 1)
	c2 := make(chan int, 1)
	c1 <- 1
	c2 <- 2

	for i := 0; i < 4; i++ {
		// 버퍼를 다시 채움
		if len(c1) == 0 {
			c1 <- 1
		}
		if len(c2) == 0 {
			c2 <- 2
		}
		select {
		case v := <-c1:
			fmt.Printf("    c1 선택됨: %d\n", v)
		case v := <-c2:
			fmt.Printf("    c2 선택됨: %d\n", v)
		}
	}
}

// ─────────────────────────────────────────
// 2. 타임아웃 패턴 (time.After)
// ─────────────────────────────────────────

// fetchData는 느린 외부 API 호출을 시뮬레이션합니다.
func fetchData(delay time.Duration) <-chan string {
	ch := make(chan string, 1)
	go func() {
		time.Sleep(delay)
		ch <- fmt.Sprintf("데이터 수신 완료 (지연: %v)", delay)
	}()
	return ch
}

func timeoutPattern() {
	fmt.Println("\n--- 2. 타임아웃 패턴 ---")

	// time.After(d): d 시간 후에 현재 시간을 전송하는 채널 반환
	timeout := 150 * time.Millisecond

	// 케이스 1: 타임아웃 이내에 응답
	fmt.Println("  케이스 1: 빠른 응답 (100ms) vs 타임아웃 (150ms):")
	dataCh := fetchData(100 * time.Millisecond)
	select {
	case data := <-dataCh:
		fmt.Printf("    성공: %s\n", data)
	case <-time.After(timeout):
		fmt.Println("    실패: 타임아웃!")
	}

	// 케이스 2: 타임아웃 초과
	fmt.Println("\n  케이스 2: 느린 응답 (200ms) vs 타임아웃 (150ms):")
	dataCh2 := fetchData(200 * time.Millisecond)
	select {
	case data := <-dataCh2:
		fmt.Printf("    성공: %s\n", data)
	case <-time.After(timeout):
		fmt.Println("    실패: 타임아웃! (응답이 너무 느림)")
	}

	// time.NewTimer: 재사용 가능한 타이머 (time.After보다 메모리 효율적)
	fmt.Println("\n  time.NewTimer 사용:")
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop() // 중요: 타이머 사용 후 반드시 Stop() 호출 (리소스 해제)

	dataCh3 := fetchData(100 * time.Millisecond)
	select {
	case data := <-dataCh3:
		fmt.Printf("    성공: %s\n", data)
		timer.Stop() // 타이머 조기 중단
	case <-timer.C:
		fmt.Println("    실패: 타이머 만료!")
	}
}

// ─────────────────────────────────────────
// 3. 논블로킹 채널 연산 (default case)
// ─────────────────────────────────────────

func nonBlockingChannel() {
	fmt.Println("\n--- 3. 논블로킹 채널 연산 (default) ---")

	ch := make(chan int, 1)

	// 논블로킹 송신: default가 있어 채널이 가득 차도 블로킹 안 됨
	fmt.Println("  논블로킹 송신:")
	select {
	case ch <- 42:
		fmt.Println("    채널에 42 송신 성공")
	default:
		fmt.Println("    채널이 가득 참, 송신 건너뜀")
	}

	// 두 번째 시도: 이미 버퍼가 가득 찬 경우
	select {
	case ch <- 99:
		fmt.Println("    채널에 99 송신 성공")
	default:
		fmt.Println("    채널이 가득 참, 두 번째 송신 건너뜀")
	}

	// 논블로킹 수신
	fmt.Println("\n  논블로킹 수신:")
	select {
	case v := <-ch:
		fmt.Printf("    수신: %d\n", v)
	default:
		fmt.Println("    채널이 비어있음")
	}

	// 빈 채널에서 논블로킹 수신 시도
	select {
	case v := <-ch:
		fmt.Printf("    수신: %d\n", v)
	default:
		fmt.Println("    채널이 비어있어 수신 건너뜀")
	}

	// 실용적 예: 워커 상태 확인 (블로킹 없이)
	fmt.Println("\n  실용적 예: 작업 큐 상태 확인:")
	workQueue := make(chan string, 5)
	workQueue <- "작업1"
	workQueue <- "작업2"

	for {
		select {
		case job := <-workQueue:
			fmt.Printf("    처리 중: %s\n", job)
		default:
			fmt.Println("    작업 큐 비어있음, 대기 중...")
			goto doneNonBlocking
		}
	}
doneNonBlocking:
}

// ─────────────────────────────────────────
// 4. done 채널 + select로 취소 구현
// ─────────────────────────────────────────

// worker는 작업을 처리하다가 done 신호를 받으면 종료합니다.
func worker(id int, jobs <-chan int, done <-chan struct{}, results chan<- int) {
	for {
		select {
		case <-done:
			// 취소 신호 수신: 즉시 종료
			fmt.Printf("  워커 #%d: 취소 신호 받음, 종료\n", id)
			return
		case job, ok := <-jobs:
			if !ok {
				// jobs 채널이 닫힘: 정상 종료
				fmt.Printf("  워커 #%d: 작업 채널 닫힘, 정상 종료\n", id)
				return
			}
			// 작업 처리
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			result := job * job
			select {
			case results <- result:
				fmt.Printf("  워커 #%d: 작업 %d 처리 완료 → 결과 %d\n", id, job, result)
			case <-done:
				fmt.Printf("  워커 #%d: 결과 전송 중 취소됨\n", id)
				return
			}
		}
	}
}

func cancellationWithDone() {
	fmt.Println("\n--- 4. done 채널 + select로 취소 구현 ---")

	jobs := make(chan int, 10)
	results := make(chan int, 10)
	done := make(chan struct{})

	// 3개 워커 시작
	for i := 1; i <= 3; i++ {
		go worker(i, jobs, done, results)
	}

	// 10개 작업 등록
	for i := 1; i <= 10; i++ {
		jobs <- i
	}

	// 3개만 처리하고 나머지는 취소
	fmt.Println("  처음 3개 결과 수신 후 취소:")
	for i := 0; i < 3; i++ {
		result := <-results
		fmt.Printf("  결과 수신: %d\n", result)
	}

	// done 채널 닫기: 모든 워커에게 취소 신호 브로드캐스트
	fmt.Println("  취소 신호 전송...")
	close(done)

	time.Sleep(100 * time.Millisecond) // 워커들이 정리될 시간
	fmt.Println("  취소 완료")
}

// ─────────────────────────────────────────
// 5. select를 이용한 주기적 작업 (time.Tick)
// ─────────────────────────────────────────

func periodicWork() {
	fmt.Println("\n--- 5. 주기적 작업과 select ---")

	// time.Tick: 주기적으로 신호를 보내는 채널
	// 주의: time.Tick은 GC되지 않음 → 장기 실행 프로그램에서는 time.NewTicker 사용
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()

	boom := time.After(250 * time.Millisecond)

	fmt.Println("  50ms마다 tick, 250ms 후 boom:")
	count := 0
loop:
	for {
		select {
		case t := <-tick.C:
			count++
			fmt.Printf("    tick #%d at %v\n", count, t.Format("15:04:05.000"))
		case <-boom:
			fmt.Println("    BOOM! 종료")
			break loop // for 루프 탈출 (break만 쓰면 select를 탈출)
		}
	}
}

// ─────────────────────────────────────────
// 6. for-select 패턴: 고루틴의 메인 루프
// ─────────────────────────────────────────

// aggregator는 여러 채널의 데이터를 하나로 합칩니다.
func aggregator(done <-chan struct{}, channels ...<-chan int) <-chan int {
	merged := make(chan int, 10)

	// 각 입력 채널에 대해 고루틴 생성
	for _, ch := range channels {
		ch := ch // 루프 변수 캡처
		go func() {
			for {
				select {
				case <-done:
					return
				case v, ok := <-ch:
					if !ok {
						return
					}
					select {
					case merged <- v:
					case <-done:
						return
					}
				}
			}
		}()
	}

	return merged
}

func forSelectPattern() {
	fmt.Println("\n--- 6. for-select 패턴 ---")

	done := make(chan struct{})

	// 두 개의 수 생성기
	gen1 := make(chan int, 3)
	gen2 := make(chan int, 3)

	gen1 <- 1
	gen1 <- 2
	gen1 <- 3
	close(gen1)

	gen2 <- 10
	gen2 <- 20
	gen2 <- 30
	close(gen2)

	merged := aggregator(done, gen1, gen2)

	// 수집 (채널이 닫히지 않으므로 타임아웃 사용)
	fmt.Println("  두 채널 병합 결과:")
	timeout := time.After(500 * time.Millisecond)
	var results []int
	for {
		select {
		case v := <-merged:
			results = append(results, v)
			if len(results) == 6 {
				goto done6
			}
		case <-timeout:
			fmt.Println("  타임아웃")
			goto done6
		}
	}
done6:
	close(done)
	fmt.Printf("  수집된 값: %v\n", results)
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== Go 동시성: select 문 ===")

	basicSelect()
	timeoutPattern()
	nonBlockingChannel()
	cancellationWithDone()
	periodicWork()
	forSelectPattern()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
