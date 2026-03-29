// 패키지 선언
package fetcher

import (
	"context"
	"io"
	"net/http"
	"time"
)

// FetchResult는 단일 URL 요청의 결과를 나타냅니다.
type FetchResult struct {
	URL        string // 요청한 URL
	Body       string // 응답 본문 (에러 시 빈 문자열)
	StatusCode int    // HTTP 상태 코드 (에러 시 0)
	Error      error  // 요청 실패 시 에러
}

// FetchAll은 urls 목록을 maxConcurrency 제한 하에 병렬로 가져옵니다.
// 결과는 urls와 동일한 순서로 반환됩니다.
// 일부 URL이 실패해도 나머지 결과를 반환합니다 (부분 실패 허용).
// ctx가 취소되면 아직 시작하지 않은 요청은 건너뜁니다.
//
// 매개변수:
//   - ctx: 취소 및 타임아웃 제어
//   - urls: 가져올 URL 목록
//   - maxConcurrency: 동시에 실행할 최대 고루틴 수 (1 이상)
//
// 반환값: urls와 동일한 길이의 FetchResult 슬라이스
func FetchAll(ctx context.Context, urls []string, maxConcurrency int) []FetchResult {
	// TODO: 구현하세요
	// 1. results := make([]FetchResult, len(urls)) 로 결과 슬라이스 초기화
	// 2. errgroup.WithContext(ctx) 로 그룹 생성
	// 3. g.SetLimit(maxConcurrency) 로 동시성 제한
	// 4. 각 URL에 대해 g.Go(func() error { ... }) 로 고루틴 시작
	//    - 각 고루틴은 해당 인덱스의 results[i]에 결과를 저장
	//    - fetchOne(gCtx, url) 로 실제 요청 수행
	//    - 에러가 있어도 nil 반환 (부분 실패 허용)
	// 5. g.Wait() 로 모든 고루틴 완료 대기
	// 6. results 반환
	panic("구현 필요")
}

// FetchWithRetry는 실패 시 지수 백오프로 재시도합니다.
// 성공하면 FetchResult를 반환합니다.
// maxRetries 초과 시 마지막 에러를 담은 FetchResult를 반환합니다.
//
// 재시도 간격: 100ms → 200ms → 400ms → 800ms → ...
//
// 매개변수:
//   - ctx: 취소 및 타임아웃 제어 (대기 중에도 취소 가능)
//   - url: 가져올 URL
//   - maxRetries: 최초 시도 이후 재시도 횟수 (0이면 재시도 없음)
func FetchWithRetry(ctx context.Context, url string, maxRetries int) FetchResult {
	// TODO: 구현하세요
	// 1. delay := 100 * time.Millisecond 로 초기 대기 시간 설정
	// 2. 최대 maxRetries+1 번 반복:
	//    a. fetchOne(ctx, url) 로 요청
	//    b. 성공하면 (result.Error == nil) 즉시 반환
	//    c. 마지막 시도면 result 반환
	//    d. 아니면 select로 delay 대기 또는 ctx.Done() 처리
	//    e. delay *= 2 로 다음 대기 시간 두 배로
	// 3. 최종 result 반환
	panic("구현 필요")
}

// fetchOne은 단일 URL에 HTTP GET 요청을 수행합니다.
// 구현이 이미 제공되어 있습니다 - 수정하지 마세요.
func fetchOne(ctx context.Context, url string) FetchResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return FetchResult{URL: url, Error: err}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return FetchResult{URL: url, Error: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FetchResult{URL: url, Error: err}
	}

	return FetchResult{
		URL:        url,
		Body:       string(body),
		StatusCode: resp.StatusCode,
	}
}
