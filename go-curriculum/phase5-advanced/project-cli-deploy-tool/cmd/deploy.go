// cmd/deploy.go
// 'deploy' 서브커맨드: 애플리케이션 배포
//
// 사용 예시:
//   deploy-tool deploy --app myapp --image nginx:latest
//   deploy-tool deploy --app myapp --image nginx:latest --port 8080 --replicas 2
//   deploy-tool deploy --app myapp --image nginx:latest --env KEY=VALUE --env DB=postgres
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/curriculum/deploy-tool/internal/deployer"
	"github.com/curriculum/deploy-tool/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	deployAppName  string
	deployImage    string
	deployPort     int
	deployReplicas int
	deployEnvVars  []string
	deployDryRun   bool
	deployWait     bool
)

// deployCmd는 애플리케이션을 배포합니다.
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "애플리케이션을 배포합니다",
	Long: `지정한 Docker 이미지로 애플리케이션을 배포합니다.

배포 과정:
  1. 이미지 Pull 확인
  2. 기존 컨테이너 중지 (있는 경우)
  3. 새 컨테이너 시작
  4. 헬스 체크 대기 (--wait 옵션 시)`,

	Example: `  deploy-tool deploy --app myapp --image nginx:latest
  deploy-tool deploy --app api --image myapi:v1.2.3 --port 8080
  deploy-tool deploy --app web --image frontend:latest --env NODE_ENV=prod
  deploy-tool deploy --app myapp --image nginx:latest --dry-run`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// 필수 플래그 검증
		if deployAppName == "" {
			return fmt.Errorf("--app 플래그가 필요합니다")
		}
		if deployImage == "" {
			return fmt.Errorf("--image 플래그가 필요합니다")
		}

		// 환경 변수 파싱 (KEY=VALUE 형식)
		envMap, err := parseEnvVars(deployEnvVars)
		if err != nil {
			return err
		}

		// 배포 요청 구성
		req := &deployer.DeployRequest{
			AppName:  deployAppName,
			Image:    deployImage,
			Port:     deployPort,
			Replicas: deployReplicas,
			Env:      envMap,
			DryRun:   deployDryRun,
		}

		// Deployer 생성 (Docker 기반)
		registry := viper.GetString("registry")
		d := deployer.NewDockerDeployer(registry, verbose)

		if deployDryRun {
			fmt.Println("[DRY-RUN] 실제 배포는 실행되지 않습니다.")
			return showDeployPlan(req)
		}

		// 배포 실행
		fmt.Printf("배포 시작: %s (%s)\n", deployAppName, deployImage)

		result, err := d.Deploy(req)
		if err != nil {
			return fmt.Errorf("배포 실패: %w", err)
		}

		// 결과 출력
		printer := output.NewPrinter(os.Stdout, viper.GetString("output"))
		return printer.PrintDeployResult(result)
	},
}

// parseEnvVars는 "KEY=VALUE" 형식의 문자열 슬라이스를 맵으로 변환합니다.
func parseEnvVars(envVars []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, e := range envVars {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("환경 변수 형식 오류 (%q): KEY=VALUE 형식이어야 합니다", e)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
}

// showDeployPlan은 DryRun 시 실행될 계획을 출력합니다.
func showDeployPlan(req *deployer.DeployRequest) error {
	fmt.Println("=== 배포 계획 ===")
	fmt.Printf("  앱 이름:   %s\n", req.AppName)
	fmt.Printf("  이미지:    %s\n", req.Image)
	if req.Port > 0 {
		fmt.Printf("  포트:      %d\n", req.Port)
	}
	fmt.Printf("  레플리카:  %d\n", req.Replicas)
	if len(req.Env) > 0 {
		fmt.Println("  환경 변수:")
		for k, v := range req.Env {
			fmt.Printf("    %s=%s\n", k, v)
		}
	}
	return nil
}

func init() {
	deployCmd.Flags().StringVarP(&deployAppName, "app", "a", "", "애플리케이션 이름 (필수)")
	deployCmd.Flags().StringVarP(&deployImage, "image", "i", "", "Docker 이미지 (필수)")
	deployCmd.Flags().IntVarP(&deployPort, "port", "p", 0, "노출할 포트 (0 = 노출 없음)")
	deployCmd.Flags().IntVar(&deployReplicas, "replicas", 1, "실행할 컨테이너 수")
	deployCmd.Flags().StringArrayVarP(&deployEnvVars, "env", "e", nil, "환경 변수 (KEY=VALUE, 반복 가능)")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "실제 배포 없이 계획만 출력")
	deployCmd.Flags().BoolVar(&deployWait, "wait", false, "헬스 체크 완료까지 대기")
}
