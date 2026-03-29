# 과제 A5: TTL과 LRU 퇴거를 가진 동시성 안전 캐시

**난이도**: ★★★★☆ (4.5/5)
**예상 소요 시간**: 5~7시간

## 과제 설명

TTL 만료, LRU(Least Recently Used) 퇴거, 최대 크기 제한, 퇴거 콜백을 지원하는
제네릭 인메모리 캐시를 구현합니다. 모든 연산은 고루틴 안전해야 합니다.

## 구현할 타입

### `Cache[K comparable, V any]`

```go
// 기본 캐시 생성
c := New[string, int](Options{
    MaxSize:    100,                // 최대 항목 수 (0 = 무제한)
    DefaultTTL: 5 * time.Minute,   // 기본 TTL (0 = 만료 없음)
    OnEviction: func(key string, value int, reason EvictionReason) {
        fmt.Printf("퇴거: %s (%v)\n", key, reason)
    },
})

// 기본 TTL 사용
c.Set("foo", 42)

// 개별 TTL 지정
c.SetWithTTL("bar", 100, 10*time.Second)

// 조회 (히트 시 LRU 순서 갱신)
val, ok := c.Get("foo")     // 42, true
val, ok  = c.Get("missing") // 0, false

// 삭제
c.Delete("foo")

// 현재 항목 수 (만료된 항목 제외)
c.Len()

// 통계
stats := c.Stats()
// stats.Hits, stats.Misses, stats.Evictions, stats.Expirations

// 전체 삭제
c.Flush()

// 백그라운드 정리 중지 (defer 권장)
c.Close()
```

### `EvictionReason` — 퇴거 사유

```go
const (
    EvictionReasonExpired  EvictionReason = "expired"   // TTL 만료
    EvictionReasonCapacity EvictionReason = "capacity"  // 최대 크기 초과
    EvictionReasonDeleted  EvictionReason = "deleted"   // 명시적 삭제
    EvictionReasonFlushed  EvictionReason = "flushed"   // Flush 호출
)
```

### `Stats` — 캐시 통계

```go
type Stats struct {
    Hits        uint64  // Get 히트 횟수
    Misses      uint64  // Get 미스 횟수
    Evictions   uint64  // LRU 퇴거 횟수
    Expirations uint64  // TTL 만료 횟수
}
```

## 구현 요구사항

### LRU 퇴거 정책
- `MaxSize` 초과 시 가장 오래 전에 사용된 항목을 퇴거합니다.
- `Get` 호출 시 해당 항목이 최근 사용으로 표시됩니다.
- `Set`/`SetWithTTL` 호출 시에도 LRU 순서가 갱신됩니다.

### TTL 만료
- **지연 만료 (Lazy expiration)**: `Get` 호출 시 만료 여부를 확인합니다. 만료된 항목은 `false`를 반환하고 삭제합니다.
- **백그라운드 정리**: 주기적으로(기본 1분마다) 만료된 항목을 일괄 삭제합니다.
- `DefaultTTL`이 0이면 해당 항목은 만료되지 않습니다.

### 고루틴 안전성
- `sync.RWMutex`를 사용하여 읽기는 공유, 쓰기는 단독 잠금을 적용하세요.

### 퇴거 콜백
- `OnEviction`이 설정된 경우 퇴거/만료/삭제 시 호출됩니다.
- 콜백은 잠금을 **보유하지 않은** 상태에서 호출하세요 (데드락 방지).

## 채점 기준

| 항목 | 배점 |
|------|------|
| Get/Set/Delete 기본 동작 | 20점 |
| TTL 만료 (지연 + 백그라운드) | 25점 |
| LRU 퇴거 순서 | 25점 |
| 퇴거 콜백 및 통계 | 15점 |
| 동시성 안전성 (race detector 통과) | 15점 |
| **합계** | **100점** |

## 실행 방법

```bash
cd assignments/a5-cache
go test ./... -v
go test -race ./...          # 레이스 컨디션 감지
go test ./... -v -run TestGrade
```

## 힌트

- LRU 순서 관리에는 `container/list`의 이중 연결 리스트를 활용하세요.
- 각 항목을 `list.Element`로 저장하고, 맵에서 `list.Element` 포인터를 값으로 가지면 O(1) 이동이 가능합니다.
- 백그라운드 정리는 `time.NewTicker`와 `select`로 구현하고, `Close()`에서 채널로 종료 신호를 보내세요.
- TTL이 0인 경우를 구별하기 위해 만료 시각을 `time.Time` 제로값으로 표현하세요.
