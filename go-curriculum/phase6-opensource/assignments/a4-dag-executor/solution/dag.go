// solution/dag.go
// DAG 참고 풀이
package main

import (
	"context"
	"fmt"
	"sort"
)

// ============================================================
// 타입 정의
// ============================================================

type Node struct {
	ID   string
	Task func(ctx context.Context) error
}

type CycleError struct {
	Path []string
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("사이클 발견: %v", e.Path)
}

// DAG는 방향성 비순환 그래프입니다.
// edges[from] = []to: from이 완료되면 to를 실행할 수 있음
// deps[to] = []from: to를 실행하려면 from이 먼저 완료되어야 함
type DAG struct {
	nodes map[string]*Node
	edges map[string][]string // from → []to (dependents)
	deps  map[string][]string // to → []from (dependencies)
}

func NewDAG() *DAG {
	return &DAG{
		nodes: make(map[string]*Node),
		edges: make(map[string][]string),
		deps:  make(map[string][]string),
	}
}

func (d *DAG) AddNode(node *Node) {
	d.nodes[node.ID] = node
	if _, ok := d.edges[node.ID]; !ok {
		d.edges[node.ID] = nil
	}
	if _, ok := d.deps[node.ID]; !ok {
		d.deps[node.ID] = nil
	}
}

func (d *DAG) AddEdge(fromID, toID string) error {
	if _, ok := d.nodes[fromID]; !ok {
		return fmt.Errorf("노드 %q가 존재하지 않습니다", fromID)
	}
	if _, ok := d.nodes[toID]; !ok {
		return fmt.Errorf("노드 %q가 존재하지 않습니다", toID)
	}
	d.edges[fromID] = append(d.edges[fromID], toID)
	d.deps[toID] = append(d.deps[toID], fromID)
	return nil
}

func (d *DAG) Nodes() []*Node {
	result := make([]*Node, 0, len(d.nodes))
	for _, n := range d.nodes {
		result = append(result, n)
	}
	return result
}

func (d *DAG) Dependencies(nodeID string) []string {
	deps := d.deps[nodeID]
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

func (d *DAG) Dependents(nodeID string) []string {
	deps := d.edges[nodeID]
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// Validate는 DFS로 사이클을 감지합니다.
func (d *DAG) Validate() error {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	path := make(map[string]string) // child → parent (경로 추적)

	var dfs func(id string) error
	dfs = func(id string) error {
		visited[id] = true
		inStack[id] = true

		for _, neighbor := range d.edges[id] {
			if !visited[neighbor] {
				path[neighbor] = id
				if err := dfs(neighbor); err != nil {
					return err
				}
			} else if inStack[neighbor] {
				// 사이클 발견 - 경로 재구성
				cyclePath := []string{neighbor, id}
				cur := id
				for cur != neighbor {
					cur = path[cur]
					cyclePath = append(cyclePath, cur)
				}
				return &CycleError{Path: cyclePath}
			}
		}

		inStack[id] = false
		return nil
	}

	// 모든 노드에 대해 DFS
	ids := make([]string, 0, len(d.nodes))
	for id := range d.nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		if !visited[id] {
			if err := dfs(id); err != nil {
				return err
			}
		}
	}
	return nil
}

// TopologicalSort는 Kahn's algorithm으로 위상 정렬합니다.
func (d *DAG) TopologicalSort() ([]*Node, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}

	// in-degree 계산
	inDegree := make(map[string]int, len(d.nodes))
	for id := range d.nodes {
		inDegree[id] = 0
	}
	for _, tos := range d.edges {
		for _, to := range tos {
			inDegree[to]++
		}
	}

	// in-degree=0인 노드를 초기 큐에 추가 (정렬해서 결정론적 순서 보장)
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var result []*Node
	for len(queue) > 0 {
		// 큐에서 첫 번째 꺼냄
		id := queue[0]
		queue = queue[1:]
		result = append(result, d.nodes[id])

		// 이웃의 in-degree 감소
		neighbors := d.edges[id]
		nextBatch := []string{}
		for _, neighbor := range neighbors {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				nextBatch = append(nextBatch, neighbor)
			}
		}
		sort.Strings(nextBatch)
		queue = append(queue, nextBatch...)
	}

	if len(result) != len(d.nodes) {
		return nil, fmt.Errorf("위상 정렬 실패: 사이클이 존재합니다")
	}
	return result, nil
}

func main() {}
