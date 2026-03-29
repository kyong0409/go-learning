# Phase 3: 동시성 — 가장 가파른 학습 곡선

**기간:** 6~10주차
**목표:** Go의 동시성 모델(고루틴, 채널, sync)을 마스터한다
**디렉토리:** `go-curriculum/phase3-concurrency/`

---

## 레슨 목록

| # | 디렉토리 | 주제 | 핵심 개념 |
|---|---------|------|----------|
| 1 | `01-goroutines` | 고루틴 | go 키워드, GMP 스케줄러, 경량 스레드 |
| 2 | `02-channels` | 채널 | chan T, 버퍼, 방향성, 닫기 |
| 3 | `03-select` | select 문 | 다중화, 타임아웃, 논블로킹 |
| 4 | `04-sync` | sync 패키지 | Mutex, RWMutex, WaitGroup, Once, Pool |
| 5 | `05-context` | context | WithCancel, WithTimeout, 취소 전파 |
| 6 | `06-patterns` | 동시성 패턴 | 워커풀, 파이프라인, Fan-out/in, 세마포어, 속도제한, errgroup |
| 7 | `07-race-detector` | 레이스 감지 | go test -race, 데이터 레이스 디버깅 |

## 레슨 파일 경로

```
go-curriculum/phase3-concurrency/
├── 01-goroutines/THEORY.md     + main.go
├── 02-channels/THEORY.md       + main.go
├── 03-select/THEORY.md         + main.go
├── 04-sync/THEORY.md           + main.go
├── 05-context/THEORY.md        + main.go
├── 06-patterns/THEORY.md       + main.go
│   ├── pipeline/
│   ├── fan_out_fan_in/
│   ├── worker_pool/
│   ├── semaphore/
│   ├── rate_limiter/
│   └── errgroup/
└── 07-race-detector/THEORY.md  + main.go
```

---

## 과제 목록

| # | 디렉토리 | 주제 | 난이도 | 핵심 |
|---|---------|------|--------|------|
| A1 | `assignments/a1-pipeline` | 데이터 파이프라인 | ★★★☆☆ | 채널, 파이프라인 패턴 |
| A2 | `assignments/a2-worker-pool` | 워커 풀 | ★★★★☆ | 고루틴 풀, 작업 분배 |
| A3 | `assignments/a3-chat-server` | 채팅 서버 | ★★★★★ | 전체 동시성 통합 |
| A4 | `assignments/a4-fanout-errgroup` | 병렬 API 호출기 | ★★★½☆ | errgroup, 순서 보존 |
| A5 | `assignments/a5-rate-limiter` | 토큰 버킷 속도 제한기 | ★★★★½ | 속도 제한, 동시 접근 |
| A6 | `assignments/a6-graceful-server` | 우아한 종료 서버 | ★★★★½ | 시그널, 미들웨어, context |

## 과제 파일 경로

```
go-curriculum/phase3-concurrency/assignments/
├── a1-pipeline/
├── a2-worker-pool/
├── a3-chat-server/
├── a4-fanout-errgroup/
├── a5-rate-limiter/
└── a6-graceful-server/
```
각 과제: README.md + 구현 파일(.go) + 테스트 파일(_test.go) + solution/

---

## 프로젝트

**프로젝트: 동시성 웹 스크래퍼**
- 디렉토리: `go-curriculum/phase3-concurrency/project-web-scraper/`
- 목표: 여러 URL을 동시에 크롤링하는 웹 스크래퍼
- 핵심 학습: 고루틴 + 채널 실전 활용
- 서브 패키지: scraper/

---

## 핵심 개념

Phase 3에서 특히 강조할 포인트:
- **"동시성은 병렬성이 아니다"**: 구조(concurrency) vs 실행(parallelism)
- **CSP 모델**: "메모리를 공유하지 말고, 통신으로 메모리를 공유하라"
- **채널 우선, 뮤텍스는 보조**: 가능하면 채널로, 필요할 때만 sync
- **context는 필수**: 모든 장기 실행 작업은 context를 존중해야 한다
- **-race 플래그**: 개발 중 항상 `go test -race`를 사용하라
