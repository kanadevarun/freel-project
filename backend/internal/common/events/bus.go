package events

import (
	"sync"
	"time"
)

// EventType defines the type of the event.
type EventType string

// Event represents a system event.
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// EventHandler defines the signature for an event handler function.
type EventHandler func(event Event)

// Bus represents an event bus for pub/sub messaging.
type Bus interface {
	Publish(event Event)
	Subscribe(eventType EventType, handler EventHandler)
}

type inProcessBus struct {
	handlers map[EventType][]EventHandler
	mu       sync.RWMutex
}

// NewInProcessBus creates a new in-memory event bus.
func NewInProcessBus() Bus {
	return &inProcessBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Publish publishes an event to all subscribers.
func (b *inProcessBus) Publish(event Event) {
	b.mu.RLock()
	handlers, ok := b.handlers[event.Type]
	b.mu.RUnlock()

	if !ok {
		return
	}

	// Execute handlers asynchronously
	for _, handler := range handlers {
		go handler(event)
	}
}

// Subscribe subscribes a handler to a specific event type.
func (b *inProcessBus) Subscribe(eventType EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}
