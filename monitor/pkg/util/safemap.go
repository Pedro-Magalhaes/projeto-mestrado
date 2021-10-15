package util

import "sync"

type SafeBoolMap struct {
	mu    sync.Mutex
	value map[string]bool
}

func NewSafeBoolMap() *SafeBoolMap {
	return new(SafeBoolMap)
}

func (s *SafeBoolMap) GetValue(key string) bool {
	s.mu.Lock()
	val := s.value[key]
	s.mu.Unlock()
	return val
}

func (s *SafeBoolMap) SetValue(key string, val bool) {
	s.mu.Lock()
	s.value[key] = val
	s.mu.Unlock()
}
