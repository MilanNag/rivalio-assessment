package httpapi

import (
	"strings"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestValidateSignup(t *testing.T) {
	cases := []struct {
		name      string
		email     string
		password  string
		wantField string
	}{
		{"valid", "a@b.com", "password123", ""},
		{"empty email", "", "password123", "email"},
		{"bad email", "nope", "password123", "email"},
		{"email with spaces", "a b@c.com", "password123", "email"},
		{"short password", "a@b.com", "1234567", "password"},
		{"too long password", "a@b.com", strings.Repeat("x", 73), "password"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fields := validateSignup(tc.email, tc.password)
			if tc.wantField == "" {
				if fields != nil {
					t.Errorf("expected no errors, got %+v", fields)
				}
				return
			}
			if fields[tc.wantField] == "" {
				t.Errorf("expected error on %q, got %+v", tc.wantField, fields)
			}
		})
	}
}

func TestValidateTaskInput(t *testing.T) {
	t.Run("create requires title", func(t *testing.T) {
		_, _, fields := validateTaskInput(taskInput{}, true)
		if fields["title"] == "" {
			t.Errorf("expected title error, got %+v", fields)
		}
	})

	t.Run("patch does not require title", func(t *testing.T) {
		_, _, fields := validateTaskInput(taskInput{Status: strPtr("done")}, false)
		if fields != nil {
			t.Errorf("expected no errors, got %+v", fields)
		}
	})

	t.Run("parses RFC3339 due date", func(t *testing.T) {
		due, set, fields := validateTaskInput(taskInput{
			Title:   strPtr("x"),
			DueDate: strPtr("2026-07-01T10:00:00Z"),
		}, true)
		if fields != nil {
			t.Fatalf("unexpected errors %+v", fields)
		}
		if !set || due == nil {
			t.Fatal("expected due date to be set")
		}
	})

	t.Run("empty due date string clears it", func(t *testing.T) {
		due, set, fields := validateTaskInput(taskInput{DueDate: strPtr("")}, false)
		if fields != nil {
			t.Fatalf("unexpected errors %+v", fields)
		}
		if !set || due != nil {
			t.Errorf("expected set=true with nil date, got set=%v due=%v", set, due)
		}
	})

	t.Run("rejects unparseable due date", func(t *testing.T) {
		_, _, fields := validateTaskInput(taskInput{Title: strPtr("x"), DueDate: strPtr("tomorrow")}, true)
		if fields["dueDate"] == "" {
			t.Errorf("expected dueDate error, got %+v", fields)
		}
	})

	t.Run("rejects invalid enums", func(t *testing.T) {
		_, _, fields := validateTaskInput(taskInput{
			Title:    strPtr("x"),
			Status:   strPtr("blocked"),
			Priority: strPtr("urgent"),
		}, true)
		if fields["status"] == "" || fields["priority"] == "" {
			t.Errorf("expected status and priority errors, got %+v", fields)
		}
	})

	t.Run("rejects oversized description", func(t *testing.T) {
		long := strings.Repeat("d", 5001)
		_, _, fields := validateTaskInput(taskInput{Title: strPtr("x"), Description: &long}, true)
		if fields["description"] == "" {
			t.Errorf("expected description error, got %+v", fields)
		}
	})
}
