package cmd

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/moond4rk/things3/thingstest"
)

// Fixture rows that exercise the edges of target resolution.
const (
	uuidTrashedTodo      = "Tmt3omxtu61wAoTaSLweQy" // "Another Deleted Todo"
	uuidTemplate         = "N1PJHsbjct4mb1bhcs7aHa" // "Repeating To-Do", the series itself
	uuidTemplateInstance = "K9bx7h1xCJdevvyWardZDq" // "Repeating To-Do", today's copy
	uuidCompletedInToday = "56dtXSk3A373M6n4eqGyr3"
	uuidTodoInHeading    = "HbKGAeZKFDkWH5osSBNHvz"
)

func execFixture(t *testing.T, dbPath, query string, args ...any) {
	t.Helper()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("exec fixture: %v", err)
	}
}

// Every UUID a view prints must be a valid query, including the trashed items
// the trash view lists and the repeating series upcoming lists.
func TestPrintedUUIDsResolve(t *testing.T) {
	setupFixtureDB(t)

	for name, tc := range map[string]struct {
		view string
		uuid string
		want string
	}{
		"trashed todo from trash":     {view: "trash", uuid: uuidTrashedTodo, want: "Trashed:  yes"},
		"repeating series from views": {view: "upcoming", uuid: uuidTemplate, want: "the repeating series itself"},
	} {
		t.Run(name, func(t *testing.T) {
			short := tc.uuid[:8]
			listing, _, err := executeCommand(t, tc.view, "--all")
			if err != nil {
				t.Fatalf("%s: %v", tc.view, err)
			}
			if !strings.Contains(listing, short) {
				t.Fatalf("%s should print %s:\n%s", tc.view, short, listing)
			}

			detail, stderr, err := executeCommand(t, "show", short)
			if err != nil {
				t.Fatalf("show %s: %v (%s)", short, err, stderr)
			}
			if !strings.Contains(detail, tc.uuid) || !strings.Contains(detail, tc.want) {
				t.Errorf("show %s should render the item and its state %q:\n%s", short, tc.want, detail)
			}
		})
	}
}

// The repeating series carries a sentinel deadline that used to print as a date
// in 1953; it has no deadline of its own to show.
func TestTemplateSentinelDeadlineIsNotPrinted(t *testing.T) {
	setupFixtureDB(t)
	out, _, err := executeCommand(t, "upcoming", "--all")
	if err != nil {
		t.Fatalf("upcoming: %v", err)
	}
	if strings.Contains(out, "due:") {
		t.Errorf("the repeating series must not show a deadline:\n%s", out)
	}
}

func TestWritesRefuseTrashedItemsAndRepeatingSeries(t *testing.T) {
	setupFixtureDB(t)

	for name, tc := range map[string]struct {
		uuid string
		want string
	}{
		"trashed":          {uuid: uuidTrashedTodo, want: "in the Trash"},
		"repeating series": {uuid: uuidTemplate, want: "repeating series"},
	} {
		for _, args := range [][]string{
			{"done", tc.uuid},
			{"cancel", tc.uuid},
			{"schedule", tc.uuid, "today"},
			{"move", tc.uuid, "--to", "Area 1"},
			{"edit", tc.uuid, "--title", "x"},
		} {
			t.Run(name+" "+args[0], func(t *testing.T) {
				stdout, stderr, err := executeCommand(t, append(args, "--dry-run")...)
				assertExitCode(t, err, 1)
				if !strings.Contains(stderr, tc.want) {
					t.Errorf("stderr should explain why (%q): %s", tc.want, stderr)
				}
				if strings.Contains(stdout, "things://") {
					t.Errorf("no URL may be built for a read-only item, got %s", stdout)
				}
			})
		}
	}

	// Revealing is not a change, so open still works on both.
	for _, uuid := range []string{uuidTrashedTodo, uuidTemplate} {
		if u := dryRunURL(t, "open", uuid); u.Query().Get("id") != uuid {
			t.Errorf("open %s should reveal it, got %s", uuid, u)
		}
	}
}

// A title shared by one open item and its completed namesakes is what every
// repeating task looks like; naming it must reach the open one.
func TestWritePrefersTheSingleOpenItem(t *testing.T) {
	dbPath := setupFixtureDB(t)
	execFixture(t, dbPath, "UPDATE TMTask SET title = 'To-Do in Today' WHERE uuid = ?", uuidCompletedInToday)

	for _, args := range [][]string{
		{"done", "To-Do in Today"},
		{"schedule", "To-Do in Today", "tomorrow"},
		{"edit", "To-Do in Today", "--add-tags", "Office"},
	} {
		if got := dryRunURL(t, args...).Query().Get("id"); got != thingstest.UUIDTodoInToday {
			t.Errorf("%v targeted %s, want the open todo", args, got)
		}
	}

	// By title the series stays hidden behind today's copy.
	if got := dryRunURL(t, "done", "Repeating To-Do").Query().Get("id"); got != uuidTemplateInstance {
		t.Errorf("done by title targeted %s, want the live instance", got)
	}
}

func TestOpenSearch(t *testing.T) {
	setupFixtureDB(t)

	u := dryRunURL(t, "open", "--search", "buy milk")
	if u.Path != "/search" || u.Query().Get("query") != "buy milk" {
		t.Errorf("want things:///search?query=buy milk, got %s", u)
	}

	_, stderr, err := executeCommand(t, "open", "today", "--search", "milk", "--dry-run")
	assertExitCode(t, err, 1)
	if !strings.Contains(stderr, "--search") {
		t.Errorf("combining --search with a target should be rejected: %s", stderr)
	}
}

// A todo under a heading has no project column of its own; it must still be
// shown inside the heading's project, and inside that project's area.
func TestHeadingChildShowsItsProject(t *testing.T) {
	setupFixtureDB(t)

	detail, _, err := executeCommand(t, "show", uuidTodoInHeading)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	for _, want := range []string{"Project:  Project in Area 1", "Area:     Area 1", "Heading:  Heading"} {
		if !strings.Contains(detail, want) {
			t.Errorf("detail is missing %q:\n%s", want, detail)
		}
	}

	row, _, err := executeCommand(t, "deadlines", "--all")
	if err != nil {
		t.Fatalf("deadlines: %v", err)
	}
	if !strings.Contains(row, "@Project in Area 1 / Heading") {
		t.Errorf("the row should carry its project and heading:\n%s", row)
	}

	anytime, _, err := executeCommand(t, "anytime", "--all")
	if err != nil {
		t.Fatalf("anytime: %v", err)
	}
	_, group, found := strings.Cut(anytime, "Project in Area 1\n")
	group, _, _ = strings.Cut(group, "\n\n")
	if !found || !strings.Contains(group, uuidTodoInHeading[:8]) {
		t.Errorf("the heading child belongs in its project's group, got:\n%s", anytime)
	}
}
