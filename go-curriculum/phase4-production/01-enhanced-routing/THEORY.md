# 01-enhanced-routing: Go 1.22+ 향상된 HTTP 라우팅

> Go 1.22에서 `net/http`의 `ServeMux`가 대폭 개선되었습니다. 서드파티 라우터 없이도 메서드 매칭, 경로 파라미터, 와일드카드를 사용할 수 있습니다. 이 기능은 Go 1.22(2024년 2월)에 도입되어 현재(Go 1.26)까지 안정적으로 자리 잡았습니다.

---

## 1. net/http ServeMux의 역사

### Go 1.0 ~ 1.21: 기본 ServeMux의 한계

```go
// Go 1.21 이전 — 메서드 구분 불가, 경로 파라미터 없음
mux := http.NewServeMux()
mux.HandleFunc("/api/users", handleUsers)       // GET과 POST 모두 처리됨
mux.HandleFunc("/api/users/", handleUserByID)   // 접두사 매칭만 가능

// 핸들러 내부에서 직접 분기해야 했음
func handleUsers(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        // 목록 조회
    case "POST":
        // 생성
    default:
        http.Error(w, "허용되지 않는 메서드", http.StatusMethodNotAllowed)
    }
}

// 경로 파라미터 추출도 직접 구현해야 했음
// /api/users/123 → "123" 추출
parts := strings.Split(r.URL.Path, "/")
id := parts[len(parts)-1]
```

이 불편함 때문에 `gorilla/mux`, `chi`, `gin` 등 서드파티 라우터가 필수처럼 사용되었습니다.

### Go 1.22: ServeMux 개선

```go
// Go 1.22+ — 메서드 패턴과 경로 와일드카드 지원
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users", handleListUsers)
mux.HandleFunc("POST /api/users", handleCreateUser)
mux.HandleFunc("GET /api/users/{id}", handleGetUser)
mux.HandleFunc("PUT /api/users/{id}", handleUpdateUser)
mux.HandleFunc("DELETE /api/users/{id}", handleDeleteUser)
```

---

## 2. 메서드 패턴

`"METHOD /path"` 형식으로 HTTP 메서드를 라우트에 포함합니다.

```go
// 형식: "METHOD /path"
mux.HandleFunc("GET /articles", listArticles)
mux.HandleFunc("POST /articles", createArticle)
mux.HandleFunc("PUT /articles/{id}", updateArticle)
mux.HandleFunc("DELETE /articles/{id}", deleteArticle)
mux.HandleFunc("PATCH /articles/{id}", patchArticle)
```

**메서드를 지정하지 않으면** 모든 메서드에 매칭됩니다 (하위 호환성 유지):

```go
// 메서드 무관 — 이전 방식과 동일
mux.HandleFunc("/legacy", legacyHandler)
```

**OPTIONS 자동 처리**: Go 1.22에서 메서드를 지정한 경로는 자동으로 `OPTIONS` 요청에 `Allow` 헤더를 포함한 응답을 반환합니다.

---

## 3. 경로 와일드카드

### 단일 세그먼트 와일드카드: `{name}`

중괄호로 감싼 이름은 URL 경로의 단일 세그먼트(슬래시 사이)를 캡처합니다.

```go
mux.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    // r.PathValue()로 값 추출 — Go 1.22+ 신규 메서드
    id := r.PathValue("id")
    fmt.Fprintf(w, "사용자 ID: %s", id)
})

// 여러 파라미터
mux.HandleFunc("GET /api/users/{userID}/posts/{postID}", func(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("userID")
    postID := r.PathValue("postID")
    fmt.Fprintf(w, "사용자 %s의 포스트 %s", userID, postID)
})
```

**`{id}`는 슬래시를 포함하지 않습니다.** `/api/users/123/extra`는 `GET /api/users/{id}`에 매칭되지 않습니다.

### 나머지 경로 와일드카드: `{name...}`

`...`을 붙이면 슬래시를 포함한 나머지 경로 전체를 캡처합니다.

```go
mux.HandleFunc("GET /files/{path...}", func(w http.ResponseWriter, r *http.Request) {
    // GET /files/images/logo.png → path = "images/logo.png"
    // GET /files/a/b/c/d.txt    → path = "a/b/c/d.txt"
    filePath := r.PathValue("path")
    fmt.Fprintf(w, "파일: %s", filePath)
})
```

---

## 4. r.PathValue() — 경로 파라미터 추출

`r.PathValue("name")`은 Go 1.22에서 추가된 `*http.Request`의 메서드입니다.

```go
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")

    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "유효하지 않은 ID", http.StatusBadRequest)
        return
    }

    // id를 사용한 비즈니스 로직
    user, err := findUser(id)
    if err != nil {
        http.Error(w, "사용자를 찾을 수 없습니다", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

---

## 5. 패턴 우선순위 규칙

여러 패턴이 동일한 경로에 매칭될 때 **더 구체적인 패턴이 우선**합니다.

```go
mux.HandleFunc("GET /api/users/me", handleCurrentUser)   // 더 구체적
mux.HandleFunc("GET /api/users/{id}", handleGetUser)     // 일반적

// GET /api/users/me → handleCurrentUser 호출 (구체적 패턴 우선)
// GET /api/users/42 → handleGetUser 호출
```

**우선순위 규칙 요약**:

| 우선순위 | 패턴 유형 | 예시 |
|---------|-----------|------|
| 1 (최고) | 고정 경로 | `/api/users/me` |
| 2 | 와일드카드 포함 | `/api/users/{id}` |
| 3 | 나머지 경로 와일드카드 | `/files/{path...}` |
| 4 (최저) | 접두사 (`/` 로 끝나는) | `/api/` |

---

## 6. http.Handler 인터페이스

Go HTTP의 핵심 추상입니다.

```go
// net/http 패키지에 정의된 인터페이스
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

구조체에 `ServeHTTP`를 구현하면 `http.Handler`가 됩니다:

```go
type ArticleHandler struct {
    db *Database
    logger *slog.Logger
}

func (h *ArticleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        h.list(w, r)
    case http.MethodPost:
        h.create(w, r)
    }
}

// 등록
mux.Handle("/api/articles", &ArticleHandler{db: db, logger: logger})
```

---

## 7. http.HandlerFunc 타입 어댑터

일반 함수를 `http.Handler`로 변환하는 타입입니다.

```go
// net/http 패키지 정의
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}

// mux.HandleFunc 내부적으로 이렇게 동작
mux.HandleFunc("GET /hello", myFunc)
// 동등: mux.Handle("GET /hello", http.HandlerFunc(myFunc))
```

---

## 8. 라우트 그룹핑: 서브 ServeMux 패턴

Go의 `ServeMux`에는 프레임워크처럼 `Group()` 메서드가 없습니다. 대신 `http.StripPrefix`와 별도의 `ServeMux`를 조합합니다.

```go
// v1 API 전용 서브 라우터
func newAPIv1Router() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /users", listUsers)
    mux.HandleFunc("POST /users", createUser)
    mux.HandleFunc("GET /users/{id}", getUser)
    return mux
}

// v2 API 전용 서브 라우터
func newAPIv2Router() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /users", listUsersV2)
    return mux
}

// 메인 라우터에 마운트
mainMux := http.NewServeMux()

// /api/v1/users → v1 라우터의 /users
mainMux.Handle("/api/v1/", http.StripPrefix("/api/v1", newAPIv1Router()))

// /api/v2/users → v2 라우터의 /users
mainMux.Handle("/api/v2/", http.StripPrefix("/api/v2", newAPIv2Router()))
```

**주의**: `http.StripPrefix`로 접두사를 제거한 경로가 서브 라우터에 전달됩니다.

---

## 9. 기존 프레임워크와의 비교

| 기능 | gorilla/mux (deprecated) | chi | gin | Go 1.22 ServeMux |
|------|--------------------------|-----|-----|-----------------|
| 메서드 라우팅 | O | O | O | O |
| 경로 파라미터 | O | O | O | O |
| 미들웨어 | O | O | O | 직접 구현 |
| 라우트 그룹 | O | O | O | StripPrefix 패턴 |
| `http.Handler` 호환 | O | O | X (독자 Context) | O |
| 외부 의존성 | O | O | O | X |

**결론**: Go 1.22+에서는 단순한 REST API라면 표준 라이브러리로 충분합니다. 미들웨어 체인이나 라우트 그룹이 복잡해지면 `chi`를 추가하는 것이 좋습니다. `gorilla/mux`는 2023년에 아카이브(deprecated)되었습니다.

> **현황 (Go 1.26, 2026년 2월 기준)**: Go 1.22 ServeMux 개선은 이제 충분히 검증된 기능입니다. Go 1.23~1.26에 걸쳐 추가 변경 없이 안정적으로 유지되고 있으며, 표준 라이브러리 HTTP 라우팅의 사실상 표준으로 자리 잡았습니다.

---

## 10. Python / Java 비교

### Flask vs Go ServeMux

```python
# Flask (Python)
@app.route("/api/users/<int:user_id>", methods=["GET"])
def get_user(user_id):
    return jsonify(find_user(user_id))
```

```go
// Go 1.22+
mux.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.Atoi(r.PathValue("id"))
    user := findUser(id)
    json.NewEncoder(w).Encode(user)
})
```

### Spring MVC vs Go ServeMux

```java
// Spring MVC (Java)
@RestController
@RequestMapping("/api/users")
public class UserController {
    @GetMapping("/{id}")
    public User getUser(@PathVariable Long id) {
        return userService.findById(id);
    }
}
```

```go
// Go — 어노테이션 없이 명시적 등록
mux.HandleFunc("GET /api/users/{id}", userHandler.Get)
```

Go의 라우팅은 프레임워크의 매직(reflection, 어노테이션)이 없고 명시적입니다. 코드를 읽으면 라우트가 어디에 등록되는지 바로 알 수 있습니다.
