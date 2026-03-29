// solution/watch.go
// 감시 시스템 참고 풀이
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type EventType string

const (
	EventPut    EventType = "PUT"
	EventDelete EventType = "DELETE"
)

type WatchEvent struct {
	Type      EventType
	Key       string
	Value     string
	Revision  int64
	PrevValue string
}

type WatchResponse struct {
	Events   []WatchEvent
	Revision int64
}

type WatchChan <-chan WatchResponse

type WatchOption func(*watchOptions)

type watchOptions struct {
	prefix       bool
	fromRevision int64
}

func WithPrefix() WatchOption {
	return func(o *watchOptions) { o.prefix = true }
}

func WithRevision(rev int64) WatchOption {
	return func(o *watchOptions) { o.fromRevision = rev }
}

// ============================================================
// watchableStore 구현
// ============================================================

type kvEntry struct {
	key      string
	value    string
	revision int64
	deleted  bool
}

type watcher struct {
	key      string
	prefix   bool
	fromRev  int64
	ch       chan WatchResponse
	cancelCh chan struct{}
}

type watchableStore struct {
	mu         sync.RWMutex
	current    map[string]kvEntry
	history    []WatchEvent // 전체 이벤트 히스토리 (리비전 순)
	revision   int64
	watchers   []*watcher
	compactRev int64 // 이 리비전 이전 히스토리는 압축됨
}

func NewWatchableStore() *watchableStore {
	return &watchableStore{
		current: make(map[string]kvEntry),
	}
}

func (s *watchableStore) Put(key, value string) int64 {
	s.mu.Lock()

	s.revision++
	rev := s.revision

	prevValue := ""
	if entry, ok := s.current[key]; ok && !entry.deleted {
		prevValue = entry.value
	}

	s.current[key] = kvEntry{key: key, value: value, revision: rev}

	ev := WatchEvent{
		Type:      EventPut,
		Key:       key,
		Value:     value,
		Revision:  rev,
		PrevValue: prevValue,
	}
	s.history = append(s.history, ev)

	// 와처 목록 복사 (notify 중 mu 잠금 해제)
	watchers := make([]*watcher, len(s.watchers))
	copy(watchers, s.watchers)
	s.mu.Unlock()

	s.notifyWatcherList(watchers, ev)
	return rev
}

func (s *watchableStore) Delete(key string) int64 {
	s.mu.Lock()

	entry, ok := s.current[key]
	if !ok || entry.deleted {
		s.mu.Unlock()
		return -1
	}

	s.revision++
	rev := s.revision
	prevValue := entry.value

	s.current[key] = kvEntry{key: key, revision: rev, deleted: true}

	ev := WatchEvent{
		Type:      EventDelete,
		Key:       key,
		Revision:  rev,
		PrevValue: prevValue,
	}
	s.history = append(s.history, ev)

	watchers := make([]*watcher, len(s.watchers))
	copy(watchers, s.watchers)
	s.mu.Unlock()

	s.notifyWatcherList(watchers, ev)
	return rev
}

func (s *watchableStore) Get(key string) (string, int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.current[key]
	if !ok || entry.deleted {
		return "", 0, false
	}
	return entry.value, entry.revision, true
}

func (s *watchableStore) CurrentRevision() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.revision
}

func (s *watchableStore) Watch(ctx context.Context, key string, opts ...WatchOption) WatchChan {
	options := &watchOptions{}
	for _, opt := range opts {
		opt(options)
	}

	w := &watcher{
		key:      key,
		prefix:   options.prefix,
		fromRev:  options.fromRevision,
		ch:       make(chan WatchResponse, 64),
		cancelCh: make(chan struct{}),
	}

	s.mu.Lock()
	s.watchers = append(s.watchers, w)

	// fromRevision이 설정된 경우 기존 이벤트 재연 준비
	var replayEvents []WatchEvent
	if options.fromRevision > 0 {
		for _, ev := range s.history {
			if ev.Revision >= options.fromRevision && matchesWatcher(w, ev.Key) {
				replayEvents = append(replayEvents, ev)
			}
		}
	}
	s.mu.Unlock()

	// goroutine으로 ctx 취소 감지 및 재연 이벤트 전송
	go func() {
		// 재연 이벤트 전송
		for _, ev := range replayEvents {
			resp := WatchResponse{Events: []WatchEvent{ev}, Revision: ev.Revision}
			select {
			case w.ch <- resp:
			case <-ctx.Done():
				s.removeWatcher(w)
				close(w.ch)
				return
			}
		}

		// ctx 취소 대기
		<-ctx.Done()
		s.removeWatcher(w)
		close(w.ch)
	}()

	return WatchChan(w.ch)
}

func (s *watchableStore) Compact(revision int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if revision > s.revision {
		return fmt.Errorf("압축 리비전 %d이 현재 리비전 %d보다 큽니다", revision, s.revision)
	}

	// 압축 리비전 이전 히스토리 제거
	var newHistory []WatchEvent
	for _, ev := range s.history {
		if ev.Revision > revision {
			newHistory = append(newHistory, ev)
		}
	}
	s.history = newHistory
	s.compactRev = revision
	return nil
}

func (s *watchableStore) notifyWatcherList(watchers []*watcher, event WatchEvent) {
	for _, w := range watchers {
		if !matchesWatcher(w, event.Key) {
			continue
		}
		// fromRevision보다 낮은 이벤트는 건너뜀 (재연 중이 아닐 때)
		resp := WatchResponse{
			Events:   []WatchEvent{event},
			Revision: event.Revision,
		}
		select {
		case w.ch <- resp:
		case <-w.cancelCh:
		default:
			// 채널이 가득 찬 경우 드롭 (실제 etcd는 별도 처리)
		}
	}
}

func (s *watchableStore) notifyWatchers(event WatchEvent) {
	s.mu.RLock()
	watchers := make([]*watcher, len(s.watchers))
	copy(watchers, s.watchers)
	s.mu.RUnlock()
	s.notifyWatcherList(watchers, event)
}

func matchesWatcher(w *watcher, eventKey string) bool {
	if w.prefix {
		return strings.HasPrefix(eventKey, w.key)
	}
	return eventKey == w.key
}

func (s *watchableStore) removeWatcher(w *watcher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, existing := range s.watchers {
		if existing == w {
			s.watchers = append(s.watchers[:i], s.watchers[i+1:]...)
			return
		}
	}
}

func main() {}
