// solution/loadtest.go - 참고 풀이
package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Config, Result 타입은 loadtest.go와 동일 (solution 패키지용 복사본)

type Config struct {
	URL         string
	Method      string
	Concurrency int
	Requests    int
	Duration    time.Duration
	Timeout     time.Duration
	Headers     http.Header
}

type Result struct {
	TotalRequests int
	Successes     int
	Failures      int
	TotalDuration time.Duration
	RPS           float64
	AvgLatency    time.Duration
	P50           time.Duration
	P90           time.Duration
	P95           time.Duration
	P99           time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
}

// LoadTester는 HTTP 부하 테스트를 실행합니다.
type LoadTester struct {
	config Config
	client *http.Client
	stats  *Stats
}

// NewLoadTester는 LoadTester 생성자입니다.
func NewLoadTester(config Config) *LoadTester {
	if config.Method == "" {
		config.Method = "GET"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &LoadTester{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
		stats:  NewStats(),
	}
}

// Run은 부하 테스트를 실행합니다.
func (lt *LoadTester) Run(ctx context.Context) (*Result, error) {
	// 시간 기반 종료를 위한 컨텍스트 처리
	runCtx := ctx
	var cancel context.CancelFunc

	if lt.config.Duration > 0 {
		runCtx, cancel = context.WithTimeout(ctx, lt.config.Duration)
		defer cancel()
	}

	start := time.Now()

	// 작업 채널
	jobs := make(chan struct{}, lt.config.Concurrency)

	var wg sync.WaitGroup

	// 워커 시작
	for i := 0; i < lt.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				latency, code, err := lt.sendRequest()
				lt.stats.Record(latency, code, err)
			}
		}()
	}

	// 작업 전송
	if lt.config.Requests > 0 {
		// 요청 수 기반
		for i := 0; i < lt.config.Requests; i++ {
			select {
			case <-runCtx.Done():
				goto done
			case jobs <- struct{}{}:
			}
		}
	} else {
		// 시간 기반
		for {
			select {
			case <-runCtx.Done():
				goto done
			case jobs <- struct{}{}:
			}
		}
	}

done:
	close(jobs)
	wg.Wait()

	result := lt.stats.Summary(time.Since(start))
	return &result, nil
}

// sendRequest는 단일 HTTP 요청을 전송합니다.
func (lt *LoadTester) sendRequest() (time.Duration, int, error) {
	start := time.Now()

	req, err := http.NewRequest(lt.config.Method, lt.config.URL, nil)
	if err != nil {
		return time.Since(start), 0, err
	}

	for k, vals := range lt.config.Headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}

	resp, err := lt.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return latency, 0, err
	}
	defer resp.Body.Close()

	return latency, resp.StatusCode, nil
}

func main() {
	fmt.Println("참고 풀이 - 직접 실행하지 마세요")
}
