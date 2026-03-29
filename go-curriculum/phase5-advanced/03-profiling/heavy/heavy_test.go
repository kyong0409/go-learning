// heavy/heavy_test.go
// 벤치마크 테스트 - 비효율 vs 최적화 함수 성능 비교
//
// 실행 방법:
//   go test ./heavy/ -bench=. -benchmem
//   go test ./heavy/ -bench=BenchmarkCPU -benchtime=5s
//   go test ./heavy/ -bench=. -benchmem -cpuprofile=cpu.prof
//   go test ./heavy/ -bench=. -benchmem -memprofile=mem.prof
//
// 프로파일 분석:
//   go tool pprof cpu.prof
//   go tool pprof mem.prof
package heavy

import (
	"testing"
)

// ============================================================
// CPU 벤치마크
// ============================================================

// BenchmarkCPUIntensive는 비효율 CPU 함수의 성능을 측정합니다.
func BenchmarkCPUIntensive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CPUIntensive(10_000)
	}
}

// BenchmarkCPUIntensiveOptimized는 최적화된 CPU 함수의 성능을 측정합니다.
func BenchmarkCPUIntensiveOptimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CPUIntensiveOptimized(10_000)
	}
}

// BenchmarkFibonacciNaive는 재귀 피보나치의 성능을 측정합니다.
// n=35 수준에서 매우 느림을 확인할 수 있습니다.
func BenchmarkFibonacciNaive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FibonacciNaive(30) // 30 이상은 매우 느림
	}
}

// BenchmarkFibonacciMemo는 메모이제이션 피보나치의 성능을 측정합니다.
func BenchmarkFibonacciMemo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FibonacciMemo(30)
	}
}

// ============================================================
// 메모리 벤치마크
// ============================================================

// BenchmarkMemoryIntensive는 비효율 메모리 함수를 측정합니다.
// -benchmem으로 실행하면 할당 횟수와 바이트를 확인할 수 있습니다.
func BenchmarkMemoryIntensive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MemoryIntensive(1_000)
	}
}

// BenchmarkMemoryIntensiveOptimized는 최적화된 메모리 함수를 측정합니다.
func BenchmarkMemoryIntensiveOptimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MemoryIntensiveOptimized(1_000)
	}
}

// ============================================================
// 고루틴 벤치마크
// ============================================================

// BenchmarkGoroutineIntensive는 비효율 고루틴 함수를 측정합니다.
func BenchmarkGoroutineIntensive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GoroutineIntensive(100)
	}
}

// BenchmarkGoroutineIntensiveOptimized는 워커 풀 패턴을 측정합니다.
func BenchmarkGoroutineIntensiveOptimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GoroutineIntensiveOptimized(100)
	}
}

// ============================================================
// 서브벤치마크: 입력 크기별 성능 측정
// ============================================================

// BenchmarkCPUScaling은 입력 크기에 따른 성능 변화를 측정합니다.
func BenchmarkCPUScaling(b *testing.B) {
	sizes := []int{100, 1_000, 10_000, 100_000}
	for _, size := range sizes {
		b.Run(
			// 서브벤치마크 이름: BenchmarkCPUScaling/n=100 등
			"n="+itoa(size),
			func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					CPUIntensive(size)
				}
			},
		)
	}
}

// BenchmarkMemoryScaling은 입력 크기에 따른 메모리 할당 변화를 측정합니다.
func BenchmarkMemoryScaling(b *testing.B) {
	sizes := []int{100, 500, 1_000, 2_000}
	for _, size := range sizes {
		b.Run(
			"naive/n="+itoa(size),
			func(b *testing.B) {
				b.ReportAllocs() // 이 벤치마크에 대해 메모리 통계 출력
				for i := 0; i < b.N; i++ {
					MemoryIntensive(size)
				}
			},
		)
		b.Run(
			"optimized/n="+itoa(size),
			func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					MemoryIntensiveOptimized(size)
				}
			},
		)
	}
}

// ============================================================
// 유닛 테스트 (벤치마크와 함께 포함)
// ============================================================

func TestCPUIntensive(t *testing.T) {
	// 결정론적 결과 확인
	result := CPUIntensive(100)
	if result <= 0 {
		t.Errorf("CPUIntensive(100) = %d; 양수여야 합니다", result)
	}
}

func TestFibonacciConsistency(t *testing.T) {
	// 두 구현이 동일한 결과를 반환하는지 확인
	for n := 0; n <= 20; n++ {
		naive := FibonacciNaive(n)
		memo := FibonacciMemo(n)
		if naive != memo {
			t.Errorf("Fibonacci(%d): naive=%d, memo=%d; 결과가 달라야 하지 않습니다", n, naive, memo)
		}
	}
}

func TestGoroutineIntensiveConsistency(t *testing.T) {
	// 두 구현이 동일한 결과를 반환하는지 확인
	n := 10
	r1 := GoroutineIntensive(n)
	r2 := GoroutineIntensiveOptimized(n)
	if r1 != r2 {
		t.Errorf("GoroutineIntensive(%d)=%d vs Optimized=%d: 결과가 일치해야 합니다", n, r1, r2)
	}
}

// ============================================================
// 헬퍼
// ============================================================

// itoa는 int를 문자열로 변환합니다 (strconv 임포트 없이 사용).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
