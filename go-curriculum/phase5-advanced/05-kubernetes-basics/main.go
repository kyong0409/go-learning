// main.go
// client-go를 사용해 Kubernetes 클러스터의 Pod 목록을 조회하는 예제
//
// 사전 요구 사항:
//   - kubectl이 설치되어 있고 클러스터에 연결 가능해야 합니다.
//   - ~/.kube/config 파일이 존재해야 합니다.
//   - 또는 KUBECONFIG 환경 변수로 설정 파일 경로를 지정할 수 있습니다.
//
// 실행 방법:
//   go run main.go
//   go run main.go -namespace kube-system
//   go run main.go -namespace ""    (모든 네임스페이스)
//
// 클러스터 없이 테스트:
//   minikube start      (로컬 클러스터)
//   kind create cluster (Kind 사용)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ============================================================
// CLI 플래그
// ============================================================

var (
	namespace  = flag.String("namespace", "default", "조회할 네임스페이스 (빈 문자열 = 전체)")
	kubeconfig = flag.String("kubeconfig", defaultKubeconfig(), "kubeconfig 파일 경로")
	allNs      = flag.Bool("all-namespaces", false, "모든 네임스페이스 조회")
)

// defaultKubeconfig는 기본 kubeconfig 경로를 반환합니다.
func defaultKubeconfig() string {
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}
	return ""
}

// ============================================================
// 메인 함수
// ============================================================

func main() {
	flag.Parse()

	// kubeconfig 로드
	config, err := loadConfig(*kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kubeconfig 로드 실패: %v\n", err)
		fmt.Fprintln(os.Stderr, "kubectl이 설치되어 있고 클러스터에 연결 가능한지 확인하세요.")
		os.Exit(1)
	}

	// Kubernetes 클라이언트 생성
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "클라이언트 생성 실패: %v\n", err)
		os.Exit(1)
	}

	// 네임스페이스 결정
	ns := *namespace
	if *allNs {
		ns = "" // 빈 문자열 = 전체 네임스페이스
	}

	// Pod 목록 조회
	fmt.Printf("Pod 목록 조회 중 (네임스페이스: %s)...\n\n",
		map[bool]string{true: "전체", false: ns}[ns == ""])

	if err := listPods(clientset, ns); err != nil {
		fmt.Fprintf(os.Stderr, "Pod 조회 실패: %v\n", err)
		os.Exit(1)
	}

	// 노드 정보도 조회 (추가 예시)
	fmt.Println()
	if err := listNodes(clientset); err != nil {
		fmt.Fprintf(os.Stderr, "노드 조회 실패: %v\n", err)
	}
}

// ============================================================
// kubeconfig 로드
// ============================================================

// loadConfig는 kubeconfig 파일 또는 in-cluster 설정을 로드합니다.
// 우선순위:
//  1. --kubeconfig 플래그로 지정한 파일
//  2. KUBECONFIG 환경 변수
//  3. ~/.kube/config
//  4. in-cluster 설정 (Pod 안에서 실행 시)
func loadConfig(kubeconfigPath string) (*rest.Config, error) {
	// in-cluster 설정 시도 (Kubernetes Pod 안에서 실행할 때)
	if kubeconfigPath == "" {
		if config, err := rest.InClusterConfig(); err == nil {
			fmt.Println("in-cluster 설정 사용")
			return config, nil
		}
	}

	// 외부 kubeconfig 파일 로드
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		loadingRules.ExplicitPath = kubeconfigPath
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	return kubeConfig.ClientConfig()
}

// ============================================================
// Pod 목록 조회
// ============================================================

// listPods는 지정한 네임스페이스의 Pod 목록을 표로 출력합니다.
func listPods(clientset *kubernetes.Clientset, namespace string) error {
	// Pod 목록 API 호출
	podList, err := clientset.CoreV1().Pods(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return fmt.Errorf("Pod 목록 API 호출 실패: %w", err)
	}

	if len(podList.Items) == 0 {
		fmt.Printf("네임스페이스 %q에 Pod가 없습니다.\n", namespace)
		return nil
	}

	// 이름순 정렬
	sort.Slice(podList.Items, func(i, j int) bool {
		if podList.Items[i].Namespace != podList.Items[j].Namespace {
			return podList.Items[i].Namespace < podList.Items[j].Namespace
		}
		return podList.Items[i].Name < podList.Items[j].Name
	})

	// tabwriter로 정렬된 표 출력
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tNAME\tSTATUS\tRESTARTS\tAGE")
	fmt.Fprintln(w, "---------\t----\t------\t--------\t---")

	for _, pod := range podList.Items {
		status := podStatus(&pod)
		restarts := podRestarts(&pod)
		age := podAge(&pod)

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			pod.Namespace,
			pod.Name,
			status,
			restarts,
			age,
		)
	}

	w.Flush()
	fmt.Printf("\n총 %d개 Pod\n", len(podList.Items))
	return nil
}

// ============================================================
// 노드 목록 조회
// ============================================================

// listNodes는 클러스터 노드 정보를 출력합니다.
func listNodes(clientset *kubernetes.Clientset) error {
	nodeList, err := clientset.CoreV1().Nodes().List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return fmt.Errorf("노드 목록 API 호출 실패: %w", err)
	}

	fmt.Printf("=== 노드 목록 (%d개) ===\n", len(nodeList.Items))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tROLES\tVERSION")
	fmt.Fprintln(w, "----\t------\t-----\t-------")

	for _, node := range nodeList.Items {
		status := "Unknown"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				if cond.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
			}
		}

		// 역할 추출 (node-role.kubernetes.io/* 레이블)
		roles := []string{}
		for label := range node.Labels {
			if len(label) > 28 && label[:28] == "node-role.kubernetes.io/" {
				roles = append(roles, label[28:])
			}
		}
		if len(roles) == 0 {
			roles = []string{"<none>"}
		}
		sort.Strings(roles)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			node.Name,
			status,
			fmt.Sprintf("%v", roles),
			node.Status.NodeInfo.KubeletVersion,
		)
	}
	w.Flush()
	return nil
}

// ============================================================
// 헬퍼 함수
// ============================================================

// podStatus는 Pod의 현재 상태를 반환합니다.
func podStatus(pod *corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}
	return string(pod.Status.Phase)
}

// podRestarts는 Pod의 총 재시작 횟수를 반환합니다.
func podRestarts(pod *corev1.Pod) int32 {
	var total int32
	for _, cs := range pod.Status.ContainerStatuses {
		total += cs.RestartCount
	}
	return total
}

// podAge는 Pod가 생성된 이후 경과 시간을 반환합니다.
func podAge(pod *corev1.Pod) string {
	if pod.CreationTimestamp.IsZero() {
		return "<unknown>"
	}
	// 단순 표시 (실제 kubectl은 더 정교한 포맷 사용)
	return pod.CreationTimestamp.Format("2006-01-02")
}
