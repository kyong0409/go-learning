// calc_bench_test.go: 벤치마크 테스트
//
// 벤치마크 함수 규칙:
// - 함수명은 Benchmark로 시작합니다.
// - *testing.B를 매개변수로 받습니다.
// - b.N 횟수만큼 루프를 실행합니다.
//
// 실행 방법:
//   go test -bench=.                  # 모든 벤치마크 실행
//   go test -bench=BenchmarkAdd       # 특정 벤치마크만 실행
//   go test -bench=. -benchmem        # 메모리 할당도 출력
//   go test -bench=. -count=3         # 3번 반복 실행
//   go test -bench=. -benchtime=5s    # 5초 동안 실행
package calc

import (
	"math"
	"testing"
)

// ─────────────────────────────────────────
// 기본 연산 벤치마크
// ─────────────────────────────────────────

// BenchmarkAdd는 Add 함수의 성능을 측정합니다.
func BenchmarkAdd(b *testing.B) {
	// b.N: 테스팅 프레임워크가 안정적인 측정을 위해 자동으로 결정하는 반복 횟수
	for i := 0; i < b.N; i++ {
		Add(3.14, 2.72)
	}
}

// BenchmarkSubtract는 Subtract 함수의 성능을 측정합니다.
func BenchmarkSubtract(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Subtract(100.0, 42.5)
	}
}

// BenchmarkMultiply는 Multiply 함수의 성능을 측정합니다.
func BenchmarkMultiply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Multiply(6.28, 3.14)
	}
}

// BenchmarkDivide는 Divide 함수의 성능을 측정합니다.
func BenchmarkDivide(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Divide(355.0, 113.0)
	}
}

// BenchmarkSqrt는 Sqrt 함수의 성능을 측정합니다.
func BenchmarkSqrt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sqrt(2.0)
	}
}

// BenchmarkPower는 Power 함수의 성능을 측정합니다.
func BenchmarkPower(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Power(2.0, 10.0)
	}
}

// ─────────────────────────────────────────
// 비교 벤치마크: math.Sqrt vs calc.Sqrt
// ─────────────────────────────────────────
// 두 구현의 성능을 직접 비교합니다.

// BenchmarkSqrtDirect는 math.Sqrt를 직접 호출한 성능입니다.
func BenchmarkSqrtDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		math.Sqrt(2.0)
	}
}

// BenchmarkSqrtWithCheck는 calc.Sqrt (음수 체크 포함)의 성능입니다.
func BenchmarkSqrtWithCheck(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Sqrt(2.0)
	}
}

// ─────────────────────────────────────────
// Calculator 구조체 벤치마크
// ─────────────────────────────────────────

// BenchmarkCalculatorAdd는 Calculator.Add 메서드의 성능입니다.
func BenchmarkCalculatorAdd(b *testing.B) {
	calc := NewCalculator()
	b.ResetTimer() // 초기화 시간을 측정에서 제외

	for i := 0; i < b.N; i++ {
		calc.Add(float64(i), float64(i+1))
	}
}

// BenchmarkCalculatorAddNoHistory는 이력 없이 Add를 비교합니다.
// 이력 기록의 오버헤드를 측정합니다.
func BenchmarkCalculatorAddNoHistory(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(float64(i), float64(i+1))
	}
}

// ─────────────────────────────────────────
// 서브 벤치마크
// ─────────────────────────────────────────
// b.Run()으로 관련 벤치마크를 그룹화합니다.

// BenchmarkOperations는 모든 연산을 서브 벤치마크로 비교합니다.
func BenchmarkOperations(b *testing.B) {
	b.Run("Add", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Add(1.0, 2.0)
		}
	})

	b.Run("Subtract", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Subtract(10.0, 3.0)
		}
	})

	b.Run("Multiply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Multiply(4.0, 5.0)
		}
	})

	b.Run("Divide", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Divide(20.0, 4.0)
		}
	})

	b.Run("Sqrt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Sqrt(16.0)
		}
	})

	b.Run("Power", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Power(2.0, 8.0)
		}
	})
}

// ─────────────────────────────────────────
// 병렬 벤치마크
// ─────────────────────────────────────────
// b.RunParallel로 여러 고루틴에서 병렬 실행 성능을 측정합니다.

// BenchmarkAddParallel은 Add의 병렬 성능을 측정합니다.
func BenchmarkAddParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		// pb.Next()가 false를 반환할 때까지 반복
		for pb.Next() {
			Add(3.14, 2.72)
		}
	})
}
