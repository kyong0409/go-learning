// solution/report.go - 참고 풀이
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// PrintTextReport는 분석 결과를 텍스트 형식으로 출력합니다.
func PrintTextReport(w io.Writer, result *AnalysisResult, patternThreshold int) error {
	fmt.Fprintf(w, "\n=== 로그 분석 결과 ===\n")
	fmt.Fprintf(w, "파일: %s\n", result.Filename)
	fmt.Fprintf(w, "총 엔트리: %d개\n", result.TotalEntries)

	if result.ParseErrors > 0 {
		fmt.Fprintf(w, "파싱 오류: %d줄\n", result.ParseErrors)
	}

	// 레벨별 분포
	fmt.Fprintf(w, "\n레벨별 분포:\n")
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for _, level := range levels {
		count := result.LevelCounts[level]
		if count == 0 {
			continue
		}
		pct := float64(count) / float64(result.TotalEntries) * 100
		fmt.Fprintf(w, "  %-8s %d개 (%.1f%%)\n", level+":", count, pct)
	}

	// 시간별 분포 (상위 5개)
	fmt.Fprintf(w, "\n시간별 분포 (상위 5):\n")
	type hourEntry struct {
		hour  string
		count int
	}
	hours := make([]hourEntry, 0, len(result.HourlyCounts))
	for h, c := range result.HourlyCounts {
		hours = append(hours, hourEntry{h, c})
	}
	sort.Slice(hours, func(i, j int) bool { return hours[i].count > hours[j].count })
	for i, h := range hours {
		if i >= 5 {
			break
		}
		fmt.Fprintf(w, "  %s ~ +1h  %d개\n", h.hour, h.count)
	}

	// 오류 패턴
	hasPattern := false
	for _, count := range result.ErrorPatterns {
		if count >= patternThreshold {
			hasPattern = true
			break
		}
	}

	if hasPattern {
		fmt.Fprintf(w, "\n오류 패턴 (%d회 이상):\n", patternThreshold)
		type patternEntry struct {
			msg   string
			count int
		}
		patterns := make([]patternEntry, 0)
		for msg, count := range result.ErrorPatterns {
			if count >= patternThreshold {
				patterns = append(patterns, patternEntry{msg, count})
			}
		}
		sort.Slice(patterns, func(i, j int) bool { return patterns[i].count > patterns[j].count })
		for _, p := range patterns {
			fmt.Fprintf(w, "  %q  %d회\n", p.msg, p.count)
		}
	}

	return nil
}

// PrintJSONReport는 분석 결과를 JSON 형식으로 출력합니다.
func PrintJSONReport(w io.Writer, result *AnalysisResult) error {
	// HourlyCounts 정렬을 위해 상위 항목 선택
	type reportJSON struct {
		Filename      string         `json:"filename"`
		TotalEntries  int            `json:"total_entries"`
		ParseErrors   int            `json:"parse_errors,omitempty"`
		LevelCounts   map[string]int `json:"level_counts"`
		HourlyCounts  map[string]int `json:"hourly_counts"`
		ErrorPatterns map[string]int `json:"error_patterns"`
	}

	report := reportJSON{
		Filename:      result.Filename,
		TotalEntries:  result.TotalEntries,
		ParseErrors:   result.ParseErrors,
		LevelCounts:   result.LevelCounts,
		HourlyCounts:  result.HourlyCounts,
		ErrorPatterns: result.ErrorPatterns,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, strings.TrimSpace(string(data)))
	return err
}
