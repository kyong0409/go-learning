# 과제 2: 은행 계좌 시스템의 에러 처리

## 목표

커스텀 에러 타입, 에러 래핑, `errors.Is()` / `errors.As()` 사용법을 익힙니다.

## 요구사항

### 1. 커스텀 에러 타입 구현

다음 세 가지 에러 타입을 구현하세요:

```go
// InsufficientFundsError: 잔액 부족
type InsufficientFundsError struct {
    Balance float64 // 현재 잔액
    Amount  float64 // 요청 금액
}

// InvalidAmountError: 유효하지 않은 금액
type InvalidAmountError struct {
    Amount  float64
    Reason  string
}

// AccountNotFoundError: 계좌 없음
type AccountNotFoundError struct {
    AccountID string
}
```

각 타입은 `error` 인터페이스의 `Error() string` 메서드를 구현해야 합니다.

### 2. BankAccount 구현

```go
type BankAccount struct {
    ID      string
    Owner   string
    balance float64  // 소문자: 직접 접근 불가
}

func NewBankAccount(id, owner string, initialBalance float64) (*BankAccount, error)
func (a *BankAccount) Balance() float64
func (a *BankAccount) Deposit(amount float64) error
func (a *BankAccount) Withdraw(amount float64) error
```

### 3. Transfer 함수 구현

```go
// Transfer는 from 계좌에서 to 계좌로 amount를 이체합니다.
// 이체 실패 시 에러를 래핑하여 반환합니다.
func Transfer(from, to *BankAccount, amount float64) error
```

### 4. 에러 처리 규칙

- `amount <= 0`: `InvalidAmountError` 반환
- 잔액 < 출금액: `InsufficientFundsError` 반환
- `nil` 계좌 전달: `AccountNotFoundError` 반환
- 에러 래핑에 `fmt.Errorf("%w", err)` 사용

## 실행 방법

```bash
go test -v
```

## 채점

```
=== 채점 결과 ===
통과: 18/20
점수: 90/100
```

## 힌트

- `errors.Is(err, target)`: 에러 체인에서 target과 같은 에러 탐색
- `errors.As(err, &target)`: 에러 체인에서 특정 타입 꺼내기
- `fmt.Errorf("...: %w", err)`: 에러 래핑
- 참고 풀이: `solution/bank.go`
