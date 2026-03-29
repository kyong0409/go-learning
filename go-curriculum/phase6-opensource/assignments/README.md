# Phase 6 과제 모음

K8s 생태계 오픈소스 패턴을 직접 구현해보는 과제들입니다.
모든 과제는 실제 프로덕션 코드에서 쓰이는 패턴을 단순화한 것입니다.

## 과제 목록

| 과제 | 패턴 출처 | 핵심 구현 | 난이도 |
|------|-----------|-----------|--------|
| [A1 - 리소스 컨트롤러](a1-controller/) | Kubernetes | Watch+Reconcile+WorkQueue+재시도 | ★★★★½ |
| [A2 - 인포머/리스터](a2-informer/) | client-go | Store+Reflector+Informer+Lister | ★★★★★ |
| [A3 - 메트릭 레지스트리](a3-metrics-registry/) | Prometheus | Registry+Collector+레이블+텍스트 형식 | ★★★★☆ |
| [A4 - DAG 실행기](a4-dag-executor/) | Terraform/K8s | DAG+사이클감지+병렬실행+dry-run | ★★★★★ |
| [A5 - 감시 시스템](a5-watch-system/) | etcd | MVCC+Watch+리비전+압축 | ★★★★½ |
| [A6 - 오퍼레이터 시뮬레이터](a6-reconciler/) | K8s Operator | 통합 캡스톤 (모든 패턴) | ★★★★★ |

## 권장 학습 순서

난이도 순으로 진행하는 것을 권장합니다:

```
A3 (★★★★☆) → A1 (★★★★½) → A5 (★★★★½) → A2 (★★★★★) → A4 (★★★★★) → A6 (★★★★★)
```

## 진행 방법

1. 각 과제 폴더의 `README.md`를 먼저 읽어 요구사항 파악
2. 학습 자료 `01~05-*-patterns/README.md` 참고
3. 스켈레톤 파일의 함수 시그니처와 TODO 주석 확인
4. 테스트 먼저 실행해 실패 확인: `go test ./... -v`
5. 구현 후 모든 테스트 통과
6. `solution/` 디렉터리 참고 답안과 비교

## 채점 방법

각 과제의 `*_test.go`를 실행하면 한국어 점수 보고서가 출력됩니다:

```bash
go test -v -run TestGrade
```

## 공통 규칙

- 표준 라이브러리만 사용 (외부 패키지 금지, 단 테스트 도구 제외)
- 테스트 파일 수정 금지
- `solution/` 디렉터리는 막힐 때만 참고
- 코드 주석은 한국어로 작성 권장
- 모든 goroutine은 컨텍스트 취소 시 정리되어야 함

**Phase 6 총 만점: 600점**
