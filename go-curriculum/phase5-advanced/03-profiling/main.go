// main.go
// Go pprof 프로파일링 예제
//
// 이 서버는 net/http/pprof 엔드포인트를 노출해
// CPU, 메모리, 고루틴 프로파일을 실시간으로 수집할 수 있습니다.
//
// 실행:
//   go run main.go
//
// 프로파일 수집:
//   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10
//   go tool pprof http://localhost:6060/debug/pprof/heap
//   go tool pprof http://localhost:6060/debug/pprof/goroutine
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // 사이드 이펙트 임포트: /debug/pprof/ 핸들러 자동 등록

	"github.com/curriculum/profiling-example/heavy"
)

// ============================================================
// HTTP 핸들러
// ============================================================

// handleCPUHeavy는 CPU 집약적 작업을 유발합니다.
// GET /work/cpu?n=1000000
func handleCPUHeavy(w http.ResponseWriter, r *http.Request) {
	n := 1_000_000
	result := heavy.CPUIntensive(n)
	fmt.Fprintf(w, "CPU 집약적 작업 완료: n=%d, result=%d\n", n, result)
}

// handleMemoryHeavy는 메모리 할당을 많이 유발합니다.
// GET /work/memory
func handleMemoryHeavy(w http.ResponseWriter, r *http.Request) {
	data := heavy.MemoryIntensive(10_000)
	fmt.Fprintf(w, "메모리 집약적 작업 완료: %d 바이트 할당됨\n", len(data))
}

// handleGoroutineHeavy는 많은 고루틴을 생성합니다.
// GET /work/goroutine
func handleGoroutineHeavy(w http.ResponseWriter, r *http.Request) {
	count := heavy.GoroutineIntensive(100)
	fmt.Fprintf(w, "고루틴 집약적 작업 완료: %d개 고루틴 사용\n", count)
}

// handleAll은 모든 작업을 순서대로 실행합니다.
// GET /work/all
func handleAll(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "=== 전체 작업 시작 ===")

	result := heavy.CPUIntensive(500_000)
	fmt.Fprintf(w, "CPU: result=%d\n", result)

	data := heavy.MemoryIntensive(5_000)
	fmt.Fprintf(w, "메모리: %d 바이트\n", len(data))

	count := heavy.GoroutineIntensive(50)
	fmt.Fprintf(w, "고루틴: %d개\n", count)

	fmt.Fprintln(w, "=== 전체 작업 완료 ===")
}

// ============================================================
// 메인 함수
// ============================================================

func main() {
	mux := http.NewServeMux()

	// 작업 핸들러 등록
	mux.HandleFunc("GET /work/cpu", handleCPUHeavy)
	mux.HandleFunc("GET /work/memory", handleMemoryHeavy)
	mux.HandleFunc("GET /work/goroutine", handleGoroutineHeavy)
	mux.HandleFunc("GET /work/all", handleAll)

	// 안내 핸들러
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "=== pprof 프로파일링 예제 서버 ===")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "작업 엔드포인트:")
		fmt.Fprintln(w, "  GET /work/cpu       - CPU 집약적 작업")
		fmt.Fprintln(w, "  GET /work/memory    - 메모리 집약적 작업")
		fmt.Fprintln(w, "  GET /work/goroutine - 고루틴 집약적 작업")
		fmt.Fprintln(w, "  GET /work/all       - 모든 작업")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "pprof 엔드포인트:")
		fmt.Fprintln(w, "  GET /debug/pprof/           - 프로파일 목록")
		fmt.Fprintln(w, "  GET /debug/pprof/profile    - CPU 프로파일 (30초)")
		fmt.Fprintln(w, "  GET /debug/pprof/heap       - 힙 메모리 프로파일")
		fmt.Fprintln(w, "  GET /debug/pprof/goroutine  - 고루틴 프로파일")
		fmt.Fprintln(w, "  GET /debug/pprof/trace      - 실행 트레이스")
	})

	// pprof는 DefaultServeMux에 등록됩니다.
	// 별도 포트(6060)에서 pprof 전용 서버 실행
	go func() {
		log.Println("pprof 서버 시작: http://localhost:6060/debug/pprof/")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Printf("pprof 서버 오류: %v", err)
		}
	}()

	log.Println("작업 서버 시작: http://localhost:8080")
	log.Println("프로파일 확인: http://localhost:6060/debug/pprof/")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
