// 패키지 선언
package main

import (
	"fmt"
	"sort"
	"strings"
)

// ─────────────────────────────────────────
// 패키지 수준 함수 선언
// ─────────────────────────────────────────

// 기본 함수: 매개변수와 반환값
func add(a, b int) int {
	return a + b
}

// 같은 타입의 매개변수는 타입을 한 번만 쓸 수 있습니다.
func multiply(a, b int) int {
	return a * b
}

// ─────────────────────────────────────────
// 1. 다중 반환값 (Multiple Return Values)
// ─────────────────────────────────────────

// Go의 함수는 여러 값을 반환할 수 있습니다.
// 관용적으로 마지막 반환값을 에러로 사용합니다.
func minMax(nums []int) (int, int) {
	if len(nums) == 0 {
		return 0, 0
	}
	min, max := nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return min, max
}

// 에러와 함께 반환
func safeDivide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("0으로 나눌 수 없습니다: %v / %v", a, b)
	}
	return a / b, nil
}

// ─────────────────────────────────────────
// 2. Named Return Values (명명된 반환값)
// ─────────────────────────────────────────

// 반환값에 이름을 붙이면 함수 시작 시 자동으로 제로값으로 초기화됩니다.
// naked return: return 키워드만 사용하면 명명된 반환값을 반환합니다.
func circleStats(radius float64) (area, circumference float64) {
	const pi = 3.14159265358979
	area = pi * radius * radius       // 면적
	circumference = 2 * pi * radius   // 둘레
	return // naked return: area, circumference 반환
}

// 명명된 반환값은 문서화 목적으로도 유용합니다.
func parseFullName(fullName string) (firstName, lastName string, err error) {
	parts := strings.Fields(fullName) // 공백으로 분리
	if len(parts) < 2 {
		err = fmt.Errorf("전체 이름이 필요합니다: %q", fullName)
		return // firstName="", lastName="", err=...
	}
	firstName = parts[0]
	lastName = parts[len(parts)-1]
	return
}

// ─────────────────────────────────────────
// 3. 가변 인자 함수 (Variadic Functions)
// ─────────────────────────────────────────

// ...타입: 0개 이상의 인자를 슬라이스로 받습니다.
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// 가변 인자와 일반 인자 혼합 (가변 인자는 항상 마지막)
func printWithPrefix(prefix string, values ...interface{}) {
	for _, v := range values {
		fmt.Printf("[%s] %v\n", prefix, v)
	}
}

// 슬라이스를 가변 인자로 전달: 슬라이스명...
func spreadExample() {
	nums := []int{1, 2, 3, 4, 5}
	result := sum(nums...) // 슬라이스를 펼쳐서 전달
	fmt.Printf("sum(%v...) = %d\n", nums, result)
}

// ─────────────────────────────────────────
// 4. 함수 타입과 일급 함수 (First-class Functions)
// ─────────────────────────────────────────

// 함수를 타입으로 선언
type MathFunc func(int, int) int
type Predicate func(int) bool

// 함수를 매개변수로 받기
func applyOperation(a, b int, op MathFunc) int {
	return op(a, b)
}

// 함수를 반환하는 함수 (함수 팩토리)
func makeAdder(n int) func(int) int {
	return func(x int) int {
		return x + n
	}
}

func makeMultiplier(n int) func(int) int {
	return func(x int) int {
		return x * n
	}
}

// 함수 슬라이스 사용
func applyAll(n int, funcs ...func(int) int) []int {
	results := make([]int, len(funcs))
	for i, f := range funcs {
		results[i] = f(n)
	}
	return results
}

// filter: 조건에 맞는 요소만 반환
func filter(nums []int, pred Predicate) []int {
	var result []int
	for _, n := range nums {
		if pred(n) {
			result = append(result, n)
		}
	}
	return result
}

// mapInts: 모든 요소에 함수 적용
func mapInts(nums []int, f func(int) int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = f(n)
	}
	return result
}

// ─────────────────────────────────────────
// 5. 클로저 (Closures)
// ─────────────────────────────────────────

// 클로저: 자신이 정의된 범위의 변수를 "캡처"하는 함수
// 카운터 클로저: 상태를 유지합니다.
func makeCounter() func() int {
	count := 0 // 이 변수는 반환된 함수에 의해 캡처됩니다.
	return func() int {
		count++
		return count
	}
}

// 누적 합산 클로저
func makeAccumulator() func(int) int {
	total := 0
	return func(n int) int {
		total += n
		return total
	}
}

// 메모이제이션 (Memoization) 클로저
func makeFibMemo() func(int) int {
	cache := map[int]int{}
	var fib func(int) int
	fib = func(n int) int {
		if n <= 1 {
			return n
		}
		if v, ok := cache[n]; ok {
			return v
		}
		result := fib(n-1) + fib(n-2)
		cache[n] = result
		return result
	}
	return fib
}

// ─────────────────────────────────────────
// 6. defer
// ─────────────────────────────────────────

// defer: 현재 함수가 반환되기 직전에 실행을 예약합니다.
// LIFO(후입선출) 순서로 실행됩니다.
func deferDemo() {
	fmt.Println("deferDemo 시작")
	defer fmt.Println("defer 1: 마지막에 실행") // 나중에 등록 -> 먼저 실행 (LIFO)
	defer fmt.Println("defer 2: 그 다음에")
	defer fmt.Println("defer 3: 이것이 먼저")
	fmt.Println("deferDemo 끝 (defer 전)")
}

// defer와 루프 - 클로저 캡처 주의
func deferLoopDemo() {
	fmt.Println("defer 루프 (클로저 캡처):")
	for i := 0; i < 3; i++ {
		i := i // 루프 변수를 새 변수로 캡처 (중요!)
		defer func() {
			fmt.Printf("  defer: i=%d\n", i)
		}()
	}
}

// defer로 패닉 복구
func safeOperation() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  패닉 복구: %v\n", r)
		}
	}()
	panic("의도적인 패닉!")
}

// defer로 리소스 정리 시뮬레이션
func simulateFileOp(filename string) error {
	fmt.Printf("  파일 열기: %s\n", filename)
	defer fmt.Printf("  파일 닫기: %s\n", filename) // 항상 실행됨

	// ... 파일 작업 ...
	fmt.Printf("  파일 처리 중: %s\n", filename)
	return nil
}

// ─────────────────────────────────────────
// 7. 재귀 함수 (Recursive Functions)
// ─────────────────────────────────────────

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────
func main() {
	fmt.Println("=== Go 기초: 함수 (Functions) ===")
	fmt.Println()

	// 1. 기본 함수 호출
	fmt.Println("--- 1. 기본 함수 ---")
	fmt.Printf("add(3, 4) = %d\n", add(3, 4))
	fmt.Printf("multiply(6, 7) = %d\n", multiply(6, 7))
	fmt.Println()

	// 2. 다중 반환값
	fmt.Println("--- 2. 다중 반환값 ---")
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	min, max := minMax(nums)
	fmt.Printf("숫자: %v\n최솟값: %d, 최댓값: %d\n", nums, min, max)

	// 불필요한 반환값은 _ 로 무시
	_, onlyMax := minMax(nums)
	fmt.Printf("최댓값만: %d\n", onlyMax)

	result, err := safeDivide(10, 3)
	if err != nil {
		fmt.Printf("에러: %v\n", err)
	} else {
		fmt.Printf("10 / 3 = %.4f\n", result)
	}

	_, err = safeDivide(5, 0)
	if err != nil {
		fmt.Printf("에러: %v\n", err)
	}
	fmt.Println()

	// 3. Named Return Values
	fmt.Println("--- 3. Named Return Values ---")
	area, circumference := circleStats(5.0)
	fmt.Printf("반지름 5인 원: 넓이=%.4f, 둘레=%.4f\n", area, circumference)

	first, last, err2 := parseFullName("홍 길동")
	if err2 != nil {
		fmt.Printf("파싱 에러: %v\n", err2)
	} else {
		fmt.Printf("이름: %s, 성: %s\n", first, last)
	}

	_, _, err3 := parseFullName("홍길동")
	if err3 != nil {
		fmt.Printf("파싱 에러: %v\n", err3)
	}
	fmt.Println()

	// 4. 가변 인자
	fmt.Println("--- 4. 가변 인자 (Variadic) ---")
	fmt.Printf("sum() = %d\n", sum())
	fmt.Printf("sum(1) = %d\n", sum(1))
	fmt.Printf("sum(1,2,3) = %d\n", sum(1, 2, 3))
	fmt.Printf("sum(1..10) = %d\n", sum(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
	spreadExample()

	printWithPrefix("INFO", "서버 시작", 8080, true)
	fmt.Println()

	// 5. 일급 함수
	fmt.Println("--- 5. 일급 함수 (First-class Functions) ---")

	// 함수를 변수에 할당
	addFn := add
	fmt.Printf("함수 변수: addFn(3,4) = %d\n", addFn(3, 4))

	// 함수를 인자로 전달
	ops := []struct {
		name string
		fn   MathFunc
	}{
		{"더하기", func(a, b int) int { return a + b }},
		{"빼기", func(a, b int) int { return a - b }},
		{"곱하기", func(a, b int) int { return a * b }},
	}
	for _, op := range ops {
		fmt.Printf("applyOperation(10, 3, %s) = %d\n",
			op.name, applyOperation(10, 3, op.fn))
	}

	// 함수 팩토리
	add5 := makeAdder(5)
	add10 := makeAdder(10)
	triple := makeMultiplier(3)
	fmt.Printf("\nadd5(3) = %d\n", add5(3))
	fmt.Printf("add10(3) = %d\n", add10(3))
	fmt.Printf("triple(4) = %d\n", triple(4))

	// filter와 map
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evens := filter(numbers, func(n int) bool { return n%2 == 0 })
	squared := mapInts(numbers, func(n int) int { return n * n })
	fmt.Printf("\n원본: %v\n짝수만: %v\n제곱: %v\n", numbers, evens, squared)

	// 함수 슬라이스로 정렬 비교기
	words := []string{"바나나", "사과", "체리", "포도"}
	sort.Slice(words, func(i, j int) bool {
		return words[i] < words[j] // 오름차순
	})
	fmt.Printf("정렬: %v\n", words)
	fmt.Println()

	// 6. 클로저
	fmt.Println("--- 6. 클로저 (Closures) ---")

	// 독립적인 카운터 인스턴스
	counter1 := makeCounter()
	counter2 := makeCounter()
	fmt.Printf("counter1: %d, %d, %d\n", counter1(), counter1(), counter1())
	fmt.Printf("counter2: %d, %d\n", counter2(), counter2())
	fmt.Printf("counter1 계속: %d\n", counter1()) // 독립적으로 유지

	// 누적 합산
	acc := makeAccumulator()
	fmt.Printf("\n누적 합산: %d, %d, %d, %d\n",
		acc(10), acc(20), acc(30), acc(40))

	// 메모이제이션 피보나치
	fib := makeFibMemo()
	fmt.Print("\n피보나치 (메모이제이션): ")
	for i := 0; i <= 10; i++ {
		fmt.Printf("fib(%d)=%d ", i, fib(i))
	}
	fmt.Println()
	fmt.Println()

	// 7. defer
	fmt.Println("--- 7. defer ---")
	deferDemo()
	fmt.Println()

	// defer + 파일 작업 시뮬레이션
	simulateFileOp("data.txt")
	fmt.Println()

	// defer + recover (패닉 처리)
	fmt.Println("패닉 복구 예시:")
	safeOperation()
	fmt.Println("패닉 이후 계속 실행됨")
	fmt.Println()

	// defer 루프 (별도 함수에서 실행하여 defer 확인)
	fmt.Println("defer 루프 (LIFO 순서):")
	deferLoopDemo()
	fmt.Println()

	// 8. 재귀
	fmt.Println("--- 8. 재귀 ---")
	fmt.Print("팩토리얼: ")
	for i := 0; i <= 10; i++ {
		fmt.Printf("%d!=%d ", i, factorial(i))
	}
	fmt.Println()

	fmt.Print("피보나치 (재귀): ")
	for i := 0; i <= 10; i++ {
		fmt.Printf("%d ", fibonacci(i))
	}
	fmt.Println()
	fmt.Println()

	fmt.Println("=== 완료 ===")
}
