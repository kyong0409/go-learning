# 01: Kubernetes 오픈소스 기여 워크플로우

## 왜 오픈소스에 기여하는가?

코드를 읽는 것과 기여하는 것은 완전히 다른 경험입니다. Phase 6에서 K8s 패턴을 학습했다면, Phase 7에서는 **실제 프로덕션 프로젝트에 코드를 제출**합니다.

기여를 통해 얻는 것:
- 프로덕션 수준의 코드 리뷰 경험 (Google/Meta 엔지니어가 리뷰)
- Kubernetes 생태계에 대한 깊은 이해
- 오픈소스 이력 (GitHub 프로필에 kubernetes-sigs 기여 기록)
- 커뮤니티 네트워크 (SIG 미팅, Slack 채널)

---

## Kubernetes 프로젝트 구조

### 조직 체계

```
kubernetes/                  # 코어 Kubernetes
kubernetes-sigs/             # SIG (Special Interest Group) 산하 프로젝트
  ├── controller-runtime/    # 컨트롤러 프레임워크
  ├── kubebuilder/           # CRD 스캐폴딩 도구
  ├── agent-sandbox/         # AI 에이전트 샌드박스 (우리의 타겟!)
  └── ...
```

### SIG (Special Interest Group)

Kubernetes는 SIG 단위로 운영됩니다:
- **SIG Apps** — 애플리케이션 관리 (Deployment, StatefulSet, **Agent Sandbox**)
- **SIG Node** — 노드/컨테이너 런타임
- **SIG Network** — 네트워킹
- **SIG Security** — 보안

Agent Sandbox는 **SIG Apps**의 정식 서브프로젝트입니다.

---

## 기여 워크플로우

### 1단계: 커뮤니티 참여

```
Kubernetes Slack 가입 → #agent-sandbox 채널 참여 → SIG Apps 미팅 참석
```

- **Slack**: https://slack.k8s.io → `#agent-sandbox` 채널
- **메일링 리스트**: sig-apps@kubernetes.io
- **미팅**: 격주 월요일 9:00 AM PT (Agent Sandbox 전용 서브프로젝트 미팅)

### 2단계: 이슈 선택

```
GitHub Issues → "good first issue" 또는 "help wanted" 라벨 필터
```

**이슈 선택 기준:**
- `good first issue`: 코드베이스에 익숙하지 않아도 가능
- `help wanted`: 메인테이너가 외부 기여를 환영하는 이슈
- `kind/bug`: 버그 수정 (가장 기여하기 쉬움)
- `kind/feature`: 새 기능 (더 큰 범위)

### 3단계: CLA 서명

Kubernetes 프로젝트에 기여하려면 CNCF CLA(Contributor License Agreement)에 서명해야 합니다.

1. PR을 올리면 `cncf-cla: no` 라벨이 자동으로 붙음
2. CLA 서명 링크를 따라가서 서명
3. `cncf-cla: yes`로 변경됨

### 4단계: 포크 & 브랜치

```bash
# 1. 포크
gh repo fork kubernetes-sigs/agent-sandbox --clone

# 2. upstream 설정
cd agent-sandbox
git remote add upstream https://github.com/kubernetes-sigs/agent-sandbox.git

# 3. 기능 브랜치 생성
git checkout -b fix/enable-race-detector

# 4. 코드 수정 후 커밋
git add .
git commit -s -m "Enable race detector in test suite"
# -s: Signed-off-by 라인 추가 (DCO 필수!)

# 5. 푸시 & PR 생성
git push origin fix/enable-race-detector
gh pr create --title "Enable race detector in test suite" --body "..."
```

**중요: DCO (Developer Certificate of Origin)**
- 모든 커밋에 `Signed-off-by: Your Name <email>` 필요
- `git commit -s`로 자동 추가
- 빠뜨리면 CI가 실패함

### 5단계: 코드 리뷰 대응

```
PR 생성 → CI 확인 → 리뷰어 배정 → 피드백 반영 → 승인 → 머지
```

**Kubernetes 리뷰 프로세스:**
- `/lgtm` — 리뷰어가 코드를 승인 (Looks Good To Me)
- `/approve` — 메인테이너가 머지 승인
- `/hold` — 머지 보류
- `/retest` — CI 재실행

리뷰 피드백에 대한 태도:
- 빠르게 응답한다 (24-48시간 이내)
- 동의하지 않으면 근거를 들어 토론한다
- 메인테이너의 최종 결정을 존중한다

---

## 커밋 메시지 컨벤션

Kubernetes 프로젝트는 표준 커밋 메시지 형식을 따릅니다:

```
<type>: <description>

<body>

Signed-off-by: Your Name <email@example.com>
```

**타입:**
- `feat:` — 새 기능
- `fix:` — 버그 수정
- `test:` — 테스트 추가/수정
- `docs:` — 문서
- `refactor:` — 리팩토링
- `chore:` — 빌드/CI 관련

---

## 핵심 도구

| 도구 | 용도 |
|------|------|
| `gh` (GitHub CLI) | 포크, PR 생성, 이슈 관리 |
| `kind` (Kubernetes IN Docker) | 로컬 K8s 클러스터 |
| `ko` | Go 바이너리 → 컨테이너 이미지 빌드 |
| `make` | 프로젝트 빌드/테스트 자동화 |
| `controller-gen` | CRD 매니페스트 생성 |

---

## 핵심 정리

1. Kubernetes 기여는 SIG → 이슈 선택 → 포크 → PR → 리뷰 흐름
2. CLA 서명과 DCO (`git commit -s`)는 필수
3. `good first issue`와 `help wanted` 라벨로 시작
4. 코드 리뷰는 학습의 핵심 — 피드백을 환영하라
5. 커뮤니티 참여(Slack, 미팅)가 코드만큼 중요하다
