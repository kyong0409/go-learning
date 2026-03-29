# 04. Docker/containerd 핵심 Go 패턴

## 개요

Docker(Moby)와 containerd는 컨테이너 런타임의 기준점입니다.
플러그인 아키텍처, 레이어 시스템, 클라이언트-서버 패턴을 학습합니다.

---

## 1. 컨테이너 라이프사이클 관리

### 개념

컨테이너는 상태 머신입니다: Created → Running → Paused → Stopped → Deleted

```
containerd 상태 전환:
  unknown → created → running → stopped → deleted
                  ↕
               paused
```

### 실제 코드 위치

```
containerd/containerd/containers/containers.go
  - type Container struct
    - ID          string
    - Labels      map[string]string
    - Image       string
    - Runtime     RuntimeInfo
    - Spec        *types.Any     // OCI 스펙

containerd/containerd/runtime/v2/manager.go
  - type TaskManager interface
    - Create(ctx, id, bundle, opts) (Task, error)
    - Get(ctx, id) (Task, error)
    - Tasks(ctx, all) ([]Task, error)
    - Delete(ctx, taskID, opts) (*Exit, error)

moby/moby/daemon/container.go
  - type Container struct
  - func (container *Container) IsRunning() bool
  - func (container *Container) IsPaused() bool
```

### 라이프사이클 패턴

```go
// containerd 클라이언트 사용 패턴
client, _ := containerd.New("/run/containerd/containerd.sock")

// 이미지 Pull
image, _ := client.Pull(ctx, "docker.io/library/redis:latest",
    containerd.WithPullUnpack)

// 컨테이너 생성
container, _ := client.NewContainer(ctx, "redis-1",
    containerd.WithImage(image),
    containerd.WithNewSnapshot("redis-1-snapshot", image),
    containerd.WithNewSpec(oci.WithImageConfig(image)))

// 태스크(프로세스) 시작
task, _ := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
task.Start(ctx)

// 종료 대기
exitCh, _ := task.Wait(ctx)
<-exitCh
task.Delete(ctx)
container.Delete(ctx, containerd.WithSnapshotCleanup)
```

---

## 2. 이미지 레이어 시스템 (Union Filesystem 개념)

### 개념

Docker 이미지는 읽기 전용 레이어의 스택입니다.
컨테이너 실행 시 쓰기 가능한 레이어가 맨 위에 추가됩니다 (Copy-on-Write).

```
[컨테이너 레이어]  ← 쓰기 가능 (임시)
[레이어 3: app]   ← 읽기 전용
[레이어 2: python]← 읽기 전용
[레이어 1: ubuntu]← 읽기 전용
```

### 실제 코드 위치

```
containerd/containerd/snapshots/snapshots.go
  - type Snapshotter interface
    - Stat(ctx, key) (Info, error)
    - Usage(ctx, key) (Usage, error)
    - Prepare(ctx, key, parent string, opts...) ([]mount.Mount, error)
    - View(ctx, key, parent string, opts...) ([]mount.Mount, error)
    - Commit(ctx, name, key string, opts...) error
    - Remove(ctx, key string) error

moby/moby/layer/layer.go
  - type Layer interface
  - type RWLayer interface
  - type Store interface
    - Register(io.Reader, ChainID) (Layer, error)
    - Get(ChainID) (Layer, error)
    - CreateRWLayer(id, parent ChainID, opts) (RWLayer, error)
```

### Content Store 패턴

```go
// 콘텐츠 주소 기반 저장소 (내용 해시 = 주소)
cs := client.ContentStore()

// 레이어 Blob 저장
writer, _ := cs.Writer(ctx,
    content.WithRef("layer-sha256:abc123"),
    content.WithDescriptor(desc))
io.Copy(writer, layerTar)
writer.Commit(ctx, size, digest)
```

---

## 3. 클라이언트-서버 아키텍처 패턴

### 개념

Docker CLI ↔ Docker Daemon(dockerd) ↔ containerd ↔ runc

각 경계는 명확한 인터페이스(gRPC 또는 REST)로 분리됩니다.

### 실제 코드 위치

```
moby/moby/client/client.go
  - type Client struct
    - httpclient *http.Client
    - host       string
    - proto      string   // "unix", "tcp", "npipe"
    - basePath   string
  - func NewClientWithOpts(ops ...Opt) (*Client, error)

moby/moby/api/server/router/container/container.go
  - type containerRouter struct
  - func (r *containerRouter) initRoutes()
  - func (r *containerRouter) postContainersCreate(...)

containerd/containerd/services/containers/service.go
  - gRPC 서비스 구현
```

### API 클라이언트 패턴

```go
// Docker 클라이언트 - Functional Options 패턴
cli, err := client.NewClientWithOpts(
    client.FromEnv,
    client.WithAPIVersionNegotiation(),
    client.WithTimeout(30*time.Second),
)

// 컨테이너 목록
containers, _ := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
for _, c := range containers {
    fmt.Printf("%s %s\n", c.ID[:12], c.Image)
}
```

---

## 4. 플러그인 시스템

### 개념

Docker 플러그인은 별도 프로세스로 실행되며, Unix 소켓이나 TCP로 통신합니다.
플러그인 인터페이스를 구현하면 볼륨, 네트워크, 로그 드라이버를 교체할 수 있습니다.

### 실제 코드 위치

```
moby/moby/volume/drivers/extpoint.go
  - type VolumeDriver interface
    - Create(name string, opts map[string]string) error
    - Remove(name string) error
    - Mount(name, id string) (string, error)
    - Unmount(name, id string) error
    - List() ([]*volumes.Volume, error)

moby/moby/pkg/plugins/plugins.go
  - type Plugin struct
  - type Manifest struct
    - Implements []string  // 구현하는 인터페이스 목록
  - func Get(name, imp string) (*Plugin, error)
  - func (p *Plugin) Client() *pluginclient.Client
```

### 플러그인 구현 패턴

```go
// 볼륨 플러그인 구현 예시
type myVolumePlugin struct{}

func (p *myVolumePlugin) Create(name string, opts map[string]string) error {
    // 볼륨 생성 로직
    return nil
}

func (p *myVolumePlugin) Mount(name, id string) (string, error) {
    return "/mnt/volumes/" + name, nil
}

// HTTP 서버로 플러그인 노출
http.Handle("/VolumeDriver.Create", handler(p.Create))
http.Handle("/VolumeDriver.Mount", handler(p.Mount))
http.ListenAndServe("unix:///run/docker/plugins/myplugin.sock", nil)
```

---

## 5. OCI (Open Container Initiative) 표준

모든 컨테이너 런타임이 따르는 표준 인터페이스입니다.

```
github.com/opencontainers/runtime-spec/specs-go/config.go
  - type Spec struct        // 컨테이너 실행 스펙
  - type Process struct     // 실행할 프로세스
  - type Root struct        // 루트 파일시스템
  - type Mount struct       // 마운트 포인트

github.com/opencontainers/image-spec/specs-go/v1/config.go
  - type Image struct       // 이미지 메타데이터
  - type RootFS struct      // 레이어 목록
```

---

## 학습 포인트 요약

| 패턴 | 핵심 아이디어 | 활용 분야 |
|------|---------------|-----------|
| 상태 머신 | 라이프사이클 전환 | A6 (Database 인스턴스) |
| Content-Addressable Store | 해시 = 주소 | 일반 설계 |
| Unix Socket API | 로컬 IPC | 플러그인 통신 |
| Interface 플러그인 | 교체 가능한 드라이버 | A2 (Informer) |
| Functional Options | 클라이언트 설정 | 모든 과제 |

## 다음 단계

Docker 패턴은 A6 (오퍼레이터) 과제에서 컨테이너 인스턴스 라이프사이클 시뮬레이션에 활용됩니다.
