// 패키지 선언
package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== Go 기초: 문자열 (Strings) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 1. Go의 문자열과 UTF-8
	// ─────────────────────────────────────────
	fmt.Println("--- 1. 문자열과 UTF-8 인코딩 ---")

	// Go 문자열은 불변(immutable)의 바이트 슬라이스입니다.
	// 내부적으로 UTF-8로 인코딩됩니다.
	s := "Hello, 세계!" // 한국어 포함 문자열

	fmt.Printf("문자열: %q\n", s)
	fmt.Printf("바이트 수 (len): %d\n", len(s))                   // 바이트 수 (영어 1byte, 한국어 3byte)
	fmt.Printf("룬(문자) 수: %d\n", utf8.RuneCountInString(s))    // 실제 문자 수

	// 바이트 단위로 순회 (UTF-8 원시 바이트)
	fmt.Println("\n바이트 순회:")
	for i := 0; i < len(s); i++ {
		fmt.Printf("  s[%2d] = 0x%02X (%d)\n", i, s[i], s[i])
	}

	// 룬(rune) 단위로 순회 (유니코드 코드 포인트)
	fmt.Println("\n룬 순회 (range):")
	for i, r := range s {
		fmt.Printf("  인덱스[%2d]: U+%04X = %q (%d bytes)\n",
			i, r, r, utf8.RuneLen(r))
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 2. string vs []byte vs []rune
	// ─────────────────────────────────────────
	fmt.Println("--- 2. string vs []byte vs []rune ---")

	original := "Hello, 한국"

	// string -> []byte: 바이트 수준 조작 시 사용
	bytes := []byte(original)
	fmt.Printf("[]byte: %v\n", bytes)
	fmt.Printf("길이: %d bytes\n", len(bytes))

	// bytes 조작
	bytes[0] = 'h' // 소문자로 변경
	modified := string(bytes)
	fmt.Printf("bytes[0]='h' 후: %q\n", modified)

	// string -> []rune: 문자 단위 인덱싱 시 사용
	runes := []rune(original)
	fmt.Printf("\n[]rune: %v\n", runes)
	fmt.Printf("길이: %d 문자\n", len(runes))

	// 룬으로 5번째 문자 접근
	fmt.Printf("runes[7] = %q (7번째 문자)\n", runes[7])

	// string[i]는 바이트 반환 (주의!)
	fmt.Printf("\noriginal[7] = %d (바이트, 한국어 첫 바이트)\n", original[7])
	// 반면 runes[7]은 실제 '한' 문자를 반환

	// []rune -> string 복원
	restored := string(runes)
	fmt.Printf("[]rune -> string: %q\n", restored)
	fmt.Println()

	// ─────────────────────────────────────────
	// 3. 문자열 비교와 불변성
	// ─────────────────────────────────────────
	fmt.Println("--- 3. 문자열 비교와 불변성 ---")

	a, b := "hello", "hello"
	c := "world"
	fmt.Printf("%q == %q: %t\n", a, b, a == b)
	fmt.Printf("%q == %q: %t\n", a, c, a == c)
	fmt.Printf("%q < %q: %t\n", a, c, a < c)
	fmt.Printf("%q > %q: %t\n", c, a, c > a)

	// 문자열은 불변이므로 직접 수정 불가
	// s[0] = 'h' // 컴파일 에러!
	// 수정하려면 []byte나 []rune으로 변환 후 다시 string으로
	str := "Hello"
	strBytes := []byte(str)
	strBytes[0] = 'h'
	str = string(strBytes)
	fmt.Printf("수정된 문자열: %q\n", str)
	fmt.Println()

	// ─────────────────────────────────────────
	// 4. strings 패키지
	// ─────────────────────────────────────────
	fmt.Println("--- 4. strings 패키지 함수들 ---")

	text := "  Hello, Go 세계!  "
	fmt.Printf("원본: %q\n\n", text)

	// 검색
	fmt.Println("[검색]")
	fmt.Printf("Contains(\"Go\"): %t\n", strings.Contains(text, "Go"))
	fmt.Printf("Contains(\"Python\"): %t\n", strings.Contains(text, "Python"))
	fmt.Printf("HasPrefix(\"  Hello\"): %t\n", strings.HasPrefix(text, "  Hello"))
	fmt.Printf("HasSuffix(\"!  \"): %t\n", strings.HasSuffix(text, "!  "))
	fmt.Printf("Index(\"Go\"): %d\n", strings.Index(text, "Go"))
	fmt.Printf("LastIndex(\"l\"): %d\n", strings.LastIndex(text, "l"))
	fmt.Printf("Count(\"l\"): %d\n", strings.Count(text, "l"))
	fmt.Printf("ContainsAny(\"aeiou\"): %t\n", strings.ContainsAny(text, "aeiou"))

	// 변환
	fmt.Println("\n[변환]")
	fmt.Printf("ToUpper: %q\n", strings.ToUpper(text))
	fmt.Printf("ToLower: %q\n", strings.ToLower(text))
	fmt.Printf("Title (deprecated): %q\n", strings.Title("hello world")) //nolint
	fmt.Printf("TrimSpace: %q\n", strings.TrimSpace(text))
	fmt.Printf("Trim(\"H!\"): %q\n", strings.Trim("!Hello!", "!"))
	fmt.Printf("TrimLeft(\" \"): %q\n", strings.TrimLeft(text, " "))
	fmt.Printf("TrimRight(\" \"): %q\n", strings.TrimRight(text, " "))
	fmt.Printf("TrimPrefix: %q\n", strings.TrimPrefix("Hello, Go!", "Hello, "))
	fmt.Printf("TrimSuffix: %q\n", strings.TrimSuffix("Hello, Go!", ", Go!"))
	fmt.Printf("Replace: %q\n", strings.Replace(text, "l", "L", 2)) // 2번만
	fmt.Printf("ReplaceAll: %q\n", strings.ReplaceAll(text, "l", "L"))

	// 분리 및 결합
	fmt.Println("\n[분리/결합]")
	csv := "사과,바나나,체리,포도"
	parts := strings.Split(csv, ",")
	fmt.Printf("Split(%q, \",\"): %v\n", csv, parts)

	partsN := strings.SplitN(csv, ",", 2) // 최대 2개로 분리
	fmt.Printf("SplitN(..., 2): %v\n", partsN)

	fields := strings.Fields("  하나  둘   셋  ") // 공백 기준 분리
	fmt.Printf("Fields: %v\n", fields)

	joined := strings.Join(parts, " | ")
	fmt.Printf("Join: %q\n", joined)

	// 반복
	fmt.Printf("Repeat: %q\n", strings.Repeat("Go! ", 3))

	// Builder로 효율적인 문자열 조합
	fmt.Println("\n[문자열 빌더]")
	var sb strings.Builder
	for i := 0; i < 5; i++ {
		fmt.Fprintf(&sb, "항목%d", i+1)
		if i < 4 {
			sb.WriteString(", ")
		}
	}
	result := sb.String()
	fmt.Printf("Builder 결과: %q\n", result)
	fmt.Printf("Builder 길이: %d\n", sb.Len())
	sb.Reset() // 초기화
	fmt.Printf("Reset 후 길이: %d\n", sb.Len())
	fmt.Println()

	// ─────────────────────────────────────────
	// 5. strconv 패키지 (타입 변환)
	// ─────────────────────────────────────────
	fmt.Println("--- 5. strconv 패키지 ---")

	// 정수 -> 문자열
	numStr := strconv.Itoa(42)
	fmt.Printf("Itoa(42): %q\n", numStr)

	// 문자열 -> 정수
	num, err := strconv.Atoi("123")
	if err == nil {
		fmt.Printf("Atoi(\"123\"): %d\n", num)
	}
	_, err = strconv.Atoi("abc")
	fmt.Printf("Atoi(\"abc\") 에러: %v\n", err)

	// 더 범용적인 변환
	i64, _ := strconv.ParseInt("FF", 16, 64) // 16진수 파싱
	fmt.Printf("ParseInt(\"FF\", 16): %d\n", i64)

	i64_2, _ := strconv.ParseInt("0b1010", 0, 64) // 자동 진법 감지
	fmt.Printf("ParseInt(\"0b1010\", 0): %d\n", i64_2)

	f64, _ := strconv.ParseFloat("3.14159", 64)
	fmt.Printf("ParseFloat(\"3.14159\"): %.5f\n", f64)

	b1, _ := strconv.ParseBool("true")
	b2, _ := strconv.ParseBool("1")
	b3, _ := strconv.ParseBool("false")
	fmt.Printf("ParseBool: %t, %t, %t\n", b1, b2, b3)

	// 타입 -> 문자열
	fmt.Printf("FormatInt(255, 2): %q (2진수)\n", strconv.FormatInt(255, 2))
	fmt.Printf("FormatInt(255, 16): %q (16진수)\n", strconv.FormatInt(255, 16))
	fmt.Printf("FormatFloat: %q\n", strconv.FormatFloat(3.14159, 'f', 3, 64))
	fmt.Printf("FormatBool(true): %q\n", strconv.FormatBool(true))

	// 문자열에 특수문자 포함 여부 확인
	fmt.Printf("\nQuote: %q\n", strconv.Quote("Hello, \"세계\"!\n"))
	fmt.Println()

	// ─────────────────────────────────────────
	// 6. unicode 패키지
	// ─────────────────────────────────────────
	fmt.Println("--- 6. unicode 패키지 ---")

	testRunes := []rune{'A', 'a', '5', ' ', '한', '!'}
	for _, r := range testRunes {
		fmt.Printf("'%c': IsLetter=%t, IsDigit=%t, IsUpper=%t, IsLower=%t, IsSpace=%t\n",
			r,
			unicode.IsLetter(r),
			unicode.IsDigit(r),
			unicode.IsUpper(r),
			unicode.IsLower(r),
			unicode.IsSpace(r),
		)
	}

	// 대소문자 변환
	fmt.Printf("\nToUpper('a'): %c\n", unicode.ToUpper('a'))
	fmt.Printf("ToLower('A'): %c\n", unicode.ToLower('A'))
	fmt.Println()

	// ─────────────────────────────────────────
	// 7. 문자열 연산 성능 비교
	// ─────────────────────────────────────────
	fmt.Println("--- 7. 효율적인 문자열 조합 ---")

	// 비효율적: + 연산자 반복 (매번 새 문자열 할당)
	inefficient := ""
	for i := 0; i < 5; i++ {
		inefficient += fmt.Sprintf("item%d", i) // 매번 새 문자열 생성
	}
	fmt.Printf("+ 연산 결과: %q\n", inefficient)

	// 효율적: strings.Builder 사용
	var builder strings.Builder
	builder.Grow(50) // 미리 용량 예약 (선택사항)
	for i := 0; i < 5; i++ {
		fmt.Fprintf(&builder, "item%d", i)
	}
	fmt.Printf("Builder 결과: %q\n", builder.String())

	// 효율적: strings.Join 사용
	items := make([]string, 5)
	for i := range items {
		items[i] = fmt.Sprintf("item%d", i)
	}
	fmt.Printf("Join 결과: %q\n", strings.Join(items, ""))
	fmt.Println()

	// ─────────────────────────────────────────
	// 8. 실용 예제: 문자열 처리
	// ─────────────────────────────────────────
	fmt.Println("--- 8. 실용 예제 ---")

	// 단어 수 세기
	sentence := "Go is a statically typed compiled language"
	words := strings.Fields(sentence)
	fmt.Printf("문장: %q\n단어 수: %d\n", sentence, len(words))

	// 회문(palindrome) 검사
	isPalindrome := func(s string) bool {
		runes := []rune(strings.ToLower(s))
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			if runes[i] != runes[j] {
				return false
			}
		}
		return true
	}
	testWords := []string{"level", "hello", "racecar", "golang", "madam"}
	for _, w := range testWords {
		fmt.Printf("isPalindrome(%q): %t\n", w, isPalindrome(w))
	}

	// 이메일 유효성 간단 검사
	emails := []string{"user@example.com", "invalid-email", "name@domain.co.kr"}
	for _, email := range emails {
		parts2 := strings.Split(email, "@")
		valid := len(parts2) == 2 && strings.Contains(parts2[1], ".")
		fmt.Printf("이메일 %q 유효: %t\n", email, valid)
	}

	fmt.Println()
	fmt.Println("=== 완료 ===")
}
