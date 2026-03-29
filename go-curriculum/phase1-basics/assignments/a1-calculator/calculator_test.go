// calculator_test.go: 계산기 함수 채점 테스트
//
// 실행: go test -v
// 각 테스트 통과 시 점수를 획득합니다.
package calculator

import (
	"errors"
	"math"
	"testing"
)

// ─────────────────────────────────────────
// 채점 시스템
// ─────────────────────────────────────────

// grader: 테스트 결과 집계용 구조체
type grader struct {
	passed int
	total  int
	points int
	maxPoints int
}

// check: 서브테스트 결과를 기록합니다.
func (g *grader) check(t *testing.T, name string, pts int, fn func(t *testing.T)) {
	t.Helper()
	g.total++
	g.maxPoints += pts
	passed := true
	t.Run(name, func(t *testing.T) {
		fn(t)
		if t.Failed() {
			passed = false
		}
	})
	if passed {
		g.passed++
		g.points += pts
	}
}

// report: 채점 결과를 출력합니다.
func (g *grader) report(t *testing.T) {
	t.Helper()
	t.Logf("\n==================")
	t.Logf("=== GRADE REPORT ===")
	t.Logf("==================")
	t.Logf("Passed: %d/%d", g.passed, g.total)
	t.Logf("Score:  %d/%d", g.points, g.maxPoints)
	t.Logf("==================")
}

// ─────────────────────────────────────────
// float64 비교 헬퍼 (부동소수점 오차 허용)
// ─────────────────────────────────────────

const epsilon = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// ─────────────────────────────────────────
// 테스트 함수
// ─────────────────────────────────────────

func TestCalculator(t *testing.T) {
	g := &grader{}

	// ─── Add (10점) ───
	g.check(t, "Add/양수+양수", 2, func(t *testing.T) {
		got := Add(3, 4)
		if !almostEqual(got, 7) {
			t.Errorf("Add(3, 4) = %.2f; want 7.0", got)
		}
	})
	g.check(t, "Add/음수+양수", 2, func(t *testing.T) {
		got := Add(-5, 3)
		if !almostEqual(got, -2) {
			t.Errorf("Add(-5, 3) = %.2f; want -2.0", got)
		}
	})
	g.check(t, "Add/소수점", 2, func(t *testing.T) {
		got := Add(1.5, 2.3)
		if !almostEqual(got, 3.8) {
			t.Errorf("Add(1.5, 2.3) = %.4f; want 3.8", got)
		}
	})
	g.check(t, "Add/제로", 4, func(t *testing.T) {
		if !almostEqual(Add(0, 0), 0) {
			t.Error("Add(0, 0) should be 0")
		}
		if !almostEqual(Add(5, 0), 5) {
			t.Error("Add(5, 0) should be 5")
		}
	})

	// ─── Subtract (8점) ───
	g.check(t, "Subtract/기본", 4, func(t *testing.T) {
		got := Subtract(10, 3)
		if !almostEqual(got, 7) {
			t.Errorf("Subtract(10, 3) = %.2f; want 7.0", got)
		}
	})
	g.check(t, "Subtract/음수결과", 4, func(t *testing.T) {
		got := Subtract(3, 10)
		if !almostEqual(got, -7) {
			t.Errorf("Subtract(3, 10) = %.2f; want -7.0", got)
		}
	})

	// ─── Multiply (8점) ───
	g.check(t, "Multiply/기본", 4, func(t *testing.T) {
		got := Multiply(4, 5)
		if !almostEqual(got, 20) {
			t.Errorf("Multiply(4, 5) = %.2f; want 20.0", got)
		}
	})
	g.check(t, "Multiply/제로곱", 4, func(t *testing.T) {
		got := Multiply(100, 0)
		if !almostEqual(got, 0) {
			t.Errorf("Multiply(100, 0) = %.2f; want 0.0", got)
		}
	})

	// ─── Divide (16점) ───
	g.check(t, "Divide/기본", 4, func(t *testing.T) {
		got, err := Divide(10, 4)
		if err != nil {
			t.Fatalf("Divide(10, 4) 예상치 못한 에러: %v", err)
		}
		if !almostEqual(got, 2.5) {
			t.Errorf("Divide(10, 4) = %.4f; want 2.5", got)
		}
	})
	g.check(t, "Divide/정확한나눗셈", 4, func(t *testing.T) {
		got, err := Divide(9, 3)
		if err != nil {
			t.Fatalf("Divide(9, 3) 예상치 못한 에러: %v", err)
		}
		if !almostEqual(got, 3.0) {
			t.Errorf("Divide(9, 3) = %.4f; want 3.0", got)
		}
	})
	g.check(t, "Divide/0으로나누기에러반환", 4, func(t *testing.T) {
		_, err := Divide(10, 0)
		if err == nil {
			t.Error("Divide(10, 0): 에러가 반환되어야 합니다")
			return
		}
		if !errors.Is(err, ErrDivisionByZero) {
			t.Errorf("Divide(10, 0) 에러: got %v, want ErrDivisionByZero", err)
		}
	})
	g.check(t, "Divide/0으로나누기결과값", 4, func(t *testing.T) {
		result, _ := Divide(10, 0)
		if result != 0 {
			t.Errorf("Divide(10, 0) 결과값: got %.2f, want 0", result)
		}
	})

	// ─── Power (12점) ───
	g.check(t, "Power/양수거듭제곱", 4, func(t *testing.T) {
		got := Power(2, 10)
		if !almostEqual(got, 1024) {
			t.Errorf("Power(2, 10) = %.2f; want 1024.0", got)
		}
	})
	g.check(t, "Power/0승", 4, func(t *testing.T) {
		got := Power(5, 0)
		if !almostEqual(got, 1) {
			t.Errorf("Power(5, 0) = %.2f; want 1.0", got)
		}
	})
	g.check(t, "Power/1승", 4, func(t *testing.T) {
		got := Power(7, 1)
		if !almostEqual(got, 7) {
			t.Errorf("Power(7, 1) = %.2f; want 7.0", got)
		}
	})

	// ─── Sqrt (16점) ───
	g.check(t, "Sqrt/완전제곱수", 4, func(t *testing.T) {
		got, err := Sqrt(16)
		if err != nil {
			t.Fatalf("Sqrt(16) 예상치 못한 에러: %v", err)
		}
		if !almostEqual(got, 4) {
			t.Errorf("Sqrt(16) = %.4f; want 4.0", got)
		}
	})
	g.check(t, "Sqrt/소수결과", 4, func(t *testing.T) {
		got, err := Sqrt(2)
		if err != nil {
			t.Fatalf("Sqrt(2) 예상치 못한 에러: %v", err)
		}
		want := math.Sqrt(2)
		if !almostEqual(got, want) {
			t.Errorf("Sqrt(2) = %.10f; want %.10f", got, want)
		}
	})
	g.check(t, "Sqrt/음수에러반환", 4, func(t *testing.T) {
		_, err := Sqrt(-4)
		if err == nil {
			t.Error("Sqrt(-4): 에러가 반환되어야 합니다")
			return
		}
		if !errors.Is(err, ErrNegativeInput) {
			t.Errorf("Sqrt(-4) 에러: got %v, want ErrNegativeInput", err)
		}
	})
	g.check(t, "Sqrt/제로", 4, func(t *testing.T) {
		got, err := Sqrt(0)
		if err != nil {
			t.Fatalf("Sqrt(0) 예상치 못한 에러: %v", err)
		}
		if !almostEqual(got, 0) {
			t.Errorf("Sqrt(0) = %.4f; want 0.0", got)
		}
	})

	// ─── Abs (10점) ───
	g.check(t, "Abs/음수", 4, func(t *testing.T) {
		got := Abs(-5.5)
		if !almostEqual(got, 5.5) {
			t.Errorf("Abs(-5.5) = %.4f; want 5.5", got)
		}
	})
	g.check(t, "Abs/양수불변", 3, func(t *testing.T) {
		got := Abs(3.2)
		if !almostEqual(got, 3.2) {
			t.Errorf("Abs(3.2) = %.4f; want 3.2", got)
		}
	})
	g.check(t, "Abs/제로", 3, func(t *testing.T) {
		got := Abs(0)
		if !almostEqual(got, 0) {
			t.Errorf("Abs(0) = %.4f; want 0.0", got)
		}
	})

	// ─── Modulo (10점) ───
	g.check(t, "Modulo/기본", 4, func(t *testing.T) {
		got, err := Modulo(10, 3)
		if err != nil {
			t.Fatalf("Modulo(10, 3) 예상치 못한 에러: %v", err)
		}
		if got != 1 {
			t.Errorf("Modulo(10, 3) = %d; want 1", got)
		}
	})
	g.check(t, "Modulo/나누어떨어짐", 2, func(t *testing.T) {
		got, err := Modulo(12, 4)
		if err != nil {
			t.Fatalf("Modulo(12, 4) 예상치 못한 에러: %v", err)
		}
		if got != 0 {
			t.Errorf("Modulo(12, 4) = %d; want 0", got)
		}
	})
	g.check(t, "Modulo/0으로나누기", 4, func(t *testing.T) {
		_, err := Modulo(10, 0)
		if err == nil {
			t.Error("Modulo(10, 0): 에러가 반환되어야 합니다")
			return
		}
		if !errors.Is(err, ErrDivisionByZero) {
			t.Errorf("Modulo(10, 0) 에러: got %v, want ErrDivisionByZero", err)
		}
	})

	// ─── 최종 채점 리포트 ───
	g.report(t)
}
