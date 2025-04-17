package llmango

import (
	"maps"
	"sync"
)

type SyncedMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func (m *SyncedMap[K, V]) Get(key K) V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.m[key]
}

func (m *SyncedMap[K, V]) Exists(key K) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.m[key]
	return exists
}

func (m *SyncedMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	// Ensure the map is initialized
	if m.m == nil {
		m.m = make(map[K]V)
	}
	m.m[key] = value
	m.mu.Unlock()
}

func (m *SyncedMap[K, V]) Delete(key K) {
	m.mu.Lock()
	if m.m != nil { // Check if map is initialized before deleting
		delete(m.m, key)
	}
	m.mu.Unlock()
}

// GetAll returns a copy of all key-value pairs in the map.
func (m *SyncedMap[K, V]) GetAll() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid returning the internal map directly
	copiedMap := make(map[K]V, len(m.m))
	maps.Copy(copiedMap, m.m) // Use maps.Copy for efficiency
	return copiedMap
}
