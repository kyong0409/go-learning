// sliceutil 패키지: 슬라이스 유틸리티 함수 구현 과제
//
// 각 함수의 본문을 구현하세요.
// 테스트 실행: go test -v
package sliceutil

import "errors"

// ErrEmptySlice: 빈 슬라이스에 대한 연산 에러
var ErrEmptySlice = errors.New("슬라이스가 비어있습니다")

// Filter: pred(n)이 true인 요소만 모아 새 슬라이스를 반환합니다.
// 원본 슬라이스를 수정하지 않습니다.
// 예: Filter([]int{1,2,3,4,5}, func(n int) bool { return n%2==0 }) == []int{2,4}
func Filter(nums []int, pred func(int) bool) []int {
	// TODO: 구현하세요
	return nil
}

// Map: 모든 요소에 f를 적용한 새 슬라이스를 반환합니다.
// 원본 슬라이스를 수정하지 않습니다.
// 예: Map([]int{1,2,3}, func(n int) int { return n*n }) == []int{1,4,9}
func Map(nums []int, f func(int) int) []int {
	// TODO: 구현하세요
	return nil
}

// Reduce: initial부터 시작해 f(누적값, 요소)를 반복 적용해 단일 값을 반환합니다.
// 빈 슬라이스이면 initial을 반환합니다.
// 예: Reduce([]int{1,2,3,4,5}, 0, func(acc,n int) int { return acc+n }) == 15
func Reduce(nums []int, initial int, f func(int, int) int) int {
	// TODO: 구현하세요
	return initial
}

// Contains: nums에 target이 포함되어 있으면 true를 반환합니다.
// 예: Contains([]int{1,2,3}, 2) == true
// 예: Contains([]int{1,2,3}, 5) == false
func Contains(nums []int, target int) bool {
	// TODO: 구현하세요
	return false
}

// Unique: 중복을 제거한 새 슬라이스를 반환합니다. 첫 등장 순서를 유지합니다.
// 예: Unique([]int{3,1,2,1,3,4}) == []int{3,1,2,4}
func Unique(nums []int) []int {
	// TODO: 구현하세요
	// 힌트: map[int]bool을 사용하면 이미 등장한 값을 추적할 수 있습니다.
	return nil
}

// Sum: 모든 요소의 합을 반환합니다. 빈 슬라이스면 0을 반환합니다.
// 예: Sum([]int{1,2,3,4,5}) == 15
func Sum(nums []int) int {
	// TODO: 구현하세요
	return 0
}

// Max: 최댓값을 반환합니다. 빈 슬라이스면 (0, ErrEmptySlice)를 반환합니다.
// 예: Max([]int{3,1,4,1,5,9,2,6}) == (9, nil)
func Max(nums []int) (int, error) {
	// TODO: 구현하세요
	return 0, nil
}

// Min: 최솟값을 반환합니다. 빈 슬라이스면 (0, ErrEmptySlice)를 반환합니다.
// 예: Min([]int{3,1,4,1,5,9,2,6}) == (1, nil)
func Min(nums []int) (int, error) {
	// TODO: 구현하세요
	return 0, nil
}

// Reverse: 순서를 뒤집은 새 슬라이스를 반환합니다. 원본을 수정하지 않습니다.
// 예: Reverse([]int{1,2,3,4,5}) == []int{5,4,3,2,1}
func Reverse(nums []int) []int {
	// TODO: 구현하세요
	return nil
}

// Flatten: 2차원 슬라이스를 1차원으로 평탄화합니다.
// 예: Flatten([][]int{{1,2},{3,4},{5}}) == []int{1,2,3,4,5}
func Flatten(nested [][]int) []int {
	// TODO: 구현하세요
	return nil
}
