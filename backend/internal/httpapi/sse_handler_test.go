package httpapi

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/milann/taskflow/internal/auth"
	"github.com/milann/taskflow/internal/models"
	"github.com/milann/taskflow/internal/realtime"
)

func TestEventsStream(t *testing.T) {
	srv, st, handler := newTestServer(t)
	_, token := createTestUser(t, st, "alice@example.com", models.RoleUser)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		ts.URL+"/api/events?access_token="+token, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer res.Body.Close()

	if got := res.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("content type = %q", got)
	}

	reader := bufio.NewReader(res.Body)

	// First frame is the connected handshake.
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read handshake: %v", err)
	}
	if !strings.HasPrefix(line, "event: connected") {
		t.Fatalf("expected connected event, got %q", line)
	}

	// Wait for the subscription to be registered, then publish.
	deadline := time.Now().Add(2 * time.Second)
	for srv.hub.SubscriberCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	srv.hub.Publish(realtime.Event{
		Type:    "task.created",
		OwnerID: claimsUserID(t, token),
		Payload: map[string]string{"id": "t1"},
	})

	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read event: %v", err)
		}
		if strings.HasPrefix(line, "data: ") && strings.Contains(line, "task.created") {
			return // success
		}
	}
}

func TestEventsRequiresAuth(t *testing.T) {
	_, _, handler := newTestServer(t)
	rec := doRequest(t, handler, http.MethodGet, "/api/events", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func claimsUserID(t *testing.T, token string) string {
	t.Helper()
	claims, err := auth.ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	return claims.UserID
}
