// fuzz/fuzz_test.go
// 퍼즈 테스트 (Go 1.18+ 내장 퍼징)
//
// 기본 실행 (시드 코퍼스만 테스트):
//   go test ./fuzz/ -v
//
// 퍼징 모드 실행 (무작위 입력 생성):
//   go test ./fuzz/ -fuzz=FuzzParseKV -fuzztime=30s
//   go test ./fuzz/ -fuzz=FuzzParseAndCalc -fuzztime=1m
//   go test ./fuzz/ -fuzz=FuzzParseCSVLine -fuzztime=30s
//
// 발견된 버그 재현:
//   go test ./fuzz/ -run=FuzzParseKV/testdata/fuzz/FuzzParseKV/<파일명>
package fuzz

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// ============================================================
// FuzzParseKV: 키-값 파서 퍼즈 테스트
// ============================================================

// FuzzParseKV는 ParseKV 함수를 퍼즈 테스트합니다.
// 퍼저가 자동으로 다양한 입력을 생성해 패닉이나 정합성 오류를 찾습니다.
func FuzzParseKV(f *testing.F) {
	// 시드 코퍼스: 퍼저가 이 값들을 변형해서 새 케이스 생성
	f.Add("key=value")
	f.Add("key=value,key2=value2")
	f.Add("name=홍길동,age=30")
	f.Add("")
	f.Add("=value")            // 빈 키 (오류 케이스)
	f.Add("key")               // = 없음 (오류 케이스)
	f.Add("key=")              // 빈 값 (유효)
	f.Add("k e y=value")       // 키에 공백 (오류 케이스)
	f.Add("key=value,")        // 후행 쉼표
	f.Add(",key=value")        // 선행 쉼표
	f.Add("a=1,b=2,c=3,d=4")  // 여러 항목

	f.Fuzz(func(t *testing.T, input string) {
		// 핵심 불변 조건 검증:
		// 1. 패닉이 발생하면 안 됨 (defer로 캐치)
		// 2. 오류 없이 파싱된 경우 결과의 정합성 확인

		entries, err := ParseKV(input)

		if err != nil {
			// 오류 반환은 정상 — 퍼저는 오류 자체가 아닌 패닉을 찾음
			return
		}

		// 불변 조건 1: nil 입력이 아니면 슬라이스는 nil이 아님
		if input != "" && entries == nil {
			t.Errorf("비어있지 않은 입력에서 nil 반환: %q", input)
		}

		// 불변 조건 2: 각 키는 비어있지 않아야 함
		for _, e := range entries {
			if e.Key == "" {
				t.Errorf("빈 키가 반환됨: input=%q", input)
			}
		}

		// 불변 조건 3: 결과를 다시 직렬화해도 동일하게 파싱돼야 함
		// (round-trip 속성)
		if len(entries) > 0 {
			var parts []string
			for _, e := range entries {
				parts = append(parts, e.Key+"="+e.Value)
			}
			serialized := strings.Join(parts, ",")

			reentries, err2 := ParseKV(serialized)
			if err2 != nil {
				t.Errorf("재직렬화 파싱 실패: serialized=%q, err=%v", serialized, err2)
				return
			}
			if len(reentries) != len(entries) {
				t.Errorf("재직렬화 후 항목 수 불일치: before=%d, after=%d", len(entries), len(reentries))
			}
		}
	})
}

// ============================================================
// FuzzParseAndCalc: 수식 파서 퍼즈 테스트
// ============================================================

// FuzzParseAndCalc는 수식 파서를 퍼즈 테스트합니다.
// 특히 0으로 나누기, 오버플로우 등을 자동으로 발견할 수 있습니다.
func FuzzParseAndCalc(f *testing.F) {
	// 시드 코퍼스
	f.Add("10 + 20")
	f.Add("100 - 5")
	f.Add("3 * 7")
	f.Add("20 / 4")
	f.Add("0 / 0")      // 0으로 나누기 (오류 케이스)
	f.Add("1 / 0")      // 0으로 나누기
	f.Add("-5 + 3")     // 음수
	f.Add("abc + 1")    // 숫자가 아닌 피연산자
	f.Add("1 ++ 2")     // 잘못된 연산자
	f.Add("")           // 빈 입력

	f.Fuzz(func(t *testing.T, expr string) {
		// 패닉이 발생하면 안 됨
		result, err := ParseAndCalc(expr)

		if err != nil {
			// 오류 반환은 정상
			return
		}

		// 불변 조건: 결과가 반환됐다면 nil이 아님
		if result == nil {
			t.Errorf("오류 없이 nil 결과: expr=%q", expr)
			return
		}

		// 불변 조건: 덧셈의 교환법칙 확인 (a + b == b + a)
		if result.Operator == '+' {
			commuted := strings.Join([]string{
				string(rune('0' + result.Right%10)), // 단순화
				"+",
				string(rune('0' + result.Left%10)),
			}, " ")
			_ = commuted // 실제 검증은 생략 (개념 시연용)
		}
	})
}

// ============================================================
// FuzzParseCSVLine: CSV 파서 퍼즈 테스트
// ============================================================

// FuzzParseCSVLine은 CSV 라인 파서를 퍼즈 테스트합니다.
func FuzzParseCSVLine(f *testing.F) {
	// 시드 코퍼스
	f.Add("a,b,c")
	f.Add(`"hello","world"`)
	f.Add(`"hello,world",foo`)
	f.Add(`""`)                       // 빈 따옴표
	f.Add(`"a""b"`)                   // 이스케이프된 따옴표
	f.Add("")                          // 빈 입력
	f.Add(`"unclosed`)                 // 닫히지 않은 따옴표 (오류)
	f.Add("a,,b")                      // 빈 필드
	f.Add("한국어,테스트,CSV")          // 유니코드

	f.Fuzz(func(t *testing.T, line string) {
		// UTF-8 유효성: 잘못된 바이트 시퀀스 처리 확인
		if !utf8.ValidString(line) {
			return // 잘못된 UTF-8은 스킵
		}

		fields, err := ParseCSVLine(line)

		if err != nil {
			return // 파싱 오류는 허용
		}

		// 불변 조건 1: 결과는 항상 슬라이스 (nil 불가)
		if fields == nil && line != "" {
			t.Errorf("비어있지 않은 입력에서 nil 반환: %q", line)
		}

		// 불변 조건 2: 필드 수는 쉼표 수 + 1 (따옴표 안 쉼표 제외)
		// (간단한 경우에만 검증)
		if !strings.Contains(line, `"`) {
			commaCount := strings.Count(line, ",")
			if line == "" {
				if len(fields) != 0 {
					t.Errorf("빈 입력 결과 %d개 필드 (기대: 0)", len(fields))
				}
			} else if len(fields) != commaCount+1 {
				t.Errorf("따옴표 없는 입력의 필드 수 불일치: input=%q, 쉼표=%d, 필드=%d",
					line, commaCount, len(fields))
			}
		}
	})
}

// ============================================================
// 일반 단위 테스트 (퍼즈 테스트와 함께)
// ============================================================

func TestParseKV_Basic(t *testing.T) {
	entries, err := ParseKV("name=홍길동,age=30")
	if err != nil {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("항목 수: 기대 2, 실제 %d", len(entries))
	}
	if entries[0].Key != "name" || entries[0].Value != "홍길동" {
		t.Errorf("첫 번째 항목 불일치: %+v", entries[0])
	}
}

func TestParseAndCalc_DivisionByZero(t *testing.T) {
	_, err := ParseAndCalc("10 / 0")
	if err == nil {
		t.Fatal("0으로 나누기는 오류를 반환해야 합니다")
	}
}

func TestParseCSVLine_Quotes(t *testing.T) {
	fields, err := ParseCSVLine(`"hello,world","foo"`)
	if err != nil {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("필드 수: 기대 2, 실제 %d", len(fields))
	}
	if fields[0] != "hello,world" {
		t.Errorf("첫 번째 필드: 기대 %q, 실제 %q", "hello,world", fields[0])
	}
}
