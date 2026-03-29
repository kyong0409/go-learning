# 01-grpc: gRPC 서비스 구현

Go에서 gRPC를 사용해 고성능 원격 프로시저 호출(RPC) 서비스를 구현하는 방법을 학습합니다.

## 개요

**gRPC**는 Google이 개발한 고성능 오픈소스 RPC 프레임워크입니다.
HTTP/2 기반이며 Protocol Buffers를 직렬화 형식으로 사용합니다.

### REST vs gRPC 비교

| 항목 | REST | gRPC |
|------|------|------|
| 프로토콜 | HTTP/1.1 | HTTP/2 |
| 직렬화 | JSON (텍스트) | Protocol Buffers (바이너리) |
| 타입 안전성 | 없음 | 있음 (proto 스키마) |
| 스트리밍 | 제한적 | 완전 지원 |
| 언어 지원 | 모든 언어 | 20+ 언어 공식 지원 |
| 성능 | 보통 | 매우 빠름 |

## 프로젝트 구조

```
01-grpc/
├── proto/
│   └── greeter.proto         # 서비스 정의 (Protocol Buffers)
├── proto/                    # 생성된 Go 코드 (protoc 실행 후)
│   ├── greeter.pb.go         # 메시지 타입 (자동 생성)
│   └── greeter_grpc.pb.go    # 서비스 인터페이스 (자동 생성)
├── server/
│   └── main.go               # gRPC 서버
├── client/
│   └── main.go               # gRPC 클라이언트
└── go.mod
```

## 설치 사전 요구 사항

### 1. protoc 컴파일러 설치

```bash
# macOS
brew install protobuf

# Ubuntu/Debian
apt install protobuf-compiler

# Windows (winget 사용)
winget install protobuf
```

### 2. Go 플러그인 설치

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

`$GOPATH/bin`이 `PATH`에 포함되어 있는지 확인하세요.

## 코드 생성

```bash
# 프로젝트 루트(01-grpc/)에서 실행
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/greeter.proto
```

실행 후 `proto/` 디렉터리에 두 파일이 생성됩니다:
- `greeter.pb.go`: HelloRequest, HelloReply 등의 메시지 구조체
- `greeter_grpc.pb.go`: GreeterServer/GreeterClient 인터페이스

## 의존성 설치

```bash
cd 01-grpc
go mod tidy
```

## 실행 방법

### 터미널 1 - 서버 시작

```bash
go run server/main.go
```

출력:
```
2024/01/01 12:00:00 gRPC 서버 시작: :50051
```

### 터미널 2 - 클라이언트 실행

```bash
# 기본 실행 (한국어 인사)
go run client/main.go

# 이름과 언어 지정
go run client/main.go -name 홍길동 -lang ko

# 서버 스트리밍 데모
go run client/main.go -stream -count 5 -name 김철수
```

## 주요 개념

### 1. RPC 유형

```
단방향 RPC:       클라이언트 → 요청 1개 → 서버 → 응답 1개 → 클라이언트
서버 스트리밍:    클라이언트 → 요청 1개 → 서버 → 응답 N개 → 클라이언트
클라이언트 스트리밍: 클라이언트 → 요청 N개 → 서버 → 응답 1개 → 클라이언트
양방향 스트리밍:  클라이언트 ↔ 요청/응답 N개 ↔ 서버
```

### 2. 인터셉터 (미들웨어)

gRPC 인터셉터는 HTTP 미들웨어와 동일한 역할입니다:
```go
grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
    grpc.StreamInterceptor(streamInterceptor),
)
```

### 3. 컨텍스트와 타임아웃

```go
// 5초 타임아웃
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

resp, err := client.SayHello(ctx, req)
```

### 4. grpcurl로 서비스 테스트 (선택사항)

```bash
# 설치
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 서비스 목록 조회
grpcurl -plaintext localhost:50051 list

# SayHello 호출
grpcurl -plaintext -d '{"name":"테스트","language":"ko"}' \
  localhost:50051 greeter.Greeter/SayHello
```

## 학습 포인트

1. `.proto` 파일로 API 스키마를 먼저 정의한다 (스키마 우선 설계)
2. `protoc`가 Go 코드를 자동 생성한다 (보일러플레이트 없음)
3. 서버는 생성된 인터페이스를 구현하면 된다
4. `UnimplementedGreeterServer` 임베딩으로 미래 확장성 확보
5. 인터셉터로 로깅, 인증, 메트릭을 횡단 관심사로 처리
