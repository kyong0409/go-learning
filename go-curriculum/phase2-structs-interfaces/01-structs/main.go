// 패키지 선언: 실행 가능한 프로그램의 진입점
package main

import (
	"encoding/json"
	"fmt"
)

// ─────────────────────────────────────────
// 1. 구조체(Struct) 정의
// ─────────────────────────────────────────

// Person은 사람을 나타내는 기본 구조체입니다.
// 구조체는 관련된 데이터를 하나의 타입으로 묶는 방법입니다.
type Person struct {
	Name string // 이름 (공개 필드: 대문자로 시작)
	Age  int    // 나이
	City string // 도시
}

// Address는 주소 정보를 담는 구조체입니다.
type Address struct {
	Street string
	City   string
	Zip    string
}

// Employee는 중첩 구조체(nested struct)를 보여줍니다.
// Address 구조체를 필드로 포함합니다.
type Employee struct {
	Name    string
	ID      int
	Address Address // 구조체 안에 구조체 (중첩)
}

// ─────────────────────────────────────────
// 2. JSON 태그가 있는 구조체
// ─────────────────────────────────────────

// Product는 JSON 태그를 사용하는 구조체입니다.
// JSON 태그는 인코딩/디코딩 시 필드 이름을 지정합니다.
type Product struct {
	// `json:"name"`: JSON에서 "name" 키로 매핑
	Name string `json:"name"`

	// `json:"price"`: JSON에서 "price" 키로 매핑
	Price float64 `json:"price"`

	// `json:"description,omitempty"`: 값이 비어있으면 JSON 출력에서 생략
	Description string `json:"description,omitempty"`

	// `json:"-"`: JSON에서 완전히 무시 (민감한 정보 등)
	InternalCode string `json:"-"`

	// `json:"stock"`: 숫자 0도 출력 (omitempty 없음)
	Stock int `json:"stock"`

	// `json:"tags,omitempty"`: 슬라이스가 nil이면 생략
	Tags []string `json:"tags,omitempty"`
}

// ─────────────────────────────────────────
// 3. 구조체 비교를 위한 타입
// ─────────────────────────────────────────

// Point는 2D 좌표를 나타냅니다.
// 모든 필드가 비교 가능(comparable)하면 == 연산자로 비교할 수 있습니다.
type Point struct {
	X, Y int // 같은 타입의 필드는 한 줄에 선언 가능
}

// Slice를 포함하는 구조체는 == 으로 비교할 수 없습니다.
type Container struct {
	Name  string
	Items []string // 슬라이스는 비교 불가 → == 사용 불가
}

func main() {
	fmt.Println("=== Go Phase 2: 구조체(Structs) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 구조체 리터럴로 생성하기
	// ─────────────────────────────────────────
	fmt.Println("--- 1. 구조체 리터럴 ---")

	// 방법 1: 필드 이름 지정 (권장 방식 - 순서 무관, 가독성 좋음)
	p1 := Person{
		Name: "홍길동",
		Age:  30,
		City: "서울",
	}
	fmt.Printf("방법 1 (필드명 지정): %+v\n", p1) // %+v: 필드명 포함 출력

	// 방법 2: 위치 기반 초기화 (필드 순서를 정확히 알아야 함 - 비권장)
	p2 := Person{"김영희", 25, "부산"}
	fmt.Printf("방법 2 (위치 기반): %+v\n", p2)

	// 방법 3: 일부 필드만 초기화 (나머지는 zero value)
	p3 := Person{Name: "이철수"}
	fmt.Printf("방법 3 (일부 필드): %+v\n", p3) // Age=0, City=""

	// 방법 4: new() 사용 - 포인터 반환
	p4 := new(Person)
	p4.Name = "박지수"
	p4.Age = 28
	fmt.Printf("방법 4 (new 사용): %+v\n", *p4)
	fmt.Println()

	// ─────────────────────────────────────────
	// 필드 접근
	// ─────────────────────────────────────────
	fmt.Println("--- 2. 필드 접근 ---")

	person := Person{Name: "최민준", Age: 35, City: "대전"}

	// 점(.) 연산자로 필드 접근
	fmt.Printf("이름: %s\n", person.Name)
	fmt.Printf("나이: %d\n", person.Age)
	fmt.Printf("도시: %s\n", person.City)

	// 필드 수정
	person.City = "인천"
	fmt.Printf("수정 후 도시: %s\n", person.City)
	fmt.Println()

	// ─────────────────────────────────────────
	// 구조체 포인터
	// ─────────────────────────────────────────
	fmt.Println("--- 3. 구조체 포인터 ---")

	// 구조체 포인터 생성
	pPtr := &Person{Name: "강하늘", Age: 22, City: "광주"}

	// Go는 포인터를 통한 필드 접근 시 자동으로 역참조합니다.
	// (*pPtr).Name 대신 pPtr.Name 으로 접근 가능 (syntactic sugar)
	fmt.Printf("포인터를 통한 접근: %s, %d세\n", pPtr.Name, pPtr.Age)

	// 함수에 구조체를 포인터로 전달하면 원본이 수정됩니다.
	growUp(pPtr)
	fmt.Printf("growUp 후 나이: %d\n", pPtr.Age)

	// 값으로 전달하면 복사본이 수정됩니다 (원본 불변).
	original := Person{Name: "테스트", Age: 20, City: "서울"}
	modifyValue(original)
	fmt.Printf("값 전달 후 원본 나이: %d (변경 없음)\n", original.Age)
	fmt.Println()

	// ─────────────────────────────────────────
	// 중첩 구조체
	// ─────────────────────────────────────────
	fmt.Println("--- 4. 중첩 구조체 ---")

	emp := Employee{
		Name: "정수연",
		ID:   1001,
		Address: Address{
			Street: "테헤란로 152",
			City:   "서울",
			Zip:    "06236",
		},
	}

	// 중첩 필드 접근: 점(.) 체이닝
	fmt.Printf("직원: %s (ID: %d)\n", emp.Name, emp.ID)
	fmt.Printf("주소: %s, %s %s\n", emp.Address.Street, emp.Address.City, emp.Address.Zip)

	// 중첩 필드 수정
	emp.Address.City = "부산"
	fmt.Printf("이전 후 도시: %s\n", emp.Address.City)
	fmt.Println()

	// ─────────────────────────────────────────
	// JSON 태그 활용
	// ─────────────────────────────────────────
	fmt.Println("--- 5. JSON 태그 ---")

	// Description 있는 상품
	laptop := Product{
		Name:         "노트북",
		Price:        1299.99,
		Description:  "고성능 개발자용 노트북",
		InternalCode: "LAP-001", // JSON에서 제외됨
		Stock:        15,
		Tags:         []string{"전자제품", "컴퓨터"},
	}

	// 구조체 → JSON 인코딩
	laptopJSON, err := json.MarshalIndent(laptop, "", "  ")
	if err != nil {
		fmt.Printf("JSON 인코딩 에러: %v\n", err)
		return
	}
	fmt.Println("Description 있는 상품 JSON:")
	fmt.Println(string(laptopJSON))

	// Description 없는 상품 (omitempty 효과 확인)
	simple := Product{
		Name:  "USB 케이블",
		Price: 9.99,
		// Description 생략 → JSON에서 제외
		// Tags 생략 → JSON에서 제외
		Stock: 0, // 0이지만 omitempty 없으므로 출력됨
	}
	simpleJSON, _ := json.MarshalIndent(simple, "", "  ")
	fmt.Println("\nomitempty 효과 (Description, Tags 없음):")
	fmt.Println(string(simpleJSON))

	// JSON → 구조체 디코딩
	jsonStr := `{"name":"마우스","price":29.99,"stock":50,"tags":["주변기기"]}`
	var decoded Product
	if err := json.Unmarshal([]byte(jsonStr), &decoded); err != nil {
		fmt.Printf("JSON 디코딩 에러: %v\n", err)
		return
	}
	fmt.Printf("\nJSON 디코딩 결과: %+v\n", decoded)
	fmt.Println()

	// ─────────────────────────────────────────
	// 구조체 비교
	// ─────────────────────────────────────────
	fmt.Println("--- 6. 구조체 비교 ---")

	// 모든 필드가 comparable이면 == 사용 가능
	pt1 := Point{X: 3, Y: 4}
	pt2 := Point{X: 3, Y: 4}
	pt3 := Point{X: 1, Y: 2}

	fmt.Printf("pt1 == pt2: %v (같은 값)\n", pt1 == pt2)   // true
	fmt.Printf("pt1 == pt3: %v (다른 값)\n", pt1 == pt3)   // false
	fmt.Printf("pt1 != pt3: %v\n", pt1 != pt3)             // true

	// 구조체도 map의 키로 사용 가능 (comparable할 때)
	distances := map[Point]string{
		{0, 0}: "원점",
		{1, 0}: "오른쪽",
		{0, 1}: "위쪽",
	}
	fmt.Printf("좌표 (0,0): %s\n", distances[Point{0, 0}])

	// 주의: 슬라이스를 포함한 구조체는 == 으로 비교 불가
	// c1 := Container{Name: "a", Items: []string{"x"}}
	// c2 := Container{Name: "a", Items: []string{"x"}}
	// c1 == c2 // 컴파일 에러! invalid operation: cannot compare
	fmt.Println("(슬라이스 포함 구조체는 == 비교 불가 - reflect.DeepEqual 사용)")
	fmt.Println()

	// ─────────────────────────────────────────
	// 익명 구조체 (Anonymous Struct)
	// ─────────────────────────────────────────
	fmt.Println("--- 7. 익명 구조체 ---")

	// 타입 이름 없이 즉석에서 구조체 정의
	// 일회성 데이터 구조에 유용 (테스트, 임시 데이터 등)
	config := struct {
		Host string
		Port int
		TLS  bool
	}{
		Host: "localhost",
		Port: 8080,
		TLS:  false,
	}
	fmt.Printf("서버 설정: %s:%d (TLS: %v)\n", config.Host, config.Port, config.TLS)

	// 슬라이스 안의 익명 구조체 (테이블 기반 테스트에 자주 사용)
	testCases := []struct {
		input    int
		expected string
	}{
		{1, "one"},
		{2, "two"},
		{3, "three"},
	}
	for _, tc := range testCases {
		fmt.Printf("  입력 %d → 기대값 %q\n", tc.input, tc.expected)
	}

	fmt.Println()
	fmt.Println("=== 구조체 학습 완료 ===")
}

// growUp은 포인터 수신으로 원본 구조체를 수정합니다.
func growUp(p *Person) {
	p.Age++ // 원본 수정
}

// modifyValue는 값 복사본을 수정합니다 (원본 불변).
func modifyValue(p Person) {
	p.Age = 999 // 복사본만 수정됨
}
