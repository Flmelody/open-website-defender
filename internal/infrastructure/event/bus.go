package event

import "sync"

type Event string

type Handler func(Event, any)

type bus struct {
	mu       sync.RWMutex
	handlers map[Event][]Handler
}

var (
	instance *bus
	once     sync.Once
)

func Bus() *bus {
	once.Do(func() {
		instance = &bus{
			handlers: make(map[Event][]Handler),
		}
	})
	return instance
}

// Subscribe registers a handler for the given event.
func (b *bus) Subscribe(event Event, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[event] = append(b.handlers[event], handler)
}

// Publish fires the event, calling all registered handlers synchronously.
// Optional data payload is passed to handlers.
func (b *bus) Publish(event Event, data ...any) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers[event]))
	copy(handlers, b.handlers[event])
	b.mu.RUnlock()

	var payload any
	if len(data) > 0 {
		payload = data[0]
	}

	for _, h := range handlers {
		h(event, payload)
	}
}
