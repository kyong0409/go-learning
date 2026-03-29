# 01. Kubernetes 핵심 Go 패턴

## 개요

Kubernetes와 client-go는 Go 생태계에서 가장 정교한 패턴들을 담고 있습니다.
이 모듈에서는 모든 K8s 컨트롤러와 오퍼레이터의 뼈대가 되는 패턴들을 학습합니다.

---

## 1. Controller 패턴 (Watch-Reconcile Loop)

### 개념

컨트롤러는 **현재 상태(actual state)**를 **원하는 상태(desired state)**로 맞추는 루프입니다.

```
Watch(API Server)
      ↓
  Event 발생 (Create/Update/Delete)
      ↓
  WorkQueue에 키 추가
      ↓
  Worker goroutine이 키를 꺼냄
      ↓
  Reconcile(key) 호출
      ↓
  원하는 상태 == 현재 상태? → 완료
                            → 아니면 조정 후 재시도
```

### 실제 코드 위치

```
k8s.io/client-go/tools/cache/controller.go
  - type Controller interface
  - type controller struct
  - func (c *controller) Run(workers int, stopCh <-chan struct{})
  - func (c *controller) processLoop()

k8s.io/controller-runtime/pkg/internal/controller/controller.go
  - func (c *Controller) reconcileHandler(ctx context.Context, obj interface{})
  - func (c *Controller) processNextWorkItem(ctx context.Context) bool
```

### 핵심 코드 패턴

```go
// 컨트롤러의 핵심 루프 - 실제 K8s 코드와 동일한 구조
func (c *Controller) worker(ctx context.Context) {
    for c.processNextItem(ctx) {
    }
}

func (c *Controller) processNextItem(ctx context.Context) bool {
    key, quit := c.queue.Get()
    if quit {
        return false
    }
    defer c.queue.Done(key)

    err := c.reconcile(ctx, key.(string))
    if err != nil {
        // 에러 발생 시 재큐잉 (지수 백오프)
        c.queue.AddRateLimited(key)
        return true
    }
    c.queue.Forget(key)
    return true
}
```

### 왜 이 패턴인가?

- **멱등성(Idempotency)**: reconcile은 몇 번 호출해도 같은 결과여야 함
- **레벨 트리거(Level-triggered)**: 이벤트가 아닌 현재 상태 기반으로 판단
- **에러 복구**: 실패 시 자동 재시도 (지수 백오프)

---

## 2. Informer/Lister/WorkQueue 패턴

### 개념

API 서버를 직접 폴링하면 부하가 크므로, client-go는 로컬 캐시를 두는 Informer 패턴을 사용합니다.

```
API Server
    ↓ List+Watch (한 번만)
  Reflector
    ↓ Delta 이벤트
  DeltaFIFO Queue
    ↓
  Store (in-memory cache)
    ↓
  EventHandler 콜백
    ↓
  WorkQueue (deduplicated, rate-limited)
    ↓
  Controller Worker
    ↓
  Lister (캐시에서 읽기) → Reconcile
```

### 실제 코드 위치

```
k8s.io/client-go/tools/cache/reflector.go
  - type Reflector struct
  - func (r *Reflector) Run(stopCh <-chan struct{})
  - func (r *Reflector) ListAndWatch(stopCh <-chan struct{}) error

k8s.io/client-go/tools/cache/store.go
  - type Store interface
  - type ThreadSafeStore interface
  - type cache struct (Store 구현체)

k8s.io/client-go/tools/cache/shared_informer.go
  - type SharedInformer interface
  - type sharedIndexInformer struct
  - func (s *sharedIndexInformer) Run(stopCh <-chan struct{})

k8s.io/client-go/util/workqueue/queue.go
  - type Interface interface
  - type Type struct
  - func (q *Type) Add(item interface{})
  - func (q *Type) Get() (item interface{}, shutdown bool)
```

### Store 인터페이스

```go
// 실제 client-go Store 인터페이스
type Store interface {
    Add(obj interface{}) error
    Update(obj interface{}) error
    Delete(obj interface{}) error
    List() []interface{}
    ListKeys() []string
    Get(obj interface{}) (item interface{}, exists bool, err error)
    GetByKey(key string) (item interface{}, exists bool, err error)
    Replace([]interface{}, string) error
    Resync() error
}
```

---

## 3. Builder 패턴

K8s 오브젝트 생성에 널리 쓰이는 빌더 패턴입니다.

### 실제 코드 위치

```
k8s.io/client-go/tools/cache/listwatch.go
  - type ListWatch struct
  - func NewListWatchFromClient(...) *ListWatch

sigs.k8s.io/controller-runtime/pkg/builder/controller.go
  - type Builder struct
  - func (blder *Builder) For(object client.Object, ...) *Builder
  - func (blder *Builder) Owns(object client.Object, ...) *Builder
  - func (blder *Builder) Complete(r reconcile.Reconciler) error
```

### 패턴 예시

```go
// controller-runtime의 빌더 패턴
err := ctrl.NewControllerManagedBy(mgr).
    For(&appsv1.Deployment{}).
    Owns(&corev1.Pod{}).
    Complete(r)
```

---

## 4. Options 패턴 (Functional Options)

gRPC-go, Docker, K8s 전반에서 사용하는 설정 패턴입니다.

### 실제 코드 위치

```
google.golang.org/grpc/server.go
  - type ServerOption interface
  - func WithMaxRecvMsgSize(s int) ServerOption

k8s.io/client-go/tools/cache/shared_informer.go
  - type SharedInformerOption func(*sharedIndexInformer) *sharedIndexInformer
  - func WithResyncPeriod(resyncPeriod time.Duration) SharedInformerOption
```

### 패턴 예시

```go
type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.timeout = d
    }
}

func NewController(opts ...Option) *Controller {
    cfg := &Config{timeout: 30 * time.Second} // 기본값
    for _, opt := range opts {
        opt(cfg)
    }
    return &Controller{cfg: cfg}
}
```

---

## 5. Interface 기반 플러그인 아키텍처

### 실제 코드 위치

```
containerd/containerd/snapshots/snapshots.go
  - type Snapshotter interface

containerd/containerd/runtime/v2/runtime.go
  - type PlatformRuntime interface

k8s.io/kubernetes/pkg/scheduler/framework/interface.go
  - type Plugin interface
  - type FilterPlugin interface
  - type ScorePlugin interface
```

### 패턴 예시

```go
// K8s 스케줄러 플러그인 인터페이스 (실제 코드)
type FilterPlugin interface {
    Plugin
    Filter(ctx context.Context, state *CycleState, p *v1.Pod, nodeInfo *NodeInfo) *Status
}

// 플러그인 레지스트리
type Registry map[string]PluginFactory
type PluginFactory func(obj runtime.Object, h Handle) (Plugin, error)
```

---

## 학습 포인트 요약

| 패턴 | 핵심 아이디어 | 적용 과제 |
|------|---------------|-----------|
| Controller Loop | Watch → Queue → Reconcile | A1, A6 |
| Informer | 로컬 캐시 + 이벤트 핸들러 | A2 |
| WorkQueue | 중복 제거 + 레이트 리밋 재시도 | A1 |
| Store/Lister | 쓰레드 안전 캐시 | A2 |
| Functional Options | 유연한 설정 | A1, A6 |
| Interface Plugin | 교체 가능한 구현 | A2, A6 |

## 다음 단계

이 패턴들을 이해했다면:
1. [A1 - 리소스 컨트롤러](../assignments/a1-controller/) 과제 시작
2. [A2 - 인포머/리스터](../assignments/a2-informer/) 과제로 심화
