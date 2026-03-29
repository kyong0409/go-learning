// search_test.go
// 파일 검색 기능 테스트 및 채점
//
// 실행:
//   go test -v
//   go test -v -run TestGrade   (채점만)
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// ============================================================
// 테스트 픽스처
// ============================================================

func setupTestDir(t *testing.T) string {
	t.Helper()
	// testdata 디렉터리 사용
	dir := filepath.Join("testdata")
	if _, err := os.Stat(dir); err != nil {
		t.Skipf("testdata 디렉터리 없음: %v", err)
	}
	return dir
}

// ============================================================
// 기본 검색 테스트 (20점)
// ============================================================

func TestSearchFiles_BasicGlob(t *testing.T) {
	dir := setupTestDir(t)

	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   "*.go",
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("SearchFiles 오류: %v", err)
	}

	// testdata에 .go 파일이 있어야 합니다.
	if len(results) == 0 {
		t.Error("*.go 검색 결과가 0개 - testdata에 .go 파일이 있어야 합니다")
	}

	// 모든 결과가 .go 파일인지 확인
	for _, r := range results {
		if filepath.Ext(r.Path) != ".go" {
			t.Errorf("*.go 결과에 비-.go 파일 포함: %s", r.Path)
		}
	}
}

func TestSearchFiles_ExactName(t *testing.T) {
	dir := setupTestDir(t)

	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   "main.go",
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("SearchFiles 오류: %v", err)
	}

	for _, r := range results {
		base := filepath.Base(r.Path)
		if base != "main.go" {
			t.Errorf("main.go 검색에 다른 파일 포함: %s", r.Path)
		}
	}
}

// ============================================================
// 재귀 탐색 테스트 (15점)
// ============================================================

func TestSearchFiles_NonRecursive(t *testing.T) {
	dir := setupTestDir(t)

	// 비재귀: 최상위 디렉터리만 검색
	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   "*",
		Recursive: false,
	})
	if err != nil {
		t.Fatalf("SearchFiles 오류: %v", err)
	}

	// 모든 결과가 rootDir 직하위여야 함
	for _, r := range results {
		parent := filepath.Dir(r.Path)
		rel, _ := filepath.Rel(dir, parent)
		if rel != "." {
			t.Errorf("비재귀 탐색에서 하위 디렉터리 파일 포함: %s", r.Path)
		}
	}
}

func TestSearchFiles_Recursive(t *testing.T) {
	dir := setupTestDir(t)

	nonRecursive, _ := SearchFiles(dir, SearchOptions{Pattern: "*", Recursive: false})
	recursive, _ := SearchFiles(dir, SearchOptions{Pattern: "*", Recursive: true})

	if len(recursive) <= len(nonRecursive) {
		t.Errorf("재귀 결과(%d)가 비재귀 결과(%d)보다 많아야 합니다", len(recursive), len(nonRecursive))
	}
}

// ============================================================
// 정규식 테스트 (15점)
// ============================================================

func TestSearchFiles_Regex(t *testing.T) {
	dir := setupTestDir(t)

	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   `.*_test\.go$`,
		IsRegex:   true,
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("SearchFiles 정규식 오류: %v", err)
	}

	for _, r := range results {
		base := filepath.Base(r.Path)
		if len(base) < 8 || base[len(base)-8:] != "_test.go" {
			t.Errorf("정규식 *_test.go에 일치하지 않는 파일: %s", r.Path)
		}
	}
}

func TestSearchFiles_InvalidRegex(t *testing.T) {
	dir := setupTestDir(t)

	_, err := SearchFiles(dir, SearchOptions{
		Pattern: `[invalid`,
		IsRegex: true,
	})

	if err == nil {
		t.Error("잘못된 정규식에서 오류를 반환해야 합니다")
	}
}

// ============================================================
// 크기 필터 테스트 (15점)
// ============================================================

func TestSearchFiles_MinSize(t *testing.T) {
	dir := setupTestDir(t)

	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   "*",
		Recursive: true,
		MinSize:   100, // 100바이트 이상
	})
	if err != nil {
		t.Fatalf("SearchFiles 오류: %v", err)
	}

	for _, r := range results {
		if r.Size < 100 {
			t.Errorf("MinSize=100 결과에 작은 파일 포함: %s (크기: %d)", r.Path, r.Size)
		}
	}
}

func TestSearchFiles_MaxSize(t *testing.T) {
	dir := setupTestDir(t)

	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   "*",
		Recursive: true,
		MaxSize:   500, // 500바이트 이하
	})
	if err != nil {
		t.Fatalf("SearchFiles 오류: %v", err)
	}

	for _, r := range results {
		if r.Size > 500 {
			t.Errorf("MaxSize=500 결과에 큰 파일 포함: %s (크기: %d)", r.Path, r.Size)
		}
	}
}

// ============================================================
// 확장자 필터 테스트 (15점)
// ============================================================

func TestSearchFiles_ExtFilter(t *testing.T) {
	dir := setupTestDir(t)

	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   "*",
		Recursive: true,
		Exts:      []string{".go"},
	})
	if err != nil {
		t.Fatalf("SearchFiles 오류: %v", err)
	}

	for _, r := range results {
		if filepath.Ext(r.Path) != ".go" {
			t.Errorf("확장자 필터 .go에 다른 파일 포함: %s", r.Path)
		}
	}
}

func TestSearchFiles_MultipleExts(t *testing.T) {
	dir := setupTestDir(t)

	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   "*",
		Recursive: true,
		Exts:      []string{".go", ".md"},
	})
	if err != nil {
		t.Fatalf("SearchFiles 오류: %v", err)
	}

	for _, r := range results {
		ext := filepath.Ext(r.Path)
		if ext != ".go" && ext != ".md" {
			t.Errorf("확장자 필터 .go,.md에 다른 파일 포함: %s (%s)", r.Path, ext)
		}
	}
}

// ============================================================
// MaxResult 테스트
// ============================================================

func TestSearchFiles_MaxResult(t *testing.T) {
	dir := setupTestDir(t)

	results, err := SearchFiles(dir, SearchOptions{
		Pattern:   "*",
		Recursive: true,
		MaxResult: 2,
	})
	if err != nil {
		t.Fatalf("SearchFiles 오류: %v", err)
	}

	if len(results) > 2 {
		t.Errorf("MaxResult=2인데 %d개 반환됨", len(results))
	}
}

// ============================================================
// 채점 함수 (TestGrade)
// ============================================================

func TestGrade(t *testing.T) {
	dir := setupTestDir(t)
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════")
	fmt.Println("  과제 A1: 파일 검색 CLI 도구 채점 결과")
	fmt.Println("═══════════════════════════════════════════")

	// 기본 검색 (20점)
	t.Run("기본_검색", func(t *testing.T) {
		results, err := SearchFiles(dir, SearchOptions{Pattern: "*.go", Recursive: true})
		if err == nil && len(results) > 0 {
			allGo := true
			for _, r := range results {
				if filepath.Ext(r.Path) != ".go" {
					allGo = false
				}
			}
			if allGo {
				score += 20
				fmt.Printf("  ✓ 기본 패턴 검색        20/20점\n")
			} else {
				fmt.Printf("  ✗ 기본 패턴 검색 (확장자 필터 오류)  0/20점\n")
			}
		} else {
			fmt.Printf("  ✗ 기본 패턴 검색        0/20점\n")
		}
	})

	// 재귀 탐색 (15점)
	t.Run("재귀_탐색", func(t *testing.T) {
		nr, _ := SearchFiles(dir, SearchOptions{Pattern: "*", Recursive: false})
		r, _ := SearchFiles(dir, SearchOptions{Pattern: "*", Recursive: true})
		if len(r) > len(nr) {
			score += 15
			fmt.Printf("  ✓ 재귀 탐색              15/15점\n")
		} else {
			fmt.Printf("  ✗ 재귀 탐색              0/15점\n")
		}
	})

	// 정규식 (15점)
	t.Run("정규식", func(t *testing.T) {
		results, err := SearchFiles(dir, SearchOptions{Pattern: `.*\.go$`, IsRegex: true, Recursive: true})
		_, errInvalid := SearchFiles(dir, SearchOptions{Pattern: `[invalid`, IsRegex: true})
		if err == nil && len(results) > 0 && errInvalid != nil {
			score += 15
			fmt.Printf("  ✓ 정규식 지원            15/15점\n")
		} else {
			fmt.Printf("  ✗ 정규식 지원            0/15점\n")
		}
	})

	// 크기 필터 (15점)
	t.Run("크기_필터", func(t *testing.T) {
		results, err := SearchFiles(dir, SearchOptions{Pattern: "*", Recursive: true, MinSize: 1})
		ok := err == nil
		for _, r := range results {
			if r.Size < 1 {
				ok = false
			}
		}
		if ok {
			score += 15
			fmt.Printf("  ✓ 파일 크기 필터         15/15점\n")
		} else {
			fmt.Printf("  ✗ 파일 크기 필터         0/15점\n")
		}
	})

	// 확장자 필터 (15점)
	t.Run("확장자_필터", func(t *testing.T) {
		results, err := SearchFiles(dir, SearchOptions{Pattern: "*", Recursive: true, Exts: []string{".go"}})
		ok := err == nil
		for _, r := range results {
			if filepath.Ext(r.Path) != ".go" {
				ok = false
			}
		}
		if ok && len(results) > 0 {
			score += 15
			fmt.Printf("  ✓ 확장자 필터            15/15점\n")
		} else {
			fmt.Printf("  ✗ 확장자 필터            0/15점\n")
		}
	})

	// 출력 형식 - MaxResult 동작으로 간접 확인 (20점)
	t.Run("MaxResult", func(t *testing.T) {
		results, err := SearchFiles(dir, SearchOptions{Pattern: "*", Recursive: true, MaxResult: 1})
		if err == nil && len(results) <= 1 {
			score += 20
			fmt.Printf("  ✓ MaxResult 제한         20/20점\n")
		} else {
			fmt.Printf("  ✗ MaxResult 제한         0/20점\n")
		}
	})

	// 결과 정렬 보너스 확인
	results, _ := SearchFiles(dir, SearchOptions{Pattern: "*", Recursive: true})
	paths := make([]string, len(results))
	for i, r := range results {
		paths[i] = r.Path
	}
	sorted := make([]string, len(paths))
	copy(sorted, paths)
	sort.Strings(sorted)

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
	fmt.Println("═══════════════════════════════════════════\n")
}
