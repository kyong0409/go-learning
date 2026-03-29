// solution/informer.go
// 인포머/리스터 참고 풀이
//
// 이 파일은 참고용입니다. 먼저 직접 구현해보세요.
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

// ============================================================
// 타입 정의
// ============================================================

type Object struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Data      map[string]interface{}
}

type WatchEvent struct {
	Type   string
	Object Object
}

type EventHandler interface {
	OnAdd(obj Object)
	OnUpdate(oldObj, newObj Object)
	OnDelete(obj Object)
}

type IndexFunc func(obj Object) []string
type ListFunc func(ctx context.Context) ([]Object, error)
type WatchFunc func(ctx context.Context) (<-chan WatchEvent, error)

func KeyFunc(obj Object) string {
	if obj.Namespace == "" {
		return obj.Name
	}
	return fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)
}

// ============================================================
// Store 구현
// ============================================================

type store struct {
	mu       sync.RWMutex
	items    map[string]Object              // key → object
	indices  map[string]map[string][]string // indexName → indexValue → []key
	indexers map[string]IndexFunc
}

func NewStore() *store {
	return &store{
		items:    make(map[string]Object),
		indices:  make(map[string]map[string][]string),
		indexers: make(map[string]IndexFunc),
	}
}

func (s *store) Add(obj Object) error {
	key := KeyFunc(obj)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = obj
	s.updateIndices(key, obj)
	return nil
}

func (s *store) Update(obj Object) error {
	key := KeyFunc(obj)
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, exists := s.items[key]; exists {
		s.deleteIndices(key, old)
	}
	s.items[key] = obj
	s.updateIndices(key, obj)
	return nil
}

func (s *store) Delete(obj Object) error {
	key := KeyFunc(obj)
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, exists := s.items[key]; exists {
		s.deleteIndices(key, existing)
	}
	delete(s.items, key)
	return nil
}

func (s *store) Get(key string) (Object, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	obj, ok := s.items[key]
	return obj, ok
}

func (s *store) List() []Object {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Object, 0, len(s.items))
	for _, obj := range s.items {
		result = append(result, obj)
	}
	return result
}

func (s *store) AddIndexer(name string, fn IndexFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.indexers[name] = fn
	s.indices[name] = make(map[string][]string)
	// 기존 오브젝트에 대해 인덱스 생성
	for key, obj := range s.items {
		for _, val := range fn(obj) {
			s.indices[name][val] = append(s.indices[name][val], key)
		}
	}
}

func (s *store) ListByIndex(indexName, indexValue string) ([]Object, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, ok := s.indices[indexName]
	if !ok {
		return nil, fmt.Errorf("인덱서 %q를 찾을 수 없습니다", indexName)
	}
	keys := idx[indexValue]
	result := make([]Object, 0, len(keys))
	for _, key := range keys {
		if obj, exists := s.items[key]; exists {
			result = append(result, obj)
		}
	}
	return result, nil
}

func (s *store) Replace(objects []Object) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]Object)
	// 인덱스 초기화
	for name := range s.indices {
		s.indices[name] = make(map[string][]string)
	}
	for _, obj := range objects {
		key := KeyFunc(obj)
		s.items[key] = obj
		s.updateIndices(key, obj)
	}
	return nil
}

// updateIndices는 mu가 잠긴 상태에서 호출되어야 합니다.
func (s *store) updateIndices(key string, obj Object) {
	for name, fn := range s.indexers {
		for _, val := range fn(obj) {
			s.indices[name][val] = append(s.indices[name][val], key)
		}
	}
}

// deleteIndices는 mu가 잠긴 상태에서 호출되어야 합니다.
func (s *store) deleteIndices(key string, obj Object) {
	for name, fn := range s.indexers {
		for _, val := range fn(obj) {
			keys := s.indices[name][val]
			newKeys := keys[:0]
			for _, k := range keys {
				if k != key {
					newKeys = append(newKeys, k)
				}
			}
			s.indices[name][val] = newKeys
		}
	}
}

// ============================================================
// Reflector 구현
// ============================================================

type Reflector struct {
	store     *store
	listFunc  ListFunc
	watchFunc WatchFunc
	synced    atomic.Bool
	// 이벤트 핸들러 콜백 (Informer에서 설정)
	onAdd    func(Object)
	onUpdate func(old, new Object)
	onDelete func(Object)
}

func NewReflector(s *store, lf ListFunc, wf WatchFunc) *Reflector {
	return &Reflector{
		store:     s,
		listFunc:  lf,
		watchFunc: wf,
	}
}

func (r *Reflector) Run(ctx context.Context) {
	// 1. 초기 List
	objects, err := r.listFunc(ctx)
	if err != nil {
		return
	}
	r.store.Replace(objects)

	// 초기 List 완료 후 핸들러 OnAdd 호출
	if r.onAdd != nil {
		for _, obj := range objects {
			r.onAdd(obj)
		}
	}
	r.synced.Store(true)

	// 2. Watch
	if r.watchFunc == nil {
		return
	}
	watchCh, err := r.watchFunc(ctx)
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-watchCh:
			if !ok {
				return
			}
			switch ev.Type {
			case "ADDED":
				old, _ := r.store.Get(KeyFunc(ev.Object))
				r.store.Add(ev.Object)
				if r.onAdd != nil {
					r.onAdd(ev.Object)
				}
				_ = old
			case "MODIFIED":
				old, _ := r.store.Get(KeyFunc(ev.Object))
				r.store.Update(ev.Object)
				if r.onUpdate != nil {
					r.onUpdate(old, ev.Object)
				}
			case "DELETED":
				old, _ := r.store.Get(KeyFunc(ev.Object))
				r.store.Delete(ev.Object)
				if r.onDelete != nil {
					r.onDelete(old)
				}
			}
		}
	}
}

func (r *Reflector) HasSynced() bool {
	return r.synced.Load()
}

// ============================================================
// Informer 구현
// ============================================================

type Informer struct {
	store     *store
	reflector *Reflector
	mu        sync.RWMutex
	handlers  []EventHandler
}

func NewInformer(lf ListFunc, wf WatchFunc) *Informer {
	s := NewStore()
	inf := &Informer{store: s}

	r := NewReflector(s, lf, wf)
	r.onAdd = func(obj Object) {
		inf.mu.RLock()
		defer inf.mu.RUnlock()
		for _, h := range inf.handlers {
			h.OnAdd(obj)
		}
	}
	r.onUpdate = func(old, new Object) {
		inf.mu.RLock()
		defer inf.mu.RUnlock()
		for _, h := range inf.handlers {
			h.OnUpdate(old, new)
		}
	}
	r.onDelete = func(obj Object) {
		inf.mu.RLock()
		defer inf.mu.RUnlock()
		for _, h := range inf.handlers {
			h.OnDelete(obj)
		}
	}
	inf.reflector = r
	return inf
}

func (inf *Informer) AddEventHandler(h EventHandler) {
	inf.mu.Lock()
	defer inf.mu.Unlock()
	inf.handlers = append(inf.handlers, h)
}

func (inf *Informer) Run(ctx context.Context) {
	inf.reflector.Run(ctx)
}

func (inf *Informer) HasSynced() bool {
	return inf.reflector.HasSynced()
}

func (inf *Informer) GetStore() *store {
	return inf.store
}

// ============================================================
// Lister 구현
// ============================================================

type Lister struct {
	mu    sync.RWMutex
	store *store
}

func NewLister(s *store) *Lister {
	return &Lister{store: s}
}

func (l *Lister) List() []Object {
	return l.store.List()
}

func (l *Lister) Get(key string) (Object, bool) {
	return l.store.Get(key)
}

func (l *Lister) ListByLabel(labelKey, labelValue string) []Object {
	all := l.store.List()
	var result []Object
	for _, obj := range all {
		if obj.Labels[labelKey] == labelValue {
			result = append(result, obj)
		}
	}
	return result
}

// ============================================================
// 헬퍼
// ============================================================

func LabelIndexFunc(labelKey string) IndexFunc {
	return func(obj Object) []string {
		if val, ok := obj.Labels[labelKey]; ok {
			return []string{val}
		}
		return nil
	}
}

var NamespaceIndexFunc IndexFunc = func(obj Object) []string {
	if obj.Namespace == "" {
		return nil
	}
	return []string{obj.Namespace}
}

func splitKey(key string) (namespace, name string) {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", parts[0]
}

func main() {}
