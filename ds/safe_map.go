package ds

import "sync"

// SafeMap is a thread-safe wrapper around a Map.
type SafeMap[K comparable, V any] struct {
	mu sync.RWMutex
	v  map[K]V
}

// New creates a new SafeMap.
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{v: make(map[K]V)}
}

// Get return an value with key from the map.
func (s *SafeMap[K, V]) Get(key K) V {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.v[key]
}

// GetWithLookup return an value and do lookup from the map.
func (s *SafeMap[K, V]) GetWithLookup(key K) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.v[key]
	return val, ok
}

// Add inserts an key with value into the map.
func (s *SafeMap[K, V]) Add(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.v[key] = value
}

// Delete removes an key from the map.
func (s *SafeMap[K, V]) Delete(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.v, key)
}

// Items returns all values in the map as slices.
func (s *SafeMap[K, V]) Values() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]V, 0, len(s.v))
	for _, value := range s.v {
		values = append(values, value)
	}
	return values
}

// Items returns all keys in the map as slices.
func (s *SafeMap[K, V]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]K, 0, len(s.v))
	for key := range s.v {
		keys = append(keys, key)
	}
	return keys
}

// Has checks whether the key exists in the map.
func (s *SafeMap[K, V]) Has(key K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.v[key]
	return ok
}

// Len returns the number of items in the map.
func (s *SafeMap[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.v)
}
