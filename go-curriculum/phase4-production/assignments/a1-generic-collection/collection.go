// a1-generic-collection/collection.go
// 제네릭 컬렉션 라이브러리 과제입니다.
// TODO 주석이 있는 모든 메서드를 구현하세요.
package collection

// ============================================================
// Stack[T] — LIFO 스택
// ============================================================

// Stack은 후입선출(LIFO) 자료구조입니다.
type Stack[T any] struct {
	// TODO: 내부 저장소를 추가하세요.
}

// Push는 스택의 맨 위에 요소를 추가합니다.
func (s *Stack[T]) Push(item T) {
	// TODO: 구현하세요.
}

// Pop은 스택의 맨 위 요소를 제거하고 반환합니다.
// 스택이 비어있으면 (zero value, false)를 반환합니다.
func (s *Stack[T]) Pop() (T, bool) {
	// TODO: 구현하세요.
	var zero T
	return zero, false
}

// Peek은 스택의 맨 위 요소를 제거하지 않고 반환합니다.
func (s *Stack[T]) Peek() (T, bool) {
	// TODO: 구현하세요.
	var zero T
	return zero, false
}

// Len은 스택의 원소 수를 반환합니다.
func (s *Stack[T]) Len() int {
	// TODO: 구현하세요.
	return 0
}

// IsEmpty는 스택이 비어있는지 확인합니다.
func (s *Stack[T]) IsEmpty() bool {
	// TODO: 구현하세요.
	return true
}

// ============================================================
// Queue[T] — FIFO 큐
// ============================================================

// Queue는 선입선출(FIFO) 자료구조입니다.
type Queue[T any] struct {
	// TODO: 내부 저장소를 추가하세요.
}

// Enqueue는 큐의 뒤에 요소를 추가합니다.
func (q *Queue[T]) Enqueue(item T) {
	// TODO: 구현하세요.
}

// Dequeue는 큐의 앞에서 요소를 제거하고 반환합니다.
func (q *Queue[T]) Dequeue() (T, bool) {
	// TODO: 구현하세요.
	var zero T
	return zero, false
}

// Front는 큐의 앞 요소를 제거하지 않고 반환합니다.
func (q *Queue[T]) Front() (T, bool) {
	// TODO: 구현하세요.
	var zero T
	return zero, false
}

// Len은 큐의 원소 수를 반환합니다.
func (q *Queue[T]) Len() int {
	// TODO: 구현하세요.
	return 0
}

// IsEmpty는 큐가 비어있는지 확인합니다.
func (q *Queue[T]) IsEmpty() bool {
	// TODO: 구현하세요.
	return true
}

// ============================================================
// Set[T comparable] — 집합
// ============================================================

// Set은 중복 없는 원소의 집합입니다.
type Set[T comparable] struct {
	// TODO: 내부 저장소를 추가하세요.
}

// NewSet은 주어진 원소들로 새 집합을 생성합니다.
func NewSet[T comparable](items ...T) *Set[T] {
	// TODO: 구현하세요.
	return &Set[T]{}
}

// Add는 집합에 원소를 추가합니다.
func (s *Set[T]) Add(item T) {
	// TODO: 구현하세요.
}

// Remove는 집합에서 원소를 제거합니다.
func (s *Set[T]) Remove(item T) {
	// TODO: 구현하세요.
}

// Contains는 집합에 원소가 있는지 확인합니다.
func (s *Set[T]) Contains(item T) bool {
	// TODO: 구현하세요.
	return false
}

// Len은 집합의 원소 수를 반환합니다.
func (s *Set[T]) Len() int {
	// TODO: 구현하세요.
	return 0
}

// Union은 두 집합의 합집합을 반환합니다.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	// TODO: 구현하세요.
	return NewSet[T]()
}

// Intersection은 두 집합의 교집합을 반환합니다.
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	// TODO: 구현하세요.
	return NewSet[T]()
}

// Difference는 s에 있고 other에 없는 원소의 집합을 반환합니다.
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	// TODO: 구현하세요.
	return NewSet[T]()
}

// ============================================================
// OrderedMap[K comparable, V any] — 삽입 순서 유지 맵
// ============================================================

// OrderedMap은 삽입 순서를 유지하는 맵입니다.
type OrderedMap[K comparable, V any] struct {
	// TODO: 내부 저장소를 추가하세요.
}

// NewOrderedMap은 새 OrderedMap을 생성합니다.
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	// TODO: 구현하세요.
	return &OrderedMap[K, V]{}
}

// Set은 키-값 쌍을 저장합니다. 키가 이미 존재하면 값을 업데이트합니다.
func (m *OrderedMap[K, V]) Set(key K, value V) {
	// TODO: 구현하세요.
}

// Get은 키에 해당하는 값을 반환합니다.
func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	// TODO: 구현하세요.
	var zero V
	return zero, false
}

// Delete는 키-값 쌍을 삭제합니다.
func (m *OrderedMap[K, V]) Delete(key K) {
	// TODO: 구현하세요.
}

// Keys는 삽입 순서대로 키 슬라이스를 반환합니다.
func (m *OrderedMap[K, V]) Keys() []K {
	// TODO: 구현하세요.
	return nil
}

// Values는 삽입 순서대로 값 슬라이스를 반환합니다.
func (m *OrderedMap[K, V]) Values() []V {
	// TODO: 구현하세요.
	return nil
}

// Len은 맵의 원소 수를 반환합니다.
func (m *OrderedMap[K, V]) Len() int {
	// TODO: 구현하세요.
	return 0
}
