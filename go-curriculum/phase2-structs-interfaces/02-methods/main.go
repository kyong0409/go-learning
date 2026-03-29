// 패키지 선언
package main

import (
	"fmt"
	"math"
	"strings"
)

// ─────────────────────────────────────────
// 1. 기본 구조체 정의
// ─────────────────────────────────────────

// Rectangle은 직사각형을 나타냅니다.
type Rectangle struct {
	Width  float64
	Height float64
}

// Circle은 원을 나타냅니다.
type Circle struct {
	Radius float64
}

// ─────────────────────────────────────────
// 2. 값 리시버(Value Receiver) 메서드
// ─────────────────────────────────────────
// 값 리시버: 구조체의 복사본을 받습니다.
// - 구조체를 수정하지 않는 경우에 사용
// - 구조체가 작고 복사 비용이 낮을 때 적합

// Area는 직사각형의 넓이를 계산합니다 (값 리시버).
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter는 직사각형의 둘레를 계산합니다 (값 리시버).
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// IsSquare는 정사각형인지 확인합니다 (값 리시버).
func (r Rectangle) IsSquare() bool {
	return r.Width == r.Height
}

// Area는 원의 넓이를 계산합니다 (값 리시버).
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Circumference는 원의 둘레를 계산합니다.
func (c Circle) Circumference() float64 {
	return 2 * math.Pi * c.Radius
}

// ─────────────────────────────────────────
// 3. 포인터 리시버(Pointer Receiver) 메서드
// ─────────────────────────────────────────
// 포인터 리시버: 구조체의 포인터를 받습니다.
// - 구조체를 수정해야 하는 경우에 사용
// - 구조체가 크거나 복사 비용이 높을 때 적합
// - 하나의 포인터 리시버 메서드가 있으면 나머지도 포인터 리시버로 통일 권장

// Scale은 직사각형의 크기를 배율로 조정합니다 (포인터 리시버).
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor  // 원본 수정
	r.Height *= factor // 원본 수정
}

// SetWidth는 너비를 설정합니다 (포인터 리시버).
func (r *Rectangle) SetWidth(w float64) {
	r.Width = w
}

// SetHeight는 높이를 설정합니다 (포인터 리시버).
func (r *Rectangle) SetHeight(h float64) {
	r.Height = h
}

// ─────────────────────────────────────────
// 4. String() 메서드 - fmt.Stringer 인터페이스
// ─────────────────────────────────────────
// fmt 패키지는 String() 메서드가 있는 타입을 자동으로 사용합니다.
// fmt.Println, fmt.Printf %v, %s 등에서 자동 호출됩니다.

// String은 Rectangle의 문자열 표현을 반환합니다.
// fmt.Stringer 인터페이스를 구현합니다: String() string
func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%.2f x %.2f)", r.Width, r.Height)
}

// String은 Circle의 문자열 표현을 반환합니다.
func (c Circle) String() string {
	return fmt.Sprintf("Circle(반지름=%.2f)", c.Radius)
}

// ─────────────────────────────────────────
// 5. NewXxx 생성자 패턴
// ─────────────────────────────────────────
// Go는 클래스 생성자가 없지만, NewXxx 함수로 관례적 생성자를 만듭니다.
// - 유효성 검사를 포함할 수 있습니다.
// - 포인터를 반환하는 것이 일반적입니다.
// - "인터페이스를 받아들이고 구조체를 반환하라" 원칙 적용

// BankAccount는 은행 계좌를 나타냅니다.
type BankAccount struct {
	owner   string  // 소문자: 패키지 외부에서 직접 접근 불가 (캡슐화)
	balance float64 // 잔액 직접 수정 방지
}

// NewBankAccount는 BankAccount의 생성자입니다.
// 유효성 검사를 포함하고 에러를 반환합니다.
func NewBankAccount(owner string, initialBalance float64) (*BankAccount, error) {
	// 유효성 검사
	if strings.TrimSpace(owner) == "" {
		return nil, fmt.Errorf("계좌 주인 이름은 비어있을 수 없습니다")
	}
	if initialBalance < 0 {
		return nil, fmt.Errorf("초기 잔액은 0 이상이어야 합니다: %.2f", initialBalance)
	}

	return &BankAccount{
		owner:   owner,
		balance: initialBalance,
	}, nil
}

// Owner는 계좌 주인을 반환합니다 (Getter).
func (a *BankAccount) Owner() string {
	return a.owner
}

// Balance는 현재 잔액을 반환합니다 (Getter).
func (a *BankAccount) Balance() float64 {
	return a.balance
}

// Deposit은 금액을 입금합니다 (포인터 리시버 - 상태 변경).
func (a *BankAccount) Deposit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("입금액은 0보다 커야 합니다: %.2f", amount)
	}
	a.balance += amount
	return nil
}

// Withdraw는 금액을 출금합니다 (포인터 리시버 - 상태 변경).
func (a *BankAccount) Withdraw(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("출금액은 0보다 커야 합니다: %.2f", amount)
	}
	if amount > a.balance {
		return fmt.Errorf("잔액 부족: 잔액 %.2f, 요청 %.2f", a.balance, amount)
	}
	a.balance -= amount
	return nil
}

// String은 BankAccount의 문자열 표현입니다 (Stringer 구현).
func (a *BankAccount) String() string {
	return fmt.Sprintf("BankAccount{주인: %s, 잔액: ₩%.2f}", a.owner, a.balance)
}

// ─────────────────────────────────────────
// 6. 메서드 세트(Method Set) 개념
// ─────────────────────────────────────────
// - 값 타입 T의 메서드 세트: 값 리시버(T) 메서드만 포함
// - 포인터 타입 *T의 메서드 세트: 값 리시버(T) + 포인터 리시버(*T) 모두 포함
//
// 실용적 규칙:
// - 메서드가 구조체를 수정 → 포인터 리시버
// - 구조체가 크다 → 포인터 리시버 (복사 비용 절감)
// - 위 두 경우 외에는 값 리시버도 괜찮음
// - 한 타입의 메서드는 리시버 종류를 통일하는 것이 권장됨

// Counter는 메서드 세트 개념을 보여줍니다.
type Counter struct {
	count int
}

// NewCounter는 Counter 생성자입니다.
func NewCounter(initial int) *Counter {
	return &Counter{count: initial}
}

// Value는 현재 값을 반환합니다 (값 리시버).
func (c Counter) Value() int {
	return c.count
}

// Increment는 카운터를 1 증가시킵니다 (포인터 리시버).
func (c *Counter) Increment() {
	c.count++
}

// Add는 카운터에 n을 더합니다 (포인터 리시버).
func (c *Counter) Add(n int) {
	c.count += n
}

// Reset은 카운터를 0으로 초기화합니다 (포인터 리시버).
func (c *Counter) Reset() {
	c.count = 0
}

// String은 Counter의 문자열 표현입니다.
func (c Counter) String() string {
	return fmt.Sprintf("Counter(%d)", c.count)
}

func main() {
	fmt.Println("=== Go Phase 2: 메서드(Methods) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 값 리시버 메서드 사용
	// ─────────────────────────────────────────
	fmt.Println("--- 1. 값 리시버 메서드 ---")

	rect := Rectangle{Width: 5.0, Height: 3.0}

	// String() 메서드 덕분에 fmt.Println이 보기 좋게 출력
	fmt.Printf("도형: %v\n", rect)
	fmt.Printf("넓이: %.2f\n", rect.Area())
	fmt.Printf("둘레: %.2f\n", rect.Perimeter())
	fmt.Printf("정사각형인가? %v\n", rect.IsSquare())

	square := Rectangle{Width: 4.0, Height: 4.0}
	fmt.Printf("\n%v\n", square)
	fmt.Printf("정사각형인가? %v\n", square.IsSquare())

	circle := Circle{Radius: 7.0}
	fmt.Printf("\n%v\n", circle)
	fmt.Printf("넓이: %.4f\n", circle.Area())
	fmt.Printf("둘레: %.4f\n", circle.Circumference())
	fmt.Println()

	// ─────────────────────────────────────────
	// 포인터 리시버 메서드 사용
	// ─────────────────────────────────────────
	fmt.Println("--- 2. 포인터 리시버 메서드 ---")

	r := Rectangle{Width: 2.0, Height: 3.0}
	fmt.Printf("Scale 전: %v, 넓이=%.2f\n", r, r.Area())

	// Go는 값 변수에서도 포인터 리시버 메서드를 자동으로 처리합니다.
	// r.Scale(3) 은 (&r).Scale(3) 으로 자동 변환됩니다.
	r.Scale(3)
	fmt.Printf("Scale(3) 후: %v, 넓이=%.2f\n", r, r.Area())

	// 포인터 변수에서도 값 리시버 메서드 호출 가능
	rPtr := &Rectangle{Width: 10.0, Height: 5.0}
	fmt.Printf("포인터 변수의 값 리시버 메서드: 넓이=%.2f\n", rPtr.Area())
	rPtr.Scale(0.5)
	fmt.Printf("Scale(0.5) 후: %v\n", *rPtr)
	fmt.Println()

	// ─────────────────────────────────────────
	// 생성자 패턴 (NewXxx)
	// ─────────────────────────────────────────
	fmt.Println("--- 3. NewXxx 생성자 패턴 ---")

	// 정상 생성
	account, err := NewBankAccount("홍길동", 100000)
	if err != nil {
		fmt.Printf("계좌 생성 실패: %v\n", err)
		return
	}
	fmt.Printf("계좌 생성: %v\n", account)

	// 입금
	if err := account.Deposit(50000); err != nil {
		fmt.Printf("입금 실패: %v\n", err)
	} else {
		fmt.Printf("50000 입금 후: %v\n", account)
	}

	// 출금
	if err := account.Withdraw(30000); err != nil {
		fmt.Printf("출금 실패: %v\n", err)
	} else {
		fmt.Printf("30000 출금 후: %v\n", account)
	}

	// 잔액 부족 출금 시도
	if err := account.Withdraw(200000); err != nil {
		fmt.Printf("출금 실패 (정상): %v\n", err)
	}

	// 잘못된 생성 시도
	_, err = NewBankAccount("", 1000)
	fmt.Printf("빈 이름으로 생성 시도: %v\n", err)

	_, err = NewBankAccount("테스트", -500)
	fmt.Printf("음수 잔액으로 생성 시도: %v\n", err)
	fmt.Println()

	// ─────────────────────────────────────────
	// 메서드 체이닝 스타일 (포인터 리시버)
	// ─────────────────────────────────────────
	fmt.Println("--- 4. Counter와 메서드 세트 ---")

	counter := NewCounter(0)
	fmt.Printf("초기 카운터: %v\n", counter)

	counter.Increment()
	counter.Increment()
	counter.Add(5)
	fmt.Printf("Increment×2 + Add(5) 후: %v\n", counter)

	// 값 복사 후 수정 - 원본에 영향 없음
	counterCopy := *counter // 값 복사
	counterCopy.Reset()     // 복사본만 초기화
	fmt.Printf("복사본 Reset 후 - 원본: %v, 복사본: %v\n", counter, counterCopy)
	fmt.Println()

	// ─────────────────────────────────────────
	// String() 메서드 자동 호출 확인
	// ─────────────────────────────────────────
	fmt.Println("--- 5. Stringer 인터페이스 (String() 메서드) ---")

	shapes := []Rectangle{
		{Width: 3, Height: 4},
		{Width: 5, Height: 5},
		{Width: 10, Height: 2},
	}

	// fmt.Println이 String() 메서드를 자동으로 호출합니다.
	for i, s := range shapes {
		fmt.Printf("  도형[%d]: %v (넓이=%.0f)\n", i, s, s.Area())
	}

	// fmt.Sprintf에서도 동일하게 작동
	desc := fmt.Sprintf("가장 큰 도형: %v", shapes[2])
	fmt.Println(desc)

	fmt.Println()
	fmt.Println("=== 메서드 학습 완료 ===")
}
