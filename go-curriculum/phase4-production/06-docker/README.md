# 06-docker: Go 멀티스테이지 Docker 빌드

Go 애플리케이션을 위한 최적화된 Docker 이미지 빌드 방법을 학습합니다.

## 개념

### 멀티스테이지 빌드란?
하나의 Dockerfile에서 여러 `FROM` 명령을 사용해 빌드 단계를 분리합니다.
- **빌더 스테이지**: Go 컴파일러와 모든 도구 포함 (~800MB)
- **최종 스테이지**: 컴파일된 바이너리만 포함 (~5MB)

## 빠른 시작

### 1. Go 모듈 초기화
```bash
cd 06-docker
go mod init docker-demo
```

### 2. Docker 이미지 빌드
```bash
# 기본 빌드
docker build -t go-docker-demo .

# 빌드 정보 주입
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  -t go-docker-demo:1.0.0 .
```

### 3. 컨테이너 실행
```bash
# 기본 실행
docker run -p 8080:8080 go-docker-demo

# 포트 변경
docker run -p 3000:3000 -e PORT=3000 go-docker-demo

# 백그라운드 실행
docker run -d --name my-server -p 8080:8080 go-docker-demo
```

### 4. API 테스트
```bash
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/hello?name=Go개발자
curl http://localhost:8080/info
```

## 이미지 크기 비교

| 베이스 이미지      | 최종 크기 | 비고                    |
|-------------------|----------|-------------------------|
| golang:1.26       | ~850MB   | 빌드 도구 전체 포함      |
| golang:1.26-alpine| ~350MB   | Alpine Linux 기반        |
| alpine:latest     | ~12MB    | Alpine + 바이너리        |
| scratch           | ~5MB     | 바이너리만 포함 (최소)   |

## Docker 명령어 참고

```bash
# 이미지 목록 확인
docker images go-docker-demo

# 컨테이너 로그 확인
docker logs my-server

# 컨테이너 중지/삭제
docker stop my-server && docker rm my-server

# 이미지 삭제
docker rmi go-docker-demo

# 빌드 캐시 삭제
docker builder prune
```

## 핵심 학습 포인트

1. **레이어 캐시 최적화**: `go.mod`를 소스보다 먼저 복사
2. **정적 바이너리**: `CGO_ENABLED=0`으로 C 의존성 제거
3. **빌드 정보 주입**: `-ldflags -X`로 버전 정보 삽입
4. **보안**: 비루트 사용자, scratch 이미지로 공격 표면 최소화
5. **.dockerignore**: 불필요한 파일을 빌드 컨텍스트에서 제외
