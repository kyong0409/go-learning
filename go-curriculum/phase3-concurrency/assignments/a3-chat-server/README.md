# 과제 A3: 채널 기반 채팅 서버

## 과제 설명

채널과 고루틴으로 멀티 클라이언트 채팅 서버를 구현하세요.
네트워크 없이 순수하게 채널로 메시지를 라우팅합니다.

## 요구사항

### ChatServer 구조체

```go
type ChatServer struct { ... }

func NewChatServer() *ChatServer
func (s *ChatServer) Start(ctx context.Context)
func (s *ChatServer) Connect(username string) (*Client, error)
func (s *ChatServer) Disconnect(username string) error
func (s *ChatServer) UserCount() int
func (s *ChatServer) UserList() []string
```

### Client 구조체

```go
type Client struct {
    Username string
    Inbox    <-chan Message  // 수신 메시지
}

func (c *Client) Send(msg string) error           // 브로드캐스트
func (c *Client) SendTo(to, msg string) error     // 귓속말
```

### Message 타입

```go
type Message struct {
    From    string
    To      string  // "" = 브로드캐스트
    Content string
    Time    time.Time
    Type    MessageType
}

type MessageType int
const (
    Broadcast  MessageType = iota
    Private
    System      // 입장/퇴장 알림
)
```

## 기능 요구사항

1. **입장/퇴장**: `Connect`/`Disconnect`로 클라이언트 관리
2. **브로드캐스트**: 한 명이 보내면 모든 클라이언트 수신 (자신 포함 여부 자유)
3. **귓속말**: 특정 클라이언트에게만 전송
4. **입장/퇴장 알림**: 다른 클라이언트들에게 시스템 메시지 전송
5. **중복 사용자명 방지**: 동일 username 재접속 시 에러 반환
6. **Context 취소**: ctx 취소 시 서버와 모든 클라이언트 종료

## 실행 방법

```bash
cd a3-chat-server
go test -v .
go test -race -v .
go test -v -run TestGrade .
```

## 채점 기준 (100점)

| 항목 | 점수 | 설명 |
|------|------|------|
| 연결/해제 | 15점 | Connect/Disconnect 정상 작동 |
| 브로드캐스트 | 25점 | 모든 클라이언트에게 메시지 전달 |
| 귓속말 | 20점 | 특정 클라이언트에게만 전달 |
| 시스템 알림 | 15점 | 입장/퇴장 시 알림 |
| 동시성 안전 | 15점 | 레이스 컨디션 없음 |
| 고루틴 누수 없음 | 10점 | 종료 후 고루틴 정리 |

## 힌트

- 서버의 메인 루프를 단일 고루틴에서 실행하면 뮤텍스 없이 클라이언트 맵 관리 가능
- 각 클라이언트의 `Inbox`는 버퍼 채널로 만들어 블로킹 방지
- `Connect()`/`Disconnect()`는 서버 루프에 요청 채널로 전달해 직렬화
- 브로드캐스트는 모든 클라이언트의 Inbox 채널로 전송
- Context 취소 시 서버 루프가 종료되면 모든 클라이언트 채널도 닫기
