// calc_test.go: calc 패키지의 단위 테스트
//
// 이 파일은 다음 테스트 기법을 보여줍니다:
// - 테이블 기반 테스트 (table-driven tests)
// - t.Run()을 이용한 서브테스트 (subtests)
// - 테스트 헬퍼 함수 (test helpers)
// - errors.Is()를 이용한 에러 타입 검사
package calc

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

// ─────────────────────────────────────────
// 테스트 헬퍼 함수
// ─────────────────────────────────────────

// almostEqual은 부동소수점 비교 헬퍼입니다.
// 부동소수점은 정확한 == 비교가 불가능하므로 허용 오차를 사용합니다.
func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

// assertFloat은 float64 값을 검사하는 테스트 헬퍼입니다.
// t.Helper()를 호출하면 실패 시 호출한 줄이 표시됩니다.
func assertFloat(t *testing.T, got, want, tolerance float64) {
	t.Helper() // 이 함수를 호출한 라인을 에러 위치로 표시
	if !almostEqual(got, want, tolerance) {
		t.Errorf("값 불일치: got=%.6f, want=%.6f (허용오차=%.6f)", got, want, tolerance)
	}
}

// assertError는 에러 발생을 검사하는 테스트 헬퍼입니다.
func assertError(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if wantErr && err == nil {
		t.Error("에러가 예상되었지만 nil을 받았습니다")
	}
	if !wantErr && err != nil {
		t.Errorf("에러가 예상되지 않았지만 받았습니다: %v", err)
	}
}

// ─────────────────────────────────────────
// Add 테스트
// ─────────────────────────────────────────

// TestAdd는 Add 함수의 테이블 기반 테스트입니다.
func TestAdd(t *testing.T) {
	// 테이블 기반 테스트: 테스트 케이스를 슬라이스로 정의합니다.
	tests := []struct {
		name string  // 테스트 이름
		a, b float64 // 입력값
		want float64 // 기대값
	}{
		// 양수
		{"양수 + 양수", 3, 4, 7},
		{"큰 수 + 작은 수", 100, 0.5, 100.5},

		// 음수
		{"음수 + 음수", -3, -4, -7},
		{"양수 + 음수", 10, -3, 7},
		{"음수 + 양수", -5, 8, 3},

		// 경계값
		{"0 + 0", 0, 0, 0},
		{"0 + 양수", 0, 5, 5},
		{"음수 + 0", -3, 0, -3},

		// 부동소수점
		{"소수 더하기", 1.1, 2.2, 3.3},
		{"작은 소수", 0.1, 0.2, 0.3},
	}

	for _, tt := range tests {
		// t.Run으로 서브테스트 실행: 개별 케이스를 격리하여 실행
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			// 부동소수점 허용 오차 1e-9
			assertFloat(t, got, tt.want, 1e-9)
		})
	}
}

// ─────────────────────────────────────────
// Subtract 테스트
// ─────────────────────────────────────────

func TestSubtract(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"양수 - 양수", 10, 3, 7},
		{"같은 수 빼기", 5, 5, 0},
		{"양수 - 큰 양수", 3, 10, -7},
		{"음수 - 음수", -5, -3, -2},
		{"0 - 양수", 0, 5, -5},
		{"소수 빼기", 1.5, 0.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Subtract(tt.a, tt.b)
			assertFloat(t, got, tt.want, 1e-9)
		})
	}
}

// ─────────────────────────────────────────
// Multiply 테스트
// ─────────────────────────────────────────

func TestMultiply(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"양수 × 양수", 3, 4, 12},
		{"음수 × 음수", -3, -4, 12},
		{"양수 × 음수", 3, -4, -12},
		{"0 × 수", 0, 999, 0},
		{"수 × 0", 999, 0, 0},
		{"1 × 수", 1, 42, 42},
		{"소수 곱하기", 2.5, 4, 10},
		{"작은 수 곱하기", 0.1, 0.1, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Multiply(tt.a, tt.b)
			assertFloat(t, got, tt.want, 1e-9)
		})
	}
}

// ─────────────────────────────────────────
// Divide 테스트
// ─────────────────────────────────────────

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool
		errIs   error // errors.Is()로 확인할 특정 에러
	}{
		// 정상 케이스
		{"정수 나눗셈", 10, 2, 5, false, nil},
		{"소수 결과", 7, 2, 3.5, false, nil},
		{"음수 ÷ 양수", -10, 2, -5, false, nil},
		{"음수 ÷ 음수", -10, -2, 5, false, nil},
		{"1 ÷ 수", 1, 4, 0.25, false, nil},

		// 에러 케이스
		{"0으로 나누기", 10, 0, 0, true, ErrDivisionByZero},
		{"음수 ÷ 0", -5, 0, 0, true, ErrDivisionByZero},
		{"0 ÷ 0", 0, 0, 0, true, ErrDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Divide(tt.a, tt.b)

			// 에러 검사
			assertError(t, err, tt.wantErr)

			// 특정 에러 타입 검사
			if tt.errIs != nil {
				if !errors.Is(err, tt.errIs) {
					t.Errorf("에러 타입 불일치: got=%v, want=%v", err, tt.errIs)
				}
			}

			// 정상 케이스에서 값 검사
			if !tt.wantErr {
				assertFloat(t, got, tt.want, 1e-9)
			}
		})
	}
}

// ─────────────────────────────────────────
// Sqrt 테스트
// ─────────────────────────────────────────

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		want    float64
		wantErr bool
		errIs   error
	}{
		{"0의 제곱근", 0, 0, false, nil},
		{"1의 제곱근", 1, 1, false, nil},
		{"4의 제곱근", 4, 2, false, nil},
		{"9의 제곱근", 9, 3, false, nil},
		{"2의 제곱근", 2, math.Sqrt2, false, nil},
		{"소수 제곱근", 2.25, 1.5, false, nil},
		// 에러 케이스
		{"음수 제곱근", -1, 0, true, ErrNegativeSqrt},
		{"음수 제곱근 2", -4, 0, true, ErrNegativeSqrt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sqrt(tt.input)
			assertError(t, err, tt.wantErr)
			if tt.errIs != nil && !errors.Is(err, tt.errIs) {
				t.Errorf("에러 타입 불일치: got=%v, want=%v", err, tt.errIs)
			}
			if !tt.wantErr {
				assertFloat(t, got, tt.want, 1e-9)
			}
		})
	}
}

// ─────────────────────────────────────────
// Power 테스트
// ─────────────────────────────────────────

func TestPower(t *testing.T) {
	tests := []struct {
		name     string
		base     float64
		exp      float64
		want     float64
		wantErr  bool
	}{
		{"2의 3제곱", 2, 3, 8, false},
		{"10의 2제곱", 10, 2, 100, false},
		{"수의 0제곱", 5, 0, 1, false},
		{"수의 1제곱", 5, 1, 5, false},
		{"음수의 짝수 제곱", -2, 2, 4, false},
		{"음수의 홀수 제곱", -2, 3, -8, false},
		{"소수 제곱", 2, 0.5, math.Sqrt2, false},
		// 오버플로
		{"오버플로", math.MaxFloat64, 2, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Power(tt.base, tt.exp)
			assertError(t, err, tt.wantErr)
			if !tt.wantErr {
				assertFloat(t, got, tt.want, 1e-9)
			}
		})
	}
}

// ─────────────────────────────────────────
// Calculator 구조체 테스트
// ─────────────────────────────────────────

func TestCalculator(t *testing.T) {
	// t.Run으로 기능별 서브테스트 그룹화
	t.Run("기본연산", func(t *testing.T) {
		calc := NewCalculator()

		t.Run("덧셈", func(t *testing.T) {
			got := calc.Add(3, 4)
			assertFloat(t, got, 7, 1e-9)
		})

		t.Run("뺄셈", func(t *testing.T) {
			got := calc.Subtract(10, 3)
			assertFloat(t, got, 7, 1e-9)
		})

		t.Run("곱셈", func(t *testing.T) {
			got := calc.Multiply(6, 7)
			assertFloat(t, got, 42, 1e-9)
		})

		t.Run("나눗셈", func(t *testing.T) {
			got, err := calc.Divide(15, 3)
			if err != nil {
				t.Fatalf("예상치 못한 에러: %v", err)
			}
			assertFloat(t, got, 5, 1e-9)
		})
	})

	t.Run("이력기록", func(t *testing.T) {
		calc := NewCalculator()
		calc.Add(1, 2)
		calc.Subtract(5, 3)
		calc.Multiply(2, 4)

		history := calc.History()
		if len(history) != 3 {
			t.Errorf("이력 길이: got=%d, want=3", len(history))
		}
	})

	t.Run("이력초기화", func(t *testing.T) {
		calc := NewCalculator()
		calc.Add(1, 2)
		calc.Add(3, 4)
		calc.ClearHistory()

		history := calc.History()
		if len(history) != 0 {
			t.Errorf("초기화 후 이력 길이: got=%d, want=0", len(history))
		}
	})

	t.Run("0으로나누기에러", func(t *testing.T) {
		calc := NewCalculator()
		_, err := calc.Divide(10, 0)
		if err == nil {
			t.Fatal("에러가 예상되었지만 nil을 받았습니다")
		}
		if !errors.Is(err, ErrDivisionByZero) {
			t.Errorf("에러 타입: got=%v, want=ErrDivisionByZero", err)
		}
	})
}

// ─────────────────────────────────────────
// Abs 테스트
// ─────────────────────────────────────────

func TestAbs(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-3.14, 3.14},
		{3.14, 3.14},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Abs(%.2f)", tt.input), func(t *testing.T) {
			got := Abs(tt.input)
			assertFloat(t, got, tt.want, 1e-9)
		})
	}
}

