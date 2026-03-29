// 05-iterators/main.go
// Go 이터레이터 패턴을 학습합니다.
//
// NOTE: Go 1.23에서 range-over-function 이터레이터가 정식 도입되었습니다.
//   - iter.Seq[V]  = func(yield func(V) bool)
//   - iter.Seq2[K,V] = func(yield func(K, V) bool)
//   - for v := range someSeq { ... } 문법 지원
//   - slices.All, slices.Values, slices.Backward, slices.Collect
//   - maps.All, maps.Keys, maps.Values
//
// 이 파일은 Go 1.22에서 실행 가능하도록 동일한 패턴을 수동으로 구현합니다.
// Go 1.23 문법과의 대응 관계를 주석으로 표시합니다.
package main

import (
	"fmt"
	"slices"
	"sort"
)

// ============================================================
// iter.Seq 시뮬레이션 (Go 1.23에서는 내장 타입)
//
// Go 1.23:  type iter.Seq[V any] = func(yield func(V) bool)
// Go 1.22:  아래처럼 직접 타입 정의
// ============================================================

// Seq는 Go 1.23의 iter.Seq[V]에 해당합니다.
type Seq[V any] func(yield func(V) bool)

// Seq2는 Go 1.23의 iter.Seq2[K,V]에 해당합니다.
type Seq2[K, V any] func(yield func(K, V) bool)

// Collect는 Seq를 슬라이스로 수집합니다.
// Go 1.23: slices.Collect(seq)
func Collect[V any](seq Seq[V]) []V {
	var result []V
	seq(func(v V) bool {
		result = append(result, v)
		return true
	})
	return result
}

// ============================================================
// 1. 기본 이터레이터: Fibonacci
// Go 1.23에서는 for n := range Fibonacci() { ... } 로 사용
// ============================================================

// Fibonacci는 피보나치 수열을 생성하는 이터레이터입니다.
func Fibonacci() Seq[int] {
	return func(yield func(int) bool) {
		a, b := 0, 1
		for {
			if !yield(a) {
				return // yield가 false를 반환하면 (break 등) 중단
			}
			a, b = b, a+b
		}
	}
}

// Range는 [start, end) 범위의 정수 이터레이터입니다.
func Range(start, end int) Seq[int] {
	return func(yield func(int) bool) {
		for i := start; i < end; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

// ============================================================
// 2. 이터레이터 변환 함수 (고차 함수)
// ============================================================

// Filter는 조건을 만족하는 요소만 통과시킵니다.
func Filter[V any](seq Seq[V], predicate func(V) bool) Seq[V] {
	return func(yield func(V) bool) {
		seq(func(v V) bool {
			if predicate(v) {
				return yield(v)
			}
			return true
		})
	}
}

// MapSeq는 각 요소에 함수를 적용합니다.
func MapSeq[V, U any](seq Seq[V], f func(V) U) Seq[U] {
	return func(yield func(U) bool) {
		seq(func(v V) bool {
			return yield(f(v))
		})
	}
}

// Take는 처음 n개의 요소만 반환합니다.
func Take[V any](seq Seq[V], n int) Seq[V] {
	return func(yield func(V) bool) {
		count := 0
		seq(func(v V) bool {
			if count >= n {
				return false
			}
			count++
			return yield(v)
		})
	}
}

// ============================================================
// 3. iter.Seq2 — 키-값 쌍 이터레이터
// ============================================================

// Enumerate는 슬라이스에 인덱스를 붙인 Seq2를 반환합니다.
// Go 1.23: slices.All(s)
func Enumerate[V any](s []V) Seq2[int, V] {
	return func(yield func(int, V) bool) {
		for i, v := range s {
			if !yield(i, v) {
				return
			}
		}
	}
}

// SlicesAll은 슬라이스의 모든 인덱스-값 쌍을 반환합니다.
// Go 1.23: slices.All(s)
func SlicesAll[V any](s []V) Seq2[int, V] {
	return Enumerate(s)
}

// SlicesValues는 슬라이스의 값만 반환합니다.
// Go 1.23: slices.Values(s)
func SlicesValues[V any](s []V) Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// SlicesBackward는 슬라이스를 역순으로 순회합니다.
// Go 1.23: slices.Backward(s)
func SlicesBackward[V any](s []V) Seq2[int, V] {
	return func(yield func(int, V) bool) {
		for i := len(s) - 1; i >= 0; i-- {
			if !yield(i, s[i]) {
				return
			}
		}
	}
}

// Zip은 두 슬라이스를 짝지어 Seq2를 반환합니다.
func Zip[A, B any](as []A, bs []B) Seq2[A, B] {
	return func(yield func(A, B) bool) {
		n := len(as)
		if len(bs) < n {
			n = len(bs)
		}
		for i := range n {
			if !yield(as[i], bs[i]) {
				return
			}
		}
	}
}

// ============================================================
// 4. 트리 순회 이터레이터
// ============================================================

// TreeNode는 이진 트리 노드입니다.
type TreeNode struct {
	Value       int
	Left, Right *TreeNode
}

// NewTree는 샘플 이진 탐색 트리를 생성합니다.
func NewTree() *TreeNode {
	return &TreeNode{
		Value: 5,
		Left: &TreeNode{
			Value: 3,
			Left:  &TreeNode{Value: 1},
			Right: &TreeNode{Value: 4},
		},
		Right: &TreeNode{
			Value: 8,
			Left:  &TreeNode{Value: 6},
			Right: &TreeNode{Value: 9},
		},
	}
}

// InOrder는 중위 순회 이터레이터를 반환합니다.
func (n *TreeNode) InOrder() Seq[int] {
	return func(yield func(int) bool) {
		var traverse func(*TreeNode) bool
		traverse = func(node *TreeNode) bool {
			if node == nil {
				return true
			}
			if !traverse(node.Left) {
				return false
			}
			if !yield(node.Value) {
				return false
			}
			return traverse(node.Right)
		}
		traverse(n)
	}
}

// ============================================================
// 5. slices 패키지 이터레이터 함수 (Go 1.22 호환)
// ============================================================

func demonstrateSlicesIterators() {
	fmt.Println("\n=== 슬라이스 이터레이터 패턴 ===")
	// Go 1.23: slices.All, slices.Values, slices.Backward, slices.Collect

	nums := []int{10, 20, 30, 40, 50}

	// slices.All (인덱스 + 값)
	fmt.Println("SlicesAll (인덱스, 값):")
	SlicesAll(nums)(func(i int, v int) bool {
		fmt.Printf("  [%d] = %d\n", i, v)
		return true
	})

	// slices.Values (값만)
	fmt.Print("SlicesValues: ")
	SlicesValues(nums)(func(v int) bool {
		fmt.Printf("%d ", v)
		return true
	})
	fmt.Println()

	// slices.Backward (역순)
	fmt.Println("SlicesBackward:")
	SlicesBackward(nums)(func(i int, v int) bool {
		fmt.Printf("  [%d] = %d\n", i, v)
		return true
	})

	// slices.Collect + MapSeq
	doubled := Collect(MapSeq(SlicesValues(nums), func(n int) int { return n * 2 }))
	fmt.Printf("두 배 수집: %v\n", doubled)

	// slices.Sorted (정렬 수집)
	unordered := []int{5, 2, 8, 1, 9, 3}
	sortedVals := Collect(SlicesValues(unordered))
	slices.Sort(sortedVals)
	fmt.Printf("정렬 수집: %v\n", sortedVals)
}

// ============================================================
// 6. maps 패키지 이터레이터 패턴
// ============================================================

func demonstrateMapsIterators() {
	fmt.Println("\n=== 맵 이터레이터 패턴 ===")
	// Go 1.23: maps.All, maps.Keys, maps.Values

	m := map[string]int{"사과": 3, "바나나": 1, "체리": 5}

	// maps.Keys 시뮬레이션
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("맵 키 (정렬): %v\n", keys)

	// maps.Values 시뮬레이션 — 합계 계산
	total := 0
	for _, v := range m {
		total += v
	}
	fmt.Printf("맵 값 합계: %d\n", total)

	// maps.All 시뮬레이션 — 키-값 순회
	fmt.Println("맵 순회 (정렬된 키 순):")
	for _, k := range keys {
		fmt.Printf("  %s: %d\n", k, m[k])
	}
}

// ============================================================
// 7. 이터레이터 체이닝
// ============================================================

func demonstrateChaining() {
	fmt.Println("\n=== 이터레이터 체이닝 ===")

	// 피보나치에서 짝수만 골라 처음 5개 수집
	// Go 1.23: slices.Collect(Take(Filter(Fibonacci(), even), 5))
	evenFibs := Collect(
		Take(
			Filter(Fibonacci(), func(n int) bool { return n%2 == 0 }),
			5,
		),
	)
	fmt.Printf("짝수 피보나치 (처음 5개): %v\n", evenFibs)

	// 1-20 범위에서 3의 배수를 제곱하여 수집
	result := Collect(
		MapSeq(
			Filter(Range(1, 21), func(n int) bool { return n%3 == 0 }),
			func(n int) int { return n * n },
		),
	)
	fmt.Printf("3의 배수 제곱 (1-20): %v\n", result)

	// 트리 중위 순회 후 짝수만 수집
	tree := NewTree()
	evenNodes := Collect(
		Filter(tree.InOrder(), func(n int) bool { return n%2 == 0 }),
	)
	fmt.Printf("트리 짝수 노드 (중위 순회): %v\n", evenNodes)
}

// ============================================================
// 메인 함수
// ============================================================

func main() {
	fmt.Println("=== Go 이터레이터 패턴 학습 ===")
	fmt.Println("(이 파일은 Go 1.22 호환 구현입니다)")
	fmt.Println("Go 1.23에서는 iter.Seq, range-over-function을 기본 제공합니다.")

	// 1. Fibonacci 이터레이터
	fmt.Println("\n--- 1. Fibonacci 이터레이터 ---")
	fmt.Print("피보나치 수열 (처음 10개): ")
	count := 0
	Fibonacci()(func(n int) bool {
		if count >= 10 {
			return false
		}
		fmt.Printf("%d ", n)
		count++
		return true
	})
	fmt.Println()

	// 2. Range 이터레이터
	fmt.Println("\n--- 2. Range 이터레이터 ---")
	fmt.Print("1부터 5까지: ")
	Range(1, 6)(func(n int) bool {
		fmt.Printf("%d ", n)
		return true
	})
	fmt.Println()

	// 3. Seq2 — Enumerate와 Zip
	fmt.Println("\n--- 3. Seq2 이터레이터 ---")
	fruits := []string{"사과", "바나나", "체리"}
	fmt.Println("Enumerate:")
	Enumerate(fruits)(func(i int, fruit string) bool {
		fmt.Printf("  %d: %s\n", i, fruit)
		return true
	})

	prices := []int{1200, 800, 2000}
	fmt.Println("Zip (과일, 가격):")
	Zip(fruits, prices)(func(fruit string, price int) bool {
		fmt.Printf("  %s: %d원\n", fruit, price)
		return true
	})

	// 4. 트리 순회
	fmt.Println("\n--- 4. 트리 중위 순회 ---")
	tree := NewTree()
	fmt.Print("중위 순회 (정렬된 결과): ")
	tree.InOrder()(func(v int) bool {
		fmt.Printf("%d ", v)
		return true
	})
	fmt.Println()

	// 5. 슬라이스 이터레이터 패턴
	demonstrateSlicesIterators()

	// 6. 맵 이터레이터 패턴
	demonstrateMapsIterators()

	// 7. 체이닝
	demonstrateChaining()

	fmt.Println()
	fmt.Println("=== Go 1.23 iter 패키지 참고 ===")
	fmt.Println("Go 1.23+에서는 다음과 같이 사용합니다:")
	fmt.Println("  for n := range Fibonacci() { ... }  // range-over-function")
	fmt.Println("  for i, v := range slices.All(s) { ... }")
	fmt.Println("  result := slices.Collect(Filter(seq, pred))")

	fmt.Println("\n=== 이터레이터 학습 완료 ===")
}
