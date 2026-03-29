// cmd/config.go
// 'config' 서브커맨드: 도구 설정 관리
//
// 사용 예시:
//   deploy-tool config show
//   deploy-tool config get registry
//   deploy-tool config set registry myregistry.io
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd는 설정을 관리합니다.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "도구 설정을 관리합니다",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "현재 설정 전체 출력",
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := viper.AllSettings()
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		if f := viper.ConfigFileUsed(); f != "" {
			fmt.Printf("\n설정 파일: %s\n", f)
		}
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <키>",
	Short: "특정 설정 값 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if !viper.IsSet(key) {
			return fmt.Errorf("설정 키 없음: %s", key)
		}
		fmt.Printf("%s = %v\n", key, viper.Get(key))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <키> <값>",
	Short: "설정 값 변경 (현재 세션)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set(args[0], args[1])
		fmt.Printf("설정됨: %s = %s\n", args[0], args[1])
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}
