// calculator 패키지: 사칙연산 및 수학 함수 구현 과제
//
// 각 함수의 본문을 구현하세요.
// 테스트 실행: go test -v
package calculator

import "errors"

// ─────────────────────────────────────────
// 에러 변수 (수정하지 마세요)
// ─────────────────────────────────────────

// ErrDivisionByZero: 0으로 나누기 에러
var ErrDivisionByZero = errors.New("0으로 나눌 수 없습니다")

// ErrNegativeInput: 음수 입력 에러 (제곱근 등)
var ErrNegativeInput = errors.New("음수는 허용되지 않습니다")

// ─────────────────────────────────────────
// 구현할 함수들
// ─────────────────────────────────────────

// Add: a와 b를 더합니다.
// 예: Add(3, 4) == 7.0
func Add(a, b float64) float64 {
	// TODO: 구현하세요
	return 0
}

// Subtract: a에서 b를 뺍니다.
// 예: Subtract(10, 3) == 7.0
func Subtract(a, b float64) float64 {
	// TODO: 구현하세요
	return 0
}

// Multiply: a와 b를 곱합니다.
// 예: Multiply(4, 5) == 20.0
func Multiply(a, b float64) float64 {
	// TODO: 구현하세요
	return 0
}

// Divide: a를 b로 나눕니다.
// b가 0이면 (0, ErrDivisionByZero)를 반환합니다.
// 예: Divide(10, 4) == (2.5, nil)
func Divide(a, b float64) (float64, error) {
	// TODO: 구현하세요
	// 힌트: b == 0 이면 에러 반환
	return 0, nil
}

// Power: base의 exp 거듭제곱을 반환합니다.
// 예: Power(2, 10) == 1024.0
// 힌트: math.Pow 사용 (import "math" 필요)
func Power(base, exp float64) float64 {
	// TODO: 구현하세요
	return 0
}

// Sqrt: n의 제곱근을 반환합니다.
// n이 음수이면 (0, ErrNegativeInput)를 반환합니다.
// 예: Sqrt(16) == (4.0, nil)
// 힌트: math.Sqrt 사용
func Sqrt(n float64) (float64, error) {
	// TODO: 구현하세요
	return 0, nil
}

// Abs: n의 절댓값을 반환합니다.
// 예: Abs(-5.5) == 5.5, Abs(3.2) == 3.2
func Abs(n float64) float64 {
	// TODO: 구현하세요
	return 0
}

// Modulo: a를 b로 나눈 나머지를 반환합니다.
// b가 0이면 (0, ErrDivisionByZero)를 반환합니다.
// 예: Modulo(10, 3) == (1, nil)
func Modulo(a, b int) (int, error) {
	// TODO: 구현하세요
	return 0, nil
}
