# 05. Terraform 핵심 Go 패턴

## 개요

Terraform은 인프라를 코드로 관리하는 도구입니다.
DAG 기반 의존성 해결, Provider 플러그인 모델, Plan/Apply 2단계 실행이 핵심 패턴입니다.

---

## 1. DAG (Directed Acyclic Graph) 기반 의존성 해결

### 개념

인프라 리소스들은 의존 관계가 있습니다.
`aws_instance`는 `aws_subnet`이 먼저 생성되어야 합니다.
DAG로 의존성을 모델링하고, 위상 정렬로 실행 순서를 결정합니다.
의존성이 없는 리소스는 병렬로 실행합니다.

```
aws_vpc ←── aws_subnet ←── aws_instance
                        ←── aws_security_group

위상 정렬 후 실행:
  단계 1: aws_vpc (병렬)
  단계 2: aws_subnet, aws_security_group (병렬)
  단계 3: aws_instance (위 둘 완료 후)
```

### 실제 코드 위치

```
github.com/hashicorp/terraform/internal/dag/graph.go
  - type Graph struct
    - vertices Set
    - edges    Set
  - func (g *Graph) Add(v Vertex) Vertex
  - func (g *Graph) Connect(edge Edge)
  - func (g *Graph) Remove(v Vertex)
  - func (g *Graph) HasVertex(v Vertex) bool
  - func (g *Graph) Edges() []Edge

github.com/hashicorp/terraform/internal/dag/walk.go
  - type Walker struct
  - func (w *Walker) Update(g *AcyclicGraph)
  - func (w *Walker) Wait() tfdiags.Diagnostics

github.com/hashicorp/terraform/internal/dag/tarjan.go
  - func StronglyConnected(g *Graph) [][]Vertex  // SCC = 사이클 감지
```

### DAG 구현 패턴

```go
// 실제 Terraform 패턴
type Graph struct {
    vertices map[string]Vertex
    edges    map[string][]string  // vertex → 의존하는 vertex 목록
}

func (g *Graph) Add(v Vertex) {
    g.vertices[v.ID()] = v
}

func (g *Graph) Connect(from, to Vertex) {
    g.edges[from.ID()] = append(g.edges[from.ID()], to.ID())
}

// 위상 정렬 (Kahn's algorithm)
func (g *Graph) TopologicalSort() ([]Vertex, error) {
    inDegree := make(map[string]int)
    // ...
}
```

### 사이클 감지 (DFS)

```go
// DFS 기반 사이클 감지
func (g *Graph) detectCycle() bool {
    visited := make(map[string]bool)
    inStack := make(map[string]bool)

    var dfs func(id string) bool
    dfs = func(id string) bool {
        visited[id] = true
        inStack[id] = true
        for _, neighbor := range g.edges[id] {
            if !visited[neighbor] && dfs(neighbor) {
                return true
            } else if inStack[neighbor] {
                return true // 사이클 발견
            }
        }
        inStack[id] = false
        return false
    }
    // ...
}
```

---

## 2. Provider 플러그인 모델 (RPC 기반)

### 개념

Terraform Provider는 별도 프로세스로 실행되며, gRPC로 통신합니다.
각 Provider는 `terraform-provider-*` 바이너리로 배포됩니다.

```
terraform (core)
    ↓ gRPC
terraform-provider-aws   ← AWS API 호출
terraform-provider-gcp   ← GCP API 호출
terraform-provider-k8s   ← K8s API 호출
```

### 실제 코드 위치

```
github.com/hashicorp/terraform-plugin-go/tfprotov6/provider.go
  - type ProviderServer interface
    - GetProviderSchema(context.Context, *GetProviderSchemaRequest) (*GetProviderSchemaResponse, error)
    - ConfigureProvider(context.Context, *ConfigureProviderRequest) (*ConfigureProviderResponse, error)

github.com/hashicorp/terraform-plugin-framework/provider/provider.go
  - type Provider interface
    - Metadata(context.Context, MetadataRequest, *MetadataResponse)
    - Schema(context.Context, SchemaRequest, *SchemaResponse)
    - Configure(context.Context, ConfigureRequest, *ConfigureResponse)
    - Resources(context.Context) []func() resource.Resource

github.com/hashicorp/terraform/internal/plugin/grpc_provider.go
  - type GRPCProvider struct     // core → provider gRPC 클라이언트
```

### Provider 구현 패턴

```go
// Provider 구현 예시
type MyProvider struct{}

func (p *MyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "endpoint": schema.StringAttribute{Required: true},
        },
    }
}

func (p *MyProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewServerResource,
        NewDatabaseResource,
    }
}
```

---

## 3. Plan/Apply 2단계 실행

### 개념

Terraform은 실행 전에 변경 계획(Plan)을 생성하고, 사용자 확인 후 적용(Apply)합니다.

```
현재 상태 (state.json) ─┐
                         ├→ Plan → 변경 목록 (diff)
원하는 상태 (*.tf) ─────┘           ↓
                                사용자 확인
                                    ↓
                                  Apply → 실제 인프라 변경
                                    ↓
                              새 상태 저장 (state.json)
```

### 실제 코드 위치

```
github.com/hashicorp/terraform/internal/command/plan.go
  - type PlanCommand struct
  - func (c *PlanCommand) Run(args []string) int

github.com/hashicorp/terraform/internal/command/apply.go
  - type ApplyCommand struct
  - func (c *ApplyCommand) Run(args []string) int

github.com/hashicorp/terraform/internal/plans/plan.go
  - type Plan struct
    - Changes  *Changes
    - State    *states.State
    - PrevRunState *states.State

github.com/hashicorp/terraform/internal/plans/changes.go
  - type Changes struct
    - Resources []*ResourceInstanceChangeSrc
  - type Change struct
    - Action Action   // Create, Update, Delete, Replace, NoOp
    - Before DynamicValue
    - After  DynamicValue
```

### Plan/Apply 패턴

```go
// 2단계 실행 패턴
type Executor struct{}

func (e *Executor) Plan(ctx context.Context, desired State) (*Plan, error) {
    current := e.loadState()
    diff := computeDiff(current, desired)
    return &Plan{Actions: diff}, nil
}

func (e *Executor) Apply(ctx context.Context, plan *Plan) error {
    for _, action := range plan.Actions {
        if err := e.execute(ctx, action); err != nil {
            return err
        }
        e.saveState()
    }
    return nil
}
```

---

## 4. 상태 관리 패턴

### 개념

Terraform은 인프라의 현재 상태를 `terraform.tfstate` 파일에 저장합니다.
이 상태 파일이 "진실의 원천(source of truth)"입니다.

### 실제 코드 위치

```
github.com/hashicorp/terraform/internal/states/state.go
  - type State struct
    - Modules map[string]*Module
  - type Module struct
    - Resources map[string]*Resource
  - type Resource struct
    - Instances map[addrs.InstanceKey]*ResourceInstance

github.com/hashicorp/terraform/internal/states/statefile/file.go
  - type File struct
    - TerraformVersion *version.Version
    - Serial           uint64
    - Lineage          string
    - State            *states.State
  - func ReadStateFile(src io.Reader) (*File, error)
  - func WriteStateFile(stateFile *File, dst io.Writer) error
```

---

## 5. HCL 파싱 패턴

```
github.com/hashicorp/hcl/v2/hclsyntax/parser.go
  - func ParseConfig(src []byte, filename string, start hcl.Pos) (*hcl.File, hcl.Diagnostics)

github.com/hashicorp/hcl/v2/gohcl/schema.go
  - func DecodeBody(body hcl.Body, ctx *hcl.EvalContext, val interface{}) hcl.Diagnostics
```

```go
// HCL 파싱 예시
type Config struct {
    Server []ServerBlock `hcl:"server,block"`
}
type ServerBlock struct {
    Port int    `hcl:"port"`
    Host string `hcl:"host"`
}

var cfg Config
diags := gohcl.DecodeBody(f.Body, nil, &cfg)
```

---

## 학습 포인트 요약

| 패턴 | 핵심 아이디어 | 적용 과제 |
|------|---------------|-----------|
| DAG | 의존성 그래프, 위상 정렬 | A4 |
| 사이클 감지 | DFS, 강연결요소 | A4 |
| 병렬 실행 | 의존성 없는 노드 동시 실행 | A4 |
| Plan/Apply | 2단계 실행, dry-run | A4 |
| Provider 플러그인 | 인터페이스 + RPC | A6 (심화) |
| 상태 관리 | 직렬화/역직렬화 | A6 |

## 다음 단계

[A4 - DAG 실행기](../assignments/a4-dag-executor/) 과제를 진행하세요.
DAG와 병렬 실행을 직접 구현하는 가장 도전적인 과제 중 하나입니다.
