// service/service.go
// 외부 의존성이 있는 서비스 - 테스트 가능한 구조 설계
//
// 핵심 원칙: 의존성을 인터페이스로 추상화하면
// 테스트에서 목(mock)으로 교체할 수 있습니다.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ============================================================
// 도메인 타입
// ============================================================

// User는 사용자 정보를 담는 도메인 타입입니다.
type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
}

// Order는 주문 정보를 담는 도메인 타입입니다.
type Order struct {
	ID     int
	UserID int
	Items  []string
	Total  float64
}

// Notification은 알림 메시지를 표현합니다.
type Notification struct {
	UserID  int
	Subject string
	Body    string
}

// ============================================================
// 의존성 인터페이스 (모킹 가능한 경계)
// ============================================================

// UserRepository는 사용자 데이터 접근 인터페이스입니다.
// 실제 구현: PostgreSQL, MySQL 등
// 테스트 구현: 인메모리 또는 testify mock
type UserRepository interface {
	GetByID(ctx context.Context, id int) (*User, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
}

// OrderRepository는 주문 데이터 접근 인터페이스입니다.
type OrderRepository interface {
	GetByUserID(ctx context.Context, userID int) ([]*Order, error)
	Create(ctx context.Context, order *Order) error
}

// Notifier는 알림 전송 인터페이스입니다.
// 실제 구현: 이메일, SMS, Slack 등
// 테스트 구현: 목(mock)
type Notifier interface {
	Send(ctx context.Context, n *Notification) error
}

// ============================================================
// 서비스 구현
// ============================================================

// UserService는 사용자 관련 비즈니스 로직을 처리합니다.
// 생성자 주입(constructor injection)으로 의존성을 받습니다.
type UserService struct {
	users    UserRepository
	orders   OrderRepository
	notifier Notifier
}

// NewUserService는 UserService 생성자입니다.
func NewUserService(users UserRepository, orders OrderRepository, notifier Notifier) *UserService {
	return &UserService{
		users:    users,
		orders:   orders,
		notifier: notifier,
	}
}

// GetUser는 ID로 사용자를 조회합니다.
func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	if id <= 0 {
		return nil, errors.New("유효하지 않은 사용자 ID")
	}

	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("사용자 조회 실패: %w", err)
	}
	return user, nil
}

// DeleteUser는 사용자를 삭제하고 알림을 전송합니다.
// 비즈니스 규칙:
//   - 주문이 있는 사용자는 삭제할 수 없음
//   - 삭제 성공 시 사용자에게 알림 전송
func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("유효하지 않은 사용자 ID")
	}

	// 1. 사용자 존재 확인
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("사용자 조회 실패: %w", err)
	}

	// 2. 주문 이력 확인 (비즈니스 규칙)
	orders, err := s.orders.GetByUserID(ctx, id)
	if err != nil {
		return fmt.Errorf("주문 조회 실패: %w", err)
	}
	if len(orders) > 0 {
		return fmt.Errorf("사용자 %d는 %d개의 주문이 있어 삭제할 수 없습니다", id, len(orders))
	}

	// 3. 삭제 실행
	if err := s.users.Delete(ctx, id); err != nil {
		return fmt.Errorf("사용자 삭제 실패: %w", err)
	}

	// 4. 알림 전송 (실패해도 삭제는 성공 처리)
	notif := &Notification{
		UserID:  id,
		Subject: "계정 삭제 완료",
		Body:    fmt.Sprintf("%s님의 계정이 삭제되었습니다.", user.Name),
	}
	if err := s.notifier.Send(ctx, notif); err != nil {
		// 알림 실패는 로그만 남기고 계속 진행
		fmt.Printf("[경고] 알림 전송 실패 (사용자 %d): %v\n", id, err)
	}

	return nil
}

// CreateOrder는 새 주문을 생성하고 확인 알림을 전송합니다.
func (s *UserService) CreateOrder(ctx context.Context, order *Order) error {
	if order == nil {
		return errors.New("주문 정보가 없습니다")
	}
	if order.UserID <= 0 {
		return errors.New("유효하지 않은 사용자 ID")
	}
	if len(order.Items) == 0 {
		return errors.New("주문 항목이 없습니다")
	}
	if order.Total <= 0 {
		return errors.New("주문 금액이 유효하지 않습니다")
	}

	// 1. 사용자 존재 확인
	user, err := s.users.GetByID(ctx, order.UserID)
	if err != nil {
		return fmt.Errorf("사용자 조회 실패: %w", err)
	}

	// 2. 주문 저장
	if err := s.orders.Create(ctx, order); err != nil {
		return fmt.Errorf("주문 저장 실패: %w", err)
	}

	// 3. 주문 확인 알림
	notif := &Notification{
		UserID:  order.UserID,
		Subject: "주문 접수 완료",
		Body:    fmt.Sprintf("%s님, %.0f원 주문이 접수되었습니다.", user.Name, order.Total),
	}
	return s.notifier.Send(ctx, notif)
}
