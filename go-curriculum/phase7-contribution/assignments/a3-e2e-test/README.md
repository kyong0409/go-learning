# A3: E2E 테스트 작성 (★★★★☆)

## 목표

Issue #168을 기반으로, Sandbox가 Ready 상태일 때 pod-name 어노테이션이 존재하는지 검증하는 E2E 테스트를 작성한다.

## 배경

Sandbox 컨트롤러는 관리하는 Pod의 이름을 Sandbox 리소스의 어노테이션에 기록합니다. 이를 검증하는 E2E 테스트가 아직 없습니다.

## 요구사항

### 1. 기존 E2E 테스트 분석
- [ ] `test/e2e/` 디렉토리 구조 파악
- [ ] `test/e2e/framework/` 유틸리티 이해
- [ ] 기존 E2E 테스트 패턴 파악 (Setup → Create → Assert → Cleanup)

### 2. 테스트 작성
- [ ] Sandbox를 생성한다
- [ ] Ready 상태가 될 때까지 대기한다 (polling/watch)
- [ ] pod-name 어노테이션이 존재하는지 검증한다
- [ ] 어노테이션의 값이 실제 Pod 이름과 일치하는지 검증한다
- [ ] 정리 (Sandbox 삭제)

### 3. 테스트 실행
- [ ] kind 클러스터에서 테스트 통과 확인
- [ ] `-race` 플래그와 함께 실행

## 테스트 구조 가이드

```go
func TestSandboxPodNameAnnotation(t *testing.T) {
    // 1. Setup
    ctx := context.Background()
    sandbox := &v1alpha1.Sandbox{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-pod-name-annotation",
            Namespace: "default",
        },
        Spec: v1alpha1.SandboxSpec{
            // ... 최소한의 스펙
        },
    }

    // 2. Create
    // client.Create(ctx, sandbox)

    // 3. Wait for Ready
    // poll until sandbox.Status.Phase == "Ready"

    // 4. Assert annotation
    // assert sandbox has pod-name annotation
    // assert annotation value matches actual pod name

    // 5. Cleanup
    // client.Delete(ctx, sandbox)
}
```

## 평가 기준

| 항목 | 배점 |
|------|------|
| 기존 E2E 패턴 분석 | 15 |
| 테스트 코드 작성 | 35 |
| kind 클러스터에서 통과 | 25 |
| 에러 케이스 처리 (타임아웃 등) | 15 |
| 코드 스타일 (기존 코드와 일관성) | 10 |
| **합계** | **100** |
