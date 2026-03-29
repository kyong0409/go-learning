// internal/deployer/docker.go
// Docker 기반 Deployer 구현
//
// Docker CLI를 exec.Command로 호출해 컨테이너를 관리합니다.
// (실제 프로덕션에서는 docker SDK: github.com/docker/docker/client 사용 권장)
package deployer

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// ============================================================
// DockerDeployer 구현
// ============================================================

// DockerDeployer는 Docker CLI를 사용해 컨테이너를 관리합니다.
type DockerDeployer struct {
	registry string // 기본 레지스트리 (예: docker.io)
	verbose  bool   // 상세 출력
}

// NewDockerDeployer는 DockerDeployer 생성자입니다.
func NewDockerDeployer(registry string, verbose bool) *DockerDeployer {
	if registry == "" {
		registry = "docker.io"
	}
	return &DockerDeployer{
		registry: registry,
		verbose:  verbose,
	}
}

// ============================================================
// Deployer 인터페이스 구현
// ============================================================

// Deploy는 Docker 컨테이너를 배포합니다.
func (d *DockerDeployer) Deploy(req *DeployRequest) (*DeployResult, error) {
	if req.Replicas <= 0 {
		req.Replicas = 1
	}

	d.logf("이미지 Pull 확인: %s", req.Image)

	// 1. 이미지 Pull
	if err := d.pullImage(req.Image); err != nil {
		return nil, fmt.Errorf("이미지 Pull 실패: %w", err)
	}

	// 2. 기존 컨테이너 정리
	containerName := containerName(req.AppName)
	d.logf("기존 컨테이너 확인: %s", containerName)
	_ = d.stopContainer(containerName)  // 없으면 무시
	_ = d.removeContainer(containerName) // 없으면 무시

	// 3. 새 컨테이너 실행
	d.logf("컨테이너 시작: %s", containerName)
	containerID, err := d.runContainer(req, containerName)
	if err != nil {
		return nil, fmt.Errorf("컨테이너 실행 실패: %w", err)
	}

	// 짧은 컨테이너 ID (12자)
	shortID := containerID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}

	result := &DeployResult{
		AppName:     req.AppName,
		Image:       req.Image,
		ContainerID: shortID,
		Status:      "running",
		Port:        req.Port,
		DeployedAt:  time.Now(),
		Message:     fmt.Sprintf("컨테이너 %s 시작됨", shortID),
	}

	return result, nil
}

// Status는 특정 앱의 현재 상태를 반환합니다.
func (d *DockerDeployer) Status(appName string) (*AppStatus, error) {
	name := containerName(appName)

	// docker inspect로 상태 조회
	out, err := d.runDockerCmd("inspect",
		"--format", "{{.Id}}\t{{.Config.Image}}\t{{.State.Status}}\t{{.State.RestartCount}}",
		name,
	)
	if err != nil {
		return nil, fmt.Errorf("컨테이너 %q를 찾을 수 없습니다", appName)
	}

	parts := strings.Split(strings.TrimSpace(out), "\t")
	if len(parts) < 4 {
		return nil, fmt.Errorf("컨테이너 정보 파싱 실패")
	}

	shortID := parts[0]
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}

	restarts := 0
	fmt.Sscanf(parts[3], "%d", &restarts)

	status := &AppStatus{
		AppName:     appName,
		Image:       parts[1],
		ContainerID: shortID,
		Status:      parts[2],
		Restarts:    restarts,
		CreatedAt:   time.Now(), // 단순화
		Uptime:      "계산 중",
	}

	return status, nil
}

// ListAll은 deploy-tool로 관리하는 모든 컨테이너 목록을 반환합니다.
func (d *DockerDeployer) ListAll() ([]*AppStatus, error) {
	// 라벨로 deploy-tool이 만든 컨테이너만 필터링
	out, err := d.runDockerCmd("ps", "-a",
		"--filter", "label=managed-by=deploy-tool",
		"--format", "{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.ID}}",
	)
	if err != nil {
		return nil, fmt.Errorf("컨테이너 목록 조회 실패: %w", err)
	}

	var statuses []*AppStatus
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}

		name := strings.TrimPrefix(parts[0], "deploy-")
		shortID := parts[3]
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}

		statuses = append(statuses, &AppStatus{
			AppName:     name,
			Image:       parts[1],
			Status:      parts[2],
			ContainerID: shortID,
		})
	}

	return statuses, nil
}

// Logs는 컨테이너 로그를 w에 씁니다.
func (d *DockerDeployer) Logs(w io.Writer, appName string, opts *LogOptions) error {
	name := containerName(appName)

	args := []string{"logs"}
	if opts.Follow {
		args = append(args, "--follow")
	}
	if opts.Tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", opts.Tail))
	}
	if opts.Since != "" {
		args = append(args, "--since", opts.Since)
	}
	args = append(args, "--timestamps", name)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = w
	cmd.Stderr = w

	d.logf("실행: docker %s", strings.Join(args, " "))

	return cmd.Run()
}

// Stop은 컨테이너를 중지합니다.
func (d *DockerDeployer) Stop(appName string) error {
	return d.stopContainer(containerName(appName))
}

// Remove는 컨테이너를 삭제합니다.
func (d *DockerDeployer) Remove(appName string) error {
	name := containerName(appName)
	if err := d.stopContainer(name); err != nil {
		d.logf("중지 실패 (무시): %v", err)
	}
	return d.removeContainer(name)
}

// ============================================================
// 내부 헬퍼 메서드
// ============================================================

// pullImage는 Docker 이미지를 Pull합니다.
func (d *DockerDeployer) pullImage(image string) error {
	_, err := d.runDockerCmd("pull", image)
	return err
}

// runContainer는 새 컨테이너를 실행하고 ID를 반환합니다.
func (d *DockerDeployer) runContainer(req *DeployRequest, name string) (string, error) {
	args := []string{
		"run", "-d",
		"--name", name,
		"--label", "managed-by=deploy-tool",
		"--label", "app=" + req.AppName,
		"--restart", "unless-stopped",
	}

	// 포트 매핑
	if req.Port > 0 {
		args = append(args, "-p", fmt.Sprintf("%d:%d", req.Port, req.Port))
	}

	// 환경 변수
	for k, v := range req.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, req.Image)

	return d.runDockerCmd(args...)
}

// stopContainer는 컨테이너를 중지합니다.
func (d *DockerDeployer) stopContainer(name string) error {
	_, err := d.runDockerCmd("stop", name)
	return err
}

// removeContainer는 컨테이너를 삭제합니다.
func (d *DockerDeployer) removeContainer(name string) error {
	_, err := d.runDockerCmd("rm", "-f", name)
	return err
}

// runDockerCmd는 docker 명령어를 실행하고 stdout을 반환합니다.
func (d *DockerDeployer) runDockerCmd(args ...string) (string, error) {
	d.logf("실행: docker %s", strings.Join(args, " "))

	cmd := exec.Command("docker", args...)
	out, err := cmd.Output()
	if err != nil {
		// stderr 내용을 오류 메시지에 포함
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("docker 명령 실패: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// logf는 verbose 모드일 때 디버그 로그를 출력합니다.
func (d *DockerDeployer) logf(format string, args ...interface{}) {
	if d.verbose {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// containerName은 앱 이름에서 Docker 컨테이너 이름을 생성합니다.
// 예: "myapp" → "deploy-myapp"
func containerName(appName string) string {
	return "deploy-" + appName
}
