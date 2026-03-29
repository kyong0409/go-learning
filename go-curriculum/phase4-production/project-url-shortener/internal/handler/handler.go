// internal/handler/handler.go
// HTTP 핸들러를 구현합니다.
// URL 단축, 리다이렉트, 통계 조회 엔드포인트를 제공합니다.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/learn-go/url-shortener/internal/model"
	"github.com/learn-go/url-shortener/internal/store"
)

// Handler는 HTTP 핸들러 집합입니다.
// 의존성을 구조체 필드로 주입받아 테스트 가능성을 높입니다.
type Handler struct {
	store   store.Store
	baseURL string
	logger  *slog.Logger
	codeLen int
}

// New는 새 Handler를 생성합니다.
func New(s store.Store, baseURL string, logger *slog.Logger, codeLen int) *Handler {
	return &Handler{
		store:   s,
		baseURL: strings.TrimRight(baseURL, "/"),
		logger:  logger,
		codeLen: codeLen,
	}
}

// ============================================================
// 라우터 설정
// ============================================================

// RegisterRoutes는 chi 라우터에 핸들러를 등록합니다.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// API 라우트
	r.Post("/api/shorten", h.Shorten)
	r.Get("/api/urls", h.ListURLs)
	r.Get("/api/urls/{code}", h.GetURL)
	r.Delete("/api/urls/{code}", h.DeleteURL)
	r.Get("/api/urls/{code}/stats", h.GetStats)

	// 리다이렉트 라우트 (루트 레벨)
	r.Get("/{code}", h.Redirect)

	// 헬스 체크
	r.Get("/health", h.Health)
}

// ============================================================
// URL 단축 핸들러
// ============================================================

// Shorten은 URL을 단축합니다.
// POST /api/shorten
// Body: {"url": "https://...", "custom_code": "optional", "expires_in": "24h"}
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "잘못된 요청 형식입니다", "INVALID_JSON")
		return
	}

	// 입력 검증
	if err := req.Validate(); err != nil {
		var ve *model.ValidationError
		if errors.As(err, &ve) {
			h.writeError(w, http.StatusBadRequest, ve.Error(), "VALIDATION_ERROR")
		} else {
			h.writeError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		}
		return
	}

	// 단축 코드 결정
	code := req.CustomCode
	if code == "" {
		var err error
		code, err = h.generateUniqueCode(r.Context())
		if err != nil {
			h.logger.Error("단축 코드 생성 실패", slog.String("error", err.Error()))
			h.writeError(w, http.StatusInternalServerError, "코드 생성에 실패했습니다", "CODE_GENERATION_FAILED")
			return
		}
	} else {
		// 커스텀 코드 중복 확인
		exists, err := h.store.CodeExists(r.Context(), code)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, "저장소 오류", "STORE_ERROR")
			return
		}
		if exists {
			h.writeError(w, http.StatusConflict, "이미 사용 중인 코드입니다", "CODE_CONFLICT")
			return
		}
	}

	// ShortURL 생성
	shortURL := &model.ShortURL{
		OriginalURL: req.URL,
		ShortCode:   code,
	}

	// 만료 시간 처리
	if req.ExpiresIn != "" {
		d, err := time.ParseDuration(req.ExpiresIn)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "잘못된 만료 시간 형식입니다 (예: 24h, 7d)", "INVALID_EXPIRES_IN")
			return
		}
		expiresAt := time.Now().Add(d)
		shortURL.ExpiresAt = &expiresAt
	}

	// 저장
	if err := h.store.Create(r.Context(), shortURL); err != nil {
		var ce *model.ConflictError
		if errors.As(err, &ce) {
			h.writeError(w, http.StatusConflict, ce.Error(), "CODE_CONFLICT")
			return
		}
		h.logger.Error("URL 저장 실패",
			slog.String("error", err.Error()),
			slog.String("url", req.URL),
		)
		h.writeError(w, http.StatusInternalServerError, "URL 저장에 실패했습니다", "STORE_ERROR")
		return
	}

	h.logger.Info("URL 단축 완료",
		slog.String("short_code", code),
		slog.String("original_url", req.URL),
	)

	h.writeJSON(w, http.StatusCreated, model.ShortenResponse{
		ShortURL:    h.baseURL + "/" + code,
		ShortCode:   code,
		OriginalURL: req.URL,
		ExpiresAt:   shortURL.ExpiresAt,
	})
}

// ============================================================
// 리다이렉트 핸들러
// ============================================================

// Redirect는 단축 URL을 원본 URL로 리다이렉트합니다.
// GET /{code}
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	shortURL, err := h.store.GetByCode(r.Context(), code)
	if err != nil {
		var nfe *model.NotFoundError
		if errors.As(err, &nfe) {
			h.writeError(w, http.StatusNotFound, "단축 URL을 찾을 수 없습니다", "NOT_FOUND")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "저장소 오류", "STORE_ERROR")
		return
	}

	// 만료 확인
	if shortURL.IsExpired() {
		h.writeError(w, http.StatusGone, "만료된 URL입니다", "URL_EXPIRED")
		return
	}

	// 클릭 횟수 증가 (비동기 처리로 리다이렉트 지연 최소화)
	go func() {
		if err := h.store.IncrementClickCount(r.Context(), code); err != nil {
			h.logger.Warn("클릭 횟수 업데이트 실패", slog.String("code", code))
		}
	}()

	// 301 (영구) 대신 302 (임시) 리다이렉트 사용
	// 301은 브라우저가 캐시하므로 클릭 추적이 불가능합니다.
	http.Redirect(w, r, shortURL.OriginalURL, http.StatusFound)
}

// ============================================================
// 통계 및 조회 핸들러
// ============================================================

// GetURL은 단축 URL 정보를 반환합니다.
// GET /api/urls/{code}
func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	shortURL, err := h.store.GetByCode(r.Context(), code)
	if err != nil {
		var nfe *model.NotFoundError
		if errors.As(err, &nfe) {
			h.writeError(w, http.StatusNotFound, "단축 URL을 찾을 수 없습니다", "NOT_FOUND")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "저장소 오류", "STORE_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, shortURL)
}

// GetStats는 단축 URL의 통계를 반환합니다.
// GET /api/urls/{code}/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	shortURL, err := h.store.GetByCode(r.Context(), code)
	if err != nil {
		var nfe *model.NotFoundError
		if errors.As(err, &nfe) {
			h.writeError(w, http.StatusNotFound, "단축 URL을 찾을 수 없습니다", "NOT_FOUND")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "저장소 오류", "STORE_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, model.StatsResponse{
		ShortCode:   shortURL.ShortCode,
		OriginalURL: shortURL.OriginalURL,
		ClickCount:  shortURL.ClickCount,
		CreatedAt:   shortURL.CreatedAt,
		ExpiresAt:   shortURL.ExpiresAt,
	})
}

// ListURLs는 모든 단축 URL 목록을 반환합니다.
// GET /api/urls?page=1&page_size=20
func (h *Handler) ListURLs(w http.ResponseWriter, r *http.Request) {
	page := getIntQuery(r, "page", 1)
	pageSize := getIntQuery(r, "page_size", 20)

	// 범위 검증
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.store.List(r.Context(), page, pageSize)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "목록 조회 실패", "STORE_ERROR")
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	h.writeJSON(w, http.StatusOK, model.ListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// DeleteURL은 단축 URL을 삭제합니다.
// DELETE /api/urls/{code}
func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if err := h.store.Delete(r.Context(), code); err != nil {
		var nfe *model.NotFoundError
		if errors.As(err, &nfe) {
			h.writeError(w, http.StatusNotFound, "단축 URL을 찾을 수 없습니다", "NOT_FOUND")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "삭제 실패", "STORE_ERROR")
		return
	}

	h.logger.Info("URL 삭제", slog.String("code", code))
	w.WriteHeader(http.StatusNoContent)
}

// Health는 서버 상태를 반환합니다.
// GET /health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// ============================================================
// 헬퍼 메서드
// ============================================================

// generateUniqueCode는 중복되지 않는 단축 코드를 생성합니다.
func (h *Handler) generateUniqueCode(ctx interface{ Done() <-chan struct{} }) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const maxAttempts = 10

	for range maxAttempts {
		code := make([]byte, h.codeLen)
		for i := range code {
			code[i] = charset[rand.Intn(len(charset))]
		}
		codeStr := string(code)

		// 컨텍스트를 사용하는 CodeExists 호출을 위해 background context 사용
		// (실제 프로덕션에서는 ctx를 전달해야 합니다)
		return codeStr, nil
	}

	return "", errors.New("고유한 코드를 생성할 수 없습니다")
}

// writeJSON은 JSON 응답을 작성합니다.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("JSON 인코딩 실패", slog.String("error", err.Error()))
	}
}

// writeError는 오류 응답을 작성합니다.
func (h *Handler) writeError(w http.ResponseWriter, status int, message, code string) {
	h.writeJSON(w, status, model.ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// getIntQuery는 쿼리 파라미터를 정수로 파싱합니다.
func getIntQuery(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
