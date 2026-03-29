// informer_test.go
// 인포머/리스터 패턴 테스트 및 채점
//
// 실행:
//
//	go test -v -race
//	go test -v -run TestGrade
package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ============================================================
// 테스트 픽스처
// ============================================================

func makeObj(name, ns string, labels map[string]string) Object {
	return Object{Name: name, Namespace: ns, Labels: labels}
}

func makeWatchSource(events []WatchEvent) WatchFunc {
	return func(ctx context.Context) (<-chan WatchEvent, error) {
		ch := make(chan WatchEvent, len(events)+1)
		go func() {
			for _, ev := range events {
				select {
				case <-ctx.Done():
					return
				case ch <- ev:
					time.Sleep(5 * time.Millisecond)
				}
			}
		}()
		return ch, nil
	}
}

func makeListSource(objects []Object) ListFunc {
	return func(ctx context.Context) ([]Object, error) {
		result := make([]Object, len(objects))
		copy(result, objects)
		return result, nil
	}
}

// ============================================================
// Store 테스트 (20점)
// ============================================================

func TestStore_AddGet(t *testing.T) {
	s := NewStore()
	obj := makeObj("pod-1", "default", map[string]string{"app": "web"})

	if err := s.Add(obj); err != nil {
		t.Fatalf("Add 오류: %v", err)
	}

	got, ok := s.Get(KeyFunc(obj))
	if !ok {
		t.Error("Get: 추가한 오브젝트를 찾을 수 없습니다")
	}
	if got.Name != obj.Name {
		t.Errorf("Get: Name = %s, 원하는 값: %s", got.Name, obj.Name)
	}
}

func TestStore_Update(t *testing.T) {
	s := NewStore()
	obj := makeObj("pod-1", "default", map[string]string{"app": "web"})
	s.Add(obj)

	updated := obj
	updated.Labels = map[string]string{"app": "api"}
	if err := s.Update(updated); err != nil {
		t.Fatalf("Update 오류: %v", err)
	}

	got, _ := s.Get(KeyFunc(obj))
	if got.Labels["app"] != "api" {
		t.Errorf("Update 후 Labels[app] = %s, 원하는 값: api", got.Labels["app"])
	}
}

func TestStore_Delete(t *testing.T) {
	s := NewStore()
	obj := makeObj("pod-1", "default", nil)
	s.Add(obj)
	s.Delete(obj)

	_, ok := s.Get(KeyFunc(obj))
	if ok {
		t.Error("Delete 후 Get이 여전히 오브젝트를 반환합니다")
	}
}

func TestStore_List(t *testing.T) {
	s := NewStore()
	objects := []Object{
		makeObj("pod-1", "default", nil),
		makeObj("pod-2", "default", nil),
		makeObj("pod-3", "kube-system", nil),
	}
	for _, obj := range objects {
		s.Add(obj)
	}

	list := s.List()
	if len(list) != 3 {
		t.Errorf("List() = %d개, 원하는 값: 3", len(list))
	}
}

func TestStore_Replace(t *testing.T) {
	s := NewStore()
	s.Add(makeObj("old-1", "default", nil))
	s.Add(makeObj("old-2", "default", nil))

	newObjects := []Object{
		makeObj("new-1", "default", nil),
		makeObj("new-2", "default", nil),
		makeObj("new-3", "default", nil),
	}
	s.Replace(newObjects)

	list := s.List()
	if len(list) != 3 {
		t.Errorf("Replace 후 List() = %d개, 원하는 값: 3", len(list))
	}

	_, ok := s.Get("default/old-1")
	if ok {
		t.Error("Replace 후 old-1이 여전히 존재합니다")
	}
}

// ============================================================
// Store 인덱싱 테스트 (15점)
// ============================================================

func TestStore_AddIndexer(t *testing.T) {
	s := NewStore()
	s.AddIndexer("byApp", LabelIndexFunc("app"))

	s.Add(makeObj("pod-1", "default", map[string]string{"app": "web"}))
	s.Add(makeObj("pod-2", "default", map[string]string{"app": "web"}))
	s.Add(makeObj("pod-3", "default", map[string]string{"app": "api"}))

	webPods, err := s.ListByIndex("byApp", "web")
	if err != nil {
		t.Fatalf("ListByIndex 오류: %v", err)
	}
	if len(webPods) != 2 {
		t.Errorf("byApp=web: %d개, 원하는 값: 2", len(webPods))
	}
}

func TestStore_IndexAfterUpdate(t *testing.T) {
	s := NewStore()
	s.AddIndexer("byApp", LabelIndexFunc("app"))

	obj := makeObj("pod-1", "default", map[string]string{"app": "web"})
	s.Add(obj)

	// 레이블 변경
	obj.Labels = map[string]string{"app": "api"}
	s.Update(obj)

	webPods, _ := s.ListByIndex("byApp", "web")
	if len(webPods) != 0 {
		t.Errorf("Update 후 byApp=web에 아직 %d개 있음", len(webPods))
	}

	apiPods, _ := s.ListByIndex("byApp", "api")
	if len(apiPods) != 1 {
		t.Errorf("Update 후 byApp=api: %d개, 원하는 값: 1", len(apiPods))
	}
}

func TestStore_NamespaceIndex(t *testing.T) {
	s := NewStore()
	s.AddIndexer("byNamespace", NamespaceIndexFunc)

	s.Add(makeObj("pod-1", "default", nil))
	s.Add(makeObj("pod-2", "default", nil))
	s.Add(makeObj("pod-3", "kube-system", nil))

	defaultPods, _ := s.ListByIndex("byNamespace", "default")
	if len(defaultPods) != 2 {
		t.Errorf("byNamespace=default: %d개, 원하는 값: 2", len(defaultPods))
	}
}

// ============================================================
// Reflector 테스트 (15점)
// ============================================================

func TestReflector_InitialList(t *testing.T) {
	initial := []Object{
		makeObj("pod-1", "default", nil),
		makeObj("pod-2", "default", nil),
	}
	s := NewStore()
	r := NewReflector(s, makeListSource(initial), makeWatchSource(nil))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go r.Run(ctx)

	// HasSynced가 true가 될 때까지 대기
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if r.HasSynced() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !r.HasSynced() {
		t.Error("초기 List 후 HasSynced()가 true여야 합니다")
	}
	if len(s.List()) != 2 {
		t.Errorf("초기 List 후 Store에 %d개, 원하는 값: 2", len(s.List()))
	}
}

func TestReflector_WatchEvents(t *testing.T) {
	events := []WatchEvent{
		{Type: "ADDED", Object: makeObj("pod-3", "default", nil)},
		{Type: "MODIFIED", Object: makeObj("pod-1", "default", map[string]string{"updated": "true"})},
		{Type: "DELETED", Object: makeObj("pod-2", "default", nil)},
	}

	initial := []Object{
		makeObj("pod-1", "default", nil),
		makeObj("pod-2", "default", nil),
	}

	s := NewStore()
	r := NewReflector(s, makeListSource(initial), makeWatchSource(events))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go r.Run(ctx)

	time.Sleep(300 * time.Millisecond)

	// pod-3이 추가되었는지
	_, ok := s.Get("default/pod-3")
	if !ok {
		t.Error("ADDED 이벤트 후 pod-3이 Store에 없습니다")
	}

	// pod-2가 삭제되었는지
	_, ok = s.Get("default/pod-2")
	if ok {
		t.Error("DELETED 이벤트 후 pod-2가 여전히 Store에 있습니다")
	}

	// pod-1이 업데이트되었는지
	pod1, ok := s.Get("default/pod-1")
	if !ok {
		t.Error("pod-1이 Store에 없습니다")
	} else if pod1.Labels["updated"] != "true" {
		t.Error("MODIFIED 이벤트 후 pod-1 레이블이 업데이트되지 않았습니다")
	}
}

// ============================================================
// Informer 이벤트 핸들러 테스트 (20점)
// ============================================================

// trackingHandler는 호출된 이벤트를 기록하는 EventHandler입니다.
type trackingHandler struct {
	mu      sync.Mutex
	added   []Object
	updated []Object
	deleted []Object
}

func (h *trackingHandler) OnAdd(obj Object) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.added = append(h.added, obj)
}

func (h *trackingHandler) OnUpdate(_, newObj Object) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.updated = append(h.updated, newObj)
}

func (h *trackingHandler) OnDelete(obj Object) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.deleted = append(h.deleted, obj)
}

func (h *trackingHandler) counts() (added, updated, deleted int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.added), len(h.updated), len(h.deleted)
}

func TestInformer_EventHandlers(t *testing.T) {
	initial := []Object{makeObj("pod-1", "default", nil)}
	events := []WatchEvent{
		{Type: "ADDED", Object: makeObj("pod-2", "default", nil)},
		{Type: "MODIFIED", Object: makeObj("pod-1", "default", map[string]string{"v": "2"})},
		{Type: "DELETED", Object: makeObj("pod-2", "default", nil)},
	}

	inf := NewInformer(makeListSource(initial), makeWatchSource(events))
	h := &trackingHandler{}
	inf.AddEventHandler(h)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go inf.Run(ctx)

	// 모든 이벤트가 처리될 때까지 대기
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		added, _, deleted := h.counts()
		if added >= 2 && deleted >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	added, updated, deleted := h.counts()
	if added < 1 {
		t.Errorf("OnAdd 호출 횟수 = %d, 1 이상이어야 합니다", added)
	}
	if updated < 1 {
		t.Errorf("OnUpdate 호출 횟수 = %d, 1 이상이어야 합니다", updated)
	}
	if deleted < 1 {
		t.Errorf("OnDelete 호출 횟수 = %d, 1 이상이어야 합니다", deleted)
	}
}

func TestInformer_MultipleHandlers(t *testing.T) {
	initial := []Object{makeObj("pod-1", "default", nil)}
	inf := NewInformer(makeListSource(initial), makeWatchSource(nil))

	h1 := &trackingHandler{}
	h2 := &trackingHandler{}
	inf.AddEventHandler(h1)
	inf.AddEventHandler(h2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go inf.Run(ctx)

	time.Sleep(200 * time.Millisecond)

	a1, _, _ := h1.counts()
	a2, _, _ := h2.counts()
	if a1 == 0 || a2 == 0 {
		t.Errorf("두 핸들러 모두 OnAdd를 받아야 합니다: h1=%d, h2=%d", a1, a2)
	}
}

// ============================================================
// HasSynced 테스트 (10점)
// ============================================================

func TestInformer_HasSynced(t *testing.T) {
	initial := []Object{
		makeObj("pod-1", "default", nil),
		makeObj("pod-2", "default", nil),
	}
	inf := NewInformer(makeListSource(initial), makeWatchSource(nil))

	if inf.HasSynced() {
		t.Error("Run 전에는 HasSynced()가 false여야 합니다")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	go inf.Run(ctx)

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if inf.HasSynced() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !inf.HasSynced() {
		t.Error("초기 List 완료 후 HasSynced()가 true여야 합니다")
	}
}

// ============================================================
// Lister 테스트 (10점)
// ============================================================

func TestLister_ListAndGet(t *testing.T) {
	s := NewStore()
	s.Add(makeObj("pod-1", "default", map[string]string{"app": "web"}))
	s.Add(makeObj("pod-2", "default", map[string]string{"app": "api"}))
	s.Add(makeObj("pod-3", "kube-system", map[string]string{"app": "web"}))

	lister := NewLister(s)

	all := lister.List()
	if len(all) != 3 {
		t.Errorf("Lister.List() = %d개, 원하는 값: 3", len(all))
	}

	obj, ok := lister.Get("default/pod-1")
	if !ok {
		t.Error("Lister.Get(default/pod-1) 실패")
	}
	if obj.Name != "pod-1" {
		t.Errorf("Lister.Get Name = %s, 원하는 값: pod-1", obj.Name)
	}
}

func TestLister_ListByLabel(t *testing.T) {
	s := NewStore()
	s.Add(makeObj("pod-1", "default", map[string]string{"app": "web"}))
	s.Add(makeObj("pod-2", "default", map[string]string{"app": "web"}))
	s.Add(makeObj("pod-3", "default", map[string]string{"app": "api"}))

	lister := NewLister(s)

	webPods := lister.ListByLabel("app", "web")
	if len(webPods) != 2 {
		t.Errorf("ListByLabel(app=web) = %d개, 원하는 값: 2", len(webPods))
	}
}

// ============================================================
// 동시성 안전 테스트 (10점) - go test -race로 검증
// ============================================================

func TestStore_ConcurrentAccess(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup

	// 동시 쓰기
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			obj := makeObj(fmt.Sprintf("pod-%d", i), "default", nil)
			s.Add(obj)
		}(i)
	}

	// 동시 읽기
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.List()
		}()
	}

	wg.Wait()

	if len(s.List()) != 20 {
		t.Errorf("동시 접근 후 List() = %d개, 원하는 값: 20", len(s.List()))
	}
}

// ============================================================
// 채점 함수 (TestGrade)
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════════════")
	fmt.Println("  과제 A2: 인포머/리스터 패턴 채점 결과")
	fmt.Println("  패턴: client-go Informer/Store/Reflector/Lister")
	fmt.Println("═══════════════════════════════════════════════════")

	// Store CRUD (20점)
	t.Run("Store_CRUD", func(t *testing.T) {
		s := NewStore()
		obj := makeObj("p1", "ns", map[string]string{"k": "v"})
		s.Add(obj)
		got, ok := s.Get("ns/p1")
		if !ok || got.Name != "p1" {
			fmt.Printf("  ✗ Store CRUD                           0/20점\n")
			return
		}
		obj.Labels = map[string]string{"k": "v2"}
		s.Update(obj)
		got, _ = s.Get("ns/p1")
		if got.Labels["k"] != "v2" {
			fmt.Printf("  ✗ Store Update                         0/20점\n")
			return
		}
		s.Delete(obj)
		_, ok = s.Get("ns/p1")
		if ok {
			fmt.Printf("  ✗ Store Delete                         0/20점\n")
			return
		}
		score += 20
		fmt.Printf("  ✓ Store CRUD (Add/Get/Update/Delete)  20/20점\n")
	})

	// Store 인덱싱 (15점)
	t.Run("Store_인덱싱", func(t *testing.T) {
		s := NewStore()
		s.AddIndexer("byApp", LabelIndexFunc("app"))
		s.Add(makeObj("p1", "ns", map[string]string{"app": "web"}))
		s.Add(makeObj("p2", "ns", map[string]string{"app": "web"}))
		s.Add(makeObj("p3", "ns", map[string]string{"app": "api"}))

		webList, err := s.ListByIndex("byApp", "web")
		if err != nil || len(webList) != 2 {
			fmt.Printf("  ✗ Store 인덱싱                         0/15점\n")
			return
		}
		score += 15
		fmt.Printf("  ✓ Store 인덱싱 (AddIndexer/ListByIndex)15/15점\n")
	})

	// Reflector → Store 동기화 (15점)
	t.Run("Reflector_동기화", func(t *testing.T) {
		initial := []Object{makeObj("p1", "ns", nil), makeObj("p2", "ns", nil)}
		events := []WatchEvent{{Type: "ADDED", Object: makeObj("p3", "ns", nil)}}
		s := NewStore()
		r := NewReflector(s, makeListSource(initial), makeWatchSource(events))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		go r.Run(ctx)
		time.Sleep(300 * time.Millisecond)

		if !r.HasSynced() || len(s.List()) < 3 {
			fmt.Printf("  ✗ Reflector 동기화 (synced=%v, count=%d)  0/15점\n", r.HasSynced(), len(s.List()))
			return
		}
		score += 15
		fmt.Printf("  ✓ Reflector → Store 동기화            15/15점\n")
	})

	// Informer EventHandler 콜백 (20점)
	t.Run("Informer_EventHandler", func(t *testing.T) {
		initial := []Object{makeObj("p1", "ns", nil)}
		events := []WatchEvent{
			{Type: "ADDED", Object: makeObj("p2", "ns", nil)},
			{Type: "DELETED", Object: makeObj("p1", "ns", nil)},
		}
		inf := NewInformer(makeListSource(initial), makeWatchSource(events))
		h := &trackingHandler{}
		inf.AddEventHandler(h)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		go inf.Run(ctx)
		time.Sleep(400 * time.Millisecond)

		added, _, deleted := h.counts()
		if added >= 1 && deleted >= 1 {
			score += 20
			fmt.Printf("  ✓ Informer EventHandler 콜백          20/20점\n")
		} else {
			fmt.Printf("  ✗ Informer EventHandler (add=%d,del=%d)  0/20점\n", added, deleted)
		}
	})

	// HasSynced (10점)
	t.Run("HasSynced", func(t *testing.T) {
		inf := NewInformer(makeListSource([]Object{makeObj("p1", "ns", nil)}), makeWatchSource(nil))
		if inf.HasSynced() {
			fmt.Printf("  ✗ HasSynced (Run 전 true여서는 안 됨)   0/10점\n")
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		go inf.Run(ctx)
		time.Sleep(300 * time.Millisecond)
		if inf.HasSynced() {
			score += 10
			fmt.Printf("  ✓ HasSynced                           10/10점\n")
		} else {
			fmt.Printf("  ✗ HasSynced (List 후에도 false)         0/10점\n")
		}
	})

	// Lister (10점)
	t.Run("Lister", func(t *testing.T) {
		s := NewStore()
		s.Add(makeObj("p1", "ns", map[string]string{"env": "prod"}))
		s.Add(makeObj("p2", "ns", map[string]string{"env": "dev"}))
		lister := NewLister(s)

		all := lister.List()
		_, ok := lister.Get("ns/p1")
		prod := lister.ListByLabel("env", "prod")

		if len(all) == 2 && ok && len(prod) == 1 {
			score += 10
			fmt.Printf("  ✓ Lister (List/Get/ListByLabel)       10/10점\n")
		} else {
			fmt.Printf("  ✗ Lister (all=%d,get=%v,prod=%d)        0/10점\n", len(all), ok, len(prod))
		}
	})

	fmt.Println("───────────────────────────────────────────────────")
	fmt.Printf("  최종 점수: %d / %d점\n", score, total)
	fmt.Printf("  (Race Detector 통과 여부는 go test -race로 확인)\n")

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
