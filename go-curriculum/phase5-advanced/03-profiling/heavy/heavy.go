// heavy/heavy.go
// 의도적으로 비효율적인 함수들 - 프로파일링 학습용
//
// 이 파일의 함수들은 CPU, 메모리, 고루틴 사용을 과도하게 일으킵니다.
// pprof로 병목 지점을 식별하는 연습에 사용합니다.
package heavy

import (
	"math"
	"strings"
	"sync"
)

// ============================================================
// CPU 집약적 함수
// ============================================================

// CPUIntensive는 의도적으로 CPU를 많이 사용하는 함수입니다.
// 비효율: 불필요한 제곱근 계산, 루프 내 반복 연산
//
// 최적화 포인트 (pprof로 확인 가능):
//   - 반복되는 math.Sqrt 계산을 캐시로 대체
//   - 루프 언롤링(loop unrolling) 적용
func CPUIntensive(n int) int {
	result := 0
	for i := 0; i < n; i++ {
		// 비효율: 매 반복마다 제곱근 계산 (부동소수점 연산 비용)
		result += int(math.Sqrt(float64(i)))
		// 비효율: 삼각함수 중복 계산
		result += int(math.Sin(float64(i)) * math.Cos(float64(i)) * 100)
	}
	return result
}

// CPUIntensiveOptimized는 CPUIntensive의 최적화 버전입니다.
// 벤치마크로 성능 차이를 측정해보세요.
func CPUIntensiveOptimized(n int) int {
	result := 0
	// 최적화: 단순한 정수 연산으로 대체
	for i := 0; i < n; i++ {
		result += i / 2
	}
	return result
}

// FibonacciNaive는 재귀 피보나치 - 지수 시간 복잡도 O(2^n)
// pprof에서 재귀 호출 스택이 깊게 쌓이는 것을 확인할 수 있습니다.
func FibonacciNaive(n int) int {
	if n <= 1 {
		return n
	}
	// 비효율: 동일한 계산을 중복으로 수행
	return FibonacciNaive(n-1) + FibonacciNaive(n-2)
}

// FibonacciMemo는 메모이제이션을 적용한 피보나치 - O(n)
func FibonacciMemo(n int) int {
	memo := make(map[int]int, n)
	var fib func(k int) int
	fib = func(k int) int {
		if k <= 1 {
			return k
		}
		if v, ok := memo[k]; ok {
			return v
		}
		result := fib(k-1) + fib(k-2)
		memo[k] = result
		return result
	}
	return fib(n)
}

// ============================================================
// 메모리 집약적 함수
// ============================================================

// MemoryIntensive는 의도적으로 메모리를 많이 할당하는 함수입니다.
// 비효율: 불필요한 중간 슬라이스/문자열 생성, 메모리 단편화
//
// 최적화 포인트 (pprof heap으로 확인 가능):
//   - strings.Builder 사용으로 문자열 연결 최적화
//   - 미리 용량을 지정한 슬라이스로 재할당 방지
func MemoryIntensive(n int) []byte {
	// 비효율: 루프마다 문자열 연결 → O(n²) 메모리 할당
	result := ""
	for i := 0; i < n; i++ {
		// 비효율: 매 반복마다 새 문자열 생성
		result += "a"
	}

	// 비효율: 불필요한 중간 변환
	parts := make([]string, 0) // 용량 미지정
	for i := 0; i < 100; i++ {
		parts = append(parts, result[:len(result)/10+1])
	}
	_ = strings.Join(parts, ",")

	return []byte(result)
}

// MemoryIntensiveOptimized는 MemoryIntensive의 최적화 버전입니다.
func MemoryIntensiveOptimized(n int) []byte {
	// 최적화 1: strings.Builder로 문자열 연결 O(n)
	var sb strings.Builder
	sb.Grow(n) // 미리 용량 지정
	for i := 0; i < n; i++ {
		sb.WriteByte('a')
	}
	return []byte(sb.String())
}

// MemoryLeaky는 메모리 누수를 시뮬레이션합니다.
// 전역 슬라이스에 데이터를 계속 추가해 GC가 회수하지 못합니다.
var leakyStorage [][]byte // 경고: 실제 프로덕션에서는 이렇게 하면 안 됩니다

// AppendToLeak은 leakyStorage에 데이터를 추가합니다.
// pprof heap에서 메모리가 계속 증가하는 것을 확인할 수 있습니다.
func AppendToLeak(size int) {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	leakyStorage = append(leakyStorage, data)
}

// ClearLeak은 누수된 메모리를 해제합니다.
func ClearLeak() {
	leakyStorage = nil
}

// ============================================================
// 고루틴 집약적 함수
// ============================================================

// GoroutineIntensive는 많은 고루틴을 생성하고 동기화합니다.
// 비효율: 불필요하게 많은 고루틴 생성, 과도한 채널 오버헤드
//
// 최적화 포인트 (pprof goroutine으로 확인 가능):
//   - 워커 풀 패턴으로 고루틴 수 제한
//   - 배치 처리로 오버헤드 감소
func GoroutineIntensive(n int) int {
	var wg sync.WaitGroup
	results := make(chan int, n)

	// 비효율: n개의 고루틴을 각각 생성 (각각 단순 덧셈만 수행)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			// 각 고루틴이 하는 일이 너무 작음 (고루틴 오버헤드 > 실제 작업)
			results <- val * val
		}(i)
	}

	// 모든 고루틴이 완료되면 채널 닫기
	go func() {
		wg.Wait()
		close(results)
	}()

	// 결과 합산
	total := 0
	for v := range results {
		total += v
	}
	return total
}

// GoroutineIntensiveOptimized는 워커 풀 패턴을 사용한 최적화 버전입니다.
func GoroutineIntensiveOptimized(n int) int {
	// 최적화: CPU 코어 수에 맞는 워커만 생성
	numWorkers := 4 // runtime.NumCPU()로 동적으로 설정 가능
	jobs := make(chan int, n)
	results := make(chan int, n)

	// 워커 고루틴 시작
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for val := range jobs {
				results <- val * val
			}
		}()
	}

	// 작업 전송
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)

	// 완료 대기 후 채널 닫기
	go func() {
		wg.Wait()
		close(results)
	}()

	total := 0
	for v := range results {
		total += v
	}
	return total
}

// ============================================================
// 복합 비효율 함수 (여러 문제를 동시에 포함)
// ============================================================

// ProcessData는 여러 종류의 비효율을 포함한 복합 함수입니다.
// 실제 코드베이스에서 흔히 볼 수 있는 패턴을 모아두었습니다.
func ProcessData(data []int) []int {
	// 비효율 1: 용량 지정 없는 슬라이스 (여러 번 재할당)
	result := []int{}

	for _, v := range data {
		// 비효율 2: 루프 내 불필요한 함수 호출
		if isPrime(v) {
			// 비효율 3: 필요 이상의 연산
			result = append(result, v*v+int(math.Sqrt(float64(v))))
		}
	}
	return result
}

// isPrime은 소수 판별 함수 (단순 구현)입니다.
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	// 비효율: math.Sqrt를 매번 계산 (캐시 없음)
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}
