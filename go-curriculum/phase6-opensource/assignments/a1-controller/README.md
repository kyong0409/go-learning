# 과제 A1: 리소스 컨트롤러 구현

**난이도**: ★★★★½
**예상 소요 시간**: 5~7시간
**참고 패턴**: Kubernetes Controller + client-go WorkQueue

## 배경

Kubernetes의 모든 컨트롤러(Deployment, ReplicaSet, StatefulSet 등)는
동일한 패턴으로 동작합니다: **Watch → Enqueue → Reconcile**.

이 과제에서는 그 핵심 구조를 직접 구현합니다.

## 요구사항

### 핵심 개념

- **Resource**: `Spec`(원하는 상태)과 `Status`(현재 상태)를 가진 오브젝트
- **Event**: 리소스에 발생한 변경 (CREATE/UPDATE/DELETE)
- **WorkQueue**: 중복 제거 + 실패 시 재시도(지수 백오프)를 지원하는 큐
- **Controller**: 이벤트를 받아 큐에 넣고, 워커가 꺼내 Reconcile하는 루프

### 구현할 인터페이스

```go
// Reconciler는 리소스의 원하는 상태로 실제 상태를 맞추는 인터페이스입니다.
type Reconciler interface {
    Reconcile(ctx context.Context, key string) error
}

// WorkQueue는 중복 제거와 재시도를 지원하는 작업 큐입니다.
type WorkQueue interface {
    Add(key string)
    Get() (string, bool)   // (key, shutdown)
    Done(key string)
    Forget(key string)     // 재시도 횟수 초기화
    AddRateLimited(key string)  // 백오프 후 재추가
    Len() int
    ShutDown()
}
```

### Controller 동작

1. `Watch(eventCh <-chan Event)`: 이벤트 채널을 감시해 큐에 키 추가
2. `Run(ctx, workers int)`: workers개의 goroutine으로 큐에서 꺼내 Reconcile
3. Reconcile 실패 시: 지수 백오프로 재시도 (최대 5회)
4. 중복 이벤트: 같은 키가 큐에 이미 있으면 한 번만 처리
5. `ctx` 취소 시: 모든 goroutine 정리 후 반환

### WorkQueue 세부 동작

- `Add(key)`: 처리 중인 키는 다음 번에 재추가 (중복 방지)
- `Get()`: 블로킹, 처리 가능한 키 반환
- `Done(key)`: 처리 완료 표시
- `AddRateLimited(key)`: 백오프 딜레이 후 Add (baseDelay * 2^retryCount)
- `Forget(key)`: 성공 시 재시도 카운터 초기화

### Resource 구조

```go
type ResourceSpec struct {
    Replicas int
    Image    string
}

type ResourceStatus struct {
    ReadyReplicas int
    Phase         string  // "Pending", "Running", "Failed"
}

type Resource struct {
    Name   string
    Spec   ResourceSpec
    Status ResourceStatus
}
```

## 채점 기준 (100점)

| 항목 | 점수 |
|------|------|
| WorkQueue 기본 동작 (Add/Get/Done) | 20점 |
| 중복 키 제거 | 15점 |
| Controller Watch → Enqueue | 15점 |
| Reconcile 호출 | 15점 |
| 실패 시 지수 백오프 재시도 | 20점 |
| 컨텍스트 취소 시 정상 종료 | 15점 |

## 실행 방법

```bash
cd a1-controller
go mod tidy
go test ./... -v
go test -v -run TestGrade
```

## 참고 자료

- `k8s.io/client-go/util/workqueue/queue.go`
- `k8s.io/client-go/tools/cache/controller.go`
- `../01-k8s-patterns/README.md`
