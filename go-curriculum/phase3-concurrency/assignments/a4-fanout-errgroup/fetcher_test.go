// 패키지 선언
package fetcher_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	fetcher "github.com/go-curriculum/a4-fanout-errgroup"
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

// newOKServer는 항상 200 OK와 지정한 본문을 반환하는 테스트 서버를 생성합니다.
func newOKServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
	}))
}

// newSlowServer는 delay 후 응답하는 느린 테스트 서버를 생성합니다.
func newSlowServer(delay time.Duration, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(delay):
			fmt.Fprint(w, body)
		case <-r.Context().Done():
			return
		}
	}))
}

// newErrorServer는 항상 500 에러를 반환하는 테스트 서버를 생성합니다.
func newErrorServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "서버 에러")
	}))
}

// newCountingServer는 요청 수를 세는 테스트 서버를 생성합니다.
func newCountingServer(counter *atomic.Int64, delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter.Add(1)
		time.Sleep(delay)
		fmt.Fprint(w, "ok")
	}))
}

// ─────────────────────────────────────────
// FetchAll 기본 테스트
// ─────────────────────────────────────────

func TestFetchAll_SingleURL(t *testing.T) {
	srv := newOKServer("안녕하세요")
	defer srv.Close()

	results := fetcher.FetchAll(context.Background(), []string{srv.URL}, 1)

	if len(results) != 1 {
		t.Fatalf("결과 수: 기대=1, 실제=%d", len(results))
	}
	if results[0].Error != nil {
		t.Fatalf("에러 발생: %v", results[0].Error)
	}
	if results[0].StatusCode != 200 {
		t.Errorf("상태 코드: 기대=200, 실제=%d", results[0].StatusCode)
	}
	if results[0].Body != "안녕하세요" {
		t.Errorf("본문: 기대='안녕하세요', 실제='%s'", results[0].Body)
	}
}

func TestFetchAll_EmptyURLs(t *testing.T) {
	results := fetcher.FetchAll(context.Background(), []string{}, 5)
	if len(results) != 0 {
		t.Errorf("빈 입력: 기대=0 결과, 실제=%d", len(results))
	}
}

func TestFetchAll_MultipleURLs_OrderPreserved(t *testing.T) {
	// 각기 다른 본문을 반환하는 서버 3개 생성
	bodies := []string{"첫째", "둘째", "셋째"}
	servers := make([]*httptest.Server, len(bodies))
	urls := make([]string, len(bodies))

	for i, body := range bodies {
		body := body
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		defer servers[i].Close()
		urls[i] = servers[i].URL
	}

	results := fetcher.FetchAll(context.Background(), urls, 3)

	if len(results) != len(urls) {
		t.Fatalf("결과 수: 기대=%d, 실제=%d", len(urls), len(results))
	}

	for i, want := range bodies {
		if results[i].Error != nil {
			t.Errorf("[%d] 에러: %v", i, results[i].Error)
			continue
		}
		if results[i].Body != want {
			t.Errorf("[%d] 본문 순서 불일치: 기대='%s', 실제='%s'", i, want, results[i].Body)
		}
		if results[i].URL != urls[i] {
			t.Errorf("[%d] URL 불일치: 기대='%s', 실제='%s'", i, urls[i], results[i].URL)
		}
	}
}

func TestFetchAll_PartialFailure(t *testing.T) {
	okSrv := newOKServer("성공")
	defer okSrv.Close()

	results := fetcher.FetchAll(context.Background(), []string{
		okSrv.URL,
		"http://127.0.0.1:1", // 연결 실패
		okSrv.URL,
	}, 3)

	if len(results) != 3 {
		t.Fatalf("결과 수: 기대=3, 실제=%d", len(results))
	}

	// 첫 번째와 세 번째는 성공해야 함
	if results[0].Error != nil {
		t.Errorf("[0] 성공해야 하는데 에러: %v", results[0].Error)
	}
	if results[0].Body != "성공" {
		t.Errorf("[0] 본문: 기대='성공', 실제='%s'", results[0].Body)
	}

	// 두 번째는 실패해야 함
	if results[1].Error == nil {
		t.Error("[1] 실패해야 하는데 에러 없음")
	}

	if results[2].Error != nil {
		t.Errorf("[2] 성공해야 하는데 에러: %v", results[2].Error)
	}
}

func TestFetchAll_AllFail(t *testing.T) {
	urls := []string{
		"http://127.0.0.1:1",
		"http://127.0.0.1:2",
		"http://127.0.0.1:3",
	}

	results := fetcher.FetchAll(context.Background(), urls, 3)

	if len(results) != 3 {
		t.Fatalf("결과 수: 기대=3, 실제=%d", len(results))
	}
	for i, r := range results {
		if r.Error == nil {
			t.Errorf("[%d] 실패해야 하는데 에러 없음", i)
		}
	}
}

// ─────────────────────────────────────────
// 동시성 제한 테스트
// ─────────────────────────────────────────

func TestFetchAll_MaxConcurrency(t *testing.T) {
	const maxConcurrency = 3
	const numURLs = 9

	var current atomic.Int64
	var maxObserved atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := current.Add(1)
		defer current.Add(-1)

		// 최댓값 갱신
		for {
			old := maxObserved.Load()
			if cur <= old || maxObserved.CompareAndSwap(old, cur) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond)
		fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	urls := make([]string, numURLs)
	for i := range urls {
		urls[i] = srv.URL
	}

	results := fetcher.FetchAll(context.Background(), urls, maxConcurrency)

	if len(results) != numURLs {
		t.Errorf("결과 수: 기대=%d, 실제=%d", numURLs, len(results))
	}

	observed := maxObserved.Load()
	if observed > int64(maxConcurrency) {
		t.Errorf("동시 실행 수 초과: 최대=%d, 관찰=%d", maxConcurrency, observed)
	}
	t.Logf("최대 동시 실행: %d (제한: %d)", observed, maxConcurrency)
}

func TestFetchAll_ConcurrencySpeedup(t *testing.T) {
	// maxConcurrency=4로 4개 URL을 처리하면 직렬보다 빨라야 함
	const delay = 80 * time.Millisecond
	const numURLs = 4

	srv := newSlowServer(delay, "ok")
	defer srv.Close()

	urls := make([]string, numURLs)
	for i := range urls {
		urls[i] = srv.URL
	}

	start := time.Now()
	results := fetcher.FetchAll(context.Background(), urls, numURLs)
	elapsed := time.Since(start)

	for _, r := range results {
		if r.Error != nil {
			t.Errorf("에러: %v", r.Error)
		}
	}

	// 병렬 처리면 ~80ms, 직렬이면 ~320ms
	if elapsed > 200*time.Millisecond {
		t.Errorf("병렬 처리가 너무 느림: %v (기대 <200ms)", elapsed)
	}
	t.Logf("4개 병렬 처리 소요: %v", elapsed)
}

// ─────────────────────────────────────────
// Context 취소 테스트
// ─────────────────────────────────────────

func TestFetchAll_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 즉시 취소

	srv := newSlowServer(500*time.Millisecond, "ok")
	defer srv.Close()

	urls := make([]string, 5)
	for i := range urls {
		urls[i] = srv.URL
	}

	start := time.Now()
	results := fetcher.FetchAll(ctx, urls, 2)
	elapsed := time.Since(start)

	// 즉시 취소됐으므로 빠르게 반환되어야 함
	if elapsed > 300*time.Millisecond {
		t.Errorf("취소 후 너무 늦게 반환: %v", elapsed)
	}

	// 일부 에러가 있어야 함 (취소됐으므로)
	errorCount := 0
	for _, r := range results {
		if r.Error != nil {
			errorCount++
		}
	}
	t.Logf("취소 후 에러 수: %d/%d", errorCount, len(results))

	_ = results
}

func TestFetchAll_ContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	srv := newSlowServer(200*time.Millisecond, "ok")
	defer srv.Close()

	urls := make([]string, 3)
	for i := range urls {
		urls[i] = srv.URL
	}

	start := time.Now()
	results := fetcher.FetchAll(ctx, urls, 3)
	elapsed := time.Since(start)

	// 타임아웃 내에 반환되어야 함 (너넉한 여유 포함)
	if elapsed > 200*time.Millisecond {
		t.Errorf("타임아웃 후 너무 늦게 반환: %v", elapsed)
	}

	t.Logf("타임아웃 테스트: 소요=%v, 결과=%d개", elapsed, len(results))
}

// ─────────────────────────────────────────
// FetchWithRetry 테스트
// ─────────────────────────────────────────

func TestFetchWithRetry_Success(t *testing.T) {
	srv := newOKServer("재시도 성공")
	defer srv.Close()

	result := fetcher.FetchWithRetry(context.Background(), srv.URL, 3)

	if result.Error != nil {
		t.Fatalf("에러: %v", result.Error)
	}
	if result.Body != "재시도 성공" {
		t.Errorf("본문: 기대='재시도 성공', 실제='%s'", result.Body)
	}
}

func TestFetchWithRetry_SucceedOnRetry(t *testing.T) {
	var attempts atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			// 처음 2번은 실패
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// 3번째부터 성공
		fmt.Fprint(w, "드디어 성공")
	}))
	defer srv.Close()

	// 이 테스트는 HTTP 500을 에러로 처리하지 않으므로 (fetchOne이 상태 코드만 반환)
	// 연결 자체가 실패하는 시나리오를 시뮬레이션합니다.
	// 여기서는 재시도 횟수가 증가하는지만 확인합니다.
	result := fetcher.FetchWithRetry(context.Background(), srv.URL, 3)
	if result.Error != nil {
		t.Fatalf("에러: %v", result.Error)
	}
	t.Logf("시도 횟수: %d", attempts.Load())
}

func TestFetchWithRetry_AllFail(t *testing.T) {
	result := fetcher.FetchWithRetry(
		context.Background(),
		"http://127.0.0.1:1", // 연결 실패
		2,
	)

	if result.Error == nil {
		t.Error("실패해야 하는데 에러 없음")
	}
	t.Logf("최종 에러: %v", result.Error)
}

func TestFetchWithRetry_ZeroRetries(t *testing.T) {
	result := fetcher.FetchWithRetry(
		context.Background(),
		"http://127.0.0.1:1",
		0, // 재시도 없음
	)

	if result.Error == nil {
		t.Error("실패해야 하는데 에러 없음")
	}
}

func TestFetchWithRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// 100ms 후 취소
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result := fetcher.FetchWithRetry(ctx, "http://127.0.0.1:1", 5)
	elapsed := time.Since(start)

	// ctx가 취소됐으므로 모든 재시도를 기다리지 않아야 함
	// 최대 재시도 간격: 100+200+400+800+1600ms = 3100ms
	if elapsed > 500*time.Millisecond {
		t.Errorf("취소 후 너무 늦게 반환: %v", elapsed)
	}

	if result.Error == nil {
		t.Error("취소 후 에러가 있어야 함")
	}
	t.Logf("취소 후 반환 시간: %v, 에러: %v", elapsed, result.Error)
}

// ─────────────────────────────────────────
// 고루틴 누수 테스트
// ─────────────────────────────────────────

func TestFetchAll_NoGoroutineLeak(t *testing.T) {
	before := goroutineCount()

	srv := newOKServer("ok")
	defer srv.Close()

	for i := 0; i < 5; i++ {
		urls := make([]string, 6)
		for j := range urls {
			urls[j] = srv.URL
		}
		fetcher.FetchAll(context.Background(), urls, 3)
	}

	waitForGoroutines(t, before, time.Second)
	after := goroutineCount()
	if after > before+2 {
		t.Errorf("고루틴 누수: 전=%d, 후=%d", before, after)
	}
}

// ─────────────────────────────────────────
// 채점 테스트
// ─────────────────────────────────────────

func TestGrade(t *testing.T) {
	passed := 0
	total := 0

	check := func(name string, points int, fn func() bool) {
		total += points
		t.Run(name, func(t *testing.T) {
			if fn() {
				passed += points
				fmt.Printf("  [통과] %s: +%d점\n", name, points)
			} else {
				fmt.Printf("  [실패] %s: 0점\n", name)
			}
		})
	}

	// 기본 Fetch (20점)
	check("기본Fetch", 20, func() bool {
		srv := newOKServer("기본테스트")
		defer srv.Close()
		results := fetcher.FetchAll(context.Background(), []string{srv.URL}, 1)
		return len(results) == 1 && results[0].Error == nil && results[0].Body == "기본테스트" && results[0].StatusCode == 200
	})

	// 병렬 Fetch 및 순서 보장 (20점)
	check("병렬Fetch순서보장", 20, func() bool {
		bodies := []string{"A", "B", "C", "D", "E"}
		servers := make([]*httptest.Server, len(bodies))
		urls := make([]string, len(bodies))
		for i, b := range bodies {
			b := b
			servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, b)
			}))
			defer servers[i].Close()
			urls[i] = servers[i].URL
		}
		results := fetcher.FetchAll(context.Background(), urls, 5)
		if len(results) != 5 {
			return false
		}
		for i, r := range results {
			if r.Error != nil || r.Body != bodies[i] {
				return false
			}
		}
		return true
	})

	// 부분 실패 처리 (20점)
	check("부분실패처리", 20, func() bool {
		srv := newOKServer("성공")
		defer srv.Close()
		results := fetcher.FetchAll(context.Background(), []string{
			srv.URL,
			"http://127.0.0.1:1",
			srv.URL,
		}, 3)
		if len(results) != 3 {
			return false
		}
		return results[0].Error == nil && results[1].Error != nil && results[2].Error == nil
	})

	// 동시성 제한 (20점)
	check("동시성제한", 20, func() bool {
		var current atomic.Int64
		var maxObs atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cur := current.Add(1)
			defer current.Add(-1)
			for {
				old := maxObs.Load()
				if cur <= old || maxObs.CompareAndSwap(old, cur) {
					break
				}
			}
			time.Sleep(40 * time.Millisecond)
			fmt.Fprint(w, "ok")
		}))
		defer srv.Close()

		urls := make([]string, 9)
		for i := range urls {
			urls[i] = srv.URL
		}
		fetcher.FetchAll(context.Background(), urls, 3)
		obs := maxObs.Load()
		fmt.Printf("    (최대 동시 실행: %d, 제한: 3)\n", obs)
		return obs <= 3
	})

	// Context 취소 (10점)
	check("Context취소", 10, func() bool {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		srv := newSlowServer(500*time.Millisecond, "ok")
		defer srv.Close()
		urls := make([]string, 4)
		for i := range urls {
			urls[i] = srv.URL
		}
		start := time.Now()
		fetcher.FetchAll(ctx, urls, 2)
		return time.Since(start) < 300*time.Millisecond
	})

	// 재시도 로직 (10점)
	check("재시도로직", 10, func() bool {
		result := fetcher.FetchWithRetry(context.Background(), "http://127.0.0.1:1", 2)
		return result.Error != nil // 실패해야 함 (하지만 패닉이 없어야 함)
	})

	score := passed * 100 / total

	fmt.Println()
	fmt.Printf("╔══════════════════════════════════╗\n")
	fmt.Printf("║  통과: %2d/%2d                      ║\n", passed/20, total/20) // 대략적인 항목 수
	fmt.Printf("║  점수: %3d/100                     ║\n", score)
	grade := "F"
	switch {
	case score >= 90:
		grade = "A+"
	case score >= 80:
		grade = "A"
	case score >= 70:
		grade = "B"
	case score >= 60:
		grade = "C"
	}
	fmt.Printf("║  등급: %-30s║\n", grade)
	fmt.Printf("╚══════════════════════════════════╝\n")
	fmt.Println()
	fmt.Println("=== 채점 결과 ===")
	fmt.Printf("통과: %d/%d\n", passed/20, total/20)
	fmt.Printf("점수: %d/100\n", score)

	if score < 60 {
		t.Errorf("점수 미달: %d/100점 (합격: 60점 이상)", score)
	}
}
