package event

import "testing"

func TestSubscribeAndUnsubscribe(t *testing.T) {
	b := &bus{
		handlers: make(map[Event]map[uint64]Handler),
	}

	var called int
	unsubscribe := b.Subscribe("test:event", func(_ Event, _ any) {
		called++
	})

	b.Publish("test:event")
	if called != 1 {
		t.Fatalf("expected handler to be called once, got %d", called)
	}

	unsubscribe()
	b.Publish("test:event")
	if called != 1 {
		t.Fatalf("expected handler to stay unsubscribed, got %d calls", called)
	}
}
