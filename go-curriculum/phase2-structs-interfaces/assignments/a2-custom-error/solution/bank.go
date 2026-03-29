// Package main은 은행 계좌 에러 처리 과제의 참고 풀이입니다.
package main

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────
// 커스텀 에러 타입
// ─────────────────────────────────────────

// InsufficientFundsError는 잔액이 부족할 때 발생합니다.
type InsufficientFundsError struct {
	Balance float64
	Amount  float64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("잔액 부족: 잔액 %.2f, 요청 %.2f", e.Balance, e.Amount)
}

// InvalidAmountError는 유효하지 않은 금액일 때 발생합니다.
type InvalidAmountError struct {
	Amount float64
	Reason string
}

func (e *InvalidAmountError) Error() string {
	return fmt.Sprintf("유효하지 않은 금액 %.2f: %s", e.Amount, e.Reason)
}

// AccountNotFoundError는 계좌를 찾지 못할 때 발생합니다.
type AccountNotFoundError struct {
	AccountID string
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("계좌를 찾을 수 없습니다: %s", e.AccountID)
}

// ─────────────────────────────────────────
// BankAccount
// ─────────────────────────────────────────

// BankAccount는 은행 계좌를 나타냅니다.
type BankAccount struct {
	ID      string
	Owner   string
	balance float64
}

// NewBankAccount는 BankAccount 생성자입니다.
func NewBankAccount(id, owner string, initialBalance float64) (*BankAccount, error) {
	if strings.TrimSpace(id) == "" {
		return nil, &InvalidAmountError{Amount: 0, Reason: "계좌 ID는 비어있을 수 없습니다"}
	}
	if strings.TrimSpace(owner) == "" {
		return nil, &InvalidAmountError{Amount: 0, Reason: "계좌 주인은 비어있을 수 없습니다"}
	}
	if initialBalance < 0 {
		return nil, &InvalidAmountError{
			Amount: initialBalance,
			Reason: "초기 잔액은 0 이상이어야 합니다",
		}
	}
	return &BankAccount{
		ID:      id,
		Owner:   owner,
		balance: initialBalance,
	}, nil
}

// Balance는 현재 잔액을 반환합니다.
func (a *BankAccount) Balance() float64 {
	return a.balance
}

// Deposit은 금액을 입금합니다.
func (a *BankAccount) Deposit(amount float64) error {
	if amount <= 0 {
		return &InvalidAmountError{
			Amount: amount,
			Reason: "입금액은 양수여야 합니다",
		}
	}
	a.balance += amount
	return nil
}

// Withdraw는 금액을 출금합니다.
func (a *BankAccount) Withdraw(amount float64) error {
	if amount <= 0 {
		return &InvalidAmountError{
			Amount: amount,
			Reason: "출금액은 양수여야 합니다",
		}
	}
	if amount > a.balance {
		return &InsufficientFundsError{
			Balance: a.balance,
			Amount:  amount,
		}
	}
	a.balance -= amount
	return nil
}

// ─────────────────────────────────────────
// Transfer 함수
// ─────────────────────────────────────────

// Transfer는 from 계좌에서 to 계좌로 amount를 이체합니다.
func Transfer(from, to *BankAccount, amount float64) error {
	if from == nil {
		return fmt.Errorf("Transfer: 출금 계좌: %w", &AccountNotFoundError{AccountID: "unknown"})
	}
	if to == nil {
		return fmt.Errorf("Transfer: 입금 계좌: %w", &AccountNotFoundError{AccountID: "unknown"})
	}

	// 출금 시도 (실패 시 에러 래핑)
	if err := from.Withdraw(amount); err != nil {
		return fmt.Errorf("Transfer(%s→%s): %w", from.ID, to.ID, err)
	}

	// 입금 (출금이 성공했으므로 입금도 진행)
	if err := to.Deposit(amount); err != nil {
		// 입금 실패 시 출금 롤백
		from.balance += amount
		return fmt.Errorf("Transfer(%s→%s) 입금 실패: %w", from.ID, to.ID, err)
	}

	return nil
}

func main() {
	fmt.Println("=== 은행 계좌 에러 처리 참고 풀이 ===")

	acc1, _ := NewBankAccount("ACC-001", "홍길동", 100000)
	acc2, _ := NewBankAccount("ACC-002", "김영희", 50000)

	fmt.Printf("초기 상태: %s=%.0f, %s=%.0f\n",
		acc1.Owner, acc1.Balance(), acc2.Owner, acc2.Balance())

	if err := Transfer(acc1, acc2, 30000); err != nil {
		fmt.Printf("이체 실패: %v\n", err)
	} else {
		fmt.Printf("30000 이체 후: %s=%.0f, %s=%.0f\n",
			acc1.Owner, acc1.Balance(), acc2.Owner, acc2.Balance())
	}

	if err := Transfer(acc1, acc2, 999999); err != nil {
		fmt.Printf("잔액 부족 이체 실패 (예상): %v\n", err)
	}
}
