# Phase 6: K8s 오픈소스 딥다이브

**기간:** 24주차+
**목표:** 프로덕션 오픈소스 코드를 읽고 기여할 수 있는 역량을 갖춘다
**디렉토리:** `go-curriculum/phase6-opensource/`

---

## 레슨 목록

| # | 디렉토리 | 주제 | 핵심 개념 |
|---|---------|------|----------|
| 1 | `01-k8s-patterns` | K8s 패턴 | Controller, Informer, WorkQueue, Reconciliation |
| 2 | `02-prometheus-patterns` | Prometheus 패턴 | Registry, Collector, 메트릭 수집 아키텍처 |
| 3 | `03-etcd-patterns` | etcd 패턴 | Watch, MVCC, Lease, 분산 합의 |
| 4 | `04-docker-patterns` | Docker 패턴 | 컨테이너 라이프사이클, 이미지 빌드, 네트워킹 |
| 5 | `05-terraform-patterns` | Terraform 패턴 | DAG, Plan/Apply, Provider, State 관리 |

## 레슨 파일 경로

```
go-curriculum/phase6-opensource/
├── 01-k8s-patterns/THEORY.md         + main.go
├── 02-prometheus-patterns/THEORY.md  + main.go
├── 03-etcd-patterns/THEORY.md        + main.go
├── 04-docker-patterns/THEORY.md      + main.go
└── 05-terraform-patterns/THEORY.md   + main.go
```

---

## 과제 목록

| # | 디렉토리 | 주제 | 난이도 | 핵심 |
|---|---------|------|--------|------|
| A1 | `assignments/a1-controller` | 리소스 컨트롤러 | ★★★★½ | K8s Controller 패턴 |
| A2 | `assignments/a2-informer` | 인포머/리스터 | ★★★★★ | client-go Informer |
| A3 | `assignments/a3-metrics-registry` | 메트릭 레지스트리 | ★★★★☆ | Prometheus 패턴 |
| A4 | `assignments/a4-dag-executor` | DAG 태스크 실행기 | ★★★★★ | Terraform DAG 패턴 |
| A5 | `assignments/a5-watch-system` | 이벤트 감시 시스템 | ★★★★½ | etcd Watch 패턴 |
| A6 | `assignments/a6-reconciler` | 오퍼레이터 시뮬레이터 | ★★★★★ | 캡스톤 — 전체 통합 |

## 과제 파일 경로

```
go-curriculum/phase6-opensource/assignments/
├── a1-controller/
├── a2-informer/
├── a3-metrics-registry/
├── a4-dag-executor/
├── a5-watch-system/
└── a6-reconciler/
```
각 과제: README.md + 구현 파일(.go) + 테스트 파일(_test.go) + solution/

---

## 프로젝트

Phase 6에는 별도 프로젝트가 없습니다.
A6 (오퍼레이터 시뮬레이터)가 캡스톤 프로젝트 역할을 합니다.

---

## 오픈소스 코드 리딩 가이드

Phase 6의 핵심은 **실제 프로덕션 코드를 읽는 것**:

### 추천 코드 리딩 순서
1. **client-go/tools/cache**: Informer, Lister, SharedIndexInformer
2. **controller-runtime/pkg/reconcile**: Reconciler 인터페이스
3. **prometheus/client_golang**: Registry, Collector
4. **etcd/client/v3**: Watch, Lease
5. **terraform/dag**: DAG 구현, Walk 알고리즘

### 코드 리딩 팁
- 인터페이스부터 읽어라 (구현보다 계약이 중요)
- 테스트 코드를 먼저 읽어라 (사용법을 보여준다)
- `godoc` 주석을 무시하지 말라 (설계 의도가 담겨 있다)
