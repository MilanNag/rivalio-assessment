package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/milann/taskflow/internal/models"
	"github.com/milann/taskflow/internal/store"
)

// Store is the PostgreSQL implementation of store.Store.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ---- Users ----

func (s *Store) CreateUser(ctx context.Context, email, passwordHash, role string) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, role, created_at`,
		email, passwordHash, role,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, store.ErrDuplicateEmail
		}
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	u := &models.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

// ---- Tasks ----

const taskColumns = `t.id, t.user_id, u.email, t.title, t.description, t.status, t.priority, t.due_date, t.created_at, t.updated_at`

func scanTask(row pgx.Row) (*models.Task, error) {
	t := &models.Task{}
	err := row.Scan(&t.ID, &t.UserID, &t.UserEmail, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Store) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	var id string
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tasks (user_id, title, description, status, priority, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		task.UserID, task.Title, task.Description, task.Status, task.Priority, task.DueDate,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return s.GetTask(ctx, id)
}

func (s *Store) GetTask(ctx context.Context, id string) (*models.Task, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT `+taskColumns+` FROM tasks t JOIN users u ON u.id = t.user_id WHERE t.id = $1`, id)
	return scanTask(row)
}

func (s *Store) ListTasks(ctx context.Context, f models.TaskFilter) ([]models.Task, int, error) {
	where := []string{"1=1"}
	args := []any{}
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if f.UserID != "" {
		where = append(where, "t.user_id = "+arg(f.UserID))
	}
	if f.Status != "" {
		where = append(where, "t.status = "+arg(f.Status))
	}
	if f.Search != "" {
		where = append(where, "t.title ILIKE "+arg("%"+escapeLike(f.Search)+"%"))
	}
	whereClause := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tasks t WHERE `+whereClause, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := orderClause(f.Sort, f.Order)
	offset := (f.Page - 1) * f.Limit
	query := `SELECT ` + taskColumns + ` FROM tasks t JOIN users u ON u.id = t.user_id
		WHERE ` + whereClause + ` ORDER BY ` + orderBy +
		` LIMIT ` + arg(f.Limit) + ` OFFSET ` + arg(offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, total, rows.Err()
}

// orderClause maps the validated sort/order inputs onto SQL. Inputs are
// whitelisted by the handler, but default defensively anyway.
func orderClause(sort, order string) string {
	dir := "DESC"
	if strings.EqualFold(order, "asc") {
		dir = "ASC"
	}
	switch sort {
	case "due_date":
		// Keep tasks without a due date at the end in both directions.
		return "t.due_date " + dir + " NULLS LAST, t.created_at DESC"
	case "priority":
		return "CASE t.priority WHEN 'high' THEN 3 WHEN 'medium' THEN 2 ELSE 1 END " + dir + ", t.created_at DESC"
	default:
		return "t.created_at " + dir
	}
}

func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func (s *Store) UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE tasks SET title = $1, description = $2, status = $3, priority = $4,
		 due_date = $5, updated_at = now() WHERE id = $6`,
		task.Title, task.Description, task.Status, task.Priority, task.DueDate, task.ID,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, store.ErrNotFound
	}
	return s.GetTask(ctx, task.ID)
}

func (s *Store) DeleteTask(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

// ---- Attachments ----

func (s *Store) CreateAttachment(ctx context.Context, a *models.Attachment) (*models.Attachment, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO attachments (task_id, file_name, stored_name, content_type, size_bytes)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		a.TaskID, a.FileName, a.StoredName, a.ContentType, a.SizeBytes,
	).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) ListAttachments(ctx context.Context, taskID string) ([]models.Attachment, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, task_id, file_name, stored_name, content_type, size_bytes, created_at
		 FROM attachments WHERE task_id = $1 ORDER BY created_at DESC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Attachment{}
	for rows.Next() {
		var a models.Attachment
		if err := rows.Scan(&a.ID, &a.TaskID, &a.FileName, &a.StoredName,
			&a.ContentType, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetAttachment(ctx context.Context, id string) (*models.Attachment, error) {
	a := &models.Attachment{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, task_id, file_name, stored_name, content_type, size_bytes, created_at
		 FROM attachments WHERE id = $1`, id,
	).Scan(&a.ID, &a.TaskID, &a.FileName, &a.StoredName, &a.ContentType, &a.SizeBytes, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) DeleteAttachment(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM attachments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

// ---- Activity ----

func (s *Store) CreateActivity(ctx context.Context, log *models.ActivityLog) error {
	return s.pool.QueryRow(ctx,
		`INSERT INTO activity_logs (task_id, user_id, user_email, action, detail)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		log.TaskID, log.UserID, log.UserEmail, log.Action, log.Detail,
	).Scan(&log.ID, &log.CreatedAt)
}

func (s *Store) ListActivity(ctx context.Context, taskID string) ([]models.ActivityLog, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, task_id, user_id, user_email, action, detail, created_at
		 FROM activity_logs WHERE task_id = $1 ORDER BY created_at DESC, id DESC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.ActivityLog{}
	for rows.Next() {
		var l models.ActivityLog
		if err := rows.Scan(&l.ID, &l.TaskID, &l.UserID, &l.UserEmail,
			&l.Action, &l.Detail, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
