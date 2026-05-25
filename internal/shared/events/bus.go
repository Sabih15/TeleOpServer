package events

import (
	"context"
	"sync"
)

// Event is the envelope every inter-module message is wrapped in.
// Type follows "module.action" convention, e.g. "user.registered", "order.placed".
type Event struct {
	Type    string // identifies what happened
	Payload any    // the data — cast to the expected struct in the handler
}

// Bus is the inter-module communication contract.
//
// Rules:
//   - Modules publish events instead of calling each other directly.
//   - Modules subscribe to events they care about in their own startup.
//   - No module ever imports another module's package.
//
// Swap strategy: replace InProcessBus with a Kafka/NATS/RabbitMQ implementation
// of this same interface when extracting a module into its own microservice.
// The publishing/subscribing code inside each module stays unchanged.
type Bus interface {
	Publish(ctx context.Context, e Event)
	Subscribe(eventType string, fn func(Event))
}

// InProcessBus is the default in-memory implementation for the monolith.
// Handlers run synchronously in the same goroutine as the publisher.
type InProcessBus struct {
	mu       sync.RWMutex
	handlers map[string][]func(Event)
}

func NewInProcessBus() *InProcessBus {
	return &InProcessBus{
		handlers: make(map[string][]func(Event)),
	}
}

// Subscribe registers a handler for a specific event type.
// Safe to call concurrently.
func (b *InProcessBus) Subscribe(eventType string, fn func(Event)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], fn)
}

// Publish delivers the event to all subscribers registered for its type.
// If no subscriber exists the event is silently dropped (fire-and-forget).
func (b *InProcessBus) Publish(_ context.Context, e Event) {
	b.mu.RLock()
	fns := make([]func(Event), len(b.handlers[e.Type]))
	copy(fns, b.handlers[e.Type])
	b.mu.RUnlock()

	for _, fn := range fns {
		fn(e)
	}
}
