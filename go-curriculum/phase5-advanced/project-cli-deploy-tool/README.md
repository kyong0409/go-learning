# project-cli-deploy-tool: kubectl 유사 배포 관리 CLI

Cobra + Viper를 사용해 Docker 기반 애플리케이션 배포를 관리하는 CLI 도구입니다.

## 프로젝트 구조

```
project-cli-deploy-tool/
├── main.go
├── cmd/
│   ├── root.go       # 루트 커맨드 + 전역 플래그
│   ├── deploy.go     # deploy 서브커맨드
│   ├── status.go     # status 서브커맨드
│   ├── logs.go       # logs 서브커맨드
│   └── config.go     # config 서브커맨드
├── internal/
│   ├── deployer/
│   │   ├── deployer.go   # Deployer 인터페이스 + 타입
│   │   └── docker.go     # Docker CLI 기반 구현
│   └── output/
│       ├── output.go     # table/JSON/YAML 출력 포맷터
│       └── output_test.go
├── go.mod
├── Makefile
└── .goreleaser.yml
```

## 빌드 및 실행

```bash
cd project-cli-deploy-tool
go mod tidy

# 직접 실행
go run main.go --help

# 바이너리 빌드
make build

# 테스트 실행
make test
```

## 커맨드 구조

```
deploy-tool
├── deploy    애플리케이션 배포
├── status    배포 상태 확인
├── logs      컨테이너 로그 출력
└── config    설정 관리
    ├── show
    ├── get
    └── set
```

## 사용 예시

```bash
# 배포
deploy-tool deploy --app myapp --image nginx:latest
deploy-tool deploy --app api --image myapi:v1.2.3 --port 8080
deploy-tool deploy --app web --image frontend:latest --env NODE_ENV=prod --env PORT=3000
deploy-tool deploy --app myapp --image nginx:latest --dry-run

# 상태 확인
deploy-tool status            # 전체 앱
deploy-tool status myapp      # 특정 앱
deploy-tool status -o json    # JSON 형식

# 로그
deploy-tool logs myapp
deploy-tool logs myapp --follow
deploy-tool logs myapp --tail 200
deploy-tool logs myapp --since 1h

# 설정
deploy-tool config show
deploy-tool config get registry
deploy-tool config set registry myregistry.io
```

## 학습 포인트

1. Cobra 서브커맨드 구조로 kubectl과 유사한 UX 구현
2. `Deployer` 인터페이스로 Docker/K8s 백엔드 교체 가능
3. `output.Printer`로 table/JSON/YAML을 단일 API로 처리
4. `internal/` 패키지로 외부 노출 없이 캡슐화
5. `--dry-run` 플래그로 실제 실행 없이 계획 확인
