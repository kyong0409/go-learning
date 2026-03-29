// Package main은 Shape 인터페이스 과제 구현 파일입니다.
// TODO 주석을 찾아 구현을 완성하세요.
package main

import "fmt"

// ─────────────────────────────────────────
// Shape 인터페이스
// ─────────────────────────────────────────

// Shape는 도형 인터페이스입니다.
// TODO: Area(), Perimeter(), String() 메서드를 정의하세요.
type Shape interface {
	// TODO: 넓이를 반환하는 메서드
	// TODO: 둘레를 반환하는 메서드
	// TODO: 문자열 표현을 반환하는 메서드
}

// ─────────────────────────────────────────
// Circle (원)
// ─────────────────────────────────────────

// Circle은 원을 나타냅니다.
type Circle struct {
	Radius float64
}

// TODO: Circle의 Area() 메서드를 구현하세요.
// 공식: π × r²
// func (c Circle) Area() float64 { ... }

// TODO: Circle의 Perimeter() 메서드를 구현하세요.
// 공식: 2 × π × r
// func (c Circle) Perimeter() float64 { ... }

// TODO: Circle의 String() 메서드를 구현하세요.
// 형식: "Circle(r=X.XX)" (소수점 2자리)
// func (c Circle) String() string { ... }

// ─────────────────────────────────────────
// Rectangle (직사각형)
// ─────────────────────────────────────────

// Rectangle은 직사각형을 나타냅니다.
type Rectangle struct {
	Width  float64
	Height float64
}

// TODO: Rectangle의 Area() 메서드를 구현하세요.
// 공식: width × height
// func (r Rectangle) Area() float64 { ... }

// TODO: Rectangle의 Perimeter() 메서드를 구현하세요.
// 공식: 2 × (width + height)
// func (r Rectangle) Perimeter() float64 { ... }

// TODO: Rectangle의 String() 메서드를 구현하세요.
// 형식: "Rect(W.XXxH.XX)" (소수점 2자리)
// func (r Rectangle) String() string { ... }

// ─────────────────────────────────────────
// Triangle (삼각형)
// ─────────────────────────────────────────

// Triangle은 삼각형을 나타냅니다 (세 변의 길이로 정의).
type Triangle struct {
	A, B, C float64 // 세 변의 길이
}

// TODO: Triangle의 Area() 메서드를 구현하세요.
// 헤론의 공식: s = (A+B+C)/2, 넓이 = √(s(s-A)(s-B)(s-C))
// func (t Triangle) Area() float64 { ... }

// TODO: Triangle의 Perimeter() 메서드를 구현하세요.
// 공식: A + B + C
// func (t Triangle) Perimeter() float64 { ... }

// TODO: Triangle의 String() 메서드를 구현하세요.
// 형식: "Triangle(A.XX,B.XX,C.XX)" (소수점 2자리)
// func (t Triangle) String() string { ... }

// ─────────────────────────────────────────
// 유틸리티 함수
// ─────────────────────────────────────────

// TotalArea는 도형 슬라이스의 총 넓이를 반환합니다.
// TODO: 구현하세요.
func TotalArea(shapes []Shape) float64 {
	// TODO: 모든 도형의 넓이를 합산하여 반환하세요.
	return 0
}

// LargestShape는 가장 넓은 도형을 반환합니다.
// shapes가 비어있으면 nil을 반환합니다.
// TODO: 구현하세요.
func LargestShape(shapes []Shape) Shape {
	// TODO: 가장 큰 넓이를 가진 도형을 찾아 반환하세요.
	return nil
}

// FilterByMinArea는 최소 넓이 이상인 도형만 반환합니다.
// TODO: 구현하세요.
func FilterByMinArea(shapes []Shape, minArea float64) []Shape {
	// TODO: minArea 이상인 도형만 모아서 반환하세요.
	return nil
}

// main은 간단한 데모를 실행합니다.
func main() {
	fmt.Println("Shape 인터페이스 과제")
	fmt.Println("shape.go의 TODO를 구현하고 'go test -v'로 테스트하세요.")
}
