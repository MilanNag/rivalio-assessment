package store

import (
	"context"
	"errors"

	"github.com/milann/taskflow/internal/models"
)

// Sentinel errors returned by store implementations so handlers can map
// them to HTTP status codes without depending on driver-specific errors.
var (
	ErrNotFound       = errors.New("resource not found")
	ErrDuplicateEmail = errors.New("email already registered")
)

type UserStore interface {
	CreateUser(ctx context.Context, email, passwordHash, role string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	CountUsers(ctx context.Context) (int, error)
}

type TaskStore interface {
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	ListTasks(ctx context.Context, filter models.TaskFilter) ([]models.Task, int, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	DeleteTask(ctx context.Context, id string) error
}

type AttachmentStore interface {
	CreateAttachment(ctx context.Context, a *models.Attachment) (*models.Attachment, error)
	ListAttachments(ctx context.Context, taskID string) ([]models.Attachment, error)
	GetAttachment(ctx context.Context, id string) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, id string) error
}

type ActivityStore interface {
	CreateActivity(ctx context.Context, log *models.ActivityLog) error
	ListActivity(ctx context.Context, taskID string) ([]models.ActivityLog, error)
}

// Store aggregates every persistence concern the API needs.
type Store interface {
	UserStore
	TaskStore
	AttachmentStore
	ActivityStore
}
