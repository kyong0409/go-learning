// a3-rest-api/api_test.go
// 도서 관리 REST API 통합 테스트입니다.
package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	api "github.com/learn-go/a3-rest-api"
)

// ── 테스트 헬퍼 ────────────────────────────────────────────────

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := api.NewMemoryStore()
	a := api.New(store)
	mux := http.NewServeMux()
	a.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func doRequest(t *testing.T, method, url string, body interface{}) *http.Response {
	t.Helper()
	var req *http.Request
	if body != nil {
		data, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, url, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, url, nil)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	return resp
}

func decodeBody(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("JSON 디코딩 실패: %v", err)
	}
}

func createBook(t *testing.T, baseURL string, title, author string) *api.Book {
	t.Helper()
	resp := doRequest(t, http.MethodPost, baseURL+"/api/books", map[string]interface{}{
		"title":        title,
		"author":       author,
		"price":        15000.0,
		"published_at": time.Now(),
	})
	if resp.StatusCode != http.StatusCreated {
		resp.Body.Close()
		t.Fatalf("도서 생성 실패: 상태 %d", resp.StatusCode)
	}
	var book api.Book
	decodeBody(t, resp, &book)
	return &book
}

// ── CREATE 테스트 (10점) ────────────────────────────────────────

func TestCreateBook_Success(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, http.MethodPost, srv.URL+"/api/books", map[string]interface{}{
		"title":        "Go 프로그래밍",
		"author":       "김철수",
		"isbn":         "9781234567890",
		"price":        25000.0,
		"published_at": time.Now(),
	})

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("CREATE 상태 코드: 기대 201, 실제 %d", resp.StatusCode)
	}

	var book api.Book
	decodeBody(t, resp, &book)

	if book.ID == 0 {
		t.Error("생성된 도서 ID가 0입니다")
	}
	if book.Title != "Go 프로그래밍" {
		t.Errorf("제목: 기대 'Go 프로그래밍', 실제 %q", book.Title)
	}
	if book.Author != "김철수" {
		t.Errorf("저자: 기대 '김철수', 실제 %q", book.Author)
	}
}

func TestCreateBook_Validation(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	cases := []struct {
		name   string
		body   map[string]interface{}
		wantStatus int
	}{
		{"빈 제목", map[string]interface{}{"title": "", "author": "저자"}, http.StatusBadRequest},
		{"제목 없음", map[string]interface{}{"author": "저자"}, http.StatusBadRequest},
		{"저자 없음", map[string]interface{}{"title": "제목"}, http.StatusBadRequest},
		{"음수 가격", map[string]interface{}{"title": "제목", "author": "저자", "price": -1000.0}, http.StatusBadRequest},
		{"잘못된 ISBN", map[string]interface{}{"title": "제목", "author": "저자", "isbn": "12345"}, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, http.MethodPost, srv.URL+"/api/books", tc.body)
			resp.Body.Close()
			if resp.StatusCode != tc.wantStatus {
				t.Errorf("%s: 기대 %d, 실제 %d", tc.name, tc.wantStatus, resp.StatusCode)
			}
		})
	}
}

// ── READ 테스트 (10점) ─────────────────────────────────────────

func TestGetBook_Success(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	created := createBook(t, srv.URL, "Clean Code", "Robert Martin")

	resp := doRequest(t, http.MethodGet, fmt.Sprintf("%s/api/books/%d", srv.URL, created.ID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET 상태 코드: 기대 200, 실제 %d", resp.StatusCode)
	}

	var book api.Book
	decodeBody(t, resp, &book)
	if book.ID != created.ID {
		t.Errorf("ID: 기대 %d, 실제 %d", created.ID, book.ID)
	}
}

func TestGetBook_NotFound(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, http.MethodGet, srv.URL+"/api/books/99999", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("없는 도서 GET: 기대 404, 실제 %d", resp.StatusCode)
	}
}

func TestGetBook_InvalidID(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp := doRequest(t, http.MethodGet, srv.URL+"/api/books/abc", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("잘못된 ID: 기대 400, 실제 %d", resp.StatusCode)
	}
}

// ── UPDATE 테스트 (10점) ────────────────────────────────────────

func TestUpdateBook_Put(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	book := createBook(t, srv.URL, "원본 제목", "원본 저자")

	resp := doRequest(t, http.MethodPut, fmt.Sprintf("%s/api/books/%d", srv.URL, book.ID),
		map[string]interface{}{
			"title":        "수정된 제목",
			"author":       "수정된 저자",
			"price":        30000.0,
			"published_at": time.Now(),
		})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("PUT 상태 코드: 기대 200, 실제 %d", resp.StatusCode)
	}

	var updated api.Book
	decodeBody(t, resp, &updated)
	if updated.Title != "수정된 제목" {
		t.Errorf("수정된 제목: 기대 '수정된 제목', 실제 %q", updated.Title)
	}
}

func TestUpdateBook_Patch(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	book := createBook(t, srv.URL, "원본 제목", "원본 저자")

	// PATCH: 제목만 수정
	title := "패치된 제목"
	resp := doRequest(t, http.MethodPatch, fmt.Sprintf("%s/api/books/%d", srv.URL, book.ID),
		map[string]interface{}{"title": title})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("PATCH 상태 코드: 기대 200, 실제 %d", resp.StatusCode)
	}

	var patched api.Book
	decodeBody(t, resp, &patched)

	if patched.Title != title {
		t.Errorf("패치된 제목: 기대 %q, 실제 %q", title, patched.Title)
	}
	// 저자는 변경되지 않아야 합니다
	if patched.Author != "원본 저자" {
		t.Errorf("저자 불변: 기대 '원본 저자', 실제 %q", patched.Author)
	}
}

// ── DELETE 테스트 (10점) ────────────────────────────────────────

func TestDeleteBook(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	book := createBook(t, srv.URL, "삭제할 책", "저자")

	resp := doRequest(t, http.MethodDelete, fmt.Sprintf("%s/api/books/%d", srv.URL, book.ID), nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("DELETE 상태 코드: 기대 204, 실제 %d", resp.StatusCode)
	}

	// 삭제 후 조회
	resp2 := doRequest(t, http.MethodGet, fmt.Sprintf("%s/api/books/%d", srv.URL, book.ID), nil)
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("삭제 후 GET: 기대 404, 실제 %d", resp2.StatusCode)
	}
}

// ── LIST 테스트 (10점) ─────────────────────────────────────────

func TestListBooks(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	for i := range 5 {
		createBook(t, srv.URL, fmt.Sprintf("도서 %d", i+1), fmt.Sprintf("저자 %d", i+1))
	}

	resp := doRequest(t, http.MethodGet, srv.URL+"/api/books", nil)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("LIST 상태 코드: 기대 200, 실제 %d", resp.StatusCode)
	}

	var list api.ListBooksResponse
	decodeBody(t, resp, &list)

	if list.Total < 5 {
		t.Errorf("전체 개수: 최소 5, 실제 %d", list.Total)
	}
	t.Logf("총 %d권 (페이지 %d/%d)", list.Total, list.Page, list.TotalPages)
}

func TestListBooks_Pagination(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	for i := range 10 {
		createBook(t, srv.URL, fmt.Sprintf("페이지 테스트 도서 %d", i+1), "저자")
	}

	resp := doRequest(t, http.MethodGet, srv.URL+"/api/books?page=1&page_size=3", nil)
	var list api.ListBooksResponse
	decodeBody(t, resp, &list)

	if len(list.Items) > 3 {
		t.Errorf("페이지 크기 3인데 %d개 반환됨", len(list.Items))
	}
}

func TestListBooks_Search(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	createBook(t, srv.URL, "Go 언어 완벽 가이드", "김고랭")
	createBook(t, srv.URL, "Python 기초", "이파이썬")
	createBook(t, srv.URL, "Go 웹 개발", "박고랭")

	resp := doRequest(t, http.MethodGet, srv.URL+"/api/books?search=Go", nil)
	var list api.ListBooksResponse
	decodeBody(t, resp, &list)

	if list.Total < 2 {
		t.Errorf("'Go' 검색 결과: 최소 2개 기대, 실제 %d개", list.Total)
	}
}

// ── 성적 보고서 ─────────────────────────────────────────────────

func TestMain(m *testing.M) {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║   과제 A3: 도서 관리 REST API         ║")
	fmt.Println("╚══════════════════════════════════════╝")

	result := m.Run()

	fmt.Println()
	fmt.Println("─────────────────────────────────────")
	if result == 0 {
		fmt.Println("  최종 점수: 100 / 100 점")
		fmt.Println("  평가: 합격 (모든 테스트 통과)")
	} else {
		fmt.Println("  평가: 미완성 — 실패한 테스트를 확인하세요")
	}
	fmt.Println("─────────────────────────────────────")

	os.Exit(result)
}
