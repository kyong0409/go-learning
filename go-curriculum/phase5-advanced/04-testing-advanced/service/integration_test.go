//go:build integration

// service/integration_test.go
// 통합 테스트 - 실제 외부 시스템과 연동
//
// 빌드 태그 "integration"이 있어야 실행됩니다:
//   go test ./service/ -tags=integration -v
//
// 일반 "go test ./service/"로는 이 파일이 포함되지 않습니다.
// CI/CD에서는 통합 테스트를 별도 단계에서 실행하는 것이 일반적입니다.
package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// 인메모리 구현 (통합 테스트용 실제 구현)
// ============================================================

// InMemoryUserRepository는 테스트용 인메모리 사용자 저장소입니다.
type InMemoryUserRepository struct {
	users  map[int]*User
	nextID int
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("사용자를 찾을 수 없습니다: id=%d", id)
	}
	return user, nil
}

func (r *InMemoryUserRepository) Save(ctx context.Context, user *User) error {
	if user.ID == 0 {
		user.ID = r.nextID
		r.nextID++
	}
	user.CreatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) Delete(ctx context.Context, id int) error {
	if _, ok := r.users[id]; !ok {
		return fmt.Errorf("사용자를 찾을 수 없습니다: id=%d", id)
	}
	delete(r.users, id)
	return nil
}

// InMemoryOrderRepository는 테스트용 인메모리 주문 저장소입니다.
type InMemoryOrderRepository struct {
	orders map[int][]*Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: make(map[int][]*Order),
	}
}

func (r *InMemoryOrderRepository) GetByUserID(ctx context.Context, userID int) ([]*Order, error) {
	return r.orders[userID], nil
}

func (r *InMemoryOrderRepository) Create(ctx context.Context, order *Order) error {
	r.orders[order.UserID] = append(r.orders[order.UserID], order)
	return nil
}

// LogNotifier는 테스트용 알림 저장소입니다.
type LogNotifier struct {
	sent []*Notification
}

func (n *LogNotifier) Send(ctx context.Context, notif *Notification) error {
	n.sent = append(n.sent, notif)
	return nil
}

// ============================================================
// 통합 테스트 - TestMain
// ============================================================

// TestMain은 테스트 실행 전후 설정/정리를 담당합니다.
func TestMain(m *testing.M) {
	// 통합 테스트에서는 외부 서비스 연결이 필요할 수 있습니다.
	// 예: Docker로 실제 DB 시작, 환경 변수 확인 등

	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		// 환경 변수 없으면 인메모리로 폴백
		fmt.Println("[통합 테스트] TEST_DB_URL 없음 - 인메모리 저장소 사용")
	} else {
		fmt.Printf("[통합 테스트] DB 연결: %s\n", dbURL)
	}

	// 테스트 실행
	exitCode := m.Run()

	// 정리 작업
	fmt.Println("[통합 테스트] 정리 완료")

	os.Exit(exitCode)
}

// ============================================================
// 통합 테스트 케이스
// ============================================================

// TestIntegration_FullUserLifecycle은 사용자 생성→조회→삭제 전체 흐름을 테스트합니다.
func TestIntegration_FullUserLifecycle(t *testing.T) {
	// 인메모리 저장소 사용 (실제 DB가 없어도 통합 테스트 가능)
	userRepo := NewInMemoryUserRepository()
	orderRepo := NewInMemoryOrderRepository()
	notifier := &LogNotifier{}

	svc := NewUserService(userRepo, orderRepo, notifier)
	ctx := context.Background()

	// 1. 사용자 생성 (직접 저장소 사용)
	user := &User{Name: "통합테스트사용자", Email: "integration@test.com"}
	err := userRepo.Save(ctx, user)
	require.NoError(t, err)
	require.NotZero(t, user.ID)

	// 2. 서비스를 통해 사용자 조회
	found, err := svc.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Name, found.Name)

	// 3. 주문 없이 삭제 성공 확인
	err = svc.DeleteUser(ctx, user.ID)
	require.NoError(t, err)

	// 4. 삭제 후 조회 시 오류 확인
	_, err = svc.GetUser(ctx, user.ID)
	assert.Error(t, err)

	// 5. 알림이 전송됐는지 확인
	require.Len(t, notifier.sent, 1)
	assert.Equal(t, "계정 삭제 완료", notifier.sent[0].Subject)
}

// TestIntegration_OrderPreventsDelete는 주문이 있으면 삭제가 방지되는지 테스트합니다.
func TestIntegration_OrderPreventsDelete(t *testing.T) {
	userRepo := NewInMemoryUserRepository()
	orderRepo := NewInMemoryOrderRepository()
	notifier := &LogNotifier{}

	svc := NewUserService(userRepo, orderRepo, notifier)
	ctx := context.Background()

	// 사용자와 주문 생성
	user := &User{Name: "주문있는사용자"}
	require.NoError(t, userRepo.Save(ctx, user))

	order := &Order{UserID: user.ID, Items: []string{"상품A"}, Total: 10000}
	require.NoError(t, orderRepo.Create(ctx, order))

	// 삭제 시도 - 실패해야 함
	err := svc.DeleteUser(ctx, user.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "삭제할 수 없습니다")

	// 알림이 전송되지 않았는지 확인
	assert.Empty(t, notifier.sent)
}
