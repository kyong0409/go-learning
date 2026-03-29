// solution/search.go
// 과제 A1 참고 풀이 - SearchFiles 구현
//
// 주의: 막혔을 때만 참고하세요!
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// SearchOptions는 검색 조건을 담는 구조체입니다.
type SearchOptions struct {
	Pattern   string
	IsRegex   bool
	Recursive bool
	MaxSize   int64
	MinSize   int64
	Exts      []string
	MaxResult int
}

// FileResult는 검색 결과 파일 하나를 표현합니다.
type FileResult struct {
	Path     string
	Size     int64
	Modified time.Time
	Mode     os.FileMode
}

// SearchFiles는 rootDir에서 조건에 맞는 파일을 검색합니다.
func SearchFiles(rootDir string, opts SearchOptions) ([]FileResult, error) {
	// 정규식 미리 컴파일 (IsRegex인 경우)
	var re *regexp.Regexp
	if opts.IsRegex {
		var err error
		re, err = regexp.Compile(opts.Pattern)
		if err != nil {
			return nil, fmt.Errorf("정규식 컴파일 오류: %w", err)
		}
	}

	var results []FileResult

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // 접근 불가 파일 무시
		}

		// MaxResult 초과 시 탐색 중단
		if opts.MaxResult > 0 && len(results) >= opts.MaxResult {
			return filepath.SkipAll
		}

		// 디렉터리 처리
		if d.IsDir() {
			// 루트가 아닌 디렉터리이고 비재귀면 스킵
			if !opts.Recursive && path != rootDir {
				return filepath.SkipDir
			}
			return nil
		}

		base := filepath.Base(path)

		// 패턴 매칭
		matched := false
		if opts.IsRegex {
			matched = re.MatchString(base)
		} else {
			var err error
			matched, err = filepath.Match(opts.Pattern, base)
			if err != nil {
				return fmt.Errorf("패턴 오류: %w", err)
			}
		}

		if !matched {
			return nil
		}

		// 파일 정보 조회
		info, err := d.Info()
		if err != nil {
			return nil
		}

		// 필터 적용
		if !matchesFilter(info, opts) {
			return nil
		}

		results = append(results, FileResult{
			Path:     path,
			Size:     info.Size(),
			Modified: info.ModTime(),
			Mode:     info.Mode(),
		})

		return nil
	}

	if err := filepath.WalkDir(rootDir, walkFn); err != nil {
		return nil, fmt.Errorf("디렉터리 탐색 실패: %w", err)
	}

	return results, nil
}

// matchesFilter는 파일이 크기/확장자 조건에 맞는지 확인합니다.
func matchesFilter(info os.FileInfo, opts SearchOptions) bool {
	// 최소 크기 확인
	if opts.MinSize > 0 && info.Size() < opts.MinSize {
		return false
	}

	// 최대 크기 확인
	if opts.MaxSize > 0 && info.Size() > opts.MaxSize {
		return false
	}

	// 확장자 필터
	if len(opts.Exts) > 0 {
		ext := filepath.Ext(info.Name())
		found := false
		for _, e := range opts.Exts {
			if e == ext {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func main() {}
