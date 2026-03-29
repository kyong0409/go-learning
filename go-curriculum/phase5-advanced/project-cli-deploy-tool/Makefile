# Makefile - deploy-tool 빌드/테스트 자동화

# ============================================================
# 변수 정의
# ============================================================

BINARY_NAME := deploy-tool
MODULE      := github.com/curriculum/deploy-tool
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "\
  -X $(MODULE)/cmd.Version=$(VERSION) \
  -X $(MODULE)/cmd.GitCommit=$(GIT_COMMIT) \
  -X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)"

# ============================================================
# 기본 타겟
# ============================================================

.PHONY: all
all: lint test build ## 전체 빌드 (lint + test + build)

.PHONY: build
build: ## 바이너리 빌드
	@echo "빌드 중: $(BINARY_NAME) $(VERSION)"
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .
	@echo "완료: bin/$(BINARY_NAME)"

.PHONY: install
install: ## 시스템에 설치 ($GOPATH/bin)
	go install $(LDFLAGS) .

# ============================================================
# 테스트
# ============================================================

.PHONY: test
test: ## 단위 테스트 실행
	go test ./... -v -race -count=1

.PHONY: test-coverage
test-coverage: ## 커버리지 포함 테스트
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "커버리지 리포트: coverage.html"

.PHONY: bench
bench: ## 벤치마크 테스트
	go test ./... -bench=. -benchmem

# ============================================================
# 코드 품질
# ============================================================

.PHONY: lint
lint: ## golangci-lint 실행
	@which golangci-lint > /dev/null || (echo "golangci-lint 없음. 설치: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

.PHONY: fmt
fmt: ## gofmt 포맷팅
	gofmt -w .
	goimports -w . 2>/dev/null || true

.PHONY: vet
vet: ## go vet 실행
	go vet ./...

# ============================================================
# 의존성
# ============================================================

.PHONY: tidy
tidy: ## go.mod/go.sum 정리
	go mod tidy

.PHONY: deps
deps: ## 의존성 다운로드
	go mod download

# ============================================================
# 배포 빌드 (크로스 컴파일)
# ============================================================

.PHONY: build-all
build-all: ## 모든 플랫폼용 빌드
	GOOS=linux  GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .
	@echo "빌드 완료:"
	@ls -lh bin/

# GoReleaser로 공식 릴리즈
.PHONY: release
release: ## GoReleaser로 릴리즈 빌드
	@which goreleaser > /dev/null || (echo "goreleaser 없음. 설치: https://goreleaser.com" && exit 1)
	goreleaser release --clean

.PHONY: release-snapshot
release-snapshot: ## GoReleaser 스냅샷 빌드 (태그 없이)
	goreleaser release --snapshot --clean

# ============================================================
# 정리
# ============================================================

.PHONY: clean
clean: ## 빌드 산출물 삭제
	rm -rf bin/ dist/ coverage.out coverage.html

# ============================================================
# 도움말
# ============================================================

.PHONY: help
help: ## 사용 가능한 타겟 목록
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
