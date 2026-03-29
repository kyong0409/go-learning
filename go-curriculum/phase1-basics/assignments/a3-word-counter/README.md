# 과제 A3: 파일에서 단어 빈도수 세기

## 목표

`wordcount.go` 파일에 선언된 함수들을 구현하여, 텍스트 파일에서 단어의 출현 빈도수를 계산하고 상위 N개 단어를 반환하는 프로그램을 완성하세요.

## 함수 목록

| 함수 | 설명 |
|------|------|
| `CountWords(text string) map[string]int` | 텍스트에서 단어 빈도수를 계산해 맵으로 반환 |
| `TopN(counts map[string]int, n int) []WordCount` | 빈도수 상위 N개 단어를 내림차순으로 반환 |
| `CountWordsFromFile(filename string) (map[string]int, error)` | 파일을 읽어 단어 빈도수를 계산 |
| `TotalWords(counts map[string]int) int` | 전체 단어 수 (중복 포함) |
| `UniqueWords(counts map[string]int) int` | 고유 단어 수 |

## 처리 규칙

- **대소문자 무시**: "Go"와 "go"는 같은 단어로 처리 (모두 소문자로 변환)
- **구두점 제거**: "hello,"와 "hello"는 같은 단어로 처리
- **빈 문자열 무시**: 구두점 제거 후 빈 문자열이 된 토큰은 제외
- **TopN 정렬**: 빈도수 내림차순. 빈도수가 같으면 알파벳 오름차순으로 정렬

## WordCount 구조체

```go
type WordCount struct {
    Word  string
    Count int
}
```

## 테스트 실행

```bash
go test -v
go test -v -run TestCountWords   # 특정 테스트만
```

## 채점

- 총 15개 테스트, 100점 만점
- `=== GRADE REPORT ===` 에서 최종 점수 확인

## 힌트

- `strings.ToLower()`: 소문자 변환
- `strings.Fields()`: 공백으로 단어 분리
- `strings.Trim(word, ".,!?;:\"'()[]")`: 구두점 제거
- `sort.Slice()`: 커스텀 정렬
- `os.ReadFile()`: 파일 읽기
