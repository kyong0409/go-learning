// a4-plugin-system/solution/plugin.go
// 플러그인 시스템 참고 답안입니다.
package plugin

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ============================================================
// 타입 정의
// ============================================================

type HealthStatus struct {
	Healthy bool
	Message string
}

type Plugin interface {
	Name() string
	Version() string
	Dependencies() []string
	Init(ctx context.Context, cfg map[string]any) error
	Execute(ctx context.Context, input any) (any, error)
	Shutdown(ctx context.Context) error
	Health() HealthStatus
}

type PluginConfigs map[string]map[string]any

// ============================================================
// PluginManager
// ============================================================

type PluginManager struct {
	mu           sync.RWMutex
	plugins      map[string]Plugin
	initOrder    []string // 초기화된 순서 (ShutdownAll 역순 사용)
	initialized  map[string]bool
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins:     make(map[string]Plugin),
		initialized: make(map[string]bool),
	}
}

func (pm *PluginManager) Register(p Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if _, exists := pm.plugins[p.Name()]; exists {
		return fmt.Errorf("플러그인 %q이 이미 등록되어 있습니다", p.Name())
	}
	pm.plugins[p.Name()] = p
	return nil
}

func (pm *PluginManager) Get(name string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, ok := pm.plugins[name]
	return p, ok
}

func (pm *PluginManager) List() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// InitAll은 Kahn 알고리즘으로 위상 정렬 후 의존성 순서로 Init을 호출합니다.
func (pm *PluginManager) InitAll(ctx context.Context, configs PluginConfigs) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 모든 플러그인 이름 수집
	names := make([]string, 0, len(pm.plugins))
	for name := range pm.plugins {
		names = append(names, name)
	}

	// 의존성 검증 + 진입 차수 계산
	inDegree := make(map[string]int, len(names))
	adjList := make(map[string][]string, len(names)) // 의존자 -> 이 플러그인에 의존하는 플러그인 목록

	for _, name := range names {
		inDegree[name] = 0
		adjList[name] = nil
	}

	for _, name := range names {
		p := pm.plugins[name]
		for _, dep := range p.Dependencies() {
			if _, exists := pm.plugins[dep]; !exists {
				return fmt.Errorf("플러그인 %q의 의존성 %q가 등록되지 않았습니다", name, dep)
			}
			inDegree[name]++
			adjList[dep] = append(adjList[dep], name)
		}
	}

	// Kahn 알고리즘: 진입 차수 0인 노드부터 처리
	queue := make([]string, 0)
	for _, name := range names {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}
	// 결정론적 순서를 위해 정렬
	sort.Strings(queue)

	order := make([]string, 0, len(names))
	for len(queue) > 0 {
		// 가장 앞의 노드 처리
		cur := queue[0]
		queue = queue[1:]
		order = append(order, cur)

		// 이 노드에 의존하는 노드들의 진입 차수 감소
		dependents := adjList[cur]
		sort.Strings(dependents)
		for _, dep := range dependents {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
				sort.Strings(queue)
			}
		}
	}

	if len(order) != len(names) {
		return fmt.Errorf("플러그인 의존성 순환이 감지되었습니다")
	}

	// 순서대로 Init 호출
	pm.initOrder = make([]string, 0, len(order))
	for _, name := range order {
		p := pm.plugins[name]
		var cfg map[string]any
		if configs != nil {
			cfg = configs[name]
		}
		if err := p.Init(ctx, cfg); err != nil {
			return fmt.Errorf("플러그인 %q 초기화 실패: %w", name, err)
		}
		pm.initialized[name] = true
		pm.initOrder = append(pm.initOrder, name)
	}

	return nil
}

// ShutdownAll은 초기화 역순으로 Shutdown을 호출하고 모든 오류를 수집합니다.
func (pm *PluginManager) ShutdownAll(ctx context.Context) error {
	pm.mu.Lock()
	order := make([]string, len(pm.initOrder))
	copy(order, pm.initOrder)
	pm.mu.Unlock()

	var errs []string
	for i := len(order) - 1; i >= 0; i-- {
		name := order[i]
		pm.mu.RLock()
		p, ok := pm.plugins[name]
		pm.mu.RUnlock()

		if !ok {
			continue
		}
		if err := p.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}

	pm.mu.Lock()
	pm.initialized = make(map[string]bool)
	pm.initOrder = nil
	pm.mu.Unlock()

	if len(errs) > 0 {
		return fmt.Errorf("종료 오류: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (pm *PluginManager) Execute(ctx context.Context, name string, input any) (any, error) {
	pm.mu.RLock()
	p, ok := pm.plugins[name]
	initialized := pm.initialized[name]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("플러그인 %q을 찾을 수 없습니다", name)
	}
	if !initialized {
		return nil, fmt.Errorf("플러그인 %q이 초기화되지 않았습니다", name)
	}
	return p.Execute(ctx, input)
}

func (pm *PluginManager) HealthCheck() map[string]HealthStatus {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]HealthStatus, len(pm.plugins))
	for name, p := range pm.plugins {
		if !pm.initialized[name] {
			result[name] = HealthStatus{Healthy: false, Message: "초기화되지 않음"}
			continue
		}
		result[name] = p.Health()
	}
	return result
}
