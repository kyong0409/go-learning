# Phase 6: K8s 생태계 오픈소스 딥다이브

실제 오픈소스 코드를 연구하고, 그 패턴을 직접 구현해보는 고급 과정입니다.

## 학습 목표

- Kubernetes, Prometheus, etcd, Docker, Terraform 등 실제 프로덕션 코드에서 반복적으로 등장하는 Go 패턴 이해
- 복잡한 시스템의 핵심 추상화를 직접 구현하며 설계 감각 습득
- 오픈소스 코드를 읽고 참고하는 능력 개발

## 왜 오픈소스 코드를 공부하는가?

K8s 생태계 프로젝트들은 Go 언어의 가장 정교한 패턴들을 담고 있습니다:

- **Kubernetes**: Controller 패턴, Informer/WorkQueue, 선언적 API
- **client-go**: Reflector, Store, Lister - 캐시 기반 로컬 상태 관리
- **Prometheus**: Registry 패턴, Collector 인터페이스, 레이블 기반 메트릭
- **etcd**: MVCC 기반 KV 스토어, Watch 스트리밍, Raft 합의
- **Docker/containerd**: 플러그인 아키텍처, 레이어 시스템, OCI 표준
- **Terraform**: DAG 기반 실행, Provider 플러그인, Plan/Apply 2단계

이 패턴들은 실제 면접에서 자주 나오고, 클라우드 네이티브 개발에서 매일 사용합니다.

## 모듈 구성

| 모듈 | 학습 내용 | 참고 프로젝트 |
|------|-----------|---------------|
| [01-k8s-patterns](01-k8s-patterns/) | Controller, Informer, WorkQueue | kubernetes/kubernetes, client-go |
| [02-prometheus-patterns](02-prometheus-patterns/) | Registry, Collector, 메트릭 타입 | prometheus/client_golang |
| [03-etcd-patterns](03-etcd-patterns/) | Watch, MVCC, Lease | etcd-io/etcd |
| [04-docker-patterns](04-docker-patterns/) | 플러그인, 레이어, 라이프사이클 | moby/moby, containerd/containerd |
| [05-terraform-patterns](05-terraform-patterns/) | DAG, Provider, Plan/Apply | hashicorp/terraform |

## 과제 목록

| 과제 | 패턴 출처 | 핵심 구현 | 난이도 |
|------|-----------|-----------|--------|
| [A1 - 리소스 컨트롤러](assignments/a1-controller/) | Kubernetes | Watch+Reconcile+WorkQueue | ★★★★½ |
| [A2 - 인포머/리스터](assignments/a2-informer/) | client-go | Store+Reflector+Informer+Lister | ★★★★★ |
| [A3 - 메트릭 레지스트리](assignments/a3-metrics-registry/) | Prometheus | Registry+Collector+레이블 | ★★★★☆ |
| [A4 - DAG 실행기](assignments/a4-dag-executor/) | Terraform/K8s | DAG+사이클감지+병렬실행 | ★★★★★ |
| [A5 - 감시 시스템](assignments/a5-watch-system/) | etcd | MVCC+Watch+리비전 | ★★★★½ |
| [A6 - 오퍼레이터 시뮬레이터](assignments/a6-reconciler/) | Kubernetes Operator | 통합 캡스톤 과제 | ★★★★★ |

## 권장 학습 순서

```
학습 자료 (01~05) 읽기
        ↓
A3 메트릭 레지스트리 (★★★★☆) - 가장 독립적, 시작하기 좋음
        ↓
A1 리소스 컨트롤러 (★★★★½) - K8s 핵심 패턴
        ↓
A5 감시 시스템 (★★★★½) - etcd Watch 패턴
        ↓
A2 인포머/리스터 (★★★★★) - client-go 심화
        ↓
A4 DAG 실행기 (★★★★★) - Terraform 패턴
        ↓
A6 오퍼레이터 시뮬레이터 (★★★★★) - 캡스톤: 모든 패턴 통합
```

## 실제 소스코드 링크

학습 자료의 패턴들이 실제로 어느 파일에 있는지 확인하세요:

### Kubernetes / client-go
- Controller 패턴: `k8s.io/client-go/tools/cache/controller.go`
- Reflector: `k8s.io/client-go/tools/cache/reflector.go`
- WorkQueue: `k8s.io/client-go/util/workqueue/queue.go`
- DeltaFIFO: `k8s.io/client-go/tools/cache/delta_fifo.go`

### Prometheus
- Registry: `github.com/prometheus/client_golang/prometheus/registry.go`
- Counter: `github.com/prometheus/client_golang/prometheus/counter.go`
- Histogram: `github.com/prometheus/client_golang/prometheus/histogram.go`

### etcd
- MVCC Store: `go.etcd.io/etcd/server/mvcc/kvstore.go`
- Watcher: `go.etcd.io/etcd/server/mvcc/watcher.go`

### Terraform
- Graph: `github.com/hashicorp/terraform/internal/dag/graph.go`
- Walker: `github.com/hashicorp/terraform/internal/dag/walk.go`

## 진행 방법

1. 각 `0N-*-patterns/README.md`로 이론과 실제 코드 파일 위치를 먼저 파악
2. 과제 `README.md`의 요구사항 숙지
3. 스켈레톤 파일의 함수 시그니처 확인
4. `go test ./... -v`로 현재 실패 확인
5. 구현 후 모든 테스트 통과
6. `solution/` 참고 답안과 비교

## 채점

```bash
# 단일 과제 채점
cd assignments/a1-controller
go test -v -run TestGrade

# 전체 Phase 6 채점
cd assignments
for d in a1-controller a2-informer a3-metrics-registry a4-dag-executor a5-watch-system a6-reconciler; do
  echo "=== $d ==="; cd $d && go test -v -run TestGrade 2>&1 | tail -8; cd ..
done
```

**Phase 6 총 만점: 600점** (각 과제 100점)
