# 04-testing-advanced: 고급 테스트 패턴

testify/mock을 사용한 목 기반 테스트, 빌드 태그 통합 테스트, 퍼즈 테스트를 학습합니다.

## 개요

| 테스트 종류 | 목적 | 속도 | 격리 |
|-------------|------|------|------|
| 단위 테스트 + 목 | 비즈니스 로직 검증 | 빠름 | 완전 격리 |
| 통합 테스트 | 실제 의존성 연동 검증 | 느림 | 실제 시스템 사용 |
| 퍼즈 테스트 | 엣지 케이스/버그 자동 발견 | 매우 느림 | 격리 가능 |
| 벤치마크 | 성능 측정 | 느림 | 격리 필요 |

## 프로젝트 구조

```
04-testing-advanced/
├── service/
│   ├── service.go           # 서비스 (인터페이스 기반 의존성)
│   ├── service_test.go      # 목 기반 단위 테스트
│   └── integration_test.go  # 통합 테스트 (//go:build integration)
├── fuzz/
│   ├── fuzz.go              # 파서 함수들
│   └── fuzz_test.go         # 퍼즈 테스트
└── go.mod
```

## 실행 방법

```bash
cd 04-testing-advanced
go mod tidy

# 단위 테스트 (목 기반)
go test ./service/ -v

# 통합 테스트 (빌드 태그 필요)
go test ./service/ -tags=integration -v

# 퍼즈 테스트 - 시드 코퍼스 실행 (빠름)
go test ./fuzz/ -v

# 퍼즈 테스트 - 실제 퍼징 (무작위 입력 생성)
go test ./fuzz/ -fuzz=FuzzParseKV -fuzztime=30s
go test ./fuzz/ -fuzz=FuzzParseAndCalc -fuzztime=30s
go test ./fuzz/ -fuzz=FuzzParseCSVLine -fuzztime=30s
```

## 목(Mock) 기반 테스트

### testify/mock 기본 패턴

```go
// 1. Mock 구조체 정의
type MockUserRepository struct {
    mock.Mock
}

// 2. 인터페이스 메서드 구현
func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*User), args.Error(1)
}

// 3. 테스트에서 기대 동작 설정
mockRepo.On("GetByID", mock.Anything, 1).Return(user, nil)

// 4. 기대값 검증
mockRepo.AssertExpectations(t)
```

### 유용한 매처

```go
mock.Anything              // 어떤 값이든 매칭
mock.MatchedBy(func(v T) bool { ... })  // 조건부 매칭
mock.AnythingOfType("int") // 특정 타입 매칭
```

## 빌드 태그 (Build Tags)

```go
//go:build integration   // 이 파일은 -tags=integration 시에만 포함
```

```bash
# 통합 테스트 포함 실행
go test ./... -tags=integration

# 통합 테스트 제외 (기본)
go test ./...
```

CI/CD 활용:
```yaml
# GitHub Actions 예시
- name: 단위 테스트
  run: go test ./...

- name: 통합 테스트
  run: go test ./... -tags=integration
  env:
    TEST_DB_URL: postgres://localhost/testdb
```

## 퍼즈 테스트 (Go 1.18+)

### 기본 구조

```go
func FuzzParseKV(f *testing.F) {
    // 시드 코퍼스: 퍼저의 시작점
    f.Add("key=value")
    f.Add("")

    f.Fuzz(func(t *testing.T, input string) {
        // 패닉이 발생하면 안 됨
        result, err := ParseKV(input)

        // 불변 조건 검증
        if err == nil {
            // result가 유효한지 확인
        }
    })
}
```

### 발견된 버그 재현

퍼저가 버그를 발견하면 `testdata/fuzz/FuzzXxx/` 디렉터리에 저장됩니다.

```bash
# 발견된 케이스 재현
go test ./fuzz/ -run=FuzzParseKV/testdata/fuzz/FuzzParseKV/<해시>
```

### 퍼즈 테스트가 잘 찾는 버그

- 패닉 (nil 포인터, 범위 초과 등)
- 0으로 나누기
- 정수 오버플로우
- 무한 루프
- 인코딩/디코딩 round-trip 불일치

## 학습 포인트

1. 의존성을 **인터페이스로 추상화**해야 목 교체가 가능하다
2. `mock.Anything`으로 컨텍스트처럼 매번 바뀌는 값을 처리
3. `AssertExpectations(t)`로 기대한 호출이 실제로 일어났는지 검증
4. 빌드 태그로 단위/통합 테스트를 분리하면 CI 속도 향상
5. 퍼즈 테스트는 **불변 조건**을 먼저 정의하는 것이 핵심
6. `f.Add()`로 의미 있는 시드를 제공하면 더 빠르게 버그 발견

## 테스트 컨테이너 (참고)

실제 DB가 필요한 통합 테스트에는 `testcontainers-go`를 사용합니다:

```go
import "github.com/testcontainers/testcontainers-go/modules/postgres"

func TestMain(m *testing.M) {
    ctx := context.Background()
    pgContainer, _ := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16"),
        postgres.WithDatabase("testdb"),
    )
    defer pgContainer.Terminate(ctx)
    // ...
}
```
