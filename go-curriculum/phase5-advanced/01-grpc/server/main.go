// server/main.go
// gRPC Greeter 서버 구현
//
// 실행 방법:
//   go run server/main.go
//
// 서버는 기본적으로 :50051 포트에서 대기합니다.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	// proto 패키지는 protoc로 생성됩니다.
	// 생성 전에는 아래 주석을 참고해 코드를 생성하세요.
	// protoc --go_out=. --go_opt=paths=source_relative \
	//        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	//        proto/greeter.proto
	pb "github.com/curriculum/grpc-example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// ============================================================
// 상수 정의
// ============================================================

const (
	port = ":50051" // gRPC 서버 포트
)

// ============================================================
// 서버 구현체
// ============================================================

// greeterServer는 pb.GreeterServer 인터페이스를 구현합니다.
// pb.UnimplementedGreeterServer를 임베딩해 미구현 메서드를 안전하게 처리합니다.
type greeterServer struct {
	pb.UnimplementedGreeterServer
}

// SayHello는 단순 단방향 RPC를 처리합니다.
// 클라이언트의 HelloRequest를 받아 HelloReply를 반환합니다.
func (s *greeterServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("SayHello 호출 수신: name=%s, language=%s", req.Name, req.Language)

	// 언어에 따라 다른 인사말 생성
	var greeting string
	switch req.Language {
	case "ko":
		greeting = fmt.Sprintf("안녕하세요, %s님!", req.Name)
	case "ja":
		greeting = fmt.Sprintf("こんにちは、%sさん！", req.Name)
	case "es":
		greeting = fmt.Sprintf("¡Hola, %s!", req.Name)
	default:
		// 기본값은 영어
		greeting = fmt.Sprintf("Hello, %s!", req.Name)
	}

	return &pb.HelloReply{
		Message:   greeting,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// SayHelloServerStream은 서버 스트리밍 RPC를 처리합니다.
// 요청된 횟수만큼 메시지를 스트리밍합니다.
func (s *greeterServer) SayHelloServerStream(req *pb.StreamRequest, stream pb.Greeter_SayHelloServerStreamServer) error {
	log.Printf("SayHelloServerStream 호출 수신: name=%s, count=%d", req.Name, req.Count)

	for i := 0; i < int(req.Count); i++ {
		// 컨텍스트 취소 확인 (클라이언트가 연결을 끊었는지 체크)
		if err := stream.Context().Err(); err != nil {
			log.Printf("스트림 컨텍스트 취소: %v", err)
			return err
		}

		reply := &pb.HelloReply{
			Message:   fmt.Sprintf("[%d/%d] Hello, %s!", i+1, req.Count, req.Name),
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := stream.Send(reply); err != nil {
			return fmt.Errorf("스트림 전송 실패: %w", err)
		}

		// 메시지 간 짧은 딜레이 (시연용)
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("스트리밍 완료: %d개 메시지 전송", req.Count)
	return nil
}

// ============================================================
// 메인 함수
// ============================================================

func main() {
	// TCP 리스너 생성
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("리스너 생성 실패: %v", err)
	}

	// gRPC 서버 생성 (인터셉터 포함)
	s := grpc.NewServer(
		// 로깅 인터셉터 추가
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	// Greeter 서비스 등록
	pb.RegisterGreeterServer(s, &greeterServer{})

	// reflection 등록: grpcurl 같은 도구로 서비스 탐색 가능
	reflection.Register(s)

	log.Printf("gRPC 서버 시작: %s", port)
	log.Printf("서버 종료하려면 Ctrl+C를 누르세요")

	// 서버 시작 (블로킹)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("서버 실행 실패: %v", err)
	}
}

// ============================================================
// 인터셉터 (미들웨어)
// ============================================================

// loggingInterceptor는 모든 단방향 RPC 호출을 로깅하는 인터셉터입니다.
// gRPC의 인터셉터는 HTTP 미들웨어와 유사한 개념입니다.
func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	// 다음 핸들러 호출
	resp, err := handler(ctx, req)

	// 결과 로깅
	duration := time.Since(start)
	if err != nil {
		log.Printf("[gRPC] %s | 오류: %v | 소요시간: %v", info.FullMethod, err, duration)
	} else {
		log.Printf("[gRPC] %s | 성공 | 소요시간: %v", info.FullMethod, duration)
	}

	return resp, err
}
