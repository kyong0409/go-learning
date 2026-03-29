// solution/analyzer.go - 참고 풀이
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// LogEntry는 slog JSON 형식의 단일 로그 엔트리입니다.
type LogEntry struct {
	Time   time.Time              `json:"time"`
	Level  string                 `json:"level"`
	Msg    string                 `json:"msg"`
	Source string                 `json:"source,omitempty"`
	Error  string                 `json:"error,omitempty"`
	Extras map[string]interface{} `json:"-"`
}

// AnalysisResult는 파일 분석 결과입니다.
type AnalysisResult struct {
	Filename      string
	TotalEntries  int
	LevelCounts   map[string]int
	HourlyCounts  map[string]int
	ErrorPatterns map[string]int
	ParseErrors   int
}

// Analyzer는 로그 파일을 분석합니다.
type Analyzer struct {
	PatternThreshold int
}

// NewAnalyzer는 Analyzer 생성자입니다.
func NewAnalyzer(threshold int) *Analyzer {
	if threshold <= 0 {
		threshold = 5
	}
	return &Analyzer{PatternThreshold: threshold}
}

// AnalyzeFile은 단일 로그 파일을 분석합니다.
func (a *Analyzer) AnalyzeFile(filename string) (*AnalysisResult, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패 (%s): %w", filename, err)
	}
	defer f.Close()

	result := &AnalysisResult{
		Filename:      filename,
		LevelCounts:   make(map[string]int),
		HourlyCounts:  make(map[string]int),
		ErrorPatterns: make(map[string]int),
	}

	scanner := bufio.NewScanner(f)
	// 큰 줄 처리를 위해 버퍼 확장
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		result.TotalEntries++

		entry, err := parseLogEntry(line)
		if err != nil {
			result.ParseErrors++
			continue
		}
		if entry == nil {
			result.TotalEntries-- // 빈 줄은 카운트 제외
			continue
		}

		// 레벨 집계
		result.LevelCounts[entry.Level]++

		// 시간별 집계 (1시간 단위, 키: "2006-01-02T15")
		hourKey := entry.Time.UTC().Format("2006-01-02T15")
		result.HourlyCounts[hourKey]++

		// 오류 패턴 집계 (ERROR 레벨만)
		if entry.Level == "ERROR" && entry.Msg != "" {
			result.ErrorPatterns[entry.Msg]++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("파일 읽기 오류: %w", err)
	}

	return result, nil
}

// AnalyzeFiles는 여러 파일을 동시에 분석합니다.
func (a *Analyzer) AnalyzeFiles(filenames []string) ([]*AnalysisResult, []error) {
	results := make([]*AnalysisResult, len(filenames))
	errs := make([]error, len(filenames))

	var wg sync.WaitGroup

	for i, filename := range filenames {
		wg.Add(1)
		go func(idx int, fname string) {
			defer wg.Done()
			result, err := a.AnalyzeFile(fname)
			results[idx] = result
			errs[idx] = err
		}(i, filename)
	}

	wg.Wait()
	return results, errs
}

// parseLogEntry는 JSON 로그 라인을 파싱합니다.
func parseLogEntry(line string) (*LogEntry, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
	}

	entry := &LogEntry{}

	// time 필드
	if t, ok := raw["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			entry.Time = parsed
		}
	}
	if entry.Time.IsZero() {
		entry.Time = time.Now()
	}

	// level 필드 (대문자 정규화)
	if l, ok := raw["level"].(string); ok {
		entry.Level = strings.ToUpper(l)
	}

	// msg 필드
	if m, ok := raw["msg"].(string); ok {
		entry.Msg = m
	}

	// error 필드
	if e, ok := raw["error"].(string); ok {
		entry.Error = e
	}

	return entry, nil
}

func main() {
	fmt.Println("참고 풀이 - 직접 실행하지 마세요")
}
