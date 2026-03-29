# Phase 5: 고급 시스템

**기간:** 16~24주차+
**목표:** 프로덕션 인프라 도구 개발 역량을 갖춘다
**디렉토리:** `go-curriculum/phase5-advanced/`

---

## 레슨 목록

| # | 디렉토리 | 주제 | 핵심 개념 |
|---|---------|------|----------|
| 1 | `01-grpc` | gRPC | Protocol Buffers, 서버/클라이언트, 스트리밍 |
| 2 | `02-cobra-cli` | Cobra CLI | CLI 도구 설계, 명령/플래그/인자 |
| 3 | `03-profiling` | 프로파일링 | pprof, CPU/메모리 프로파일, 벤치마크 |
| 4 | `04-testing-advanced` | 고급 테스트 | 모킹, 퍼징, 빌드 태그, 테스트 더블 |
| 5 | `05-kubernetes-basics` | Kubernetes 기초 | client-go, API 서버, 리소스 관리 |

## 레슨 파일 경로

```
go-curriculum/phase5-advanced/
├── 01-grpc/THEORY.md           + client/ + proto/ + server/
├── 02-cobra-cli/THEORY.md      + main.go + cmd/
├── 03-profiling/THEORY.md      + main.go + heavy/
├── 04-testing-advanced/THEORY.md + fuzz/ + service/
└── 05-kubernetes-basics/THEORY.md + main.go
```

---

## 과제 목록

| # | 디렉토리 | 주제 | 난이도 | 핵심 |
|---|---------|------|--------|------|
| A1 | `assignments/a1-cli-tool` | 파일 검색 CLI | ★★★☆☆ | Cobra 설계 |
| A2 | `assignments/a2-load-tester` | HTTP 부하 테스터 | ★★★★☆ | 동시성 + CLI |
| A3 | `assignments/a3-log-analyzer` | 로그 분석기 | ★★★½☆ | 파이프라인 + 파일 I/O |
| A4 | `assignments/a4-plugin-system` | 플러그인 시스템 | ★★★★½ | 인터페이스 + 동적 로딩 |
| A5 | `assignments/a5-distributed-lock` | 분산 잠금 | ★★★★★ | 펜싱 토큰, 교착 감지 |

## 과제 파일 경로

```
go-curriculum/phase5-advanced/assignments/
├── a1-cli-tool/
├── a2-load-tester/
├── a3-log-analyzer/
├── a4-plugin-system/
└── a5-distributed-lock/
```
각 과제: README.md + 구현 파일(.go) + 테스트 파일(_test.go) + solution/

---

## 프로젝트

**프로젝트: CLI 배포 도구**
- 디렉토리: `go-curriculum/phase5-advanced/project-cli-deploy-tool/`
- 목표: Cobra + Viper 기반 Docker 애플리케이션 배포 관리 도구
- 핵심 학습: CLI 설계, 설정 관리, Docker 통합
- 서브 패키지: cmd/, internal/

---

## Go가 인프라 언어인 이유

Phase 5에서 체감할 Go의 강점:
- Kubernetes, Docker, Prometheus, etcd, Terraform — 모두 Go
- CNCF 프로젝트의 75%+ Go 작성
- 단일 바이너리 배포, 빠른 컴파일, 낮은 메모리
- 크로스 컴파일: `GOOS=linux GOARCH=amd64 go build`
