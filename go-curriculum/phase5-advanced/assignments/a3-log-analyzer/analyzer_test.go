// analyzer_test.go
// 로그 분석기 테스트 및 채점
//
// 실행:
//   go test -v
//   go test -v -run TestGrade
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testLogFile = "testdata/app.log"

// ============================================================
// parseLogEntry 테스트 (20점)
// ============================================================

func TestParseLogEntry_Valid(t *testing.T) {
	line := `{"time":"2024-01-15T10:00:01Z","level":"INFO","msg":"서버 시작","port":8080}`

	entry, err := parseLogEntry(line)
	if err != nil {
		t.Fatalf("parseLogEntry 오류: %v", err)
	}
	if entry == nil {
		t.Fatal("parseLogEntry가 nil을 반환했습니다")
	}

	if entry.Level != "INFO" {
		t.Errorf("Level: 기대 INFO, 실제 %s", entry.Level)
	}
	if entry.Msg != "서버 시작" {
		t.Errorf("Msg: 기대 '서버 시작', 실제 %s", entry.Msg)
	}
	if entry.Time.Year() != 2024 {
		t.Errorf("Time.Year: 기대 2024, 실제 %d", entry.Time.Year())
	}
}

func TestParseLogEntry_LevelNormalization(t *testing.T) {
	// 소문자 level도 대문자로 변환되어야 합니다.
	cases := []struct {
		input    string
		expected string
	}{
		{`{"level":"info","msg":"test"}`, "INFO"},
		{`{"level":"warn","msg":"test"}`, "WARN"},
		{`{"level":"ERROR","msg":"test"}`, "ERROR"},
	}

	for _, tc := range cases {
		entry, err := parseLogEntry(tc.input)
		if err != nil || entry == nil {
			t.Errorf("parseLogEntry(%s) 실패: %v", tc.input, err)
			continue
		}
		if entry.Level != tc.expected {
			t.Errorf("Level 정규화: 기대 %s, 실제 %s", tc.expected, entry.Level)
		}
	}
}

func TestParseLogEntry_EmptyLine(t *testing.T) {
	entry, err := parseLogEntry("")
	if err != nil {
		t.Errorf("빈 줄에서 오류 반환: %v", err)
	}
	if entry != nil {
		t.Error("빈 줄에서 nil 반환해야 합니다")
	}
}

func TestParseLogEntry_InvalidJSON(t *testing.T) {
	_, err := parseLogEntry("{invalid json}")
	if err == nil {
		t.Error("잘못된 JSON에서 오류를 반환해야 합니다")
	}
}

// ============================================================
// AnalyzeFile 테스트 (레벨 집계 20점, 시간 집계 20점)
// ============================================================

func TestAnalyzeFile_LevelCounts(t *testing.T) {
	if _, err := os.Stat(testLogFile); err != nil {
		t.Skipf("테스트 파일 없음: %s", testLogFile)
	}

	a := NewAnalyzer(5)
	result, err := a.AnalyzeFile(testLogFile)
	if err != nil {
		t.Fatalf("AnalyzeFile 오류: %v", err)
	}

	if result.TotalEntries == 0 {
		t.Error("TotalEntries가 0 - 파일을 읽지 못했습니다")
	}

	// 레벨 카운트가 있어야 합니다
	if len(result.LevelCounts) == 0 {
		t.Error("LevelCounts가 비어 있습니다")
	}

	// INFO가 가장 많아야 합니다 (테스트 파일 기준)
	if result.LevelCounts["INFO"] == 0 {
		t.Error("INFO 카운트가 0입니다")
	}
	if result.LevelCounts["ERROR"] == 0 {
		t.Error("ERROR 카운트가 0입니다")
	}

	// 합계가 총 엔트리 수와 일치해야 합니다
	total := 0
	for _, count := range result.LevelCounts {
		total += count
	}
	if total != result.TotalEntries-result.ParseErrors {
		t.Errorf("레벨 합계(%d) != TotalEntries(%d) - ParseErrors(%d)",
			total, result.TotalEntries, result.ParseErrors)
	}
}

func TestAnalyzeFile_HourlyCounts(t *testing.T) {
	if _, err := os.Stat(testLogFile); err != nil {
		t.Skipf("테스트 파일 없음: %s", testLogFile)
	}

	a := NewAnalyzer(5)
	result, err := a.AnalyzeFile(testLogFile)
	if err != nil {
		t.Fatalf("AnalyzeFile 오류: %v", err)
	}

	if len(result.HourlyCounts) == 0 {
		t.Error("HourlyCounts가 비어 있습니다")
	}

	// 테스트 파일에는 10시, 11시, 12시 데이터가 있습니다
	hasHour10 := false
	hasHour11 := false
	for key := range result.HourlyCounts {
		if strings.Contains(key, "T10") {
			hasHour10 = true
		}
		if strings.Contains(key, "T11") {
			hasHour11 = true
		}
	}

	if !hasHour10 {
		t.Error("10시 데이터가 HourlyCounts에 없습니다")
	}
	if !hasHour11 {
		t.Error("11시 데이터가 HourlyCounts에 없습니다")
	}
}

func TestAnalyzeFile_ErrorPatterns(t *testing.T) {
	if _, err := os.Stat(testLogFile); err != nil {
		t.Skipf("테스트 파일 없음: %s", testLogFile)
	}

	a := NewAnalyzer(5)
	result, err := a.AnalyzeFile(testLogFile)
	if err != nil {
		t.Fatalf("AnalyzeFile 오류: %v", err)
	}

	// "DB 연결 실패"가 5회 이상 반복됩니다
	dbConnFail := result.ErrorPatterns["DB 연결 실패"]
	if dbConnFail < 5 {
		t.Errorf("'DB 연결 실패' 패턴: 기대 >= 5, 실제 %d", dbConnFail)
	}
}

func TestAnalyzeFile_Streaming(t *testing.T) {
	// 큰 파일도 bufio.Scanner로 처리하는지 간접 확인
	// (파일 전체를 한번에 읽지 않는지)
	if _, err := os.Stat(testLogFile); err != nil {
		t.Skipf("테스트 파일 없음: %s", testLogFile)
	}

	a := NewAnalyzer(5)
	result, err := a.AnalyzeFile(testLogFile)

	if err != nil {
		t.Fatalf("AnalyzeFile 오류: %v", err)
	}
	if result == nil {
		t.Fatal("AnalyzeFile이 nil을 반환했습니다")
	}
	// 결과가 있으면 스트리밍 동작 간접 확인
	if result.TotalEntries == 0 {
		t.Error("엔트리가 0개 — 파일을 읽지 못했습니다")
	}
}

// ============================================================
// AnalyzeFiles (다중 파일) 테스트 (20점)
// ============================================================

func TestAnalyzeFiles_Multiple(t *testing.T) {
	if _, err := os.Stat(testLogFile); err != nil {
		t.Skipf("테스트 파일 없음: %s", testLogFile)
	}

	a := NewAnalyzer(5)
	// 같은 파일을 두 번 분석
	results, errs := a.AnalyzeFiles([]string{testLogFile, testLogFile})

	if len(results) != 2 {
		t.Fatalf("결과 수: 기대 2, 실제 %d", len(results))
	}

	// 각 결과가 유효해야 합니다
	for i, r := range results {
		if r == nil {
			t.Errorf("결과[%d]가 nil입니다 (오류: %v)", i, errs[i])
			continue
		}
		if r.TotalEntries == 0 {
			t.Errorf("결과[%d].TotalEntries가 0입니다", i)
		}
	}
}

func TestAnalyzeFiles_OrderPreservation(t *testing.T) {
	// 결과 순서가 입력 파일 순서와 동일해야 합니다
	if _, err := os.Stat(testLogFile); err != nil {
		t.Skipf("테스트 파일 없음: %s", testLogFile)
	}

	a := NewAnalyzer(5)
	files := []string{testLogFile, "nonexistent.log", testLogFile}
	results, _ := a.AnalyzeFiles(files)

	if len(results) != 3 {
		t.Fatalf("결과 수: 기대 3, 실제 %d", len(results))
	}

	// 첫 번째와 세 번째 결과는 유효해야 합니다
	if results[0] == nil {
		t.Error("results[0]이 nil - 유효한 파일 결과가 있어야 합니다")
	}
	// 두 번째는 존재하지 않는 파일이므로 nil이거나 오류
	// (nil 또는 오류 결과 어느 쪽도 허용)
	if results[2] == nil {
		t.Error("results[2]이 nil - 유효한 파일 결과가 있어야 합니다")
	}
}

// ============================================================
// 채점 함수
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════")
	fmt.Println("  과제 A3: 로그 분석기 채점 결과")
	fmt.Println("═══════════════════════════════════════════")

	logFile := testLogFile
	if _, err := os.Stat(logFile); err != nil {
		fmt.Println("  [오류] testdata/app.log 파일 없음 - 일부 테스트 스킵")
	}

	// JSON 파싱 (20점)
	t.Run("JSON_파싱", func(t *testing.T) {
		line := `{"time":"2024-01-15T10:00:00Z","level":"info","msg":"테스트"}`
		entry, err := parseLogEntry(line)
		if err == nil && entry != nil && entry.Level == "INFO" && entry.Msg == "테스트" {
			score += 20
			fmt.Printf("  ✓ JSON 파싱              20/20점\n")
		} else {
			fmt.Printf("  ✗ JSON 파싱              0/20점\n")
		}
	})

	// 레벨 집계 (20점)
	t.Run("레벨_집계", func(t *testing.T) {
		if _, err := os.Stat(logFile); err != nil {
			fmt.Printf("  - 레벨 집계 (파일 없음)  0/20점\n")
			return
		}
		a := NewAnalyzer(5)
		result, err := a.AnalyzeFile(logFile)
		if err == nil && result != nil &&
			result.LevelCounts["INFO"] > 0 &&
			result.LevelCounts["ERROR"] > 0 &&
			result.TotalEntries > 0 {
			score += 20
			fmt.Printf("  ✓ 레벨별 집계            20/20점\n")
		} else {
			fmt.Printf("  ✗ 레벨별 집계            0/20점\n")
		}
	})

	// 오류 패턴 (20점)
	t.Run("오류_패턴", func(t *testing.T) {
		if _, err := os.Stat(logFile); err != nil {
			fmt.Printf("  - 오류 패턴 (파일 없음)  0/20점\n")
			return
		}
		a := NewAnalyzer(5)
		result, err := a.AnalyzeFile(logFile)
		if err == nil && result != nil && result.ErrorPatterns["DB 연결 실패"] >= 5 {
			score += 20
			fmt.Printf("  ✓ 오류 패턴 탐지         20/20점\n")
		} else {
			fmt.Printf("  ✗ 오류 패턴 탐지         0/20점\n")
		}
	})

	// 시간 집계 (20점)
	t.Run("시간_집계", func(t *testing.T) {
		if _, err := os.Stat(logFile); err != nil {
			fmt.Printf("  - 시간 집계 (파일 없음)  0/20점\n")
			return
		}
		a := NewAnalyzer(5)
		result, err := a.AnalyzeFile(logFile)
		hasHour10, hasHour11 := false, false
		for k := range result.HourlyCounts {
			if strings.Contains(k, "T10") {
				hasHour10 = true
			}
			if strings.Contains(k, "T11") {
				hasHour11 = true
			}
		}
		if err == nil && result != nil && len(result.HourlyCounts) >= 2 && hasHour10 && hasHour11 {
			score += 20
			fmt.Printf("  ✓ 시간별 집계            20/20점\n")
		} else {
			fmt.Printf("  ✗ 시간별 집계            0/20점\n")
		}
	})

	// 다중 파일 (20점)
	t.Run("다중_파일", func(t *testing.T) {
		if _, err := os.Stat(logFile); err != nil {
			fmt.Printf("  - 다중 파일 (파일 없음)  0/20점\n")
			return
		}
		a := NewAnalyzer(5)
		// 유효한 파일과 없는 파일 혼합
		results, _ := a.AnalyzeFiles([]string{logFile, filepath.Join("testdata", "nonexistent.log"), logFile})
		if len(results) == 3 && results[0] != nil && results[2] != nil && results[0].TotalEntries > 0 {
			score += 20
			fmt.Printf("  ✓ 다중 파일 동시 처리    20/20점\n")
		} else {
			fmt.Printf("  ✗ 다중 파일 동시 처리    0/20점\n")
		}
	})

	fmt.Println("───────────────────────────────────────────")
	fmt.Printf("  최종 점수: %d / %d점\n", score, total)
	grade := "F"
	switch {
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	}
	fmt.Printf("  등급: %s\n", grade)
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println()
}
