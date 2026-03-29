# 과제 A5: 분산 잠금 시스템 구현 (로컬 시뮬레이션)

**난이도**: ★★★★★ (5/5)
**예상 소요 시간**: 6~9시간

## 과제 설명

etcd나 Redis의 분산 잠금 패턴을 인-프로세스에서 시뮬레이션합니다.
펜싱 토큰, TTL 만료, 대기 큐, 데드락 감지를 구현하여
실제 분산 시스템에서 사용되는 조율 패턴을 이해합니다.

## 구현할 타입

### `Lock` — 잠금 토큰

```go
type Lock struct {
    Key          string        // 잠금 대상 리소스 키
    Owner        string        // 잠금 보유자 식별자
    Token        uint64        // 펜싱 토큰 (단조 증가)
    AcquiredAt   time.Time     // 획득 시각
    ExpiresAt    time.Time     // 만료 시각 (TTL 기반)
}

// 잠금이 현재 유효한지 (만료되지 않았는지) 확인
func (l *Lock) IsValid() bool
```

### `LockService` — 잠금 서비스

```go
svc := NewLockService()

// 잠금 획득 (ctx 취소 시 대기 중단)
// - 이미 잠긴 경우: 해제될 때까지 대기
// - TTL 만료된 잠금은 자동으로 해제 처리
lock, err := svc.Acquire(ctx, key, owner, ttl)

// 잠금 해제
// - lock.Token이 현재 보유 토큰과 다르면 error (펜싱 실패)
err = svc.Release(lock)

// 잠금 갱신 (TTL 연장)
// - 유효한 잠금만 갱신 가능
// - 갱신된 Lock 반환 (ExpiresAt 업데이트)
lock, err = svc.Refresh(lock, newTTL)

// 잠금 보유 여부 확인
owned := svc.IsLocked(key)

// 현재 잠금 정보 조회 (없거나 만료 시 nil, false)
lock, ok := svc.GetLock(key)
```

## 핵심 개념 구현

### 펜싱 토큰 (Fencing Token)
- 잠금을 획득할 때마다 전역적으로 단조 증가하는 토큰을 발급합니다.
- `Release`/`Refresh` 시 토큰을 검증합니다.
- 만료된 잠금을 다른 소유자가 재획득하면 **새 토큰**이 발급됩니다.
- 이전 토큰으로 `Release`를 시도하면 `ErrInvalidToken` 오류를 반환합니다.

### TTL 만료
- 잠금은 TTL이 지나면 자동으로 만료됩니다.
- `Acquire` 호출 시 현재 잠금이 만료되었으면 즉시 새 잠금을 발급합니다.
- 백그라운드 고루틴이 주기적으로 만료된 잠금을 정리합니다.

### 대기 큐 (Wait Queue)
- 잠금이 이미 점유된 경우 `Acquire`는 블로킹됩니다.
- 잠금이 해제(또는 만료)될 때 대기 중인 요청 중 하나가 깨어나 획득합니다.
- `ctx`가 취소되면 대기 중인 `Acquire`는 `context.Canceled` 오류를 반환합니다.

### 데드락 감지
```go
// DetectDeadlock은 소유자 간 대기 사이클을 탐지합니다.
// 예: A가 key1을 보유하고 key2를 대기, B가 key2를 보유하고 key1을 대기
// 사이클이 있으면 관련된 소유자 목록을 반환합니다.
cycle, err := svc.DetectDeadlock()
// cycle: []string{"ownerA", "ownerB"} 또는 nil
```

## 오류 타입

```go
var (
    ErrLockNotFound   = errors.New("잠금을 찾을 수 없습니다")
    ErrInvalidToken   = errors.New("유효하지 않은 펜싱 토큰입니다")
    ErrLockExpired    = errors.New("잠금이 만료되었습니다")
    ErrContextCanceled = errors.New("컨텍스트가 취소되었습니다")
)
```

## 채점 기준

| 항목 | 배점 |
|------|------|
| Acquire / Release 기본 동작 | 20점 |
| 펜싱 토큰 (단조 증가, 검증) | 20점 |
| TTL 만료 (자동 해제, 백그라운드 정리) | 20점 |
| 대기 큐 (블로킹 Acquire, ctx 취소) | 25점 |
| 데드락 감지 | 15점 |
| **합계** | **100점** |

## 실행 방법

```bash
cd assignments/a5-distributed-lock
go test ./... -v
go test -race ./...           # 레이스 컨디션 감지
go test ./... -v -run TestGrade
```

## 힌트

- 각 키에 대해 `sync.Mutex` + `sync.Cond`를 사용하면 대기/깨우기를 깔끔하게 구현할 수 있습니다.
- 또는 키별 `chan struct{}`를 만들고 해제 시 close하여 대기자를 깨울 수 있습니다.
- 펜싱 토큰은 `sync/atomic`의 `atomic.Uint64`로 전역 카운터를 관리하세요.
- 데드락 감지: "소유자 -> 대기 중인 키 -> 그 키의 현재 소유자" 그래프에서 DFS/BFS로 사이클을 찾으세요.
- `context.WithDeadline`이나 `select { case <-ch: case <-ctx.Done(): }`으로 취소 가능한 대기를 구현하세요.
