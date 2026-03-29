// client/main.go
// gRPC Greeter 클라이언트 구현
//
// 실행 방법:
//   go run client/main.go
//   go run client/main.go -name 홍길동 -lang ko
//
// 서버가 먼저 실행되어 있어야 합니다.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/curriculum/grpc-example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ============================================================
// CLI 플래그
// ============================================================

var (
	serverAddr = flag.String("addr", "localhost:50051", "gRPC 서버 주소")
	name       = flag.String("name", "Go학습자", "인사할 이름")
	language   = flag.String("lang", "ko", "언어 코드 (ko/en/ja/es)")
	streaming  = flag.Bool("stream", false, "서버 스트리밍 데모 실행")
	streamCnt  = flag.Int("count", 3, "스트리밍 메시지 수")
)

func main() {
	flag.Parse()

	// --------------------------------------------------------
	// gRPC 연결 생성
	// insecure.NewCredentials(): 개발 환경용 (TLS 없음)
	// 프로덕션에서는 TLS 인증서를 사용해야 합니다.
	// --------------------------------------------------------
	conn, err := grpc.NewClient(
		*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("서버 연결 실패: %v", err)
	}
	defer conn.Close()

	log.Printf("서버 연결 성공: %s", *serverAddr)

	// gRPC 클라이언트 스텁 생성
	client := pb.NewGreeterClient(conn)

	if *streaming {
		// 서버 스트리밍 RPC 데모
		runServerStreamingDemo(client)
	} else {
		// 단방향 RPC 데모
		runUnaryDemo(client)
	}
}

// ============================================================
// 단방향 RPC 데모
// ============================================================

// runUnaryDemo는 단순 SayHello RPC를 호출합니다.
func runUnaryDemo(client pb.GreeterClient) {
	fmt.Println("\n=== 단방향 RPC 데모 ===")

	// 컨텍스트: 5초 타임아웃 설정
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.HelloRequest{
		Name:     *name,
		Language: *language,
	}

	fmt.Printf("요청 전송: name=%s, language=%s\n", req.Name, req.Language)

	resp, err := client.SayHello(ctx, req)
	if err != nil {
		log.Fatalf("SayHello 호출 실패: %v", err)
	}

	fmt.Printf("서버 응답:\n")
	fmt.Printf("  메시지: %s\n", resp.Message)
	fmt.Printf("  타임스탬프: %s\n", resp.Timestamp)

	// 여러 언어로 테스트
	fmt.Println("\n--- 다국어 테스트 ---")
	langs := []string{"ko", "en", "ja", "es"}
	for _, lang := range langs {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)

		r, err := client.SayHello(ctx2, &pb.HelloRequest{
			Name:     *name,
			Language: lang,
		})
		cancel2()

		if err != nil {
			log.Printf("언어 %s 실패: %v", lang, err)
			continue
		}
		fmt.Printf("  [%s] %s\n", lang, r.Message)
	}
}

// ============================================================
// 서버 스트리밍 RPC 데모
// ============================================================

// runServerStreamingDemo는 서버 스트리밍 SayHelloServerStream을 호출합니다.
func runServerStreamingDemo(client pb.GreeterClient) {
	fmt.Println("\n=== 서버 스트리밍 RPC 데모 ===")

	// 스트리밍은 타임아웃 대신 취소 가능한 컨텍스트 사용
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := &pb.StreamRequest{
		Name:  *name,
		Count: int32(*streamCnt),
	}

	fmt.Printf("스트리밍 요청: name=%s, count=%d\n\n", req.Name, req.Count)

	// 스트림 열기
	stream, err := client.SayHelloServerStream(ctx, req)
	if err != nil {
		log.Fatalf("스트림 시작 실패: %v", err)
	}

	// 스트림에서 메시지를 하나씩 수신
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			// 서버가 스트림을 종료함
			fmt.Println("\n스트림 종료 (서버가 모든 메시지 전송 완료)")
			break
		}
		if err != nil {
			log.Fatalf("스트림 수신 오류: %v", err)
		}

		fmt.Printf("수신: %s (시각: %s)\n", msg.Message, msg.Timestamp)
	}
}
