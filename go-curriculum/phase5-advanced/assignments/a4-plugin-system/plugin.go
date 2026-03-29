// a4-plugin-system/plugin.go
// 플러그인 시스템 구현 과제입니다.
// TODO 주석이 있는 모든 함수/메서드를 구현하세요.
package plugin

import (
	"context"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// HealthStatus는 플러그인의 현재 상태를 나타냅니다.
type HealthStatus struct {
	Healthy bool
	Message string
}

// Plugin은 모든 플러그인이 구현해야 하는 인터페이스입니다.
type Plugin interface {
	// Name은 플러그인의 고유 이름을 반환합니다.
	Name() string
	// Version은 플러그인 버전 문자열을 반환합니다 (예: "1.0.0").
	Version() string
	// Dependencies는 이 플러그인이 의존하는 플러그인 이름 목록을 반환합니다.
	Dependencies() []string
	// Init은 플러그인을 초기화합니다.
	Init(ctx context.Context, cfg map[string]any) error
	// Execute는 플러그인 로직을 실행합니다.
	Execute(ctx context.Context, input any) (any, error)
	// Shutdown은 플러그인을 종료합니다.
	Shutdown(ctx context.Context) error
	// Health는 플러그인의 현재 상태를 반환합니다.
	Health() HealthStatus
}

// PluginConfigs는 InitAll에 전달되는 플러그인별 설정 맵입니다.
type PluginConfigs map[string]map[string]any

// ============================================================
// PluginManager
// ============================================================

// PluginManager는 플러그인의 등록, 초기화, 실행, 종료를 관리합니다.
// 고루틴 안전해야 합니다.
type PluginManager struct {
	// TODO: 필드를 추가하세요.
}

// NewPluginManager는 새 PluginManager를 생성합니다.
// TODO: 구현하세요.
func NewPluginManager() *PluginManager {
	return &PluginManager{}
}

// Register는 플러그인을 매니저에 등록합니다.
// 같은 이름의 플러그인이 이미 등록된 경우 error를 반환합니다.
// TODO: 구현하세요.
func (pm *PluginManager) Register(p Plugin) error {
	return nil
}

// Get은 이름으로 플러그인을 조회합니다.
// TODO: 구현하세요.
func (pm *PluginManager) Get(name string) (Plugin, bool) {
	return nil, false
}

// List는 등록된 플러그인 이름을 오름차순으로 반환합니다.
// TODO: 구현하세요.
func (pm *PluginManager) List() []string {
	return nil
}

// InitAll은 의존성 순서(위상 정렬)에 따라 모든 플러그인을 초기화합니다.
//
// 요구사항:
//   - 의존하는 플러그인이 먼저 Init되어야 합니다.
//   - 순환 의존성이 있으면 error를 반환합니다.
//   - 의존하는 플러그인이 등록되지 않았으면 error를 반환합니다.
//   - 초기화 순서를 내부에 저장해 ShutdownAll에서 역순으로 사용합니다.
//
// TODO: 구현하세요.
func (pm *PluginManager) InitAll(ctx context.Context, configs PluginConfigs) error {
	return nil
}

// ShutdownAll은 InitAll의 역순으로 모든 플러그인을 종료합니다.
// 일부 플러그인의 Shutdown이 실패해도 나머지는 계속 진행합니다.
// 발생한 모든 오류를 하나의 error로 합쳐 반환합니다 (없으면 nil).
// TODO: 구현하세요.
func (pm *PluginManager) ShutdownAll(ctx context.Context) error {
	return nil
}

// Execute는 이름으로 플러그인을 찾아 실행합니다.
// 플러그인이 없거나 초기화되지 않았으면 error를 반환합니다.
// TODO: 구현하세요.
func (pm *PluginManager) Execute(ctx context.Context, name string, input any) (any, error) {
	return nil, nil
}

// HealthCheck는 등록된 모든 플러그인의 상태를 반환합니다.
// 초기화되지 않은 플러그인은 Healthy: false로 보고합니다.
// TODO: 구현하세요.
func (pm *PluginManager) HealthCheck() map[string]HealthStatus {
	return nil
}
