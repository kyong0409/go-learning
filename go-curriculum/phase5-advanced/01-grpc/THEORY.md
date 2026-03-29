# gRPC - 고성능 RPC 프레임워크

## gRPC란?

gRPC는 Google이 개발한 오픈소스 RPC(Remote Procedure Call) 프레임워크입니다.
2015년 공개된 이후 마이크로서비스 아키텍처의 표준 통신 방식으로 자리잡았습니다.

**핵심 구성 요소:**
- **전송 계층**: HTTP/2
- **직렬화 형식**: Protocol Buffers (Protobuf)
- **인터페이스 정의**: .proto 파일
- **코드 생성**: protoc 컴파일러

---

## HTTP/2 기반 전송

gRPC가 HTTP/1.1 대신 HTTP/2를 사용하는 이유:

```
HTTP/1.1 문제점:
- 요청당 하나의 커넥션 (또는 파이프라이닝 한계)
- 헤더 중복 전송 (쿠키, Content-Type 매번)
- HOL(Head-of-Line) 블로킹

HTTP/2 해결책:
- 멀티플렉싱: 단일 커넥션에서 여러 스트림 동시 처리
- 헤더 압축 (HPACK): 반복 헤더 90% 이상 감소
- 서버 푸시: 클라이언트 요청 없이 데이터 전송
- 바이너리 프레이밍: 텍스트 대신 바이너리로 효율적 전송
```

---

## Protocol Buffers (Protobuf)

### .proto 파일 문법

```protobuf
// proto/bookmark.proto
syntax = "proto3";

package bookmark;
option go_package = "github.com/myproject/proto/bookmark";

// message: 데이터 구조 정의 (JSON의 object와 유사)
message Bookmark {
  int64 id = 1;          // 필드 번호 (직렬화에 사용, 변경 불가)
  string url = 2;
  string title = 3;
  repeated string tags = 4;  // 배열/슬라이스
  bool is_public = 5;
  int64 created_at = 6;      // Unix timestamp
}

message CreateBookmarkRequest {
  string url = 1;
  string title = 2;
  repeated string tags = 3;
}

message CreateBookmarkResponse {
  Bookmark bookmark = 1;
}

message ListBookmarksRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message ListBookmarksResponse {
  repeated Bookmark bookmarks = 1;
  string next_page_token = 2;
}

// service: RPC 메서드 정의
service BookmarkService {
  // Unary RPC
  rpc CreateBookmark(CreateBookmarkRequest) returns (CreateBookmarkResponse);

  // Server Streaming RPC
  rpc ListBookmarks(ListBookmarksRequest) returns (stream Bookmark);

  // Client Streaming RPC
  rpc BulkCreate(stream CreateBookmarkRequest) returns (CreateBookmarkResponse);

  // Bidirectional Streaming RPC
  rpc Sync(stream Bookmark) returns (stream Bookmark);
}
```

### 직렬화 성능

```
JSON (텍스트):
{"id":1,"url":"https://example.com","title":"Example","tags":["go","grpc"]}
→ 약 74 바이트

Protobuf (바이너리):
→ 약 35 바이트 (53% 감소)

성능 차이:
- 크기: JSON 대비 3~10배 작음
- 직렬화 속도: JSON 대비 5~7배 빠름
- 역직렬화 속도: JSON 대비 6~10배 빠름
```

### protoc 컴파일러 사용

```bash
# 설치
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# .proto → .pb.go 생성
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/bookmark.proto

# 생성되는 파일:
# proto/bookmark.pb.go       - 메시지 타입 (직렬화/역직렬화)
# proto/bookmark_grpc.pb.go  - 서버/클라이언트 인터페이스
```

---

## 4가지 통신 패턴

### 1. Unary RPC (단순 요청-응답)

REST와 가장 유사한 패턴입니다.

```go
// 서버 구현
func (s *BookmarkServer) CreateBookmark(
    ctx context.Context,
    req *pb.CreateBookmarkRequest,
) (*pb.CreateBookmarkResponse, error) {
    bookmark := &pb.Bookmark{
        Id:    generateID(),
        Url:   req.Url,
        Title: req.Title,
        Tags:  req.Tags,
    }
    if err := s.store.Save(bookmark); err != nil {
        return nil, status.Errorf(codes.Internal, "저장 실패: %v", err)
    }
    return &pb.CreateBookmarkResponse{Bookmark: bookmark}, nil
}

// 클라이언트 호출
resp, err := client.CreateBookmark(ctx, &pb.CreateBookmarkRequest{
    Url:   "https://golang.org",
    Title: "Go 언어 공식 사이트",
    Tags:  []string{"go", "programming"},
})
```

### 2. Server Streaming RPC (서버가 스트림 반환)

서버가 여러 응답을 순차적으로 전송합니다. 로그 스트리밍, 실시간 업데이트에 적합합니다.

```go
// 서버 구현
func (s *BookmarkServer) ListBookmarks(
    req *pb.ListBookmarksRequest,
    stream pb.BookmarkService_ListBookmarksServer,
) error {
    bookmarks, err := s.store.List()
    if err != nil {
        return status.Errorf(codes.Internal, "조회 실패: %v", err)
    }
    for _, b := range bookmarks {
        // 스트림에 하나씩 전송
        if err := stream.Send(b); err != nil {
            return err
        }
    }
    return nil
}

// 클라이언트 수신
stream, err := client.ListBookmarks(ctx, &pb.ListBookmarksRequest{PageSize: 100})
for {
    bookmark, err := stream.Recv()
    if err == io.EOF {
        break // 스트림 종료
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(bookmark.Title)
}
```

### 3. Client Streaming RPC (클라이언트가 스트림 전송)

클라이언트가 여러 요청을 보내고 서버가 하나의 응답을 반환합니다. 파일 업로드, 배치 처리에 적합합니다.

```go
// 서버 구현
func (s *BookmarkServer) BulkCreate(
    stream pb.BookmarkService_BulkCreateServer,
) error {
    var count int32
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            // 모든 요청을 받았으면 최종 응답 전송
            return stream.SendAndClose(&pb.CreateBookmarkResponse{
                // 결과 반환
            })
        }
        if err != nil {
            return err
        }
        s.store.Save(&pb.Bookmark{Url: req.Url, Title: req.Title})
        count++
    }
}

// 클라이언트 전송
stream, _ := client.BulkCreate(ctx)
for _, url := range urls {
    stream.Send(&pb.CreateBookmarkRequest{Url: url})
}
resp, err := stream.CloseAndRecv()
```

### 4. Bidirectional Streaming RPC (양방향 스트리밍)

채팅, 실시간 동기화 등 양방향 실시간 통신에 사용합니다.

```go
// 서버 구현
func (s *BookmarkServer) Sync(stream pb.BookmarkService_SyncServer) error {
    for {
        in, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }
        // 처리 후 응답 전송
        updated := s.syncBookmark(in)
        if err := stream.Send(updated); err != nil {
            return err
        }
    }
}
```

---

## gRPC 에러 처리

gRPC는 HTTP 상태 코드 대신 `codes` 패키지의 코드를 사용합니다.

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// 에러 반환
return nil, status.Errorf(codes.NotFound, "북마크 %d를 찾을 수 없습니다", id)
return nil, status.Errorf(codes.InvalidArgument, "URL이 비어있습니다")
return nil, status.Errorf(codes.Internal, "데이터베이스 오류: %v", err)
return nil, status.Errorf(codes.Unauthenticated, "인증이 필요합니다")
return nil, status.Errorf(codes.PermissionDenied, "권한이 없습니다")
return nil, status.Errorf(codes.AlreadyExists, "이미 존재하는 북마크입니다")
return nil, status.Errorf(codes.ResourceExhausted, "요청 한도 초과")

// 클라이언트에서 에러 처리
_, err := client.CreateBookmark(ctx, req)
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.NotFound:
            fmt.Println("리소스 없음:", st.Message())
        case codes.InvalidArgument:
            fmt.Println("잘못된 입력:", st.Message())
        default:
            fmt.Println("기타 오류:", st.Code(), st.Message())
        }
    }
}
```

---

## 인터셉터 (Interceptor)

HTTP 미들웨어와 동일한 개념입니다. 로깅, 인증, 재시도 등을 AOP 방식으로 처리합니다.

```go
// 서버 단항 인터셉터: 로깅
func loggingInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    start := time.Now()
    resp, err := handler(ctx, req) // 실제 핸들러 호출
    log.Printf("메서드: %s, 소요시간: %v, 에러: %v",
        info.FullMethod, time.Since(start), err)
    return resp, err
}

// 서버에 인터셉터 등록
s := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
    // 여러 인터셉터: grpc.ChainUnaryInterceptor(i1, i2, i3)
)
```

---

## 서비스 간 통신에서 gRPC의 위치

```
[외부 클라이언트]
     │
     │ HTTPS + JSON (REST)
     ▼
[API Gateway]  ← gRPC-Gateway로 REST ↔ gRPC 변환 가능
     │
     │ gRPC (HTTP/2 + Protobuf)
  ┌──┴──────────┐
  ▼             ▼
[User Service] [Bookmark Service]
  │
  │ gRPC
  ▼
[Auth Service]
```

**실전 사용 사례:**
- Kubernetes API Server: etcd와 gRPC 통신
- etcd: 클러스터 노드 간 gRPC 사용
- CockroachDB: 노드 간 데이터 복제에 gRPC 사용
- Prometheus: 일부 익스포터가 gRPC 사용

---

## 서버/클라이언트 기본 구조

```go
// 서버 시작
func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("리슨 실패: %v", err)
    }

    s := grpc.NewServer()
    pb.RegisterBookmarkServiceServer(s, &BookmarkServer{})
    reflection.Register(s) // grpcurl 등 디버깅 도구용

    log.Println("gRPC 서버 시작: :50051")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("서버 실패: %v", err)
    }
}

// 클라이언트 연결
func main() {
    conn, err := grpc.NewClient("localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("연결 실패: %v", err)
    }
    defer conn.Close()

    client := pb.NewBookmarkServiceClient(conn)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // RPC 호출
    resp, err := client.CreateBookmark(ctx, &pb.CreateBookmarkRequest{
        Url:   "https://golang.org",
        Title: "Go 공식 사이트",
    })
}
```
