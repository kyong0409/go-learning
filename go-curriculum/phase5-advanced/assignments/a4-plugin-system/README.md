# 과제 A4: 플러그인 시스템 구현

**난이도**: ★★★★☆ (4.5/5)
**예상 소요 시간**: 4~6시간

## 과제 설명

Go 인터페이스와 등록 패턴을 사용하여 플러그인 시스템을 구현합니다.
플러그인 생명주기 관리, 의존성 순서 초기화, 상태 모니터링을 포함합니다.

## 구현할 인터페이스 및 타입

### `Plugin` 인터페이스

```go
type Plugin interface {
    Name() string                      // 플러그인 고유 이름
    Version() string                   // 버전 문자열 (예: "1.0.0")
    Dependencies() []string            // 의존하는 플러그인 이름 목록
    Init(ctx context.Context, cfg map[string]any) error  // 초기화
    Execute(ctx context.Context, input any) (any, error) // 실행
    Shutdown(ctx context.Context) error // 종료
    Health() HealthStatus              // 현재 상태 반환
}
```

### `HealthStatus` — 플러그인 상태

```go
type HealthStatus struct {
    Healthy bool
    Message string
    // 추가 필드는 자유롭게 정의 가능
}
```

### `PluginManager` — 플러그인 관리자

```go
mgr := NewPluginManager()

// 등록 (이름 중복 시 error)
err := mgr.Register(myPlugin)

// 이름으로 조회
p, ok := mgr.Get("my-plugin")

// 등록된 플러그인 목록 (이름 정렬)
names := mgr.List()

// 의존성 순서로 모든 플러그인 초기화
// 의존성이 먼저 Init되어야 합니다.
err = mgr.InitAll(ctx, configs)

// 모든 플러그인 종료 (초기화 역순)
err = mgr.ShutdownAll(ctx)

// 이름으로 단일 플러그인 실행
result, err := mgr.Execute(ctx, "my-plugin", input)

// 모든 플러그인 상태 확인
statuses := mgr.HealthCheck()  // map[string]HealthStatus
```

### 설정 맵 타입

```go
// InitAll에 전달되는 설정: 플러그인 이름 -> 설정 맵
type PluginConfigs map[string]map[string]any
```

## 구현 요구사항

### 의존성 순서 초기화 (위상 정렬)
- `InitAll` 호출 시 의존성 그래프를 분석하여 올바른 순서로 `Init`을 호출합니다.
- 순환 의존성이 있으면 `error`를 반환합니다.
- 의존하는 플러그인이 등록되지 않았으면 `error`를 반환합니다.

### 역순 종료
- `ShutdownAll`은 `InitAll`의 역순으로 `Shutdown`을 호출합니다.
- 일부 플러그인의 `Shutdown`이 실패해도 나머지는 계속 진행합니다.
- 모든 오류를 수집하여 반환합니다.

### 상태 모니터링
- `HealthCheck`는 등록된 모든 플러그인의 `Health()`를 호출합니다.
- 초기화되지 않은 플러그인은 `Healthy: false`로 보고합니다.

### 고루틴 안전성
- `PluginManager`의 모든 메서드는 동시 호출에 안전해야 합니다.

## 채점 기준

| 항목 | 배점 |
|------|------|
| Plugin 인터페이스 및 Register/Get/List | 20점 |
| 의존성 순서 위상 정렬 | 30점 |
| 생명주기 (InitAll / ShutdownAll) | 25점 |
| 상태 모니터링 (HealthCheck) | 15점 |
| 동시성 안전성 | 10점 |
| **합계** | **100점** |

## 실행 방법

```bash
cd assignments/a4-plugin-system
go test ./... -v
go test -race ./...
go test ./... -v -run TestGrade
```

## 힌트

- 위상 정렬은 Kahn 알고리즘(진입 차수 0인 노드 먼저)을 사용하세요.
- 순환 의존성 탐지: 위상 정렬 후 처리된 노드 수가 전체보다 적으면 순환 존재.
- `InitAll`이 완료된 순서를 `[]string` 슬라이스에 저장해두면 역순 종료가 쉽습니다.
- `Execute`는 플러그인이 초기화되지 않았을 때 적절한 오류를 반환하세요.
