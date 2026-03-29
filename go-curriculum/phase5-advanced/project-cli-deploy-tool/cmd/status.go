// cmd/status.go
// 'status' 서브커맨드: 배포 상태 확인
//
// 사용 예시:
//   deploy-tool status myapp
//   deploy-tool status          (모든 앱)
//   deploy-tool status -o json
package cmd

import (
	"fmt"
	"os"

	"github.com/curriculum/deploy-tool/internal/deployer"
	"github.com/curriculum/deploy-tool/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusWatch bool

// statusCmd는 배포 상태를 확인합니다.
var statusCmd = &cobra.Command{
	Use:   "status [앱이름]",
	Short: "배포 상태를 확인합니다",
	Long: `배포된 애플리케이션의 현재 상태를 조회합니다.
앱 이름을 지정하면 해당 앱만, 생략하면 모든 앱의 상태를 표시합니다.`,

	Example: `  deploy-tool status
  deploy-tool status myapp
  deploy-tool status -o json
  deploy-tool status --watch`,

	Args: cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		appName := ""
		if len(args) > 0 {
			appName = args[0]
		}

		registry := viper.GetString("registry")
		d := deployer.NewDockerDeployer(registry, verbose)
		printer := output.NewPrinter(os.Stdout, viper.GetString("output"))

		if appName != "" {
			// 단일 앱 상태 조회
			status, err := d.Status(appName)
			if err != nil {
				return fmt.Errorf("상태 조회 실패 (%s): %w", appName, err)
			}
			return printer.PrintStatus([]*deployer.AppStatus{status})
		}

		// 전체 앱 상태 조회
		statuses, err := d.ListAll()
		if err != nil {
			return fmt.Errorf("전체 상태 조회 실패: %w", err)
		}

		if len(statuses) == 0 {
			fmt.Println("배포된 애플리케이션이 없습니다.")
			return nil
		}

		return printer.PrintStatus(statuses)
	},
}

func init() {
	statusCmd.Flags().BoolVarP(&statusWatch, "watch", "w", false, "상태를 주기적으로 갱신")
}
