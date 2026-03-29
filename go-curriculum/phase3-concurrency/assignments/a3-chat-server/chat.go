// 패키지 선언
package chat

import (
	"context"
	"errors"
	"time"
)

// MessageType은 메시지 종류를 나타냅니다.
type MessageType int

const (
	Broadcast MessageType = iota // 전체 브로드캐스트
	Private                      // 귓속말
	System                       // 시스템 알림 (입장/퇴장)
)

// Message는 채팅 메시지를 나타냅니다.
type Message struct {
	From    string      // 발신자
	To      string      // 수신자 ("" = 전체)
	Content string      // 내용
	Time    time.Time   // 전송 시각
	Type    MessageType // 메시지 종류
}

// 에러 상수들
var (
	ErrUserAlreadyExists = errors.New("이미 사용 중인 사용자명")
	ErrUserNotFound      = errors.New("사용자를 찾을 수 없음")
	ErrServerStopped     = errors.New("서버가 종료됨")
	ErrSelfMessage       = errors.New("자신에게 귓속말 불가")
)

// Client는 채팅 서버에 연결된 클라이언트입니다.
type Client struct {
	Username string
	Inbox    <-chan Message // 수신 메시지 채널 (읽기 전용)

	// TODO: 서버에 메시지를 보내기 위한 내부 채널 필드 추가
}

// Send는 모든 클라이언트에게 메시지를 브로드캐스트합니다.
func (c *Client) Send(msg string) error {
	// TODO: 구현하세요
	// - 서버의 send 채널로 Broadcast 메시지 전송
	panic("구현 필요")
}

// SendTo는 특정 클라이언트에게 귓속말을 보냅니다.
func (c *Client) SendTo(to, msg string) error {
	// TODO: 구현하세요
	// - to == c.Username이면 ErrSelfMessage 반환
	// - 서버의 send 채널로 Private 메시지 전송
	panic("구현 필요")
}

// ChatServer는 채널 기반 채팅 서버입니다.
type ChatServer struct {
	// TODO: 필요한 필드를 추가하세요
	// 힌트:
	// - clients map[string]*clientState: 연결된 클라이언트 맵
	// - connect chan connectReq: 접속 요청 채널
	// - disconnect chan disconnectReq: 해제 요청 채널
	// - send chan Message: 메시지 전송 채널
	// - stopped bool: 서버 종료 여부
}

// NewChatServer는 새 채팅 서버를 생성합니다.
func NewChatServer() *ChatServer {
	// TODO: 구현하세요
	panic("구현 필요")
}

// Start는 서버를 시작합니다.
// ctx가 취소되면 서버와 모든 클라이언트가 종료됩니다.
// 이 함수는 별도의 고루틴에서 실행되거나, 고루틴을 내부에서 시작해야 합니다.
func (s *ChatServer) Start(ctx context.Context) {
	// TODO: 구현하세요
	// 메인 이벤트 루프:
	// for {
	//   select {
	//   case req := <-s.connect:   // 접속 처리
	//   case req := <-s.disconnect: // 해제 처리
	//   case msg := <-s.send:       // 메시지 라우팅
	//   case <-ctx.Done():          // 서버 종료
	//   }
	// }
	panic("구현 필요")
}

// Connect는 새 클라이언트를 서버에 연결합니다.
// 이미 사용 중인 username이면 ErrUserAlreadyExists를 반환합니다.
func (s *ChatServer) Connect(username string) (*Client, error) {
	// TODO: 구현하세요
	// - 서버 루프에 접속 요청 전송
	// - 응답(Client 또는 에러) 대기
	panic("구현 필요")
}

// Disconnect는 클라이언트를 서버에서 해제합니다.
func (s *ChatServer) Disconnect(username string) error {
	// TODO: 구현하세요
	// - 서버 루프에 해제 요청 전송
	// - 응답 대기
	panic("구현 필요")
}

// UserCount는 현재 연결된 클라이언트 수를 반환합니다.
func (s *ChatServer) UserCount() int {
	// TODO: 구현하세요
	panic("구현 필요")
}

// UserList는 현재 연결된 사용자명 목록을 반환합니다.
func (s *ChatServer) UserList() []string {
	// TODO: 구현하세요
	panic("구현 필요")
}
