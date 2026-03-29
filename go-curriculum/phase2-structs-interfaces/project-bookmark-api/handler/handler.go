// Package handler는 HTTP 요청 핸들러를 정의합니다.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"bookmark-api/model"
)

// ─────────────────────────────────────────
// Handler 구조체
// ─────────────────────────────────────────

// Handler는 모든 HTTP 핸들러를 보유합니다.
// 의존성(store)을 구조체 필드로 주입합니다.
type Handler struct {
	store model.Store // 인터페이스 타입으로 저장 (교체 가능)
}

// New는 Handler 생성자입니다.
// Store 인터페이스를 받아 의존성을 주입합니다.
func New(store model.Store) *Handler {
	return &Handler{store: store}
}

// ─────────────────────────────────────────
// 응답 헬퍼
// ─────────────────────────────────────────

// apiResponse는 API 응답의 공통 구조입니다.
type apiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// writeJSON은 JSON 응답을 작성합니다.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// 인코딩 실패 시 내부 서버 에러 (이미 헤더는 작성됨)
		http.Error(w, "JSON 인코딩 실패", http.StatusInternalServerError)
	}
}

// writeSuccess는 성공 응답을 작성합니다.
func writeSuccess(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, apiResponse{
		Success: true,
		Data:    data,
	})
}

// writeError는 에러 응답을 작성합니다.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiResponse{
		Success: false,
		Message: message,
	})
}

// parseID는 URL 경로에서 정수 ID를 파싱합니다.
// Go 1.22+ ServeMux의 {id} 패턴 변수를 사용합니다.
func parseID(r *http.Request) (int, error) {
	idStr := r.PathValue("id") // Go 1.22+ 기능
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, errors.New("유효하지 않은 ID입니다")
	}
	return id, nil
}

// ─────────────────────────────────────────
// 핸들러 메서드들
// ─────────────────────────────────────────

// GetAll은 모든 북마크를 반환합니다.
// GET /bookmarks
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	bookmarks, err := h.store.GetAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "북마크 조회 실패: "+err.Error())
		return
	}

	// nil 슬라이스 대신 빈 슬라이스 반환 (JSON: [] vs null)
	if bookmarks == nil {
		bookmarks = []*model.Bookmark{}
	}

	writeSuccess(w, http.StatusOK, bookmarks)
}

// GetByID는 특정 ID의 북마크를 반환합니다.
// GET /bookmarks/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	bookmark, err := h.store.GetByID(id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ID "+strconv.Itoa(id)+" 북마크를 찾을 수 없습니다")
			return
		}
		writeError(w, http.StatusInternalServerError, "북마크 조회 실패: "+err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, bookmark)
}

// Create는 새 북마크를 생성합니다.
// POST /bookmarks
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookmarkRequest

	// 요청 본문 디코딩
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "요청 본문 파싱 실패: "+err.Error())
		return
	}

	bookmark, err := h.store.Create(req)
	if err != nil {
		if errors.Is(err, model.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "필수 필드 누락: title과 url은 필수입니다")
			return
		}
		writeError(w, http.StatusInternalServerError, "북마크 생성 실패: "+err.Error())
		return
	}

	// 201 Created + 생성된 리소스 반환
	writeSuccess(w, http.StatusCreated, bookmark)
}

// Update는 기존 북마크를 수정합니다.
// PUT /bookmarks/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req model.UpdateBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "요청 본문 파싱 실패: "+err.Error())
		return
	}

	bookmark, err := h.store.Update(id, req)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ID "+strconv.Itoa(id)+" 북마크를 찾을 수 없습니다")
			return
		}
		writeError(w, http.StatusInternalServerError, "북마크 수정 실패: "+err.Error())
		return
	}

	writeSuccess(w, http.StatusOK, bookmark)
}

// Delete는 북마크를 삭제합니다.
// DELETE /bookmarks/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.store.Delete(id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			writeError(w, http.StatusNotFound, "ID "+strconv.Itoa(id)+" 북마크를 찾을 수 없습니다")
			return
		}
		writeError(w, http.StatusInternalServerError, "북마크 삭제 실패: "+err.Error())
		return
	}

	// 204 No Content: 삭제 성공, 응답 본문 없음
	w.WriteHeader(http.StatusNoContent)
}
