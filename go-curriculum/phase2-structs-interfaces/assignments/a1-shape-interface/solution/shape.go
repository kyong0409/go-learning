// Package main은 Shape 인터페이스 과제의 참고 풀이입니다.
package main

import (
	"fmt"
	"math"
)

// ─────────────────────────────────────────
// Shape 인터페이스
// ─────────────────────────────────────────

// Shape는 도형 인터페이스입니다.
type Shape interface {
	Area() float64
	Perimeter() float64
	String() string
}

// ─────────────────────────────────────────
// Circle (원)
// ─────────────────────────────────────────

// Circle은 원을 나타냅니다.
type Circle struct {
	Radius float64
}

// Area는 원의 넓이를 반환합니다: π × r²
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Perimeter는 원의 둘레를 반환합니다: 2 × π × r
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// String은 원의 문자열 표현을 반환합니다.
func (c Circle) String() string {
	return fmt.Sprintf("Circle(r=%.2f)", c.Radius)
}

// ─────────────────────────────────────────
// Rectangle (직사각형)
// ─────────────────────────────────────────

// Rectangle은 직사각형을 나타냅니다.
type Rectangle struct {
	Width  float64
	Height float64
}

// Area는 직사각형의 넓이를 반환합니다: width × height
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter는 직사각형의 둘레를 반환합니다: 2 × (width + height)
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// String은 직사각형의 문자열 표현을 반환합니다.
func (r Rectangle) String() string {
	return fmt.Sprintf("Rect(%.2fx%.2f)", r.Width, r.Height)
}

// ─────────────────────────────────────────
// Triangle (삼각형)
// ─────────────────────────────────────────

// Triangle은 삼각형을 나타냅니다 (세 변의 길이로 정의).
type Triangle struct {
	A, B, C float64
}

// Area는 헤론의 공식으로 삼각형의 넓이를 반환합니다.
func (t Triangle) Area() float64 {
	s := (t.A + t.B + t.C) / 2 // 반둘레
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

// Perimeter는 삼각형의 둘레를 반환합니다: A + B + C
func (t Triangle) Perimeter() float64 {
	return t.A + t.B + t.C
}

// String은 삼각형의 문자열 표현을 반환합니다.
func (t Triangle) String() string {
	return fmt.Sprintf("Triangle(%.2f,%.2f,%.2f)", t.A, t.B, t.C)
}

// ─────────────────────────────────────────
// 유틸리티 함수
// ─────────────────────────────────────────

// TotalArea는 도형 슬라이스의 총 넓이를 반환합니다.
func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// LargestShape는 가장 넓은 도형을 반환합니다.
// shapes가 비어있으면 nil을 반환합니다.
func LargestShape(shapes []Shape) Shape {
	if len(shapes) == 0 {
		return nil
	}
	largest := shapes[0]
	for _, s := range shapes[1:] {
		if s.Area() > largest.Area() {
			largest = s
		}
	}
	return largest
}

// FilterByMinArea는 최소 넓이 이상인 도형만 반환합니다.
func FilterByMinArea(shapes []Shape, minArea float64) []Shape {
	result := make([]Shape, 0)
	for _, s := range shapes {
		if s.Area() >= minArea {
			result = append(result, s)
		}
	}
	return result
}

func main() {
	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 4, Height: 6},
		Triangle{A: 3, B: 4, C: 5},
	}

	fmt.Println("=== Shape 인터페이스 참고 풀이 ===")
	for _, s := range shapes {
		fmt.Printf("%v → 넓이=%.4f, 둘레=%.4f\n", s, s.Area(), s.Perimeter())
	}

	fmt.Printf("\n총 넓이: %.4f\n", TotalArea(shapes))
	fmt.Printf("가장 큰 도형: %v\n", LargestShape(shapes))

	filtered := FilterByMinArea(shapes, 10.0)
	fmt.Printf("넓이 10 이상: %d개\n", len(filtered))
}
