package workflow

import (
	"encoding/json"
	"fmt"
	"sync"
)

// State 不是全局类型安全容器，只是节点间共享的执行状态。
type State struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewState() *State {
	return &State{data: make(map[string]any)}
}

func (s *State) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *State) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key]
	return value, ok
}

func (s *State) Bind(key string, dst any) error {
	value, ok := s.Get(key)
	if !ok {
		return fmt.Errorf("state key not found: %s", key)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal state value: %w", err)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("bind state value: %w", err)
	}
	return nil
}
