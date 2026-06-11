package httpapi

import (
	"errors"
	"net/http"
	"testing"

	"github.com/milann/taskflow/internal/models"
)

// Exercises the 500 branches by forcing store failures.
func TestStoreFailuresReturn500(t *testing.T) {
	boom := errors.New("db down")

	t.Run("create task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		_, token := createTestUser(t, st, "a@b.com", models.RoleUser)
		st.failNext["CreateTask"] = boom
		rec := doRequest(t, h, http.MethodPost, "/api/tasks", token, map[string]any{"title": "x"})
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
		resp := decodeBody[errorEnvelope](t, rec)
		if resp.Error.Code != "internal_error" {
			t.Errorf("code = %s", resp.Error.Code)
		}
	})

	t.Run("list tasks", func(t *testing.T) {
		_, st, h := newTestServer(t)
		_, token := createTestUser(t, st, "a@b.com", models.RoleUser)
		st.failNext["ListTasks"] = boom
		rec := doRequest(t, h, http.MethodGet, "/api/tasks", token, nil)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("get task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "a@b.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)
		st.failNext["GetTask"] = boom
		rec := doRequest(t, h, http.MethodGet, "/api/tasks/"+task.ID, token, nil)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("update task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "a@b.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)
		st.failNext["UpdateTask"] = boom
		rec := doRequest(t, h, http.MethodPatch, "/api/tasks/"+task.ID, token, map[string]any{"status": "done"})
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("delete task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "a@b.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)
		st.failNext["DeleteTask"] = boom
		rec := doRequest(t, h, http.MethodDelete, "/api/tasks/"+task.ID, token, nil)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("signup", func(t *testing.T) {
		_, st, h := newTestServer(t)
		st.failNext["CreateUser"] = boom
		rec := doRequest(t, h, http.MethodPost, "/api/auth/signup", "",
			map[string]string{"email": "a@b.com", "password": "password123"})
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("login", func(t *testing.T) {
		_, st, h := newTestServer(t)
		createTestUser(t, st, "a@b.com", models.RoleUser)
		st.failNext["GetUserByEmail"] = boom
		rec := doRequest(t, h, http.MethodPost, "/api/auth/login", "",
			map[string]string{"email": "a@b.com", "password": "password123"})
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("list activity", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "a@b.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)
		st.failNext["ListActivity"] = boom
		rec := doRequest(t, h, http.MethodGet, "/api/tasks/"+task.ID+"/activity", token, nil)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("list attachments", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "a@b.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)
		st.failNext["ListAttachments"] = boom
		rec := doRequest(t, h, http.MethodGet, "/api/tasks/"+task.ID+"/attachments", token, nil)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestMeWithDeletedAccount(t *testing.T) {
	_, st, h := newTestServer(t)
	_, token := createTestUser(t, st, "ghost@b.com", models.RoleUser)
	// Simulate the account being removed after the token was issued.
	for id := range st.users {
		delete(st.users, id)
	}
	rec := doRequest(t, h, http.MethodGet, "/api/auth/me", token, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	_, _, h := newTestServer(t)
	rec := doRequest(t, h, http.MethodGet, "/healthz", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
