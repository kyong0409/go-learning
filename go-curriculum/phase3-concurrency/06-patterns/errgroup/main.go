// 패키지 선언
package main

// 동시성 패턴: errgroup
//
// golang.org/x/sync/errgroup는 여러 고루틴의 에러를 수집하고
// 하나라도 실패하면 나머지를 취소하는 패키지입니다.
//
// sync.WaitGroup + 에러 처리 + Context 취소를 한 번에 제공합니다.
//
// 핵심 API:
// - errgroup.WithContext(ctx): 그룹과 파생 Context 생성
// - g.Go(func() error): 고루틴 시작
// - g.Wait(): 모든 고루틴 완료 대기, 첫 번째 에러 반환

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	"golang.org/x/sync/errgroup"
)

// ─────────────────────────────────────────
// 1. 기본 errgroup 사용
// ─────────────────────────────────────────

// fetchURL은 URL에서 데이터를 가져옵니다.
func fetchURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s 요청 실패: %w", url, err)
	}
	defer resp.Body.Close()

	return fmt.Sprintf("%s → 상태: %d", url, resp.StatusCode), nil
}

func basicErrgroupDemo(serverURL string) {
	fmt.Println("\n--- 1. 기본 errgroup 사용 ---")

	// errgroup.WithContext: Context가 있는 그룹
	// 하나라도 에러 반환하면 ctx가 자동으로 취소됨
	g, ctx := errgroup.WithContext(context.Background())

	urls := []string{
		serverURL + "/fast",
		serverURL + "/slow",
		serverURL + "/fast",
	}

	results := make([]string, len(urls))

	for i, url := range urls {
		i, url := i, url // 루프 변수 캡처
		g.Go(func() error {
			result, err := fetchURL(ctx, url)
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}

	// 모든 고루틴 완료 대기
	if err := g.Wait(); err != nil {
		fmt.Printf("  에러 발생: %v\n", err)
		return
	}

	fmt.Println("  모든 요청 성공:")
	for _, r := range results {
		fmt.Printf("    %s\n", r)
	}
}

// ─────────────────────────────────────────
// 2. 에러 전파와 자동 취소
// ─────────────────────────────────────────

// processItem은 아이템을 처리하고 가끔 에러를 반환합니다.
func processItem(ctx context.Context, id int) error {
	delay := time.Duration(rand.Intn(200)) * time.Millisecond

	select {
	case <-time.After(delay):
		// 3번 아이템은 에러 반환
		if id == 3 {
			return fmt.Errorf("아이템 #%d 처리 실패: 데이터 손상", id)
		}
		fmt.Printf("  아이템 #%d 처리 완료 (%.0fms)\n",
			id, float64(delay.Milliseconds()))
		return nil
	case <-ctx.Done():
		fmt.Printf("  아이템 #%d 취소됨 (ctx: %v)\n", id, ctx.Err())
		return ctx.Err()
	}
}

func errorPropagationDemo() {
	fmt.Println("\n--- 2. 에러 전파와 자동 취소 ---")
	fmt.Println("  아이템 #3이 에러를 반환하면 나머지 취소:")

	g, ctx := errgroup.WithContext(context.Background())

	for i := 1; i <= 6; i++ {
		i := i
		g.Go(func() error {
			return processItem(ctx, i)
		})
	}

	err := g.Wait()
	if err != nil {
		fmt.Printf("  그룹 에러: %v\n", err)
		fmt.Println("  (errgroup은 첫 번째 에러만 반환)")
	}
}

// ─────────────────────────────────────────
// 3. 결과 수집 패턴
// ─────────────────────────────────────────

// APIResponse는 API 응답을 나타냅니다.
type APIResponse struct {
	ServiceName string
	Data        string
	Latency     time.Duration
}

// callService는 마이크로서비스를 호출합니다.
func callService(ctx context.Context, name string, delay time.Duration) (APIResponse, error) {
	select {
	case <-time.After(delay):
		return APIResponse{
			ServiceName: name,
			Data:        fmt.Sprintf("%s 응답 데이터", name),
			Latency:     delay,
		}, nil
	case <-ctx.Done():
		return APIResponse{}, fmt.Errorf("%s 서비스 취소: %w", name, ctx.Err())
	}
}

func resultCollectionDemo() {
	fmt.Println("\n--- 3. 여러 서비스 병렬 호출 ---")

	type serviceCall struct {
		name  string
		delay time.Duration
	}

	services := []serviceCall{
		{"UserService", 80 * time.Millisecond},
		{"OrderService", 120 * time.Millisecond},
		{"ProductService", 60 * time.Millisecond},
		{"PaymentService", 100 * time.Millisecond},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	responses := make([]APIResponse, len(services))
	g, ctx := errgroup.WithContext(ctx)

	start := time.Now()
	for i, svc := range services {
		i, svc := i, svc
		g.Go(func() error {
			resp, err := callService(ctx, svc.name, svc.delay)
			if err != nil {
				return err
			}
			responses[i] = resp
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("  서비스 호출 실패: %v\n", err)
		return
	}

	fmt.Printf("  전체 소요 시간: %v (순차 시 %v)\n",
		time.Since(start).Round(time.Millisecond),
		func() time.Duration {
			total := time.Duration(0)
			for _, s := range services {
				total += s.delay
			}
			return total
		}())

	for _, r := range responses {
		fmt.Printf("  %s: %s (지연: %v)\n",
			r.ServiceName, r.Data, r.Latency)
	}
}

// ─────────────────────────────────────────
// 4. errgroup 재사용 (SetLimit)
// ─────────────────────────────────────────

func setLimitDemo() {
	fmt.Println("\n--- 4. errgroup.SetLimit (동시 실행 제한) ---")
	fmt.Println("  Go 1.20+: SetLimit으로 동시 고루틴 수 제한")

	g := new(errgroup.Group)
	g.SetLimit(3) // 최대 3개 동시 실행

	start := time.Now()
	for i := 1; i <= 9; i++ {
		i := i
		g.Go(func() error {
			fmt.Printf("  작업 #%d 시작 (%.0fms)\n",
				i, float64(time.Since(start).Milliseconds()))
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("  작업 #%d 완료 (%.0fms)\n",
				i, float64(time.Since(start).Milliseconds()))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("  에러: %v\n", err)
		return
	}
	fmt.Printf("  전체 소요 시간: %v\n", time.Since(start).Round(time.Millisecond))
	fmt.Println("  (9개 작업을 3개씩 = 3배치 = 약 300ms)")
}

// ─────────────────────────────────────────
// 5. errgroup vs WaitGroup 비교
// ─────────────────────────────────────────

func comparisonDemo() {
	fmt.Println("\n--- 5. errgroup vs sync.WaitGroup 비교 ---")

	fmt.Println("  sync.WaitGroup 방식 (에러 처리가 번거로움):")
	fmt.Println("  var wg sync.WaitGroup")
	fmt.Println("  errs := make(chan error, n)")
	fmt.Println("  for ...: wg.Add(1); go func() { defer wg.Done(); if err ...: errs <- err }()")
	fmt.Println("  wg.Wait(); close(errs)")
	fmt.Println("  for err := range errs { ... }")
	fmt.Println()
	fmt.Println("  errgroup 방식 (에러 처리 내장):")
	fmt.Println("  g, ctx := errgroup.WithContext(ctx)")
	fmt.Println("  for ...: g.Go(func() error { ...; return err })")
	fmt.Println("  if err := g.Wait(); err != nil { ... }")
	fmt.Println()
	fmt.Println("  errgroup 장점:")
	fmt.Println("  - 첫 번째 에러 자동 전파")
	fmt.Println("  - 에러 발생 시 Context 자동 취소")
	fmt.Println("  - 보일러플레이트 코드 감소")
	fmt.Println("  - SetLimit으로 동시 실행 제한")
	fmt.Println()
	fmt.Println("  errgroup 주의사항:")
	fmt.Println("  - 여러 에러 중 첫 번째만 반환 (나머지 무시)")
	fmt.Println("  - 모든 에러 수집이 필요하면 별도 수집 채널 사용")
	fmt.Println("  - g.Go()에 nil 함수 전달 금지 (패닉)")
}

// ─────────────────────────────────────────
// 6. 실전: 파일 병렬 처리 with errgroup
// ─────────────────────────────────────────

// FileProcessor는 파일을 처리하는 시뮬레이션입니다.
type FileProcessor struct {
	maxWorkers int
}

func (fp *FileProcessor) ProcessAll(ctx context.Context, files []string) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(fp.maxWorkers)

	results := make([]string, len(files))

	for i, file := range files {
		i, file := i, file
		g.Go(func() error {
			// 처리 시뮬레이션
			delay := time.Duration(rand.Intn(80)+20) * time.Millisecond
			select {
			case <-time.After(delay):
				// 특정 파일 에러 시뮬레이션
				if file == "corrupt.dat" {
					return fmt.Errorf("파일 손상: %s", file)
				}
				results[i] = fmt.Sprintf("처리 완료: %s", file)
				return nil
			case <-ctx.Done():
				return fmt.Errorf("취소됨: %s", file)
			}
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	for _, r := range results {
		if r != "" {
			fmt.Printf("    %s\n", r)
		}
	}
	return nil
}

func fileProcessorDemo() {
	fmt.Println("\n--- 6. 실전: 파일 병렬 처리 ---")

	processor := &FileProcessor{maxWorkers: 3}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 성공 케이스
	fmt.Println("  정상 파일 처리:")
	files := []string{"data1.csv", "data2.csv", "data3.csv", "data4.csv", "data5.csv"}
	if err := processor.ProcessAll(ctx, files); err != nil {
		fmt.Printf("  에러: %v\n", err)
	} else {
		fmt.Println("  모든 파일 처리 완료!")
	}

	// 에러 케이스
	fmt.Println("\n  손상 파일 포함 처리:")
	filesWithError := []string{"data1.csv", "corrupt.dat", "data3.csv", "data4.csv"}
	if err := processor.ProcessAll(ctx, filesWithError); err != nil {
		fmt.Printf("  처리 실패: %v\n", err)
	}
}

// ─────────────────────────────────────────
// 테스트 서버 설정
// ─────────────────────────────────────────

func setupTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/fast", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		fmt.Fprintf(w, "fast response")
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		fmt.Fprintf(w, "slow response")
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})

	return httptest.NewServer(mux)
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== 동시성 패턴: errgroup ===")

	// 테스트용 HTTP 서버 시작
	server := setupTestServer()
	defer server.Close()
	fmt.Printf("테스트 서버: %s\n", server.URL)

	basicErrgroupDemo(server.URL)
	errorPropagationDemo()
	resultCollectionDemo()
	setLimitDemo()
	comparisonDemo()
	fileProcessorDemo()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
