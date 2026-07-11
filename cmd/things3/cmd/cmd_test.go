package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

// setupFixtureDB points the CLI at a writable copy of the shared fixture and
// returns its path.
func setupFixtureDB(t *testing.T) string {
	t.Helper()
	path := thingstest.DatabasePath(t)
	t.Setenv("THINGSDB", path)
	return path
}

// executeCommand runs the root command like main does: it executes, and on
// error renders it (so stderr reflects real user-facing output).
func executeCommand(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	root := NewRootCmd()
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs(args)
	err = root.Execute()
	if err != nil {
		RenderError(root, err)
	}
	return outBuf.String(), errBuf.String(), err
}

func assertExitCode(t *testing.T, err error, want int) {
	t.Helper()
	if got := ExitCode(err); got != want {
		t.Errorf("exit code = %d, want %d (err: %v)", got, want, err)
	}
}

// listEnvelope mirrors the json/yaml shape a list command emits: an items array
// plus the total/page/pages pagination metadata.
type listEnvelope struct {
	Items []map[string]any `json:"items" yaml:"items"`
	Total int              `json:"total" yaml:"total"`
	Page  int              `json:"page" yaml:"page"`
	Pages int              `json:"pages" yaml:"pages"`
}

// decodeList unmarshals a list-command json envelope.
func decodeList(t *testing.T, s string) listEnvelope {
	t.Helper()
	var env listEnvelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &env); err != nil {
		t.Fatalf("output is not a list envelope: %v\n%s", err, s)
	}
	return env
}

// decodeItems unmarshals only the envelope's items array into dst.
func decodeItems(t *testing.T, s string, dst any) {
	t.Helper()
	var env struct {
		Items json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &env); err != nil {
		t.Fatalf("output is not a list envelope: %v\n%s", err, s)
	}
	if err := json.Unmarshal(env.Items, dst); err != nil {
		t.Fatalf("decode items: %v\n%s", err, env.Items)
	}
}

// jsonTotal returns the envelope's reported total, the full count before paging.
func jsonTotal(t *testing.T, s string) int {
	t.Helper()
	return decodeList(t, s).Total
}

// runJSON executes a command expected to succeed and returns its stdout.
func runJSON(t *testing.T, args ...string) string {
	t.Helper()
	out, stderr, err := executeCommand(t, args...)
	if err != nil {
		t.Fatalf("%v: %v (stderr %s)", args, err, stderr)
	}
	return out
}

// hasHeaderLine reports whether text contains a group/section header - a
// non-empty line that is neither a todo row nor the flat column header.
func hasHeaderLine(text string) bool {
	for line := range strings.SplitSeq(strings.TrimSpace(text), "\n") {
		if line == "" || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "STATUS") {
			continue
		}
		return true
	}
	return false
}

func fetchTitle(t *testing.T, dbPath, uuid string) string {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(dbPath))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer func() { _ = client.Close() }()
	todo, err := client.Todos().WithUUID(uuid).Status().Any().First(context.Background())
	if err != nil {
		t.Fatalf("fetch title: %v", err)
	}
	return todo.Title
}

func injectEvening(t *testing.T, dbPath, uuid string) {
	t.Helper()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = db.Close() }()
	res, err := db.ExecContext(context.Background(), "UPDATE TMTask SET startBucket = 1 WHERE uuid = ?", uuid)
	if err != nil {
		t.Fatalf("inject evening: %v", err)
	}
	if n, _ := res.RowsAffected(); n != 1 {
		t.Fatalf("expected to update 1 row, updated %d", n)
	}
}

func TestViewRowCounts(t *testing.T) {
	tests := []struct {
		view string
		want int
	}{
		{"inbox", thingstest.Inbox},
		{"deadlines", thingstest.Deadlines},
	}
	for _, tt := range tests {
		t.Run(tt.view, func(t *testing.T) {
			setupFixtureDB(t)
			stdout, stderr, err := executeCommand(t, tt.view, "--json")
			if err != nil {
				t.Fatalf("%s: %v (stderr %s)", tt.view, err, stderr)
			}
			if got := jsonTotal(t, stdout); got != tt.want {
				t.Errorf("%s row count = %d, want %d", tt.view, got, tt.want)
			}
		})
	}
}

func TestTodayContainsInToday(t *testing.T) {
	setupFixtureDB(t)
	stdout, stderr, err := executeCommand(t, "today", "--json")
	if err != nil {
		t.Fatalf("today: %v (stderr %s)", err, stderr)
	}
	if !strings.Contains(stdout, thingstest.UUIDTodoInToday) {
		t.Errorf("today should contain %s\n%s", thingstest.UUIDTodoInToday, stdout)
	}
}

func TestTodayEveningSection(t *testing.T) {
	setupFixtureDB(t)
	plain, _, err := executeCommand(t, "today")
	if err != nil {
		t.Fatalf("today: %v", err)
	}
	if strings.Contains(plain, "This Evening") {
		t.Errorf("today without evening todos must not show a This Evening section:\n%s", plain)
	}

	path := thingstest.DatabasePath(t)
	injectEvening(t, path, thingstest.UUIDTodoInToday)
	t.Setenv("THINGSDB", path)
	sectioned, _, err := executeCommand(t, "today")
	if err != nil {
		t.Fatalf("today (evening): %v", err)
	}
	todayIdx := strings.Index(sectioned, "Today")
	eveningIdx := strings.Index(sectioned, "This Evening")
	if todayIdx < 0 || eveningIdx < 0 {
		t.Fatalf("expected Today and This Evening sections:\n%s", sectioned)
	}
	if todayIdx > eveningIdx {
		t.Errorf("Today section must precede This Evening:\n%s", sectioned)
	}
	if evening := sectioned[eveningIdx:]; !strings.Contains(evening, thingstest.UUIDTodoInToday[:8]) {
		t.Errorf("injected todo %s should appear under This Evening:\n%s", thingstest.UUIDTodoInToday[:8], sectioned)
	}
}

func TestUpcomingGrouped(t *testing.T) {
	setupFixtureDB(t)
	text, _, err := executeCommand(t, "upcoming")
	if err != nil {
		t.Fatalf("upcoming: %v", err)
	}
	if strings.TrimSpace(text) != "" && !hasHeaderLine(text) {
		t.Errorf("upcoming text should have date group headers:\n%s", text)
	}
	j, _, err := executeCommand(t, "upcoming", "--json")
	if err != nil {
		t.Fatalf("upcoming json: %v", err)
	}
	decodeList(t, j)
}

func TestAnytimeGrouped(t *testing.T) {
	setupFixtureDB(t)
	text, _, err := executeCommand(t, "anytime")
	if err != nil {
		t.Fatalf("anytime: %v", err)
	}
	if strings.TrimSpace(text) != "" && !hasHeaderLine(text) {
		t.Errorf("anytime text should have container headers:\n%s", text)
	}
	j, _, err := executeCommand(t, "anytime", "--json")
	if err != nil {
		t.Fatalf("anytime json: %v", err)
	}
	decodeList(t, j)
}

func TestTrashMixed(t *testing.T) {
	dbPath := setupFixtureDB(t)
	ctx := context.Background()
	client, err := things3.NewClient(things3.WithDatabasePath(dbPath))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	trashedTodos, err := client.Todos().Trashed(true).Status().Any().All(ctx)
	if err != nil {
		t.Fatalf("trashed todos: %v", err)
	}
	trashedProjects, err := client.Projects().Trashed(true).Status().Any().All(ctx)
	if err != nil {
		t.Fatalf("trashed projects: %v", err)
	}
	want := len(trashedTodos) + len(trashedProjects)
	_ = client.Close()

	text, _, err := executeCommand(t, "trash")
	if err != nil {
		t.Fatalf("trash: %v", err)
	}
	if !strings.Contains(text, "TYPE") {
		t.Errorf("trash text should have a TYPE column:\n%s", text)
	}
	j, _, err := executeCommand(t, "trash", "--json")
	if err != nil {
		t.Fatalf("trash json: %v", err)
	}
	if got := jsonTotal(t, j); got != want {
		t.Errorf("trash count = %d, want %d", got, want)
	}
}

type logbookRow struct {
	UUID        string     `json:"uuid"`
	CompletedAt *time.Time `json:"completed_at"`
	CanceledAt  *time.Time `json:"canceled_at"`
}

func (r *logbookRow) stop() *time.Time {
	if r.CompletedAt != nil {
		return r.CompletedAt
	}
	return r.CanceledAt
}

func TestLogbookOrdering(t *testing.T) {
	setupFixtureDB(t)
	out, stderr, err := executeCommand(t, "logbook", "--days", "0", "--json")
	if err != nil {
		t.Fatalf("logbook: %v (%s)", err, stderr)
	}
	var rows []logbookRow
	decodeItems(t, out, &rows)
	if len(rows) < 2 {
		t.Fatalf("want >= 2 logbook rows, got %d", len(rows))
	}
	for i := 1; i < len(rows); i++ {
		prev, curr := rows[i-1].stop(), rows[i].stop()
		if prev != nil && curr != nil && curr.After(*prev) {
			t.Errorf("logbook not sorted desc: row %d is newer than row %d", i, i-1)
		}
	}

	limited, _, err := executeCommand(t, "logbook", "--days", "0", "-n", "2", "--json")
	if err != nil {
		t.Fatalf("logbook -n 2: %v", err)
	}
	var limitedRows []logbookRow
	decodeItems(t, limited, &limitedRows)
	if len(limitedRows) != 2 {
		t.Fatalf("want 2 rows with -n 2, got %d", len(limitedRows))
	}
	for i := range limitedRows {
		if limitedRows[i].UUID != rows[i].UUID {
			t.Errorf("-n 2 row %d = %s, want most-recent %s", i, limitedRows[i].UUID, rows[i].UUID)
		}
	}
}

func TestShowRoundTrips(t *testing.T) {
	dbPath := setupFixtureDB(t)
	full := thingstest.UUIDTodoInToday

	for _, q := range []string{full, full[:8], fetchTitle(t, dbPath, full)} {
		out, stderr, err := executeCommand(t, "show", q)
		if err != nil {
			t.Fatalf("show %q: %v (%s)", q, err, stderr)
		}
		if !strings.Contains(out, "Title:") {
			t.Errorf("show %q should render a detail view:\n%s", q, out)
		}
	}

	// substring matching several -> mixed list, exit 0
	out, _, err := executeCommand(t, "show", "in Today")
	assertExitCode(t, err, 0)
	if !strings.Contains(out, "TYPE") {
		t.Errorf("multi-match show should render a mixed list:\n%s", out)
	}

	// gibberish -> exit 1, Error: on stderr
	_, stderr, err := executeCommand(t, "show", "zzznope")
	assertExitCode(t, err, 1)
	if !strings.Contains(stderr, "Error:") || !strings.Contains(stderr, "zzznope") {
		t.Errorf("not-found should print Error with the query, stderr: %s", stderr)
	}

	// gibberish --json -> {"error":...} JSON on stderr
	_, stderr, err = executeCommand(t, "show", "zzznope", "--json")
	assertExitCode(t, err, 1)
	if !strings.Contains(stderr, `"error"`) {
		t.Errorf("not-found under --json should print a JSON error, stderr: %s", stderr)
	}
}

func TestSearch(t *testing.T) {
	setupFixtureDB(t)

	out, _, err := executeCommand(t, "search", "zzznomatch", "--json")
	assertExitCode(t, err, 0)
	if !strings.Contains(out, `"items": []`) {
		t.Errorf("empty search should carry an empty items array, got %q", out)
	}
	if env := decodeList(t, out); env.Total != 0 || len(env.Items) != 0 {
		t.Errorf("empty search envelope should report no items, got total=%d items=%d", env.Total, len(env.Items))
	}

	out, _, err = executeCommand(t, "search", "Project in Today", "--json")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !strings.Contains(out, `"type": "project"`) {
		t.Errorf("cross-type search should include a project:\n%s", out)
	}
}

func TestGlobalFlags(t *testing.T) {
	t.Run("db flag overrides THINGSDB", func(t *testing.T) {
		t.Setenv("THINGSDB", "/nonexistent/path/main.sqlite")
		_, _, err := executeCommand(t, "inbox", "--db", thingstest.DatabasePath(t))
		if err != nil {
			t.Errorf("--db should override THINGSDB: %v", err)
		}
	})

	t.Run("stray arg errors", func(t *testing.T) {
		setupFixtureDB(t)
		_, _, err := executeCommand(t, "inbox", "extra")
		if err == nil {
			t.Errorf("stray arg to inbox should error")
		}
	})
}

func TestOutputFormatFlags(t *testing.T) {
	setupFixtureDB(t)

	t.Run("default is text", func(t *testing.T) {
		out, stderr, err := executeCommand(t, "inbox")
		if err != nil {
			t.Fatalf("inbox: %v (stderr %s)", err, stderr)
		}
		if strings.HasPrefix(strings.TrimSpace(out), "{") {
			t.Errorf("default output should be text, got JSON:\n%s", out)
		}
	})

	for _, flag := range []string{"--json", "-j"} {
		t.Run("json via "+flag, func(t *testing.T) {
			out := runJSON(t, "inbox", flag)
			if got := jsonTotal(t, out); got != thingstest.Inbox {
				t.Errorf("inbox %s total = %d, want %d", flag, got, thingstest.Inbox)
			}
		})
	}

	for _, flag := range []string{"--yaml", "-y"} {
		t.Run("yaml via "+flag, func(t *testing.T) {
			out := runJSON(t, "inbox", flag)
			var env struct {
				Total int `yaml:"total"`
			}
			if err := yaml.Unmarshal([]byte(out), &env); err != nil {
				t.Fatalf("inbox %s is not yaml: %v\n%s", flag, err, out)
			}
			if env.Total != thingstest.Inbox {
				t.Errorf("inbox %s total = %d, want %d", flag, env.Total, thingstest.Inbox)
			}
		})
	}

	t.Run("json and yaml are mutually exclusive", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "inbox", "--json", "--yaml")
		assertExitCode(t, err, 1)
		if !strings.Contains(stderr, flagJSON) || !strings.Contains(stderr, flagYAML) {
			t.Errorf("mutual-exclusion error should name both flags, got %q", stderr)
		}
	})
}

func TestListFlagsGlobalOnListCommands(t *testing.T) {
	setupFixtureDB(t)

	t.Run("sort title orders projects", func(t *testing.T) {
		out := runJSON(t, "projects", "--all", "--sort", "title", "--json")
		var rows []struct {
			Title string `json:"title"`
		}
		decodeItems(t, out, &rows)
		if len(rows) < 2 {
			t.Fatalf("want >= 2 projects to test sorting, got %d", len(rows))
		}
		for i := 1; i < len(rows); i++ {
			if strings.ToLower(rows[i-1].Title) > strings.ToLower(rows[i].Title) {
				t.Errorf("projects --sort title not ascending: %q before %q", rows[i-1].Title, rows[i].Title)
			}
		}
	})

	t.Run("page selects a distinct page", func(t *testing.T) {
		p1 := decodeList(t, runJSON(t, "logbook", "--days", "0", "-n", "1", "--page", "1", "--json"))
		p2 := decodeList(t, runJSON(t, "logbook", "--days", "0", "-n", "1", "--page", "2", "--json"))
		if p1.Page != 1 || p2.Page != 2 {
			t.Fatalf("page metadata wrong: p1=%d p2=%d", p1.Page, p2.Page)
		}
		if len(p1.Items) != 1 || len(p2.Items) != 1 {
			t.Fatalf("each page should carry one item, got %d and %d", len(p1.Items), len(p2.Items))
		}
		if p1.Items[0]["uuid"] == p2.Items[0]["uuid"] {
			t.Errorf("page 1 and page 2 must differ, both %v", p1.Items[0]["uuid"])
		}
	})
}

func TestCollectionsHonorListFlags(t *testing.T) {
	setupFixtureDB(t)

	t.Run("areas envelope reports full total under -n", func(t *testing.T) {
		env := decodeList(t, runJSON(t, "areas", "-n", "1", "--json"))
		if env.Total != thingstest.Areas {
			t.Errorf("areas total = %d, want %d", env.Total, thingstest.Areas)
		}
		if len(env.Items) != 1 {
			t.Errorf("areas -n 1 should return one item, got %d", len(env.Items))
		}
		if env.Pages != thingstest.Areas {
			t.Errorf("areas -n 1 pages = %d, want %d", env.Pages, thingstest.Areas)
		}
	})

	t.Run("areas text shows a footer when paginated", func(t *testing.T) {
		text, _, err := executeCommand(t, "areas", "-n", "1")
		if err != nil {
			t.Fatalf("areas text: %v", err)
		}
		if !strings.Contains(text, "page 1/") {
			t.Errorf("paginated areas text should carry a footer:\n%s", text)
		}
	})

	t.Run("areas --page selects the second page", func(t *testing.T) {
		p2 := decodeList(t, runJSON(t, "areas", "-n", "1", "--page", "2", "--json"))
		if p2.Page != 2 || len(p2.Items) != 1 {
			t.Errorf("areas page 2 wrong: page=%d items=%d", p2.Page, len(p2.Items))
		}
	})

	t.Run("tags honor -n", func(t *testing.T) {
		env := decodeList(t, runJSON(t, "tags", "-n", "2", "--json"))
		if env.Total != thingstest.Tags {
			t.Errorf("tags total = %d, want %d", env.Total, thingstest.Tags)
		}
		if len(env.Items) != 2 {
			t.Errorf("tags -n 2 should return two items, got %d", len(env.Items))
		}
	})
}

func TestHelpGroupsAndExamples(t *testing.T) {
	out, _, err := executeCommand(t, "--help")
	if err != nil {
		t.Fatalf("--help: %v", err)
	}
	for _, group := range []string{"Views:", "Collections:", "Lookup:"} {
		if !strings.Contains(out, group) {
			t.Errorf("--help missing group %q", group)
		}
	}

	var walk func(parent *cobra.Command, topLevel bool)
	walk = func(parent *cobra.Command, topLevel bool) {
		for _, c := range parent.Commands() {
			if c.Hidden {
				continue
			}
			switch c.Name() {
			case "help", "completion", "version":
				continue
			}
			if c.Short == "" || !startsUpper(c.Short) {
				t.Errorf("command %q Short must be non-empty and capitalized, got %q", c.Name(), c.Short)
			}
			if c.Example == "" {
				t.Errorf("command %q has no Example", c.Name())
			}
			if topLevel && c.GroupID == "" {
				t.Errorf("top-level command %q has no GroupID", c.Name())
			}
			walk(c, false)
		}
	}
	walk(NewRootCmd(), true)
}

func startsUpper(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

func TestProjects(t *testing.T) {
	dbPath := setupFixtureDB(t)
	ctx := context.Background()
	client, err := things3.NewClient(things3.WithDatabasePath(dbPath))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	incomplete, err := client.Projects().Status().Incomplete().All(ctx)
	if err != nil {
		t.Fatalf("incomplete projects: %v", err)
	}
	anyStatus, err := client.Projects().Status().Any().All(ctx)
	if err != nil {
		t.Fatalf("all projects: %v", err)
	}
	area1, err := client.Areas().WithTitle("Area 1").First(ctx)
	if err != nil {
		t.Fatalf("area 1: %v", err)
	}
	inArea1, err := client.Projects().Status().Incomplete().InArea(area1.UUID).All(ctx)
	if err != nil {
		t.Fatalf("projects in area 1: %v", err)
	}
	_ = client.Close()

	t.Run("default incomplete", func(t *testing.T) {
		out, _, err := executeCommand(t, "projects", "--json")
		if err != nil {
			t.Fatalf("projects: %v", err)
		}
		if got := jsonTotal(t, out); got != len(incomplete) {
			t.Errorf("projects count = %d, want %d", got, len(incomplete))
		}
	})

	t.Run("all includes closed", func(t *testing.T) {
		out, _, err := executeCommand(t, "projects", "--all", "--json")
		if err != nil {
			t.Fatalf("projects --all: %v", err)
		}
		if got := jsonTotal(t, out); got != len(anyStatus) {
			t.Errorf("projects --all count = %d, want %d", got, len(anyStatus))
		}
		if len(anyStatus) <= len(incomplete) {
			t.Errorf("fixture should have closed projects (any=%d, incomplete=%d)", len(anyStatus), len(incomplete))
		}
	})

	t.Run("area filter", func(t *testing.T) {
		out, _, err := executeCommand(t, "projects", "--area", "Area 1", "--json")
		if err != nil {
			t.Fatalf("projects --area Area 1: %v", err)
		}
		if got := jsonTotal(t, out); got != len(inArea1) {
			t.Errorf("projects --area Area 1 count = %d, want %d", got, len(inArea1))
		}
	})

	t.Run("area not found exits 1", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "projects", "--area", "zzznope")
		assertExitCode(t, err, 1)
		if !strings.Contains(stderr, "Error:") {
			t.Errorf("expected Error on stderr, got %q", stderr)
		}
	})

	t.Run("ambiguous area exits 2 with candidates", func(t *testing.T) {
		_, stderr, err := executeCommand(t, "projects", "--area", "Area")
		assertExitCode(t, err, 2)
		if !strings.Contains(stderr, "Error:") || !strings.Contains(stderr, "area") {
			t.Errorf("text ambiguity should list area candidates:\n%s", stderr)
		}

		_, stderrJSON, err := executeCommand(t, "projects", "--area", "Area", "--json")
		assertExitCode(t, err, 2)
		if !strings.Contains(stderrJSON, "candidates") || !strings.Contains(stderrJSON, "area") {
			t.Errorf("json ambiguity should carry a candidates array:\n%s", stderrJSON)
		}
	})
}

func TestShowProjectDetail(t *testing.T) {
	setupFixtureDB(t)
	out, _, err := executeCommand(t, "show", thingstest.UUIDProject)
	if err != nil {
		t.Fatalf("show project: %v", err)
	}
	if !strings.Contains(out, "Title:") {
		t.Errorf("project show should render a detail view:\n%s", out)
	}

	outJSON, _, err := executeCommand(t, "show", thingstest.UUIDProject, "--json")
	if err != nil {
		t.Fatalf("show project json: %v", err)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(outJSON), &obj); err != nil {
		t.Fatalf("project detail json invalid: %v\n%s", err, outJSON)
	}
	if _, ok := obj["project"]; !ok {
		t.Errorf("project detail json missing 'project' key")
	}
	if _, ok := obj["todos"]; !ok {
		t.Errorf("project detail json missing 'todos' key")
	}
}

func TestOutputYAML(t *testing.T) {
	setupFixtureDB(t)
	out, _, err := executeCommand(t, "trash", "--yaml")
	if err != nil {
		t.Fatalf("trash --yaml: %v", err)
	}
	var env struct {
		Items []map[string]any `yaml:"items"`
		Total int              `yaml:"total"`
		Page  int              `yaml:"page"`
		Pages int              `yaml:"pages"`
	}
	if err := yaml.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("trash yaml is invalid: %v\n%s", err, out)
	}
	if len(env.Items) == 0 {
		t.Fatalf("expected trashed items in yaml output:\n%s", out)
	}
	if env.Page != 1 || env.Pages < 1 || env.Total < len(env.Items) {
		t.Errorf("yaml envelope keys wrong: total=%d page=%d pages=%d items=%d", env.Total, env.Page, env.Pages, len(env.Items))
	}
	for _, it := range env.Items {
		if _, ok := it["type"]; !ok {
			t.Errorf("mixed yaml item missing 'type': %v", it)
		}
	}
}

func TestCollectionsAndSomeday(t *testing.T) {
	dbPath := setupFixtureDB(t)
	ctx := context.Background()
	client, err := things3.NewClient(things3.WithDatabasePath(dbPath))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	someday, err := client.Todos().StartDate().Exists(false).Start().Someday().Status().Incomplete().All(ctx)
	if err != nil {
		t.Fatalf("someday: %v", err)
	}
	_ = client.Close()

	sd, _, err := executeCommand(t, "someday", "--json")
	if err != nil {
		t.Fatalf("someday: %v", err)
	}
	if got := jsonTotal(t, sd); got != len(someday) {
		t.Errorf("someday count = %d, want %d", got, len(someday))
	}

	areasText, _, err := executeCommand(t, "areas")
	if err != nil {
		t.Fatalf("areas: %v", err)
	}
	if !strings.Contains(areasText, "- Area 1") {
		t.Errorf("areas text should list '- Area 1':\n%s", areasText)
	}
	areasJSON, _, err := executeCommand(t, "areas", "--json")
	if err != nil {
		t.Fatalf("areas json: %v", err)
	}
	if got := jsonTotal(t, areasJSON); got != thingstest.Areas {
		t.Errorf("areas count = %d, want %d", got, thingstest.Areas)
	}

	tagsJSON, _, err := executeCommand(t, "tags", "--json")
	if err != nil {
		t.Fatalf("tags json: %v", err)
	}
	if got := jsonTotal(t, tagsJSON); got != thingstest.Tags {
		t.Errorf("tags count = %d, want %d", got, thingstest.Tags)
	}
}

func TestDeadlinesAscending(t *testing.T) {
	setupFixtureDB(t)
	out, _, err := executeCommand(t, "deadlines", "--json")
	if err != nil {
		t.Fatalf("deadlines: %v", err)
	}
	var rows []struct {
		Deadline *time.Time `json:"deadline"`
	}
	decodeItems(t, out, &rows)
	if len(rows) < 2 {
		t.Fatalf("want >= 2 deadlines to test ordering, got %d", len(rows))
	}
	for i := 1; i < len(rows); i++ {
		if rows[i-1].Deadline != nil && rows[i].Deadline != nil && rows[i].Deadline.Before(*rows[i-1].Deadline) {
			t.Errorf("deadlines not ascending: row %d precedes earlier row %d", i-1, i)
		}
	}
}

// TestDeadlinesDaysWindow covers the --days flag on deadlines: a tight window
// narrows the list, and a window so wide it leaves the encodable date range still
// yields only todos that actually carry a deadline. Before the date clamp such a
// window dropped the deadline condition entirely and printed every incomplete todo.
func TestDeadlinesDaysWindow(t *testing.T) {
	setupFixtureDB(t)
	deadlineTotals := func(args ...string) (total, undated int) {
		t.Helper()
		out := runJSON(t, append([]string{"deadlines", "--json", "--all"}, args...)...)
		var rows []struct {
			Deadline *time.Time `json:"deadline"`
		}
		decodeItems(t, out, &rows)
		for _, r := range rows {
			if r.Deadline == nil {
				undated++
			}
		}
		return decodeList(t, out).Total, undated
	}

	all, undated := deadlineTotals()
	if all != thingstest.Deadlines || undated != 0 {
		t.Fatalf("deadlines = %d rows (%d undated), want %d with none undated", all, undated, thingstest.Deadlines)
	}
	if near, _ := deadlineTotals("--days", "1"); near >= all {
		t.Errorf("--days 1 = %d rows, want fewer than the unwindowed %d", near, all)
	}
	absurd, absurdUndated := deadlineTotals("--days", "3000000")
	if absurd != all || absurdUndated != 0 {
		t.Errorf("--days 3000000 = %d rows (%d undated), want all %d deadlines and no undated todo", absurd, absurdUndated, all)
	}
}

// TestDaysRejectsNegative pins the flag contract shared with the MCP days
// argument: a negative window is an error, not a silently ignored no-op.
func TestDaysRejectsNegative(t *testing.T) {
	setupFixtureDB(t)
	for _, view := range []string{"upcoming", "logbook", "deadlines"} {
		t.Run(view, func(t *testing.T) {
			_, _, err := executeCommand(t, view, "--days", "-7")
			if err == nil {
				t.Fatalf("%s --days -7 should be rejected", view)
			}
			assertExitCode(t, err, 1)
		})
	}
}

func TestGroupedViewsJSONFlat(t *testing.T) {
	dbPath := setupFixtureDB(t)
	ctx := context.Background()
	client, err := things3.NewClient(things3.WithDatabasePath(dbPath))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	anytime, err := client.Todos().Start().Anytime().Status().Incomplete().All(ctx)
	if err != nil {
		t.Fatalf("anytime: %v", err)
	}
	_ = client.Close()

	for _, view := range []string{"upcoming", "anytime"} {
		viewOut, _, runErr := executeCommand(t, view, "--json")
		if runErr != nil {
			t.Fatalf("%s json: %v", view, runErr)
		}
		var arr []map[string]any
		decodeItems(t, viewOut, &arr)
		for _, el := range arr {
			if _, ok := el["uuid"]; !ok {
				t.Errorf("%s json element is not a todo (no uuid): %v", view, el)
			}
			if _, ok := el["header"]; ok {
				t.Errorf("%s json leaked a group header: %v", view, el)
			}
		}
	}

	out, _, err := executeCommand(t, "anytime", "--all", "--json")
	if err != nil {
		t.Fatalf("anytime json: %v", err)
	}
	if got := jsonTotal(t, out); got != len(anytime) {
		t.Errorf("anytime json count = %d, want %d (grouping must not drop rows)", got, len(anytime))
	}
}

func TestRenderErrorTruncatesCandidates(t *testing.T) {
	root := NewRootCmd()
	var errBuf bytes.Buffer
	root.SetErr(&errBuf)
	candidates := make([]Candidate, 15)
	for i := range candidates {
		candidates[i] = Candidate{UUID: "abcd1234", Type: "todo", Title: "item"}
	}
	RenderError(root, &AmbiguousError{Query: "x", Candidates: candidates})
	if !strings.Contains(errBuf.String(), "... and 5 more") {
		t.Errorf("expected candidate truncation, got:\n%s", errBuf.String())
	}
}
