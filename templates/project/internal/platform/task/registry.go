//go:build ignore

package task

import (
	"context"
	"encoding/json"
	"fmt"
)

type Handler interface {
	Handle(ctx context.Context, message Message) (json.RawMessage, error)
}

type HandlerFunc func(ctx context.Context, message Message) (json.RawMessage, error)

func (f HandlerFunc) Handle(ctx context.Context, message Message) (json.RawMessage, error) {
	return f(ctx, message)
}

type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(taskType string, handler Handler) error {
	if taskType == "" {
		return fmt.Errorf("task type is required")
	}
	if handler == nil {
		return fmt.Errorf("handler is required")
	}
	if _, exists := r.handlers[taskType]; exists {
		return fmt.Errorf("task handler already registered: %s", taskType)
	}
	r.handlers[taskType] = handler
	return nil
}

func (r *Registry) Lookup(taskType string) (Handler, bool) {
	handler, ok := r.handlers[taskType]
	return handler, ok
}
