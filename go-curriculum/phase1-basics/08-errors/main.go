// 패키지 선언
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// ─────────────────────────────────────────
// 커스텀 에러 타입 정의
// ─────────────────────────────────────────

// ValidationError: 유효성 검사 에러
type ValidationError struct {
	Field   string // 에러가 발생한 필드 이름
	Message string // 에러 메시지
	Value   any    // 잘못된 값
}

// error 인터페이스 구현: Error() string 메서드
func (e *ValidationError) Error() string {
	return fmt.Sprintf("유효성 검사 실패 - 필드 %q: %s (값: %v)", e.Field, e.Message, e.Value)
}

// NotFoundError: 찾을 수 없음 에러
type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s ID=%d를 찾을 수 없습니다", e.Resource, e.ID)
}

// ─────────────────────────────────────────
// 에러 래핑을 위한 커스텀 에러
// ─────────────────────────────────────────

// DatabaseError: 데이터베이스 에러 (래핑 예시)
type DatabaseError struct {
	Operation string
	Table     string
	Err       error // 래핑된 원인 에러
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("DB 에러 [%s on %s]: %v", e.Operation, e.Table, e.Err)
}

// Unwrap: errors.Is/As가 래핑된 에러를 탐색할 수 있게 합니다.
func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// ─────────────────────────────────────────
// 센티넬 에러 (Sentinel Errors)
// ─────────────────────────────────────────

// 패키지 수준에서 선언된 잘 알려진 에러값 (비교에 사용)
var (
	ErrNotFound     = errors.New("찾을 수 없음")
	ErrUnauthorized = errors.New("권한 없음")
	ErrInvalidInput = errors.New("잘못된 입력")
)

func main() {
	fmt.Println("=== Go 기초: 에러 처리 (Error Handling) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 1. error 인터페이스
	// ─────────────────────────────────────────
	fmt.Println("--- 1. error 인터페이스 ---")
	// error는 내장 인터페이스입니다:
	// type error interface {
	//     Error() string
	// }
	// nil error는 "에러 없음"을 의미합니다.

	var err error // nil
	fmt.Printf("nil error: %v, nil=%t\n", err, err == nil)

	err = errors.New("무언가 잘못됨")
	fmt.Printf("에러: %v, nil=%t\n", err, err == nil)
	fmt.Printf("에러 타입: %T\n", err)
	fmt.Println()

	// ─────────────────────────────────────────
	// 2. errors.New와 fmt.Errorf
	// ─────────────────────────────────────────
	fmt.Println("--- 2. errors.New와 fmt.Errorf ---")

	// errors.New: 간단한 에러 메시지
	err1 := errors.New("파일을 찾을 수 없습니다")
	fmt.Printf("errors.New: %v\n", err1)

	// fmt.Errorf: 형식화된 에러 메시지
	filename := "config.json"
	err2 := fmt.Errorf("파일 %q를 열 수 없습니다: 권한 거부됨", filename)
	fmt.Printf("fmt.Errorf: %v\n", err2)

	// fmt.Errorf + %w: 에러 래핑 (Go 1.13+)
	baseErr := errors.New("연결 시간 초과")
	wrappedErr := fmt.Errorf("데이터베이스 조회 실패: %w", baseErr)
	fmt.Printf("래핑된 에러: %v\n", wrappedErr)
	fmt.Printf("errors.Is(wrappedErr, baseErr): %t\n", errors.Is(wrappedErr, baseErr))
	fmt.Println()

	// ─────────────────────────────────────────
	// 3. if err != nil 패턴
	// ─────────────────────────────────────────
	fmt.Println("--- 3. if err != nil 패턴 ---")

	// Go의 표준 에러 처리 패턴
	result, err := divide(10.0, 3.0)
	if err != nil {
		fmt.Printf("에러: %v\n", err)
	} else {
		fmt.Printf("10 / 3 = %.4f\n", result)
	}

	_, err = divide(5.0, 0.0)
	if err != nil {
		fmt.Printf("에러: %v\n", err)
	}

	// 여러 에러 처리의 연쇄
	fmt.Println("\n여러 에러 처리 연쇄:")
	if err := step1(); err != nil {
		fmt.Printf("step1 에러: %v\n", err)
		return
	}
	if err := step2(); err != nil {
		fmt.Printf("step2 에러: %v\n", err)
		return
	}
	fmt.Println("모든 단계 성공")
	fmt.Println()

	// ─────────────────────────────────────────
	// 4. 센티넬 에러 비교 (errors.Is)
	// ─────────────────────────────────────────
	fmt.Println("--- 4. 센티넬 에러와 errors.Is ---")

	// errors.Is: 에러 체인에서 특정 에러 값 찾기
	err3 := findUser(99)
	if errors.Is(err3, ErrNotFound) {
		fmt.Printf("사용자를 찾지 못함: %v\n", err3)
	}

	err4 := findUser(1)
	if err4 == nil {
		fmt.Println("사용자 찾기 성공")
	}

	// 래핑된 에러에서도 errors.Is 동작
	wrappedNotFound := fmt.Errorf("조회 중 에러: %w", ErrNotFound)
	fmt.Printf("errors.Is(wrapped, ErrNotFound): %t\n",
		errors.Is(wrappedNotFound, ErrNotFound))

	// os 패키지의 센티넬 에러
	_, err5 := os.Open("존재하지않는파일.txt")
	if err5 != nil {
		fmt.Printf("\nos.Open 에러: %v\n", err5)
		fmt.Printf("errors.Is(err, os.ErrNotExist): %t\n",
			errors.Is(err5, os.ErrNotExist))
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 5. 커스텀 에러 타입 (errors.As)
	// ─────────────────────────────────────────
	fmt.Println("--- 5. 커스텀 에러와 errors.As ---")

	// errors.As: 에러 체인에서 특정 타입의 에러 추출
	err6 := validateAge(-5)
	if err6 != nil {
		var valErr *ValidationError
		if errors.As(err6, &valErr) {
			fmt.Printf("유효성 검사 에러!\n")
			fmt.Printf("  필드: %s\n", valErr.Field)
			fmt.Printf("  메시지: %s\n", valErr.Message)
			fmt.Printf("  잘못된 값: %v\n", valErr.Value)
		}
	}

	// 래핑된 커스텀 에러에서도 errors.As 동작
	dbErr := &DatabaseError{
		Operation: "SELECT",
		Table:     "users",
		Err: &NotFoundError{
			Resource: "User",
			ID:       42,
		},
	}
	var notFoundErr *NotFoundError
	if errors.As(dbErr, &notFoundErr) {
		fmt.Printf("\nDB 에러에서 NotFoundError 추출: %v\n", notFoundErr)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 6. 에러 래핑과 언래핑
	// ─────────────────────────────────────────
	fmt.Println("--- 6. 에러 래핑 (Wrapping) ---")

	// 에러 체인: layer3 -> layer2 -> layer1
	layer1 := errors.New("원인: 네트워크 연결 실패")
	layer2 := fmt.Errorf("서비스 호출 실패: %w", layer1)
	layer3 := fmt.Errorf("요청 처리 실패: %w", layer2)

	fmt.Printf("최상위 에러: %v\n", layer3)
	fmt.Printf("errors.Unwrap(layer3): %v\n", errors.Unwrap(layer3))
	fmt.Printf("errors.Is(layer3, layer1): %t\n", errors.Is(layer3, layer1))

	// errors.Join (Go 1.20+): 여러 에러를 하나로 합치기
	err7 := errors.Join(
		errors.New("에러 1"),
		errors.New("에러 2"),
		errors.New("에러 3"),
	)
	fmt.Printf("\nerrors.Join: %v\n", err7)
	fmt.Println()

	// ─────────────────────────────────────────
	// 7. 파일 읽기에서의 에러 처리
	// ─────────────────────────────────────────
	fmt.Println("--- 7. 파일 읽기 에러 처리 ---")

	// 임시 파일 생성
	tmpFile := filepath.Join(os.TempDir(), "go_test.txt")
	if err := os.WriteFile(tmpFile, []byte("Hello, Go!\n두 번째 줄"), 0644); err != nil {
		fmt.Printf("파일 생성 실패: %v\n", err)
	} else {
		fmt.Printf("임시 파일 생성: %s\n", tmpFile)
		readFile(tmpFile)
		os.Remove(tmpFile)
	}

	// 존재하지 않는 파일 읽기
	readFile("/존재하지않는/경로/파일.txt")
	fmt.Println()

	// ─────────────────────────────────────────
	// 8. panic과 recover
	// ─────────────────────────────────────────
	fmt.Println("--- 8. panic과 recover ---")

	// panic: 프로그램을 즉시 중단시키는 내장 함수
	// 일반적으로 복구 불가능한 에러에 사용
	// recover: panic을 잡아서 프로그램을 계속 실행 (defer 내에서만 동작)

	result2, err8 := safeOperation(func() int {
		return riskyOperation(10)
	})
	fmt.Printf("safeOperation(10): result=%d, err=%v\n", result2, err8)

	result3, err9 := safeOperation(func() int {
		return riskyOperation(0) // 패닉 발생
	})
	fmt.Printf("safeOperation(0): result=%d, err=%v\n", result3, err9)
	fmt.Println()

	// ─────────────────────────────────────────
	// 9. strconv 에러 처리 예시
	// ─────────────────────────────────────────
	fmt.Println("--- 9. strconv 에러 처리 ---")

	inputs := []string{"42", "3.14", "abc", "9999999999999999999"}
	for _, input := range inputs {
		n, err := strconv.Atoi(input)
		if err != nil {
			// *strconv.NumError 타입으로 캐스팅해서 세부 정보 확인
			var numErr *strconv.NumError
			if errors.As(err, &numErr) {
				fmt.Printf("  %q 변환 실패: 함수=%s, 숫자=%s, 에러=%v\n",
					input, numErr.Func, numErr.Num, numErr.Err)
			}
		} else {
			fmt.Printf("  %q -> %d\n", input, n)
		}
	}

	fmt.Println()
	fmt.Println("=== 완료 ===")
}

// ─────────────────────────────────────────
// 헬퍼 함수들
// ─────────────────────────────────────────

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("나눗셈 에러: %w", ErrInvalidInput)
	}
	return a / b, nil
}

func step1() error {
	fmt.Println("  step1 실행 중...")
	return nil // 성공
}

func step2() error {
	fmt.Println("  step2 실행 중...")
	return nil // 성공
}

func findUser(id int) error {
	if id != 1 {
		return fmt.Errorf("findUser(%d): %w", id, ErrNotFound)
	}
	return nil
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{
			Field:   "age",
			Message: "나이는 0 이상이어야 합니다",
			Value:   age,
		}
	}
	if age > 150 {
		return &ValidationError{
			Field:   "age",
			Message: "나이는 150 이하여야 합니다",
			Value:   age,
		}
	}
	return nil
}

func readFile(path string) {
	// os.Open으로 파일 열기
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("  파일 없음: %s\n", path)
		} else if errors.Is(err, os.ErrPermission) {
			fmt.Printf("  권한 없음: %s\n", path)
		} else {
			fmt.Printf("  알 수 없는 에러: %v\n", err)
		}
		return
	}
	defer f.Close() // 함수 종료 시 파일 닫기

	// 파일 읽기
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  파일 읽기 에러: %v\n", err)
		return
	}
	fmt.Printf("  파일 내용 (%d bytes):\n", len(data))
	fmt.Printf("  %q\n", string(data))
}

func riskyOperation(n int) int {
	if n == 0 {
		panic("0으로 나눌 수 없습니다!")
	}
	return 100 / n
}

// safeOperation: panic을 recover해서 에러로 변환
func safeOperation(fn func() int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("패닉 복구: %v", r)
		}
	}()
	result = fn()
	return
}
