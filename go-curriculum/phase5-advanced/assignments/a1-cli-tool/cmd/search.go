// cmd/search.go - search 서브커맨드 (골격)
//
// TODO: 아래 searchCmd를 완성하세요.
// - 플래그를 정의하고
// - SearchFiles를 호출하고
// - 결과를 지정한 형식으로 출력하세요.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// TODO: 플래그 변수를 선언하세요.
// 예:
//   var searchRecursive bool
//   var searchRegex bool
//   ...

var searchCmd = &cobra.Command{
	Use:   "search <패턴> [디렉터리]",
	Short: "패턴으로 파일을 검색합니다",
	Long: `지정한 패턴으로 파일을 검색합니다.
패턴은 glob (기본) 또는 정규식 (--regex)을 사용할 수 있습니다.`,

	Example: `  filesearch search "*.go" .
  filesearch search --regex ".*_test\.go$" ./src
  filesearch search --recursive --output json "main" .`,

	Args: cobra.RangeArgs(1, 2),

	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: 구현하세요.
		// 1. args에서 패턴과 디렉터리를 추출하세요.
		// 2. SearchOptions를 구성하세요.
		// 3. SearchFiles를 호출하세요.
		// 4. 결과를 출력하세요 (table/json/count).
		fmt.Println("TODO: search 커맨드를 구현하세요")
		return nil
	},
}

func init() {
	// TODO: 플래그를 등록하세요.
	// searchCmd.Flags().BoolVarP(&searchRecursive, "recursive", "r", false, "재귀 탐색")
	// searchCmd.Flags().BoolVar(&searchRegex, "regex", false, "정규식 패턴 사용")
	// searchCmd.Flags().StringP("output", "o", "table", "출력 형식 (table|json|count)")
	// searchCmd.Flags().Int64("min-size", 0, "최소 파일 크기 (바이트)")
	// searchCmd.Flags().Int64("max-size", 0, "최대 파일 크기 (바이트)")
	// searchCmd.Flags().StringSlice("ext", nil, "파일 확장자 필터 (.go,.md)")
	// searchCmd.Flags().Int("max", 0, "최대 결과 수")
}
