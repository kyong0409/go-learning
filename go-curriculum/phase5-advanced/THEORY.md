# Phase 5: 고급 Go - 마이크로서비스, DevOps, 마스터리

## Phase 5 학습 목표

Phase 5는 Go를 실제 프로덕션 환경에서 사용하는 데 필요한 고급 기술을 다룹니다.

- gRPC로 고성능 마이크로서비스 간 통신 구현
- Cobra/Viper로 전문가 수준의 CLI 도구 개발
- pprof와 벤치마크로 성능 병목 찾기 및 최적화
- 퍼즈 테스트, 통합 테스트 등 고급 테스트 전략 적용
- client-go로 Kubernetes 클러스터와 상호작용

---

## Go가 인프라의 언어인 이유

CNCF(Cloud Native Computing Foundation) 프로젝트의 75% 이상이 Go로 작성되어 있습니다.

| 프로젝트 | 언어 | 설명 |
|---------|------|------|
| Kubernetes | Go | 컨테이너 오케스트레이션 |
| Docker / containerd | Go | 컨테이너 런타임 |
| Prometheus | Go | 모니터링 시스템 |
| etcd | Go | 분산 키-값 저장소 |
| Istio | Go | 서비스 메시 |
| Terraform | Go | 인프라 as Code |
| Helm | Go | K8s 패키지 관리자 |
| Hugo | Go | 정적 사이트 생성기 |

Go가 인프라 도구로 선택받는 이유:

```
1. 단일 바이너리 배포: 의존성 없이 복사만 하면 실행
2. 빠른 컴파일: 대규모 코드베이스도 수 초 내 빌드
3. 낮은 메모리 사용량: Python/Java 대비 5~10배 효율적
4. 뛰어난 동시성: 고루틴으로 수천 개 연결 처리
5. 정적 타입: 런타임 오류 최소화
6. 크로스 컴파일: linux/amd64, linux/arm64, darwin 등 단일 명령으로 빌드
```

```go
// 크로스 컴파일 예시
// GOOS=linux GOARCH=amd64 go build -o myapp-linux ./cmd/myapp
// GOOS=darwin GOARCH=arm64 go build -o myapp-mac ./cmd/myapp
// GOOS=windows GOARCH=amd64 go build -o myapp.exe ./cmd/myapp
```

---

## gRPC vs REST: 언제 무엇을 쓸지

### REST가 적합한 경우

- 외부 API (브라우저, 모바일 클라이언트, 서드파티)
- 단순한 CRUD 작업
- 캐싱이 중요한 경우 (HTTP GET 캐싱)
- 팀이 OpenAPI/Swagger에 익숙한 경우

### gRPC가 적합한 경우

- 마이크로서비스 간 내부 통신
- 스트리밍이 필요한 경우 (실시간 데이터, 로그 스트리밍)
- 다국어 환경 (Go 서버 ↔ Python 클라이언트 등)
- 성능이 중요한 경우 (직렬화 오버헤드 최소화)

### 성능 비교 (대략적인 수치)

```
JSON over HTTP/1.1  : 기준 (1x)
JSON over HTTP/2    : ~1.5x 빠름 (커넥션 재사용)
Protobuf over gRPC  : ~7~10x 빠름 (바이너리 직렬화 + HTTP/2)
```

### 실전 아키텍처 패턴

```
[클라이언트 (브라우저/앱)]
        │ HTTP/REST (JSON)
        ▼
[API Gateway / BFF]
        │ gRPC (Protobuf)
   ┌────┴────┐
   ▼         ▼
[서비스 A] [서비스 B]
   │              │ gRPC
   ▼         ▼
[서비스 C] [서비스 D]
```

---

## CLI 도구 설계 원칙

### Unix 철학

```
1. 한 가지 일을 잘 하라
2. 함께 작동하도록 만들어라 (파이프 지원)
3. 텍스트 스트림을 범용 인터페이스로 사용하라
4. 출력은 stdout, 에러는 stderr
5. 종료 코드: 0은 성공, 비-0은 실패
```

### 좋은 CLI의 특징

```go
// 좋은 CLI 설계
myapp deploy --env production --image myapp:v1.2.3 --dry-run

// 피해야 할 패턴
myapp deploy production myapp:v1.2.3 1  // 순서에 의존하는 인수
```

### Cobra/Viper 생태계

- **Cobra**: 명령어 구조, 플래그, 자동 완성, 도움말 생성
- **Viper**: 설정 파일, 환경 변수, 기본값 통합 관리
- 두 라이브러리는 함께 사용하도록 설계됨

---

## 성능 최적화의 Go 접근법

### 프로파일링 주도 최적화 (PDO)

"추측하지 말고 측정하라" - Rob Pike

```
1단계: 벤치마크로 현재 성능 측정
2단계: pprof로 병목 구간 식별
3단계: 가장 큰 병목 하나만 최적화
4단계: 다시 벤치마크로 개선 확인
5단계: 반복
```

### 고루틴 누수 방지

```go
// 나쁜 패턴: 고루틴이 영원히 실행됨
go func() {
    for {
        data := <-ch // ch가 닫히지 않으면 영원히 대기
        process(data)
    }
}()

// 좋은 패턴: context로 취소 가능
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        case data, ok := <-ch:
            if !ok {
                return
            }
            process(data)
        }
    }
}()
```

---

## Phase 5 학습 순서

```
01-grpc          → gRPC 기초, Protobuf, 4가지 통신 패턴
02-cobra-cli     → CLI 도구 설계, Cobra/Viper
03-profiling     → pprof, 벤치마크, 이스케이프 분석
04-testing-advanced → 목킹, 통합 테스트, 퍼즈 테스트
05-kubernetes-basics → client-go, Informer, Operator 패턴
```

### Python/Java 개발자를 위한 사전 지식 체크

| 개념 | Python 유사 | Java 유사 | Go |
|------|------------|-----------|-----|
| 인터페이스 기반 모킹 | unittest.mock | Mockito | testify/mock |
| 직렬화 | pickle/json | Jackson/Gson | encoding/json, protobuf |
| CLI 프레임워크 | Click/Typer | picocli | Cobra |
| 컨테이너 오케스트레이션 | - | - | client-go (Go가 원어민) |

Phase 5를 마치면 Kubernetes 생태계에 기여하거나, 프로덕션급 CLI 도구와 마이크로서비스를 독립적으로 설계·구현할 수 있는 수준이 됩니다.
