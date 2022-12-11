package syncwrap

import "sync"

// Map is a map safe for asynchronous use.
// It is just a ordinary map wrapped with a RWMutex.
type Map[K comparable, V any] struct {
	kv map[K]V
	mu sync.RWMutex
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{kv: make(map[K]V)}
}

// Set the value in the synchronized map
func (m *Map[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kv[key] = value
}

// Get retries a value from the synchronized map and returns success status and the value (or the zero-value of the given type if ok=false)
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok = m.kv[key]
	return
}

// SyncMap is a map safe for asynchronous use.
// Typefull wrapper around sync.Map
type SyncMap[K comparable, V any] struct {
	kv sync.Map
}

// Load returns the value from the sync.Map.
// If ok=false the zero-value of the value-type is returned as value.
func (m *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	var mapval any
	mapval, ok = m.kv.Load(key)
	if !ok {
		return
	}
	return mapval.(V), ok
}

// LoadOrStore saves the value to the sync.Map if it is not already present.
// Returns the value currently present in the sync.Map.
// The return value loaded is true if the value has been loaded from the map and not been stored.
func (m *SyncMap[K, V]) LoadOrStore(key K, value V) (result V, loaded bool) {
	var mapval any
	mapval, loaded = m.kv.LoadOrStore(key, value)
	return mapval.(V), loaded
}
