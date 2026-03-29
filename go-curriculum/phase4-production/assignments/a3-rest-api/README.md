# 과제 A3: 도서 관리 REST API

## 목표

Go 1.22+ 라우팅을 사용하여 완전한 CRUD REST API를 구현합니다.

## 요구사항

### 엔드포인트

| 메서드 | 경로 | 설명 |
|--------|------|------|
| `GET` | `/api/books` | 목록 조회 (페이지네이션, 검색) |
| `POST` | `/api/books` | 도서 생성 |
| `GET` | `/api/books/{id}` | 단건 조회 |
| `PUT` | `/api/books/{id}` | 전체 수정 |
| `PATCH` | `/api/books/{id}` | 부분 수정 |
| `DELETE` | `/api/books/{id}` | 삭제 |

### Book 모델
```go
type Book struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Author      string    `json:"author"`
    ISBN        string    `json:"isbn"`
    PublishedAt time.Time `json:"published_at"`
    Price       float64   `json:"price"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### 입력 검증
- `title`: 필수, 1~200자
- `author`: 필수, 1~100자
- `isbn`: 선택, 10자 또는 13자 숫자
- `price`: 0 이상

### 검색 및 페이지네이션
```
GET /api/books?page=1&page_size=10&search=golang&author=김철수
```

### 오류 응답
```json
{"error": "설명", "code": "ERROR_CODE", "field": "필드명(검증 오류 시)"}
```

## 채점 기준

| 항목 | 배점 |
|------|------|
| CRUD 핸들러 구현 | 40점 |
| 입력 검증 | 20점 |
| 페이지네이션 + 검색 | 20점 |
| 오류 처리 | 20점 |
| **합계** | **100점** |

## 실행 방법

```bash
cd assignments/a3-rest-api
go test ./... -v
```

## 힌트

- Go 1.22+ `mux.HandleFunc("GET /api/books/{id}", handler)` 사용
- `r.PathValue("id")`로 경로 파라미터 추출
- `strconv.Atoi`로 ID 변환 후 검증
- `strings.Contains`로 제목/저자 검색
