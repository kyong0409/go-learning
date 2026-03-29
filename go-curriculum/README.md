# Go 완전 학습 커리큘럼 - 실습 자료

경험 있는 개발자를 위한 Go 학습 커리큘럼의 실습 예제, 프로젝트, 과제 모음입니다.

## 디렉토리 구조

```
go-curriculum/
├── phase1-basics/              # Phase 1: Go 기초 (1~3주차)
│   ├── 01-hello/               # Hello World, fmt 패키지
│   ├── 02-variables/           # 변수, 상수, 타입
│   ├── 03-control-flow/        # for, switch, if
│   ├── 04-functions/           # 다중 반환, 클로저, 일급 함수
│   ├── 05-data-structures/     # 배열, 슬라이스, 맵
│   ├── 06-strings/             # UTF-8, rune, strings 패키지
│   ├── 07-pointers/            # 포인터 기초
│   ├── 08-errors/              # 에러 처리 기초
│   ├── 09-packages/            # 패키지, 모듈
│   ├── project-todo-cli/       # [프로젝트] CLI 할일 관리자
│   └── assignments/            # 과제 3개
│
├── phase2-structs-interfaces/  # Phase 2: 구조체, 인터페이스 (3~6주차)
│   ├── 01-structs/             # 구조체, JSON 태그
│   ├── 02-methods/             # 메서드, 리시버
│   ├── 03-interfaces/          # 인터페이스, 타입 단언
│   ├── 04-composition/         # 컴포지션, 임베딩
│   ├── 05-error-handling/      # 에러 래핑, errors.Is/As
│   ├── 06-defer-panic/         # defer, panic/recover
│   ├── 07-testing/             # 테이블 주도 테스트, 벤치마크
│   ├── project-bookmark-api/   # [프로젝트] HTTP JSON API
│   └── assignments/            # 과제 3개
│
├── phase3-concurrency/         # Phase 3: 동시성 (6~10주차)
│   ├── 01-goroutines/          # 고루틴
│   ├── 02-channels/            # 채널
│   ├── 03-select/              # select 문
│   ├── 04-sync/                # sync 패키지
│   ├── 05-context/             # context 패키지
│   ├── 06-patterns/            # 동시성 패턴 6종
│   ├── 07-race-detector/       # 레이스 감지기
│   ├── project-web-scraper/    # [프로젝트] 동시성 웹 스크래퍼
│   └── assignments/            # 과제 6개 (중간 단계 포함)
│
├── phase4-production/          # Phase 4: 프로덕션 Go (10~16주차)
│   ├── 01-enhanced-routing/    # Go 1.22+ 라우팅
│   ├── 02-middleware/          # 미들웨어 패턴
│   ├── 03-generics/           # 제네릭
│   ├── 04-slog/               # 구조화된 로깅
│   ├── 05-iterators/          # Range-over-function
│   ├── 06-docker/             # Docker 멀티스테이지 빌드
│   ├── project-url-shortener/ # [프로젝트] URL 단축 서비스
│   └── assignments/           # 과제 5개 (고난이도 추가)
│
├── phase5-advanced/           # Phase 5: 고급 시스템 (16~24주차+)
│   ├── 01-grpc/               # gRPC 서버/클라이언트
│   ├── 02-cobra-cli/          # Cobra CLI 도구
│   ├── 03-profiling/          # pprof, 벤치마크
│   ├── 04-testing-advanced/   # 모킹, 퍼징, 빌드 태그
│   ├── 05-kubernetes-basics/  # client-go 기초
│   ├── project-cli-deploy-tool/ # [프로젝트] CLI 배포 도구
│   └── assignments/           # 과제 5개 (분산 시스템 추가)
│
├── phase6-opensource/          # Phase 6: K8s 생태계 오픈소스 딥다이브
│   ├── 01-k8s-patterns/       # K8s Controller, Informer, WorkQueue
│   ├── 02-prometheus-patterns/# Prometheus Registry, Collector
│   ├── 03-etcd-patterns/      # etcd Watch, MVCC, Lease
│   ├── 04-docker-patterns/    # Docker 컨테이너 라이프사이클
│   ├── 05-terraform-patterns/ # Terraform DAG, Plan/Apply
│   └── assignments/           # 과제 6개 (전부 ★4~5 고난이도)
│
├── grade.sh                   # 전체 과제 채점 스크립트
└── README.md                  # 이 파일
```

## 사용법

### 예제 실행

```bash
# 각 예제 디렉토리에서
cd phase1-basics/01-hello
go run main.go

# 패키지가 있는 예제
cd phase1-basics/09-packages
go run main.go
```

### 프로젝트 실행

```bash
# 프로젝트 디렉토리에서
cd phase1-basics/project-todo-cli
go run main.go add "Go 공부하기"
go run main.go list
```

### 과제 진행 방법

1. 과제 디렉토리의 `README.md`를 읽고 문제를 파악합니다
2. 스텁 파일(TODO가 있는 파일)에 구현을 작성합니다
3. 테스트를 실행하여 채점합니다:

```bash
cd phase1-basics/assignments/a1-calculator
go test -v
```

4. 채점 결과가 출력됩니다:

```
=== 채점 결과 ===
통과: 20/25
점수: 72/90
```

5. 막히면 `solution/` 디렉토리의 참고 답안을 확인합니다

### 전체 채점

```bash
# 모든 과제 한번에 채점
bash grade.sh

# 특정 Phase만 채점
bash grade.sh phase1
bash grade.sh phase3
bash grade.sh phase6
```

## 과제 목록

### Phase 1: Go 기초 (3개)

| 과제 | 주제 | 난이도 | 만점 |
|------|------|--------|------|
| A1 계산기 | 함수, 에러 처리 | ★☆☆☆☆ | 90 |
| A2 슬라이스 유틸리티 | 슬라이스, 고차함수 | ★★☆☆☆ | 90 |
| A3 단어 빈도수 | 파일 I/O, 맵, 정렬 | ★★½☆☆ | 100 |

### Phase 2: 구조체와 인터페이스 (3개)

| 과제 | 주제 | 난이도 | 만점 |
|------|------|--------|------|
| A1 Shape 인터페이스 | 인터페이스, 다형성 | ★★☆☆☆ | 100 |
| A2 은행 계좌 에러 처리 | 커스텀 에러, 래핑 | ★★★☆☆ | 100 |
| A3 JSON 설정 파서 | 구조체, JSON, 검증 | ★★★☆☆ | 100 |

### Phase 3: 동시성 (6개)

| 과제 | 주제 | 난이도 | 만점 |
|------|------|--------|------|
| A1 데이터 파이프라인 | 채널, 파이프라인 | ★★★☆☆ | 100 |
| A2 워커 풀 | 고루틴, context, WaitGroup | ★★★★☆ | 100 |
| A3 채팅 서버 | 이벤트 루프, 채널 라우팅 | ★★★★★ | 100 |
| **A4 병렬 API 호출기** | **errgroup, 순서 보존** | **★★★½☆** | **100** |
| **A5 토큰 버킷 속도 제한기** | **sync, 시간 기반 로직** | **★★★★½** | **100** |
| **A6 우아한 종료 서버** | **시그널, 미들웨어, 헬스체크** | **★★★★½** | **100** |

### Phase 4: 프로덕션 Go (5개)

| 과제 | 주제 | 난이도 | 만점 |
|------|------|--------|------|
| A1 제네릭 컬렉션 | 제네릭, 자료구조 | ★★★☆☆ | 100 |
| A2 미들웨어 체인 | HTTP, 미들웨어 | ★★★★☆ | 100 |
| A3 도서 관리 REST API | REST, 통합 테스트 | ★★★½☆ | 100 |
| **A4 메트릭 수집기** | **Prometheus 패턴, 동시성** | **★★★★½** | **100** |
| **A5 TTL/LRU 캐시** | **제네릭, 동시성, 퇴거 정책** | **★★★★½** | **100** |

### Phase 5: 고급 시스템 (5개)

| 과제 | 주제 | 난이도 | 만점 |
|------|------|--------|------|
| A1 파일 검색 CLI | Cobra, CLI 설계 | ★★★☆☆ | 100 |
| A2 HTTP 부하 테스터 | 동시성, 통계 | ★★★★☆ | 100 |
| A3 로그 분석기 | 스트리밍, JSON 파싱 | ★★★½☆ | 100 |
| **A4 플러그인 시스템** | **인터페이스, 의존성 정렬** | **★★★★½** | **100** |
| **A5 분산 잠금** | **펜싱 토큰, 대기 큐, 교착 감지** | **★★★★★** | **100** |

### Phase 6: K8s 오픈소스 딥다이브 (6개) - NEW

| 과제 | 원본 프로젝트 | 난이도 | 만점 |
|------|-------------|--------|------|
| **A1 리소스 컨트롤러** | **Kubernetes Controller** | **★★★★½** | **100** |
| **A2 인포머/리스터** | **client-go Informer** | **★★★★★** | **100** |
| **A3 메트릭 레지스트리** | **Prometheus** | **★★★★☆** | **100** |
| **A4 DAG 태스크 실행기** | **Terraform** | **★★★★★** | **100** |
| **A5 이벤트 감시 시스템** | **etcd Watch** | **★★★★½** | **100** |
| **A6 오퍼레이터 시뮬레이터** | **K8s Operator (캡스톤)** | **★★★★★** | **100** |

### 난이도 분포

```
★☆☆☆☆  ■                                    (1개)
★★☆☆☆  ■■                                   (2개)
★★½~★★★ ■■■■■■■                              (7개)
★★★½☆   ■■■■                                 (4개)
★★★★☆   ■■■■                                 (4개)
★★★★½   ■■■■■■                               (6개)
★★★★★   ■■■■■                                (5개)
```

**총 28개 과제 | 총 만점: 2,780점**

## 추천 학습 순서

Phase 1~5는 순서대로 진행합니다. Phase 6은 Phase 3~5 완료 후 도전합니다.
Phase 3부터 A 번호 순서대로 진행하되, 어려우면 다음 Phase로 넘어간 뒤 돌아와도 됩니다.

```
Phase 1 (★1~2.5) → Phase 2 (★2~3) → Phase 3 (★3~5) → Phase 4 (★3~4.5) → Phase 5 (★3~5) → Phase 6 (★4~5)
```

## 과제 제출 방법

과제를 구현한 후 `go test -v` 결과를 공유해주시면 채점하겠습니다.
혹은 구현한 코드 파일을 직접 공유해주셔도 됩니다.
