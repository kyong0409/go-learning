// handler_test.go: HTTP 핸들러 테이블 기반 테스트
//
// httptest 패키지를 사용하여 실제 HTTP 서버 없이 핸들러를 테스트합니다.
// - httptest.NewRecorder(): 응답을 캡처하는 ResponseWriter
// - httptest.NewRequest(): 테스트용 HTTP 요청 생성
package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"bookmark-api/handler"
	"bookmark-api/model"
)

// ─────────────────────────────────────────
// 테스트 헬퍼
// ─────────────────────────────────────────

// setupHandler는 테스트용 핸들러와 미리 채워진 저장소를 반환합니다.
func setupHandler(t *testing.T) (*handler.Handler, model.Store) {
	t.Helper()
	store := model.NewInMemoryStore()
	h := handler.New(store)
	return h, store
}

// seedBookmarks는 테스트용 초기 데이터를 저장소에 추가합니다.
func seedBookmarks(t *testing.T, store model.Store) []*model.Bookmark {
	t.Helper()
	seeds := []model.CreateBookmarkRequest{
		{Title: "Go 공식 문서", URL: "https://go.dev/doc", Tags: []string{"go", "docs"}},
		{Title: "GitHub", URL: "https://github.com", Tags: []string{"git", "dev"}},
		{Title: "Stack Overflow", URL: "https://stackoverflow.com"},
	}
	created := make([]*model.Bookmark, 0, len(seeds))
	for _, s := range seeds {
		b, err := store.Create(s)
		if err != nil {
			t.Fatalf("시드 데이터 생성 실패: %v", err)
		}
		created = append(created, b)
	}
	return created
}

// mustJSON은 값을 JSON bytes로 변환합니다.
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("JSON 인코딩 실패: %v", err)
	}
	return b
}

// decodeResponse는 응답 본문을 디코딩합니다.
type testResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeResponse(t *testing.T, body []byte) testResponse {
	t.Helper()
	var resp testResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("응답 파싱 실패: %v\n본문: %s", err, body)
	}
	return resp
}

// ─────────────────────────────────────────
// GET /bookmarks 테스트
// ─────────────────────────────────────────

func TestGetAll(t *testing.T) {
	tests := []struct {
		name       string
		seedCount  int    // 사전에 생성할 북마크 수
		wantStatus int
		wantCount  int
	}{
		{"빈 목록", 0, http.StatusOK, 0},
		{"1개 있을 때", 1, http.StatusOK, 1},
		{"3개 있을 때", 3, http.StatusOK, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, store := setupHandler(t)

			// 시드 데이터 생성
			for i := 0; i < tt.seedCount; i++ {
				_, err := store.Create(model.CreateBookmarkRequest{
					Title: fmt.Sprintf("북마크 %d", i+1),
					URL:   fmt.Sprintf("https://example%d.com", i+1),
				})
				if err != nil {
					t.Fatalf("시드 생성 실패: %v", err)
				}
			}

			// 요청 실행
			req := httptest.NewRequest("GET", "/bookmarks", nil)
			rec := httptest.NewRecorder()
			h.GetAll(rec, req)

			// 상태 코드 확인
			if rec.Code != tt.wantStatus {
				t.Errorf("상태 코드: got=%d, want=%d", rec.Code, tt.wantStatus)
			}

			// Content-Type 확인
			ct := rec.Header().Get("Content-Type")
			if ct != "application/json" {
				t.Errorf("Content-Type: got=%q, want=%q", ct, "application/json")
			}

			// 응답 파싱
			resp := decodeResponse(t, rec.Body.Bytes())
			if !resp.Success {
				t.Errorf("success=false, message=%q", resp.Message)
			}

			// 북마크 수 확인
			var bookmarks []*model.Bookmark
			if err := json.Unmarshal(resp.Data, &bookmarks); err != nil {
				t.Fatalf("data 파싱 실패: %v", err)
			}
			if len(bookmarks) != tt.wantCount {
				t.Errorf("북마크 수: got=%d, want=%d", len(bookmarks), tt.wantCount)
			}
		})
	}
}

// ─────────────────────────────────────────
// GET /bookmarks/{id} 테스트
// ─────────────────────────────────────────

func TestGetByID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantTitle  string
	}{
		{"존재하는 ID", "1", http.StatusOK, "Go 공식 문서"},
		{"존재하지 않는 ID", "999", http.StatusNotFound, ""},
		{"잘못된 ID (문자)", "abc", http.StatusBadRequest, ""},
		{"잘못된 ID (음수)", "-1", http.StatusBadRequest, ""},
		{"잘못된 ID (0)", "0", http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, store := setupHandler(t)
			seedBookmarks(t, store)

			// Go 1.22 ServeMux를 통해 PathValue가 설정되도록 mux를 경유
			mux := http.NewServeMux()
			mux.HandleFunc("GET /bookmarks/{id}", h.GetByID)

			req := httptest.NewRequest("GET", "/bookmarks/"+tt.id, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("상태 코드: got=%d, want=%d\n본문: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantTitle != "" {
				resp := decodeResponse(t, rec.Body.Bytes())
				var bookmark model.Bookmark
				if err := json.Unmarshal(resp.Data, &bookmark); err != nil {
					t.Fatalf("data 파싱 실패: %v", err)
				}
				if bookmark.Title != tt.wantTitle {
					t.Errorf("title: got=%q, want=%q", bookmark.Title, tt.wantTitle)
				}
			}
		})
	}
}

// ─────────────────────────────────────────
// POST /bookmarks 테스트
// ─────────────────────────────────────────

func TestCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
		wantTitle  string
	}{
		{
			"정상 생성 (tags 포함)",
			model.CreateBookmarkRequest{Title: "테스트", URL: "https://test.com", Tags: []string{"test"}},
			http.StatusCreated,
			"테스트",
		},
		{
			"정상 생성 (tags 없음)",
			model.CreateBookmarkRequest{Title: "심플", URL: "https://simple.com"},
			http.StatusCreated,
			"심플",
		},
		{
			"title 누락",
			model.CreateBookmarkRequest{URL: "https://notitle.com"},
			http.StatusBadRequest,
			"",
		},
		{
			"url 누락",
			model.CreateBookmarkRequest{Title: "URL 없음"},
			http.StatusBadRequest,
			"",
		},
		{
			"잘못된 JSON",
			"not-valid-json{",
			http.StatusBadRequest,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := setupHandler(t)

			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes = mustJSON(t, tt.body)
			}

			req := httptest.NewRequest("POST", "/bookmarks", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.Create(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("상태 코드: got=%d, want=%d\n본문: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantTitle != "" {
				resp := decodeResponse(t, rec.Body.Bytes())
				if !resp.Success {
					t.Errorf("success=false, message=%q", resp.Message)
				}
				var bookmark model.Bookmark
				if err := json.Unmarshal(resp.Data, &bookmark); err != nil {
					t.Fatalf("data 파싱 실패: %v", err)
				}
				if bookmark.Title != tt.wantTitle {
					t.Errorf("title: got=%q, want=%q", bookmark.Title, tt.wantTitle)
				}
				if bookmark.ID == 0 {
					t.Error("생성된 북마크의 ID가 0입니다")
				}
			}
		})
	}
}

// ─────────────────────────────────────────
// PUT /bookmarks/{id} 테스트
// ─────────────────────────────────────────

func TestUpdate(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       model.UpdateBookmarkRequest
		wantStatus int
		wantTitle  string
	}{
		{
			"제목 수정",
			"1",
			model.UpdateBookmarkRequest{Title: "새로운 제목"},
			http.StatusOK,
			"새로운 제목",
		},
		{
			"URL 수정",
			"1",
			model.UpdateBookmarkRequest{URL: "https://newurl.com"},
			http.StatusOK,
			"Go 공식 문서", // 제목은 변경 없음
		},
		{
			"존재하지 않는 ID",
			"999",
			model.UpdateBookmarkRequest{Title: "없는 항목"},
			http.StatusNotFound,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, store := setupHandler(t)
			seedBookmarks(t, store)

			mux := http.NewServeMux()
			mux.HandleFunc("PUT /bookmarks/{id}", h.Update)

			bodyBytes := mustJSON(t, tt.body)
			req := httptest.NewRequest("PUT", "/bookmarks/"+tt.id, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("상태 코드: got=%d, want=%d\n본문: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantTitle != "" {
				resp := decodeResponse(t, rec.Body.Bytes())
				var bookmark model.Bookmark
				if err := json.Unmarshal(resp.Data, &bookmark); err != nil {
					t.Fatalf("data 파싱 실패: %v", err)
				}
				if bookmark.Title != tt.wantTitle {
					t.Errorf("title: got=%q, want=%q", bookmark.Title, tt.wantTitle)
				}
			}
		})
	}
}

// ─────────────────────────────────────────
// DELETE /bookmarks/{id} 테스트
// ─────────────────────────────────────────

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{"존재하는 북마크 삭제", "1", http.StatusNoContent},
		{"존재하지 않는 북마크 삭제", "999", http.StatusNotFound},
		{"잘못된 ID", "abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, store := setupHandler(t)
			seedBookmarks(t, store)

			mux := http.NewServeMux()
			mux.HandleFunc("DELETE /bookmarks/{id}", h.Delete)

			req := httptest.NewRequest("DELETE", "/bookmarks/"+tt.id, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("상태 코드: got=%d, want=%d\n본문: %s",
					rec.Code, tt.wantStatus, rec.Body.String())
			}

			// 삭제 후 실제로 없어졌는지 확인
			if tt.wantStatus == http.StatusNoContent {
				id := 1
				_, err := store.GetByID(id)
				if err == nil {
					t.Errorf("삭제 후 ID=%d가 여전히 존재합니다", id)
				}
			}
		})
	}
}
