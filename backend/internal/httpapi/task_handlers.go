package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/milann/taskflow/internal/models"
	"github.com/milann/taskflow/internal/realtime"
	"github.com/milann/taskflow/internal/store"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

var validSorts = []string{"created_at", "due_date", "priority"}

type listMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())

	var in taskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	dueDate, _, fields := validateTaskInput(in, true)
	if fields != nil {
		writeValidationError(w, fields)
		return
	}

	task := &models.Task{
		UserID:   claims.UserID,
		Title:    strings.TrimSpace(*in.Title),
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		DueDate:  dueDate,
	}
	if in.Description != nil {
		task.Description = *in.Description
	}
	if in.Status != nil {
		task.Status = *in.Status
	}
	if in.Priority != nil {
		task.Priority = *in.Priority
	}

	created, err := s.store.CreateTask(r.Context(), task)
	if err != nil {
		s.logger.Error("create task", "error", err)
		writeInternalError(w)
		return
	}

	s.recordActivity(r, created.ID, "created", fmt.Sprintf("Created task %q", created.Title))
	s.hub.Publish(realtime.Event{Type: "task.created", OwnerID: created.UserID, ActorID: claims.UserID, Payload: created})

	writeJSON(w, http.StatusCreated, map[string]any{"data": created})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())
	q := r.URL.Query()

	filter := models.TaskFilter{
		UserID: claims.UserID,
		Status: q.Get("status"),
		Search: strings.TrimSpace(q.Get("q")),
		Sort:   q.Get("sort"),
		Order:  q.Get("order"),
		Page:   1,
		Limit:  defaultPageSize,
	}

	// Admins may request a global view across all users.
	if claims.Role == models.RoleAdmin && q.Get("all") == "true" {
		filter.UserID = ""
	}

	if filter.Status != "" && !slices.Contains(models.ValidStatuses, filter.Status) {
		writeValidationError(w, map[string]string{"status": "Status must be one of: todo, in_progress, done."})
		return
	}
	if filter.Sort == "" {
		filter.Sort = "created_at"
	} else if !slices.Contains(validSorts, filter.Sort) {
		writeValidationError(w, map[string]string{"sort": "Sort must be one of: created_at, due_date, priority."})
		return
	}
	if filter.Order == "" {
		filter.Order = "desc"
	} else if filter.Order != "asc" && filter.Order != "desc" {
		writeValidationError(w, map[string]string{"order": "Order must be asc or desc."})
		return
	}

	if p := q.Get("page"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 {
			writeValidationError(w, map[string]string{"page": "Page must be a positive integer."})
			return
		}
		filter.Page = n
	}
	if l := q.Get("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n < 1 || n > maxPageSize {
			writeValidationError(w, map[string]string{"limit": "Limit must be between 1 and 100."})
			return
		}
		filter.Limit = n
	}

	tasks, total, err := s.store.ListTasks(r.Context(), filter)
	if err != nil {
		s.logger.Error("list tasks", "error", err)
		writeInternalError(w)
		return
	}

	totalPages := (total + filter.Limit - 1) / filter.Limit
	writeJSON(w, http.StatusOK, map[string]any{
		"data": tasks,
		"meta": listMeta{Page: filter.Page, Limit: filter.Limit, Total: total, TotalPages: totalPages},
	})
}

// loadTaskAuthorized fetches a task and enforces that the requester owns it
// (or is an admin). Writes the error response itself when returning nil.
func (s *Server) loadTaskAuthorized(w http.ResponseWriter, r *http.Request) *models.Task {
	claims := claimsFrom(r.Context())
	id := chi.URLParam(r, "id")

	task, err := s.store.GetTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Task not found.")
			return nil
		}
		s.logger.Error("get task", "error", err)
		writeInternalError(w)
		return nil
	}

	if task.UserID != claims.UserID && claims.Role != models.RoleAdmin {
		// Hide the task's existence from other users.
		writeError(w, http.StatusNotFound, "not_found", "Task not found.")
		return nil
	}
	return task
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	task := s.loadTaskAuthorized(w, r)
	if task == nil {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": task})
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())
	task := s.loadTaskAuthorized(w, r)
	if task == nil {
		return
	}

	var in taskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	dueDate, dueDateSet, fields := validateTaskInput(in, false)
	if fields != nil {
		writeValidationError(w, fields)
		return
	}

	changes := applyTaskPatch(task, in, dueDate, dueDateSet)
	if len(changes) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"data": task})
		return
	}

	updated, err := s.store.UpdateTask(r.Context(), task)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Task not found.")
			return
		}
		s.logger.Error("update task", "error", err)
		writeInternalError(w)
		return
	}

	s.recordActivity(r, updated.ID, "updated", strings.Join(changes, "; "))
	s.hub.Publish(realtime.Event{Type: "task.updated", OwnerID: updated.UserID, ActorID: claims.UserID, Payload: updated})

	writeJSON(w, http.StatusOK, map[string]any{"data": updated})
}

// applyTaskPatch mutates task with the provided fields and returns
// human-readable change descriptions for the activity log.
func applyTaskPatch(task *models.Task, in taskInput, dueDate *time.Time, dueDateSet bool) []string {
	var changes []string

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title != task.Title {
			changes = append(changes, fmt.Sprintf("Title changed from %q to %q", task.Title, title))
			task.Title = title
		}
	}
	if in.Description != nil && *in.Description != task.Description {
		changes = append(changes, "Description updated")
		task.Description = *in.Description
	}
	if in.Status != nil && *in.Status != task.Status {
		changes = append(changes, fmt.Sprintf("Status changed from %s to %s", task.Status, *in.Status))
		task.Status = *in.Status
	}
	if in.Priority != nil && *in.Priority != task.Priority {
		changes = append(changes, fmt.Sprintf("Priority changed from %s to %s", task.Priority, *in.Priority))
		task.Priority = *in.Priority
	}
	if dueDateSet && !equalTimePtr(task.DueDate, dueDate) {
		changes = append(changes, fmt.Sprintf("Due date changed from %s to %s",
			formatTimePtr(task.DueDate), formatTimePtr(dueDate)))
		task.DueDate = dueDate
	}
	return changes
}

func equalTimePtr(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return "none"
	}
	return t.Format("2006-01-02")
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	claims := claimsFrom(r.Context())
	task := s.loadTaskAuthorized(w, r)
	if task == nil {
		return
	}

	if err := s.store.DeleteTask(r.Context(), task.ID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Task not found.")
			return
		}
		s.logger.Error("delete task", "error", err)
		writeInternalError(w)
		return
	}

	s.hub.Publish(realtime.Event{Type: "task.deleted", OwnerID: task.UserID, ActorID: claims.UserID, Payload: map[string]string{"id": task.ID}})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListActivity(w http.ResponseWriter, r *http.Request) {
	task := s.loadTaskAuthorized(w, r)
	if task == nil {
		return
	}

	logs, err := s.store.ListActivity(r.Context(), task.ID)
	if err != nil {
		s.logger.Error("list activity", "error", err)
		writeInternalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": logs})
}

// recordActivity persists an activity entry; failures are logged but never
// block the main operation.
func (s *Server) recordActivity(r *http.Request, taskID, action, detail string) {
	claims := claimsFrom(r.Context())
	entry := &models.ActivityLog{
		TaskID:    taskID,
		UserID:    claims.UserID,
		UserEmail: claims.Email,
		Action:    action,
		Detail:    detail,
	}
	if err := s.store.CreateActivity(r.Context(), entry); err != nil {
		s.logger.Error("record activity", "error", err, "task_id", taskID)
	}
}