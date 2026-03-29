// cmd/greet.go
// 'greet' 서브커맨드 구현
//
// 사용 예시:
//   mycli greet --name 홍길동
//   mycli greet --name 홍길동 --count 3 --lang ko
//   mycli greet Alice Bob Charlie   (위치 인자로 여러 이름)
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ============================================================
// 로컬 플래그 변수
// ============================================================

var (
	greetName  string // --name 플래그
	greetCount int    // --count 플래그
	greetLang  string // --lang 플래그
)

// ============================================================
// greet 커맨드 정의
// ============================================================

var greetCmd = &cobra.Command{
	Use:   "greet [이름...]",
	Short: "지정한 이름에게 인사합니다",
	Long: `greet 커맨드는 지정한 이름에게 인사 메시지를 출력합니다.

이름은 --name 플래그 또는 위치 인자로 지정할 수 있습니다.
--count로 반복 횟수를, --lang으로 언어를 선택할 수 있습니다.`,

	// 예시 (--help에 표시됨)
	Example: `  mycli greet --name 홍길동
  mycli greet --name 홍길동 --count 3
  mycli greet --lang en --name Alice
  mycli greet 홍길동 김철수 이영희   (위치 인자)`,

	// Args: 인자 유효성 검사
	// 위치 인자는 0개 이상 허용 (--name 플래그와 함께 사용 가능)
	Args: cobra.ArbitraryArgs,

	// RunE: 오류를 반환하는 실행 함수 (Run 대신 RunE 권장)
	RunE: func(cmd *cobra.Command, args []string) error {
		// 인사할 대상 목록 수집
		names := collectNames(greetName, args)
		if len(names) == 0 {
			// 기본값: viper에서 읽기 (config.yaml의 greet.default_name)
			defaultName := viper.GetString("greet.default_name")
			if defaultName == "" {
				defaultName = "세상"
			}
			names = []string{defaultName}
		}

		// 인사 횟수 결정 (플래그 > viper 설정 > 기본값 1)
		count := greetCount
		if count == 0 {
			count = viper.GetInt("greet.count")
		}
		if count <= 0 {
			count = 1
		}

		// 각 이름에게 인사
		for _, name := range names {
			for i := 0; i < count; i++ {
				msg := buildGreeting(name, greetLang)
				fmt.Println(msg)
			}
		}

		return nil
	},
}

// ============================================================
// 헬퍼 함수
// ============================================================

// collectNames는 --name 플래그와 위치 인자에서 이름 목록을 수집합니다.
func collectNames(flagName string, args []string) []string {
	var names []string
	if flagName != "" {
		names = append(names, flagName)
	}
	names = append(names, args...)
	return names
}

// buildGreeting은 언어에 맞는 인사 메시지를 생성합니다.
func buildGreeting(name, lang string) string {
	hour := time.Now().Hour()

	// 시간대별 수식어
	var timeOfDay string
	switch {
	case hour < 12:
		timeOfDay = "좋은 아침"
	case hour < 18:
		timeOfDay = "좋은 오후"
	default:
		timeOfDay = "좋은 저녁"
	}

	switch strings.ToLower(lang) {
	case "en":
		return fmt.Sprintf("Hello, %s! Have a great day!", name)
	case "ja":
		return fmt.Sprintf("こんにちは、%sさん！", name)
	case "es":
		return fmt.Sprintf("¡Hola, %s! ¿Cómo estás?", name)
	case "zh":
		return fmt.Sprintf("你好，%s！", name)
	default:
		// 기본: 한국어
		return fmt.Sprintf("%s, %s님!", timeOfDay, name)
	}
}

// ============================================================
// init: 플래그 등록
// ============================================================

func init() {
	// Flags(): 이 커맨드에만 적용되는 로컬 플래그
	greetCmd.Flags().StringVarP(&greetName, "name", "n", "", "인사할 대상의 이름")
	greetCmd.Flags().IntVarP(&greetCount, "count", "c", 1, "인사 반복 횟수")
	greetCmd.Flags().StringVarP(&greetLang, "lang", "l", "ko", "언어 (ko|en|ja|es|zh)")

	// viper 연동: 설정 파일에서 기본값을 읽을 수 있습니다.
	_ = viper.BindPFlag("greet.name", greetCmd.Flags().Lookup("name"))
	_ = viper.BindPFlag("greet.lang", greetCmd.Flags().Lookup("lang"))
}
