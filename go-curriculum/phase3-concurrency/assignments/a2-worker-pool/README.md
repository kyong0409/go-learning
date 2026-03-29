# 과제 A2: 동시성 파일 처리 워커 풀

## 과제 설명

설정 가능한 워커 수로 작업을 병렬 처리하는 워커 풀을 구현하세요.
파일 처리를 시뮬레이션하며, 실제 I/O 없이 처리 로직에 집중합니다.

## 요구사항

### Pool 구조체와 메서드

```go
type Pool struct { ... }

func NewPool(workers int) *Pool
func (p *Pool) Start(ctx context.Context)
func (p *Pool) Submit(job Job) error
func (p *Pool) Progress() <-chan ProgressEvent
func (p *Pool) Results() <-chan Result
func (p *Pool) Shutdown()
func (p *Pool) Wait() []Result
```

### Job과 Result 타입

```go
type Job struct {
    ID       int
    Filename string
    SizeKB   int
}

type Result struct {
    Job      Job
    Output   string
    Duration time.Duration
    Err      error
}

type ProgressEvent struct {
    Completed int
    Total     int
    Current   Job
}
```

## 기능 요구사항

1. **설정 가능한 워커 수**: `NewPool(n)`으로 워커 수 지정
2. **우아한 종료**: `Shutdown()` 호출 시 진행 중인 작업 완료 후 종료
3. **Context 취소**: ctx 취소 시 즉시 모든 워커 중단
4. **에러 수집**: 실패한 작업의 에러를 Result에 저장
5. **진행률 보고**: `Progress()` 채널로 완료된 작업 수 실시간 전달
6. **중복 Submit 방지**: `Shutdown()` 후 `Submit()` 호출 시 에러 반환

## 실행 방법

```bash
cd a2-worker-pool
go test -v .
go test -race -v .
go test -v -run TestGrade .
```

## 채점 기준 (100점)

| 항목 | 점수 | 설명 |
|------|------|------|
| 기본 처리 | 25점 | 모든 작업 처리, 올바른 결과 |
| 워커 수 제한 | 20점 | 지정한 수만큼만 동시 실행 |
| Context 취소 | 20점 | 취소 시 즉시 중단 |
| 우아한 종료 | 15점 | Shutdown 후 진행 중 완료 |
| 진행률 보고 | 10점 | Progress 채널 정상 작동 |
| 고루틴 누수 없음 | 10점 | 모든 고루틴 정상 종료 |

## 힌트

- `Start()`에서 워커 고루틴들을 시작하세요.
- `Submit()`은 작업 채널에 Job을 전송합니다.
- `Shutdown()`은 작업 채널을 닫아 워커들이 range로 종료되게 합니다.
- `sync.WaitGroup`으로 모든 워커 완료를 추적하세요.
- Progress 채널은 버퍼를 충분히 주거나 비동기로 전송하세요.
