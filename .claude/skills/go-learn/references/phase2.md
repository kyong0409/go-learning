# Phase 2: 구조체, 인터페이스, 관용적 패턴

**기간:** 3~6주차
**목표:** Go의 타입 시스템과 인터페이스 기반 설계를 익힌다
**디렉토리:** `go-curriculum/phase2-structs-interfaces/`

---

## 레슨 목록

| # | 디렉토리 | 주제 | 핵심 개념 |
|---|---------|------|----------|
| 1 | `01-structs` | 구조체 | struct 정의, JSON 태그, 생성자 패턴 |
| 2 | `02-methods` | 메서드 | 값/포인터 리시버, 메서드 세트 |
| 3 | `03-interfaces` | 인터페이스 | 암묵적 충족, 타입 단언, io.Reader/Writer |
| 4 | `04-composition` | 컴포지션 | 임베딩, 상속 대체, 위임 |
| 5 | `05-error-handling` | 에러 처리 심화 | 에러 래핑(%w), errors.Is/As, 센티널 에러 |
| 6 | `06-defer-panic` | defer/panic | defer 스택, panic/recover, 정리 패턴 |
| 7 | `07-testing` | 테스트 | 테이블 주도 테스트, 벤치마크, go test |

## 레슨 파일 경로

```
go-curriculum/phase2-structs-interfaces/
├── 01-structs/THEORY.md         + main.go
├── 02-methods/THEORY.md         + main.go
├── 03-interfaces/THEORY.md      + main.go
├── 04-composition/THEORY.md     + main.go
├── 05-error-handling/THEORY.md  + main.go
├── 06-defer-panic/THEORY.md     + main.go
└── 07-testing/THEORY.md         + main.go + calc/
```

---

## 과제 목록

| # | 디렉토리 | 주제 | 난이도 | 핵심 |
|---|---------|------|--------|------|
| A1 | `assignments/a1-shape-interface` | Shape 인터페이스 | ★★☆☆☆ | 인터페이스, 다형성 |
| A2 | `assignments/a2-custom-error` | 은행 계좌 에러 | ★★★☆☆ | 커스텀 에러, 래핑 |
| A3 | `assignments/a3-json-config` | JSON 설정 파서 | ★★★☆☆ | 구조체, 검증, JSON |

## 과제 파일 경로

```
go-curriculum/phase2-structs-interfaces/assignments/
├── a1-shape-interface/
│   ├── README.md
│   ├── shape.go / shape_test.go
│   └── solution/
├── a2-custom-error/
│   ├── README.md
│   ├── bank.go / bank_test.go
│   └── solution/
└── a3-json-config/
    ├── README.md
    ├── config.go / config_test.go
    └── solution/
```

---

## 프로젝트

**프로젝트: HTTP JSON API 북마크 관리자**
- 디렉토리: `go-curriculum/phase2-structs-interfaces/project-bookmark-api/`
- 목표: 표준 라이브러리만으로 RESTful 북마크 API 구현
- 핵심 학습: net/http, JSON 마셜링, 핸들러 테스트
- 서브 패키지: handler/, model/

---

## 핵심 설계 원칙

Phase 2에서 강조할 Go의 철학:
- **"인터페이스를 받고, 구조체를 반환하라"**: 의존성 역전의 Go 버전
- **암묵적 인터페이스 충족**: implements 없이 메서드만 맞으면 자동 구현
- **작은 인터페이스**: io.Reader(1메서드)가 Go I/O 시스템의 기반
- **상속 없음**: 임베딩으로 코드 재사용, 하위 타입 관계 없음
- **에러는 값이다**: 에러를 래핑하고, 비교하고, 타입 단언할 수 있다
