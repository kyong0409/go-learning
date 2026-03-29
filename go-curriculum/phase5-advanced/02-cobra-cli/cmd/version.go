// cmd/version.go
// 'version' 서브커맨드 구현
//
// 사용 예시:
//   mycli version
//   mycli version --short
//   mycli version -o json
package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ============================================================
// 버전 정보 (빌드 시 -ldflags로 주입 가능)
// ============================================================

// 빌드 시 주입하는 방법:
//   go build -ldflags "-X github.com/curriculum/cobra-example/cmd.Version=1.2.3" .

var (
	Version   = "0.1.0"           // 의미론적 버전 (semver)
	GitCommit = "unknown"          // git commit 해시
	BuildDate = "unknown"          // 빌드 날짜
)

// ============================================================
// VersionInfo 구조체
// ============================================================

// VersionInfo는 버전 관련 모든 정보를 담습니다.
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// ============================================================
// version 커맨드 정의
// ============================================================

var versionShort bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "버전 정보를 출력합니다",
	Long:  `현재 CLI 도구의 버전, 빌드 정보, Go 런타임 정보를 출력합니다.`,

	Example: `  mycli version
  mycli version --short
  mycli version -o json`,

	RunE: func(cmd *cobra.Command, args []string) error {
		info := VersionInfo{
			Version:   Version,
			GitCommit: GitCommit,
			BuildDate: BuildDate,
			GoVersion: runtime.Version(),
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		}

		// --short 플래그: 버전만 출력
		if versionShort {
			fmt.Println(info.Version)
			return nil
		}

		// 출력 형식: viper에서 읽기 (전역 --output 플래그 반영)
		format := viper.GetString("output")

		switch format {
		case "json":
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return fmt.Errorf("JSON 직렬화 실패: %w", err)
			}
			fmt.Println(string(data))

		case "yaml":
			fmt.Printf("version: %s\n", info.Version)
			fmt.Printf("git_commit: %s\n", info.GitCommit)
			fmt.Printf("build_date: %s\n", info.BuildDate)
			fmt.Printf("go_version: %s\n", info.GoVersion)
			fmt.Printf("os: %s\n", info.OS)
			fmt.Printf("arch: %s\n", info.Arch)

		default:
			// 기본: 표 형식
			fmt.Printf("%-12s %s\n", "버전:", info.Version)
			fmt.Printf("%-12s %s\n", "Git 커밋:", info.GitCommit)
			fmt.Printf("%-12s %s\n", "빌드 날짜:", info.BuildDate)
			fmt.Printf("%-12s %s\n", "Go 버전:", info.GoVersion)
			fmt.Printf("%-12s %s/%s\n", "플랫폼:", info.OS, info.Arch)
		}

		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "버전 번호만 출력")
}
