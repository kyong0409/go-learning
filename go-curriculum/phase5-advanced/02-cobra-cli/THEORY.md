# Cobra & Viper - 전문가 수준의 CLI 도구 개발

## CLI 도구 설계 철학

### Unix 원칙

```
1. 한 가지 일을 잘 하라 (Single Responsibility)
2. 조합 가능하게 만들어라 (파이프: cmd1 | cmd2)
3. stdout은 데이터 출력, stderr는 진단 메시지
4. 종료 코드: 0 = 성공, 1 = 일반 오류, 2 = 잘못된 사용법
5. --help는 항상 동작해야 한다
6. --version은 항상 동작해야 한다
```

### POSIX 플래그 컨벤션

```bash
# 짧은 플래그: 단일 대시 + 단일 문자
myapp deploy -e production -n 3

# 긴 플래그: 이중 대시 + 단어
myapp deploy --env production --replicas 3

# 불리언 플래그: 값 없이 사용
myapp deploy --dry-run

# 서브커맨드 패턴 (kubectl, git, docker 스타일)
myapp <resource> <verb> [flags]
myapp bookmark list --format json
myapp bookmark create --url https://golang.org
```

---

## Cobra 프레임워크

kubectl, helm, gh(GitHub CLI), hugo가 모두 Cobra를 사용합니다.

### Command 구조

```go
// cmd/root.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
    Use:   "bookmarkctl",                        // 명령어 이름
    Short: "북마크 관리 CLI 도구",                 // 한 줄 설명
    Long: `bookmarkctl은 북마크를 관리하는 CLI 도구입니다.
다양한 형식으로 출력하고 원격 서버와 동기화합니다.`, // --help에 표시
    // Run이 없으면 서브커맨드 없이 실행 시 도움말 표시
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)
    // PersistentFlags: 이 명령어와 모든 하위 명령어에 적용
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
        "설정 파일 경로 (기본값: $HOME/.bookmarkctl.yaml)")
    rootCmd.PersistentFlags().StringP("output", "o", "table",
        "출력 형식 (table|json|yaml)")
    // Flags: 이 명령어에만 적용 (로컬)
    rootCmd.Flags().BoolP("version", "v", false, "버전 출력")
}
```

### 서브커맨드 추가

```go
// cmd/bookmark.go
var bookmarkCmd = &cobra.Command{
    Use:   "bookmark",
    Short: "북마크 관리",
    Aliases: []string{"bm"}, // 별칭: bookmarkctl bm list
}

// cmd/bookmark_create.go
var createCmd = &cobra.Command{
    Use:   "create",
    Short: "새 북마크 생성",
    Example: `  bookmarkctl bookmark create --url https://golang.org --title "Go 공식 사이트"
  bookmarkctl bookmark create -u https://pkg.go.dev -t "Go 패키지 문서"`,

    // Args 검증
    Args: cobra.NoArgs, // 위치 인수 없음

    // 실행 전 검증
    PreRunE: func(cmd *cobra.Command, args []string) error {
        url, _ := cmd.Flags().GetString("url")
        if url == "" {
            return fmt.Errorf("--url 플래그가 필요합니다")
        }
        return nil
    },

    // 실제 실행 (에러 반환 가능한 RunE 권장)
    RunE: func(cmd *cobra.Command, args []string) error {
        url, _ := cmd.Flags().GetString("url")
        title, _ := cmd.Flags().GetString("title")
        tags, _ := cmd.Flags().GetStringSlice("tags")

        return createBookmark(url, title, tags)
    },
}

func init() {
    bookmarkCmd.AddCommand(createCmd)
    rootCmd.AddCommand(bookmarkCmd)

    createCmd.Flags().StringP("url", "u", "", "북마크 URL (필수)")
    createCmd.Flags().StringP("title", "t", "", "북마크 제목")
    createCmd.Flags().StringSliceP("tags", "g", nil, "태그 목록 (쉼표 구분)")

    // Required flag 설정
    createCmd.MarkFlagRequired("url")
}
```

### Args 검증 옵션

```go
Args: cobra.NoArgs,               // 위치 인수 없음
Args: cobra.ExactArgs(1),         // 정확히 1개
Args: cobra.MinimumNArgs(1),      // 최소 1개
Args: cobra.MaximumNArgs(3),      // 최대 3개
Args: cobra.RangeArgs(1, 3),      // 1~3개
Args: cobra.MatchAll(             // 여러 조건 AND
    cobra.ExactArgs(1),
    cobra.OnlyValidArgs,
),
ValidArgs: []string{"json", "yaml", "table"}, // OnlyValidArgs와 함께 사용
```

### PreRun / PostRun 훅

```go
var deployCmd = &cobra.Command{
    Use: "deploy",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // 모든 하위 명령어 실행 전: 인증 확인
        return checkAuth()
    },
    PreRunE: func(cmd *cobra.Command, args []string) error {
        // 이 명령어만: 환경 확인
        return validateEnvironment()
    },
    RunE: func(cmd *cobra.Command, args []string) error {
        return runDeploy()
    },
    PostRunE: func(cmd *cobra.Command, args []string) error {
        // 성공 후: 알림 전송
        return sendNotification()
    },
}
```

### 자동 완성 생성

```bash
# bash 자동 완성
bookmarkctl completion bash > /etc/bash_completion.d/bookmarkctl

# zsh 자동 완성
bookmarkctl completion zsh > "${fpath[1]}/_bookmarkctl"

# fish 자동 완성
bookmarkctl completion fish > ~/.config/fish/completions/bookmarkctl.fish

# PowerShell 자동 완성
bookmarkctl completion powershell | Out-String | Invoke-Expression
```

---

## Viper 설정 관리

### 설정 우선순위 (높은 순서)

```
1. 명시적 Set (코드에서 viper.Set())
2. 플래그 (--flag-name)
3. 환경 변수 (MYAPP_SERVER_URL)
4. 설정 파일 (config.yaml)
5. 기본값 (viper.SetDefault())
```

### 기본 설정

```go
// cmd/root.go
func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, _ := os.UserHomeDir()
        viper.AddConfigPath(home)       // ~/.bookmarkctl.yaml
        viper.AddConfigPath(".")        // ./config.yaml
        viper.SetConfigName(".bookmarkctl")
        viper.SetConfigType("yaml")
    }

    // 환경 변수: BOOKMARKCTL_SERVER_URL → server.url
    viper.SetEnvPrefix("BOOKMARKCTL")
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    if err := viper.ReadInConfig(); err == nil {
        fmt.Fprintln(os.Stderr, "설정 파일 사용:", viper.ConfigFileUsed())
    }
}
```

### 설정 파일 (config.yaml)

```yaml
# ~/.bookmarkctl.yaml
server:
  url: https://api.bookmarks.example.com
  timeout: 30s

auth:
  token: eyJhbGciOiJIUzI1NiJ9...

output:
  format: table
  color: true

defaults:
  page_size: 20
```

### 플래그와 Viper 바인딩

```go
func init() {
    rootCmd.PersistentFlags().String("server-url", "", "API 서버 URL")
    // 플래그를 Viper 키에 바인딩
    viper.BindPFlag("server.url", rootCmd.PersistentFlags().Lookup("server-url"))

    // 기본값 설정
    viper.SetDefault("server.timeout", "30s")
    viper.SetDefault("output.format", "table")
}

// 설정값 읽기
func runCreate() error {
    serverURL := viper.GetString("server.url")
    timeout := viper.GetDuration("server.timeout")
    format := viper.GetString("output.format")
    // ...
}
```

### 환경 변수 자동 매핑

```bash
# 설정 파일의 server.url은 환경 변수로 덮어씌울 수 있음
export BOOKMARKCTL_SERVER_URL=https://staging.example.com
export BOOKMARKCTL_AUTH_TOKEN=mytoken123

# 실행 시 환경 변수가 설정 파일보다 우선
bookmarkctl bookmark list
```

---

## 출력 형식 패턴

### 테이블 / JSON / YAML 출력

```go
func printBookmarks(bookmarks []Bookmark, format string) error {
    switch format {
    case "json":
        return json.NewEncoder(os.Stdout).Encode(bookmarks)

    case "yaml":
        return yaml.NewEncoder(os.Stdout).Encode(bookmarks)

    case "table":
        // tabwriter로 정렬된 테이블 출력
        w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
        fmt.Fprintln(w, "ID\tURL\tTITLE\tTAGS")
        fmt.Fprintln(w, "--\t---\t-----\t----")
        for _, b := range bookmarks {
            fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
                b.ID, b.URL, b.Title, strings.Join(b.Tags, ","))
        }
        return w.Flush()

    default:
        return fmt.Errorf("알 수 없는 출력 형식: %s", format)
    }
}
```

### 색상 출력

```go
import "github.com/fatih/color"

var (
    green  = color.New(color.FgGreen).SprintFunc()
    red    = color.New(color.FgRed).SprintFunc()
    yellow = color.New(color.FgYellow).SprintFunc()
    bold   = color.New(color.Bold).SprintFunc()
)

func printStatus(ok bool, msg string) {
    if ok {
        fmt.Printf("%s %s\n", green("✓"), msg)
    } else {
        fmt.Printf("%s %s\n", red("✗"), msg)
    }
}
```

---

## 배포: GoReleaser

```yaml
# .goreleaser.yaml
project_name: bookmarkctl
builds:
  - main: ./cmd/bookmarkctl
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      # 빌드 시 버전 정보 주입
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.buildDate={{.Date}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

changelog:
  sort: asc
  filters:
    exclude: ['^docs:', '^test:']
```

```go
// main.go: 버전 정보 (빌드 시 ldflags로 주입)
var (
    version   = "dev"
    commit    = "none"
    buildDate = "unknown"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "버전 정보 출력",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("bookmarkctl %s (commit: %s, built: %s)\n",
            version, commit, buildDate)
    },
}
```

---

## 실전 팁

```go
// 1. stderr로 진단 메시지 출력 (stdout은 데이터만)
fmt.Fprintln(os.Stderr, "경고: 서버에 연결할 수 없습니다")

// 2. os.Exit 대신 RunE에서 에러 반환 (defer 실행 보장)
RunE: func(cmd *cobra.Command, args []string) error {
    return fmt.Errorf("배포 실패: %w", err)
}

// 3. 인터랙티브 여부 감지 (파이프 사용 시 프롬프트 비활성화)
func isInteractive() bool {
    fi, _ := os.Stdin.Stat()
    return (fi.Mode() & os.ModeCharDevice) != 0
}

// 4. 시그널 처리로 graceful shutdown
ctx, cancel := signal.NotifyContext(context.Background(),
    os.Interrupt, syscall.SIGTERM)
defer cancel()
```
