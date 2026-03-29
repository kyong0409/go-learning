// 06-docker/main.go
// Docker 멀티스테이지 빌드를 위한 간단한 HTTP 서버입니다.
// 이 서버는 scratch 이미지에서 실행되도록 최소한의 의존성만 사용합니다.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

// BuildInfo는 빌드 정보를 담습니다.
// Docker 빌드 시 --build-arg로 주입할 수 있습니다.
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// HealthResponse는 헬스 체크 응답 구조입니다.
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
	Uptime    string `json:"uptime"`
}

// startTime은 서버 시작 시간입니다.
var startTime = time.Now()

// handleHealth는 서버 상태를 반환합니다.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "healthy",
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GoVersion: runtime.Version(),
		Uptime:    time.Since(startTime).Round(time.Second).String(),
	})
}

// handleHello는 간단한 인사 메시지를 반환합니다.
func handleHello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "세계"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("안녕하세요, %s!", name),
		"time":    time.Now().Format(time.RFC3339),
	})
}

// handleInfo는 서버 환경 정보를 반환합니다.
func handleInfo(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hostname":   hostname,
		"go_version": runtime.Version(),
		"go_os":      runtime.GOOS,
		"go_arch":    runtime.GOARCH,
		"num_cpu":    runtime.NumCPU(),
		"goroutines": runtime.NumGoroutine(),
	})
}

func main() {
	// 환경 변수에서 포트 읽기 (Docker 환경 변수 지원)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /hello", handleHello)
	mux.HandleFunc("GET /info", handleInfo)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   "Docker 멀티스테이지 빌드 Go 서버",
			"version":   Version,
			"endpoints": []string{"/health", "/hello?name=이름", "/info"},
		})
	})

	addr := ":" + port
	log.Printf("서버 시작: http://localhost%s (버전: %s)", addr, Version)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("서버 종료: %v", err)
	}
}
