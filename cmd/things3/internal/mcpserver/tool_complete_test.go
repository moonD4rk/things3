package mcpserver

import (
	"context"
	"testing"

	"github.com/moond4rk/things3/thingstest"
)

func complete(t *testing.T, srv *Server, in CompleteInput) WriteResult {
	t.Helper()
	_, res, err := srv.handleComplete(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	return res
}

func TestCompleteURL(t *testing.T) {
	srv, rec := newWriteServer(t)

	t.Run("completed", func(t *testing.T) {
		complete(t, srv, CompleteInput{Target: "To-Do in Today", Status: "completed"})
		u := rec.last(t)
		if u.Path != "/update" {
			t.Fatalf("path = %q, want /update", u.Path)
		}
		q := u.Query()
		if q.Get("id") != thingstest.UUIDTodoInToday {
			t.Errorf("id = %q, want %q", q.Get("id"), thingstest.UUIDTodoInToday)
		}
		if q.Get("completed") != "true" {
			t.Errorf("completed = %q", q.Get("completed"))
		}
		if q.Get("auth-token") != thingstest.AuthToken {
			t.Errorf("auth-token = %q, want %q", q.Get("auth-token"), thingstest.AuthToken)
		}
	})

	t.Run("canceled", func(t *testing.T) {
		complete(t, srv, CompleteInput{Target: "To-Do in Today", Status: "canceled"})
		if got := rec.last(t).Query().Get("canceled"); got != "true" {
			t.Errorf("canceled = %q", got)
		}
	})

	t.Run("reopen clears completed", func(t *testing.T) {
		complete(t, srv, CompleteInput{Target: "To-Do in Today", Status: "incomplete"})
		if got := rec.last(t).Query().Get("completed"); got != "false" {
			t.Errorf("reopen completed = %q, want false", got)
		}
	})
}

func TestCompleteConfirmedReopen(t *testing.T) {
	srv, _ := newWriteServer(t)
	// The fixture todo is already incomplete, so reopening it verifies at once.
	res := complete(t, srv, CompleteInput{Target: thingstest.UUIDTodoInToday, Status: "incomplete"})
	if !res.Success || !res.Verified {
		t.Fatalf("reopen of an incomplete todo should confirm: %+v", res)
	}
	if res.Item == nil || res.Item.UUID != thingstest.UUIDTodoInToday || res.Item.Status != statusIncomplete {
		t.Errorf("confirmed item = %+v", res.Item)
	}
}

func TestCompleteProjectTarget(t *testing.T) {
	srv, rec := newWriteServer(t)
	want := fixtureProjectUUID(t, "Project in Today")
	complete(t, srv, CompleteInput{Target: "Project in Today", Status: "completed"})
	u := rec.last(t)
	if u.Path != "/update-project" {
		t.Fatalf("path = %q, want /update-project", u.Path)
	}
	if got := u.Query().Get("id"); got != want {
		t.Errorf("id = %q, want %q", got, want)
	}
	if u.Query().Get("completed") != "true" {
		t.Errorf("completed = %q", u.Query().Get("completed"))
	}
}

func TestCompleteErrors(t *testing.T) {
	srv, _ := newWriteServer(t)

	nf := complete(t, srv, CompleteInput{Target: "zzznope", Status: "completed"})
	if nf.Success || nf.Error == nil || nf.Error.Code != codeNotFound {
		t.Errorf("want not_found, got %+v", nf)
	}

	amb := complete(t, srv, CompleteInput{Target: "Project in", Status: "completed"})
	if amb.Success || amb.Error == nil || amb.Error.Code != codeAmbiguous {
		t.Fatalf("want ambiguous, got %+v", amb)
	}
	if len(amb.Error.Candidates) < 2 {
		t.Errorf("ambiguous should carry candidates, got %+v", amb.Error.Candidates)
	}
}
