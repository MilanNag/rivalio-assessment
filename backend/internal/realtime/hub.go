package realtime

import (
	"encoding/json"
	"sync"
)

// Event is a realtime notification pushed to connected clients over SSE.
type Event struct {
	Type    string `json:"type"` // task.created | task.updated | task.deleted
	OwnerID string `json:"-"`    // task owner; controls fan-out
	ActorID string `json:"-"`    // user who performed the change
	Payload any    `json:"payload"`
}

type subscriber struct {
	userID  string
	isAdmin bool
	ch      chan []byte
}

// Hub fans out task events to SSE subscribers. Regular users receive events
// for their own tasks; admins receive everything.
type Hub struct {
	mu   sync.RWMutex
	subs map[*subscriber]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[*subscriber]struct{})}
}

// Subscribe registers a client and returns its channel plus an unsubscribe func.
func (h *Hub) Subscribe(userID string, isAdmin bool) (<-chan []byte, func()) {
	sub := &subscriber{userID: userID, isAdmin: isAdmin, ch: make(chan []byte, 16)}
	h.mu.Lock()
	h.subs[sub] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if _, ok := h.subs[sub]; ok {
			delete(h.subs, sub)
			close(sub.ch)
		}
		h.mu.Unlock()
	}
	return sub.ch, unsubscribe
}

// Publish sends the event to every subscriber allowed to see it.
// Slow clients are skipped rather than blocking the publisher.
func (h *Hub) Publish(event Event) {
	body, err := json.Marshal(map[string]any{
		"type":    event.Type,
		"payload": event.Payload,
		"actorId": event.ActorID,
	})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.subs {
		if !sub.isAdmin && sub.userID != event.OwnerID {
			continue
		}
		select {
		case sub.ch <- body:
		default:
		}
	}
}

// SubscriberCount is used by tests and health diagnostics.
func (h *Hub) SubscriberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs)
}
