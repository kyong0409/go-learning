// 패키지 scraper 테스트
package scraper_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-curriculum/web-scraper/scraper"
)

// ─────────────────────────────────────────
// 테스트 서버 헬퍼
// ─────────────────────────────────────────

// setupTestServer는 테스트용 HTTP 서버를 설정합니다.
// 실제 외부 네트워크 없이 스크레이퍼를 테스트할 수 있습니다.
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// 정상 페이지
	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>페이지 1</title></head>
<body>
  <h1>테스트 페이지 1</h1>
  <a href="/page2">페이지 2 링크</a>
  <a href="/page3">페이지 3 링크</a>
  <a href="https://external.example.com">외부 링크</a>
</body>
</html>`)
	})

	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>페이지 2</title></head>
<body>
  <h1>테스트 페이지 2</h1>
  <a href="/page1">페이지 1로 돌아가기</a>
</body>
</html>`)
	})

	mux.HandleFunc("/page3", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>페이지 3</title></head>
<body><p>내용이 없는 페이지</p></body>
</html>`)
	})

	// 느린 페이지 (타임아웃 테스트용)
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second):
			fmt.Fprintf(w, "<html><body>느린 응답</body></html>")
		case <-r.Context().Done():
			return
		}
	})

	// 에러 페이지
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	})

	// 리다이렉트
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/page1", http.StatusMovedPermanently)
	})

	// 여러 링크가 있는 페이지
	mux.HandleFunc("/many-links", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		var sb strings.Builder
		sb.WriteString("<html><head><title>많은 링크</title></head><body>")
		for i := 1; i <= 20; i++ {
			fmt.Fprintf(&sb, `<a href="/page%d">링크 %d</a>`, i, i)
		}
		sb.WriteString("</body></html>")
		fmt.Fprint(w, sb.String())
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// ─────────────────────────────────────────
// ParseHTML 테스트
// ─────────────────────────────────────────

func TestParseHTML_Title(t *testing.T) {
	html := []byte(`<html><head><title>테스트 제목</title></head><body></body></html>`)
	title, _ := scraper.ParseHTML(html, "http://example.com")

	if title != "테스트 제목" {
		t.Errorf("제목 파싱 실패: 기대='테스트 제목', 실제='%s'", title)
	}
}

func TestParseHTML_Links(t *testing.T) {
	html := []byte(`<html><body>
		<a href="/page1">링크1</a>
		<a href="https://external.com">외부링크</a>
		<a href="javascript:void(0)">JS 링크 (무시됨)</a>
		<a href="mailto:test@test.com">메일 (무시됨)</a>
		<a href="#">앵커 (무시됨)</a>
	</body></html>`)

	_, links := scraper.ParseHTML(html, "http://example.com")

	// javascript:, mailto:, # 링크는 제외
	if len(links) != 2 {
		t.Errorf("링크 수 오류: 기대=2, 실제=%d, 링크=%v", len(links), links)
	}

	// 절대 URL로 변환 확인
	found := false
	for _, l := range links {
		if l == "http://example.com/page1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("상대 URL 변환 실패: http://example.com/page1 없음, 링크=%v", links)
	}
}

func TestParseHTML_EmptyHTML(t *testing.T) {
	title, links := scraper.ParseHTML([]byte(""), "http://example.com")
	if title != "" {
		t.Errorf("빈 HTML에서 제목 파싱: '%s'", title)
	}
	if len(links) != 0 {
		t.Errorf("빈 HTML에서 링크 파싱: %v", links)
	}
}

func TestParseHTML_DuplicateLinks(t *testing.T) {
	html := []byte(`<html><body>
		<a href="/page1">링크1</a>
		<a href="/page1">중복 링크1</a>
		<a href="/page2">링크2</a>
	</body></html>`)

	_, links := scraper.ParseHTML(html, "http://example.com")

	// 중복 제거 확인
	if len(links) != 2 {
		t.Errorf("중복 링크 제거 실패: 기대=2, 실제=%d", len(links))
	}
}

// ─────────────────────────────────────────
// Scraper 통합 테스트
// ─────────────────────────────────────────

func TestScraper_BasicScraping(t *testing.T) {
	server := setupTestServer(t)

	cfg := scraper.Config{
		Workers:    2,
		RatePerSec: 100, // 테스트에서는 빠르게
		Timeout:    10 * time.Second,
		Depth:      1,
	}
	s := scraper.New(cfg)

	ctx := context.Background()
	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
	}

	results := s.Scrape(ctx, urls)

	if len(results) != 2 {
		t.Fatalf("결과 수 오류: 기대=2, 실제=%d", len(results))
	}

	for _, r := range results {
		if r.Error != nil {
			t.Errorf("URL %s 스크레이핑 실패: %v", r.URL, r.Error)
		}
		if r.StatusCode != http.StatusOK {
			t.Errorf("URL %s 상태 코드 오류: 기대=200, 실제=%d", r.URL, r.StatusCode)
		}
		if r.Title == "" {
			t.Errorf("URL %s 제목 없음", r.URL)
		}
		if r.ContentLength == 0 {
			t.Errorf("URL %s 내용 없음", r.URL)
		}
	}
}

func TestScraper_TitleExtraction(t *testing.T) {
	server := setupTestServer(t)

	cfg := scraper.Config{
		Workers:    1,
		RatePerSec: 100,
		Timeout:    10 * time.Second,
		Depth:      1,
	}
	s := scraper.New(cfg)

	results := s.Scrape(context.Background(), []string{server.URL + "/page1"})

	if len(results) != 1 {
		t.Fatalf("결과 수 오류: %d", len(results))
	}

	if results[0].Title != "페이지 1" {
		t.Errorf("제목 오류: 기대='페이지 1', 실제='%s'", results[0].Title)
	}
}

func TestScraper_LinkExtraction(t *testing.T) {
	server := setupTestServer(t)

	cfg := scraper.Config{
		Workers:    1,
		RatePerSec: 100,
		Timeout:    10 * time.Second,
		Depth:      1,
	}
	s := scraper.New(cfg)

	results := s.Scrape(context.Background(), []string{server.URL + "/page1"})

	if len(results) != 1 {
		t.Fatalf("결과 수 오류: %d", len(results))
	}

	// /page1에는 3개 링크 있어야 함 (/page2, /page3, 외부링크)
	if len(results[0].Links) < 2 {
		t.Errorf("링크 수 부족: 기대>=2, 실제=%d", len(results[0].Links))
	}
}

func TestScraper_ContextCancellation(t *testing.T) {
	server := setupTestServer(t)

	cfg := scraper.Config{
		Workers:    2,
		RatePerSec: 100,
		Timeout:    10 * time.Second,
		Depth:      1,
	}
	s := scraper.New(cfg)

	// 즉시 취소
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 바로 취소

	results := s.Scrape(ctx, []string{
		server.URL + "/slow",
		server.URL + "/slow",
	})

	// 취소된 경우 결과가 없거나 에러여야 함
	for _, r := range results {
		if r.Error == nil {
			t.Errorf("취소된 요청이 성공함: %s", r.URL)
		}
	}
}

func TestScraper_Timeout(t *testing.T) {
	server := setupTestServer(t)

	cfg := scraper.Config{
		Workers:    1,
		RatePerSec: 100,
		Timeout:    200 * time.Millisecond, // 짧은 타임아웃
		Depth:      1,
	}
	s := scraper.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	results := s.Scrape(ctx, []string{server.URL + "/slow"})

	// 타임아웃으로 에러여야 함
	if len(results) > 0 && results[0].Error == nil {
		t.Error("타임아웃 요청이 성공함")
	}
}

func TestScraper_HTTPError(t *testing.T) {
	server := setupTestServer(t)

	cfg := scraper.Config{
		Workers:    1,
		RatePerSec: 100,
		Timeout:    10 * time.Second,
		Depth:      1,
	}
	s := scraper.New(cfg)

	results := s.Scrape(context.Background(), []string{server.URL + "/error"})

	if len(results) != 1 {
		t.Fatalf("결과 수 오류: %d", len(results))
	}

	// 500 에러는 에러 없이 반환되지만 StatusCode가 500
	if results[0].StatusCode != http.StatusInternalServerError {
		t.Errorf("HTTP 500 처리 오류: 상태=%d", results[0].StatusCode)
	}
}

func TestScraper_ParallelPerformance(t *testing.T) {
	server := setupTestServer(t)

	// 5개 URL을 순차/병렬 처리 시간 비교
	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
		server.URL + "/page3",
		server.URL + "/page1",
		server.URL + "/page2",
	}

	// 병렬 처리
	cfg := scraper.Config{
		Workers:    5,
		RatePerSec: 100,
		Timeout:    10 * time.Second,
		Depth:      1,
	}
	s := scraper.New(cfg)

	start := time.Now()
	results := s.Scrape(context.Background(), urls)
	parallelTime := time.Since(start)

	if len(results) != len(urls) {
		t.Errorf("결과 수 오류: 기대=%d, 실제=%d", len(urls), len(results))
	}

	t.Logf("병렬 처리 시간: %v (워커: %d개)", parallelTime, cfg.Workers)

	// 모든 요청 성공 확인
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("요청 실패: %s: %v", r.URL, r.Error)
		}
	}
}

func TestScraper_RateLimit(t *testing.T) {
	server := setupTestServer(t)

	// 초당 2개로 제한
	cfg := scraper.Config{
		Workers:    10,
		RatePerSec: 2.0,
		Timeout:    10 * time.Second,
		Depth:      1,
	}
	s := scraper.New(cfg)

	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
		server.URL + "/page3",
		server.URL + "/page1",
	}

	start := time.Now()
	results := s.Scrape(context.Background(), urls)
	elapsed := time.Since(start)

	// 4개 요청을 초당 2개 속도로 처리하면 최소 1.5초 소요
	minExpected := 1200 * time.Millisecond
	if elapsed < minExpected {
		t.Errorf("속도 제한 미작동: 소요=%v, 최소기대=%v", elapsed, minExpected)
	}

	t.Logf("속도 제한 테스트: %d개 요청, %v 소요", len(results), elapsed)
}
