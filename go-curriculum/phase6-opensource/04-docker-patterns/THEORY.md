# 04. Docker/containerd 핵심 Go 패턴 - 이론 심화

## Docker/Moby 전체 아키텍처

Docker는 단일 바이너리처럼 보이지만 내부적으로 여러 컴포넌트로 구성됩니다.

```
┌─────────────────────────────────────────────────────────────┐
│                     사용자 영역                               │
│                                                              │
│  docker run nginx         ← Docker CLI                       │
│       │                                                      │
│       │ REST API (Unix Socket /var/run/docker.sock)          │
│       ▼                                                      │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              dockerd (Docker Daemon)                  │   │
│  │  - API 서버                                          │   │
│  │  - 이미지 관리 (pull, push, build)                   │   │
│  │  - 네트워크 관리                                      │   │
│  │  - 볼륨 관리                                          │   │
│  └──────────────────────────┬───────────────────────────┘   │
│                             │ gRPC                           │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              containerd                               │   │
│  │  - 컨테이너 라이프사이클 관리                          │   │
│  │  - 이미지 저장 (Content Store)                        │   │
│  │  - 스냅샷 관리 (레이어 시스템)                        │   │
│  └──────────────────────────┬───────────────────────────┘   │
│                             │ Unix Socket                    │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         containerd-shim-runc-v2                       │   │
│  │  - 컨테이너 프로세스 감시                              │   │
│  │  - containerd와 runc 사이 중간자                       │   │
│  └──────────────────────────┬───────────────────────────┘   │
│                             │ exec                           │
│                             ▼                                │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                  runc                                 │   │
│  │  - OCI 스펙 구현                                      │   │
│  │  - Linux namespace, cgroup 설정                       │   │
│  │  - 실제 프로세스 시작                                  │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘

계층 분리의 이유:
  - dockerd 없이 containerd만으로 Kubernetes가 직접 사용 가능
  - runc를 다른 OCI 호환 런타임으로 교체 가능 (gVisor, Kata)
  - 각 계층을 독립적으로 업그레이드 가능
```

---

## 컨테이너 핵심 기술

컨테이너는 가상머신과 달리 호스트 OS 커널을 공유하되, Linux 커널의 격리 기능을 사용합니다.

### Linux Namespaces: 프로세스 격리

```
Namespace 종류와 격리 범위:

PID Namespace:
  호스트: PID 1(init), PID 1234(dockerd), PID 5678(컨테이너 내부 nginx)
  컨테이너 내부: PID 1(nginx)  ← 호스트 PID 5678이 컨테이너 내 PID 1로 보임
  효과: 컨테이너 내부에서 다른 컨테이너 프로세스를 볼 수 없음

Network Namespace:
  각 컨테이너가 독립적인 네트워크 스택 보유
  고유한 eth0, 라우팅 테이블, 방화벽 규칙
  veth pair로 호스트와 연결

Mount Namespace:
  컨테이너별 독립적인 파일시스템 뷰
  호스트 /proc, /sys를 마운트하지 않으면 보이지 않음

UTS Namespace:
  컨테이너별 독립적인 hostname, domainname
  각 컨테이너가 자신의 이름을 가질 수 있음

IPC Namespace:
  독립적인 System V IPC, POSIX 메시지 큐
  컨테이너 간 IPC 통신 격리

User Namespace:
  컨테이너 내 root(UID 0) = 호스트의 일반 사용자(UID 65534)
  컨테이너 탈출해도 호스트에서 일반 권한만 가짐 (보안 강화)
```

Go에서 Namespace 사용:
```go
// runc 내부 - Linux 시스템 콜로 Namespace 생성
// github.com/opencontainers/runc/libcontainer/configs/namespaces_linux.go

import "syscall"

// 새 프로세스를 새 Namespace에서 시작
cmd := &exec.Cmd{
    Path: "/bin/sh",
    SysProcAttr: &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWPID |   // 새 PID namespace
                    syscall.CLONE_NEWNET |   // 새 Network namespace
                    syscall.CLONE_NEWNS  |   // 새 Mount namespace
                    syscall.CLONE_NEWUTS |   // 새 UTS namespace
                    syscall.CLONE_NEWIPC,    // 새 IPC namespace
    },
}
```

### Cgroups: 리소스 제한

```
Cgroups v2 계층 구조:

/sys/fs/cgroup/
├── system.slice/
│   └── docker-{container-id}.scope/
│       ├── cpu.max          ← CPU 제한
│       ├── memory.max       ← 메모리 제한
│       ├── io.max           ← 디스크 I/O 제한
│       └── pids.max         ← 프로세스 수 제한
└── ...

예: CPU 제한
  cpu.max = "100000 1000000"
  → 1000000 마이크로초(1초) 중 100000 마이크로초만 사용
  → CPU 10% 제한
```

```go
// runc가 cgroup 설정하는 방식 (단순화)
func setCgroupLimits(containerID string, resources Resources) error {
    cgroupPath := "/sys/fs/cgroup/system.slice/docker-" + containerID + ".scope"

    // CPU 제한 설정
    if resources.CPUQuota > 0 {
        cpuMax := fmt.Sprintf("%d %d", resources.CPUQuota, resources.CPUPeriod)
        if err := os.WriteFile(cgroupPath+"/cpu.max", []byte(cpuMax), 0644); err != nil {
            return err
        }
    }

    // 메모리 제한 설정
    if resources.Memory > 0 {
        memStr := strconv.FormatInt(resources.Memory, 10)
        if err := os.WriteFile(cgroupPath+"/memory.max", []byte(memStr), 0644); err != nil {
            return err
        }
    }

    return nil
}
```

### Union Filesystem: 이미지 레이어

```
OverlayFS 구조:

  upperdir (컨테이너 쓰기 레이어)  ← 컨테이너에서 수정한 파일만
      +
  lowerdir-3 (앱 레이어)           ← 읽기 전용
      +
  lowerdir-2 (Python 레이어)       ← 읽기 전용
      +
  lowerdir-1 (Ubuntu 레이어)       ← 읽기 전용
      =
  merged (통합 뷰)                 ← 컨테이너가 보는 파일시스템

마운트 명령:
  mount -t overlay overlay \
    -o lowerdir=/var/lib/containerd/layers/ubuntu:/var/lib/containerd/layers/python,\
       upperdir=/var/lib/containerd/containers/abc123/diff,\
       workdir=/var/lib/containerd/containers/abc123/work \
    /var/lib/containerd/containers/abc123/rootfs

Copy-on-Write:
  컨테이너가 /etc/nginx/nginx.conf 수정 시:
  1. lowerdir에서 upperdir로 파일 복사
  2. upperdir의 파일 수정
  3. 원본 lowerdir는 변경 없음
  → 같은 이미지를 쓰는 100개 컨테이너가 lowerdir를 공유
```

---

## Docker 소스코드 분석 포인트

### moby/moby: Docker 엔진

```
moby/moby/
├── cmd/dockerd/         ← dockerd 진입점
│   └── daemon.go        ← 데몬 초기화
├── daemon/
│   ├── container.go     ← Container 구조체
│   ├── start.go         ← docker start 처리
│   ├── create.go        ← docker create 처리
│   └── images/          ← 이미지 관리
├── client/
│   └── client.go        ← Docker SDK 클라이언트
└── api/
    └── server/
        └── router/      ← REST API 라우터
```

### containerd/containerd: 컨테이너 런타임

```
containerd/containerd/
├── cmd/containerd/      ← containerd 진입점
├── snapshots/           ← Snapshotter 인터페이스 (레이어 관리)
├── content/             ← Content Store (이미지 Blob 저장)
├── runtime/v2/          ← 런타임 관리 (shim과 통신)
├── services/            ← gRPC 서비스 구현
└── client.go            ← Go 클라이언트 API
```

### opencontainers/runc: OCI 런타임

```
opencontainers/runc/
├── main.go              ← CLI 진입점 (create, start, exec, delete)
├── libcontainer/
│   ├── container_linux.go  ← 컨테이너 생성/시작 핵심
│   ├── namespaces_linux.go ← Namespace 설정
│   └── cgroups/            ← Cgroup 설정
└── spec_linux.go        ← OCI 스펙 파싱
```

---

## 컨테이너 라이프사이클: 상태 머신 패턴

```
상태 전환 다이어그램:

  [없음]
    │ Create
    ▼
 [Created]
    │ Start
    ▼
 [Running] ←──── Unpause ────┐
    │                         │
    │ Pause                   │
    ▼                         │
 [Paused] ───────────────────┘
    │
    │ (또는 Running에서)
    │ Stop / Kill
    ▼
 [Stopped/Exited]
    │ Remove
    ▼
  [없음]
```

Go에서 상태 머신 구현 패턴:

```go
// containerd의 상태 표현 (단순화)
type Status int

const (
    StatusUnknown  Status = 0
    StatusCreated  Status = 1
    StatusRunning  Status = 2
    StatusStopped  Status = 3
    StatusPaused   Status = 4
    StatusDeleted  Status = 5
)

type Container struct {
    id     string
    status Status
    mu     sync.RWMutex
}

// 상태 전환: 허용된 전환만 허용
func (c *Container) transition(from, to Status) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    if c.status != from {
        return fmt.Errorf("cannot transition from %v to %v: current state is %v",
            from, to, c.status)
    }
    c.status = to
    return nil
}

func (c *Container) Start(ctx context.Context) error {
    if err := c.transition(StatusCreated, StatusRunning); err != nil {
        return err
    }
    // 실제 프로세스 시작...
    return nil
}

func (c *Container) Stop(ctx context.Context) error {
    if err := c.transition(StatusRunning, StatusStopped); err != nil {
        return err
    }
    // SIGTERM 전송...
    return nil
}
```

---

## 이미지 레이어 시스템: Content-Addressable Store

### 내용 기반 주소 지정

```
Content-Addressable Storage의 원리:
  데이터의 SHA256 해시 = 데이터의 주소

  nginx 이미지 레이어 (50MB tar):
  SHA256(데이터) = "sha256:a1b2c3d4e5..."
  저장 위치: /var/lib/containerd/io.containerd.content.v1.content/blobs/sha256/a1b2c3d4e5...

장점:
  - 중복 제거: 같은 내용은 한 번만 저장
  - 무결성 검증: 주소(해시)로 데이터 손상 자동 감지
  - 캐싱: 같은 레이어는 다운로드 불필요

같은 Ubuntu 레이어를 쓰는 nginx, redis, python 이미지:
  nginx 이미지: [ubuntu-layer-sha256:aaa, nginx-layer-sha256:bbb]
  redis 이미지: [ubuntu-layer-sha256:aaa, redis-layer-sha256:ccc]  ← ubuntu 공유
  python 이미지: [ubuntu-layer-sha256:aaa, python-layer-sha256:ddd] ← ubuntu 공유
  → ubuntu 레이어는 디스크에 한 번만 저장됨
```

```go
// containerd Content Store 인터페이스
// containerd/containerd/content/content.go
type Store interface {
    // 콘텐츠 조회
    Info(ctx context.Context, dgst digest.Digest) (Info, error)

    // 콘텐츠 읽기
    ReaderAt(ctx context.Context, desc ocispec.Descriptor) (ReaderAt, error)

    // 콘텐츠 쓰기 (스트리밍)
    Writer(ctx context.Context, opts ...WriterOpt) (Writer, error)

    // 콘텐츠 삭제
    Delete(ctx context.Context, dgst digest.Digest) error
}

// Digest = "sha256:abcdef..."
type Digest = digest.Digest  // github.com/opencontainers/go-digest

// 콘텐츠 쓰기 예시
func writeContent(ctx context.Context, cs content.Store,
    reader io.Reader, expectedDigest digest.Digest) error {

    writer, err := cs.Writer(ctx,
        content.WithRef("layer-"+expectedDigest.String()),
        content.WithDescriptor(ocispec.Descriptor{
            Digest: expectedDigest,
        }),
    )
    if err != nil {
        return err
    }
    defer writer.Close()

    // 스트리밍으로 저장 (메모리에 전체 로드 불필요)
    if _, err := io.Copy(writer, reader); err != nil {
        return err
    }

    // 커밋 (해시 검증 포함)
    return writer.Commit(ctx, 0, expectedDigest)
}
```

### Snapshotter 인터페이스: 레이어 시스템 추상화

```go
// containerd/containerd/snapshots/snapshots.go
// 다양한 파일시스템 구현(OverlayFS, AUFS, btrfs 등)을 동일 인터페이스로
type Snapshotter interface {
    // Stat: 스냅샷 정보 조회
    Stat(ctx context.Context, key string) (Info, error)

    // Prepare: 쓰기 가능한 스냅샷 생성 (parent 위에 새 레이어)
    // 컨테이너 시작 시 사용 (upperdir 생성)
    Prepare(ctx context.Context, key, parent string, opts ...Opt) ([]mount.Mount, error)

    // View: 읽기 전용 스냅샷 생성
    // 이미지 레이어 마운트 시 사용
    View(ctx context.Context, key, parent string, opts ...Opt) ([]mount.Mount, error)

    // Commit: 쓰기 레이어를 읽기 전용으로 확정
    // docker commit 또는 이미지 빌드 시 사용
    Commit(ctx context.Context, name, key string, opts ...Opt) error

    // Remove: 스냅샷 삭제
    Remove(ctx context.Context, key string) error

    // Walk: 모든 스냅샷 순회
    Walk(ctx context.Context, fn WalkFunc, filters ...string) error
}

// OverlayFS 구현 (단순화)
type snapshotter struct {
    root string  // /var/lib/containerd/snapshots
}

func (s *snapshotter) Prepare(ctx context.Context, key, parent string,
    opts ...snapshots.Opt) ([]mount.Mount, error) {

    // parent 스냅샷의 레이어들을 lowerdir로 수집
    lowers, err := s.getLowers(parent)
    if err != nil {
        return nil, err
    }

    // 새 upperdir 생성
    upperDir := filepath.Join(s.root, "snapshots", key, "fs")
    workDir := filepath.Join(s.root, "snapshots", key, "work")
    os.MkdirAll(upperDir, 0755)
    os.MkdirAll(workDir, 0755)

    // OverlayFS 마운트 옵션 반환
    return []mount.Mount{
        {
            Type:   "overlay",
            Source: "overlay",
            Options: []string{
                "lowerdir=" + strings.Join(lowers, ":"),
                "upperdir=" + upperDir,
                "workdir=" + workDir,
            },
        },
    }, nil
}
```

---

## Plugin 아키텍처

### 인터페이스 기반 드라이버 교체

Docker의 플러그인 시스템은 인터페이스를 통해 볼륨, 네트워크, 로그 드라이버를 런타임에 교체할 수 있게 합니다.

```go
// Volume Driver 인터페이스 (moby/moby/volume/drivers/extpoint.go)
type VolumeDriver interface {
    // 드라이버 이름 (예: "local", "nfs", "s3")
    Name() string

    // 볼륨 생성
    Create(name string, opts map[string]string) error

    // 볼륨 삭제
    Remove(name string) error

    // 컨테이너에 볼륨 마운트 (마운트 경로 반환)
    Mount(name string, id string) (string, error)

    // 컨테이너에서 볼륨 언마운트
    Unmount(name string, id string) error

    // 볼륨 목록
    List() ([]*volumes.Volume, error)

    // 볼륨 존재 여부
    Get(name string) (*volumes.Volume, error)
}

// 드라이버 레지스트리
type DriverExtpoint struct {
    extensions map[string]VolumeDriver  // name → driver
    mu         sync.Mutex
}

func (e *DriverExtpoint) Register(d VolumeDriver, name string) bool {
    e.mu.Lock()
    defer e.mu.Unlock()
    if _, exists := e.extensions[name]; exists {
        return false
    }
    e.extensions[name] = d
    return true
}

func (e *DriverExtpoint) Lookup(name string) (VolumeDriver, error) {
    e.mu.Lock()
    d, ok := e.extensions[name]
    e.mu.Unlock()
    if ok {
        return d, nil
    }
    // 플러그인 프로세스에서 로드 시도
    return e.loadPlugin(name)
}
```

### 외부 플러그인: HTTP over Unix Socket

```go
// 외부 플러그인은 Unix Socket을 통해 HTTP 통신
// 플러그인 프로세스는 Docker Daemon과 별도 프로세스

// 플러그인 구현 (볼륨 드라이버 예시)
func main() {
    plugin := &myVolumePlugin{
        volumes: map[string]string{},
    }

    // Docker 플러그인 HTTP 핸들러
    http.HandleFunc("/Plugin.Activate", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "Implements": []string{"VolumeDriver"},
        })
    })

    http.HandleFunc("/VolumeDriver.Create", func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            Name    string
            Options map[string]string
        }
        json.NewDecoder(r.Body).Decode(&req)

        if err := plugin.Create(req.Name, req.Options); err != nil {
            json.NewEncoder(w).Encode(map[string]string{"Err": err.Error()})
            return
        }
        json.NewEncoder(w).Encode(map[string]string{"Err": ""})
    })

    http.HandleFunc("/VolumeDriver.Mount", func(w http.ResponseWriter, r *http.Request) {
        var req struct{ Name, ID string }
        json.NewDecoder(r.Body).Decode(&req)

        mountpoint, err := plugin.Mount(req.Name, req.ID)
        if err != nil {
            json.NewEncoder(w).Encode(map[string]string{"Err": err.Error()})
            return
        }
        json.NewEncoder(w).Encode(map[string]string{
            "Mountpoint": mountpoint,
            "Err":        "",
        })
    })

    // Unix Socket에서 수신 대기
    listener, _ := net.Listen("unix", "/run/docker/plugins/myplugin.sock")
    http.Serve(listener, nil)
}
```

---

## Go에서의 시스템 프로그래밍

Docker/containerd/runc는 Linux 커널 기능을 직접 사용하는 시스템 프로그래밍의 좋은 예입니다.

### os/exec: 프로세스 관리

```go
// exec.Cmd로 프로세스 제어
cmd := exec.CommandContext(ctx, "runc", "create", "--bundle", bundlePath, containerID)

// 표준 입출력 연결
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr

// 환경 변수 설정
cmd.Env = append(os.Environ(), "CONTAINER_ID="+containerID)

// 프로세스 그룹 설정 (자식 프로세스도 함께 종료)
cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

// 시작
if err := cmd.Start(); err != nil {
    return fmt.Errorf("start runc: %w", err)
}

// 비동기 대기
doneCh := make(chan error, 1)
go func() {
    doneCh <- cmd.Wait()
}()

select {
case err := <-doneCh:
    return err
case <-ctx.Done():
    // 타임아웃: 프로세스 그룹 전체 종료
    syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
    return ctx.Err()
}
```

### OCI 스펙: 표준화된 컨테이너 설명

```go
// github.com/opencontainers/runtime-spec/specs-go/config.go
// 모든 OCI 호환 런타임이 이해하는 표준 형식

type Spec struct {
    Version  string   // OCI 스펙 버전
    Process  *Process // 실행할 프로세스
    Root     *Root    // 루트 파일시스템
    Hostname string   // UTS namespace hostname
    Mounts   []Mount  // 추가 마운트
    Linux    *Linux   // Linux 특화 설정 (Namespaces, Cgroups 등)
}

type Process struct {
    Terminal bool     // 터미널 할당 여부
    Cwd      string   // 작업 디렉토리
    Env      []string // 환경 변수
    Args     []string // 실행 명령어 + 인자
    User     User     // 실행 사용자 (UID, GID)
    Capabilities *LinuxCapabilities // Linux capability 설정
}

type Linux struct {
    Namespaces  []LinuxNamespace  // 사용할 namespace
    Resources   *LinuxResources   // cgroup 리소스 제한
    Seccomp     *LinuxSeccomp     // 시스템 콜 필터링
    RootfsPropagation string      // 마운트 전파 설정
}

// 실제 bundle/config.json 예시
spec := specs.Spec{
    Version: "1.0.2",
    Process: &specs.Process{
        Args: []string{"/bin/nginx", "-g", "daemon off;"},
        Env:  []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
        Cwd:  "/",
    },
    Root: &specs.Root{
        Path:     "rootfs",
        Readonly: false,
    },
    Linux: &specs.Linux{
        Namespaces: []specs.LinuxNamespace{
            {Type: specs.PIDNamespace},
            {Type: specs.NetworkNamespace},
            {Type: specs.MountNamespace},
        },
        Resources: &specs.LinuxResources{
            Memory: &specs.LinuxMemory{
                Limit: int64Ptr(512 * 1024 * 1024), // 512MB
            },
        },
    },
}
```

---

## 패턴 적용: 과제 A6와의 연결

Docker 패턴은 A6 오퍼레이터 시뮬레이터에서 데이터베이스 인스턴스 라이프사이클 관리에 직접 적용됩니다.

```
실제 Kubernetes PostgreSQL Operator (예: Zalando postgres-operator):
  spec.instances: 3 선언
       ↓
  Controller Reconcile
       ↓
  현재 StatefulSet replicas 조회 (Lister → 캐시)
       ↓
  3개 안되면? → StatefulSet 업데이트 (API 서버 요청)
       ↓
  K8s가 Pod 생성 → kubelet → containerd → runc → PostgreSQL 프로세스

A6 시뮬레이터:
  spec.replicas: 3 선언
       ↓
  Reconcile
       ↓
  현재 DB 인스턴스 수 조회 (인메모리 상태)
       ↓
  부족하면 → DBInstance 생성 (상태: Created → Running)
       ↓
  DBInstance의 라이프사이클 관리 (Docker 컨테이너와 동일한 상태 머신)
```
