// cmd/root.go
// deploy-tool 루트 커맨드
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	outputFormat string // table | json | yaml
	verbose      bool
)

// rootCmd는 CLI의 최상위 커맨드입니다.
var rootCmd = &cobra.Command{
	Use:   "deploy-tool",
	Short: "deploy-tool - 컨테이너 애플리케이션 배포 관리 CLI",
	Long: `deploy-tool은 Docker 기반 애플리케이션의 배포, 상태 확인,
로그 스트리밍을 관리하는 CLI 도구입니다.

주요 기능:
  deploy   애플리케이션 배포 (이미지 pull + 컨테이너 실행)
  status   배포 상태 확인
  logs     컨테이너 로그 스트리밍
  config   도구 설정 관리`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] 커맨드: %s\n", cmd.CommandPath())
		}
		return nil
	},
}

// Execute는 루트 커맨드를 실행합니다.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "설정 파일 경로 (기본값: ~/.deploy-tool/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "출력 형식 (table|json|yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "상세 출력 모드")

	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// 서브커맨드 등록
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(configCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, _ := os.UserHomeDir()
		viper.AddConfigPath(home + "/.deploy-tool")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("DEPLOY")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 기본값
	viper.SetDefault("registry", "docker.io")
	viper.SetDefault("default_namespace", "default")
	viper.SetDefault("log_tail", 100)
	viper.SetDefault("output", "table")

	_ = viper.ReadInConfig()
}
