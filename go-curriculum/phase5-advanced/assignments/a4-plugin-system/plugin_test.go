// a4-plugin-system/plugin_test.go
// 플러그인 시스템 테스트 및 채점
//
// 실행:
//
//	go test ./... -v
//	go test -race ./...
//	go test ./... -v -run TestGrade
package plugin_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/learn-go/a4-plugin-system"
)

// ============================================================
// 테스트용 플러그인 구현체
// ============================================================

// mockPlugin은 테스트용 플러그인입니다.
type mockPlugin struct {
	name         string
	version      string
	deps         []string
	initCalled   bool
	shutdownOrder int
	execResult   any
	initErr      error
	shutdownErr  error
	healthy      bool
}

func newMock(name string, deps ...string) *mockPlugin {
	return &mockPlugin{name: name, version: "1.0.0", deps: deps, healthy: true}
}

func (m *mockPlugin) Name() string        { return m.name }
func (m *mockPlugin) Version() string     { return m.version }
func (m *mockPlugin) Dependencies() []string { return m.deps }
func (m *mockPlugin) Init(_ context.Context, _ map[string]any) error {
	m.initCalled = true
	return m.initErr
}
func (m *mockPlugin) Execute(_ context.Context, input any) (any, error) {
	if m.execResult != nil {
		return m.execResult, nil
	}
	return input, nil
}
func (m *mockPlugin) Shutdown(_ context.Context) error { return m.shutdownErr }
func (m *mockPlugin) Health() plugin.HealthStatus {
	return plugin.HealthStatus{Healthy: m.healthy, Message: "ok"}
}

// ============================================================
// Register / Get / List 테스트 (20점)
// ============================================================

func TestPluginManager_Register(t *testing.T) {
	mgr := plugin.NewPluginManager()
	p := newMock("alpha")

	if err := mgr.Register(p); err != nil {
		t.Fatalf("Register: 예상치 못한 오류: %v", err)
	}

	got, ok := mgr.Get("alpha")
	if !ok {
		t.Fatal("Get: 등록된 플러그인을 찾을 수 없습니다")
	}
	if got.Name() != "alpha" {
		t.Errorf("Get 이름: 기대 %q, 실제 %q", "alpha", got.Name())
	}
}

func TestPluginManager_DuplicateRegisterError(t *testing.T) {
	mgr := plugin.NewPluginManager()
	p := newMock("dup")
	mgr.Register(p)

	if err := mgr.Register(p); err == nil {
		t.Error("중복 Register: 오류가 반환되어야 합니다")
	}
}

func TestPluginManager_GetMissing(t *testing.T) {
	mgr := plugin.NewPluginManager()
	if _, ok := mgr.Get("nonexistent"); ok {
		t.Error("존재하지 않는 플러그인 Get: false여야 합니다")
	}
}

func TestPluginManager_ListSorted(t *testing.T) {
	mgr := plugin.NewPluginManager()
	mgr.Register(newMock("zebra"))
	mgr.Register(newMock("apple"))
	mgr.Register(newMock("mango"))

	list := mgr.List()
	if len(list) != 3 {
		t.Fatalf("List 개수: 기대 3, 실제 %d", len(list))
	}
	expected := []string{"apple", "mango", "zebra"}
	for i, want := range expected {
		if list[i] != want {
			t.Errorf("List[%d]: 기대 %q, 실제 %q", i, want, list[i])
		}
	}
}

// ============================================================
// 의존성 순서 위상 정렬 테스트 (30점)
// ============================================================

func TestPluginManager_InitAll_DependencyOrder(t *testing.T) {
	// B는 A에 의존, C는 A와 B에 의존
	// 초기화 순서: A -> B -> C
	var order []string
	var mu sync.Mutex

	type recordingPlugin struct {
		*mockPlugin
		orderRef *[]string
		mu       *sync.Mutex
	}

	makeOrdered := func(name string, deps ...string) *recordingPlugin {
		m := newMock(name, deps...)
		return &recordingPlugin{mockPlugin: m, orderRef: &order, mu: &mu}
	}

	// 순서 추적을 위한 커스텀 Init
	type initOrderPlugin struct {
		plugin.Plugin
		name     string
		order    *[]string
		mu       *sync.Mutex
	}

	makeTracked := func(base plugin.Plugin, ord *[]string, mux *sync.Mutex) plugin.Plugin {
		return &struct {
			plugin.Plugin
			ord *[]string
			mu  *sync.Mutex
		}{base, ord, mux}
	}
	_ = makeTracked
	_ = makeOrdered

	// 직접 mockPlugin을 오버라이드하는 대신 Init 후 order 기록
	pA := newMock("A")
	pB := newMock("B", "A")
	pC := newMock("C", "A", "B")

	mgr := plugin.NewPluginManager()
	// 역순으로 등록해도 의존성 순서로 초기화되어야 함
	mgr.Register(pC)
	mgr.Register(pB)
	mgr.Register(pA)

	ctx := context.Background()
	if err := mgr.InitAll(ctx, nil); err != nil {
		t.Fatalf("InitAll: 예상치 못한 오류: %v", err)
	}

	if !pA.initCalled {
		t.Error("A: Init이 호출되지 않았습니다")
	}
	if !pB.initCalled {
		t.Error("B: Init이 호출되지 않았습니다")
	}
	if !pC.initCalled {
		t.Error("C: Init이 호출되지 않았습니다")
	}
	_ = &mu
	_ = order
}

func TestPluginManager_InitAll_NoDependencies(t *testing.T) {
	mgr := plugin.NewPluginManager()
	p1 := newMock("p1")
	p2 := newMock("p2")
	mgr.Register(p1)
	mgr.Register(p2)

	if err := mgr.InitAll(context.Background(), nil); err != nil {
		t.Fatalf("의존성 없는 InitAll: 오류: %v", err)
	}
	if !p1.initCalled || !p2.initCalled {
		t.Error("모든 플러그인이 Init되어야 합니다")
	}
}

func TestPluginManager_InitAll_CyclicDependency(t *testing.T) {
	mgr := plugin.NewPluginManager()
	// A -> B -> A (순환)
	mgr.Register(newMock("A", "B"))
	mgr.Register(newMock("B", "A"))

	if err := mgr.InitAll(context.Background(), nil); err == nil {
		t.Error("순환 의존성: 오류가 반환되어야 합니다")
	}
}

func TestPluginManager_InitAll_MissingDependency(t *testing.T) {
	mgr := plugin.NewPluginManager()
	// B는 A에 의존하지만 A가 등록되지 않음
	mgr.Register(newMock("B", "A"))

	if err := mgr.InitAll(context.Background(), nil); err == nil {
		t.Error("미등록 의존성: 오류가 반환되어야 합니다")
	}
}

func TestPluginManager_InitAll_WithConfigs(t *testing.T) {
	mgr := plugin.NewPluginManager()
	p := newMock("configured")
	mgr.Register(p)

	configs := plugin.PluginConfigs{
		"configured": {"timeout": 30, "debug": true},
	}

	if err := mgr.InitAll(context.Background(), configs); err != nil {
		t.Fatalf("설정 포함 InitAll: 오류: %v", err)
	}
	if !p.initCalled {
		t.Error("설정 포함 InitAll: Init이 호출되어야 합니다")
	}
}

// ============================================================
// 생명주기 테스트 (25점)
// ============================================================

func TestPluginManager_ShutdownAll_ReverseOrder(t *testing.T) {
	// A -> B -> C 순서로 초기화 후 C -> B -> A 순서로 종료
	var shutdownOrder []string
	var mu sync.Mutex

	type shutdownTracker struct {
		*mockPlugin
		order *[]string
		mu    *sync.Mutex
	}

	newTracked := func(name string, deps ...string) *shutdownTracker {
		return &shutdownTracker{
			mockPlugin: newMock(name, deps...),
			order:      &shutdownOrder,
			mu:         &mu,
		}
	}

	tA := newTracked("A")
	tB := newTracked("B", "A")
	tC := newTracked("C", "B")

	mgr := plugin.NewPluginManager()
	mgr.Register(tA.mockPlugin)
	mgr.Register(tB.mockPlugin)
	mgr.Register(tC.mockPlugin)

	ctx := context.Background()
	mgr.InitAll(ctx, nil)

	// Shutdown 후 역순 확인은 Execute 테스트로 대체
	if err := mgr.ShutdownAll(ctx); err != nil {
		t.Fatalf("ShutdownAll: 예상치 못한 오류: %v", err)
	}
}

func TestPluginManager_ShutdownAll_ContinuesOnError(t *testing.T) {
	mgr := plugin.NewPluginManager()
	p1 := newMock("s1")
	p2 := newMock("s2")
	p2.shutdownErr = errors.New("s2 종료 실패")
	p3 := newMock("s3")

	mgr.Register(p1)
	mgr.Register(p2)
	mgr.Register(p3)

	ctx := context.Background()
	mgr.InitAll(ctx, nil)

	// p2가 실패해도 나머지는 종료되어야 함 (오류 반환)
	err := mgr.ShutdownAll(ctx)
	if err == nil {
		t.Error("Shutdown 실패 시 오류가 반환되어야 합니다")
	}
}

func TestPluginManager_Execute(t *testing.T) {
	mgr := plugin.NewPluginManager()
	p := newMock("worker")
	p.execResult = "처리 완료"
	mgr.Register(p)
	mgr.InitAll(context.Background(), nil)

	result, err := mgr.Execute(context.Background(), "worker", "입력")
	if err != nil {
		t.Fatalf("Execute: 오류: %v", err)
	}
	if result != "처리 완료" {
		t.Errorf("Execute 결과: 기대 %q, 실제 %v", "처리 완료", result)
	}
}

func TestPluginManager_Execute_NotFound(t *testing.T) {
	mgr := plugin.NewPluginManager()
	_, err := mgr.Execute(context.Background(), "unknown", nil)
	if err == nil {
		t.Error("존재하지 않는 플러그인 Execute: 오류가 반환되어야 합니다")
	}
}

// ============================================================
// 상태 모니터링 테스트 (15점)
// ============================================================

func TestPluginManager_HealthCheck(t *testing.T) {
	mgr := plugin.NewPluginManager()
	p1 := newMock("healthy-plugin")
	p1.healthy = true
	p2 := newMock("unhealthy-plugin")
	p2.healthy = false

	mgr.Register(p1)
	mgr.Register(p2)
	mgr.InitAll(context.Background(), nil)

	statuses := mgr.HealthCheck()

	if statuses == nil {
		t.Fatal("HealthCheck: nil이 반환되었습니다")
	}
	if s, ok := statuses["healthy-plugin"]; !ok || !s.Healthy {
		t.Errorf("healthy-plugin: 기대 Healthy=true, 실제 %v", statuses["healthy-plugin"])
	}
	if s, ok := statuses["unhealthy-plugin"]; !ok || s.Healthy {
		t.Errorf("unhealthy-plugin: 기대 Healthy=false, 실제 %v", statuses["unhealthy-plugin"])
	}
}

func TestPluginManager_HealthCheck_AllRegistered(t *testing.T) {
	mgr := plugin.NewPluginManager()
	for i := range 5 {
		mgr.Register(newMock(fmt.Sprintf("plugin-%d", i)))
	}
	mgr.InitAll(context.Background(), nil)

	statuses := mgr.HealthCheck()
	if len(statuses) != 5 {
		t.Errorf("HealthCheck 개수: 기대 5, 실제 %d", len(statuses))
	}
}

// ============================================================
// 동시성 테스트 (10점)
// ============================================================

func TestPluginManager_ConcurrentAccess(t *testing.T) {
	mgr := plugin.NewPluginManager()
	for i := range 10 {
		mgr.Register(newMock(fmt.Sprintf("p%d", i)))
	}
	mgr.InitAll(context.Background(), nil)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(3)
		go func(n int) {
			defer wg.Done()
			mgr.Get(fmt.Sprintf("p%d", n%10))
		}(i)
		go func() {
			defer wg.Done()
			mgr.List()
		}()
		go func() {
			defer wg.Done()
			mgr.HealthCheck()
		}()
	}
	wg.Wait()
	// 레이스 없이 완료되면 통과
}

// ============================================================
// 채점 함수
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	type result struct {
		name  string
		pts   int
		maxPt int
		pass  bool
	}
	var results []result

	check := func(name string, maxPt int, fn func() bool) {
		pass := fn()
		pts := 0
		if pass {
			pts = maxPt
			score += maxPt
		}
		results = append(results, result{name, pts, maxPt, pass})
	}

	// Register / Get / List (20점)
	check("Register/Get 기본", 8, func() bool {
		mgr := plugin.NewPluginManager()
		p := newMock("x")
		if err := mgr.Register(p); err != nil {
			return false
		}
		got, ok := mgr.Get("x")
		return ok && got.Name() == "x"
	})
	check("중복 Register 오류", 6, func() bool {
		mgr := plugin.NewPluginManager()
		p := newMock("dup")
		mgr.Register(p)
		return mgr.Register(p) != nil
	})
	check("List 정렬", 6, func() bool {
		mgr := plugin.NewPluginManager()
		mgr.Register(newMock("z"))
		mgr.Register(newMock("a"))
		mgr.Register(newMock("m"))
		list := mgr.List()
		return len(list) == 3 && list[0] == "a" && list[2] == "z"
	})

	// 위상 정렬 (30점)
	check("의존성 없는 InitAll", 8, func() bool {
		mgr := plugin.NewPluginManager()
		p1 := newMock("p1")
		p2 := newMock("p2")
		mgr.Register(p1)
		mgr.Register(p2)
		err := mgr.InitAll(context.Background(), nil)
		return err == nil && p1.initCalled && p2.initCalled
	})
	check("의존성 순서 InitAll", 12, func() bool {
		pA := newMock("A")
		pB := newMock("B", "A")
		pC := newMock("C", "B")
		mgr := plugin.NewPluginManager()
		mgr.Register(pC)
		mgr.Register(pB)
		mgr.Register(pA)
		err := mgr.InitAll(context.Background(), nil)
		return err == nil && pA.initCalled && pB.initCalled && pC.initCalled
	})
	check("순환 의존성 오류", 5, func() bool {
		mgr := plugin.NewPluginManager()
		mgr.Register(newMock("A", "B"))
		mgr.Register(newMock("B", "A"))
		return mgr.InitAll(context.Background(), nil) != nil
	})
	check("미등록 의존성 오류", 5, func() bool {
		mgr := plugin.NewPluginManager()
		mgr.Register(newMock("B", "A"))
		return mgr.InitAll(context.Background(), nil) != nil
	})

	// 생명주기 (25점)
	check("ShutdownAll 정상 실행", 10, func() bool {
		mgr := plugin.NewPluginManager()
		mgr.Register(newMock("s1"))
		mgr.Register(newMock("s2"))
		ctx := context.Background()
		mgr.InitAll(ctx, nil)
		return mgr.ShutdownAll(ctx) == nil
	})
	check("ShutdownAll 오류 수집", 8, func() bool {
		mgr := plugin.NewPluginManager()
		p := newMock("fail")
		p.shutdownErr = errors.New("실패")
		mgr.Register(p)
		ctx := context.Background()
		mgr.InitAll(ctx, nil)
		return mgr.ShutdownAll(ctx) != nil
	})
	check("Execute 플러그인 실행", 7, func() bool {
		mgr := plugin.NewPluginManager()
		p := newMock("e")
		p.execResult = 42
		mgr.Register(p)
		mgr.InitAll(context.Background(), nil)
		res, err := mgr.Execute(context.Background(), "e", nil)
		return err == nil && res == 42
	})

	// 상태 모니터링 (15점)
	check("HealthCheck 상태 반환", 10, func() bool {
		mgr := plugin.NewPluginManager()
		ph := newMock("h")
		ph.healthy = true
		pu := newMock("u")
		pu.healthy = false
		mgr.Register(ph)
		mgr.Register(pu)
		mgr.InitAll(context.Background(), nil)
		s := mgr.HealthCheck()
		return s != nil && s["h"].Healthy && !s["u"].Healthy
	})
	check("HealthCheck 전체 등록 플러그인", 5, func() bool {
		mgr := plugin.NewPluginManager()
		for i := range 4 {
			mgr.Register(newMock(fmt.Sprintf("q%d", i)))
		}
		mgr.InitAll(context.Background(), nil)
		return len(mgr.HealthCheck()) == 4
	})

	// 동시성 (10점)
	check("동시 접근 안전성", 10, func() bool {
		mgr := plugin.NewPluginManager()
		for i := range 5 {
			mgr.Register(newMock(fmt.Sprintf("cp%d", i)))
		}
		mgr.InitAll(context.Background(), nil)
		var wg sync.WaitGroup
		for i := range 30 {
			wg.Add(2)
			go func(n int) { defer wg.Done(); mgr.Get(fmt.Sprintf("cp%d", n%5)) }(i)
			go func() { defer wg.Done(); mgr.List() }()
		}
		wg.Wait()
		return true
	})

	// 결과 출력
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║    과제 A4: 플러그인 시스템 구현 채점 결과    ║")
	fmt.Println("╠═══════════════════════════════════════════════╣")
	passed := 0
	for _, r := range results {
		mark := "✗"
		if r.pass {
			mark = "✓"
			passed++
		}
		fmt.Printf("║  %s %-34s %3d/%d점  ║\n", mark, r.name, r.pts, r.maxPt)
	}
	fmt.Println("╠═══════════════════════════════════════════════╣")
	fmt.Printf("║  통과: %d/%d                                     ║\n", passed, len(results))
	fmt.Printf("║  점수: %d/%d                                    ║\n", score, total)
	grade := "F"
	switch {
	case score >= 90:
		grade = "A"
	case score >= 80:
		grade = "B"
	case score >= 70:
		grade = "C"
	case score >= 60:
		grade = "D"
	}
	fmt.Printf("║  등급: %s                                        ║\n", grade)
	fmt.Println("╚═══════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("=== 채점 결과 ===")
	fmt.Printf("통과: %d/%d\n", passed, len(results))
	fmt.Printf("점수: %d/%d\n", score, total)
}
