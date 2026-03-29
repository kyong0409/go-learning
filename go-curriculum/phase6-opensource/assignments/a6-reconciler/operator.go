// operator.go
// 쿠버네티스 오퍼레이터 시뮬레이터 - 캡스톤 과제
//
// TODO: 아래 타입과 함수들을 완성하세요.
// Phase 6의 모든 패턴(Controller, WorkQueue, Watch, Informer, Lifecycle)을 통합합니다.
package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// EventType은 Database 리소스 변경 종류입니다.
type EventType string

const (
	EventCreate EventType = "CREATE"
	EventUpdate EventType = "UPDATE"
	EventDelete EventType = "DELETE"
)

// DatabaseEvent는 Database 리소스 변경 이벤트입니다.
type DatabaseEvent struct {
	Type     EventType
	Database *Database
}

// DatabaseSpec은 Database의 원하는 상태입니다.
type DatabaseSpec struct {
	Replicas int    // 원하는 인스턴스 수 (1~5)
	Version  string // DB 버전 (예: "5.7", "8.0")
	Storage  int    // 스토리지 크기(GB)
}

// ConditionType은 Database 상태 조건 종류입니다.
type ConditionType string

const (
	ConditionReady       ConditionType = "Ready"
	ConditionProgressing ConditionType = "Progressing"
	ConditionDegraded    ConditionType = "Degraded"
)

// Condition은 Database의 특정 상태 조건입니다.
type Condition struct {
	Type    ConditionType
	Status  bool
	Reason  string
	Message string
}

// DatabaseStatus는 Database의 현재 상태입니다.
type DatabaseStatus struct {
	ReadyReplicas int
	Phase         string      // "Pending", "Running", "Upgrading", "Failed"
	Conditions    []Condition
	Instances     []string    // 실행 중인 인스턴스 ID 목록
}

// Database는 오퍼레이터가 관리하는 커스텀 리소스입니다.
type Database struct {
	Name              string
	Spec              DatabaseSpec
	Status            DatabaseStatus
	Finalizers        []string
	DeletionRequested bool
}

// Instance는 실제 실행 중인 DB 인스턴스를 시뮬레이션합니다.
type Instance struct {
	ID      string
	DBName  string
	Version string
	Phase   string // "Starting", "Running", "Stopping", "Stopped"
}

// ============================================================
// LeaderElection 구현
// ============================================================

// LeaderElection은 단순화된 리더 선출입니다.
// 실제 K8s는 etcd lease를 사용하지만, 여기서는 인메모리 mutex를 사용합니다.
type LeaderElection struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - mu: sync.Mutex
	//   - leader: bool
}

// NewLeaderElection은 새 LeaderElection을 생성합니다.
func NewLeaderElection() *LeaderElection {
	// TODO: 구현하세요
	panic("NewLeaderElection: 아직 구현되지 않았습니다")
}

// TryAcquire는 리더 획득을 시도합니다.
// 이미 리더면 true, 아니면 false를 반환합니다.
func (le *LeaderElection) TryAcquire() bool {
	// TODO: 구현하세요
	panic("TryAcquire: 아직 구현되지 않았습니다")
}

// IsLeader는 현재 리더인지 반환합니다.
func (le *LeaderElection) IsLeader() bool {
	// TODO: 구현하세요
	panic("IsLeader: 아직 구현되지 않았습니다")
}

// Release는 리더를 반납합니다.
func (le *LeaderElection) Release() {
	// TODO: 구현하세요
	panic("Release: 아직 구현되지 않았습니다")
}

// ============================================================
// InstanceManager 구현
// ============================================================

// InstanceManager는 DB 인스턴스의 생명주기를 관리합니다.
type InstanceManager struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - mu: sync.RWMutex
	//   - instances: map[string]*Instance (instanceID → Instance)
	//   - nextID: int (자동 증가 ID)
}

// NewInstanceManager는 새 InstanceManager를 생성합니다.
func NewInstanceManager() *InstanceManager {
	// TODO: 구현하세요
	panic("NewInstanceManager: 아직 구현되지 않았습니다")
}

// CreateInstance는 새 인스턴스를 생성합니다.
// 반환값: 생성된 인스턴스 ID
func (m *InstanceManager) CreateInstance(dbName, version string) string {
	// TODO: 구현하세요
	// 힌트: "{dbName}-{nextID}" 형태의 ID 생성, Phase = "Running"
	panic("CreateInstance: 아직 구현되지 않았습니다")
}

// DeleteInstance는 인스턴스를 삭제합니다.
func (m *InstanceManager) DeleteInstance(instanceID string) error {
	// TODO: 구현하세요
	panic("DeleteInstance: 아직 구현되지 않았습니다")
}

// GetInstancesByDB는 특정 DB의 모든 인스턴스를 반환합니다.
func (m *InstanceManager) GetInstancesByDB(dbName string) []*Instance {
	// TODO: 구현하세요
	panic("GetInstancesByDB: 아직 구현되지 않았습니다")
}

// UpgradeInstance는 인스턴스의 버전을 업그레이드합니다.
// 업그레이드 중에는 Phase = "Stopping" → "Running" (시뮬레이션)
func (m *InstanceManager) UpgradeInstance(instanceID, newVersion string) error {
	// TODO: 구현하세요
	panic("UpgradeInstance: 아직 구현되지 않았습니다")
}

// ============================================================
// DatabaseStore 구현
// ============================================================

// DatabaseStore는 Database 리소스의 인메모리 저장소입니다.
type DatabaseStore struct {
	// TODO: 필요한 필드를 추가하세요.
	mu  sync.RWMutex
	dbs map[string]*Database
}

// NewDatabaseStore는 새 DatabaseStore를 생성합니다.
func NewDatabaseStore() *DatabaseStore {
	return &DatabaseStore{dbs: make(map[string]*Database)}
}

// Get은 이름으로 Database를 조회합니다.
func (ds *DatabaseStore) Get(name string) (*Database, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	db, ok := ds.dbs[name]
	return db, ok
}

// Set은 Database를 저장합니다.
func (ds *DatabaseStore) Set(db *Database) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.dbs[db.Name] = db
}

// Delete는 Database를 삭제합니다.
func (ds *DatabaseStore) Delete(name string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.dbs, name)
}

// ============================================================
// DatabaseOperator 구현
// ============================================================

// DatabaseOperator는 Database CR을 관리하는 오퍼레이터입니다.
type DatabaseOperator struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - store: *DatabaseStore
	//   - instanceMgr: *InstanceManager
	//   - leaderElection: *LeaderElection
	//   - queue: 작업 큐 (A1 패턴의 workQueue 또는 간단한 채널)
	//   - maxRetries: int
	//   - mu: sync.Mutex
	//   - retries: map[string]int
}

// NewDatabaseOperator는 새 DatabaseOperator를 생성합니다.
func NewDatabaseOperator() *DatabaseOperator {
	// TODO: 구현하세요
	panic("NewDatabaseOperator: 아직 구현되지 않았습니다")
}

// Watch는 이벤트 채널을 감시해 작업 큐에 추가합니다.
func (op *DatabaseOperator) Watch(ctx context.Context, eventCh <-chan DatabaseEvent) {
	// TODO: 구현하세요
	panic("Watch: 아직 구현되지 않았습니다")
}

// Run은 오퍼레이터의 Reconcile 루프를 시작합니다.
// 리더가 아니면 Reconcile하지 않습니다.
// ctx 취소 시 정상 종료합니다.
func (op *DatabaseOperator) Run(ctx context.Context, workers int) {
	// TODO: 구현하세요
	panic("Run: 아직 구현되지 않았습니다")
}

// Reconcile은 Database의 원하는 상태와 실제 상태를 비교해 조정합니다.
//
// 처리 순서:
//  1. Database 조회 (없으면 종료)
//  2. DeletionRequested 확인 → Finalizer 처리
//  3. 인스턴스 수 조정 (스케일 업/다운)
//  4. 버전 업그레이드 (필요 시)
//  5. Status 업데이트 (ReadyReplicas, Phase, Conditions)
func (op *DatabaseOperator) Reconcile(ctx context.Context, name string) error {
	// TODO: 구현하세요
	panic("Reconcile: 아직 구현되지 않았습니다")
}

// handleDeletion은 DeletionRequested=true인 Database를 처리합니다.
// 1. 모든 인스턴스 삭제
// 2. Finalizer 제거
// 3. Store에서 Database 삭제
func (op *DatabaseOperator) handleDeletion(ctx context.Context, db *Database) error {
	// TODO: 구현하세요
	panic("handleDeletion: 아직 구현되지 않았습니다")
}

// reconcileInstances는 인스턴스 수를 원하는 상태로 맞춥니다.
func (op *DatabaseOperator) reconcileInstances(ctx context.Context, db *Database) error {
	// TODO: 구현하세요
	// 힌트:
	//   - instanceMgr.GetInstancesByDB(db.Name)로 현재 인스턴스 조회
	//   - 부족하면 CreateInstance, 초과하면 DeleteInstance
	panic("reconcileInstances: 아직 구현되지 않았습니다")
}

// reconcileVersion은 인스턴스의 버전을 원하는 버전으로 롤링 업그레이드합니다.
func (op *DatabaseOperator) reconcileVersion(ctx context.Context, db *Database) error {
	// TODO: 구현하세요
	// 힌트: 버전이 다른 인스턴스를 하나씩 UpgradeInstance 호출
	panic("reconcileVersion: 아직 구현되지 않았습니다")
}

// updateStatus는 Database의 Status를 현재 인스턴스 상태로 업데이트합니다.
func (op *DatabaseOperator) updateStatus(db *Database) {
	// TODO: 구현하세요
	// 힌트:
	//   - GetInstancesByDB로 현재 인스턴스 수집
	//   - ReadyReplicas, Phase, Instances 업데이트
	//   - Conditions 설정 (Ready, Progressing, Degraded)
	panic("updateStatus: 아직 구현되지 않았습니다")
}

// setCondition은 특정 Condition을 설정합니다.
func setCondition(db *Database, condType ConditionType, status bool, reason, message string) {
	for i, c := range db.Status.Conditions {
		if c.Type == condType {
			db.Status.Conditions[i] = Condition{
				Type: condType, Status: status,
				Reason: reason, Message: message,
			}
			return
		}
	}
	db.Status.Conditions = append(db.Status.Conditions, Condition{
		Type: condType, Status: status,
		Reason: reason, Message: message,
	})
}

// getCondition은 특정 Condition을 반환합니다.
func getCondition(db *Database, condType ConditionType) (Condition, bool) {
	for _, c := range db.Status.Conditions {
		if c.Type == condType {
			return c, true
		}
	}
	return Condition{}, false
}

// hasFinalizer는 Database에 특정 finalizer가 있는지 확인합니다.
func hasFinalizer(db *Database, finalizer string) bool {
	for _, f := range db.Finalizers {
		if f == finalizer {
			return true
		}
	}
	return false
}

// removeFinalizer는 Database에서 특정 finalizer를 제거합니다.
func removeFinalizer(db *Database, finalizer string) {
	var newFinalizers []string
	for _, f := range db.Finalizers {
		if f != finalizer {
			newFinalizers = append(newFinalizers, f)
		}
	}
	db.Finalizers = newFinalizers
}

// ============================================================
// 지수 백오프 헬퍼
// ============================================================

func exponentialBackoff(base time.Duration, retry int, maxDelay time.Duration) time.Duration {
	if retry < 0 {
		retry = 0
	}
	delay := time.Duration(float64(base) * math.Pow(2, float64(retry)))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// ============================================================
// 에러 타입
// ============================================================

// ReconcileError는 Reconcile 실패 시 반환하는 에러입니다.
type ReconcileError struct {
	Name    string
	Message string
}

func (e *ReconcileError) Error() string {
	return fmt.Sprintf("reconcile 실패 [%s]: %s", e.Name, e.Message)
}

func main() {
	// 테스트를 실행하세요: go test ./... -v
}
