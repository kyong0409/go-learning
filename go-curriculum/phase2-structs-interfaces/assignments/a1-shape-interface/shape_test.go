// shape_test.go: Shape 인터페이스 과제 채점 테스트
package main

import (
	"fmt"
	"math"
	"testing"
)

// ─────────────────────────────────────────
// 채점 시스템
// ─────────────────────────────────────────

// scorer는 채점 결과를 추적합니다.
type scorer struct {
	total  int
	passed int
}

func newScorer() *scorer { return &scorer{} }

func (s *scorer) check(t *testing.T, name string, got, want float64, tolerance float64) {
	t.Helper()
	s.total++
	if math.Abs(got-want) <= tolerance {
		s.passed++
	} else {
		t.Errorf("  FAIL [%s]: got=%.6f, want=%.6f", name, got, want)
	}
}

func (s *scorer) checkBool(t *testing.T, name string, cond bool) {
	t.Helper()
	s.total++
	if cond {
		s.passed++
	} else {
		t.Errorf("  FAIL [%s]: 조건이 충족되지 않았습니다", name)
	}
}

func (s *scorer) report(t *testing.T) {
	t.Helper()
	score := s.passed * 100 / s.total
	fmt.Printf("\n=== 채점 결과 ===\n")
	fmt.Printf("통과: %d/%d\n", s.passed, s.total)
	fmt.Printf("점수: %d/100\n", score)
	if s.passed == s.total {
		fmt.Println("완벽합니다! 모든 테스트를 통과했습니다.")
	}
}

const eps = 1e-6

// ─────────────────────────────────────────
// Circle 테스트
// ─────────────────────────────────────────

func TestCircleArea(t *testing.T) {
	tests := []struct {
		radius float64
		want   float64
	}{
		{1, math.Pi},
		{2, 4 * math.Pi},
		{3, 9 * math.Pi},
		{0.5, 0.25 * math.Pi},
		{10, 100 * math.Pi},
	}
	sc := newScorer()
	for _, tt := range tests {
		c := Circle{Radius: tt.radius}
		sc.check(t, fmt.Sprintf("Circle(r=%.1f).Area()", tt.radius), c.Area(), tt.want, eps)
	}
	sc.report(t)
}

func TestCirclePerimeter(t *testing.T) {
	tests := []struct {
		radius float64
		want   float64
	}{
		{1, 2 * math.Pi},
		{2, 4 * math.Pi},
		{5, 10 * math.Pi},
		{0.5, math.Pi},
	}
	sc := newScorer()
	for _, tt := range tests {
		c := Circle{Radius: tt.radius}
		sc.check(t, fmt.Sprintf("Circle(r=%.1f).Perimeter()", tt.radius), c.Perimeter(), tt.want, eps)
	}
	sc.report(t)
}

// ─────────────────────────────────────────
// Rectangle 테스트
// ─────────────────────────────────────────

func TestRectangleArea(t *testing.T) {
	tests := []struct {
		w, h float64
		want float64
	}{
		{3, 4, 12},
		{5, 5, 25},
		{10, 2, 20},
		{1.5, 2.0, 3.0},
		{0, 5, 0},
	}
	sc := newScorer()
	for _, tt := range tests {
		r := Rectangle{Width: tt.w, Height: tt.h}
		sc.check(t, fmt.Sprintf("Rect(%.1fx%.1f).Area()", tt.w, tt.h), r.Area(), tt.want, eps)
	}
	sc.report(t)
}

func TestRectanglePerimeter(t *testing.T) {
	tests := []struct {
		w, h float64
		want float64
	}{
		{3, 4, 14},
		{5, 5, 20},
		{10, 2, 24},
		{1, 1, 4},
	}
	sc := newScorer()
	for _, tt := range tests {
		r := Rectangle{Width: tt.w, Height: tt.h}
		sc.check(t, fmt.Sprintf("Rect(%.1fx%.1f).Perimeter()", tt.w, tt.h), r.Perimeter(), tt.want, eps)
	}
	sc.report(t)
}

// ─────────────────────────────────────────
// Triangle 테스트
// ─────────────────────────────────────────

func TestTriangleArea(t *testing.T) {
	tests := []struct {
		a, b, c float64
		want    float64
	}{
		{3, 4, 5, 6},                        // 직각삼각형
		{5, 5, 5, math.Sqrt(3) / 4 * 25},    // 정삼각형
		{6, 8, 10, 24},                       // 직각삼각형
		{1, 1, 1, math.Sqrt(3) / 4},          // 단위 정삼각형
	}
	sc := newScorer()
	for _, tt := range tests {
		tri := Triangle{A: tt.a, B: tt.b, C: tt.c}
		sc.check(t, fmt.Sprintf("Triangle(%.0f,%.0f,%.0f).Area()", tt.a, tt.b, tt.c), tri.Area(), tt.want, eps)
	}
	sc.report(t)
}

func TestTrianglePerimeter(t *testing.T) {
	tests := []struct {
		a, b, c float64
		want    float64
	}{
		{3, 4, 5, 12},
		{5, 5, 5, 15},
		{1, 2, 3, 6},
	}
	sc := newScorer()
	for _, tt := range tests {
		tri := Triangle{A: tt.a, B: tt.b, C: tt.c}
		sc.check(t, fmt.Sprintf("Triangle(%.0f,%.0f,%.0f).Perimeter()", tt.a, tt.b, tt.c), tri.Perimeter(), tt.want, eps)
	}
	sc.report(t)
}

// ─────────────────────────────────────────
// 인터페이스 다형성 테스트
// ─────────────────────────────────────────

func TestShapeInterface(t *testing.T) {
	sc := newScorer()

	// 모든 타입이 Shape 인터페이스를 만족하는지 컴파일 타임 검사
	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 4, Height: 6},
		Triangle{A: 3, B: 4, C: 5},
	}

	sc.checkBool(t, "Shape 슬라이스 길이", len(shapes) == 3)

	// TotalArea 검사
	wantTotal := Circle{Radius: 5}.Area() +
		Rectangle{Width: 4, Height: 6}.Area() +
		Triangle{A: 3, B: 4, C: 5}.Area()
	sc.check(t, "TotalArea", TotalArea(shapes), wantTotal, eps)

	// LargestShape 검사 (Circle r=5: ~78.5, Rect 4x6: 24, Tri 3-4-5: 6)
	largest := LargestShape(shapes)
	sc.checkBool(t, "LargestShape not nil", largest != nil)
	if largest != nil {
		_, isCircle := largest.(Circle)
		sc.checkBool(t, "LargestShape is Circle", isCircle)
	} else {
		sc.total++ // 위 if 블록 건너뜀
	}

	// LargestShape 빈 슬라이스
	sc.checkBool(t, "LargestShape(empty) == nil", LargestShape([]Shape{}) == nil)

	// FilterByMinArea 검사
	filtered := FilterByMinArea(shapes, 20.0)
	// Circle(~78.5)과 Rectangle(24)이 통과, Triangle(6)은 제외
	sc.checkBool(t, "FilterByMinArea count", len(filtered) == 2)

	filtered0 := FilterByMinArea(shapes, 100.0)
	sc.checkBool(t, "FilterByMinArea none", len(filtered0) == 0)

	sc.report(t)
}

// ─────────────────────────────────────────
// 종합 채점 테스트
// ─────────────────────────────────────────

func TestFinalScore(t *testing.T) {
	sc := newScorer()

	// Circle 5개
	circleTests := []struct{ r, wantArea, wantPerim float64 }{
		{1, math.Pi, 2 * math.Pi},
		{2, 4 * math.Pi, 4 * math.Pi},
		{3, 9 * math.Pi, 6 * math.Pi},
		{5, 25 * math.Pi, 10 * math.Pi},
		{7, 49 * math.Pi, 14 * math.Pi},
	}
	for _, tt := range circleTests {
		c := Circle{Radius: tt.r}
		sc.check(t, fmt.Sprintf("Circle(%.0f) Area", tt.r), c.Area(), tt.wantArea, eps)
		sc.check(t, fmt.Sprintf("Circle(%.0f) Perim", tt.r), c.Perimeter(), tt.wantPerim, eps)
	}

	// Rectangle 5개
	rectTests := []struct{ w, h, wantArea, wantPerim float64 }{
		{3, 4, 12, 14},
		{5, 5, 25, 20},
		{10, 2, 20, 24},
		{1, 1, 1, 4},
		{6, 7, 42, 26},
	}
	for _, tt := range rectTests {
		r := Rectangle{Width: tt.w, Height: tt.h}
		sc.check(t, fmt.Sprintf("Rect(%.0fx%.0f) Area", tt.w, tt.h), r.Area(), tt.wantArea, eps)
		sc.check(t, fmt.Sprintf("Rect(%.0fx%.0f) Perim", tt.w, tt.h), r.Perimeter(), tt.wantPerim, eps)
	}

	// Triangle 3개
	triTests := []struct {
		a, b, c  float64
		wantArea float64
	}{
		{3, 4, 5, 6},
		{5, 12, 13, 30},
		{6, 8, 10, 24},
	}
	for _, tt := range triTests {
		tri := Triangle{A: tt.a, B: tt.b, C: tt.c}
		sc.check(t, fmt.Sprintf("Tri(%.0f,%.0f,%.0f) Area", tt.a, tt.b, tt.c), tri.Area(), tt.wantArea, eps)
	}

	// 유틸리티 함수
	shapes := []Shape{
		Circle{Radius: 1},
		Rectangle{Width: 2, Height: 3},
		Triangle{A: 3, B: 4, C: 5},
	}
	wantTotal := math.Pi + 6 + 6
	sc.check(t, "TotalArea", TotalArea(shapes), wantTotal, eps)

	largest := LargestShape(shapes)
	sc.checkBool(t, "LargestShape", largest != nil)

	filtered := FilterByMinArea(shapes, 5.0)
	sc.checkBool(t, "FilterByMinArea", len(filtered) == 2)

	sc.report(t)
}
