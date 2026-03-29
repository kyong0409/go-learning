// 패키지 선언
package graceful_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gs "github.com/go-curriculum/a6-graceful-server"
)

// ─────────────────────────────────────────
// 헬퍼
// ─────────────────────────────────────────

func goroutineCount() int { return runtime.NumGoroutine() }

func waitForGoroutines(t *testing.T, before int, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// startTestServer는 서버를 시작하고 준비될 때까지 기다립니다.
func startTestServer(t *testing.T, s *gs.Server) (context.CancelFunc, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start(ctx)
	}()
	// 서버가 준비될 때까지 대기
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if s.Addr() != "" && s.Addr() != ":0" {
			resp, err := http.Get("http://" + s.Addr() + "/healthz")
			if err == nil {
				resp.Body.Close()
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Cleanup(cancel)
	return cancel, errCh
}

// get은 HTTP GET 요청을 수행합니다.
func get(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	return client.Get(url)
}

// ─────────────────────────────────────────
// 기본 시작/종료 테스트
// ─────────────────────────────────────────

func TestServer_StartAndShutdown(t *testing.T) {
	s := gs.NewServer(":0")
	cancel, errCh := startTestServer(t, s)

	addr := s.Addr()
	if addr == "" {
		t.Fatal("Addr()이 빈 문자열")
	}
	t.Logf("서버 주소: %s", addr)

	// 서버가 실제로 응답하는지 확인
	resp, err := get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("healthz 요청 실패: %v", err)
	}
	resp.Body.Close()

	// 종료
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Start 반환 에러: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("종료 타임아웃")
	}
}

func TestServer_Addr_DynamicPort(t *testing.T) {
	s := gs.NewServer(":0")
	cancel, _ := startTestServer(t, s)
	defer cancel()

	addr := s.Addr()
	if addr == "" || addr == ":0" {
		t.Errorf("동적 포트 할당 실패: addr=%q", addr)
	}
	if !strings.Contains(addr, ":") {
		t.Errorf("포트가 없는 주소: %q", addr)
	}
	t.Logf("동적 포트: %s", addr)
}

// ─────────────────────────────────────────
// 헬스 체크 테스트
// ─────────────────────────────────────────

func TestHealthz_OK(t *testing.T) {
	s := gs.NewServer(":0")
	cancel, _ := startTestServer(t, s)
	defer cancel()

	resp, err := get("http://" + s.Addr() + "/healthz")
	if err != nil {
		t.Fatalf("healthz 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("상태 코드: 기대=200, 실제=%d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("JSON 파싱 실패: %v (본문: %s)", err, body)
	}
	if result["status"] != "ok" {
		t.Errorf("status: 기대='ok', 실제='%s'", result["status"])
	}
}

func TestHealthz_ContentType(t *testing.T) {
	s := gs.NewServer(":0")
	cancel, _ := startTestServer(t, s)
	defer cancel()

	resp, err := get("http://" + s.Addr() + "/healthz")
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: 기대=application/json 포함, 실제=%q", ct)
	}
}

// ─────────────────────────────────────────
// 핸들러 등록 테스트
// ─────────────────────────────────────────

func TestHandle_Basic(t *testing.T) {
	s := gs.NewServer(":0")
	s.Handle("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "안녕하세요")
	}))

	cancel, _ := startTestServer(t, s)
	defer cancel()

	resp, err := get("http://" + s.Addr() + "/hello")
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "안녕하세요" {
		t.Errorf("본문: 기대='안녕하세요', 실제='%s'", body)
	}
}

func TestHandle_MultipleRoutes(t *testing.T) {
	s := gs.NewServer(":0")
	s.Handle("/a", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "A") }))
	s.Handle("/b", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "B") }))
	s.Handle("/c", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "C") }))

	cancel, _ := startTestServer(t, s)
	defer cancel()

	for _, tc := range []struct{ path, want string }{{"/a", "A"}, {"/b", "B"}, {"/c", "C"}} {
		resp, err := get("http://" + s.Addr() + tc.path)
		if err != nil {
			t.Fatalf("%s 요청 실패: %v", tc.path, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if string(body) != tc.want {
			t.Errorf("%s: 기대='%s', 실제='%s'", tc.path, tc.want, body)
		}
	}
}

// ─────────────────────────────────────────
// 미들웨어 테스트
// ─────────────────────────────────────────

func TestMiddleware_SingleMiddleware(t *testing.T) {
	s := gs.NewServer(":0")

	// 헤더를 추가하는 미들웨어
	s.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "middleware-applied")
			next.ServeHTTP(w, r)
		})
	})

	s.Handle("/mw", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))

	cancel, _ := startTestServer(t, s)
	defer cancel()

	resp, err := get("http://" + s.Addr() + "/mw")
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Test") != "middleware-applied" {
		t.Errorf("미들웨어 헤더 없음: X-Test=%q", resp.Header.Get("X-Test"))
	}
}

func TestMiddleware_OrderPreserved(t *testing.T) {
	s := gs.NewServer(":0")

	var order []string
	var mu sync.Mutex

	addOrder := func(label string) {
		mu.Lock()
		order = append(order, label)
		mu.Unlock()
	}

	// A가 먼저 등록, B가 나중 → 순서: A → B → 핸들러
	s.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			addOrder("A")
			next.ServeHTTP(w, r)
		})
	})
	s.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			addOrder("B")
			next.ServeHTTP(w, r)
		})
	})

	s.Handle("/order", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addOrder("H")
		fmt.Fprint(w, "ok")
	}))

	cancel, _ := startTestServer(t, s)
	defer cancel()

	resp, err := get("http://" + s.Addr() + "/order")
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	resp.Body.Close()

	mu.Lock()
	got := strings.Join(order, "→")
	mu.Unlock()

	if got != "A→B→H" {
		t.Errorf("미들웨어 순서: 기대=A→B→H, 실제=%s", got)
	}
}

func TestMiddleware_AppliesToAllRoutes(t *testing.T) {
	s := gs.NewServer(":0")

	var count atomic.Int64
	s.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/healthz" {
				count.Add(1)
			}
			next.ServeHTTP(w, r)
		})
	})

	s.Handle("/r1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "1") }))
	s.Handle("/r2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "2") }))

	cancel, _ := startTestServer(t, s)
	defer cancel()

	get("http://" + s.Addr() + "/r1")
	get("http://" + s.Addr() + "/r2")

	if count.Load() != 2 {
		t.Errorf("미들웨어 적용 횟수: 기대=2, 실제=%d", count.Load())
	}
}

// ─────────────────────────────────────────
// 우아한 종료 테스트
// ─────────────────────────────────────────

func TestGracefulShutdown_DrainRequests(t *testing.T) {
	s := gs.NewServer(":0")

	requestStarted := make(chan struct{})
	requestCanFinish := make(chan struct{})

	s.Handle("/slow", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-requestCanFinish
		fmt.Fprint(w, "완료")
	}))

	cancel, errCh := startTestServer(t, s)
	_ = cancel

	// 느린 요청 시작
	respCh := make(chan *http.Response, 1)
	go func() {
		resp, err := get("http://" + s.Addr() + "/slow")
		if err == nil {
			respCh <- resp
		} else {
			respCh <- nil
		}
	}()

	// 요청이 시작될 때까지 대기
	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("요청 시작 타임아웃")
	}

	// 서버 종료 시작 (별도 고루틴)
	shutdownDone := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		shutdownDone <- s.Shutdown(ctx)
	}()

	// 잠시 후 요청 완료 허용
	time.Sleep(50 * time.Millisecond)
	close(requestCanFinish)

	// 요청 응답 확인
	select {
	case resp := <-respCh:
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if string(body) != "완료" {
				t.Errorf("우아한 종료 중 응답: 기대='완료', 실제='%s'", body)
			}
		}
	case <-time.After(2 * time.Second):
		t.Error("요청 응답 타임아웃")
	}

	// 종료 완료 확인
	select {
	case err := <-shutdownDone:
		if err != nil {
			t.Errorf("Shutdown 에러: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Shutdown 타임아웃")
	}

	<-errCh
}

// ─────────────────────────────────────────
// OnShutdown 훅 테스트
// ─────────────────────────────────────────

func TestOnShutdown_HookCalled(t *testing.T) {
	s := gs.NewServer(":0")

	var hookCalled atomic.Bool
	s.OnShutdown(func() {
		hookCalled.Store(true)
	})

	cancel, errCh := startTestServer(t, s)

	cancel()
	<-errCh

	time.Sleep(100 * time.Millisecond)
	if !hookCalled.Load() {
		t.Error("OnShutdown 훅이 호출되지 않음")
	}
}

func TestOnShutdown_MultipleHooks(t *testing.T) {
	s := gs.NewServer(":0")

	var order []string
	var mu sync.Mutex

	for _, label := range []string{"hook1", "hook2", "hook3"} {
		label := label
		s.OnShutdown(func() {
			mu.Lock()
			order = append(order, label)
			mu.Unlock()
		})
	}

	cancel, errCh := startTestServer(t, s)
	cancel()
	<-errCh

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	got := strings.Join(order, ",")
	mu.Unlock()

	if len(order) != 3 {
		t.Errorf("훅 호출 수: 기대=3, 실제=%d (%s)", len(order), got)
	}
	t.Logf("훅 실행 순서: %s", got)
}

// ─────────────────────────────────────────
// 고루틴 누수 테스트
// ─────────────────────────────────────────

func TestServer_NoGoroutineLeak(t *testing.T) {
	before := goroutineCount()

	for i := 0; i < 3; i++ {
		s := gs.NewServer(":0")
		s.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "pong")
		}))

		cancel, errCh := startTestServer(t, s)

		// 몇 개 요청
		get("http://" + s.Addr() + "/ping")
		get("http://" + s.Addr() + "/healthz")

		cancel()
		<-errCh
	}

	waitForGoroutines(t, before, 2*time.Second)
	after := goroutineCount()
	if after > before+3 {
		t.Errorf("고루틴 누수: 전=%d, 후=%d", before, after)
	}
}

// ─────────────────────────────────────────
// 채점 테스트
// ─────────────────────────────────────────

func TestGrade(t *testing.T) {
	passedScore := 0

	check := func(name string, points int, fn func(t *testing.T) bool) {
		t.Run(name, func(t *testing.T) {
			if fn(t) {
				passedScore += points
				fmt.Printf("  [통과] %s: +%d점\n", name, points)
			} else {
				fmt.Printf("  [실패] %s: 0점\n", name)
			}
		})
	}

	// 기본 서버 시작/종료 (20점)
	check("기본시작종료", 20, func(t *testing.T) bool {
		s := gs.NewServer(":0")
		cancel, errCh := startTestServer(t, s)

		addr := s.Addr()
		if addr == "" || addr == ":0" {
			return false
		}

		resp, err := get("http://" + addr + "/healthz")
		if err != nil {
			return false
		}
		resp.Body.Close()

		cancel()
		select {
		case <-errCh:
			return true
		case <-time.After(2 * time.Second):
			return false
		}
	})

	// 헬스 체크 (15점)
	check("헬스체크", 15, func(t *testing.T) bool {
		s := gs.NewServer(":0")
		cancel, _ := startTestServer(t, s)
		defer cancel()

		resp, err := get("http://" + s.Addr() + "/healthz")
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var result map[string]string
		if err := json.Unmarshal(body, &result); err != nil {
			return false
		}
		return resp.StatusCode == 200 && result["status"] == "ok"
	})

	// 미들웨어 체인 (20점)
	check("미들웨어체인", 20, func(t *testing.T) bool {
		s := gs.NewServer(":0")

		var called []string
		var mu sync.Mutex

		s.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/healthz" {
					mu.Lock()
					called = append(called, "MW1")
					mu.Unlock()
				}
				next.ServeHTTP(w, r)
			})
		})
		s.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/healthz" {
					mu.Lock()
					called = append(called, "MW2")
					mu.Unlock()
				}
				next.ServeHTTP(w, r)
			})
		})
		s.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "ok")
		}))

		cancel, _ := startTestServer(t, s)
		defer cancel()

		get("http://" + s.Addr() + "/test")

		mu.Lock()
		result := strings.Join(called, ",")
		mu.Unlock()

		return result == "MW1,MW2"
	})

	// 우아한 종료 (20점)
	check("우아한종료", 20, func(t *testing.T) bool {
		s := gs.NewServer(":0")

		started := make(chan struct{})
		canFinish := make(chan struct{})

		s.Handle("/slow", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			<-canFinish
			fmt.Fprint(w, "done")
		}))

		cancel, errCh := startTestServer(t, s)
		_ = cancel

		respCh := make(chan string, 1)
		go func() {
			resp, err := get("http://" + s.Addr() + "/slow")
			if err != nil {
				respCh <- ""
				return
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			respCh <- string(body)
		}()

		<-started

		shutDone := make(chan error, 1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			shutDone <- s.Shutdown(ctx)
		}()

		time.Sleep(30 * time.Millisecond)
		close(canFinish)

		body := <-respCh
		<-shutDone
		<-errCh

		return body == "done"
	})

	// OnShutdown 훅 (15점)
	check("OnShutdown훅", 15, func(t *testing.T) bool {
		s := gs.NewServer(":0")
		var called atomic.Bool
		s.OnShutdown(func() { called.Store(true) })

		cancel, errCh := startTestServer(t, s)
		cancel()
		<-errCh
		time.Sleep(100 * time.Millisecond)
		return called.Load()
	})

	// 고루틴 누수 없음 (10점)
	check("고루틴누수없음", 10, func(t *testing.T) bool {
		before := goroutineCount()

		for i := 0; i < 2; i++ {
			s := gs.NewServer(":0")
			cancel, errCh := startTestServer(t, s)
			get("http://" + s.Addr() + "/healthz")
			cancel()
			<-errCh
		}

		waitForGoroutines(t, before, 2*time.Second)
		after := goroutineCount()
		return after <= before+3
	})

	fmt.Println()
	fmt.Printf("╔══════════════════════════════════╗\n")
	fmt.Printf("║  점수: %3d/100                     ║\n", passedScore)
	grade := "F"
	switch {
	case passedScore >= 90:
		grade = "A+"
	case passedScore >= 80:
		grade = "A"
	case passedScore >= 70:
		grade = "B"
	case passedScore >= 60:
		grade = "C"
	}
	fmt.Printf("║  등급: %-30s║\n", grade)
	fmt.Printf("╚══════════════════════════════════╝\n")
	fmt.Println()
	fmt.Println("=== 채점 결과 ===")
	fmt.Printf("통과: %d/%d\n", passedScore/20, 6) // 대략적인 항목 수
	fmt.Printf("점수: %d/100\n", passedScore)

	if passedScore < 60 {
		t.Errorf("점수 미달: %d/100점 (합격: 60점 이상)", passedScore)
	}
}
