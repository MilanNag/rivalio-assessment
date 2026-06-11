package realtime

import (
	"encoding/json"
	"testing"
	"time"
)

func receive(t *testing.T, ch <-chan []byte) map[string]any {
	t.Helper()
	select {
	case msg := <-ch:
		var out map[string]any
		if err := json.Unmarshal(msg, &out); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}
		return out
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
		return nil
	}
}

func expectNoEvent(t *testing.T, ch <-chan []byte) {
	t.Helper()
	select {
	case msg := <-ch:
		t.Fatalf("unexpected event: %s", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHubDeliversToOwner(t *testing.T) {
	hub := NewHub()
	ch, unsubscribe := hub.Subscribe("user-1", false)
	defer unsubscribe()

	hub.Publish(Event{Type: "task.created", OwnerID: "user-1", Payload: map[string]string{"id": "t1"}})

	event := receive(t, ch)
	if event["type"] != "task.created" {
		t.Errorf("unexpected event %+v", event)
	}
}

func TestHubDoesNotLeakAcrossUsers(t *testing.T) {
	hub := NewHub()
	ch, unsubscribe := hub.Subscribe("user-2", false)
	defer unsubscribe()

	hub.Publish(Event{Type: "task.created", OwnerID: "user-1", Payload: nil})
	expectNoEvent(t, ch)
}

func TestHubAdminReceivesAllEvents(t *testing.T) {
	hub := NewHub()
	ch, unsubscribe := hub.Subscribe("admin-1", true)
	defer unsubscribe()

	hub.Publish(Event{Type: "task.updated", OwnerID: "user-1", Payload: nil})
	event := receive(t, ch)
	if event["type"] != "task.updated" {
		t.Errorf("unexpected event %+v", event)
	}
}

func TestHubUnsubscribeRemovesSubscriber(t *testing.T) {
	hub := NewHub()
	_, unsubscribe := hub.Subscribe("user-1", false)
	if hub.SubscriberCount() != 1 {
		t.Fatalf("expected 1 subscriber, got %d", hub.SubscriberCount())
	}
	unsubscribe()
	if hub.SubscriberCount() != 0 {
		t.Fatalf("expected 0 subscribers, got %d", hub.SubscriberCount())
	}
	// Publishing after unsubscribe must not panic.
	hub.Publish(Event{Type: "task.deleted", OwnerID: "user-1"})
	// Double-unsubscribe must be safe.
	unsubscribe()
}

func TestHubSkipsSlowSubscribers(t *testing.T) {
	hub := NewHub()
	ch, unsubscribe := hub.Subscribe("user-1", false)
	defer unsubscribe()

	// Fill the buffer beyond capacity; extra events are dropped, not blocking.
	for i := 0; i < 50; i++ {
		hub.Publish(Event{Type: "task.updated", OwnerID: "user-1"})
	}
	if len(ch) == 0 {
		t.Error("expected buffered events")
	}
}
