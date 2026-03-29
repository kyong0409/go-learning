# 과제 A1: 파일 검색 CLI 도구

**난이도**: ★★★☆☆
**예상 소요 시간**: 3-4시간

## 과제 설명

Cobra를 사용해 파일 시스템에서 패턴으로 파일을 검색하는 CLI 도구를 만드세요.

## 요구사항

### 기능 요구사항

1. **기본 검색**: 지정한 디렉터리에서 파일 이름 패턴으로 검색
2. **정규식 지원**: `--regex` 플래그로 정규식 패턴 사용
3. **재귀 탐색**: `--recursive` 플래그로 하위 디렉터리 포함
4. **출력 형식**: `--output table|json|count` 지원
5. **결과 제한**: `--max N`으로 최대 결과 수 제한
6. **파일 크기 필터**: `--min-size`, `--max-size` 바이트 단위
7. **확장자 필터**: `--ext .go,.md` 형식

### CLI 인터페이스

```bash
# 기본 검색 (현재 디렉터리, 재귀)
filesearch search "*.go"

# 정규식 검색
filesearch search --regex ".*_test\.go$" ./src

# 재귀 + JSON 출력
filesearch search --recursive --output json "main" .

# 확장자 필터 + 크기 제한
filesearch search --ext .go --max-size 10240 .

# 결과 수만 출력
filesearch search --output count "*.go" .
```

### 출력 형식

**table (기본)**:
```
경로                          크기      수정일
----                          ----      ------
./main.go                     1234      2024-01-15
./cmd/root.go                 567       2024-01-14
```

**json**:
```json
[
  {"path": "./main.go", "size": 1234, "modified": "2024-01-15T..."}
]
```

**count**:
```
12개 파일 발견
```

## 구현 파일

- `search.go`: 핵심 검색 로직 (`SearchFiles` 함수)
- `cmd/root.go`: Cobra 루트 커맨드
- `cmd/search.go`: search 서브커맨드

## 채점 기준 (100점)

| 항목 | 점수 |
|------|------|
| 기본 패턴 검색 | 20점 |
| 재귀 탐색 | 15점 |
| 정규식 지원 | 15점 |
| 출력 형식 3종 | 20점 |
| 파일 크기 필터 | 15점 |
| 확장자 필터 | 15점 |

## 실행 방법

```bash
cd a1-cli-tool
go mod tidy
go run main.go search "*.go" .
go test ./... -v
```
