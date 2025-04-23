package ds

import "sync"

// SafeSet is a thread-safe wrapper around a Set.
type SafeSet[T comparable] struct {
	mu  sync.RWMutex
	set *Set[T]
}

// New creates a new SafeSet.
func NewSafeSet[T comparable]() *SafeSet[T] {
	return &SafeSet[T]{set: NewSet[T]()}
}

// Add inserts an item into the set.
func (s *SafeSet[T]) Add(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Add(item)
}

// Delete removes an item from the set.
func (s *SafeSet[T]) Delete(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Delete(item)
}

// Items returns all items in the set as a slice.
func (s *SafeSet[T]) Items() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Items()
}

// Has checks whether the item exists in the set.
func (s *SafeSet[T]) Has(item T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Has(item)
}

// Len returns the number of items in the set.
func (s *SafeSet[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Len()
}
