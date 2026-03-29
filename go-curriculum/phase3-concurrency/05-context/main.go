// 패키지 선언
package main

// Go 동시성: context 패키지
//
// context.Context는 요청 범위의 데이터, 취소 신호, 데드라인을
// API 경계와 고루틴 사이에 전달하는 표준 방법입니다.
//
// 핵심 원칙:
// 1. Context는 함수의 첫 번째 인자로 전달 (관례: ctx context.Context)
// 2. nil Context를 전달하지 마세요 (context.Background() 또는 context.TODO() 사용)
// 3. Context를 구조체에 저장하지 마세요 (함수 인자로만 사용)
// 4. 부모 Context 취소 시 자식도 자동 취소됨

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// ─────────────────────────────────────────
// 1. context.Background()와 context.TODO()
// ─────────────────────────────────────────

func backgroundAndTodo() {
	fmt.Println("\n--- 1. context.Background()와 context.TODO() ---")

	// context.Background(): 최상위 루트 Context
	// - 취소, 값, 데드라인 없음
	// - main 함수, 테스트, 최상위 요청 핸들러에서 사용
	bg := context.Background()
	fmt.Printf("  Background: %v\n", bg)
	fmt.Printf("  Background Err: %v\n", bg.Err()) // nil: 취소되지 않음

	// context.TODO(): 아직 어떤 Context를 써야 할지 불확실할 때
	// - Background와 기술적으로 동일하지만 의미가 다름
	// - "나중에 채워야 함"을 나타내는 플레이스홀더
	// - 정적 분석 도구가 감지 가능
	todo := context.TODO()
	fmt.Printf("  TODO: %v\n", todo)

	fmt.Println("  Background: 최상위 루트 컨텍스트 (프로덕션 코드)")
	fmt.Println("  TODO: 컨텍스트 전파 경로가 불확실할 때 임시 사용")
}

// ─────────────────────────────────────────
// 2. context.WithCancel: 수동 취소
// ─────────────────────────────────────────

// longRunningTask는 취소 가능한 장기 실행 작업입니다.
func longRunningTask(ctx context.Context, id int) error {
	for step := 1; step <= 10; step++ {
		// 각 단계마다 취소 여부 확인
		select {
		case <-ctx.Done():
			// ctx.Err(): 취소 이유 반환
			// - context.Canceled: 명시적 취소
			// - context.DeadlineExceeded: 데드라인/타임아웃 초과
			fmt.Printf("  작업 #%d: 단계 %d에서 취소됨 (이유: %v)\n",
				id, step, ctx.Err())
			return ctx.Err()
		default:
			// 작업 수행
			fmt.Printf("  작업 #%d: 단계 %d 처리 중...\n", id, step)
			time.Sleep(30 * time.Millisecond)
		}
	}
	fmt.Printf("  작업 #%d: 모든 단계 완료!\n", id)
	return nil
}

func withCancelDemo() {
	fmt.Println("\n--- 2. context.WithCancel ---")

	// WithCancel: 부모 Context로부터 취소 가능한 자식 Context 생성
	// cancel 함수를 반드시 호출해야 리소스 누수 방지
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 함수 종료 시 항상 cancel 호출 (여러 번 호출해도 안전)

	var wg sync.WaitGroup

	// 여러 작업 동시 시작
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := longRunningTask(ctx, id); err != nil {
				fmt.Printf("  작업 #%d 에러: %v\n", id, err)
			}
		}(i)
	}

	// 100ms 후 모든 작업 취소
	time.Sleep(100 * time.Millisecond)
	fmt.Println("  ---취소 신호 전송---")
	cancel() // 이 ctx를 사용하는 모든 고루틴에게 취소 신호

	wg.Wait()
	fmt.Printf("  ctx.Err() 확인: %v\n", ctx.Err())
}

// ─────────────────────────────────────────
// 3. context.WithTimeout: 타임아웃
// ─────────────────────────────────────────

// simulateDBQuery는 데이터베이스 쿼리를 시뮬레이션합니다.
func simulateDBQuery(ctx context.Context, query string) (string, error) {
	// 쿼리 소요 시간 시뮬레이션 (0~300ms)
	delay := time.Duration(rand.Intn(300)) * time.Millisecond

	resultCh := make(chan string, 1)
	go func() {
		time.Sleep(delay)
		resultCh <- fmt.Sprintf("쿼리 결과 [%s]", query)
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case <-ctx.Done():
		return "", fmt.Errorf("DB 쿼리 실패: %w", ctx.Err())
	}
}

func withTimeoutDemo() {
	fmt.Println("\n--- 3. context.WithTimeout ---")

	// WithTimeout: 지정된 시간 후 자동 취소되는 Context
	// deadline = now + timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel() // 타임아웃 전에 완료되면 cancel로 리소스 즉시 해제

	fmt.Printf("  타임아웃: 200ms\n")

	start := time.Now()
	result, err := simulateDBQuery(ctx, "SELECT * FROM users")
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("  실패 (%.0fms): %v\n", float64(elapsed.Milliseconds()), err)
	} else {
		fmt.Printf("  성공 (%.0fms): %s\n", float64(elapsed.Milliseconds()), result)
	}

	// 마감 시간 확인
	if deadline, ok := ctx.Deadline(); ok {
		fmt.Printf("  마감 시간: %v\n", deadline.Format("15:04:05.000"))
		fmt.Printf("  남은 시간: %v\n", time.Until(deadline))
	}
}

// ─────────────────────────────────────────
// 4. context.WithDeadline: 절대 시각 기반 데드라인
// ─────────────────────────────────────────

func withDeadlineDemo() {
	fmt.Println("\n--- 4. context.WithDeadline ---")

	// WithDeadline: 특정 시각에 취소되는 Context
	// WithTimeout(ctx, d) == WithDeadline(ctx, time.Now().Add(d))
	deadline := time.Now().Add(150 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	fmt.Printf("  데드라인: %v\n", deadline.Format("15:04:05.000"))

	// 여러 작업 수행
	tasks := []string{"태스크A", "태스크B", "태스크C"}
	for _, task := range tasks {
		select {
		case <-ctx.Done():
			fmt.Printf("  %s: 데드라인 초과로 건너뜀\n", task)
		default:
			fmt.Printf("  %s: 실행 중...\n", task)
			time.Sleep(60 * time.Millisecond) // 각 태스크 60ms
		}
	}

	fmt.Printf("  최종 상태: %v\n", ctx.Err())
}

// ─────────────────────────────────────────
// 5. context.WithValue: 값 전달
// ─────────────────────────────────────────

// 키 타입: 기본 타입(string, int 등)을 직접 키로 쓰면 충돌 위험
// → unexported 커스텀 타입 사용 (다른 패키지와 키 충돌 방지)
type contextKey string

const (
	requestIDKey contextKey = "requestID"
	userIDKey    contextKey = "userID"
	traceKey     contextKey = "traceID"
)

// withRequestID는 Context에 요청 ID를 추가합니다.
func withRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}

// getRequestID는 Context에서 요청 ID를 꺼냅니다.
func getRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

// middleware는 미들웨어처럼 Context에 값을 추가합니다.
func middleware(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, requestIDKey, "req-abc-123")
	ctx = context.WithValue(ctx, userIDKey, "user-42")
	ctx = context.WithValue(ctx, traceKey, "trace-xyz-789")
	return ctx
}

// handler는 Context에서 값을 꺼내 사용합니다.
func handler(ctx context.Context) {
	reqID, _ := ctx.Value(requestIDKey).(string)
	userID, _ := ctx.Value(userIDKey).(string)
	traceID, _ := ctx.Value(traceKey).(string)

	fmt.Printf("  핸들러: reqID=%s, userID=%s, traceID=%s\n",
		reqID, userID, traceID)
}

func withValueDemo() {
	fmt.Println("\n--- 5. context.WithValue ---")
	fmt.Println("  주의사항:")
	fmt.Println("  - 함수 인자로 전달해야 할 값은 Context에 넣지 마세요")
	fmt.Println("  - 요청 범위 메타데이터만 사용 (요청ID, 인증토큰, 추적ID 등)")
	fmt.Println("  - Context 값은 타입 안전하지 않음 → 래퍼 함수 사용 권장")

	ctx := context.Background()
	ctx = middleware(ctx)
	handler(ctx)

	// Context 체인: 값은 부모에서 자식으로 전파됨
	childCtx := context.WithValue(ctx, contextKey("extra"), "추가값")
	fmt.Printf("  자식 ctx 요청ID: %v\n", childCtx.Value(requestIDKey)) // 부모 값 접근 가능
	fmt.Printf("  자식 ctx 추가값: %v\n", childCtx.Value(contextKey("extra")))
	fmt.Printf("  부모 ctx 추가값: %v\n", ctx.Value(contextKey("extra"))) // nil: 자식 값은 부모에 없음
}

// ─────────────────────────────────────────
// 6. HTTP 핸들러에서 context 사용
// ─────────────────────────────────────────

// slowDBQuery는 느린 DB 쿼리를 시뮬레이션합니다.
func slowDBQuery(ctx context.Context) (string, error) {
	// 실제 DB 드라이버들은 ctx를 받아 쿼리 취소를 지원합니다.
	select {
	case <-time.After(500 * time.Millisecond): // 500ms 소요
		return "데이터베이스 결과", nil
	case <-ctx.Done():
		return "", fmt.Errorf("쿼리 취소: %w", ctx.Err())
	}
}

// apiHandler는 HTTP 요청을 처리하는 핸들러입니다.
func apiHandler(w http.ResponseWriter, r *http.Request) {
	// r.Context(): HTTP 요청의 Context (클라이언트 연결 끊기면 자동 취소)
	ctx := r.Context()

	// 요청별 타임아웃 설정
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	// 요청 ID 추가 (미들웨어에서 설정했다고 가정)
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = "auto-generated-id"
	}
	ctx = withRequestID(ctx, reqID)

	// DB 쿼리 수행 (Context 전달로 취소 가능)
	result, err := slowDBQuery(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("에러: %v", err), http.StatusGatewayTimeout)
		return
	}

	fmt.Fprintf(w, "결과: %s\n", result)
}

func httpContextDemo() {
	fmt.Println("\n--- 6. HTTP 핸들러에서 context 사용 ---")

	// httptest로 실제 서버 없이 테스트
	handler := http.HandlerFunc(apiHandler)

	// 케이스 1: 타임아웃 이내 (빠른 쿼리 - 300ms 타임아웃, 쿼리 500ms → 실패)
	fmt.Println("  케이스 1: 타임아웃 초과 (쿼리 500ms > 타임아웃 300ms):")
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-Request-ID", "req-001")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	fmt.Printf("    상태 코드: %d\n", w.Code)
	fmt.Printf("    응답 본문: %s", w.Body.String())

	// Context 취소 전파 데모
	fmt.Println("\n  Context 취소 전파:")
	ctx, cancel := context.WithCancel(context.Background())

	// 자식 Context들
	child1, cancel1 := context.WithCancel(ctx)
	child2, cancel2 := context.WithTimeout(ctx, 1*time.Second)
	defer cancel1()
	defer cancel2()

	var wg sync.WaitGroup
	for i, c := range []context.Context{ctx, child1, child2} {
		wg.Add(1)
		go func(id int, ctx context.Context) {
			defer wg.Done()
			<-ctx.Done()
			fmt.Printf("    Context #%d 취소됨: %v\n", id, ctx.Err())
		}(i, c)
	}

	time.Sleep(50 * time.Millisecond)
	fmt.Println("  부모 Context 취소 → 자식 모두 자동 취소:")
	cancel() // 부모 취소 → child1, child2 모두 자동 취소
	wg.Wait()
}

// ─────────────────────────────────────────
// 7. Context 체인 시각화
// ─────────────────────────────────────────

func contextChainVisualization() {
	fmt.Println("\n--- 7. Context 트리 구조 ---")

	// Background (루트)
	//   └── WithValue (requestID)
	//         ├── WithTimeout (300ms)
	//         │     └── WithValue (userID)
	//         └── WithCancel
	//               └── WithValue (traceID)

	root := context.Background()
	withReq := context.WithValue(root, requestIDKey, "req-999")

	timeoutCtx, cancelTimeout := context.WithTimeout(withReq, 300*time.Millisecond)
	defer cancelTimeout()
	withUser := context.WithValue(timeoutCtx, userIDKey, "user-777")

	cancelCtx, cancelManual := context.WithCancel(withReq)
	defer cancelManual()
	withTrace := context.WithValue(cancelCtx, traceKey, "trace-abc")

	fmt.Printf("  withUser에서 requestID: %v\n", withUser.Value(requestIDKey)) // 부모에서 상속
	fmt.Printf("  withUser에서 userID: %v\n", withUser.Value(userIDKey))
	fmt.Printf("  withTrace에서 requestID: %v\n", withTrace.Value(requestIDKey)) // 부모에서 상속
	fmt.Printf("  withTrace에서 userID: %v (없음)\n", withTrace.Value(userIDKey))
	fmt.Printf("  withTrace에서 traceID: %v\n", withTrace.Value(traceKey))

	// 데드라인 확인
	if dl, ok := withUser.Deadline(); ok {
		fmt.Printf("  withUser 데드라인: %v\n", dl.Format("15:04:05.000"))
	}
	if _, ok := withTrace.Deadline(); !ok {
		fmt.Println("  withTrace 데드라인: 없음 (WithCancel은 데드라인 없음)")
	}
}

// ─────────────────────────────────────────
// main 함수
// ─────────────────────────────────────────

func main() {
	fmt.Println("=== Go 동시성: context 패키지 ===")

	backgroundAndTodo()
	withCancelDemo()
	withTimeoutDemo()
	withDeadlineDemo()
	withValueDemo()
	httpContextDemo()
	contextChainVisualization()

	fmt.Println("\n=== 프로그램 정상 종료 ===")
}
