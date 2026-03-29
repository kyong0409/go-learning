# 과제 1: Shape 인터페이스와 다형성 구현

## 목표

Go 인터페이스의 암묵적 구현과 다형성을 이해합니다.

## 요구사항

### 1. Shape 인터페이스 정의

```go
type Shape interface {
    Area() float64      // 넓이 반환
    Perimeter() float64 // 둘레 반환
    String() string     // 문자열 표현 반환
}
```

### 2. 구체 타입 구현

다음 세 가지 도형을 구현하세요:

**Circle (원)**
- 필드: `Radius float64`
- 넓이: π × r²
- 둘레: 2 × π × r
- String: `"Circle(r=X.XX)"`

**Rectangle (직사각형)**
- 필드: `Width, Height float64`
- 넓이: width × height
- 둘레: 2 × (width + height)
- String: `"Rect(WxH)"`

**Triangle (삼각형)**
- 필드: `A, B, C float64` (세 변의 길이)
- 넓이: 헤론의 공식 `√(s(s-a)(s-b)(s-c))`, `s = (a+b+c)/2`
- 둘레: a + b + c
- String: `"Triangle(A,B,C)"`

### 3. 유틸리티 함수 구현

```go
// TotalArea는 도형 슬라이스의 총 넓이를 반환합니다.
func TotalArea(shapes []Shape) float64

// LargestShape는 가장 넓은 도형을 반환합니다.
// shapes가 비어있으면 nil을 반환합니다.
func LargestShape(shapes []Shape) Shape

// FilterByMinArea는 최소 넓이 이상인 도형만 반환합니다.
func FilterByMinArea(shapes []Shape, minArea float64) []Shape
```

## 실행 방법

```bash
# 테스트 실행
go test -v

# 특정 테스트만
go test -run TestCircle -v
```

## 채점

테스트 실행 시 채점 결과가 출력됩니다:

```
=== 채점 결과 ===
통과: 20/20
점수: 100/100
```

## 힌트

- `math.Pi` 상수를 사용하세요.
- `math.Sqrt()` 함수로 제곱근을 계산하세요.
- 부동소수점 비교는 `math.Abs(a-b) < 1e-9`를 사용하세요.
- 참고 풀이: `solution/shape.go`
