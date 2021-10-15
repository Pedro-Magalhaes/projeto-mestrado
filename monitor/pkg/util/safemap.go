package util

import "sync"

type SafeBoolMap struct {
	Mu    sync.Mutex
	Value map[string]bool
}

func NewSafeBoolMap() *SafeBoolMap {
	return new(SafeBoolMap)
}

func (s *SafeBoolMap) GetValue(key string) bool {
	s.Mu.Lock()
	val := s.Value[key]
	s.Mu.Unlock()
	return val
}

func (s *SafeBoolMap) SetValue(key string, val bool) {
	s.Mu.Lock()
	s.Value[key] = val
	s.Mu.Unlock()
}
