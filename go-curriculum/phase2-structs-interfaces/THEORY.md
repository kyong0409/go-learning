# Phase 2: 구조체와 인터페이스 - 학습 가이드

> 대상: Python/Java 경험이 있는 개발자. Go의 OOP 대안을 이해하고 실무에 적용하는 것이 목표입니다.

---

## Phase 2 학습 목표

이 Phase를 마치면 다음을 할 수 있습니다.

- 구조체로 도메인 모델을 정의하고 JSON 직렬화까지 처리한다
- 값 리시버/포인터 리시버를 상황에 맞게 선택한다
- 인터페이스로 느슨하게 결합된 코드를 작성한다
- 상속 없이 임베딩과 컴포지션으로 기능을 조합한다
- 에러를 래핑하고 `errors.Is` / `errors.As`로 분류한다
- `defer` / `panic` / `recover`의 실행 흐름을 이해한다
- 테이블 주도 테스트와 벤치마크를 작성한다

---

## Go에는 클래스가 없다 — 왜, 그리고 대안은 무엇인가

Java나 Python에서 객체 지향은 `class`가 중심입니다. Go 설계자들은 클래스를 의도적으로 뺐습니다.

**클래스를 뺀 이유**

- 깊은 상속 계층은 코드를 읽기 어렵게 만들고, 변경이 파급된다
- 다중 상속(Python) 또는 인터페이스 명시 구현(Java)은 타입 간 결합을 높인다
- Go는 단순성을 최우선 가치로 삼는다

**Go의 대안**

| 개념 | Go 방식 |
|------|---------|
| 데이터 묶기 | `struct` |
| 메서드 | 리시버 함수 `func (r T) Method()` |
| 다형성 | 인터페이스 (암묵적 충족) |
| 코드 재사용 | 구조체 임베딩 (컴포지션) |
| 생성자 | `NewXxx()` 관례 함수 |

---

## 상속 vs 컴포지션: Go의 선택과 그 이유

```
// Java 스타일 — 상속
class Animal { speak() {} }
class Dog extends Animal { @Override speak() {} }

// Go 스타일 — 임베딩(컴포지션)
type Animal struct { Name string }
func (a Animal) Speak() string { ... }

type Dog struct {
    Animal        // 임베딩: Animal의 모든 메서드가 Dog로 프로모션됨
    Breed string
}
func (d Dog) Speak() string { ... } // 선택적 오버라이드
```

Go는 임베딩을 통해 코드를 재사용하지만, `Dog`가 `Animal`의 하위 타입이 되지는 않습니다. 대신 인터페이스로 다형성을 표현합니다.

**컴포지션의 장점**

- 어떤 기능을 쓸지 명확하게 보인다
- 여러 타입을 조합해도 다이아몬드 문제가 없다
- 테스트 시 의존성을 인터페이스로 교체하기 쉽다

---

## "인터페이스를 받아들이고, 구조체를 반환하라"

Go의 핵심 설계 원칙입니다.

```go
// 좋은 예: 파라미터는 인터페이스, 반환은 구체 타입
func NewService(db Database, logger Logger) *UserService { ... }

// 피해야 할 예: 파라미터가 구체 타입 — 테스트와 교체가 어렵다
func NewService(db *PostgresDB, logger *ZapLogger) *UserService { ... }
```

**왜 이 원칙이 중요한가**

- 파라미터를 인터페이스로 받으면 호출자가 어떤 구현체든 넘길 수 있다 (테스트용 mock 포함)
- 반환 타입을 구체 타입(구조체 포인터)으로 하면 호출자가 모든 메서드에 접근할 수 있다
- 반환 타입을 인터페이스로 하면 호출자는 인터페이스에 없는 메서드를 쓸 수 없어 오히려 불편해진다

---

## Phase 2 학습 순서와 토픽 연결 관계

```
01-structs        데이터 모델 정의, JSON 태그
      ↓
02-methods        구조체에 행동 추가, 값/포인터 리시버, NewXxx 패턴
      ↓
03-interfaces     암묵적 충족, io.Reader/Writer, 타입 단언/스위치
      ↓
04-composition    임베딩으로 상속 대체, 인터페이스 합성
      ↓
05-error-handling 커스텀 에러, %w 래핑, errors.Is/As
      ↓
06-defer-panic    리소스 정리, panic/recover 패턴
      ↓
07-testing        테이블 주도 테스트, 벤치마크, 예제 함수
```

각 토픽은 이전 토픽을 기반으로 합니다. `03-interfaces`는 `02-methods`에서 배운 메서드 구현이 인터페이스 충족의 핵심이고, `04-composition`은 `03-interfaces`의 인터페이스 합성까지 확장됩니다.

---

## Python / Java 대응표

| 개념 | Python | Java | Go |
|------|--------|------|----|
| 데이터 클래스 | `@dataclass` / `class` | `record` / `class` | `struct` |
| 생성자 | `__init__` | 생성자 메서드 | `NewXxx()` 함수 |
| 인터페이스 | ABC / Protocol | `interface` | `interface` (암묵적) |
| 상속 | `class Dog(Animal)` | `extends` | 없음 (임베딩 사용) |
| 문자열 표현 | `__str__` | `toString()` | `String() string` |
| 예외 | `raise` / `try/except` | `throw` / `try/catch` | `error` 반환 |
| 리소스 정리 | `with` / `__exit__` | `try-with-resources` | `defer` |
| 런타임 오류 | 예외 | 예외 | `panic` / `recover` |

---

## 추천 학습 자료

- **Learn Go with Tests** (https://quii.gitbook.io/learn-go-with-tests) — TDD로 Go를 배우는 무료 온라인 책. 이 Phase의 모든 토픽을 커버합니다.
- **100 Go Mistakes and How to Avoid Them** (Teiva Harsanyi) — 실수 패턴 100가지. 인터페이스 남용, 에러 처리 안티패턴 챕터가 특히 유용합니다.
- **Exercism Go Track** (https://exercism.org/tracks/go) — 멘토 피드백이 있는 실습 문제. Phase 2 수준에서는 `two-fer`, `clock`, `bank-account` 문제를 추천합니다.
- **Effective Go** (https://go.dev/doc/effective_go) — Go 공식 관례 문서. 짧고 실용적입니다.
- **The Go Blog** (https://go.dev/blog) — 인터페이스, 에러, 컴포지션 관련 공식 아티클이 있습니다.
