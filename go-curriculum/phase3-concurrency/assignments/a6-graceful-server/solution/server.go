// 패키지 선언
// 참고 솔루션 - 풀기 전에 보지 마세요!
package graceful

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// Middleware는 HTTP 핸들러를 감싸는 미들웨어 함수 타입입니다.
type Middleware func(http.Handler) http.Handler

// Server는 우아한 종료를 지원하는 HTTP 서버 래퍼입니다.
type Server struct {
	addr          string
	mux           *http.ServeMux
	middlewares   []Middleware
	shutdownHooks []func()
	httpServer    *http.Server
	listener      net.Listener
	shuttingDown  atomic.Bool
}

// NewServer는 새 서버를 생성합니다.
func NewServer(addr string) *Server {
	s := &Server{
		addr:        addr,
		mux:         http.NewServeMux(),
		middlewares: []Middleware{},
	}
	// /healthz 핸들러 등록
	s.mux.HandleFunc("/healthz", s.healthHandler())
	return s
}

// Handle은 지정한 패턴에 핸들러를 등록합니다.
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

// Use는 미들웨어를 체인에 추가합니다.
func (s *Server) Use(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

// OnShutdown은 서버 종료 시 실행할 훅을 등록합니다.
func (s *Server) OnShutdown(hook func()) {
	s.shutdownHooks = append(s.shutdownHooks, hook)
}

// Addr은 서버가 실제로 바인딩된 주소를 반환합니다.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

// Start는 서버를 시작합니다.
func (s *Server) Start(ctx context.Context) error {
	// 미들웨어 체인 적용
	handler := s.applyMiddlewares()

	// 리스너 생성 (실제 포트 확인 가능)
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	// http.Server 생성
	s.httpServer = &http.Server{
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// OnShutdown 훅 등록
	for _, hook := range s.shutdownHooks {
		hook := hook
		s.httpServer.RegisterOnShutdown(hook)
	}

	// 서버 에러 채널
	serverErr := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		} else {
			serverErr <- nil
		}
	}()

	// ctx 취소 또는 서버 에러 대기
	select {
	case <-ctx.Done():
		// ctx 취소 시 우아한 종료
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.Shutdown(shutCtx) //nolint:errcheck
		<-serverErr
		return nil
	case err := <-serverErr:
		return err
	}
}

// Shutdown은 서버를 우아하게 종료합니다.
func (s *Server) Shutdown(ctx context.Context) error {
	s.shuttingDown.Store(true)
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// applyMiddlewares는 등록된 미들웨어를 적용하여 최종 핸들러를 반환합니다.
func (s *Server) applyMiddlewares() http.Handler {
	var h http.Handler = s.mux
	// 역순으로 감싸야 등록 순서대로 실행됨
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		h = s.middlewares[i](h)
	}
	return h
}

// healthHandler는 /healthz 엔드포인트 핸들러를 반환합니다.
func (s *Server) healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if s.shuttingDown.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"shutting_down"}`)) //nolint:errcheck
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
		}
	}
}
