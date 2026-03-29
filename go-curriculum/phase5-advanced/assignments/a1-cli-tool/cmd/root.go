// cmd/root.go - filesearch 루트 커맨드 (골격)
//
// TODO: 루트 커맨드를 완성하세요.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "filesearch",
	Short: "파일 시스템 검색 도구",
	Long:  `filesearch는 파일 이름 패턴으로 파일을 검색하는 CLI 도구입니다.`,
}

// Execute는 루트 커맨드를 실행합니다.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// TODO: 전역 플래그와 서브커맨드를 등록하세요.
	rootCmd.AddCommand(searchCmd)
}
