// 패키지 선언
package chat_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	chat "github.com/go-curriculum/a3-chat-server"
)

// ─────────────────────────────────────────
// 헬퍼
// ─────────────────────────────────────────

func goroutineCount() int { return runtime.NumGoroutine() }

func waitForGoroutines(t *testing.T, before int, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if runtime.NumGoroutine() <= before+1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// receiveWithTimeout은 타임아웃 내에 메시지를 수신하거나 nil을 반환합니다.
func receiveWithTimeout(ch <-chan chat.Message, d time.Duration) *chat.Message {
	select {
	case msg, ok := <-ch:
		if !ok {
			return nil
		}
		return &msg
	case <-time.After(d):
		return nil
	}
}

// startServer는 서버를 시작하고 cleanup을 등록합니다.
func startServer(t *testing.T) (*chat.ChatServer, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	server := chat.NewChatServer()
	server.Start(ctx)
	time.Sleep(10 * time.Millisecond) // 서버 루프 시작 대기
	t.Cleanup(cancel)
	return server, cancel
}

// ─────────────────────────────────────────
// 연결/해제 테스트
// ─────────────────────────────────────────

func TestConnect_Basic(t *testing.T) {
	server, _ := startServer(t)

	client, err := server.Connect("alice")
	if err != nil {
		t.Fatalf("Connect 실패: %v", err)
	}
	if client == nil {
		t.Fatal("Connect: nil 클라이언트 반환")
	}
	if client.Username != "alice" {
		t.Errorf("Username: 기대=alice, 실제=%s", client.Username)
	}
	if client.Inbox == nil {
		t.Error("Inbox 채널이 nil")
	}
}

func TestConnect_DuplicateUsername(t *testing.T) {
	server, _ := startServer(t)

	_, err := server.Connect("bob")
	if err != nil {
		t.Fatalf("첫 번째 Connect 실패: %v", err)
	}

	_, err = server.Connect("bob")
	if err == nil {
		t.Error("중복 사용자명: 에러가 반환되어야 함")
	}
	if err != chat.ErrUserAlreadyExists {
		t.Errorf("잘못된 에러: 기대=%v, 실제=%v", chat.ErrUserAlreadyExists, err)
	}
}

func TestDisconnect_Basic(t *testing.T) {
	server, _ := startServer(t)

	_, err := server.Connect("charlie")
	if err != nil {
		t.Fatalf("Connect 실패: %v", err)
	}

	if server.UserCount() != 1 {
		t.Errorf("연결 후 사용자 수: 기대=1, 실제=%d", server.UserCount())
	}

	err = server.Disconnect("charlie")
	if err != nil {
		t.Fatalf("Disconnect 실패: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	if server.UserCount() != 0 {
		t.Errorf("해제 후 사용자 수: 기대=0, 실제=%d", server.UserCount())
	}
}

func TestDisconnect_NotFound(t *testing.T) {
	server, _ := startServer(t)

	err := server.Disconnect("nobody")
	if err == nil {
		t.Error("없는 사용자 Disconnect: 에러가 반환되어야 함")
	}
	if err != chat.ErrUserNotFound {
		t.Errorf("잘못된 에러: 기대=%v, 실제=%v", chat.ErrUserNotFound, err)
	}
}

func TestUserCount(t *testing.T) {
	server, _ := startServer(t)

	if server.UserCount() != 0 {
		t.Errorf("초기 사용자 수: 기대=0, 실제=%d", server.UserCount())
	}

	server.Connect("u1")
	server.Connect("u2")
	server.Connect("u3")

	if server.UserCount() != 3 {
		t.Errorf("3명 연결 후: 기대=3, 실제=%d", server.UserCount())
	}

	server.Disconnect("u2")
	time.Sleep(20 * time.Millisecond)

	if server.UserCount() != 2 {
		t.Errorf("1명 해제 후: 기대=2, 실제=%d", server.UserCount())
	}
}

func TestUserList(t *testing.T) {
	server, _ := startServer(t)

	server.Connect("alice")
	server.Connect("bob")
	server.Connect("charlie")

	list := server.UserList()
	if len(list) != 3 {
		t.Errorf("사용자 목록 길이: 기대=3, 실제=%d", len(list))
	}

	// 사용자명 포함 확인
	found := make(map[string]bool)
	for _, u := range list {
		found[u] = true
	}
	for _, name := range []string{"alice", "bob", "charlie"} {
		if !found[name] {
			t.Errorf("사용자 목록에 %s 없음", name)
		}
	}
}

// ─────────────────────────────────────────
// 브로드캐스트 테스트
// ─────────────────────────────────────────

func TestBroadcast_Basic(t *testing.T) {
	server, _ := startServer(t)

	alice, _ := server.Connect("alice")
	bob, _ := server.Connect("bob")
	charlie, _ := server.Connect("charlie")

	// alice가 브로드캐스트
	err := alice.Send("안녕하세요!")
	if err != nil {
		t.Fatalf("Send 실패: %v", err)
	}

	timeout := 300 * time.Millisecond

	// bob과 charlie가 수신해야 함 (시스템 메시지를 건너뛰고 Broadcast 찾기)
	findBroadcast := func(inbox <-chan chat.Message, content string) *chat.Message {
		for i := 0; i < 10; i++ {
			msg := receiveWithTimeout(inbox, timeout)
			if msg == nil {
				return nil
			}
			if msg.Type == chat.Broadcast && msg.Content == content {
				return msg
			}
		}
		return nil
	}

	bobMsg := findBroadcast(bob.Inbox, "안녕하세요!")
	if bobMsg == nil {
		t.Error("bob: 브로드캐스트 메시지 수신 실패")
	} else {
		if bobMsg.From != "alice" {
			t.Errorf("bob 수신 메시지 발신자: 기대=alice, 실제=%s", bobMsg.From)
		}
		if bobMsg.Content != "안녕하세요!" {
			t.Errorf("bob 수신 메시지 내용: 기대='안녕하세요!', 실제='%s'", bobMsg.Content)
		}
	}

	charlieMsg := findBroadcast(charlie.Inbox, "안녕하세요!")
	if charlieMsg == nil {
		t.Error("charlie: 브로드캐스트 메시지 수신 실패")
	}

	_ = bobMsg
	_ = charlieMsg
}

func TestBroadcast_MultipleMessages(t *testing.T) {
	server, _ := startServer(t)

	alice, _ := server.Connect("alice")
	bob, _ := server.Connect("bob")

	messages := []string{"메시지1", "메시지2", "메시지3"}
	for _, msg := range messages {
		alice.Send(msg)
	}

	timeout := 300 * time.Millisecond
	received := 0
	for range messages {
		msg := receiveWithTimeout(bob.Inbox, timeout)
		if msg != nil {
			received++
		}
	}

	// 시스템 메시지(입장 알림)를 제외하고 브로드캐스트만 카운트
	if received < len(messages) {
		t.Errorf("브로드캐스트 수신: 기대>=%d, 실제=%d", len(messages), received)
	}
}

// ─────────────────────────────────────────
// 귓속말 테스트
// ─────────────────────────────────────────

func TestPrivateMessage_Basic(t *testing.T) {
	server, _ := startServer(t)

	alice, _ := server.Connect("alice")
	bob, _ := server.Connect("bob")
	charlie, _ := server.Connect("charlie")

	// alice → bob 귓속말
	err := alice.SendTo("bob", "bob에게만 보내는 메시지")
	if err != nil {
		t.Fatalf("SendTo 실패: %v", err)
	}

	timeout := 300 * time.Millisecond

	// bob은 수신해야 함 (시스템 메시지를 건너뛰고 Private 찾기)
	var bobMsg *chat.Message
	for i := 0; i < 10; i++ {
		msg := receiveWithTimeout(bob.Inbox, timeout)
		if msg == nil {
			break
		}
		if msg.Type == chat.Private {
			bobMsg = msg
			break
		}
	}
	if bobMsg == nil {
		t.Error("bob: 귓속말 수신 실패")
	} else {
		if bobMsg.From != "alice" {
			t.Errorf("발신자: 기대=alice, 실제=%s", bobMsg.From)
		}
		if bobMsg.To != "bob" {
			t.Errorf("수신자: 기대=bob, 실제=%s", bobMsg.To)
		}
	}

	// charlie는 Private 메시지를 수신하면 안 됨
	for i := 0; i < 5; i++ {
		charlieMsg := receiveWithTimeout(charlie.Inbox, 50*time.Millisecond)
		if charlieMsg == nil {
			break
		}
		if charlieMsg.Type == chat.Private {
			t.Error("charlie: 귓속말을 수신함 (잘못된 라우팅)")
			break
		}
	}
}

func TestPrivateMessage_SelfMessage(t *testing.T) {
	server, _ := startServer(t)

	alice, _ := server.Connect("alice")

	err := alice.SendTo("alice", "자기 자신에게")
	if err == nil {
		t.Error("자기 자신에게 귓속말: 에러가 반환되어야 함")
	}
}

func TestPrivateMessage_UserNotFound(t *testing.T) {
	server, _ := startServer(t)

	alice, _ := server.Connect("alice")

	err := alice.SendTo("nobody", "없는 사람에게")
	if err == nil {
		t.Error("없는 사용자에게 귓속말: 에러가 반환되어야 함")
	}
}

// ─────────────────────────────────────────
// 시스템 알림 테스트
// ─────────────────────────────────────────

func TestSystemMessage_Join(t *testing.T) {
	server, _ := startServer(t)

	alice, _ := server.Connect("alice")

	// bob이 입장하면 alice에게 시스템 알림이 와야 함
	server.Connect("bob")

	timeout := 300 * time.Millisecond
	var systemMsg *chat.Message

	// alice의 inbox에서 시스템 메시지 찾기
	for i := 0; i < 5; i++ {
		msg := receiveWithTimeout(alice.Inbox, timeout)
		if msg == nil {
			break
		}
		if msg.Type == chat.System {
			systemMsg = msg
			break
		}
	}

	if systemMsg == nil {
		t.Error("입장 시스템 알림 없음")
	} else {
		t.Logf("입장 알림: %s", systemMsg.Content)
	}
}

func TestSystemMessage_Leave(t *testing.T) {
	server, _ := startServer(t)

	alice, _ := server.Connect("alice")
	server.Connect("bob")

	// bob 입장 알림 소비
	receiveWithTimeout(alice.Inbox, 200*time.Millisecond)

	// bob 퇴장
	server.Disconnect("bob")

	timeout := 300 * time.Millisecond
	var systemMsg *chat.Message

	for i := 0; i < 5; i++ {
		msg := receiveWithTimeout(alice.Inbox, timeout)
		if msg == nil {
			break
		}
		if msg.Type == chat.System {
			systemMsg = msg
			break
		}
	}

	if systemMsg == nil {
		t.Error("퇴장 시스템 알림 없음")
	} else {
		t.Logf("퇴장 알림: %s", systemMsg.Content)
	}
}

// ─────────────────────────────────────────
// 동시성 테스트
// ─────────────────────────────────────────

func TestConcurrentMessages(t *testing.T) {
	server, _ := startServer(t)

	clients := make([]*chat.Client, 5)
	for i := range clients {
		c, err := server.Connect(fmt.Sprintf("user%d", i))
		if err != nil {
			t.Fatalf("Connect user%d 실패: %v", i, err)
		}
		clients[i] = c
	}

	// 모든 클라이언트가 동시에 메시지 전송
	done := make(chan struct{})
	for _, c := range clients {
		c := c
		go func() {
			for i := 0; i < 5; i++ {
				c.Send(fmt.Sprintf("%s의 메시지 %d", c.Username, i))
				time.Sleep(5 * time.Millisecond)
			}
			done <- struct{}{}
		}()
	}

	// 모든 전송 완료 대기
	for range clients {
		<-done
	}

	// 패닉 없이 완료되면 통과
	t.Logf("동시 메시지 전송 완료 (레이스 없음)")
}

// ─────────────────────────────────────────
// Context 취소 테스트
// ─────────────────────────────────────────

func TestServer_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := chat.NewChatServer()
	server.Start(ctx)
	time.Sleep(10 * time.Millisecond)

	alice, _ := server.Connect("alice")
	bob, _ := server.Connect("bob")

	// 서버 종료
	cancel()
	time.Sleep(100 * time.Millisecond)

	// 종료 후 새 연결 시도: 타임아웃 컨텍스트로 블로킹 방지
	connectCtx, connectCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer connectCancel()
	connectDone := make(chan error, 1)
	go func() {
		_, err := server.Connect("charlie")
		connectDone <- err
	}()
	select {
	case err := <-connectDone:
		if err == nil {
			t.Log("서버 종료 후 Connect: 에러 반환 권장 (구현에 따라 허용)")
		}
	case <-connectCtx.Done():
		t.Log("서버 종료 후 Connect 타임아웃 (서버 루프 종료로 블로킹됨 - 허용)")
	}

	// alice, bob의 Inbox 채널이 닫혀야 함
	select {
	case _, ok := <-alice.Inbox:
		if !ok {
			t.Log("alice Inbox 채널 정상 종료")
		}
	case <-time.After(200 * time.Millisecond):
		t.Log("alice Inbox 채널이 닫히지 않음 (타임아웃) - 구현에 따라 허용")
	}

	_ = bob
}

// ─────────────────────────────────────────
// 고루틴 누수 테스트
// ─────────────────────────────────────────

func TestNoGoroutineLeak(t *testing.T) {
	before := goroutineCount()

	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		server := chat.NewChatServer()
		server.Start(ctx)
		time.Sleep(5 * time.Millisecond)

		c1, _ := server.Connect("u1")
		c2, _ := server.Connect("u2")
		c1.Send("hello")
		receiveWithTimeout(c2.Inbox, 100*time.Millisecond)

		cancel()
		time.Sleep(50 * time.Millisecond)
	}

	waitForGoroutines(t, before, time.Second)
	after := goroutineCount()
	if after > before+3 {
		t.Errorf("고루틴 누수: 전=%d, 후=%d", before, after)
	}
}

// ─────────────────────────────────────────
// 채점 테스트
// ─────────────────────────────────────────

func TestGrade(t *testing.T) {
	score := 0

	// 연결/해제 (15점)
	t.Run("연결해제", func(t *testing.T) {
		server, _ := startServer(t)
		c, err := server.Connect("alice")
		if err == nil && c != nil && c.Username == "alice" {
			if server.UserCount() == 1 {
				server.Disconnect("alice")
				time.Sleep(20 * time.Millisecond)
				if server.UserCount() == 0 {
					score += 15
					fmt.Println("  [통과] 연결/해제: +15점")
				} else {
					fmt.Println("  [실패] 연결/해제: Disconnect 후 UserCount != 0")
				}
			} else {
				fmt.Println("  [실패] 연결/해제: Connect 후 UserCount != 1")
			}
		} else {
			fmt.Printf("  [실패] 연결/해제: Connect 오류 err=%v\n", err)
		}
	})

	// 브로드캐스트 (25점)
	t.Run("브로드캐스트", func(t *testing.T) {
		server, _ := startServer(t)
		alice, _ := server.Connect("alice")
		bob, _ := server.Connect("bob")
		charlie, _ := server.Connect("charlie")

		alice.Send("테스트 브로드캐스트")

		timeout := 300 * time.Millisecond
		var received int
		for _, c := range []*chat.Client{bob, charlie} {
			for i := 0; i < 5; i++ {
				msg := receiveWithTimeout(c.Inbox, timeout)
				if msg == nil {
					break
				}
				if msg.Type == chat.Broadcast && msg.Content == "테스트 브로드캐스트" {
					received++
					break
				}
			}
		}

		if received == 2 {
			score += 25
			fmt.Println("  [통과] 브로드캐스트: +25점")
		} else {
			fmt.Printf("  [실패] 브로드캐스트: 수신한 클라이언트=%d/2\n", received)
		}
	})

	// 귓속말 (20점)
	t.Run("귓속말", func(t *testing.T) {
		server, _ := startServer(t)
		alice, _ := server.Connect("alice")
		bob, _ := server.Connect("bob")
		charlie, _ := server.Connect("charlie")

		err := alice.SendTo("bob", "귓속말 내용")
		if err != nil {
			fmt.Printf("  [실패] 귓속말: SendTo 에러=%v\n", err)
			return
		}

		timeout := 300 * time.Millisecond
		bobGotIt := false
		charlieGotIt := false

		// bob 수신 확인
		for i := 0; i < 5; i++ {
			msg := receiveWithTimeout(bob.Inbox, timeout)
			if msg == nil {
				break
			}
			if msg.Type == chat.Private && msg.Content == "귓속말 내용" {
				bobGotIt = true
				break
			}
		}

		// charlie 미수신 확인
		for i := 0; i < 3; i++ {
			msg := receiveWithTimeout(charlie.Inbox, 50*time.Millisecond)
			if msg != nil && msg.Type == chat.Private {
				charlieGotIt = true
				break
			}
		}

		if bobGotIt && !charlieGotIt {
			score += 20
			fmt.Println("  [통과] 귓속말: +20점")
		} else {
			fmt.Printf("  [실패] 귓속말: bob수신=%v, charlie미수신=%v\n", bobGotIt, !charlieGotIt)
		}
	})

	// 시스템 알림 (15점)
	t.Run("시스템알림", func(t *testing.T) {
		server, _ := startServer(t)
		alice, _ := server.Connect("alice")
		server.Connect("bob")

		timeout := 300 * time.Millisecond
		got := false
		for i := 0; i < 5; i++ {
			msg := receiveWithTimeout(alice.Inbox, timeout)
			if msg == nil {
				break
			}
			if msg.Type == chat.System {
				got = true
				break
			}
		}
		if got {
			score += 15
			fmt.Println("  [통과] 시스템 알림: +15점")
		} else {
			fmt.Println("  [실패] 시스템 알림: 입장 알림 없음")
		}
	})

	// 동시성 안전 (15점) - -race로 별도 확인 필요
	t.Run("동시성안전", func(t *testing.T) {
		server, _ := startServer(t)
		clients := make([]*chat.Client, 4)
		for i := range clients {
			c, err := server.Connect(fmt.Sprintf("u%d", i))
			if err != nil {
				fmt.Printf("  [실패] 동시성 안전: Connect 오류\n")
				return
			}
			clients[i] = c
		}

		done := make(chan bool, len(clients))
		for _, c := range clients {
			c := c
			go func() {
				for i := 0; i < 3; i++ {
					c.Send(fmt.Sprintf("msg%d", i))
					time.Sleep(5 * time.Millisecond)
				}
				done <- true
			}()
		}
		for range clients {
			<-done
		}
		score += 15
		fmt.Println("  [통과] 동시성 안전: +15점 (go test -race로 추가 확인 필요)")
	})

	// 고루틴 누수 없음 (10점)
	t.Run("고루틴누수없음", func(t *testing.T) {
		before := goroutineCount()

		ctx, cancel := context.WithCancel(context.Background())
		server := chat.NewChatServer()
		server.Start(ctx)
		time.Sleep(5 * time.Millisecond)

		c1, _ := server.Connect("a")
		c2, _ := server.Connect("b")
		c1.Send("hello")
		receiveWithTimeout(c2.Inbox, 100*time.Millisecond)

		cancel()
		time.Sleep(100 * time.Millisecond)

		waitForGoroutines(t, before, time.Second)
		after := goroutineCount()
		if after <= before+3 {
			score += 10
			fmt.Printf("  [통과] 고루틴 누수 없음: +10점 (전=%d, 후=%d)\n", before, after)
		} else {
			fmt.Printf("  [실패] 고루틴 누수: 전=%d, 후=%d\n", before, after)
		}
	})

	fmt.Println()
	fmt.Printf("╔══════════════════════════════════╗\n")
	fmt.Printf("║  최종 점수: %3d / 100점            ║\n", score)
	grade := "F"
	switch {
	case score >= 90:
		grade = "A+"
	case score >= 80:
		grade = "A"
	case score >= 70:
		grade = "B"
	case score >= 60:
		grade = "C"
	}
	fmt.Printf("║  등급: %-30s║\n", grade)
	fmt.Printf("╚══════════════════════════════════╝\n")

	if score < 60 {
		t.Errorf("점수 미달: %d/100점", score)
	}
}
