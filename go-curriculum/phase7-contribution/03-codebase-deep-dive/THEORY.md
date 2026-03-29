# 03: Agent Sandbox 코드베이스 딥다이브

## 코드 읽기 전략

대규모 Go 프로젝트를 처음 접할 때의 접근법:

```
1. 인터페이스부터 읽어라 (구현보다 계약이 중요)
2. 테스트 코드를 먼저 읽어라 (사용법을 보여준다)
3. godoc 주석을 무시하지 말라 (설계 의도가 담겨 있다)
4. cmd/ → api/ → controllers/ 순서로 진입하라
```

---

## 진입점: cmd/agent-sandbox-controller

모든 것은 여기서 시작됩니다:

```
cmd/agent-sandbox-controller/main.go
```

이 파일은:
1. controller-runtime의 `ctrl.NewManager()`로 매니저를 생성
2. 각 컨트롤러를 매니저에 등록
3. `mgr.Start()`로 reconciliation 루프를 시작

**Phase 6에서 배운 Controller 패턴이 그대로 적용됩니다.**

---

## API 타입: api/v1alpha1/

### Sandbox 타입 분석

```go
// api/v1alpha1/types.go (예상 구조)

// Sandbox는 격리된 단일 컨테이너 워크로드를 나타냅니다
type Sandbox struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec              SandboxSpec   `json:"spec,omitempty"`
    Status            SandboxStatus `json:"status,omitempty"`
}

type SandboxSpec struct {
    PodTemplate corev1.PodTemplateSpec `json:"podTemplate"`
    // ... 추가 필드
}

type SandboxStatus struct {
    Phase      SandboxPhase       `json:"phase,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    PodName    string             `json:"podName,omitempty"`
    // ...
}
```

**읽기 포인트:**
- `Spec` — 사용자가 선언하는 desired state
- `Status` — 컨트롤러가 관리하는 observed state
- `Conditions` — 리소스의 현재 상태를 표현하는 표준 패턴

### 확장 API: extensions/api/v1alpha1/

SandboxTemplate, SandboxClaim, SandboxWarmPool 타입이 여기에 정의됩니다. 코어 API와 분리하여 선택적으로 설치할 수 있게 설계되었습니다.

---

## 컨트롤러: controllers/

### Sandbox 컨트롤러의 Reconcile 루프

```go
// controllers/sandbox_controller.go (예상 구조)

func (r *SandboxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Sandbox 리소스를 가져온다
    sandbox := &v1alpha1.Sandbox{}
    if err := r.Get(ctx, req.NamespacedName, sandbox); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. 소유한 Pod을 확인한다
    // 3. Pod이 없으면 생성한다
    // 4. Pod이 있으면 spec과 동기화한다
    // 5. Status를 업데이트한다

    return ctrl.Result{}, nil
}
```

**Phase 6의 Controller 패턴과 직접 연결:**
- `Reconcile()` — 선언적 상태 동기화
- `client.IgnoreNotFound()` — 삭제된 리소스 처리
- `ctrl.Result{}` — 재큐 제어
- Owner Reference — Sandbox가 삭제되면 Pod도 삭제

### 확장 컨트롤러: extensions/controllers/

| 컨트롤러 | 역할 |
|----------|------|
| SandboxTemplate | 템플릿 유효성 검증, 참조 관리 |
| SandboxClaim | Claim → Sandbox 바인딩 |
| SandboxWarmPool | 미리 워밍된 Pod 풀 관리, 스케일링 |

---

## 테스트 구조

### 단위 테스트

```
controllers/*_test.go          # 컨트롤러 단위 테스트
api/v1alpha1/*_test.go         # API 유효성 검증 테스트
```

controller-runtime의 `envtest` 패키지를 사용하여 etcd + API 서버를 로컬에서 실행하고 테스트합니다.

### E2E 테스트

```
test/e2e/
├── framework/                 # 테스트 유틸리티 프레임워크
├── sandbox_test.go            # Sandbox E2E
└── ...
```

E2E 테스트는 실제 Kubernetes 클러스터(kind)에서 실행됩니다.

### 테스트 실행

```bash
# 단위 테스트
make test

# E2E 테스트 (kind 클러스터 필요)
make test-e2e

# 특정 패키지만
go test ./controllers/... -v

# race detector와 함께 (이슈 #331 관련!)
go test -race ./...
```

---

## 빌드 시스템

### Makefile 핵심 타겟

```bash
make build              # 컨트롤러 바이너리 빌드
make test               # 단위 테스트
make test-e2e           # E2E 테스트
make manifests          # CRD 매니페스트 생성 (controller-gen)
make generate           # DeepCopy, 클라이언트 코드 생성
make docker-build       # Docker 이미지 빌드
make install            # CRD를 클러스터에 설치
make deploy             # 컨트롤러를 클러스터에 배포
```

### 코드 생성

Kubernetes 프로젝트는 마커 주석으로 코드를 자동 생성합니다:

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
type Sandbox struct { ... }
```

`make generate`를 실행하면:
- `zz_generated.deepcopy.go` — DeepCopyObject 구현
- `clients/k8s/clientset/` — 타입드 클라이언트셋
- CRD YAML 매니페스트

---

## 메트릭과 옵저버빌리티

### internal/metrics/

Prometheus 메트릭이 정의됩니다:
- Sandbox 생성/삭제 카운터
- Reconciliation 레이턴시 히스토그램
- WarmPool 크기/사용률 게이지

이 부분은 Phase 6의 Prometheus 패턴과 직접 연결됩니다.

---

## 코드 읽기 실습 순서

1. **`api/v1alpha1/types.go`** — Sandbox, SandboxSpec, SandboxStatus 타입 이해
2. **`controllers/sandbox_controller.go`** — Reconcile 로직의 큰 그림
3. **`controllers/sandbox_controller_test.go`** — 테스트로 사용법 확인
4. **`cmd/agent-sandbox-controller/main.go`** — 진입점과 매니저 구성
5. **`extensions/api/v1alpha1/`** — 확장 CRD 타입
6. **`test/e2e/`** — E2E 테스트로 실제 동작 확인

---

## 핵심 정리

1. `cmd/` → `api/` → `controllers/` 순서로 코드를 읽어라
2. controller-runtime 기반 표준 Reconcile 패턴
3. `make test`, `make test-e2e`로 테스트 실행
4. kubebuilder 마커 주석으로 코드 자동 생성
5. 테스트 코드가 가장 좋은 문서다
