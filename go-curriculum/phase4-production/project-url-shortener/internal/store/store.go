// internal/store/store.go
// 데이터 저장소 인터페이스를 정의합니다.
// 인터페이스를 통해 메모리 저장소와 PostgreSQL 저장소를 교체 가능하게 합니다.
package store

import (
	"context"

	"github.com/learn-go/url-shortener/internal/model"
)

// Store는 URL 저장소 인터페이스입니다.
// 이 인터페이스를 구현하면 메모리, PostgreSQL, Redis 등 어떤 백엔드도 사용할 수 있습니다.
type Store interface {
	// Create는 새 단축 URL을 저장합니다.
	// 단축 코드가 이미 존재하면 ConflictError를 반환합니다.
	Create(ctx context.Context, shortURL *model.ShortURL) error

	// GetByCode는 단축 코드로 URL을 조회합니다.
	// 존재하지 않으면 NotFoundError를 반환합니다.
	GetByCode(ctx context.Context, code string) (*model.ShortURL, error)

	// GetByID는 ID로 URL을 조회합니다.
	GetByID(ctx context.Context, id int64) (*model.ShortURL, error)

	// IncrementClickCount는 클릭 횟수를 1 증가시킵니다.
	IncrementClickCount(ctx context.Context, code string) error

	// List는 페이지네이션된 URL 목록을 반환합니다.
	List(ctx context.Context, page, pageSize int) ([]*model.ShortURL, int, error)

	// Delete는 단축 URL을 삭제합니다.
	Delete(ctx context.Context, code string) error

	// CodeExists는 단축 코드가 이미 존재하는지 확인합니다.
	CodeExists(ctx context.Context, code string) (bool, error)

	// Close는 저장소 연결을 닫습니다.
	Close()
}
