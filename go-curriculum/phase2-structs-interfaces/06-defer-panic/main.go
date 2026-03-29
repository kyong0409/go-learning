// 패키지 선언
package main

import (
	"fmt"
	"os"
)

// ─────────────────────────────────────────
// 1. defer 기본
// ─────────────────────────────────────────
// defer는 함수가 반환되기 직전에 실행됩니다.
// 여러 defer는 LIFO(Last In, First Out) 순서로 실행됩니다.
// 주로 리소스 정리(파일 닫기, 잠금 해제 등)에 사용합니다.

func demoBasicDefer() {
	fmt.Println("  함수 시작")

	// defer는 선언 즉시 실행되지 않고, 함수 종료 시 실행됩니다.
	defer fmt.Println("  defer 1: 첫 번째 등록")
	defer fmt.Println("  defer 2: 두 번째 등록")
	defer fmt.Println("  defer 3: 세 번째 등록")

	fmt.Println("  함수 중간")
	fmt.Println("  함수 끝 (return 직전)")
	// 함수 반환 시: defer 3 → defer 2 → defer 1 순서로 실행 (LIFO)
}

// ─────────────────────────────────────────
// 2. defer와 반복문
// ─────────────────────────────────────────

func demoLoopDefer() {
	fmt.Println("  반복문 defer (LIFO 확인):")
	for i := 0; i < 5; i++ {
		// i 값이 defer 등록 시점에 평가됩니다 (값이 복사됨).
		defer fmt.Printf("    defer[%d]\n", i)
	}
	fmt.Println("  반복문 종료")
}

// ─────────────────────────────────────────
// 3. defer와 클로저
// ─────────────────────────────────────────
// defer에 클로저를 사용하면 변수를 캡처합니다.
// 캡처된 변수는 defer 실행 시점의 값을 사용합니다.

func demoClosureDefer() string {
	result := "초기값"

	// 클로저: result 변수를 캡처 (defer 실행 시점의 값 사용)
	defer func() {
		fmt.Printf("  클로저 defer에서 result 값: %q\n", result)
	}()

	result = "변경된 값"
	fmt.Printf("  함수 내에서 result: %q\n", result)
	return result
	// defer 실행 시 result는 "변경된 값" 입니다.
}

// ─────────────────────────────────────────
// 4. defer로 반환값 수정 (Named Return)
// ─────────────────────────────────────────
// 이름 있는 반환값(named return)과 defer를 결합하면
// 함수 반환 직전에 반환값을 수정할 수 있습니다.

// doubleResult는 반환값을 defer로 2배로 만듭니다.
func doubleResult() (result int) {
	defer func() {
		result *= 2 // 반환값을 수정
	}()

	result = 21
	return // result = 21 이지만 defer가 42로 만듦
}

// withErrorHandling은 defer로 에러를 처리합니다.
func withErrorHandling(input int) (result string, err error) {
	defer func() {
		if err != nil {
			// 에러가 발생하면 로그 추가
			err = fmt.Errorf("withErrorHandling(%d): %w", input, err)
		}
	}()

	if input < 0 {
		err = fmt.Errorf("음수 입력: %d", input)
		return
	}
	result = fmt.Sprintf("처리 결과: %d", input*2)
	return
}

// ─────────────────────────────────────────
// 5. defer로 파일 처리
// ─────────────────────────────────────────
// defer의 가장 일반적인 사용: 파일/리소스 닫기

// writeToFile은 defer로 파일을 안전하게 닫습니다.
func writeToFile(filename, content string) error {
	// 파일 열기
	f, err := os.CreateTemp("", filename)
	if err != nil {
		return fmt.Errorf("파일 생성 실패: %w", err)
	}
	// defer로 파일 닫기를 보장합니다.
	// 함수 어디서 return하든 파일이 닫힙니다.
	defer func() {
		f.Close()
		os.Remove(f.Name()) // 예제이므로 임시 파일 삭제
		fmt.Printf("  파일 닫힘: %s\n", f.Name())
	}()

	fmt.Printf("  파일 생성: %s\n", f.Name())

	// 파일에 쓰기
	_, err = f.WriteString(content)
	if err != nil {
		return fmt.Errorf("파일 쓰기 실패: %w", err) // defer가 파일을 닫음
	}

	fmt.Printf("  파일 쓰기 완료: %d bytes\n", len(content))
	return nil
	// return 후 defer 실행 → f.Close() 호출
}

// ─────────────────────────────────────────
// 6. panic 기본
// ─────────────────────────────────────────
// panic은 프로그램을 즉시 중단합니다.
// defer는 panic 중에도 실행됩니다.
//
// panic을 사용해야 하는 경우:
// - 프로그래머 실수로 절대 일어나면 안 되는 상황 (nil 맵 접근 등)
// - 초기화 실패 (서버 시작 시 필수 설정 없음)
//
// panic을 사용하지 말아야 하는 경우:
// - 일반적인 에러 처리 (error 반환 사용)
// - 외부 입력 오류 (사용자 입력, 파일 없음 등)

func mustPositive(n int) int {
	if n <= 0 {
		// 라이브러리 내부에서 절대 일어나면 안 되는 상황
		panic(fmt.Sprintf("mustPositive: 양수여야 하지만 %d를 받았습니다", n))
	}
	return n
}

// ─────────────────────────────────────────
// 7. recover - panic 복구
// ─────────────────────────────────────────
// recover는 panic을 잡아 프로그램 중단을 막습니다.
// 반드시 defer 내에서만 동작합니다.

// safeDiv는 패닉을 recover로 처리합니다.
func safeDiv(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			// panic 값을 에러로 변환
			err = fmt.Errorf("패닉 복구: %v", r)
		}
	}()

	// b가 0이면 런타임 패닉 발생 (divide by zero)
	result = a / b
	return
}

// runWithRecovery는 패닉이 발생할 수 있는 함수를 안전하게 실행합니다.
// 서버에서 한 요청의 패닉이 전체를 죽이지 않게 할 때 사용합니다.
func runWithRecovery(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// recover()의 반환값은 panic()에 전달된 값
			switch v := r.(type) {
			case error:
				err = fmt.Errorf("패닉(에러): %w", v)
			case string:
				err = fmt.Errorf("패닉(문자열): %s", v)
			default:
				err = fmt.Errorf("패닉(기타): %v", v)
			}
		}
	}()

	fn()
	return nil
}

// ─────────────────────────────────────────
// 8. panic/recover를 사용하는 실제 패턴
// ─────────────────────────────────────────
// JSON 파서처럼 내부적으로 panic을 사용하고
// 외부에는 error를 반환하는 패턴

// jsonParser는 내부에서 panic을 사용하는 예시입니다.
type jsonParser struct {
	data string
	pos  int
}

// parseError는 파서 에러를 나타냅니다.
type parseError struct {
	message string
	pos     int
}

// Error는 error 인터페이스를 구현합니다.
func (e *parseError) Error() string {
	return fmt.Sprintf("파싱 에러 (위치 %d): %s", e.pos, e.message)
}

// expect는 예상과 다른 문자면 패닉을 발생시킵니다.
func (p *jsonParser) expect(ch byte) {
	if p.pos >= len(p.data) || p.data[p.pos] != ch {
		panic(&parseError{
			message: fmt.Sprintf("'%c' 기대했지만 없음", ch),
			pos:     p.pos,
		})
	}
	p.pos++
}

// Parse는 외부에 error를 반환합니다 (내부 panic을 숨깁니다).
func (p *jsonParser) Parse() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*parseError); ok {
				err = pe
			} else {
				err = fmt.Errorf("예상치 못한 에러: %v", r)
			}
		}
	}()

	// 예시: 간단한 JSON 배열 파싱
	p.expect('[')
	p.expect(']')
	return nil
}

func main() {
	fmt.Println("=== Go Phase 2: defer / panic / recover ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// defer LIFO 순서
	// ─────────────────────────────────────────
	fmt.Println("--- 1. defer LIFO 실행 순서 ---")
	demoBasicDefer()
	// 출력: defer 3 → defer 2 → defer 1 순서 확인
	fmt.Println()

	// ─────────────────────────────────────────
	// defer와 반복문
	// ─────────────────────────────────────────
	fmt.Println("--- 2. defer와 반복문 ---")
	demoLoopDefer()
	// 출력: 4 → 3 → 2 → 1 → 0 순서 확인
	fmt.Println()

	// ─────────────────────────────────────────
	// defer와 클로저
	// ─────────────────────────────────────────
	fmt.Println("--- 3. defer와 클로저 ---")
	ret := demoClosureDefer()
	fmt.Printf("  반환값: %q\n", ret)
	// 클로저 defer는 return 후 값인 "변경된 값"을 봅니다.
	fmt.Println()

	// ─────────────────────────────────────────
	// Named Return + defer
	// ─────────────────────────────────────────
	fmt.Println("--- 4. Named Return + defer ---")

	result := doubleResult()
	fmt.Printf("  doubleResult(): %d (21 * 2 = 42)\n", result)

	r1, e1 := withErrorHandling(5)
	fmt.Printf("  withErrorHandling(5): result=%q, err=%v\n", r1, e1)

	r2, e2 := withErrorHandling(-3)
	fmt.Printf("  withErrorHandling(-3): result=%q, err=%v\n", r2, e2)
	fmt.Println()

	// ─────────────────────────────────────────
	// 파일 처리에서의 defer
	// ─────────────────────────────────────────
	fmt.Println("--- 5. 파일 처리에서의 defer ---")
	if err := writeToFile("example", "Hello, Go defer!\n안전한 파일 처리"); err != nil {
		fmt.Printf("  파일 쓰기 에러: %v\n", err)
	} else {
		fmt.Println("  파일 작업 완료")
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// safeDiv: recover로 나누기 패닉 처리
	// ─────────────────────────────────────────
	fmt.Println("--- 6. recover로 패닉 복구 ---")

	if result, err := safeDiv(10, 2); err != nil {
		fmt.Printf("  10/2 에러: %v\n", err)
	} else {
		fmt.Printf("  10/2 = %d\n", result)
	}

	if result, err := safeDiv(10, 0); err != nil {
		fmt.Printf("  10/0 패닉 복구: %v\n", err)
	} else {
		fmt.Printf("  10/0 = %d\n", result)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// runWithRecovery: 다양한 패닉 복구
	// ─────────────────────────────────────────
	fmt.Println("--- 7. runWithRecovery ---")

	// 정상 실행
	err := runWithRecovery(func() {
		fmt.Println("  정상 실행: 문제 없음")
	})
	fmt.Printf("  에러: %v\n", err)

	// string panic
	err = runWithRecovery(func() {
		panic("뭔가 심각한 문제!")
	})
	fmt.Printf("  string panic 복구: %v\n", err)

	// error panic
	err = runWithRecovery(func() {
		panic(fmt.Errorf("에러 패닉"))
	})
	fmt.Printf("  error panic 복구: %v\n", err)

	// 런타임 패닉 (nil 포인터)
	err = runWithRecovery(func() {
		var s *string
		_ = *s // nil 포인터 역참조
	})
	fmt.Printf("  nil 포인터 패닉 복구: %v\n", err)
	fmt.Println()

	// ─────────────────────────────────────────
	// 내부 panic, 외부 error 패턴
	// ─────────────────────────────────────────
	fmt.Println("--- 8. 내부 panic / 외부 error 패턴 ---")

	// 정상 파싱
	p1 := &jsonParser{data: "[]"}
	if err := p1.Parse(); err != nil {
		fmt.Printf("  파싱 에러: %v\n", err)
	} else {
		fmt.Println("  '[]' 파싱 성공")
	}

	// 잘못된 입력
	p2 := &jsonParser{data: "{]"}
	if err := p2.Parse(); err != nil {
		fmt.Printf("  '{]' 파싱 에러: %v\n", err)
	}

	fmt.Println()
	fmt.Println("=== defer/panic/recover 학습 완료 ===")
	fmt.Println()
	fmt.Println("핵심 정리:")
	fmt.Println("  defer  : LIFO 순서, 함수 종료 시 실행, 리소스 정리에 사용")
	fmt.Println("  panic  : 프로그래머 실수/불가능한 상황에만 사용, 일반 에러에는 error 반환")
	fmt.Println("  recover: defer 내에서만 동작, 패닉을 에러로 변환, HTTP 서버 등에서 유용")
}
