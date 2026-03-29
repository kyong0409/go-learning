# 과제 A5: 이벤트 감시 시스템

**난이도**: ★★★★½
**예상 소요 시간**: 5~7시간
**참고 패턴**: etcd Watch 메커니즘 + MVCC

## 배경

etcd는 키-값 저장소로, 모든 쓰기에 전역 리비전(revision)을 부여합니다.
클라이언트는 특정 키나 접두사를 Watch하여 변경 스트림을 실시간으로 받을 수 있습니다.
이 과제에서는 그 핵심 구조를 직접 구현합니다.

## 요구사항

### WatchableStore

```go
type WatchableStore interface {
    // 기본 KV 연산
    Put(key, value string) int64   // 반환값: 새 리비전
    Delete(key string) int64       // 반환값: 새 리비전 (키 없으면 -1)
    Get(key string) (value string, revision int64, ok bool)

    // Watch 연산
    Watch(ctx context.Context, key string, opts ...WatchOption) WatchChan

    // 유지보수
    Compact(revision int64) error  // revision 이전 데이터 삭제
    CurrentRevision() int64
}

type WatchChan <-chan WatchResponse

type WatchResponse struct {
    Events   []WatchEvent
    Revision int64  // 이 응답의 리비전
}

type WatchEvent struct {
    Type     EventType  // PUT, DELETE
    Key      string
    Value    string     // DELETE면 빈 문자열
    Revision int64      // 이 이벤트의 리비전
    PrevValue string    // 이전 값 (PUT 업데이트 시)
}
```

### WatchOption

```go
// WithPrefix()는 키 접두사로 Watch합니다.
// Watch("/pods/", WithPrefix())는 "/pods/"로 시작하는 모든 키를 감시합니다.

// WithRevision(rev)은 특정 리비전 이후 이벤트부터 수신합니다.
// 이미 발생한 이벤트를 재연(replay)할 수 있습니다.
```

### 핵심 동작

1. **리비전 추적**: Put/Delete마다 전역 리비전이 1씩 증가
2. **이벤트 히스토리**: 모든 이벤트를 리비전 순으로 보존
3. **다중 와처**: 같은 키에 여러 Watch를 등록할 수 있음
4. **접두사 Watch**: `/pods/`로 시작하는 모든 키 변경 수신
5. **리비전 재연**: `WithRevision(5)` → 리비전 5 이후 이벤트를 즉시 전송 후 새 이벤트 수신
6. **Compact**: 지정 리비전 이전 이벤트 히스토리 삭제 (Compact 후 그 리비전 이전으로 Watch 불가)
7. **ctx 취소**: Watch 채널이 닫힘

## 채점 기준 (100점)

| 항목 | 점수 |
|------|------|
| 기본 Put/Delete/Get + 리비전 | 15점 |
| Watch 단일 키 | 20점 |
| Watch 접두사 | 15점 |
| Watch fromRevision (이벤트 재연) | 20점 |
| 다중 와처 | 10점 |
| Compact | 10점 |
| 컨텍스트 취소 시 채널 닫힘 | 10점 |

## 실행 방법

```bash
cd a5-watch-system
go mod tidy
go test ./... -v
go test -v -run TestGrade
```

## 참고 자료

- `go.etcd.io/etcd/server/mvcc/watcher.go`
- `go.etcd.io/etcd/server/mvcc/kvstore.go`
- `go.etcd.io/etcd/client/v3/watch.go`
- `../03-etcd-patterns/README.md`
