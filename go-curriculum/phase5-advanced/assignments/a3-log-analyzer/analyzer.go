// analyzer.go
// 구조화된 로그 분석기 핵심 로직
//
// TODO: Analyzer 구조체와 메서드들을 구현하세요.
package main

import (
	"time"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// LogEntry는 slog JSON 형식의 단일 로그 엔트리를 표현합니다.
type LogEntry struct {
	Time    time.Time              `json:"time"`
	Level   string                 `json:"level"`
	Msg     string                 `json:"msg"`
	Source  string                 `json:"source,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Extras  map[string]interface{} `json:"-"` // 추가 필드
}

// AnalysisResult는 파일 하나의 분석 결과를 담습니다.
type AnalysisResult struct {
	Filename     string            // 파일 이름
	TotalEntries int               // 총 엔트리 수
	LevelCounts  map[string]int    // 레벨별 개수 (DEBUG/INFO/WARN/ERROR)
	HourlyCounts map[string]int    // 시간별 개수 (키: "2024-01-15T10")
	ErrorPatterns map[string]int   // 오류 메시지별 등장 횟수
	ParseErrors  int               // 파싱 실패 줄 수
}

// ============================================================
// TODO: 아래를 구현하세요
// ============================================================

// Analyzer는 로그 파일을 분석합니다.
type Analyzer struct {
	// TODO: 필요한 필드를 추가하세요
	// 오류 패턴 임계값 (이 값 이상 반복되면 감지)
	PatternThreshold int
}

// NewAnalyzer는 Analyzer 생성자입니다.
func NewAnalyzer(threshold int) *Analyzer {
	// TODO: 구현하세요
	panic("NewAnalyzer: 아직 구현되지 않았습니다")
}

// AnalyzeFile은 단일 로그 파일을 분석합니다.
//
// 구현 요구사항:
//   - bufio.Scanner로 한 줄씩 읽어 메모리를 효율적으로 사용합니다.
//   - 각 줄을 parseLogEntry로 파싱합니다.
//   - 파싱 실패한 줄은 ParseErrors에 카운트하고 계속 진행합니다.
//   - LevelCounts, HourlyCounts, ErrorPatterns를 집계합니다.
//
// HourlyCounts 키 형식: "2024-01-15T10" (RFC3339의 시간 부분까지)
func (a *Analyzer) AnalyzeFile(filename string) (*AnalysisResult, error) {
	// TODO: 구현하세요
	panic("AnalyzeFile: 아직 구현되지 않았습니다")
}

// AnalyzeFiles는 여러 파일을 동시에 분석합니다.
//
// 구현 요구사항:
//   - 각 파일을 별도 고루틴에서 분석합니다.
//   - 모든 고루틴 완료 후 결과를 반환합니다.
//   - 파일 하나의 오류는 다른 파일 분석에 영향을 주지 않습니다.
//   - 결과 순서는 입력 파일 순서와 동일해야 합니다.
//
// 힌트:
//   - sync.WaitGroup 또는 결과 채널 사용
//   - 오류가 있는 파일은 nil 대신 오류 정보를 담아 반환
func (a *Analyzer) AnalyzeFiles(filenames []string) ([]*AnalysisResult, []error) {
	// TODO: 구현하세요
	panic("AnalyzeFiles: 아직 구현되지 않았습니다")
}

// parseLogEntry는 JSON 로그 라인을 LogEntry로 파싱합니다.
//
// 구현 요구사항:
//   - encoding/json으로 파싱합니다.
//   - "time" 필드가 없으면 현재 시각을 사용합니다.
//   - "level" 필드는 대문자로 정규화합니다 (info → INFO).
//   - 빈 줄은 nil, nil을 반환합니다.
func parseLogEntry(line string) (*LogEntry, error) {
	// TODO: 구현하세요
	panic("parseLogEntry: 아직 구현되지 않았습니다")
}
