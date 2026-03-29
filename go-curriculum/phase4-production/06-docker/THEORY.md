# 06-docker: Go와 Docker 컨테이너화

> Go는 정적 바이너리를 생성하기 때문에 Docker와 궁합이 완벽합니다. 멀티스테이지 빌드로 5~20MB의 최소 이미지를 만들 수 있습니다. Python(1GB+), Java(500MB+)와 비교할 수 없는 차이입니다.

---

## 1. Go가 Docker에 최적인 이유

### 정적 바이너리

```bash
# CGO_ENABLED=0 으로 빌드하면 외부 라이브러리 의존성이 없음
CGO_ENABLED=0 go build -o server .

# 빌드된 바이너리 확인
file server
# server: ELF 64-bit LSB executable, statically linked
# "statically linked" — OS의 libc 조차 필요 없음
```

Python이나 Java는 인터프리터/JVM + 수십~수백 개의 라이브러리가 런타임에 필요합니다. Go 바이너리는 모든 것이 단일 파일에 포함됩니다.

### 컨테이너 이미지 크기 비교

| 언어 | 기반 이미지 | 앱 이미지 크기 |
|------|------------|--------------|
| Python (Flask) | python:3.12 (1.1GB) | ~200MB |
| Java (Spring) | eclipse-temurin:21 (460MB) | ~300MB |
| Node.js | node:20 (1.1GB) | ~150MB |
| Go | golang:1.26 (빌드용) | ~5MB (scratch 기반) |

---

## 2. 멀티스테이지 빌드

단일 Dockerfile에 여러 빌드 스테이지를 정의합니다. 최종 이미지에는 실행에 필요한 파일만 복사합니다.

```dockerfile
# ============================================================
# 스테이지 1: 빌드 환경
# golang:1.26-alpine은 Go 컴파일러가 포함된 ~250MB 이미지
# ============================================================
FROM golang:1.26-alpine AS builder

# 보안: 루트가 아닌 사용자를 미리 생성 (최종 이미지에서 사용)
RUN adduser -D -g '' appuser

WORKDIR /app

# 의존성 레이어 캐싱 최적화 — 소스보다 먼저 복사
# go.mod/go.sum이 변경되지 않으면 이 레이어는 캐시에서 재사용
COPY go.mod go.sum* ./
RUN go mod download && go mod verify

# 소스 복사 (go.mod 변경 여부와 무관)
COPY . .

# 빌드 인수 — docker build --build-arg VERSION=1.0.0 으로 주입
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# 정적 바이너리 빌드
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
      -X main.Version=${VERSION} \
      -X main.BuildTime=${BUILD_TIME} \
      -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/server ./main.go

# ============================================================
# 스테이지 2: 최소 실행 환경
# scratch = 완전히 빈 이미지 (0MB)
# ============================================================
FROM scratch

# 빌드 스테이지에서 필요한 파일만 선택적으로 복사
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /app/server /server

# 비루트 사용자로 실행 (보안)
USER appuser

EXPOSE 8080
ENV PORT=8080

ENTRYPOINT ["/server"]
```

**최종 이미지에는 다음만 포함**:
- Go 바이너리 (`/server`)
- SSL/TLS 인증서 (HTTPS 요청용)
- 사용자 정보 (`/etc/passwd`)

---

## 3. 빌드 플래그 상세 설명

```bash
CGO_ENABLED=0   # C 라이브러리(CGO) 비활성화 → 순수 정적 바이너리
GOOS=linux      # 대상 OS: linux (빌드 머신이 Mac/Windows여도 동작)
GOARCH=amd64    # 대상 아키텍처: x86-64

-ldflags:
  -w    # DWARF 디버그 정보 제거 → 바이너리 크기 감소 (~30%)
  -s    # 심볼 테이블 제거 → 바이너리 크기 추가 감소
  -X main.Version=1.0.0  # main 패키지의 Version 변수에 값 주입
```

### 링커 플래그로 빌드 정보 주입

```go
// main.go
var (
    Version   = "dev"      // -X main.Version=1.2.3 으로 덮어씀
    BuildTime = "unknown"  // -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    GitCommit = "unknown"  // -X main.GitCommit=$(git rev-parse --short HEAD)
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "version":    Version,
        "build_time": BuildTime,
        "git_commit": GitCommit,
    })
}
```

```bash
# 빌드 시 정보 주입
docker build \
  --build-arg VERSION=1.2.3 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  -t myapp:1.2.3 .
```

---

## 4. Docker 레이어 캐싱 최적화

Docker는 Dockerfile의 각 명령을 레이어로 저장하고, 변경이 없으면 캐시를 재사용합니다.

```dockerfile
# 나쁜 예 — 소스 변경 시 항상 의존성을 다시 다운로드
COPY . .
RUN go mod download
RUN go build -o server .

# 좋은 예 — go.mod가 변경되지 않으면 의존성 레이어 캐시 재사용
COPY go.mod go.sum* ./      # 레이어 1: 거의 변경 안 됨
RUN go mod download          # 레이어 2: 레이어 1이 캐시되면 이것도 캐시

COPY . .                     # 레이어 3: 소스 변경 시마다 무효화
RUN go build -o server .    # 레이어 4: 레이어 3이 변경되면 항상 재실행
```

**원칙**: 자주 변경되는 파일은 늦게 복사합니다.

---

## 5. .dockerignore

빌드 컨텍스트에서 불필요한 파일을 제외합니다. `.gitignore`와 유사한 역할입니다.

```
# .dockerignore
.git/
*.exe
*.test
tmp/
vendor/
.env
.env.*
README.md
*.md
```

`.dockerignore`가 없으면 `.git` 디렉토리 전체가 빌드 컨텍스트에 포함되어 빌드가 느려집니다.

---

## 6. scratch vs alpine vs distroless

| 기반 이미지 | 크기 | 쉘 | 패키지 관리자 | 보안 | 사용 사례 |
|------------|------|-----|--------------|------|-----------|
| `scratch` | 0MB | X | X | 최고 | 완전 정적 바이너리 |
| `alpine` | ~5MB | O | apk | 높음 | 디버깅 도구 필요 시 |
| `distroless` | ~20MB | X | X | 매우 높음 | glibc 의존 바이너리 |
| `ubuntu` | ~80MB | O | apt | 보통 | 일반 목적 |

```dockerfile
# scratch: 가장 작고 안전 — CGO_ENABLED=0 필수
FROM scratch

# alpine: 쉘이 있어 디버깅 가능 — 개발/스테이징 환경
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

# distroless: glibc 기반 바이너리, 쉘 없음
FROM gcr.io/distroless/base-debian12
```

---

## 7. docker-compose: 로컬 개발 환경

```yaml
# docker-compose.yml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        VERSION: dev
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://user:password@db:5432/mydb
      - REDIS_URL=redis://cache:6379
    depends_on:
      db:
        condition: service_healthy
      cache:
        condition: service_started
    # 개발용: 소스 마운트 + air 핫 리로드
    volumes:
      - .:/app
    command: air  # go run 대신 air로 핫 리로드

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: mydb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d mydb"]
      interval: 5s
      timeout: 5s
      retries: 5

  cache:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

---

## 8. 헬스체크

컨테이너 오케스트레이터(Kubernetes, ECS 등)가 서비스 상태를 확인합니다.

```dockerfile
# Dockerfile — HEALTHCHECK 지시문
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1
```

```go
// Go 서버 — /health 엔드포인트
func handleHealth(w http.ResponseWriter, r *http.Request) {
    // DB 연결 확인
    if err := db.PingContext(r.Context()); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "error":  err.Error(),
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "status":  "healthy",
        "version": Version,
        "uptime":  time.Since(startTime).String(),
    })
}
```

**헬스체크 종류**:
- **Liveness probe**: 서버가 살아있는지 (`/health`)
- **Readiness probe**: 트래픽을 받을 준비가 됐는지 (`/ready`) — DB 연결, 초기화 완료 등 확인

---

## 9. Makefile로 빌드 자동화

```makefile
VERSION  := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
IMAGE_NAME := myapp

.PHONY: build docker-build docker-run

build:
	CGO_ENABLED=0 go build \
	  -ldflags="-w -s -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)" \
	  -o server ./main.go

docker-build:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg BUILD_TIME=$(BUILD_TIME) \
	  --build-arg GIT_COMMIT=$(GIT_COMMIT) \
	  -t $(IMAGE_NAME):$(VERSION) \
	  -t $(IMAGE_NAME):latest \
	  .

docker-run:
	docker run --rm -p 8080:8080 $(IMAGE_NAME):latest
```

---

## 10. GoReleaser: 크로스 플랫폼 배포 자동화

GoReleaser는 CI/CD 파이프라인에서 다중 플랫폼 바이너리와 Docker 이미지를 자동으로 빌드하고 배포합니다.

```yaml
# .goreleaser.yml
builds:
  - binary: server
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -w -s
      - -X main.Version={{.Version}}
      - -X main.BuildTime={{.Date}}
      - -X main.GitCommit={{.ShortCommit}}

dockers:
  - image_templates:
      - "ghcr.io/myorg/myapp:{{ .Version }}"
      - "ghcr.io/myorg/myapp:latest"
    dockerfile: Dockerfile
    use: buildx
    build_flag_templates:
      - "--platform=linux/amd64,linux/arm64"

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

changelog:
  sort: asc
  filters:
    exclude: ['^docs:', '^test:']
```

```bash
# CI에서 실행 (GitHub Actions 등)
goreleaser release --clean
```

GoReleaser 하나로 다음을 자동화합니다:
- Linux/macOS/Windows × amd64/arm64 바이너리 빌드
- 압축 파일 (.tar.gz, .zip) 생성
- Docker 멀티 아키텍처 이미지 빌드 및 푸시
- GitHub Release 생성 및 체인지로그 작성
- Homebrew 포뮬러 업데이트
