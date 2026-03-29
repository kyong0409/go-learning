// 패키지 선언
package pipeline

import "context"

// Generate는 nums 슬라이스의 숫자를 채널로 전송합니다.
// Context가 취소되면 즉시 종료합니다.
// 반환된 채널은 모든 숫자를 전송한 후 닫힙니다.
func Generate(ctx context.Context, nums ...int) <-chan int {
	// TODO: 구현하세요
	// 1. 출력 채널 생성
	// 2. 고루틴 시작
	// 3. nums의 각 값을 채널로 전송 (ctx.Done() 확인)
	// 4. 전송 완료 후 채널 닫기
	panic("구현 필요")
}

// Filter는 입력 채널에서 짝수만 통과시킵니다.
// Context가 취소되면 즉시 종료합니다.
func Filter(ctx context.Context, in <-chan int) <-chan int {
	// TODO: 구현하세요
	// 1. 출력 채널 생성
	// 2. 고루틴 시작
	// 3. in에서 값을 수신하며 짝수만 출력 채널로 전송
	// 4. in이 닫히거나 ctx가 취소되면 종료
	panic("구현 필요")
}

// Square는 입력 채널의 각 값을 제곱해서 출력합니다.
// Context가 취소되면 즉시 종료합니다.
func Square(ctx context.Context, in <-chan int) <-chan int {
	// TODO: 구현하세요
	// 1. 출력 채널 생성
	// 2. 고루틴 시작
	// 3. in에서 값을 수신하며 n*n을 출력 채널로 전송
	// 4. in이 닫히거나 ctx가 취소되면 종료
	panic("구현 필요")
}

// Sum은 입력 채널의 모든 값을 합산해 반환합니다.
// Context가 취소되면 현재까지의 합산 값과 ctx.Err()를 반환합니다.
func Sum(ctx context.Context, in <-chan int) (int, error) {
	// TODO: 구현하세요
	// 1. total 변수 초기화
	// 2. select로 in 수신과 ctx.Done() 동시 처리
	// 3. in이 닫히면 total, nil 반환
	// 4. ctx가 취소되면 total, ctx.Err() 반환
	panic("구현 필요")
}

// RunPipeline은 전체 파이프라인을 실행합니다.
// Generate → Filter → Square → Sum 순서로 연결합니다.
// Context가 취소되면 에러를 반환합니다.
func RunPipeline(ctx context.Context, nums []int) (int, error) {
	// TODO: 구현하세요
	// 1. Generate로 숫자 스트림 생성
	// 2. Filter로 짝수 필터링
	// 3. Square로 제곱
	// 4. Sum으로 합산 후 반환
	panic("구현 필요")
}
