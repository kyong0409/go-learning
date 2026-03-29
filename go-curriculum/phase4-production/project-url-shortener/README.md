# URL 단축 서비스

Go 1.22+로 구현한 REST API URL 단축 서비스입니다.

## 기술 스택

| 항목 | 기술 |
|------|------|
| 언어 | Go 1.22+ |
| 라우터 | chi v5 |
| 데이터베이스 | PostgreSQL (pgx v5) |
| 로깅 | log/slog |
| 컨테이너 | Docker + docker-compose |

## 빠른 시작

### 방법 1: Docker Compose (권장)

```bash
# 전체 스택 시작 (앱 + PostgreSQL)
make docker-run

# 또는 직접
docker-compose up -d
```

### 방법 2: 로컬 실행 (인메모리 저장소)

```bash
# 의존성 설치
go mod download

# 개발 서버 시작
make dev

# 또는
go run ./cmd/server/
```

서버가 시작되면 `http://localhost:8080`에서 접근할 수 있습니다.

## API 문서

### URL 단축

**POST** `/api/shorten`

URL을 단축합니다.

요청:
```json
{
  "url": "https://www.example.com/very/long/url",
  "custom_code": "my-link",
  "expires_in": "24h"
}
```

응답 (`201 Created`):
```json
{
  "short_url": "http://localhost:8080/my-link",
  "short_code": "my-link",
  "original_url": "https://www.example.com/very/long/url",
  "expires_at": "2024-01-02T15:04:05Z"
}
```

### 리다이렉트

**GET** `/{code}`

단축 URL로 원본 URL에 리다이렉트합니다 (`302 Found`).

```bash
curl -L http://localhost:8080/my-link
```

### URL 목록 조회

**GET** `/api/urls?page=1&page_size=20`

```bash
curl http://localhost:8080/api/urls
```

응답:
```json
{
  "items": [...],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

### URL 정보 조회

**GET** `/api/urls/{code}`

```bash
curl http://localhost:8080/api/urls/my-link
```

### 통계 조회

**GET** `/api/urls/{code}/stats`

클릭 횟수 등 통계를 조회합니다.

```bash
curl http://localhost:8080/api/urls/my-link/stats
```

응답:
```json
{
  "short_code": "my-link",
  "original_url": "https://www.example.com/...",
  "click_count": 42,
  "created_at": "2024-01-01T12:00:00Z"
}
```

### URL 삭제

**DELETE** `/api/urls/{code}`

```bash
curl -X DELETE http://localhost:8080/api/urls/my-link
```

### 헬스 체크

**GET** `/health`

```bash
curl http://localhost:8080/health
```

## 오류 응답 형식

모든 오류는 다음 형식으로 반환됩니다:

```json
{
  "error": "오류 메시지",
  "code": "ERROR_CODE"
}
```

| 코드 | 설명 |
|------|------|
| `INVALID_JSON` | 잘못된 JSON 형식 |
| `VALIDATION_ERROR` | 입력 검증 실패 |
| `CODE_CONFLICT` | 단축 코드 중복 |
| `NOT_FOUND` | 리소스 없음 |
| `URL_EXPIRED` | 만료된 URL |
| `RATE_LIMIT_EXCEEDED` | 요청 초과 |
| `INTERNAL_ERROR` | 서버 내부 오류 |

## 환경 변수

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `SERVER_ADDR` | `:8080` | 서버 바인딩 주소 |
| `DATABASE_URL` | (메모리) | PostgreSQL 연결 URL |
| `BASE_URL` | `http://localhost:8080` | 단축 URL 기본 도메인 |
| `SHORT_CODE_LENGTH` | `6` | 단축 코드 길이 |
| `RATE_LIMIT` | `100` | 분당 최대 요청 수 |
| `LOG_LEVEL` | `info` | 로그 레벨 (debug/info/warn/error) |
| `ENVIRONMENT` | `development` | 실행 환경 |

## 개발 명령어

```bash
make build          # 바이너리 빌드
make test           # 테스트 실행
make test-coverage  # 커버리지 측정
make lint           # 린트 검사
make fmt            # 코드 포맷팅
make docker-build   # Docker 이미지 빌드
make migrate-up     # DB 마이그레이션 적용
make help           # 전체 명령어 목록
```

## 프로젝트 구조

```
url-shortener/
├── cmd/server/         # 진입점
├── internal/
│   ├── config/         # 환경 변수 설정
│   ├── handler/        # HTTP 핸들러 + 미들웨어
│   ├── model/          # 도메인 모델
│   └── store/          # 저장소 인터페이스 + 구현
├── migrations/         # SQL 마이그레이션
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .golangci.yml
```
