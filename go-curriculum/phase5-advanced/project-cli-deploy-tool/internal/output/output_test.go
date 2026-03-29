// internal/output/output_test.go
// 출력 포맷터 테스트
package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/curriculum/deploy-tool/internal/deployer"
)

// ============================================================
// PrintDeployResult 테스트
// ============================================================

func TestPrintDeployResult_Table(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, "table")

	result := &deployer.DeployResult{
		AppName:     "myapp",
		Image:       "nginx:latest",
		ContainerID: "abc123def456",
		Status:      "running",
		Port:        8080,
		DeployedAt:  time.Now(),
		Message:     "컨테이너 시작됨",
	}

	if err := p.PrintDeployResult(result); err != nil {
		t.Fatalf("PrintDeployResult 실패: %v", err)
	}

	output := buf.String()

	// 필수 내용 포함 확인
	assertContains(t, output, "myapp", "앱 이름")
	assertContains(t, output, "nginx:latest", "이미지")
	assertContains(t, output, "abc123def456", "컨테이너 ID")
	assertContains(t, output, "8080", "포트")
}

func TestPrintDeployResult_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, "json")

	deployedAt := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	result := &deployer.DeployResult{
		AppName:     "testapp",
		Image:       "redis:7",
		ContainerID: "deadbeef1234",
		Status:      "running",
		DeployedAt:  deployedAt,
	}

	if err := p.PrintDeployResult(result); err != nil {
		t.Fatalf("PrintDeployResult JSON 실패: %v", err)
	}

	// JSON 파싱 유효성 확인
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON 파싱 실패: %v\n출력: %s", err, buf.String())
	}

	if parsed["AppName"] != "testapp" {
		t.Errorf("JSON AppName: 기대 testapp, 실제 %v", parsed["AppName"])
	}
	if parsed["Image"] != "redis:7" {
		t.Errorf("JSON Image: 기대 redis:7, 실제 %v", parsed["Image"])
	}
}

func TestPrintDeployResult_NoPort(t *testing.T) {
	// Port = 0이면 포트 행이 출력되지 않아야 합니다.
	var buf bytes.Buffer
	p := NewPrinter(&buf, "table")

	result := &deployer.DeployResult{
		AppName:     "noport",
		Image:       "busybox",
		ContainerID: "abc123",
		Status:      "running",
		Port:        0, // 포트 없음
		DeployedAt:  time.Now(),
	}

	if err := p.PrintDeployResult(result); err != nil {
		t.Fatalf("예상치 못한 오류: %v", err)
	}

	output := buf.String()
	// 포트 행이 없어야 함
	if strings.Contains(output, "포트:") {
		t.Error("포트가 0인 경우 포트 행이 출력되면 안 됩니다")
	}
}

// ============================================================
// PrintStatus 테스트
// ============================================================

func TestPrintStatus_Table(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, "table")

	statuses := []*deployer.AppStatus{
		{
			AppName:     "app1",
			Image:       "nginx:1.25",
			ContainerID: "aaa111",
			Status:      "running",
			Restarts:    0,
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		},
		{
			AppName:     "app2",
			Image:       "redis:7",
			ContainerID: "bbb222",
			Status:      "stopped",
			Restarts:    3,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
	}

	if err := p.PrintStatus(statuses); err != nil {
		t.Fatalf("PrintStatus 실패: %v", err)
	}

	output := buf.String()

	assertContains(t, output, "app1", "앱1 이름")
	assertContains(t, output, "app2", "앱2 이름")
	assertContains(t, output, "nginx:1.25", "앱1 이미지")
	assertContains(t, output, "Running", "앱1 상태")
	assertContains(t, output, "Stopped", "앱2 상태")
	assertContains(t, output, "총 2개 앱", "총 앱 수")
}

func TestPrintStatus_Empty(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, "table")

	if err := p.PrintStatus([]*deployer.AppStatus{}); err != nil {
		t.Fatalf("PrintStatus 빈 목록 실패: %v", err)
	}

	output := buf.String()
	assertContains(t, output, "0", "0개 앱")
}

func TestPrintStatus_JSON(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, "json")

	statuses := []*deployer.AppStatus{
		{AppName: "myapp", Status: "running", ContainerID: "abc"},
	}

	if err := p.PrintStatus(statuses); err != nil {
		t.Fatalf("PrintStatus JSON 실패: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON 파싱 실패: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("JSON 항목 수: 기대 1, 실제 %d", len(parsed))
	}
	if parsed[0]["AppName"] != "myapp" {
		t.Errorf("JSON AppName: 기대 myapp, 실제 %v", parsed[0]["AppName"])
	}
}

// ============================================================
// 헬퍼 함수 테스트
// ============================================================

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{90 * time.Minute, "1h 30m"},
		{50 * time.Hour, "2d 2h"},
	}

	for _, tc := range testCases {
		result := formatDuration(tc.duration)
		if result != tc.expected {
			t.Errorf("formatDuration(%v) = %q; 기대: %q", tc.duration, result, tc.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	testCases := []struct {
		input    string
		n        int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long string", 10, "this is..."},
	}

	for _, tc := range testCases {
		result := truncate(tc.input, tc.n)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q; 기대: %q", tc.input, tc.n, result, tc.expected)
		}
	}
}

// ============================================================
// 헬퍼 assertion
// ============================================================

func assertContains(t *testing.T, s, substr, label string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("%s: %q를 포함해야 합니다\n실제 출력:\n%s", label, substr, s)
	}
}
