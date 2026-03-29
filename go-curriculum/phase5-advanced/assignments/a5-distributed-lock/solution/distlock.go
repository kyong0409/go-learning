// a5-distributed-lock/solution/distlock.go
// 분산 잠금 시스템 참고 답안입니다.
package distlock

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// 오류 타입
// ============================================================

var (
	ErrLockNotFound = errors.New("잠금을 찾을 수 없습니다")
	ErrInvalidToken = errors.New("유효하지 않은 펜싱 토큰입니다")
	ErrLockExpired  = errors.New("잠금이 만료되었습니다")
)

// ============================================================
// Lock
// ============================================================

type Lock struct {
	Key        string
	Owner      string
	Token      uint64
	AcquiredAt time.Time
	ExpiresAt  time.Time
}

func (l *Lock) IsValid() bool {
	return time.Now().Before(l.ExpiresAt)
}

// ============================================================
// 내부 상태
// ============================================================

// keyState는 특정 키의 잠금 상태를 관리합니다.
type keyState struct {
	lock    *Lock         // 현재 잠금 (nil이면 미잠금)
	waitCh  chan struct{}  // 해제 시 close하여 대기자 깨우기
}

// ============================================================
// LockService
// ============================================================

type LockService struct {
	mu      sync.Mutex
	keys    map[string]*keyState
	token   atomic.Uint64
	// 소유자 -> 현재 대기 중인 키 (데드락 감지용)
	waiting map[string]string
	stopCh  chan struct{}
}

func NewLockService() *LockService {
	s := &LockService{
		keys:    make(map[string]*keyState),
		waiting: make(map[string]string),
		stopCh:  make(chan struct{}),
	}
	// 백그라운드 만료 정리
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.cleanExpired()
			case <-s.stopCh:
				return
			}
		}
	}()
	return s
}

func (s *LockService) getOrCreateKey(key string) *keyState {
	ks, ok := s.keys[key]
	if !ok {
		ks = &keyState{waitCh: make(chan struct{})}
		s.keys[key] = ks
	}
	return ks
}

// Acquire는 key에 대한 잠금을 획득합니다.
func (s *LockService) Acquire(ctx context.Context, key, owner string, ttl time.Duration) (*Lock, error) {
	for {
		s.mu.Lock()
		ks := s.getOrCreateKey(key)

		// 만료 확인 또는 미잠금
		if ks.lock == nil || !ks.lock.IsValid() {
			// 즉시 획득
			token := s.token.Add(1)
			now := time.Now()
			lock := &Lock{
				Key:        key,
				Owner:      owner,
				Token:      token,
				AcquiredAt: now,
				ExpiresAt:  now.Add(ttl),
			}
			ks.lock = lock
			// 대기자에게 알리기 위한 새 채널 (이전 채널은 close됐을 수 있음)
			delete(s.waiting, owner)
			s.mu.Unlock()
			return lock, nil
		}

		// 잠금이 점유 중: 대기 채널 캡처 후 대기
		waitCh := ks.waitCh
		s.waiting[owner] = key
		s.mu.Unlock()

		select {
		case <-waitCh:
			// 잠금이 해제됨, 재시도
			s.mu.Lock()
			delete(s.waiting, owner)
			s.mu.Unlock()
		case <-ctx.Done():
			s.mu.Lock()
			delete(s.waiting, owner)
			s.mu.Unlock()
			return nil, fmt.Errorf("%w: %v", context.Canceled, ctx.Err())
		}
	}
}

// Release는 잠금을 해제합니다.
func (s *LockService) Release(lock *Lock) error {
	s.mu.Lock()

	ks, ok := s.keys[lock.Key]
	if !ok || ks.lock == nil {
		s.mu.Unlock()
		return ErrLockNotFound
	}

	if ks.lock.Token != lock.Token {
		s.mu.Unlock()
		return ErrInvalidToken
	}

	// 잠금 해제 + 대기자 깨우기
	ks.lock = nil
	oldWaitCh := ks.waitCh
	ks.waitCh = make(chan struct{}) // 다음 대기자를 위한 새 채널
	s.mu.Unlock()

	close(oldWaitCh) // 모든 대기자 깨우기
	return nil
}

// Refresh는 잠금의 TTL을 갱신합니다.
func (s *LockService) Refresh(lock *Lock, newTTL time.Duration) (*Lock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ks, ok := s.keys[lock.Key]
	if !ok || ks.lock == nil {
		return nil, ErrLockNotFound
	}

	if ks.lock.Token != lock.Token {
		return nil, ErrInvalidToken
	}

	if !ks.lock.IsValid() {
		return nil, ErrLockExpired
	}

	newLock := &Lock{
		Key:        lock.Key,
		Owner:      lock.Owner,
		Token:      lock.Token,
		AcquiredAt: lock.AcquiredAt,
		ExpiresAt:  time.Now().Add(newTTL),
	}
	ks.lock = newLock
	return newLock, nil
}

func (s *LockService) IsLocked(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	ks, ok := s.keys[key]
	if !ok || ks.lock == nil {
		return false
	}
	return ks.lock.IsValid()
}

func (s *LockService) GetLock(key string) (*Lock, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ks, ok := s.keys[key]
	if !ok || ks.lock == nil || !ks.lock.IsValid() {
		return nil, false
	}
	// 복사본 반환
	cp := *ks.lock
	return &cp, true
}

// cleanExpired는 만료된 잠금을 정리하고 대기자를 깨웁니다.
func (s *LockService) cleanExpired() {
	s.mu.Lock()
	var toWake []chan struct{}

	for _, ks := range s.keys {
		if ks.lock != nil && !ks.lock.IsValid() {
			ks.lock = nil
			oldCh := ks.waitCh
			ks.waitCh = make(chan struct{})
			toWake = append(toWake, oldCh)
		}
	}
	s.mu.Unlock()

	for _, ch := range toWake {
		close(ch)
	}
}

// DetectDeadlock은 소유자 간 대기 사이클을 탐지합니다 (DFS).
func (s *LockService) DetectDeadlock() ([]string, error) {
	s.mu.Lock()

	// 소유자 -> 대기 중인 키 스냅샷
	waiting := make(map[string]string, len(s.waiting))
	for owner, key := range s.waiting {
		waiting[owner] = key
	}

	// 키 -> 현재 소유자 스냅샷
	holders := make(map[string]string)
	for key, ks := range s.keys {
		if ks.lock != nil && ks.lock.IsValid() {
			holders[key] = ks.lock.Owner
		}
	}
	s.mu.Unlock()

	// 소유자 그래프: ownerA -> ownerA가 대기 중인 키의 현재 소유자
	// 즉, ownerA는 ownerB가 보유한 키를 기다림 => edge: A -> B
	graph := make(map[string]string)
	for owner, key := range waiting {
		if holder, ok := holders[key]; ok && holder != owner {
			graph[owner] = holder
		}
	}

	// DFS로 사이클 탐지
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var dfs func(node string, path []string) []string
	dfs = func(node string, path []string) []string {
		if inStack[node] {
			// 사이클 발견: path에서 node부터 잘라냄
			start := 0
			for i, n := range path {
				if n == node {
					start = i
					break
				}
			}
			return path[start:]
		}
		if visited[node] {
			return nil
		}
		visited[node] = true
		inStack[node] = true

		if next, ok := graph[node]; ok {
			if cycle := dfs(next, append(path, node)); cycle != nil {
				return cycle
			}
		}

		inStack[node] = false
		return nil
	}

	for owner := range graph {
		if !visited[owner] {
			if cycle := dfs(owner, nil); cycle != nil {
				return cycle, nil
			}
		}
	}

	return nil, nil
}

func (s *LockService) Close() {
	close(s.stopCh)
}
