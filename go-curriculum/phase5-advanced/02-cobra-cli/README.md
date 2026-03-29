# 02-cobra-cli: Cobra + Viper CLI 도구 구현

Go에서 Cobra와 Viper를 사용해 전문적인 수준의 CLI 도구를 만드는 방법을 학습합니다.

## 개요

**Cobra**: Go의 가장 인기 있는 CLI 프레임워크
- `kubectl`, `hugo`, `gh` (GitHub CLI), `docker` 등이 Cobra 기반
- 서브커맨드 트리, 플래그 관리, 자동 도움말 생성

**Viper**: Cobra와 함께 쓰이는 설정 관리 라이브러리
- 설정 파일(YAML/TOML/JSON), 환경 변수, 플래그를 단일 API로 통합
- 우선순위: 명시적 Set > 플래그 > 환경 변수 > 설정 파일 > 기본값

## 프로젝트 구조

```
02-cobra-cli/
├── main.go           # 진입점 (최대한 얇게 유지)
├── cmd/
│   ├── root.go       # 루트 커맨드 + Viper 초기화
│   ├── greet.go      # greet 서브커맨드
│   ├── version.go    # version 서브커맨드
│   └── config.go     # config 서브커맨드 (중첩 서브커맨드 예시)
├── config.yaml       # 기본 설정 파일
└── go.mod
```

## 설치 및 실행

```bash
cd 02-cobra-cli
go mod tidy
go run main.go --help
```

## 커맨드 구조

```
mycli
├── greet          인사 메시지 출력
├── version        버전 정보 출력
└── config         설정 관리
    ├── show       전체 설정 조회
    ├── get <키>   특정 값 조회
    ├── set <키> <값>  값 변경
    └── path       설정 파일 경로 출력
```

## 사용 예시

```bash
# 기본 도움말
go run main.go --help

# greet 커맨드
go run main.go greet --name 홍길동
go run main.go greet --name 홍길동 --count 3
go run main.go greet --lang en --name Alice
go run main.go greet 홍길동 김철수 이영희

# version 커맨드
go run main.go version
go run main.go version --short
go run main.go version -o json

# config 커맨드
go run main.go config show
go run main.go config get greet.default_name
go run main.go config set output json
go run main.go config path

# 전역 플래그 조합
go run main.go --verbose greet --name 홍길동
go run main.go -o json version
```

## 바이너리 빌드

```bash
# 빌드 (버전 정보 주입)
go build -ldflags "-X github.com/curriculum/cobra-example/cmd.Version=1.0.0 \
                   -X 'github.com/curriculum/cobra-example/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' \
                   -X github.com/curriculum/cobra-example/cmd.GitCommit=$(git rev-parse --short HEAD)" \
  -o mycli .

# 실행
./mycli version
./mycli greet --name 홍길동
```

## 주요 패턴

### 1. 서브커맨드 등록

```go
// cmd/root.go
rootCmd.AddCommand(greetCmd)
rootCmd.AddCommand(versionCmd)
rootCmd.AddCommand(configCmd)

// 중첩 서브커맨드 (config show, config get 등)
configCmd.AddCommand(configShowCmd)
configCmd.AddCommand(configGetCmd)
```

### 2. 플래그 종류

```go
// PersistentFlags: 이 커맨드 + 모든 자식 커맨드에 적용
rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "상세 출력")

// Flags: 이 커맨드에만 적용 (로컬)
greetCmd.Flags().StringVarP(&greetName, "name", "n", "", "이름")

// MarkRequired: 필수 플래그 지정
cmd.MarkFlagRequired("name")
```

### 3. Viper와 플래그 연동

```go
viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
// 이제 viper.GetString("output")이 플래그 값을 반환
```

### 4. 환경 변수 자동 매핑

```go
viper.SetEnvPrefix("MYCLI")
viper.AutomaticEnv()
// MYCLI_OUTPUT=json → viper.GetString("output") == "json"
```

## 학습 포인트

1. `main.go`는 최대한 얇게 — 실제 로직은 `cmd/` 패키지에
2. `RunE` 사용 (`Run` 대신) — 오류 반환으로 일관성 유지
3. `cobra.OnInitialize`로 공통 초기화 처리
4. Viper의 우선순위를 이해하면 유연한 설정 구조 가능
5. `-ldflags`로 빌드 시 버전 정보를 바이너리에 주입
