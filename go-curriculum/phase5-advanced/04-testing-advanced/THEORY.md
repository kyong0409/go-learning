# 고급 테스트 전략 - 목킹, 통합 테스트, 퍼즈 테스트

## 테스트 피라미드

```
        /\
       /  \
      / E2E \       ← 적게: 느리고 불안정, 하지만 실제 환경 검증
     /──────\
    / 통합 테스트 \   ← 적당히: 실제 DB/서비스 사용
   /────────────\
  /   단위 테스트   \ ← 많이: 빠르고 안정적, 목(mock) 사용
 /────────────────\

단위 테스트: 밀리초 단위, 수천 개 가능
통합 테스트: 초 단위, 수백 개 가능
E2E 테스트: 분 단위, 수십 개 가능
```

---

## 인터페이스 기반 모킹

Go에서 모킹의 핵심은 **의존성 역전 원칙(DIP)**입니다.
구체 타입 대신 인터페이스에 의존하면 테스트에서 가짜 구현을 주입할 수 있습니다.

### 의존성 역전 원칙 적용

```go
// 나쁜 패턴: 구체 타입에 직접 의존
type BookmarkService struct {
    db *sql.DB // 구체 타입 - 테스트에서 실제 DB 필요
}

// 좋은 패턴: 인터페이스에 의존
type BookmarkRepository interface {
    Save(ctx context.Context, b Bookmark) error
    FindByID(ctx context.Context, id int64) (Bookmark, error)
    List(ctx context.Context, filter Filter) ([]Bookmark, error)
    Delete(ctx context.Context, id int64) error
}

type BookmarkService struct {
    repo BookmarkRepository // 인터페이스 - 테스트에서 목 주입 가능
}

func NewBookmarkService(repo BookmarkRepository) *BookmarkService {
    return &BookmarkService{repo: repo}
}
```

### 방법 1: 직접 목 작성 (간단한 경우)

```go
// 테스트 파일에서 직접 구현
type mockRepo struct {
    bookmarks map[int64]Bookmark
    saveErr   error
    findErr   error
}

func (m *mockRepo) Save(_ context.Context, b Bookmark) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    m.bookmarks[b.ID] = b
    return nil
}

func (m *mockRepo) FindByID(_ context.Context, id int64) (Bookmark, error) {
    if m.findErr != nil {
        return Bookmark{}, m.findErr
    }
    b, ok := m.bookmarks[id]
    if !ok {
        return Bookmark{}, ErrNotFound
    }
    return b, nil
}

// 테스트에서 사용
func TestCreateBookmark(t *testing.T) {
    repo := &mockRepo{bookmarks: make(map[int64]Bookmark)}
    svc := NewBookmarkService(repo)

    err := svc.Create(context.Background(), "https://golang.org", "Go")
    if err != nil {
        t.Fatalf("예상치 못한 에러: %v", err)
    }
}
```

### 방법 2: testify/mock

```go
import "github.com/stretchr/testify/mock"

// 목 구조체 정의
type MockBookmarkRepo struct {
    mock.Mock // 임베딩
}

func (m *MockBookmarkRepo) Save(ctx context.Context, b Bookmark) error {
    args := m.Called(ctx, b) // 호출 기록
    return args.Error(0)
}

func (m *MockBookmarkRepo) FindByID(ctx context.Context, id int64) (Bookmark, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(Bookmark), args.Error(1)
}

// 테스트에서 사용
func TestBookmarkService(t *testing.T) {
    repo := new(MockBookmarkRepo)

    expected := Bookmark{ID: 1, URL: "https://golang.org"}

    // 기대 동작 설정
    repo.On("FindByID", mock.Anything, int64(1)).Return(expected, nil)
    repo.On("Save", mock.Anything, mock.AnythingOfType("Bookmark")).Return(nil)

    svc := NewBookmarkService(repo)
    got, err := svc.GetBookmark(context.Background(), 1)

    assert.NoError(t, err)
    assert.Equal(t, expected, got)
    repo.AssertExpectations(t) // 기대한 메서드가 실제 호출되었는지 검증
}
```

### 방법 3: mockery로 자동 생성

```bash
# mockery 설치
go install github.com/vektra/mockery/v2@latest

# 인터페이스에서 목 자동 생성
mockery --name=BookmarkRepository --output=mocks --outpkg=mocks

# 생성된 파일: mocks/BookmarkRepository.go
```

---

## 통합 테스트

### 빌드 태그로 분리

```go
//go:build integration

// integration_test.go
package bookmark_test

import (
    "context"
    "database/sql"
    "testing"
    _ "github.com/lib/pq"
)

func TestBookmarkRepository_Integration(t *testing.T) {
    db, err := sql.Open("postgres", "postgres://localhost/testdb?sslmode=disable")
    if err != nil {
        t.Skipf("데이터베이스 연결 실패 (통합 테스트 건너뜀): %v", err)
    }
    defer db.Close()

    repo := NewPostgresRepository(db)
    // ... 실제 DB로 테스트
}
```

```bash
# 단위 테스트만 (기본)
go test ./...

# 통합 테스트 포함
go test -tags=integration ./...
```

### testcontainers-go: Docker 기반 실제 DB 테스트

```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestWithRealPostgres(t *testing.T) {
    ctx := context.Background()

    // Docker로 PostgreSQL 컨테이너 시작
    pgContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections"),
        ),
    )
    if err != nil {
        t.Fatalf("컨테이너 시작 실패: %v", err)
    }
    defer pgContainer.Terminate(ctx) // 테스트 후 컨테이너 삭제

    // 연결 문자열 가져오기
    connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")

    db, _ := sql.Open("postgres", connStr)
    repo := NewPostgresRepository(db)

    // 실제 PostgreSQL로 테스트
    err = repo.Save(ctx, Bookmark{URL: "https://golang.org"})
    assert.NoError(t, err)
}
```

---

## 퍼즈 테스트 (Fuzz Testing)

Go 1.18에서 도입된 퍼즈 테스트는 자동으로 다양한 입력값을 생성하여 버그를 찾습니다.
파서, URL 검증기, 직렬화 코드에 특히 유용합니다.

### 퍼즈 테스트 작성

```go
// fuzz_test.go
package bookmark

import (
    "testing"
    "unicode/utf8"
)

// FuzzXxx 형식의 이름
func FuzzParseURL(f *testing.F) {
    // 시드 코퍼스: 초기 입력값 (자동 변형의 기반)
    f.Add("https://golang.org")
    f.Add("http://example.com/path?query=1")
    f.Add("")
    f.Add("not-a-url")
    f.Add("https://")

    // 퍼저가 생성한 입력으로 반복 호출
    f.Fuzz(func(t *testing.T, url string) {
        // 크래시하지 않아야 함 (패닉 없이 에러 반환)
        result, err := ParseURL(url)
        if err != nil {
            return // 에러는 OK
        }

        // 불변 조건(invariant) 검증
        if !utf8.ValidString(result.Host) {
            t.Errorf("Host가 유효한 UTF-8이 아님: %q", result.Host)
        }
        if result.Scheme != "http" && result.Scheme != "https" {
            t.Errorf("예상치 못한 스킴: %q", result.Scheme)
        }
    })
}

// 왕복 속성 테스트 (marshal → unmarshal → 동일해야 함)
func FuzzMarshalRoundtrip(f *testing.F) {
    f.Add("https://golang.org", "Go 공식 사이트")
    f.Add("", "")

    f.Fuzz(func(t *testing.T, url, title string) {
        original := Bookmark{URL: url, Title: title}
        data, err := json.Marshal(original)
        if err != nil {
            return
        }
        var restored Bookmark
        if err := json.Unmarshal(data, &restored); err != nil {
            t.Errorf("역직렬화 실패: %v", err)
        }
        if original.URL != restored.URL || original.Title != restored.Title {
            t.Errorf("왕복 후 데이터 불일치")
        }
    })
}
```

### 퍼즈 테스트 실행

```bash
# 단위 테스트 모드 (시드 코퍼스만 테스트, CI에서 사용)
go test ./...

# 퍼징 모드 (새 입력 자동 생성, 버그 찾기)
go test -fuzz=FuzzParseURL

# 시간 제한
go test -fuzz=FuzzParseURL -fuzztime=60s

# 크래시 재현: 발견된 케이스는 자동으로 저장됨
# testdata/fuzz/FuzzParseURL/abc123def456
go test -run=FuzzParseURL/abc123def456 # 저장된 케이스로 재현
```

---

## testing/synctest (Go 1.25 안정화)

동시성 코드를 결정론적으로 테스트하기 위한 패키지입니다. Go 1.24에서 실험적(`GOEXPERIMENT=synctest`)으로 도입되었고, **Go 1.25(2025년 8월)에 정식 안정화**되었습니다.

`testing/synctest`의 핵심 아이디어: `Run` 블록 내부의 고루틴들은 **가짜 시계**를 사용합니다. `time.Sleep`이나 타이머가 실제 시간을 소비하지 않고, `synctest.Wait()`를 호출하면 모든 고루틴이 대기 상태가 될 때까지 가짜 시계를 앞으로 진행시킵니다.

```go
import "testing/synctest"

func TestCacheExpiry(t *testing.T) {
    synctest.Run(func() {
        cache := NewCache(5 * time.Minute)
        cache.Set("key", "value")

        // 가짜 시계를 5분 뒤로 이동 (실제로 5분을 기다리지 않음)
        time.Sleep(5 * time.Minute)
        synctest.Wait() // 모든 고루틴이 대기 상태가 될 때까지 기다림

        // 실제로 5분이 지나지 않았지만 캐시는 만료됨
        _, ok := cache.Get("key")
        if ok {
            t.Error("캐시가 만료되어야 함")
        }
    })
}

// 타이머 기반 재시도 로직 테스트
func TestRetryWithBackoff(t *testing.T) {
    synctest.Run(func() {
        attempts := 0
        go func() {
            for i := 0; i < 3; i++ {
                attempts++
                time.Sleep(time.Duration(i+1) * time.Second) // 지수 백오프
            }
        }()

        synctest.Wait() // 모든 Sleep 완료까지 가짜 시계 진행
        if attempts != 3 {
            t.Errorf("재시도 횟수: got %d, want 3", attempts)
        }
    })
}
```

**Go 1.24 대비 Go 1.25 변경점**: Go 1.24에서는 `GOEXPERIMENT=synctest` 환경 변수 설정이 필요했으나, Go 1.25부터는 일반 import만으로 사용 가능합니다.

---

## sync.WaitGroup.Go() — Go 1.25

Go 1.25에서 `sync.WaitGroup`에 `Go()` 메서드가 추가되었습니다. `Add(1)` + `go func()` + `defer wg.Done()` 패턴을 하나의 호출로 줄여줍니다.

```go
// Go 1.24 이하 — 번거로운 패턴
var wg sync.WaitGroup
for _, url := range urls {
    wg.Add(1)
    go func(u string) {
        defer wg.Done()
        fetch(u)
    }(url)
}
wg.Wait()

// Go 1.25+ — sync.WaitGroup.Go() 사용
var wg sync.WaitGroup
for _, url := range urls {
    wg.Go(func() {  // Add(1) + go + Done() 자동 처리
        fetch(url)  // 클로저 캡처 버그도 없음 (Go 1.22+ 루프 변수 수정과 함께)
    })
}
wg.Wait()
```
```

---

## 고루틴 누수 감지: goleak

```go
import "go.uber.org/goleak"

func TestMain(m *testing.M) {
    // 테스트 종료 후 고루틴 누수 검사
    goleak.VerifyTestMain(m)
}

// 또는 개별 테스트에서
func TestMyFunction(t *testing.T) {
    defer goleak.VerifyNone(t) // 이 테스트가 고루틴을 누수하면 실패

    // 고루틴을 생성하는 코드 테스트
    svc := NewService()
    svc.Start()
    // svc.Stop()을 빠뜨리면 goleak이 감지
}
```

---

## 테스트 환경 헬퍼

```go
func TestFileProcessor(t *testing.T) {
    // 1. 임시 디렉토리: 테스트 후 자동 삭제
    tmpDir := t.TempDir()

    // 2. 환경 변수 임시 설정: 테스트 후 원래값 복원
    t.Setenv("DATABASE_URL", "postgres://localhost/testdb")

    // 3. 정리 함수 등록
    t.Cleanup(func() {
        // DB 연결 닫기, 외부 리소스 해제 등
    })

    // 4. 병렬 실행
    t.Parallel()

    // 5. 서브테스트
    t.Run("빈 파일", func(t *testing.T) {
        // ...
    })
    t.Run("대용량 파일", func(t *testing.T) {
        // ...
    })
}
```

### Golden File 패턴

```go
// 복잡한 출력을 파일로 저장하고 비교
func TestRenderTable(t *testing.T) {
    bookmarks := createTestBookmarks()
    got := renderTable(bookmarks)

    goldenFile := filepath.Join("testdata", "table_output.golden")

    if *update { // -update 플래그로 갱신
        os.WriteFile(goldenFile, []byte(got), 0644)
        return
    }

    want, err := os.ReadFile(goldenFile)
    if err != nil {
        t.Fatalf("골든 파일 읽기 실패: %v", err)
    }
    if got != string(want) {
        t.Errorf("출력이 골든 파일과 다름\ngot:\n%s\nwant:\n%s", got, want)
    }
}

// 플래그 등록
var update = flag.Bool("update", false, "골든 파일 업데이트")
```

---

## 테스트 전략 요약

| 전략 | 도구 | 속도 | 신뢰도 | 용도 |
|------|------|------|--------|------|
| 단위 테스트 | testify | 빠름 | 중간 | 비즈니스 로직 |
| 목킹 | testify/mock | 빠름 | 중간 | 외부 의존성 격리 |
| 통합 테스트 | testcontainers | 느림 | 높음 | DB/외부 서비스 |
| 퍼즈 테스트 | 내장 | 가변 | 높음 | 파서/검증기 |
| 고루틴 검사 | goleak | 빠름 | 높음 | 동시성 코드 |
| 벤치마크 | 내장 | 가변 | - | 성능 측정 |
