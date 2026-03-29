// service/service_test.go
// testify/mock을 사용한 목(mock) 기반 단위 테스트
//
// 실행:
//   go test ./service/ -v
//   go test ./service/ -v -run TestDeleteUser
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Mock 구현 (testify/mock 사용)
// ============================================================

// MockUserRepository는 UserRepository 인터페이스의 목 구현입니다.
// testify/mock.Mock을 임베딩해서 호출 기록과 기대값을 관리합니다.
type MockUserRepository struct {
	mock.Mock // testify mock 기능 임베딩
}

// GetByID는 MockUserRepository의 목 메서드입니다.
func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
	// Called()는 이 메서드가 id 인자로 호출됐음을 기록합니다.
	args := m.Called(ctx, id)
	// Get(0)은 첫 번째 반환값을 가져옵니다.
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Save(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockOrderRepository는 OrderRepository 인터페이스의 목 구현입니다.
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) GetByUserID(ctx context.Context, userID int) ([]*Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Order), args.Error(1)
}

func (m *MockOrderRepository) Create(ctx context.Context, order *Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

// MockNotifier는 Notifier 인터페이스의 목 구현입니다.
type MockNotifier struct {
	mock.Mock
}

func (m *MockNotifier) Send(ctx context.Context, n *Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

// ============================================================
// GetUser 테스트
// ============================================================

func TestGetUser_Success(t *testing.T) {
	// Arrange: 목 객체 생성
	mockUsers := new(MockUserRepository)
	mockOrders := new(MockOrderRepository)
	mockNotifier := new(MockNotifier)

	// 기대 동작 설정: GetByID(ctx, 1) 호출 시 user를 반환
	expectedUser := &User{
		ID:        1,
		Name:      "홍길동",
		Email:     "hong@example.com",
		CreatedAt: time.Now(),
	}
	mockUsers.On("GetByID", mock.Anything, 1).Return(expectedUser, nil)

	svc := NewUserService(mockUsers, mockOrders, mockNotifier)

	// Act
	user, err := svc.GetUser(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Name, user.Name)

	// 목 검증: 기대한 메서드가 실제로 호출됐는지 확인
	mockUsers.AssertExpectations(t)
}

func TestGetUser_InvalidID(t *testing.T) {
	mockUsers := new(MockUserRepository)
	mockOrders := new(MockOrderRepository)
	mockNotifier := new(MockNotifier)

	svc := NewUserService(mockUsers, mockOrders, mockNotifier)

	// ID가 0 이하인 경우
	user, err := svc.GetUser(context.Background(), 0)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "유효하지 않은")

	// GetByID가 호출되지 않았는지 확인 (조기 반환)
	mockUsers.AssertNotCalled(t, "GetByID")
}

func TestGetUser_RepositoryError(t *testing.T) {
	mockUsers := new(MockUserRepository)
	mockOrders := new(MockOrderRepository)
	mockNotifier := new(MockNotifier)

	// DB 오류 시뮬레이션
	mockUsers.On("GetByID", mock.Anything, 99).
		Return(nil, errors.New("DB 연결 실패"))

	svc := NewUserService(mockUsers, mockOrders, mockNotifier)

	user, err := svc.GetUser(context.Background(), 99)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "사용자 조회 실패")

	mockUsers.AssertExpectations(t)
}

// ============================================================
// DeleteUser 테스트
// ============================================================

func TestDeleteUser_Success(t *testing.T) {
	mockUsers := new(MockUserRepository)
	mockOrders := new(MockOrderRepository)
	mockNotifier := new(MockNotifier)

	existingUser := &User{ID: 1, Name: "홍길동"}

	// 기대 동작 체인 설정
	mockUsers.On("GetByID", mock.Anything, 1).Return(existingUser, nil)
	mockOrders.On("GetByUserID", mock.Anything, 1).Return([]*Order{}, nil) // 주문 없음
	mockUsers.On("Delete", mock.Anything, 1).Return(nil)
	mockNotifier.On("Send", mock.Anything, mock.MatchedBy(func(n *Notification) bool {
		// 알림 내용 검증: 삭제 관련 알림인지 확인
		return n.UserID == 1 && n.Subject == "계정 삭제 완료"
	})).Return(nil)

	svc := NewUserService(mockUsers, mockOrders, mockNotifier)

	err := svc.DeleteUser(context.Background(), 1)

	assert.NoError(t, err)
	mockUsers.AssertExpectations(t)
	mockOrders.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestDeleteUser_HasOrders(t *testing.T) {
	mockUsers := new(MockUserRepository)
	mockOrders := new(MockOrderRepository)
	mockNotifier := new(MockNotifier)

	existingUser := &User{ID: 2, Name: "김철수"}
	existingOrders := []*Order{
		{ID: 101, UserID: 2, Items: []string{"상품A"}},
	}

	mockUsers.On("GetByID", mock.Anything, 2).Return(existingUser, nil)
	mockOrders.On("GetByUserID", mock.Anything, 2).Return(existingOrders, nil)

	svc := NewUserService(mockUsers, mockOrders, mockNotifier)

	err := svc.DeleteUser(context.Background(), 2)

	// 주문이 있으므로 오류 반환
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "삭제할 수 없습니다")

	// Delete와 Send는 호출되지 않아야 함
	mockUsers.AssertNotCalled(t, "Delete")
	mockNotifier.AssertNotCalled(t, "Send")
}

func TestDeleteUser_NotificationFailure(t *testing.T) {
	// 알림 실패는 삭제 성공에 영향을 주지 않아야 합니다.
	mockUsers := new(MockUserRepository)
	mockOrders := new(MockOrderRepository)
	mockNotifier := new(MockNotifier)

	existingUser := &User{ID: 3, Name: "이영희"}

	mockUsers.On("GetByID", mock.Anything, 3).Return(existingUser, nil)
	mockOrders.On("GetByUserID", mock.Anything, 3).Return([]*Order{}, nil)
	mockUsers.On("Delete", mock.Anything, 3).Return(nil)
	// 알림 전송 실패
	mockNotifier.On("Send", mock.Anything, mock.Anything).
		Return(errors.New("이메일 서버 연결 실패"))

	svc := NewUserService(mockUsers, mockOrders, mockNotifier)

	// 알림이 실패해도 DeleteUser는 성공해야 합니다.
	err := svc.DeleteUser(context.Background(), 3)
	assert.NoError(t, err) // 알림 실패는 에러가 아님

	mockUsers.AssertExpectations(t)
}

// ============================================================
// CreateOrder 테스트
// ============================================================

func TestCreateOrder_Success(t *testing.T) {
	mockUsers := new(MockUserRepository)
	mockOrders := new(MockOrderRepository)
	mockNotifier := new(MockNotifier)

	user := &User{ID: 1, Name: "홍길동"}
	order := &Order{
		UserID: 1,
		Items:  []string{"상품A", "상품B"},
		Total:  29_000,
	}

	mockUsers.On("GetByID", mock.Anything, 1).Return(user, nil)
	mockOrders.On("Create", mock.Anything, order).Return(nil)
	mockNotifier.On("Send", mock.Anything, mock.MatchedBy(func(n *Notification) bool {
		return n.Subject == "주문 접수 완료"
	})).Return(nil)

	svc := NewUserService(mockUsers, mockOrders, mockNotifier)

	err := svc.CreateOrder(context.Background(), order)
	assert.NoError(t, err)

	mockUsers.AssertExpectations(t)
	mockOrders.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}

func TestCreateOrder_ValidationErrors(t *testing.T) {
	// 표 기반 테스트: 다양한 유효성 검사 오류 케이스
	testCases := []struct {
		name      string
		order     *Order
		expectErr string
	}{
		{
			name:      "nil 주문",
			order:     nil,
			expectErr: "주문 정보가 없습니다",
		},
		{
			name:      "유효하지 않은 사용자 ID",
			order:     &Order{UserID: 0, Items: []string{"상품A"}, Total: 1000},
			expectErr: "유효하지 않은 사용자 ID",
		},
		{
			name:      "빈 주문 항목",
			order:     &Order{UserID: 1, Items: []string{}, Total: 1000},
			expectErr: "주문 항목이 없습니다",
		},
		{
			name:      "음수 금액",
			order:     &Order{UserID: 1, Items: []string{"상품A"}, Total: -100},
			expectErr: "주문 금액이 유효하지 않습니다",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers := new(MockUserRepository)
			mockOrders := new(MockOrderRepository)
			mockNotifier := new(MockNotifier)

			svc := NewUserService(mockUsers, mockOrders, mockNotifier)

			err := svc.CreateOrder(context.Background(), tc.order)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectErr)

			// 유효성 검사 실패 시 저장소는 호출되지 않아야 함
			mockUsers.AssertNotCalled(t, "GetByID")
			mockOrders.AssertNotCalled(t, "Create")
		})
	}
}
