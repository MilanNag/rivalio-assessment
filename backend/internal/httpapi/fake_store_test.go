package httpapi

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/milann/taskflow/internal/models"
	"github.com/milann/taskflow/internal/store"
)

// fakeStore is an in-memory store.Store used for handler unit tests.
type fakeStore struct {
	mu          sync.Mutex
	users       map[string]*models.User
	tasks       map[string]*models.Task
	attachments map[string]*models.Attachment
	activity    []models.ActivityLog
	nextLogID   int64

	// failNext forces the next call of the named method to fail, letting
	// tests exercise 500 paths.
	failNext map[string]error
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users:       map[string]*models.User{},
		tasks:       map[string]*models.Task{},
		attachments: map[string]*models.Attachment{},
		failNext:    map[string]error{},
	}
}

func (f *fakeStore) fail(method string) error {
	if err, ok := f.failNext[method]; ok {
		delete(f.failNext, method)
		return err
	}
	return nil
}

func (f *fakeStore) CreateUser(_ context.Context, email, passwordHash, role string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("CreateUser"); err != nil {
		return nil, err
	}
	for _, u := range f.users {
		if u.Email == email {
			return nil, store.ErrDuplicateEmail
		}
	}
	u := &models.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
	}
	f.users[u.ID] = u
	return u, nil
}

func (f *fakeStore) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("GetUserByEmail"); err != nil {
		return nil, err
	}
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, store.ErrNotFound
}

func (f *fakeStore) GetUserByID(_ context.Context, id string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("GetUserByID"); err != nil {
		return nil, err
	}
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, store.ErrNotFound
}

func (f *fakeStore) CountUsers(_ context.Context) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("CountUsers"); err != nil {
		return 0, err
	}
	return len(f.users), nil
}

func (f *fakeStore) CreateTask(_ context.Context, task *models.Task) (*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("CreateTask"); err != nil {
		return nil, err
	}
	now := time.Now()
	t := *task
	t.ID = uuid.NewString()
	t.CreatedAt = now
	t.UpdatedAt = now
	if u, ok := f.users[t.UserID]; ok {
		t.UserEmail = u.Email
	}
	f.tasks[t.ID] = &t
	out := t
	return &out, nil
}

func (f *fakeStore) ListTasks(_ context.Context, filter models.TaskFilter) ([]models.Task, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("ListTasks"); err != nil {
		return nil, 0, err
	}

	matched := []models.Task{}
	for _, t := range f.tasks {
		if filter.UserID != "" && t.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(filter.Search)) {
			continue
		}
		matched = append(matched, *t)
	}

	priorityRank := map[string]int{"low": 1, "medium": 2, "high": 3}
	sort.Slice(matched, func(i, j int) bool {
		a, b := matched[i], matched[j]
		var less bool
		switch filter.Sort {
		case "priority":
			less = priorityRank[a.Priority] < priorityRank[b.Priority]
		case "due_date":
			switch {
			case a.DueDate == nil:
				return false
			case b.DueDate == nil:
				return true
			default:
				less = a.DueDate.Before(*b.DueDate)
			}
		default:
			less = a.CreatedAt.Before(b.CreatedAt)
		}
		if filter.Order == "desc" {
			return !less
		}
		return less
	})

	total := len(matched)
	start := (filter.Page - 1) * filter.Limit
	if start > total {
		start = total
	}
	end := start + filter.Limit
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

func (f *fakeStore) GetTask(_ context.Context, id string) (*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("GetTask"); err != nil {
		return nil, err
	}
	if t, ok := f.tasks[id]; ok {
		out := *t
		return &out, nil
	}
	return nil, store.ErrNotFound
}

func (f *fakeStore) UpdateTask(_ context.Context, task *models.Task) (*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("UpdateTask"); err != nil {
		return nil, err
	}
	existing, ok := f.tasks[task.ID]
	if !ok {
		return nil, store.ErrNotFound
	}
	t := *task
	t.CreatedAt = existing.CreatedAt
	t.UpdatedAt = time.Now()
	f.tasks[t.ID] = &t
	out := t
	return &out, nil
}

func (f *fakeStore) DeleteTask(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("DeleteTask"); err != nil {
		return err
	}
	if _, ok := f.tasks[id]; !ok {
		return store.ErrNotFound
	}
	delete(f.tasks, id)
	return nil
}

func (f *fakeStore) CreateAttachment(_ context.Context, a *models.Attachment) (*models.Attachment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("CreateAttachment"); err != nil {
		return nil, err
	}
	a.ID = uuid.NewString()
	a.CreatedAt = time.Now()
	copied := *a
	f.attachments[a.ID] = &copied
	return a, nil
}

func (f *fakeStore) ListAttachments(_ context.Context, taskID string) ([]models.Attachment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("ListAttachments"); err != nil {
		return nil, err
	}
	out := []models.Attachment{}
	for _, a := range f.attachments {
		if a.TaskID == taskID {
			out = append(out, *a)
		}
	}
	return out, nil
}

func (f *fakeStore) GetAttachment(_ context.Context, id string) (*models.Attachment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("GetAttachment"); err != nil {
		return nil, err
	}
	if a, ok := f.attachments[id]; ok {
		out := *a
		return &out, nil
	}
	return nil, store.ErrNotFound
}

func (f *fakeStore) DeleteAttachment(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("DeleteAttachment"); err != nil {
		return err
	}
	if _, ok := f.attachments[id]; !ok {
		return store.ErrNotFound
	}
	delete(f.attachments, id)
	return nil
}

func (f *fakeStore) CreateActivity(_ context.Context, log *models.ActivityLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("CreateActivity"); err != nil {
		return err
	}
	f.nextLogID++
	log.ID = f.nextLogID
	log.CreatedAt = time.Now()
	f.activity = append(f.activity, *log)
	return nil
}

func (f *fakeStore) ListActivity(_ context.Context, taskID string) ([]models.ActivityLog, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.fail("ListActivity"); err != nil {
		return nil, err
	}
	out := []models.ActivityLog{}
	for _, l := range f.activity {
		if l.TaskID == taskID {
			out = append(out, l)
		}
	}
	return out, nil
}
