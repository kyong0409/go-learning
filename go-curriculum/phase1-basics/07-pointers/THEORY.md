# 07-pointers: 포인터

## 포인터란 무엇인가

포인터는 **메모리 주소를 담는 변수**입니다. 값 자체가 아니라 값이 저장된 위치를 가리킵니다.

```
메모리 레이아웃:

주소    값
0x1000  42    ← x (int 변수)
0x1008  0x1000 ← ptr (*int 포인터, x의 주소를 담음)
```

```go
x := 42
ptr := &x  // ptr은 x의 주소를 담음

fmt.Println(x)    // 42     (x의 값)
fmt.Println(&x)   // 0x1000 (x의 메모리 주소)
fmt.Println(ptr)  // 0x1000 (ptr이 담고 있는 주소)
fmt.Println(*ptr) // 42     (ptr이 가리키는 값)
```

---

## & 와 * 연산자

```go
x := 42

// & (주소 연산자): 변수의 메모리 주소를 반환
ptr := &x
fmt.Printf("타입: %T\n", ptr)   // *int

// * (역참조/간접참조 연산자): 포인터가 가리키는 값에 접근
fmt.Println(*ptr)   // 42

// 포인터를 통해 원본 값 수정
*ptr = 100
fmt.Println(x)      // 100 (x도 바뀜!)

// 포인터도 변수: 다른 주소를 가리킬 수 있음
y := 200
ptr = &y
fmt.Println(*ptr)   // 200
```

### 포인터 타입 표기
```go
var p *int         // int를 가리키는 포인터
var pp **int       // *int를 가리키는 포인터 (포인터의 포인터)
var ps *string     // string을 가리키는 포인터
var pf *float64    // float64를 가리키는 포인터
```

---

## 값 전달 vs 포인터 전달

Go에서 **모든 것은 값으로 전달됩니다**. 함수에 인자를 넘기면 항상 복사가 일어납니다.

### 값 전달 (Pass by Value)
```go
func incrementByValue(n int) {
    n++  // 복사본만 변경, 원본 불변
    fmt.Printf("  함수 내부: %d\n", n)
}

num := 10
incrementByValue(num)
fmt.Println(num)  // 10 (변경 안 됨)
```

### 포인터 전달 (Pass by Pointer)
```go
func incrementByPointer(n *int) {
    *n++  // 포인터가 가리키는 원본 값 변경
    fmt.Printf("  함수 내부: %d\n", *n)
}

num := 10
incrementByPointer(&num)
fmt.Println(num)  // 11 (변경됨)
```

### 구조체 전달
```go
type Person struct {
    Name string
    Age  int
}

// 값 전달: 구조체 전체가 복사됨 (크면 비효율)
func birthdayByValue(p Person) {
    p.Age++  // 복사본만 변경
}

// 포인터 전달: 주소(8바이트)만 복사됨 (효율적 + 수정 가능)
func birthdayByPointer(p *Person) {
    p.Age++  // 원본 변경
}

alice := Person{"Alice", 30}
birthdayByValue(alice)
fmt.Println(alice.Age)  // 30 (변경 안 됨)

birthdayByPointer(&alice)
fmt.Println(alice.Age)  // 31 (변경됨)
```

---

## 슬라이스와 맵 — "이미 참조"

슬라이스와 맵은 내부적으로 포인터를 포함하는 헤더 구조체입니다. 함수에 전달할 때 헤더가 복사되지만, 헤더 안의 포인터는 같은 데이터를 가리킵니다.

```go
// 슬라이스: 요소 수정은 원본에 반영됨
func doubleAll(s []int) {
    for i := range s {
        s[i] *= 2  // 원본 배열 수정됨
    }
}

s := []int{1, 2, 3}
doubleAll(s)
fmt.Println(s)  // [2 4 6] (변경됨!)

// 하지만 append는 헤더를 변경하므로 포인터가 필요
func appendTo(s *[]int, val int) {
    *s = append(*s, val)
}

appendTo(&s, 100)
fmt.Println(s)  // [2 4 6 100]

// 잘못된 방법 (원본 변경 안 됨):
func wrongAppend(s []int, val int) {
    s = append(s, val)  // 로컬 복사본만 변경
}
wrongAppend(s, 999)
fmt.Println(s)  // [2 4 6 100] (999 없음!)
```

```go
// 맵: 함수 내에서 수정 가능 (참조 타입)
func addEntry(m map[string]int, key string, val int) {
    m[key] = val  // 원본 맵 수정됨
}

m := map[string]int{"a": 1}
addEntry(m, "b", 2)
fmt.Println(m)  // map[a:1 b:2]
```

---

## nil 포인터와 안전한 처리

```go
var p *int  // nil 포인터 (제로값)
fmt.Println(p == nil)  // true
fmt.Println(p)         // <nil>

// nil 포인터 역참조는 런타임 패닉!
// fmt.Println(*p)  // panic: runtime error: nil pointer dereference

// 안전한 처리: 사용 전 nil 확인
if p != nil {
    fmt.Println(*p)
} else {
    fmt.Println("포인터가 nil입니다")
}

// 안전한 역참조 헬퍼 함수 패턴
func safeDeref(p *int, defaultVal int) int {
    if p == nil {
        return defaultVal
    }
    return *p
}
```

---

## new() vs &T{}

```go
// new(T): 타입 T의 제로값을 힙에 할당하고 *T 반환
p1 := new(int)         // *int, 값 = 0
p2 := new(string)      // *string, 값 = ""
p3 := new(bool)        // *bool, 값 = false

*p1 = 42
fmt.Println(*p1)  // 42

// &T{}: 구조체/리터럴의 주소 반환 (초기값 지정 가능)
type Point struct{ X, Y int }
p4 := &Point{X: 1, Y: 2}   // 초기값 지정 가능
p5 := &Point{}               // 제로값 (new(Point)와 동일)

// 기본 타입에 리터럴 주소 반환
n := 42
p6 := &n          // 변수가 필요
// p7 := &42      // 컴파일 에러! 리터럴에 & 불가

// new vs &T{} 선택
// - 구조체: &T{field: val} 이 더 관용적
// - 기본 타입: new(int) 보다는 n := 0; &n 이 더 명확
// - 실제로 new()는 잘 안 씀
```

---

## 포인터 수신자 (Pointer Receivers)

메서드에서 구조체를 수정하거나, 큰 구조체를 효율적으로 전달하려면 포인터 수신자를 사용합니다.

```go
type Counter struct {
    value int
    name  string
}

// 값 수신자: 복사본에서 동작, 원본 불변
func (c Counter) Get() int {
    return c.value
}

// 포인터 수신자: 원본 수정 가능
func (c *Counter) Increment() {
    c.value++
}

func (c *Counter) Reset() {
    c.value = 0
}

c := Counter{name: "test", value: 0}
c.Increment()   // Go가 자동으로 (&c).Increment() 로 변환
c.Increment()
fmt.Println(c.Get())  // 2

// 포인터 변수에서 값 수신자 메서드도 자동 역참조
cp := &Counter{name: "ptr", value: 10}
cp.Increment()         // (*cp).Increment() — 포인터 수신자
fmt.Println(cp.Get())  // cp.Get() 도 자동 (*cp).Get()
```

**언제 포인터 수신자를 써야 하는가?**
1. 메서드가 수신자를 수정해야 할 때
2. 구조체가 클 때 (복사 비용 절약)
3. 뮤텍스(sync.Mutex) 포함 구조체 — 복사하면 안 됨
4. 일관성: 하나라도 포인터 수신자면 나머지도 포인터 수신자

---

## 이스케이프 분석: 스택 vs 힙

Go 컴파일러는 자동으로 변수를 스택 또는 힙에 할당합니다.

```go
// 스택 할당: 함수 반환 후 자동 해제
func stackAlloc() {
    x := 42         // 스택
    _ = x
}

// 힙 할당: 함수 반환 후에도 살아있어야 할 때 (이스케이프)
func heapAlloc() *int {
    x := 42
    return &x   // x의 주소를 반환 → x는 힙으로 이스케이프
}
// &x를 반환하면 x가 스택에 있으면 위험
// Go 컴파일러가 이를 감지하고 자동으로 힙에 할당

// 이스케이프 분석 확인
// go build -gcflags="-m" main.go
```

Go는 가비지 컬렉터가 있으므로 힙 메모리를 직접 해제할 필요가 없습니다. 하지만 힙 할당은 스택보다 느리고 GC 부하를 유발합니다.

---

## 포인터로 Optional 값 표현

Go에는 `Optional<T>`나 `null` 안전 타입이 없습니다. 포인터의 nil을 "값 없음"으로 활용합니다.

```go
type User struct {
    Name     string
    Age      int
    Nickname *string  // nil이면 닉네임 없음
    Score    *int     // nil이면 점수 미등록
}

// 포인터 리터럴 헬퍼 (Go에서 자주 쓰는 패턴)
func strPtr(s string) *string { return &s }
func intPtr(n int) *int       { return &n }

u1 := User{Name: "Alice", Age: 30, Nickname: strPtr("앨리스"), Score: intPtr(95)}
u2 := User{Name: "Bob", Age: 25}  // Nickname=nil, Score=nil

if u1.Nickname != nil {
    fmt.Printf("닉네임: %s\n", *u1.Nickname)
}
if u2.Score == nil {
    fmt.Println("점수 미등록")
}
```

---

## C와의 차이 — 포인터 산술 없음

```c
// C: 포인터 산술 가능 (위험)
int arr[] = {1, 2, 3};
int *p = arr;
p++;        // 다음 요소로 이동
*(p + 2);   // 2칸 앞 요소
```

```go
// Go: 포인터 산술 없음 (안전)
arr := [3]int{1, 2, 3}
p := &arr[0]
// p++  // 컴파일 에러!
// 배열/슬라이스 인덱싱으로 대체
arr[1]   // 인덱스로 접근

// unsafe 패키지로 포인터 산술이 가능하지만 사용 금지
// (CGo, 시스템 프로그래밍 등 특수 목적에만)
```

---

## Python/Java와의 비교

### Python: 모든 것이 참조
```python
# Python에서 변수는 객체 참조
a = [1, 2, 3]
b = a          # 같은 객체를 가리킴
b.append(4)
print(a)       # [1, 2, 3, 4] (a도 변경!)

b = b[:]       # 슬라이스로 복사
b.append(5)
print(a)       # [1, 2, 3, 4] (a는 변경 안 됨)
```

### Java: 기본 타입은 값, 객체는 참조
```java
// Java: int는 값, Integer는 참조
int x = 42;
modify(x);  // 복사본 전달
System.out.println(x);  // 42

Person p = new Person("Alice");
modifyPerson(p);  // 참조 전달 (포인터처럼)
```

### Go: 명시적 포인터
```go
// Go: 명시적으로 포인터를 전달
x := 42
modify(&x)  // 주소를 명시적으로 전달
fmt.Println(x)  // 수정됨

// 또는 값으로 전달하면 복사
modifyByValue(x)  // 복사 전달
fmt.Println(x)    // 수정 안 됨
```

Go는 Python처럼 모든 것이 자동으로 참조되지도 않고, Java처럼 타입에 따라 다르지도 않습니다. **모든 것이 값으로 전달되지만, 그 값이 포인터일 수 있다**는 원칙이 명확합니다.

---

## 핵심 정리

1. `&` = 주소 가져오기, `*` = 역참조 (가리키는 값 접근)
2. Go는 모든 것을 값으로 전달 — 포인터를 전달하면 주소값이 복사됨
3. 슬라이스와 맵은 내부에 포인터를 품고 있어 포인터 없이도 원본을 수정 가능 (단, append는 헤더 교체이므로 `*[]T` 필요)
4. nil 포인터 역참조는 패닉 — 사용 전 nil 체크 필수
5. 포인터 수신자: 구조체 수정 또는 대형 구조체 전달 시 사용
6. Go에는 포인터 산술이 없음 — C의 위험한 패턴 불가
