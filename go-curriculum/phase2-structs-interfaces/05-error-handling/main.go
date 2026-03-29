// 패키지 선언
package main

import (
	"errors"
	"fmt"
	"strconv"
)

// ─────────────────────────────────────────
// 1. Sentinel 에러 (sentinel errors)
// ─────────────────────────────────────────
// 센티넬 에러는 패키지 수준의 변수로 선언된 특정 에러입니다.
// errors.Is()로 비교할 수 있습니다.
// 관례: var ErrXxx = errors.New("...") 형태로 선언

var (
	ErrNotFound       = errors.New("항목을 찾을 수 없습니다")
	ErrUnauthorized   = errors.New("권한이 없습니다")
	ErrInvalidInput   = errors.New("유효하지 않은 입력입니다")
	ErrAlreadyExists  = errors.New("항목이 이미 존재합니다")
)

// ─────────────────────────────────────────
// 2. 커스텀 에러 타입
// ─────────────────────────────────────────
// error 인터페이스: Error() string 메서드만 구현하면 됩니다.
//
// type error interface {
//     Error() string
// }

// ValidationError는 유효성 검사 에러를 나타냅니다.
type ValidationError struct {
	Field   string // 에러가 발생한 필드명
	Message string // 에러 메시지
	Value   any    // 잘못된 값
}

// Error는 error 인터페이스를 구현합니다.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("유효성 검사 실패 - 필드 '%s': %s (값: %v)", e.Field, e.Message, e.Value)
}

// DatabaseError는 데이터베이스 에러를 나타냅니다.
type DatabaseError struct {
	Operation string // 실행된 작업 (SELECT, INSERT 등)
	Table     string // 에러가 발생한 테이블
	Err       error  // 원인 에러 (래핑용)
}

// Error는 error 인터페이스를 구현합니다.
func (e *DatabaseError) Error() string {
	return fmt.Sprintf("DB 에러 [%s on %s]: %v", e.Operation, e.Table, e.Err)
}

// Unwrap은 래핑된 에러를 반환합니다.
// errors.Is(), errors.As()가 에러 체인을 탐색할 수 있게 합니다.
func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NotFoundError는 항목을 찾지 못한 에러입니다.
type NotFoundError struct {
	Resource string // 리소스 종류 (User, Product 등)
	ID       int    // 찾으려 했던 ID
}

// Error는 error 인터페이스를 구현합니다.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s ID=%d를 찾을 수 없습니다", e.Resource, e.ID)
}

// Is는 errors.Is()와 통합됩니다.
// ErrNotFound 와 비교할 때 true를 반환하게 합니다.
func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// ─────────────────────────────────────────
// 3. 에러를 반환하는 함수들
// ─────────────────────────────────────────

// User는 사용자를 나타냅니다.
type User struct {
	ID    int
	Name  string
	Email string
	Age   int
}

// 간단한 인메모리 사용자 저장소
var userStore = map[int]User{
	1: {ID: 1, Name: "홍길동", Email: "hong@example.com", Age: 30},
	2: {ID: 2, Name: "김영희", Email: "kim@example.com", Age: 25},
}

// findUser는 사용자를 조회합니다.
// 없으면 NotFoundError를 반환합니다.
func findUser(id int) (*User, error) {
	user, ok := userStore[id]
	if !ok {
		// 커스텀 에러 반환
		return nil, &NotFoundError{Resource: "User", ID: id}
	}
	return &user, nil
}

// validateUser는 사용자 데이터를 유효성 검사합니다.
func validateUser(u User) error {
	if u.Name == "" {
		return &ValidationError{
			Field:   "Name",
			Message: "이름은 필수입니다",
			Value:   u.Name,
		}
	}
	if u.Age < 0 || u.Age > 150 {
		return &ValidationError{
			Field:   "Age",
			Message: "나이는 0~150 사이여야 합니다",
			Value:   u.Age,
		}
	}
	if u.Email == "" {
		return &ValidationError{
			Field:   "Email",
			Message: "이메일은 필수입니다",
			Value:   u.Email,
		}
	}
	return nil
}

// getUser는 DB에서 사용자를 가져오는 서비스 함수입니다.
// 에러를 래핑하여 컨텍스트를 추가합니다.
func getUser(id int) (*User, error) {
	user, err := findUser(id)
	if err != nil {
		// fmt.Errorf + %w: 에러를 래핑하면서 메시지를 추가합니다.
		// %w를 사용하면 errors.Is()와 errors.As()로 내부 에러를 찾을 수 있습니다.
		return nil, fmt.Errorf("getUser(%d): %w", id, err)
	}
	return user, nil
}

// dbQueryUser는 DatabaseError로 한 번 더 래핑합니다.
func dbQueryUser(id int) (*User, error) {
	user, err := getUser(id)
	if err != nil {
		return nil, &DatabaseError{
			Operation: "SELECT",
			Table:     "users",
			Err:       err, // 이미 래핑된 에러를 다시 래핑
		}
	}
	return user, nil
}

// parseAge는 문자열을 나이로 파싱합니다.
// 여러 에러 케이스를 보여줍니다.
func parseAge(s string) (int, error) {
	age, err := strconv.Atoi(s)
	if err != nil {
		// 외부 에러를 %w로 래핑
		return 0, fmt.Errorf("나이 파싱 실패 %q: %w", s, ErrInvalidInput)
	}
	if age < 0 || age > 150 {
		return 0, fmt.Errorf("나이 범위 초과 %d: %w", age, ErrInvalidInput)
	}
	return age, nil
}

func main() {
	fmt.Println("=== Go Phase 2: 에러 처리(Error Handling) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// Sentinel 에러
	// ─────────────────────────────────────────
	fmt.Println("--- 1. Sentinel 에러 ---")

	// errors.Is()로 센티넬 에러 비교
	err1 := ErrNotFound
	fmt.Printf("err1 == ErrNotFound: %v\n", err1 == ErrNotFound)
	fmt.Printf("errors.Is(err1, ErrNotFound): %v\n", errors.Is(err1, ErrNotFound))

	// 래핑된 에러에서도 errors.Is() 동작
	wrapped := fmt.Errorf("조회 실패: %w", ErrNotFound)
	fmt.Printf("래핑 후 == : %v\n", wrapped == ErrNotFound)           // false
	fmt.Printf("래핑 후 errors.Is: %v\n", errors.Is(wrapped, ErrNotFound)) // true!
	fmt.Println()

	// ─────────────────────────────────────────
	// 커스텀 에러 타입
	// ─────────────────────────────────────────
	fmt.Println("--- 2. 커스텀 에러 타입 ---")

	// ValidationError
	validErr := &ValidationError{
		Field:   "Age",
		Message: "음수는 허용되지 않습니다",
		Value:   -5,
	}
	fmt.Printf("ValidationError: %v\n", validErr)
	fmt.Printf("에러 타입: %T\n", validErr)

	// 사용자 유효성 검사
	badUser := User{Name: "테스트", Age: 200, Email: "test@example.com"}
	if err := validateUser(badUser); err != nil {
		fmt.Printf("유효성 검사 실패: %v\n", err)
	}

	goodUser := User{Name: "홍길동", Age: 30, Email: "hong@example.com"}
	if err := validateUser(goodUser); err != nil {
		fmt.Printf("유효성 검사 실패: %v\n", err)
	} else {
		fmt.Println("유효성 검사 통과!")
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// fmt.Errorf + %w 에러 래핑
	// ─────────────────────────────────────────
	fmt.Println("--- 3. fmt.Errorf + %w 에러 래핑 ---")

	// 정상 조회
	user, err := getUser(1)
	if err != nil {
		fmt.Printf("에러: %v\n", err)
	} else {
		fmt.Printf("사용자 조회 성공: %+v\n", *user)
	}

	// 없는 사용자 조회 → 에러 체인 발생
	_, err = getUser(999)
	if err != nil {
		fmt.Printf("\n에러 메시지: %v\n", err)
		// 에러 체인: getUser(999) → findUser(999) → NotFoundError
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// errors.Is() - 에러 체인 탐색
	// ─────────────────────────────────────────
	fmt.Println("--- 4. errors.Is() ---")

	// 깊이 래핑된 에러에서 센티넬 에러 찾기
	_, dbErr := dbQueryUser(999)
	if dbErr != nil {
		fmt.Printf("최상위 에러: %v\n\n", dbErr)

		// errors.Is는 에러 체인을 재귀적으로 탐색합니다.
		// DatabaseError → fmt.Errorf → NotFoundError → Is(ErrNotFound) == true
		fmt.Printf("errors.Is(dbErr, ErrNotFound): %v\n", errors.Is(dbErr, ErrNotFound))
		fmt.Printf("errors.Is(dbErr, ErrUnauthorized): %v\n", errors.Is(dbErr, ErrUnauthorized))
	}

	// parseAge 에러
	_, parseErr := parseAge("abc")
	if parseErr != nil {
		fmt.Printf("\nparseAge 에러: %v\n", parseErr)
		fmt.Printf("errors.Is(parseErr, ErrInvalidInput): %v\n", errors.Is(parseErr, ErrInvalidInput))
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// errors.As() - 특정 타입으로 언래핑
	// ─────────────────────────────────────────
	fmt.Println("--- 5. errors.As() ---")

	// 깊이 래핑된 에러에서 특정 타입 꺼내기
	_, dbErr2 := dbQueryUser(999)
	if dbErr2 != nil {
		// DatabaseError 타입 꺼내기
		var dbError *DatabaseError
		if errors.As(dbErr2, &dbError) {
			fmt.Printf("DatabaseError 발견!\n")
			fmt.Printf("  작업: %s\n", dbError.Operation)
			fmt.Printf("  테이블: %s\n", dbError.Table)
		}

		// NotFoundError 타입 꺼내기 (더 깊은 체인)
		var notFoundErr *NotFoundError
		if errors.As(dbErr2, &notFoundErr) {
			fmt.Printf("NotFoundError 발견!\n")
			fmt.Printf("  리소스: %s\n", notFoundErr.Resource)
			fmt.Printf("  ID: %d\n", notFoundErr.ID)
		}

		// ValidationError는 없음
		var validationErr *ValidationError
		if errors.As(dbErr2, &validationErr) {
			fmt.Printf("ValidationError 발견\n")
		} else {
			fmt.Printf("ValidationError 없음 (예상된 결과)\n")
		}
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 실전 에러 처리 패턴
	// ─────────────────────────────────────────
	fmt.Println("--- 6. 실전 에러 처리 패턴 ---")

	// 여러 사용자 ID 조회
	ids := []int{1, 2, 999, 404}
	for _, id := range ids {
		user, err := dbQueryUser(id)
		if err != nil {
			// 에러 종류에 따른 처리
			switch {
			case errors.Is(err, ErrNotFound):
				fmt.Printf("  ID=%d: 사용자 없음 (404 처리)\n", id)
			case errors.Is(err, ErrUnauthorized):
				fmt.Printf("  ID=%d: 권한 없음 (401 처리)\n", id)
			default:
				fmt.Printf("  ID=%d: 알 수 없는 에러: %v\n", id, err)
			}
		} else {
			fmt.Printf("  ID=%d: 조회 성공 → %s\n", id, user.Name)
		}
	}

	// 나이 파싱 테스트
	fmt.Println("\n나이 파싱 테스트:")
	ageInputs := []string{"25", "-1", "200", "abc", "0"}
	for _, input := range ageInputs {
		age, err := parseAge(input)
		if err != nil {
			if errors.Is(err, ErrInvalidInput) {
				fmt.Printf("  %q → 잘못된 입력: %v\n", input, err)
			} else {
				fmt.Printf("  %q → 에러: %v\n", input, err)
			}
		} else {
			fmt.Printf("  %q → 나이: %d\n", input, age)
		}
	}

	fmt.Println()
	fmt.Println("=== 에러 처리 학습 완료 ===")
}
