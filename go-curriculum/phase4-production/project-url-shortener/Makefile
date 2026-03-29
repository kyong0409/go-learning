# Makefile
# URL 단축 서비스 빌드 및 관리 명령어
# 사용법: make <target>

# ============================================================
# 변수 설정
# ============================================================
APP_NAME    := url-shortener
CMD_PATH    := ./cmd/server
BINARY      := bin/$(APP_NAME)
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS     := -ldflags="-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Docker 설정
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG   := $(VERSION)

# 데이터베이스 설정
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable

# 색상 출력
CYAN  := \033[36m
GREEN := \033[32m
RESET := \033[0m

.PHONY: all build clean test test-verbose lint fmt vet \
        run dev docker-build docker-run docker-stop \
        migrate-up migrate-down migrate-create \
        help

# ============================================================
# 기본 타겟
# ============================================================

## all: 코드 검사 후 빌드
all: fmt vet lint test build

## help: 사용 가능한 명령어 목록
help:
	@echo "$(CYAN)사용 가능한 명령어:$(RESET)"
	@grep -E '^## ' Makefile | sed 's/## /  /'

# ============================================================
# 빌드
# ============================================================

## build: 바이너리 빌드
build:
	@echo "$(CYAN)빌드 중... (버전: $(VERSION))$(RESET)"
	@mkdir -p bin
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) $(CMD_PATH)
	@echo "$(GREEN)빌드 완료: $(BINARY)$(RESET)"

## clean: 빌드 결과물 삭제
clean:
	@echo "$(CYAN)정리 중...$(RESET)"
	@rm -rf bin/ coverage.out coverage.html
	@echo "$(GREEN)정리 완료$(RESET)"

# ============================================================
# 실행
# ============================================================

## run: 서버 실행 (빌드 후)
run: build
	@echo "$(CYAN)서버 시작...$(RESET)"
	./$(BINARY)

## dev: 환경 변수 설정 후 개발 서버 실행
dev:
	@echo "$(CYAN)개발 서버 시작...$(RESET)"
	LOG_LEVEL=debug ENVIRONMENT=development go run $(CMD_PATH)/main.go

# ============================================================
# 테스트
# ============================================================

## test: 모든 테스트 실행
test:
	@echo "$(CYAN)테스트 실행 중...$(RESET)"
	go test ./... -timeout 60s
	@echo "$(GREEN)테스트 완료$(RESET)"

## test-verbose: 상세 출력으로 테스트 실행
test-verbose:
	go test ./... -v -timeout 60s

## test-coverage: 커버리지 측정 후 HTML 리포트 생성
test-coverage:
	@echo "$(CYAN)커버리지 측정 중...$(RESET)"
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)커버리지 리포트: coverage.html$(RESET)"

## test-race: 경쟁 조건 감지 모드로 테스트
test-race:
	go test -race ./... -timeout 60s

# ============================================================
# 코드 품질
# ============================================================

## fmt: 코드 포맷팅
fmt:
	@echo "$(CYAN)코드 포맷팅...$(RESET)"
	gofmt -w .
	goimports -w . 2>/dev/null || true

## vet: 정적 분석
vet:
	@echo "$(CYAN)정적 분석...$(RESET)"
	go vet ./...

## lint: golangci-lint 실행
lint:
	@echo "$(CYAN)린트 검사...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint 미설치. 설치: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## tidy: go.mod 정리
tidy:
	go mod tidy
	go mod verify

# ============================================================
# Docker
# ============================================================

## docker-build: Docker 이미지 빌드
docker-build:
	@echo "$(CYAN)Docker 이미지 빌드: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.
	@echo "$(GREEN)이미지 빌드 완료$(RESET)"

## docker-run: Docker Compose로 전체 스택 실행
docker-run:
	@echo "$(CYAN)Docker Compose 시작...$(RESET)"
	BUILD_TIME=$(BUILD_TIME) GIT_COMMIT=$(GIT_COMMIT) docker-compose up -d
	@echo "$(GREEN)서비스 시작 완료. http://localhost:8080$(RESET)"

## docker-stop: Docker Compose 중지
docker-stop:
	docker-compose down

## docker-logs: 서비스 로그 확인
docker-logs:
	docker-compose logs -f app

## docker-clean: Docker 리소스 정리
docker-clean:
	docker-compose down -v --remove-orphans
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest 2>/dev/null || true

# ============================================================
# 데이터베이스 마이그레이션
# ============================================================

## migrate-up: 마이그레이션 적용
migrate-up:
	@echo "$(CYAN)마이그레이션 적용...$(RESET)"
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -database "$(DATABASE_URL)" -path migrations up; \
	else \
		echo "migrate 미설치. 설치: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

## migrate-down: 마이그레이션 롤백
migrate-down:
	@echo "$(CYAN)마이그레이션 롤백...$(RESET)"
	migrate -database "$(DATABASE_URL)" -path migrations down 1

## migrate-create: 새 마이그레이션 파일 생성
## 사용법: make migrate-create NAME=add_user_column
migrate-create:
	@if [ -z "$(NAME)" ]; then echo "NAME 변수가 필요합니다. 예: make migrate-create NAME=add_index"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(NAME)
