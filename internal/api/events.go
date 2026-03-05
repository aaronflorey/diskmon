package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Event struct {
	Type      string    `json:"type"`
	Device    string    `json:"device,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type EventBroker struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

func NewEventBroker() *EventBroker {
	return &EventBroker{
		subs: make(map[chan Event]struct{}),
	}
}

func (b *EventBroker) Subscribe() (chan Event, func()) {
	ch := make(chan Event, 16)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		if _, ok := b.subs[ch]; ok {
			delete(b.subs, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
}

func (b *EventBroker) Publish(eventType string, device string) {
	if b == nil {
		return
	}
	ev := Event{
		Type:      eventType,
		Device:    device,
		Timestamp: time.Now().UTC(),
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, ev Event) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\n", ev.Type); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}
