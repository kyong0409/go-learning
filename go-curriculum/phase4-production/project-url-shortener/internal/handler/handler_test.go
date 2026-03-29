// internal/handler/handler_test.go
// httptest를 사용한 HTTP 핸들러 통합 테스트입니다.
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/learn-go/url-shortener/internal/handler"
	"github.com/learn-go/url-shortener/internal/model"
	"github.com/learn-go/url-shortener/internal/store"
)

// ============================================================
// 테스트 설정
// ============================================================

// testServer는 테스트용 서버를 생성합니다.
func testServer(t *testing.T) (*httptest.Server, *store.MemoryStore) {
	t.Helper()

	memStore := store.NewMemoryStore()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // 테스트 중 불필요한 로그 억제
	}))

	h := handler.New(memStore, "http://localhost:8080", logger, 6)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	return httptest.NewServer(r), memStore
}

// postJSON은 JSON POST 요청을 보내는 헬퍼입니다.
func postJSON(t *testing.T, url string, body interface{}) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("JSON 마샬링 실패: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST 요청 실패: %v", err)
	}
	return resp
}

// decodeJSON은 응답 본문을 JSON으로 디코딩하는 헬퍼입니다.
func decodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("JSON 디코딩 실패: %v", err)
	}
}

// ============================================================
// URL 단축 테스트
// ============================================================

func TestShorten(t *testing.T) {
	srv, _ := testServer(t)
	defer srv.Close()

	t.Run("정상 URL 단축", func(t *testing.T) {
		resp := postJSON(t, srv.URL+"/api/shorten", map[string]string{
			"url": "https://www.google.com",
		})

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("상태 코드: 기대 %d, 실제 %d", http.StatusCreated, resp.StatusCode)
		}

		var result model.ShortenResponse
		decodeJSON(t, resp, &result)

		if result.ShortCode == "" {
			t.Error("단축 코드가 비어있습니다")
		}
		if result.OriginalURL != "https://www.google.com" {
			t.Errorf("원본 URL: 기대 %s, 실제 %s", "https://www.google.com", result.OriginalURL)
		}
		if result.ShortURL == "" {
			t.Error("단축 URL이 비어있습니다")
		}
		t.Logf("단축 코드: %s, 단축 URL: %s", result.ShortCode, result.ShortURL)
	})

	t.Run("커스텀 코드 지정", func(t *testing.T) {
		resp := postJSON(t, srv.URL+"/api/shorten", map[string]string{
			"url":         "https://github.com",
			"custom_code": "github",
		})

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("상태 코드: 기대 %d, 실제 %d", http.StatusCreated, resp.StatusCode)
		}

		var result model.ShortenResponse
		decodeJSON(t, resp, &result)

		if result.ShortCode != "github" {
			t.Errorf("단축 코드: 기대 %s, 실제 %s", "github", result.ShortCode)
		}
	})

	t.Run("중복 커스텀 코드", func(t *testing.T) {
		// 첫 번째 등록
		postJSON(t, srv.URL+"/api/shorten", map[string]string{
			"url":         "https://example1.com",
			"custom_code": "dup-test",
		})

		// 중복 등록 시도
		resp := postJSON(t, srv.URL+"/api/shorten", map[string]string{
			"url":         "https://example2.com",
			"custom_code": "dup-test",
		})

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("상태 코드: 기대 %d, 실제 %d", http.StatusConflict, resp.StatusCode)
		}
	})
}

func TestShortenValidation(t *testing.T) {
	srv, _ := testServer(t)
	defer srv.Close()

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
		desc           string
	}{
		{
			name:           "빈 URL",
			body:           map[string]string{"url": ""},
			expectedStatus: http.StatusBadRequest,
			desc:           "URL이 비어있으면 400을 반환해야 합니다",
		},
		{
			name:           "잘못된 URL 형식",
			body:           map[string]string{"url": "not-a-url"},
			expectedStatus: http.StatusBadRequest,
			desc:           "유효하지 않은 URL이면 400을 반환해야 합니다",
		},
		{
			name:           "http/https 외 스킴",
			body:           map[string]string{"url": "ftp://files.example.com"},
			expectedStatus: http.StatusBadRequest,
			desc:           "ftp:// 같은 스킴은 거부해야 합니다",
		},
		{
			name:           "너무 짧은 커스텀 코드",
			body:           map[string]string{"url": "https://example.com", "custom_code": "ab"},
			expectedStatus: http.StatusBadRequest,
			desc:           "커스텀 코드가 3자 미만이면 400을 반환해야 합니다",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := postJSON(t, srv.URL+"/api/shorten", tt.body)
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("%s: 기대 %d, 실제 %d", tt.desc, tt.expectedStatus, resp.StatusCode)
			}
			resp.Body.Close()
		})
	}
}

// ============================================================
// 리다이렉트 테스트
// ============================================================

func TestRedirect(t *testing.T) {
	srv, _ := testServer(t)
	defer srv.Close()

	// URL 단축
	resp := postJSON(t, srv.URL+"/api/shorten", map[string]string{
		"url":         "https://www.example.com",
		"custom_code": "redir-test",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("URL 단축 실패: %d", resp.StatusCode)
	}
	resp.Body.Close()

	t.Run("올바른 코드로 리다이렉트", func(t *testing.T) {
		// 리다이렉트를 따르지 않는 클라이언트 사용
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		resp, err := client.Get(srv.URL + "/redir-test")
		if err != nil {
			t.Fatalf("GET 요청 실패: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusFound {
			t.Errorf("상태 코드: 기대 %d, 실제 %d", http.StatusFound, resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if location != "https://www.example.com" {
			t.Errorf("Location 헤더: 기대 %s, 실제 %s", "https://www.example.com", location)
		}
	})

	t.Run("존재하지 않는 코드", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/nonexistent")
		if err != nil {
			t.Fatalf("GET 요청 실패: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("상태 코드: 기대 %d, 실제 %d", http.StatusNotFound, resp.StatusCode)
		}
	})
}

// ============================================================
// 통계 테스트
// ============================================================

func TestGetStats(t *testing.T) {
	srv, _ := testServer(t)
	defer srv.Close()

	// URL 단축
	resp := postJSON(t, srv.URL+"/api/shorten", map[string]string{
		"url":         "https://stats-test.com",
		"custom_code": "stats",
	})
	resp.Body.Close()

	// 몇 번 클릭 시뮬레이션
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	for range 3 {
		r, _ := client.Get(srv.URL + "/stats")
		if r != nil {
			r.Body.Close()
		}
	}

	t.Run("통계 조회", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/api/urls/stats/stats")
		if err != nil {
			t.Fatalf("GET 요청 실패: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("상태 코드: 기대 %d, 실제 %d", http.StatusOK, resp.StatusCode)
		}

		var stats model.StatsResponse
		decodeJSON(t, resp, &stats)

		if stats.ShortCode != "stats" {
			t.Errorf("단축 코드: 기대 stats, 실제 %s", stats.ShortCode)
		}
		t.Logf("클릭 수: %d", stats.ClickCount)
	})
}

// ============================================================
// 목록 조회 테스트
// ============================================================

func TestListURLs(t *testing.T) {
	srv, _ := testServer(t)
	defer srv.Close()

	// 여러 URL 등록
	for i := range 5 {
		resp := postJSON(t, srv.URL+"/api/shorten", map[string]string{
			"url":         fmt.Sprintf("https://example%d.com", i),
			"custom_code": fmt.Sprintf("list-test-%d", i),
		})
		resp.Body.Close()
	}

	t.Run("목록 조회", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/api/urls")
		if err != nil {
			t.Fatalf("GET 요청 실패: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("상태 코드: 기대 %d, 실제 %d", http.StatusOK, resp.StatusCode)
		}

		var list model.ListResponse
		decodeJSON(t, resp, &list)

		if list.Total < 5 {
			t.Errorf("전체 개수: 최소 5, 실제 %d", list.Total)
		}
		t.Logf("총 %d개 URL (페이지: %d/%d)", list.Total, list.Page, list.TotalPages)
	})

	t.Run("페이지네이션", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/api/urls?page=1&page_size=2")
		if err != nil {
			t.Fatalf("GET 요청 실패: %v", err)
		}

		var list model.ListResponse
		decodeJSON(t, resp, &list)

		if len(list.Items) > 2 {
			t.Errorf("페이지 크기 2인데 %d개 반환됨", len(list.Items))
		}
		if list.PageSize != 2 {
			t.Errorf("page_size: 기대 2, 실제 %d", list.PageSize)
		}
	})
}

// ============================================================
// 삭제 테스트
// ============================================================

func TestDeleteURL(t *testing.T) {
	srv, _ := testServer(t)
	defer srv.Close()

	// URL 등록
	resp := postJSON(t, srv.URL+"/api/shorten", map[string]string{
		"url":         "https://delete-test.com",
		"custom_code": "del-me",
	})
	resp.Body.Close()

	t.Run("URL 삭제", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/urls/del-me", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("DELETE 요청 실패: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("상태 코드: 기대 %d, 실제 %d", http.StatusNoContent, resp.StatusCode)
		}
	})

	t.Run("삭제 후 조회 시 404", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/api/urls/del-me")
		if err != nil {
			t.Fatalf("GET 요청 실패: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("삭제된 URL 조회 시 404가 와야 합니다: 실제 %d", resp.StatusCode)
		}
	})
}

// ============================================================
// 헬스 체크 테스트
// ============================================================

func TestHealth(t *testing.T) {
	srv, _ := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("헬스 체크 상태 코드: 기대 %d, 실제 %d", http.StatusOK, resp.StatusCode)
	}

	var result map[string]string
	decodeJSON(t, resp, &result)

	if result["status"] != "healthy" {
		t.Errorf("상태: 기대 healthy, 실제 %s", result["status"])
	}
}

// ============================================================
// 성적 보고서 출력
// ============================================================

func TestMain(m *testing.M) {
	fmt.Println("======================================")
	fmt.Println("  URL 단축기 핸들러 테스트 시작")
	fmt.Println("======================================")

	code := m.Run()

	fmt.Println()
	fmt.Println("======================================")
	if code == 0 {
		fmt.Println("  결과: 모든 테스트 통과!")
		fmt.Println("  점수: 100/100")
	} else {
		fmt.Println("  결과: 일부 테스트 실패")
		fmt.Println("  실패한 테스트를 확인하세요")
	}
	fmt.Println("======================================")

	os.Exit(code)
}
