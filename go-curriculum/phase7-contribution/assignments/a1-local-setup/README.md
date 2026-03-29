# A1: 로컬 개발 환경 구축 (★★★☆☆)

## 목표

agent-sandbox 리포지토리를 포크하고, 로컬에서 빌드 및 테스트를 성공적으로 실행한다.

## 요구사항

### 1. 환경 준비
- [ ] Go 1.22+ 설치 확인
- [ ] Docker 실행 확인
- [ ] kind 설치 (`go install sigs.k8s.io/kind@latest`)
- [ ] kubectl 설치
- [ ] GitHub CLI 설치 및 인증 (`gh auth login`)

### 2. 리포지토리 설정
- [ ] `gh repo fork kubernetes-sigs/agent-sandbox --clone`
- [ ] `git remote add upstream https://github.com/kubernetes-sigs/agent-sandbox.git`
- [ ] `go mod download` 성공

### 3. 빌드 & 테스트
- [ ] `make build` 성공
- [ ] `make test` 전체 통과
- [ ] `go test -race ./...` 실행 (실패해도 OK — 현재 상태 기록)

### 4. 로컬 클러스터
- [ ] `kind create cluster --name agent-sandbox-dev`
- [ ] `make install` (CRD 설치)
- [ ] `kubectl get crd | grep sandbox` 로 CRD 확인

## 제출물

아래 내용을 정리하여 보고:
1. 각 단계의 성공/실패 스크린샷 또는 출력
2. `go test -race ./...` 실행 시 발견된 레이스 목록 (있다면)
3. `kubectl get crd` 출력 결과

## 평가 기준

| 항목 | 배점 |
|------|------|
| 빌드 성공 | 25 |
| 단위 테스트 통과 | 25 |
| race detector 실행 | 25 |
| kind 클러스터 + CRD 설치 | 25 |
| **합계** | **100** |
