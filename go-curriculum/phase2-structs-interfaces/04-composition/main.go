// 패키지 선언
package main

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────
// 1. 구조체 임베딩(Embedding) 기본
// ─────────────────────────────────────────
// Go에는 상속(inheritance)이 없습니다.
// 대신 임베딩(embedding)으로 코드를 재사용합니다.
// 임베딩하면 내부 타입의 메서드와 필드가 외부 타입으로 "프로모션"됩니다.

// Animal은 모든 동물의 공통 속성입니다.
type Animal struct {
	Name string
	Age  int
}

// Speak는 Animal의 기본 메서드입니다.
func (a Animal) Speak() string {
	return fmt.Sprintf("%s(이)가 소리를 냅니다.", a.Name)
}

// Describe는 동물을 설명합니다.
func (a Animal) Describe() string {
	return fmt.Sprintf("이름: %s, 나이: %d살", a.Name, a.Age)
}

// Dog는 Animal을 임베딩합니다.
// Dog는 Animal의 모든 필드와 메서드를 자동으로 가집니다 (프로모션).
type Dog struct {
	Animal      // 임베딩: 필드명 없이 타입명만 씀
	Breed string // Dog 고유 필드
}

// Speak는 Dog가 Animal의 Speak를 오버라이드합니다.
// 임베딩된 메서드를 같은 이름으로 재정의할 수 있습니다.
func (d Dog) Speak() string {
	return fmt.Sprintf("%s(이)가 '멍멍!'이라고 짖습니다.", d.Name)
}

// Fetch는 Dog 고유 메서드입니다.
func (d Dog) Fetch() string {
	return fmt.Sprintf("%s(이)가 공을 물어옵니다!", d.Name)
}

// Cat은 Animal을 임베딩합니다.
type Cat struct {
	Animal        // 임베딩
	Indoor bool   // Cat 고유 필드
}

// Speak는 Cat이 Animal의 Speak를 오버라이드합니다.
func (c Cat) Speak() string {
	return fmt.Sprintf("%s(이)가 '야옹~'이라고 웁니다.", c.Name)
}

// Purr는 Cat 고유 메서드입니다.
func (c Cat) Purr() string {
	return fmt.Sprintf("%s(이)가 그루룩 그루룩 소리를 냅니다.", c.Name)
}

// ─────────────────────────────────────────
// 2. 다중 임베딩 (Multiple Embedding)
// ─────────────────────────────────────────

// Logger는 로깅 기능을 제공합니다.
type Logger struct {
	prefix string
	logs   []string
}

// NewLogger는 Logger 생성자입니다.
func NewLogger(prefix string) Logger {
	return Logger{prefix: prefix}
}

// Log는 메시지를 기록합니다.
func (l *Logger) Log(msg string) {
	entry := fmt.Sprintf("[%s] %s", l.prefix, msg)
	l.logs = append(l.logs, entry)
}

// Logs는 모든 로그를 반환합니다.
func (l *Logger) Logs() []string {
	return l.logs
}

// Validator는 유효성 검사 기능을 제공합니다.
type Validator struct {
	errors []string
}

// AddError는 에러를 추가합니다.
func (v *Validator) AddError(msg string) {
	v.errors = append(v.errors, msg)
}

// IsValid는 에러가 없으면 true를 반환합니다.
func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

// Errors는 모든 에러를 반환합니다.
func (v *Validator) Errors() []string {
	return v.errors
}

// UserService는 Logger와 Validator를 모두 임베딩합니다.
type UserService struct {
	Logger    // 임베딩 1: 로깅 기능
	Validator // 임베딩 2: 유효성 검사 기능
	users map[string]string
}

// NewUserService는 UserService 생성자입니다.
func NewUserService() *UserService {
	return &UserService{
		Logger:  NewLogger("UserService"),
		users:   make(map[string]string),
	}
}

// CreateUser는 새 사용자를 생성합니다.
func (us *UserService) CreateUser(name, email string) bool {
	us.Log(fmt.Sprintf("사용자 생성 시도: %s <%s>", name, email))

	// 유효성 검사
	if strings.TrimSpace(name) == "" {
		us.AddError("이름은 비어있을 수 없습니다")
	}
	if !strings.Contains(email, "@") {
		us.AddError(fmt.Sprintf("유효하지 않은 이메일: %s", email))
	}

	if !us.IsValid() {
		us.Log("사용자 생성 실패: 유효성 검사 오류")
		return false
	}

	us.users[name] = email
	us.Log(fmt.Sprintf("사용자 생성 성공: %s", name))
	return true
}

// ─────────────────────────────────────────
// 3. 인터페이스 조합 (Interface Composition)
// ─────────────────────────────────────────
// 인터페이스도 다른 인터페이스를 임베딩하여 조합할 수 있습니다.
// io.ReadWriter = io.Reader + io.Writer 가 대표적인 예시입니다.

// Reader는 읽기 인터페이스입니다.
type Reader interface {
	Read() string
}

// Writer는 쓰기 인터페이스입니다.
type Writer interface {
	Write(s string)
}

// ReadWriter는 Reader와 Writer를 조합합니다.
// io.ReadWriter와 동일한 패턴입니다.
type ReadWriter interface {
	Reader
	Writer
}

// Closer는 닫기 인터페이스입니다.
type Closer interface {
	Close() error
}

// ReadWriteCloser는 세 인터페이스를 모두 조합합니다.
// io.ReadWriteCloser와 동일한 패턴입니다.
type ReadWriteCloser interface {
	ReadWriter
	Closer
}

// Buffer는 ReadWriteCloser를 구현합니다.
type Buffer struct {
	data   strings.Builder
	closed bool
}

// Read는 버퍼 내용을 읽습니다.
func (b *Buffer) Read() string {
	return b.data.String()
}

// Write는 버퍼에 데이터를 씁니다.
func (b *Buffer) Write(s string) {
	if !b.closed {
		b.data.WriteString(s)
	}
}

// Close는 버퍼를 닫습니다.
func (b *Buffer) Close() error {
	if b.closed {
		return fmt.Errorf("이미 닫힌 버퍼입니다")
	}
	b.closed = true
	return nil
}

// ─────────────────────────────────────────
// 4. 상속 대신 컴포지션 패턴
// ─────────────────────────────────────────
// "상속보다 컴포지션을 선호하라" - 고전적인 소프트웨어 설계 원칙
// Go는 언어 수준에서 이를 강제합니다.

// Notifier는 알림 기능 인터페이스입니다.
type Notifier interface {
	Notify(message string) error
}

// EmailNotifier는 이메일 알림을 보냅니다.
type EmailNotifier struct {
	From string
}

// Notify는 이메일 알림을 전송합니다.
func (e EmailNotifier) Notify(message string) error {
	fmt.Printf("  [이메일] %s → %s\n", e.From, message)
	return nil
}

// SMSNotifier는 SMS 알림을 보냅니다.
type SMSNotifier struct {
	PhoneNumber string
}

// Notify는 SMS 알림을 전송합니다.
func (s SMSNotifier) Notify(message string) error {
	fmt.Printf("  [SMS] %s: %s\n", s.PhoneNumber, message)
	return nil
}

// MultiNotifier는 여러 Notifier를 조합합니다.
// 이것이 컴포지션 패턴입니다: 상속 없이 기능을 조합합니다.
type MultiNotifier struct {
	notifiers []Notifier // 인터페이스 슬라이스
}

// AddNotifier는 알림 수단을 추가합니다.
func (m *MultiNotifier) AddNotifier(n Notifier) {
	m.notifiers = append(m.notifiers, n)
}

// Notify는 모든 알림 수단으로 메시지를 전송합니다.
func (m *MultiNotifier) Notify(message string) error {
	var errs []string
	for _, n := range m.notifiers {
		if err := n.Notify(message); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("알림 전송 실패: %s", strings.Join(errs, "; "))
	}
	return nil
}

func main() {
	fmt.Println("=== Go Phase 2: 컴포지션(Composition) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 임베딩과 메서드 프로모션
	// ─────────────────────────────────────────
	fmt.Println("--- 1. 구조체 임베딩과 메서드 프로모션 ---")

	dog := Dog{
		Animal: Animal{Name: "바둑이", Age: 3},
		Breed:  "진돗개",
	}

	// 프로모션된 필드 직접 접근 (dog.Animal.Name 또는 dog.Name)
	fmt.Printf("이름: %s (dog.Name으로 접근)\n", dog.Name)
	fmt.Printf("나이: %d (dog.Age로 접근)\n", dog.Age)
	fmt.Printf("품종: %s\n", dog.Breed)

	// 프로모션된 메서드 호출
	fmt.Println(dog.Describe())  // Animal.Describe() 프로모션됨
	fmt.Println(dog.Speak())     // Dog가 오버라이드한 Speak()
	fmt.Println(dog.Fetch())     // Dog 고유 메서드

	// 임베딩된 원본에 직접 접근도 가능
	fmt.Println(dog.Animal.Speak()) // Animal의 원본 Speak() 호출

	fmt.Println()

	cat := Cat{
		Animal: Animal{Name: "나비", Age: 5},
		Indoor: true,
	}
	fmt.Println(cat.Describe())
	fmt.Println(cat.Speak())
	fmt.Println(cat.Purr())
	fmt.Printf("실내 고양이: %v\n", cat.Indoor)
	fmt.Println()

	// ─────────────────────────────────────────
	// 다중 임베딩
	// ─────────────────────────────────────────
	fmt.Println("--- 2. 다중 임베딩 (UserService) ---")

	svc := NewUserService()

	// 정상 사용자 생성
	svc.CreateUser("홍길동", "hong@example.com")
	svc.CreateUser("김영희", "kim@example.com")

	// 잘못된 사용자 생성 (유효성 검사 실패)
	svc.CreateUser("", "invalid-email")

	// 로그 출력 (임베딩된 Logger 메서드 직접 호출)
	fmt.Println("\n로그 기록:")
	for _, log := range svc.Logs() {
		fmt.Printf("  %s\n", log)
	}

	// 에러 출력 (임베딩된 Validator 메서드 직접 호출)
	fmt.Println("\n유효성 검사 에러:")
	for _, e := range svc.Errors() {
		fmt.Printf("  - %s\n", e)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 인터페이스 조합
	// ─────────────────────────────────────────
	fmt.Println("--- 3. 인터페이스 조합 (ReadWriteCloser) ---")

	buf := &Buffer{}

	// ReadWriteCloser 인터페이스로 사용
	var rwc ReadWriteCloser = buf

	rwc.Write("Hello, ")
	rwc.Write("Go 컴포지션!")
	fmt.Printf("버퍼 내용: %q\n", rwc.Read())

	// Close 후 Write 시도
	if err := rwc.Close(); err != nil {
		fmt.Printf("Close 에러: %v\n", err)
	} else {
		fmt.Println("버퍼 닫힘")
	}

	rwc.Write("닫힌 후 쓰기") // 무시됨
	fmt.Printf("Close 후 버퍼 내용: %q (변경 없음)\n", rwc.Read())

	// 이미 닫힌 버퍼 다시 닫기
	if err := rwc.Close(); err != nil {
		fmt.Printf("두 번째 Close 에러: %v\n", err)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 컴포지션 패턴 (MultiNotifier)
	// ─────────────────────────────────────────
	fmt.Println("--- 4. 상속 대신 컴포지션 패턴 ---")

	// 각각의 알림 수단
	emailNotifier := EmailNotifier{From: "system@example.com"}
	smsNotifier := SMSNotifier{PhoneNumber: "010-1234-5678"}

	// 개별 사용
	fmt.Println("개별 알림:")
	emailNotifier.Notify("개별 이메일 테스트")
	smsNotifier.Notify("개별 SMS 테스트")

	// 조합하여 사용 (MultiNotifier)
	fmt.Println("\n다중 알림:")
	multi := &MultiNotifier{}
	multi.AddNotifier(emailNotifier)
	multi.AddNotifier(smsNotifier)
	multi.Notify("주문이 완료되었습니다!")

	// Notifier 인터페이스로 다형성 활용
	fmt.Println("\n인터페이스 다형성:")
	notifiers := []Notifier{
		EmailNotifier{From: "noreply@shop.com"},
		SMSNotifier{PhoneNumber: "010-9876-5432"},
		multi, // MultiNotifier도 Notifier를 구현
	}
	for _, n := range notifiers {
		n.Notify(fmt.Sprintf("결제 완료 (%T)", n))
	}

	fmt.Println()
	fmt.Println("=== 컴포지션 학습 완료 ===")
}
