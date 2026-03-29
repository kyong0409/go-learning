// internal/output/output.go
// 다양한 출력 형식 지원 (table, JSON, YAML)
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/curriculum/deploy-tool/internal/deployer"
	"gopkg.in/yaml.v3"
)

// ============================================================
// Printer
// ============================================================

// Printer는 출력 형식에 맞게 데이터를 포맷합니다.
type Printer struct {
	w      io.Writer // 출력 대상
	format string    // table | json | yaml
}

// NewPrinter는 Printer 생성자입니다.
func NewPrinter(w io.Writer, format string) *Printer {
	if format == "" {
		format = "table"
	}
	return &Printer{w: w, format: format}
}

// ============================================================
// 배포 결과 출력
// ============================================================

// PrintDeployResult는 배포 결과를 출력합니다.
func (p *Printer) PrintDeployResult(result *deployer.DeployResult) error {
	switch p.format {
	case "json":
		return p.printJSON(result)
	case "yaml":
		return p.printYAML(result)
	default:
		return p.printDeployResultTable(result)
	}
}

func (p *Printer) printDeployResultTable(result *deployer.DeployResult) error {
	fmt.Fprintf(p.w, "\n배포 완료!\n")
	fmt.Fprintf(p.w, "%s\n", strings.Repeat("─", 40))

	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "앱 이름:\t%s\n", result.AppName)
	fmt.Fprintf(tw, "이미지:\t%s\n", result.Image)
	fmt.Fprintf(tw, "컨테이너 ID:\t%s\n", result.ContainerID)
	fmt.Fprintf(tw, "상태:\t%s\n", colorStatus(result.Status))
	if result.Port > 0 {
		fmt.Fprintf(tw, "포트:\t%d\n", result.Port)
	}
	fmt.Fprintf(tw, "배포 시각:\t%s\n", result.DeployedAt.Format(time.RFC3339))
	if result.Message != "" {
		fmt.Fprintf(tw, "메시지:\t%s\n", result.Message)
	}
	return tw.Flush()
}

// ============================================================
// 상태 목록 출력
// ============================================================

// PrintStatus는 앱 상태 목록을 출력합니다.
func (p *Printer) PrintStatus(statuses []*deployer.AppStatus) error {
	switch p.format {
	case "json":
		return p.printJSON(statuses)
	case "yaml":
		return p.printYAML(statuses)
	default:
		return p.printStatusTable(statuses)
	}
}

func (p *Printer) printStatusTable(statuses []*deployer.AppStatus) error {
	tw := tabwriter.NewWriter(p.w, 0, 0, 3, ' ', 0)

	// 헤더
	fmt.Fprintln(tw, "NAME\tIMAGE\tSTATUS\tCONTAINER ID\tRESTARTS\tUPTIME")
	fmt.Fprintln(tw, "----\t-----\t------\t------------\t--------\t------")

	for _, s := range statuses {
		uptime := s.Uptime
		if uptime == "" && !s.CreatedAt.IsZero() {
			uptime = formatDuration(time.Since(s.CreatedAt))
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n",
			s.AppName,
			truncate(s.Image, 30),
			colorStatus(s.Status),
			s.ContainerID,
			s.Restarts,
			uptime,
		)
	}

	if err := tw.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(p.w, "\n총 %d개 앱\n", len(statuses))
	return nil
}

// ============================================================
// JSON / YAML 출력
// ============================================================

func (p *Printer) printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 변환 실패: %w", err)
	}
	fmt.Fprintln(p.w, string(data))
	return nil
}

func (p *Printer) printYAML(v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("YAML 변환 실패: %w", err)
	}
	fmt.Fprint(p.w, string(data))
	return nil
}

// ============================================================
// 헬퍼 함수
// ============================================================

// colorStatus는 상태에 따라 (터미널 컬러 없이) 표시 문자열을 반환합니다.
// 실제 구현에서는 github.com/fatih/color 등을 사용할 수 있습니다.
func colorStatus(status string) string {
	switch strings.ToLower(status) {
	case "running":
		return "Running"
	case "stopped", "exited":
		return "Stopped"
	case "failed":
		return "Failed"
	default:
		return status
	}
}

// truncate는 문자열을 최대 n자로 자릅니다.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

// formatDuration은 시간 간격을 사람이 읽기 쉬운 형태로 변환합니다.
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	case minutes > 0:
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}
