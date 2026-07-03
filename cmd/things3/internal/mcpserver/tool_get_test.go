package mcpserver

import (
	"context"
	"testing"

	"github.com/moond4rk/things3/thingstest"
)

func get(t *testing.T, srv *Server, id string) GetResult {
	t.Helper()
	_, res, err := srv.handleGet(context.Background(), nil, GetInput{ID: id})
	if err != nil {
		t.Fatalf("get %q: %v", id, err)
	}
	return res
}

func TestGetTodoByUUIDAndPrefix(t *testing.T) {
	srv := newTestServer(t, Config{})
	for _, id := range []string{thingstest.UUIDTodoInToday, thingstest.UUIDTodoInToday[:8]} {
		got := get(t, srv, id)
		if !got.Success || got.Item == nil {
			t.Fatalf("get %q should succeed, got %+v", id, got)
		}
		if got.Item.UUID != thingstest.UUIDTodoInToday {
			t.Errorf("get %q resolved to %s", id, got.Item.UUID)
		}
		if got.Item.Type != typeTodo {
			t.Errorf("get %q type = %q", id, got.Item.Type)
		}
	}
}

func TestGetTodoChecklist(t *testing.T) {
	srv := newTestServer(t, Config{})
	got := get(t, srv, thingstest.UUIDTodoChecklist)
	if got.Item == nil {
		t.Fatalf("get checklist todo failed: %+v", got)
	}
	if len(got.Item.Checklist) == 0 {
		t.Errorf("todo %s should carry a checklist", thingstest.UUIDTodoChecklist)
	}
}

func TestGetProjectNesting(t *testing.T) {
	srv := newTestServer(t, Config{})
	client := testClient(t)
	ctx := context.Background()

	wantTodos, err := client.Todos().InProject(thingstest.UUIDProject).Status().Incomplete().All(ctx)
	if err != nil {
		t.Fatalf("expected todos: %v", err)
	}
	wantHeadings, err := client.Headings().InProject(thingstest.UUIDProject).All(ctx)
	if err != nil {
		t.Fatalf("expected headings: %v", err)
	}

	got := get(t, srv, thingstest.UUIDProject)
	if !got.Success || got.Item == nil || got.Item.Type != typeProject {
		t.Fatalf("get project failed: %+v", got)
	}
	if len(got.Todos) != len(wantTodos) {
		t.Errorf("nested todos = %d, want %d", len(got.Todos), len(wantTodos))
	}
	if len(got.Headings) != len(wantHeadings) {
		t.Errorf("nested headings = %d, want %d", len(got.Headings), len(wantHeadings))
	}
}

func TestGetNotFound(t *testing.T) {
	srv := newTestServer(t, Config{})
	got := get(t, srv, "zzznope")
	if got.Success {
		t.Errorf("want success=false")
	}
	if got.Error == nil || got.Error.Code != codeNotFound {
		t.Errorf("want not_found error, got %+v", got.Error)
	}
}

func TestGetAmbiguous(t *testing.T) {
	srv := newTestServer(t, Config{})
	got := get(t, srv, "Project in")
	if got.Success {
		t.Errorf("want success=false")
	}
	if got.Error == nil || got.Error.Code != codeAmbiguous {
		t.Fatalf("want ambiguous error, got %+v", got.Error)
	}
	if len(got.Error.Candidates) < 2 {
		t.Errorf("ambiguous error should carry >= 2 candidates, got %+v", got.Error.Candidates)
	}
	for _, c := range got.Error.Candidates {
		if len(c.UUID) < 20 {
			t.Errorf("candidate UUID should be full, got %q", c.UUID)
		}
	}
}
