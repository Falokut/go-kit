package ds

// Set is a basic non-thread-safe generic set.
type Set[T comparable] struct {
	items map[T]struct{}
}

// New creates a new Set.
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{items: make(map[T]struct{})}
}

// Add inserts an item into the set.
func (s *Set[T]) Add(item T) {
	s.items[item] = struct{}{}
}

// Delete removes an item from the set.
func (s *Set[T]) Delete(item T) {
	delete(s.items, item)
}

// Items returns all items in the set as a slice.
func (s *Set[T]) Items() []T {
	result := make([]T, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	return result
}

// Has checks whether the item exists in the set.
func (s *Set[T]) Has(item T) bool {
	_, ok := s.items[item]
	return ok
}

// Len returns the number of items in the set.
func (s *Set[T]) Len() int {
	return len(s.items)
}
