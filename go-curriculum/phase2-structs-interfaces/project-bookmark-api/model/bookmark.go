// Package model은 북마크 데이터 모델과 저장소를 정의합니다.
package model

import (
	"errors"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 에러 정의
// ─────────────────────────────────────────

// ErrNotFound는 북마크를 찾지 못했을 때 반환됩니다.
var ErrNotFound = errors.New("북마크를 찾을 수 없습니다")

// ErrInvalidInput은 입력값이 유효하지 않을 때 반환됩니다.
var ErrInvalidInput = errors.New("유효하지 않은 입력입니다")

// ─────────────────────────────────────────
// Bookmark 구조체
// ─────────────────────────────────────────

// Bookmark는 저장된 URL 북마크를 나타냅니다.
type Bookmark struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateBookmarkRequest는 북마크 생성 요청 본문입니다.
type CreateBookmarkRequest struct {
	Title string   `json:"title"`
	URL   string   `json:"url"`
	Tags  []string `json:"tags,omitempty"`
}

// UpdateBookmarkRequest는 북마크 수정 요청 본문입니다.
type UpdateBookmarkRequest struct {
	Title string   `json:"title,omitempty"`
	URL   string   `json:"url,omitempty"`
	Tags  []string `json:"tags,omitempty"`
}

// ─────────────────────────────────────────
// Store 인터페이스
// ─────────────────────────────────────────
// 인터페이스를 정의하여 구현을 교체할 수 있게 합니다.
// (인메모리 → PostgreSQL 등으로 교체 가능)

// Store는 북마크 저장소 인터페이스입니다.
type Store interface {
	// GetAll은 모든 북마크를 반환합니다.
	GetAll() ([]*Bookmark, error)

	// GetByID는 특정 ID의 북마크를 반환합니다.
	// 없으면 ErrNotFound를 반환합니다.
	GetByID(id int) (*Bookmark, error)

	// Create는 새 북마크를 생성하고 반환합니다.
	Create(req CreateBookmarkRequest) (*Bookmark, error)

	// Update는 기존 북마크를 수정하고 반환합니다.
	// 없으면 ErrNotFound를 반환합니다.
	Update(id int, req UpdateBookmarkRequest) (*Bookmark, error)

	// Delete는 북마크를 삭제합니다.
	// 없으면 ErrNotFound를 반환합니다.
	Delete(id int) error
}

// ─────────────────────────────────────────
// InMemoryStore 구현
// ─────────────────────────────────────────

// InMemoryStore는 메모리 기반 북마크 저장소입니다.
// Store 인터페이스를 구현합니다.
type InMemoryStore struct {
	mu       sync.RWMutex    // 동시 접근 보호
	items    map[int]*Bookmark
	nextID   int
}

// NewInMemoryStore는 InMemoryStore 생성자입니다.
// Store 인터페이스를 반환합니다. ("인터페이스를 반환하라" 패턴)
func NewInMemoryStore() Store {
	return &InMemoryStore{
		items:  make(map[int]*Bookmark),
		nextID: 1,
	}
}

// GetAll은 모든 북마크를 ID 순으로 반환합니다.
func (s *InMemoryStore) GetAll() ([]*Bookmark, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Bookmark, 0, len(s.items))
	for _, b := range s.items {
		// 복사본 반환 (외부 수정 방지)
		copy := *b
		result = append(result, &copy)
	}

	// ID 순으로 정렬 (간단한 삽입 정렬)
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j].ID < result[j-1].ID; j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}

	return result, nil
}

// GetByID는 ID로 북마크를 조회합니다.
func (s *InMemoryStore) GetByID(id int) (*Bookmark, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.items[id]
	if !ok {
		return nil, ErrNotFound
	}

	// 복사본 반환
	copy := *b
	return &copy, nil
}

// Create는 새 북마크를 생성합니다.
func (s *InMemoryStore) Create(req CreateBookmarkRequest) (*Bookmark, error) {
	// 유효성 검사
	if req.Title == "" {
		return nil, ErrInvalidInput
	}
	if req.URL == "" {
		return nil, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	b := &Bookmark{
		ID:        s.nextID,
		Title:     req.Title,
		URL:       req.URL,
		Tags:      req.Tags,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.items[s.nextID] = b
	s.nextID++

	// 복사본 반환
	copy := *b
	return &copy, nil
}

// Update는 기존 북마크를 수정합니다.
func (s *InMemoryStore) Update(id int, req UpdateBookmarkRequest) (*Bookmark, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.items[id]
	if !ok {
		return nil, ErrNotFound
	}

	// 제공된 필드만 업데이트 (partial update)
	if req.Title != "" {
		b.Title = req.Title
	}
	if req.URL != "" {
		b.URL = req.URL
	}
	if req.Tags != nil {
		b.Tags = req.Tags
	}
	b.UpdatedAt = time.Now()

	// 복사본 반환
	copy := *b
	return &copy, nil
}

// Delete는 북마크를 삭제합니다.
func (s *InMemoryStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}

	delete(s.items, id)
	return nil
}
