// a1-generic-collection/solution/collection.go
// 제네릭 컬렉션 라이브러리 참고 답안입니다.
package collection

// ============================================================
// Stack[T] — LIFO 스택
// ============================================================

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int    { return len(s.items) }
func (s *Stack[T]) IsEmpty() bool { return len(s.items) == 0 }

// ============================================================
// Queue[T] — FIFO 큐
// ============================================================

type Queue[T any] struct {
	items []T
}

func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}
	front := q.items[0]
	q.items = q.items[1:]
	return front, true
}

func (q *Queue[T]) Front() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}
	return q.items[0], true
}

func (q *Queue[T]) Len() int      { return len(q.items) }
func (q *Queue[T]) IsEmpty() bool { return len(q.items) == 0 }

// ============================================================
// Set[T comparable] — 집합
// ============================================================

type Set[T comparable] struct {
	items map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{items: make(map[T]struct{})}
	for _, item := range items {
		s.items[item] = struct{}{}
	}
	return s
}

func (s *Set[T]) Add(item T)          { s.items[item] = struct{}{} }
func (s *Set[T]) Remove(item T)       { delete(s.items, item) }
func (s *Set[T]) Len() int            { return len(s.items) }
func (s *Set[T]) Contains(item T) bool {
	_, ok := s.items[item]
	return ok
}

func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		result.Add(item)
	}
	for item := range other.items {
		result.Add(item)
	}
	return result
}

func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// ============================================================
// OrderedMap[K comparable, V any] — 삽입 순서 유지 맵
// ============================================================

type OrderedMap[K comparable, V any] struct {
	m    map[K]V
	keys []K
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{m: make(map[K]V)}
}

func (m *OrderedMap[K, V]) Set(key K, value V) {
	if _, exists := m.m[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.m[key] = value
}

func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	v, ok := m.m[key]
	return v, ok
}

func (m *OrderedMap[K, V]) Delete(key K) {
	if _, exists := m.m[key]; !exists {
		return
	}
	delete(m.m, key)
	for i, k := range m.keys {
		if k == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func (m *OrderedMap[K, V]) Keys() []K {
	result := make([]K, len(m.keys))
	copy(result, m.keys)
	return result
}

func (m *OrderedMap[K, V]) Values() []V {
	result := make([]V, len(m.keys))
	for i, k := range m.keys {
		result[i] = m.m[k]
	}
	return result
}

func (m *OrderedMap[K, V]) Len() int { return len(m.m) }
