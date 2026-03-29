// 패키지 선언
package main

import "fmt"

// ─────────────────────────────────────────
// 구조체 예시 (포인터 수신자 설명용)
// ─────────────────────────────────────────

// Counter: 값 수신자 vs 포인터 수신자 비교용 구조체
type Counter struct {
	value int
	name  string
}

// 값 수신자: 복사본에서 동작하므로 원본이 변경되지 않습니다.
func (c Counter) IncrementByValue() {
	c.value++ // 복사본만 변경
}

// 포인터 수신자: 원본을 직접 변경합니다.
func (c *Counter) IncrementByPointer() {
	c.value++ // 원본 변경
}

func (c Counter) String() string {
	return fmt.Sprintf("Counter{name=%q, value=%d}", c.name, c.value)
}

func main() {
	fmt.Println("=== Go 기초: 포인터 (Pointers) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 1. & 와 * 연산자
	// ─────────────────────────────────────────
	fmt.Println("--- 1. & 와 * 연산자 ---")

	x := 42
	// & (주소 연산자): 변수의 메모리 주소를 반환합니다.
	ptr := &x
	fmt.Printf("x의 값: %d\n", x)
	fmt.Printf("x의 주소 (&x): %p\n", ptr)
	fmt.Printf("ptr이 가리키는 값 (*ptr): %d\n", *ptr)
	fmt.Printf("ptr의 타입: %T\n", ptr)

	// * (역참조 연산자): 포인터가 가리키는 값에 접근합니다.
	*ptr = 100 // 포인터를 통해 원본 변수 수정
	fmt.Printf("*ptr=100 후, x의 값: %d\n", x) // x도 100으로 변경됨

	// 포인터도 변수: 다른 주소를 가리킬 수 있습니다.
	y := 200
	ptr = &y // 이제 ptr은 y를 가리킵니다.
	fmt.Printf("ptr = &y 후, *ptr = %d\n", *ptr)
	fmt.Println()

	// ─────────────────────────────────────────
	// 2. 값 전달 vs 포인터 전달
	// ─────────────────────────────────────────
	fmt.Println("--- 2. 값 전달 vs 포인터 전달 ---")

	// 값 전달: 함수가 복사본을 받으므로 원본이 변경되지 않습니다.
	num := 10
	fmt.Printf("increment 호출 전: num=%d\n", num)
	incrementByValue(num)
	fmt.Printf("incrementByValue 후: num=%d (변경 없음)\n", num)

	// 포인터 전달: 함수가 원본의 주소를 받으므로 원본이 변경됩니다.
	incrementByPointer(&num)
	fmt.Printf("incrementByPointer 후: num=%d (변경됨)\n", num)

	// 구조체 값 전달
	person := Person{name: "Alice", age: 30}
	fmt.Printf("\n원본 Person: %v\n", person)
	birthDayByValue(person)
	fmt.Printf("birthDayByValue 후: %v (변경 없음)\n", person)

	// 구조체 포인터 전달
	birthDayByPointer(&person)
	fmt.Printf("birthDayByPointer 후: %v (변경됨)\n", person)
	fmt.Println()

	// ─────────────────────────────────────────
	// 3. 포인터의 포인터
	// ─────────────────────────────────────────
	fmt.Println("--- 3. 포인터의 포인터 ---")

	a := 42
	p1 := &a    // *int: a의 주소
	p2 := &p1   // **int: p1의 주소

	fmt.Printf("a = %d\n", a)
	fmt.Printf("p1 = %p (a의 주소)\n", p1)
	fmt.Printf("*p1 = %d\n", *p1)
	fmt.Printf("p2 = %p (p1의 주소)\n", p2)
	fmt.Printf("*p2 = %p (p1의 값 = a의 주소)\n", *p2)
	fmt.Printf("**p2 = %d (a의 값)\n", **p2)

	**p2 = 999
	fmt.Printf("**p2=999 후, a = %d\n", a)
	fmt.Println()

	// ─────────────────────────────────────────
	// 4. nil 포인터
	// ─────────────────────────────────────────
	fmt.Println("--- 4. nil 포인터 ---")

	var nilPtr *int // 초기화되지 않은 포인터: nil
	fmt.Printf("nil 포인터: %v\n", nilPtr)
	fmt.Printf("nil 포인터 == nil: %t\n", nilPtr == nil)

	// nil 포인터 역참조는 패닉!
	// fmt.Println(*nilPtr) // 런타임 패닉: nil pointer dereference

	// 안전한 포인터 사용: nil 확인
	if nilPtr != nil {
		fmt.Printf("값: %d\n", *nilPtr)
	} else {
		fmt.Println("포인터가 nil입니다. 역참조 불가.")
	}

	// 포인터를 안전하게 사용하는 함수
	safeValue := safeDeref(nilPtr, -1)
	fmt.Printf("safeDeref(nil, -1) = %d\n", safeValue)
	realPtr := &num
	safeValue2 := safeDeref(realPtr, -1)
	fmt.Printf("safeDeref(&num, -1) = %d\n", safeValue2)
	fmt.Println()

	// ─────────────────────────────────────────
	// 5. new() 함수
	// ─────────────────────────────────────────
	fmt.Println("--- 5. new() 함수 ---")

	// new(T): 타입 T의 제로값을 할당하고 *T를 반환합니다.
	// &T{} 리터럴과 유사하지만 기본 타입에 편리합니다.
	pInt := new(int)         // *int, 값=0
	pFloat := new(float64)   // *float64, 값=0.0
	pBool := new(bool)       // *bool, 값=false
	pString := new(string)   // *string, 값=""

	fmt.Printf("new(int):     %p -> %d\n", pInt, *pInt)
	fmt.Printf("new(float64): %p -> %f\n", pFloat, *pFloat)
	fmt.Printf("new(bool):    %p -> %t\n", pBool, *pBool)
	fmt.Printf("new(string):  %p -> %q\n", pString, *pString)

	// 값 설정
	*pInt = 42
	fmt.Printf("*pInt = 42 후: %d\n", *pInt)

	// new()는 구조체에도 사용 가능
	pPerson := new(Person)
	pPerson.name = "Bob"
	pPerson.age = 25
	fmt.Printf("new(Person): %v\n", *pPerson)
	fmt.Println()

	// ─────────────────────────────────────────
	// 6. 포인터와 슬라이스/맵
	// ─────────────────────────────────────────
	fmt.Println("--- 6. 슬라이스/맵은 이미 참조 타입 ---")

	// 슬라이스와 맵은 이미 내부적으로 참조를 포함합니다.
	// 함수에 전달할 때 포인터 없이도 수정이 반영됩니다.
	slice := []int{1, 2, 3}
	fmt.Printf("modifySlice 전: %v\n", slice)
	modifySlice(slice)
	fmt.Printf("modifySlice 후: %v (변경됨)\n", slice)

	// 하지만 append는 슬라이스 헤더(포인터, 길이, 용량)를 변경하므로
	// 새 슬라이스를 반환하거나 *[]int로 전달해야 합니다.
	fmt.Printf("appendToSlice 전: %v\n", slice)
	appendToSlice(&slice, 4, 5)
	fmt.Printf("appendToSlice 후: %v (포인터로 전달)\n", slice)

	// 맵도 참조 타입
	m := map[string]int{"a": 1}
	fmt.Printf("modifyMap 전: %v\n", m)
	modifyMap(m)
	fmt.Printf("modifyMap 후: %v (변경됨)\n", m)
	fmt.Println()

	// ─────────────────────────────────────────
	// 7. 포인터 수신자 (Pointer Receivers)
	// ─────────────────────────────────────────
	fmt.Println("--- 7. 포인터 수신자 ---")

	c1 := Counter{name: "counter1", value: 0}
	fmt.Printf("초기: %s\n", c1)

	c1.IncrementByValue()
	fmt.Printf("IncrementByValue 후: %s (변경 없음)\n", c1)

	c1.IncrementByPointer()
	fmt.Printf("IncrementByPointer 후: %s (변경됨)\n", c1)

	// 포인터에서 값 수신자 메서드 호출 가능 (자동 역참조)
	c2ptr := &Counter{name: "counter2", value: 10}
	c2ptr.IncrementByPointer() // (*c2ptr).IncrementByPointer()와 동일
	fmt.Printf("포인터에서 호출: %s\n", *c2ptr)
	fmt.Println()

	// ─────────────────────────────────────────
	// 8. 실용 패턴: 선택적 값 (Optional values)
	// ─────────────────────────────────────────
	fmt.Println("--- 8. 포인터로 선택적 값 표현 ---")

	// nil 포인터를 "값 없음"으로 사용 (Go에 Optional 타입이 없으므로)
	user1 := User{name: "Alice", age: 30, nickname: strPtr("앨리스")}
	user2 := User{name: "Bob", age: 25, nickname: nil} // 닉네임 없음

	printUser(user1)
	printUser(user2)
	fmt.Println()

	fmt.Println("=== 완료 ===")
}

// ─────────────────────────────────────────
// 헬퍼 함수들
// ─────────────────────────────────────────

// 값으로 전달: 복사본 수정, 원본 불변
func incrementByValue(n int) {
	n++
	fmt.Printf("  incrementByValue 내부: n=%d\n", n)
}

// 포인터로 전달: 원본 수정
func incrementByPointer(n *int) {
	*n++
	fmt.Printf("  incrementByPointer 내부: *n=%d\n", *n)
}

// Person 구조체
type Person struct {
	name string
	age  int
}

func (p Person) String() string {
	return fmt.Sprintf("Person{%s, %d세}", p.name, p.age)
}

func birthDayByValue(p Person) {
	p.age++
}

func birthDayByPointer(p *Person) {
	p.age++
}

// nil 포인터 안전 역참조
func safeDeref(p *int, defaultVal int) int {
	if p == nil {
		return defaultVal
	}
	return *p
}

// 슬라이스 요소 수정 (가능: 내부 배열 공유)
func modifySlice(s []int) {
	for i := range s {
		s[i] *= 2
	}
}

// 슬라이스 append (포인터 필요: 헤더 변경)
func appendToSlice(s *[]int, vals ...int) {
	*s = append(*s, vals...)
}

// 맵 수정 (가능: 참조 타입)
func modifyMap(m map[string]int) {
	m["b"] = 2
	m["c"] = 3
}

// User 구조체 (선택적 필드 예시)
type User struct {
	name     string
	age      int
	nickname *string // nil이면 닉네임 없음
}

func strPtr(s string) *string {
	return &s
}

func printUser(u User) {
	if u.nickname != nil {
		fmt.Printf("사용자: %s (%d세), 닉네임: %s\n", u.name, u.age, *u.nickname)
	} else {
		fmt.Printf("사용자: %s (%d세), 닉네임: (없음)\n", u.name, u.age)
	}
}
