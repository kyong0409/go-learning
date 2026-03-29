# A2: Race Detector 분석 리포트 (★★★½☆)

## 목표

Issue #331을 기반으로, agent-sandbox 테스트에 race detector를 적용하고 결과를 분석한다.

## 배경

Go의 race detector (`go test -race`)는 데이터 레이스를 런타임에 감지합니다. agent-sandbox에는 아직 이것이 기본 활성화되어 있지 않습니다.

## 요구사항

### 1. 현재 상태 분석
- [ ] `go test -race ./controllers/...` 실행
- [ ] `go test -race ./extensions/controllers/...` 실행
- [ ] `go test -race ./api/...` 실행
- [ ] 발견된 모든 레이스를 목록화

### 2. 레이스 분석
발견된 각 레이스에 대해:
- [ ] 어떤 변수/필드에서 발생하는가?
- [ ] 어떤 고루틴들이 충돌하는가?
- [ ] 심각도 평가 (높음/중간/낮음)
- [ ] 수정 방향 제안 (Mutex? 채널? atomic?)

### 3. Makefile 수정안 작성
- [ ] `make test`에 `-race` 추가하는 패치 작성
- [ ] 또는 별도 `make test-race` 타겟 추가
- [ ] CI 워크플로우 수정안 (해당 시)

## 제출물

`report.md` 파일에 아래 내용 정리:

```markdown
# Race Detector 분석 리포트

## 실행 환경
- Go 버전:
- OS:
- agent-sandbox 버전/커밋:

## 발견된 레이스 목록
### Race 1: [위치]
- 충돌 변수:
- 고루틴 1:
- 고루틴 2:
- 심각도:
- 수정 제안:

## Makefile 수정안
[diff 또는 코드]

## CI 워크플로우 수정안
[diff 또는 코드]
```

## 평가 기준

| 항목 | 배점 |
|------|------|
| 모든 패키지에서 race detector 실행 | 20 |
| 발견된 레이스 정확한 목록화 | 25 |
| 각 레이스의 원인 분석 | 25 |
| Makefile/CI 수정안 | 20 |
| 리포트 품질 | 10 |
| **합계** | **100** |
