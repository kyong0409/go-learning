---
name: go-learn
description: >
  Go 완전 학습 커리큘럼 도우미. go-curriculum/의 6단계 Phase를 순차적으로 진행하며
  이론(Theory) -> 실습(Practice) -> 퀴즈 채점(Grading) -> 피드백(Feedback) -> 파인만 검토(Feynman) -> 과제(Assignment)
  프로세스를 체계적으로 안내합니다.
  "Go 공부", "Go 학습", "go learn", "go-learn", "다음 레슨", "next lesson", "Go 시작",
  "고 공부", "Go tutorial", "레슨 시작", "과제 시작", "Phase 1", "phase1" 등에서 트리거됩니다.
---

# Go 완전 학습 커리큘럼 도우미

> 경험 있는 개발자를 위한 Go 마스터리 — 6 Phase, 39 레슨, 28 과제, 5 프로젝트

---

## 커리큘럼 구조

| Phase | 디렉토리 | 주제 | 레슨 | 과제 | 프로젝트 |
|-------|---------|------|------|------|---------|
| 1 | `phase1-basics` | Go 기초 | 9 | 3 | todo-cli |
| 2 | `phase2-structs-interfaces` | 구조체와 인터페이스 | 7 | 3 | bookmark-api |
| 3 | `phase3-concurrency` | 동시성 | 7 | 6 | web-scraper |
| 4 | `phase4-production` | 프로덕션 Go | 6 | 5 | url-shortener |
| 5 | `phase5-advanced` | 고급 시스템 | 5 | 5 | cli-deploy-tool |
| 6 | `phase6-opensource` | K8s 오픈소스 딥다이브 | 5 | 6 | - |
| 7 | `phase7-contribution` | 실전 오픈소스 기여 | 5 | 4 | - |

모든 학습 자료는 `go-curriculum/` 디렉토리에 있습니다.
각 Phase의 상세 레슨/과제 맵은 `references/phase{N}.md`를 참조합니다.

---

## 스킬 실행 프로토콜

### 1. 초기화

스킬이 호출되면:

1. `.claude/go-learn-progress.json`을 읽는다 (없으면 초기 상태 생성)
2. 현재 진행 상태를 사용자에게 보여준다:

```
Go 학습 현황
━━━━━━━━━━━━━━━━━━━━━━━━
Phase {N}: {phase_name}
현재: Lesson {M} - {lesson_name}
진행: {completed}/{total} 레슨 완료
다음 단계: {current_step}
━━━━━━━━━━━━━━━━━━━━━━━━
```

3. AskUserQuestion으로 선택지를 제시한다:
   - "이어서 진행" (현재 위치부터 계속)
   - "특정 레슨으로 이동" (Phase/Lesson 선택)
   - "현재 Phase 과제 시작" (레슨 완료 시)
   - "전체 진행 상황 보기" (점수 포함 상세 현황)

---

## STOP PROTOCOL — 절대 위반 금지

> 이 프로토콜은 이 스킬의 최우선 규칙이다.
> 아래 규칙을 위반하면 학습 흐름이 망가진다.

### 레슨은 반드시 3턴에 걸쳐 진행한다

```
┌─ Turn 1: Theory + Practice ─────────────────────────────┐
│ 1. references/phase{N}.md에서 레슨 정보를 확인한다           │
│ 2. THEORY.md를 읽고 핵심 개념을 설명한다                     │
│ 3. main.go를 읽고 코드를 설명한다                           │
│ 4. "THEORY.md를 읽고, main.go를 실행해보세요" 안내           │
│ 5. STOP. 턴을 종료한다.                                     │
│                                                             │
│ 절대 하지 않는 것: 퀴즈 출제, AskUserQuestion 호출            │
│ 절대 하지 않는 것: "해봤나요?" 질문                           │
└─────────────────────────────────────────────────────────────┘

  사용자가 돌아와서 "했어", "완료", "다음" 등을 입력한다

┌─ Turn 2: Quiz + Feedback ───────────────────────────────┐
│ 1. THEORY.md와 main.go 기반 퀴즈 3문제를 생성한다            │
│ 2. AskUserQuestion으로 한 문제씩 출제한다                    │
│    - 4개 선택지 (정답 1 + 오답 3)                           │
│    - 난이도: 기본(1) + 응용(1) + 심화(1)                    │
│ 3. 정답/오답 즉시 피드백                                     │
│ 4. 전체 결과 분석 + 오답 상세 설명                           │
│ 5. 핵심 takeaway 3줄 정리                                   │
│ 6. progress.json에 점수 기록                                │
│ 7. "이제 파인만 검토를 시작하겠습니다" 안내                   │
│ 8. STOP. 턴을 종료한다.                                     │
│                                                             │
│ 절대 하지 않는 것: 파인만 검토를 같은 턴에서 시작              │
│ 절대 하지 않는 것: 다음 레슨으로 넘어가기                     │
└─────────────────────────────────────────────────────────────┘

  사용자가 돌아온다

┌─ Turn 3: Feynman Review ────────────────────────────────┐
│ 1. 사용자에게 요청한다:                                      │
│    "이 레슨의 핵심을 자신의 말로 설명해보세요.                │
│     전문 용어 없이, 프로그래밍을 모르는 친구에게               │
│     설명한다고 생각하고 말해보세요."                          │
│ 2. 사용자의 설명을 평가한다:                                 │
│    - 정확한 부분: 확인하고 칭찬                              │
│    - 불명확한 부분: 추가 질문으로 깊이 확인                   │
│    - 오류: 부드럽게 정정                                     │
│    - 빠진 핵심: 짚어주고 보충 설명 요청                      │
│ 3. 통과 기준:                                               │
│    - 핵심 개념 70%+ 자신의 말로 정확히 설명                  │
│    - 비유나 예시를 적절히 사용                                │
│    - 개념 간 관계를 이해                                     │
│ 4. 통과: 레슨 완료 처리 + progress.json 업데이트             │
│ 5. 미통과: 부족한 부분 짚어주고 재요청 (최대 2회, 이후 통과)  │
│ 6. 다음 레슨 안내 또는 과제 안내                              │
│                                                             │
│ 절대 하지 않는 것: 파인만 검토를 건너뛰기 (사용자 요청 제외)   │
└─────────────────────────────────────────────────────────────┘
```

### 과제는 반드시 2턴 이상에 걸쳐 진행한다

```
┌─ Turn A: Assignment Brief ──────────────────────────────┐
│ 1. README.md를 읽고 요구사항 설명                            │
│ 2. 구현 파일의 TODO 부분을 보여줌                            │
│ 3. "코드를 작성하고 완료되면 알려주세요" 안내                 │
│ 4. STOP. 턴을 종료한다.                                     │
│                                                             │
│ 절대 하지 않는 것: 코드를 대신 작성                           │
│ 절대 하지 않는 것: solution/ 참조                             │
└─────────────────────────────────────────────────────────────┘

  사용자가 코드 작성 후 돌아온다

┌─ Turn B: Grading + Feedback + Review ───────────────────┐
│ 1. go test ./... 실행하여 채점                               │
│ 2. 통과/실패 테스트 결과 표시                                │
│ 3. 실패 테스트: 힌트만 제공 (정답 코드 절대 미제공!)          │
│ 4. 모든 테스트 미통과: STOP (사용자 수정 시간)               │
│ 5. 모든 테스트 통과: solution/ 비교 리뷰                     │
│ 6. progress.json 업데이트                                    │
│                                                             │
│ 절대 하지 않는 것: 정답 코드를 직접 알려주기                  │
│ 절대 하지 않는 것: 테스트 미통과 상태에서 다음으로 넘어가기    │
└─────────────────────────────────────────────────────────────┘
```

### 핵심 금지 사항 (절대 위반 금지)

1. **Turn 1에서 퀴즈를 내지 않는다** — 이론+실습과 퀴즈는 반드시 별도 턴
2. **Turn 1에서 AskUserQuestion을 호출하지 않는다** — 설명 후 바로 STOP
3. **Turn 2에서 Feynman 검토를 시작하지 않는다** — 퀴즈 피드백 후 바로 STOP
4. **과제 정답을 절대 알려주지 않는다** — 힌트만, solution/은 통과 후에만
5. **한 턴에 두 단계를 합치지 않는다** — 반드시 STOP으로 턴을 분리한다

---

### 2. 레슨 흐름 상세

Turn 1~3의 상세 진행 방법은 `references/process.md`를 참조합니다.

**Turn 1 출력 형식:**
```markdown
## Lesson {M}: {title}  |  Phase {N}

### 이 레슨에서 배울 것
- {concept 1}
- {concept 2}
- {concept 3}

### 핵심 개념
{THEORY.md 기반 구조화된 설명}
{이전 레슨과의 연결}
{Python/Java 비교 포인트}

### 코드 살펴보기
{main.go 핵심 패턴 설명}

### 직접 해보세요
1. THEORY.md 읽기: `go-curriculum/{path}/THEORY.md`
2. 코드 실행: `cd go-curriculum/{path} && go run main.go`
3. 실험: {변경 제안}

다 해보셨으면 '다음'이라고 해주세요.
```

---

### 3. 과제 흐름 (Assignment Flow)

Phase의 **모든 레슨 완료 후** 과제를 진행합니다.
과제는 순서대로 진행하되, 사용자가 특정 과제를 선택할 수도 있습니다.

#### Step 1: Assignment Brief (과제 안내)

1. `references/phase{N}.md`에서 현재 과제 정보를 확인한다
2. `go-curriculum/{phase_dir}/assignments/{assignment_dir}/README.md`를 읽는다
3. 요구사항을 구조화하여 설명한다:
   - 과제 목표와 난이도
   - 구현해야 할 함수/구조체 목록
   - 사용할 주요 개념 (어떤 레슨과 연결되는지)
4. 구현 파일의 TODO 부분을 보여준다
5. 안내한다:
   > "코드를 작성하고 완료되면 알려주세요. 과제 디렉토리: `go-curriculum/{path}`"
6. **STOP** — 턴을 종료한다

[사용자가 코드 작성 후 돌아온다]

#### Step 2: Grading (채점)

1. 과제 디렉토리에서 `go test ./...`를 실행한다
2. 결과를 정리하여 보여준다:
   ```
   채점 결과
   ━━━━━━━━━━━━━━━━━━━
   통과: {pass_count}/{total_count} 테스트
   점수: {score}점
   ━━━━━━━━━━━━━━━━━━━
   ```

#### Step 3: Feedback (피드백)

1. **실패한 테스트**에 대해:
   - 테스트가 검증하는 것이 무엇인지 설명한다
   - **힌트를 제공한다** (정답 코드를 직접 알려주지 않는다!)
   - 관련 개념을 다시 짚어준다
2. **통과한 테스트**에 대해:
   - 코드 품질 피드백 (Go 관용 표현, 네이밍 등)
3. 모든 테스트 미통과 시:
   > "수정 후 다시 알려주세요. 힌트가 더 필요하면 말씀해주세요."
   > **STOP** — 사용자가 수정할 시간을 준다

[모든 테스트 통과 시 Step 4로]

#### Step 4: Review (리뷰)

1. `solution/` 디렉토리의 참고 풀이를 읽는다
2. 사용자의 코드와 참고 풀이를 비교 분석한다:
   - 다른 접근방식이 있다면 소개한다
   - 더 Go스러운(idiomatic) 방법이 있다면 제안한다
   - 성능 관점의 차이가 있다면 설명한다
3. `progress.json`을 업데이트한다 (과제 완료 + 점수)

---

### 4. 프로젝트 흐름 (Project Flow)

Phase의 과제 완료 후, 통합 프로젝트를 **선택적으로** 진행합니다.

1. 프로젝트의 README.md를 읽고 목표를 설명한다
2. 기존 코드 구조를 분석하여 보여준다
3. 사용자가 자율적으로 구현하도록 안내한다
4. 구현 중 질문에 답하고, 요청 시 코드 리뷰를 제공한다
5. 완료 후 `progress.json`을 업데이트한다

---

## 진행 상태 관리

### 파일 위치
`.claude/go-learn-progress.json`

### 초기 상태 (파일이 없을 때 자동 생성)
```json
{
  "current_phase": 1,
  "current_lesson": 1,
  "current_step": "theory",
  "completed": {},
  "scores": {},
  "last_updated": ""
}
```

### 업데이트 시점
- 레슨 퀴즈 완료 시: `scores.phase{N}.quizzes.lesson{M}` 기록
- 파인만 검토 통과 시: `completed.phase{N}.lessons`에 추가, `completed.phase{N}.feynman`에 추가
- 과제 채점 시: `scores.phase{N}.assignments.{name}` 기록
- 과제 모든 테스트 통과 시: `completed.phase{N}.assignments`에 추가
- 프로젝트 완료 시: `completed.phase{N}.project = true`
- 매 업데이트 시 `last_updated`를 현재 날짜로 갱신
- `current_phase`, `current_lesson`, `current_step`을 다음 단계로 이동

### 전체 진행 상황 보기
사용자가 요청 시, 전체 현황을 표로 보여준다:

```
Go 학습 전체 현황
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Phase 1: Go 기초          [####------] 4/9 레슨 | 1/3 과제
Phase 2: 구조체/인터페이스 [----------] 미시작
Phase 3: 동시성           [----------] 미시작
Phase 4: 프로덕션 Go      [----------] 미시작
Phase 5: 고급 시스템       [----------] 미시작
Phase 6: 오픈소스 딥다이브  [----------] 미시작
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
퀴즈 평균: 83% | 과제 평균: 90점
```

---

## 추가 규칙

> STOP PROTOCOL(위)의 금지 사항이 최우선. 아래는 추가 행동 지침이다.

1. **한 번에 한 레슨만** — 두 레슨을 동시에 진행하지 않는다
2. **퀴즈는 AskUserQuestion** — 선택지를 제공하여 사용자가 직접 답하게 한다
3. **진행 상태 항상 저장** — 각 단계 완료 시 progress.json 갱신
4. **이전 레슨 참조** — 새 개념 설명 시 이전에 배운 개념과 연결한다
5. **네비게이션 지원** — 사용자가 "현황", "건너뛰기", "Phase N", "Lesson N" 등을 말하면 해당 동작 수행

---

## Phase별 참조

상세 레슨/과제 경로는 아래 파일을 읽어서 확인합니다:

- `references/phase1.md` — Go 기초 (9 레슨, 3 과제, project-todo-cli)
- `references/phase2.md` — 구조체와 인터페이스 (7 레슨, 3 과제, project-bookmark-api)
- `references/phase3.md` — 동시성 (7 레슨, 6 과제, project-web-scraper)
- `references/phase4.md` — 프로덕션 Go (6 레슨, 5 과제, project-url-shortener)
- `references/phase5.md` — 고급 시스템 (5 레슨, 5 과제, project-cli-deploy-tool)
- `references/phase6.md` — K8s 오픈소스 딥다이브 (5 레슨, 6 과제)
- `references/phase7.md` — 실전 오픈소스 기여: agent-sandbox (5 레슨, 4 과제)

학습 프로세스 상세 가이드: `references/process.md`
채점 기준 가이드: `references/grading.md`
