# 과제 A2: HTTP 부하 테스트 도구

**난이도**: ★★★★☆
**예상 소요 시간**: 4-5시간

## 과제 설명

동시 HTTP 요청을 전송하고 실시간 통계를 수집하는 부하 테스트 도구를 만드세요.

## 요구사항

### 기능 요구사항

1. **동시 요청**: 설정 가능한 동시성(concurrency) 수준으로 HTTP 요청 전송
2. **실행 시간 제어**: 요청 수(`-n`) 또는 실행 시간(`-d 30s`)으로 종료 조건 지정
3. **실시간 통계**: 매 초마다 현재 RPS, 평균 지연시간 출력
4. **지연시간 백분위수**: 테스트 완료 후 P50, P90, P95, P99 출력
5. **오류율**: HTTP 4xx/5xx 및 연결 오류 집계
6. **우아한 종료**: Ctrl+C로 즉시 종료 후 집계 결과 출력

### 출력 예시

```
부하 테스트 시작: GET http://localhost:8080/
동시성: 10 | 요청 수: 1000

진행 중: 523개 완료 | 87.2 RPS | 평균 지연: 114ms | 오류: 2 (0.4%)

=== 최종 결과 ===
총 요청:      1000개
성공:         998개 (99.8%)
실패:         2개 (0.2%)
총 시간:      11.46초
평균 RPS:     87.2
지연시간:
  P50:  98ms
  P90:  187ms
  P95:  234ms
  P99:  412ms
  최소: 23ms
  최대: 678ms
```

### 인터페이스

```go
// LoadTester 구조체와 Run 메서드를 구현하세요.
type LoadTester struct { ... }
func NewLoadTester(config Config) *LoadTester
func (lt *LoadTester) Run(ctx context.Context) (*Result, error)

// Stats 구조체로 통계를 수집하세요.
type Stats struct { ... }
func (s *Stats) Record(latency time.Duration, statusCode int, err error)
func (s *Stats) Percentile(p float64) time.Duration
func (s *Stats) Summary() Result
```

## 구현 파일

- `loadtest.go`: LoadTester 구조체와 Run 로직
- `stats.go`: 통계 수집 및 백분위수 계산
- `main.go`: CLI 진입점 (선택)

## 채점 기준 (100점)

| 항목 | 점수 |
|------|------|
| 동시 요청 실행 | 25점 |
| 요청 수 / 시간 종료 | 15점 |
| 통계 수집 (카운터) | 20점 |
| 백분위수 계산 | 25점 |
| 오류율 집계 | 15점 |

## 실행 방법

```bash
cd a2-load-tester
go mod tidy
go test ./... -v
go test -bench=. -benchmem
```
