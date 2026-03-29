// example_test.go: Example 함수 테스트
//
// Example 함수는 두 가지 역할을 합니다:
// 1. 문서화: go doc 명령어와 pkg.go.dev에 예제로 표시됩니다.
// 2. 테스트: // Output: 주석과 실제 출력이 일치하는지 자동으로 검증됩니다.
//
// 명명 규칙:
// - Example()         : 패키지 예제
// - ExampleAdd()      : Add 함수 예제
// - ExampleAdd_case1(): Add 함수의 두 번째 예제 (접미사로 구분)
package calc_test // 외부 테스트 패키지 (_test 접미사)

import (
	"errors"
	"fmt"

	"curriculum/phase2/testing/calc"
)

// Example은 패키지 전체 사용 예제입니다.
func Example() {
	// 기본 사칙연산
	sum := calc.Add(3, 4)
	diff := calc.Subtract(10, 3)
	product := calc.Multiply(6, 7)
	quotient, _ := calc.Divide(15, 3)

	fmt.Printf("덧셈: %.0f\n", sum)
	fmt.Printf("뺄셈: %.0f\n", diff)
	fmt.Printf("곱셈: %.0f\n", product)
	fmt.Printf("나눗셈: %.0f\n", quotient)

	// Output:
	// 덧셈: 7
	// 뺄셈: 7
	// 곱셈: 42
	// 나눗셈: 5
}

// ExampleAdd는 Add 함수의 예제입니다.
func ExampleAdd() {
	result := calc.Add(3, 4)
	fmt.Println(result)

	// Output:
	// 7
}

// ExampleSubtract는 Subtract 함수의 예제입니다.
func ExampleSubtract() {
	result := calc.Subtract(10, 3)
	fmt.Println(result)

	// Output:
	// 7
}

// ExampleMultiply는 Multiply 함수의 예제입니다.
func ExampleMultiply() {
	result := calc.Multiply(6, 7)
	fmt.Println(result)

	// Output:
	// 42
}

// ExampleDivide는 Divide 함수의 정상 케이스 예제입니다.
func ExampleDivide() {
	result, err := calc.Divide(10, 2)
	if err != nil {
		fmt.Println("에러:", err)
		return
	}
	fmt.Println(result)

	// Output:
	// 5
}

// ExampleDivide_zero는 Divide 함수의 0으로 나누기 에러 예제입니다.
// 접미사(_zero)로 두 번째 예제임을 나타냅니다.
func ExampleDivide_zero() {
	_, err := calc.Divide(10, 0)
	if errors.Is(err, calc.ErrDivisionByZero) {
		fmt.Println("0으로 나눌 수 없습니다")
	}

	// Output:
	// 0으로 나눌 수 없습니다
}

// ExampleSqrt는 Sqrt 함수의 예제입니다.
func ExampleSqrt() {
	result, err := calc.Sqrt(9)
	if err != nil {
		fmt.Println("에러:", err)
		return
	}
	fmt.Printf("%.0f\n", result)

	// Output:
	// 3
}

// ExampleSqrt_negative는 Sqrt 함수의 음수 입력 예제입니다.
func ExampleSqrt_negative() {
	_, err := calc.Sqrt(-4)
	if errors.Is(err, calc.ErrNegativeSqrt) {
		fmt.Println("음수의 제곱근은 실수가 아닙니다")
	}

	// Output:
	// 음수의 제곱근은 실수가 아닙니다
}

// ExampleCalculator는 Calculator 구조체 사용 예제입니다.
func ExampleCalculator() {
	c := calc.NewCalculator()

	c.Add(5, 3)
	c.Multiply(2, 6)
	c.Subtract(10, 4)

	fmt.Printf("이력 수: %d\n", len(c.History()))

	// Output:
	// 이력 수: 3
}
