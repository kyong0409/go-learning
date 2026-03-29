# Phase 7: 실전 오픈소스 기여 — kubernetes-sigs/agent-sandbox

**기간:** Phase 6 완료 후
**목표:** 실제 Kubernetes 오픈소스 프로젝트에 기여한다
**디렉토리:** `go-curriculum/phase7-contribution/`

---

## 레슨 목록

| # | 디렉토리 | 주제 | 핵심 개념 |
|---|---------|------|----------|
| 1 | `01-open-source-workflow` | K8s 오픈소스 기여 워크플로우 | SIG, CLA, DCO, 포크/PR 프로세스 |
| 2 | `02-agent-sandbox-overview` | Agent Sandbox 프로젝트 이해 | CRD 4종, gVisor 격리, 아키텍처 |
| 3 | `03-codebase-deep-dive` | 코드베이스 딥다이브 | api/, controllers/, 테스트 구조, 빌드 시스템 |
| 4 | `04-contribution-practice` | 기여 실습 | 로컬 환경, kind, 첫 PR 작성법 |
| 5 | `05-real-contribution` | 실전 기여 도전 | 이슈 분석, PR 전략, 리뷰 대응 |

## 레슨 파일 경로

```
go-curriculum/phase7-contribution/
├── 01-open-source-workflow/THEORY.md
├── 02-agent-sandbox-overview/THEORY.md
├── 03-codebase-deep-dive/THEORY.md
├── 04-contribution-practice/THEORY.md
└── 05-real-contribution/THEORY.md
```

---

## 과제 목록

| # | 디렉토리 | 주제 | 난이도 | 핵심 |
|---|---------|------|--------|------|
| A1 | `assignments/a1-local-setup` | 로컬 개발 환경 구축 | ★★★☆☆ | 포크, 빌드, kind 클러스터 |
| A2 | `assignments/a2-race-detector` | Race Detector 분석 리포트 | ★★★½☆ | go test -race, 레이스 분석 |
| A3 | `assignments/a3-e2e-test` | E2E 테스트 작성 | ★★★★☆ | E2E 프레임워크, 어노테이션 검증 |
| A4 | `assignments/a4-real-issue` | 실전 기여 PR 제출 | ★★★★★ | 캡스톤 — 실제 PR 머지 |

## 과제 파일 경로

```
go-curriculum/phase7-contribution/assignments/
├── a1-local-setup/README.md
├── a2-race-detector/README.md
├── a3-e2e-test/README.md
└── a4-real-issue/README.md
```

---

## 프로젝트

Phase 7에는 별도 프로젝트가 없습니다.
A4 (실전 기여 PR)가 캡스톤 프로젝트 역할을 합니다.

---

## 타겟 프로젝트: kubernetes-sigs/agent-sandbox

| 항목 | 값 |
|------|-----|
| SIG | SIG Apps |
| Slack | `#agent-sandbox` (kubernetes.slack.com) |
| 리포지토리 | https://github.com/kubernetes-sigs/agent-sandbox |
| 문서 | https://agent-sandbox.sigs.k8s.io |
| 미팅 | 격주 월요일 9:00 AM PT |
| 현재 버전 | v0.2.1 (pre-stable, v1alpha1) |
| 핵심 기술 | controller-runtime, client-go, gVisor |

### 추천 기여 진입점
- Issue #331 — Race detector 활성화 (good first issue)
- Issue #168 — E2E 테스트: pod-name 어노테이션 (help wanted)
- Issue #403 — 지속적 벤치마킹 (help wanted)
- Issue #166 — 웹사이트 리디자인 (help wanted)
