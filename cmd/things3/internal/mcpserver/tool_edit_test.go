package mcpserver

import (
	"context"
	"strings"
	"testing"
)

func edit(t *testing.T, srv *Server, in EditInput) WriteResult {
	t.Helper()
	_, res, err := srv.handleEdit(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	return res
}

func TestEditURL(t *testing.T) {
	srv, rec := newWriteServer(t)
	cases := []struct {
		name, param, want string
		in                EditInput
	}{
		{"title", "title", "New title", EditInput{Target: "To-Do in Today", Title: "New title"}},
		{"append_notes", "append-notes", "more", EditInput{Target: "To-Do in Today", AppendNotes: "more"}},
		{"deadline", "deadline", "2026-08-01", EditInput{Target: "To-Do in Today", Deadline: "2026-08-01"}},
		{"add_tags", "add-tags", "urgent", EditInput{Target: "To-Do in Today", AddTags: []string{"urgent"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			edit(t, srv, tc.in)
			u := rec.last(t)
			if u.Path != "/update" {
				t.Fatalf("path = %q", u.Path)
			}
			if got := u.Query().Get(tc.param); got != tc.want {
				t.Errorf("%s = %q, want %q", tc.param, got, tc.want)
			}
		})
	}

	t.Run("clear_deadline sends an empty deadline", func(t *testing.T) {
		edit(t, srv, EditInput{Target: "To-Do in Today", ClearDeadline: true})
		if _, ok := rec.last(t).Query()["deadline"]; !ok {
			t.Errorf("clear_deadline should send an empty deadline param")
		}
	})

	t.Run("reminder preserves the scheduled date and folds in the time", func(t *testing.T) {
		edit(t, srv, EditInput{Target: "To-Do in Today", Reminder: "09:30"})
		got := rec.last(t).Query().Get("when")
		if !strings.Contains(got, "09:30") {
			t.Errorf("reminder 09:30 should fold into when, got %q", got)
		}
		if strings.HasPrefix(got, "today") {
			t.Errorf("reminder must preserve the existing scheduled date, not reschedule to today: %q", got)
		}
	})
}

// TestEditReminderRequiresSchedule proves editing the reminder of an item with no
// concrete scheduled date is rejected as invalid_input instead of silently
// rescheduling it to today.
func TestEditReminderRequiresSchedule(t *testing.T) {
	srv, rec := newWriteServer(t)
	// "To-Do in Someday" sits in the someday bucket with no scheduled date.
	res := edit(t, srv, EditInput{Target: "To-Do in Someday", Reminder: "09:00"})
	if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
		t.Fatalf("want invalid_input, got %+v", res)
	}
	if rec.count() != 0 {
		t.Errorf("a rejected reminder must not execute a URL")
	}
}

func TestEditValidation(t *testing.T) {
	srv, _ := newWriteServer(t)

	t.Run("nothing to edit", func(t *testing.T) {
		res := edit(t, srv, EditInput{Target: "To-Do in Today"})
		if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
			t.Errorf("want invalid_input, got %+v", res)
		}
		if !strings.Contains(res.Error.Message, "nothing to edit") {
			t.Errorf("message = %q", res.Error.Message)
		}
	})

	t.Run("deadline and clear are exclusive", func(t *testing.T) {
		res := edit(t, srv, EditInput{Target: "To-Do in Today", Deadline: "2026-08-01", ClearDeadline: true})
		if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
			t.Errorf("want invalid_input, got %+v", res)
		}
	})

	t.Run("bad deadline", func(t *testing.T) {
		res := edit(t, srv, EditInput{Target: "To-Do in Today", Deadline: "nope"})
		if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
			t.Errorf("want invalid_input, got %+v", res)
		}
	})
}
