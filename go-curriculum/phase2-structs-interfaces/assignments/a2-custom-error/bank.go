// Package main은 은행 계좌 에러 처리 과제 구현 파일입니다.
// TODO 주석을 찾아 구현을 완성하세요.
package main

import "fmt"

// ─────────────────────────────────────────
// Sentinel 에러
// ─────────────────────────────────────────

// TODO: 필요한 sentinel 에러를 추가하세요.

// ─────────────────────────────────────────
// 커스텀 에러 타입
// ─────────────────────────────────────────

// InsufficientFundsError는 잔액이 부족할 때 발생합니다.
type InsufficientFundsError struct {
	Balance float64 // 현재 잔액
	Amount  float64 // 요청 금액
}

// TODO: InsufficientFundsError의 Error() string 메서드를 구현하세요.
// 예: "잔액 부족: 잔액 50.00, 요청 100.00"
// func (e *InsufficientFundsError) Error() string { ... }

// InvalidAmountError는 유효하지 않은 금액일 때 발생합니다.
type InvalidAmountError struct {
	Amount float64
	Reason string
}

// TODO: InvalidAmountError의 Error() string 메서드를 구현하세요.
// 예: "유효하지 않은 금액 -50.00: 금액은 양수여야 합니다"
// func (e *InvalidAmountError) Error() string { ... }

// AccountNotFoundError는 계좌를 찾지 못할 때 발생합니다.
type AccountNotFoundError struct {
	AccountID string
}

// TODO: AccountNotFoundError의 Error() string 메서드를 구현하세요.
// 예: "계좌를 찾을 수 없습니다: ACC-001"
// func (e *AccountNotFoundError) Error() string { ... }

// ─────────────────────────────────────────
// BankAccount
// ─────────────────────────────────────────

// BankAccount는 은행 계좌를 나타냅니다.
type BankAccount struct {
	ID      string
	Owner   string
	balance float64 // 소문자: 패키지 외부에서 직접 접근 불가
}

// NewBankAccount는 BankAccount 생성자입니다.
// id 또는 owner가 비어있거나 initialBalance가 음수이면 에러를 반환합니다.
// TODO: 구현하세요.
func NewBankAccount(id, owner string, initialBalance float64) (*BankAccount, error) {
	// TODO: 유효성 검사 후 BankAccount 반환
	return nil, fmt.Errorf("구현되지 않았습니다")
}

// Balance는 현재 잔액을 반환합니다.
// TODO: 구현하세요.
func (a *BankAccount) Balance() float64 {
	return 0
}

// Deposit은 금액을 입금합니다.
// amount <= 0이면 InvalidAmountError를 반환합니다.
// TODO: 구현하세요.
func (a *BankAccount) Deposit(amount float64) error {
	// TODO: 유효성 검사 후 잔액 증가
	return fmt.Errorf("구현되지 않았습니다")
}

// Withdraw는 금액을 출금합니다.
// amount <= 0이면 InvalidAmountError를 반환합니다.
// 잔액 부족이면 InsufficientFundsError를 반환합니다.
// TODO: 구현하세요.
func (a *BankAccount) Withdraw(amount float64) error {
	// TODO: 유효성 검사 및 잔액 체크 후 잔액 감소
	return fmt.Errorf("구현되지 않았습니다")
}

// ─────────────────────────────────────────
// Transfer 함수
// ─────────────────────────────────────────

// Transfer는 from 계좌에서 to 계좌로 amount를 이체합니다.
// from 또는 to가 nil이면 AccountNotFoundError를 반환합니다.
// 실패 시 에러를 fmt.Errorf + %w로 래핑하여 반환합니다.
// TODO: 구현하세요.
func Transfer(from, to *BankAccount, amount float64) error {
	// TODO: nil 체크, 출금, 입금 순서로 구현
	// 출금 실패 시 에러를 래핑하여 반환
	return fmt.Errorf("구현되지 않았습니다")
}

func main() {
	fmt.Println("은행 계좌 에러 처리 과제")
	fmt.Println("bank.go의 TODO를 구현하고 'go test -v'로 테스트하세요.")
}
