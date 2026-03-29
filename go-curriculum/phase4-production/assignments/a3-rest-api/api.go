// a3-rest-api/api.go
// 도서 관리 REST API 핸들러 과제입니다.
// TODO 주석이 있는 모든 메서드를 구현하세요.
package api

import (
	"net/http"
)

// API는 HTTP 핸들러 집합입니다.
type API struct {
	store Store
}

// New는 새 API 인스턴스를 생성합니다.
func New(store Store) *API {
	return &API{store: store}
}

// RegisterRoutes는 ServeMux에 라우트를 등록합니다.
// Go 1.22+ 메서드 기반 패턴을 사용합니다.
func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/books", a.ListBooks)
	mux.HandleFunc("POST /api/books", a.CreateBook)
	mux.HandleFunc("GET /api/books/{id}", a.GetBook)
	mux.HandleFunc("PUT /api/books/{id}", a.UpdateBook)
	mux.HandleFunc("PATCH /api/books/{id}", a.PatchBook)
	mux.HandleFunc("DELETE /api/books/{id}", a.DeleteBook)
}

// ListBooks는 도서 목록을 반환합니다.
// GET /api/books?page=1&page_size=10&search=키워드&author=저자
// TODO: 구현하세요.
func (a *API) ListBooks(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. 쿼리 파라미터 파싱 (page, page_size, search, author)
	// 2. 기본값 설정 (page=1, page_size=10)
	// 3. store.List 호출
	// 4. ListBooksResponse JSON 응답
	writeJSON(w, http.StatusOK, map[string]string{"message": "구현 필요"})
}

// CreateBook은 새 도서를 생성합니다.
// POST /api/books
// Body: CreateBookRequest JSON
// TODO: 구현하세요.
func (a *API) CreateBook(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. JSON 요청 파싱
	// 2. 입력 검증 (Validate 메서드 사용)
	// 3. Book 생성 후 store.Create 호출
	// 4. 201 Created + Book JSON 응답
	writeJSON(w, http.StatusCreated, map[string]string{"message": "구현 필요"})
}

// GetBook은 단건 도서를 반환합니다.
// GET /api/books/{id}
// TODO: 구현하세요.
func (a *API) GetBook(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. r.PathValue("id")로 ID 추출
	// 2. ID를 정수로 변환 (잘못된 경우 400)
	// 3. store.GetByID 호출
	// 4. 없으면 404, 있으면 200 + Book JSON 응답
	writeJSON(w, http.StatusOK, map[string]string{"message": "구현 필요"})
}

// UpdateBook은 도서 전체를 수정합니다.
// PUT /api/books/{id}
// TODO: 구현하세요.
func (a *API) UpdateBook(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. ID 추출 및 검증
	// 2. 기존 도서 조회
	// 3. JSON 파싱 및 검증
	// 4. store.Update 호출
	// 5. 200 + 수정된 Book JSON 응답
	writeJSON(w, http.StatusOK, map[string]string{"message": "구현 필요"})
}

// PatchBook은 도서를 부분 수정합니다.
// PATCH /api/books/{id}
// 요청에 포함된 필드만 업데이트합니다.
// TODO: 구현하세요.
func (a *API) PatchBook(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. ID 추출 및 검증
	// 2. 기존 도서 조회
	// 3. UpdateBookRequest JSON 파싱
	// 4. nil이 아닌 필드만 업데이트
	// 5. store.Update 호출
	// 6. 200 + 수정된 Book JSON 응답
	writeJSON(w, http.StatusOK, map[string]string{"message": "구현 필요"})
}

// DeleteBook은 도서를 삭제합니다.
// DELETE /api/books/{id}
// TODO: 구현하세요.
func (a *API) DeleteBook(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. ID 추출 및 검증
	// 2. store.Delete 호출
	// 3. 없으면 404, 성공하면 204 No Content
	writeJSON(w, http.StatusNoContent, nil)
}

// ============================================================
// 헬퍼 함수 (자유롭게 사용하세요)
// ============================================================

// writeJSON은 JSON 응답을 작성합니다.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	// TODO: Content-Type 설정, 상태 코드 설정, JSON 인코딩
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data == nil {
		return
	}
	// json.NewEncoder(w).Encode(data) — 구현 시 주석 해제
}

// writeError는 오류 JSON 응답을 작성합니다.
func writeError(w http.ResponseWriter, status int, message, code, field string) {
	writeJSON(w, status, ErrorResponse{Error: message, Code: code, Field: field})
}
