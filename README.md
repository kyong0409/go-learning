# Go 완전 학습 커리큘럼

경험 있는 개발자를 위한 Go 마스터리 — 6 Phase, 39 레슨, 28 과제, 5 프로젝트

## 구조

```
go-curriculum/          # 학습 자료 (이론, 예제, 과제, 풀이)
.claude/
  skills/
    go-learn/           # 학습 도우미 스킬
    feynman-learning/   # 파인만 테크닉 스킬
  go-learn-progress.json # 진행 상태 추적
```

## 커리큘럼 개요

| Phase | 주제 | 기간 | 레슨 | 과제 | 프로젝트 |
|-------|------|------|------|------|---------|
| 1 | Go 기초 | 1-3주 | 9 | 3 | CLI 할일 관리자 |
| 2 | 구조체와 인터페이스 | 3-6주 | 7 | 3 | HTTP 북마크 API |
| 3 | 동시성 | 6-10주 | 7 | 6 | 웹 스크래퍼 |
| 4 | 프로덕션 Go | 10-16주 | 6 | 5 | URL 단축기 |
| 5 | 고급 시스템 | 16-24주 | 5 | 5 | CLI 배포 도구 |
| 6 | K8s 오픈소스 딥다이브 | 24주+ | 5 | 6 | - |

## 학습 프로세스

각 레슨은 아래 흐름으로 진행됩니다:

```
Theory (이론)  -->  Practice (실습)  -->  Quiz (퀴즈)
     |                                       |
     v                                       v
Feedback (피드백)  <--  Feynman Review (파인만 검토)
```

1. **Theory** — THEORY.md를 읽고 핵심 개념을 학습
2. **Practice** — main.go를 실행하며 코드 체험
3. **Quiz** — 3문제 퀴즈로 이해도 확인
4. **Feedback** — 오답 분석 및 핵심 정리
5. **Feynman Review** — 자신의 말로 설명하여 깊은 이해 검증
6. **Assignment** — 과제 코드 작성 후 `go test`로 채점

## 사용법

Claude Code에서 `/go-learn` 또는 "Go 학습 시작"으로 학습 도우미를 호출합니다.

```bash
# 과제 채점
cd go-curriculum/phase1-basics/assignments/a1-calculator
go test ./... -v

# 전체 채점
bash go-curriculum/grade.sh
bash go-curriculum/grade.sh phase1   # Phase별
```

## 난이도 분포

Phase 1-2는 기초~중급, Phase 3부터 본격적으로 어려워집니다.

```
Phase 1  ★~★★½     입문
Phase 2  ★★~★★★    초중급
Phase 3  ★★★~★★★★★ 중상급 (가장 가파른 곡선)
Phase 4  ★★★~★★★★½ 중상급
Phase 5  ★★★~★★★★★ 상급
Phase 6  ★★★★~★★★★★ 최상급
```
