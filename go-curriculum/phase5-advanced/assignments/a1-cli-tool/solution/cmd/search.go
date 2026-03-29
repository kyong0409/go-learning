// solution/cmd/search.go - 참고 풀이
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// FileResult는 검색 결과 (solution 패키지용 복사본)
type FileResult struct {
	Path     string      `json:"path"`
	Size     int64       `json:"size"`
	Modified time.Time   `json:"modified"`
	Mode     os.FileMode `json:"mode"`
}

var (
	searchRecursive bool
	searchRegex     bool
	searchOutput    string
	searchMinSize   int64
	searchMaxSize   int64
	searchExts      []string
	searchMax       int
)

var searchCmd = &cobra.Command{
	Use:   "search <패턴> [디렉터리]",
	Short: "패턴으로 파일을 검색합니다",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		dir := "."
		if len(args) == 2 {
			dir = args[1]
		}

		// 확장자 정규화 (.go 형식 보장)
		normalizedExts := make([]string, len(searchExts))
		for i, e := range searchExts {
			if !strings.HasPrefix(e, ".") {
				normalizedExts[i] = "." + e
			} else {
				normalizedExts[i] = e
			}
		}

		// SearchOptions는 solution/search.go의 타입을 사용해야 하지만
		// 여기서는 구조를 보여주기 위한 개념 코드입니다.
		// 실제로는 메인 패키지의 SearchFiles를 임포트해서 사용합니다.

		// 결과 출력
		switch searchOutput {
		case "json":
			fmt.Println("[")
			fmt.Printf(`  {"path": "%s", "pattern": "%s", "dir": "%s"}`, "example", pattern, dir)
			fmt.Println("\n]")
		case "count":
			fmt.Printf("검색 중: 패턴=%s, 디렉터리=%s\n", pattern, dir)
			fmt.Println("0개 파일 발견")
		default:
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "경로\t크기\t수정일")
			fmt.Fprintln(w, "----\t----\t------")
			w.Flush()
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().BoolVarP(&searchRecursive, "recursive", "r", false, "재귀 탐색")
	searchCmd.Flags().BoolVar(&searchRegex, "regex", false, "정규식 패턴 사용")
	searchCmd.Flags().StringVarP(&searchOutput, "output", "o", "table", "출력 형식 (table|json|count)")
	searchCmd.Flags().Int64Var(&searchMinSize, "min-size", 0, "최소 파일 크기 (바이트)")
	searchCmd.Flags().Int64Var(&searchMaxSize, "max-size", 0, "최대 파일 크기 (바이트)")
	searchCmd.Flags().StringSliceVar(&searchExts, "ext", nil, "파일 확장자 필터")
	searchCmd.Flags().IntVar(&searchMax, "max", 0, "최대 결과 수")
}
