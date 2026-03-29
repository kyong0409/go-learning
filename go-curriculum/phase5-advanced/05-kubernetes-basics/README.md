# 05-kubernetes-basics: Kubernetes client-go 기초

client-go를 사용해 Kubernetes API 서버와 통신하는 방법을 학습합니다.

## 개요

**client-go**는 Kubernetes 공식 Go 클라이언트 라이브러리입니다.
`kubectl`도 내부적으로 client-go를 사용합니다.

### 활용 사례
- kubectl 대체 CLI 도구 개발
- Kubernetes Operator 구현
- 자동화 스크립트
- 모니터링/관리 도구

## 사전 요구 사항

```bash
# 1. kubectl 설치 및 클러스터 연결 확인
kubectl cluster-info

# 2. 로컬 클러스터가 없다면 minikube 또는 kind 사용
# minikube
minikube start

# kind (Kubernetes in Docker)
kind create cluster
```

## 실행 방법

```bash
cd 05-kubernetes-basics
go mod tidy
go run main.go

# 다른 네임스페이스
go run main.go -namespace kube-system

# 모든 네임스페이스
go run main.go -all-namespaces

# kubeconfig 명시적 지정
go run main.go -kubeconfig /path/to/kubeconfig
```

## 주요 개념

### 1. client-go 구조

```
kubernetes.Clientset
├── CoreV1()          → Pods, Services, ConfigMaps, Secrets, Nodes
├── AppsV1()          → Deployments, StatefulSets, DaemonSets
├── BatchV1()         → Jobs, CronJobs
├── NetworkingV1()    → Ingresses, NetworkPolicies
└── RbacV1()          → Roles, ClusterRoles, RoleBindings
```

### 2. 기본 CRUD 패턴

```go
// 목록 조회
pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})

// 단일 조회
pod, err := clientset.CoreV1().Pods("default").Get(ctx, "my-pod", metav1.GetOptions{})

// 생성
created, err := clientset.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{})

// 업데이트
updated, err := clientset.CoreV1().Pods("default").Update(ctx, pod, metav1.UpdateOptions{})

// 삭제
err = clientset.CoreV1().Pods("default").Delete(ctx, "my-pod", metav1.DeleteOptions{})
```

### 3. Watch (실시간 이벤트 감시)

```go
watcher, err := clientset.CoreV1().Pods("default").Watch(ctx, metav1.ListOptions{})
for event := range watcher.ResultChan() {
    pod := event.Object.(*corev1.Pod)
    switch event.Type {
    case watch.Added:
        fmt.Printf("Pod 추가: %s\n", pod.Name)
    case watch.Modified:
        fmt.Printf("Pod 변경: %s\n", pod.Name)
    case watch.Deleted:
        fmt.Printf("Pod 삭제: %s\n", pod.Name)
    }
}
```

### 4. Informer (고성능 캐싱 Watch)

Operator 개발에서는 Watch 대신 Informer를 사용합니다:
```go
factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
podInformer := factory.Core().V1().Pods()

podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
    AddFunc:    func(obj interface{}) { /* ... */ },
    UpdateFunc: func(old, new interface{}) { /* ... */ },
    DeleteFunc: func(obj interface{}) { /* ... */ },
})

factory.Start(stopCh)
factory.WaitForCacheSync(stopCh)
```

### 5. kubeconfig 로드 우선순위

```
1. --kubeconfig 플래그
2. KUBECONFIG 환경 변수
3. ~/.kube/config
4. in-cluster 설정 (Pod 내부 실행 시)
```

## Kubernetes Operator 개념

Operator는 custom resource(CRD)와 controller-runtime을 사용합니다:

```
사용자 → kubectl apply -f myapp.yaml (CRD)
         ↓
     API Server (상태 저장)
         ↓
     Operator Controller (Reconcile 루프)
     │  실제 상태 조회 → 원하는 상태와 비교 → 차이 수정
     └→ Pod/Deployment 생성/수정/삭제
```

### controller-runtime 기본 구조

```go
// Reconciler 인터페이스 구현
type MyAppReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. CRD 리소스 조회
    var myApp myv1.MyApp
    if err := r.Get(ctx, req.NamespacedName, &myApp); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. 원하는 상태와 실제 상태 비교
    // 3. 필요한 변경 적용

    return ctrl.Result{}, nil
}
```

## 학습 포인트

1. client-go는 Kubernetes 모든 리소스에 타입 안전한 API 제공
2. `Watch`와 `Informer`의 차이: Informer는 로컬 캐시로 API 서버 부하 감소
3. Operator 패턴: 선언적 설정 + Reconcile 루프 = 자가 치유 시스템
4. in-cluster vs out-of-cluster: 배포 환경에 따라 설정 방법이 다름
5. `kubebuilder`나 `operator-sdk`로 Operator 개발 시작 가능
