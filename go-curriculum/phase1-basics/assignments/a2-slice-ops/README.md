# 과제 A2: 슬라이스 유틸리티 함수 구현

## 목표

`sliceutil.go` 파일에 선언된 `[]int` 슬라이스 처리 함수들을 구현하세요.

## 함수 목록

| 함수 | 설명 |
|------|------|
| `Filter(nums []int, pred func(int) bool) []int` | 조건 함수가 true인 요소만 반환 |
| `Map(nums []int, f func(int) int) []int` | 모든 요소에 변환 함수를 적용한 새 슬라이스 반환 |
| `Reduce(nums []int, initial int, f func(int, int) int) int` | 슬라이스를 단일 값으로 축약 |
| `Contains(nums []int, target int) bool` | target이 포함되어 있으면 true |
| `Unique(nums []int) []int` | 중복 제거 (첫 등장 순서 유지) |
| `Sum(nums []int) int` | 모든 요소의 합 |
| `Max(nums []int) (int, error)` | 최댓값 (빈 슬라이스면 에러) |
| `Min(nums []int) (int, error)` | 최솟값 (빈 슬라이스면 에러) |
| `Reverse(nums []int) []int` | 순서를 뒤집은 새 슬라이스 반환 |
| `Flatten(nested [][]int) []int` | 2차원 슬라이스를 1차원으로 평탄화 |

## 구현 예시

```go
// Filter 예시:
// Filter([]int{1,2,3,4,5}, func(n int) bool { return n%2 == 0 })
// => []int{2, 4}

// Map 예시:
// Map([]int{1,2,3}, func(n int) int { return n * n })
// => []int{1, 4, 9}

// Reduce 예시:
// Reduce([]int{1,2,3,4,5}, 0, func(acc, n int) int { return acc + n })
// => 15
```

## 테스트 실행

```bash
go test -v
go test -v -run TestFilter   # 특정 테스트만
```

## 채점

- 총 20개 테스트, 100점 만점
- `=== GRADE REPORT ===` 에서 최종 점수 확인

## 힌트

- `Filter`와 `Map`은 새 슬라이스를 반환합니다 (원본 수정 금지).
- `Unique`는 첫 등장 순서를 유지해야 합니다. 맵을 활용하세요.
- `Reverse`도 원본을 수정하지 않고 새 슬라이스를 반환합니다.
- 빈 슬라이스에 대한 처리를 잊지 마세요.
