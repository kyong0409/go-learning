# 05. Terraform 핵심 Go 패턴 - 이론 심화

## Terraform 아키텍처: Core + Providers + State

```
┌─────────────────────────────────────────────────────────────┐
│                   Terraform 아키텍처                          │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Terraform Core                          │   │
│  │                                                      │   │
│  │  HCL 파서 → 설정 로더 → 그래프 빌더 → 플래너          │   │
│  │                              ↓                       │   │
│  │                        DAG 워커 (병렬 실행기)         │   │
│  │                              ↓                       │   │
│  │                        상태 관리자                    │   │
│  └──────────────────────────┬──────────────────────────┘   │
│                             │ go-plugin (gRPC)              │
│         ┌───────────────────┼───────────────────┐          │
│         ▼                   ▼                   ▼          │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐      │
│  │  Provider   │   │  Provider   │   │  Provider   │      │
│  │    AWS      │   │    GCP      │   │  Kubernetes │      │
│  │  (별도 프로세스)│  │  (별도 프로세스)│  │  (별도 프로세스)│    │
│  └──────┬──────┘   └──────┬──────┘   └──────┬──────┘      │
│         │                 │                 │               │
│         ▼                 ▼                 ▼               │
│      AWS API           GCP API           K8s API            │
└─────────────────────────────────────────────────────────────┘

State:
  terraform.tfstate (로컬)
  또는 S3, GCS, etcd (원격 백엔드)
```

세 구성요소의 역할:
- **Core**: HCL 파싱, 그래프 구성, 실행 계획, 상태 동기화
- **Provider**: 특정 플랫폼(AWS, GCP 등)의 API를 Terraform 인터페이스로 래핑
- **State**: 인프라의 현재 상태를 JSON으로 저장 (진실의 원천)

---

## DAG (Directed Acyclic Graph) 심층 분석

### 왜 DAG인가

인프라 리소스는 의존 관계를 가집니다:

```hcl
resource "aws_vpc" "main" { ... }

resource "aws_subnet" "public" {
  vpc_id = aws_vpc.main.id  # vpc에 의존
}

resource "aws_instance" "web" {
  subnet_id = aws_subnet.public.id  # subnet에 의존
  security_groups = [aws_security_group.web.id]  # sg에도 의존
}

resource "aws_security_group" "web" {
  vpc_id = aws_vpc.main.id  # vpc에 의존
}
```

이 의존성을 그래프로 표현:
```
aws_vpc
  ├── aws_subnet ──→ aws_instance
  └── aws_security_group ──→ aws_instance
```

DAG로 결정되는 것:
1. 실행 순서 (위상 정렬)
2. 병렬 실행 가능 여부 (같은 레벨)
3. 사이클 감지 (순환 의존 오류)

### 그래프 자료구조

```go
// hashicorp/terraform/internal/dag/graph.go (단순화)

type Vertex interface{}  // 모든 타입이 정점이 될 수 있음

type Edge interface {
    Source() Vertex
    Target() Vertex
}

type BasicEdge struct {
    source Vertex
    target Vertex
}

type Graph struct {
    vertices *Set  // 정점 집합
    edges    *Set  // 엣지 집합
    // downEdges: 정점 → 이 정점이 의존하는 정점들
    // upEdges:   정점 → 이 정점에 의존하는 정점들
    downEdges map[interface{}]*Set
    upEdges   map[interface{}]*Set
    mu        sync.RWMutex
}

func (g *Graph) Add(v Vertex) Vertex {
    g.mu.Lock()
    defer g.mu.Unlock()
    g.vertices.Add(v)
    return v
}

// Connect: from이 to에 의존함을 표현
// "aws_instance는 aws_subnet이 필요하다"
// → g.Connect(BasicEdge{aws_instance, aws_subnet})
func (g *Graph) Connect(edge Edge) {
    g.mu.Lock()
    defer g.mu.Unlock()
    source := edge.Source()
    target := edge.Target()

    // source의 다운스트림에 target 추가
    if _, ok := g.downEdges[vertexID(source)]; !ok {
        g.downEdges[vertexID(source)] = new(Set)
    }
    g.downEdges[vertexID(source)].Add(target)

    // target의 업스트림에 source 추가
    if _, ok := g.upEdges[vertexID(target)]; !ok {
        g.upEdges[vertexID(target)] = new(Set)
    }
    g.upEdges[vertexID(target)].Add(source)

    g.edges.Add(edge)
}
```

### 위상 정렬 (Topological Sort)

의존성 순서를 지키는 실행 순서를 결정합니다.

```go
// Kahn's Algorithm: 진입 차수(in-degree) 기반
func (g *Graph) TopologicalSort() ([]Vertex, error) {
    // 1. 모든 정점의 진입 차수 계산
    inDegree := make(map[interface{}]int)
    for _, v := range g.Vertices() {
        if _, ok := inDegree[vertexID(v)]; !ok {
            inDegree[vertexID(v)] = 0
        }
        // 이 정점이 의존하는 정점들의 진입 차수 증가
        for _, dep := range g.DownEdges(v) {
            inDegree[vertexID(dep)]++
        }
    }

    // 2. 진입 차수 0인 정점(아무도 의존 안 함)을 큐에 추가
    queue := []Vertex{}
    for _, v := range g.Vertices() {
        if inDegree[vertexID(v)] == 0 {
            queue = append(queue, v)
        }
    }

    // 3. BFS로 위상 정렬
    var sorted []Vertex
    for len(queue) > 0 {
        v := queue[0]
        queue = queue[1:]
        sorted = append(sorted, v)

        // 이 정점에 의존하는 정점들의 진입 차수 감소
        for _, dependent := range g.UpEdges(v) {
            inDegree[vertexID(dependent)]--
            if inDegree[vertexID(dependent)] == 0 {
                queue = append(queue, dependent)
            }
        }
    }

    // 4. 사이클 감지: 모든 정점이 처리되지 않으면 사이클 존재
    if len(sorted) != g.Vertices().Len() {
        return nil, fmt.Errorf("graph has cycles")
    }

    return sorted, nil
}
```

### 사이클 감지: DFS 기반

```go
// DFS 기반 사이클 감지 (Terraform에서 실제 사용)
// hashicorp/terraform/internal/dag/tarjan.go

type CycleError struct {
    Cycles [][]Vertex
}

func (g *Graph) Cycles() [][]Vertex {
    var cycles [][]Vertex

    // White: 미방문, Gray: 방문 중(스택), Black: 완료
    const (
        WHITE = 0
        GRAY  = 1
        BLACK = 2
    )

    color := make(map[interface{}]int)
    var path []Vertex

    var dfs func(v Vertex)
    dfs = func(v Vertex) {
        color[vertexID(v)] = GRAY
        path = append(path, v)

        for _, neighbor := range g.DownEdges(v) {
            switch color[vertexID(neighbor)] {
            case WHITE:
                dfs(neighbor)
            case GRAY:
                // GRAY 정점에 도달 = 사이클 발견
                // path에서 사이클 부분 추출
                cycleStart := -1
                for i, pv := range path {
                    if vertexID(pv) == vertexID(neighbor) {
                        cycleStart = i
                        break
                    }
                }
                if cycleStart >= 0 {
                    cycle := make([]Vertex, len(path)-cycleStart)
                    copy(cycle, path[cycleStart:])
                    cycles = append(cycles, cycle)
                }
            }
        }

        path = path[:len(path)-1]
        color[vertexID(v)] = BLACK
    }

    for _, v := range g.Vertices() {
        if color[vertexID(v)] == WHITE {
            dfs(v)
        }
    }

    return cycles
}
```

### 병렬 실행: Walker 패턴

```go
// hashicorp/terraform/internal/dag/walk.go (단순화)
// 의존성이 없는 정점은 동시에 실행

type Walker struct {
    Callback   func(Vertex) error  // 각 정점에서 실행할 작업
    Concurrency int                // 최대 동시 실행 수
}

func (w *Walker) Walk(g *Graph) error {
    // 세마포어로 동시 실행 수 제한
    sem := make(chan struct{}, w.Concurrency)

    // 각 정점의 완료 채널
    done := make(map[interface{}]<-chan struct{})
    var mu sync.Mutex
    var wg sync.WaitGroup
    var firstErr error

    // 위상 정렬 순서로 처리
    for _, v := range topologicalOrder(g) {
        v := v  // goroutine 캡처

        // 이 정점의 모든 의존성(선행 정점)이 완료될 때까지 대기
        deps := g.DownEdges(v)
        depChans := make([]<-chan struct{}, 0, len(deps))
        mu.Lock()
        for _, dep := range deps {
            if ch, ok := done[vertexID(dep)]; ok {
                depChans = append(depChans, ch)
            }
        }
        mu.Unlock()

        // 이 정점의 완료 채널 등록
        doneCh := make(chan struct{})
        mu.Lock()
        done[vertexID(v)] = doneCh
        mu.Unlock()

        wg.Add(1)
        go func() {
            defer wg.Done()
            defer close(doneCh)

            // 모든 의존성 완료 대기
            for _, dep := range depChans {
                <-dep
            }

            // 세마포어 획득 (동시 실행 수 제한)
            sem <- struct{}{}
            defer func() { <-sem }()

            // 실제 작업 실행
            if err := w.Callback(v); err != nil {
                mu.Lock()
                if firstErr == nil {
                    firstErr = err
                }
                mu.Unlock()
            }
        }()
    }

    wg.Wait()
    return firstErr
}
```

실행 시각화:
```
aws_vpc (t=0 시작)
    │ 완료 (t=2)
    ├── aws_subnet 시작 (t=2)          ← 동시 실행
    └── aws_security_group 시작 (t=2)  ← 동시 실행
              │ 완료 (t=5)              │ 완료 (t=4)
              └──────────┬─────────────┘
                    aws_instance 시작 (t=5)
                         │ 완료 (t=8)
                    총 소요 시간: 8초
                    (순차 실행: 2+3+2+3=10초 대비 20% 단축)
```

---

## Plan/Apply 2단계 실행

### 왜 2단계인가

인프라 변경은 취소 불가능한 경우가 많습니다. 먼저 변경 계획을 확인하고 승인 후 실행하는 패턴은 안전성을 보장합니다.

```
현재 상태 (state.json):
  aws_instance.web: {id: "i-1234", instance_type: "t2.micro"}

원하는 상태 (main.tf):
  resource "aws_instance" "web" {
    instance_type = "t3.small"  # 변경됨
  }

Plan 단계:
  diff 계산:
    ~ aws_instance.web (수정)
        instance_type: "t2.micro" → "t3.small"  # 재시작 필요!

  Plan 출력:
    Plan: 0 to add, 1 to change, 0 to destroy.
    (사용자가 검토 후 yes/no 결정)

Apply 단계:
  Plan 실행:
    1. Provider에 UpdateResource 요청
    2. AWS API: 인스턴스 타입 변경 (재시작 수반)
    3. 새 상태를 state.json에 저장
```

### Plan 자료구조

```go
// hashicorp/terraform/internal/plans/changes.go

// 변경 행동의 종류
type Action string

const (
    NoOp    Action = "no-op"    // 변경 없음
    Create  Action = "create"   // 새로 생성
    Read    Action = "read"     // 데이터 소스 읽기
    Update  Action = "update"   // 수정
    Replace Action = "replace"  // 삭제 후 재생성 (ForceNew 속성 변경 시)
    Delete  Action = "delete"   // 삭제
)

// 단일 리소스의 변경 계획
type ResourceInstanceChange struct {
    Addr          addrs.AbsResourceInstance  // 리소스 주소
    PrevRunAddr   addrs.AbsResourceInstance  // 이전 주소 (이동 시)
    ProviderAddr  addrs.AbsProviderConfig    // 사용할 Provider
    Change        Change                     // 변경 내용
}

type Change struct {
    Action Action       // 무엇을 할 것인가
    Before DynamicValue // 현재 값 (삭제/수정 시)
    After  DynamicValue // 원하는 값 (생성/수정 시)
    // AfterSensitivePaths: 민감 정보 마스킹 경로
}

// Plan 전체
type Plan struct {
    Changes     *Changes      // 모든 리소스 변경 목록
    Variables   InputValues   // 입력 변수
    State       *states.State // Apply 시점의 상태 스냅샷
    UIMode      UIMode        // 실행 모드 (normal, destroy, refresh-only)
}
```

### Plan/Apply 구현 패턴

```go
// 2단계 실행 패턴 (A4 과제의 핵심 아이디어)

type Planner struct {
    stateStore StateStore    // 현재 상태 저장소
    providers  ProviderMap   // 사용 가능한 Provider
}

// Plan: 변경이 필요한 것만 계산 (실제 변경 없음)
func (p *Planner) Plan(ctx context.Context, desired Config) (*Plan, error) {
    // 1. 현재 상태 로드
    current, err := p.stateStore.Load()
    if err != nil {
        return nil, fmt.Errorf("load state: %w", err)
    }

    // 2. 의존성 그래프 구성
    graph, err := buildGraph(desired)
    if err != nil {
        return nil, fmt.Errorf("build graph: %w", err)
    }

    // 3. 각 리소스에 대해 변경 계획 수립
    var changes []ResourceChange
    for _, resource := range desired.Resources {
        currentState := current.Get(resource.Addr)

        var action Action
        switch {
        case currentState == nil:
            action = Create
        case resource.shouldDestroy:
            action = Delete
        case needsReplace(currentState, resource):
            action = Replace  // ForceNew 속성 변경
        case isDifferent(currentState, resource):
            action = Update
        default:
            action = NoOp
        }

        if action != NoOp {
            changes = append(changes, ResourceChange{
                Addr:   resource.Addr,
                Action: action,
                Before: currentState,
                After:  resource.Config,
            })
        }
    }

    return &Plan{
        Changes: changes,
        Graph:   graph,
    }, nil
}

// Apply: Plan을 실제로 실행
func (p *Planner) Apply(ctx context.Context, plan *Plan) error {
    // Walker로 DAG 병렬 실행
    walker := &Walker{
        Callback: func(v Vertex) error {
            change := plan.GetChange(v)
            if change == nil {
                return nil
            }

            provider := p.providers.Get(v.ProviderType())

            switch change.Action {
            case Create:
                newState, err := provider.CreateResource(ctx, change.After)
                if err != nil {
                    return fmt.Errorf("create %s: %w", v.Addr(), err)
                }
                // 즉시 상태 저장 (Apply 중 크래시 대비)
                return p.stateStore.Set(v.Addr(), newState)

            case Update:
                newState, err := provider.UpdateResource(ctx, change.Before, change.After)
                if err != nil {
                    return fmt.Errorf("update %s: %w", v.Addr(), err)
                }
                return p.stateStore.Set(v.Addr(), newState)

            case Delete:
                if err := provider.DeleteResource(ctx, change.Before); err != nil {
                    return fmt.Errorf("delete %s: %w", v.Addr(), err)
                }
                return p.stateStore.Delete(v.Addr())
            }

            return nil
        },
    }

    return walker.Walk(plan.Graph)
}
```

---

## Provider 플러그인 모델

### go-plugin: HashiCorp의 플러그인 시스템

go-plugin은 Go 프로그램을 별도 프로세스로 실행하고 gRPC로 통신하는 라이브러리입니다.

```
왜 별도 프로세스인가?

1. 격리: Provider 크래시가 Terraform Core에 영향 없음
2. 독립 릴리즈: Core와 Provider를 별도로 배포 가능
3. 다중 언어: gRPC이므로 Go 외 언어로 Provider 작성 가능
4. 보안: Provider가 Core 메모리에 접근 불가

통신 방식:
  Terraform Core
      │ os/exec로 Provider 바이너리 실행
      │ 환경 변수로 핸드셰이크 정보 전달
      ▼
  terraform-provider-aws (별도 프로세스)
      │ 준비 완료 시 표준 출력에 포트 번호 출력
      ▼
  gRPC 연결 수립
      │
  Terraform Core ←──gRPC──→ Provider 프로세스
```

### Provider 인터페이스

```go
// hashicorp/terraform-plugin-framework/provider/provider.go

type Provider interface {
    // 메타데이터: Provider 이름, 버전
    Metadata(context.Context, MetadataRequest, *MetadataResponse)

    // 스키마: Provider 설정 스키마 (required 필드 등)
    Schema(context.Context, SchemaRequest, *SchemaResponse)

    // Configure: Provider 초기화 (API 클라이언트 생성)
    Configure(context.Context, ConfigureRequest, *ConfigureResponse)

    // Resources: 이 Provider가 관리하는 리소스 목록
    Resources(context.Context) []func() resource.Resource

    // DataSources: 데이터 소스 목록
    DataSources(context.Context) []func() datasource.DataSource
}

// 리소스 인터페이스
type Resource interface {
    // 메타데이터: 리소스 타입 이름
    Metadata(context.Context, resource.MetadataRequest, *resource.MetadataResponse)

    // 스키마: 리소스 속성 정의
    Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)

    // CRUD 작업
    Create(context.Context, resource.CreateRequest, *resource.CreateResponse)
    Read(context.Context, resource.ReadRequest, *resource.ReadResponse)
    Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse)
    Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse)
}
```

### 간단한 Provider 구현

```go
// "mycloud" Provider 구현 예시
type MycloudProvider struct {
    client *MycloudClient  // Configure에서 초기화
}

func (p *MycloudProvider) Configure(
    ctx context.Context,
    req provider.ConfigureRequest,
    resp *provider.ConfigureResponse,
) {
    var config struct {
        Endpoint types.String `tfsdk:"endpoint"`
        APIKey   types.String `tfsdk:"api_key"`
    }

    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // API 클라이언트 초기화
    p.client = NewMycloudClient(config.Endpoint.ValueString(), config.APIKey.ValueString())

    // 리소스와 데이터 소스에 클라이언트 전달
    resp.ResourceData = p.client
    resp.DataSourceData = p.client
}

// 서버 리소스 구현
type ServerResource struct {
    client *MycloudClient
}

func (r *ServerResource) Create(
    ctx context.Context,
    req resource.CreateRequest,
    resp *resource.CreateResponse,
) {
    var plan struct {
        Name   types.String `tfsdk:"name"`
        Size   types.String `tfsdk:"size"`
        ID     types.String `tfsdk:"id"`  // Computed: API가 할당
    }

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 실제 API 호출
    server, err := r.client.CreateServer(ctx, CreateServerInput{
        Name: plan.Name.ValueString(),
        Size: plan.Size.ValueString(),
    })
    if err != nil {
        resp.Diagnostics.AddError("Create Server Error", err.Error())
        return
    }

    // API가 할당한 ID를 상태에 저장
    plan.ID = types.StringValue(server.ID)
    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
```

---

## State 관리

### terraform.tfstate 구조

```json
{
  "version": 4,
  "terraform_version": "1.6.0",
  "serial": 42,
  "lineage": "abc123-...",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "type": "aws_instance",
      "name": "web",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [
        {
          "schema_version": 1,
          "attributes": {
            "id": "i-1234567890abcdef0",
            "instance_type": "t3.small",
            "ami": "ami-0abcdef1234567890",
            "tags": {"Name": "web-server"}
          }
        }
      ]
    }
  ]
}
```

Go 구조체:
```go
// hashicorp/terraform/internal/states/state.go

type State struct {
    Modules map[string]*Module  // 모듈별 상태
}

type Module struct {
    Resources   map[string]*Resource   // 리소스 상태
    OutputValues map[string]*OutputValue
}

type Resource struct {
    Addr      addrs.AbsResource
    Instances map[addrs.InstanceKey]*ResourceInstance
}

type ResourceInstance struct {
    Current  *ResourceInstanceObjectSrc  // 현재 상태
    Deposed  []*ResourceInstanceObjectSrc // 교체 대기 중인 이전 상태
}

type ResourceInstanceObjectSrc struct {
    AttrsJSON []byte  // JSON 직렬화된 속성
    Status    ObjectStatus  // Tainted, Ready 등
}
```

### State Locking: 동시 수정 방지

```go
// 원격 백엔드의 State Lock
// 여러 사람이 동시에 terraform apply 실행 시 충돌 방지

type Locker interface {
    // Lock: 잠금 획득 (이미 잠겨있으면 LockError)
    Lock(info *LockInfo) (string, error)

    // Unlock: 잠금 해제
    Unlock(id string) error
}

type LockInfo struct {
    ID        string    // 잠금 식별자 (UUID)
    Operation string    // "apply", "plan" 등
    Who       string    // 잠금 보유자 (username@hostname)
    Version   string    // Terraform 버전
    Created   time.Time // 잠금 생성 시각
    Path      string    // State 경로
}

// S3 백엔드 Lock 구현 (DynamoDB 사용)
func (s *S3Backend) Lock(info *LockInfo) (string, error) {
    info.ID = uuid.NewString()

    // DynamoDB에 조건부 쓰기 (이미 존재하면 실패)
    _, err := s.dynamo.PutItem(&dynamodb.PutItemInput{
        TableName: aws.String(s.lockTable),
        Item:      marshalLockInfo(info),
        // 이미 ID가 있으면 실패 (낙관적 잠금)
        ConditionExpression: aws.String("attribute_not_exists(LockID)"),
    })
    if err != nil {
        // 기존 잠금 정보 조회 후 LockError 반환
        return "", &LockError{Info: s.getLockInfo()}
    }
    return info.ID, nil
}
```

---

## HCL (HashiCorp Configuration Language)

### HCL 파싱 개념

HCL은 Go의 `encoding/json`처럼 구조체에 직접 디코딩할 수 있습니다.

```go
// HCL 설정 파일 (config.hcl)
/*
server {
  host = "localhost"
  port = 8080

  tls {
    cert = "/etc/ssl/cert.pem"
    key  = "/etc/ssl/key.pem"
  }
}

timeout = "30s"
*/

// Go 구조체 정의
type Config struct {
    Server  []ServerConfig `hcl:"server,block"`
    Timeout string         `hcl:"timeout,optional"`
}

type ServerConfig struct {
    Host string      `hcl:"host"`
    Port int         `hcl:"port"`
    TLS  []TLSConfig `hcl:"tls,block"`
}

type TLSConfig struct {
    Cert string `hcl:"cert"`
    Key  string `hcl:"key"`
}

// 파싱
func loadConfig(filename string) (*Config, error) {
    src, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    // HCL 파일 파싱 → AST
    file, diags := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
    if diags.HasErrors() {
        return nil, diags
    }

    // AST → Go 구조체
    var cfg Config
    diags = gohcl.DecodeBody(file.Body, nil, &cfg)
    if diags.HasErrors() {
        return nil, diags
    }

    return &cfg, nil
}
```

### HCL 표현식과 변수 참조

```go
// Terraform HCL은 변수 참조와 함수를 지원
// ${var.name}, ${aws_vpc.main.id} 같은 표현식

// EvalContext로 변수 값 제공
evalCtx := &hcl.EvalContext{
    Variables: map[string]cty.Value{
        "var": cty.ObjectVal(map[string]cty.Value{
            "region": cty.StringVal("us-east-1"),
            "prefix": cty.StringVal("myapp"),
        }),
        "aws_vpc": cty.ObjectVal(map[string]cty.Value{
            "main": cty.ObjectVal(map[string]cty.Value{
                "id": cty.StringVal("vpc-12345"),
            }),
        }),
    },
    Functions: map[string]function.Function{
        "format": stdlib.FormatFunc,
        "length": stdlib.LengthFunc,
    },
}

// 표현식 평가
expr, _ := hclsyntax.ParseExpression(
    []byte(`"${var.prefix}-server-${var.region}"`),
    "config.hcl", hcl.Pos{},
)
val, diags := expr.Value(evalCtx)
// val = "myapp-server-us-east-1"
```

---

## Terraform 패턴을 A4에 적용

A4 DAG 실행기 과제는 Terraform의 핵심 메커니즘을 구현합니다:

```
Terraform 실제 코드               A4 과제 구현
─────────────────────────────────────────────────────
dag.Graph                    →  Graph struct
graph.Add(vertex)            →  g.AddNode(id)
graph.Connect(edge)          →  g.AddEdge(from, to)
StronglyConnected()          →  detectCycle() bool
Walker.Walk()                →  Execute(ctx, fn)
goroutine + semaphore        →  병렬 실행 (WaitGroup + chan)
변경 계획 (Plan)              →  DryRun 모드
실제 실행 (Apply)             →  Execute 모드
```

이 과제를 완성하면 Terraform의 실행 엔진과 정확히 동일한 로직을 구현한 것입니다.
