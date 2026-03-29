// 패키지 선언
package graceful

import (
	"context"
	"net/http"
)

// Middleware는 HTTP 핸들러를 감싸는 미들웨어 함수 타입입니다.
// 등록 순서대로 적용됩니다 (첫 번째 등록이 가장 바깥쪽 레이어).
type Middleware func(http.Handler) http.Handler

// Server는 우아한 종료를 지원하는 HTTP 서버 래퍼입니다.
type Server struct {
	// TODO: 필요한 필드를 추가하세요
	// 힌트:
	// - addr string: 바인딩 주소 (예: ":8080")
	// - mux *http.ServeMux: 라우터
	// - middlewares []Middleware: 미들웨어 체인
	// - shutdownHooks []func(): 종료 훅 목록
	// - httpServer *http.Server: 내부 표준 라이브러리 서버
	// - listener net.Listener: 네트워크 리스너 (실제 포트 확인용)
	// - shuttingDown atomic.Bool: 종료 중 여부
}

// NewServer는 새 서버를 생성합니다.
// addr 예시: ":8080", ":0" (OS가 포트 자동 할당)
func NewServer(addr string) *Server {
	// TODO: 구현하세요
	// - 필드 초기화
	// - mux에 /healthz 핸들러 등록
	panic("구현 필요")
}

// Handle은 지정한 패턴에 핸들러를 등록합니다.
// 미들웨어는 Start() 시점에 적용되므로 여기서는 mux에만 등록합니다.
func (s *Server) Handle(pattern string, handler http.Handler) {
	// TODO: 구현하세요
	// s.mux.Handle(pattern, handler)
	panic("구현 필요")
}

// Use는 미들웨어를 체인에 추가합니다.
// 등록 순서: Use(A), Use(B) → 요청은 A → B → 핸들러 순으로 처리됩니다.
func (s *Server) Use(middleware Middleware) {
	// TODO: 구현하세요
	// s.middlewares = append(s.middlewares, middleware)
	panic("구현 필요")
}

// OnShutdown은 서버 종료 시 실행할 훅을 등록합니다.
// 훅은 등록 순서대로 실행됩니다.
func (s *Server) OnShutdown(hook func()) {
	// TODO: 구현하세요
	// s.shutdownHooks = append(s.shutdownHooks, hook)
	panic("구현 필요")
}

// Addr은 서버가 실제로 바인딩된 주소를 반환합니다.
// Start() 호출 전에는 빈 문자열을 반환할 수 있습니다.
// ":0"으로 시작한 경우 OS가 할당한 실제 포트를 반환합니다.
func (s *Server) Addr() string {
	// TODO: 구현하세요
	// listener가 nil이 아니면 s.listener.Addr().String() 반환
	// 아니면 s.addr 반환
	panic("구현 필요")
}

// Start는 서버를 시작합니다.
// 미들웨어 체인을 적용하고 리스너를 생성한 뒤 요청을 수락합니다.
// ctx가 취소되면 Shutdown을 호출하고 nil을 반환합니다.
// 예상치 못한 에러가 발생하면 해당 에러를 반환합니다.
func (s *Server) Start(ctx context.Context) error {
	// TODO: 구현하세요
	// 1. applyMiddlewares()로 미들웨어 체인 적용
	// 2. net.Listen("tcp", s.addr)으로 리스너 생성 후 s.listener에 저장
	// 3. s.httpServer 생성 및 OnShutdown 훅 등록
	// 4. 고루틴에서 s.httpServer.Serve(s.listener) 실행
	// 5. select { case <-ctx.Done(): ... case err := <-serverErr: ... }
	// 6. ctx 취소 시 Shutdown(context.WithTimeout 5초) 호출 후 nil 반환
	// 7. http.ErrServerClosed는 정상 종료이므로 nil 반환
	panic("구현 필요")
}

// Shutdown은 서버를 우아하게 종료합니다.
// 진행 중인 요청이 완료될 때까지 기다리거나 ctx 타임아웃 시 강제 종료합니다.
func (s *Server) Shutdown(ctx context.Context) error {
	// TODO: 구현하세요
	// 1. s.shuttingDown.Store(true)
	// 2. s.httpServer.Shutdown(ctx) 호출
	// 3. 에러 반환
	panic("구현 필요")
}

// applyMiddlewares는 등록된 미들웨어를 mux에 적용하여 최종 핸들러를 반환합니다.
// 내부 함수입니다.
func (s *Server) applyMiddlewares() http.Handler {
	// TODO: 구현하세요
	// var h http.Handler = s.mux
	// 역순으로 감싸야 등록 순서대로 실행됨:
	// for i := len(s.middlewares) - 1; i >= 0; i-- {
	//     h = s.middlewares[i](h)
	// }
	// return h
	panic("구현 필요")
}

// healthHandler는 /healthz 엔드포인트 핸들러입니다.
// 종료 중이 아니면 {"status":"ok"}, 종료 중이면 {"status":"shutting_down"}을 반환합니다.
func (s *Server) healthHandler() http.HandlerFunc {
	// TODO: 구현하세요
	// return func(w http.ResponseWriter, r *http.Request) {
	//     w.Header().Set("Content-Type", "application/json")
	//     if s.shuttingDown.Load() {
	//         w.WriteHeader(http.StatusServiceUnavailable)
	//         w.Write([]byte(`{"status":"shutting_down"}`))
	//     } else {
	//         w.Write([]byte(`{"status":"ok"}`))
	//     }
	// }
	panic("구현 필요")
}
