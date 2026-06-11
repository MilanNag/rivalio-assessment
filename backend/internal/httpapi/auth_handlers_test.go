package httpapi

import (
	"net/http"
	"testing"

	"github.com/milann/taskflow/internal/models"
)

func TestSignup(t *testing.T) {
	t.Run("creates user and returns token", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodPost, "/api/auth/signup", "",
			map[string]string{"email": "alice@example.com", "password": "password123"})

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeBody[authResponse](t, rec)
		if resp.Token == "" {
			t.Error("expected a token")
		}
		if resp.User.Email != "alice@example.com" {
			t.Errorf("expected email alice@example.com, got %s", resp.User.Email)
		}
	})

	t.Run("first user becomes admin, second is regular user", func(t *testing.T) {
		_, _, h := newTestServer(t)
		first := decodeBody[authResponse](t, doRequest(t, h, http.MethodPost, "/api/auth/signup", "",
			map[string]string{"email": "admin@example.com", "password": "password123"}))
		second := decodeBody[authResponse](t, doRequest(t, h, http.MethodPost, "/api/auth/signup", "",
			map[string]string{"email": "user@example.com", "password": "password123"}))

		if first.User.Role != models.RoleAdmin {
			t.Errorf("expected first user to be admin, got %s", first.User.Role)
		}
		if second.User.Role != models.RoleUser {
			t.Errorf("expected second user to be user, got %s", second.User.Role)
		}
	})

	t.Run("rejects invalid email", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodPost, "/api/auth/signup", "",
			map[string]string{"email": "not-an-email", "password": "password123"})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
		resp := decodeBody[errorEnvelope](t, rec)
		if resp.Error.Fields["email"] == "" {
			t.Error("expected email field error")
		}
	})

	t.Run("rejects short password", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodPost, "/api/auth/signup", "",
			map[string]string{"email": "alice@example.com", "password": "short"})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		_, st, h := newTestServer(t)
		createTestUser(t, st, "alice@example.com", models.RoleUser)
		rec := doRequest(t, h, http.MethodPost, "/api/auth/signup", "",
			map[string]string{"email": "alice@example.com", "password": "password123"})
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodPost, "/api/auth/signup", "", "{not json")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})
}

func TestLogin(t *testing.T) {
	t.Run("returns token for valid credentials", func(t *testing.T) {
		_, st, h := newTestServer(t)
		createTestUser(t, st, "alice@example.com", models.RoleUser)

		rec := doRequest(t, h, http.MethodPost, "/api/auth/login", "",
			map[string]string{"email": "alice@example.com", "password": "password123"})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeBody[authResponse](t, rec)
		if resp.Token == "" {
			t.Error("expected a token")
		}
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		_, st, h := newTestServer(t)
		createTestUser(t, st, "alice@example.com", models.RoleUser)

		rec := doRequest(t, h, http.MethodPost, "/api/auth/login", "",
			map[string]string{"email": "alice@example.com", "password": "wrong-password"})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("rejects unknown email with same status as wrong password", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodPost, "/api/auth/login", "",
			map[string]string{"email": "ghost@example.com", "password": "password123"})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("rejects empty body fields", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodPost, "/api/auth/login", "", map[string]string{})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})
}

func TestMe(t *testing.T) {
	t.Run("returns current user", func(t *testing.T) {
		_, st, h := newTestServer(t)
		user, token := createTestUser(t, st, "alice@example.com", models.RoleUser)

		rec := doRequest(t, h, http.MethodGet, "/api/auth/me", token, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		resp := decodeBody[map[string]models.User](t, rec)
		if resp["user"].ID != user.ID {
			t.Errorf("expected user %s, got %s", user.ID, resp["user"].ID)
		}
	})

	t.Run("rejects missing token", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodGet, "/api/auth/me", "", nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("rejects garbage token", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodGet, "/api/auth/me", "garbage.token.value", nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}
