# 07-testing: 테스팅 (Testing)

> Go의 테스팅은 프레임워크 없이 표준 라이브러리만으로 완성됩니다. JUnit 어노테이션이나 pytest 픽스처 없이도 강력한 테스트를 작성할 수 있습니다.

---

## 1. Go의 테스팅 철학

- **표준 라이브러리 우선**: `testing` 패키지 하나로 유닛 테스트, 벤치마크, 예제 문서화가 모두 가능합니다
- **외부 assert 라이브러리 최소화**: `if got != want { t.Errorf(...) }` 패턴이 Go의 관례입니다
- **테이블 주도 테스트**: 테스트 케이스를 데이터로 표현하는 패턴이 Go 커뮤니티 표준입니다
- **테스트는 코드와 함께**: 테스트 파일은 같은 디렉터리에 위치합니다

---

## 2. go test 명령어

```bash
# 현재 패키지 테스트 실행
go test ./...

# 출력 상세 모드
go test -v ./...

# 특정 테스트만 실행 (정규식 사용)
go test -run TestAdd ./...
go test -run TestCalculator/divide ./...  # 서브테스트

# 테스트 반복 실행 (캐시 무효화)
go test -count=1 ./...

# 코드 커버리지
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out         # HTML 리포트

# 레이스 컨디션 감지
go test -race ./...

# 벤치마크 실행 (-bench는 정규식)
go test -bench=. ./...
go test -bench=BenchmarkAdd -benchmem ./...

# 타임아웃 설정
go test -timeout 30s ./...
```

---

## 3. 테스트 파일 규칙

```
calc/
  calc.go        ← 구현 코드
  calc_test.go   ← 테스트 코드 (_test.go 접미사 필수)
```

테스트 파일의 패키지 선언 두 가지 방식:

```go
// 방식 1: 같은 패키지 — 비공개(소문자) 함수/변수에 접근 가능
package calc

// 방식 2: 외부 패키지 — 공개 API만 테스트 (블랙박스 테스트)
package calc_test
```

실무에서는 두 방식을 함께 쓰기도 합니다. 비공개 함수 테스트가 필요하면 `package calc`, 공개 API만 테스트하면 `package calc_test`를 사용합니다.

---

## 4. 테스트 함수 작성

```go
// 테스트 함수: func Test로 시작, *testing.T 파라미터 필수
func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5.0

    if got != want {
        t.Errorf("Add(2, 3) = %v, want %v", got, want)
    }
}
```

### t.Error vs t.Fatal vs t.Skip

| 메서드 | 동작 |
|--------|------|
| `t.Error(args...)` | 테스트 실패 표시, **계속 실행** |
| `t.Errorf(format, args...)` | 포맷된 메시지로 실패 표시, 계속 실행 |
| `t.Fatal(args...)` | 테스트 실패 표시, **즉시 중단** |
| `t.Fatalf(format, args...)` | 포맷된 메시지로 실패, 즉시 중단 |
| `t.Skip(args...)` | 테스트 건너뜀 (조건부 skip) |
| `t.Log(args...)` | `-v` 플래그 시 출력 |

```go
func TestDivide(t *testing.T) {
    result, err := Divide(10, 0)

    // 에러가 반드시 있어야 하면 Fatal (이후 코드가 의미 없음)
    if err == nil {
        t.Fatal("0으로 나누기에서 에러가 반환되어야 합니다")
    }

    // result는 계속 확인 가능
    if result != 0 {
        t.Errorf("에러 시 result는 0이어야 합니다, got %v", result)
    }
}
```

---

## 5. 테이블 주도 테스트 (Table-Driven Tests)

Go의 가장 중요한 테스트 패턴입니다. 테스트 케이스를 구조체 슬라이스로 정의합니다.

```go
func TestAdd(t *testing.T) {
    // 테스트 케이스를 데이터로 표현
    tests := []struct {
        name string  // 테스트 이름 (t.Run에 사용)
        a, b float64
        want float64
    }{
        {"양수 더하기", 2, 3, 5},
        {"음수 더하기", -1, -2, -3},
        {"제로 더하기", 0, 5, 5},
        {"소수 더하기", 1.5, 2.5, 4.0},
    }

    for _, tt := range tests {
        // t.Run으로 각 케이스를 서브테스트로 실행
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

**t.Run()의 장점**

```bash
# 전체 실행
go test -v -run TestAdd

# 특정 서브테스트만 실행
go test -v -run "TestAdd/음수_더하기"  # 공백은 _로 대체
```

**왜 이 패턴이 Go 표준인가**

- 새 테스트 케이스 추가가 데이터 한 줄 추가로 끝남
- 실패 시 어떤 케이스가 실패했는지 이름으로 명확히 알 수 있음
- 중복 코드 없이 다양한 케이스 커버 가능
- 병렬 실행이 용이 (`t.Parallel()` 추가)

---

## 6. 에러를 반환하는 함수 테스트

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool        // 에러 기대 여부
    }{
        {"정상 나눗셈", 10, 2, 5, false},
        {"0으로 나누기", 10, 0, 0, true},
        {"음수 나눗셈", -6, 2, -3, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)

            // 에러 발생 여부 확인
            if (err != nil) != tt.wantErr {
                t.Errorf("Divide() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // 에러가 없을 때만 결과값 확인
            if !tt.wantErr && got != tt.want {
                t.Errorf("Divide() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## 7. 테스트 헬퍼: t.Helper()

`t.Helper()`를 호출한 함수는 실패 시 그 함수 내부가 아닌 **호출한 위치**를 에러 줄로 표시합니다.

```go
// t.Helper() 없이
func assertEqual(t *testing.T, got, want float64) {
    if got != want {
        t.Errorf("got %v, want %v", got, want)  // 이 줄이 에러 위치로 표시됨
    }
}

// t.Helper() 사용
func assertEqual(t *testing.T, got, want float64) {
    t.Helper()  // 이 함수를 호출한 위치가 에러 위치로 표시됨
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}

func TestSomething(t *testing.T) {
    assertEqual(t, Add(1, 2), 3)  // ← t.Helper() 사용 시 이 줄이 에러 위치
}
```

---

## 8. Example 함수: 문서화 + 테스트

`Example` 함수는 문서화와 테스트를 동시에 합니다. `go doc`에서 예제로 표시되고, `go test`로 실행되어 출력이 검증됩니다.

```go
func ExampleAdd() {
    result := Add(2, 3)
    fmt.Println(result)
    // Output:
    // 5
}

func ExampleDivide() {
    result, err := Divide(10, 2)
    fmt.Println(result, err)
    // Output:
    // 5 <nil>
}

func ExampleDivide_byZero() {  // _byZero: 같은 함수의 두 번째 예제
    _, err := Divide(10, 0)
    fmt.Println(err)
    // Output:
    // 0으로 나눌 수 없습니다
}
```

`// Output:` 주석이 없으면 컴파일은 되지만 출력 검증은 하지 않습니다.

---

## 9. Benchmark: func BenchmarkXxx(b *testing.B)

```go
func BenchmarkAdd(b *testing.B) {
    // b.N은 go test가 자동으로 결정하는 반복 횟수
    for i := 0; i < b.N; i++ {
        Add(1.5, 2.5)
    }
}

func BenchmarkDivide(b *testing.B) {
    b.ResetTimer()  // 셋업 시간 제외
    for i := 0; i < b.N; i++ {
        Divide(float64(i+1), 2.0)
    }
}

// 병렬 벤치마크
func BenchmarkAddParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            Add(1.5, 2.5)
        }
    })
}
```

```bash
# 벤치마크 실행 결과 예시
go test -bench=BenchmarkAdd -benchmem

# BenchmarkAdd-8    1000000000    0.3 ns/op    0 B/op    0 allocs/op
# 이름-CPU코어수    반복횟수      1회당 시간    1회당 메모리  1회당 할당수
```

### testing.B.Loop() — Go 1.24 현대적 대안

Go 1.24(Feb 2025)에서 `b.Loop()`가 추가되었습니다. `for i := 0; i < b.N; i++` 패턴의 현대적 대체입니다. 루프 본문이 컴파일러에 의해 최적화(제거)되는 것을 방지하고, 첫/마지막 반복에서 셋업/티어다운 코드를 실행하기에 더 적합합니다.

```go
// 기존 방식 (여전히 유효)
func BenchmarkAddOld(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(1.5, 2.5)
    }
}

// Go 1.24+ 방식: b.Loop() 권장
func BenchmarkAdd(b *testing.B) {
    for b.Loop() {
        Add(1.5, 2.5)
    }
}
```

### testing.T.Context() — Go 1.24 신규

테스트가 종료될 때 자동으로 취소되는 컨텍스트를 반환합니다. `context.Background()` + `defer cancel()` 패턴을 대체합니다.

```go
func TestFetch(t *testing.T) {
    ctx := t.Context() // 테스트 종료 시 자동 취소

    result, err := fetchData(ctx, "https://example.com")
    if err != nil {
        t.Fatalf("fetchData 실패: %v", err)
    }
    _ = result
}
```

### testing/synctest — Go 1.25 동시성 테스트

Go 1.25(Aug 2025)에서 `testing/synctest` 패키지가 안정화되었습니다(API: `synctest.Test()`). 시간 기반 동시성 코드를 결정론적으로 테스트할 수 있게 해줍니다. `synctest.Test()` 내부에서는 `time.Sleep`, `time.After`, `time.Now` 등이 가상 시계를 사용하므로 실제로 기다리지 않아도 됩니다.

```go
import "testing/synctest"

func TestRateLimiter(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        rl := NewRateLimiter(1, time.Second)

        // 가상 시계 — 실제로 1초를 기다리지 않음
        go rl.Allow() // 즉시 허용
        time.Sleep(time.Second) // 가상 시간 1초 진행
        go rl.Allow() // 이제 다시 허용
    })
}
```

---

## 10. TestMain: 전체 테스트 셋업/티어다운

```go
func TestMain(m *testing.M) {
    // 모든 테스트 실행 전 셋업
    fmt.Println("테스트 셋업: DB 연결")
    setupTestDB()

    // 테스트 실행 (m.Run()의 반환값이 종료 코드)
    code := m.Run()

    // 모든 테스트 실행 후 티어다운
    fmt.Println("테스트 티어다운: DB 연결 해제")
    teardownTestDB()

    os.Exit(code)
}
```

`TestMain`은 패키지당 하나만 정의할 수 있습니다.

---

## 11. 코드 커버리지

```bash
# 커버리지 퍼센트 출력
go test -cover ./...

# 커버리지 프로파일 저장
go test -coverprofile=coverage.out ./...

# HTML 리포트로 어떤 줄이 커버됐는지 확인
go tool cover -html=coverage.out

# 함수별 커버리지 출력
go tool cover -func=coverage.out
```

커버리지 목표를 100%로 잡는 것은 비실용적입니다. 의미 있는 경계값, 에러 경로, 정상 경로를 커버하는 것이 중요합니다.

---

## 12. 테스트 더블: 인터페이스를 이용한 모킹

인터페이스를 사용하면 테스트에서 실제 의존성을 mock으로 교체할 수 있습니다.

```go
// 실제 코드: 인터페이스 파라미터
type UserRepository interface {
    FindByID(id int) (*User, error)
    Save(u *User) error
}

type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(id int) (*User, error) {
    return s.repo.FindByID(id)
}

// 테스트 코드: mock 구현
type mockUserRepo struct {
    users map[int]*User
}

func (m *mockUserRepo) FindByID(id int) (*User, error) {
    u, ok := m.users[id]
    if !ok {
        return nil, fmt.Errorf("user %d not found", id)
    }
    return u, nil
}

func (m *mockUserRepo) Save(u *User) error { return nil }

func TestGetUser(t *testing.T) {
    repo := &mockUserRepo{
        users: map[int]*User{
            1: {ID: 1, Name: "홍길동"},
        },
    }
    svc := &UserService{repo: repo}

    user, err := svc.GetUser(1)
    if err != nil {
        t.Fatalf("예상치 못한 에러: %v", err)
    }
    if user.Name != "홍길동" {
        t.Errorf("이름 = %q, want %q", user.Name, "홍길동")
    }
}
```

외부 mock 라이브러리(`testify/mock`, `gomock`)도 있지만, 인터페이스가 작으면 직접 구현하는 것이 더 단순합니다.

---

## 13. Python / Java 비교

### pytest vs Go testing

```python
# pytest
import pytest

def test_add():
    assert add(2, 3) == 5

# 파라미터화 테스트 (테이블 주도 테스트와 유사)
@pytest.mark.parametrize("a,b,expected", [
    (2, 3, 5),
    (-1, -2, -3),
    (0, 5, 5),
])
def test_add_parametrized(a, b, expected):
    assert add(a, b) == expected
```

```go
// Go
func TestAdd(t *testing.T) {
    tests := []struct{ a, b, want float64 }{
        {2, 3, 5},
        {-1, -2, -3},
        {0, 5, 5},
    }
    for _, tt := range tests {
        t.Run(fmt.Sprintf("%v+%v", tt.a, tt.b), func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### pytest fixture vs Go TestMain/서브테스트 셋업

```python
# pytest fixture
@pytest.fixture
def db_connection():
    conn = setup_db()
    yield conn          # 테스트에 제공
    conn.close()        # 테스트 후 정리

def test_query(db_connection):
    result = db_connection.query("SELECT 1")
    assert result == 1
```

```go
// Go: TestMain (패키지 수준)
var testDB *sql.DB

func TestMain(m *testing.M) {
    testDB = setupDB()
    code := m.Run()
    testDB.Close()
    os.Exit(code)
}

func TestQuery(t *testing.T) {
    result := testDB.QueryRow("SELECT 1")
    // ...
}
```

### JUnit vs Go testing

```java
// JUnit 5
@Test
@DisplayName("Add two positive numbers")
void testAdd() {
    assertEquals(5.0, calculator.add(2, 3));
}

@ParameterizedTest
@CsvSource({"2,3,5", "-1,-2,-3", "0,5,5"})
void testAddParametrized(double a, double b, double expected) {
    assertEquals(expected, calculator.add(a, b));
}
```

```go
// Go: 어노테이션 없음, 순수 코드
func TestAdd(t *testing.T) {
    tests := []struct{ a, b, want float64 }{
        {2, 3, 5}, {-1, -2, -3}, {0, 5, 5},
    }
    for _, tt := range tests {
        t.Run(fmt.Sprintf("%.0f+%.0f", tt.a, tt.b), func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.want {
                t.Errorf("Add(%v,%v) = %v; want %v", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

**핵심 차이**
- Go는 어노테이션/데코레이터 없이 순수 코드로 테스트를 구성합니다
- `assert` 라이브러리 없이 `if` 문으로 직접 비교합니다 (간결하지만 장황해 보일 수 있음)
- 테이블 주도 테스트는 pytest `@parametrize`, JUnit `@ParameterizedTest`와 개념이 같지만 외부 의존성이 없습니다
- `t.Parallel()`로 서브테스트를 병렬 실행할 수 있습니다 (pytest의 `pytest-xdist`와 유사)
