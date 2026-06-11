package models

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"

	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"

	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

var (
	ValidStatuses   = []string{StatusTodo, StatusInProgress, StatusDone}
	ValidPriorities = []string{PriorityLow, PriorityMedium, PriorityHigh}
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Task struct {
	ID          string     `json:"id"`
	UserID      string     `json:"userId"`
	UserEmail   string     `json:"userEmail,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"dueDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Attachment struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"taskId"`
	FileName    string    `json:"fileName"`
	StoredName  string    `json:"-"`
	ContentType string    `json:"contentType"`
	SizeBytes   int64     `json:"sizeBytes"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ActivityLog struct {
	ID        int64     `json:"id"`
	TaskID    string    `json:"taskId"`
	UserID    string    `json:"userId"`
	UserEmail string    `json:"userEmail"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createdAt"`
}

// TaskFilter captures the supported list query parameters.
type TaskFilter struct {
	// UserID scopes results to a single owner. Empty means all users (admin only).
	UserID string
	Status string
	Search string
	Sort   string // due_date | priority | created_at
	Order  string // asc | desc
	Page   int
	Limit  int
}
