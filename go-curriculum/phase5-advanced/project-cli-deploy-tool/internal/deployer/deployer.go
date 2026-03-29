// internal/deployer/deployer.go
// Deployer 인터페이스와 공통 타입 정의
package deployer

import (
	"io"
	"time"
)

// ============================================================
// 요청/응답 타입
// ============================================================

// DeployRequest는 배포 요청 파라미터를 담습니다.
type DeployRequest struct {
	AppName  string            // 애플리케이션 이름
	Image    string            // Docker 이미지 (예: nginx:latest)
	Port     int               // 호스트에 노출할 포트 (0 = 없음)
	Replicas int               // 실행할 컨테이너 수
	Env      map[string]string // 환경 변수
	DryRun   bool              // 실제 배포 없이 계획만 반환
}

// DeployResult는 배포 결과를 담습니다.
type DeployResult struct {
	AppName     string    // 배포된 앱 이름
	Image       string    // 사용된 이미지
	ContainerID string    // 생성된 컨테이너 ID (짧은 형식)
	Status      string    // running | failed | pending
	Port        int       // 노출된 포트
	DeployedAt  time.Time // 배포 완료 시각
	Message     string    // 추가 메시지
}

// AppStatus는 애플리케이션의 현재 상태를 담습니다.
type AppStatus struct {
	AppName     string    // 앱 이름
	Image       string    // 현재 이미지
	ContainerID string    // 컨테이너 ID
	Status      string    // running | stopped | exited
	Port        int       // 노출된 포트
	Restarts    int       // 재시작 횟수
	CreatedAt   time.Time // 컨테이너 생성 시각
	Uptime      string    // 가동 시간 (예: "2h 30m")
}

// LogOptions는 로그 조회 옵션을 담습니다.
type LogOptions struct {
	Follow bool   // 실시간 스트리밍
	Tail   int    // 마지막 N줄
	Since  string // 특정 시간 이후 (예: 1h, 30m)
}

// ============================================================
// Deployer 인터페이스
// ============================================================

// Deployer는 배포 작업을 처리하는 인터페이스입니다.
// Docker, Kubernetes, 기타 백엔드에 대해 구현할 수 있습니다.
type Deployer interface {
	// Deploy는 애플리케이션을 배포합니다.
	Deploy(req *DeployRequest) (*DeployResult, error)

	// Status는 특정 앱의 현재 상태를 반환합니다.
	Status(appName string) (*AppStatus, error)

	// ListAll은 모든 배포된 앱의 상태 목록을 반환합니다.
	ListAll() ([]*AppStatus, error)

	// Logs는 컨테이너 로그를 w에 씁니다.
	Logs(w io.Writer, appName string, opts *LogOptions) error

	// Stop은 실행 중인 컨테이너를 중지합니다.
	Stop(appName string) error

	// Remove는 컨테이너를 삭제합니다.
	Remove(appName string) error
}
