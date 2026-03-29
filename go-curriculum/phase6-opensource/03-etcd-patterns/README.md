# 03. etcd 핵심 Go 패턴

## 개요

etcd는 Kubernetes의 모든 상태를 저장하는 분산 키-값 저장소입니다.
Watch 메커니즘, MVCC, Lease 기반 TTL이 핵심 패턴입니다.

---

## 1. Watch 메커니즘 (서버사이드 이벤트 스트리밍)

### 개념

클라이언트는 키 또는 키 접두사를 Watch하면, 변경 발생 시 서버가 스트림으로 이벤트를 푸시합니다.
HTTP 폴링과 달리 서버가 능동적으로 통지합니다.

```
Client: Watch("/pods/", fromRevision=5)
        ↓
Server: 리비전 5 이후 이벤트를 스트림으로 전송
        ↓
Client: WatchChan (<-chan WatchResponse) 수신
```

### 실제 코드 위치

```
go.etcd.io/etcd/client/v3/watch.go
  - type Watcher interface
    - Watch(ctx, key, opts...) WatchChan
    - Close() error
  - type WatchChan <-chan WatchResponse
  - type WatchResponse struct
    - Events []*Event
    - Revision int64
    - Canceled bool

go.etcd.io/etcd/server/mvcc/watcher.go
  - type watcherGroup struct
  - func (wg *watcherGroup) add(wa watcher)
  - func (wg *watcherGroup) notify(rev int64, ev mvccpb.Event)
```

### Watch 사용 패턴

```go
// 실제 etcd 클라이언트 사용 패턴
cli, _ := clientv3.New(clientv3.Config{Endpoints: []string{"localhost:2379"}})
defer cli.Close()

// 접두사 Watch
watchChan := cli.Watch(ctx, "/pods/", clientv3.WithPrefix())

for resp := range watchChan {
    for _, ev := range resp.Events {
        switch ev.Type {
        case mvccpb.PUT:
            fmt.Printf("PUT %s = %s\n", ev.Kv.Key, ev.Kv.Value)
        case mvccpb.DELETE:
            fmt.Printf("DELETE %s\n", ev.Kv.Key)
        }
    }
}
```

---

## 2. MVCC (Multi-Version Concurrency Control)

### 개념

etcd는 모든 쓰기에 전역 리비전(revision)을 부여합니다.
각 키-값 쌍은 `ModRevision`(마지막 수정 리비전)과 `CreateRevision`을 가집니다.
과거 리비전을 조회할 수 있고, Watch 시작 리비전을 지정할 수 있습니다.

```
초기 상태: revision=0
Put("a", "1") → revision=1, a.ModRevision=1
Put("b", "2") → revision=2, b.ModRevision=2
Put("a", "3") → revision=3, a.ModRevision=3  (이전 값 보존)
Delete("b")   → revision=4, b 삭제 이벤트 기록
```

### 실제 코드 위치

```
go.etcd.io/etcd/server/mvcc/kvstore.go
  - type store struct
  - func (s *store) Put(key, value []byte, lease lease.LeaseID) (rev int64)
  - func (s *store) Range(key, end []byte, ro RangeOptions) (r *RangeResult, err error)

go.etcd.io/etcd/server/mvcc/revision.go
  - type revision struct
    - main int64   // 전역 단조 증가 리비전
    - sub  int64   // 같은 트랜잭션 내 서브 리비전
```

### 리비전 기반 Watch

```go
// 리비전 5 이후 이벤트만 수신 (재연 가능)
watchChan := cli.Watch(ctx, "/pods/",
    clientv3.WithPrefix(),
    clientv3.WithRev(5),
)
```

---

## 3. Lease 기반 TTL

### 개념

Lease는 TTL(Time-To-Live)을 가진 토큰입니다.
키를 Lease에 연결하면, Lease가 만료될 때 키도 자동 삭제됩니다.
keepalive로 TTL을 갱신해 살아있음을 알립니다.
→ 서비스 디스커버리, 리더 선출, 세션 관리에 활용됩니다.

### 실제 코드 위치

```
go.etcd.io/etcd/client/v3/lease.go
  - type Lease interface
    - Grant(ctx, ttl int64) (*LeaseGrantResponse, error)
    - Revoke(ctx, id LeaseID) (*LeaseRevokeResponse, error)
    - KeepAlive(ctx, id LeaseID) (<-chan *LeaseKeepAliveResponse, error)
    - KeepAliveOnce(ctx, id LeaseID) (*LeaseKeepAliveResponse, error)

go.etcd.io/etcd/server/lease/lessor.go
  - type Lessor interface
  - type lessor struct
  - func (le *lessor) Grant(id LeaseID, ttl int64) (*Lease, error)
```

### Lease 사용 패턴

```go
// 서비스 등록 패턴 (실제 사용 예)
lease, _ := cli.Grant(ctx, 10) // 10초 TTL

// 키에 Lease 연결
cli.Put(ctx, "/services/my-api", "127.0.0.1:8080",
    clientv3.WithLease(lease.ID))

// keepalive goroutine
keepAlive, _ := cli.KeepAlive(ctx, lease.ID)
go func() {
    for range keepAlive {
        // TTL 자동 갱신됨
    }
}()
```

---

## 4. Raft 합의 알고리즘 (개념)

etcd는 Raft로 분산 합의를 구현합니다. 직접 구현은 매우 복잡하지만 개념은 중요합니다.

### 핵심 개념

```
리더 선출:
  - 모든 노드는 Follower로 시작
  - 타임아웃 후 Candidate로 전환, 투표 요청
  - 과반수 득표 시 Leader가 됨

로그 복제:
  - 클라이언트 요청 → Leader만 처리
  - Leader → 모든 Follower에 로그 항목 전송 (AppendEntries RPC)
  - 과반수 응답 → Committed (안전하게 적용됨)
  - Committed 후 Leader가 상태 머신에 적용
```

### 실제 코드 위치

```
go.etcd.io/etcd/server/etcdserver/raft.go
  - type raftNode struct
  - func (r *raftNode) run()

go.etcd.io/raft/raft.go     (분리된 raft 라이브러리)
  - type raft struct
  - func (r *raft) Step(m pb.Message) error
  - func stepLeader(r *raft, m pb.Message) error
  - func stepFollower(r *raft, m pb.Message) error
```

---

## 5. Compact (압축)

오래된 리비전 데이터를 정리합니다. 무한정 쌓이면 디스크가 가득 찹니다.

```go
// 리비전 100 이전 데이터 모두 삭제
cli.Compact(ctx, 100)

// Compact 이후에는 리비전 100 이전으로 Watch 불가
// → Watch fromRevision=50 → ErrCompacted 에러
```

### 실제 코드 위치

```
go.etcd.io/etcd/server/mvcc/kvstore.go
  - func (s *store) Compact(trace *traceutil.Trace, rev int64) error
  - func (s *store) scheduleCompaction(compactMainRev int64, keep map[revision]struct{})
```

---

## 학습 포인트 요약

| 패턴 | 핵심 아이디어 | 적용 과제 |
|------|---------------|-----------|
| Watch 스트리밍 | 이벤트 채널, 접두사 Watch | A5 |
| MVCC 리비전 | 단조 증가, 특정 리비전부터 Watch | A5 |
| 접두사 Watch | 키 범위 구독 | A5 |
| Compact | 오래된 리비전 정리 | A5 |
| Lease/TTL | 만료 기반 키 관리 | A6 (심화) |

## 다음 단계

[A5 - 감시 시스템](../assignments/a5-watch-system/) 과제를 진행하세요.
