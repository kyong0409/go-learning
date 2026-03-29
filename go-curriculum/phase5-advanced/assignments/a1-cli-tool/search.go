// search.go
// 파일 검색 핵심 로직
//
// TODO: 아래 함수들을 완성하세요.
// 각 함수의 시그니처와 주석을 참고해 구현합니다.
package main

import (
	"os"
	"time"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// SearchOptions는 검색 조건을 담는 구조체입니다.
type SearchOptions struct {
	Pattern   string   // 검색 패턴 (glob 또는 정규식)
	IsRegex   bool     // true면 정규식, false면 glob 패턴
	Recursive bool     // true면 하위 디렉터리 포함
	MaxSize   int64    // 최대 파일 크기 (바이트, 0 = 제한 없음)
	MinSize   int64    // 최소 파일 크기 (바이트, 0 = 제한 없음)
	Exts      []string // 허용 확장자 목록 (예: [".go", ".md"], 빈 슬라이스 = 전체)
	MaxResult int      // 최대 결과 수 (0 = 제한 없음)
}

// FileResult는 검색 결과 파일 하나를 표현합니다.
type FileResult struct {
	Path     string      // 파일 경로
	Size     int64       // 파일 크기 (바이트)
	Modified time.Time   // 마지막 수정 시각
	Mode     os.FileMode // 파일 권한
}

// ============================================================
// TODO: 아래 함수들을 구현하세요
// ============================================================

// SearchFiles는 rootDir 아래에서 조건에 맞는 파일을 검색합니다.
//
// 구현 요구사항:
//   - opts.Recursive가 true면 모든 하위 디렉터리를 탐색합니다.
//   - opts.IsRegex가 true면 regexp 패키지로 패턴을 컴파일합니다.
//   - opts.IsRegex가 false면 filepath.Match로 glob 패턴을 사용합니다.
//   - 패턴은 파일 이름(베이스)에만 적용합니다 (경로 전체 아님).
//   - opts.MaxResult > 0이면 결과가 MaxResult개를 초과하면 탐색을 중단합니다.
//   - 디렉터리는 결과에 포함하지 않습니다.
//
// 힌트:
//   - filepath.WalkDir 또는 os.ReadDir + 재귀 사용
//   - matchesFilter 헬퍼를 먼저 구현하세요
func SearchFiles(rootDir string, opts SearchOptions) ([]FileResult, error) {
	// TODO: 구현하세요
	panic("SearchFiles: 아직 구현되지 않았습니다")
}

// matchesFilter는 파일이 검색 조건에 맞는지 확인합니다.
//
// 확인 항목:
//   - 파일 크기가 MinSize 이상인지 (MinSize > 0인 경우)
//   - 파일 크기가 MaxSize 이하인지 (MaxSize > 0인 경우)
//   - 확장자가 Exts 목록에 포함되는지 (Exts가 비어있지 않은 경우)
//
// 참고: 패턴 매칭은 SearchFiles에서 처리하므로 여기서는 하지 않습니다.
func matchesFilter(info os.FileInfo, opts SearchOptions) bool {
	// TODO: 구현하세요
	panic("matchesFilter: 아직 구현되지 않았습니다")
}
