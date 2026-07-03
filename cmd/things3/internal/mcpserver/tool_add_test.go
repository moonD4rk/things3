package mcpserver

import (
	"context"
	"strings"
	"testing"

	"github.com/moond4rk/things3/thingstest"
)

func addTodo(t *testing.T, srv *Server, in AddTodoInput) WriteResult {
	t.Helper()
	_, res, err := srv.handleAddTodo(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("add_todo: %v", err)
	}
	return res
}

func addProject(t *testing.T, srv *Server, in AddProjectInput) WriteResult {
	t.Helper()
	_, res, err := srv.handleAddProject(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("add_project: %v", err)
	}
	return res
}

func TestAddTodoURL(t *testing.T) {
	srv, rec := newWriteServer(t)
	res := addTodo(t, srv, AddTodoInput{
		Title:     "Buy milk",
		When:      "evening",
		Notes:     "at the store",
		Tags:      []string{"home", "errand"},
		Checklist: []string{"one", "two"},
		Deadline:  "2026-08-01",
	})
	if !res.Success {
		t.Fatalf("add_todo failed: %+v", res.Error)
	}
	if res.Verified {
		t.Errorf("a novel title cannot be verified against the fixture")
	}
	u := rec.last(t)
	if u.Path != "/add" {
		t.Fatalf("path = %q, want /add", u.Path)
	}
	q := u.Query()
	for k, want := range map[string]string{
		"title": "Buy milk", "when": "evening", "notes": "at the store",
		"tags": "home,errand", "deadline": "2026-08-01",
	} {
		if q.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, q.Get(k), want)
		}
	}
	if !strings.Contains(q.Get("checklist-items"), "one") || !strings.Contains(q.Get("checklist-items"), "two") {
		t.Errorf("checklist-items = %q", q.Get("checklist-items"))
	}
}

func TestAddTodoReminderInWhen(t *testing.T) {
	srv, rec := newWriteServer(t)
	addTodo(t, srv, AddTodoInput{Title: "Wake", Reminder: "09:30"})
	if got := rec.last(t).Query().Get("when"); !strings.Contains(got, "09:30") {
		t.Errorf("reminder should fold into when, got %q", got)
	}
}

func TestAddTodoInvalidInputs(t *testing.T) {
	srv, _ := newWriteServer(t)
	cases := []struct {
		name string
		in   AddTodoInput
	}{
		{"bad when", AddTodoInput{Title: "x", When: "tdoay"}},
		{"bad deadline", AddTodoInput{Title: "x", Deadline: "not-a-date"}},
		{"bad reminder", AddTodoInput{Title: "x", Reminder: "25:00"}},
		{"project and area", AddTodoInput{Title: "x", Project: "Work", Area: "Home"}},
		{"heading without project", AddTodoInput{Title: "x", Heading: "Heading"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := addTodo(t, srv, tc.in)
			if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
				t.Errorf("want invalid_input, got %+v", res)
			}
		})
	}
}

func TestAddTodoPlacement(t *testing.T) {
	srv, rec := newWriteServer(t)

	t.Run("project resolves to full uuid", func(t *testing.T) {
		addTodo(t, srv, AddTodoInput{Title: "Email Bob", Project: "Project in Area 1"})
		if got := rec.last(t).Query().Get("list-id"); got != thingstest.UUIDProject {
			t.Errorf("list-id = %q, want %q", got, thingstest.UUIDProject)
		}
	})

	t.Run("heading within project", func(t *testing.T) {
		addTodo(t, srv, AddTodoInput{Title: "Sub-task", Project: "Project in Area 1", Heading: "Heading"})
		q := rec.last(t).Query()
		if q.Get("list-id") != thingstest.UUIDProject {
			t.Errorf("list-id = %q", q.Get("list-id"))
		}
		if q.Get("heading-id") == "" {
			t.Errorf("expected a heading-id")
		}
	})

	t.Run("area resolves to full uuid", func(t *testing.T) {
		want := fixtureAreaUUID(t, "Area 1")
		addTodo(t, srv, AddTodoInput{Title: "Chore", Area: "Area 1"})
		if got := rec.last(t).Query().Get("list-id"); got != want {
			t.Errorf("list-id = %q, want %q", got, want)
		}
	})

	t.Run("ambiguous project rides the envelope", func(t *testing.T) {
		res := addTodo(t, srv, AddTodoInput{Title: "x", Project: "Project in"})
		if res.Success || res.Error == nil || res.Error.Code != codeAmbiguous {
			t.Errorf("want ambiguous, got %+v", res)
		}
	})
}

func TestAddProjectURL(t *testing.T) {
	srv, rec := newWriteServer(t)
	res := addProject(t, srv, AddProjectInput{
		Title:    "Website",
		Area:     "Area 1",
		Todos:    []string{"Flights", "Hotel"},
		Deadline: "2026-08-01",
	})
	if !res.Success {
		t.Fatalf("add_project failed: %+v", res.Error)
	}
	u := rec.last(t)
	if u.Path != "/add-project" {
		t.Fatalf("path = %q, want /add-project", u.Path)
	}
	q := u.Query()
	if q.Get("title") != "Website" {
		t.Errorf("title = %q", q.Get("title"))
	}
	if q.Get("area-id") == "" {
		t.Errorf("expected area-id for --area")
	}
	if q.Get("deadline") != "2026-08-01" {
		t.Errorf("deadline = %q", q.Get("deadline"))
	}
	if !strings.Contains(q.Get("to-dos"), "Flights") || !strings.Contains(q.Get("to-dos"), "Hotel") {
		t.Errorf("to-dos = %q", q.Get("to-dos"))
	}
}

// TestAddTodoInvalidTagIsInvalidInput proves a build-time validation failure -- a
// tag containing a comma -- reports invalid_input (not execution_failed) and sends
// nothing, so the machine-parseable error steers a caller to fix the input.
func TestAddTodoInvalidTagIsInvalidInput(t *testing.T) {
	srv, rec := newWriteServer(t)
	res := addTodo(t, srv, AddTodoInput{Title: "x", Tags: []string{"a,b"}})
	if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
		t.Fatalf("want invalid_input, got %+v", res)
	}
	if rec.count() != 0 {
		t.Errorf("invalid input must not execute a URL")
	}
}
