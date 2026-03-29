// 패키지 선언
// 참고 솔루션 - 풀기 전에 보지 마세요!
package chat

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// MessageType은 메시지 종류를 나타냅니다.
type MessageType int

const (
	Broadcast MessageType = iota
	Private
	System
)

// Message는 채팅 메시지를 나타냅니다.
type Message struct {
	From    string
	To      string
	Content string
	Time    time.Time
	Type    MessageType
}

var (
	ErrUserAlreadyExists = errors.New("이미 사용 중인 사용자명")
	ErrUserNotFound      = errors.New("사용자를 찾을 수 없음")
	ErrServerStopped     = errors.New("서버가 종료됨")
	ErrSelfMessage       = errors.New("자신에게 귓속말 불가")
)

// Client는 채팅 클라이언트입니다.
type Client struct {
	Username string
	Inbox    <-chan Message

	inbox  chan Message // 쓰기용 내부 채널
	sendCh chan<- Message // 서버로 메시지 전송
	server *ChatServer
}

// Send는 브로드캐스트 메시지를 전송합니다.
func (c *Client) Send(msg string) error {
	m := Message{
		From:    c.Username,
		Content: msg,
		Time:    time.Now(),
		Type:    Broadcast,
	}
	select {
	case c.sendCh <- m:
		return nil
	default:
		c.sendCh <- m
		return nil
	}
}

// SendTo는 귓속말을 전송합니다.
func (c *Client) SendTo(to, msg string) error {
	if to == c.Username {
		return ErrSelfMessage
	}

	// 수신자 존재 확인은 서버가 처리
	m := Message{
		From:    c.Username,
		To:      to,
		Content: msg,
		Time:    time.Now(),
		Type:    Private,
	}

	// 서버에 귓속말 요청 (응답 채널 포함)
	respCh := make(chan error, 1)
	req := sendPrivateReq{msg: m, resp: respCh}
	select {
	case c.server.privateReqs <- req:
	case <-c.server.stopped:
		return ErrServerStopped
	}
	return <-respCh
}

// ─────────────────────────────────────────
// 내부 요청 타입들
// ─────────────────────────────────────────

type connectReq struct {
	username string
	resp     chan connectResp
}

type connectResp struct {
	client *Client
	err    error
}

type disconnectReq struct {
	username string
	resp     chan error
}

type sendPrivateReq struct {
	msg  Message
	resp chan error
}

type userCountReq struct {
	resp chan int
}

type userListReq struct {
	resp chan []string
}

// ─────────────────────────────────────────
// ChatServer
// ─────────────────────────────────────────

// clientState는 서버 내부의 클라이언트 상태입니다.
type clientState struct {
	client *Client
}

// ChatServer는 채널 기반 채팅 서버입니다.
type ChatServer struct {
	clients     map[string]*clientState
	connectReqs chan connectReq
	disconnReqs chan disconnectReq
	broadcast   chan Message
	privateReqs chan sendPrivateReq
	countReqs   chan userCountReq
	listReqs    chan userListReq
	stopped     chan struct{}
	stopOnce    sync.Once
}

// NewChatServer는 새 채팅 서버를 생성합니다.
func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:     make(map[string]*clientState),
		connectReqs: make(chan connectReq, 10),
		disconnReqs: make(chan disconnectReq, 10),
		broadcast:   make(chan Message, 100),
		privateReqs: make(chan sendPrivateReq, 100),
		countReqs:   make(chan userCountReq, 10),
		listReqs:    make(chan userListReq, 10),
		stopped:     make(chan struct{}),
	}
}

// Start는 서버 이벤트 루프를 시작합니다.
func (s *ChatServer) Start(ctx context.Context) {
	go func() {
		defer s.stopOnce.Do(func() {
			close(s.stopped)
			// 모든 클라이언트 Inbox 닫기
			for _, cs := range s.clients {
				close(cs.client.inbox)
			}
			s.clients = make(map[string]*clientState)
		})

		for {
			select {
			case <-ctx.Done():
				return

			case req := <-s.connectReqs:
				if _, exists := s.clients[req.username]; exists {
					req.resp <- connectResp{err: ErrUserAlreadyExists}
					continue
				}

				inbox := make(chan Message, 50)
				client := &Client{
					Username: req.username,
					Inbox:    inbox,
					inbox:    inbox,
					sendCh:   s.broadcast,
					server:   s,
				}
				s.clients[req.username] = &clientState{client: client}
				req.resp <- connectResp{client: client}

				// 입장 알림 브로드캐스트 (다른 사용자에게)
				sysMsg := Message{
					From:    "서버",
					Content: fmt.Sprintf("%s님이 입장했습니다.", req.username),
					Time:    time.Now(),
					Type:    System,
				}
				s.broadcastToOthers(req.username, sysMsg)

			case req := <-s.disconnReqs:
				cs, exists := s.clients[req.username]
				if !exists {
					req.resp <- ErrUserNotFound
					continue
				}
				delete(s.clients, req.username)
				close(cs.client.inbox)
				req.resp <- nil

				// 퇴장 알림 브로드캐스트
				sysMsg := Message{
					From:    "서버",
					Content: fmt.Sprintf("%s님이 퇴장했습니다.", req.username),
					Time:    time.Now(),
					Type:    System,
				}
				s.broadcastToAll(sysMsg)

			case msg := <-s.broadcast:
				// 브로드캐스트: 모든 클라이언트에게 전송
				for _, cs := range s.clients {
					select {
					case cs.client.inbox <- msg:
					default:
						// 가득 찬 경우 건너뜀 (블로킹 방지)
					}
				}

			case req := <-s.privateReqs:
				// 귓속말 라우팅
				target, exists := s.clients[req.msg.To]
				if !exists {
					req.resp <- ErrUserNotFound
					continue
				}
				select {
				case target.client.inbox <- req.msg:
					req.resp <- nil
				default:
					target.client.inbox <- req.msg
					req.resp <- nil
				}

			case req := <-s.countReqs:
				req.resp <- len(s.clients)

			case req := <-s.listReqs:
				list := make([]string, 0, len(s.clients))
				for name := range s.clients {
					list = append(list, name)
				}
				req.resp <- list
			}
		}
	}()
}

// broadcastToOthers는 특정 사용자를 제외한 모든 클라이언트에게 전송합니다.
func (s *ChatServer) broadcastToOthers(exclude string, msg Message) {
	for name, cs := range s.clients {
		if name == exclude {
			continue
		}
		select {
		case cs.client.inbox <- msg:
		default:
		}
	}
}

// broadcastToAll은 모든 클라이언트에게 전송합니다.
func (s *ChatServer) broadcastToAll(msg Message) {
	for _, cs := range s.clients {
		select {
		case cs.client.inbox <- msg:
		default:
		}
	}
}

// Connect는 새 클라이언트를 연결합니다.
func (s *ChatServer) Connect(username string) (*Client, error) {
	resp := make(chan connectResp, 1)
	req := connectReq{username: username, resp: resp}
	select {
	case s.connectReqs <- req:
	case <-s.stopped:
		return nil, ErrServerStopped
	}
	r := <-resp
	return r.client, r.err
}

// Disconnect는 클라이언트를 해제합니다.
func (s *ChatServer) Disconnect(username string) error {
	resp := make(chan error, 1)
	req := disconnectReq{username: username, resp: resp}
	select {
	case s.disconnReqs <- req:
	case <-s.stopped:
		return ErrServerStopped
	}
	return <-resp
}

// UserCount는 현재 연결된 사용자 수를 반환합니다.
func (s *ChatServer) UserCount() int {
	resp := make(chan int, 1)
	select {
	case s.countReqs <- userCountReq{resp: resp}:
	case <-s.stopped:
		return 0
	}
	return <-resp
}

// UserList는 사용자명 목록을 반환합니다.
func (s *ChatServer) UserList() []string {
	resp := make(chan []string, 1)
	select {
	case s.listReqs <- userListReq{resp: resp}:
	case <-s.stopped:
		return nil
	}
	return <-resp
}
