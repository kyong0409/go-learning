// solution/calculator.go: 계산기 과제 참고 풀이
//
// 이 파일은 참고용입니다. 먼저 스스로 구현해 보세요!
package calculator

import (
	"errors"
	"math"
)

// 에러 변수
var ErrDivisionByZero = errors.New("0으로 나눌 수 없습니다")
var ErrNegativeInput = errors.New("음수는 허용되지 않습니다")

// Add: 덧셈
func Add(a, b float64) float64 {
	return a + b
}

// Subtract: 뺄셈
func Subtract(a, b float64) float64 {
	return a - b
}

// Multiply: 곱셈
func Multiply(a, b float64) float64 {
	return a * b
}

// Divide: 나눗셈 (0으로 나누면 에러)
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// Power: 거듭제곱
func Power(base, exp float64) float64 {
	return math.Pow(base, exp)
}

// Sqrt: 제곱근 (음수면 에러)
func Sqrt(n float64) (float64, error) {
	if n < 0 {
		return 0, ErrNegativeInput
	}
	return math.Sqrt(n), nil
}

// Abs: 절댓값
func Abs(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
}

// Modulo: 나머지 (0으로 나누면 에러)
func Modulo(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a % b, nil
}
