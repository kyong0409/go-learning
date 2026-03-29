// Package main은 북마크 API 서버의 진입점입니다.
// Go 1.22+의 향상된 ServeMux를 사용하여 RESTful API를 구현합니다.
package main

import (
	"fmt"
	"log"
	"net/http"

	"bookmark-api/handler"
	"bookmark-api/model"
)

func main() {
	// ─────────────────────────────────────────
	// 의존성 생성 (Dependency Setup)
	// ─────────────────────────────────────────

	// 저장소 생성: 인터페이스를 통해 구현을 주입합니다.
	// 나중에 PostgreSQL 등 다른 구현으로 교체 가능합니다.
	store := model.NewInMemoryStore()

	// 핸들러 생성: 저장소를 주입합니다.
	h := handler.New(store)

	// ─────────────────────────────────────────
	// 라우터 설정 (Go 1.22+ enhanced ServeMux)
	// ─────────────────────────────────────────
	// Go 1.22에서 ServeMux가 크게 개선되었습니다:
	// - HTTP 메서드 패턴: "GET /path", "POST /path"
	// - 경로 변수: "/bookmarks/{id}"
	// - 와일드카드: "/files/{path...}"

	mux := http.NewServeMux()

	// 북마크 CRUD 라우트
	// GET    /bookmarks        → 전체 목록 조회
	// POST   /bookmarks        → 새 북마크 생성
	// GET    /bookmarks/{id}   → 특정 북마크 조회
	// PUT    /bookmarks/{id}   → 북마크 수정
	// DELETE /bookmarks/{id}   → 북마크 삭제
	mux.HandleFunc("GET /bookmarks", h.GetAll)
	mux.HandleFunc("POST /bookmarks", h.Create)
	mux.HandleFunc("GET /bookmarks/{id}", h.GetByID)
	mux.HandleFunc("PUT /bookmarks/{id}", h.Update)
	mux.HandleFunc("DELETE /bookmarks/{id}", h.Delete)

	// 헬스체크 엔드포인트
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","service":"bookmark-api"}`)
	})

	// ─────────────────────────────────────────
	// 서버 시작
	// ─────────────────────────────────────────
	addr := ":8080"
	log.Printf("북마크 API 서버 시작: http://localhost%s", addr)
	log.Println("사용 가능한 엔드포인트:")
	log.Println("  GET    /health")
	log.Println("  GET    /bookmarks")
	log.Println("  POST   /bookmarks")
	log.Println("  GET    /bookmarks/{id}")
	log.Println("  PUT    /bookmarks/{id}")
	log.Println("  DELETE /bookmarks/{id}")

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
