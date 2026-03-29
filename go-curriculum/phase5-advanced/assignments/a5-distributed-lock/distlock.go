// a5-distributed-lock/distlock.go
// 분산 잠금 시스템 (로컬 시뮬레이션) 과제입니다.
// TODO 주석이 있는 모든 함수/메서드를 구현하세요.
package distlock

import (
	"context"
	"errors"
	"time"
)

// ============================================================
// 오류 타입 (수정하지 마세요)
// ============================================================

var (
	// ErrLockNotFound는 잠금이 존재하지 않을 때 반환됩니다.
	ErrLockNotFound = errors.New("잠금을 찾을 수 없습니다")
	// ErrInvalidToken은 펜싱 토큰이 유효하지 않을 때 반환됩니다.
	ErrInvalidToken = errors.New("유효하지 않은 펜싱 토큰입니다")
	// ErrLockExpired는 잠금이 만료되었을 때 반환됩니다.
	ErrLockExpired = errors.New("잠금이 만료되었습니다")
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// Lock은 획득한 잠금을 나타냅니다.
type Lock struct {
	Key        string    // 잠금 대상 리소스 키
	Owner      string    // 잠금 보유자 식별자
	Token      uint64    // 펜싱 토큰 (단조 증가)
	AcquiredAt time.Time // 획득 시각
	ExpiresAt  time.Time // 만료 시각
}

// IsValid는 잠금이 현재 시각 기준으로 유효한지(만료되지 않았는지) 반환합니다.
// TODO: 구현하세요.
func (l *Lock) IsValid() bool {
	return false
}

// ============================================================
// LockService
// ============================================================

// LockService는 분산 잠금을 관리합니다.
// 여러 고루틴에서 동시에 호출되어도 안전해야 합니다.
type LockService struct {
	// TODO: 필드를 추가하세요.
	// 힌트:
	//   - 키별 현재 잠금 정보를 저장하는 맵
	//   - 펜싱 토큰 카운터 (atomic)
	//   - 키별 대기자에게 알리는 채널 또는 sync.Cond
	//   - 소유자가 대기 중인 키를 추적하는 맵 (데드락 감지용)
}

// NewLockService는 새 LockService를 생성합니다.
// 백그라운드 만료 정리 고루틴을 시작합니다.
// TODO: 구현하세요.
func NewLockService() *LockService {
	panic("NewLockService: 아직 구현되지 않았습니다")
}

// Acquire는 key에 대한 잠금을 획득합니다.
//
// 요구사항:
//   - 잠금이 없거나 만료된 경우: 즉시 새 잠금 발급
//   - 잠금이 이미 점유된 경우: 해제 또는 만료될 때까지 블로킹 대기
//   - ctx가 취소되면 대기를 중단하고 context.Canceled 오류 반환
//   - 획득 성공 시 단조 증가하는 펜싱 토큰이 포함된 Lock 반환
//
// TODO: 구현하세요.
func (s *LockService) Acquire(ctx context.Context, key, owner string, ttl time.Duration) (*Lock, error) {
	panic("Acquire: 아직 구현되지 않았습니다")
}

// Release는 잠금을 해제합니다.
//
// 요구사항:
//   - lock.Token이 현재 보유 토큰과 다르면 ErrInvalidToken 반환
//   - 키에 해당하는 잠금이 없으면 ErrLockNotFound 반환
//   - 성공 시 대기 중인 Acquire 호출자 중 하나를 깨웁니다
//
// TODO: 구현하세요.
func (s *LockService) Release(lock *Lock) error {
	panic("Release: 아직 구현되지 않았습니다")
}

// Refresh는 잠금의 TTL을 갱신합니다.
//
// 요구사항:
//   - lock.Token이 현재 보유 토큰과 다르면 ErrInvalidToken 반환
//   - 잠금이 만료된 경우 ErrLockExpired 반환
//   - 성공 시 ExpiresAt이 갱신된 새 Lock 포인터를 반환합니다
//
// TODO: 구현하세요.
func (s *LockService) Refresh(lock *Lock, newTTL time.Duration) (*Lock, error) {
	panic("Refresh: 아직 구현되지 않았습니다")
}

// IsLocked는 키에 유효한 잠금이 존재하는지 반환합니다.
// 만료된 잠금은 false를 반환합니다.
// TODO: 구현하세요.
func (s *LockService) IsLocked(key string) bool {
	return false
}

// GetLock은 키의 현재 잠금 정보를 반환합니다.
// 잠금이 없거나 만료된 경우 nil, false를 반환합니다.
// TODO: 구현하세요.
func (s *LockService) GetLock(key string) (*Lock, bool) {
	return nil, false
}

// DetectDeadlock은 소유자 간 대기 사이클을 탐지합니다.
//
// 예시:
//   - ownerA가 key1을 보유하고 key2를 대기 중
//   - ownerB가 key2를 보유하고 key1을 대기 중
//   => 사이클: ["ownerA", "ownerB"] (또는 동등한 순열)
//
// 사이클이 없으면 nil, nil을 반환합니다.
// TODO: 구현하세요.
func (s *LockService) DetectDeadlock() ([]string, error) {
	return nil, nil
}

// Close는 백그라운드 고루틴을 중지합니다.
// TODO: 구현하세요.
func (s *LockService) Close() {
}
