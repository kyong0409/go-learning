// internal/model/url.go
// URL 단축 서비스의 도메인 모델을 정의합니다.
package model

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ShortURL은 단축 URL 엔티티입니다.
type ShortURL struct {
	// ID는 데이터베이스 기본 키입니다.
	ID int64 `json:"id"`
	// OriginalURL은 원본 긴 URL입니다.
	OriginalURL string `json:"original_url"`
	// ShortCode는 단축 코드입니다 (예: "abc123").
	ShortCode string `json:"short_code"`
	// ClickCount는 클릭 횟수입니다.
	ClickCount int64 `json:"click_count"`
	// CreatedAt은 생성 시각입니다.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt은 최종 수정 시각입니다.
	UpdatedAt time.Time `json:"updated_at"`
	// ExpiresAt은 만료 시각입니다 (nil이면 만료 없음).
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// IsExpired는 URL이 만료되었는지 확인합니다.
func (s *ShortURL) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

// ShortenRequest는 URL 단축 요청 구조입니다.
type ShortenRequest struct {
	// URL은 단축할 원본 URL입니다 (필수).
	URL string `json:"url"`
	// CustomCode는 사용자 지정 단축 코드입니다 (선택).
	CustomCode string `json:"custom_code,omitempty"`
	// ExpiresIn은 만료까지의 시간입니다 (예: "24h", "7d").
	ExpiresIn string `json:"expires_in,omitempty"`
}

// Validate는 요청의 유효성을 검사합니다.
func (r *ShortenRequest) Validate() error {
	// URL 필수 확인
	if strings.TrimSpace(r.URL) == "" {
		return &ValidationError{Field: "url", Message: "URL은 필수입니다"}
	}

	// URL 형식 검사
	parsed, err := url.ParseRequestURI(r.URL)
	if err != nil {
		return &ValidationError{Field: "url", Message: "유효하지 않은 URL 형식입니다"}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return &ValidationError{Field: "url", Message: "http 또는 https URL만 허용됩니다"}
	}

	// 커스텀 코드 검사
	if r.CustomCode != "" {
		if len(r.CustomCode) < 3 || len(r.CustomCode) > 20 {
			return &ValidationError{Field: "custom_code", Message: "커스텀 코드는 3~20자여야 합니다"}
		}
		for _, c := range r.CustomCode {
			if !isAllowedChar(c) {
				return &ValidationError{
					Field:   "custom_code",
					Message: "커스텀 코드는 영문자, 숫자, 하이픈만 사용할 수 있습니다",
				}
			}
		}
	}

	return nil
}

// isAllowedChar는 단축 코드에 허용된 문자인지 확인합니다.
func isAllowedChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_'
}

// ShortenResponse는 URL 단축 응답 구조입니다.
type ShortenResponse struct {
	// ShortURL은 완성된 단축 URL입니다 (예: "http://short.ly/abc123").
	ShortURL string `json:"short_url"`
	// ShortCode는 단축 코드입니다.
	ShortCode string `json:"short_code"`
	// OriginalURL은 원본 URL입니다.
	OriginalURL string `json:"original_url"`
	// ExpiresAt은 만료 시각입니다.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// StatsResponse는 URL 통계 응답 구조입니다.
type StatsResponse struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	ClickCount  int64      `json:"click_count"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ListResponse는 URL 목록 응답 구조입니다.
type ListResponse struct {
	Items      []*ShortURL `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// ErrorResponse는 API 오류 응답 구조입니다.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ============================================================
// 오류 타입
// ============================================================

// ValidationError는 입력 검증 오류입니다.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("검증 오류 [%s]: %s", e.Field, e.Message)
}

// NotFoundError는 리소스를 찾을 수 없을 때 발생하는 오류입니다.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s '%s'를 찾을 수 없습니다", e.Resource, e.ID)
}

// ConflictError는 리소스 충돌 오류입니다 (예: 중복 단축 코드).
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}
