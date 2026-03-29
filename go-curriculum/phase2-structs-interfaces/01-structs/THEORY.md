# 01-structs: 구조체 (Struct)

> Go에서 데이터를 묶는 유일한 방법. Python의 `dataclass`, Java의 `record`/`class`에 해당합니다.

---

## 1. 구조체 정의와 초기화

### 정의

```go
type Person struct {
    Name string
    Age  int
    City string
}
```

필드명이 대문자로 시작하면 패키지 외부에서 접근 가능(공개), 소문자면 패키지 내부에서만 접근 가능(비공개)입니다.

### 초기화 방법 4가지

```go
// 1. 필드명 지정 — 권장. 순서 무관, 가독성 좋음
p1 := Person{Name: "홍길동", Age: 30, City: "서울"}

// 2. 위치 기반 — 비권장. 필드 추가 시 모든 호출부가 깨진다
p2 := Person{"홍길동", 30, "서울"}

// 3. 부분 초기화 — 나머지 필드는 제로값
p3 := Person{Name: "이철수"}  // Age=0, City=""

// 4. new() — 포인터 반환, 모든 필드 제로값으로 시작
p4 := new(Person)
p4.Name = "박지수"
```

### 제로값 구조체

Go의 모든 타입에는 제로값이 있습니다. 구조체의 제로값은 각 필드의 제로값으로 채워진 구조체입니다.

```go
var p Person   // Name="", Age=0, City=""
// new(Person)과 동일한 상태 (차이: var은 값, new는 포인터)
```

---

## 2. 필드 접근과 수정

점(`.`) 연산자로 접근합니다.

```go
person := Person{Name: "최민준", Age: 35, City: "대전"}

fmt.Println(person.Name)  // 읽기
person.City = "인천"       // 수정
```

---

## 3. 구조체 포인터: & 연산자와 자동 역참조

```go
// & 연산자로 포인터 생성
pPtr := &Person{Name: "강하늘", Age: 22}

// 포인터를 통한 필드 접근 — Go가 자동으로 역참조
fmt.Println(pPtr.Name)   // (*pPtr).Name 과 동일, Go가 자동 처리

// 함수에 포인터 전달 — 원본 수정 가능
func birthday(p *Person) {
    p.Age++  // 원본 수정
}
birthday(pPtr)
```

**값 전달 vs 포인터 전달**

```go
// 값 전달: 복사본이 전달됨 — 원본 불변
func noEffect(p Person) { p.Age = 999 }

original := Person{Age: 20}
noEffect(original)
fmt.Println(original.Age)  // 20 (변경 없음)
```

---

## 4. 익명(anonymous) 구조체

타입 이름 없이 즉석에서 정의합니다. 일회성 데이터 구조나 테스트 케이스에 유용합니다.

```go
// 설정 데이터
config := struct {
    Host string
    Port int
    TLS  bool
}{
    Host: "localhost",
    Port: 8080,
    TLS:  false,
}

// 테이블 주도 테스트 케이스 — Go에서 매우 자주 쓰이는 패턴
testCases := []struct {
    input    int
    expected string
}{
    {1, "one"},
    {2, "two"},
    {3, "three"},
}
```

---

## 5. 구조체 태그

백틱(`` ` ``)으로 감싼 메타데이터입니다. 리플렉션을 통해 라이브러리가 읽습니다.

```go
type Product struct {
    Name         string   `json:"name"`
    Price        float64  `json:"price"`
    Description  string   `json:"description,omitempty"` // 빈 값이면 JSON 출력 생략
    InternalCode string   `json:"-"`                     // JSON에서 완전히 제외
    Stock        int      `json:"stock"`
    Tags         []string `json:"tags,omitempty"`
}
```

**주요 태그 종류**

| 태그 | 패키지 | 용도 |
|------|--------|------|
| `json:"name"` | `encoding/json` | JSON 키 이름 지정 |
| `xml:"name"` | `encoding/xml` | XML 요소 이름 지정 |
| `db:"column_name"` | `sqlx`, `gorm` | DB 컬럼 매핑 |
| `validate:"required,min=1"` | `go-playground/validator` | 유효성 검사 규칙 |

---

## 6. JSON 직렬화 / 역직렬화

```go
// 구조체 → JSON (직렬화)
product := Product{Name: "노트북", Price: 1299.99, Stock: 15}
data, err := json.Marshal(product)
// data = []byte(`{"name":"노트북","price":1299.99,"stock":15}`)

// 보기 좋게 출력 (들여쓰기)
pretty, _ := json.MarshalIndent(product, "", "  ")

// JSON → 구조체 (역직렬화)
jsonStr := `{"name":"마우스","price":29.99,"stock":50}`
var decoded Product
err = json.Unmarshal([]byte(jsonStr), &decoded)
// decoded.Name == "마우스"
```

**omitempty 동작**

- `string`: 빈 문자열 `""` 이면 생략
- `int`, `float64`: `0`이면 생략
- `bool`: `false`이면 생략
- `slice`, `map`: `nil`이면 생략 (길이 0인 빈 슬라이스는 생략 안 됨)

---

## 7. 구조체 비교

모든 필드가 `comparable`(비교 가능)한 타입이면 `==` 연산자로 비교할 수 있습니다.

```go
type Point struct{ X, Y int }

pt1 := Point{3, 4}
pt2 := Point{3, 4}
pt3 := Point{1, 2}

fmt.Println(pt1 == pt2)  // true
fmt.Println(pt1 == pt3)  // false

// comparable 구조체는 map의 키로도 사용 가능
distances := map[Point]string{
    {0, 0}: "원점",
    {1, 0}: "오른쪽",
}
```

**슬라이스/맵을 포함한 구조체는 == 불가**

```go
type Container struct {
    Items []string  // 슬라이스는 comparable이 아님
}

c1 := Container{Items: []string{"a"}}
c2 := Container{Items: []string{"a"}}
// c1 == c2  // 컴파일 에러!

// 대신 reflect.DeepEqual 사용
import "reflect"
fmt.Println(reflect.DeepEqual(c1, c2))  // true
```

---

## 8. 구조체와 메모리 레이아웃

Go는 구조체 필드 사이에 패딩(padding)을 삽입해 각 필드를 자신의 크기에 맞게 정렬합니다.

```go
// 비효율적 정렬 — 내부적으로 패딩 발생
type Bad struct {
    a bool    // 1바이트 + 7바이트 패딩
    b float64 // 8바이트
    c bool    // 1바이트 + 7바이트 패딩
}
// 총 24바이트

// 효율적 정렬 — 큰 타입 먼저
type Good struct {
    b float64 // 8바이트
    a bool    // 1바이트
    c bool    // 1바이트 + 6바이트 패딩
}
// 총 16바이트
```

실무에서는 대부분 신경 쓸 필요가 없지만, 메모리에 민감한 대량 데이터 구조에서는 고려합니다.

---

## 9. Python / Java 비교

### Python dataclass vs Go struct

```python
# Python
from dataclasses import dataclass

@dataclass
class Person:
    name: str
    age: int
    city: str = ""  # 기본값

p = Person(name="홍길동", age=30)
```

```go
// Go
type Person struct {
    Name string
    Age  int
    City string  // 기본값 없음 — 제로값("")으로 초기화됨
}

p := Person{Name: "홍길동", Age: 30}
```

**핵심 차이**
- Go 구조체에는 필드 기본값이 없다. 대신 `NewXxx()` 생성자 함수에서 기본값을 설정한다.
- Python의 `__init__`, Java의 생성자 없이 구조체 리터럴로 직접 초기화한다.
- Go의 `omitempty` JSON 태그는 Python의 `field(default=None)` + 직렬화 커스터마이징을 한 번에 처리한다.

### Java record vs Go struct

```java
// Java
record Person(String name, int age, String city) {}
var p = new Person("홍길동", 30, "서울");
```

```go
// Go
type Person struct {
    Name string
    Age  int
    City string
}
p := Person{Name: "홍길동", Age: 30, City: "서울"}
```

Java `record`는 자동으로 `equals`, `hashCode`, `toString`을 생성합니다. Go에서는 직접 구현하거나(`==` 비교, `String() string` 메서드), 라이브러리를 사용합니다.
