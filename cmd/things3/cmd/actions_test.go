package cmd

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

func fixtureProjectUUID(t *testing.T, dbPath, title string) string {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(dbPath))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer func() { _ = client.Close() }()
	p, err := client.Projects().WithTitle(title).Status().Any().First(context.Background())
	if err != nil {
		t.Fatalf("project %q: %v", title, err)
	}
	return p.UUID
}

func fixtureAreaUUID(t *testing.T, dbPath, title string) string {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(dbPath))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer func() { _ = client.Close() }()
	a, err := client.Areas().WithTitle(title).First(context.Background())
	if err != nil {
		t.Fatalf("area %q: %v", title, err)
	}
	return a.UUID
}

// dryRunURL runs a write command with --dry-run and parses the emitted URL.
func dryRunURL(t *testing.T, args ...string) *url.URL {
	t.Helper()
	stdout, stderr, err := executeCommand(t, append(args, "--dry-run")...)
	if err != nil {
		t.Fatalf("args %v: unexpected error %v (stderr %s)", args, err, stderr)
	}
	u, perr := url.Parse(strings.TrimSpace(stdout))
	if perr != nil {
		t.Fatalf("dry-run output is not a URL: %q (%v)", stdout, perr)
	}
	if u.Scheme != "things" {
		t.Fatalf("want a things:// URL, got %q", stdout)
	}
	return u
}

func TestAddDryRun(t *testing.T) {
	setupFixtureDB(t)

	t.Run("title when notes tags", func(t *testing.T) {
		u := dryRunURL(t, "add", "Buy milk", "--when", "evening", "--notes", "at the store", "--tags", "home,errand")
		if u.Path != "/add" {
			t.Fatalf("want /add, got %q", u.Path)
		}
		q := u.Query()
		if q.Get("title") != "Buy milk" {
			t.Errorf("title = %q", q.Get("title"))
		}
		if q.Get("when") != "evening" {
			t.Errorf("when = %q", q.Get("when"))
		}
		if q.Get("notes") != "at the store" {
			t.Errorf("notes = %q", q.Get("notes"))
		}
		if q.Get("tags") != "home,errand" {
			t.Errorf("tags = %q", q.Get("tags"))
		}
	})

	t.Run("checklist and deadline", func(t *testing.T) {
		u := dryRunURL(t, "add", "Plan", "--checklist", "one", "--checklist", "two", "--deadline", "2026-08-01")
		q := u.Query()
		if !strings.Contains(q.Get("checklist-items"), "one") || !strings.Contains(q.Get("checklist-items"), "two") {
			t.Errorf("checklist-items = %q", q.Get("checklist-items"))
		}
		if q.Get("deadline") != "2026-08-01" {
			t.Errorf("deadline = %q", q.Get("deadline"))
		}
	})

	t.Run("reminder time in when", func(t *testing.T) {
		u := dryRunURL(t, "add", "Wake", "--reminder", "09:30")
		if !strings.Contains(u.Query().Get("when"), "09:30") {
			t.Errorf("reminder should appear in when, got %q", u.Query().Get("when"))
		}
	})

	t.Run("invalid reminder exits 1", func(t *testing.T) {
		_, _, err := executeCommand(t, "add", "Wake", "--reminder", "25:00", "--dry-run")
		assertExitCode(t, err, 1)
	})
}

func TestAddDestinationFlags(t *testing.T) {
	setupFixtureDB(t)

	t.Run("project resolves to full uuid", func(t *testing.T) {
		u := dryRunURL(t, "add", "Email Bob", "--project", "Project in Area 1")
		if got := u.Query().Get("list-id"); got != thingstest.UUIDProject {
			t.Errorf("list-id = %q, want %q", got, thingstest.UUIDProject)
		}
	})

	t.Run("ambiguous project exits 2", func(t *testing.T) {
		stdout, stderr, err := executeCommand(t, "add", "X", "--project", "Project in", "--dry-run")
		assertExitCode(t, err, 2)
		if strings.TrimSpace(stdout) != "" {
			t.Errorf("ambiguous add must not print a URL, got %q", stdout)
		}
		if !strings.Contains(stderr, "Error:") || !strings.Contains(stderr, "project") {
			t.Errorf("ambiguity should list project candidates:\n%s", stderr)
		}
	})

	t.Run("project and area are mutually exclusive", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "add", "X", "--project", "Work", "--area", "Home", "--dry-run")
		assertExitCode(t, err, 1)
		if !strings.Contains(stderr, flagProject) || !strings.Contains(stderr, flagArea) {
			t.Errorf("exclusion error should name both flags, got %q", stderr)
		}
	})
}

func TestWriteAcceptsIgnoredListFlags(t *testing.T) {
	setupFixtureDB(t)
	// List flags are global for a uniform surface; a write command accepts and
	// ignores them, still emitting its URL.
	u := dryRunURL(t, "add", "Buy milk", "--page", "3")
	if u.Path != "/add" {
		t.Fatalf("want /add, got %q", u.Path)
	}
	if got := u.Query().Get("title"); got != "Buy milk" {
		t.Errorf("title = %q, want %q", got, "Buy milk")
	}
}

func TestDoneCancelDryRun(t *testing.T) {
	setupFixtureDB(t)

	u := dryRunURL(t, "done", "To-Do in Today")
	q := u.Query()
	if u.Path != "/update" {
		t.Fatalf("want /update, got %q", u.Path)
	}
	if q.Get("id") != thingstest.UUIDTodoInToday {
		t.Errorf("id = %q, want %q", q.Get("id"), thingstest.UUIDTodoInToday)
	}
	if q.Get("completed") != "true" {
		t.Errorf("completed = %q", q.Get("completed"))
	}
	if q.Get("auth-token") != thingstest.AuthToken {
		t.Errorf("auth-token = %q, want %q", q.Get("auth-token"), thingstest.AuthToken)
	}

	c := dryRunURL(t, "cancel", "To-Do in Today")
	if c.Query().Get("canceled") != "true" {
		t.Errorf("canceled = %q", c.Query().Get("canceled"))
	}
}

func TestScheduleDryRun(t *testing.T) {
	setupFixtureDB(t)

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
			u := dryRunURL(t, "schedule", "To-Do in Today", tc.when)
			if got := u.Query().Get("when"); got != tc.want {
				t.Errorf("when = %q, want %q", got, tc.want)
			}
		})
	}

	t.Run("bad when exits 1", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "schedule", "To-Do in Today", "tdoay", "--dry-run")
		assertExitCode(t, err, 1)
		if !strings.Contains(stderr, "today") {
			t.Errorf("bad when should list allowed values, got %q", stderr)
		}
	})
}

func TestMoveDryRun(t *testing.T) {
	dbPath := setupFixtureDB(t)

	t.Run("to project", func(t *testing.T) {
		want := fixtureProjectUUID(t, dbPath, "Project in Today")
		u := dryRunURL(t, "move", "To-Do in Today", "--to", "Project in Today")
		if got := u.Query().Get("list-id"); got != want {
			t.Errorf("list-id = %q, want %q", got, want)
		}
	})

	t.Run("to inbox rejected", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "move", "To-Do in Today", "--to", "inbox")
		assertExitCode(t, err, 1)
		if !strings.Contains(stderr, "Inbox") {
			t.Errorf("expected Inbox message, got %q", stderr)
		}
	})

	t.Run("when keyword rejected", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "move", "To-Do in Today", "--to", "someday")
		assertExitCode(t, err, 1)
		if !strings.Contains(stderr, "schedule") {
			t.Errorf("expected pointer to schedule, got %q", stderr)
		}
	})

	t.Run("project into project rejected", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "move", "Project in Area 1", "--to", "Project in Today")
		assertExitCode(t, err, 1)
		if !strings.Contains(stderr, "area") {
			t.Errorf("expected a project-to-area message, got %q", stderr)
		}
	})
}

func TestEditDryRun(t *testing.T) {
	setupFixtureDB(t)

	cases := []struct{ name, flag, val, param, want string }{
		{"title", "--title", "New title", "title", "New title"},
		{"append-notes", "--append-notes", "more", "append-notes", "more"},
		{"deadline", "--deadline", "2026-08-01", "deadline", "2026-08-01"},
		{"add-tags", "--add-tags", "urgent", "add-tags", "urgent"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := dryRunURL(t, "edit", "To-Do in Today", tc.flag, tc.val)
			if got := u.Query().Get(tc.param); got != tc.want {
				t.Errorf("%s = %q, want %q", tc.param, got, tc.want)
			}
		})
	}

	t.Run("clear-deadline sets empty deadline", func(t *testing.T) {
		u := dryRunURL(t, "edit", "To-Do in Today", "--clear-deadline")
		if _, ok := u.Query()["deadline"]; !ok {
			t.Errorf("clear-deadline should send an empty deadline param, got %s", u)
		}
	})

	t.Run("no flags exits 1", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "edit", "To-Do in Today")
		assertExitCode(t, err, 1)
		if !strings.Contains(stderr, "nothing to edit") {
			t.Errorf("expected 'nothing to edit', got %q", stderr)
		}
	})
}

func TestWriteTargetAmbiguity(t *testing.T) {
	setupFixtureDB(t)

	// "Project in" matches several projects; a write target must be unambiguous.
	stdout, stderr, err := executeCommand(t, "done", "Project in")
	assertExitCode(t, err, 2)
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("ambiguous write must not print to stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "Error:") || !strings.Contains(stderr, "project") {
		t.Errorf("ambiguity should print candidate rows:\n%s", stderr)
	}
	if !strings.Contains(stderr, "hint: rerun with a UUID") {
		t.Errorf("ambiguity should teach the UUID-query hint:\n%s", stderr)
	}

	_, stderrJSON, err := executeCommand(t, "done", "Project in", "--json")
	assertExitCode(t, err, 2)
	if !strings.Contains(stderrJSON, "error") || !strings.Contains(stderrJSON, "candidates") {
		t.Errorf("json ambiguity should carry error+candidates:\n%s", stderrJSON)
	}
}

func TestOpenDryRun(t *testing.T) {
	setupFixtureDB(t)

	t.Run("view name", func(t *testing.T) {
		u := dryRunURL(t, "open", "today")
		if u.Path != "/show" || u.Query().Get("id") != "today" {
			t.Errorf("want show?id=today, got %s", u)
		}
	})

	t.Run("query resolves to uuid", func(t *testing.T) {
		u := dryRunURL(t, "open", "To-Do in Today")
		if u.Query().Get("id") != thingstest.UUIDTodoInToday {
			t.Errorf("id = %q, want %q", u.Query().Get("id"), thingstest.UUIDTodoInToday)
		}
	})

	t.Run("no args defaults to today", func(t *testing.T) {
		u := dryRunURL(t, "open")
		if u.Query().Get("id") != "today" {
			t.Errorf("no-arg open should target today, got %s", u)
		}
	})
}

func TestWriteWriteResult(t *testing.T) {
	todo := &things3.Todo{UUID: "abcd1234efgh5678", Title: "Buy milk", Status: things3.StatusCompleted}

	cases := []struct {
		name         string
		result       writeResult
		format       outputFormat
		wantContains string
	}{
		{"dry-run text is url only", writeResult{Action: "add", DryRun: true, URL: "things:///add?title=x"}, formatText, "things:///add?title=x"},
		{"confirmed text", writeResult{Action: "done", Verified: true, Todo: todo}, formatText, "done: "},
		{"unverified text", writeResult{Action: "add", Verified: false, Message: "not yet"}, formatText, "sent to Things (not yet confirmed)"},
		{"open text uses opened verb", writeResult{Action: "open", Verified: true, Message: "today"}, formatText, "opened: today"},
		{"confirmed json", writeResult{Action: "done", Verified: true, Todo: todo}, formatJSON, `"verified": true`},
		{"dry-run json", writeResult{Action: "add", DryRun: true, URL: "u"}, formatJSON, `"dry_run": true`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			result := tc.result
			if err := writeWriteResult(&buf, &result, tc.format); err != nil {
				t.Fatalf("writeWriteResult: %v", err)
			}
			if !strings.Contains(buf.String(), tc.wantContains) {
				t.Errorf("output %q does not contain %q", buf.String(), tc.wantContains)
			}
		})
	}

	var yb bytes.Buffer
	if err := writeWriteResult(&yb, &writeResult{Action: "done", Verified: true, Todo: todo}, formatYAML); err != nil {
		t.Fatalf("yaml: %v", err)
	}
	if !strings.Contains(yb.String(), "verified") {
		t.Errorf("yaml result missing verified key:\n%s", yb.String())
	}
}

func TestAddProjectDryRun(t *testing.T) {
	setupFixtureDB(t)
	u := dryRunURL(t, "add", "project", "Website", "--area", "Area 1",
		"--todos", "Flights", "--todos", "Hotel", "--deadline", "2026-08-01")
	if u.Path != "/add-project" {
		t.Fatalf("want /add-project, got %q", u.Path)
	}
	q := u.Query()
	if q.Get("title") != "Website" {
		t.Errorf("title = %q", q.Get("title"))
	}
	if q.Get("area-id") == "" {
		t.Errorf("expected area-id for --area, got %s", u)
	}
	if q.Get("deadline") != "2026-08-01" {
		t.Errorf("deadline = %q", q.Get("deadline"))
	}
	if !strings.Contains(q.Get("to-dos"), "Flights") || !strings.Contains(q.Get("to-dos"), "Hotel") {
		t.Errorf("to-dos = %q", q.Get("to-dos"))
	}

	_, _, err := executeCommand(t, "add", "project", "Website", "--deadline", "not-a-date", "--dry-run")
	assertExitCode(t, err, 1)
}

func TestAddHeadingDryRun(t *testing.T) {
	setupFixtureDB(t)
	u := dryRunURL(t, "add", "Sub-task", "--project", "Project in Area 1", "--heading", "Heading")
	q := u.Query()
	if q.Get("list-id") != thingstest.UUIDProject {
		t.Errorf("list-id = %q, want %q", q.Get("list-id"), thingstest.UUIDProject)
	}
	if q.Get("heading-id") == "" {
		t.Errorf("expected heading-id, got %s", u)
	}

	_, stderr, err := executeCommand(t, "add", "X", "--heading", "Heading", "--dry-run")
	assertExitCode(t, err, 1)
	if !strings.Contains(stderr, "requires --project") {
		t.Errorf("want '--heading requires --project', got %q", stderr)
	}
}

func TestMoveDestinations(t *testing.T) {
	dbPath := setupFixtureDB(t)
	// "Area 2" is not a substring of any project title, so it exercises the
	// todo->area fallback and the project->area path cleanly.
	area2 := fixtureAreaUUID(t, dbPath, "Area 2")

	t.Run("todo to area", func(t *testing.T) {
		u := dryRunURL(t, "move", "To-Do in Today", "--to", "Area 2")
		if got := u.Query().Get("list-id"); got != area2 {
			t.Errorf("list-id = %q, want area %q", got, area2)
		}
	})

	t.Run("project to area", func(t *testing.T) {
		u := dryRunURL(t, "move", "Project in Area 1", "--to", "Area 2")
		if u.Path != "/update-project" || u.Query().Get("area-id") != area2 {
			t.Errorf("want update-project area-id=%s, got %s", area2, u)
		}
	})

	t.Run("to is required", func(t *testing.T) {
		_, _, err := executeCommand(t, "move", "To-Do in Today")
		if err == nil {
			t.Error("move without --to should error")
		}
	})
}

func TestDoneProjectTarget(t *testing.T) {
	dbPath := setupFixtureDB(t)
	want := fixtureProjectUUID(t, dbPath, "Project in Today")
	u := dryRunURL(t, "done", "Project in Today")
	if u.Path != "/update-project" {
		t.Fatalf("want /update-project, got %q", u.Path)
	}
	if got := u.Query().Get("id"); got != want {
		t.Errorf("id = %q, want %q", got, want)
	}
	if u.Query().Get("completed") != "true" {
		t.Errorf("completed = %q", u.Query().Get("completed"))
	}
}
