// 패키지 선언: 모든 Go 프로그램은 패키지에 속합니다.
// main 패키지는 실행 가능한 프로그램의 진입점입니다.
package main

// import: 다른 패키지를 가져옵니다.
// fmt 패키지는 형식화된 I/O (입출력) 기능을 제공합니다.
import (
	"fmt"
	"math"
	"os"
)

// main 함수: 프로그램의 시작점입니다.
// Go 프로그램은 항상 main 패키지의 main() 함수에서 시작됩니다.
func main() {
	fmt.Println("=== Go 기초: Hello World와 fmt 패키지 ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 1. fmt.Println: 줄바꿈 포함 출력
	// ─────────────────────────────────────────
	fmt.Println("--- 1. fmt.Println ---")

	// Println은 인자들 사이에 자동으로 공백을 추가하고 마지막에 줄바꿈을 추가합니다.
	fmt.Println("안녕하세요, 세계!")                  // 문자열 출력
	fmt.Println("값:", 42, true, 3.14)               // 여러 타입 동시 출력
	fmt.Println("첫 번째 줄")                          // 각 Println은 새 줄에서 시작
	fmt.Println("두 번째 줄")
	fmt.Println() // 빈 줄 출력

	// ─────────────────────────────────────────
	// 2. fmt.Print: 줄바꿈 없는 출력
	// ─────────────────────────────────────────
	fmt.Println("--- 2. fmt.Print ---")

	// Print는 줄바꿈을 추가하지 않습니다.
	// 숫자가 아닌 타입 사이에만 공백이 추가됩니다.
	fmt.Print("Hello ")
	fmt.Print("World")
	fmt.Print("\n") // 수동으로 줄바꿈
	fmt.Print("하나 ", "둘 ", "셋\n")
	fmt.Println() // 구분용 빈 줄

	// ─────────────────────────────────────────
	// 3. fmt.Printf: 형식 지정 출력
	// ─────────────────────────────────────────
	fmt.Println("--- 3. fmt.Printf (형식 동사 verbs) ---")

	name := "Go 언어"
	version := 1.21
	year := 2009
	isAwesome := true

	// %s: 문자열
	fmt.Printf("언어 이름: %s\n", name)

	// %d: 정수 (decimal)
	fmt.Printf("출시 연도: %d년\n", year)

	// %f: 부동소수점
	fmt.Printf("버전: %f\n", version)

	// %.2f: 소수점 2자리까지 출력
	fmt.Printf("버전 (소수점 2자리): %.2f\n", version)

	// %t: 불리언
	fmt.Printf("훌륭한가요? %t\n", isAwesome)

	// %v: 기본 형식 (Go가 타입에 맞게 자동 선택)
	fmt.Printf("기본 형식: %v, %v, %v, %v\n", name, version, year, isAwesome)

	// %+v: 구조체일 때 필드 이름도 함께 출력 (나중에 유용)
	// %#v: Go 문법 형식으로 출력
	fmt.Printf("Go 문법 형식: %#v\n", name)

	// %T: 타입 출력
	fmt.Printf("타입들: %T, %T, %T, %T\n", name, version, year, isAwesome)

	// %p: 포인터 주소
	x := 42
	fmt.Printf("변수 x의 주소: %p\n", &x)

	// 너비와 정렬
	fmt.Printf("오른쪽 정렬 (너비 10): '%10s'\n", "Go")
	fmt.Printf("왼쪽 정렬 (너비 10): '%-10s'\n", "Go")
	fmt.Printf("0으로 패딩 (너비 5): '%05d'\n", 42)
	fmt.Println()

	// ─────────────────────────────────────────
	// 4. fmt.Sprintf: 문자열로 형식화 (출력 안 함)
	// ─────────────────────────────────────────
	fmt.Println("--- 4. fmt.Sprintf (문자열 반환) ---")

	// Sprintf는 Printf와 같지만 출력하지 않고 문자열을 반환합니다.
	greeting := fmt.Sprintf("안녕하세요, %s! 당신은 %d살이군요.", "홍길동", 25)
	fmt.Println(greeting)

	pi := math.Pi
	piStr := fmt.Sprintf("파이 값: %.4f", pi)
	fmt.Println(piStr)

	// Sprintf로 만든 문자열은 변수에 저장하거나 다른 함수에 전달할 수 있습니다.
	message := fmt.Sprintf("총 %d개의 아이템이 있고, %.1f%%가 완료되었습니다.", 100, 75.5)
	fmt.Println(message)
	fmt.Println()

	// ─────────────────────────────────────────
	// 5. fmt.Fprintf: 특정 Writer에 출력
	// ─────────────────────────────────────────
	fmt.Println("--- 5. fmt.Fprintf (Writer에 출력) ---")

	// Fprintf는 os.Stdout, os.Stderr, 파일 등 io.Writer 인터페이스를 구현한 모든 곳에 출력합니다.
	fmt.Fprintf(os.Stdout, "표준 출력(stdout)으로: %s\n", "Hello, Stdout!")
	fmt.Fprintf(os.Stderr, "표준 에러(stderr)로: %s\n", "Hello, Stderr!")
	fmt.Println()

	// ─────────────────────────────────────────
	// 6. fmt.Scan, fmt.Scanln, fmt.Scanf: 입력 받기
	// ─────────────────────────────────────────
	fmt.Println("--- 6. fmt.Sscan (문자열에서 파싱) ---")

	// Sscan: 문자열에서 값을 파싱합니다 (실제 콘솔 입력 예시 대신 사용)
	var parsedName string
	var parsedAge int
	input := "길동 30"
	n, err := fmt.Sscan(input, &parsedName, &parsedAge)
	if err != nil {
		fmt.Printf("파싱 에러: %v\n", err)
	} else {
		fmt.Printf("파싱 성공: %d개 값 읽음 - 이름=%s, 나이=%d\n", n, parsedName, parsedAge)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 7. fmt.Errorf: 에러 메시지 형식화
	// ─────────────────────────────────────────
	fmt.Println("--- 7. fmt.Errorf (에러 생성) ---")

	// Errorf는 형식화된 에러 메시지를 만듭니다.
	userID := 404
	notFoundErr := fmt.Errorf("사용자 ID %d를 찾을 수 없습니다", userID)
	fmt.Printf("에러 타입: %T\n", notFoundErr)
	fmt.Printf("에러 메시지: %v\n", notFoundErr)
	fmt.Println()

	// ─────────────────────────────────────────
	// 8. 특수 문자 이스케이프
	// ─────────────────────────────────────────
	fmt.Println("--- 8. 특수 문자 이스케이프 ---")
	fmt.Println("탭:\t[여기]")
	fmt.Println("줄바꿈:\n[다음 줄]")
	fmt.Println("큰따옴표: \"Go\"")
	fmt.Println("백슬래시: C:\\Users\\홍길동")
	fmt.Println("퍼센트 기호 출력: 100%%")

	// 원시 문자열 리터럴 (raw string literal): 백틱(`) 사용
	// 이스케이프 시퀀스가 처리되지 않습니다.
	rawStr := `이것은 "원시 문자열"입니다.
줄바꿈도 그대로입니다.
탭 \t 도 이스케이프되지 않습니다.`
	fmt.Println("원시 문자열:")
	fmt.Println(rawStr)
	fmt.Println()

	// ─────────────────────────────────────────
	// 9. 프로그램 종료
	// ─────────────────────────────────────────
	fmt.Println("=== 프로그램 정상 종료 ===")
	// os.Exit(0): 정상 종료 (0 = 성공, 1 이상 = 에러)
	// main()이 정상적으로 끝나면 자동으로 0으로 종료됩니다.
}
