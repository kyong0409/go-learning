// todo 패키지 테스트
// 실행: go test ./todo/ -v
package todo

import (
	"errors"
	"os"
	"testing"
	"time"
)

// ─────────────────────────────────────────
// 테스트 헬퍼
// ─────────────────────────────────────────

// newTestList: 테스트용 TodoList 생성 (항목 3개 포함)
func newTestList() *TodoList {
	l := NewTodoList()
	l.Add("Go 기초 학습")
	l.Add("프로젝트 만들기")
	l.Add("테스트 작성")
	return l
}

// ─────────────────────────────────────────
// NewTodoList 테스트
// ─────────────────────────────────────────

func TestNewTodoList(t *testing.T) {
	l := NewTodoList()
	if l == nil {
		t.Fatal("NewTodoList()이 nil을 반환했습니다")
	}
	if len(l.Items) != 0 {
		t.Errorf("새 목록의 Items 길이: got %d, want 0", len(l.Items))
	}
	if l.NextID != 1 {
		t.Errorf("새 목록의 NextID: got %d, want 1", l.NextID)
	}
}

// ─────────────────────────────────────────
// Add 테스트
// ─────────────────────────────────────────

func TestAdd(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr bool
		errType error
	}{
		{
			name:    "정상 추가",
			title:   "할 일 1",
			wantErr: false,
		},
		{
			name:    "빈 제목 거부",
			title:   "",
			wantErr: true,
			errType: ErrEmptyTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewTodoList()
			todo, err := l.Add(tt.title)

			if tt.wantErr {
				if err == nil {
					t.Error("에러가 발생해야 하는데 nil이 반환됨")
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("에러 타입: got %v, want %v", err, tt.errType)
				}
				return
			}

			if err != nil {
				t.Fatalf("예상치 못한 에러: %v", err)
			}
			if todo == nil {
				t.Fatal("todo가 nil입니다")
			}
			if todo.Title != tt.title {
				t.Errorf("제목: got %q, want %q", todo.Title, tt.title)
			}
			if todo.Done {
				t.Error("새 todo는 완료 상태이면 안 됩니다")
			}
			if todo.ID != 1 {
				t.Errorf("첫 번째 ID: got %d, want 1", todo.ID)
			}
			if len(l.Items) != 1 {
				t.Errorf("Items 길이: got %d, want 1", len(l.Items))
			}
		})
	}
}

func TestAdd_IDIncrement(t *testing.T) {
	l := NewTodoList()
	for i := 1; i <= 5; i++ {
		todo, err := l.Add("항목")
		if err != nil {
			t.Fatalf("Add 실패: %v", err)
		}
		if todo.ID != i {
			t.Errorf("ID: got %d, want %d", todo.ID, i)
		}
	}
}

// ─────────────────────────────────────────
// List 테스트
// ─────────────────────────────────────────

func TestList_All(t *testing.T) {
	l := newTestList()
	items := l.List(nil) // 전체 조회
	if len(items) != 3 {
		t.Errorf("전체 목록 길이: got %d, want 3", len(items))
	}
}

func TestList_Filter(t *testing.T) {
	l := newTestList()
	// ID 1을 완료 처리
	l.Complete(1)

	trueVal := true
	falseVal := false

	doneItems := l.List(&trueVal)
	if len(doneItems) != 1 {
		t.Errorf("완료 항목 수: got %d, want 1", len(doneItems))
	}

	pendingItems := l.List(&falseVal)
	if len(pendingItems) != 2 {
		t.Errorf("미완료 항목 수: got %d, want 2", len(pendingItems))
	}
}

// ─────────────────────────────────────────
// Complete 테스트
// ─────────────────────────────────────────

func TestComplete(t *testing.T) {
	l := newTestList()

	// 정상 완료
	if err := l.Complete(1); err != nil {
		t.Fatalf("Complete(1) 실패: %v", err)
	}
	todo, _ := l.Get(1)
	if !todo.Done {
		t.Error("완료 처리 후 Done이 true여야 합니다")
	}
	if todo.CompletedAt == nil {
		t.Error("완료 처리 후 CompletedAt이 설정되어야 합니다")
	}
	// 완료 시간이 현재 시간에 가까운지 확인
	if time.Since(*todo.CompletedAt) > time.Second {
		t.Error("CompletedAt이 현재 시간에 가깝지 않습니다")
	}

	// 이미 완료된 항목 재완료 시도
	if err := l.Complete(1); err == nil {
		t.Error("이미 완료된 항목에 Complete를 호출하면 에러가 발생해야 합니다")
	}

	// 존재하지 않는 ID
	err := l.Complete(999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("존재하지 않는 ID: got %v, want ErrNotFound", err)
	}
}

// ─────────────────────────────────────────
// Delete 테스트
// ─────────────────────────────────────────

func TestDelete(t *testing.T) {
	l := newTestList()

	// 정상 삭제
	if err := l.Delete(2); err != nil {
		t.Fatalf("Delete(2) 실패: %v", err)
	}
	if len(l.Items) != 2 {
		t.Errorf("삭제 후 Items 길이: got %d, want 2", len(l.Items))
	}

	// 삭제된 항목이 없는지 확인
	_, err := l.Get(2)
	if !errors.Is(err, ErrNotFound) {
		t.Error("삭제된 항목이 조회되면 안 됩니다")
	}

	// 존재하지 않는 ID 삭제
	err = l.Delete(999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("존재하지 않는 ID: got %v, want ErrNotFound", err)
	}
}

// ─────────────────────────────────────────
// Update 테스트
// ─────────────────────────────────────────

func TestUpdate(t *testing.T) {
	l := newTestList()

	// 정상 업데이트
	if err := l.Update(1, "새 제목"); err != nil {
		t.Fatalf("Update(1) 실패: %v", err)
	}
	todo, _ := l.Get(1)
	if todo.Title != "새 제목" {
		t.Errorf("업데이트 후 제목: got %q, want %q", todo.Title, "새 제목")
	}

	// 빈 제목
	if err := l.Update(1, ""); !errors.Is(err, ErrEmptyTitle) {
		t.Errorf("빈 제목: got %v, want ErrEmptyTitle", err)
	}

	// 존재하지 않는 ID
	if err := l.Update(999, "제목"); !errors.Is(err, ErrNotFound) {
		t.Errorf("존재하지 않는 ID: got %v, want ErrNotFound", err)
	}
}

// ─────────────────────────────────────────
// Count 테스트
// ─────────────────────────────────────────

func TestCount(t *testing.T) {
	l := newTestList()
	l.Complete(1)
	l.Complete(2)

	total, done, pending := l.Count()
	if total != 3 {
		t.Errorf("total: got %d, want 3", total)
	}
	if done != 2 {
		t.Errorf("done: got %d, want 2", done)
	}
	if pending != 1 {
		t.Errorf("pending: got %d, want 1", pending)
	}
}

// ─────────────────────────────────────────
// ClearDone 테스트
// ─────────────────────────────────────────

func TestClearDone(t *testing.T) {
	l := newTestList()
	l.Complete(1)
	l.Complete(3)

	removed := l.ClearDone()
	if removed != 2 {
		t.Errorf("삭제된 수: got %d, want 2", removed)
	}
	if len(l.Items) != 1 {
		t.Errorf("남은 항목 수: got %d, want 1", len(l.Items))
	}
	if l.Items[0].ID != 2 {
		t.Errorf("남은 항목 ID: got %d, want 2", l.Items[0].ID)
	}
}

// ─────────────────────────────────────────
// Save/Load 테스트
// ─────────────────────────────────────────

func TestSaveLoad(t *testing.T) {
	// 임시 파일 생성
	tmpFile, err := os.CreateTemp("", "todo_test_*.json")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// 저장할 목록 생성
	original := newTestList()
	original.Complete(1)

	// 저장
	if err := original.Save(tmpFile.Name()); err != nil {
		t.Fatalf("Save 실패: %v", err)
	}

	// 불러오기
	loaded, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load 실패: %v", err)
	}

	// 비교
	if len(loaded.Items) != len(original.Items) {
		t.Errorf("불러온 항목 수: got %d, want %d",
			len(loaded.Items), len(original.Items))
	}

	for i, item := range loaded.Items {
		orig := original.Items[i]
		if item.ID != orig.ID {
			t.Errorf("[%d] ID: got %d, want %d", i, item.ID, orig.ID)
		}
		if item.Title != orig.Title {
			t.Errorf("[%d] Title: got %q, want %q", i, item.Title, orig.Title)
		}
		if item.Done != orig.Done {
			t.Errorf("[%d] Done: got %t, want %t", i, item.Done, orig.Done)
		}
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	// 존재하지 않는 파일 로드 -> 빈 목록 반환 (에러 아님)
	list, err := Load("/tmp/nonexistent_todo_test.json")
	if err != nil {
		t.Fatalf("존재하지 않는 파일 Load가 에러를 반환했습니다: %v", err)
	}
	if list == nil {
		t.Fatal("list가 nil입니다")
	}
	if len(list.Items) != 0 {
		t.Errorf("빈 목록이어야 합니다: got %d items", len(list.Items))
	}
}
