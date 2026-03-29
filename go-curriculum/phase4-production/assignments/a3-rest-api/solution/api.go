// a3-rest-api/solution/api.go
// 도서 관리 REST API 참고 답안입니다.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// API는 HTTP 핸들러 집합입니다.
type API struct {
	store Store
}

func New(store Store) *API {
	return &API{store: store}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/books", a.ListBooks)
	mux.HandleFunc("POST /api/books", a.CreateBook)
	mux.HandleFunc("GET /api/books/{id}", a.GetBook)
	mux.HandleFunc("PUT /api/books/{id}", a.UpdateBook)
	mux.HandleFunc("PATCH /api/books/{id}", a.PatchBook)
	mux.HandleFunc("DELETE /api/books/{id}", a.DeleteBook)
}

func (a *API) ListBooks(w http.ResponseWriter, r *http.Request) {
	q := ListBooksQuery{
		Page:     getIntQuery(r, "page", 1),
		PageSize: getIntQuery(r, "page_size", 10),
		Search:   r.URL.Query().Get("search"),
		Author:   r.URL.Query().Get("author"),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 100 {
		q.PageSize = 10
	}

	items, total, err := a.store.List(r.Context(), q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "목록 조회 실패", "STORE_ERROR", "")
		return
	}
	if items == nil {
		items = []*Book{}
	}
	totalPages := (total + q.PageSize - 1) / q.PageSize
	writeJSON(w, http.StatusOK, ListBooksResponse{
		Items:      items,
		Total:      total,
		Page:       q.Page,
		PageSize:   q.PageSize,
		TotalPages: totalPages,
	})
}

func (a *API) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 JSON 형식", "INVALID_JSON", "")
		return
	}
	if ve := req.Validate(); ve != nil {
		writeError(w, http.StatusBadRequest, ve.Message, "VALIDATION_ERROR", ve.Field)
		return
	}
	book := &Book{
		Title:       strings.TrimSpace(req.Title),
		Author:      strings.TrimSpace(req.Author),
		ISBN:        req.ISBN,
		PublishedAt: req.PublishedAt,
		Price:       req.Price,
	}
	if err := a.store.Create(r.Context(), book); err != nil {
		writeError(w, http.StatusInternalServerError, "도서 생성 실패", "STORE_ERROR", "")
		return
	}
	writeJSON(w, http.StatusCreated, book)
}

func (a *API) GetBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	book, err := a.store.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "도서를 찾을 수 없습니다", "NOT_FOUND", "")
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func (a *API) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	existing, err := a.store.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "도서를 찾을 수 없습니다", "NOT_FOUND", "")
		return
	}
	var req CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 JSON 형식", "INVALID_JSON", "")
		return
	}
	if ve := req.Validate(); ve != nil {
		writeError(w, http.StatusBadRequest, ve.Message, "VALIDATION_ERROR", ve.Field)
		return
	}
	existing.Title = strings.TrimSpace(req.Title)
	existing.Author = strings.TrimSpace(req.Author)
	existing.ISBN = req.ISBN
	existing.PublishedAt = req.PublishedAt
	existing.Price = req.Price
	if err := a.store.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "도서 수정 실패", "STORE_ERROR", "")
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (a *API) PatchBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	existing, err := a.store.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "도서를 찾을 수 없습니다", "NOT_FOUND", "")
		return
	}
	var req UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 JSON 형식", "INVALID_JSON", "")
		return
	}
	if req.Title != nil {
		existing.Title = strings.TrimSpace(*req.Title)
	}
	if req.Author != nil {
		existing.Author = strings.TrimSpace(*req.Author)
	}
	if req.ISBN != nil {
		existing.ISBN = *req.ISBN
	}
	if req.PublishedAt != nil {
		existing.PublishedAt = *req.PublishedAt
	}
	if req.Price != nil {
		if *req.Price < 0 {
			writeError(w, http.StatusBadRequest, "가격은 0 이상이어야 합니다", "VALIDATION_ERROR", "price")
			return
		}
		existing.Price = *req.Price
	}
	if err := a.store.Update(r.Context(), existing); err != nil {
		writeError(w, http.StatusInternalServerError, "도서 수정 실패", "STORE_ERROR", "")
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (a *API) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := a.store.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, "도서를 찾을 수 없습니다", "NOT_FOUND", "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── helpers ──────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message, code, field string) {
	writeJSON(w, status, ErrorResponse{Error: message, Code: code, Field: field})
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("유효하지 않은 ID: %s", idStr), "INVALID_ID", "id")
		return 0, false
	}
	return id, true
}

func getIntQuery(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
