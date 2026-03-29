// operator_test.go
// 쿠버네티스 오퍼레이터 시뮬레이터 테스트 및 채점
//
// 실행:
//
//	go test -v
//	go test -v -run TestGrade
package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ============================================================
// 테스트 헬퍼
// ============================================================

func makeDatabase(name string, replicas int, version string) *Database {
	return &Database{
		Name: name,
		Spec: DatabaseSpec{
			Replicas: replicas,
			Version:  version,
			Storage:  10,
		},
		Finalizers: []string{"database.cleanup"},
	}
}

// waitFor는 조건이 참이 될 때까지 최대 timeout 동안 대기합니다.
func waitFor(t *testing.T, desc string, timeout time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("타임아웃: %s (%v)", desc, timeout)
	return false
}

// ============================================================
// LeaderElection 테스트
// ============================================================

func TestLeaderElection_Acquire(t *testing.T) {
	le := NewLeaderElection()

	if !le.TryAcquire() {
		t.Error("첫 번째 TryAcquire는 true여야 합니다")
	}
	if !le.IsLeader() {
		t.Error("TryAcquire 후 IsLeader()가 true여야 합니다")
	}
}

func TestLeaderElection_AlreadyAcquired(t *testing.T) {
	le := NewLeaderElection()
	le.TryAcquire()

	// 이미 리더이므로 두 번 호출해도 true (자신이 이미 획득)
	if !le.TryAcquire() {
		t.Error("이미 리더인 경우 TryAcquire는 true여야 합니다")
	}
}

func TestLeaderElection_Release(t *testing.T) {
	le := NewLeaderElection()
	le.TryAcquire()
	le.Release()

	if le.IsLeader() {
		t.Error("Release 후 IsLeader()가 false여야 합니다")
	}
}

// ============================================================
// InstanceManager 테스트
// ============================================================

func TestInstanceManager_CreateDelete(t *testing.T) {
	m := NewInstanceManager()

	id := m.CreateInstance("db-1", "5.7")
	if id == "" {
		t.Error("CreateInstance가 빈 ID를 반환했습니다")
	}

	instances := m.GetInstancesByDB("db-1")
	if len(instances) != 1 {
		t.Errorf("GetInstancesByDB = %d개, 원하는 값: 1", len(instances))
	}
	if instances[0].Phase != "Running" {
		t.Errorf("새 인스턴스 Phase = %s, 원하는 값: Running", instances[0].Phase)
	}

	if err := m.DeleteInstance(id); err != nil {
		t.Errorf("DeleteInstance 오류: %v", err)
	}
	if len(m.GetInstancesByDB("db-1")) != 0 {
		t.Error("DeleteInstance 후 인스턴스가 남아있습니다")
	}
}

func TestInstanceManager_MultipleInstances(t *testing.T) {
	m := NewInstanceManager()
	for i := 0; i < 3; i++ {
		m.CreateInstance("db-multi", "8.0")
	}

	instances := m.GetInstancesByDB("db-multi")
	if len(instances) != 3 {
		t.Errorf("3개 생성 후 GetInstancesByDB = %d개", len(instances))
	}
}

func TestInstanceManager_Upgrade(t *testing.T) {
	m := NewInstanceManager()
	id := m.CreateInstance("db-upg", "5.7")

	if err := m.UpgradeInstance(id, "8.0"); err != nil {
		t.Fatalf("UpgradeInstance 오류: %v", err)
	}

	instances := m.GetInstancesByDB("db-upg")
	if instances[0].Version != "8.0" {
		t.Errorf("업그레이드 후 Version = %s, 원하는 값: 8.0", instances[0].Version)
	}
}

// ============================================================
// Operator: Create → Reconcile 테스트 (15점)
// ============================================================

func TestOperator_CreateReconcile(t *testing.T) {
	op := NewDatabaseOperator()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)
	go op.Run(ctx, 1)

	db := makeDatabase("testdb", 2, "8.0")
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

	waitFor(t, "인스턴스 2개 생성", 2*time.Second, func() bool {
		instances := op.instanceMgr.GetInstancesByDB("testdb")
		return len(instances) == 2
	})

	instances := op.instanceMgr.GetInstancesByDB("testdb")
	if len(instances) != 2 {
		t.Errorf("Reconcile 후 인스턴스 = %d개, 원하는 값: 2", len(instances))
	}
}

// ============================================================
// Operator: 스케일 업 테스트 (10점)
// ============================================================

func TestOperator_ScaleUp(t *testing.T) {
	op := NewDatabaseOperator()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)
	go op.Run(ctx, 1)

	db := makeDatabase("scaledb", 1, "8.0")
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

	waitFor(t, "초기 1개 생성", time.Second, func() bool {
		return len(op.instanceMgr.GetInstancesByDB("scaledb")) == 1
	})

	// 스케일 업: 1 → 3
	db.Spec.Replicas = 3
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventUpdate, Database: db}

	waitFor(t, "스케일 업 3개", 2*time.Second, func() bool {
		return len(op.instanceMgr.GetInstancesByDB("scaledb")) == 3
	})

	if len(op.instanceMgr.GetInstancesByDB("scaledb")) != 3 {
		t.Errorf("스케일 업 후 인스턴스 = %d개, 원하는 값: 3",
			len(op.instanceMgr.GetInstancesByDB("scaledb")))
	}
}

// ============================================================
// Operator: 스케일 다운 테스트 (10점)
// ============================================================

func TestOperator_ScaleDown(t *testing.T) {
	op := NewDatabaseOperator()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)
	go op.Run(ctx, 1)

	db := makeDatabase("downdb", 3, "8.0")
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

	waitFor(t, "초기 3개 생성", time.Second, func() bool {
		return len(op.instanceMgr.GetInstancesByDB("downdb")) == 3
	})

	// 스케일 다운: 3 → 1
	db.Spec.Replicas = 1
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventUpdate, Database: db}

	waitFor(t, "스케일 다운 1개", 2*time.Second, func() bool {
		return len(op.instanceMgr.GetInstancesByDB("downdb")) == 1
	})

	if len(op.instanceMgr.GetInstancesByDB("downdb")) != 1 {
		t.Errorf("스케일 다운 후 인스턴스 = %d개, 원하는 값: 1",
			len(op.instanceMgr.GetInstancesByDB("downdb")))
	}
}

// ============================================================
// Operator: 버전 업그레이드 테스트 (15점)
// ============================================================

func TestOperator_VersionUpgrade(t *testing.T) {
	op := NewDatabaseOperator()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)
	go op.Run(ctx, 1)

	db := makeDatabase("upgradedb", 2, "5.7")
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

	waitFor(t, "초기 인스턴스 생성", time.Second, func() bool {
		return len(op.instanceMgr.GetInstancesByDB("upgradedb")) == 2
	})

	// 버전 업그레이드: 5.7 → 8.0
	db.Spec.Version = "8.0"
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventUpdate, Database: db}

	waitFor(t, "버전 업그레이드 완료", 3*time.Second, func() bool {
		instances := op.instanceMgr.GetInstancesByDB("upgradedb")
		if len(instances) == 0 {
			return false
		}
		for _, inst := range instances {
			if inst.Version != "8.0" {
				return false
			}
		}
		return true
	})

	instances := op.instanceMgr.GetInstancesByDB("upgradedb")
	for _, inst := range instances {
		if inst.Version != "8.0" {
			t.Errorf("업그레이드 후 인스턴스 %s의 버전 = %s, 원하는 값: 8.0",
				inst.ID, inst.Version)
		}
	}
}

// ============================================================
// Operator: Finalizer 테스트 (15점)
// ============================================================

func TestOperator_Finalizer(t *testing.T) {
	op := NewDatabaseOperator()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)
	go op.Run(ctx, 1)

	db := makeDatabase("findb", 2, "8.0")
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

	waitFor(t, "초기 인스턴스 생성", time.Second, func() bool {
		return len(op.instanceMgr.GetInstancesByDB("findb")) == 2
	})

	// 삭제 요청
	db.DeletionRequested = true
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventDelete, Database: db}

	// 인스턴스가 모두 삭제되고 DB가 Store에서 제거되어야 함
	waitFor(t, "Finalizer 처리 완료", 2*time.Second, func() bool {
		_, exists := op.store.Get("findb")
		return !exists
	})

	if len(op.instanceMgr.GetInstancesByDB("findb")) != 0 {
		t.Errorf("Finalizer 처리 후 인스턴스가 %d개 남아있습니다",
			len(op.instanceMgr.GetInstancesByDB("findb")))
	}
}

// ============================================================
// Operator: Status Conditions 테스트 (15점)
// ============================================================

func TestOperator_StatusConditions(t *testing.T) {
	op := NewDatabaseOperator()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)
	go op.Run(ctx, 1)

	db := makeDatabase("conddb", 2, "8.0")
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

	// 안정 상태가 될 때까지 대기
	waitFor(t, "Ready 상태", 2*time.Second, func() bool {
		current, ok := op.store.Get("conddb")
		if !ok {
			return false
		}
		cond, found := getCondition(current, ConditionReady)
		return found && cond.Status
	})

	current, ok := op.store.Get("conddb")
	if !ok {
		t.Fatal("conddb를 Store에서 찾을 수 없습니다")
	}

	readyCond, found := getCondition(current, ConditionReady)
	if !found {
		t.Error("Ready Condition이 없습니다")
	} else if !readyCond.Status {
		t.Error("모든 인스턴스가 Running인데 Ready=false입니다")
	}

	if current.Status.ReadyReplicas != 2 {
		t.Errorf("ReadyReplicas = %d, 원하는 값: 2", current.Status.ReadyReplicas)
	}
}

// ============================================================
// Operator: 리더 선출 테스트 (10점)
// ============================================================

func TestOperator_LeaderElection(t *testing.T) {
	op := NewDatabaseOperator()

	// 리더 획득
	if !op.leaderElection.TryAcquire() {
		t.Error("TryAcquire 실패")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)
	go op.Run(ctx, 1)

	db := makeDatabase("leaderdb", 1, "8.0")
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

	waitFor(t, "리더로 Reconcile", time.Second, func() bool {
		return len(op.instanceMgr.GetInstancesByDB("leaderdb")) == 1
	})

	if len(op.instanceMgr.GetInstancesByDB("leaderdb")) != 1 {
		t.Error("리더인데 Reconcile이 실행되지 않았습니다")
	}
}

func TestOperator_NotLeaderSkipsReconcile(t *testing.T) {
	op := NewDatabaseOperator()
	// 리더 획득 안 함

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)
	go op.Run(ctx, 1)

	db := makeDatabase("notleaderdb", 1, "8.0")
	op.store.Set(db)
	eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

	time.Sleep(300 * time.Millisecond)

	// 리더가 아니므로 인스턴스가 생성되지 않아야 함
	if len(op.instanceMgr.GetInstancesByDB("notleaderdb")) != 0 {
		t.Error("리더가 아닌데 Reconcile이 실행되었습니다")
	}
}

// ============================================================
// Operator: 에러 복구 + ctx 취소 테스트 (10점)
// ============================================================

func TestOperator_ContextCancel(t *testing.T) {
	op := NewDatabaseOperator()
	ctx, cancel := context.WithCancel(context.Background())

	eventCh := make(chan DatabaseEvent, 10)
	go op.Watch(ctx, eventCh)

	done := make(chan struct{})
	go func() {
		op.Run(ctx, 2)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("ctx 취소 후 Run이 1초 내에 반환되지 않음")
	}
}

// ============================================================
// 채점 함수 (TestGrade)
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════════════")
	fmt.Println("  과제 A6: 오퍼레이터 시뮬레이터 채점 결과")
	fmt.Println("  패턴: Kubernetes Operator (캡스톤 - 모든 패턴 통합)")
	fmt.Println("═══════════════════════════════════════════════════")

	// Create → Reconcile (15점)
	t.Run("Create_Reconcile", func(t *testing.T) {
		op := NewDatabaseOperator()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		eventCh := make(chan DatabaseEvent, 10)
		go op.Watch(ctx, eventCh)
		go op.Run(ctx, 1)

		db := makeDatabase("g-create", 2, "8.0")
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

		ok := waitFor(t, "인스턴스 2개", 2*time.Second, func() bool {
			return len(op.instanceMgr.GetInstancesByDB("g-create")) == 2
		})
		if ok {
			score += 15
			fmt.Printf("  ✓ Create → Reconcile → 인스턴스 생성  15/15점\n")
		} else {
			n := len(op.instanceMgr.GetInstancesByDB("g-create"))
			fmt.Printf("  ✗ Create → Reconcile (인스턴스=%d)      0/15점\n", n)
		}
	})

	// 스케일 업 (10점)
	t.Run("스케일_업", func(t *testing.T) {
		op := NewDatabaseOperator()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		eventCh := make(chan DatabaseEvent, 10)
		go op.Watch(ctx, eventCh)
		go op.Run(ctx, 1)

		db := makeDatabase("g-scale", 1, "8.0")
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventCreate, Database: db}
		waitFor(t, "초기", time.Second, func() bool {
			return len(op.instanceMgr.GetInstancesByDB("g-scale")) == 1
		})

		db.Spec.Replicas = 3
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventUpdate, Database: db}

		ok := waitFor(t, "스케일 업 3개", 2*time.Second, func() bool {
			return len(op.instanceMgr.GetInstancesByDB("g-scale")) == 3
		})
		if ok {
			score += 10
			fmt.Printf("  ✓ 스케일 업 (1→3)                     10/10점\n")
		} else {
			n := len(op.instanceMgr.GetInstancesByDB("g-scale"))
			fmt.Printf("  ✗ 스케일 업 (인스턴스=%d, 원하는:3)     0/10점\n", n)
		}
	})

	// 스케일 다운 (10점)
	t.Run("스케일_다운", func(t *testing.T) {
		op := NewDatabaseOperator()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		eventCh := make(chan DatabaseEvent, 10)
		go op.Watch(ctx, eventCh)
		go op.Run(ctx, 1)

		db := makeDatabase("g-down", 3, "8.0")
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventCreate, Database: db}
		waitFor(t, "초기", time.Second, func() bool {
			return len(op.instanceMgr.GetInstancesByDB("g-down")) == 3
		})

		db.Spec.Replicas = 1
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventUpdate, Database: db}

		ok := waitFor(t, "스케일 다운 1개", 2*time.Second, func() bool {
			return len(op.instanceMgr.GetInstancesByDB("g-down")) == 1
		})
		if ok {
			score += 10
			fmt.Printf("  ✓ 스케일 다운 (3→1)                   10/10점\n")
		} else {
			n := len(op.instanceMgr.GetInstancesByDB("g-down"))
			fmt.Printf("  ✗ 스케일 다운 (인스턴스=%d, 원하는:1)   0/10점\n", n)
		}
	})

	// 버전 업그레이드 (15점)
	t.Run("버전_업그레이드", func(t *testing.T) {
		op := NewDatabaseOperator()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		eventCh := make(chan DatabaseEvent, 10)
		go op.Watch(ctx, eventCh)
		go op.Run(ctx, 1)

		db := makeDatabase("g-upg", 2, "5.7")
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventCreate, Database: db}
		waitFor(t, "초기", time.Second, func() bool {
			return len(op.instanceMgr.GetInstancesByDB("g-upg")) == 2
		})

		db.Spec.Version = "8.0"
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventUpdate, Database: db}

		ok := waitFor(t, "업그레이드", 3*time.Second, func() bool {
			for _, inst := range op.instanceMgr.GetInstancesByDB("g-upg") {
				if inst.Version != "8.0" {
					return false
				}
			}
			return len(op.instanceMgr.GetInstancesByDB("g-upg")) > 0
		})
		if ok {
			score += 15
			fmt.Printf("  ✓ 버전 업그레이드 (5.7→8.0)           15/15점\n")
		} else {
			fmt.Printf("  ✗ 버전 업그레이드 미완료               0/15점\n")
		}
	})

	// Finalizer (15점)
	t.Run("Finalizer", func(t *testing.T) {
		op := NewDatabaseOperator()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		eventCh := make(chan DatabaseEvent, 10)
		go op.Watch(ctx, eventCh)
		go op.Run(ctx, 1)

		db := makeDatabase("g-fin", 2, "8.0")
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventCreate, Database: db}
		waitFor(t, "초기", time.Second, func() bool {
			return len(op.instanceMgr.GetInstancesByDB("g-fin")) == 2
		})

		db.DeletionRequested = true
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventDelete, Database: db}

		ok := waitFor(t, "Finalizer 처리", 2*time.Second, func() bool {
			_, exists := op.store.Get("g-fin")
			return !exists
		})
		if ok && len(op.instanceMgr.GetInstancesByDB("g-fin")) == 0 {
			score += 15
			fmt.Printf("  ✓ Finalizer 패턴 (삭제 정리)          15/15점\n")
		} else {
			_, exists := op.store.Get("g-fin")
			fmt.Printf("  ✗ Finalizer (db_exists=%v, instances=%d)  0/15점\n",
				exists, len(op.instanceMgr.GetInstancesByDB("g-fin")))
		}
	})

	// Status Conditions (15점)
	t.Run("Status_Conditions", func(t *testing.T) {
		op := NewDatabaseOperator()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		eventCh := make(chan DatabaseEvent, 10)
		go op.Watch(ctx, eventCh)
		go op.Run(ctx, 1)

		db := makeDatabase("g-cond", 2, "8.0")
		op.store.Set(db)
		eventCh <- DatabaseEvent{Type: EventCreate, Database: db}

		ok := waitFor(t, "Ready 조건", 2*time.Second, func() bool {
			current, exists := op.store.Get("g-cond")
			if !exists {
				return false
			}
			cond, found := getCondition(current, ConditionReady)
			return found && cond.Status && current.Status.ReadyReplicas == 2
		})
		if ok {
			score += 15
			fmt.Printf("  ✓ Status Conditions (Ready/ReadyReplicas) 15/15점\n")
		} else {
			current, exists := op.store.Get("g-cond")
			if exists {
				fmt.Printf("  ✗ Status Conditions (ready=%d)          0/15점\n",
					current.Status.ReadyReplicas)
			} else {
				fmt.Printf("  ✗ Status Conditions (DB 없음)           0/15점\n")
			}
		}
	})

	// 리더 선출 (10점) - 별도 테스트에서 확인
	t.Run("리더_선출", func(t *testing.T) {
		le := NewLeaderElection()
		acquired := le.TryAcquire()
		isLeader := le.IsLeader()
		le.Release()
		notLeader := !le.IsLeader()

		if acquired && isLeader && notLeader {
			score += 10
			fmt.Printf("  ✓ 리더 선출 (Acquire/IsLeader/Release)10/10점\n")
		} else {
			fmt.Printf("  ✗ 리더 선출 (acq=%v,leader=%v,notLeader=%v)  0/10점\n",
				acquired, isLeader, notLeader)
		}
	})

	fmt.Println("───────────────────────────────────────────────────")
	fmt.Printf("  최종 점수: %d / %d점\n", score, total)

	grade := "F"
	switch {
	case score >= 95:
		grade = "A+"
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	}
	fmt.Printf("  등급: %s\n", grade)
	fmt.Print("═══════════════════════════════════════════════════\n\n")
}
