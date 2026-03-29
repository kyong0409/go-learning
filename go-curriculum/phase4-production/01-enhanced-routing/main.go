// 01-enhanced-routing/main.go
// Go 1.22+의 향상된 HTTP 라우팅 기능을 학습합니다.
// 메서드 기반 패턴, 경로 와일드카드, 라우트 그룹핑 패턴을 다룹니다.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ============================================================
// 데이터 모델
// ============================================================

// User는 API 응답에 사용되는 사용자 모델입니다.
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Post는 블로그 포스트 모델입니다.
type Post struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

// APIResponse는 일관된 API 응답 구조입니다.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ============================================================
// 샘플 데이터 저장소
// ============================================================

var users = []User{
	{ID: 1, Name: "김철수", Email: "kim@example.com", CreatedAt: time.Now().Add(-24 * time.Hour)},
	{ID: 2, Name: "이영희", Email: "lee@example.com", CreatedAt: time.Now().Add(-48 * time.Hour)},
	{ID: 3, Name: "박민준", Email: "park@example.com", CreatedAt: time.Now().Add(-72 * time.Hour)},
}

var posts = []Post{
	{ID: 1, UserID: 1, Title: "첫 번째 포스트", Content: "안녕하세요!"},
	{ID: 2, UserID: 1, Title: "두 번째 포스트", Content: "Go 1.22 최고!"},
	{ID: 3, UserID: 2, Title: "이영희의 포스트", Content: "라우팅이 편해졌어요."},
}

// ============================================================
// 헬퍼 함수
// ============================================================

// writeJSON은 JSON 응답을 작성합니다.
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON 인코딩 오류: %v", err)
	}
}

// writeError는 오류 응답을 작성합니다.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, APIResponse{Success: false, Error: message})
}

// ============================================================
// 핸들러 함수들
// ============================================================

// handleListUsers는 모든 사용자 목록을 반환합니다.
// Go 1.22 이전: mux.HandleFunc("/api/users", handler)
// Go 1.22 이후: mux.HandleFunc("GET /api/users", handler) — 메서드 지정 가능!
func handleListUsers(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    users,
	})
}

// handleGetUser는 특정 ID의 사용자를 반환합니다.
// {id}는 경로 와일드카드입니다. r.PathValue("id")로 추출합니다.
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Go 1.22+: r.PathValue()로 경로 파라미터 추출
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("유효하지 않은 ID: %s", idStr))
		return
	}

	for _, u := range users {
		if u.ID == id {
			writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: u})
			return
		}
	}

	writeError(w, http.StatusNotFound, fmt.Sprintf("사용자 ID %d를 찾을 수 없습니다", id))
}

// handleCreateUser는 새 사용자를 생성합니다.
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청 본문")
		return
	}

	if input.Name == "" || input.Email == "" {
		writeError(w, http.StatusBadRequest, "이름과 이메일은 필수입니다")
		return
	}

	newUser := User{
		ID:        len(users) + 1,
		Name:      input.Name,
		Email:     input.Email,
		CreatedAt: time.Now(),
	}
	users = append(users, newUser)

	writeJSON(w, http.StatusCreated, APIResponse{Success: true, Data: newUser})
}

// handleUpdateUser는 사용자 정보를 업데이트합니다.
func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "유효하지 않은 ID")
		return
	}

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "잘못된 요청 본문")
		return
	}

	for i, u := range users {
		if u.ID == id {
			if input.Name != "" {
				users[i].Name = input.Name
			}
			if input.Email != "" {
				users[i].Email = input.Email
			}
			writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: users[i]})
			return
		}
	}

	writeError(w, http.StatusNotFound, "사용자를 찾을 수 없습니다")
}

// handleDeleteUser는 사용자를 삭제합니다.
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "유효하지 않은 ID")
		return
	}

	for i, u := range users {
		if u.ID == id {
			users = append(users[:i], users[i+1:]...)
			writeJSON(w, http.StatusOK, APIResponse{
				Success: true,
				Data:    map[string]string{"message": "사용자가 삭제되었습니다"},
			})
			return
		}
	}

	writeError(w, http.StatusNotFound, "사용자를 찾을 수 없습니다")
}

// ============================================================
// 중첩 리소스 핸들러 (사용자의 포스트)
// ============================================================

// handleListUserPosts는 특정 사용자의 모든 포스트를 반환합니다.
// 경로: GET /api/users/{userID}/posts
func handleListUserPosts(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "유효하지 않은 사용자 ID")
		return
	}

	var userPosts []Post
	for _, p := range posts {
		if p.UserID == userID {
			userPosts = append(userPosts, p)
		}
	}

	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: userPosts})
}

// handleGetUserPost는 특정 사용자의 특정 포스트를 반환합니다.
// 경로: GET /api/users/{userID}/posts/{postID}
// 두 개의 경로 파라미터를 동시에 추출하는 예제입니다.
func handleGetUserPost(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("userID")
	postIDStr := r.PathValue("postID")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "유효하지 않은 사용자 ID")
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "유효하지 않은 포스트 ID")
		return
	}

	for _, p := range posts {
		if p.UserID == userID && p.ID == postID {
			writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: p})
			return
		}
	}

	writeError(w, http.StatusNotFound, "포스트를 찾을 수 없습니다")
}

// ============================================================
// 와일드카드 경로 패턴
// ============================================================

// handleFiles는 파일 경로 와일드카드를 처리합니다.
// 패턴: GET /files/{path...}  — 나머지 모든 경로 세그먼트를 캡처
func handleFiles(w http.ResponseWriter, r *http.Request) {
	// {path...}는 슬래시를 포함한 나머지 경로를 모두 캡처합니다.
	filePath := r.PathValue("path")
	fmt.Fprintf(w, "요청된 파일 경로: %s\n", filePath)

	// 예: GET /files/images/logo.png -> filePath = "images/logo.png"
	parts := strings.Split(filePath, "/")
	fmt.Fprintf(w, "경로 세그먼트 수: %d\n", len(parts))
	for i, part := range parts {
		fmt.Fprintf(w, "  세그먼트[%d]: %s\n", i, part)
	}
}

// ============================================================
// 라우트 그룹핑 패턴
// ============================================================

// apiV1Handler는 /api/v1/ 접두사를 가진 라우트 그룹을 처리합니다.
// Go의 ServeMux는 자체적인 그룹핑 기능이 없으므로,
// 별도의 ServeMux를 생성하고 Strip으로 접두사를 제거하는 패턴을 사용합니다.
func newAPIv1Handler() http.Handler {
	mux := http.NewServeMux()

	// v1 전용 엔드포인트
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"version": "v1",
			"status":  "운영 중",
		})
	})

	mux.HandleFunc("GET /info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"version":    "1.0.0",
			"go_version": "1.22+",
			"features":   []string{"enhanced-routing", "method-patterns", "path-values"},
		})
	})

	return mux
}

// ============================================================
// 메인 라우터 설정
// ============================================================

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// --------------------------------------------------------
	// Go 1.22+ 메서드 기반 라우팅
	// 형식: "METHOD /path"
	// --------------------------------------------------------

	// 사용자 CRUD
	mux.HandleFunc("GET /api/users", handleListUsers)
	mux.HandleFunc("POST /api/users", handleCreateUser)
	mux.HandleFunc("GET /api/users/{id}", handleGetUser)
	mux.HandleFunc("PUT /api/users/{id}", handleUpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", handleDeleteUser)

	// 중첩 리소스: 사용자의 포스트
	mux.HandleFunc("GET /api/users/{userID}/posts", handleListUserPosts)
	mux.HandleFunc("GET /api/users/{userID}/posts/{postID}", handleGetUserPost)

	// 와일드카드 경로: {path...}는 나머지 경로 전체를 캡처
	mux.HandleFunc("GET /files/{path...}", handleFiles)

	// 라우트 그룹핑: http.StripPrefix로 접두사 제거 후 서브 ServeMux에 위임
	// /api/v1/status  -> 서브 mux의 /status
	// /api/v1/info    -> 서브 mux의 /info
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", newAPIv1Handler()))

	// 루트 핸들러: 사용 가능한 엔드포인트 목록 출력
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			writeError(w, http.StatusNotFound, "엔드포인트를 찾을 수 없습니다")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Go 1.22+ 향상된 라우팅 데모",
			"endpoints": []string{
				"GET  /api/users",
				"POST /api/users",
				"GET  /api/users/{id}",
				"PUT  /api/users/{id}",
				"DELETE /api/users/{id}",
				"GET  /api/users/{userID}/posts",
				"GET  /api/users/{userID}/posts/{postID}",
				"GET  /files/{path...}",
				"GET  /api/v1/status",
				"GET  /api/v1/info",
			},
		})
	})

	return mux
}

func main() {
	mux := setupRouter()

	addr := ":8080"
	fmt.Printf("서버 시작: http://localhost%s\n", addr)
	fmt.Println()
	fmt.Println("테스트 명령어:")
	fmt.Println("  curl http://localhost:8080/")
	fmt.Println("  curl http://localhost:8080/api/users")
	fmt.Println("  curl http://localhost:8080/api/users/1")
	fmt.Println("  curl -X POST http://localhost:8080/api/users -d '{\"name\":\"홍길동\",\"email\":\"hong@example.com\"}'")
	fmt.Println("  curl http://localhost:8080/api/users/1/posts")
	fmt.Println("  curl http://localhost:8080/files/images/logo.png")
	fmt.Println("  curl http://localhost:8080/api/v1/status")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
