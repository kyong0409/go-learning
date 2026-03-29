// a3-rest-api/model.go
// 도서 관리 REST API 도메인 모델입니다.
package api

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Book은 도서 엔티티입니다.
type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	ISBN        string    `json:"isbn,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateBookRequest는 도서 생성 요청입니다.
type CreateBookRequest struct {
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	ISBN        string    `json:"isbn"`
	PublishedAt time.Time `json:"published_at"`
	Price       float64   `json:"price"`
}

// Validate는 입력 유효성을 검사합니다.
func (r *CreateBookRequest) Validate() *ValidationError {
	if strings.TrimSpace(r.Title) == "" {
		return &ValidationError{Field: "title", Message: "제목은 필수입니다"}
	}
	if len(r.Title) > 200 {
		return &ValidationError{Field: "title", Message: "제목은 200자 이하여야 합니다"}
	}
	if strings.TrimSpace(r.Author) == "" {
		return &ValidationError{Field: "author", Message: "저자는 필수입니다"}
	}
	if len(r.Author) > 100 {
		return &ValidationError{Field: "author", Message: "저자는 100자 이하여야 합니다"}
	}
	if r.ISBN != "" && len(r.ISBN) != 10 && len(r.ISBN) != 13 {
		return &ValidationError{Field: "isbn", Message: "ISBN은 10자 또는 13자여야 합니다"}
	}
	if r.Price < 0 {
		return &ValidationError{Field: "price", Message: "가격은 0 이상이어야 합니다"}
	}
	return nil
}

// UpdateBookRequest는 도서 수정 요청입니다 (부분 수정 지원).
type UpdateBookRequest struct {
	Title       *string    `json:"title"`
	Author      *string    `json:"author"`
	ISBN        *string    `json:"isbn"`
	PublishedAt *time.Time `json:"published_at"`
	Price       *float64   `json:"price"`
}

// ListBooksQuery는 목록 조회 쿼리 파라미터입니다.
type ListBooksQuery struct {
	Page     int
	PageSize int
	Search   string // 제목 또는 저자 검색
	Author   string // 저자 필터
}

// ListBooksResponse는 목록 조회 응답입니다.
type ListBooksResponse struct {
	Items      []*Book `json:"items"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

// ErrorResponse는 API 오류 응답입니다.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Field   string `json:"field,omitempty"`
}

// ValidationError는 입력 검증 오류입니다.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Field, e.Message)
}

// Store는 도서 저장소 인터페이스입니다.
type Store interface {
	Create(ctx context.Context, book *Book) error
	GetByID(ctx context.Context, id int) (*Book, error)
	Update(ctx context.Context, book *Book) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, q ListBooksQuery) ([]*Book, int, error)
}

// MemoryStore는 테스트용 인메모리 저장소입니다.
type MemoryStore struct {
	books  map[int]*Book
	nextID int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{books: make(map[int]*Book), nextID: 1}
}

func (s *MemoryStore) Create(ctx context.Context, book *Book) error {
	book.ID = s.nextID
	s.nextID++
	now := time.Now()
	book.CreatedAt = now
	book.UpdatedAt = now
	cp := *book
	s.books[book.ID] = &cp
	return nil
}

func (s *MemoryStore) GetByID(ctx context.Context, id int) (*Book, error) {
	b, ok := s.books[id]
	if !ok {
		return nil, fmt.Errorf("book %d not found", id)
	}
	cp := *b
	return &cp, nil
}

func (s *MemoryStore) Update(ctx context.Context, book *Book) error {
	if _, ok := s.books[book.ID]; !ok {
		return fmt.Errorf("book %d not found", book.ID)
	}
	book.UpdatedAt = time.Now()
	cp := *book
	s.books[book.ID] = &cp
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, id int) error {
	if _, ok := s.books[id]; !ok {
		return fmt.Errorf("book %d not found", id)
	}
	delete(s.books, id)
	return nil
}

func (s *MemoryStore) List(ctx context.Context, q ListBooksQuery) ([]*Book, int, error) {
	var filtered []*Book
	for _, b := range s.books {
		if q.Search != "" {
			if !strings.Contains(strings.ToLower(b.Title), strings.ToLower(q.Search)) &&
				!strings.Contains(strings.ToLower(b.Author), strings.ToLower(q.Search)) {
				continue
			}
		}
		if q.Author != "" && !strings.Contains(strings.ToLower(b.Author), strings.ToLower(q.Author)) {
			continue
		}
		cp := *b
		filtered = append(filtered, &cp)
	}
	total := len(filtered)
	start := (q.Page - 1) * q.PageSize
	if start >= total {
		return []*Book{}, total, nil
	}
	end := start + q.PageSize
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}
