# 과제 A5: 토큰 버킷 속도 제한기 구현

## 과제 설명

`sync.Mutex`와 시간 계산을 사용하여 스레드 안전한 토큰 버킷(Token Bucket) 속도 제한기를 밑바닥부터 구현하세요.

## 배경: 토큰 버킷 알고리즘

- **버킷**에는 최대 `burst`개의 토큰이 들어갑니다.
- 초당 `rate`개의 토큰이 버킷에 채워집니다.
- 요청 하나를 처리할 때마다 토큰 하나를 소비합니다.
- 토큰이 없으면 요청을 거부하거나 토큰이 생길 때까지 기다립니다.

## 요구사항

### TokenBucket 구조체

```go
type TokenBucket struct { ... }

func NewTokenBucket(rate float64, burst int) *TokenBucket
func (tb *TokenBucket) Allow() bool
func (tb *TokenBucket) AllowN(n int) bool
func (tb *TokenBucket) Wait(ctx context.Context) error
func (tb *TokenBucket) SetRate(rate float64, burst int)
```

### KeyedRateLimiter 구조체

```go
type KeyedRateLimiter struct { ... }

func NewKeyedRateLimiter(rate float64, burst int) *KeyedRateLimiter
func (krl *KeyedRateLimiter) Allow(key string) bool
func (krl *KeyedRateLimiter) Wait(ctx context.Context, key string) error
```

### 메서드 설명

| 메서드 | 설명 |
|--------|------|
| `Allow()` | 토큰이 있으면 소비하고 `true`, 없으면 `false` 반환 (비블로킹) |
| `AllowN(n)` | 토큰 n개가 있으면 소비하고 `true`, 없으면 `false` 반환 |
| `Wait(ctx)` | 토큰이 생길 때까지 블로킹 대기. ctx 취소 시 에러 반환 |
| `SetRate(rate, burst)` | 속도와 버스트 크기를 동적으로 변경 |

### KeyedRateLimiter 설명

IP 주소 등 키별로 독립적인 속도 제한을 적용합니다.
- 각 키마다 독립적인 `TokenBucket`을 관리합니다.
- 처음 보는 키는 자동으로 새 버킷을 생성합니다.

## 실행 방법

```bash
cd a5-rate-limiter
go test -v .
go test -race -v .
go test -v -run TestGrade .
```

## 채점 기준 (100점)

| 항목 | 점수 | 설명 |
|------|------|------|
| Allow 기본 동작 | 20점 | 토큰 소비 및 고갈 처리 |
| AllowN | 10점 | N개 토큰 한 번에 소비 |
| 토큰 보충 | 20점 | 시간 경과에 따른 토큰 자동 보충 |
| Wait 블로킹 | 15점 | 토큰 생길 때까지 대기 |
| Context 취소 | 15점 | Wait 중 ctx 취소 처리 |
| SetRate | 10점 | 동적 속도 변경 |
| KeyedRateLimiter | 10점 | 키별 독립 제한 |

## 힌트

- 토큰 수를 `float64`로 관리하면 소수 토큰을 처리할 수 있습니다.
- `lastRefill time.Time`을 기록하고 `Allow()` 호출 시마다 경과 시간에 비례해 토큰을 보충하세요.
- `Wait()`는 폴링 루프 대신 정확한 대기 시간을 계산하세요: `waitDuration = (1 - tokens) / rate`
- `sync.Mutex`로 `tokens`와 `lastRefill` 접근을 보호하세요.
- `KeyedRateLimiter`는 `sync.Mutex`와 `map[string]*TokenBucket`을 사용하세요.
- 타이밍 테스트는 넉넉한 허용 오차(±50ms)가 있습니다.
