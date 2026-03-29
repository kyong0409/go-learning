# Kubernetes와 Go - client-go, Informer, Operator 패턴

## Kubernetes 아키텍처 개요

```
Control Plane (마스터 노드)
┌────────────────────────────────────────┐
│  kube-apiserver  ← 모든 통신의 중심    │
│  etcd            ← 클러스터 상태 저장  │
│  kube-scheduler  ← Pod 배치 결정       │
│  controller-mgr  ← 상태 조정 루프      │
└────────────────────────────────────────┘
           ↕ (gRPC + REST)
Worker Nodes
┌──────────────────┐  ┌──────────────────┐
│  kubelet         │  │  kubelet         │
│  kube-proxy      │  │  kube-proxy      │
│  container-rt    │  │  container-rt    │
│  [Pod] [Pod]     │  │  [Pod] [Pod]     │
└──────────────────┘  └──────────────────┘
```

**Go와 Kubernetes의 관계:**
Kubernetes 자체가 Go로 작성되어 있습니다. client-go를 사용하면
kubectl이 하는 모든 작업을 프로그래밍 방식으로 수행할 수 있습니다.

---

## client-go 라이브러리

### kubeconfig 로드

```go
import (
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/rest"
)

func getK8sClient() (*kubernetes.Clientset, error) {
    var config *rest.Config
    var err error

    // 클러스터 외부: ~/.kube/config 사용
    kubeconfig := os.Getenv("KUBECONFIG")
    if kubeconfig == "" {
        home, _ := os.UserHomeDir()
        kubeconfig = filepath.Join(home, ".kube", "config")
    }

    config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        // 클러스터 내부 (Pod에서 실행): 서비스 어카운트 사용
        config, err = rest.InClusterConfig()
        if err != nil {
            return nil, fmt.Errorf("kubeconfig 로드 실패: %w", err)
        }
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("클라이언트 생성 실패: %w", err)
    }
    return clientset, nil
}
```

### 기본 연산: List, Get, Create, Delete

```go
ctx := context.Background()
namespace := "default"

// Pod 목록 조회
pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
    LabelSelector: "app=bookmarkapi", // 레이블 필터
})
for _, pod := range pods.Items {
    fmt.Printf("Pod: %s, 상태: %s\n", pod.Name, pod.Status.Phase)
}

// 특정 Pod 조회
pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, "bookmarkapi-abc123", metav1.GetOptions{})

// Deployment 생성
deployment := &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "bookmarkapi",
        Namespace: namespace,
    },
    Spec: appsv1.DeploymentSpec{
        Replicas: int32Ptr(3),
        Selector: &metav1.LabelSelector{
            MatchLabels: map[string]string{"app": "bookmarkapi"},
        },
        Template: corev1.PodTemplateSpec{
            ObjectMeta: metav1.ObjectMeta{
                Labels: map[string]string{"app": "bookmarkapi"},
            },
            Spec: corev1.PodSpec{
                Containers: []corev1.Container{
                    {
                        Name:  "bookmarkapi",
                        Image: "myregistry/bookmarkapi:v1.0.0",
                        Ports: []corev1.ContainerPort{{ContainerPort: 8080}},
                    },
                },
            },
        },
    },
}

created, err := clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})

// 리소스 삭제
err = clientset.CoreV1().Pods(namespace).Delete(ctx, "old-pod", metav1.DeleteOptions{})
```

### Watch: 실시간 이벤트 감시

```go
// Pod 변경 사항 실시간 감시
watcher, err := clientset.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{
    LabelSelector: "app=bookmarkapi",
})
if err != nil {
    log.Fatal(err)
}
defer watcher.Stop()

for event := range watcher.ResultChan() {
    pod, ok := event.Object.(*corev1.Pod)
    if !ok {
        continue
    }
    switch event.Type {
    case watch.Added:
        fmt.Printf("Pod 추가됨: %s\n", pod.Name)
    case watch.Modified:
        fmt.Printf("Pod 변경됨: %s → %s\n", pod.Name, pod.Status.Phase)
    case watch.Deleted:
        fmt.Printf("Pod 삭제됨: %s\n", pod.Name)
    }
}
```

---

## 핵심 패턴: Informer

직접 Watch를 사용하면 API 서버 부하가 크고, 재연결 처리가 복잡합니다.
Informer는 로컬 캐시 + 이벤트 알림으로 이 문제를 해결합니다.

```
Informer 동작 방식:
┌─────────────────────────────────────────┐
│              Informer                   │
│  ┌──────────┐     ┌─────────────────┐  │
│  │ Reflector│────▶│  Thread-safe    │  │
│  │ (Watch)  │     │  Store (캐시)   │  │
│  └──────────┘     └────────┬────────┘  │
│                            │            │
│                   ┌────────▼────────┐  │
│                   │  Event Handler  │  │
│                   │ OnAdd/OnUpdate  │  │
│                   │ OnDelete        │  │
│                   └─────────────────┘  │
└─────────────────────────────────────────┘
```

```go
import (
    "k8s.io/client-go/informers"
    "k8s.io/client-go/tools/cache"
)

func startPodInformer(clientset *kubernetes.Clientset) {
    // 공유 Informer 팩토리 생성 (30초마다 전체 재동기화)
    factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)

    // Pod Informer 등록
    podInformer := factory.Core().V1().Pods()

    // 이벤트 핸들러 등록
    podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
        AddFunc: func(obj interface{}) {
            pod := obj.(*corev1.Pod)
            fmt.Printf("Pod 추가: %s/%s\n", pod.Namespace, pod.Name)
        },
        UpdateFunc: func(old, new interface{}) {
            oldPod := old.(*corev1.Pod)
            newPod := new.(*corev1.Pod)
            if oldPod.ResourceVersion == newPod.ResourceVersion {
                return // 실제 변경 없음
            }
            fmt.Printf("Pod 업데이트: %s\n", newPod.Name)
        },
        DeleteFunc: func(obj interface{}) {
            pod := obj.(*corev1.Pod)
            fmt.Printf("Pod 삭제: %s\n", pod.Name)
        },
    })

    stopCh := make(chan struct{})
    defer close(stopCh)

    factory.Start(stopCh)           // 모든 Informer 시작
    factory.WaitForCacheSync(stopCh) // 초기 캐시 동기화 완료 대기

    // Lister: API 서버 대신 캐시에서 읽기 (빠르고 서버 부하 없음)
    lister := podInformer.Lister()
    pods, err := lister.Pods("default").List(labels.Everything())

    <-stopCh
}
```

---

## 조정 루프 (Reconciliation Loop)

Kubernetes의 핵심 패턴입니다. "원하는 상태(Desired State)"와
"현재 상태(Current State)"를 비교하여 일치시킵니다.

```go
// 조정 루프의 기본 구조
func reconcile(ctx context.Context, key string) error {
    // 1. 원하는 상태 읽기 (CRD 또는 리소스)
    namespace, name, _ := cache.SplitMetaNamespaceKey(key)
    desired, err := myResourceLister.MyResources(namespace).Get(name)
    if errors.IsNotFound(err) {
        // 리소스가 삭제됨 → 정리 작업
        return cleanup(namespace, name)
    }

    // 2. 현재 상태 확인
    current, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
    if errors.IsNotFound(err) {
        // 현재 상태 없음 → 생성
        return createDeployment(ctx, desired)
    }

    // 3. 차이 계산 및 동기화
    if needsUpdate(desired, current) {
        return updateDeployment(ctx, desired, current)
    }

    // 4. 상태 업데이트
    return updateStatus(ctx, desired, "Ready")
}
```

---

## Custom Resource Definition (CRD)

CRD로 Kubernetes API를 확장하여 나만의 리소스 타입을 정의할 수 있습니다.

```yaml
# bookmarksite-crd.yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: bookmarksites.bookmark.example.com
spec:
  group: bookmark.example.com
  names:
    kind: BookmarkSite
    plural: bookmarksites
    singular: bookmarksite
    shortNames: [bms]
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                replicas:
                  type: integer
                image:
                  type: string
            status:
              type: object
```

```go
// Go 타입 정의 (controller-runtime 스타일)
type BookmarkSite struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   BookmarkSiteSpec   `json:"spec,omitempty"`
    Status BookmarkSiteStatus `json:"status,omitempty"`
}

type BookmarkSiteSpec struct {
    Replicas int32  `json:"replicas"`
    Image    string `json:"image"`
}

type BookmarkSiteStatus struct {
    ReadyReplicas int32  `json:"readyReplicas"`
    Phase         string `json:"phase"` // Pending, Running, Failed
}
```

---

## Operator 패턴

**오퍼레이터 = CRD + 커스텀 컨트롤러**

```
사용자 → kubectl apply -f bookmarksite.yaml
                ↓
        [kube-apiserver]
                ↓ (저장)
            [etcd]
                ↓ (Watch 이벤트)
    [BookmarkSite Controller] ← 오퍼레이터
                ↓ (조정)
    [Deployment] [Service] [ConfigMap]
        (실제 K8s 리소스 생성/관리)
```

### controller-runtime으로 Reconciler 구현

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
    ctrl "sigs.k8s.io/controller-runtime"
)

type BookmarkSiteReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// Reconcile: 원하는 상태로 조정
func (r *BookmarkSiteReconciler) Reconcile(
    ctx context.Context,
    req ctrl.Request, // namespace/name
) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 1. 리소스 조회
    site := &bookmarkv1.BookmarkSite{}
    if err := r.Get(ctx, req.NamespacedName, site); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil // 삭제된 경우 무시
        }
        return ctrl.Result{}, err
    }

    // 2. Deployment 존재 확인
    deploy := &appsv1.Deployment{}
    err := r.Get(ctx, types.NamespacedName{
        Name:      site.Name,
        Namespace: site.Namespace,
    }, deploy)

    if errors.IsNotFound(err) {
        // 3. Deployment 생성
        deploy = r.buildDeployment(site)
        if err := r.Create(ctx, deploy); err != nil {
            return ctrl.Result{}, err
        }
        log.Info("Deployment 생성됨", "name", deploy.Name)
        return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
    }

    // 4. 상태 업데이트
    site.Status.ReadyReplicas = deploy.Status.ReadyReplicas
    if err := r.Status().Update(ctx, site); err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}
```

### Finalizer: 삭제 전 정리 작업

```go
const finalizerName = "bookmark.example.com/cleanup"

func (r *BookmarkSiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    site := &bookmarkv1.BookmarkSite{}
    r.Get(ctx, req.NamespacedName, site)

    // 삭제 중인지 확인
    if !site.DeletionTimestamp.IsZero() {
        if controllerutil.ContainsFinalizer(site, finalizerName) {
            // 정리 작업 (외부 리소스 삭제, 데이터 백업 등)
            if err := r.cleanup(ctx, site); err != nil {
                return ctrl.Result{}, err
            }
            // Finalizer 제거 → 실제 삭제 진행
            controllerutil.RemoveFinalizer(site, finalizerName)
            r.Update(ctx, site)
        }
        return ctrl.Result{}, nil
    }

    // Finalizer 추가 (처음 생성 시)
    if !controllerutil.ContainsFinalizer(site, finalizerName) {
        controllerutil.AddFinalizer(site, finalizerName)
        r.Update(ctx, site)
    }

    // 정상 조정 로직...
    return ctrl.Result{}, nil
}
```

---

## kubebuilder로 빠른 시작

> **버전 현황 (2026년 3월 기준)**: kubebuilder v4.13.1, controller-runtime v0.23.3이 최신 안정 버전입니다. kubebuilder v4.11+부터 Go 1.25 이상을 요구합니다.

```bash
# kubebuilder 설치 및 프로젝트 초기화
# kubebuilder v4.13.1 (latest), controller-runtime v0.23.3
kubebuilder init --domain example.com --repo github.com/myorg/bookmark-operator

# API (CRD + Controller) 생성
kubebuilder create api --group bookmark --version v1 --kind BookmarkSite

# 생성되는 파일 구조:
# api/v1/bookmarksite_types.go     ← 타입 정의
# controllers/bookmarksite_controller.go ← Reconciler
# config/crd/                       ← CRD YAML
# config/rbac/                      ← RBAC 권한

# CRD 설치
make install

# 로컬에서 컨트롤러 실행 (디버깅용)
make run
```

---

## Status Conditions 패턴

```go
// Kubernetes 표준 조건 패턴
meta.SetStatusCondition(&site.Status.Conditions, metav1.Condition{
    Type:               "Ready",
    Status:             metav1.ConditionTrue,
    Reason:             "DeploymentReady",
    Message:            "모든 레플리카가 준비됨",
    ObservedGeneration: site.Generation,
})

meta.SetStatusCondition(&site.Status.Conditions, metav1.Condition{
    Type:    "Progressing",
    Status:  metav1.ConditionFalse,
    Reason:  "DeploymentComplete",
    Message: "배포 완료",
})

// kubectl get bookmarksites -o wide 로 조건 확인 가능
// READY   PROGRESSING
// True    False
```

---

## 학습 참고: Kubernetes 생태계 기여

```bash
# kubernetes/kubernetes 코드 구조 이해
# staging/src/k8s.io/client-go/  ← client-go 소스
# pkg/controller/                ← 내장 컨트롤러
# cmd/kube-controller-manager/   ← 컨트롤러 매니저

# 기여 시작:
# 1. Good First Issue 레이블 찾기
# 2. /sig-cli, /sig-apps 등 SIG 참여
# 3. E2E 테스트, 단위 테스트 작성
# 4. 코드 리뷰: /lgtm, /approve 프로세스

# 필수 도구
kind create cluster       # 로컬 K8s 클러스터
kubectl apply -f crd.yaml
kubectl logs -f my-operator-pod
kubectl describe bookmarksite my-site
```
