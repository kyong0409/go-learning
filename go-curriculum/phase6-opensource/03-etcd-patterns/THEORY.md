# 03. etcd 핵심 Go 패턴 - 이론 심화

## etcd란 무엇인가

etcd는 **분산 신뢰성 있는 키-값 스토어**입니다. 단순한 KV 스토어가 아니라 다음 특성을 보장합니다:

- **일관성(Consistency)**: 모든 노드가 동일한 데이터를 봄 (Raft 합의)
- **고가용성(High Availability)**: 과반수 노드 장애에도 서비스 유지
- **감시 가능성(Watchability)**: 키 변경을 실시간 스트림으로 구독
- **트랜잭션**: 원자적 비교-교환(CAS) 지원

Kubernetes는 모든 클러스터 상태(Pod, Service, Deployment 등)를 etcd에 저장합니다.

```
etcd 클러스터 (일반적으로 3 또는 5 노드):

┌──────────┐    Raft    ┌──────────┐    Raft    ┌──────────┐
│  etcd-1  │◄──────────►│  etcd-2  │◄──────────►│  etcd-3  │
│ (Leader) │            │(Follower)│            │(Follower)│
└──────────┘            └──────────┘            └──────────┘
     ▲
     │ 클라이언트 요청
     │ (모든 쓰기는 Leader 경유)
┌────┴────────────────────────────────────────────┐
│              Kubernetes API Server               │
└─────────────────────────────────────────────────┘
```

---

## Raft 합의 알고리즘 개요

Raft는 분산 시스템에서 여러 노드가 동일한 데이터에 합의하도록 하는 알고리즘입니다. etcd의 핵심입니다.

### 노드 역할

```
세 가지 상태:
  Follower  → 기본 상태. Leader의 지시를 따름
  Candidate → 선거 진행 중. 투표 요청
  Leader    → 클라이언트 요청 처리. 로그 복제 지시

상태 전환:
  Follower
      ↓ (election timeout 경과, Leader로부터 heartbeat 없음)
  Candidate
      ↓ (과반수 득표)
  Leader
      ↓ (더 높은 term의 메시지 수신)
  Follower
```

### Leader Election (리더 선출)

```
1. 모든 노드가 Follower로 시작
2. 랜덤 타임아웃(150~300ms) 후 Follower → Candidate
3. Candidate: term 증가, 자신에게 투표, 다른 노드에 RequestVote RPC
4. 과반수(3/5, 2/3) 득표 → Leader
5. Leader: 주기적 heartbeat(AppendEntries) 전송
6. Follower가 heartbeat 수신 → 선거 타임아웃 리셋

랜덤 타임아웃의 의미:
  동시에 여러 Candidate가 나오는 split vote를 줄임
  (같은 타임아웃이면 항상 동시 선거 → 무한 반복)
```

### Log Replication (로그 복제)

```
클라이언트 쓰기 요청:
  1. 클라이언트 → Leader에 Put("key", "value") 요청
  2. Leader: 로그에 항목 추가 (아직 커밋 전)
  3. Leader → 모든 Follower에 AppendEntries RPC (로그 복제)
  4. Follower들: 로그 저장 후 응답
  5. Leader: 과반수 응답 수신 → 커밋 (상태 머신 적용)
  6. Leader → 클라이언트에 응답
  7. 다음 heartbeat에서 Follower들도 커밋

커밋된 항목의 보장:
  - 절대 변경되지 않음 (Safety)
  - 결국 모든 노드에 적용됨 (Liveness)
```

### 실제 코드 위치

```
go.etcd.io/raft/raft.go
  - type raft struct       ← Raft 상태 머신
  - func (r *raft) Step(m pb.Message) error  ← 메시지 처리
  - func stepLeader(r *raft, m pb.Message) error
  - func stepFollower(r *raft, m pb.Message) error
  - func stepCandidate(r *raft, m pb.Message) error

go.etcd.io/etcd/server/etcdserver/raft.go
  - type raftNode struct   ← etcd와 raft 라이브러리 연결
  - func (r *raftNode) run()  ← Raft 이벤트 처리 루프
```

---

## MVCC (Multi-Version Concurrency Control)

### 리비전(Revision)의 개념

etcd의 모든 쓰기 연산은 **전역 단조 증가 리비전(revision)**을 생성합니다.

```
초기 상태: currentRevision = 0

트랜잭션 1: Put("pods/pod-1", "Running")
  → revision = 1
  → pod-1.CreateRevision = 1
  → pod-1.ModRevision = 1

트랜잭션 2: Put("pods/pod-2", "Pending")
  → revision = 2
  → pod-2.CreateRevision = 2
  → pod-2.ModRevision = 2

트랜잭션 3: Put("pods/pod-1", "Succeeded")  (pod-1 상태 변경)
  → revision = 3
  → pod-1.ModRevision = 3
  → (이전 버전 보존: revision=1의 "Running"도 여전히 조회 가능)

트랜잭션 4: Delete("pods/pod-2")
  → revision = 4
  → pod-2에 tombstone 기록 (실제 데이터는 compact 전까지 보존)
```

### KV 구조

```go
// etcd가 저장하는 KeyValue 구조
type KeyValue struct {
    Key            []byte  // 키
    CreateRevision int64   // 이 키가 처음 생성된 리비전
    ModRevision    int64   // 이 키가 마지막으로 수정된 리비전
    Version        int64   // 이 키의 버전 (수정 횟수, 삭제 후 재생성 시 1로 초기화)
    Value          []byte  // 값
    Lease          int64   // 연결된 Lease ID (없으면 0)
}
```

### 과거 리비전 조회

```go
// 현재 값 조회
resp, _ := cli.Get(ctx, "pods/pod-1")

// 특정 리비전 시점의 값 조회 (히스토리)
resp, _ := cli.Get(ctx, "pods/pod-1",
    clientv3.WithRev(1),  // revision=1 시점의 값 ("Running")
)

// 리비전 기반 조회의 활용:
// Kubernetes는 이를 통해 Watch 재연 구현
// "내가 마지막으로 처리한 revision=100 이후의 변경을 다시 줘"
```

### 내부 저장 구조

```
etcd MVCC 내부 (BoltDB/bbolt 기반):

Key-Version Index (in-memory):
  "pods/pod-1" → [rev=1, rev=3]     (생성 및 수정 리비전 목록)
  "pods/pod-2" → [rev=2, rev=4(삭제)]

Value Store (디스크, BoltDB):
  (rev=1, sub=0) → "Running"
  (rev=2, sub=0) → "Pending"
  (rev=3, sub=0) → "Succeeded"
  (rev=4, sub=0) → tombstone

조회 과정:
  Get("pods/pod-1") → 인덱스에서 최신 rev(3) 찾기 → BoltDB에서 값 읽기
  Get("pods/pod-1", rev=1) → 인덱스에서 rev=1 찾기 → BoltDB에서 값 읽기
```

### Compact: 오래된 리비전 정리

```go
// Compact 없이는 디스크가 무한정 증가
// "revision 100 이전의 모든 히스토리 삭제"
resp, err := cli.Compact(ctx, 100)
// revision 1~99의 데이터 삭제
// revision 100 이후는 유지

// Compact 후의 영향:
watchChan := cli.Watch(ctx, "/pods/",
    clientv3.WithRev(50),  // 이미 compact된 리비전
)
// → WatchResponse.Canceled = true, Err = ErrCompacted

// Kubernetes etcd 자동 compact 설정:
// kube-apiserver --etcd-compaction-interval=5m
// (5분마다 자동 compact)
```

---

## Watch 메커니즘 심층 분석

### 서버사이드 이벤트 스트리밍

```
클라이언트 Watch 요청 흐름:

Client                          etcd Server
  │                                  │
  │── Watch("/pods/", rev=100) ──→  │
  │                                  │ WatchStream 생성
  │                                  │ WatchGroup에 등록
  │                              (대기)
  │                                  │
  │  (다른 클라이언트가 Put 수행)      │
  │                                  │ revision=101: Put("/pods/pod-1")
  │                                  │ watcherGroup에서 매칭 watch 찾기
  │                                  │ 매칭된 모든 watch에 이벤트 전송
  │←── WatchResponse{Events: [...]} ─│
  │←── WatchResponse{Events: [...]} ─│ (계속 스트리밍)
```

### 접두사 Watch

```go
// 단일 키 Watch
cli.Watch(ctx, "/pods/pod-1")

// 접두사 Watch (하위 모든 키)
cli.Watch(ctx, "/pods/", clientv3.WithPrefix())
// "/pods/pod-1", "/pods/pod-2", "/pods/ns1/pod-3" 등 모두 감시

// 범위 Watch (from 이상 to 미만)
cli.Watch(ctx, "/pods/a",
    clientv3.WithRange("/pods/z"),
)
// "/pods/a" ~ "/pods/z" 범위의 키 감시

// 리비전 지정 Watch (재연 가능)
cli.Watch(ctx, "/pods/",
    clientv3.WithPrefix(),
    clientv3.WithRev(100),  // revision 100부터 이벤트 재생
)
```

### WatchResponse 처리

```go
// 완전한 Watch 처리 패턴
func watchPods(ctx context.Context, cli *clientv3.Client) {
    watchChan := cli.Watch(ctx, "/pods/",
        clientv3.WithPrefix(),
        clientv3.WithPrevKV(),  // 이전 값도 포함
    )

    for {
        select {
        case resp, ok := <-watchChan:
            if !ok {
                // 채널이 닫힘 (context 취소 등)
                return
            }
            if resp.Canceled {
                // Watch 강제 취소 (ErrCompacted 등)
                if resp.Err() == rpctypes.ErrCompacted {
                    // compact된 리비전 요청: 재시작 필요
                    log.Println("watch compacted, restart from latest")
                }
                return
            }
            for _, ev := range resp.Events {
                switch ev.Type {
                case mvccpb.PUT:
                    if ev.IsCreate() {
                        // 새 키 생성
                        fmt.Printf("CREATE %s = %s\n", ev.Kv.Key, ev.Kv.Value)
                    } else {
                        // 기존 키 수정
                        fmt.Printf("UPDATE %s: %s → %s\n",
                            ev.Kv.Key, ev.PrevKv.Value, ev.Kv.Value)
                    }
                case mvccpb.DELETE:
                    fmt.Printf("DELETE %s (was: %s)\n",
                        ev.Kv.Key, ev.PrevKv.Value)
                }
            }

        case <-ctx.Done():
            return
        }
    }
}
```

### 서버 내부: watchableStore

```go
// go.etcd.io/etcd/server/mvcc/watchable_store.go (단순화)

type watchableStore struct {
    *store                              // MVCC 스토어
    mu       sync.RWMutex
    victims  []watcherBatch            // 느린 watcher (버퍼 오버플로)
    watchers map[WatchID]*watcher      // 활성 watcher 맵
}

// 쓰기 발생 시 watcher 알림
func (s *watchableStore) notify(rev int64, evs []mvccpb.Event) {
    for _, ev := range evs {
        // 이 이벤트를 구독하는 모든 watcher 찾기
        for _, w := range s.watchers {
            if w.contains(ev.Kv.Key) && w.minRev <= rev {
                select {
                case w.ch <- WatchResponse{Events: []mvccpb.Event{ev}}:
                    // 전송 성공
                default:
                    // 채널 풀: victim으로 이동 (느린 consumer)
                    s.victims = append(s.victims, w)
                }
            }
        }
    }
}
```

---

## Lease 시스템

### TTL 기반 임대의 동작 원리

```
Lease 생명주기:

  Grant(TTL=10s) → LeaseID=12345 생성
       │
       │ 10초 카운트다운 시작
       │
  Put("key", "value", Lease=12345) → 키에 Lease 연결
       │
  KeepAlive(12345) → TTL 갱신 (카운트다운 리셋)
  KeepAlive(12345) → TTL 갱신
  KeepAlive(12345) → TTL 갱신
       │
  (KeepAlive 중단 - 프로세스 죽음)
       │
  TTL 카운트다운 0에 도달
       │
  Lease 만료 → 연결된 모든 키 자동 삭제
       │
  Watch를 통해 삭제 이벤트 전파
```

### 서비스 디스커버리 구현

```go
// 서비스 등록: 서비스가 살아있는 동안만 키 유지
func registerService(ctx context.Context, cli *clientv3.Client,
    serviceName, addr string) error {

    // 10초 TTL Lease 생성
    lease, err := cli.Grant(ctx, 10)
    if err != nil {
        return fmt.Errorf("grant lease: %w", err)
    }

    // 서비스 주소 등록 (Lease와 연결)
    key := fmt.Sprintf("/services/%s/%s", serviceName, addr)
    _, err = cli.Put(ctx, key, addr, clientv3.WithLease(lease.ID))
    if err != nil {
        return fmt.Errorf("register: %w", err)
    }

    // 백그라운드에서 Lease 갱신
    keepAliveCh, err := cli.KeepAlive(ctx, lease.ID)
    if err != nil {
        return fmt.Errorf("keepalive: %w", err)
    }

    go func() {
        for {
            select {
            case _, ok := <-keepAliveCh:
                if !ok {
                    // 갱신 중단 (context 취소 또는 에러)
                    return
                }
                // 갱신 성공 (로그 생략 가능)
            case <-ctx.Done():
                return
            }
        }
    }()

    return nil
}

// 서비스 조회: 현재 살아있는 서비스 목록
func discoverServices(ctx context.Context, cli *clientv3.Client,
    serviceName string) ([]string, error) {

    resp, err := cli.Get(ctx,
        fmt.Sprintf("/services/%s/", serviceName),
        clientv3.WithPrefix(),
    )
    if err != nil {
        return nil, err
    }

    var addrs []string
    for _, kv := range resp.Kvs {
        addrs = append(addrs, string(kv.Value))
    }
    return addrs, nil
}
```

---

## 분산 잠금과 Leader Election

### etcd 기반 분산 잠금

```go
// go.etcd.io/etcd/client/v3/concurrency 패키지 사용
import "go.etcd.io/etcd/client/v3/concurrency"

func distributedLock(cli *clientv3.Client) error {
    // 세션 생성 (내부적으로 Lease 사용)
    session, err := concurrency.NewSession(cli, concurrency.WithTTL(15))
    if err != nil {
        return err
    }
    defer session.Close()

    // 뮤텍스 생성 (키 접두사 기반)
    mutex := concurrency.NewMutex(session, "/my-lock/")

    // 잠금 획득 (블로킹, context로 타임아웃 가능)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := mutex.Lock(ctx); err != nil {
        return fmt.Errorf("lock: %w", err)
    }
    defer mutex.Unlock(context.Background())

    // 임계 구역: 분산 잠금 보호 하에 실행
    doExclusiveWork()

    return nil
}
```

### 펜싱 토큰(Fencing Token)

분산 잠금에서 잠금 보유자가 느리거나 일시 정지된 경우 오래된 잠금 보유자가 데이터를 덮어쓰는 문제가 있습니다. etcd는 리비전을 펜싱 토큰으로 사용하여 이를 해결합니다.

```
펜싱 토큰 없는 문제:

  Process A: 잠금 획득 (Lease 기반)
  Process A: GC pause 30초 (잠금 만료)
  Process B: 잠금 획득 (새 잠금)
  Process B: 공유 리소스 수정 중...
  Process A: GC 재개, 자신이 잠금 보유자라 생각하고 리소스 덮어씀!

펜싱 토큰으로 해결:

  Process A: 잠금 획득, revision=100
  Process B: 잠금 획득, revision=101

  공유 리소스(DB 등): 마지막으로 처리한 revision 추적
  Process A가 revision=100으로 쓰기 시도 → 101 이미 처리, 거부
  Process B가 revision=101로 쓰기 시도 → 허용
```

```go
// etcd 트랜잭션으로 펜싱 구현
// "내 Lease가 아직 살아있을 때만 쓰기 허용"
txn := cli.Txn(ctx)
resp, err := txn.
    If(clientv3.Compare(clientv3.LeaseValue(myKey), "=", string(leaseID))).
    Then(clientv3.OpPut(targetKey, newValue)).
    Else(clientv3.OpGet(targetKey)).
    Commit()

if !resp.Succeeded {
    return errors.New("lease expired, write rejected")
}
```

### errors.AsType() — Go 1.26 (에러 처리 개선)

Go 1.26에서 `errors.AsType[T]()` 제네릭 함수가 추가되어 etcd 에러 타입 검사가 더 간결해졌습니다.

```go
import (
    "errors"
    "go.etcd.io/etcd/client/v3/rpctypes"
)

// Go 1.25 이하 — 타입 단언 방식
func handleEtcdError(err error) {
    var etcdErr rpctypes.EtcdError
    if errors.As(err, &etcdErr) {
        switch etcdErr.Code() {
        case codes.NotFound:
            log.Println("키를 찾을 수 없음")
        case codes.PermissionDenied:
            log.Println("권한 없음")
        }
    }
}

// Go 1.26+ — errors.AsType[T]() 사용 (더 간결)
func handleEtcdError(err error) {
    if etcdErr, ok := errors.AsType[rpctypes.EtcdError](err); ok {
        switch etcdErr.Code() {
        case codes.NotFound:
            log.Println("키를 찾을 수 없음")
        case codes.PermissionDenied:
            log.Println("권한 없음")
        }
    }
}
```

### Leader Election

```go
import "go.etcd.io/etcd/client/v3/concurrency"

func runWithLeadership(cli *clientv3.Client, componentName string) {
    session, _ := concurrency.NewSession(cli)
    defer session.Close()

    election := concurrency.NewElection(session, "/leader-election/"+componentName)

    for {
        // 리더 선출에 참여 (블로킹: 리더가 될 때까지 대기)
        if err := election.Campaign(context.Background(), "my-identity"); err != nil {
            log.Printf("campaign error: %v", err)
            time.Sleep(time.Second)
            continue
        }

        log.Println("I am the leader!")

        // 리더로서 작업 수행
        ctx, cancel := context.WithCancel(context.Background())
        go doLeaderWork(ctx)

        // 리더십 유지 감시
        observeCh := election.Observe(ctx)
        for resp := range observeCh {
            if string(resp.Kvs[0].Value) != "my-identity" {
                // 다른 노드가 리더가 됨
                cancel()
                break
            }
        }

        // 리더십 반납
        election.Resign(context.Background())
    }
}
```

---

## Kubernetes와 etcd의 관계

Kubernetes의 모든 상태는 etcd에 저장됩니다:

```
etcd 키 구조 (Kubernetes):
  /registry/pods/default/pod-1
  /registry/pods/kube-system/coredns-xxx
  /registry/deployments/default/my-app
  /registry/services/default/kubernetes
  /registry/nodes/node-1
  /registry/leases/kube-node-lease/node-1  ← 노드 heartbeat
  /registry/events/default/...

API 서버의 etcd 사용:
  - 모든 객체는 protobuf로 직렬화하여 etcd 저장
  - Watch: etcd Watch → API 서버 Watch → client-go Informer Watch
  - ResourceVersion = etcd revision (동일한 개념)
```

Watch 체인:
```
etcd Watch
    ↓ (etcd 이벤트)
kube-apiserver Watch 핸들러
    ↓ (HTTP Chunked 또는 WebSocket)
client-go Reflector (ListAndWatch)
    ↓ (DeltaFIFO에 추가)
Informer EventHandler
    ↓ (WorkQueue에 키 추가)
Controller Reconcile
```

이 체인이 Kubernetes의 실시간 상태 동기화를 가능하게 합니다.
