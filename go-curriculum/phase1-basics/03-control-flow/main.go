// 패키지 선언
package main

import "fmt"

func main() {
	fmt.Println("=== Go 기초: 제어 흐름 (Control Flow) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 1. for 루프 - C 스타일 (3-component loop)
	// ─────────────────────────────────────────
	fmt.Println("--- 1. for 루프: C 스타일 ---")

	// Go에는 while이 없습니다. for 하나로 모든 반복을 처리합니다.
	// 기본 형태: for 초기화; 조건; 후처리 { }
	for i := 0; i < 5; i++ {
		fmt.Printf("i = %d\n", i)
	}

	// 카운트다운
	fmt.Println("카운트다운:")
	for i := 5; i > 0; i-- {
		fmt.Printf("%d ", i)
	}
	fmt.Println("발사!")

	// 2씩 증가
	fmt.Print("짝수: ")
	for i := 0; i <= 10; i += 2 {
		fmt.Printf("%d ", i)
	}
	fmt.Println()
	fmt.Println()

	// ─────────────────────────────────────────
	// 2. for 루프 - while 스타일 (조건만)
	// ─────────────────────────────────────────
	fmt.Println("--- 2. for 루프: while 스타일 ---")

	// 초기화와 후처리 없이 조건만 사용하면 while처럼 동작합니다.
	n := 1
	for n < 100 {
		n *= 2
	}
	fmt.Printf("1부터 시작해 2배씩: %d (100 이상 첫 번째 값)\n", n)

	// 무한 루프 + break
	fmt.Println("무한 루프 예시:")
	count := 0
	for {
		count++
		if count >= 3 {
			fmt.Printf("  count=%d에서 break\n", count)
			break
		}
		fmt.Printf("  count=%d\n", count)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 3. for 루프 - range
	// ─────────────────────────────────────────
	fmt.Println("--- 3. for 루프: range ---")

	// range는 슬라이스, 배열, 맵, 문자열, 채널을 순회합니다.
	// 슬라이스/배열: (인덱스, 값) 반환
	fruits := []string{"사과", "바나나", "체리", "두리안"}
	fmt.Println("과일 목록:")
	for i, fruit := range fruits {
		fmt.Printf("  [%d] %s\n", i, fruit)
	}

	// 인덱스만 필요할 때: 값을 _ 로 무시
	fmt.Print("인덱스만: ")
	for i := range fruits {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// 값만 필요할 때: 인덱스를 _ 로 무시
	fmt.Print("값만: ")
	for _, fruit := range fruits {
		fmt.Printf("%s ", fruit)
	}
	fmt.Println()

	// 문자열 range: 룬(rune/Unicode 코드 포인트) 단위로 순회
	fmt.Println("\n문자열 range (룬 단위):")
	msg := "Hello, 한국!"
	for i, r := range msg {
		fmt.Printf("  인덱스[%2d]: U+%04X = '%c'\n", i, r, r)
	}
	// 참고: 한국어 문자는 3바이트이므로 인덱스가 1씩 증가하지 않음

	// 맵 range
	fmt.Println("\n맵 range:")
	scores := map[string]int{
		"Alice": 95,
		"Bob":   87,
		"Carol": 92,
	}
	for name, score := range scores {
		fmt.Printf("  %s: %d점\n", name, score)
	}

	// Go 1.22+: 정수 range (for i := range N)
	fmt.Print("\n정수 range (0~4): ")
	for i := range 5 {
		fmt.Printf("%d ", i)
	}
	fmt.Println()
	fmt.Println()

	// ─────────────────────────────────────────
	// 4. break와 continue
	// ─────────────────────────────────────────
	fmt.Println("--- 4. break와 continue ---")

	// continue: 현재 반복을 건너뛰고 다음 반복으로
	fmt.Print("홀수만 출력: ")
	for i := 1; i <= 10; i++ {
		if i%2 == 0 {
			continue // 짝수이면 건너뜀
		}
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// 레이블(label) + break: 중첩 루프 탈출
	fmt.Println("레이블 break (중첩 루프 탈출):")
outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == 1 && j == 1 {
				fmt.Printf("  i=%d, j=%d에서 outer break\n", i, j)
				break outer // 바깥 루프까지 탈출
			}
			fmt.Printf("  i=%d, j=%d\n", i, j)
		}
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 5. if 문
	// ─────────────────────────────────────────
	fmt.Println("--- 5. if 문 ---")

	// 기본 if-else
	temperature := 25
	if temperature >= 30 {
		fmt.Println("덥습니다.")
	} else if temperature >= 20 {
		fmt.Println("따뜻합니다.")
	} else if temperature >= 10 {
		fmt.Println("시원합니다.")
	} else {
		fmt.Println("춥습니다.")
	}

	// if with init statement (초기화 구문이 있는 if)
	// 변수를 if 블록 범위(scope) 내에서만 사용할 수 있습니다.
	fmt.Println("\nif 초기화 구문:")
	if val := getScore(); val >= 90 {
		fmt.Printf("점수 %d: 우수!\n", val)
	} else if val >= 60 {
		fmt.Printf("점수 %d: 합격\n", val)
	} else {
		fmt.Printf("점수 %d: 불합격\n", val)
	}
	// fmt.Println(val) // 에러! val은 if 블록 밖에서 접근 불가

	// 에러 처리 패턴에서 자주 사용
	fmt.Println("\n에러 처리 패턴:")
	if result, err := divide(10, 3); err != nil {
		fmt.Printf("에러: %v\n", err)
	} else {
		fmt.Printf("10 / 3 = %.4f\n", result)
	}

	if _, err := divide(10, 0); err != nil {
		fmt.Printf("에러: %v\n", err)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 6. switch 문 - Expression switch
	// ─────────────────────────────────────────
	fmt.Println("--- 6. switch: Expression switch ---")

	// Go의 switch는 C와 달리 자동으로 break됩니다.
	// fallthrough가 필요한 경우에만 명시적으로 작성합니다.
	day := "월요일"
	switch day {
	case "월요일":
		fmt.Println("한 주의 시작!")
	case "금요일":
		fmt.Println("주말이 코앞!")
	case "토요일", "일요일": // 여러 값을 쉼표로 묶기
		fmt.Println("주말!")
	default:
		fmt.Println("평범한 평일")
	}

	// switch with init statement
	switch grade := getScore(); {
	case grade >= 90:
		fmt.Printf("점수 %d: A학점\n", grade)
	case grade >= 80:
		fmt.Printf("점수 %d: B학점\n", grade)
	case grade >= 70:
		fmt.Printf("점수 %d: C학점\n", grade)
	default:
		fmt.Printf("점수 %d: F학점\n", grade)
	}

	// 조건 없는 switch (if-else if 체인 대체)
	hour := 14
	switch {
	case hour < 12:
		fmt.Println("오전입니다.")
	case hour < 18:
		fmt.Println("오후입니다.")
	default:
		fmt.Println("저녁입니다.")
	}

	// fallthrough: 명시적으로 다음 case 실행
	fmt.Println("\nfallthrough 예시:")
	num2 := 3
	switch num2 {
	case 1:
		fmt.Println("case 1")
	case 2:
		fmt.Println("case 2")
	case 3:
		fmt.Println("case 3")
		fallthrough // 다음 case도 실행 (조건 무시)
	case 4:
		fmt.Println("case 4 (fallthrough로 실행됨)")
	case 5:
		fmt.Println("case 5")
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 7. switch 문 - Type switch
	// ─────────────────────────────────────────
	fmt.Println("--- 7. switch: Type switch ---")

	// interface{} 타입의 변수가 실제로 어떤 타입인지 확인
	values := []interface{}{42, "hello", 3.14, true, nil, []int{1, 2, 3}}
	for _, v := range values {
		describeType(v)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 8. goto (드물게 사용)
	// ─────────────────────────────────────────
	fmt.Println("--- 8. goto (참고용) ---")
	// Go는 goto를 지원하지만 일반적으로 사용을 권장하지 않습니다.
	// 에러 처리 정리 코드(cleanup code)에서 가끔 사용됩니다.
	i := 0
loop:
	if i < 3 {
		fmt.Printf("goto loop: i=%d\n", i)
		i++
		goto loop
	}
	fmt.Println("goto 루프 종료")
}

// 점수를 반환하는 헬퍼 함수
func getScore() int {
	return 85
}

// 나눗셈 함수 (에러 반환 예시)
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("0으로 나눌 수 없습니다")
	}
	return a / b, nil
}

// 타입 스위치 예시 함수
func describeType(v interface{}) {
	switch t := v.(type) {
	case int:
		fmt.Printf("  정수: %d\n", t)
	case string:
		fmt.Printf("  문자열: %q (길이: %d)\n", t, len(t))
	case float64:
		fmt.Printf("  부동소수점: %.2f\n", t)
	case bool:
		fmt.Printf("  불리언: %t\n", t)
	case nil:
		fmt.Println("  nil 값")
	default:
		fmt.Printf("  알 수 없는 타입: %T = %v\n", t, t)
	}
}
