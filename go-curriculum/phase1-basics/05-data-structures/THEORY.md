# 05-data-structures: 배열, 슬라이스, 맵

## 배열 (Array)

배열은 **고정 크기**의 동일 타입 요소들의 연속된 메모리입니다.

```go
// 선언 방법들
var arr1 [5]int                         // 제로값으로 초기화 [0,0,0,0,0]
arr2 := [5]int{10, 20, 30, 40, 50}     // 리터럴
arr3 := [...]string{"사과", "바나나"}   // ... 으로 길이 자동 결정
arr4 := [5]int{0: 100, 2: 200, 4: 300} // 특정 인덱스만 초기화

// 접근
fmt.Println(arr2[0])    // 10
arr2[0] = 999
fmt.Println(len(arr2))  // 5

// 2차원 배열
var matrix [3][3]int
matrix[0][0] = 1
```

### 배열의 핵심 특성: 값 타입
```go
original := [3]int{1, 2, 3}
copied := original  // 완전한 복사 (참조가 아님!)
copied[0] = 999

fmt.Println(original)  // [1 2 3] (변경 안 됨)
fmt.Println(copied)    // [999 2 3]
```

### 배열 크기는 타입의 일부
```go
var a [3]int
var b [4]int
// a = b  // 컴파일 에러! [3]int != [4]int
```

**언제 배열을 사용?** 크기가 컴파일 타임에 고정되고 절대 변하지 않을 때. Go에서 배열은 직접보다 슬라이스의 기반으로 더 많이 쓰입니다.

---

## 슬라이스 (Slice) — Go의 핵심 자료구조

슬라이스는 동적 크기의 배열입니다. 내부적으로 배열에 대한 **뷰(view)**입니다.

### 슬라이스의 내부 구조
```
slice 헤더 (3개 필드):
┌──────────┬──────┬─────────┐
│ ptr      │ len  │  cap    │
│ (배열 주소)│ (길이)│ (용량)  │
└──────────┴──────┴─────────┘
       │
       ▼
┌───┬───┬───┬───┬───┬───┐
│ 0 │ 1 │ 2 │ 3 │ 4 │ 5 │  ← 실제 배열
└───┴───┴───┴───┴───┴───┘
```
- `len`: 슬라이스가 현재 담고 있는 요소 수
- `cap`: 재할당 없이 담을 수 있는 최대 요소 수

### 슬라이스 생성 방법
```go
// 1. nil 슬라이스 (len=0, cap=0, ptr=nil)
var s []int

// 2. 리터럴 (len=cap=요소 수)
s2 := []int{1, 2, 3, 4, 5}

// 3. make([]T, len, cap)
s3 := make([]int, 5)      // len=5, cap=5, 모두 0
s4 := make([]int, 3, 10)  // len=3, cap=10, 처음 3개만 0

// len과 cap 확인
fmt.Println(len(s4), cap(s4))  // 3 10
```

---

## 슬라이싱 (Slicing)

기존 슬라이스/배열에서 부분 슬라이스를 만듭니다.

```go
base := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

// s[low:high]: base[low] 부터 base[high-1] 까지
fmt.Println(base[2:5])  // [2 3 4]  (인덱스 2,3,4)
fmt.Println(base[:3])   // [0 1 2]  (처음부터 3개)
fmt.Println(base[7:])   // [7 8 9]  (인덱스 7부터 끝까지)
fmt.Println(base[:])    // [0..9]   (전체)
```

### 슬라이스 공유 주의
```go
base := []int{0, 1, 2, 3, 4, 5}
shared := base[2:5]  // [2 3 4] — 원본 배열 공유!

shared[0] = 999      // base[2]도 변경됨!
fmt.Println(base)    // [0 1 999 3 4 5]  원본도 바뀜!

// 독립적인 복사본이 필요하면 copy 사용
independent := make([]int, 3)
copy(independent, base[2:5])
independent[0] = 777
fmt.Println(base)    // 원본 변경 없음
```

### 3-index 슬라이싱 `s[low:high:max]`
```go
base := []int{0, 1, 2, 3, 4, 5}
// base[1:3:4]: 인덱스 1~2 요소, cap = max-low = 4-1 = 3
limited := base[1:3:4]
fmt.Println(limited, len(limited), cap(limited))
// [1 2] 2 3

// cap을 제한하여 실수로 원본을 덮어쓰는 것을 방지
```

---

## append와 용량 증가 전략

```go
var s []int
for i := 1; i <= 10; i++ {
    prevCap := cap(s)
    s = append(s, i)
    if cap(s) != prevCap {
        fmt.Printf("len=%d, cap: %d -> %d\n", len(s), prevCap, cap(s))
    }
}
// len=1, cap: 0 -> 1
// len=2, cap: 1 -> 2
// len=3, cap: 2 -> 4
// len=5, cap: 4 -> 8
// len=9, cap: 8 -> 16
```

**용량 증가 전략 (구현별로 다를 수 있음):**
- 작은 슬라이스 (cap < 256): 2배씩 증가
- 큰 슬라이스 (cap >= 256): 1.25배 + 추가량 (점진적 감소)

**append의 주의점**: 용량이 초과되면 새 배열을 할당하고 데이터를 복사합니다. 이때 이전 슬라이스와의 메모리 공유가 끊어집니다.

```go
// append는 반드시 반환값을 받아야 합니다
s = append(s, 1)       // 정상
append(s, 1)           // 컴파일되지만 결과가 버려짐 (vet 경고)

// 여러 요소 추가
s = append(s, 4, 5, 6)

// 슬라이스 붙이기
s2 := []int{7, 8, 9}
s = append(s, s2...)   // ... 로 펼쳐서 전달
```

---

## copy

```go
src := []int{1, 2, 3, 4, 5}
dst := make([]int, len(src))

n := copy(dst, src)  // n = 복사된 요소 수 = min(len(dst), len(src))
fmt.Println(n, dst)  // 5 [1 2 3 4 5]

// 복사 후 수정은 독립적
dst[0] = 999
fmt.Println(src)  // [1 2 3 4 5] 원본 변경 없음

// 부분 복사
small := make([]int, 3)
copy(small, src)     // src의 앞 3개만 복사
fmt.Println(small)   // [1 2 3]

// 슬라이스 내 이동 (overlapping 가능)
data := []int{1, 2, 3, 4, 5}
copy(data[1:], data)  // data[1:]에 data를 복사
fmt.Println(data)     // [1 1 2 3 4]
```

---

## nil 슬라이스 vs 빈 슬라이스

```go
var nilSlice []int          // nil 슬라이스: ptr=nil, len=0, cap=0
emptySlice := []int{}       // 빈 슬라이스: ptr != nil, len=0, cap=0
emptyMake := make([]int, 0) // 빈 슬라이스

fmt.Println(nilSlice == nil)   // true
fmt.Println(emptySlice == nil) // false
```

**nil 슬라이스에도 append, len, cap 사용 가능**합니다. `append`는 nil 슬라이스도 안전하게 처리합니다.

```go
var s []int
s = append(s, 1, 2, 3)  // 정상 동작
```

**JSON 직렬화 차이:**
```go
// encoding/json에서:
// nil 슬라이스   -> null
// 빈 슬라이스   -> []
type Response struct {
    Items []string
}
r1 := Response{Items: nil}        // {"Items":null}
r2 := Response{Items: []string{}} // {"Items":[]}
```

API 응답에서 `[]`와 `null`을 구분해야 할 때 이 차이가 중요합니다.

---

## 슬라이스 실용 패턴

```go
// 스택 (Stack)
stack := []int{}
stack = append(stack, 1, 2, 3)   // push
top := stack[len(stack)-1]         // peek
stack = stack[:len(stack)-1]       // pop

// 큐 (Queue)  — 성능이 중요하면 container/list 사용
queue := []int{1, 2, 3}
front := queue[0]
queue = queue[1:]  // dequeue

// 요소 삭제 (순서 유지): append로 이어붙이기
data := []int{1, 2, 3, 4, 5}
i := 2
data = append(data[:i], data[i+1:]...)
// [1 2 4 5]

// 요소 삭제 (순서 무관, 더 빠름): 마지막 요소로 교체
data[i] = data[len(data)-1]
data = data[:len(data)-1]
```

---

## 맵 (Map)

맵은 키-값 쌍의 비순서 컬렉션입니다. Go의 맵은 해시맵으로 구현됩니다.

### 생성과 기본 연산
```go
// make로 생성
scores := make(map[string]int)
scores["Alice"] = 95
scores["Bob"] = 87

// 리터럴로 생성
capitals := map[string]string{
    "한국": "서울",
    "일본": "도쿄",
    "미국": "워싱턴 D.C.",
}

// 조회
v := scores["Alice"]       // 95
v2 := scores["존재않는키"] // 0 (int 제로값, 패닉 없음)

// 삭제
delete(scores, "Bob")
delete(scores, "없는키")  // 안전, 패닉 없음

// 길이
fmt.Println(len(scores))
```

### comma-ok 패턴 — 키 존재 여부 확인
```go
// score가 0인 것인지, 키가 없어서 0인 것인지 구분할 수 없을 때
score, ok := scores["Dave"]
if ok {
    fmt.Printf("Dave의 점수: %d\n", score)
} else {
    fmt.Println("Dave는 없음")
}

// 한 줄로
if score, ok := scores["Eve"]; ok {
    fmt.Printf("Eve: %d\n", score)
}
```

### 맵 순회 — 순서 보장 안 됨
```go
for country, capital := range capitals {
    fmt.Printf("%s -> %s\n", country, capital)
}
// 순서는 실행마다 다름! (Go가 의도적으로 랜덤화)
```

정렬된 순서로 순회하려면:
```go
keys := make([]string, 0, len(capitals))
for k := range capitals {
    keys = append(keys, k)
}
sort.Strings(keys)
for _, k := range keys {
    fmt.Printf("%s -> %s\n", k, capitals[k])
}
```

### nil 맵 주의
```go
var m map[string]int  // nil 맵
v := m["key"]         // 0 반환 (안전)
delete(m, "key")      // 안전 (패닉 없음)
m["key"] = 1          // 패닉! nil 맵에는 쓸 수 없음

// 반드시 초기화 후 사용
m = make(map[string]int)
m["key"] = 1  // 정상
```

### 동시성 주의
```go
// 맵은 동시성에서 안전하지 않습니다!
// 여러 고루틴에서 동시에 읽고 쓰면 런타임 패닉

// 해결책 1: sync.Mutex
var mu sync.Mutex
mu.Lock()
m["key"] = value
mu.Unlock()

// 해결책 2: sync.Map (읽기가 많은 경우)
var sm sync.Map
sm.Store("key", value)
v, ok := sm.Load("key")
```

---

## Python/Java와의 비교

### list vs slice
```python
# Python list: 동적, 다양한 타입 가능
lst = [1, "hello", True]
lst.append(4)
```
```go
// Go slice: 동적이지만 단일 타입
s := []int{1, 2, 3}
s = append(s, 4)
// 다양한 타입: []interface{} 또는 []any
mixed := []any{1, "hello", true}
```

### dict vs map
```python
# Python dict: 3.7+ 삽입 순서 보장
d = {"a": 1, "b": 2}
d["c"] = 3
del d["a"]
if "b" in d: ...
```
```go
// Go map: 순서 보장 없음
m := map[string]int{"a": 1, "b": 2}
m["c"] = 3
delete(m, "a")
if _, ok := m["b"]; ok { ... }
```

### ArrayList vs slice
```java
// Java ArrayList
ArrayList<Integer> list = new ArrayList<>();
list.add(1);
list.add(2);
int size = list.size();
list.get(0);
list.remove(0);
```
```go
// Go slice
s := []int{}
s = append(s, 1, 2)
size := len(s)
first := s[0]
s = append(s[:0], s[1:]...)  // 인덱스 0 삭제
```

---

## 핵심 정리

1. **배열**은 값 타입 (복사됨), **슬라이스**는 참조 타입 (배열을 공유)
2. 슬라이스 내부 구조: `[ptr | len | cap]` — 이 셋을 항상 인식
3. `append` 후 반환값을 반드시 받아야 함 (`s = append(s, ...)`)
4. 슬라이스 공유로 인한 의도치 않은 수정 주의 → 독립 복사 필요시 `copy`
5. 맵 순회 순서는 비결정적 — 정렬이 필요하면 키를 추출 후 정렬
6. nil 맵에 쓰기는 패닉 — `make`로 초기화 필수
7. 맵은 goroutine-safe하지 않음 — 동시성 환경에서 뮤텍스 필요
