// dag.go
// DAG (Directed Acyclic Graph) 구현 - Terraform 패턴
//
// TODO: 아래 타입과 함수들을 완성하세요.
package main

import (
	"context"
	"fmt"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// Node는 DAG의 단일 노드(태스크)입니다.
type Node struct {
	ID   string
	Task func(ctx context.Context) error
}

// DAG는 방향성 비순환 그래프입니다.
// AddEdge(from, to)는 "from이 완료돼야 to를 실행할 수 있다"는 의미입니다.
// 즉 to는 from에 의존합니다.
type DAG struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - nodes: map[string]*Node
	//   - edges: map[string][]string  (from → []to)
	//   - 또는 deps: map[string][]string (node → 이 노드가 의존하는 노드 목록)
}

// NewDAG는 새 DAG를 생성합니다.
func NewDAG() *DAG {
	// TODO: 구현하세요
	panic("NewDAG: 아직 구현되지 않았습니다")
}

// AddNode는 노드를 DAG에 추가합니다.
// 같은 ID가 이미 있으면 덮어씁니다.
func (d *DAG) AddNode(node *Node) {
	// TODO: 구현하세요
	panic("AddNode: 아직 구현되지 않았습니다")
}

// AddEdge는 from → to 방향의 엣지를 추가합니다.
// 의미: "from이 완료된 후에 to를 실행한다" (to는 from에 의존)
// from 또는 to가 존재하지 않으면 에러를 반환합니다.
func (d *DAG) AddEdge(fromID, toID string) error {
	// TODO: 구현하세요
	panic("AddEdge: 아직 구현되지 않았습니다")
}

// Nodes는 모든 노드를 반환합니다.
func (d *DAG) Nodes() []*Node {
	// TODO: 구현하세요
	panic("Nodes: 아직 구현되지 않았습니다")
}

// Dependencies는 nodeID가 의존하는 노드 ID 목록을 반환합니다.
// 즉, nodeID를 실행하기 전에 완료되어야 하는 노드들입니다.
func (d *DAG) Dependencies(nodeID string) []string {
	// TODO: 구현하세요
	panic("Dependencies: 아직 구현되지 않았습니다")
}

// Dependents는 nodeID가 완료되면 실행 가능해지는 노드 ID 목록을 반환합니다.
func (d *DAG) Dependents(nodeID string) []string {
	// TODO: 구현하세요
	panic("Dependents: 아직 구현되지 않았습니다")
}

// Validate는 DAG에 사이클이 없는지 검증합니다.
// 사이클이 있으면 에러를 반환합니다.
//
// 힌트: DFS(깊이 우선 탐색) + inStack 집합으로 사이클 감지
//
//	visited: 방문한 노드
//	inStack: 현재 DFS 경로에 있는 노드
//	어떤 노드의 이웃을 방문하는데 그 이웃이 inStack에 있으면 사이클
func (d *DAG) Validate() error {
	// TODO: 구현하세요
	panic("Validate: 아직 구현되지 않았습니다")
}

// TopologicalSort는 의존성 순서로 정렬된 노드 목록을 반환합니다.
// 의존 노드가 항상 의존하는 노드보다 뒤에 옵니다.
//
// 힌트: Kahn's algorithm
//
//	1. 각 노드의 in-degree(들어오는 엣지 수) 계산
//	2. in-degree=0인 노드를 큐에 추가
//	3. 큐에서 꺼내 결과에 추가, 이웃의 in-degree를 1 감소
//	4. in-degree=0이 된 이웃을 큐에 추가
//	5. 결과 노드 수 != 전체 노드 수 → 사이클 존재
func (d *DAG) TopologicalSort() ([]*Node, error) {
	// TODO: 구현하세요
	panic("TopologicalSort: 아직 구현되지 않았습니다")
}

// ============================================================
// 에러 타입
// ============================================================

// CycleError는 DAG에 사이클이 발견됐을 때 반환하는 에러입니다.
type CycleError struct {
	Path []string // 사이클을 형성하는 노드 경로
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("사이클 발견: %v", e.Path)
}

func main() {
	// 테스트를 실행하세요: go test ./... -v
}
