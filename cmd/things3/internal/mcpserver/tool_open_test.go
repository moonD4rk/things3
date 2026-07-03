package mcpserver

import (
	"context"
	"testing"

	"github.com/moond4rk/things3/thingstest"
)

func openTool(t *testing.T, srv *Server, in OpenInput) WriteResult {
	t.Helper()
	_, res, err := srv.handleOpen(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return res
}

func TestOpenView(t *testing.T) {
	srv, rec := newWriteServer(t)

	t.Run("today", func(t *testing.T) {
		res := openTool(t, srv, OpenInput{View: "today"})
		if !res.Success {
			t.Fatalf("open today failed: %+v", res.Error)
		}
		u := rec.last(t)
		if u.Path != "/show" || u.Query().Get("id") != "today" {
			t.Errorf("want show?id=today, got %s", u)
		}
	})

	t.Run("projects maps to a list id", func(t *testing.T) {
		res := openTool(t, srv, OpenInput{View: "projects"})
		if !res.Success {
			t.Fatalf("open projects failed: %+v", res.Error)
		}
		if rec.last(t).Query().Get("id") == "" {
			t.Errorf("open projects should target a list id")
		}
	})
}

func TestOpenTarget(t *testing.T) {
	srv, rec := newWriteServer(t)
	res := openTool(t, srv, OpenInput{Target: "To-Do in Today"})
	if !res.Success {
		t.Fatalf("open target failed: %+v", res.Error)
	}
	if got := rec.last(t).Query().Get("id"); got != thingstest.UUIDTodoInToday {
		t.Errorf("id = %q, want %q", got, thingstest.UUIDTodoInToday)
	}
	if res.Item == nil || res.Item.UUID != thingstest.UUIDTodoInToday {
		t.Errorf("open result should carry the resolved item, got %+v", res.Item)
	}
}

func TestOpenQuery(t *testing.T) {
	srv, rec := newWriteServer(t)
	openTool(t, srv, OpenInput{Query: "milk"})
	if got := rec.last(t).Query().Get("query"); got != "milk" {
		t.Errorf("query = %q, want milk", got)
	}
}

func TestOpenExactlyOne(t *testing.T) {
	srv, _ := newWriteServer(t)

	none := openTool(t, srv, OpenInput{})
	if none.Success || none.Error == nil || none.Error.Code != codeInvalidInput {
		t.Errorf("no argument should be invalid_input, got %+v", none)
	}

	two := openTool(t, srv, OpenInput{View: "today", Query: "x"})
	if two.Success || two.Error == nil || two.Error.Code != codeInvalidInput {
		t.Errorf("two arguments should be invalid_input, got %+v", two)
	}
}
