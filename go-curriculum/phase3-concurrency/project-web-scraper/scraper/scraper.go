// 패키지 scraper는 병렬 웹 스크레이퍼의 핵심 로직을 제공합니다.
package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// Config는 스크레이퍼 설정입니다.
type Config struct {
	Workers    int           // 동시 워커 수
	RatePerSec float64       // 초당 최대 요청 수
	Timeout    time.Duration // 전체 타임아웃
	Depth      int           // 링크 크롤링 깊이
	UserAgent  string        // HTTP User-Agent 헤더
}

// Result는 단일 URL 스크레이핑 결과입니다.
type Result struct {
	URL           string        // 요청 URL
	StatusCode    int           // HTTP 상태 코드
	ContentLength int           // 응답 본문 크기 (bytes)
	Title         string        // HTML <title> 텍스트
	Links         []string      // 추출된 링크 목록
	Latency       time.Duration // 요청 소요 시간
	Error         error         // 에러 (nil이면 성공)
	Depth         int           // 크롤링 깊이
}

// Scraper는 병렬 웹 스크레이퍼입니다.
type Scraper struct {
	cfg    Config
	client *http.Client

	// 속도 제한: 토큰 버킷
	rateLimiter <-chan time.Time

	// 방문 추적: 중복 URL 방지
	visited   map[string]bool
	visitedMu sync.Mutex
}

// New는 새 Scraper를 생성합니다.
func New(cfg Config) *Scraper {
	if cfg.Workers <= 0 {
		cfg.Workers = 3
	}
	if cfg.RatePerSec <= 0 {
		cfg.RatePerSec = 5.0
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "Go-Curriculum-Scraper/1.0"
	}

	// HTTP 클라이언트: 연결당 타임아웃 설정
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("리다이렉트 최대 횟수 초과")
			}
			return nil
		},
	}

	// 토큰 버킷 속도 제한 채널
	interval := time.Duration(float64(time.Second) / cfg.RatePerSec)
	rateLimiter := time.NewTicker(interval)

	return &Scraper{
		cfg:         cfg,
		client:      client,
		rateLimiter: rateLimiter.C,
		visited:     make(map[string]bool),
	}
}

// Scrape는 주어진 URL 목록을 병렬로 스크레이핑합니다.
// Context 취소 또는 타임아웃 시 진행 중인 작업을 중단합니다.
func (s *Scraper) Scrape(ctx context.Context, urls []string) []Result {
	// 작업 채널: URL 큐
	jobs := make(chan scrapeJob, len(urls)*s.cfg.Depth*10)
	// 결과 채널
	results := make(chan Result, len(urls)*s.cfg.Depth*10)

	// 초기 URL 큐에 추가
	for _, url := range urls {
		jobs <- scrapeJob{URL: url, Depth: 0}
	}

	// errgroup으로 워커 관리
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(s.cfg.Workers)

	// 활성 작업 추적
	var activeJobs sync.WaitGroup
	for _, url := range urls {
		activeJobs.Add(1)
		_ = url
	}

	// 결과 수집 고루틴
	var allResults []Result
	var resultsMu sync.Mutex
	collectDone := make(chan struct{})
	go func() {
		defer close(collectDone)
		for r := range results {
			resultsMu.Lock()
			allResults = append(allResults, r)
			resultsMu.Unlock()
		}
	}()

	// 워커 풀
	var pendingJobs sync.WaitGroup
	for _, url := range urls {
		url := url
		pendingJobs.Add(1)
		g.Go(func() error {
			defer pendingJobs.Done()
			result := s.scrapeOne(ctx, url, 0)
			select {
			case results <- result:
			case <-ctx.Done():
				return nil
			}

			// 링크 크롤링 (깊이 제한)
			if result.Error == nil && result.Depth < s.cfg.Depth-1 {
				for _, link := range result.Links {
					if s.markVisited(link) {
						continue // 이미 방문
					}
					link := link
					pendingJobs.Add(1)
					g.Go(func() error {
						defer pendingJobs.Done()
						childResult := s.scrapeOne(ctx, link, result.Depth+1)
						select {
						case results <- childResult:
						case <-ctx.Done():
						}
						return nil
					})
				}
			}
			return nil
		})
	}

	// 모든 작업 완료 후 results 채널 닫기
	go func() {
		pendingJobs.Wait()
		close(results)
	}()

	g.Wait()
	<-collectDone

	return allResults
}

// scrapeJob은 스크레이핑할 작업을 나타냅니다.
type scrapeJob struct {
	URL   string
	Depth int
}

// scrapeOne은 단일 URL을 스크레이핑합니다.
func (s *Scraper) scrapeOne(ctx context.Context, url string, depth int) Result {
	// 속도 제한: 다음 토큰까지 대기
	select {
	case <-s.rateLimiter:
	case <-ctx.Done():
		return Result{URL: url, Error: ctx.Err(), Depth: depth}
	}

	start := time.Now()
	fmt.Printf("  스크레이핑: %s (깊이: %d)\n", url, depth)

	// HTTP 요청
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Result{
			URL:     url,
			Error:   fmt.Errorf("요청 생성 실패: %w", err),
			Latency: time.Since(start),
			Depth:   depth,
		}
	}
	req.Header.Set("User-Agent", s.cfg.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := s.client.Do(req)
	if err != nil {
		return Result{
			URL:     url,
			Error:   fmt.Errorf("요청 실패: %w", err),
			Latency: time.Since(start),
			Depth:   depth,
		}
	}
	defer resp.Body.Close()

	// 본문 읽기 (최대 1MB)
	const maxBodySize = 1 * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return Result{
			URL:        url,
			StatusCode: resp.StatusCode,
			Error:      fmt.Errorf("본문 읽기 실패: %w", err),
			Latency:    time.Since(start),
			Depth:      depth,
		}
	}

	// HTML 파싱: 링크와 제목 추출
	title, links := ParseHTML(body, url)

	return Result{
		URL:           url,
		StatusCode:    resp.StatusCode,
		ContentLength: len(body),
		Title:         title,
		Links:         links,
		Latency:       time.Since(start),
		Depth:         depth,
	}
}

// markVisited는 URL을 방문 처리하고 이미 방문했으면 true를 반환합니다.
func (s *Scraper) markVisited(url string) bool {
	s.visitedMu.Lock()
	defer s.visitedMu.Unlock()
	if s.visited[url] {
		return true
	}
	s.visited[url] = true
	return false
}
