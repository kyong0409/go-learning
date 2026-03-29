// bank_test.go: 은행 계좌 에러 처리 과제 채점 테스트
package main

import (
	"errors"
	"fmt"
	"testing"
)

// ─────────────────────────────────────────
// 채점 시스템
// ─────────────────────────────────────────

type scorer struct {
	total  int
	passed int
}

func newScorer() *scorer { return &scorer{} }

func (s *scorer) pass(t *testing.T, name string) {
	t.Helper()
	s.total++
	s.passed++
}

func (s *scorer) fail(t *testing.T, name, msg string) {
	t.Helper()
	s.total++
	t.Errorf("  FAIL [%s]: %s", name, msg)
}

func (s *scorer) check(t *testing.T, name string, cond bool, msg string) {
	t.Helper()
	s.total++
	if cond {
		s.passed++
	} else {
		t.Errorf("  FAIL [%s]: %s", name, msg)
	}
}

func (s *scorer) report(t *testing.T) {
	score := 0
	if s.total > 0 {
		score = s.passed * 100 / s.total
	}
	fmt.Printf("\n=== 채점 결과 ===\n")
	fmt.Printf("통과: %d/%d\n", s.passed, s.total)
	fmt.Printf("점수: %d/100\n", score)
}

// ─────────────────────────────────────────
// 에러 타입 테스트
// ─────────────────────────────────────────

func TestErrorTypes(t *testing.T) {
	sc := newScorer()

	// InsufficientFundsError
	t.Run("InsufficientFundsError", func(t *testing.T) {
		err := &InsufficientFundsError{Balance: 50, Amount: 100}
		sc.check(t, "InsufficientFundsError not nil", err != nil, "에러가 nil입니다")
		sc.check(t, "InsufficientFundsError implements error",
			err.Error() != "", "Error()가 빈 문자열입니다")
		sc.check(t, "InsufficientFundsError contains balance",
			containsFloat(err.Error(), 50), "Error()에 잔액(50)이 포함되어야 합니다")
		sc.check(t, "InsufficientFundsError contains amount",
			containsFloat(err.Error(), 100), "Error()에 요청금액(100)이 포함되어야 합니다")
	})

	// InvalidAmountError
	t.Run("InvalidAmountError", func(t *testing.T) {
		err := &InvalidAmountError{Amount: -50, Reason: "음수 불가"}
		sc.check(t, "InvalidAmountError not nil", err != nil, "에러가 nil입니다")
		sc.check(t, "InvalidAmountError implements error",
			err.Error() != "", "Error()가 빈 문자열입니다")
	})

	// AccountNotFoundError
	t.Run("AccountNotFoundError", func(t *testing.T) {
		err := &AccountNotFoundError{AccountID: "ACC-001"}
		sc.check(t, "AccountNotFoundError not nil", err != nil, "에러가 nil입니다")
		sc.check(t, "AccountNotFoundError contains ID",
			containsStr(err.Error(), "ACC-001"), "Error()에 계좌ID가 포함되어야 합니다")
	})

	sc.report(t)
}

// ─────────────────────────────────────────
// NewBankAccount 테스트
// ─────────────────────────────────────────

func TestNewBankAccount(t *testing.T) {
	sc := newScorer()

	// 정상 생성
	acc, err := NewBankAccount("ACC-001", "홍길동", 100000)
	sc.check(t, "정상 생성 에러 없음", err == nil, fmt.Sprintf("에러 발생: %v", err))
	sc.check(t, "정상 생성 nil 아님", acc != nil, "계좌가 nil입니다")
	if acc != nil {
		sc.check(t, "초기 잔액", acc.Balance() == 100000, fmt.Sprintf("잔액: got=%.0f, want=100000", acc.Balance()))
		sc.check(t, "ID 설정", acc.ID == "ACC-001", "ID가 설정되지 않았습니다")
		sc.check(t, "Owner 설정", acc.Owner == "홍길동", "Owner가 설정되지 않았습니다")
	}

	// 빈 ID
	_, err = NewBankAccount("", "홍길동", 1000)
	sc.check(t, "빈 ID 에러", err != nil, "빈 ID로 생성 시 에러가 없습니다")

	// 빈 Owner
	_, err = NewBankAccount("ACC-002", "", 1000)
	sc.check(t, "빈 Owner 에러", err != nil, "빈 Owner로 생성 시 에러가 없습니다")

	// 음수 초기 잔액
	_, err = NewBankAccount("ACC-003", "테스트", -100)
	sc.check(t, "음수 잔액 에러", err != nil, "음수 잔액으로 생성 시 에러가 없습니다")

	sc.report(t)
}

// ─────────────────────────────────────────
// Deposit 테스트
// ─────────────────────────────────────────

func TestDeposit(t *testing.T) {
	sc := newScorer()

	acc, _ := NewBankAccount("ACC-001", "테스트", 0)
	if acc == nil {
		t.Skip("NewBankAccount 구현 필요")
	}

	// 정상 입금
	err := acc.Deposit(50000)
	sc.check(t, "입금 에러 없음", err == nil, fmt.Sprintf("에러: %v", err))
	sc.check(t, "입금 후 잔액", acc.Balance() == 50000, fmt.Sprintf("잔액: %.0f", acc.Balance()))

	// 추가 입금
	acc.Deposit(30000)
	sc.check(t, "추가 입금 잔액", acc.Balance() == 80000, fmt.Sprintf("잔액: %.0f", acc.Balance()))

	// 0 입금
	err = acc.Deposit(0)
	sc.check(t, "0 입금 에러", err != nil, "0 입금 시 에러가 없습니다")
	var invErr *InvalidAmountError
	sc.check(t, "0 입금 InvalidAmountError", errors.As(err, &invErr), "InvalidAmountError 타입이어야 합니다")

	// 음수 입금
	err = acc.Deposit(-1000)
	sc.check(t, "음수 입금 에러", err != nil, "음수 입금 시 에러가 없습니다")

	sc.report(t)
}

// ─────────────────────────────────────────
// Withdraw 테스트
// ─────────────────────────────────────────

func TestWithdraw(t *testing.T) {
	sc := newScorer()

	acc, _ := NewBankAccount("ACC-001", "테스트", 100000)
	if acc == nil {
		t.Skip("NewBankAccount 구현 필요")
	}

	// 정상 출금
	err := acc.Withdraw(30000)
	sc.check(t, "출금 에러 없음", err == nil, fmt.Sprintf("에러: %v", err))
	sc.check(t, "출금 후 잔액", acc.Balance() == 70000, fmt.Sprintf("잔액: %.0f", acc.Balance()))

	// 잔액 부족
	err = acc.Withdraw(200000)
	sc.check(t, "잔액 부족 에러", err != nil, "잔액 부족 시 에러가 없습니다")
	var insErr *InsufficientFundsError
	sc.check(t, "InsufficientFundsError 타입", errors.As(err, &insErr), "InsufficientFundsError 타입이어야 합니다")
	if insErr != nil {
		sc.check(t, "InsufficientFundsError.Balance", insErr.Balance == 70000,
			fmt.Sprintf("Balance: got=%.0f, want=70000", insErr.Balance))
		sc.check(t, "InsufficientFundsError.Amount", insErr.Amount == 200000,
			fmt.Sprintf("Amount: got=%.0f, want=200000", insErr.Amount))
	} else {
		sc.total += 2 // 위 if 건너뜀
	}

	// 0 출금
	err = acc.Withdraw(0)
	sc.check(t, "0 출금 에러", err != nil, "0 출금 시 에러가 없습니다")

	// 음수 출금
	err = acc.Withdraw(-5000)
	sc.check(t, "음수 출금 에러", err != nil, "음수 출금 시 에러가 없습니다")

	sc.report(t)
}

// ─────────────────────────────────────────
// Transfer 테스트
// ─────────────────────────────────────────

func TestTransfer(t *testing.T) {
	sc := newScorer()

	acc1, _ := NewBankAccount("ACC-001", "홍길동", 100000)
	acc2, _ := NewBankAccount("ACC-002", "김영희", 50000)

	if acc1 == nil || acc2 == nil {
		t.Skip("NewBankAccount 구현 필요")
	}

	// 정상 이체
	err := Transfer(acc1, acc2, 30000)
	sc.check(t, "이체 에러 없음", err == nil, fmt.Sprintf("에러: %v", err))
	sc.check(t, "출금 계좌 잔액", acc1.Balance() == 70000, fmt.Sprintf("acc1 잔액: %.0f", acc1.Balance()))
	sc.check(t, "입금 계좌 잔액", acc2.Balance() == 80000, fmt.Sprintf("acc2 잔액: %.0f", acc2.Balance()))

	// 잔액 부족 이체
	err = Transfer(acc1, acc2, 200000)
	sc.check(t, "잔액 부족 이체 에러", err != nil, "잔액 부족 시 에러가 없습니다")
	// 에러 래핑 후에도 InsufficientFundsError를 찾을 수 있어야 함
	var insErr *InsufficientFundsError
	sc.check(t, "래핑된 InsufficientFundsError", errors.As(err, &insErr), "래핑 후에도 InsufficientFundsError를 찾을 수 있어야 합니다")

	// nil 계좌 이체
	err = Transfer(nil, acc2, 1000)
	sc.check(t, "nil from 에러", err != nil, "nil from 계좌 시 에러가 없습니다")
	var accErr *AccountNotFoundError
	sc.check(t, "nil from AccountNotFoundError", errors.As(err, &accErr), "AccountNotFoundError 타입이어야 합니다")

	err = Transfer(acc1, nil, 1000)
	sc.check(t, "nil to 에러", err != nil, "nil to 계좌 시 에러가 없습니다")

	sc.report(t)
}

// ─────────────────────────────────────────
// 헬퍼
// ─────────────────────────────────────────

func containsFloat(s string, f float64) bool {
	return containsStr(s, fmt.Sprintf("%.2f", f)) ||
		containsStr(s, fmt.Sprintf("%.0f", f)) ||
		containsStr(s, fmt.Sprintf("%g", f))
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
