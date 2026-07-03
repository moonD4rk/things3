package mcpserver

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/verify"
	"github.com/moond4rk/things3/thingstest"
)

// oneShotSleep makes verification check exactly once and then give up: a
// condition already true (the confirmed path) returns before any sleep, while an
// unverified condition resolves immediately instead of polling to the budget.
func oneShotSleep(context.Context, time.Duration) error { return context.DeadlineExceeded }

// recorder captures the URLs write tools would execute instead of running them,
// mirroring the CLI's --dry-run assertions without touching osascript.
type recorder struct {
	mu   sync.Mutex
	urls []string
}

func (r *recorder) exec(_ context.Context, b URLBuilder) error {
	uri, err := b.Build()
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.urls = append(r.urls, uri)
	r.mu.Unlock()
	return nil
}

func (r *recorder) last(t *testing.T) *url.URL {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.urls) == 0 {
		t.Fatalf("no URL was executed")
	}
	raw := r.urls[len(r.urls)-1]
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("executed URL is unparseable: %q (%v)", raw, err)
	}
	if u.Scheme != "things" {
		t.Fatalf("want a things:// URL, got %q", raw)
	}
	return u
}

// newWriteServer builds a server whose writes are recorded, not executed, and
// whose verification uses an instant clock.
func newWriteServer(t *testing.T) (*Server, *recorder) {
	t.Helper()
	rec := &recorder{}
	srv := newTestServer(t, Config{
		Execute: rec.exec,
		Verify:  verify.Options{Sleep: oneShotSleep},
	})
	return srv, rec
}

// fixtureProjectUUID looks up a project UUID by title from a fresh fixture copy.
func fixtureProjectUUID(t *testing.T, title string) string {
	t.Helper()
	p, err := testClient(t).Projects().WithTitle(title).Status().Any().First(context.Background())
	if err != nil {
		t.Fatalf("project %q: %v", title, err)
	}
	return p.UUID
}

// fixtureAreaUUID looks up an area UUID by title from a fresh fixture copy.
func fixtureAreaUUID(t *testing.T, title string) string {
	t.Helper()
	a, err := testClient(t).Areas().WithTitle(title).First(context.Background())
	if err != nil {
		t.Fatalf("area %q: %v", title, err)
	}
	return a.UUID
}

func TestRunWriteExecutionFailed(t *testing.T) {
	srv := newTestServer(t, Config{
		Execute: func(context.Context, URLBuilder) error { return errors.New("boom") },
	})
	res := srv.runWrite(context.Background(), srv.client.AddTodo().Title("x"),
		func(context.Context) WriteResult { return WriteResult{Success: true, Verified: true} })
	if res.Success {
		t.Errorf("execution failure should not be success")
	}
	if res.Error == nil || res.Error.Code != codeExecutionFailed {
		t.Errorf("want execution_failed, got %+v", res.Error)
	}
}

func TestTodoOutcome(t *testing.T) {
	todo := &things3.Todo{UUID: "u1", Title: "t", Status: things3.StatusCompleted}

	confirmed := todoOutcome(todo, verify.Confirmed, nil)
	if !confirmed.Success || !confirmed.Verified || confirmed.Item == nil || confirmed.Item.UUID != "u1" {
		t.Errorf("confirmed = %+v", confirmed)
	}

	unverified := todoOutcome(nil, verify.Unverified, nil)
	if !unverified.Success || unverified.Verified {
		t.Errorf("unverified should be success but not verified, got %+v", unverified)
	}

	ambiguous := todoOutcome(nil, verify.AmbiguousMatch, nil)
	if ambiguous.Verified || !strings.Contains(ambiguous.Message, "same-title") {
		t.Errorf("ambiguous message = %+v", ambiguous)
	}
}

func TestParseHelpers(t *testing.T) {
	if _, te := parseDate("2026-08-01"); te != nil {
		t.Errorf("valid date rejected: %v", te)
	}
	if _, te := parseDate("not-a-date"); te == nil || te.Code != codeInvalidInput {
		t.Errorf("bad date should be invalid_input, got %+v", te)
	}
	if hour, minute, te := parseReminder("09:30"); te != nil || hour != 9 || minute != 30 {
		t.Errorf("parseReminder(09:30) = %d:%d %v", hour, minute, te)
	}
	if _, _, te := parseReminder("25:00"); te == nil || te.Code != codeInvalidInput {
		t.Errorf("bad reminder should be invalid_input, got %+v", te)
	}
}

// count returns how many URLs the recorder captured, so a test can assert a
// rejected write executed nothing.
func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.urls)
}

// bumpEpoch is a far-future Unix timestamp, later than every fixture date, so
// writing it to a row's userModificationDate or creationDate confirms a
// modification or a create against any baseline.
const bumpEpoch = 4102444800 // 2100-01-01 UTC

const (
	// sqlBumpModified advances a row's modification timestamp past any baseline.
	sqlBumpModified = "UPDATE TMTask SET userModificationDate = ? WHERE uuid = ?"
	// sqlBumpCreated retitles a spare row and dates its creation to bumpEpoch, so a
	// create verification finds it by title as if Things had just made it.
	sqlBumpCreated = "UPDATE TMTask SET title = ?, creationDate = ? WHERE uuid = ?"
)

// confirmingServer builds a recording server over one fixture copy whose Execute,
// after recording the URL, runs bump to mutate that same copy -- simulating Things
// landing the write so the database verifier confirms it on the next poll.
func confirmingServer(t *testing.T, path string, bump func()) *Server {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(path))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	rec := &recorder{}
	srv, err := New(client, Config{
		Execute: func(ctx context.Context, b URLBuilder) error {
			if execErr := rec.exec(ctx, b); execErr != nil {
				return execErr
			}
			bump()
			return nil
		},
		Verify: verify.Options{Sleep: oneShotSleep},
	})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	return srv
}

// execFixture applies a write to the fixture copy at path over a short-lived
// connection, letting a test simulate Things mutating a row mid-write.
func execFixture(t *testing.T, path, query string, args ...any) {
	t.Helper()
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("exec fixture: %v", err)
	}
}

// assertConfirmed asserts a write landed as a verified success carrying wantUUID,
// exercising the resolve/baseline -> verify -> outcome handler wiring end to end.
func assertConfirmed(t *testing.T, res WriteResult, wantUUID string) {
	t.Helper()
	if !res.Success || !res.Verified {
		t.Fatalf("want verified success, got %+v (error %+v)", res, res.Error)
	}
	if res.Item == nil || res.Item.UUID != wantUUID {
		t.Fatalf("want confirmed item %s, got %+v", wantUUID, res.Item)
	}
}

func TestScheduleConfirmed(t *testing.T) {
	path := thingstest.DatabasePath(t)
	srv := confirmingServer(t, path, func() {
		execFixture(t, path, sqlBumpModified, bumpEpoch, thingstest.UUIDTodoInToday)
	})
	res := schedule(t, srv, ScheduleInput{Target: thingstest.UUIDTodoInToday, When: "tomorrow"})
	assertConfirmed(t, res, thingstest.UUIDTodoInToday)
}

func TestMoveConfirmed(t *testing.T) {
	path := thingstest.DatabasePath(t)
	srv := confirmingServer(t, path, func() {
		execFixture(t, path, sqlBumpModified, bumpEpoch, thingstest.UUIDTodoInToday)
	})
	res := move(t, srv, MoveInput{Target: thingstest.UUIDTodoInToday, To: thingstest.UUIDProject})
	assertConfirmed(t, res, thingstest.UUIDTodoInToday)
}

func TestEditConfirmed(t *testing.T) {
	path := thingstest.DatabasePath(t)
	srv := confirmingServer(t, path, func() {
		execFixture(t, path, sqlBumpModified, bumpEpoch, thingstest.UUIDTodoInToday)
	})
	res := edit(t, srv, EditInput{Target: thingstest.UUIDTodoInToday, Title: "Renamed"})
	assertConfirmed(t, res, thingstest.UUIDTodoInToday)
}

func TestAddTodoConfirmed(t *testing.T) {
	path := thingstest.DatabasePath(t)
	const title = "ZZ-Confirmed-Todo"
	// UUIDTodoChecklist is a non-trashed, non-template todo; retitling it and dating
	// its creation to bumpEpoch makes AddedTodo recover it as the freshly created one.
	srv := confirmingServer(t, path, func() {
		execFixture(t, path, sqlBumpCreated, title, bumpEpoch, thingstest.UUIDTodoChecklist)
	})
	res := addTodo(t, srv, AddTodoInput{Title: title})
	assertConfirmed(t, res, thingstest.UUIDTodoChecklist)
}

func TestAddProjectConfirmed(t *testing.T) {
	path := thingstest.DatabasePath(t)
	const title = "ZZ-Confirmed-Project"
	srv := confirmingServer(t, path, func() {
		execFixture(t, path, sqlBumpCreated, title, bumpEpoch, thingstest.UUIDSomedayProject)
	})
	res := addProject(t, srv, AddProjectInput{Title: title})
	assertConfirmed(t, res, thingstest.UUIDSomedayProject)
}

// TestWriteExecErrorClassification proves a failed write is classified by error
// identity: URL-scheme input-validation sentinels are invalid_input, while a missing
// auth token or any other failure is execution_failed. This guards against
// reclassifying a token/IO failure - raised inside URL building for update tools -
// as bad input.
func TestWriteExecErrorClassification(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"tag comma", things3.ErrTagContainsComma, codeInvalidInput},
		{"reminder needs date", things3.ErrReminderNeedsDate, codeInvalidInput},
		{"title too long", things3.ErrTitleTooLong, codeInvalidInput},
		{"checklist newline", things3.ErrChecklistItemContainsNewline, codeInvalidInput},
		{"missing auth token", things3.ErrAuthTokenNotFound, codeExecutionFailed},
		{"osascript failure", errors.New("osascript exited 1"), codeExecutionFailed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := writeExecError(tc.err); got.Code != tc.want {
				t.Errorf("writeExecError(%v).Code = %q, want %q", tc.err, got.Code, tc.want)
			}
		})
	}
}

// TestUpdateMissingTokenIsExecutionFailed proves that when the auth token cannot be
// read - resolved inside URL building for an update tool - the write reports
// execution_failed with the diagnostic hint, not invalid_input. The fixture's token
// is cleared to reproduce the first-run "URL scheme not authorized" condition.
func TestUpdateMissingTokenIsExecutionFailed(t *testing.T) {
	path := thingstest.DatabasePath(t)
	execFixture(t, path, "UPDATE TMSettings SET uriSchemeAuthenticationToken = NULL")
	client, err := things3.NewClient(things3.WithDatabasePath(path))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	rec := &recorder{}
	srv, err := New(client, Config{Execute: rec.exec, Verify: verify.Options{Sleep: oneShotSleep}})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	res := schedule(t, srv, ScheduleInput{Target: thingstest.UUIDTodoInToday, When: "today"})
	if res.Success || res.Error == nil || res.Error.Code != codeExecutionFailed {
		t.Fatalf("a missing auth token on an update must be execution_failed, got %+v", res.Error)
	}
	if res.Error.Hint == "" {
		t.Errorf("execution_failed should carry the diagnostic hint")
	}
}
