# Phase 4: 프로덕션 Go - 학습 가이드

> 대상: Phase 1~3을 마친 개발자. "Go를 아는 사람"에서 "Go로 프로덕션 시스템을 만드는 사람"으로 성장하는 것이 목표입니다.

---

## Phase 4 학습 목표

이 Phase를 마치면 다음을 할 수 있습니다.

- Go 1.22+의 향상된 HTTP 라우팅으로 RESTful API를 표준 라이브러리만으로 구축한다
- 미들웨어 체인으로 인증, 로깅, 복구 등 횡단 관심사를 분리한다
- 제네릭으로 타입 안전하고 재사용 가능한 컬렉션과 유틸리티를 작성한다
- `log/slog`로 구조화된 로깅을 구현하고 프로덕션 환경에서 운영한다
- Go 1.23 이터레이터 패턴으로 지연 평가 파이프라인을 구성한다
- Docker 멀티스테이지 빌드로 5~20MB의 최소 컨테이너 이미지를 만든다

---

## 모던 Go 기능 타임라인

Go는 하위 호환성을 철저히 지키면서도 꾸준히 발전해왔습니다.

| 버전 | 출시 | 주요 기능 |
|------|------|-----------|
| Go 1.18 | 2022.03 | 제네릭 (타입 파라미터) 도입 |
| Go 1.19 | 2022.08 | `//go:build` 태그 표준화 |
| Go 1.20 | 2023.02 | `errors.Join`, `http.ResponseController` |
| Go 1.21 | 2023.08 | `log/slog`, `slices`, `maps`, `cmp` 표준 패키지 |
| Go 1.22 | 2024.02 | 향상된 `ServeMux` (메서드 패턴, 경로 와일드카드), `for` 루프 변수 수정 |
| Go 1.23 | 2024.08 | `iter` 패키지, range-over-function 이터레이터 |
| Go 1.24 | 2025.02 | 제네릭 타입 별칭, `os.Root`, 약한 포인터(`weak`) |

**"지금 Go를 배우기 가장 좋은 시기"인 이유**: Go 1.21~1.24에 걸쳐 제네릭, 구조화된 로깅, 이터레이터, 향상된 라우팅이 모두 표준 라이브러리로 들어왔습니다. 서드파티 의존성 없이도 프로덕션 품질의 코드를 작성할 수 있습니다.

---

## 프레임워크 선택 가이드

### 표준 라이브러리 (`net/http`)

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users/{id}", handleGetUser)  // Go 1.22+
```

**선택 기준**: 새 프로젝트, 의존성 최소화, Go 1.22+ 전용 서비스.
Go 1.22 이전에는 메서드 매칭과 경로 파라미터가 없어서 서드파티가 필수였지만, 이제는 많은 경우 표준 라이브러리로 충분합니다.

### Chi

```go
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Get("/api/users/{id}", handleGetUser)
```

**선택 기준**: 미들웨어 생태계가 필요할 때, 표준 라이브러리와 100% 호환되는 라우터를 원할 때. Chi는 `net/http`의 얇은 래퍼이며 `http.Handler`를 그대로 사용합니다.

### Gin

```go
r := gin.New()
r.GET("/api/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})
```

**선택 기준**: 빠른 프로토타이핑, 바인딩/유효성 검사 기능이 내장되어야 할 때. Gin은 독자적인 컨텍스트 타입을 사용하므로 표준 미들웨어와 조합이 어렵습니다.

### Echo

```go
e := echo.New()
e.GET("/api/users/:id", func(c echo.Context) error {
    return c.JSON(200, map[string]string{"id": c.Param("id")})
})
```

**선택 기준**: Gin과 유사하지만 더 Go스러운 에러 처리를 원할 때.

### 요약

```
표준 라이브러리 → Chi → Gin/Echo
(의존성 적음)           (기능 많음)
```

Go 1.22+를 사용할 수 있다면 대부분의 API 서버는 표준 라이브러리 + Chi 조합으로 충분합니다.

---

## 데이터베이스 전략

### pgx (직접 쿼리)

```go
import "github.com/jackc/pgx/v5/pgxpool"

pool, _ := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
row := pool.QueryRow(ctx, "SELECT id, name FROM users WHERE id = $1", id)
```

**선택 기준**: 성능이 최우선, SQL을 완전히 제어하고 싶을 때. PostgreSQL 전용.

### sqlc (코드 생성)

```sql
-- query.sql
-- name: GetUser :one
SELECT id, name, email FROM users WHERE id = $1;
```

```go
// sqlc가 생성한 타입 안전 코드
user, err := queries.GetUser(ctx, id)
```

**선택 기준**: SQL을 직접 작성하되 타입 안전성을 원할 때. SQL → Go 코드를 자동 생성합니다. Phase 5에서 다룹니다.

### GORM (ORM)

```go
db.First(&user, id)
db.Create(&User{Name: "홍길동", Email: "hong@example.com"})
```

**선택 기준**: 빠른 프로토타이핑, 간단한 CRUD. 복잡한 쿼리는 Raw SQL로 빠져나올 수 있어야 합니다.

---

## 프로덕션 도구 체인

| 도구 | 역할 | 설치 |
|------|------|------|
| `golangci-lint` | 린터 집합체 (50+ 린터) | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| `air` | 핫 리로드 (개발용) | `go install github.com/air-verse/air@latest` |
| `delve` | 디버거 | `go install github.com/go-delve/delve/cmd/dlv@latest` |
| `govulncheck` | 보안 취약점 검사 | `go install golang.org/x/vuln/cmd/govulncheck@latest` |
| `GoReleaser` | 크로스 플랫폼 릴리스 자동화 | `go install github.com/goreleaser/goreleaser@latest` |

### golangci-lint 기본 설정 (`.golangci.yml`)

```yaml
linters:
  enable:
    - errcheck      # 에러 반환값 무시 검출
    - gosimple      # 단순화 제안
    - govet         # go vet 검사
    - ineffassign   # 불필요한 대입 검출
    - staticcheck   # 정적 분석
    - unused        # 미사용 코드 검출

linters-settings:
  errcheck:
    check-type-assertions: true
```

### Air 설정 (`.air.toml`)

```toml
[build]
cmd = "go build -o ./tmp/main ."
bin = "./tmp/main"
include_ext = ["go"]
exclude_dir = ["vendor", "tmp"]
```

---

## Phase 4 학습 순서

```
01-enhanced-routing   Go 1.22 ServeMux: 메서드 패턴, 경로 와일드카드
        ↓
02-middleware          미들웨어 체인: 로깅, 복구, CORS, 요청 ID
        ↓
03-generics            타입 파라미터, 제약조건, slices/maps/cmp
        ↓
04-slog               log/slog: 구조화된 로깅, JSONHandler, LevelVar
        ↓
05-iterators           iter.Seq, range-over-function (Go 1.23)
        ↓
06-docker              멀티스테이지 빌드, scratch 이미지, GoReleaser
```

각 토픽은 독립적으로 학습할 수 있지만, `01 → 02`는 연속성이 강합니다. 라우팅과 미들웨어는 함께 실제 HTTP 서버를 구성합니다.

---

## Python / Java 대응표

| 개념 | Python | Java | Go |
|------|--------|------|----|
| HTTP 라우터 | FastAPI / Flask | Spring MVC | `net/http` ServeMux (1.22+) |
| 미들웨어 | WSGI Middleware | Servlet Filter | `func(http.Handler) http.Handler` |
| 제네릭 | `typing.Generic[T]` | `<T>` (erasure) | `[T any]` (monomorphization) |
| 구조화된 로깅 | `structlog` / `loguru` | Logback + JSON | `log/slog` |
| 이터레이터 | `generator` / `yield` | `Stream<T>` | `iter.Seq[V]` (1.23+) |
| 컨테이너화 | Dockerfile (대형 이미지) | Dockerfile (JVM 포함) | scratch 기반 5MB 이미지 |

---

## 추천 학습 자료

- **Go 공식 블로그** (https://go.dev/blog) — 제네릭, slog, 이터레이터 도입 배경 글이 있습니다.
- **Ardan Labs Blog** (https://www.ardanlabs.com/blog) — 프로덕션 Go 패턴의 최고 출처입니다.
- **100 Go Mistakes** (Teiva Harsanyi) — 실수 패턴 100가지. HTTP, 동시성, 성능 챕터가 이 Phase에 해당합니다.
- **go.dev/doc/go1.22** ~ **go.dev/doc/go1.24** — 버전별 릴리스 노트. 짧고 명확합니다.
