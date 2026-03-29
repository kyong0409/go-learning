// internal/store/memory.go
// 테스트 및 개발용 인메모리 저장소 구현입니다.
// 실제 데이터베이스 없이 서버를 실행하거나 테스트할 때 사용합니다.
package store

import (
	"context"
	"sync"
	"time"

	"github.com/learn-go/url-shortener/internal/model"
)

// MemoryStore는 메모리 기반 URL 저장소입니다.
// 동시성 안전을 위해 sync.RWMutex를 사용합니다.
type MemoryStore struct {
	mu      sync.RWMutex
	byCode  map[string]*model.ShortURL // 단축 코드 -> ShortURL
	byID    map[int64]*model.ShortURL  // ID -> ShortURL
	nextID  int64
	ordered []string // 삽입 순서 유지 (List용)
}

// NewMemoryStore는 새 인메모리 저장소를 생성합니다.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byCode: make(map[string]*model.ShortURL),
		byID:   make(map[int64]*model.ShortURL),
		nextID: 1,
	}
}

// Create는 새 단축 URL을 저장합니다.
func (s *MemoryStore) Create(ctx context.Context, shortURL *model.ShortURL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 중복 코드 확인
	if _, exists := s.byCode[shortURL.ShortCode]; exists {
		return &model.ConflictError{
			Message: "단축 코드 '" + shortURL.ShortCode + "'가 이미 존재합니다",
		}
	}

	// ID 및 시간 설정
	shortURL.ID = s.nextID
	s.nextID++
	now := time.Now()
	shortURL.CreatedAt = now
	shortURL.UpdatedAt = now

	// 복사본 저장 (외부에서 수정되지 않도록)
	copy := *shortURL
	s.byCode[shortURL.ShortCode] = &copy
	s.byID[shortURL.ID] = &copy
	s.ordered = append(s.ordered, shortURL.ShortCode)

	return nil
}

// GetByCode는 단축 코드로 URL을 조회합니다.
func (s *MemoryStore) GetByCode(ctx context.Context, code string) (*model.ShortURL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortURL, exists := s.byCode[code]
	if !exists {
		return nil, &model.NotFoundError{Resource: "ShortURL", ID: code}
	}

	// 복사본 반환
	copy := *shortURL
	return &copy, nil
}

// GetByID는 ID로 URL을 조회합니다.
func (s *MemoryStore) GetByID(ctx context.Context, id int64) (*model.ShortURL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortURL, exists := s.byID[id]
	if !exists {
		return nil, &model.NotFoundError{
			Resource: "ShortURL",
			ID:       string(rune(id + '0')),
		}
	}

	copy := *shortURL
	return &copy, nil
}

// IncrementClickCount는 클릭 횟수를 1 증가시킵니다.
func (s *MemoryStore) IncrementClickCount(ctx context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	shortURL, exists := s.byCode[code]
	if !exists {
		return &model.NotFoundError{Resource: "ShortURL", ID: code}
	}

	shortURL.ClickCount++
	shortURL.UpdatedAt = time.Now()
	// byID도 동일한 포인터이므로 자동으로 업데이트됩니다.

	return nil
}

// List는 페이지네이션된 URL 목록을 반환합니다.
func (s *MemoryStore) List(ctx context.Context, page, pageSize int) ([]*model.ShortURL, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.ordered)

	// 페이지네이션 계산
	start := (page - 1) * pageSize
	if start >= total {
		return []*model.ShortURL{}, total, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	items := make([]*model.ShortURL, 0, end-start)
	for _, code := range s.ordered[start:end] {
		if shortURL, exists := s.byCode[code]; exists {
			copy := *shortURL
			items = append(items, &copy)
		}
	}

	return items, total, nil
}

// Delete는 단축 URL을 삭제합니다.
func (s *MemoryStore) Delete(ctx context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	shortURL, exists := s.byCode[code]
	if !exists {
		return &model.NotFoundError{Resource: "ShortURL", ID: code}
	}

	// 모든 맵에서 제거
	delete(s.byID, shortURL.ID)
	delete(s.byCode, code)

	// ordered 슬라이스에서 제거
	for i, c := range s.ordered {
		if c == code {
			s.ordered = append(s.ordered[:i], s.ordered[i+1:]...)
			break
		}
	}

	return nil
}

// CodeExists는 단축 코드가 이미 존재하는지 확인합니다.
func (s *MemoryStore) CodeExists(ctx context.Context, code string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.byCode[code]
	return exists, nil
}

// Close는 인메모리 저장소의 정리 작업을 수행합니다 (실제로는 아무것도 하지 않음).
func (s *MemoryStore) Close() {}

// Count는 저장된 URL의 총 개수를 반환합니다 (테스트 헬퍼).
func (s *MemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.byCode)
}
