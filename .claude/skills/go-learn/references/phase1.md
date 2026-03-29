# Phase 1: Go 기초와 사고방식 전환

**기간:** 1~3주차
**목표:** Go의 기본 문법과 패러다임 전환을 체득한다
**디렉토리:** `go-curriculum/phase1-basics/`

---

## 레슨 목록

| # | 디렉토리 | 주제 | 핵심 개념 |
|---|---------|------|----------|
| 1 | `01-hello` | 첫 프로그램 | 프로그램 구조, fmt 패키지, go run/build |
| 2 | `02-variables` | 변수와 타입 | var, :=, 타입 시스템, 제로값, const/iota |
| 3 | `03-control-flow` | 제어 흐름 | for, if, switch, range (루프는 for 하나뿐) |
| 4 | `04-functions` | 함수 | 다중 반환, 가변 인자, 클로저, defer |
| 5 | `05-data-structures` | 자료구조 | 배열, 슬라이스(길이/용량), 맵 |
| 6 | `06-strings` | 문자열 | UTF-8, rune, strings 패키지, 불변성 |
| 7 | `07-pointers` | 포인터 | &, *, 값 전달 vs 참조, 산술 없음 |
| 8 | `08-errors` | 에러 처리 기초 | error 인터페이스, if err != nil, fmt.Errorf |
| 9 | `09-packages` | 패키지와 모듈 | go mod, 가시성(대소문자), 순환 의존 금지 |

## 레슨 파일 경로

```
go-curriculum/phase1-basics/
├── 01-hello/THEORY.md          + main.go
├── 02-variables/THEORY.md      + main.go
├── 03-control-flow/THEORY.md   + main.go
├── 04-functions/THEORY.md      + main.go
├── 05-data-structures/THEORY.md + main.go
├── 06-strings/THEORY.md        + main.go
├── 07-pointers/THEORY.md       + main.go
├── 08-errors/THEORY.md         + main.go
└── 09-packages/THEORY.md       + main.go + mathutil/
```

---

## 과제 목록

| # | 디렉토리 | 주제 | 난이도 | 핵심 |
|---|---------|------|--------|------|
| A1 | `assignments/a1-calculator` | 계산기 | ★☆☆☆☆ | 함수, 에러 처리 |
| A2 | `assignments/a2-slice-ops` | 슬라이스 유틸리티 | ★★☆☆☆ | 고차함수, 슬라이스 |
| A3 | `assignments/a3-word-counter` | 단어 빈도수 | ★★½☆☆ | 파일 I/O, 맵, 정렬 |

## 과제 파일 경로

```
go-curriculum/phase1-basics/assignments/
├── a1-calculator/
│   ├── README.md        # 문제 설명
│   ├── calculator.go    # 구현 파일 (TODO)
│   ├── calculator_test.go # 채점 테스트
│   └── solution/        # 참고 풀이
├── a2-slice-ops/
│   ├── README.md
│   ├── sliceutil.go
│   ├── sliceutil_test.go
│   └── solution/
└── a3-word-counter/
    ├── README.md
    ├── wordcount.go
    ├── wordcount_test.go
    ├── testdata/
    └── solution/
```

---

## 프로젝트

**프로젝트: CLI 할일 관리자**
- 디렉토리: `go-curriculum/phase1-basics/project-todo-cli/`
- 목표: 커맨드라인 할일 관리자 (add, list, done, delete)
- 핵심 학습: 패키지 구조, JSON 파일 I/O, CLI 인자 파싱
- 서드파티 라이브러리 없이 표준 라이브러리만 사용

---

## 패러다임 전환 포인트

Phase 1에서 특히 강조할 Go의 특성:
- **25개 예약어**: 단순함이 Go의 핵심 철학
- **에러는 반환값**: try/catch가 아닌 if err != nil
- **대소문자 가시성**: public/private 대신 첫 글자 대소문자
- **제로값**: 초기화하지 않아도 안전한 기본값
- **for만 존재**: while, do-while 없이 for 하나로 해결
