// wordcount 패키지: 단어 빈도수 계산 과제
//
// 각 함수의 본문을 구현하세요.
// 테스트 실행: go test -v
package wordcount

// WordCount: 단어와 그 출현 횟수를 나타내는 구조체
type WordCount struct {
	Word  string // 단어
	Count int    // 출현 횟수
}

// CountWords: 텍스트에서 단어의 빈도수를 계산합니다.
//
// 처리 규칙:
//  1. 공백으로 토큰을 분리합니다.
//  2. 각 토큰을 소문자로 변환합니다.
//  3. 앞뒤의 구두점(.,!?;:"'()[]) 을 제거합니다.
//  4. 빈 문자열이 된 토큰은 무시합니다.
//
// 예:
//
//	CountWords("Go is great. Go is fast!")
//	=> map[string]int{"go": 2, "is": 2, "great": 1, "fast": 1}
func CountWords(text string) map[string]int {
	// TODO: 구현하세요
	// 힌트:
	//   - strings.Fields(text)로 공백 기준 분리
	//   - strings.ToLower(word)로 소문자 변환
	//   - strings.Trim(word, ".,!?;:\"'()[]")로 구두점 제거
	return make(map[string]int)
}

// TopN: 빈도수 상위 n개 단어를 반환합니다.
//
// 정렬 기준:
//  1. 빈도수 내림차순
//  2. 빈도수가 같으면 단어 알파벳 오름차순
//
// n이 단어 수보다 크면 전체를 반환합니다.
//
// 예:
//
//	TopN(map[string]int{"go": 3, "is": 2, "great": 1}, 2)
//	=> []WordCount{{"go", 3}, {"is", 2}}
func TopN(counts map[string]int, n int) []WordCount {
	// TODO: 구현하세요
	// 힌트:
	//   - 맵을 WordCount 슬라이스로 변환
	//   - sort.Slice로 정렬
	//   - n개만 잘라서 반환
	return nil
}

// CountWordsFromFile: 파일을 읽어 단어 빈도수를 계산합니다.
//
// 파일이 없거나 읽기 실패 시 에러를 반환합니다.
// 파일을 성공적으로 읽으면 CountWords를 호출하여 결과를 반환합니다.
//
// 예:
//
//	counts, err := CountWordsFromFile("testdata/sample.txt")
func CountWordsFromFile(filename string) (map[string]int, error) {
	// TODO: 구현하세요
	// 힌트:
	//   - os.ReadFile(filename)으로 파일 읽기
	//   - 에러 발생 시 nil, err 반환
	//   - 성공 시 CountWords(string(data)) 호출
	return nil, nil
}

// TotalWords: 전체 단어 수 (중복 포함)를 반환합니다.
//
// 예:
//
//	TotalWords(map[string]int{"go": 3, "is": 2}) == 5
func TotalWords(counts map[string]int) int {
	// TODO: 구현하세요
	return 0
}

// UniqueWords: 고유 단어 수를 반환합니다.
//
// 예:
//
//	UniqueWords(map[string]int{"go": 3, "is": 2, "great": 1}) == 3
func UniqueWords(counts map[string]int) int {
	// TODO: 구현하세요
	return 0
}
