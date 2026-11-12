package mcpserver

import (
	"context"
	"testing"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

func schedule(t *testing.T, srv *Server, in ScheduleInput) WriteResult {
	t.Helper()
	_, res, err := srv.handleSchedule(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("schedule: %v", err)
	}
	return res
}

func TestScheduleURL(t *testing.T) {
	srv, rec := newWriteServer(t)
	today := things3.Today().Format("2006-01-02")
	cases := []struct{ when, want string }{
		{"evening", "evening"},
		{"someday", "someday"},
		{"anytime", "anytime"},
		{"2026-08-01", "2026-08-01"},
		{"today", today},
	}
	for _, tc := range cases {
		t.Run(tc.when, func(t *testing.T) {
			schedule(t, srv, ScheduleInput{Target: "To-Do in Today", When: tc.when})
			u := rec.last(t)
			if u.Path != "/update" {
				t.Fatalf("path = %q, want /update", u.Path)
			}
			if got := u.Query().Get("when"); got != tc.want {
				t.Errorf("when = %q, want %q", got, tc.want)
			}
			if u.Query().Get("id") != thingstest.UUIDTodoInToday {
				t.Errorf("id = %q", u.Query().Get("id"))
			}
		})
	}
}

func TestScheduleBadWhen(t *testing.T) {
	srv, _ := newWriteServer(t)
	res := schedule(t, srv, ScheduleInput{Target: "To-Do in Today", When: "tdoay"})
	if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
		t.Errorf("want invalid_input, got %+v", res)
	}
}

func TestScheduleProjectTarget(t *testing.T) {
	srv, rec := newWriteServer(t)
	schedule(t, srv, ScheduleInput{Target: "Project in Today", When: "today"})
	if u := rec.last(t); u.Path != "/update-project" {
		t.Errorf("path = %q, want /update-project", u.Path)
	}
}

func TestScheduleReminder(t *testing.T) {
	srv, rec := newWriteServer(t)

	_, res, err := srv.handleSchedule(context.Background(), nil,
		ScheduleInput{Target: "To-Do in Today", When: "2026-08-01", Reminder: "09:00"})
	if err != nil || !res.Success {
		t.Fatalf("schedule: %v %+v", err, res.Error)
	}
	if got := rec.last(t).Query().Get("when"); got != "2026-08-01@09:00" {
		t.Errorf("when = %q, want the day with the reminder", got)
	}

	for name, in := range map[string]ScheduleInput{
		"a reminder needs a day": {Target: "To-Do in Today", When: "someday", Reminder: "09:00"},
		"bad reminder":           {Target: "To-Do in Today", When: "today", Reminder: "nine"},
	} {
		_, res, err := srv.handleSchedule(context.Background(), nil, in)
		if err != nil || res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
			t.Errorf("%s should be invalid_input, got %+v (%v)", name, res, err)
		}
	}
}
