# 과제 A6: 쿠버네티스 오퍼레이터 시뮬레이터 (캡스톤)

**난이도**: ★★★★★
**예상 소요 시간**: 8~12시간
**참고 패턴**: Kubernetes Operator Pattern (모든 Phase 6 패턴 통합)

## 배경

Kubernetes Operator는 커스텀 리소스(CR)를 감시하고,
실제 시스템을 원하는 상태(spec)로 맞추는 컨트롤러입니다.
이 과제는 Phase 6의 모든 패턴을 통합하는 캡스톤 과제입니다.

## 요구사항

### 관리할 리소스: Database CR

```go
type DatabaseSpec struct {
    Replicas int    // 원하는 인스턴스 수 (1~5)
    Version  string // DB 버전 (예: "5.7", "8.0")
    Storage  int    // 스토리지 크기(GB)
}

type ConditionType string
const (
    ConditionReady       ConditionType = "Ready"
    ConditionProgressing ConditionType = "Progressing"
    ConditionDegraded    ConditionType = "Degraded"
)

type Condition struct {
    Type    ConditionType
    Status  bool
    Reason  string
    Message string
}

type DatabaseStatus struct {
    ReadyReplicas int
    Phase         string      // "Pending", "Running", "Upgrading", "Failed"
    Conditions    []Condition
    Instances     []string    // 실행 중인 인스턴스 ID 목록
}

type Database struct {
    Name       string
    Spec       DatabaseSpec
    Status     DatabaseStatus
    Finalizers []string  // 삭제 전 정리 작업 목록
    DeletionRequested bool
}
```

### DatabaseOperator가 구현해야 할 기능

#### 1. Watch + Reconcile 루프 (A1 패턴)
- `Watch(ctx, eventCh)`로 Database 변경 이벤트 수신
- WorkQueue로 키를 관리하고 Reconcile 호출
- 실패 시 지수 백오프 재시도

#### 2. 인스턴스 라이프사이클 (Docker 패턴)
- `Reconcile(ctx, name)`: Database의 spec과 실제 인스턴스 수를 비교
- 인스턴스 부족: 새 Instance 생성 (simulated)
- 인스턴스 초과: 초과 Instance 삭제
- 상태 업데이트: `Status.ReadyReplicas`, `Status.Phase`

```go
type Instance struct {
    ID      string
    DBName  string
    Version string
    Phase   string // "Starting", "Running", "Stopping", "Stopped"
}
```

#### 3. 버전 업그레이드 (롤링)
- `spec.Version`이 변경되면 롤링 업그레이드 수행
- 한 번에 하나의 인스턴스만 업그레이드
- 업그레이드 중 `Status.Phase = "Upgrading"`, `Condition{Progressing: true}`

#### 4. Finalizer 패턴
- Database에 `"database.cleanup"` finalizer가 있으면
- `DeletionRequested=true` 시 먼저 cleanup(모든 인스턴스 삭제) 후 finalizer 제거
- finalizer 제거 후 실제 삭제 처리

#### 5. Status Conditions
- `Ready`: 모든 인스턴스가 Running 상태
- `Progressing`: 인스턴스 생성/삭제/업그레이드 진행 중
- `Degraded`: 일부 인스턴스 실패

#### 6. 리더 선출 (단순화)
```go
type LeaderElection struct {
    // TryAcquire() bool  - 리더 획득 시도
    // IsLeader() bool    - 현재 리더인지 확인
    // Release()          - 리더 반납
}
```
- 오퍼레이터가 리더가 아니면 Reconcile 하지 않음
- 단순화: 인메모리 mutex 기반 구현

## 채점 기준 (100점)

| 항목 | 점수 |
|------|------|
| 기본 Create → Reconcile → 인스턴스 생성 | 15점 |
| 스케일 업 (replicas 증가) | 10점 |
| 스케일 다운 (replicas 감소) | 10점 |
| 버전 업그레이드 (롤링) | 15점 |
| Finalizer 패턴 (삭제 정리) | 15점 |
| Status Conditions 업데이트 | 15점 |
| 리더 선출 | 10점 |
| 에러 복구 + 컨텍스트 취소 | 10점 |

## 실행 방법

```bash
cd a6-reconciler
go mod tidy
go test ./... -v
go test -v -run TestGrade
```

## 참고 자료

모든 Phase 6 학습 자료를 참고하세요:
- `../01-k8s-patterns/README.md` - Controller, WorkQueue
- `../03-etcd-patterns/README.md` - Watch 패턴
- `../04-docker-patterns/README.md` - 인스턴스 라이프사이클
- Phase 6 과제 A1~A5의 solution/

이 과제는 Phase 6의 모든 패턴을 통합합니다.
막히면 각 과제의 solution을 참고하세요.
