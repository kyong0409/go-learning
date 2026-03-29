// wordcount_test.go: 단어 빈도수 함수 채점 테스트
//
// 실행: go test -v
package wordcount

import (
	"os"
	"path/filepath"
	"testing"
)

// ─────────────────────────────────────────
// 채점 시스템
// ─────────────────────────────────────────

type grader struct {
	passed    int
	total     int
	points    int
	maxPoints int
}

func (g *grader) check(t *testing.T, name string, pts int, fn func(t *testing.T)) {
	t.Helper()
	g.total++
	g.maxPoints += pts
	passed := true
	t.Run(name, func(t *testing.T) {
		fn(t)
		if t.Failed() {
			passed = false
		}
	})
	if passed {
		g.passed++
		g.points += pts
	}
}

func (g *grader) report(t *testing.T) {
	t.Helper()
	t.Logf("\n==================")
	t.Logf("=== GRADE REPORT ===")
	t.Logf("==================")
	t.Logf("Passed: %d/%d", g.passed, g.total)
	t.Logf("Score:  %d/%d", g.points, g.maxPoints)
	t.Logf("==================")
}

// ─────────────────────────────────────────
// 테스트
// ─────────────────────────────────────────

func TestWordCount(t *testing.T) {
	g := &grader{}

	// ─── CountWords (30점) ───
	g.check(t, "CountWords/기본빈도수", 8, func(t *testing.T) {
		counts := CountWords("Go is great. Go is fast!")
		if counts["go"] != 2 {
			t.Errorf("\"go\" 빈도수: got %d, want 2", counts["go"])
		}
		if counts["is"] != 2 {
			t.Errorf("\"is\" 빈도수: got %d, want 2", counts["is"])
		}
		if counts["great"] != 1 {
			t.Errorf("\"great\" 빈도수: got %d, want 1", counts["great"])
		}
		if counts["fast"] != 1 {
			t.Errorf("\"fast\" 빈도수: got %d, want 1", counts["fast"])
		}
	})

	g.check(t, "CountWords/대소문자무시", 7, func(t *testing.T) {
		counts := CountWords("Go GO go gO")
		if counts["go"] != 4 {
			t.Errorf("\"go\" 대소문자 무시: got %d, want 4", counts["go"])
		}
		if len(counts) != 1 {
			t.Errorf("고유 단어 수: got %d, want 1 (모두 \"go\")", len(counts))
		}
	})

	g.check(t, "CountWords/구두점제거", 8, func(t *testing.T) {
		counts := CountWords("hello, world! hello. world?")
		if counts["hello"] != 2 {
			t.Errorf("\"hello\" 구두점 제거: got %d, want 2", counts["hello"])
		}
		if counts["world"] != 2 {
			t.Errorf("\"world\" 구두점 제거: got %d, want 2", counts["world"])
		}
	})

	g.check(t, "CountWords/빈문자열", 4, func(t *testing.T) {
		counts := CountWords("")
		if len(counts) != 0 {
			t.Errorf("빈 문자열: got %d words, want 0", len(counts))
		}
	})

	g.check(t, "CountWords/단일단어", 3, func(t *testing.T) {
		counts := CountWords("hello")
		if counts["hello"] != 1 {
			t.Errorf("단일 단어: got %d, want 1", counts["hello"])
		}
	})

	// ─── TopN (25점) ───
	g.check(t, "TopN/상위2개", 8, func(t *testing.T) {
		counts := map[string]int{"go": 5, "is": 3, "great": 1, "fast": 2}
		top := TopN(counts, 2)
		if len(top) != 2 {
			t.Fatalf("TopN(2): got %d items, want 2", len(top))
		}
		if top[0].Word != "go" || top[0].Count != 5 {
			t.Errorf("1위: got {%s,%d}, want {go,5}", top[0].Word, top[0].Count)
		}
		if top[1].Word != "is" || top[1].Count != 3 {
			t.Errorf("2위: got {%s,%d}, want {is,3}", top[1].Word, top[1].Count)
		}
	})

	g.check(t, "TopN/동점시알파벳순", 9, func(t *testing.T) {
		counts := map[string]int{"banana": 3, "apple": 3, "cherry": 3}
		top := TopN(counts, 3)
		if len(top) != 3 {
			t.Fatalf("TopN 동점: got %d items, want 3", len(top))
		}
		// 동점이면 알파벳 오름차순: apple < banana < cherry
		if top[0].Word != "apple" {
			t.Errorf("동점 1위: got %q, want \"apple\"", top[0].Word)
		}
		if top[1].Word != "banana" {
			t.Errorf("동점 2위: got %q, want \"banana\"", top[1].Word)
		}
		if top[2].Word != "cherry" {
			t.Errorf("동점 3위: got %q, want \"cherry\"", top[2].Word)
		}
	})

	g.check(t, "TopN/N이단어수보다클때", 5, func(t *testing.T) {
		counts := map[string]int{"a": 1, "b": 2}
		top := TopN(counts, 10)
		if len(top) != 2 {
			t.Errorf("TopN n>len: got %d items, want 2", len(top))
		}
	})

	g.check(t, "TopN/빈맵", 3, func(t *testing.T) {
		top := TopN(map[string]int{}, 5)
		if len(top) != 0 {
			t.Errorf("TopN 빈맵: got %d items, want 0", len(top))
		}
	})

	// ─── CountWordsFromFile (20점) ───
	g.check(t, "CountWordsFromFile/샘플파일", 10, func(t *testing.T) {
		counts, err := CountWordsFromFile(filepath.Join("testdata", "sample.txt"))
		if err != nil {
			t.Fatalf("CountWordsFromFile 에러: %v", err)
		}
		// "go"는 샘플 텍스트에서 여러 번 등장
		if counts["go"] < 5 {
			t.Errorf("\"go\" 빈도수: got %d, want >= 5", counts["go"])
		}
		if len(counts) == 0 {
			t.Error("결과 맵이 비어있습니다")
		}
	})

	g.check(t, "CountWordsFromFile/없는파일에러", 10, func(t *testing.T) {
		_, err := CountWordsFromFile("/tmp/nonexistent_wordcount_test.txt")
		if err == nil {
			t.Error("존재하지 않는 파일: 에러가 반환되어야 합니다")
		}
	})

	// ─── TotalWords (10점) ───
	g.check(t, "TotalWords/기본", 7, func(t *testing.T) {
		counts := map[string]int{"go": 3, "is": 2, "great": 1}
		got := TotalWords(counts)
		if got != 6 {
			t.Errorf("TotalWords: got %d, want 6", got)
		}
	})

	g.check(t, "TotalWords/빈맵", 3, func(t *testing.T) {
		got := TotalWords(map[string]int{})
		if got != 0 {
			t.Errorf("TotalWords 빈맵: got %d, want 0", got)
		}
	})

	// ─── UniqueWords (15점) ───
	g.check(t, "UniqueWords/기본", 10, func(t *testing.T) {
		counts := map[string]int{"go": 3, "is": 2, "great": 1}
		got := UniqueWords(counts)
		if got != 3 {
			t.Errorf("UniqueWords: got %d, want 3", got)
		}
	})

	g.check(t, "UniqueWords/빈맵", 5, func(t *testing.T) {
		got := UniqueWords(map[string]int{})
		if got != 0 {
			t.Errorf("UniqueWords 빈맵: got %d, want 0", got)
		}
	})

	// ─── 통합 테스트 (임시 파일 사용) ───
	g.check(t, "통합/파일생성후카운트", 0, func(t *testing.T) {
		// 임시 파일 생성
		tmpFile, err := os.CreateTemp("", "wordcount_test_*.txt")
		if err != nil {
			t.Skip("임시 파일 생성 실패, 건너뜁니다")
		}
		defer os.Remove(tmpFile.Name())

		content := "The quick brown fox jumps over the lazy dog. The fox is quick."
		if _, err := tmpFile.WriteString(content); err != nil {
			t.Skip("파일 쓰기 실패, 건너뜁니다")
		}
		tmpFile.Close()

		counts, err := CountWordsFromFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("CountWordsFromFile 에러: %v", err)
		}
		if counts["the"] != 3 {
			t.Errorf("\"the\" 빈도수: got %d, want 3", counts["the"])
		}
		if counts["fox"] != 2 {
			t.Errorf("\"fox\" 빈도수: got %d, want 2", counts["fox"])
		}
		if counts["quick"] != 2 {
			t.Errorf("\"quick\" 빈도수: got %d, want 2", counts["quick"])
		}
	})

	// ─── 최종 채점 리포트 ───
	g.report(t)
}
