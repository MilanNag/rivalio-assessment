package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/milann/taskflow/internal/models"
)

func seedTask(t *testing.T, st *fakeStore, userID, title, status, priority string, due *time.Time) *models.Task {
	t.Helper()
	task, err := st.CreateTask(context.Background(), &models.Task{
		UserID:   userID,
		Title:    title,
		Status:   status,
		Priority: priority,
		DueDate:  due,
	})
	if err != nil {
		t.Fatalf("seed task: %v", err)
	}
	return task
}

func TestCreateTask(t *testing.T) {
	t.Run("creates task with defaults", func(t *testing.T) {
		_, st, h := newTestServer(t)
		_, token := createTestUser(t, st, "alice@example.com", models.RoleUser)

		rec := doRequest(t, h, http.MethodPost, "/api/tasks", token,
			map[string]any{"title": "Write report"})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeBody[taskEnvelope](t, rec)
		if resp.Data.Title != "Write report" {
			t.Errorf("title = %q", resp.Data.Title)
		}
		if resp.Data.Status != models.StatusTodo || resp.Data.Priority != models.PriorityMedium {
			t.Errorf("expected default status/priority, got %s/%s", resp.Data.Status, resp.Data.Priority)
		}
	})

	t.Run("creates task with all fields and records activity", func(t *testing.T) {
		_, st, h := newTestServer(t)
		_, token := createTestUser(t, st, "alice@example.com", models.RoleUser)

		rec := doRequest(t, h, http.MethodPost, "/api/tasks", token, map[string]any{
			"title":       "Ship release",
			"description": "v2 launch",
			"status":      "in_progress",
			"priority":    "high",
			"dueDate":     "2026-07-01T00:00:00Z",
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeBody[taskEnvelope](t, rec)
		if resp.Data.DueDate == nil {
			t.Fatal("expected due date to be set")
		}
		logs, _ := st.ListActivity(context.Background(), resp.Data.ID)
		if len(logs) != 1 || logs[0].Action != "created" {
			t.Errorf("expected one 'created' activity entry, got %+v", logs)
		}
	})

	t.Run("accepts date-only due date", func(t *testing.T) {
		_, st, h := newTestServer(t)
		_, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		rec := doRequest(t, h, http.MethodPost, "/api/tasks", token,
			map[string]any{"title": "Task", "dueDate": "2026-07-01"})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("validation failures", func(t *testing.T) {
		cases := []struct {
			name  string
			body  map[string]any
			field string
		}{
			{"missing title", map[string]any{}, "title"},
			{"blank title", map[string]any{"title": "   "}, "title"},
			{"title too long", map[string]any{"title": string(make([]byte, 201))}, "title"},
			{"bad status", map[string]any{"title": "x", "status": "later"}, "status"},
			{"bad priority", map[string]any{"title": "x", "priority": "urgent"}, "priority"},
			{"bad due date", map[string]any{"title": "x", "dueDate": "next week"}, "dueDate"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				_, st, h := newTestServer(t)
				_, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
				rec := doRequest(t, h, http.MethodPost, "/api/tasks", token, tc.body)
				if rec.Code != http.StatusUnprocessableEntity {
					t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
				}
				resp := decodeBody[errorEnvelope](t, rec)
				if resp.Error.Fields[tc.field] == "" {
					t.Errorf("expected error for field %q, got %+v", tc.field, resp.Error.Fields)
				}
			})
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		_, _, h := newTestServer(t)
		rec := doRequest(t, h, http.MethodPost, "/api/tasks", "", map[string]any{"title": "x"})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})
}

func TestListTasks(t *testing.T) {
	t.Run("returns only own tasks with pagination meta", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		bob, _ := createTestUser(t, st, "bob@example.com", models.RoleUser)
		seedTask(t, st, alice.ID, "Mine", models.StatusTodo, models.PriorityLow, nil)
		seedTask(t, st, bob.ID, "Bobs", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodGet, "/api/tasks", token, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		resp := decodeBody[taskListEnvelope](t, rec)
		if len(resp.Data) != 1 || resp.Data[0].Title != "Mine" {
			t.Errorf("expected only own task, got %+v", resp.Data)
		}
		if resp.Meta.Total != 1 || resp.Meta.Page != 1 {
			t.Errorf("unexpected meta %+v", resp.Meta)
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		seedTask(t, st, alice.ID, "A", models.StatusTodo, models.PriorityLow, nil)
		seedTask(t, st, alice.ID, "B", models.StatusDone, models.PriorityLow, nil)

		resp := decodeBody[taskListEnvelope](t,
			doRequest(t, h, http.MethodGet, "/api/tasks?status=done", token, nil))
		if len(resp.Data) != 1 || resp.Data[0].Title != "B" {
			t.Errorf("expected only done task, got %+v", resp.Data)
		}
	})

	t.Run("searches by title", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		seedTask(t, st, alice.ID, "Buy groceries", models.StatusTodo, models.PriorityLow, nil)
		seedTask(t, st, alice.ID, "Write code", models.StatusTodo, models.PriorityLow, nil)

		resp := decodeBody[taskListEnvelope](t,
			doRequest(t, h, http.MethodGet, "/api/tasks?q=groc", token, nil))
		if len(resp.Data) != 1 || resp.Data[0].Title != "Buy groceries" {
			t.Errorf("expected search match, got %+v", resp.Data)
		}
	})

	t.Run("sorts by priority descending", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		seedTask(t, st, alice.ID, "low", models.StatusTodo, models.PriorityLow, nil)
		seedTask(t, st, alice.ID, "high", models.StatusTodo, models.PriorityHigh, nil)
		seedTask(t, st, alice.ID, "medium", models.StatusTodo, models.PriorityMedium, nil)

		resp := decodeBody[taskListEnvelope](t,
			doRequest(t, h, http.MethodGet, "/api/tasks?sort=priority&order=desc", token, nil))
		got := []string{resp.Data[0].Title, resp.Data[1].Title, resp.Data[2].Title}
		want := []string{"high", "medium", "low"}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("expected order %v, got %v", want, got)
			}
		}
	})

	t.Run("combines filter, search and sort", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		due1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		due2 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		seedTask(t, st, alice.ID, "report alpha", models.StatusTodo, models.PriorityLow, &due1)
		seedTask(t, st, alice.ID, "report beta", models.StatusTodo, models.PriorityLow, &due2)
		seedTask(t, st, alice.ID, "report done", models.StatusDone, models.PriorityLow, nil)
		seedTask(t, st, alice.ID, "other", models.StatusTodo, models.PriorityLow, nil)

		resp := decodeBody[taskListEnvelope](t, doRequest(t, h, http.MethodGet,
			"/api/tasks?status=todo&q=report&sort=due_date&order=asc", token, nil))
		if len(resp.Data) != 2 {
			t.Fatalf("expected 2 tasks, got %d", len(resp.Data))
		}
		if resp.Data[0].Title != "report beta" {
			t.Errorf("expected earliest due first, got %+v", resp.Data)
		}
	})

	t.Run("paginates", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		for i := 0; i < 15; i++ {
			seedTask(t, st, alice.ID, fmt.Sprintf("task %02d", i), models.StatusTodo, models.PriorityLow, nil)
		}

		resp := decodeBody[taskListEnvelope](t,
			doRequest(t, h, http.MethodGet, "/api/tasks?page=2&limit=10", token, nil))
		if len(resp.Data) != 5 {
			t.Errorf("expected 5 tasks on page 2, got %d", len(resp.Data))
		}
		if resp.Meta.Total != 15 || resp.Meta.TotalPages != 2 {
			t.Errorf("unexpected meta %+v", resp.Meta)
		}
	})

	t.Run("rejects invalid query params", func(t *testing.T) {
		_, st, h := newTestServer(t)
		_, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		for _, q := range []string{"status=bogus", "sort=bogus", "order=sideways", "page=0", "page=abc", "limit=101"} {
			rec := doRequest(t, h, http.MethodGet, "/api/tasks?"+q, token, nil)
			if rec.Code != http.StatusUnprocessableEntity {
				t.Errorf("query %q: expected 422, got %d", q, rec.Code)
			}
		}
	})

	t.Run("admin can list all tasks with all=true", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, _ := createTestUser(t, st, "alice@example.com", models.RoleUser)
		admin, adminToken := createTestUser(t, st, "admin@example.com", models.RoleAdmin)
		seedTask(t, st, alice.ID, "Alices", models.StatusTodo, models.PriorityLow, nil)
		seedTask(t, st, admin.ID, "Admins", models.StatusTodo, models.PriorityLow, nil)

		resp := decodeBody[taskListEnvelope](t,
			doRequest(t, h, http.MethodGet, "/api/tasks?all=true", adminToken, nil))
		if len(resp.Data) != 2 {
			t.Errorf("expected admin to see 2 tasks, got %d", len(resp.Data))
		}
	})

	t.Run("regular user cannot use all=true", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, _ := createTestUser(t, st, "alice@example.com", models.RoleUser)
		bob, bobToken := createTestUser(t, st, "bob@example.com", models.RoleUser)
		seedTask(t, st, alice.ID, "Alices", models.StatusTodo, models.PriorityLow, nil)
		seedTask(t, st, bob.ID, "Bobs", models.StatusTodo, models.PriorityLow, nil)

		resp := decodeBody[taskListEnvelope](t,
			doRequest(t, h, http.MethodGet, "/api/tasks?all=true", bobToken, nil))
		if len(resp.Data) != 1 || resp.Data[0].Title != "Bobs" {
			t.Errorf("expected only own task despite all=true, got %+v", resp.Data)
		}
	})
}

func TestGetTask(t *testing.T) {
	t.Run("returns own task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "Mine", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodGet, "/api/tasks/"+task.ID, token, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("hides other users' tasks as 404", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, _ := createTestUser(t, st, "alice@example.com", models.RoleUser)
		_, bobToken := createTestUser(t, st, "bob@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "Mine", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodGet, "/api/tasks/"+task.ID, bobToken, nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("admin can view any task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, _ := createTestUser(t, st, "alice@example.com", models.RoleUser)
		_, adminToken := createTestUser(t, st, "admin@example.com", models.RoleAdmin)
		task := seedTask(t, st, alice.ID, "Mine", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodGet, "/api/tasks/"+task.ID, adminToken, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("404 for unknown id", func(t *testing.T) {
		_, st, h := newTestServer(t)
		_, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		rec := doRequest(t, h, http.MethodGet, "/api/tasks/00000000-0000-0000-0000-000000000000", token, nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}

func TestUpdateTask(t *testing.T) {
	t.Run("patches provided fields only", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "Original", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodPatch, "/api/tasks/"+task.ID, token,
			map[string]any{"status": "done"})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		resp := decodeBody[taskEnvelope](t, rec)
		if resp.Data.Status != models.StatusDone {
			t.Errorf("status = %s", resp.Data.Status)
		}
		if resp.Data.Title != "Original" {
			t.Errorf("title should be unchanged, got %q", resp.Data.Title)
		}
	})

	t.Run("clears due date with empty string", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		due := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, &due)

		rec := doRequest(t, h, http.MethodPatch, "/api/tasks/"+task.ID, token,
			map[string]any{"dueDate": ""})
		resp := decodeBody[taskEnvelope](t, rec)
		if resp.Data.DueDate != nil {
			t.Errorf("expected due date cleared, got %v", resp.Data.DueDate)
		}
	})

	t.Run("records activity describing changes", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

		doRequest(t, h, http.MethodPatch, "/api/tasks/"+task.ID, token,
			map[string]any{"status": "in_progress", "priority": "high"})
		logs, _ := st.ListActivity(context.Background(), task.ID)
		if len(logs) != 1 || logs[0].Action != "updated" {
			t.Fatalf("expected one updated entry, got %+v", logs)
		}
	})

	t.Run("no-op patch records no activity", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodPatch, "/api/tasks/"+task.ID, token,
			map[string]any{"status": "todo"})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		logs, _ := st.ListActivity(context.Background(), task.ID)
		if len(logs) != 0 {
			t.Errorf("expected no activity, got %+v", logs)
		}
	})

	t.Run("cannot update another user's task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, _ := createTestUser(t, st, "alice@example.com", models.RoleUser)
		_, bobToken := createTestUser(t, st, "bob@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "Mine", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodPatch, "/api/tasks/"+task.ID, bobToken,
			map[string]any{"title": "Hijacked"})
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("rejects invalid patch values", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodPatch, "/api/tasks/"+task.ID, token,
			map[string]any{"status": "archived"})
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", rec.Code)
		}
	})
}

func TestDeleteTask(t *testing.T) {
	t.Run("deletes own task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodDelete, "/api/tasks/"+task.ID, token, nil)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
		if _, err := st.GetTask(context.Background(), task.ID); err == nil {
			t.Error("expected task to be deleted")
		}
	})

	t.Run("cannot delete another user's task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, _ := createTestUser(t, st, "alice@example.com", models.RoleUser)
		_, bobToken := createTestUser(t, st, "bob@example.com", models.RoleUser)
		task := seedTask(t, st, alice.ID, "Mine", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodDelete, "/api/tasks/"+task.ID, bobToken, nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("admin can delete any task", func(t *testing.T) {
		_, st, h := newTestServer(t)
		alice, _ := createTestUser(t, st, "alice@example.com", models.RoleUser)
		_, adminToken := createTestUser(t, st, "admin@example.com", models.RoleAdmin)
		task := seedTask(t, st, alice.ID, "Mine", models.StatusTodo, models.PriorityLow, nil)

		rec := doRequest(t, h, http.MethodDelete, "/api/tasks/"+task.ID, adminToken, nil)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}

func TestListActivityEndpoint(t *testing.T) {
	_, st, h := newTestServer(t)
	alice, token := createTestUser(t, st, "alice@example.com", models.RoleUser)
	task := seedTask(t, st, alice.ID, "T", models.StatusTodo, models.PriorityLow, nil)

	doRequest(t, h, http.MethodPatch, "/api/tasks/"+task.ID, token, map[string]any{"status": "done"})

	rec := doRequest(t, h, http.MethodGet, "/api/tasks/"+task.ID+"/activity", token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	resp := decodeBody[map[string][]models.ActivityLog](t, rec)
	if len(resp["data"]) != 1 {
		t.Errorf("expected 1 activity entry, got %d", len(resp["data"]))
	}
}
