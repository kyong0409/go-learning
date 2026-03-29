// 패키지 선언
package main

import "fmt"

func main() {
	fmt.Println("=== Go 기초: 데이터 구조 (배열, 슬라이스, 맵) ===")
	fmt.Println()

	// ─────────────────────────────────────────
	// 1. 배열 (Array)
	// ─────────────────────────────────────────
	fmt.Println("--- 1. 배열 (Array) ---")

	// 배열: 고정 크기, 동일 타입 요소의 연속된 메모리
	// 크기는 타입의 일부입니다: [3]int != [4]int
	var arr1 [5]int // 제로값으로 초기화
	fmt.Printf("기본 배열: %v\n", arr1)

	// 배열 리터럴로 초기화
	arr2 := [5]int{10, 20, 30, 40, 50}
	fmt.Printf("초기화된 배열: %v\n", arr2)

	// ... 으로 길이 자동 결정
	arr3 := [...]string{"사과", "바나나", "체리"}
	fmt.Printf("자동 길이 배열: %v (길이: %d)\n", arr3, len(arr3))

	// 특정 인덱스만 초기화
	arr4 := [5]int{0: 100, 2: 200, 4: 300}
	fmt.Printf("인덱스 지정 초기화: %v\n", arr4)

	// 배열 요소 접근 및 수정
	arr2[0] = 999
	fmt.Printf("arr2[0] 변경 후: %v\n", arr2)

	// 2차원 배열
	var matrix [3][3]int
	for i := range 3 {
		for j := range 3 {
			matrix[i][j] = i*3 + j + 1
		}
	}
	fmt.Printf("3x3 행렬:\n")
	for _, row := range matrix {
		fmt.Printf("  %v\n", row)
	}

	// 배열은 값 타입: 대입 시 복사됨
	original := [3]int{1, 2, 3}
	copied := original // 완전한 복사
	copied[0] = 999
	fmt.Printf("원본: %v, 복사본(수정됨): %v\n", original, copied)
	fmt.Println()

	// ─────────────────────────────────────────
	// 2. 슬라이스 (Slice) - 기초
	// ─────────────────────────────────────────
	fmt.Println("--- 2. 슬라이스 기초 ---")

	// 슬라이스: 동적 크기 배열. 배열에 대한 참조입니다.
	// 내부 구조: [포인터 | 길이(len) | 용량(cap)]
	var s1 []int // nil 슬라이스 (길이=0, 용량=0, 포인터=nil)
	fmt.Printf("nil 슬라이스: %v, len=%d, cap=%d, nil=%t\n",
		s1, len(s1), cap(s1), s1 == nil)

	// 슬라이스 리터럴
	s2 := []int{1, 2, 3, 4, 5}
	fmt.Printf("슬라이스 리터럴: %v, len=%d, cap=%d\n", s2, len(s2), cap(s2))

	// make()로 슬라이스 생성: make([]T, len, cap)
	s3 := make([]int, 3)     // len=3, cap=3
	s4 := make([]int, 3, 10) // len=3, cap=10
	fmt.Printf("make([]int, 3):    %v, len=%d, cap=%d\n", s3, len(s3), cap(s3))
	fmt.Printf("make([]int, 3, 10): %v, len=%d, cap=%d\n", s4, len(s4), cap(s4))
	fmt.Println()

	// ─────────────────────────────────────────
	// 3. 슬라이싱 (Slicing)
	// ─────────────────────────────────────────
	fmt.Println("--- 3. 슬라이싱 ---")

	base := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Printf("기본: %v\n", base)

	// s[low:high]: low <= 인덱스 < high
	fmt.Printf("s[2:5]:  %v\n", base[2:5])  // [2,3,4]
	fmt.Printf("s[:3]:   %v\n", base[:3])    // [0,1,2]
	fmt.Printf("s[7:]:   %v\n", base[7:])    // [7,8,9]
	fmt.Printf("s[:]:    %v\n", base[:])     // 전체 복사본 참조

	// 슬라이스는 원본 배열을 공유합니다!
	shared := base[2:5]
	fmt.Printf("\nshared = base[2:5]: %v\n", shared)
	shared[0] = 999 // base[2]도 변경됨
	fmt.Printf("shared[0]=999 후:\n")
	fmt.Printf("  shared: %v\n", shared)
	fmt.Printf("  base:   %v (원본도 변경!)\n", base)

	// 3-인덱스 슬라이싱: s[low:high:max] - 용량 제한
	base2 := []int{0, 1, 2, 3, 4, 5}
	limited := base2[1:3:4] // len=2, cap=3 (max-low=4-1=3)
	fmt.Printf("\n3-인덱스 슬라이싱 base2[1:3:4]:\n")
	fmt.Printf("  %v, len=%d, cap=%d\n", limited, len(limited), cap(limited))
	fmt.Println()

	// ─────────────────────────────────────────
	// 4. append와 용량 증가
	// ─────────────────────────────────────────
	fmt.Println("--- 4. append와 용량 증가 ---")

	var growing []int
	fmt.Printf("초기: len=%d, cap=%d\n", len(growing), cap(growing))

	for i := 1; i <= 10; i++ {
		prevCap := cap(growing)
		growing = append(growing, i)
		if cap(growing) != prevCap {
			fmt.Printf("용량 증가! len=%d, cap: %d -> %d\n",
				len(growing), prevCap, cap(growing))
		}
	}
	fmt.Printf("최종: %v\n", growing)

	// append로 여러 요소 추가
	s5 := []int{1, 2, 3}
	s5 = append(s5, 4, 5, 6)
	fmt.Printf("\n여러 요소 append: %v\n", s5)

	// 슬라이스 붙이기: append(s1, s2...)
	s6 := []int{7, 8, 9}
	s5 = append(s5, s6...)
	fmt.Printf("슬라이스 붙이기: %v\n", s5)
	fmt.Println()

	// ─────────────────────────────────────────
	// 5. copy
	// ─────────────────────────────────────────
	fmt.Println("--- 5. copy ---")

	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, len(src))
	n := copy(dst, src) // n = 복사된 요소 수
	fmt.Printf("copy 결과: src=%v, dst=%v, 복사됨=%d\n", src, dst, n)

	// 독립적인 복사본이므로 수정해도 원본에 영향 없음
	dst[0] = 999
	fmt.Printf("dst[0]=999 후: src=%v, dst=%v\n", src, dst)

	// 부분 복사 (목적지가 더 작으면 min(len(dst), len(src))만큼 복사)
	small := make([]int, 3)
	copy(small, src)
	fmt.Printf("부분 복사: %v\n", small)

	// 슬라이스 내에서 이동 (overlapping)
	overlap := []int{1, 2, 3, 4, 5}
	copy(overlap[1:], overlap) // [1,1,2,3,4]
	fmt.Printf("겹치는 복사: %v\n", overlap)
	fmt.Println()

	// ─────────────────────────────────────────
	// 6. nil 슬라이스 vs 빈 슬라이스
	// ─────────────────────────────────────────
	fmt.Println("--- 6. nil 슬라이스 vs 빈 슬라이스 ---")

	var nilSlice []int       // nil 슬라이스
	emptySlice := []int{}    // 빈 슬라이스 (리터럴)
	emptyMake := make([]int, 0) // 빈 슬라이스 (make)

	fmt.Printf("nil 슬라이스:     %v, len=%d, nil=%t\n",
		nilSlice, len(nilSlice), nilSlice == nil)
	fmt.Printf("빈 슬라이스(리터럴): %v, len=%d, nil=%t\n",
		emptySlice, len(emptySlice), emptySlice == nil)
	fmt.Printf("빈 슬라이스(make):   %v, len=%d, nil=%t\n",
		emptyMake, len(emptyMake), emptyMake == nil)

	// nil 슬라이스에도 append 가능 (안전함)
	nilSlice = append(nilSlice, 1, 2, 3)
	fmt.Printf("nil 슬라이스에 append 후: %v\n", nilSlice)

	// 주의: JSON 직렬화 시 nil은 null, 빈 슬라이스는 []
	fmt.Println("주의: encoding/json에서 nil=null, 빈슬라이스=[]")
	fmt.Println()

	// ─────────────────────────────────────────
	// 7. 맵 (Map)
	// ─────────────────────────────────────────
	fmt.Println("--- 7. 맵 (Map) ---")

	// 맵 생성 방법 1: make
	scores := make(map[string]int)
	scores["Alice"] = 95
	scores["Bob"] = 87
	scores["Carol"] = 92

	// 맵 생성 방법 2: 리터럴
	capitals := map[string]string{
		"한국":  "서울",
		"일본":  "도쿄",
		"미국":  "워싱턴 D.C.",
		"영국":  "런던",
		"프랑스": "파리",
	}

	fmt.Printf("점수: %v\n", scores)
	fmt.Printf("수도: %v\n", capitals)

	// 맵 조회
	fmt.Printf("\nAlice 점수: %d\n", scores["Alice"])
	fmt.Printf("한국 수도: %s\n", capitals["한국"])

	// comma-ok 패턴: 키 존재 여부 확인
	if score, ok := scores["Dave"]; ok {
		fmt.Printf("Dave 점수: %d\n", score)
	} else {
		fmt.Println("Dave는 없습니다 (ok=false)")
	}

	// 존재하지 않는 키 조회: 제로값 반환 (안전함)
	missing := scores["없는키"]
	fmt.Printf("없는 키의 값: %d (int 제로값)\n", missing)

	// 맵 수정
	scores["Alice"] = 100 // 업데이트
	fmt.Printf("Alice 점수 업데이트: %d\n", scores["Alice"])

	// 맵에서 삭제
	delete(scores, "Bob")
	fmt.Printf("Bob 삭제 후: %v\n", scores)

	// 키 존재 확인 후 삭제 (안전, delete는 없는 키도 안전하게 처리)
	delete(scores, "없는키") // 패닉 없음

	// 맵 순회 (순서 보장 없음!)
	fmt.Println("\n수도 순회:")
	for country, capital := range capitals {
		fmt.Printf("  %s의 수도: %s\n", country, capital)
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 8. 중첩 맵과 슬라이스 맵
	// ─────────────────────────────────────────
	fmt.Println("--- 8. 중첩 데이터 구조 ---")

	// 슬라이스의 맵
	tags := map[string][]string{
		"Go":     {"컴파일", "정적타입", "병렬"},
		"Python": {"인터프리터", "동적타입", "쉬운문법"},
	}
	for lang, ts := range tags {
		fmt.Printf("%s: %v\n", lang, ts)
	}

	// 맵의 맵 (중첩 맵)
	students := map[string]map[string]int{
		"Alice": {"수학": 95, "영어": 88},
		"Bob":   {"수학": 72, "영어": 91},
	}
	for student, subjectScores := range students {
		fmt.Printf("%s: ", student)
		for subject, score := range subjectScores {
			fmt.Printf("%s=%d ", subject, score)
		}
		fmt.Println()
	}
	fmt.Println()

	// ─────────────────────────────────────────
	// 9. 슬라이스 실용 패턴
	// ─────────────────────────────────────────
	fmt.Println("--- 9. 슬라이스 실용 패턴 ---")

	// 스택 (Stack)
	stack := []int{}
	stack = append(stack, 1) // push
	stack = append(stack, 2)
	stack = append(stack, 3)
	fmt.Printf("스택: %v\n", stack)
	top := stack[len(stack)-1]         // peek
	stack = stack[:len(stack)-1]       // pop
	fmt.Printf("pop: %d, 나머지: %v\n", top, stack)

	// 큐 (Queue)
	queue := []int{1, 2, 3}
	front := queue[0]       // dequeue
	queue = queue[1:]
	fmt.Printf("dequeue: %d, 나머지: %v\n", front, queue)

	// 요소 삭제 (순서 유지)
	data := []int{1, 2, 3, 4, 5}
	i := 2 // 인덱스 2 삭제
	data = append(data[:i], data[i+1:]...)
	fmt.Printf("인덱스 2 삭제 (순서 유지): %v\n", data)

	// 요소 삭제 (순서 무관, 더 빠름)
	data2 := []int{1, 2, 3, 4, 5}
	data2[i] = data2[len(data2)-1]
	data2 = data2[:len(data2)-1]
	fmt.Printf("인덱스 2 삭제 (순서 무관): %v\n", data2)

	// 요소 삽입
	data3 := []int{1, 2, 4, 5}
	insertIdx := 2
	data3 = append(data3[:insertIdx+1], data3[insertIdx:]...)
	data3[insertIdx] = 3
	fmt.Printf("인덱스 2에 3 삽입: %v\n", data3)

	fmt.Println()
	fmt.Println("=== 완료 ===")
}
