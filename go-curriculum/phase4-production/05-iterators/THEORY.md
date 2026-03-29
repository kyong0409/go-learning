# 05-iterators: Go 이터레이터 패턴 (Go 1.23)

> Go 1.23에서 `range-over-function` 이터레이터가 정식 도입되었습니다. `iter.Seq`와 `iter.Seq2` 타입으로 일관된 순회 패턴을 정의하고, `for-range` 루프에서 직접 사용할 수 있습니다.

---

## 1. 배경: 이터레이터가 필요했던 이유

### Go 1.22 이전의 문제

Go에는 컬렉션 순회를 위한 통일된 패턴이 없었습니다.

```go
// 슬라이스 순회
for i, v := range slice { ... }

// 맵 순회
for k, v := range myMap { ... }

// 채널 순회
for v := range ch { ... }

// 커스텀 컬렉션? — 제각각
tree.Walk(func(v int) { ... })         // 콜백
for iter := tree.Iterator(); iter.Next(); { iter.Value() }  // 이터레이터 객체
vals := tree.All()                     // 슬라이스로 모두 반환
```

**문제점**:
1. 언어 내장 순회(`for-range`)를 커스텀 타입에 사용할 수 없었음
2. 라이브러리마다 다른 순회 API
3. 무한 시퀀스나 지연 평가를 표현하기 어려움

---

## 2. iter.Seq와 iter.Seq2

Go 1.23에서 `iter` 패키지에 두 가지 이터레이터 타입이 추가되었습니다.

```go
package iter

// 단일 값 이터레이터
type Seq[V any] func(yield func(V) bool)

// 키-값 쌍 이터레이터
type Seq2[K, V any] func(yield func(K, V) bool)
```

이터레이터는 단순히 함수 타입입니다. `yield` 함수를 받아서 값을 하나씩 `yield`에 전달합니다. `yield`가 `false`를 반환하면(break 등) 순회를 중단합니다.

---

## 3. range-over-function: for 루프에서 이터레이터 사용

Go 1.23의 핵심 변경사항입니다. `iter.Seq`를 반환하는 함수를 `for-range`에 직접 사용할 수 있습니다.

```go
// 피보나치 이터레이터 정의
func Fibonacci() iter.Seq[int] {
    return func(yield func(int) bool) {
        a, b := 0, 1
        for {
            if !yield(a) { return }  // break 시 yield가 false 반환
            a, b = b, a+b
        }
    }
}

// Go 1.23 — for-range로 직접 사용
for n := range Fibonacci() {
    if n > 100 { break }
    fmt.Println(n)
}
```

```go
// iter.Seq2 — 키-값 쌍
func Enumerate[V any](s []V) iter.Seq2[int, V] {
    return func(yield func(int, V) bool) {
        for i, v := range s {
            if !yield(i, v) { return }
        }
    }
}

// 사용
for i, fruit := range Enumerate([]string{"사과", "바나나", "체리"}) {
    fmt.Printf("%d: %s\n", i, fruit)
}
```

---

## 4. yield 함수의 동작 원리

`yield` 함수는 Go 컴파일러가 `for-range` 루프의 내용을 래핑하여 생성합니다.

```go
// 컴파일러가 이것을
for v := range someSeq {
    if v > 5 { break }
    fmt.Println(v)
}

// 대략 이렇게 변환
someSeq(func(v int) bool {
    if v > 5 { return false }  // break → false 반환
    fmt.Println(v)
    return true  // continue
})
```

`yield(v)`의 반환값 의미:
- `true`: 계속 순회
- `false`: 순회 중단 (`break`, `return`, `goto` 등)

---

## 5. 이터레이터 작성 패턴

### 유한 시퀀스

```go
func Range(start, end int) iter.Seq[int] {
    return func(yield func(int) bool) {
        for i := start; i < end; i++ {
            if !yield(i) { return }
        }
    }
}

// 사용
for n := range Range(1, 6) {
    fmt.Print(n, " ")  // 1 2 3 4 5
}
```

### 무한 시퀀스

```go
func Naturals() iter.Seq[int] {
    return func(yield func(int) bool) {
        for n := 1; ; n++ {
            if !yield(n) { return }
        }
    }
}

// 무한 시퀀스지만 break로 안전하게 중단
for n := range Naturals() {
    if n > 10 { break }
    fmt.Print(n, " ")
}
```

### 트리 순회 (재귀 이터레이터)

```go
type TreeNode struct {
    Value       int
    Left, Right *TreeNode
}

func (n *TreeNode) InOrder() iter.Seq[int] {
    return func(yield func(int) bool) {
        var traverse func(*TreeNode) bool
        traverse = func(node *TreeNode) bool {
            if node == nil { return true }
            if !traverse(node.Left) { return false }
            if !yield(node.Value) { return false }
            return traverse(node.Right)
        }
        traverse(n)
    }
}

// 사용
for v := range tree.InOrder() {
    fmt.Print(v, " ")  // 정렬된 순서로 출력
}
```

---

## 6. 고차 이터레이터 함수

이터레이터를 변환하는 함수들입니다. **지연 평가(lazy evaluation)**가 핵심입니다.

```go
// Filter — 조건을 만족하는 요소만 통과
func Filter[V any](seq iter.Seq[V], pred func(V) bool) iter.Seq[V] {
    return func(yield func(V) bool) {
        for v := range seq {
            if pred(v) {
                if !yield(v) { return }
            }
        }
    }
}

// Map — 각 요소 변환
func Map[V, U any](seq iter.Seq[V], f func(V) U) iter.Seq[U] {
    return func(yield func(U) bool) {
        for v := range seq {
            if !yield(f(v)) { return }
        }
    }
}

// Take — 처음 n개만
func Take[V any](seq iter.Seq[V], n int) iter.Seq[V] {
    return func(yield func(V) bool) {
        count := 0
        for v := range seq {
            if count >= n { return }
            if !yield(v) { return }
            count++
        }
    }
}

// 체이닝: 짝수 피보나치의 처음 5개
for n := range Take(Filter(Fibonacci(), func(n int) bool { return n%2 == 0 }), 5) {
    fmt.Print(n, " ")  // 0 2 8 34 144
}
```

---

## 7. 표준 라이브러리 이터레이터 (Go 1.23+)

### strings / bytes 패키지 — Go 1.24 추가 함수

Go 1.24에서 `strings`와 `bytes` 패키지에 이터레이터를 반환하는 함수들이 추가되었습니다.

```go
import "strings"

text := "첫 번째 줄\n두 번째 줄\n세 번째 줄"

// strings.Lines: 줄 단위 이터레이터 (iter.Seq[string])
for line := range strings.Lines(text) {
    fmt.Println(line)
}
// "첫 번째 줄\n", "두 번째 줄\n", "세 번째 줄"

// strings.SplitSeq: 구분자 기준 분리 이터레이터
for part := range strings.SplitSeq("a,b,c,d", ",") {
    fmt.Print(part, " ")  // a b c d
}

// strings.FieldsSeq: 공백 기준 분리 이터레이터
for word := range strings.FieldsSeq("  Go  is  fun  ") {
    fmt.Print(word, " ")  // Go is fun
}

// bytes 패키지도 동일한 함수 제공
import "bytes"

data := []byte("line1\nline2\nline3")
for line := range bytes.Lines(data) {
    fmt.Println(string(line))
}
```

**이전 방식과 비교**:
```go
// Go 1.23 이하 — 슬라이스로 모두 분리 후 순회 (메모리 낭비)
for _, line := range strings.Split(text, "\n") { ... }

// Go 1.24+ — 지연 평가 이터레이터 (메모리 효율적)
for line := range strings.Lines(text) { ... }
```

### slices 패키지

```go
import "slices"

s := []int{10, 20, 30, 40, 50}

// slices.All: 인덱스-값 쌍 (iter.Seq2[int, V])
for i, v := range slices.All(s) {
    fmt.Printf("[%d]=%d ", i, v)
}

// slices.Values: 값만 (iter.Seq[V])
for v := range slices.Values(s) {
    fmt.Print(v, " ")
}

// slices.Backward: 역순 (iter.Seq2[int, V])
for i, v := range slices.Backward(s) {
    fmt.Printf("[%d]=%d ", i, v)
}

// slices.Collect: iter.Seq → []T 수집
doubled := slices.Collect(Map(slices.Values(s), func(n int) int { return n * 2 }))
// [20 40 60 80 100]

// slices.Sorted: iter.Seq → 정렬된 []T
sorted := slices.Sorted(slices.Values([]int{5, 2, 8, 1}))
// [1 2 5 8]
```

### maps 패키지

```go
import "maps"

m := map[string]int{"사과": 3, "바나나": 1, "체리": 5}

// maps.All: 키-값 쌍 (iter.Seq2[K, V])
for k, v := range maps.All(m) {
    fmt.Printf("%s: %d\n", k, v)
}

// maps.Keys: 키만 (iter.Seq[K])
for k := range maps.Keys(m) {
    fmt.Println(k)
}

// maps.Values: 값만 (iter.Seq[V])
for v := range maps.Values(m) {
    fmt.Println(v)
}

// maps.Collect: iter.Seq2[K,V] → map[K]V
filtered := maps.Collect(Filter2(maps.All(m), func(k string, v int) bool {
    return v > 2
}))
```

---

## 8. Pull 이터레이터: iter.Pull과 iter.Pull2

`iter.Pull`은 "밀기(push)" 방식의 이터레이터를 "당기기(pull)" 방식으로 변환합니다. 두 이터레이터를 동시에 순회하거나 외부에서 순회를 제어할 때 유용합니다.

```go
// Zip: 두 이터레이터를 쌍으로 묶기
func Zip[A, B any](a iter.Seq[A], b iter.Seq[B]) iter.Seq2[A, B] {
    return func(yield func(A, B) bool) {
        // Pull 이터레이터로 변환: next()로 하나씩 당김
        nextA, stopA := iter.Pull(a)
        nextB, stopB := iter.Pull(b)
        defer stopA()
        defer stopB()

        for {
            va, okA := nextA()
            vb, okB := nextB()
            if !okA || !okB { return }
            if !yield(va, vb) { return }
        }
    }
}

// 사용
names := slices.Values([]string{"Alice", "Bob", "Carol"})
scores := slices.Values([]int{95, 82, 78})

for name, score := range Zip(names, scores) {
    fmt.Printf("%s: %d\n", name, score)
}
```

---

## 9. 언제 이터레이터를, 언제 슬라이스를 반환할지

| 상황 | 권장 |
|------|------|
| 무한 시퀀스 | 이터레이터 |
| 지연 평가가 필요한 큰 데이터 | 이터레이터 |
| 체이닝 파이프라인 | 이터레이터 |
| 결과를 여러 번 사용 | 슬라이스 |
| 길이를 미리 알아야 함 | 슬라이스 |
| 간단한 목록 반환 | 슬라이스 |

```go
// 큰 파일 라인별 처리 — 이터레이터 (메모리 효율적)
func Lines(r io.Reader) iter.Seq[string] {
    return func(yield func(string) bool) {
        scanner := bufio.NewScanner(r)
        for scanner.Scan() {
            if !yield(scanner.Text()) { return }
        }
    }
}

// 짧은 목록 — 슬라이스가 더 간단
func GetTopUsers(n int) []User {
    return users[:n]
}
```

---

## 10. Python / Java 비교

### Python generator vs Go iterator

```python
# Python generator
def fibonacci():
    a, b = 0, 1
    while True:
        yield a          # yield 키워드로 값 생성
        a, b = b, a + b

# 사용
for n in fibonacci():
    if n > 100: break
    print(n)
```

```go
// Go iterator
func Fibonacci() iter.Seq[int] {
    return func(yield func(int) bool) {
        a, b := 0, 1
        for {
            if !yield(a) { return }  // yield 함수 호출
            a, b = b, a+b
        }
    }
}

// 사용
for n := range Fibonacci() {
    if n > 100 { break }
    fmt.Println(n)
}
```

Python `generator`는 언어 키워드(`yield`)를 사용하지만, Go 이터레이터는 함수 타입과 클로저로만 구현됩니다. 언어 확장 없이 라이브러리 수준에서 동일한 패턴을 표현합니다.

### Java Stream vs Go iter.Seq

```java
// Java Stream — 지연 평가 파이프라인
Stream.iterate(0, n -> n + 1)
    .filter(n -> n % 2 == 0)
    .map(n -> n * n)
    .limit(5)
    .forEach(System.out::println);
```

```go
// Go — 이터레이터 체이닝
for n := range Take(
    Map(
        Filter(Naturals(), func(n int) bool { return n%2 == 0 }),
        func(n int) int { return n * n },
    ),
    5,
) {
    fmt.Println(n)
}
```

Java `Stream`은 풍부한 메서드가 빌트인되어 있지만, Go 이터레이터는 일반 함수를 조합합니다. Go 방식이 더 명시적이고 커스텀 연산자를 추가하기 쉽습니다.
