// a1-generic-collection/collection_test.go
// 제네릭 컬렉션 라이브러리 테스트입니다.
// 모든 테스트를 통과하면 100점입니다.
package collection_test

import (
	"fmt"
	"os"
	"testing"

	collection "github.com/learn-go/a1-generic-collection"
)

// ============================================================
// Stack 테스트 (25점)
// ============================================================

func TestStack_BasicOperations(t *testing.T) {
	s := &collection.Stack[int]{}

	// 빈 스택 확인
	if !s.IsEmpty() {
		t.Error("새 스택은 비어있어야 합니다")
	}
	if s.Len() != 0 {
		t.Errorf("새 스택의 길이는 0이어야 합니다. 실제: %d", s.Len())
	}

	// Push
	s.Push(1)
	s.Push(2)
	s.Push(3)

	if s.Len() != 3 {
		t.Errorf("Push 3회 후 길이는 3이어야 합니다. 실제: %d", s.Len())
	}
	if s.IsEmpty() {
		t.Error("원소 추가 후 IsEmpty는 false여야 합니다")
	}
}

func TestStack_PopOrder(t *testing.T) {
	s := &collection.Stack[string]{}
	s.Push("first")
	s.Push("second")
	s.Push("third")

	// LIFO 순서 확인
	expected := []string{"third", "second", "first"}
	for i, want := range expected {
		got, ok := s.Pop()
		if !ok {
			t.Errorf("Pop %d번째: ok=false, 원소가 있어야 합니다", i+1)
		}
		if got != want {
			t.Errorf("Pop %d번째: 기대 %q, 실제 %q", i+1, want, got)
		}
	}

	// 빈 스택에서 Pop
	_, ok := s.Pop()
	if ok {
		t.Error("빈 스택에서 Pop 시 ok=false여야 합니다")
	}
}

func TestStack_Peek(t *testing.T) {
	s := &collection.Stack[int]{}
	s.Push(42)

	// Peek은 제거하지 않아야 함
	v, ok := s.Peek()
	if !ok || v != 42 {
		t.Errorf("Peek: 기대 (42, true), 실제 (%d, %v)", v, ok)
	}
	if s.Len() != 1 {
		t.Error("Peek 후 길이가 변하면 안 됩니다")
	}

	// 빈 스택 Peek
	s.Pop()
	_, ok = s.Peek()
	if ok {
		t.Error("빈 스택에서 Peek 시 ok=false여야 합니다")
	}
}

// ============================================================
// Queue 테스트 (25점)
// ============================================================

func TestQueue_BasicOperations(t *testing.T) {
	q := &collection.Queue[int]{}

	if !q.IsEmpty() {
		t.Error("새 큐는 비어있어야 합니다")
	}

	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	if q.Len() != 3 {
		t.Errorf("Enqueue 3회 후 길이는 3이어야 합니다. 실제: %d", q.Len())
	}
}

func TestQueue_DequeueOrder(t *testing.T) {
	q := &collection.Queue[string]{}
	q.Enqueue("first")
	q.Enqueue("second")
	q.Enqueue("third")

	// FIFO 순서 확인
	expected := []string{"first", "second", "third"}
	for i, want := range expected {
		got, ok := q.Dequeue()
		if !ok {
			t.Errorf("Dequeue %d번째: ok=false", i+1)
		}
		if got != want {
			t.Errorf("Dequeue %d번째: 기대 %q, 실제 %q", i+1, want, got)
		}
	}

	_, ok := q.Dequeue()
	if ok {
		t.Error("빈 큐에서 Dequeue 시 ok=false여야 합니다")
	}
}

func TestQueue_Front(t *testing.T) {
	q := &collection.Queue[int]{}
	q.Enqueue(10)
	q.Enqueue(20)

	v, ok := q.Front()
	if !ok || v != 10 {
		t.Errorf("Front: 기대 (10, true), 실제 (%d, %v)", v, ok)
	}
	if q.Len() != 2 {
		t.Error("Front 후 길이가 변하면 안 됩니다")
	}
}

// ============================================================
// Set 테스트 (25점)
// ============================================================

func TestSet_BasicOperations(t *testing.T) {
	s := collection.NewSet(1, 2, 3, 4, 5)

	if s.Len() != 5 {
		t.Errorf("초기 크기: 기대 5, 실제 %d", s.Len())
	}

	// Contains
	if !s.Contains(3) {
		t.Error("Contains(3)은 true여야 합니다")
	}
	if s.Contains(99) {
		t.Error("Contains(99)는 false여야 합니다")
	}

	// Add 중복
	s.Add(3) // 이미 있음
	if s.Len() != 5 {
		t.Error("중복 Add는 크기를 변경하면 안 됩니다")
	}

	// Add 새 원소
	s.Add(6)
	if s.Len() != 6 {
		t.Errorf("Add 후 크기: 기대 6, 실제 %d", s.Len())
	}

	// Remove
	s.Remove(1)
	if s.Contains(1) {
		t.Error("Remove 후 Contains(1)은 false여야 합니다")
	}
	if s.Len() != 5 {
		t.Errorf("Remove 후 크기: 기대 5, 실제 %d", s.Len())
	}
}

func TestSet_SetOperations(t *testing.T) {
	a := collection.NewSet(1, 2, 3, 4, 5)
	b := collection.NewSet(3, 4, 5, 6, 7)

	// Union
	u := a.Union(b)
	if u.Len() != 7 {
		t.Errorf("합집합 크기: 기대 7, 실제 %d", u.Len())
	}
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7} {
		if !u.Contains(v) {
			t.Errorf("합집합에 %d가 없습니다", v)
		}
	}

	// Intersection
	i := a.Intersection(b)
	if i.Len() != 3 {
		t.Errorf("교집합 크기: 기대 3, 실제 %d", i.Len())
	}
	for _, v := range []int{3, 4, 5} {
		if !i.Contains(v) {
			t.Errorf("교집합에 %d가 없습니다", v)
		}
	}

	// Difference
	d := a.Difference(b)
	if d.Len() != 2 {
		t.Errorf("차집합 크기: 기대 2, 실제 %d", d.Len())
	}
	for _, v := range []int{1, 2} {
		if !d.Contains(v) {
			t.Errorf("차집합에 %d가 없습니다", v)
		}
	}
}

// ============================================================
// OrderedMap 테스트 (25점)
// ============================================================

func TestOrderedMap_BasicOperations(t *testing.T) {
	m := collection.NewOrderedMap[string, int]()

	if m.Len() != 0 {
		t.Errorf("초기 크기: 기대 0, 실제 %d", m.Len())
	}

	// Set
	m.Set("banana", 2)
	m.Set("apple", 1)
	m.Set("cherry", 3)

	if m.Len() != 3 {
		t.Errorf("Set 3회 후 크기: 기대 3, 실제 %d", m.Len())
	}

	// Get
	v, ok := m.Get("apple")
	if !ok || v != 1 {
		t.Errorf("Get(apple): 기대 (1, true), 실제 (%d, %v)", v, ok)
	}

	_, ok = m.Get("mango")
	if ok {
		t.Error("없는 키 Get 시 ok=false여야 합니다")
	}

	// 값 업데이트
	m.Set("apple", 99)
	v, _ = m.Get("apple")
	if v != 99 {
		t.Errorf("업데이트 후 Get(apple): 기대 99, 실제 %d", v)
	}
}

func TestOrderedMap_InsertionOrder(t *testing.T) {
	m := collection.NewOrderedMap[string, int]()
	m.Set("banana", 2)
	m.Set("apple", 1)
	m.Set("cherry", 3)

	// Keys는 삽입 순서여야 합니다
	keys := m.Keys()
	expectedKeys := []string{"banana", "apple", "cherry"}
	if len(keys) != len(expectedKeys) {
		t.Fatalf("Keys 길이: 기대 %d, 실제 %d", len(expectedKeys), len(keys))
	}
	for i, want := range expectedKeys {
		if keys[i] != want {
			t.Errorf("Keys[%d]: 기대 %q, 실제 %q", i, want, keys[i])
		}
	}

	// Values도 삽입 순서여야 합니다
	values := m.Values()
	expectedValues := []int{2, 1, 3}
	for i, want := range expectedValues {
		if values[i] != want {
			t.Errorf("Values[%d]: 기대 %d, 실제 %d", i, want, values[i])
		}
	}
}

func TestOrderedMap_Delete(t *testing.T) {
	m := collection.NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	m.Delete("b")

	if m.Len() != 2 {
		t.Errorf("Delete 후 크기: 기대 2, 실제 %d", m.Len())
	}

	_, ok := m.Get("b")
	if ok {
		t.Error("삭제 후 Get(b)는 ok=false여야 합니다")
	}

	// 순서 유지 확인
	keys := m.Keys()
	if len(keys) != 2 || keys[0] != "a" || keys[1] != "c" {
		t.Errorf("삭제 후 순서: 기대 [a c], 실제 %v", keys)
	}
}

// ============================================================
// 성적 보고서
// ============================================================

func TestMain(m *testing.M) {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║   과제 A1: 제네릭 컬렉션 라이브러리   ║")
	fmt.Println("╚══════════════════════════════════════╝")

	result := m.Run()

	fmt.Println()
	fmt.Println("─────────────────────────────────────")
	if result == 0 {
		fmt.Println("  최종 점수: 100 / 100 점")
		fmt.Println("  평가: 합격 (모든 테스트 통과)")
	} else {
		fmt.Println("  평가: 미완성 — 실패한 테스트를 확인하세요")
	}
	fmt.Println("─────────────────────────────────────")

	os.Exit(result)
}
