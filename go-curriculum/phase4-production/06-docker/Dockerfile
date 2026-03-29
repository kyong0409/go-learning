# 06-docker/Dockerfile
# Go 멀티스테이지 빌드 예제
#
# 스테이지 1 (builder): Go 소스를 컴파일하여 정적 바이너리 생성
# 스테이지 2 (final):   scratch 이미지에 바이너리만 복사하여 최소 이미지 생성
#
# 결과 이미지 크기 비교:
#   golang:1.26 기반:  ~800MB
#   alpine 기반:       ~10MB
#   scratch 기반:      ~5MB (바이너리만 포함)

# ============================================================
# 스테이지 1: 빌드
# ============================================================
FROM golang:1.26-alpine AS builder

# 보안: 루트가 아닌 사용자로 실행 (최종 이미지에서도 사용)
RUN adduser -D -g '' appuser

# 작업 디렉토리 설정
WORKDIR /app

# 의존성 파일을 먼저 복사하여 Docker 레이어 캐시를 활용합니다.
# 소스 코드가 변경되어도 go.mod/go.sum이 같으면 의존성을 다시 다운로드하지 않습니다.
COPY go.mod go.sum* ./

# 의존성 다운로드
# -mod=readonly: go.sum이 최신 상태인지 확인
RUN go mod download && go mod verify

# 소스 코드 복사
COPY . .

# 빌드 인수 (선택적)
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# 정적 바이너리 빌드
# CGO_ENABLED=0: C 라이브러리 비활성화 (scratch 이미지 호환)
# -ldflags:
#   -w: DWARF 디버그 정보 제거 (크기 감소)
#   -s: 심볼 테이블 제거 (크기 감소)
#   -X: 변수에 빌드 정보 주입
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
      -X main.Version=${VERSION} \
      -X main.BuildTime=${BUILD_TIME} \
      -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/server \
    ./main.go

# 바이너리가 올바르게 생성되었는지 확인
RUN ls -la /app/server

# ============================================================
# 스테이지 2: 최종 이미지
# scratch는 완전히 빈 이미지입니다.
# OS, 쉘, 패키지 관리자가 없어 공격 표면이 최소화됩니다.
# ============================================================
FROM scratch

# 빌드 스테이지에서 필요한 파일만 복사

# 1. SSL/TLS 인증서 (HTTPS 요청을 위해 필요)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 2. 비루트 사용자 정보 (보안을 위해 루트로 실행하지 않음)
COPY --from=builder /etc/passwd /etc/passwd

# 3. 컴파일된 바이너리
COPY --from=builder /app/server /server

# 비루트 사용자로 실행
USER appuser

# 컨테이너가 노출할 포트
EXPOSE 8080

# 환경 변수 기본값
ENV PORT=8080

# 헬스 체크 (Docker가 컨테이너 상태를 모니터링)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/server", "-health"] || exit 1

# 컨테이너 시작 명령
ENTRYPOINT ["/server"]
