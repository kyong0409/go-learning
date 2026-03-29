// solution/operator.go
// 오퍼레이터 시뮬레이터 참고 풀이 (~420줄)
//
// 이 파일은 참고용입니다. 먼저 직접 구현해보세요.
package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// ============================================================
// 타입 정의
// ============================================================

type EventType string

const (
	EventCreate EventType = "CREATE"
	EventUpdate EventType = "UPDATE"
	EventDelete EventType = "DELETE"
)

type DatabaseEvent struct {
	Type     EventType
	Database *Database
}

type DatabaseSpec struct {
	Replicas int
	Version  string
	Storage  int
}

type ConditionType string

const (
	ConditionReady       ConditionType = "Ready"
	ConditionProgressing ConditionType = "Progressing"
	ConditionDegraded    ConditionType = "Degraded"
)

type Condition struct {
	Type    ConditionType
	Status  bool
	Reason  string
	Message string
}

type DatabaseStatus struct {
	ReadyReplicas int
	Phase         string
	Conditions    []Condition
	Instances     []string
}

type Database struct {
	Name              string
	Spec              DatabaseSpec
	Status            DatabaseStatus
	Finalizers        []string
	DeletionRequested bool
}

type Instance struct {
	ID      string
	DBName  string
	Version string
	Phase   string
}

// ============================================================
// LeaderElection
// ============================================================

type LeaderElection struct {
	mu     sync.Mutex
	leader bool
}

func NewLeaderElection() *LeaderElection {
	return &LeaderElection{}
}

func (le *LeaderElection) TryAcquire() bool {
	le.mu.Lock()
	defer le.mu.Unlock()
	le.leader = true
	return true
}

func (le *LeaderElection) IsLeader() bool {
	le.mu.Lock()
	defer le.mu.Unlock()
	return le.leader
}

func (le *LeaderElection) Release() {
	le.mu.Lock()
	defer le.mu.Unlock()
	le.leader = false
}

// ============================================================
// InstanceManager
// ============================================================

type InstanceManager struct {
	mu        sync.RWMutex
	instances map[string]*Instance
	nextID    int
}

func NewInstanceManager() *InstanceManager {
	return &InstanceManager{instances: make(map[string]*Instance)}
}

func (m *InstanceManager) CreateInstance(dbName, version string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	id := fmt.Sprintf("%s-%d", dbName, m.nextID)
	m.instances[id] = &Instance{
		ID:      id,
		DBName:  dbName,
		Version: version,
		Phase:   "Running",
	}
	return id
}

func (m *InstanceManager) DeleteInstance(instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.instances[instanceID]; !ok {
		return fmt.Errorf("인스턴스 %q를 찾을 수 없습니다", instanceID)
	}
	delete(m.instances, instanceID)
	return nil
}

func (m *InstanceManager) GetInstancesByDB(dbName string) []*Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*Instance
	for _, inst := range m.instances {
		if inst.DBName == dbName {
			result = append(result, inst)
		}
	}
	return result
}

func (m *InstanceManager) UpgradeInstance(instanceID, newVersion string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	inst, ok := m.instances[instanceID]
	if !ok {
		return fmt.Errorf("인스턴스 %q를 찾을 수 없습니다", instanceID)
	}
	inst.Phase = "Stopping"
	inst.Version = newVersion
	inst.Phase = "Running"
	return nil
}

// ============================================================
// DatabaseStore
// ============================================================

type DatabaseStore struct {
	mu  sync.RWMutex
	dbs map[string]*Database
}

func NewDatabaseStore() *DatabaseStore {
	return &DatabaseStore{dbs: make(map[string]*Database)}
}

func (ds *DatabaseStore) Get(name string) (*Database, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	db, ok := ds.dbs[name]
	return db, ok
}

func (ds *DatabaseStore) Set(db *Database) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.dbs[db.Name] = db
}

func (ds *DatabaseStore) Delete(name string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.dbs, name)
}

// ============================================================
// 내부 WorkQueue (A1 패턴 재사용)
// ============================================================

type workQueue struct {
	mu         sync.Mutex
	cond       *sync.Cond
	queue      []string
	queued     map[string]bool
	processing map[string]bool
	dirty      map[string]bool
	retries    map[string]int
	shutdown   bool
	baseDelay  time.Duration
}

func newWorkQueue() *workQueue {
	q := &workQueue{
		queued:     make(map[string]bool),
		processing: make(map[string]bool),
		dirty:      make(map[string]bool),
		retries:    make(map[string]int),
		baseDelay:  5 * time.Millisecond,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *workQueue) add(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.shutdown {
		return
	}
	if q.processing[key] {
		q.dirty[key] = true
		return
	}
	if q.queued[key] {
		return
	}
	q.queue = append(q.queue, key)
	q.queued[key] = true
	q.cond.Signal()
}

func (q *workQueue) get() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.queue) == 0 && !q.shutdown {
		q.cond.Wait()
	}
	if q.shutdown && len(q.queue) == 0 {
		return "", true
	}
	key := q.queue[0]
	q.queue = q.queue[1:]
	delete(q.queued, key)
	q.processing[key] = true
	return key, false
}

func (q *workQueue) done(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.processing, key)
	if q.dirty[key] {
		delete(q.dirty, key)
		q.queue = append(q.queue, key)
		q.queued[key] = true
		q.cond.Signal()
	}
}

func (q *workQueue) forget(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.retries, key)
}

func (q *workQueue) addRateLimited(key string) {
	q.mu.Lock()
	retries := q.retries[key]
	q.retries[key] = retries + 1
	delay := exponentialBackoff(q.baseDelay, retries, time.Second)
	q.mu.Unlock()
	go func() {
		time.Sleep(delay)
		q.add(key)
	}()
}

func (q *workQueue) retryCount(key string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.retries[key]
}

func (q *workQueue) shutDown() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.shutdown = true
	q.cond.Broadcast()
}

// ============================================================
// DatabaseOperator
// ============================================================

type DatabaseOperator struct {
	store          *DatabaseStore
	instanceMgr    *InstanceManager
	leaderElection *LeaderElection
	queue          *workQueue
	maxRetries     int
}

func NewDatabaseOperator() *DatabaseOperator {
	op := &DatabaseOperator{
		store:          NewDatabaseStore(),
		instanceMgr:    NewInstanceManager(),
		leaderElection: NewLeaderElection(),
		queue:          newWorkQueue(),
		maxRetries:     5,
	}
	// 오퍼레이터는 기본적으로 리더를 획득하고 시작합니다.
	op.leaderElection.TryAcquire()
	return op
}

func (op *DatabaseOperator) Watch(ctx context.Context, eventCh <-chan DatabaseEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-eventCh:
			if !ok {
				return
			}
			op.queue.add(ev.Database.Name)
		}
	}
}

func (op *DatabaseOperator) Run(ctx context.Context, workers int) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for op.processNextItem(ctx) {
			}
		}()
	}
	<-ctx.Done()
	op.queue.shutDown()
	wg.Wait()
}

func (op *DatabaseOperator) processNextItem(ctx context.Context) bool {
	key, shutdown := op.queue.get()
	if shutdown {
		return false
	}
	defer op.queue.done(key)

	// 리더가 아니면 처리하지 않음
	if !op.leaderElection.IsLeader() {
		return true
	}

	err := op.Reconcile(ctx, key)
	if err == nil {
		op.queue.forget(key)
		return true
	}

	if op.queue.retryCount(key) < op.maxRetries {
		op.queue.addRateLimited(key)
	} else {
		op.queue.forget(key)
	}
	return true
}

func (op *DatabaseOperator) Reconcile(ctx context.Context, name string) error {
	db, ok := op.store.Get(name)
	if !ok {
		return nil
	}

	// 삭제 처리
	if db.DeletionRequested {
		return op.handleDeletion(ctx, db)
	}

	// 인스턴스 수 조정
	if err := op.reconcileInstances(ctx, db); err != nil {
		return err
	}

	// 버전 업그레이드
	if err := op.reconcileVersion(ctx, db); err != nil {
		return err
	}

	// Status 업데이트
	op.updateStatus(db)
	op.store.Set(db)
	return nil
}

func (op *DatabaseOperator) handleDeletion(_ context.Context, db *Database) error {
	if !hasFinalizer(db, "database.cleanup") {
		op.store.Delete(db.Name)
		return nil
	}

	// 모든 인스턴스 삭제
	instances := op.instanceMgr.GetInstancesByDB(db.Name)
	for _, inst := range instances {
		op.instanceMgr.DeleteInstance(inst.ID)
	}

	// Finalizer 제거 후 저장
	removeFinalizer(db, "database.cleanup")
	if len(db.Finalizers) == 0 {
		op.store.Delete(db.Name)
	} else {
		op.store.Set(db)
	}
	return nil
}

func (op *DatabaseOperator) reconcileInstances(_ context.Context, db *Database) error {
	instances := op.instanceMgr.GetInstancesByDB(db.Name)
	current := len(instances)
	desired := db.Spec.Replicas

	// 스케일 업
	for current < desired {
		op.instanceMgr.CreateInstance(db.Name, db.Spec.Version)
		current++
	}

	// 스케일 다운
	for current > desired {
		last := instances[current-1]
		op.instanceMgr.DeleteInstance(last.ID)
		instances = instances[:current-1]
		current--
	}
	return nil
}

func (op *DatabaseOperator) reconcileVersion(_ context.Context, db *Database) error {
	instances := op.instanceMgr.GetInstancesByDB(db.Name)
	for _, inst := range instances {
		if inst.Version != db.Spec.Version {
			// 롤링 업그레이드: 한 번에 하나
			if err := op.instanceMgr.UpgradeInstance(inst.ID, db.Spec.Version); err != nil {
				return err
			}
			// 실제 롤링 업그레이드는 완료를 확인해야 하지만, 시뮬레이션에서는 즉시 완료
		}
	}
	return nil
}

func (op *DatabaseOperator) updateStatus(db *Database) {
	instances := op.instanceMgr.GetInstancesByDB(db.Name)

	runningCount := 0
	var instanceIDs []string
	upgradingCount := 0

	for _, inst := range instances {
		instanceIDs = append(instanceIDs, inst.ID)
		switch inst.Phase {
		case "Running":
			runningCount++
		case "Stopping", "Starting":
			upgradingCount++
		}
	}

	db.Status.ReadyReplicas = runningCount
	db.Status.Instances = instanceIDs

	// Phase 결정
	if runningCount == db.Spec.Replicas {
		db.Status.Phase = "Running"
	} else if upgradingCount > 0 {
		db.Status.Phase = "Upgrading"
	} else if runningCount < db.Spec.Replicas {
		db.Status.Phase = "Pending"
	}

	// Conditions 설정
	setCondition(db, ConditionReady,
		runningCount == db.Spec.Replicas,
		"AllRunning", fmt.Sprintf("%d/%d 인스턴스 준비됨", runningCount, db.Spec.Replicas))

	setCondition(db, ConditionProgressing,
		runningCount != db.Spec.Replicas || upgradingCount > 0,
		"Reconciling", "상태 조정 중")

	setCondition(db, ConditionDegraded,
		false, "OK", "모든 인스턴스 정상")
}

// ============================================================
// 헬퍼 함수
// ============================================================

func setCondition(db *Database, condType ConditionType, status bool, reason, message string) {
	for i, c := range db.Status.Conditions {
		if c.Type == condType {
			db.Status.Conditions[i] = Condition{Type: condType, Status: status, Reason: reason, Message: message}
			return
		}
	}
	db.Status.Conditions = append(db.Status.Conditions, Condition{
		Type: condType, Status: status, Reason: reason, Message: message,
	})
}

func getCondition(db *Database, condType ConditionType) (Condition, bool) {
	for _, c := range db.Status.Conditions {
		if c.Type == condType {
			return c, true
		}
	}
	return Condition{}, false
}

func hasFinalizer(db *Database, finalizer string) bool {
	for _, f := range db.Finalizers {
		if f == finalizer {
			return true
		}
	}
	return false
}

func removeFinalizer(db *Database, finalizer string) {
	var newFinalizers []string
	for _, f := range db.Finalizers {
		if f != finalizer {
			newFinalizers = append(newFinalizers, f)
		}
	}
	db.Finalizers = newFinalizers
}

func exponentialBackoff(base time.Duration, retry int, maxDelay time.Duration) time.Duration {
	delay := time.Duration(float64(base) * math.Pow(2, float64(retry)))
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}

type ReconcileError struct {
	Name    string
	Message string
}

func (e *ReconcileError) Error() string {
	return fmt.Sprintf("reconcile 실패 [%s]: %s", e.Name, e.Message)
}

func main() {}
