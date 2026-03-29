# 03-generics: Go 제네릭 (타입 파라미터)

> Go 1.18에서 도입된 제네릭은 "타입 파라미터를 가진 함수와 타입"을 작성할 수 있게 합니다. 코드 중복 없이 여러 타입에서 동작하는 범용 컬렉션과 유틸리티를 만들 수 있습니다.

---

## 1. 제네릭이 해결하는 문제

### 제네릭 이전: 타입별 중복 구현

```go
// int 슬라이스용 Contains
func ContainsInt(s []int, v int) bool {
    for _, x := range s { if x == v { return true } }
    return false
}

// string 슬라이스용 Contains
func ContainsString(s []string, v string) bool {
    for _, x := range s { if x == v { return true } }
    return false
}

// interface{} 사용 — 타입 안전성 없음
func ContainsAny(s []interface{}, v interface{}) bool {
    for _, x := range s { if x == v { return true } }
    return false
}
```

### 제네릭 이후: 단일 범용 함수

```go
// comparable 제약: == 연산자를 사용할 수 있는 모든 타입
func Contains[T comparable](s []T, v T) bool {
    for _, x := range s {
        if x == v {
            return true
        }
    }
    return false
}

// 사용: 타입 추론으로 [int], [string] 생략 가능
Contains([]int{1, 2, 3}, 2)         // true
Contains([]string{"a", "b"}, "c")   // false
```

---

## 2. Go에 제네릭이 늦게 온 이유

Go는 2009년에 등장했지만 제네릭은 2022년(Go 1.18)에야 도입되었습니다.

**10년간 논쟁의 핵심**:

1. **단순성**: Go의 핵심 가치는 단순성입니다. 제네릭은 언어를 복잡하게 만듭니다.
2. **인터페이스로 충분?**: `interface{}`와 타입 단언으로 많은 경우를 처리할 수 있었습니다.
3. **구현 방식 논쟁**: C++ 스타일(템플릿)? Java 스타일(타입 소거)? Go 팀은 최종적으로 **monomorphization**을 선택했습니다.

**Monomorphization**: 컴파일 시 각 타입 조합에 대해 별도의 코드를 생성합니다. `Contains[int]`와 `Contains[string]`은 별도의 기계어 코드로 컴파일됩니다. Java의 타입 소거(erasure)와 달리 런타임 오버헤드가 없습니다.

---

## 3. 타입 파라미터 문법

```go
// 함수 타입 파라미터: 함수명 뒤 [ ] 안에 선언
func Map[T, U any](s []T, f func(T) U) []U {
    result := make([]U, len(s))
    for i, v := range s {
        result[i] = f(v)
    }
    return result
}

// 타입 타입 파라미터: 타입명 뒤 [ ] 안에 선언
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

// 사용
var s Stack[string]
s.Push("hello")

nums := Map([]int{1, 2, 3}, func(n int) string {
    return fmt.Sprintf("%d", n)
})
// nums = []string{"1", "2", "3"}
```

---

## 4. 타입 제약조건 (Constraints)

제약조건은 타입 파라미터가 충족해야 할 조건을 정의하는 인터페이스입니다.

### any — 모든 타입

```go
// any = interface{} 의 별칭
func Identity[T any](v T) T { return v }
```

### comparable — 동등 비교 가능

```go
// == 연산자를 지원하는 타입: 기본 타입, 포인터, 배열, comparable 필드만 가진 구조체
func Index[T comparable](s []T, v T) int {
    for i, x := range s {
        if x == v { return i }
    }
    return -1
}
```

### cmp.Ordered — 순서 비교 가능

```go
import "cmp"

// cmp.Ordered = ~int | ~int8 | ~int16 | ... | ~float32 | ~float64 | ~string
func Min[T cmp.Ordered](a, b T) T {
    if a < b { return a }
    return b
}

Min(3, 7)           // 3 (int)
Min(3.14, 2.72)     // 2.72 (float64)
Min("apple", "banana") // "apple" (string)
```

### 인터페이스 제약조건 — 메서드 요구

```go
type Stringer interface {
    String() string
}

func Print[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}
```

### 타입 집합 — underlying type 제약

```go
// ~ 접두사: 해당 타입을 기반으로 정의된 모든 타입 포함
type Number interface {
    ~int | ~int32 | ~int64 | ~float32 | ~float64
}

func Sum[T Number](s []T) T {
    var total T
    for _, v := range s { total += v }
    return total
}

// 커스텀 타입도 동작
type Celsius float64
temps := []Celsius{36.5, 37.2, 38.0}
Sum(temps)  // Celsius 타입으로 반환
```

---

## 5. 제네릭 함수: Map, Filter, Reduce

함수형 프로그래밍의 핵심 패턴을 타입 안전하게 구현합니다.

```go
// Map: T 슬라이스 → U 슬라이스 변환
func Map[T, U any](s []T, f func(T) U) []U {
    result := make([]U, len(s))
    for i, v := range s { result[i] = f(v) }
    return result
}

// Filter: 조건을 만족하는 요소만 반환
func Filter[T any](s []T, pred func(T) bool) []T {
    var result []T
    for _, v := range s {
        if pred(v) { result = append(result, v) }
    }
    return result
}

// Reduce: 슬라이스 → 단일 값 축약
func Reduce[T, U any](s []T, init U, f func(U, T) U) U {
    result := init
    for _, v := range s { result = f(result, v) }
    return result
}

// 조합 사용
nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// 짝수만 골라 제곱한 후 합산
result := Reduce(
    Map(
        Filter(nums, func(n int) bool { return n%2 == 0 }),
        func(n int) int { return n * n },
    ),
    0,
    func(acc, n int) int { return acc + n },
)
// (2²+4²+6²+8²+10²) = 220
```

---

## 6. 제네릭 타입: Stack, Set, Cache

```go
// Stack[T]: 범용 LIFO 자료구조
type Stack[T any] struct{ items []T }

func (s *Stack[T]) Push(item T)        { s.items = append(s.items, item) }
func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 { var zero T; return zero, false }
    top := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return top, true
}

// Set[T]: comparable 제약 — 맵 키로 사용하기 위해 필요
type Set[T comparable] struct{ items map[T]struct{} }

func NewSet[T comparable](items ...T) *Set[T] {
    s := &Set[T]{items: make(map[T]struct{})}
    for _, item := range items { s.items[item] = struct{}{} }
    return s
}

// Cache[K, V]: 동시성 안전 제네릭 캐시
type Cache[K comparable, V any] struct {
    mu    sync.RWMutex
    items map[K]V
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    v, ok := c.items[key]
    return v, ok
}
```

---

## 7. 표준 라이브러리의 제네릭 (Go 1.21+)

Go 1.21에서 제네릭을 사용한 표준 패키지가 추가되었습니다.

### slices 패키지

```go
import "slices"

nums := []int{3, 1, 4, 1, 5, 9}

slices.Sort(nums)                            // 정렬 (제자리)
sorted := slices.Sorted(slices.Values(nums)) // 원본 유지 정렬 (Go 1.23+)
slices.Contains(nums, 5)                     // 포함 여부
slices.Index(nums, 9)                        // 인덱스 반환
slices.Max(nums)                             // 최댓값
slices.Min(nums)                             // 최솟값
slices.Reverse(nums)                         // 역순 (제자리)
slices.Compact(nums)                         // 연속 중복 제거
slices.Clone(nums)                           // 얕은 복사

// 커스텀 비교 함수
slices.SortFunc(people, func(a, b Person) int {
    return cmp.Compare(a.Age, b.Age)
})
```

### maps 패키지

```go
import "maps"

m := map[string]int{"a": 1, "b": 2}
clone := maps.Clone(m)             // 맵 복사
maps.Copy(dst, src)                // 한 맵을 다른 맵에 복사
maps.DeleteFunc(m, func(k string, v int) bool {
    return v < 2  // 조건에 맞는 항목 삭제
})
maps.Equal(m1, m2)                 // 두 맵 동등 비교
```

### cmp 패키지

```go
import "cmp"

cmp.Compare(1, 2)    // -1 (1 < 2)
cmp.Compare(2, 2)    //  0 (같음)
cmp.Compare(3, 2)    //  1 (3 > 2)

// Or: 첫 번째 비제로 값 반환 (nil 대체 패턴)
cmp.Or(0, 0, 42)     // 42
cmp.Or("", "기본값") // "기본값"
```

---

## 8. 제네릭을 쓰지 말아야 할 때

제네릭이 항상 정답은 아닙니다.

```go
// 과도한 추상화 — 인터페이스가 더 명확한 경우
// 나쁜 예
func Process[T interface{ Process() error }](items []T) error {
    for _, item := range items { item.Process() }
    return nil
}

// 좋은 예 — 인터페이스로 충분
type Processor interface{ Process() error }
func Process(items []Processor) error {
    for _, item := range items { item.Process() }
    return nil
}
```

**제네릭을 사용해야 할 때**:
- 다양한 타입의 슬라이스/맵을 다루는 유틸리티 함수
- `Stack[T]`, `Queue[T]`, `Set[T]` 같은 자료구조
- `Map[T,U]`, `Filter[T]`, `Reduce[T,U]` 같은 함수형 유틸리티

**인터페이스를 사용해야 할 때**:
- 런타임에 동적으로 타입이 결정되는 경우
- 메서드 기반 다형성이 필요한 경우
- 이미 `interface`로 표현할 수 있는 경우

---

## 9. 제네릭 타입 별칭 (Go 1.24 안정화)

Go 1.23에서 실험적으로 도입되었던 제네릭 타입 별칭(alias type parameters)이 **Go 1.24(2025년 2월)에 정식 안정화**되었습니다.

```go
// Go 1.24+ — 타입 별칭에도 타입 파라미터 사용 가능 (stable)
type StringMap[V any] = map[string]V

// 사용
var m StringMap[int] = map[string]int{"a": 1}

// 더 복잡한 예: 제네릭 함수 타입 별칭
type Predicate[T any] = func(T) bool
type Transform[T, U any] = func(T) U

// 활용
var isEven Predicate[int] = func(n int) bool { return n%2 == 0 }
```

### Go 1.26: 자기 참조 제네릭 타입

Go 1.26(2026년 2월)에서는 자기 참조(self-referential) 제네릭 타입을 더 자연스럽게 표현할 수 있게 되었습니다.

```go
// Go 1.26+ — 자기 참조 제네릭 타입 (재귀 타입 제약 개선)
// 예: 정렬 가능한 트리 노드
type TreeNode[T interface{ Compare(T) int }] struct {
    Value       T
    Left, Right *TreeNode[T]
}
```

---

## 10. Python / Java 비교

### Python typing vs Go generics

```python
# Python — 런타임 타입 힌트 (실제 강제 없음)
from typing import TypeVar, List

T = TypeVar('T')

def first(items: List[T]) -> T:
    return items[0]  # 런타임에 타입 검사 없음
```

```go
// Go — 컴파일 타임 타입 검사
func First[T any](items []T) (T, bool) {
    if len(items) == 0 {
        var zero T
        return zero, false
    }
    return items[0], true
}
```

Python의 타입 힌트는 도구(mypy)가 검사하지만 런타임에는 무시됩니다. Go 제네릭은 컴파일 타임에 완전히 검사됩니다.

### Java generics vs Go generics

```java
// Java — 타입 소거(erasure): 런타임에 List<String>과 List<Integer>는 모두 List
public <T> List<T> filter(List<T> list, Predicate<T> pred) {
    return list.stream().filter(pred).collect(Collectors.toList());
}
// 런타임에 T의 타입 정보 없음 → instanceof T 불가
```

```go
// Go — monomorphization: 컴파일 시 각 타입에 대해 별도 코드 생성
// Filter[int]와 Filter[string]은 별도의 기계어
func Filter[T any](s []T, pred func(T) bool) []T { ... }
```

Java 제네릭의 타입 소거로 인한 `InstanceOf<T>` 불가, 기본 타입 사용 불가(`List<int>` 불가, `List<Integer>` 강제) 같은 제약이 Go에는 없습니다.
