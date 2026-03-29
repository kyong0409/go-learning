// todo 패키지: Todo 항목 관리 로직
//
// 이 패키지는 Todo 데이터 구조체와 JSON 파일 기반의
// 저장/불러오기, CRUD 기능을 제공합니다.
package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// ─────────────────────────────────────────
// 타입 정의
// ─────────────────────────────────────────

// Todo: 할 일 항목을 나타내는 구조체
type Todo struct {
	ID          int       `json:"id"`           // 고유 식별자
	Title       string    `json:"title"`        // 할 일 제목
	Done        bool      `json:"done"`         // 완료 여부
	CreatedAt   time.Time `json:"created_at"`   // 생성 시간
	CompletedAt *time.Time `json:"completed_at,omitempty"` // 완료 시간 (nil이면 미완료)
}

// String: Todo의 문자열 표현
func (t Todo) String() string {
	status := "[ ]"
	if t.Done {
		status = "[x]"
	}
	return fmt.Sprintf("%s [%d] %s", status, t.ID, t.Title)
}

// TodoList: Todo 항목들의 컬렉션
type TodoList struct {
	Items  []Todo `json:"items"`
	NextID int    `json:"next_id"` // 다음에 사용할 ID
}

// ─────────────────────────────────────────
// 센티넬 에러
// ─────────────────────────────────────────

// ErrNotFound: 해당 ID의 Todo를 찾을 수 없을 때
var ErrNotFound = errors.New("todo를 찾을 수 없습니다")

// ErrEmptyTitle: 빈 제목으로 Todo를 추가하려 할 때
var ErrEmptyTitle = errors.New("제목이 비어있습니다")

// ─────────────────────────────────────────
// TodoList 생성
// ─────────────────────────────────────────

// NewTodoList: 빈 TodoList를 생성합니다.
func NewTodoList() *TodoList {
	return &TodoList{
		Items:  []Todo{},
		NextID: 1,
	}
}

// ─────────────────────────────────────────
// JSON 파일 저장/불러오기
// ─────────────────────────────────────────

// Load: JSON 파일에서 TodoList를 불러옵니다.
// 파일이 없으면 새 빈 목록을 반환합니다.
func Load(filename string) (*TodoList, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// 파일이 없으면 새 목록 반환 (에러 아님)
			return NewTodoList(), nil
		}
		return nil, fmt.Errorf("파일 읽기 실패 %q: %w", filename, err)
	}

	var list TodoList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
	}

	// NextID 초기화: 기존 항목보다 큰 값으로 설정
	if list.NextID == 0 {
		list.NextID = 1
		for _, item := range list.Items {
			if item.ID >= list.NextID {
				list.NextID = item.ID + 1
			}
		}
	}

	return &list, nil
}

// Save: TodoList를 JSON 파일에 저장합니다.
func (l *TodoList) Save(filename string) error {
	// 들여쓰기 포함해서 읽기 좋게 저장
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 직렬화 실패: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("파일 쓰기 실패 %q: %w", filename, err)
	}

	return nil
}

// ─────────────────────────────────────────
// CRUD 함수들
// ─────────────────────────────────────────

// Add: 새 Todo 항목을 추가합니다.
func (l *TodoList) Add(title string) (*Todo, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}

	todo := Todo{
		ID:        l.NextID,
		Title:     title,
		Done:      false,
		CreatedAt: time.Now(),
	}

	l.Items = append(l.Items, todo)
	l.NextID++

	return &l.Items[len(l.Items)-1], nil
}

// List: 모든 Todo 항목을 반환합니다.
// doneOnly가 true이면 완료된 항목만, false이면 미완료 항목만, nil이면 전체를 반환합니다.
func (l *TodoList) List(doneFilter *bool) []Todo {
	if doneFilter == nil {
		// 전체 반환 (복사본)
		result := make([]Todo, len(l.Items))
		copy(result, l.Items)
		return result
	}

	var result []Todo
	for _, item := range l.Items {
		if item.Done == *doneFilter {
			result = append(result, item)
		}
	}
	return result
}

// findByID: ID로 항목의 인덱스를 찾습니다 (비공개).
func (l *TodoList) findByID(id int) (int, bool) {
	for i, item := range l.Items {
		if item.ID == id {
			return i, true
		}
	}
	return -1, false
}

// Get: ID로 단일 Todo를 조회합니다.
func (l *TodoList) Get(id int) (*Todo, error) {
	idx, found := l.findByID(id)
	if !found {
		return nil, fmt.Errorf("ID %d: %w", id, ErrNotFound)
	}
	return &l.Items[idx], nil
}

// Complete: 지정한 ID의 Todo를 완료 처리합니다.
func (l *TodoList) Complete(id int) error {
	idx, found := l.findByID(id)
	if !found {
		return fmt.Errorf("ID %d: %w", id, ErrNotFound)
	}

	if l.Items[idx].Done {
		return fmt.Errorf("ID %d는 이미 완료된 항목입니다", id)
	}

	now := time.Now()
	l.Items[idx].Done = true
	l.Items[idx].CompletedAt = &now
	return nil
}

// Uncomplete: 지정한 ID의 Todo를 미완료로 되돌립니다.
func (l *TodoList) Uncomplete(id int) error {
	idx, found := l.findByID(id)
	if !found {
		return fmt.Errorf("ID %d: %w", id, ErrNotFound)
	}

	l.Items[idx].Done = false
	l.Items[idx].CompletedAt = nil
	return nil
}

// Delete: 지정한 ID의 Todo를 삭제합니다.
func (l *TodoList) Delete(id int) error {
	idx, found := l.findByID(id)
	if !found {
		return fmt.Errorf("ID %d: %w", id, ErrNotFound)
	}

	// 순서를 유지하면서 삭제
	l.Items = append(l.Items[:idx], l.Items[idx+1:]...)
	return nil
}

// Update: 지정한 ID의 Todo 제목을 수정합니다.
func (l *TodoList) Update(id int, newTitle string) error {
	if newTitle == "" {
		return ErrEmptyTitle
	}

	idx, found := l.findByID(id)
	if !found {
		return fmt.Errorf("ID %d: %w", id, ErrNotFound)
	}

	l.Items[idx].Title = newTitle
	return nil
}

// Count: 전체, 완료, 미완료 항목 수를 반환합니다.
func (l *TodoList) Count() (total, done, pending int) {
	total = len(l.Items)
	for _, item := range l.Items {
		if item.Done {
			done++
		}
	}
	pending = total - done
	return
}

// Clear: 완료된 항목을 모두 삭제합니다.
func (l *TodoList) ClearDone() int {
	original := len(l.Items)
	var remaining []Todo
	for _, item := range l.Items {
		if !item.Done {
			remaining = append(remaining, item)
		}
	}
	l.Items = remaining
	if l.Items == nil {
		l.Items = []Todo{}
	}
	return original - len(l.Items)
}
