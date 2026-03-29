// mathutil 패키지: 수학 유틸리티 함수 모음
//
// Go에서 패키지는 코드를 논리적으로 구성하는 기본 단위입니다.
// 대문자로 시작하는 이름은 "공개(exported)" - 다른 패키지에서 접근 가능
// 소문자로 시작하는 이름은 "비공개(unexported)" - 이 패키지 내부에서만 접근 가능
package mathutil

import (
	"errors"
	"math"
)

// ─────────────────────────────────────────
// 공개 상수 (Exported Constants)
// ─────────────────────────────────────────

// Pi: 원주율 (공개 상수 - 외부에서 mathutil.Pi 로 접근 가능)
const Pi = math.Pi

// E: 자연 상수 (공개 상수)
const E = math.E

// 비공개 상수 (unexported - 이 패키지 내부에서만 사용)
const maxIterations = 1000

// ─────────────────────────────────────────
// 공개 에러 변수 (Exported Sentinel Errors)
// ─────────────────────────────────────────

// ErrDivisionByZero: 0으로 나누기 에러 (센티넬 에러)
var ErrDivisionByZero = errors.New("0으로 나눌 수 없습니다")

// ErrNegativeInput: 음수 입력 에러
var ErrNegativeInput = errors.New("음수 입력은 허용되지 않습니다")

// ErrInvalidRange: 범위 초과 에러
var ErrInvalidRange = errors.New("입력값이 유효한 범위를 벗어났습니다")

// ─────────────────────────────────────────
// 공개 함수 (Exported Functions)
// ─────────────────────────────────────────

// Add: 두 정수를 더합니다.
// 가장 간단한 공개 함수 예시입니다.
func Add(a, b int) int {
	return a + b
}

// Subtract: 두 정수를 뺍니다.
func Subtract(a, b int) int {
	return a - b
}

// Multiply: 두 정수를 곱합니다.
func Multiply(a, b int) int {
	return a * b
}

// Divide: 두 정수를 나눕니다. 0으로 나누면 에러를 반환합니다.
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// DivideFloat: 두 float64를 나눕니다.
func DivideFloat(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// Abs: 절댓값을 반환합니다.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// AbsFloat: float64의 절댓값을 반환합니다.
func AbsFloat(f float64) float64 {
	return math.Abs(f)
}

// Sqrt: 제곱근을 반환합니다. 음수 입력 시 에러를 반환합니다.
func Sqrt(n float64) (float64, error) {
	if n < 0 {
		return 0, ErrNegativeInput
	}
	return math.Sqrt(n), nil
}

// Power: base의 exp 거듭제곱을 반환합니다.
func Power(base, exp float64) float64 {
	return math.Pow(base, exp)
}

// Max: 두 정수 중 큰 값을 반환합니다.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min: 두 정수 중 작은 값을 반환합니다.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Clamp: 값을 [min, max] 범위로 제한합니다.
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Sum: 슬라이스의 모든 요소의 합을 반환합니다.
func Sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Average: 슬라이스의 평균을 반환합니다.
func Average(nums []int) (float64, error) {
	if len(nums) == 0 {
		return 0, errors.New("빈 슬라이스의 평균을 계산할 수 없습니다")
	}
	return float64(Sum(nums)) / float64(len(nums)), nil
}

// IsPrime: 주어진 수가 소수인지 확인합니다.
func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	// 비공개 헬퍼 함수 사용
	return isPrimeInternal(n)
}

// Fibonacci: n번째 피보나치 수를 반환합니다 (0-indexed).
// n이 음수이면 에러를 반환합니다.
func Fibonacci(n int) (int, error) {
	if n < 0 {
		return 0, ErrNegativeInput
	}
	return fibonacci(n), nil
}

// GCD: 두 수의 최대공약수(Greatest Common Divisor)를 반환합니다.
func GCD(a, b int) int {
	a = Abs(a)
	b = Abs(b)
	return gcd(a, b)
}

// LCM: 두 수의 최소공배수(Least Common Multiple)를 반환합니다.
func LCM(a, b int) (int, error) {
	if a == 0 || b == 0 {
		return 0, ErrDivisionByZero
	}
	return Abs(a*b) / GCD(a, b), nil
}

// Factorial: n!을 반환합니다.
func Factorial(n int) (int, error) {
	if n < 0 {
		return 0, ErrNegativeInput
	}
	return factorial(n), nil
}

// ─────────────────────────────────────────
// 비공개 함수 (Unexported Functions)
// 패키지 내부 구현 세부사항 - 외부에서 접근 불가
// ─────────────────────────────────────────

// isPrimeInternal: 소수 판별 내부 구현 (비공개)
// 3 이상의 홀수만 입력받는다고 가정
func isPrimeInternal(n int) bool {
	limit := int(math.Sqrt(float64(n)))
	for i := 3; i <= limit; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// fibonacci: 피보나치 내부 재귀 구현 (비공개)
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	// 메모이제이션으로 효율화
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// gcd: 유클리드 호제법으로 최대공약수 계산 (비공개)
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// factorial: 팩토리얼 내부 구현 (비공개)
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}
