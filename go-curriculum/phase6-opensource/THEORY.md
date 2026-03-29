# Phase 6 이론: K8s 생태계 오픈소스 딥다이브

## Phase 6 학습 목표

Phase 6는 단순히 Go 문법을 배우는 단계를 넘어, **실제 프로덕션 시스템이 어떻게 설계되고 구현되는가**를 직접 코드로 확인하는 과정입니다.

목표:
1. K8s 생태계 오픈소스 코드를 읽고 핵심 패턴을 추출하는 능력
2. 수십만 줄 규모의 코드베이스에서 진입점을 찾는 방법론
3. 반복적으로 나타나는 Go 관용 패턴을 인식하고 직접 구현
4. 오픈소스 기여를 위한 코드 컨벤션과 PR 프로세스 이해

---

## 왜 오픈소스 코드를 읽어야 하는가

### 교과서 코드와 프로덕션 코드의 차이

교과서 예제:
```go
// 단순하고 이해하기 쉬움, 하지만 실제로 쓰기 어려움
func NewServer(port int) *Server {
    return &Server{port: port}
}
```

실제 Kubernetes 코드 (client-go):
```go
// 복잡하지만 이유가 있음: 확장성, 안전성, 하위 호환성
func NewSharedIndexInformer(
    lw ListerWatcher,
    exampleObject runtime.Object,
    defaultEventHandlerResyncPeriod time.Duration,
    indexers Indexers,
) SharedIndexInformer {
    realClock := &clock.RealClock{}
    sharedIndexInformer := &sharedIndexInformer{
        processor:                       &sharedProcessor{clock: realClock},
        store:                           NewIndexer(DeletionHandlingMetaNamespaceKeyFunc, indexers),
        listerWatcher:                   lw,
        objectType:                      exampleObject,
        resyncCheckPeriod:               defaultEventHandlerResyncPeriod,
        defaultEventHandlerResyncPeriod: defaultEventHandlerResyncPeriod,
        cacheMutationDetector:           NewCacheMutationDetector(fmt.Sprintf("%T", exampleObject)),
        clock:                           realClock,
    }
    return sharedIndexInformer
}
```

프로덕션 코드는 다음 문제들을 모두 해결해야 합니다:
- **동시성**: 여러 goroutine에서 안전하게 접근
- **오류 처리**: 네트워크 오류, 타임아웃, 재시도
- **관측 가능성**: 메트릭, 로그, 트레이싱
- **하위 호환성**: 이전 버전 API 유지
- **테스트 가능성**: 인터페이스 분리, 목(Mock) 주입

---

## 오픈소스 코드 리딩 방법론

### 1단계: 진입점 찾기

대규모 Go 프로젝트의 표준 구조:
```
프로젝트 루트/
├── main.go          (또는 없음)
├── cmd/             ← 진입점들. 여기서 시작
│   ├── server/
│   │   └── main.go  ← 실제 main 함수
│   └── cli/
│       └── main.go
├── pkg/             ← 외부 공개 패키지 (import 가능)
├── internal/        ← 내부 패키지 (외부 import 불가)
├── vendor/          ← 의존성 (또는 go.sum으로 관리)
└── staging/         ← Kubernetes 특유: 별도 모듈로 분리될 패키지
```

Kubernetes 진입점 탐색 예시:
```
kubernetes/
├── cmd/
│   ├── kube-apiserver/main.go     ← API 서버
│   ├── kube-controller-manager/   ← 컨트롤러 매니저
│   ├── kube-scheduler/            ← 스케줄러
│   └── kubelet/                   ← 노드 에이전트
└── pkg/
    ├── controller/                ← 컨트롤러 구현체들
    ├── scheduler/                 ← 스케줄러 구현
    └── kubelet/                   ← Kubelet 구현
```

### 2단계: 인터페이스 먼저 읽기

Go에서 인터페이스는 "계약서"입니다. 구현 코드보다 인터페이스를 먼저 읽으면 전체 구조가 보입니다.

```go
// 인터페이스만 보면 시스템이 무엇을 하는지 알 수 있음
// k8s.io/client-go/tools/cache/store.go

type Store interface {
    Add(obj interface{}) error      // 객체 추가
    Update(obj interface{}) error   // 객체 갱신
    Delete(obj interface{}) error   // 객체 삭제
    List() []interface{}            // 모든 객체 반환
    Get(obj interface{}) (item interface{}, exists bool, err error)
    GetByKey(key string) (item interface{}, exists bool, err error)
}
// → 이 인터페이스만 보면 "키-값 저장소구나"를 즉시 알 수 있음
```

인터페이스 찾는 방법:
```bash
# 프로젝트에서 인터페이스 정의만 검색
grep -r "^type.*interface" --include="*.go" pkg/ | head -30

# 특정 인터페이스의 구현체 찾기
grep -r "func.*Reconcile\b" --include="*.go" .
```

### 3단계: 테스트 코드로 기대 동작 파악

테스트는 "이 코드가 어떻게 사용되어야 하는지"를 보여주는 최고의 문서입니다.

```go
// 테스트 코드에서 사용 패턴 파악
// k8s.io/client-go/tools/cache/store_test.go

func TestCache(t *testing.T) {
    store := NewStore(MetaNamespaceKeyFunc)

    // Add
    store.Add(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns"}})

    // Get - 이렇게 사용한다
    item, exists, err := store.GetByKey("ns/pod1")
    if !exists {
        t.Errorf("expected pod to exist")
    }
    // → 키 형식이 "namespace/name"임을 테스트에서 확인
}
```

### 4단계: git blame으로 변경 이유 추적

코드가 왜 이렇게 생겼는지 이해하는 가장 강력한 방법:

```bash
# 특정 줄이 왜 이렇게 작성되었는지 확인
git blame pkg/controller/deployment/deployment_controller.go

# 커밋 메시지에서 맥락 파악
git log --follow -p pkg/controller/deployment/deployment_controller.go | less

# 특정 변경의 배경 Issue/PR 확인
# 커밋 메시지에 "Fixes #12345" 또는 "kubernetes/kubernetes#12345" 형태로 링크됨
```

---

## K8s 생태계 프로젝트 맵

```
┌─────────────────────────────────────────────────────────────────┐
│                        K8s 생태계                                │
│                                                                   │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │  Terraform  │    │    Helm     │    │   kubectl/client-go │  │
│  │ (프로비저닝) │    │ (패키지관리) │    │    (클라이언트)     │  │
│  └──────┬──────┘    └──────┬──────┘    └──────────┬──────────┘  │
│         │                  │                       │             │
│         └──────────────────┴───────────────────────┘            │
│                            │                                     │
│                  ┌─────────▼──────────┐                         │
│                  │    Kubernetes      │                          │
│                  │  (오케스트레이터)   │                          │
│                  │  API Server        │                          │
│                  │  Controller Manager│                          │
│                  │  Scheduler         │                          │
│                  └────────┬───────────┘                         │
│                           │                                     │
│              ┌────────────┴─────────────┐                      │
│              │                          │                       │
│    ┌─────────▼──────┐       ┌──────────▼────────┐             │
│    │      etcd      │       │   containerd/runc  │             │
│    │ (상태 저장소)   │       │  (컨테이너 런타임)  │             │
│    └────────────────┘       └───────────────────┘             │
│                                                                  │
│  ┌─────────────────────────────────────────────┐               │
│  │               Prometheus                     │               │
│  │         (전체 스택 모니터링)                  │               │
│  └─────────────────────────────────────────────┘               │
└─────────────────────────────────────────────────────────────────┘
```

### 각 프로젝트의 역할과 GitHub 위치

| 프로젝트 | 역할 | GitHub | 핵심 Go 패턴 |
|---------|------|--------|-------------|
| **Kubernetes** | 컨테이너 오케스트레이션 | `kubernetes/kubernetes` | Controller, Informer, WorkQueue |
| **client-go** | K8s Go 클라이언트 | `kubernetes/client-go` | Reflector, Store, Lister |
| **etcd** | 분산 KV 스토어 | `etcd-io/etcd` | MVCC, Watch, Raft, Lease |
| **Prometheus** | 메트릭 모니터링 | `prometheus/prometheus` | Registry, Collector, Pull 모델 |
| **client_golang** | Prometheus Go 클라이언트 | `prometheus/client_golang` | Collector 인터페이스, Vec 타입 |
| **Docker (Moby)** | 컨테이너 엔진 | `moby/moby` | Plugin 아키텍처, 레이어 |
| **containerd** | 컨테이너 런타임 | `containerd/containerd` | Snapshotter, OCI 표준 |
| **Terraform** | 인프라 프로비저닝 | `hashicorp/terraform` | DAG, Provider 플러그인 |
| **Helm** | K8s 패키지 매니저 | `helm/helm` | 템플릿 엔진, 릴리즈 관리 |

### 핵심 디렉토리 안내

**kubernetes/kubernetes:**
```
kubernetes/
├── cmd/kube-apiserver/          ← API 서버 진입점
├── cmd/kube-controller-manager/ ← 컨트롤러 매니저 진입점
├── pkg/controller/              ← 내장 컨트롤러들
│   ├── deployment/              ← Deployment 컨트롤러
│   ├── replicaset/              ← ReplicaSet 컨트롤러
│   └── statefulset/             ← StatefulSet 컨트롤러
├── staging/src/k8s.io/          ← client-go, api-machinery 등
└── vendor/                      ← 의존성
```

**prometheus/client_golang:**
```
prometheus/
├── counter.go        ← Counter 타입 구현
├── gauge.go          ← Gauge 타입 구현
├── histogram.go      ← Histogram 타입 구현
├── registry.go       ← Registry 구현 (핵심)
├── collector.go      ← Collector 인터페이스 정의
└── promhttp/
    └── http.go       ← HTTP 핸들러
```

---

## 오픈소스 기여 첫걸음

### good-first-issue 찾기

1. GitHub에서 `label:good-first-issue` 필터 사용
2. K8s 관련 좋은 시작점:
   - `kubernetes/kubernetes`: `https://github.com/kubernetes/kubernetes/issues?q=label%3A"good+first+issue"`
   - `prometheus/client_golang`: 비교적 작은 코드베이스로 시작하기 좋음
   - `etcd-io/etcd`: 문서화 이슈가 많음

### 코딩 컨벤션

K8s 생태계는 엄격한 코딩 컨벤션을 따릅니다:

```go
// 1. 에러 처리: 반드시 처리하고 wrapping
if err != nil {
    return fmt.Errorf("failed to reconcile deployment %s: %w", key, err)
}

// 2. 로깅: klog 또는 logr 인터페이스 사용 (fmt.Println 금지)
logger := klog.FromContext(ctx)
logger.V(4).Info("reconciling", "key", key)

// 3. 컨텍스트: 모든 장기 실행 함수에 ctx 전달
func (c *Controller) Reconcile(ctx context.Context, key string) error { ... }

// 4. 인터페이스 크기: 작고 명확하게
// 나쁜 예:
type BigInterface interface { Method1(); Method2(); ...; Method20() }
// 좋은 예:
type Lister interface { List() []interface{} }
type Getter interface { Get(key string) (interface{}, bool, error) }
```

### PR 프로세스 (Kubernetes 기준)

```
1. Issue 생성 또는 기존 Issue 코멘트로 작업 의사 표명
        ↓
2. Fork → 브랜치 생성 (feature/fix-typo-in-controller)
        ↓
3. 변경 사항 구현 + 테스트 추가
        ↓
4. go test ./... 통과 확인
        ↓
5. PR 생성: 제목에 영향 범위 명시
   예: "fix: correct error message in deployment controller reconcile loop"
        ↓
6. /lgtm, /approve 자동화 봇 시스템
   - Reviewer가 /lgtm 코멘트 → "Looks Good To Me"
   - Approver가 /approve 코멘트 → 머지 가능 상태
        ↓
7. CI 통과 (prow 기반 CI 시스템)
        ↓
8. 자동 머지
```

---

## Phase 6 학습 흐름과 패턴 연결

```
Prometheus (02) ──→ 메트릭 계측 개념 확립
      ↓
Kubernetes (01) ──→ Controller, Informer 패턴 이해
      ↓
etcd (03) ──────→ Watch, MVCC 패턴 이해
      ↓
Docker (04) ────→ 플러그인, 레이어 패턴 이해
      ↓
Terraform (05) ─→ DAG, Provider 패턴 이해
      ↓
과제 A1~A6: 모든 패턴을 직접 구현
```

각 패턴은 독립적으로 보이지만 실제로는 긴밀히 연결됩니다. Kubernetes 컨트롤러는 Prometheus로 메트릭을 내보내고, etcd로 상태를 저장하며, containerd로 컨테이너를 실행합니다. 이 전체 그림을 이해한 상태에서 각 패턴을 학습하면 훨씬 깊은 이해가 가능합니다.
