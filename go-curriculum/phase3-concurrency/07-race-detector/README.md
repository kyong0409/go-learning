# 레이스 디텍터 (Race Detector)

Go에 내장된 레이스 디텍터를 사용해 레이스 컨디션을 찾고 수정하는 방법을 배웁니다.

## 레이스 디텍터란?

Go 1.1부터 내장된 동적 분석 도구입니다. 프로그램 실행 중 동기화 없이
여러 고루틴이 같은 메모리에 접근하는 상황을 감지합니다.

**구현 방식**: ThreadSanitizer (TSan) 기반, C/C++ 생태계와 동일한 검사기

## 사용법

### 프로그램 실행 시
```bash
go run -race main.go
```

### 빌드 시
```bash
go build -race -o myapp
./myapp
```

### 테스트 시 (가장 중요!)
```bash
go test -race ./...
go test -race -count=10 ./...   # 반복 실행으로 간헐적 레이스 탐지
```

## 이 디렉터리의 파일

| 파일 | 설명 |
|------|------|
| `main.go` | 의도적 레이스 컨디션 예제 (4가지 패턴) |
| `main_fixed.go` | 수정된 버전 (`//go:build ignore` 태그로 일반 빌드 제외) |

### main.go 실행
```bash
# 레이스 감지 없이 실행 (결과가 불규칙)
go run main.go

# 레이스 디텍터 활성화 (경고 메시지 확인)
go run -race main.go
```

### main_fixed.go 실행
```bash
# build 태그 제거 후 실행
go run main_fixed.go
go run -race main_fixed.go   # 경고 없음 확인
```

## 레이스 디텍터 출력 해석

```
==================
WARNING: DATA RACE
Read at 0x00c0001b4010 by goroutine 7:       ← 읽기 위치와 고루틴
  main.incrementRacy()
      /path/main.go:30 +0x44

Previous write at 0x00c0001b4010 by goroutine 6:  ← 이전 쓰기
  main.incrementRacy()
      /path/main.go:30 +0x58

Goroutine 7 (running) created at:            ← 고루틴 생성 위치
  main.main()
      /path/main.go:55 +0x9c
==================
```

## 레이스 컨디션 수정 전략

| 문제 패턴 | 수정 방법 |
|-----------|-----------|
| 카운터/플래그 동시 접근 | `sync/atomic` |
| 슬라이스/맵 동시 쓰기 | `sync.Mutex` |
| 맵 동시 읽기/쓰기 | `sync.Map` |
| 루프 클로저 변수 캡처 | 루프 변수를 인자로 전달 |
| 고루틴 간 데이터 전달 | 채널 사용 |
| 초기화 중복 실행 | `sync.Once` |

## 레이스 디텍터 주의사항

- **성능 오버헤드**: CPU 5~15배, 메모리 5~10배 증가
- **프로덕션 배포 금지**: 개발/테스트 환경에서만 사용
- **모든 경로 커버 필요**: 실행되지 않은 코드의 레이스는 감지 불가
- **CI/CD 통합 권장**: `go test -race ./...`를 CI 파이프라인에 포함

## CI/CD 통합 예시 (GitHub Actions)

```yaml
- name: Race condition check
  run: go test -race -timeout 60s ./...
```
