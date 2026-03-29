// main 패키지: 09-packages 예제의 진입점
//
// 이 파일은 mathutil 패키지를 임포트하여 사용합니다.
// 실행: go run main.go  (go.mod가 있는 디렉토리에서)
package main

import (
	"errors"
	"fmt"

	// 로컬 패키지 임포트: 모듈 경로 + 패키지 경로
	"go-curriculum/phase1/packages/mathutil"
)

func main() {
	fmt.Println("=== Go 기초: 패키지 (Packages) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 1. 공개 함수 사용
	// ─────────────────────────────────────────
	fmt.Println("--- 1. mathutil 패키지 기본 함수 ---")

	// 패키지명.함수명 형식으로 호출
	fmt.Printf("Add(3, 4) = %d\n", mathutil.Add(3, 4))
	fmt.Printf("Subtract(10, 3) = %d\n", mathutil.Subtract(10, 3))
	fmt.Printf("Multiply(6, 7) = %d\n", mathutil.Multiply(6, 7))

	// 에러를 반환하는 함수
	q, err := mathutil.Divide(15, 4)
	if err != nil {
		fmt.Printf("Divide 에러: %v\n", err)
	} else {
		fmt.Printf("Divide(15, 4) = %d\n", q)
	}

	_, err = mathutil.Divide(10, 0)
	if err != nil {
		// 센티넬 에러 비교
		if errors.Is(err, mathutil.ErrDivisionByZero) {
			fmt.Printf("0으로 나누기 에러 감지: %v\n", err)
		}
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 2. 공개 상수 사용
	// ─────────────────────────────────────────
	fmt.Println("--- 2. 공개 상수 ---")
	fmt.Printf("mathutil.Pi = %.10f\n", mathutil.Pi)
	fmt.Printf("mathutil.E  = %.10f\n", mathutil.E)

	// 원 넓이 계산
	radius := 5.0
	area := mathutil.Pi * radius * radius
	fmt.Printf("반지름 %.1f인 원의 넓이: %.4f\n", radius, area)
	fmt.Println()

	// ─────────────────────────────────────────
	// 3. 수학 함수들
	// ─────────────────────────────────────────
	fmt.Println("--- 3. 수학 함수들 ---")

	fmt.Printf("Abs(-42) = %d\n", mathutil.Abs(-42))
	fmt.Printf("Abs(42)  = %d\n", mathutil.Abs(42))

	sqrt, err := mathutil.Sqrt(16.0)
	if err == nil {
		fmt.Printf("Sqrt(16) = %.2f\n", sqrt)
	}
	_, err = mathutil.Sqrt(-1.0)
	if errors.Is(err, mathutil.ErrNegativeInput) {
		fmt.Printf("Sqrt(-1) 에러: %v\n", err)
	}

	fmt.Printf("Power(2, 10) = %.0f\n", mathutil.Power(2, 10))
	fmt.Printf("Max(17, 42) = %d\n", mathutil.Max(17, 42))
	fmt.Printf("Min(17, 42) = %d\n", mathutil.Min(17, 42))
	fmt.Printf("Clamp(150, 0, 100) = %d\n", mathutil.Clamp(150, 0, 100))
	fmt.Printf("Clamp(-5, 0, 100) = %d\n", mathutil.Clamp(-5, 0, 100))
	fmt.Printf("Clamp(50, 0, 100) = %d\n", mathutil.Clamp(50, 0, 100))
	fmt.Println()

	// ─────────────────────────────────────────
	// 4. 슬라이스 연산
	// ─────────────────────────────────────────
	fmt.Println("--- 4. 슬라이스 연산 ---")

	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Printf("숫자: %v\n", nums)
	fmt.Printf("Sum: %d\n", mathutil.Sum(nums))

	avg, err := mathutil.Average(nums)
	if err == nil {
		fmt.Printf("Average: %.1f\n", avg)
	}

	_, err = mathutil.Average([]int{})
	if err != nil {
		fmt.Printf("빈 슬라이스 Average 에러: %v\n", err)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 5. 소수 판별
	// ─────────────────────────────────────────
	fmt.Println("--- 5. 소수 판별 ---")

	primes := []int{}
	for i := 2; i <= 30; i++ {
		if mathutil.IsPrime(i) {
			primes = append(primes, i)
		}
	}
	fmt.Printf("2~30 소수: %v\n", primes)
	fmt.Println()

	// ─────────────────────────────────────────
	// 6. 피보나치, GCD, LCM, 팩토리얼
	// ─────────────────────────────────────────
	fmt.Println("--- 6. 피보나치 / GCD / LCM / 팩토리얼 ---")

	fmt.Print("피보나치(0~10): ")
	for i := 0; i <= 10; i++ {
		f, _ := mathutil.Fibonacci(i)
		fmt.Printf("%d ", f)
	}
	fmt.Println()

	fmt.Printf("GCD(48, 18) = %d\n", mathutil.GCD(48, 18))
	fmt.Printf("GCD(100, 75) = %d\n", mathutil.GCD(100, 75))

	lcm, err := mathutil.LCM(4, 6)
	if err == nil {
		fmt.Printf("LCM(4, 6) = %d\n", lcm)
	}

	for n := 0; n <= 10; n++ {
		f, _ := mathutil.Factorial(n)
		fmt.Printf("%d! = %d\n", n, f)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 7. 패키지 개념 요약
	// ─────────────────────────────────────────
	fmt.Println("--- 7. 패키지 핵심 개념 요약 ---")
	fmt.Println("1. 대문자 이름 (Add, Pi, ErrDivisionByZero): 공개(exported) - 외부 패키지 접근 가능")
	fmt.Println("2. 소문자 이름 (gcd, factorial, maxIterations): 비공개(unexported) - 패키지 내부 전용")
	fmt.Println("3. import 경로는 go.mod의 module 이름 + 디렉토리 경로")
	fmt.Println("4. 패키지명은 디렉토리명과 일치시키는 것이 관례")
	fmt.Println("5. main 패키지만 실행 가능한 프로그램을 만들 수 있음")

	fmt.Println()
	fmt.Println("=== 완료 ===")
}
