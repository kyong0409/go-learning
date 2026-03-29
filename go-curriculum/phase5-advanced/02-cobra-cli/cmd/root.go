// cmd/root.go
// Cobra CLI의 루트 커맨드 정의
//
// 루트 커맨드는 CLI의 최상위 진입점입니다.
// 모든 서브커맨드는 이 루트에 붙습니다.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ============================================================
// 전역 변수
// ============================================================

var (
	// cfgFile은 --config 플래그로 지정하는 설정 파일 경로입니다.
	cfgFile string

	// verbose는 -v/--verbose 플래그 값입니다. (전역 플래그)
	verbose bool

	// outputFormat은 -o/--output 플래그 값입니다. (table/json/yaml)
	outputFormat string
)

// ============================================================
// 루트 커맨드 정의
// ============================================================

// rootCmd는 CLI의 최상위 커맨드입니다.
// 서브커맨드 없이 실행하면 도움말이 표시됩니다.
var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "mycli - Go 학습용 CLI 도구",
	Long: `mycli는 Cobra와 Viper를 사용해 만든 Go CLI 예제 도구입니다.

이 도구는 다음 기능을 보여줍니다:
  - Cobra를 이용한 서브커맨드 구조
  - Viper를 이용한 설정 관리
  - 전역 플래그와 로컬 플래그
  - 다양한 출력 형식 지원

자세한 내용은 각 서브커맨드의 --help를 참고하세요.`,

	// PersistentPreRunE는 모든 서브커맨드 실행 전에 호출됩니다.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] 커맨드: %s\n", cmd.CommandPath())
			fmt.Fprintf(os.Stderr, "[DEBUG] 인자: %v\n", args)
		}
		return nil
	},
}

// ============================================================
// Execute (공개 함수 - main.go에서 호출)
// ============================================================

// Execute는 루트 커맨드를 실행합니다.
// 오류가 발생하면 stderr에 출력 후 os.Exit(1)을 호출합니다.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ============================================================
// init: 커맨드 초기화
// ============================================================

func init() {
	// cobra.OnInitialize: 어떤 커맨드를 실행하든 먼저 호출되는 함수 등록
	cobra.OnInitialize(initConfig)

	// --------------------------------------------------------
	// PersistentFlags: 루트 및 모든 서브커맨드에 적용되는 플래그
	// --------------------------------------------------------
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "설정 파일 경로 (기본값: ./config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "상세 출력 모드")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "출력 형식 (table|json|yaml)")

	// viper와 플래그 연동: viper가 플래그 값을 읽을 수 있게 됩니다.
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// --------------------------------------------------------
	// 서브커맨드 등록
	// --------------------------------------------------------
	rootCmd.AddCommand(greetCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

// ============================================================
// Viper 설정 초기화
// ============================================================

// initConfig는 설정 파일과 환경 변수를 로드합니다.
func initConfig() {
	if cfgFile != "" {
		// --config 플래그로 명시적 파일 지정
		viper.SetConfigFile(cfgFile)
	} else {
		// 기본 위치에서 설정 파일 탐색
		// 현재 디렉터리 → $HOME/.mycli/
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.mycli")
		viper.SetConfigName("config") // config.yaml, config.json, config.toml 탐색
		viper.SetConfigType("yaml")
	}

	// 환경 변수 자동 바인딩
	// MYCLI_OUTPUT=json → viper.GetString("output") == "json"
	viper.SetEnvPrefix("MYCLI")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// 설정 파일 로드 (없어도 오류 아님)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// 파일을 찾았지만 읽기 실패한 경우만 경고
			fmt.Fprintf(os.Stderr, "[경고] 설정 파일 읽기 실패: %v\n", err)
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] 설정 파일 로드됨: %s\n", viper.ConfigFileUsed())
	}
}
