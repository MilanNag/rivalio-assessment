package httpapi

import (
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/milann/taskflow/internal/models"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

const (
	maxTitleLen       = 200
	maxDescriptionLen = 5000
	minPasswordLen    = 8
	maxPasswordLen    = 72 // bcrypt input limit
)

func validateSignup(email, password string) map[string]string {
	fields := map[string]string{}
	email = strings.TrimSpace(email)
	if email == "" {
		fields["email"] = "Email is required."
	} else if !emailRegex.MatchString(email) || len(email) > 254 {
		fields["email"] = "Email is not a valid address."
	}
	if len(password) < minPasswordLen {
		fields["password"] = "Password must be at least 8 characters."
	} else if len(password) > maxPasswordLen {
		fields["password"] = "Password must be at most 72 characters."
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}

// taskInput uses pointers so PATCH can distinguish "absent" from "set to zero".
type taskInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	DueDate     *string `json:"dueDate"` // RFC3339 or empty string to clear
}

// validateTaskInput validates the provided fields. When isCreate is true,
// title is mandatory. It returns the parsed due date (if provided) so the
// handler doesn't have to parse twice.
func validateTaskInput(in taskInput, isCreate bool) (dueDate *time.Time, dueDateSet bool, fields map[string]string) {
	fields = map[string]string{}

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			fields["title"] = "Title is required."
		} else if len(title) > maxTitleLen {
			fields["title"] = "Title must be at most 200 characters."
		}
	} else if isCreate {
		fields["title"] = "Title is required."
	}

	if in.Description != nil && len(*in.Description) > maxDescriptionLen {
		fields["description"] = "Description must be at most 5000 characters."
	}

	if in.Status != nil && !slices.Contains(models.ValidStatuses, *in.Status) {
		fields["status"] = "Status must be one of: todo, in_progress, done."
	}

	if in.Priority != nil && !slices.Contains(models.ValidPriorities, *in.Priority) {
		fields["priority"] = "Priority must be one of: low, medium, high."
	}

	if in.DueDate != nil {
		dueDateSet = true
		raw := strings.TrimSpace(*in.DueDate)
		if raw != "" {
			parsed, err := parseDueDate(raw)
			if err != nil {
				fields["dueDate"] = "Due date must be an RFC3339 timestamp or YYYY-MM-DD date."
			} else {
				dueDate = &parsed
			}
		}
	}

	if len(fields) == 0 {
		fields = nil
	}
	return dueDate, dueDateSet, fields
}

func parseDueDate(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", raw)
}
