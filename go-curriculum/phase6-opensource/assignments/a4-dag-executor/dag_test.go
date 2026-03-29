// dag_test.go
// DAG 실행기 테스트 및 채점
//
// 실행:
//
//	go test -v
//	go test -v -run TestGrade
package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
// 테스트 헬퍼
// ============================================================

func makeNode(id string, delay time.Duration) *Node {
	return &Node{
		ID: id,
		Task: func(ctx context.Context) error {
			select {
			case <-time.After(delay):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
}

func makeFailNode(id string) *Node {
	return &Node{
		ID: id,
		Task: func(ctx context.Context) error {
			return fmt.Errorf("노드 %s 실패", id)
		},
	}
}

func makeRecordNode(id string, order *[]string, mu *sync.Mutex) *Node {
	return &Node{
		ID: id,
		Task: func(ctx context.Context) error {
			mu.Lock()
			*order = append(*order, id)
			mu.Unlock()
			return nil
		},
	}
}

// ============================================================
// DAG 기본 테스트 (10점)
// ============================================================

func TestDAG_AddNode(t *testing.T) {
	d := NewDAG()
	d.AddNode(&Node{ID: "a"})
	d.AddNode(&Node{ID: "b"})
	d.AddNode(&Node{ID: "c"})

	nodes := d.Nodes()
	if len(nodes) != 3 {
		t.Errorf("Nodes() = %d개, 원하는 값: 3", len(nodes))
	}
}

func TestDAG_AddEdge(t *testing.T) {
	d := NewDAG()
	d.AddNode(&Node{ID: "a"})
	d.AddNode(&Node{ID: "b"})

	if err := d.AddEdge("a", "b"); err != nil {
		t.Errorf("AddEdge 오류: %v", err)
	}

	deps := d.Dependencies("b")
	if len(deps) != 1 || deps[0] != "a" {
		t.Errorf("Dependencies(b) = %v, 원하는 값: [a]", deps)
	}
}

func TestDAG_AddEdge_UnknownNode(t *testing.T) {
	d := NewDAG()
	d.AddNode(&Node{ID: "a"})

	err := d.AddEdge("a", "unknown")
	if err == nil {
		t.Error("존재하지 않는 노드로의 AddEdge는 에러를 반환해야 합니다")
	}
}

// ============================================================
// 사이클 감지 테스트 (15점)
// ============================================================

func TestDAG_Validate_NoCycle(t *testing.T) {
	d := NewDAG()
	d.AddNode(&Node{ID: "a"})
	d.AddNode(&Node{ID: "b"})
	d.AddNode(&Node{ID: "c"})
	d.AddEdge("a", "b")
	d.AddEdge("b", "c")

	if err := d.Validate(); err != nil {
		t.Errorf("사이클 없는 DAG에서 Validate 에러: %v", err)
	}
}

func TestDAG_Validate_DirectCycle(t *testing.T) {
	d := NewDAG()
	d.AddNode(&Node{ID: "a"})
	d.AddNode(&Node{ID: "b"})
	d.AddEdge("a", "b")
	d.AddEdge("b", "a") // 사이클!

	if err := d.Validate(); err == nil {
		t.Error("사이클이 있는데 Validate가 nil을 반환했습니다")
	}
}

func TestDAG_Validate_IndirectCycle(t *testing.T) {
	d := NewDAG()
	for _, id := range []string{"a", "b", "c", "d"} {
		d.AddNode(&Node{ID: id})
	}
	d.AddEdge("a", "b")
	d.AddEdge("b", "c")
	d.AddEdge("c", "d")
	d.AddEdge("d", "b") // 간접 사이클: b→c→d→b

	if err := d.Validate(); err == nil {
		t.Error("간접 사이클이 있는데 Validate가 nil을 반환했습니다")
	}
}

func TestDAG_Validate_SelfLoop(t *testing.T) {
	d := NewDAG()
	d.AddNode(&Node{ID: "a"})
	d.AddEdge("a", "a") // 자기 자신에 대한 엣지

	if err := d.Validate(); err == nil {
		t.Error("자기 참조 사이클에서 Validate가 nil을 반환했습니다")
	}
}

// ============================================================
// 위상 정렬 테스트 (15점)
// ============================================================

func TestDAG_TopologicalSort_Linear(t *testing.T) {
	// a → b → c → d
	d := NewDAG()
	for _, id := range []string{"a", "b", "c", "d"} {
		d.AddNode(&Node{ID: id})
	}
	d.AddEdge("a", "b")
	d.AddEdge("b", "c")
	d.AddEdge("c", "d")

	sorted, err := d.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort 오류: %v", err)
	}
	if len(sorted) != 4 {
		t.Fatalf("정렬 결과 %d개, 원하는 값: 4", len(sorted))
	}

	// a가 b보다, b가 c보다, c가 d보다 앞에 있어야 함
	pos := make(map[string]int)
	for i, n := range sorted {
		pos[n.ID] = i
	}
	if pos["a"] > pos["b"] || pos["b"] > pos["c"] || pos["c"] > pos["d"] {
		ids := make([]string, len(sorted))
		for i, n := range sorted {
			ids[i] = n.ID
		}
		t.Errorf("위상 정렬 순서 오류: %v", ids)
	}
}

func TestDAG_TopologicalSort_Diamond(t *testing.T) {
	//   a
	//  / \
	// b   c
	//  \ /
	//   d
	d := NewDAG()
	for _, id := range []string{"a", "b", "c", "d"} {
		d.AddNode(&Node{ID: id})
	}
	d.AddEdge("a", "b")
	d.AddEdge("a", "c")
	d.AddEdge("b", "d")
	d.AddEdge("c", "d")

	sorted, err := d.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort 오류: %v", err)
	}
	pos := make(map[string]int)
	for i, n := range sorted {
		pos[n.ID] = i
	}
	if pos["a"] > pos["b"] || pos["a"] > pos["c"] {
		t.Errorf("a는 b, c보다 앞에 있어야 합니다")
	}
	if pos["b"] > pos["d"] || pos["c"] > pos["d"] {
		t.Errorf("b, c는 d보다 앞에 있어야 합니다")
	}
}

// ============================================================
// 실행 순서 테스트 (20점)
// ============================================================

func TestExecutor_DependencyOrder(t *testing.T) {
	var order []string
	var mu sync.Mutex

	d := NewDAG()
	d.AddNode(makeRecordNode("a", &order, &mu))
	d.AddNode(makeRecordNode("b", &order, &mu))
	d.AddNode(makeRecordNode("c", &order, &mu))
	d.AddEdge("a", "b")
	d.AddEdge("b", "c")

	e := NewExecutor()
	results, err := e.Execute(context.Background(), d, ExecuteOptions{})
	if err != nil {
		t.Fatalf("Execute 오류: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("결과 %d개, 원하는 값: 3", len(results))
	}

	mu.Lock()
	defer mu.Unlock()
	// a → b → c 순서 확인
	pos := make(map[string]int)
	for i, id := range order {
		pos[id] = i
	}
	if pos["a"] > pos["b"] || pos["b"] > pos["c"] {
		t.Errorf("실행 순서 오류: %v", order)
	}
}

func TestExecutor_DiamondOrder(t *testing.T) {
	var order []string
	var mu sync.Mutex

	d := NewDAG()
	d.AddNode(makeRecordNode("a", &order, &mu))
	d.AddNode(makeRecordNode("b", &order, &mu))
	d.AddNode(makeRecordNode("c", &order, &mu))
	d.AddNode(makeRecordNode("d", &order, &mu))
	d.AddEdge("a", "b")
	d.AddEdge("a", "c")
	d.AddEdge("b", "d")
	d.AddEdge("c", "d")

	e := NewExecutor()
	e.Execute(context.Background(), d, ExecuteOptions{})

	mu.Lock()
	defer mu.Unlock()
	pos := make(map[string]int)
	for i, id := range order {
		pos[id] = i
	}
	if pos["a"] > pos["d"] {
		t.Errorf("a는 d보다 먼저 실행되어야 합니다: %v", order)
	}
	if pos["b"] > pos["d"] || pos["c"] > pos["d"] {
		t.Errorf("b,c는 d보다 먼저 실행되어야 합니다: %v", order)
	}
}

// ============================================================
// 병렬 실행 테스트 (20점)
// ============================================================

func TestExecutor_ParallelExecution(t *testing.T) {
	// a, b, c는 서로 독립적 → 병렬 실행되어야 함
	var maxConcurrent int64
	var current int64

	makeParallelNode := func(id string, delay time.Duration) *Node {
		return &Node{
			ID: id,
			Task: func(ctx context.Context) error {
				c := atomic.AddInt64(&current, 1)
				for {
					old := atomic.LoadInt64(&maxConcurrent)
					if c <= old || atomic.CompareAndSwapInt64(&maxConcurrent, old, c) {
						break
					}
				}
				select {
				case <-time.After(delay):
				case <-ctx.Done():
				}
				atomic.AddInt64(&current, -1)
				return nil
			},
		}
	}

	d := NewDAG()
	d.AddNode(makeParallelNode("a", 50*time.Millisecond))
	d.AddNode(makeParallelNode("b", 50*time.Millisecond))
	d.AddNode(makeParallelNode("c", 50*time.Millisecond))
	// 모두 독립적

	e := NewExecutor()
	start := time.Now()
	e.Execute(context.Background(), d, ExecuteOptions{})
	elapsed := time.Since(start)

	// 병렬이면 ~50ms, 순차면 ~150ms
	if elapsed > 120*time.Millisecond {
		t.Errorf("병렬 실행이 되지 않음: %v (50ms~100ms 예상)", elapsed)
	}
	if atomic.LoadInt64(&maxConcurrent) < 2 {
		t.Errorf("최대 동시 실행 수 = %d, 2 이상이어야 합니다", maxConcurrent)
	}
}

func TestExecutor_MixedParallel(t *testing.T) {
	// a → c
	// b → c
	// a, b는 병렬, c는 둘 다 끝나면 실행
	var order []string
	var mu sync.Mutex

	d := NewDAG()
	for _, id := range []string{"a", "b", "c"} {
		d.AddNode(makeRecordNode(id, &order, &mu))
	}
	d.AddEdge("a", "c")
	d.AddEdge("b", "c")

	e := NewExecutor()
	e.Execute(context.Background(), d, ExecuteOptions{})

	mu.Lock()
	defer mu.Unlock()
	pos := make(map[string]int)
	for i, id := range order {
		pos[id] = i
	}
	if pos["a"] > pos["c"] || pos["b"] > pos["c"] {
		t.Errorf("a,b 모두 c보다 먼저여야 합니다: %v", order)
	}
}

// ============================================================
// DryRun 테스트 (10점)
// ============================================================

func TestExecutor_DryRun(t *testing.T) {
	executed := false
	d := NewDAG()
	d.AddNode(&Node{
		ID: "a",
		Task: func(ctx context.Context) error {
			executed = true
			return nil
		},
	})

	e := NewExecutor()
	results, err := e.Execute(context.Background(), d, ExecuteOptions{DryRun: true})
	if err != nil {
		t.Fatalf("DryRun Execute 오류: %v", err)
	}
	if executed {
		t.Error("DryRun=true인데 Task가 실행됨")
	}
	if len(results) != 1 {
		t.Errorf("DryRun 결과 %d개, 원하는 값: 1", len(results))
	}
	if results[0].NodeID != "a" {
		t.Errorf("DryRun 결과 NodeID = %s, 원하는 값: a", results[0].NodeID)
	}
}

// ============================================================
// 컨텍스트 취소 테스트 (10점)
// ============================================================

func TestExecutor_ContextCancel(t *testing.T) {
	d := NewDAG()
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("node-%d", i)
		d.AddNode(&Node{
			ID: id,
			Task: func(ctx context.Context) error {
				select {
				case <-time.After(10 * time.Second):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	e := NewExecutor()
	start := time.Now()
	e.Execute(ctx, d, ExecuteOptions{ContinueOnError: false})
	elapsed := time.Since(start)

	if elapsed > 500*time.Millisecond {
		t.Errorf("컨텍스트 취소 후 너무 오래 걸림: %v", elapsed)
	}
}

func TestExecutor_ContinueOnError(t *testing.T) {
	var completed []string
	var mu sync.Mutex

	d := NewDAG()
	d.AddNode(makeFailNode("fail"))
	d.AddNode(makeRecordNode("independent", &completed, &mu))

	e := NewExecutor()
	results, _ := e.Execute(context.Background(), d, ExecuteOptions{ContinueOnError: true})

	// 실패 노드와 독립 노드 모두 결과에 있어야 함
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.NodeID
	}
	sort.Strings(ids)

	hasIndependent := false
	for _, id := range ids {
		if id == "independent" {
			hasIndependent = true
		}
	}
	if !hasIndependent {
		t.Errorf("ContinueOnError=true인데 독립 노드가 실행되지 않음: %v", ids)
	}
}

func TestExecutor_Validate_BeforeExecute(t *testing.T) {
	d := NewDAG()
	d.AddNode(&Node{ID: "a"})
	d.AddNode(&Node{ID: "b"})
	d.AddEdge("a", "b")
	d.AddEdge("b", "a") // 사이클

	e := NewExecutor()
	_, err := e.Execute(context.Background(), d, ExecuteOptions{})
	if err == nil {
		t.Error("사이클이 있는 DAG 실행 시 에러를 반환해야 합니다")
	}
}

// ============================================================
// 채점 함수 (TestGrade)
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════════════")
	fmt.Println("  과제 A4: DAG 실행기 채점 결과")
	fmt.Println("  패턴: Terraform DAG + 병렬 실행")
	fmt.Println("═══════════════════════════════════════════════════")

	// DAG 노드/엣지 (10점)
	t.Run("DAG_기본", func(t *testing.T) {
		d := NewDAG()
		d.AddNode(&Node{ID: "x"})
		d.AddNode(&Node{ID: "y"})
		err := d.AddEdge("x", "y")
		deps := d.Dependencies("y")
		if err == nil && len(d.Nodes()) == 2 && len(deps) == 1 && deps[0] == "x" {
			score += 10
			fmt.Printf("  ✓ DAG AddNode/AddEdge/Dependencies    10/10점\n")
		} else {
			fmt.Printf("  ✗ DAG 기본 (err=%v,nodes=%d,deps=%v)    0/10점\n",
				err, len(d.Nodes()), deps)
		}
	})

	// 사이클 감지 (15점)
	t.Run("사이클_감지", func(t *testing.T) {
		// 사이클 없음
		d1 := NewDAG()
		d1.AddNode(&Node{ID: "a"})
		d1.AddNode(&Node{ID: "b"})
		d1.AddEdge("a", "b")
		noErr := d1.Validate()

		// 사이클 있음
		d2 := NewDAG()
		d2.AddNode(&Node{ID: "a"})
		d2.AddNode(&Node{ID: "b"})
		d2.AddEdge("a", "b")
		d2.AddEdge("b", "a")
		cycleErr := d2.Validate()

		if noErr == nil && cycleErr != nil {
			score += 15
			fmt.Printf("  ✓ 사이클 감지                         15/15점\n")
		} else {
			fmt.Printf("  ✗ 사이클 감지 (noErr=%v, cycleErr=%v)   0/15점\n", noErr, cycleErr)
		}
	})

	// 위상 정렬 (15점)
	t.Run("위상_정렬", func(t *testing.T) {
		d := NewDAG()
		for _, id := range []string{"a", "b", "c"} {
			d.AddNode(&Node{ID: id})
		}
		d.AddEdge("a", "b")
		d.AddEdge("b", "c")

		sorted, err := d.TopologicalSort()
		if err != nil || len(sorted) != 3 {
			fmt.Printf("  ✗ 위상 정렬 (err=%v, len=%d)            0/15점\n", err, len(sorted))
			return
		}
		pos := make(map[string]int)
		for i, n := range sorted {
			pos[n.ID] = i
		}
		if pos["a"] < pos["b"] && pos["b"] < pos["c"] {
			score += 15
			fmt.Printf("  ✓ 위상 정렬                           15/15점\n")
		} else {
			ids := make([]string, len(sorted))
			for i, n := range sorted {
				ids[i] = n.ID
			}
			fmt.Printf("  ✗ 위상 정렬 순서 오류: %v              0/15점\n", ids)
		}
	})

	// 순차 실행 (20점)
	t.Run("순차_실행", func(t *testing.T) {
		var order []string
		var mu sync.Mutex
		d := NewDAG()
		d.AddNode(makeRecordNode("x", &order, &mu))
		d.AddNode(makeRecordNode("y", &order, &mu))
		d.AddNode(makeRecordNode("z", &order, &mu))
		d.AddEdge("x", "y")
		d.AddEdge("y", "z")

		e := NewExecutor()
		e.Execute(context.Background(), d, ExecuteOptions{})

		mu.Lock()
		pos := make(map[string]int)
		for i, id := range order {
			pos[id] = i
		}
		mu.Unlock()

		if pos["x"] < pos["y"] && pos["y"] < pos["z"] {
			score += 20
			fmt.Printf("  ✓ 순차 실행 (의존성 순서 보장)        20/20점\n")
		} else {
			mu.Lock()
			fmt.Printf("  ✗ 순차 실행 순서 오류: %v              0/20점\n", order)
			mu.Unlock()
		}
	})

	// 병렬 실행 (20점)
	t.Run("병렬_실행", func(t *testing.T) {
		var maxC int64
		var curr int64

		makeTimedNode := func(id string) *Node {
			return &Node{ID: id, Task: func(ctx context.Context) error {
				c := atomic.AddInt64(&curr, 1)
				for {
					old := atomic.LoadInt64(&maxC)
					if c <= old || atomic.CompareAndSwapInt64(&maxC, old, c) {
						break
					}
				}
				time.Sleep(30 * time.Millisecond)
				atomic.AddInt64(&curr, -1)
				return nil
			}}
		}

		d := NewDAG()
		d.AddNode(makeTimedNode("p1"))
		d.AddNode(makeTimedNode("p2"))
		d.AddNode(makeTimedNode("p3"))

		e := NewExecutor()
		start := time.Now()
		e.Execute(context.Background(), d, ExecuteOptions{})
		elapsed := time.Since(start)

		if elapsed < 80*time.Millisecond && atomic.LoadInt64(&maxC) >= 2 {
			score += 20
			fmt.Printf("  ✓ 병렬 실행 (elapsed=%v, maxConc=%d)  20/20점\n",
				elapsed.Round(time.Millisecond), maxC)
		} else {
			fmt.Printf("  ✗ 병렬 실행 (elapsed=%v, maxConc=%d)   0/20점\n",
				elapsed.Round(time.Millisecond), maxC)
		}
	})

	// DryRun (10점)
	t.Run("DryRun", func(t *testing.T) {
		executed := false
		d := NewDAG()
		d.AddNode(&Node{ID: "dr", Task: func(ctx context.Context) error {
			executed = true
			return nil
		}})
		e := NewExecutor()
		results, _ := e.Execute(context.Background(), d, ExecuteOptions{DryRun: true})
		if !executed && len(results) == 1 {
			score += 10
			fmt.Printf("  ✓ DryRun 모드                         10/10점\n")
		} else {
			fmt.Printf("  ✗ DryRun (executed=%v, results=%d)      0/10점\n", executed, len(results))
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
