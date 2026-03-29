# 04: 기여 실습 — 로컬 개발 환경과 첫 PR

## 로컬 개발 환경 구축

### 필수 도구 설치

```bash
# 1. Go (1.22+)
go version

# 2. Docker
docker version

# 3. kind (Kubernetes IN Docker)
go install sigs.k8s.io/kind@latest
kind version

# 4. kubectl
# Windows: winget install Kubernetes.kubectl
kubectl version --client

# 5. GitHub CLI
gh auth status

# 6. ko (컨테이너 이미지 빌드)
go install github.com/google/ko@latest
```

### 리포지토리 포크 & 클론

```bash
# 포크 + 클론 한 번에
gh repo fork kubernetes-sigs/agent-sandbox --clone
cd agent-sandbox

# upstream 설정
git remote add upstream https://github.com/kubernetes-sigs/agent-sandbox.git
git fetch upstream

# 의존성 확인
go mod download
```

### 로컬 클러스터 생성

```bash
# kind 클러스터 생성
kind create cluster --name agent-sandbox-dev

# CRD 설치
make install

# 컨트롤러 로컬 실행 (클러스터 밖에서)
make run
```

---

## 빌드 & 테스트 사이클

### 일상적인 개발 루프

```bash
# 1. 코드 수정
# 2. 빌드 확인
make build

# 3. 단위 테스트
make test

# 4. 특정 패키지만 테스트
go test ./controllers/... -v -run TestSandboxReconciler

# 5. race detector (이슈 #331)
go test -race ./controllers/...

# 6. 린트
make lint

# 7. CRD 매니페스트 재생성 (API 변경 시)
make manifests
make generate
```

### E2E 테스트

```bash
# kind 클러스터가 실행 중이어야 함
make test-e2e

# 특정 테스트만
go test ./test/e2e/... -v -run TestSandboxCreation
```

---

## 첫 번째 기여: Issue #331 (Race Detector)

이 이슈는 `good first issue`로 라벨링되어 있고, 기여의 진입점으로 적합합니다.

### 문제

현재 테스트에 Go의 race detector (`-race` 플래그)가 활성화되어 있지 않습니다. 데이터 레이스가 잠재적으로 존재할 수 있습니다.

### 해결 방향

1. **테스트 실행에 `-race` 플래그 추가**
   - Makefile의 `test` 타겟에 `-race` 추가
   - 또는 별도의 `test-race` 타겟 생성

2. **CI 파이프라인에 race detection 추가**
   - GitHub Actions 워크플로우에 race detector 단계 추가

3. **발견된 레이스 수정**
   - `-race`를 켜면 기존 테스트에서 레이스가 발견될 수 있음
   - 각 레이스를 분석하고 수정

### PR 작성 가이드

```bash
# 1. 브랜치 생성
git checkout -b fix/enable-race-detector

# 2. Makefile 수정
# test 타겟에 -race 추가

# 3. 로컬에서 확인
make test  # race detector와 함께 실행

# 4. 커밋
git add Makefile
git commit -s -m "test: Enable Go race detector in unit tests

Fixes #331

This enables the -race flag for all unit test runs, both locally
and in CI, to detect data races early.

Signed-off-by: Your Name <your-email@example.com>"

# 5. 푸시 & PR 생성
git push origin fix/enable-race-detector
gh pr create \
  --title "test: Enable Go race detector in unit tests" \
  --body "Fixes #331

## What this PR does
Enables Go's race detector in the test suite and CI pipeline.

## How to test
\`\`\`bash
make test  # Now includes -race flag
\`\`\`
"
```

---

## PR 체크리스트

PR을 올리기 전에 확인:

- [ ] `make build` 성공
- [ ] `make test` 통과
- [ ] `make lint` 통과
- [ ] 커밋 메시지에 `Signed-off-by` 포함 (`git commit -s`)
- [ ] 관련 이슈 번호 참조 (`Fixes #NNN`)
- [ ] 변경 사항에 대한 테스트 추가 (해당 시)
- [ ] `make manifests` / `make generate` 실행 (API 변경 시)

---

## 리뷰 대응 전략

### 흔한 리뷰 피드백과 대응

| 피드백 | 의미 | 대응 |
|--------|------|------|
| "nit:" | 사소한 스타일 이슈 | 바로 수정 |
| "Can you add a test?" | 테스트 커버리지 부족 | 테이블 주도 테스트 추가 |
| "This could race" | 동시성 이슈 가능성 | sync 패키지로 보호 |
| "Please squash" | 커밋 정리 요청 | `git rebase -i` |
| "/hold" | 머지 보류 | 이유 확인 후 논의 |

### Force Push 주의

```bash
# 리뷰 후 수정 시 — 새 커밋으로 추가 (리뷰 추적 가능)
git commit -s -m "address review feedback"

# 메인테이너가 squash 요청 시에만
git rebase -i HEAD~3
git push --force-with-lease
```

---

## 핵심 정리

1. kind + make + go test가 로컬 개발의 핵심 도구
2. Issue #331 (race detector)이 첫 기여로 가장 적합
3. 모든 커밋에 `Signed-off-by` 필수 (`git commit -s`)
4. PR은 작게, 한 가지 이슈만 해결
5. 리뷰 피드백에 빠르게, 겸손하게 대응
