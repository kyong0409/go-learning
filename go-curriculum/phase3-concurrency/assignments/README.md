# Phase 3 동시성 과제

Go 동시성 프로그래밍 실력을 기르기 위한 6가지 과제입니다.
각 과제는 독립적으로 풀 수 있으며, 난이도 순으로 정렬되어 있습니다.

## 과제 목록

| 과제 | 주제 | 핵심 개념 | 난이도 |
|------|------|-----------|--------|
| [A1 - 데이터 처리 파이프라인](./a1-pipeline/) | 채널 파이프라인 | 채널, 고루틴, Context 취소 | ★★☆ |
| [A2 - 워커 풀 파일 처리](./a2-worker-pool/) | 워커 풀 패턴 | sync.WaitGroup, 에러 수집, 진행률 | ★★★ |
| [A3 - 채널 기반 채팅 서버](./a3-chat-server/) | 채널 메시지 라우팅 | 채널, select, 브로드캐스트 | ★★★ |
| [A4 - 병렬 API 호출기](./a4-fanout-errgroup/) | Fan-out / Fan-in | errgroup, 동시성 제한, 부분 실패 | ★★★½ |
| [A5 - 토큰 버킷 속도 제한기](./a5-rate-limiter/) | 속도 제한 알고리즘 | sync.Mutex, 시간 계산, 블로킹 대기 | ★★★★½ |
| [A6 - 우아한 종료 HTTP 서버](./a6-graceful-server/) | 서버 생명주기 관리 | graceful shutdown, 미들웨어, 훅 | ★★★★½ |

## 권장 학습 순서

```
A1 (★★) → A4 (★★★½) → A2 (★★★) → A3 (★★★) → A5 (★★★★½) → A6 (★★★★½)
```

- **A1 → A4**: 채널 파이프라인을 익힌 뒤, errgroup으로 Fan-out/Fan-in 패턴 학습
- **A2 → A3**: 워커 풀로 동시성 제어를 익힌 뒤, 채널 기반 메시지 라우팅 학습
- **A5 → A6**: 저수준 동기화 프리미티브(Mutex, 시간 계산)를 다룬 뒤, 서버 생명주기 관리 학습

## 과제 구조

각 과제 디렉터리는 다음 파일을 포함합니다:

```
a1-pipeline/
├── README.md           # 과제 설명 및 요구사항
├── pipeline.go         # 구현할 함수 시그니처 (스켈레톤)
├── pipeline_test.go    # 자동 채점 테스트
└── solution/
    └── pipeline.go     # 참고 솔루션 (풀기 전에 보지 마세요!)
```

## 채점 방법

```bash
# 테스트 실행 (점수 확인)
cd a1-pipeline
go test -v .

# 레이스 디텍터 포함 테스트 (추가 점수)
go test -race -v .

# 채점 테스트만 실행
go test -v -run TestGrade .

# 전체 과제 테스트
cd ..
for dir in a1-pipeline a2-worker-pool a3-chat-server a4-fanout-errgroup a5-rate-limiter a6-graceful-server; do
    echo "=== $dir ==="
    cd $dir && go test -v -run TestGrade . && cd ..
done
```

## 학습 순서 권장

1. `06-patterns/` 예제를 먼저 학습하세요.
2. 각 과제의 `README.md`를 읽고 요구사항을 파악하세요.
3. 스켈레톤 파일의 TODO 주석을 채워 구현하세요.
4. 테스트를 통과시키세요.
5. `-race` 플래그로 레이스 컨디션이 없는지 확인하세요.
6. 솔루션과 비교해 개선점을 찾으세요.

## 제출 기준

- 모든 테스트 통과 (`go test -v .`)
- 레이스 컨디션 없음 (`go test -race .`)
- 고루틴 누수 없음 (테스트가 검증)
- Context 취소 올바르게 처리

## 난이도 상세

| 과제 | 별점 | 핵심 어려움 |
|------|------|------------|
| A1 파이프라인 | ★★☆☆☆ | 채널 닫기 타이밍, Context 전파 |
| A2 워커 풀 | ★★★☆☆ | 채널 생명주기, WaitGroup 조율 |
| A3 채팅 서버 | ★★★☆☆ | 단일 루프 직렬화, 요청/응답 채널 |
| A4 Fan-out errgroup | ★★★½☆ | errgroup.SetLimit, 순서 보장, 부분 실패 |
| A5 토큰 버킷 | ★★★★½ | 시간 기반 토큰 계산, Wait 정확도 |
| A6 우아한 서버 | ★★★★½ | 서버 생명주기, 미들웨어 체인, 신호 처리 |
