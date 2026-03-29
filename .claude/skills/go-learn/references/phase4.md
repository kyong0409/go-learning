# Phase 4: 프로덕션 Go

**기간:** 10~16주차
**목표:** 모던 Go (1.22+) 기능으로 프로덕션급 코드를 작성한다
**디렉토리:** `go-curriculum/phase4-production/`

---

## 레슨 목록

| # | 디렉토리 | 주제 | 핵심 개념 |
|---|---------|------|----------|
| 1 | `01-enhanced-routing` | 향상된 라우팅 | Go 1.22+ ServeMux, 메서드 패턴, 경로 파라미터 |
| 2 | `02-middleware` | 미들웨어 | 미들웨어 체인, 함수 래핑, http.Handler 패턴 |
| 3 | `03-generics` | 제네릭 | 타입 파라미터, 제약 조건, Go 1.18+ |
| 4 | `04-slog` | 구조화된 로깅 | log/slog, 핸들러, 속성, 그룹 |
| 5 | `05-iterators` | 이터레이터 | range-over-function, Go 1.23+ iter 패키지 |
| 6 | `06-docker` | Docker | 멀티스테이지 빌드, 최적화, 배포 |

## 레슨 파일 경로

```
go-curriculum/phase4-production/
├── 01-enhanced-routing/THEORY.md  + main.go
├── 02-middleware/THEORY.md        + main.go
├── 03-generics/THEORY.md         + main.go
├── 04-slog/THEORY.md             + main.go
├── 05-iterators/THEORY.md        + main.go
└── 06-docker/THEORY.md           + main.go (+ Dockerfile)
```

---

## 과제 목록

| # | 디렉토리 | 주제 | 난이도 | 핵심 |
|---|---------|------|--------|------|
| A1 | `assignments/a1-generic-collection` | 제네릭 컬렉션 | ★★★☆☆ | 제네릭 자료구조 |
| A2 | `assignments/a2-middleware-chain` | 미들웨어 체인 | ★★★★☆ | HTTP 미들웨어 |
| A3 | `assignments/a3-rest-api` | 도서 관리 REST API | ★★★½☆ | 종합 API 개발 |
| A4 | `assignments/a4-metrics-collector` | 메트릭 수집기 | ★★★★½ | Prometheus 패턴 |
| A5 | `assignments/a5-cache` | TTL/LRU 캐시 | ★★★★½ | 제네릭 + 동시성 |

## 과제 파일 경로

```
go-curriculum/phase4-production/assignments/
├── a1-generic-collection/
├── a2-middleware-chain/
├── a3-rest-api/
├── a4-metrics-collector/
└── a5-cache/
```
각 과제: README.md + 구현 파일(.go) + 테스트 파일(_test.go) + solution/

---

## 프로젝트

**프로젝트: URL 단축 서비스**
- 디렉토리: `go-curriculum/phase4-production/project-url-shortener/`
- 목표: Go 1.22+ 표준 라이브러리 기반 URL 단축기
- 핵심 학습: 향상된 라우팅, 미들웨어, 구조화된 로깅
- 서브 패키지: cmd/, internal/, migrations/

---

## 모던 Go 타임라인

이 Phase에서 다루는 최신 Go 기능:
- Go 1.18 (2022.03) — 제네릭 도입
- Go 1.21 (2023.08) — slog, slices, maps 패키지
- Go 1.22 (2024.02) — 향상된 라우팅 (ServeMux 패턴 매칭)
- Go 1.23 (2024.08) — range-over-function 이터레이터
