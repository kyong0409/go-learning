// sliceutil_test.go: 슬라이스 유틸리티 함수 채점 테스트
//
// 실행: go test -v
package sliceutil

import (
	"errors"
	"reflect"
	"testing"
)

// ─────────────────────────────────────────
// 채점 시스템
// ─────────────────────────────────────────

type grader struct {
	passed    int
	total     int
	points    int
	maxPoints int
}

func (g *grader) check(t *testing.T, name string, pts int, fn func(t *testing.T)) {
	t.Helper()
	g.total++
	g.maxPoints += pts
	passed := true
	t.Run(name, func(t *testing.T) {
		fn(t)
		if t.Failed() {
			passed = false
		}
	})
	if passed {
		g.passed++
		g.points += pts
	}
}

func (g *grader) report(t *testing.T) {
	t.Helper()
	t.Logf("\n==================")
	t.Logf("=== GRADE REPORT ===")
	t.Logf("==================")
	t.Logf("Passed: %d/%d", g.passed, g.total)
	t.Logf("Score:  %d/%d", g.points, g.maxPoints)
	t.Logf("==================")
}

// ─────────────────────────────────────────
// 테스트 함수
// ─────────────────────────────────────────

func TestSliceUtil(t *testing.T) {
	g := &grader{}

	// ─── Filter (10점) ───
	g.check(t, "Filter/짝수추출", 4, func(t *testing.T) {
		got := Filter([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
		want := []int{2, 4}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Filter 짝수: got %v, want %v", got, want)
		}
	})
	g.check(t, "Filter/모두제외", 3, func(t *testing.T) {
		got := Filter([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 })
		if len(got) != 0 {
			t.Errorf("Filter 모두제외: got %v, want []", got)
		}
	})
	g.check(t, "Filter/빈슬라이스", 3, func(t *testing.T) {
		got := Filter([]int{}, func(n int) bool { return true })
		if len(got) != 0 {
			t.Errorf("Filter 빈슬라이스: got %v, want []", got)
		}
	})

	// ─── Map (10점) ───
	g.check(t, "Map/제곱", 4, func(t *testing.T) {
		got := Map([]int{1, 2, 3, 4}, func(n int) int { return n * n })
		want := []int{1, 4, 9, 16}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Map 제곱: got %v, want %v", got, want)
		}
	})
	g.check(t, "Map/두배", 3, func(t *testing.T) {
		got := Map([]int{1, 2, 3}, func(n int) int { return n * 2 })
		want := []int{2, 4, 6}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Map 두배: got %v, want %v", got, want)
		}
	})
	g.check(t, "Map/빈슬라이스", 3, func(t *testing.T) {
		got := Map([]int{}, func(n int) int { return n })
		if len(got) != 0 {
			t.Errorf("Map 빈슬라이스: got %v, want []", got)
		}
	})

	// ─── Reduce (10점) ───
	g.check(t, "Reduce/합산", 4, func(t *testing.T) {
		got := Reduce([]int{1, 2, 3, 4, 5}, 0, func(acc, n int) int { return acc + n })
		if got != 15 {
			t.Errorf("Reduce 합산: got %d, want 15", got)
		}
	})
	g.check(t, "Reduce/곱셈", 3, func(t *testing.T) {
		got := Reduce([]int{1, 2, 3, 4}, 1, func(acc, n int) int { return acc * n })
		if got != 24 {
			t.Errorf("Reduce 곱셈: got %d, want 24", got)
		}
	})
	g.check(t, "Reduce/빈슬라이스", 3, func(t *testing.T) {
		got := Reduce([]int{}, 42, func(acc, n int) int { return acc + n })
		if got != 42 {
			t.Errorf("Reduce 빈슬라이스: got %d, want 42 (initial)", got)
		}
	})

	// ─── Contains (8점) ───
	g.check(t, "Contains/존재함", 4, func(t *testing.T) {
		if !Contains([]int{1, 2, 3, 4, 5}, 3) {
			t.Error("Contains([1,2,3,4,5], 3) = false; want true")
		}
	})
	g.check(t, "Contains/없음", 4, func(t *testing.T) {
		if Contains([]int{1, 2, 3}, 99) {
			t.Error("Contains([1,2,3], 99) = true; want false")
		}
	})

	// ─── Unique (12점) ───
	g.check(t, "Unique/중복제거", 5, func(t *testing.T) {
		got := Unique([]int{3, 1, 2, 1, 3, 4})
		want := []int{3, 1, 2, 4}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Unique: got %v, want %v", got, want)
		}
	})
	g.check(t, "Unique/중복없음", 4, func(t *testing.T) {
		got := Unique([]int{1, 2, 3})
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Unique 중복없음: got %v, want %v", got, want)
		}
	})
	g.check(t, "Unique/빈슬라이스", 3, func(t *testing.T) {
		got := Unique([]int{})
		if len(got) != 0 {
			t.Errorf("Unique 빈슬라이스: got %v, want []", got)
		}
	})

	// ─── Sum (6점) ───
	g.check(t, "Sum/기본", 4, func(t *testing.T) {
		got := Sum([]int{1, 2, 3, 4, 5})
		if got != 15 {
			t.Errorf("Sum: got %d, want 15", got)
		}
	})
	g.check(t, "Sum/빈슬라이스", 2, func(t *testing.T) {
		got := Sum([]int{})
		if got != 0 {
			t.Errorf("Sum 빈슬라이스: got %d, want 0", got)
		}
	})

	// ─── Max (8점) ───
	g.check(t, "Max/기본", 4, func(t *testing.T) {
		got, err := Max([]int{3, 1, 4, 1, 5, 9, 2, 6})
		if err != nil {
			t.Fatalf("Max 예상치 못한 에러: %v", err)
		}
		if got != 9 {
			t.Errorf("Max: got %d, want 9", got)
		}
	})
	g.check(t, "Max/빈슬라이스에러", 4, func(t *testing.T) {
		_, err := Max([]int{})
		if !errors.Is(err, ErrEmptySlice) {
			t.Errorf("Max 빈슬라이스: got %v, want ErrEmptySlice", err)
		}
	})

	// ─── Min (8점) ───
	g.check(t, "Min/기본", 4, func(t *testing.T) {
		got, err := Min([]int{3, 1, 4, 1, 5, 9, 2, 6})
		if err != nil {
			t.Fatalf("Min 예상치 못한 에러: %v", err)
		}
		if got != 1 {
			t.Errorf("Min: got %d, want 1", got)
		}
	})
	g.check(t, "Min/빈슬라이스에러", 4, func(t *testing.T) {
		_, err := Min([]int{})
		if !errors.Is(err, ErrEmptySlice) {
			t.Errorf("Min 빈슬라이스: got %v, want ErrEmptySlice", err)
		}
	})

	// ─── Reverse (10점) ───
	g.check(t, "Reverse/기본", 5, func(t *testing.T) {
		got := Reverse([]int{1, 2, 3, 4, 5})
		want := []int{5, 4, 3, 2, 1}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Reverse: got %v, want %v", got, want)
		}
	})
	g.check(t, "Reverse/원본불변", 5, func(t *testing.T) {
		original := []int{1, 2, 3}
		_ = Reverse(original)
		if !reflect.DeepEqual(original, []int{1, 2, 3}) {
			t.Errorf("Reverse가 원본을 수정했습니다: %v", original)
		}
	})

	// ─── Flatten (8점) ───
	g.check(t, "Flatten/기본", 5, func(t *testing.T) {
		got := Flatten([][]int{{1, 2}, {3, 4}, {5}})
		want := []int{1, 2, 3, 4, 5}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Flatten: got %v, want %v", got, want)
		}
	})
	g.check(t, "Flatten/빈내부슬라이스", 3, func(t *testing.T) {
		got := Flatten([][]int{{1, 2}, {}, {3}})
		want := []int{1, 2, 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Flatten 빈내부: got %v, want %v", got, want)
		}
	})

	// ─── 최종 채점 리포트 ───
	g.report(t)
}
