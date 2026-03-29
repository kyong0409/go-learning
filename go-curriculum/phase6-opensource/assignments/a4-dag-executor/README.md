# 과제 A4: DAG 기반 태스크 실행기

**난이도**: ★★★★★
**예상 소요 시간**: 6~9시간
**참고 패턴**: Terraform DAG + Kubernetes scheduler dependency graph

## 배경

Terraform은 인프라 리소스의 의존성을 DAG(Directed Acyclic Graph)로 모델링하고,
의존성이 없는 리소스는 병렬로 생성/삭제합니다.
이 과제에서는 그 핵심 구조를 직접 구현합니다.

## 요구사항

### DAG (dag.go)

```go
// Node는 DAG의 단일 노드(태스크)입니다.
type Node struct {
    ID   string
    Task func(ctx context.Context) error  // 실행할 작업
}

// DAG는 방향성 비순환 그래프입니다.
type DAG struct {
    // AddNode(node *Node)
    // AddEdge(fromID, toID string) error  // from이 완료돼야 to 실행 가능
    // Validate() error                    // 사이클 감지
    // TopologicalSort() ([]*Node, error)  // 위상 정렬
    // Nodes() []*Node
    // Dependencies(id string) []string    // id가 의존하는 노드 목록
    // Dependents(id string) []string      // id에 의존하는 노드 목록
}
```

### Executor (executor.go)

```go
type ExecutionResult struct {
    NodeID   string
    Err      error
    Duration time.Duration
}

type ExecuteOptions struct {
    DryRun         bool   // true면 실제 실행 안 함, 순서만 출력
    ContinueOnError bool  // true면 에러 발생해도 독립적 노드는 계속 실행
    OnProgress     func(result ExecutionResult)  // 진행 상황 콜백
}

type Executor struct {
    // Execute(ctx, dag, opts) ([]ExecutionResult, error)
    // 의존성이 없는 노드는 병렬로 실행
    // 모든 의존 노드가 완료된 노드도 즉시 실행
}
```

### 실행 규칙

1. 노드 A가 노드 B에 의존하면(`AddEdge(B, A)`), B가 완료된 후에만 A를 실행
2. 의존성이 없는 노드(또는 모든 의존 노드가 완료된 노드)는 동시에 실행
3. `ContinueOnError=false`: 한 노드 실패 시 나머지 취소
4. `ContinueOnError=true`: 실패해도 독립적 노드는 계속 실행
5. `ctx` 취소 시 실행 중인 모든 노드에 취소 신호 전파
6. 사이클이 있으면 `Validate()` 에러 반환

### 예시

```
A ──→ C
B ──→ C
      C ──→ D

실행 순서:
  단계1: A, B (병렬 - 의존성 없음)
  단계2: C (A, B 완료 후)
  단계3: D (C 완료 후)
```

## 채점 기준 (100점)

| 항목 | 점수 |
|------|------|
| DAG 노드/엣지 추가 | 10점 |
| 사이클 감지 (Validate) | 15점 |
| 위상 정렬 (TopologicalSort) | 15점 |
| 순차 실행 (의존성 순서 보장) | 20점 |
| 병렬 실행 (독립 노드 동시 실행) | 20점 |
| DryRun 모드 | 10점 |
| 컨텍스트 취소 전파 | 10점 |

## 실행 방법

```bash
cd a4-dag-executor
go mod tidy
go test ./... -v
go test -v -run TestGrade
```

## 참고 자료

- `github.com/hashicorp/terraform/internal/dag/graph.go`
- `github.com/hashicorp/terraform/internal/dag/walk.go`
- `../05-terraform-patterns/README.md`
