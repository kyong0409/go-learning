// controller.go
// 리소스 컨트롤러 - Kubernetes Controller 패턴 구현
//
// TODO: 아래 타입과 함수들을 완성하세요.
// Reconciler 인터페이스와 WorkQueue 인터페이스를 구현하고,
// Controller가 Watch → Enqueue → Reconcile 루프를 돌도록 합니다.
package main

import (
	"context"
	"time"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// EventType은 리소스에 발생한 변경 종류입니다.
type EventType string

const (
	EventCreate EventType = "CREATE"
	EventUpdate EventType = "UPDATE"
	EventDelete EventType = "DELETE"
)

// Event는 리소스 변경 이벤트를 표현합니다.
type Event struct {
	Type     EventType
	Resource Resource
}

// ResourceSpec은 리소스의 원하는 상태입니다.
type ResourceSpec struct {
	Replicas int
	Image    string
}

// ResourceStatus는 리소스의 현재 상태입니다.
type ResourceStatus struct {
	ReadyReplicas int
	Phase         string // "Pending", "Running", "Failed"
}

// Resource는 컨트롤러가 관리하는 오브젝트입니다.
type Resource struct {
	Name   string
	Spec   ResourceSpec
	Status ResourceStatus
}

// Reconciler는 리소스의 원하는 상태로 실제 상태를 맞추는 인터페이스입니다.
// Kubernetes의 reconcile.Reconciler와 동일한 역할입니다.
type Reconciler interface {
	// Reconcile은 key에 해당하는 리소스를 조정합니다.
	// 에러 반환 시 컨트롤러가 재시도합니다.
	Reconcile(ctx context.Context, key string) error
}

// ============================================================
// WorkQueue 구현
// ============================================================

// workQueue는 중복 제거와 지수 백오프 재시도를 지원하는 큐입니다.
// client-go의 RateLimitingQueue를 단순화한 버전입니다.
type workQueue struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - queue: 처리 대기 중인 키 슬라이스
	//   - processing: 현재 처리 중인 키 집합
	//   - dirty: 재추가 대기 중인 키 집합 (처리 중에 들어온 중복)
	//   - retries: 키별 재시도 횟수
	//   - shutdown: 종료 여부
	//   - mu: 뮤텍스
	//   - cond: 조건 변수 (Get 블로킹용)
	//   - baseDelay: 기본 백오프 딜레이 (기본값 5ms)
}

// NewWorkQueue는 새 workQueue를 생성합니다.
func NewWorkQueue() *workQueue {
	// TODO: 구현하세요
	panic("NewWorkQueue: 아직 구현되지 않았습니다")
}

// Add는 키를 큐에 추가합니다.
// 이미 처리 중인 키라면 dirty 집합에 표시해 Done 후 재추가되도록 합니다.
// 이미 큐에 있는 키는 중복 추가하지 않습니다.
func (q *workQueue) Add(key string) {
	// TODO: 구현하세요
	panic("Add: 아직 구현되지 않았습니다")
}

// Get은 처리 가능한 키를 블로킹으로 반환합니다.
// shutdown이면 ("", true)를 반환합니다.
func (q *workQueue) Get() (string, bool) {
	// TODO: 구현하세요
	panic("Get: 아직 구현되지 않았습니다")
}

// Done은 키 처리가 완료됐음을 알립니다.
// dirty에 있는 키라면 큐에 재추가합니다.
func (q *workQueue) Done(key string) {
	// TODO: 구현하세요
	panic("Done: 아직 구현되지 않았습니다")
}

// Forget은 키의 재시도 횟수를 초기화합니다.
// Reconcile 성공 시 호출합니다.
func (q *workQueue) Forget(key string) {
	// TODO: 구현하세요
	panic("Forget: 아직 구현되지 않았습니다")
}

// AddRateLimited는 지수 백오프 딜레이 후 키를 추가합니다.
// 딜레이 = baseDelay * 2^retryCount (최대 1초)
// Reconcile 실패 시 호출합니다.
func (q *workQueue) AddRateLimited(key string) {
	// TODO: 구현하세요
	panic("AddRateLimited: 아직 구현되지 않았습니다")
}

// Len은 현재 큐에 대기 중인 항목 수를 반환합니다.
func (q *workQueue) Len() int {
	// TODO: 구현하세요
	panic("Len: 아직 구현되지 않았습니다")
}

// ShutDown은 큐를 종료합니다. 블로킹 중인 Get을 깨웁니다.
func (q *workQueue) ShutDown() {
	// TODO: 구현하세요
	panic("ShutDown: 아직 구현되지 않았습니다")
}

// ============================================================
// Controller 구현
// ============================================================

// Controller는 이벤트를 감시하고 Reconcile 루프를 실행합니다.
// Kubernetes의 controller-runtime Controller를 단순화한 버전입니다.
type Controller struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - name: 컨트롤러 이름
	//   - reconciler: Reconciler 인터페이스
	//   - queue: *workQueue
	//   - maxRetries: 최대 재시도 횟수 (기본값 5)
}

// NewController는 새 Controller를 생성합니다.
func NewController(name string, r Reconciler) *Controller {
	// TODO: 구현하세요
	panic("NewController: 아직 구현되지 않았습니다")
}

// Watch는 이벤트 채널을 감시해 키를 큐에 추가합니다.
// ctx 취소 또는 eventCh 종료 시 반환합니다.
// goroutine으로 실행됩니다.
func (c *Controller) Watch(ctx context.Context, eventCh <-chan Event) {
	// TODO: 구현하세요
	// 힌트: Event.Resource.Name을 키로 사용
	panic("Watch: 아직 구현되지 않았습니다")
}

// Run은 workers개의 goroutine으로 Reconcile 루프를 실행합니다.
// ctx 취소 시 큐를 shutdown하고 모든 worker가 끝날 때까지 기다립니다.
func (c *Controller) Run(ctx context.Context, workers int) {
	// TODO: 구현하세요
	panic("Run: 아직 구현되지 않았습니다")
}

// processNextItem은 큐에서 키를 하나 꺼내 Reconcile합니다.
// 큐가 shutdown되면 false를 반환합니다.
func (c *Controller) processNextItem(ctx context.Context) bool {
	// TODO: 구현하세요
	// 힌트:
	//   1. queue.Get()으로 키 획득
	//   2. defer queue.Done(key)
	//   3. reconciler.Reconcile(ctx, key) 호출
	//   4. 성공 → queue.Forget(key)
	//   5. 실패 + retries < maxRetries → queue.AddRateLimited(key)
	//   6. 실패 + retries >= maxRetries → queue.Forget(key) (포기)
	panic("processNextItem: 아직 구현되지 않았습니다")
}

// retryCount는 키의 현재 재시도 횟수를 반환합니다.
func (c *Controller) retryCount(key string) int {
	// TODO: 구현하세요
	panic("retryCount: 아직 구현되지 않았습니다")
}

// ============================================================
// 지수 백오프 헬퍼
// ============================================================

// exponentialBackoff는 retry번째 재시도의 딜레이를 계산합니다.
// 딜레이 = base * 2^retry, 최대 maxDelay
func exponentialBackoff(base time.Duration, retry int, maxDelay time.Duration) time.Duration {
	// TODO: 구현하세요
	panic("exponentialBackoff: 아직 구현되지 않았습니다")
}

func main() {
	// 이 파일은 라이브러리로 사용됩니다.
	// 테스트를 실행하세요: go test ./... -v
}
