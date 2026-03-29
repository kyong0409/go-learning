# 01. Kubernetes 핵심 Go 패턴 - 이론 심화

## Kubernetes 소스코드 구조

Kubernetes는 Go로 작성된 가장 큰 오픈소스 프로젝트 중 하나입니다. 전체 구조를 파악하면 어디서 무엇을 찾아야 하는지 알 수 있습니다.

```
kubernetes/
├── cmd/                          ← 바이너리 진입점
│   ├── kube-apiserver/           ← API 서버 (REST + Watch)
│   ├── kube-controller-manager/  ← 모든 내장 컨트롤러 실행
│   ├── kube-scheduler/           ← Pod 스케줄링
│   └── kubelet/                  ← 노드 에이전트
│
├── pkg/                          ← 외부 공개 패키지
│   ├── controller/               ← Deployment, ReplicaSet 등 컨트롤러
│   ├── scheduler/                ← 스케줄러 프레임워크
│   ├── kubelet/                  ← Kubelet 로직
│   └── api/                      ← API 타입 정의
│
├── staging/src/k8s.io/           ← 별도 모듈로 배포되는 패키지
│   ├── client-go/                ← Go 클라이언트 라이브러리
│   │   ├── tools/cache/          ← Informer, Store, Reflector
│   │   ├── util/workqueue/       ← WorkQueue
│   │   └── informers/            ← 타입별 Informer 자동 생성
│   ├── api/                      ← API 타입 (Pod, Deployment 등)
│   └── apimachinery/             ← API 기반 구조 (ObjectMeta 등)
│
├── vendor/                       ← 의존성 복사본
└── test/                         ← 통합 테스트 (e2e)
```

핵심 파일 위치:
```
k8s.io/client-go/tools/cache/controller.go      ← Controller 인터페이스
k8s.io/client-go/tools/cache/reflector.go       ← Reflector 구현
k8s.io/client-go/tools/cache/delta_fifo.go      ← DeltaFIFO 큐
k8s.io/client-go/tools/cache/shared_informer.go ← SharedInformer
k8s.io/client-go/tools/cache/store.go           ← Store 인터페이스
k8s.io/client-go/util/workqueue/queue.go        ← 기본 WorkQueue
k8s.io/client-go/util/workqueue/rate_limiting_queue.go ← 레이트 제한 큐
```

---

## Controller 패턴 상세

### 선언적 관리: "무엇을" vs "어떻게"

명령형(Imperative) 방식:
```
"nginx 컨테이너를 시작해라"
"포트 80을 열어라"
"3개 복제본을 만들어라"
```

선언적(Declarative) 방식:
```yaml
# 원하는 상태만 선언
apiVersion: apps/v1
kind: Deployment
spec:
  replicas: 3         # "3개 있어야 한다"
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
```

컨트롤러는 이 선언된 상태(Spec)와 실제 상태(Status)를 지속적으로 비교하고 맞춥니다.

### Reconciliation Loop 상세

```
                    ┌────────────────────────────┐
                    │    원하는 상태 (Spec)        │
                    │    replicas: 3              │
                    └──────────────┬─────────────┘
                                   │
                                   ▼
                    ┌────────────────────────────┐
                    │      Reconcile 함수         │
                    │   실제 상태 조회            │
                    │   현재 Pod 수: 1개          │
                    │   diff: 2개 부족            │
                    └──────────────┬─────────────┘
                                   │
                    ┌──────────────▼─────────────┐
                    │      조정 행동               │
                    │   Pod 2개 추가 생성          │
                    └──────────────┬─────────────┘
                                   │
                    ┌──────────────▼─────────────┐
                    │    실제 상태 (Status)        │
                    │    readyReplicas: 3         │
                    └────────────────────────────┘
```

### 멱등성(Idempotency)이 왜 중요한가

Reconcile은 언제든 재호출될 수 있습니다. 네트워크 오류, 프로세스 재시작, 단순 재시도 등으로 같은 키가 여러 번 처리됩니다. 따라서 **같은 입력에 항상 같은 결과**여야 합니다.

나쁜 예 (비멱등):
```go
func (c *Controller) reconcile(key string) error {
    // 비멱등: 호출할 때마다 Pod가 하나씩 추가됨
    return c.createPod(key)
}
```

좋은 예 (멱등):
```go
func (c *Controller) reconcile(key string) error {
    // 현재 상태 조회
    pods, err := c.lister.List(labels.SelectorFromSet(labels.Set{"app": key}))
    if err != nil {
        return err
    }

    desired := c.getDesiredReplicas(key)
    current := len(pods)

    // 차이만큼만 조정 (멱등)
    switch {
    case current < desired:
        return c.createPods(key, desired-current)
    case current > desired:
        return c.deletePods(pods[:current-desired])
    default:
        return nil // 이미 원하는 상태
    }
}
```

### 레벨 트리거 vs 엣지 트리거

Kubernetes는 **레벨 트리거(Level-triggered)** 방식을 사용합니다.

엣지 트리거 (이벤트 기반):
```
이벤트: "Pod 삭제됨" → "Pod 하나 생성"
문제: 이벤트를 놓치면 불일치 상태 유지됨
```

레벨 트리거 (상태 기반):
```
주기적으로: "현재 상태를 조회하고 원하는 상태와 비교"
이점: 이벤트를 놓쳐도 다음 Reconcile에서 자동 수정
```

실제 코드에서의 표현:
```go
// 이벤트 핸들러: 키만 큐에 넣음 (이벤트 내용은 버림)
func (c *Controller) onPodDeleted(obj interface{}) {
    key, err := cache.MetaNamespaceKeyFunc(obj)
    if err != nil {
        return
    }
    // 이벤트 타입(삭제)을 저장하지 않음 - 레벨 트리거 핵심
    c.queue.Add(key)
}

// Reconcile: 이벤트 타입 모름, 현재 상태만 봄
func (c *Controller) reconcile(key string) error {
    // "무슨 이벤트가 왔는가"가 아닌 "지금 상태가 어떤가"
    obj, exists, err := c.store.GetByKey(key)
    // ...
}
```

---

## Informer/Reflector/WorkQueue 아키텍처

### 전체 데이터 흐름

```
API Server
    │
    │ 1) List (초기 전체 목록) + Watch (이후 변경 스트림)
    │    gRPC/HTTP2 스트리밍
    ▼
┌─────────────────────────────────────────────────┐
│                   Reflector                      │
│  - ListAndWatch() 실행                           │
│  - 연결 끊기면 재연결 (exponential backoff)       │
│  - List 결과를 DeltaFIFO에 Replace로 전달        │
│  - Watch 이벤트를 DeltaFIFO에 Add/Update/Delete │
└─────────────────────────────────────────────────┘
    │
    │ 2) Delta 이벤트 (Added, Modified, Deleted, Sync)
    ▼
┌─────────────────────────────────────────────────┐
│                  DeltaFIFO                       │
│  - FIFO 큐 + 중복 제거                           │
│  - 같은 키의 여러 이벤트를 하나의 Delta 슬라이스로│
│  - Pop() 호출 시 처리기(handleDeltas)에 전달     │
└─────────────────────────────────────────────────┘
    │
    │ 3) handleDeltas() 호출
    ▼
┌────────────────────┐    ┌────────────────────────┐
│   Store (캐시)      │    │  EventHandler 콜백      │
│  - in-memory 맵    │    │  - OnAdd(obj)           │
│  - 쓰레드 안전     │    │  - OnUpdate(old, new)   │
│  - Lister가 여기서 │    │  - OnDelete(obj)        │
│    데이터 읽음     │    └──────────┬─────────────┘
└────────────────────┘               │
                                     │ 4) 키만 WorkQueue에 추가
                                     ▼
                        ┌────────────────────────┐
                        │      WorkQueue          │
                        │  - 중복 제거            │
                        │  - 처리 중 항목 추적    │
                        │  - 레이트 리밋 재시도   │
                        └──────────┬─────────────┘
                                   │
                                   │ 5) Get() → Reconcile
                                   ▼
                        ┌────────────────────────┐
                        │     Controller Worker   │
                        │  goroutine pool         │
                        │  - Reconcile(key) 호출  │
                        │  - Store에서 현재 상태  │
                        │    조회 (API 서버 미사용)│
                        └────────────────────────┘
```

### Reflector: API 서버와의 단일 연결점

```go
// k8s.io/client-go/tools/cache/reflector.go (단순화)
type Reflector struct {
    listerWatcher ListerWatcher  // List + Watch 수행
    store         Store          // 결과 저장소 (DeltaFIFO)
    expectedType  reflect.Type   // 감시할 타입 (Pod, Deployment 등)
    resyncPeriod  time.Duration  // 주기적 재동기화 간격
}

func (r *Reflector) ListAndWatch(stopCh <-chan struct{}) error {
    // 1단계: 전체 목록 가져오기
    list, err := r.listerWatcher.List(opts)
    if err != nil {
        return fmt.Errorf("failed to list: %w", err)
    }

    // Store 초기화 (Replace)
    r.store.Replace(listToItems(list), resourceVersion)

    // 2단계: 변경 감시 스트림 시작
    watcher, err := r.listerWatcher.Watch(opts)
    if err != nil {
        return err
    }

    // 3단계: 이벤트 처리 루프
    for {
        select {
        case event, ok := <-watcher.ResultChan():
            if !ok {
                return nil // 재연결 필요
            }
            switch event.Type {
            case watch.Added:
                r.store.Add(event.Object)
            case watch.Modified:
                r.store.Update(event.Object)
            case watch.Deleted:
                r.store.Delete(event.Object)
            }
        case <-stopCh:
            return nil
        }
    }
}
```

### DeltaFIFO: 이벤트 버퍼링과 순서 보장

DeltaFIFO는 일반 채널과 다릅니다. 같은 키에 대한 여러 이벤트를 하나의 Delta 슬라이스로 묶어 순서를 보장합니다.

```go
// Delta의 타입
type DeltaType string
const (
    Added    DeltaType = "Added"
    Updated  DeltaType = "Updated"
    Deleted  DeltaType = "Deleted"
    Replaced DeltaType = "Replaced" // 초기 List 결과
    Sync     DeltaType = "Sync"     // 주기적 재동기화
)

// 같은 키에 여러 이벤트가 오면 합쳐짐
// Pod "foo": Added → Updated → Updated
// → Deltas{ {Added, pod}, {Updated, pod}, {Updated, pod} }
// Pop() 시 한 번에 처리 (순서 보장)
```

### Store와 Lister의 분리

```go
// Store: 쓰기 인터페이스 (Reflector가 사용)
type Store interface {
    Add(obj interface{}) error
    Update(obj interface{}) error
    Delete(obj interface{}) error
    Replace([]interface{}, string) error
}

// Lister: 읽기 인터페이스 (Controller가 사용)
// 코드 생성으로 타입 안전한 버전 자동 생성됨
type PodLister interface {
    List(selector labels.Selector) ([]*v1.Pod, error)
    Pods(namespace string) PodNamespaceLister
}

// 같은 in-memory 맵을 공유하지만 인터페이스로 역할 분리
// Controller는 API 서버에 직접 요청하지 않고 로컬 캐시에서 읽음
```

---

## WorkQueue 종류

### 기본 큐: 중복 제거 보장

```go
// k8s.io/client-go/util/workqueue/queue.go
type Type struct {
    queue      []t              // FIFO 순서 유지
    dirty      set              // 큐에 있거나 처리 중인 항목
    processing set              // 현재 처리 중인 항목
    cond       *sync.Cond
}

// 중복 제거 동작:
// Add("foo") → queue: ["foo"], dirty: {"foo"}
// Add("bar") → queue: ["foo", "bar"], dirty: {"foo", "bar"}
// Add("foo") → dirty에 이미 있음, 무시 (중복)
// Get() → "foo" 반환, processing: {"foo"}
// Add("foo") → processing 중이므로 dirty에만 추가 (보류)
// Done("foo") → processing에서 제거, dirty에 있으면 다시 queue에 추가
```

### DelayingQueue: 재시도 지연

```go
// 에러 발생 시 즉시 재시도하지 않고 지연
type DelayingInterface interface {
    Interface
    AddAfter(item interface{}, duration time.Duration)
}

// 사용 패턴
func (c *Controller) processNextItem() bool {
    key, quit := c.queue.Get()
    // ...
    err := c.reconcile(key.(string))
    if err != nil {
        // 5초 후 재시도 (즉시 재시도하면 같은 오류 반복)
        c.queue.AddAfter(key, 5*time.Second)
    }
}
```

### RateLimitingQueue: 지수 백오프

```go
// 연속 실패 시 대기 시간이 지수적으로 증가
type RateLimitingInterface interface {
    DelayingInterface
    AddRateLimited(item interface{})
    Forget(item interface{})           // 성공 시 실패 횟수 초기화
    NumRequeues(item interface{}) int  // 재시도 횟수
}

// BucketRateLimiter + ItemExponentialFailureRateLimiter 조합
// 1회 실패: 5ms 대기
// 2회 실패: 10ms 대기
// 3회 실패: 20ms 대기
// ...최대 1000s

// 실제 사용 패턴
func (c *Controller) processNextItem() bool {
    key, quit := c.queue.Get()
    defer c.queue.Done(key)

    err := c.reconcile(key.(string))
    if err != nil {
        c.queue.AddRateLimited(key)  // 지수 백오프로 재시도
        return true
    }
    c.queue.Forget(key)  // 성공: 실패 카운터 초기화
    return true
}
```

---

## Builder/Functional Options 패턴

### Go에서 객체 생성의 관용적 방법

**문제**: Go는 함수 오버로드가 없습니다. 선택적 파라미터를 어떻게 처리할까요?

**방법 1: 설정 구조체** (단순하지만 유연성 낮음)
```go
type Config struct {
    Timeout    time.Duration
    MaxRetries int
    Logger     Logger
}

func NewController(cfg Config) *Controller { ... }

// 사용: 모든 필드를 명시해야 함
ctrl := NewController(Config{
    Timeout:    30 * time.Second,
    MaxRetries: 3,
})
```

**방법 2: Functional Options** (K8s, gRPC, Docker 등 표준)
```go
type Option func(*Controller)

func WithTimeout(d time.Duration) Option {
    return func(c *Controller) {
        c.timeout = d
    }
}

func WithMaxRetries(n int) Option {
    return func(c *Controller) {
        c.maxRetries = n
    }
}

func WithLogger(l Logger) Option {
    return func(c *Controller) {
        c.logger = l
    }
}

func NewController(opts ...Option) *Controller {
    // 합리적인 기본값 설정
    c := &Controller{
        timeout:    30 * time.Second,
        maxRetries: 3,
        logger:     defaultLogger,
    }
    // 옵션 적용
    for _, opt := range opts {
        opt(c)
    }
    return c
}

// 사용: 필요한 것만 지정
ctrl := NewController(
    WithTimeout(10 * time.Second),
    // MaxRetries, Logger는 기본값 사용
)
```

Functional Options의 장점:
- 기본값과 선택적 설정을 명확히 분리
- 새 옵션 추가 시 하위 호환성 유지
- 테스트에서 특정 설정만 변경 용이

### 체이닝 빌더 패턴

controller-runtime에서 사용하는 빌더 패턴:

```go
// sigs.k8s.io/controller-runtime/pkg/builder/controller.go
type Builder struct {
    forInput         ForInput
    ownsInput        []OwnsInput
    mgr              manager.Manager
    // ...
}

func (blder *Builder) For(object client.Object, opts ...ForOption) *Builder {
    blder.forInput = ForInput{object: object, opts: opts}
    return blder  // 자신을 반환해 체이닝 가능
}

func (blder *Builder) Owns(object client.Object, opts ...OwnsOption) *Builder {
    blder.ownsInput = append(blder.ownsInput, OwnsInput{object: object, opts: opts})
    return blder
}

func (blder *Builder) Complete(r reconcile.Reconciler) error {
    // 실제 컨트롤러 구성 및 등록
    return blder.build(r)
}

// 실제 사용
err := ctrl.NewControllerManagedBy(mgr).
    For(&appsv1.Deployment{}).           // Deployment를 감시
    Owns(&corev1.ReplicaSet{}).          // ReplicaSet도 감시 (소유자 관계)
    WithOptions(controller.Options{      // 옵션 추가
        MaxConcurrentReconciles: 5,
    }).
    Complete(&DeploymentReconciler{})
```

---

## Kubernetes가 사용하는 Go 패턴 총정리

| 패턴 | 사용 위치 | 핵심 아이디어 |
|------|-----------|--------------|
| **Reconciliation Loop** | 모든 컨트롤러 | 현재↔원하는 상태 비교, 멱등 조정 |
| **Informer/Cache** | client-go | API 서버 대신 로컬 캐시 사용 |
| **WorkQueue** | 모든 컨트롤러 | 중복 제거 + 지수 백오프 재시도 |
| **Functional Options** | client-go, gRPC | 유연한 선택적 설정 |
| **Builder** | controller-runtime | 체이닝으로 복잡한 객체 구성 |
| **Interface Plugin** | Scheduler, CSI, CNI | 교체 가능한 구현체 |
| **Context 전파** | 모든 함수 | 취소, 타임아웃, 값 전달 |
| **채널 기반 종료** | 모든 goroutine | `<-chan struct{}` stopCh |
| **code-gen** | Informer, Lister | boilerplate 자동 생성 |
| **Structured Logging** | klog, logr | 필드 기반 로그 (fmt.Println 금지) |

### Context와 stopCh 패턴

Kubernetes는 두 가지 종료 신호 방식을 혼용합니다:

```go
// 구형: stopCh 패턴 (단순 종료 신호)
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
    defer c.queue.ShutDown()

    // Informer 시작
    go c.informer.Run(stopCh)

    // 워커 시작
    for i := 0; i < workers; i++ {
        go wait.Until(c.worker, time.Second, stopCh)
    }

    <-stopCh // 종료 신호 대기
}

// 신형: Context 패턴 (취소, 타임아웃, 값 전달 모두 지원)
func (c *Controller) Run(ctx context.Context, workers int) error {
    defer c.queue.ShutDown()

    go c.informer.Run(ctx.Done())

    for i := 0; i < workers; i++ {
        go func() {
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    c.processNextItem(ctx)
                }
            }
        }()
    }

    <-ctx.Done()
    return ctx.Err()
}
```

---

## 실제 코드 읽기 연습

### DeploymentController 분석 경로

```
1. cmd/kube-controller-manager/main.go
   → NewControllerManagerCommand()
   → startDeploymentController()

2. pkg/controller/deployment/deployment_controller.go
   → type DeploymentController struct
   → func NewDeploymentController(...)
   → func (dc *DeploymentController) Run(ctx, workers)
   → func (dc *DeploymentController) syncDeployment(ctx, key)

3. pkg/controller/deployment/sync.go
   → func (dc *DeploymentController) sync(ctx, d, rsList)
   → func (dc *DeploymentController) scale(ctx, deployment, allRSs, newRS)

각 함수가 어떤 인터페이스를 통해 API를 호출하는지,
어떤 Lister를 통해 캐시에서 읽는지 추적하면
Controller 패턴의 전체 그림이 보입니다.
```
