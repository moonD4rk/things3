package mcpserver

import (
	"context"
	"strings"
	"testing"
)

func move(t *testing.T, srv *Server, in MoveInput) WriteResult {
	t.Helper()
	_, res, err := srv.handleMove(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	return res
}

func TestMoveToProject(t *testing.T) {
	srv, rec := newWriteServer(t)
	want := fixtureProjectUUID(t, "Project in Today")
	move(t, srv, MoveInput{Target: "To-Do in Today", To: "Project in Today"})
	if got := rec.last(t).Query().Get("list-id"); got != want {
		t.Errorf("list-id = %q, want %q", got, want)
	}
}

func TestMoveToArea(t *testing.T) {
	srv, rec := newWriteServer(t)
	// "Area 2" is not a substring of any project title, so this exercises the
	// todo->area fallback cleanly.
	want := fixtureAreaUUID(t, "Area 2")
	move(t, srv, MoveInput{Target: "To-Do in Today", To: "Area 2"})
	if got := rec.last(t).Query().Get("list-id"); got != want {
		t.Errorf("list-id = %q, want %q", got, want)
	}
}

func TestMoveProjectToArea(t *testing.T) {
	srv, rec := newWriteServer(t)
	want := fixtureAreaUUID(t, "Area 2")
	move(t, srv, MoveInput{Target: "Project in Area 1", To: "Area 2"})
	u := rec.last(t)
	if u.Path != "/update-project" || u.Query().Get("area-id") != want {
		t.Errorf("want update-project area-id=%s, got %s", want, u)
	}
}

func TestMoveRejections(t *testing.T) {
	srv, _ := newWriteServer(t)
	cases := []struct {
		name, to, wantMsg string
	}{
		{"inbox", "inbox", "Inbox"},
		{"when keyword", "someday", "schedule"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := move(t, srv, MoveInput{Target: "To-Do in Today", To: tc.to})
			if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
				t.Fatalf("want invalid_input, got %+v", res)
			}
			if !strings.Contains(res.Error.Message, tc.wantMsg) {
				t.Errorf("message %q should mention %q", res.Error.Message, tc.wantMsg)
			}
		})
	}

	t.Run("project into project", func(t *testing.T) {
		res := move(t, srv, MoveInput{Target: "Project in Area 1", To: "Project in Today"})
		if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
			t.Fatalf("want invalid_input, got %+v", res)
		}
		if !strings.Contains(res.Error.Message, "area") {
			t.Errorf("message %q should mention area", res.Error.Message)
		}
	})
}
