// report.go
// 분석 결과 보고서 생성
//
// TODO: Report 함수들을 구현하세요.
package main

import (
	"io"
)

// ============================================================
// TODO: 아래를 구현하세요
// ============================================================

// PrintTextReport는 분석 결과를 텍스트 형식으로 w에 출력합니다.
//
// 출력 항목:
//   - 파일명, 총 엔트리 수
//   - 레벨별 분포 (개수 + 비율)
//   - 시간별 분포 (상위 5개)
//   - 오류 패턴 (PatternThreshold 이상인 것)
//   - 파싱 오류 수 (있는 경우)
func PrintTextReport(w io.Writer, result *AnalysisResult, patternThreshold int) error {
	// TODO: 구현하세요
	panic("PrintTextReport: 아직 구현되지 않았습니다")
}

// PrintJSONReport는 분석 결과를 JSON 형식으로 w에 출력합니다.
func PrintJSONReport(w io.Writer, result *AnalysisResult) error {
	// TODO: 구현하세요
	panic("PrintJSONReport: 아직 구현되지 않았습니다")
}
