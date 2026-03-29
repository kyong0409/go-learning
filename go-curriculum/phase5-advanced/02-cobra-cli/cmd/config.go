// cmd/config.go
// 'config' 서브커맨드 구현 (Viper 설정 관리 데모)
//
// 사용 예시:
//   mycli config show
//   mycli config get greet.default_name
//   mycli config set greet.default_name 홍길동
//   mycli config path
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ============================================================
// config 커맨드 (부모)
// ============================================================

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "설정을 관리합니다",
	Long: `config 커맨드는 Viper를 통한 설정 관리 기능을 제공합니다.

Viper는 다음 우선순위로 설정을 읽습니다 (높은 순서):
  1. 명시적 Set() 호출
  2. 플래그 (cobra flags)
  3. 환경 변수 (MYCLI_*)
  4. 설정 파일 (config.yaml)
  5. 기본값 (SetDefault)`,

	// 서브커맨드 없이 실행하면 도움말 표시
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// ============================================================
// config show 서브커맨드
// ============================================================

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "현재 설정 전체를 출력합니다",
	RunE: func(cmd *cobra.Command, args []string) error {
		// viper의 모든 설정 키-값을 JSON으로 출력
		allSettings := viper.AllSettings()

		data, err := json.MarshalIndent(allSettings, "", "  ")
		if err != nil {
			return fmt.Errorf("설정 직렬화 실패: %w", err)
		}

		fmt.Println("현재 설정:")
		fmt.Println(string(data))

		configFile := viper.ConfigFileUsed()
		if configFile != "" {
			fmt.Printf("\n설정 파일: %s\n", configFile)
		} else {
			fmt.Println("\n설정 파일: 없음 (기본값 사용)")
		}

		return nil
	},
}

// ============================================================
// config get 서브커맨드
// ============================================================

var configGetCmd = &cobra.Command{
	Use:   "get <키>",
	Short: "특정 설정 값을 조회합니다",
	Args:  cobra.ExactArgs(1), // 인자 1개 필수

	Example: `  mycli config get greet.default_name
  mycli config get output`,

	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		if !viper.IsSet(key) {
			return fmt.Errorf("설정 키를 찾을 수 없습니다: %s", key)
		}

		value := viper.Get(key)
		fmt.Printf("%s = %v\n", key, value)
		return nil
	},
}

// ============================================================
// config set 서브커맨드
// ============================================================

var configSetCmd = &cobra.Command{
	Use:   "set <키> <값>",
	Short: "설정 값을 변경합니다 (현재 세션에만 적용)",
	Args:  cobra.ExactArgs(2), // 인자 2개 필수

	Example: `  mycli config set greet.default_name 홍길동
  mycli config set output json`,

	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		// 현재 세션의 viper에 설정 적용
		viper.Set(key, value)

		fmt.Printf("설정됨: %s = %s\n", key, value)
		fmt.Println("(주의: 이 변경은 현재 세션에만 적용됩니다. 파일에 저장하려면 --save 플래그를 사용하세요)")

		return nil
	},
}

// ============================================================
// config path 서브커맨드
// ============================================================

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "설정 파일 경로를 출력합니다",
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			fmt.Println("현재 로드된 설정 파일이 없습니다.")
			fmt.Println("탐색 경로:")
			fmt.Println("  1. ./config.yaml (현재 디렉터리)")
			fmt.Println("  2. $HOME/.mycli/config.yaml")
			fmt.Println("  3. --config 플래그로 직접 지정")
		} else {
			absPath, err := os.Getwd()
			if err != nil {
				absPath = "."
			}
			fmt.Printf("설정 파일: %s/%s\n", absPath, configFile)
		}
		return nil
	},
}

// ============================================================
// init: 서브커맨드 구조 등록
// ============================================================

func init() {
	// config의 서브커맨드들을 configCmd에 추가
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configPathCmd)

	// 기본 설정값 등록 (설정 파일에 없을 때 사용)
	viper.SetDefault("greet.default_name", "세상")
	viper.SetDefault("greet.count", 1)
	viper.SetDefault("greet.lang", "ko")
	viper.SetDefault("output", "table")
	viper.SetDefault("app.name", "mycli")
	viper.SetDefault("app.debug", false)
}
