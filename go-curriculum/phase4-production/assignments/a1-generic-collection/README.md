# 과제 A1: 제네릭 컬렉션 라이브러리 구현

## 목표

Go 제네릭을 사용하여 네 가지 자료구조를 구현합니다.

## 구현할 타입

### 1. `Stack[T any]` — LIFO 스택
```go
s := &Stack[int]{}
s.Push(1); s.Push(2); s.Push(3)
v, ok := s.Pop()   // v=3, ok=true
v, ok = s.Peek()   // v=2, ok=true (제거 안 됨)
s.Len()            // 2
s.IsEmpty()        // false
```

### 2. `Queue[T any]` — FIFO 큐
```go
q := &Queue[string]{}
q.Enqueue("a"); q.Enqueue("b"); q.Enqueue("c")
v, ok := q.Dequeue()  // v="a", ok=true
q.Front()             // "b", true
q.Len()               // 2
```

### 3. `Set[T comparable]` — 집합
```go
s := NewSet(1, 2, 3, 4, 5)
s.Add(6)
s.Remove(1)
s.Contains(3)          // true
s.Len()                // 5
s.Union(NewSet(4,5,6,7))       // {2,3,4,5,6,7}
s.Intersection(NewSet(3,4,5))  // {3,4,5}
s.Difference(NewSet(4,5))      // {2,3,6}
```

### 4. `OrderedMap[K comparable, V any]` — 삽입 순서를 유지하는 맵
```go
m := NewOrderedMap[string, int]()
m.Set("banana", 2); m.Set("apple", 1); m.Set("cherry", 3)
m.Get("apple")    // 1, true
m.Keys()          // ["banana", "apple", "cherry"] (삽입 순서)
m.Values()        // [2, 1, 3]
m.Delete("apple")
m.Len()           // 2
```

## 채점 기준

| 항목 | 배점 |
|------|------|
| Stack 구현 | 25점 |
| Queue 구현 | 25점 |
| Set 구현 | 25점 |
| OrderedMap 구현 | 25점 |
| **합계** | **100점** |

## 실행 방법

```bash
cd assignments/a1-generic-collection
go test ./... -v
```

## 힌트

- `Stack`과 `Queue`는 슬라이스로 구현하세요.
- `Set`은 `map[T]struct{}`를 내부적으로 사용하세요.
- `OrderedMap`은 맵과 키 순서 슬라이스를 함께 관리하세요.
- `comparable` 제약조건: `==` 연산자가 필요할 때 사용합니다.
