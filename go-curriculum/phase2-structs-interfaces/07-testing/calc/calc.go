// Package calc는 기본 사칙연산 기능을 제공합니다.
// 이 패키지는 테스트 작성 방법을 학습하기 위한 예제입니다.
package calc

import (
	"errors"
	"fmt"
	"math"
)

// ─────────────────────────────────────────
// 에러 정의
// ─────────────────────────────────────────

// ErrDivisionByZero는 0으로 나누기 시도 시 반환됩니다.
var ErrDivisionByZero = errors.New("0으로 나눌 수 없습니다")

// ErrNegativeSqrt는 음수의 제곱근 계산 시 반환됩니다.
var ErrNegativeSqrt = errors.New("음수의 제곱근은 실수가 아닙니다")

// ErrOverflow는 결과가 float64 범위를 초과할 때 반환됩니다.
var ErrOverflow = errors.New("계산 결과가 범위를 초과합니다")

// ─────────────────────────────────────────
// Calculator 타입
// ─────────────────────────────────────────

// Calculator는 계산기 구조체입니다.
type Calculator struct {
	// history는 계산 이력을 저장합니다.
	history []string
}

// NewCalculator는 Calculator 생성자입니다.
func NewCalculator() *Calculator {
	return &Calculator{
		history: make([]string, 0),
	}
}

// History는 계산 이력을 반환합니다.
func (c *Calculator) History() []string {
	// 복사본 반환 (외부에서 수정 방지)
	result := make([]string, len(c.history))
	copy(result, c.history)
	return result
}

// ClearHistory는 계산 이력을 초기화합니다.
func (c *Calculator) ClearHistory() {
	c.history = c.history[:0]
}

// record는 계산 이력을 기록합니다.
func (c *Calculator) record(expr string) {
	c.history = append(c.history, expr)
}

// ─────────────────────────────────────────
// 기본 연산 함수 (패키지 수준)
// ─────────────────────────────────────────

// Add는 두 수를 더합니다.
func Add(a, b float64) float64 {
	return a + b
}

// Subtract는 a에서 b를 뺍니다.
func Subtract(a, b float64) float64 {
	return a - b
}

// Multiply는 두 수를 곱합니다.
func Multiply(a, b float64) float64 {
	return a * b
}

// Divide는 a를 b로 나눕니다.
// b가 0이면 ErrDivisionByZero를 반환합니다.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// Sqrt는 n의 제곱근을 계산합니다.
// n이 음수이면 ErrNegativeSqrt를 반환합니다.
func Sqrt(n float64) (float64, error) {
	if n < 0 {
		return 0, fmt.Errorf("Sqrt(%.2f): %w", n, ErrNegativeSqrt)
	}
	return math.Sqrt(n), nil
}

// Power는 base의 exp 제곱을 계산합니다.
func Power(base, exp float64) (float64, error) {
	result := math.Pow(base, exp)
	if math.IsInf(result, 0) {
		return 0, fmt.Errorf("Power(%.2f, %.2f): %w", base, exp, ErrOverflow)
	}
	if math.IsNaN(result) {
		return 0, fmt.Errorf("Power(%.2f, %.2f): 결과가 NaN입니다", base, exp)
	}
	return result, nil
}

// Abs는 절댓값을 반환합니다.
func Abs(n float64) float64 {
	return math.Abs(n)
}

// ─────────────────────────────────────────
// Calculator 메서드 (이력 기록 포함)
// ─────────────────────────────────────────

// Add는 두 수를 더하고 이력을 기록합니다.
func (c *Calculator) Add(a, b float64) float64 {
	result := Add(a, b)
	c.record(fmt.Sprintf("%.4g + %.4g = %.4g", a, b, result))
	return result
}

// Subtract는 a에서 b를 빼고 이력을 기록합니다.
func (c *Calculator) Subtract(a, b float64) float64 {
	result := Subtract(a, b)
	c.record(fmt.Sprintf("%.4g - %.4g = %.4g", a, b, result))
	return result
}

// Multiply는 두 수를 곱하고 이력을 기록합니다.
func (c *Calculator) Multiply(a, b float64) float64 {
	result := Multiply(a, b)
	c.record(fmt.Sprintf("%.4g × %.4g = %.4g", a, b, result))
	return result
}

// Divide는 a를 b로 나누고 이력을 기록합니다.
func (c *Calculator) Divide(a, b float64) (float64, error) {
	result, err := Divide(a, b)
	if err != nil {
		c.record(fmt.Sprintf("%.4g ÷ %.4g = 에러: %v", a, b, err))
		return 0, err
	}
	c.record(fmt.Sprintf("%.4g ÷ %.4g = %.4g", a, b, result))
	return result, nil
}
