package events

import (
	"context"
	"fmt"
	"sync"
)

type Event interface{ Name() string }
type Handler func(context.Context, Event) error

type Bus interface {
	Publish(context.Context, ...Event) error
	Subscribe(string, Handler)
}

type InProcess struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewInProcess() *InProcess { return &InProcess{handlers: make(map[string][]Handler)} }

func (b *InProcess) Subscribe(name string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[name] = append(b.handlers[name], handler)
}

func (b *InProcess) Publish(ctx context.Context, published ...Event) error {
	for _, event := range published {
		b.mu.RLock()
		handlers := append([]Handler(nil), b.handlers[event.Name()]...)
		b.mu.RUnlock()
		for _, handler := range handlers {
			if err := handler(ctx, event); err != nil {
				return fmt.Errorf("handle %s: %w", event.Name(), err)
			}
		}
	}
	return nil
}
