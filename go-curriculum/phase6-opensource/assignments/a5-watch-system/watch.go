// watch.go
// etcd Watch 패턴 구현 - MVCC 기반 감시 시스템
//
// TODO: 아래 타입과 함수들을 완성하세요.
package main

import (
	"context"
	"strings"
	"sync"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// EventType은 이벤트 종류입니다.
type EventType string

const (
	EventPut    EventType = "PUT"
	EventDelete EventType = "DELETE"
)

// WatchEvent는 단일 키 변경 이벤트입니다.
type WatchEvent struct {
	Type      EventType
	Key       string
	Value     string // DELETE면 빈 문자열
	Revision  int64  // 이 이벤트의 리비전
	PrevValue string // 이전 값 (신규 PUT이면 빈 문자열)
}

// WatchResponse는 Watch 채널로 전달되는 응답입니다.
type WatchResponse struct {
	Events   []WatchEvent
	Revision int64
}

// WatchChan은 Watch 이벤트 채널 타입입니다.
type WatchChan <-chan WatchResponse

// WatchOption은 Watch 동작을 설정하는 함수형 옵션입니다.
type WatchOption func(*watchOptions)

type watchOptions struct {
	prefix      bool
	fromRevision int64
}

// WithPrefix는 키 접두사로 Watch합니다.
func WithPrefix() WatchOption {
	return func(o *watchOptions) {
		o.prefix = true
	}
}

// WithRevision은 특정 리비전 이후 이벤트부터 수신합니다.
// 이미 발생한 이벤트는 즉시 재연(replay)됩니다.
func WithRevision(rev int64) WatchOption {
	return func(o *watchOptions) {
		o.fromRevision = rev
	}
}

// ============================================================
// WatchableStore 구현
// ============================================================

// kvEntry는 키-값의 특정 시점 스냅샷입니다.
type kvEntry struct {
	key       string
	value     string
	revision  int64
	deleted   bool
}

// watcher는 단일 Watch 구독자입니다.
type watcher struct {
	key      string
	prefix   bool
	fromRev  int64
	ch       chan WatchResponse
	cancelCh chan struct{}
}

// watchableStore는 WatchableStore를 구현합니다.
type watchableStore struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - mu: sync.RWMutex
	//   - current: map[string]kvEntry (현재 값)
	//   - history: []WatchEvent (모든 이벤트, 리비전 순)
	//   - revision: int64 (현재 리비전, atomic 또는 mu로 보호)
	//   - watchers: []*watcher
	//   - compactRev: int64 (압축된 최소 리비전)
}

// NewWatchableStore는 새 watchableStore를 생성합니다.
func NewWatchableStore() *watchableStore {
	// TODO: 구현하세요
	panic("NewWatchableStore: 아직 구현되지 않았습니다")
}

// Put은 키-값을 저장하고 새 리비전을 반환합니다.
// 키가 이미 있으면 업데이트, 없으면 신규 생성입니다.
func (s *watchableStore) Put(key, value string) int64 {
	// TODO: 구현하세요
	// 힌트:
	//   1. revision 증가
	//   2. current[key] 업데이트
	//   3. history에 WatchEvent 추가
	//   4. 해당 키를 Watch 중인 와처에 이벤트 전송
	panic("Put: 아직 구현되지 않았습니다")
}

// Delete는 키를 삭제하고 새 리비전을 반환합니다.
// 키가 없으면 -1을 반환합니다.
func (s *watchableStore) Delete(key string) int64 {
	// TODO: 구현하세요
	panic("Delete: 아직 구현되지 않았습니다")
}

// Get은 키의 현재 값과 리비전을 반환합니다.
func (s *watchableStore) Get(key string) (value string, revision int64, ok bool) {
	// TODO: 구현하세요
	panic("Get: 아직 구현되지 않았습니다")
}

// CurrentRevision은 현재 전역 리비전을 반환합니다.
func (s *watchableStore) CurrentRevision() int64 {
	// TODO: 구현하세요
	panic("CurrentRevision: 아직 구현되지 않았습니다")
}

// Watch는 key(또는 접두사)의 변경을 감시하는 채널을 반환합니다.
// ctx 취소 시 채널이 닫힙니다.
//
// WithRevision(rev) 옵션이 있으면:
//   - rev 이후의 기존 이벤트를 먼저 채널로 전송 (재연)
//   - 그 후 새 이벤트를 실시간으로 전송
//
// 힌트:
//   - watcher를 생성해 s.watchers에 등록
//   - goroutine을 시작해 ctx 취소 시 watcher를 제거하고 채널을 닫음
//   - fromRevision이 있으면 goroutine에서 기존 이벤트를 먼저 전송
func (s *watchableStore) Watch(ctx context.Context, key string, opts ...WatchOption) WatchChan {
	// TODO: 구현하세요
	panic("Watch: 아직 구현되지 않았습니다")
}

// Compact는 revision 이전의 이벤트 히스토리를 삭제합니다.
// revision 이전으로 Watch 시작을 요청하면 에러가 되어야 합니다.
// (이 과제에서는 이미 등록된 와처에는 영향을 주지 않아도 됩니다)
func (s *watchableStore) Compact(revision int64) error {
	// TODO: 구현하세요
	panic("Compact: 아직 구현되지 않았습니다")
}

// ============================================================
// 내부 헬퍼
// ============================================================

// notifyWatchers는 이벤트를 관련 와처들에게 전송합니다.
// mu가 잠긴 상태에서 호출하지 마세요 (데드락 위험).
func (s *watchableStore) notifyWatchers(event WatchEvent) {
	// TODO: 구현하세요
	// 힌트: 각 watcher에 대해 key 또는 prefix가 매칭되면 채널로 전송
	panic("notifyWatchers: 아직 구현되지 않았습니다")
}

// matchesWatcher는 이벤트가 와처의 조건에 맞는지 확인합니다.
func matchesWatcher(w *watcher, eventKey string) bool {
	if w.prefix {
		return strings.HasPrefix(eventKey, w.key)
	}
	return eventKey == w.key
}

// removeWatcher는 와처 목록에서 특정 와처를 제거합니다.
func (s *watchableStore) removeWatcher(w *watcher) {
	// TODO: 구현하세요
	panic("removeWatcher: 아직 구현되지 않았습니다")
}

func main() {
	// 테스트를 실행하세요: go test ./... -v
}
