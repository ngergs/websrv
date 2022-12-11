package sync

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

// Get retries a value from the synchronized map and returns success status and the value (or the nil-value of the given type if ok=false)
func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok = m.kv[key]
	return
}
