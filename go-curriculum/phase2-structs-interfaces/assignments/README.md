# Phase 2 과제 (Assignments)

Go Phase 2 - 구조체, 인터페이스, 관용적 패턴 학습 과제입니다.

## 과제 목록

| 과제 | 주제 | 핵심 개념 |
|------|------|-----------|
| [a1-shape-interface](./a1-shape-interface/) | Shape 인터페이스와 다형성 | 인터페이스, 메서드, 다형성 |
| [a2-custom-error](./a2-custom-error/) | 은행 계좌 에러 처리 | 커스텀 에러, 에러 래핑, errors.Is/As |
| [a3-json-config](./a3-json-config/) | JSON 설정 파일 파서 | JSON 태그, 유효성 검사, 환경변수 |

## 진행 방법

1. 각 과제 폴더의 `README.md`를 읽고 요구사항을 파악합니다.
2. `*.go` 파일에서 `// TODO:` 주석을 찾아 구현합니다.
3. 테스트를 실행하여 채점 결과를 확인합니다.

```bash
# 특정 과제 테스트
cd a1-shape-interface
go test -v

# 전체 과제 테스트
go test ./...
```

4. 막히면 `solution/` 폴더의 참고 풀이를 확인하세요.

## 채점 기준

각 테스트 파일은 실행 후 채점 결과를 출력합니다:

```
=== 채점 결과 ===
통과: 18/20
점수: 90/100
```

## 학습 순서 권장

`a1` → `a2` → `a3` 순서로 진행하면 난이도가 점진적으로 높아집니다.
