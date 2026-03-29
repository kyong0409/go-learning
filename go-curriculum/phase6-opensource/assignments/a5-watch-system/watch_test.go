// watch_test.go
// 감시 시스템 테스트 및 채점
//
// 실행:
//
//	go test -v
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
// 기본 KV 테스트 (15점)
// ============================================================

func TestStore_Put(t *testing.T) {
	s := NewWatchableStore()

	rev1 := s.Put("key1", "value1")
	rev2 := s.Put("key2", "value2")

	if rev2 <= rev1 {
		t.Errorf("리비전이 단조 증가해야 합니다: rev1=%d, rev2=%d", rev1, rev2)
	}
	if s.CurrentRevision() != rev2 {
		t.Errorf("CurrentRevision() = %d, 원하는 값: %d", s.CurrentRevision(), rev2)
	}
}

func TestStore_Get(t *testing.T) {
	s := NewWatchableStore()
	s.Put("foo", "bar")

	val, rev, ok := s.Get("foo")
	if !ok {
		t.Error("Get: 존재하는 키를 찾지 못했습니다")
	}
	if val != "bar" {
		t.Errorf("Get value = %q, 원하는 값: %q", val, "bar")
	}
	if rev <= 0 {
		t.Errorf("Get revision = %d, 0보다 커야 합니다", rev)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	s := NewWatchableStore()
	_, _, ok := s.Get("nonexistent")
	if ok {
		t.Error("존재하지 않는 키에 대해 ok=true를 반환했습니다")
	}
}

func TestStore_Delete(t *testing.T) {
	s := NewWatchableStore()
	s.Put("k", "v")
	rev := s.Delete("k")

	if rev <= 0 {
		t.Errorf("Delete 리비전 = %d, 0보다 커야 합니다", rev)
	}
	_, _, ok := s.Get("k")
	if ok {
		t.Error("Delete 후 Get이 여전히 값을 반환합니다")
	}
}

func TestStore_DeleteNonExistent(t *testing.T) {
	s := NewWatchableStore()
	rev := s.Delete("nokey")
	if rev != -1 {
		t.Errorf("존재하지 않는 키 Delete = %d, 원하는 값: -1", rev)
	}
}

func TestStore_RevisionMonotonicallyIncreasing(t *testing.T) {
	s := NewWatchableStore()
	var prevRev int64
	for i := 0; i < 10; i++ {
		rev := s.Put(fmt.Sprintf("key-%d", i), "val")
		if rev <= prevRev {
			t.Errorf("리비전이 감소했습니다: prev=%d, curr=%d", prevRev, rev)
		}
		prevRev = rev
	}
}

// ============================================================
// Watch 단일 키 테스트 (20점)
// ============================================================

func TestWatch_SingleKey(t *testing.T) {
	s := NewWatchableStore()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := s.Watch(ctx, "mykey")

	s.Put("mykey", "v1")
	s.Put("otherkey", "v2") // 무시되어야 함
	s.Put("mykey", "v2")

	var events []WatchEvent
	timeout := time.After(500 * time.Millisecond)
loop:
	for {
		select {
		case resp, ok := <-ch:
			if !ok {
				break loop
			}
			events = append(events, resp.Events...)
			if len(events) >= 2 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	if len(events) < 2 {
		t.Errorf("mykey 이벤트 %d개, 원하는 값: 2개 이상", len(events))
		return
	}
	for _, ev := range events {
		if ev.Key != "mykey" {
			t.Errorf("Watch(mykey)에서 다른 키 이벤트 수신: %s", ev.Key)
		}
	}
}

func TestWatch_PutEventType(t *testing.T) {
	s := NewWatchableStore()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ch := s.Watch(ctx, "k")
	s.Put("k", "val")

	select {
	case resp := <-ch:
		if len(resp.Events) == 0 {
			t.Error("이벤트가 비어있습니다")
			return
		}
		if resp.Events[0].Type != EventPut {
			t.Errorf("이벤트 타입 = %s, 원하는 값: PUT", resp.Events[0].Type)
		}
		if resp.Events[0].Value != "val" {
			t.Errorf("이벤트 값 = %q, 원하는 값: val", resp.Events[0].Value)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("이벤트가 수신되지 않았습니다")
	}
}

func TestWatch_DeleteEventType(t *testing.T) {
	s := NewWatchableStore()
	s.Put("k", "val")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ch := s.Watch(ctx, "k")
	s.Delete("k")

	select {
	case resp := <-ch:
		if len(resp.Events) == 0 || resp.Events[0].Type != EventDelete {
			t.Errorf("DELETE 이벤트를 받지 못했습니다")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("이벤트가 수신되지 않았습니다")
	}
}

// ============================================================
// Watch 접두사 테스트 (15점)
// ============================================================

func TestWatch_Prefix(t *testing.T) {
	s := NewWatchableStore()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := s.Watch(ctx, "/pods/", WithPrefix())

	s.Put("/pods/pod-1", "running")
	s.Put("/pods/pod-2", "pending")
	s.Put("/services/svc-1", "ignored") // 무시되어야 함
	s.Put("/pods/pod-3", "running")

	var events []WatchEvent
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case resp, ok := <-ch:
			if !ok {
				goto done
			}
			events = append(events, resp.Events...)
			if len(events) >= 3 {
				goto done
			}
		case <-timeout:
			goto done
		}
	}
done:
	if len(events) < 3 {
		t.Errorf("접두사 /pods/ 이벤트 %d개, 원하는 값: 3개", len(events))
		return
	}
	for _, ev := range events {
		if len(ev.Key) < 7 || ev.Key[:7] != "/pods/" {
			t.Errorf("접두사 Watch에서 /pods/ 외 키 수신: %s", ev.Key)
		}
	}
}

func TestWatch_PrefixVsExact(t *testing.T) {
	s := NewWatchableStore()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	exactCh := s.Watch(ctx, "/pods/")          // 정확히 "/pods/" 키만
	prefixCh := s.Watch(ctx, "/pods/", WithPrefix()) // "/pods/"로 시작하는 모든 키

	s.Put("/pods/pod-1", "v")

	// 정확한 키 Watch는 이벤트를 받지 않아야 함
	select {
	case resp := <-exactCh:
		for _, ev := range resp.Events {
			if ev.Key == "/pods/pod-1" {
				t.Error("정확한 Watch가 다른 키의 이벤트를 받았습니다")
			}
		}
	case <-time.After(100 * time.Millisecond):
		// 정상 - 이벤트 없어야 함
	}

	// 접두사 Watch는 이벤트를 받아야 함
	select {
	case resp := <-prefixCh:
		if len(resp.Events) == 0 || resp.Events[0].Key != "/pods/pod-1" {
			t.Error("접두사 Watch가 이벤트를 받지 못했습니다")
		}
	case <-time.After(300 * time.Millisecond):
		t.Error("접두사 Watch 이벤트 타임아웃")
	}
}

// ============================================================
// Watch fromRevision (이벤트 재연) 테스트 (20점)
// ============================================================

func TestWatch_FromRevision(t *testing.T) {
	s := NewWatchableStore()

	// 이벤트 3개 생성
	rev1 := s.Put("k", "v1") // revision 1
	s.Put("k", "v2")          // revision 2
	s.Put("k", "v3")          // revision 3

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// rev1 이후 이벤트 재연
	ch := s.Watch(ctx, "k", WithRevision(rev1+1))

	var events []WatchEvent
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case resp, ok := <-ch:
			if !ok {
				goto done2
			}
			events = append(events, resp.Events...)
			if len(events) >= 2 {
				goto done2
			}
		case <-timeout:
			goto done2
		}
	}
done2:
	if len(events) < 2 {
		t.Errorf("fromRevision=%d 재연 이벤트 %d개, 원하는 값: 2개 이상", rev1+1, len(events))
	}
}

func TestWatch_FromRevisionFuture(t *testing.T) {
	s := NewWatchableStore()
	s.Put("k", "v1")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 미래 리비전부터 Watch - 기존 이벤트 없음, 새 이벤트만 수신
	futureRev := s.CurrentRevision() + 10
	ch := s.Watch(ctx, "k", WithRevision(futureRev))

	s.Put("k", "v2") // 이 이벤트는 futureRev보다 낮으므로 수신 안 됨 (구현 따라 다를 수 있음)
	s.Put("k", "v3") // 실시간 이벤트

	// 최소 1개는 받아야 함 (v3)
	select {
	case resp := <-ch:
		if len(resp.Events) == 0 {
			t.Error("futureRev Watch에서 이벤트가 없습니다")
		}
	case <-time.After(500 * time.Millisecond):
		// futureRev 이전 이벤트를 건너뛰었을 수도 있음 - 허용
	}
}

// ============================================================
// 다중 와처 테스트 (10점)
// ============================================================

func TestWatch_MultipleWatchers(t *testing.T) {
	s := NewWatchableStore()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 같은 키에 3개 Watch
	ch1 := s.Watch(ctx, "shared")
	ch2 := s.Watch(ctx, "shared")
	ch3 := s.Watch(ctx, "shared")

	s.Put("shared", "value")

	received := make([]bool, 3)
	chs := []WatchChan{ch1, ch2, ch3}
	timeout := time.After(500 * time.Millisecond)

	var wg sync.WaitGroup
	for i, ch := range chs {
		wg.Add(1)
		go func(idx int, c WatchChan) {
			defer wg.Done()
			select {
			case resp := <-c:
				if len(resp.Events) > 0 {
					received[idx] = true
				}
			case <-timeout:
			}
		}(i, ch)
	}
	wg.Wait()

	for i, got := range received {
		if !got {
			t.Errorf("와처 %d가 이벤트를 받지 못했습니다", i+1)
		}
	}
}

// ============================================================
// Compact 테스트 (10점)
// ============================================================

func TestCompact(t *testing.T) {
	s := NewWatchableStore()

	s.Put("k", "v1")
	rev2 := s.Put("k", "v2")
	s.Put("k", "v3")

	// rev2까지 압축
	if err := s.Compact(rev2); err != nil {
		t.Fatalf("Compact 오류: %v", err)
	}

	// Compact 후에도 현재 값은 유지되어야 함
	val, _, ok := s.Get("k")
	if !ok || val != "v3" {
		t.Errorf("Compact 후 현재 값 = %q (ok=%v), 원하는 값: v3", val, ok)
	}
}

func TestCompact_FutureRevisionError(t *testing.T) {
	s := NewWatchableStore()
	s.Put("k", "v")

	future := s.CurrentRevision() + 100
	err := s.Compact(future)
	if err == nil {
		t.Error("미래 리비전으로 Compact는 에러를 반환해야 합니다")
	}
}

// ============================================================
// 컨텍스트 취소 테스트 (10점)
// ============================================================

func TestWatch_ContextCancel(t *testing.T) {
	s := NewWatchableStore()
	ctx, cancel := context.WithCancel(context.Background())

	ch := s.Watch(ctx, "k")
	cancel()

	// 채널이 닫혀야 함
	select {
	case _, ok := <-ch:
		if ok {
			// 닫히기 전에 이벤트가 왔을 수 있음, 다시 읽기
			_, ok2 := <-ch
			if ok2 {
				t.Error("ctx 취소 후 채널이 닫히지 않았습니다")
			}
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("ctx 취소 후 채널이 500ms 내에 닫히지 않았습니다")
	}
}

// ============================================================
// 채점 함수 (TestGrade)
// ============================================================

func TestGrade(t *testing.T) {
	score := 0
	total := 100

	fmt.Println("\n" + "═══════════════════════════════════════════════════")
	fmt.Println("  과제 A5: 감시 시스템 채점 결과")
	fmt.Println("  패턴: etcd Watch + MVCC")
	fmt.Println("═══════════════════════════════════════════════════")

	collectEvents := func(ch WatchChan, want int, timeout time.Duration) []WatchEvent {
		var events []WatchEvent
		deadline := time.After(timeout)
		for {
			select {
			case resp, ok := <-ch:
				if !ok {
					return events
				}
				events = append(events, resp.Events...)
				if len(events) >= want {
					return events
				}
			case <-deadline:
				return events
			}
		}
	}

	// 기본 KV + 리비전 (15점)
	t.Run("기본_KV", func(t *testing.T) {
		s := NewWatchableStore()
		r1 := s.Put("a", "1")
		r2 := s.Put("b", "2")
		val, _, ok := s.Get("a")
		delRev := s.Delete("a")
		_, _, okAfter := s.Get("a")

		if r1 > 0 && r2 > r1 && ok && val == "1" && delRev > r2 && !okAfter {
			score += 15
			fmt.Printf("  ✓ 기본 KV + 리비전 단조 증가          15/15점\n")
		} else {
			fmt.Printf("  ✗ 기본 KV (r1=%d,r2=%d,ok=%v,del=%d)   0/15점\n", r1, r2, ok, delRev)
		}
	})

	// Watch 단일 키 (20점)
	t.Run("Watch_단일키", func(t *testing.T) {
		s := NewWatchableStore()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		ch := s.Watch(ctx, "watch-k")
		s.Put("watch-k", "v1")
		s.Put("other", "ignored")
		s.Put("watch-k", "v2")

		events := collectEvents(ch, 2, 500*time.Millisecond)
		ok := len(events) >= 2
		for _, ev := range events {
			if ev.Key != "watch-k" {
				ok = false
			}
		}
		if ok {
			score += 20
			fmt.Printf("  ✓ Watch 단일 키                       20/20점\n")
		} else {
			fmt.Printf("  ✗ Watch 단일 키 (events=%d)             0/20점\n", len(events))
		}
	})

	// Watch 접두사 (15점)
	t.Run("Watch_접두사", func(t *testing.T) {
		s := NewWatchableStore()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		ch := s.Watch(ctx, "/ns/", WithPrefix())
		s.Put("/ns/a", "1")
		s.Put("/ns/b", "2")
		s.Put("/other/c", "3")

		events := collectEvents(ch, 2, 500*time.Millisecond)
		ok := len(events) >= 2
		for _, ev := range events {
			if len(ev.Key) < 4 || ev.Key[:4] != "/ns/" {
				ok = false
			}
		}
		if ok {
			score += 15
			fmt.Printf("  ✓ Watch 접두사                        15/15점\n")
		} else {
			fmt.Printf("  ✗ Watch 접두사 (events=%d)              0/15점\n", len(events))
		}
	})

	// Watch fromRevision (20점)
	t.Run("Watch_fromRevision", func(t *testing.T) {
		s := NewWatchableStore()
		rev1 := s.Put("rk", "v1")
		s.Put("rk", "v2")
		s.Put("rk", "v3")

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		ch := s.Watch(ctx, "rk", WithRevision(rev1+1))
		events := collectEvents(ch, 2, 500*time.Millisecond)
		if len(events) >= 2 {
			score += 20
			fmt.Printf("  ✓ Watch fromRevision (재연)           20/20점\n")
		} else {
			fmt.Printf("  ✗ Watch fromRevision (events=%d)        0/20점\n", len(events))
		}
	})

	// 다중 와처 (10점)
	t.Run("다중_와처", func(t *testing.T) {
		s := NewWatchableStore()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		ch1 := s.Watch(ctx, "mk")
		ch2 := s.Watch(ctx, "mk")
		s.Put("mk", "val")

		e1 := collectEvents(ch1, 1, 300*time.Millisecond)
		e2 := collectEvents(ch2, 1, 300*time.Millisecond)

		if len(e1) > 0 && len(e2) > 0 {
			score += 10
			fmt.Printf("  ✓ 다중 와처                           10/10점\n")
		} else {
			fmt.Printf("  ✗ 다중 와처 (ch1=%d,ch2=%d)             0/10점\n", len(e1), len(e2))
		}
	})

	// Compact (10점)
	t.Run("Compact", func(t *testing.T) {
		s := NewWatchableStore()
		s.Put("ck", "v1")
		rev2 := s.Put("ck", "v2")
		s.Put("ck", "v3")

		err := s.Compact(rev2)
		val, _, ok := s.Get("ck")
		futErr := s.Compact(s.CurrentRevision() + 100)

		if err == nil && ok && val == "v3" && futErr != nil {
			score += 10
			fmt.Printf("  ✓ Compact                             10/10점\n")
		} else {
			fmt.Printf("  ✗ Compact (err=%v,val=%q,futErr=%v)     0/10점\n", err, val, futErr)
		}
	})

	fmt.Println("───────────────────────────────────────────────────")
	fmt.Printf("  최종 점수: %d / %d점\n", score, total)
	fmt.Printf("  (ctx 취소 10점은 TestWatch_ContextCancel로 확인)\n")

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
