# 과제 A6: 우아한 종료가 되는 HTTP 서버

## 과제 설명

프로덕션 수준의 HTTP 서버 래퍼를 구현하세요.
우아한 종료(graceful shutdown), 헬스 체크, 미들웨어 체인을 지원합니다.

## 요구사항

### Server 구조체

```go
type Server struct { ... }

func NewServer(addr string) *Server
func (s *Server) Handle(pattern string, handler http.Handler)
func (s *Server) Use(middleware Middleware)
func (s *Server) OnShutdown(hook func())
func (s *Server) Start(ctx context.Context) error
func (s *Server) Shutdown(ctx context.Context) error
func (s *Server) Addr() string
```

### 타입 정의

```go
// Middleware는 HTTP 핸들러를 감싸는 함수입니다.
type Middleware func(http.Handler) http.Handler
```

### 기능 요구사항

1. **우아한 종료**: `Shutdown(ctx)`는 진행 중인 요청이 완료될 때까지 기다린 후 서버를 종료합니다. `ctx` 타임아웃 내에 완료되지 않으면 강제 종료합니다.

2. **자동 헬스 체크**: `/healthz` 경로로 접근하면 `{"status":"ok"}` JSON을 반환합니다. 서버가 종료 중이면 `{"status":"shutting_down"}`을 반환합니다.

3. **미들웨어 체인**: `Use(middleware)`로 미들웨어를 등록합니다. 등록 순서대로 적용됩니다 (첫 번째 등록이 가장 바깥쪽).

4. **OnShutdown 훅**: 서버가 종료되기 전에 등록된 훅들을 순서대로 실행합니다.

5. **Start 동작**: `Start(ctx)`는 서버를 시작하고 `ctx`가 취소되면 자동으로 `Shutdown`을 호출합니다. 서버가 정상 종료되면 `nil`을 반환합니다.

6. **Addr()**: 실제 바인딩된 주소를 반환합니다 (`:0`으로 시작한 경우 OS가 할당한 포트).

## 실행 방법

```bash
cd a6-graceful-server
go test -v .
go test -race -v .
go test -v -run TestGrade .
```

## 채점 기준 (100점)

| 항목 | 점수 | 설명 |
|------|------|------|
| 기본 서버 시작/종료 | 20점 | Start/Shutdown 정상 작동 |
| 헬스 체크 엔드포인트 | 15점 | /healthz 응답 |
| 미들웨어 체인 | 20점 | Use()로 등록한 미들웨어 적용 |
| 우아한 종료 | 20점 | 진행 중 요청 완료 후 종료 |
| OnShutdown 훅 | 15점 | 종료 시 훅 실행 |
| 고루틴 누수 없음 | 10점 | 종료 후 고루틴 정리 |

## 힌트

- `net/http.Server`의 `Shutdown(ctx)` 메서드를 내부적으로 활용하세요.
- 종료 중 여부를 `atomic.Bool`로 추적하면 헬스 체크에서 상태를 반환할 수 있습니다.
- `net.Listen`으로 리스너를 먼저 생성하면 `:0` 포트를 사용할 때 실제 포트를 알 수 있습니다.
- 미들웨어는 `http.Handler`를 감싸는 함수입니다: `func(next http.Handler) http.Handler`
- `Start(ctx)`는 내부 고루틴에서 `httpServer.Serve(listener)`를 실행하고, ctx 취소 신호를 감지해 종료를 트리거하세요.
- `OnShutdown` 훅은 `net/http.Server.RegisterOnShutdown`을 활용할 수 있습니다.
