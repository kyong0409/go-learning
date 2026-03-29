// solution/executor.go
// DAG 실행기 참고 풀이
package main

import (
	"context"
	"sync"
	"time"
)

type ExecutionResult struct {
	NodeID   string
	Err      error
	Duration time.Duration
}

type ExecuteOptions struct {
	DryRun          bool
	ContinueOnError bool
	OnProgress      func(result ExecutionResult)
}

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(ctx context.Context, dag *DAG, opts ExecuteOptions) ([]ExecutionResult, error) {
	// 먼저 사이클 검사
	if err := dag.Validate(); err != nil {
		return nil, err
	}

	nodes := dag.Nodes()
	if len(nodes) == 0 {
		return nil, nil
	}

	// 취소 가능한 컨텍스트 생성
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 각 노드의 완료된 의존 수 추적
	depCount := make(map[string]int)      // 총 의존 수
	completedDeps := make(map[string]int) // 완료된 의존 수
	var mu sync.Mutex

	for _, n := range nodes {
		depCount[n.ID] = len(dag.Dependencies(n.ID))
		completedDeps[n.ID] = 0
	}

	// 결과 수집
	var results []ExecutionResult
	var resultsMu sync.Mutex

	// 완료 알림 채널
	doneCh := make(chan ExecutionResult, len(nodes))

	var wg sync.WaitGroup
	var firstErr error
	var errOnce sync.Once

	// 노드 실행 함수
	var runNode func(node *Node)
	runNode = func(node *Node) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result := e.runNode(execCtx, node, opts.DryRun)

			if result.Err != nil && result.Err != context.Canceled {
				errOnce.Do(func() {
					firstErr = result.Err
				})
				if !opts.ContinueOnError {
					cancel() // 나머지 모두 취소
				}
			}

			if opts.OnProgress != nil {
				opts.OnProgress(result)
			}

			resultsMu.Lock()
			results = append(results, result)
			resultsMu.Unlock()

			doneCh <- result

			// 이 노드가 완료되면 의존하는 노드들의 카운터 감소
			mu.Lock()
			dependents := dag.Dependents(node.ID)
			readyNodes := []*Node{}
			for _, depID := range dependents {
				completedDeps[depID]++
				if completedDeps[depID] == depCount[depID] {
					readyNodes = append(readyNodes, dag.nodes[depID])
				}
			}
			mu.Unlock()

			// 준비된 노드 실행
			for _, ready := range readyNodes {
				select {
				case <-execCtx.Done():
					return
				default:
					runNode(ready)
				}
			}
		}()
	}

	// in-degree=0인 노드 (의존성 없는 노드)부터 시작
	for _, n := range nodes {
		if depCount[n.ID] == 0 {
			runNode(n)
		}
	}

	// 모든 goroutine 완료 대기
	wg.Wait()
	close(doneCh)

	return results, firstErr
}

func (e *Executor) runNode(ctx context.Context, node *Node, dryRun bool) ExecutionResult {
	start := time.Now()
	result := ExecutionResult{NodeID: node.ID}

	if dryRun || node.Task == nil {
		result.Duration = time.Since(start)
		return result
	}

	err := node.Task(ctx)
	result.Err = err
	result.Duration = time.Since(start)
	return result
}
