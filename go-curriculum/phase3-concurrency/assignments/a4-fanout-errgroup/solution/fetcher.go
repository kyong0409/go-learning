// 패키지 선언
// 참고 솔루션 - 풀기 전에 보지 마세요!
package fetcher

import (
	"context"
	"io"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

// FetchResult는 단일 URL 요청의 결과를 나타냅니다.
type FetchResult struct {
	URL        string
	Body       string
	StatusCode int
	Error      error
}

// FetchAll은 urls 목록을 maxConcurrency 제한 하에 병렬로 가져옵니다.
// 결과는 urls와 동일한 순서로 반환됩니다.
func FetchAll(ctx context.Context, urls []string, maxConcurrency int) []FetchResult {
	results := make([]FetchResult, len(urls))

	if len(urls) == 0 {
		return results
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	// errgroup으로 고루틴 그룹 생성 (컨텍스트 전파 포함)
	g, gCtx := errgroup.WithContext(ctx)
	// 동시 실행 수 제한
	g.SetLimit(maxConcurrency)

	for i, url := range urls {
		i, url := i, url // 루프 변수 캡처
		g.Go(func() error {
			// 그룹 컨텍스트(취소 전파)를 사용해 요청
			results[i] = fetchOne(gCtx, url)
			// 부분 실패 허용: 에러가 있어도 nil 반환 (그룹을 중단하지 않음)
			return nil
		})
	}

	// 모든 고루틴 완료 대기
	g.Wait() //nolint:errcheck // 항상 nil 반환
	return results
}

// FetchWithRetry는 실패 시 지수 백오프로 재시도합니다.
func FetchWithRetry(ctx context.Context, url string, maxRetries int) FetchResult {
	delay := 100 * time.Millisecond
	var result FetchResult

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result = fetchOne(ctx, url)

		// 성공하면 즉시 반환
		if result.Error == nil {
			return result
		}

		// 마지막 시도면 결과 반환
		if attempt == maxRetries {
			break
		}

		// 다음 재시도 전 대기 (ctx 취소 시 즉시 종료)
		select {
		case <-time.After(delay):
			// 다음 시도 진행
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result
		}

		delay *= 2 // 지수 백오프
	}

	return result
}

// fetchOne은 단일 URL에 HTTP GET 요청을 수행합니다.
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
