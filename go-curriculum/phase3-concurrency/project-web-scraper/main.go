// 패키지 선언
package main

// 동시성 프로젝트: 병렬 웹 스크레이퍼
//
// 이 프로그램은 여러 URL을 동시에 스크레이핑하는 CLI 도구입니다.
//
// 사용법:
//   go run . -urls="https://example.com,https://golang.org" -workers=5 -rate=10
//
// 기능:
//   - 워커 풀로 동시 다운로드
//   - 속도 제한 (초당 요청 수 제한)
//   - Context 기반 타임아웃/취소
//   - HTML에서 링크 추출
//   - 결과 통계 출력

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-curriculum/web-scraper/scraper"
)

func main() {
	// CLI 플래그 정의
	urlsFlag := flag.String("urls", "", "스크레이핑할 URL 목록 (쉼표로 구분)")
	workersFlag := flag.Int("workers", 3, "워커 고루틴 수 (기본값: 3)")
	rateFlag := flag.Float64("rate", 5.0, "초당 최대 요청 수 (기본값: 5.0)")
	timeoutFlag := flag.Duration("timeout", 30*time.Second, "전체 타임아웃 (기본값: 30s)")
	depthFlag := flag.Int("depth", 1, "링크 크롤링 깊이 (기본값: 1, 최대: 3)")
	flag.Parse()

	// URL 검증
	if *urlsFlag == "" {
		fmt.Fprintln(os.Stderr, "에러: -urls 플래그가 필요합니다")
		fmt.Fprintln(os.Stderr, "사용법: go run . -urls=\"https://example.com,https://golang.org\"")
		flag.Usage()
		os.Exit(1)
	}

	urls := parseURLs(*urlsFlag)
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "에러: 유효한 URL이 없습니다")
		os.Exit(1)
	}

	// 깊이 제한
	if *depthFlag > 3 {
		*depthFlag = 3
		fmt.Fprintln(os.Stderr, "경고: 최대 깊이는 3입니다. 3으로 설정됩니다.")
	}

	// Context 설정: Ctrl+C 또는 타임아웃으로 취소
	ctx, cancel := context.WithTimeout(context.Background(), *timeoutFlag)
	defer cancel()

	// OS 시그널 처리 (Ctrl+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-sigCh:
			fmt.Printf("\n시그널 수신 (%v): 스크레이퍼 종료 중...\n", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	// 스크레이퍼 설정
	cfg := scraper.Config{
		Workers:    *workersFlag,
		RatePerSec: *rateFlag,
		Timeout:    *timeoutFlag,
		Depth:      *depthFlag,
		UserAgent:  "Go-Curriculum-Scraper/1.0",
	}

	s := scraper.New(cfg)

	// 실행
	fmt.Println("=== Go 동시성 프로젝트: 병렬 웹 스크레이퍼 ===")
	fmt.Printf("설정: 워커=%d, 속도제한=%.1f req/s, 타임아웃=%v, 깊이=%d\n",
		cfg.Workers, cfg.RatePerSec, cfg.Timeout, cfg.Depth)
	fmt.Printf("대상 URL: %d개\n", len(urls))
	fmt.Println(strings.Repeat("─", 60))

	start := time.Now()
	results := s.Scrape(ctx, urls)
	elapsed := time.Since(start)

	// 결과 출력
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("=== 스크레이핑 결과 ===")
	printResults(results, elapsed)
}

// parseURLs는 쉼표로 구분된 URL 문자열을 파싱합니다.
func parseURLs(raw string) []string {
	parts := strings.Split(raw, ",")
	var urls []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			urls = append(urls, p)
		}
	}
	return urls
}

// printResults는 스크레이핑 결과를 출력합니다.
func printResults(results []scraper.Result, elapsed time.Duration) {
	successful := 0
	failed := 0
	totalLinks := 0

	for _, r := range results {
		if r.Error != nil {
			failed++
			fmt.Printf("  [실패] %s\n", r.URL)
			fmt.Printf("         에러: %v\n", r.Error)
		} else {
			successful++
			totalLinks += len(r.Links)
			fmt.Printf("  [성공] %s\n", r.URL)
			fmt.Printf("         상태: %d, 크기: %d bytes, 링크: %d개, 소요: %v\n",
				r.StatusCode, r.ContentLength, len(r.Links), r.Latency.Round(time.Millisecond))
			if r.Title != "" {
				fmt.Printf("         제목: %s\n", r.Title)
			}
			// 첫 3개 링크만 출력
			if len(r.Links) > 0 {
				limit := 3
				if len(r.Links) < limit {
					limit = len(r.Links)
				}
				fmt.Printf("         링크 예시: ")
				for i := 0; i < limit; i++ {
					if i > 0 {
						fmt.Print(", ")
					}
					link := r.Links[i]
					if len(link) > 50 {
						link = link[:50] + "..."
					}
					fmt.Print(link)
				}
				fmt.Println()
			}
		}
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("총 소요 시간: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("처리: %d개 (성공: %d, 실패: %d)\n",
		len(results), successful, failed)
	fmt.Printf("총 링크 발견: %d개\n", totalLinks)
	if elapsed > 0 && len(results) > 0 {
		fmt.Printf("처리 속도: %.1f req/s\n",
			float64(len(results))/elapsed.Seconds())
	}
}
