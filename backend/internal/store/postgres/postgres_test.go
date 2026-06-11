package postgres

import (
	"strings"
	"testing"
)

func TestOrderClause(t *testing.T) {
	cases := []struct {
		sort, order string
		contains    string
	}{
		{"created_at", "desc", "t.created_at DESC"},
		{"created_at", "asc", "t.created_at ASC"},
		{"due_date", "asc", "t.due_date ASC NULLS LAST"},
		{"due_date", "desc", "t.due_date DESC NULLS LAST"},
		{"priority", "desc", "CASE t.priority"},
		{"bogus", "sideways", "t.created_at DESC"}, // defensive default
	}
	for _, tc := range cases {
		got := orderClause(tc.sort, tc.order)
		if !strings.Contains(got, tc.contains) {
			t.Errorf("orderClause(%q, %q) = %q, want it to contain %q",
				tc.sort, tc.order, got, tc.contains)
		}
	}
}

func TestEscapeLike(t *testing.T) {
	cases := map[string]string{
		"plain":     "plain",
		"100%":      `100\%`,
		"under_bar": `under\_bar`,
		`back\slash`: `back\\slash`,
	}
	for in, want := range cases {
		if got := escapeLike(in); got != want {
			t.Errorf("escapeLike(%q) = %q, want %q", in, got, want)
		}
	}
}
