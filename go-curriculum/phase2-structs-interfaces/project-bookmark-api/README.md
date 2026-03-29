# 북마크 API

Go 표준 라이브러리만 사용한 RESTful 북마크 관리 HTTP API입니다.

## 학습 포인트

- Go 1.22+ `net/http` 향상된 ServeMux (메서드 패턴, 경로 변수)
- 인터페이스 기반 저장소 설계 (`model.Store`)
- JSON 인코딩/디코딩 (`encoding/json`)
- `httptest` 패키지를 이용한 핸들러 테스트
- 의존성 주입 패턴

## 프로젝트 구조

```
project-bookmark-api/
├── go.mod
├── main.go                  # 서버 진입점, 라우터 설정
├── model/
│   └── bookmark.go          # 데이터 모델, Store 인터페이스, 인메모리 구현
├── handler/
│   ├── handler.go           # HTTP 핸들러
│   └── handler_test.go      # 테이블 기반 핸들러 테스트
└── README.md
```

## 실행 방법

```bash
# 서버 실행 (포트 8080)
go run main.go

# 테스트 실행
go test ./...

# 상세 테스트 출력
go test -v ./...
```

## API 엔드포인트

| 메서드 | 경로 | 설명 |
|--------|------|------|
| `GET` | `/health` | 헬스체크 |
| `GET` | `/bookmarks` | 전체 북마크 목록 조회 |
| `POST` | `/bookmarks` | 새 북마크 생성 |
| `GET` | `/bookmarks/{id}` | 특정 북마크 조회 |
| `PUT` | `/bookmarks/{id}` | 북마크 수정 |
| `DELETE` | `/bookmarks/{id}` | 북마크 삭제 |

## 요청/응답 예제

### 헬스체크

```bash
curl http://localhost:8080/health
```
```json
{"status":"ok","service":"bookmark-api"}
```

### 북마크 생성

```bash
curl -X POST http://localhost:8080/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"title":"Go 공식 문서","url":"https://go.dev/doc","tags":["go","docs"]}'
```
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "Go 공식 문서",
    "url": "https://go.dev/doc",
    "tags": ["go", "docs"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 전체 목록 조회

```bash
curl http://localhost:8080/bookmarks
```
```json
{
  "success": true,
  "data": [
    {"id": 1, "title": "Go 공식 문서", "url": "https://go.dev/doc", ...}
  ]
}
```

### 특정 북마크 조회

```bash
curl http://localhost:8080/bookmarks/1
```

### 북마크 수정

```bash
curl -X PUT http://localhost:8080/bookmarks/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Go 언어 공식 문서"}'
```

### 북마크 삭제

```bash
curl -X DELETE http://localhost:8080/bookmarks/1
# 응답: 204 No Content
```

## 에러 응답 형식

```json
{
  "success": false,
  "message": "에러 메시지"
}
```

| 상태 코드 | 상황 |
|-----------|------|
| `400 Bad Request` | 잘못된 요청 (누락된 필드, 잘못된 ID 등) |
| `404 Not Found` | 북마크를 찾을 수 없음 |
| `500 Internal Server Error` | 서버 내부 오류 |
