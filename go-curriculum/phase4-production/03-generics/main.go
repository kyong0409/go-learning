// 03-generics/main.go
// Go 제네릭(타입 파라미터)을 학습합니다.
// 타입 파라미터 기초, 제약조건, 제네릭 타입, 표준 라이브러리 패키지를 다룹니다.
package main

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"sync"
)

// ============================================================
// 1. 타입 파라미터 기초
// 함수에 타입 파라미터를 추가하면 여러 타입에서 동작하는 범용 함수를 만들 수 있습니다.
// ============================================================

// Map은 슬라이스의 각 요소에 함수를 적용하여 새 슬라이스를 반환합니다.
// T: 입력 타입, U: 출력 타입
func Map[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

// Filter는 조건을 만족하는 요소만 포함하는 새 슬라이스를 반환합니다.
func Filter[T any](s []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range s {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce는 슬라이스를 단일 값으로 축약합니다.
// T: 슬라이스 요소 타입, U: 결과 타입
func Reduce[T, U any](s []T, initial U, f func(U, T) U) U {
	result := initial
	for _, v := range s {
		result = f(result, v)
	}
	return result
}

// Contains는 슬라이스에 특정 값이 있는지 확인합니다.
// comparable 제약: == 연산자를 사용할 수 있어야 합니다.
func Contains[T comparable](s []T, target T) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

// Keys는 맵의 모든 키를 슬라이스로 반환합니다.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ============================================================
// 2. 제약조건 (Constraints)
// ============================================================

// Ordered는 순서 비교가 가능한 타입을 나타내는 제약조건입니다.
// cmp.Ordered를 사용하면 정수, 부동소수점, 문자열 등을 모두 지원합니다.
type Ordered interface {
	cmp.Ordered
}

// Min은 두 값 중 작은 값을 반환합니다.
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max는 두 값 중 큰 값을 반환합니다.
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clamp는 값을 [min, max] 범위로 제한합니다.
func Clamp[T cmp.Ordered](value, minVal, maxVal T) T {
	return Max(minVal, Min(value, maxVal))
}

// SortedKeys는 맵의 키를 정렬하여 반환합니다.
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := Keys(m)
	slices.Sort(keys)
	return keys
}

// ============================================================
// 3. 제네릭 타입: Stack[T]
// ============================================================

// Stack은 LIFO(후입선출) 스택 자료구조입니다.
type Stack[T any] struct {
	items []T
}

// Push는 스택에 요소를 추가합니다.
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop은 스택의 최상위 요소를 제거하고 반환합니다.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

// Peek은 최상위 요소를 제거하지 않고 반환합니다.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Len은 스택의 크기를 반환합니다.
func (s *Stack[T]) Len() int {
	return len(s.items)
}

// IsEmpty는 스택이 비어있는지 확인합니다.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// ============================================================
// 4. 제네릭 타입: Set[T]
// ============================================================

// Set은 중복 없는 집합 자료구조입니다.
// comparable 제약: 맵의 키로 사용하기 위해 필요합니다.
type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet은 새 집합을 생성합니다.
func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{items: make(map[T]struct{})}
	for _, item := range items {
		s.Add(item)
	}
	return s
}

// Add는 집합에 요소를 추가합니다.
func (s *Set[T]) Add(item T) {
	s.items[item] = struct{}{}
}

// Remove는 집합에서 요소를 제거합니다.
func (s *Set[T]) Remove(item T) {
	delete(s.items, item)
}

// Contains는 집합에 요소가 있는지 확인합니다.
func (s *Set[T]) Contains(item T) bool {
	_, ok := s.items[item]
	return ok
}

// Len은 집합의 크기를 반환합니다.
func (s *Set[T]) Len() int {
	return len(s.items)
}

// Union은 두 집합의 합집합을 반환합니다.
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

// Intersection은 두 집합의 교집합을 반환합니다.
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for item := range s.items {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// ============================================================
// 5. 실용적 예제: 제네릭 캐시 (동시성 안전)
// ============================================================

// Cache는 키-값 쌍을 저장하는 동시성 안전 제네릭 캐시입니다.
type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

// NewCache는 새 캐시를 생성합니다.
func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		items: make(map[K]V),
	}
}

// Set은 캐시에 값을 저장합니다.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

// Get은 캐시에서 값을 조회합니다.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[key]
	return v, ok
}

// GetOrSet은 값이 있으면 반환하고, 없으면 생성 함수를 호출하여 저장합니다.
func (c *Cache[K, V]) GetOrSet(key K, create func() V) V {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.items[key]; ok {
		return v
	}
	v := create()
	c.items[key] = v
	return v
}

// ============================================================
// 6. 제네릭 Result 타입
// 에러 처리를 타입 안전하게 만드는 패턴입니다.
// ============================================================

// Result는 성공값(T) 또는 오류를 담는 제네릭 타입입니다.
type Result[T any] struct {
	value T
	err   error
}

// Ok는 성공 Result를 생성합니다.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

// Err는 실패 Result를 생성합니다.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// IsOk는 Result가 성공인지 확인합니다.
func (r Result[T]) IsOk() bool {
	return r.err == nil
}

// Unwrap은 성공값을 반환합니다. 실패인 경우 패닉이 발생합니다.
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(fmt.Sprintf("Result.Unwrap 실패: %v", r.err))
	}
	return r.value
}

// UnwrapOr는 성공값을 반환하거나, 실패인 경우 기본값을 반환합니다.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}

// MapResult는 Result의 성공값에 함수를 적용합니다.
func MapResult[T, U any](r Result[T], f func(T) U) Result[U] {
	if r.err != nil {
		return Err[U](r.err)
	}
	return Ok(f(r.value))
}

// ============================================================
// 7. slices, maps, cmp 표준 라이브러리 사용 예제
// ============================================================

func demonstrateStdLib() {
	fmt.Println("\n=== 표준 라이브러리 제네릭 함수 ===")

	// slices 패키지
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}

	// slices.Sort: 슬라이스 정렬
	sorted := slices.Clone(nums)
	slices.Sort(sorted)
	fmt.Printf("정렬됨: %v\n", sorted)

	// slices.Contains: 요소 포함 여부
	fmt.Printf("5 포함 여부: %v\n", slices.Contains(nums, 5))

	// slices.Index: 요소의 인덱스 반환
	fmt.Printf("9의 인덱스: %d\n", slices.Index(nums, 9))

	// slices.Max / slices.Min
	fmt.Printf("최댓값: %d, 최솟값: %d\n", slices.Max(nums), slices.Min(nums))

	// slices.Reverse: 역순
	rev := slices.Clone(nums)
	slices.Reverse(rev)
	fmt.Printf("역순: %v\n", rev)

	// slices.Compact: 연속된 중복 제거 (정렬 후 사용)
	compact := slices.Compact(sorted)
	fmt.Printf("중복 제거: %v\n", compact)

	// maps 패키지
	m := map[string]int{"사과": 3, "바나나": 1, "체리": 2}

	// maps 패키지: Go 1.22에서는 maps.Clone, maps.Copy만 제공됩니다.
	// maps.Keys는 Go 1.23+에서 iter.Seq를 반환합니다.
	// Go 1.22에서는 직접 키를 추출합니다.
	keys := Keys(m) // 파일 상단에 정의한 제네릭 Keys 함수 사용
	slices.Sort(keys)
	fmt.Printf("맵 키 (정렬): %v\n", keys)

	// maps.Clone: 맵 복사 (Go 1.22에서 사용 가능)
	cloned := maps.Clone(m)
	cloned["딸기"] = 5
	fmt.Printf("원본 크기: %d, 복사본 크기: %d\n", len(m), len(cloned))

	// cmp 패키지
	fmt.Printf("\ncmp.Compare(1, 2): %d\n", cmp.Compare(1, 2))   // -1
	fmt.Printf("cmp.Compare(2, 2): %d\n", cmp.Compare(2, 2))     // 0
	fmt.Printf("cmp.Compare(3, 2): %d\n", cmp.Compare(3, 2))     // 1
	fmt.Printf("cmp.Or(0, 0, 42): %d\n", cmp.Or(0, 0, 42))       // 42 (첫 비제로값)

	// slices.SortFunc: 커스텀 비교 함수로 정렬
	type Person struct {
		Name string
		Age  int
	}
	people := []Person{
		{"김철수", 30},
		{"이영희", 25},
		{"박민준", 35},
	}
	slices.SortFunc(people, func(a, b Person) int {
		return cmp.Compare(a.Age, b.Age)
	})
	fmt.Printf("\n나이순 정렬: %v\n", people)
}

// ============================================================
// 메인 함수
// ============================================================

func main() {
	fmt.Println("=== Go 제네릭 학습 ===")

	// 1. 기본 제네릭 함수
	fmt.Println("\n--- 1. Map, Filter, Reduce ---")
	nums := []int{1, 2, 3, 4, 5}

	doubled := Map(nums, func(n int) int { return n * 2 })
	fmt.Printf("두 배: %v\n", doubled)

	squares := Map(nums, func(n int) string { return fmt.Sprintf("%d²=%d", n, n*n) })
	fmt.Printf("제곱: %v\n", squares)

	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Printf("짝수: %v\n", evens)

	sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Printf("합계: %d\n", sum)

	// 2. 제약조건
	fmt.Println("\n--- 2. Ordered 제약조건 ---")
	fmt.Printf("Min(3, 7) = %d\n", Min(3, 7))
	fmt.Printf("Max(3.14, 2.72) = %.2f\n", Max(3.14, 2.72))
	fmt.Printf("Min(\"apple\", \"banana\") = %s\n", Min("apple", "banana"))
	fmt.Printf("Clamp(15, 0, 10) = %d\n", Clamp(15, 0, 10))

	scores := map[string]int{"수학": 95, "영어": 82, "과학": 78}
	fmt.Printf("정렬된 키: %v\n", SortedKeys(scores))

	// 3. 스택
	fmt.Println("\n--- 3. Stack[T] ---")
	var intStack Stack[int]
	intStack.Push(1)
	intStack.Push(2)
	intStack.Push(3)
	fmt.Printf("스택 크기: %d\n", intStack.Len())

	if top, ok := intStack.Pop(); ok {
		fmt.Printf("Pop: %d\n", top)
	}
	if top, ok := intStack.Peek(); ok {
		fmt.Printf("Peek: %d (제거 안 됨)\n", top)
	}

	// 문자열 스택
	var strStack Stack[string]
	strStack.Push("Go")
	strStack.Push("제네릭")
	strStack.Push("최고!")
	for !strStack.IsEmpty() {
		if v, ok := strStack.Pop(); ok {
			fmt.Printf("  %s\n", v)
		}
	}

	// 4. 집합
	fmt.Println("\n--- 4. Set[T] ---")
	set1 := NewSet(1, 2, 3, 4, 5)
	set2 := NewSet(3, 4, 5, 6, 7)

	union := set1.Union(set2)
	fmt.Printf("합집합 크기: %d\n", union.Len())

	inter := set1.Intersection(set2)
	fmt.Printf("교집합 크기: %d\n", inter.Len())
	fmt.Printf("교집합에 3 포함: %v\n", inter.Contains(3))
	fmt.Printf("교집합에 1 포함: %v\n", inter.Contains(1))

	// 5. 캐시
	fmt.Println("\n--- 5. 제네릭 캐시 ---")
	cache := NewCache[string, int]()
	cache.Set("a", 1)
	cache.Set("b", 2)

	if v, ok := cache.Get("a"); ok {
		fmt.Printf("캐시 조회 'a': %d\n", v)
	}

	// GetOrSet: 없으면 생성
	v := cache.GetOrSet("c", func() int { return 42 })
	fmt.Printf("GetOrSet 'c': %d\n", v)

	// 6. Result 타입
	fmt.Println("\n--- 6. Result[T] ---")
	success := Ok(42)
	failure := Err[int](fmt.Errorf("계산 실패"))

	fmt.Printf("성공: %d\n", success.Unwrap())
	fmt.Printf("실패 기본값: %d\n", failure.UnwrapOr(-1))

	// MapResult로 변환
	doubled2 := MapResult(success, func(n int) string {
		return fmt.Sprintf("결과: %d", n*2)
	})
	fmt.Printf("변환된 결과: %s\n", doubled2.Unwrap())

	// 7. 표준 라이브러리
	demonstrateStdLib()

	fmt.Println("\n=== 학습 완료 ===")
}
