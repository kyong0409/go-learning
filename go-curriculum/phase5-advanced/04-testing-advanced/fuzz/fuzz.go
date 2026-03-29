// fuzz/fuzz.go
// 퍼즈 테스트 대상 파서 함수
//
// 퍼즈 테스트(Fuzz Testing)는 자동으로 생성된 무작위 입력으로
// 함수를 테스트해 예상치 못한 패닉이나 오류를 찾습니다.
// Go 1.18+에서 표준 라이브러리로 지원됩니다.
package fuzz

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ============================================================
// 간단한 키-값 파서 (퍼즈 테스트 대상)
// ============================================================

// KVEntry는 파싱된 키-값 쌍을 표현합니다.
type KVEntry struct {
	Key   string
	Value string
}

// ParseKV는 "key=value" 형식의 문자열을 파싱합니다.
//
// 입력 형식:
//   - "key=value"
//   - "key=value,key2=value2"
//   - 키는 영문자와 숫자, '_', '-'만 허용
//   - 값은 모든 문자 허용 (공백 trim)
//
// 오류 케이스 (퍼즈 테스트로 찾을 수 있는 엣지 케이스):
//   - 빈 입력
//   - '=' 없는 항목
//   - 빈 키
//   - 키에 허용되지 않은 문자
func ParseKV(input string) ([]KVEntry, error) {
	if input == "" {
		return nil, nil
	}

	// 쉼표로 분리
	parts := strings.Split(input, ",")
	result := make([]KVEntry, 0, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// '=' 위치 찾기
		eqIdx := strings.Index(part, "=")
		if eqIdx < 0 {
			return nil, fmt.Errorf("항목 %d: '=' 없음 (%q)", i, part)
		}

		key := strings.TrimSpace(part[:eqIdx])
		value := strings.TrimSpace(part[eqIdx+1:])

		// 키 유효성 검사
		if key == "" {
			return nil, fmt.Errorf("항목 %d: 빈 키", i)
		}
		if err := validateKey(key); err != nil {
			return nil, fmt.Errorf("항목 %d 키 오류: %w", i, err)
		}

		result = append(result, KVEntry{Key: key, Value: value})
	}

	return result, nil
}

// validateKey는 키가 허용된 문자로만 구성됐는지 검사합니다.
func validateKey(key string) error {
	for _, r := range key {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return fmt.Errorf("허용되지 않는 문자: %q", r)
		}
	}
	return nil
}

// ============================================================
// 간단한 수식 파서 (퍼즈 테스트 대상 2)
// ============================================================

// CalcResult는 수식 파싱/계산 결과를 표현합니다.
type CalcResult struct {
	Left     int
	Operator rune
	Right    int
	Result   int
}

// ParseAndCalc는 "num op num" 형식의 간단한 수식을 계산합니다.
//
// 지원 연산자: +, -, *, /
// 예시: "10 + 20", "100 - 5", "3 * 7", "20 / 4"
//
// 퍼즈 테스트로 찾을 수 있는 문제:
//   - 0으로 나누기 패닉
//   - 정수 오버플로우
//   - 예상치 못한 공백 처리
func ParseAndCalc(expr string) (*CalcResult, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil, errors.New("빈 수식")
	}

	// 공백으로 분리
	tokens := strings.Fields(expr)
	if len(tokens) != 3 {
		return nil, fmt.Errorf("수식 형식 오류: 토큰 %d개 (필요: 3)", len(tokens))
	}

	left, err := strconv.Atoi(tokens[0])
	if err != nil {
		return nil, fmt.Errorf("첫 번째 피연산자 오류: %w", err)
	}

	right, err := strconv.Atoi(tokens[2])
	if err != nil {
		return nil, fmt.Errorf("두 번째 피연산자 오류: %w", err)
	}

	if len(tokens[1]) != 1 {
		return nil, fmt.Errorf("연산자는 한 글자여야 합니다: %q", tokens[1])
	}
	op := rune(tokens[1][0])

	var result int
	switch op {
	case '+':
		result = left + right
	case '-':
		result = left - right
	case '*':
		result = left * right
	case '/':
		if right == 0 {
			return nil, errors.New("0으로 나눌 수 없습니다")
		}
		result = left / right
	default:
		return nil, fmt.Errorf("지원하지 않는 연산자: %q", op)
	}

	return &CalcResult{
		Left:     left,
		Operator: op,
		Right:    right,
		Result:   result,
	}, nil
}

// ============================================================
// 간단한 CSV 파서 (퍼즈 테스트 대상 3)
// ============================================================

// ParseCSVLine은 단일 CSV 라인을 파싱합니다.
// 따옴표 처리, 이스케이프, 빈 필드를 올바르게 처리해야 합니다.
func ParseCSVLine(line string) ([]string, error) {
	if line == "" {
		return []string{}, nil
	}

	var fields []string
	var current strings.Builder
	inQuotes := false

	for i, ch := range line {
		switch {
		case ch == '"':
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				// 이스케이프된 따옴표 ("") → "
				current.WriteRune('"')
			} else {
				inQuotes = !inQuotes
			}
		case ch == ',' && !inQuotes:
			fields = append(fields, current.String())
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if inQuotes {
		return nil, errors.New("따옴표가 닫히지 않았습니다")
	}

	fields = append(fields, current.String())
	return fields, nil
}
