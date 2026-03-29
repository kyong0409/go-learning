// solution/wordcount.go: 단어 빈도수 과제 참고 풀이
//
// 이 파일은 참고용입니다. 먼저 스스로 구현해 보세요!
package wordcount

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// WordCount: 단어와 그 출현 횟수
type WordCount struct {
	Word  string
	Count int
}

// CountWords: 텍스트에서 단어 빈도수 계산
func CountWords(text string) map[string]int {
	counts := make(map[string]int)
	// 공백으로 분리
	tokens := strings.Fields(text)
	for _, token := range tokens {
		// 소문자 변환
		word := strings.ToLower(token)
		// 앞뒤 구두점 제거
		word = strings.Trim(word, `.,!?;:"'()[]`)
		// 빈 문자열 무시
		if word == "" {
			continue
		}
		counts[word]++
	}
	return counts
}

// TopN: 빈도수 상위 N개 단어 반환
func TopN(counts map[string]int, n int) []WordCount {
	// 맵을 슬라이스로 변환
	list := make([]WordCount, 0, len(counts))
	for word, count := range counts {
		list = append(list, WordCount{word, count})
	}

	// 정렬: 빈도수 내림차순, 동점 시 알파벳 오름차순
	sort.Slice(list, func(i, j int) bool {
		if list[i].Count != list[j].Count {
			return list[i].Count > list[j].Count
		}
		return list[i].Word < list[j].Word
	})

	// 상위 N개만 반환
	if n > len(list) {
		n = len(list)
	}
	return list[:n]
}

// CountWordsFromFile: 파일을 읽어 단어 빈도수 계산
func CountWordsFromFile(filename string) (map[string]int, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("파일 읽기 실패 %q: %w", filename, err)
	}
	return CountWords(string(data)), nil
}

// TotalWords: 전체 단어 수 (중복 포함)
func TotalWords(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}

// UniqueWords: 고유 단어 수
func UniqueWords(counts map[string]int) int {
	return len(counts)
}
