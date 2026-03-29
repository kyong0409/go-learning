// informer.go
// 인포머/리스터 패턴 - client-go Informer 패턴 구현
//
// TODO: 아래 타입과 함수들을 완성하세요.
// Store, Reflector, Informer, Lister를 순서대로 구현합니다.
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// ============================================================
// 타입 정의 (수정하지 마세요)
// ============================================================

// Object는 컨트롤러가 관리하는 리소스를 표현합니다.
type Object struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Data      map[string]interface{}
}

// WatchEvent는 Watch 채널을 통해 전달되는 이벤트입니다.
type WatchEvent struct {
	Type   string // "ADDED", "MODIFIED", "DELETED"
	Object Object
}

// EventHandler는 오브젝트 변경 콜백 인터페이스입니다.
// client-go의 ResourceEventHandler와 동일한 역할입니다.
type EventHandler interface {
	OnAdd(obj Object)
	OnUpdate(oldObj, newObj Object)
	OnDelete(obj Object)
}

// IndexFunc는 오브젝트에서 인덱스 값 목록을 추출합니다.
type IndexFunc func(obj Object) []string

// ListFunc는 모든 오브젝트를 반환합니다.
type ListFunc func(ctx context.Context) ([]Object, error)

// WatchFunc는 변경 이벤트 채널을 반환합니다.
type WatchFunc func(ctx context.Context) (<-chan WatchEvent, error)

// ============================================================
// 헬퍼 함수
// ============================================================

// KeyFunc는 오브젝트의 고유 키를 반환합니다.
// namespace가 있으면 "namespace/name", 없으면 "name"
func KeyFunc(obj Object) string {
	if obj.Namespace == "" {
		return obj.Name
	}
	return fmt.Sprintf("%s/%s", obj.Namespace, obj.Name)
}

// ============================================================
// Store 구현
// ============================================================

// store는 쓰레드 안전 인메모리 캐시입니다.
// client-go의 ThreadSafeStore + cache를 단순화한 버전입니다.
type store struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - items: map[string]Object (key → object)
	//   - indices: map[string]map[string][]string (indexName → indexValue → []key)
	//   - indexers: map[string]IndexFunc (indexName → 인덱스 함수)
	//   - mu: sync.RWMutex
}

// NewStore는 새 store를 생성합니다.
func NewStore() *store {
	// TODO: 구현하세요
	panic("NewStore: 아직 구현되지 않았습니다")
}

// Add는 오브젝트를 스토어에 추가합니다.
func (s *store) Add(obj Object) error {
	// TODO: 구현하세요
	// 힌트: KeyFunc로 키 생성, items에 저장, 인덱스 업데이트
	panic("Add: 아직 구현되지 않았습니다")
}

// Update는 기존 오브젝트를 갱신합니다.
func (s *store) Update(obj Object) error {
	// TODO: 구현하세요
	// 힌트: 기존 오브젝트의 인덱스 제거 후 새 오브젝트로 교체
	panic("Update: 아직 구현되지 않았습니다")
}

// Delete는 오브젝트를 스토어에서 제거합니다.
func (s *store) Delete(obj Object) error {
	// TODO: 구현하세요
	panic("Delete: 아직 구현되지 않았습니다")
}

// Get은 이름으로 오브젝트를 조회합니다.
func (s *store) Get(key string) (Object, bool) {
	// TODO: 구현하세요
	panic("Get: 아직 구현되지 않았습니다")
}

// List는 모든 오브젝트를 반환합니다.
func (s *store) List() []Object {
	// TODO: 구현하세요
	panic("List: 아직 구현되지 않았습니다")
}

// AddIndexer는 인덱서를 등록합니다.
// 예: store.AddIndexer("byApp", func(obj Object) []string { return []string{obj.Labels["app"]} })
func (s *store) AddIndexer(name string, fn IndexFunc) {
	// TODO: 구현하세요
	panic("AddIndexer: 아직 구현되지 않았습니다")
}

// ListByIndex는 인덱스 값으로 오브젝트 목록을 반환합니다.
func (s *store) ListByIndex(indexName, indexValue string) ([]Object, error) {
	// TODO: 구현하세요
	panic("ListByIndex: 아직 구현되지 않았습니다")
}

// Replace는 스토어 전체를 새 목록으로 교체합니다. (Resync 시 사용)
func (s *store) Replace(objects []Object) error {
	// TODO: 구현하세요
	// 힌트: 기존 items/indices 모두 초기화 후 objects를 하나씩 Add
	panic("Replace: 아직 구현되지 않았습니다")
}

// updateIndices는 오브젝트의 모든 인덱스를 갱신합니다.
// mu는 이미 잠긴 상태여야 합니다.
func (s *store) updateIndices(key string, obj Object) {
	// TODO: 구현하세요
	panic("updateIndices: 아직 구현되지 않았습니다")
}

// deleteIndices는 오브젝트의 모든 인덱스를 제거합니다.
// mu는 이미 잠긴 상태여야 합니다.
func (s *store) deleteIndices(key string, obj Object) {
	// TODO: 구현하세요
	panic("deleteIndices: 아직 구현되지 않았습니다")
}

// ============================================================
// Reflector 구현
// ============================================================

// Reflector는 소스(ListFunc+WatchFunc)에서 오브젝트를 가져와
// Store에 반영합니다. client-go Reflector의 단순화 버전입니다.
type Reflector struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - store: *store
	//   - listFunc: ListFunc
	//   - watchFunc: WatchFunc
	//   - hasSynced: atomic bool (List 완료 후 true)
}

// NewReflector는 새 Reflector를 생성합니다.
func NewReflector(s *store, lf ListFunc, wf WatchFunc) *Reflector {
	// TODO: 구현하세요
	panic("NewReflector: 아직 구현되지 않았습니다")
}

// Run은 List → Watch 루프를 실행합니다.
// 1. ListFunc로 전체 목록을 가져와 store.Replace 호출 → hasSynced = true
// 2. WatchFunc로 이벤트 채널 획득 후 이벤트에 따라 store 업데이트
// ctx 취소 시 반환합니다.
func (r *Reflector) Run(ctx context.Context) {
	// TODO: 구현하세요
	// 힌트:
	//   - WatchEvent.Type == "ADDED" → store.Add
	//   - WatchEvent.Type == "MODIFIED" → store.Update
	//   - WatchEvent.Type == "DELETED" → store.Delete
	panic("Run: 아직 구현되지 않았습니다")
}

// HasSynced는 초기 List가 완료되었는지 반환합니다.
func (r *Reflector) HasSynced() bool {
	// TODO: 구현하세요
	panic("HasSynced: 아직 구현되지 않았습니다")
}

// ============================================================
// Informer 구현
// ============================================================

// Informer는 Reflector + Store + EventHandler를 조합합니다.
// client-go의 SharedIndexInformer를 단순화한 버전입니다.
type Informer struct {
	// TODO: 필요한 필드를 추가하세요.
	// 힌트:
	//   - store: *store
	//   - reflector: *Reflector
	//   - handlers: []EventHandler
	//   - mu: sync.RWMutex (handlers 보호)
}

// NewInformer는 새 Informer를 생성합니다.
// ListFunc와 WatchFunc를 받아 내부에 Reflector를 생성합니다.
func NewInformer(lf ListFunc, wf WatchFunc) *Informer {
	// TODO: 구현하세요
	// 힌트: store와 Reflector를 만들되, Reflector는 이벤트 발생 시
	// EventHandler 콜백도 호출하도록 연결해야 합니다.
	// → store의 Add/Update/Delete를 래핑하거나,
	//   Reflector가 직접 핸들러를 호출하도록 설계하세요.
	panic("NewInformer: 아직 구현되지 않았습니다")
}

// AddEventHandler는 이벤트 핸들러를 등록합니다.
func (inf *Informer) AddEventHandler(h EventHandler) {
	// TODO: 구현하세요
	panic("AddEventHandler: 아직 구현되지 않았습니다")
}

// Run은 Reflector를 시작하고 ctx 취소 시까지 실행합니다.
func (inf *Informer) Run(ctx context.Context) {
	// TODO: 구현하세요
	panic("Run: 아직 구현되지 않았습니다")
}

// HasSynced는 초기 동기화가 완료되었는지 반환합니다.
func (inf *Informer) HasSynced() bool {
	// TODO: 구현하세요
	panic("HasSynced: 아직 구현되지 않았습니다")
}

// GetStore는 내부 Store를 반환합니다 (Lister 생성용).
func (inf *Informer) GetStore() *store {
	// TODO: 구현하세요
	panic("GetStore: 아직 구현되지 않았습니다")
}

// ============================================================
// Lister 구현
// ============================================================

// Lister는 Store의 읽기 전용 뷰를 제공합니다.
// client-go의 GenericLister를 단순화한 버전입니다.
type Lister struct {
	// TODO: 필요한 필드를 추가하세요.
	mu    sync.RWMutex // 사용하지 않아도 됩니다
	store *store
}

// NewLister는 새 Lister를 생성합니다.
func NewLister(s *store) *Lister {
	// TODO: 구현하세요
	panic("NewLister: 아직 구현되지 않았습니다")
}

// List는 모든 오브젝트를 반환합니다.
func (l *Lister) List() []Object {
	// TODO: 구현하세요
	panic("Lister.List: 아직 구현되지 않았습니다")
}

// Get은 키로 오브젝트를 조회합니다.
func (l *Lister) Get(key string) (Object, bool) {
	// TODO: 구현하세요
	panic("Lister.Get: 아직 구현되지 않았습니다")
}

// ListByLabel은 레이블 키=값 으로 필터링한 오브젝트 목록을 반환합니다.
func (l *Lister) ListByLabel(labelKey, labelValue string) []Object {
	// TODO: 구현하세요
	// 힌트: List() 결과를 순회하며 obj.Labels[labelKey] == labelValue 필터
	panic("ListByLabel: 아직 구현되지 않았습니다")
}

// ============================================================
// 레이블 인덱서 헬퍼
// ============================================================

// LabelIndexFunc는 레이블 키로 인덱싱하는 IndexFunc를 생성합니다.
func LabelIndexFunc(labelKey string) IndexFunc {
	return func(obj Object) []string {
		if val, ok := obj.Labels[labelKey]; ok {
			return []string{val}
		}
		return nil
	}
}

// NamespaceIndexFunc는 네임스페이스로 인덱싱합니다.
var NamespaceIndexFunc IndexFunc = func(obj Object) []string {
	if obj.Namespace == "" {
		return nil
	}
	return []string{obj.Namespace}
}

// ============================================================
// 패키지 레벨 헬퍼
// ============================================================

// splitKey는 "namespace/name" 키를 분리합니다.
func splitKey(key string) (namespace, name string) {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", parts[0]
}

func main() {
	// 테스트를 실행하세요: go test ./... -v -race
}
