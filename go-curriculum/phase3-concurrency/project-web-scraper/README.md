# 동시성 프로젝트: 병렬 웹 스크레이퍼

Go의 동시성 기능을 활용한 실전 웹 스크레이퍼 프로젝트입니다.

## 학습 목표

이 프로젝트를 통해 다음 동시성 패턴을 실전에서 배웁니다:

- **워커 풀**: `errgroup.SetLimit`으로 동시 HTTP 요청 수 제한
- **속도 제한**: `time.Ticker` 기반 토큰으로 초당 요청 수 제한
- **Context**: 타임아웃, 취소, OS 시그널 처리
- **채널**: 작업 큐, 결과 수집
- **sync**: `sync.Mutex`로 방문 URL 추적 보호

## 프로젝트 구조

```
project-web-scraper/
├── go.mod                  # 모듈 설정 (golang.org/x/net, x/sync)
├── main.go                 # CLI 진입점
├── scraper/
│   ├── scraper.go          # 핵심 스크레이퍼 로직
│   ├── parser.go           # HTML 파싱 (링크, 제목 추출)
│   └── scraper_test.go     # httptest 기반 테스트
└── README.md
```

## 설치 및 실행

```bash
cd project-web-scraper

# 의존성 설치
go mod tidy

# 빌드
go build -o scraper .

# 실행
./scraper -urls="https://example.com" -workers=3 -rate=5

# 또는 직접 실행
go run . -urls="https://golang.org,https://pkg.go.dev" -workers=5 -rate=10
```

## CLI 옵션

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-urls` | (필수) | 쉼표로 구분된 URL 목록 |
| `-workers` | `3` | 동시 워커 고루틴 수 |
| `-rate` | `5.0` | 초당 최대 요청 수 |
| `-timeout` | `30s` | 전체 작업 타임아웃 |
| `-depth` | `1` | 링크 크롤링 깊이 (최대 3) |

## 실행 예시

```bash
# 단일 URL
go run . -urls="https://example.com"

# 여러 URL, 높은 동시성
go run . -urls="https://golang.org,https://pkg.go.dev,https://go.dev" \
         -workers=10 -rate=20

# 느린 네트워크 환경
go run . -urls="https://example.com" -workers=2 -rate=1 -timeout=60s
```

## 테스트 실행

```bash
# 일반 테스트 (httptest 서버 사용, 외부 네트워크 불필요)
go test ./scraper/...

# 레이스 디텍터 포함
go test -race ./scraper/...

# 상세 출력
go test -v ./scraper/...

# 커버리지
go test -cover ./scraper/...
```

## 핵심 동시성 설계

### 워커 풀 (errgroup.SetLimit)

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(cfg.Workers) // 최대 N개 동시 실행

for _, url := range urls {
    url := url
    g.Go(func() error {
        return scrapeOne(ctx, url)
    })
}
g.Wait()
```

### 속도 제한 (time.Ticker)

```go
interval := time.Duration(float64(time.Second) / cfg.RatePerSec)
ticker := time.NewTicker(interval)

// 각 요청 전에 토큰 획득
select {
case <-ticker.C:      // 속도 제한 통과
case <-ctx.Done():    // 취소 시 즉시 반환
    return ctx.Err()
}
```

### 취소 체인 (Context)

```
context.Background()
    └── WithTimeout(30s)          ← 전체 타임아웃
          └── errgroup.WithContext ← 에러 시 자동 취소
                └── scrapeOne(ctx) ← HTTP 요청 취소 가능
```

## 개선 아이디어

이 프로젝트를 확장해보세요:

1. **결과 저장**: CSV/JSON으로 결과 파일 저장
2. **robots.txt 준수**: robots.txt 파싱 후 허용된 경로만 크롤링
3. **재시도 로직**: 실패한 요청 지수 백오프로 재시도
4. **도메인 제한**: 동일 도메인만 크롤링
5. **진행률 표시**: `sync/atomic`으로 진행률 카운터
6. **캐싱**: 이미 방문한 URL 영구 저장
