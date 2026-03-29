# project-url-shortener/Dockerfile
# URL 단축 서비스 멀티스테이지 빌드

# ============================================================
# 스테이지 1: 의존성 캐시
# go.mod/go.sum만 먼저 복사하여 소스 변경 시 의존성을 재다운로드하지 않습니다.
# ============================================================
FROM golang:1.26-alpine AS deps

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download && go mod verify

# ============================================================
# 스테이지 2: 빌드
# ============================================================
FROM deps AS builder

# 빌드 메타데이터
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s \
      -X main.Version=${VERSION} \
      -X main.BuildTime=${BUILD_TIME} \
      -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/server \
    ./cmd/server/

# ============================================================
# 스테이지 3: 최종 최소 이미지
# ============================================================
FROM scratch

# 인증서 및 시간대 데이터 복사
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 바이너리 복사
COPY --from=builder /app/server /server

EXPOSE 8080

ENV PORT=8080 \
    ENVIRONMENT=production \
    LOG_LEVEL=info

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD ["/server", "-health"] || exit 1

ENTRYPOINT ["/server"]
