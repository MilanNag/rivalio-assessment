package httpapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/milann/taskflow/internal/models"
)

// handleEvents streams task events to the client using Server-Sent Events.
// Clients authenticate via ?access_token= since EventSource cannot set headers.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming_unsupported", "Streaming is not supported.")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe := s.hub.Subscribe(claims.UserID, claims.Role == models.RoleAdmin)
	defer unsubscribe()

	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case msg, open := <-events:
			if !open {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}
