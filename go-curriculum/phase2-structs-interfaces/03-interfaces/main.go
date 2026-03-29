// 패키지 선언
package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// ─────────────────────────────────────────
// 1. 인터페이스 정의
// ─────────────────────────────────────────
// Go의 인터페이스는 암묵적(implicit)으로 충족됩니다.
// 즉, "implements" 키워드 없이 메서드만 구현하면 자동으로 인터페이스를 만족합니다.
// 이것을 "덕 타이핑(duck typing)"이라고 합니다: "오리처럼 걷고 꽥거리면 오리다"

// Shape는 도형 인터페이스입니다.
// 작은 인터페이스 설계 원칙: 인터페이스는 가능한 작게 유지합니다.
type Shape interface {
	Area() float64      // 넓이
	Perimeter() float64 // 둘레
}

// Stringer는 문자열 표현 인터페이스입니다.
// (실제로는 fmt.Stringer가 있지만 예시로 직접 정의)
type Stringer interface {
	String() string
}

// ─────────────────────────────────────────
// 2. 인터페이스를 구현하는 타입들
// ─────────────────────────────────────────

// Circle은 원입니다.
type Circle struct {
	Radius float64
}

// Area는 원의 넓이입니다 (Shape 인터페이스 구현).
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Perimeter는 원의 둘레입니다 (Shape 인터페이스 구현).
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// String은 원의 문자열 표현입니다.
func (c Circle) String() string {
	return fmt.Sprintf("Circle(r=%.2f)", c.Radius)
}

// Rectangle은 직사각형입니다.
type Rectangle struct {
	Width, Height float64
}

// Area는 직사각형의 넓이입니다 (Shape 인터페이스 구현).
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter는 직사각형의 둘레입니다 (Shape 인터페이스 구현).
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// String은 직사각형의 문자열 표현입니다.
func (r Rectangle) String() string {
	return fmt.Sprintf("Rect(%.2f×%.2f)", r.Width, r.Height)
}

// Triangle은 삼각형입니다.
type Triangle struct {
	A, B, C float64 // 세 변의 길이
}

// Area는 헤론의 공식으로 삼각형 넓이를 계산합니다.
func (t Triangle) Area() float64 {
	s := (t.A + t.B + t.C) / 2 // 반둘레
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

// Perimeter는 삼각형의 둘레입니다.
func (t Triangle) Perimeter() float64 {
	return t.A + t.B + t.C
}

// String은 삼각형의 문자열 표현입니다.
func (t Triangle) String() string {
	return fmt.Sprintf("Triangle(%.2f, %.2f, %.2f)", t.A, t.B, t.C)
}

// ─────────────────────────────────────────
// 3. 인터페이스를 사용하는 함수들
// ─────────────────────────────────────────

// printShapeInfo는 Shape 인터페이스를 받습니다.
// 구체적인 타입(Circle, Rectangle 등)을 알 필요가 없습니다.
// "인터페이스를 받아들이고 구조체를 반환하라" 패턴의 인수 부분
func printShapeInfo(s Shape) {
	fmt.Printf("  넓이: %.4f, 둘레: %.4f\n", s.Area(), s.Perimeter())
}

// totalArea는 도형 슬라이스의 총 넓이를 계산합니다.
func totalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

// largestShape는 가장 넓은 도형을 반환합니다.
// 반환 타입도 인터페이스입니다.
func largestShape(shapes []Shape) Shape {
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

// ─────────────────────────────────────────
// 4. io.Reader / io.Writer 예제
// ─────────────────────────────────────────
// io.Reader와 io.Writer는 Go 표준 라이브러리의 핵심 인터페이스입니다.
//
// type Reader interface {
//     Read(p []byte) (n int, err error)
// }
//
// type Writer interface {
//     Write(p []byte) (n int, err error)
// }

// UpperCaseWriter는 소문자를 대문자로 변환하여 출력하는 Writer입니다.
// io.Writer 인터페이스를 구현합니다.
type UpperCaseWriter struct {
	w io.Writer // 내부 Writer에 위임
}

// Write는 io.Writer 인터페이스를 구현합니다.
func (u UpperCaseWriter) Write(p []byte) (n int, err error) {
	upper := strings.ToUpper(string(p))
	return u.w.Write([]byte(upper))
}

// WordCounter는 쓰여진 단어 수를 세는 Writer입니다.
type WordCounter struct {
	count int
}

// Write는 io.Writer 인터페이스를 구현합니다.
func (wc *WordCounter) Write(p []byte) (n int, err error) {
	words := strings.Fields(string(p))
	wc.count += len(words)
	return len(p), nil
}

// Count는 단어 수를 반환합니다.
func (wc *WordCounter) Count() int {
	return wc.count
}

// ─────────────────────────────────────────
// 5. 빈 인터페이스(any / interface{})
// ─────────────────────────────────────────
// any는 interface{}의 별칭입니다 (Go 1.18+).
// 어떤 타입의 값이든 저장할 수 있습니다.
// 하지만 타입 정보가 없어서 사용 시 타입 단언이 필요합니다.

// describe는 any 타입을 받아 설명합니다.
func describe(v any) string {
	return fmt.Sprintf("타입: %-10T  값: %v", v, v)
}

// ─────────────────────────────────────────
// 6. 타입 단언(Type Assertion)
// ─────────────────────────────────────────

// tryGetArea는 any 타입에서 Shape을 꺼내려 시도합니다.
func tryGetArea(v any) (float64, bool) {
	// 단일 반환값 형태: 실패 시 panic 발생 (주의!)
	// shape := v.(Shape)

	// 안전한 형태: ok가 false면 panic 없음
	shape, ok := v.(Shape)
	if !ok {
		return 0, false
	}
	return shape.Area(), true
}

// ─────────────────────────────────────────
// 7. 타입 스위치(Type Switch)
// ─────────────────────────────────────────

// processValue는 타입 스위치로 다양한 타입을 처리합니다.
func processValue(v any) string {
	switch val := v.(type) {
	case int:
		return fmt.Sprintf("정수: %d (2배=%d)", val, val*2)
	case float64:
		return fmt.Sprintf("실수: %.2f (반올림=%d)", val, int(math.Round(val)))
	case string:
		return fmt.Sprintf("문자열: %q (길이=%d)", val, len(val))
	case bool:
		if val {
			return "불리언: true (참)"
		}
		return "불리언: false (거짓)"
	case []int:
		return fmt.Sprintf("int 슬라이스: %v (길이=%d)", val, len(val))
	case Shape:
		return fmt.Sprintf("도형: 넓이=%.4f", val.Area())
	case nil:
		return "nil 값"
	default:
		return fmt.Sprintf("알 수 없는 타입: %T", val)
	}
}

func main() {
	fmt.Println("=== Go Phase 2: 인터페이스(Interfaces) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 암묵적 인터페이스 충족
	// ─────────────────────────────────────────
	fmt.Println("--- 1. 암묵적 인터페이스 충족 ---")

	// Circle, Rectangle, Triangle 모두 Shape 인터페이스를 구현합니다.
	// 명시적 선언(implements) 없이 메서드만 있으면 됩니다.
	var s1 Shape = Circle{Radius: 5}
	var s2 Shape = Rectangle{Width: 4, Height: 6}
	var s3 Shape = Triangle{A: 3, B: 4, C: 5}

	shapes := []Shape{s1, s2, s3}

	for _, s := range shapes {
		// %v는 String() 메서드를 자동 호출 (fmt.Stringer 구현 시)
		fmt.Printf("  %v → ", s)
		printShapeInfo(s)
	}

	fmt.Printf("\n총 넓이: %.4f\n", totalArea(shapes))

	largest := largestShape(shapes)
	fmt.Printf("가장 큰 도형: %v (넓이=%.4f)\n", largest, largest.Area())
	fmt.Println()

	// ─────────────────────────────────────────
	// io.Reader / io.Writer 예제
	// ─────────────────────────────────────────
	fmt.Println("--- 2. io.Reader / io.Writer ---")

	// strings.NewReader: 문자열을 io.Reader로 변환
	reader := strings.NewReader("Hello, Go 인터페이스!")
	buf := make([]byte, 10)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			fmt.Printf("  읽음: %q (%d bytes)\n", buf[:n], n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("  에러: %v\n", err)
			break
		}
	}

	// io.Copy: Reader에서 Writer로 복사 (인터페이스 조합의 힘)
	fmt.Println("\nio.Copy 예제 (strings.Reader → os.Stdout):")
	reader2 := strings.NewReader("  io.Copy로 출력됩니다!\n")
	io.Copy(os.Stdout, reader2)

	// UpperCaseWriter: 커스텀 Writer
	fmt.Println("\nUpperCaseWriter 예제:")
	ucw := UpperCaseWriter{w: os.Stdout}
	fmt.Fprintln(ucw, "  hello, world! 소문자가 대문자로 변환됩니다.")

	// WordCounter: 단어 수 세기
	wc := &WordCounter{}
	fmt.Fprintln(wc, "the quick brown fox jumps over the lazy dog")
	fmt.Fprintln(wc, "go is awesome and simple")
	fmt.Printf("\nWordCounter: 총 %d 단어 작성됨\n", wc.Count())
	fmt.Println()

	// ─────────────────────────────────────────
	// 빈 인터페이스 (any)
	// ─────────────────────────────────────────
	fmt.Println("--- 3. 빈 인터페이스 (any) ---")

	values := []any{
		42,
		3.14,
		"안녕하세요",
		true,
		[]int{1, 2, 3},
		Circle{Radius: 3},
		nil,
	}

	for _, v := range values {
		fmt.Printf("  %s\n", describe(v))
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 타입 단언 (Type Assertion)
	// ─────────────────────────────────────────
	fmt.Println("--- 4. 타입 단언 (Type Assertion) ---")

	var i any = "Go 언어"

	// 안전한 타입 단언 (ok 패턴)
	if str, ok := i.(string); ok {
		fmt.Printf("  문자열로 단언 성공: %q (길이=%d)\n", str, len(str))
	}

	// 잘못된 타입 단언 (안전한 방식)
	if num, ok := i.(int); ok {
		fmt.Printf("  정수로 단언: %d\n", num)
	} else {
		fmt.Printf("  정수 단언 실패 (실제 타입: %T)\n", i)
	}

	// Shape 단언
	testValues := []any{
		Circle{Radius: 2},
		Rectangle{Width: 3, Height: 4},
		"도형 아님",
		42,
	}

	fmt.Println("\n  Shape 단언 테스트:")
	for _, v := range testValues {
		if area, ok := tryGetArea(v); ok {
			fmt.Printf("    %v → 넓이=%.4f\n", v, area)
		} else {
			fmt.Printf("    %v → Shape 아님 (%T)\n", v, v)
		}
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 타입 스위치 (Type Switch)
	// ─────────────────────────────────────────
	fmt.Println("--- 5. 타입 스위치 (Type Switch) ---")

	mixed := []any{
		100,
		3.14159,
		"Hello, 세계",
		false,
		[]int{10, 20, 30},
		Rectangle{Width: 5, Height: 3},
		nil,
		complex(1, 2), // 정의된 케이스 없음
	}

	for _, v := range mixed {
		fmt.Printf("  %s\n", processValue(v))
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 인터페이스 nil 주의사항
	// ─────────────────────────────────────────
	fmt.Println("--- 6. 인터페이스 nil 주의사항 ---")

	// nil 인터페이스: 타입도 없고 값도 없음
	var nilShape Shape
	fmt.Printf("nil 인터페이스: 값=%v, nil여부=%v\n", nilShape, nilShape == nil)

	// 인터페이스는 (타입, 값) 쌍으로 구성됩니다.
	// 타입이 있으면 값이 nil이어도 인터페이스 자체는 nil이 아닙니다!
	var c *Circle = nil // nil 포인터
	var s Shape = c    // 타입(*Circle)은 있지만 값은 nil
	fmt.Printf("nil 포인터가 담긴 인터페이스: 값=%v, nil여부=%v\n", s, s == nil)
	// s == nil은 false! (타입 정보가 있으므로)

	fmt.Println("  주의: 인터페이스 == nil은 타입과 값이 모두 nil일 때만 true")

	fmt.Println()
	fmt.Println("=== 인터페이스 학습 완료 ===")
}
