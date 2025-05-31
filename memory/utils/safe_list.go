package utils

import "sync"

type SafeList[V any] struct {
	list []V
	mu   sync.RWMutex
}

func NewSafeList[K any](data ...K) *SafeList[K] {
	s := &SafeList[K]{
		list: data,
	}

	return s
}

func (s *SafeList[V]) Add(value V) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.list = append(s.list, value)
}

func (s *SafeList[V]) List() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dst := make([]V, len(s.list))
	copy(dst, s.list)

	return dst
}

func (s *SafeList[V]) Reset() []V {
	s.mu.Lock()
	defer s.mu.Unlock()

	dst := make([]V, len(s.list))
	copy(dst, s.list)
	s.list = []V{}

	return dst
}

func (s *SafeList[V]) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.list)
}
