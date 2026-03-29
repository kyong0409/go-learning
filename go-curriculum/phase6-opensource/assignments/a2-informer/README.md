# 과제 A2: 인포머/리스터 패턴 구현

**난이도**: ★★★★★
**예상 소요 시간**: 7~10시간
**참고 패턴**: client-go Informer/Store/Reflector/Lister

## 배경

Kubernetes 컨트롤러는 API 서버를 직접 폴링하지 않습니다.
대신 **Informer**를 통해 로컬 메모리 캐시를 유지하고,
변경 이벤트를 비동기로 받습니다.

이 패턴은 모든 K8s 컨트롤러와 오퍼레이터의 근간입니다.

## 요구사항

### 구현할 컴포넌트

#### 1. Store (쓰레드 안전 인메모리 캐시)

```go
type Store interface {
    Add(obj Object) error
    Update(obj Object) error
    Delete(obj Object) error
    Get(name string) (Object, bool)
    List() []Object
    ListByIndex(indexName, indexValue string) ([]Object, error)
    AddIndexer(name string, indexFunc IndexFunc)
    Replace(objects []Object) error  // 전체 교체 (Resync)
}
```

#### 2. EventHandler (이벤트 콜백 인터페이스)

```go
type EventHandler interface {
    OnAdd(obj Object)
    OnUpdate(oldObj, newObj Object)
    OnDelete(obj Object)
}
```

#### 3. Reflector (소스 → 스토어 동기화)

```go
type Reflector struct {
    // 소스에서 오브젝트를 List하고 Watch하여 Store에 반영합니다.
    // ListFunc: 전체 목록을 가져오는 함수
    // WatchFunc: 변경 이벤트를 스트리밍하는 함수 (채널 반환)
}
```

#### 4. Informer (Reflector + Store + EventHandler)

```go
type Informer struct {
    // Reflector를 실행하고 Store 변경 시 EventHandler를 호출합니다.
    // AddEventHandler(handler EventHandler)
    // Run(ctx context.Context)
    // HasSynced() bool  // 초기 List가 완료되었는지
}
```

#### 5. Lister (로컬 캐시 읽기 전용 인터페이스)

```go
type Lister struct {
    // Store의 읽기 전용 뷰
    // List() []Object
    // Get(name string) (Object, bool)
    // ListByLabel(key, value string) []Object
}
```

### Object 타입

```go
type Object struct {
    Name      string
    Namespace string
    Labels    map[string]string
    Data      map[string]interface{}
}

// 키 생성 규칙: "namespace/name" (namespace가 비면 "name")
func KeyFunc(obj Object) string
```

### IndexFunc 타입

```go
// IndexFunc는 오브젝트에서 인덱스 값 목록을 추출합니다.
// 예: Labels 인덱서 → obj.Labels["app"] 값 반환
type IndexFunc func(obj Object) []string
```

## 채점 기준 (100점)

| 항목 | 점수 |
|------|------|
| Store CRUD (Add/Update/Delete/Get/List) | 20점 |
| Store 인덱싱 (AddIndexer/ListByIndex) | 15점 |
| Reflector → Store 동기화 | 15점 |
| Informer EventHandler 콜백 | 20점 |
| HasSynced (초기 동기화 완료 감지) | 10점 |
| Lister (캐시 읽기) | 10점 |
| 동시성 안전 (Race Detector 통과) | 10점 |

## 실행 방법

```bash
cd a2-informer
go mod tidy
go test ./... -v -race
go test -v -run TestGrade
```

## 참고 자료

- `k8s.io/client-go/tools/cache/store.go`
- `k8s.io/client-go/tools/cache/reflector.go`
- `k8s.io/client-go/tools/cache/shared_informer.go`
- `../01-k8s-patterns/README.md`
