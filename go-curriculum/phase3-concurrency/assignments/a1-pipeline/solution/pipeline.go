// 패키지 선언
// 참고 솔루션 - 풀기 전에 보지 마세요!
package pipeline

import "context"

// Generate는 nums 슬라이스의 숫자를 채널로 전송합니다.
func Generate(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // 고루틴 종료 시 채널 닫기 보장
		for _, n := range nums {
			select {
			case out <- n: // 정상 전송
			case <-ctx.Done(): // 취소 신호 수신 시 즉시 종료
				return
			}
		}
	}()
	return out
}

// Filter는 입력 채널에서 짝수만 통과시킵니다.
func Filter(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case n, ok := <-in:
				if !ok { // 입력 채널 닫힘
					return
				}
				if n%2 == 0 { // 짝수만 전송
					select {
					case out <- n:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()
	return out
}

// Square는 입력 채널의 각 값을 제곱해서 출력합니다.
func Square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case n, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- n * n:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out
}

// Sum은 입력 채널의 모든 값을 합산해 반환합니다.
func Sum(ctx context.Context, in <-chan int) (int, error) {
	total := 0
	for {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		case n, ok := <-in:
			if !ok { // 채널 닫힘 = 모든 값 수신 완료
				return total, nil
			}
			total += n
		}
	}
}

// RunPipeline은 전체 파이프라인을 실행합니다.
func RunPipeline(ctx context.Context, nums []int) (int, error) {
	// 단계별 연결: Generate → Filter → Square → Sum
	generated := Generate(ctx, nums...)
	filtered := Filter(ctx, generated)
	squared := Square(ctx, filtered)
	return Sum(ctx, squared)
}
