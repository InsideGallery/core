package utils

import "sync"

type SafeMap[K string, V any] struct {
	data map[K]V
	mu   *sync.RWMutex
}

func NewSafeMap[K string, V any](data map[K]V) *SafeMap[K, V] {
	s := &SafeMap[K, V]{
		data: map[K]V{},
		mu:   &sync.RWMutex{},
	}
	for k, v := range data {
		s.data[k] = v
	}

	return s
}

func (s *SafeMap[K, V]) Set(name K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		s.data = map[K]V{}
	}

	s.data[name] = value
}

func (s *SafeMap[K, V]) Get(name K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[name]

	return v, ok
}

func (s *SafeMap[K, V]) Exists(name K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[name]

	return ok
}

func (s *SafeMap[K, V]) Remove(name K) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, name)
}

func (s *SafeMap[K, V]) GetMap() map[K]V {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := map[K]V{}
	for k, v := range s.data {
		result[k] = v
	}

	return result
}
