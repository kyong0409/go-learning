# 02: Agent Sandbox — 프로젝트 이해

## Agent Sandbox란?

Agent Sandbox는 **AI 에이전트가 안전하게 코드를 실행할 수 있는 격리된 환경**을 Kubernetes 위에서 제공하는 프로젝트입니다.

- **소속**: kubernetes-sigs/agent-sandbox (SIG Apps 서브프로젝트)
- **리포지토리**: https://github.com/kubernetes-sigs/agent-sandbox
- **문서**: https://agent-sandbox.sigs.k8s.io
- **현재 버전**: v0.2.1 (pre-stable, 모든 API가 `v1alpha1`)
- **발표**: KubeCon North America 2025 (Atlanta)
- **주요 기여자**: Google (원래 GKE 내부 프로젝트에서 오픈소스화)

---

## 해결하는 문제

### 문제 1: 레이턴시

AI 에이전트의 각 도구 호출(tool call)마다 격리된 환경이 필요합니다. 하지만 Pod 콜드 스타트는 ~1초 이상 걸려서 에이전트의 빠른 피드백 루프를 깨뜨립니다.

```
에이전트: "이 Python 코드를 실행해줘"
  → Pod 생성 대기 (~1초+)  ← 이게 문제!
  → 코드 실행 (~0.1초)
  → 결과 반환
```

### 문제 2: 스케일

엔터프라이즈 플랫폼은 수만 개의 병렬 샌드박스가 필요하고, 초당 수천 건의 쿼리를 처리해야 합니다.

### 문제 3: 보안

LLM이 생성한 코드는 신뢰할 수 없습니다. Langflow, Cursor, Replit 등에서 RCE(Remote Code Execution) 취약점이 실제로 발생했습니다. 커널 수준의 격리가 필요합니다.

### 문제 4: 기존 추상화의 한계

Kubernetes에는 "오래 실행되는 상태를 가진 단일 컨테이너"에 맞는 깔끔한 추상화가 없습니다. StatefulSet(크기 1) + Headless Service + PVC를 조합해야 하는데, 이는 번거롭고 하이버네이션 같은 라이프사이클 관리가 부족합니다.

---

## 핵심 아키텍처

### CRD (Custom Resource Definitions)

Agent Sandbox는 4개의 CRD를 정의합니다:

#### 1. Sandbox (코어)

가장 기본적인 리소스. 격리된 단일 Pod을 관리합니다.

```yaml
apiVersion: agents.x-k8s.io/v1alpha1
kind: Sandbox
metadata:
  name: my-sandbox
spec:
  podTemplate:
    spec:
      containers:
      - name: my-container
        image: python:3.12-slim
```

**특징:**
- 안정적인 아이덴티티 (고정 호스트네임, 네트워크 ID)
- 영속적 스토리지 (Pod 재시작 후에도 유지)
- 라이프사이클 관리 (생성, 일시정지/하이버네이션, 재개, 예약 삭제)

#### 2. SandboxTemplate (확장)

재사용 가능한 블루프린트. 리소스 제한, 베이스 이미지, 보안 정책을 정의합니다.

```yaml
apiVersion: agents.x-k8s.io/v1alpha1
kind: SandboxTemplate
metadata:
  name: python-sandbox
spec:
  template:
    spec:
      containers:
      - name: sandbox
        image: python:3.12-slim
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
```

#### 3. SandboxClaim (확장)

상위 프레임워크(ADK, LangChain 등)가 실행 환경을 요청할 때 사용하는 트랜잭셔널 리소스.

#### 4. SandboxWarmPool (확장)

미리 워밍된 Pod 풀을 유지. Claim이 들어오면 1초 이내에 환경을 제공합니다 (콜드 스타트 대비 ~90% 레이턴시 감소).

```yaml
apiVersion: agents.x-k8s.io/v1alpha1
kind: SandboxWarmPool
metadata:
  name: python-pool
spec:
  templateRef:
    name: python-sandbox
  replicas: 5    # 항상 5개를 미리 준비
```

### 격리 백엔드

| 백엔드 | 수준 | 설명 |
|--------|------|------|
| **gVisor** (기본) | 커널 수준 | 사용자 공간 커널로 시스템 콜 인터셉트 |
| **Kata Containers** | VM 수준 | 경량 VM으로 완전한 하드웨어 격리 |
| QEMU, Firecracker | (로드맵) | 2026년 추가 예정 |

---

## 리포지토리 구조

```
kubernetes-sigs/agent-sandbox/
├── api/v1alpha1/                     # 코어 CRD 타입 정의
├── extensions/api/v1alpha1/          # 확장 CRD 타입 정의
├── controllers/                      # 코어 Sandbox 컨트롤러 (reconciliation)
├── extensions/controllers/           # 확장 컨트롤러
├── cmd/agent-sandbox-controller/     # 컨트롤러 바이너리 진입점
├── clients/
│   ├── k8s/clientset/versioned/      # Go 타입드 클라이언트셋
│   ├── k8s/extensions/clientset/     # 확장 클라이언트셋
│   └── python/                       # Python SDK
├── internal/metrics/                 # Prometheus 메트릭
├── test/e2e/framework/              # E2E 테스트 프레임워크
├── examples/                         # 사용 예시
└── extensions/examples/              # 확장 사용 예시
```

### 핵심 Go 의존성

| 패키지 | 역할 |
|--------|------|
| `sigs.k8s.io/controller-runtime` | 컨트롤러 reconciliation 프레임워크 |
| `k8s.io/client-go` | Kubernetes API 클라이언트 |
| `k8s.io/apimachinery` | K8s 타입 정의 및 유틸리티 |
| `sigs.k8s.io/agent-sandbox/api/v1alpha1` | Sandbox CRD 타입 |

---

## 프로젝트 현황 (2026년 3월 기준)

| 지표 | 값 |
|------|-----|
| 커밋 | 312 |
| 기여자 | 44명 |
| 스타 | ~1,600 |
| 포크 | 171 |
| 오픈 이슈 | 72 |
| 오픈 PR | 77 |
| 릴리스 | 6 (최신 v0.2.1) |

### 2026 로드맵 주요 항목

- Go 클라이언트 라이브러리 (#227)
- PVC 기반 스케일다운/재개 (하이버네이션)
- 상태 조건(status condition) 개선 (#119)
- 생성 레이턴시 메트릭 (#123)
- Startup Actions (#58)
- 프레임워크 통합: CrewAI, Ray RLlib
- 추가 격리 백엔드: QEMU, Firecracker
- 웹사이트 리디자인 (#166)
- Beta/GA 릴리스

---

## 기여 기회 (현재 오픈된 이슈)

### good first issue
| 이슈 | 제목 | 내용 |
|------|------|------|
| #331 | Go race detector 활성화 | 테스트와 CI에 `-race` 플래그 구현 |

### help wanted
| 이슈 | 제목 | 내용 |
|------|------|------|
| #403 | 지속적 벤치마킹 | 성능 추적 인프라 구축 |
| #323 | SandboxWarmPool 롤아웃 | 템플릿 업데이트 시 롤아웃 전략 |
| #265 | 권한 상승 취약점 | pod-name 어노테이션 보안 이슈 (critical) |
| #168 | 컨트롤러 E2E 테스트 | pod-name 어노테이션 검증 테스트 |
| #166 | 웹사이트 리디자인 | 공식 웹사이트 개선 |

---

## 핵심 정리

1. Agent Sandbox는 AI 에이전트의 안전한 코드 실행을 위한 K8s CRD 기반 프로젝트
2. 4개 CRD: Sandbox, SandboxTemplate, SandboxClaim, SandboxWarmPool
3. controller-runtime 기반의 표준 K8s 컨트롤러 패턴
4. gVisor/Kata Containers로 커널/VM 수준 격리
5. v0.2.1 (pre-stable) — 기여하기 좋은 초기 단계
6. `good first issue` #331 (race detector)이 가장 접근하기 쉬운 진입점
