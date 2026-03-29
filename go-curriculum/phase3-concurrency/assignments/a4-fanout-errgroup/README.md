# 과제 A4: 병렬 API 호출기 (Fan-out / Fan-in with errgroup)

## 과제 설명

`errgroup` 패키지를 사용하여 여러 URL을 병렬로 가져오는 함수를 구현하세요.
워커 풀(A2)보다 간단하지만 `errgroup`과 에러 집계 패턴을 배울 수 있습니다.

## 요구사항

### FetchResult 타입

```go
type FetchResult struct {
    URL        string
    Body       string
    StatusCode int
    Error      error
}
```

### 구현할 함수

| 함수 | 설명 |
|------|------|
| `FetchAll(ctx, urls, maxConcurrency)` | 여러 URL을 병렬로 가져옴. 결과는 입력 순서와 동일하게 반환 |
| `FetchWithRetry(ctx, url, maxRetries)` | 실패 시 지수 백오프(exponential backoff)로 재시도 |

### FetchAll 세부 요구사항

1. **순서 보장**: 결과 슬라이스의 순서가 입력 `urls`의 순서와 동일해야 합니다.
2. **부분 실패 허용**: 일부 URL이 실패해도 나머지 결과를 반환합니다. 실패한 항목은 `FetchResult.Error` 필드에 에러를 담습니다.
3. **동시성 제한**: `maxConcurrency`로 동시에 실행할 최대 고루틴 수를 제한합니다. `errgroup.SetLimit`을 사용하세요.
4. **Context 취소**: `ctx`가 취소되면 진행 중이지 않은 요청은 시작하지 않습니다.

### FetchWithRetry 세부 요구사항

1. **지수 백오프**: 재시도 간격은 `100ms → 200ms → 400ms → ...` 로 두 배씩 증가합니다.
2. **최대 재시도**: `maxRetries`번 재시도 후에도 실패하면 마지막 에러를 반환합니다.
3. **Context 취소**: 재시도 대기 중에도 `ctx`가 취소되면 즉시 종료합니다.

## 실행 방법

```bash
cd a4-fanout-errgroup

# 의존성 설치
go mod tidy

# 구현 후 테스트
go test -v .

# 레이스 디텍터 포함
go test -race -v .

# 점수 확인
go test -v -run TestGrade .
```

## 채점 기준 (100점)

| 항목 | 점수 | 설명 |
|------|------|------|
| 기본 Fetch | 20점 | 단일 URL 가져오기 성공 |
| 병렬 Fetch | 20점 | 여러 URL 동시 처리, 순서 보장 |
| 부분 실패 처리 | 20점 | 일부 실패해도 나머지 결과 반환 |
| 동시성 제한 | 20점 | maxConcurrency 초과하지 않음 |
| Context 취소 | 10점 | 취소 시 즉시 중단 |
| 재시도 로직 | 10점 | FetchWithRetry 지수 백오프 |

## 힌트

- `golang.org/x/sync/errgroup` 패키지를 사용하세요.
- `errgroup.WithContext(ctx)`로 그룹을 생성하고 `g.SetLimit(maxConcurrency)`로 제한을 설정합니다.
- 순서 보장을 위해 `results := make([]FetchResult, len(urls))`로 미리 슬라이스를 할당하고 인덱스로 접근하세요.
- 각 고루틴은 `results[i]`에 직접 써도 됩니다 (각기 다른 인덱스에 쓰므로 레이스 없음).
- `FetchWithRetry`는 `time.Sleep` 대신 `select { case <-time.After(delay): case <-ctx.Done(): }`를 사용하세요.
- 테스트는 `httptest.NewServer`로 실제 HTTP 서버를 시뮬레이션합니다.
