// 패키지 선언
package main

import (
	"fmt"
	"math"
	"unsafe"
)

func main() {
	fmt.Println("=== Go 기초: 변수, 상수, 타입 ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 1. var 선언 (명시적 타입)
	// ─────────────────────────────────────────
	fmt.Println("--- 1. var 선언 ---")

	// 기본 var 선언: var 이름 타입
	var age int
	var name string
	var price float64
	var isActive bool

	// 선언 직후 값을 할당하지 않으면 "제로값(zero value)"이 됩니다.
	fmt.Printf("int 제로값:     %d\n", age)
	fmt.Printf("string 제로값:  %q\n", name)   // %q는 따옴표 포함 출력
	fmt.Printf("float64 제로값: %f\n", price)
	fmt.Printf("bool 제로값:    %t\n", isActive)

	// 선언과 동시에 초기화
	var city string = "서울"
	var population int = 9_700_000 // 숫자에 _ 사용 가능 (가독성용, Go 1.13+)
	fmt.Printf("\n도시: %s, 인구: %d\n", city, population)

	// 타입 추론: 초기값이 있으면 타입 생략 가능
	var language = "Go" // string으로 자동 추론
	var version = 1.21  // float64로 자동 추론
	fmt.Printf("언어: %s (타입: %T), 버전: %.2f (타입: %T)\n", language, language, version, version)
	fmt.Println()

	// ─────────────────────────────────────────
	// 2. := 단축 선언 (Short Variable Declaration)
	// ─────────────────────────────────────────
	fmt.Println("--- 2. := 단축 선언 ---")

	// := 는 함수 내부에서만 사용 가능합니다.
	// var + 타입 추론을 한 번에 처리합니다.
	score := 95          // int
	temperature := 36.5  // float64
	greeting := "안녕!"  // string
	running := true      // bool

	fmt.Printf("점수: %d, 온도: %.1f, 인사: %s, 실행중: %t\n",
		score, temperature, greeting, running)

	// 여러 변수를 한 번에 선언
	x, y := 10, 20
	a, b, c := "첫째", 2, true
	fmt.Printf("x=%d, y=%d\n", x, y)
	fmt.Printf("a=%s, b=%d, c=%t\n", a, b, c)

	// 값 교환 (swap) - Go의 우아한 특성
	fmt.Printf("교환 전: x=%d, y=%d\n", x, y)
	x, y = y, x
	fmt.Printf("교환 후: x=%d, y=%d\n", x, y)
	fmt.Println()

	// ─────────────────────────────────────────
	// 3. 블록 선언
	// ─────────────────────────────────────────
	fmt.Println("--- 3. 블록(괄호) 선언 ---")

	// var () 블록으로 여러 변수를 묶어서 선언
	var (
		firstName = "길동"
		lastName  = "홍"
		birthYear = 1990
		height    = 175.5
	)
	fmt.Printf("이름: %s %s, 출생: %d년, 키: %.1fcm\n",
		lastName, firstName, birthYear, height)
	fmt.Println()

	// ─────────────────────────────────────────
	// 4. 상수 (const)
	// ─────────────────────────────────────────
	fmt.Println("--- 4. 상수 (const) ---")

	// const: 컴파일 타임에 값이 결정되는 변수
	// 한 번 선언되면 변경 불가
	const Pi = 3.14159265358979
	const MaxRetry = 3
	const AppName = "MyApp"
	const IsDevelopment = true

	fmt.Printf("Pi: %.10f\n", Pi)
	fmt.Printf("최대 재시도: %d\n", MaxRetry)
	fmt.Printf("앱 이름: %s, 개발 모드: %t\n", AppName, IsDevelopment)

	// const 블록
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	fmt.Printf("1KB = %d bytes\n", KB)
	fmt.Printf("1MB = %d bytes\n", MB)
	fmt.Printf("1GB = %d bytes\n", GB)
	fmt.Println()

	// ─────────────────────────────────────────
	// 5. iota: 열거형 상수
	// ─────────────────────────────────────────
	fmt.Println("--- 5. iota (열거형 상수) ---")

	// iota는 const 블록 안에서 0부터 시작하여 1씩 증가하는 자동 증가 카운터입니다.
	const (
		Sunday    = iota // 0
		Monday           // 1
		Tuesday          // 2
		Wednesday        // 3
		Thursday         // 4
		Friday           // 5
		Saturday         // 6
	)
	fmt.Printf("일요일=%d, 월요일=%d, 화요일=%d\n", Sunday, Monday, Tuesday)
	fmt.Printf("수요일=%d, 목요일=%d, 금요일=%d, 토요일=%d\n", Wednesday, Thursday, Friday, Saturday)

	// iota 활용: 비트 플래그
	const (
		ReadPermission  = 1 << iota // 1 << 0 = 1 (001)
		WritePermission              // 1 << 1 = 2 (010)
		ExecPermission               // 1 << 2 = 4 (100)
	)
	fmt.Printf("\n읽기 권한: %d (이진수: %03b)\n", ReadPermission, ReadPermission)
	fmt.Printf("쓰기 권한: %d (이진수: %03b)\n", WritePermission, WritePermission)
	fmt.Printf("실행 권한: %d (이진수: %03b)\n", ExecPermission, ExecPermission)
	fmt.Printf("읽기+쓰기: %d (이진수: %03b)\n", ReadPermission|WritePermission, ReadPermission|WritePermission)

	// iota 활용: 크기 단위 (이전 값 건너뛰기)
	const (
		_  = iota             // 첫 번째 값(0) 버리기
		KB2 = 1 << (10 * iota) // 1 << 10 = 1024
		MB2                    // 1 << 20
		GB2                    // 1 << 30
	)
	fmt.Printf("\nKB=%d, MB=%d, GB=%d\n", KB2, MB2, GB2)
	fmt.Println()

	// ─────────────────────────────────────────
	// 6. 기본 타입 (Primitive Types)
	// ─────────────────────────────────────────
	fmt.Println("--- 6. 기본 타입 ---")

	// 정수 타입
	var i8 int8 = 127          // -128 ~ 127
	var i16 int16 = 32767      // -32768 ~ 32767
	var i32 int32 = 2147483647 // -2^31 ~ 2^31-1
	var i64 int64 = 9223372036854775807
	var u8 uint8 = 255    // 0 ~ 255
	var u16 uint16 = 65535
	var defaultInt int = 100 // 플랫폼에 따라 32 또는 64비트

	fmt.Printf("int8:  %d (크기: %d bytes)\n", i8, unsafe.Sizeof(i8))
	fmt.Printf("int16: %d (크기: %d bytes)\n", i16, unsafe.Sizeof(i16))
	fmt.Printf("int32: %d (크기: %d bytes)\n", i32, unsafe.Sizeof(i32))
	fmt.Printf("int64: %d (크기: %d bytes)\n", i64, unsafe.Sizeof(i64))
	fmt.Printf("uint8: %d (크기: %d bytes)\n", u8, unsafe.Sizeof(u8))
	fmt.Printf("uint16: %d (크기: %d bytes)\n", u16, unsafe.Sizeof(u16))
	fmt.Printf("int:  %d (크기: %d bytes)\n", defaultInt, unsafe.Sizeof(defaultInt))

	// 부동소수점 타입
	var f32 float32 = 3.14
	var f64 float64 = math.Pi
	fmt.Printf("\nfloat32: %.7f (크기: %d bytes)\n", f32, unsafe.Sizeof(f32))
	fmt.Printf("float64: %.15f (크기: %d bytes)\n", f64, unsafe.Sizeof(f64))

	// 복소수 타입
	var c64 complex64 = 1 + 2i
	var c128 complex128 = 3 + 4i
	fmt.Printf("\ncomplex64:  %v (실수부: %.1f, 허수부: %.1f)\n", c64, real(c64), imag(c64))
	fmt.Printf("complex128: %v (실수부: %.1f, 허수부: %.1f)\n", c128, real(c128), imag(c128))

	// bool 타입
	var flag bool = true
	fmt.Printf("\nbool: %t (크기: %d bytes)\n", flag, unsafe.Sizeof(flag))

	// string 타입
	str := "Hello, 世界"
	fmt.Printf("\nstring: %q\n", str)
	fmt.Printf("바이트 수: %d\n", len(str)) // len()은 바이트 수 반환

	// byte (uint8의 별칭)
	var byteVal byte = 'A'
	fmt.Printf("\nbyte: %d = '%c' (크기: %d bytes)\n", byteVal, byteVal, unsafe.Sizeof(byteVal))

	// rune (int32의 별칭, Unicode 코드 포인트)
	var r rune = '한'
	fmt.Printf("rune: %d = '%c' (크기: %d bytes)\n", r, r, unsafe.Sizeof(r))
	fmt.Println()

	// ─────────────────────────────────────────
	// 7. 제로값 (Zero Values) 요약
	// ─────────────────────────────────────────
	fmt.Println("--- 7. 제로값 (Zero Values) ---")
	// Go의 모든 변수는 선언 시 해당 타입의 제로값으로 초기화됩니다.
	// 이것은 C/C++처럼 초기화되지 않은 변수로 인한 버그를 방지합니다.

	var zeroInt int
	var zeroFloat float64
	var zeroBool bool
	var zeroString string
	var zeroRune rune

	fmt.Printf("int    제로값: %d\n", zeroInt)
	fmt.Printf("float64 제로값: %f\n", zeroFloat)
	fmt.Printf("bool   제로값: %t\n", zeroBool)
	fmt.Printf("string 제로값: %q (빈 문자열)\n", zeroString)
	fmt.Printf("rune   제로값: %d\n", zeroRune)

	// 포인터, 슬라이스, 맵, 채널, 함수의 제로값은 nil
	var zeroPtr *int
	fmt.Printf("pointer 제로값: %v\n", zeroPtr)
	fmt.Println()

	// ─────────────────────────────────────────
	// 8. 타입 변환 (Type Conversion)
	// ─────────────────────────────────────────
	fmt.Println("--- 8. 타입 변환 ---")
	// Go는 암시적 타입 변환이 없습니다. 반드시 명시적으로 변환해야 합니다.

	var intVal int = 42
	var floatVal float64 = float64(intVal) // int -> float64
	var int32Val int32 = int32(intVal)     // int -> int32
	var byteVal2 byte = byte(intVal)       // int -> byte (uint8)

	fmt.Printf("원래 int: %d\n", intVal)
	fmt.Printf("float64으로 변환: %f\n", floatVal)
	fmt.Printf("int32로 변환: %d\n", int32Val)
	fmt.Printf("byte로 변환: %d\n", byteVal2)

	// float -> int 변환 시 소수점 버려짐 (반올림 아님!)
	pi := 3.99
	piInt := int(pi) // 3 (소수점 잘림)
	fmt.Printf("\nfloat %.2f -> int: %d (소수점 잘림)\n", pi, piInt)

	// string <-> []byte <-> []rune 변환
	s := "Hello, 한국"
	b2 := []byte(s)   // string -> []byte
	r2 := []rune(s)   // string -> []rune
	s2 := string(b2)  // []byte -> string
	fmt.Printf("\n문자열: %q\n", s)
	fmt.Printf("바이트 슬라이스: %v\n", b2)
	fmt.Printf("룬 슬라이스: %v\n", r2)
	fmt.Printf("바이트에서 복원: %q\n", s2)
	fmt.Printf("바이트 수: %d, 문자(룬) 수: %d\n", len(b2), len(r2))

	// 정수 -> string 변환 (주의!)
	num := 65
	// string(num)은 아스키 코드 65 = 'A'로 변환됩니다. 숫자 문자열 "65"가 아닙니다!
	charFromNum := string(rune(num))
	fmt.Printf("\n정수 %d -> string: %q (아스키 문자)\n", num, charFromNum)
	// 숫자를 문자열 "65"로 변환하려면 fmt.Sprintf 또는 strconv 패키지를 사용하세요.
	numStr := fmt.Sprintf("%d", num)
	fmt.Printf("정수 %d -> 숫자 문자열: %q\n", num, numStr)
	fmt.Println()

	// ─────────────────────────────────────────
	// 9. 변수 사용 규칙 요약
	// ─────────────────────────────────────────
	fmt.Println("--- 9. 변수 사용 규칙 요약 ---")
	fmt.Println("1. Go에서 선언된 변수는 반드시 사용해야 합니다.")
	fmt.Println("   (사용하지 않으면 컴파일 에러 발생)")
	fmt.Println("2. := 는 함수 내부에서만 사용 가능합니다.")
	fmt.Println("3. var는 함수 외부(패키지 수준)와 내부 모두 사용 가능합니다.")
	fmt.Println("4. 상수(const)는 런타임에 변경할 수 없습니다.")
	fmt.Println("5. Go는 암시적 타입 변환을 허용하지 않습니다.")
}
