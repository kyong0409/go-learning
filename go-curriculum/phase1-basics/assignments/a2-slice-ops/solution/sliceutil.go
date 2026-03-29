// solution/sliceutil.go: 슬라이스 유틸리티 과제 참고 풀이
//
// 이 파일은 참고용입니다. 먼저 스스로 구현해 보세요!
package sliceutil

import "errors"

// ErrEmptySlice: 빈 슬라이스에 대한 연산 에러
var ErrEmptySlice = errors.New("슬라이스가 비어있습니다")

// Filter: 조건에 맞는 요소만 반환
func Filter(nums []int, pred func(int) bool) []int {
	var result []int
	for _, n := range nums {
		if pred(n) {
			result = append(result, n)
		}
	}
	if result == nil {
		return []int{}
	}
	return result
}

// Map: 모든 요소에 변환 함수 적용
func Map(nums []int, f func(int) int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = f(n)
	}
	return result
}

// Reduce: 슬라이스를 단일 값으로 축약
func Reduce(nums []int, initial int, f func(int, int) int) int {
	acc := initial
	for _, n := range nums {
		acc = f(acc, n)
	}
	return acc
}

// Contains: 특정 값 포함 여부
func Contains(nums []int, target int) bool {
	for _, n := range nums {
		if n == target {
			return true
		}
	}
	return false
}

// Unique: 중복 제거 (첫 등장 순서 유지)
func Unique(nums []int) []int {
	seen := make(map[int]bool)
	var result []int
	for _, n := range nums {
		if !seen[n] {
			seen[n] = true
			result = append(result, n)
		}
	}
	if result == nil {
		return []int{}
	}
	return result
}

// Sum: 합산
func Sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Max: 최댓값
func Max(nums []int) (int, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	max := nums[0]
	for _, n := range nums[1:] {
		if n > max {
			max = n
		}
	}
	return max, nil
}

// Min: 최솟값
func Min(nums []int) (int, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	min := nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
	}
	return min, nil
}

// Reverse: 순서 뒤집기 (원본 불변)
func Reverse(nums []int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[len(nums)-1-i] = n
	}
	return result
}

// Flatten: 2차원 -> 1차원 평탄화
func Flatten(nested [][]int) []int {
	var result []int
	for _, inner := range nested {
		result = append(result, inner...)
	}
	if result == nil {
		return []int{}
	}
	return result
}
