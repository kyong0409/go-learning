// cmd/logs.go
// 'logs' 서브커맨드: 컨테이너 로그 스트리밍
//
// 사용 예시:
//   deploy-tool logs myapp
//   deploy-tool logs myapp --follow
//   deploy-tool logs myapp --tail 50
package cmd

import (
	"fmt"
	"os"

	"github.com/curriculum/deploy-tool/internal/deployer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logsFollow bool
	logsTail   int
	logsSince  string
)

// logsCmd는 컨테이너 로그를 출력합니다.
var logsCmd = &cobra.Command{
	Use:   "logs <앱이름>",
	Short: "컨테이너 로그를 출력합니다",
	Long: `실행 중인 컨테이너의 로그를 출력합니다.
--follow 플래그로 실시간 로그 스트리밍이 가능합니다.`,

	Example: `  deploy-tool logs myapp
  deploy-tool logs myapp --follow
  deploy-tool logs myapp --tail 200
  deploy-tool logs myapp --since 1h`,

	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		// viper에서 기본 tail 줄 수 읽기
		tail := logsTail
		if !cmd.Flags().Changed("tail") {
			tail = viper.GetInt("log_tail")
		}

		registry := viper.GetString("registry")
		d := deployer.NewDockerDeployer(registry, verbose)

		opts := &deployer.LogOptions{
			Follow: logsFollow,
			Tail:   tail,
			Since:  logsSince,
		}

		if logsFollow {
			fmt.Printf("%s 로그 스트리밍 중 (Ctrl+C로 종료)...\n\n", appName)
		}

		return d.Logs(os.Stdout, appName, opts)
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "로그 실시간 스트리밍")
	logsCmd.Flags().IntVar(&logsTail, "tail", 100, "마지막 N줄만 출력")
	logsCmd.Flags().StringVar(&logsSince, "since", "", "특정 시간 이후 로그 (예: 1h, 30m, 2024-01-01T00:00:00)")
}
