# 과제 A1: 간단한 계산기 함수 구현

## 목표

`calculator.go` 파일에 선언된 함수들의 본문을 구현하세요.

## 함수 목록

| 함수 | 설명 | 반환 |
|------|------|------|
| `Add(a, b float64) float64` | 덧셈 | a + b |
| `Subtract(a, b float64) float64` | 뺄셈 | a - b |
| `Multiply(a, b float64) float64` | 곱셈 | a * b |
| `Divide(a, b float64) (float64, error)` | 나눗셈 | a / b, 0으로 나누면 에러 |
| `Power(base, exp float64) float64` | 거듭제곱 | base^exp |
| `Sqrt(n float64) (float64, error)` | 제곱근 | √n, 음수면 에러 |
| `Abs(n float64) float64` | 절댓값 | \|n\| |
| `Modulo(a, b int) (int, error)` | 나머지 | a % b, 0이면 에러 |

## 구현 방법

`calculator.go`를 열고 `// TODO: 구현하세요` 부분을 채우세요.

```go
func Add(a, b float64) float64 {
    // TODO: 구현하세요
    return 0
}
```

## 에러 처리 요구사항

- `Divide(a, 0)`: `ErrDivisionByZero` 에러 반환
- `Sqrt(-1)`: `ErrNegativeInput` 에러 반환
- `Modulo(a, 0)`: `ErrDivisionByZero` 에러 반환

## 테스트 실행

```bash
go test -v
go test -v -run TestAdd      # 특정 테스트만
go test -count=1 -v          # 캐시 없이 실행
```

## 채점

- 총 15개 테스트
- 테스트 통과 시 점수 획득
- 최종 점수는 `=== GRADE REPORT ===` 섹션에서 확인

## 힌트

- `math.Pow(base, exp)` 함수를 사용하면 거듭제곱을 계산할 수 있습니다.
- `math.Sqrt(n)` 함수를 사용하면 제곱근을 계산할 수 있습니다.
- `errors.New()` 또는 `fmt.Errorf()`로 에러를 만들 수 있습니다.
- `import "math"` 를 추가해야 합니다.
