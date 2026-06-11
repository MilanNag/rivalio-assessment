package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/milann/taskflow/internal/auth"
	"github.com/milann/taskflow/internal/config"
	"github.com/milann/taskflow/internal/models"
	"github.com/milann/taskflow/internal/realtime"
)

const testSecret = "test-secret-test-secret-test-secret!"

func newTestServer(t *testing.T) (*Server, *fakeStore, http.Handler) {
	t.Helper()
	st := newFakeStore()
	cfg := &config.Config{
		Port:           "0",
		JWTSecret:      testSecret,
		JWTExpiry:      time.Hour,
		UploadDir:      t.TempDir(),
		MaxUploadBytes: 1 << 20,
		AllowedOrigins: []string{"*"},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := NewServer(cfg, st, realtime.NewHub(), logger)
	return srv, st, srv.Routes()
}

// createTestUser inserts a user directly into the fake store and returns the
// user plus a valid bearer token.
func createTestUser(t *testing.T, st *fakeStore, email, role string) (*models.User, string) {
	t.Helper()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user, err := st.CreateUser(t.Context(), email, hash, role)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	token, err := auth.GenerateToken(testSecret, user.ID, user.Email, user.Role, time.Hour)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return user, token
}

func doRequest(t *testing.T, handler http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return out
}

type taskEnvelope struct {
	Data models.Task `json:"data"`
}

type taskListEnvelope struct {
	Data []models.Task `json:"data"`
	Meta listMeta      `json:"meta"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}
