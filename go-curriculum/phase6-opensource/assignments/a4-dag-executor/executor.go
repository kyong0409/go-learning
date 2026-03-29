// executor.go
// DAG 기반 태스크 실행기
//
// TODO: Executor를 완성하세요.
// 의존성이 없는 노드는 병렬로 실행하고,
// 모든 의존 노드가 완료된 노드도 즉시 실행합니다.
package main

import (
	"context"
	"time"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// ExecutionResult는 단일 노드의 실행 결과입니다.
type ExecutionResult struct {
	NodeID   string
	Err      error
	Duration time.Duration
}

// ExecuteOptions는 실행 옵션입니다.
type ExecuteOptions struct {
	// DryRun이 true면 실제 Task를 실행하지 않고 실행 순서만 기록합니다.
	DryRun bool
	// ContinueOnError가 true면 한 노드가 실패해도 독립적인 노드는 계속 실행합니다.
	ContinueOnError bool
	// OnProgress는 각 노드 완료 시 호출되는 콜백입니다 (nil이면 무시).
	OnProgress func(result ExecutionResult)
}

// Executor는 DAG를 받아 태스크를 실행합니다.
type Executor struct {
	// TODO: 필요한 필드를 추가하세요 (없어도 됩니다).
}

// NewExecutor는 새 Executor를 생성합니다.
func NewExecutor() *Executor {
	// TODO: 구현하세요
	panic("NewExecutor: 아직 구현되지 않았습니다")
}

// Execute는 DAG의 태스크를 의존성 순서에 따라 실행합니다.
//
// 실행 규칙:
//  1. 먼저 dag.Validate()로 사이클 없음을 확인합니다.
//  2. 의존성이 없는 노드(in-degree=0)는 즉시 병렬로 실행합니다.
//  3. 어떤 노드의 모든 의존 노드가 완료되면 그 노드를 즉시 실행합니다.
//  4. opts.DryRun=true면 Task를 실행하지 않고 NodeID만 기록합니다.
//  5. opts.ContinueOnError=false면 첫 에러 발생 시 나머지를 취소합니다.
//  6. ctx 취소 시 실행 중인 모든 goroutine에 취소를 전파합니다.
//
// 반환값: 모든 노드의 ExecutionResult 목록 (실행 순서대로)
//
// 힌트:
//   - 각 노드에 대한 "완료 카운터"를 추적하세요 (완료된 의존 노드 수)
//   - sync.WaitGroup으로 모든 goroutine 완료를 기다립니다
//   - 완료 채널(done channel)로 노드 완료를 알립니다
//   - context.WithCancel로 에러 시 전체 취소를 구현합니다
func (e *Executor) Execute(ctx context.Context, dag *DAG, opts ExecuteOptions) ([]ExecutionResult, error) {
	// TODO: 구현하세요
	panic("Execute: 아직 구현되지 않았습니다")
}

// runNode는 단일 노드를 실행하고 결과를 반환합니다.
// DryRun 모드면 Task를 실행하지 않습니다.
func (e *Executor) runNode(ctx context.Context, node *Node, dryRun bool) ExecutionResult {
	// TODO: 구현하세요
	panic("runNode: 아직 구현되지 않았습니다")
}
