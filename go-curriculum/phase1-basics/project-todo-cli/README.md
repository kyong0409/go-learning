# Todo CLI - Go 학습 프로젝트

Phase 1 학습 내용을 적용한 간단한 CLI 기반 할 일 관리 도구입니다.

## 학습 목표

이 프로젝트를 통해 다음을 실습합니다:

- 패키지 구조 설계 (`main` 패키지 + `todo` 패키지)
- 구조체(struct)와 메서드
- JSON 파일 입출력 (`encoding/json`)
- 에러 처리 패턴
- 커맨드라인 인자 파싱 (`os.Args`)
- 슬라이스 조작

## 빌드 및 실행

```bash
# 디렉토리 이동
cd phase1-basics/project-todo-cli

# 직접 실행 (빌드 없이)
go run main.go help

# 바이너리 빌드
go build -o todo .

# 테스트 실행
go test ./todo/ -v
```

## 사용법

```bash
# 할 일 추가
go run main.go add "Go 언어 공부하기"
go run main.go add "프로젝트 만들기"

# 목록 보기
go run main.go list

# 완료된 항목만 보기
go run main.go list --done

# 미완료 항목만 보기
go run main.go list --pending

# 완료 처리 (ID 사용)
go run main.go done 1

# 항목 삭제
go run main.go delete 2

# 제목 수정
go run main.go update 3 "Go 언어 마스터하기"

# 완료된 항목 모두 삭제
go run main.go clear-done
```

## 실행 예시

```
$ go run main.go list
=== Todo 목록 ===
  [ ] [  3] 제어 흐름 (for, if, switch) 학습        2026-03-03
  [ ] [  4] 함수와 클로저 이해하기                   2026-03-04
  [x] [  1] Go 기초 문법 학습                        2026-03-01
  [x] [  2] 변수와 타입 예제 실습                    2026-03-02

전체: 4 | 완료: 2 | 미완료: 2

$ go run main.go add "포인터 학습"
추가됨: [8] 포인터 학습

$ go run main.go done 3
완료됨: [3] 제어 흐름 (for, if, switch) 학습
```

## 프로젝트 구조

```
project-todo-cli/
├── main.go          # CLI 진입점, 명령 파싱
├── go.mod           # Go 모듈 파일
├── todos.json       # 데이터 저장 파일 (자동 생성)
├── README.md        # 이 파일
└── todo/
    ├── todo.go      # Todo 구조체, CRUD 로직, JSON 저장/불러오기
    └── todo_test.go # 단위 테스트
```

## 데이터 형식 (todos.json)

```json
{
  "items": [
    {
      "id": 1,
      "title": "Go 기초 문법 학습",
      "done": true,
      "created_at": "2026-03-01T09:00:00Z",
      "completed_at": "2026-03-02T18:30:00Z"
    }
  ],
  "next_id": 2
}
```

## 확장 아이디어

1. 우선순위 필드 추가 (`priority: high/medium/low`)
2. 마감일 설정 기능
3. 태그 기능
4. 대화형 TUI (터미널 UI)
5. 원격 API 연동
