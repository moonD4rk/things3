package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

// lineContaining returns the first line of text that contains needle, failing
// the test when none does.
func lineContaining(t *testing.T, text, needle string) string {
	t.Helper()
	for line := range strings.SplitSeq(text, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	t.Fatalf("no line contains %q in:\n%s", needle, text)
	return ""
}

func TestUpcomingIncludesRepeatingTemplate(t *testing.T) {
	setupFixtureDB(t)
	out, stderr, err := executeCommand(t, "upcoming")
	if err != nil {
		t.Fatalf("upcoming: %v (stderr %s)", err, stderr)
	}
	// The fixture's only repeating template surfaces at its next occurrence
	// (2040-01-01, the far-future representable date the library stores) with the
	// repeats marker.
	for _, want := range []string{"Repeating To-Do", "N1PJHsbj", "2040-01-01", "repeats"} {
		if !strings.Contains(out, want) {
			t.Errorf("upcoming should contain %q:\n%s", want, out)
		}
	}
}

func TestTodoContainerSegment(t *testing.T) {
	cases := []struct {
		name string
		todo things3.Todo
		want string
	}{
		{"project with heading", things3.Todo{ProjectTitle: "Launch", HeadingTitle: "Docs"}, "@Launch / Docs"},
		{"project only", things3.Todo{ProjectTitle: "Launch"}, "@Launch"},
		{"area only", things3.Todo{AreaTitle: "Work"}, "@Work"},
		{"project wins over area", things3.Todo{ProjectTitle: "Launch", AreaTitle: "Work"}, "@Launch"},
		{"heading without project omitted", things3.Todo{HeadingTitle: "Docs"}, ""},
		{"no container", things3.Todo{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			todo := tc.todo
			if got := todoContainer(&todo); got != tc.want {
				t.Errorf("todoContainer = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestContainerSuffixInFlatRows(t *testing.T) {
	setupFixtureDB(t)
	out, stderr, err := executeCommand(t, "logbook", "--days", "0", "--all")
	if err != nil {
		t.Fatalf("logbook: %v (stderr %s)", err, stderr)
	}

	if row := lineContaining(t, out, "5u2yGhP4"); !strings.Contains(row, "@Project without Area") {
		t.Errorf("project todo row should carry its container: %q", row)
	}
	if row := lineContaining(t, out, "UwNEL2Wd"); !strings.Contains(row, "@Area 1") {
		t.Errorf("area todo row should carry its container: %q", row)
	}
}

func TestAnytimeRowsOmitContainer(t *testing.T) {
	setupFixtureDB(t)
	out, stderr, err := executeCommand(t, "anytime", "--all")
	if err != nil {
		t.Fatalf("anytime: %v (stderr %s)", err, stderr)
	}
	for line := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		// Rows begin with a status checkbox; headers are bare container names.
		if strings.HasPrefix(line, "[") && strings.Contains(line, "@") {
			t.Errorf("anytime already groups by container, so rows must not repeat it: %q", line)
		}
	}
}

func TestPaginationFooter(t *testing.T) {
	setupFixtureDB(t)
	// The fixture has exactly four deadline todos, giving stable page math.
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"first page", []string{"deadlines", "-n", "2"}, "-- 1-2 of 4 (page 1/2) | next: --page 2 | all: --all"},
		{"middle page", []string{"deadlines", "-n", "1", "--page", "2"}, "-- 2-2 of 4 (page 2/4) | next: --page 3 | all: --all"},
		{"last page", []string{"deadlines", "-n", "2", "--page", "2"}, "-- 3-4 of 4 (page 2/2) | all: --all"},
		{"out of range", []string{"deadlines", "-n", "2", "--page", "3"}, "-- 0-0 of 4 (page 3/2) | all: --all"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, stderr, err := executeCommand(t, tc.args...)
			if err != nil {
				t.Fatalf("%v: %v (stderr %s)", tc.args, err, stderr)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("footer for %v:\ngot:\n%s\nwant line: %q", tc.args, out, tc.want)
			}
		})
	}
}

func TestNoFooterCases(t *testing.T) {
	setupFixtureDB(t)
	cases := [][]string{
		{"deadlines", "-n", "0"},
		{"deadlines", "--all"},
		{"deadlines", "--json"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			out, stderr, err := executeCommand(t, args...)
			if err != nil {
				t.Fatalf("%v: %v (stderr %s)", args, err, stderr)
			}
			if strings.Contains(out, "(page ") {
				t.Errorf("%v should print no footer:\n%s", args, out)
			}
		})
	}
}

func TestOutOfRangePageJSONIsEmptyArray(t *testing.T) {
	setupFixtureDB(t)
	// An out-of-range page must still encode items as [] (never null) and exit 0.
	out, stderr, err := executeCommand(t, "deadlines", "--json", "-n", "2", "--page", "9")
	if err != nil {
		t.Fatalf("deadlines json out of range: %v (stderr %s)", err, stderr)
	}
	assertExitCode(t, err, 0)
	if !strings.Contains(out, `"items": []`) {
		t.Errorf("out-of-range json page should carry an empty items array:\ngot: %q", out)
	}
	env := decodeList(t, out)
	if env.Items == nil {
		t.Errorf("items must decode to [] not null:\n%s", out)
	}
	if len(env.Items) != 0 {
		t.Errorf("out-of-range page should hold no items, got %d", len(env.Items))
	}
	if env.Total != thingstest.Deadlines || env.Page != 9 {
		t.Errorf("envelope should still report total=%d page=9, got total=%d page=%d", thingstest.Deadlines, env.Total, env.Page)
	}
}

func TestListDefaultsByFormat(t *testing.T) {
	setupFixtureDB(t)

	// Text mode defaults to a page so a long list stays readable.
	text, stderr, err := executeCommand(t, "logbook", "--days", "0")
	if err != nil {
		t.Fatalf("logbook text: %v (stderr %s)", err, stderr)
	}
	if strings.Count(text, "\n[") > defaultPageSize {
		t.Errorf("text default should cap the page at %d rows:\n%s", defaultPageSize, text)
	}
	if !strings.Contains(text, "(page 1/") || !strings.Contains(text, "| all: --all") {
		t.Errorf("text default over one page should print a footer:\n%s", text)
	}

	// JSON now paginates uniformly with text; the envelope self-describes the slice
	// instead of silently truncating, so items caps at the default page size while
	// total reports the full count.
	jsonOut := runJSON(t, "logbook", "--days", "0", "--json")
	env := decodeList(t, jsonOut)
	if env.Total <= defaultPageSize {
		t.Fatalf("fixture logbook (%d) must exceed the default page to test capping", env.Total)
	}
	if len(env.Items) != defaultPageSize {
		t.Errorf("default json page should cap items at %d, got %d", defaultPageSize, len(env.Items))
	}
	if env.Page != 1 {
		t.Errorf("default json page should be 1, got %d", env.Page)
	}
	if strings.Contains(jsonOut, "(page ") {
		t.Errorf("json output must not carry a text footer:\n%s", jsonOut)
	}
}

// sortRow decodes the fields the sort pipeline keys on.
type sortRow struct {
	Title       string     `json:"title"`
	CompletedAt *time.Time `json:"completed_at"`
	CanceledAt  *time.Time `json:"canceled_at"`
	Deadline    *time.Time `json:"deadline"`
	StartDate   *time.Time `json:"start_date"`
}

func (r sortRow) date() *time.Time {
	switch {
	case r.CompletedAt != nil:
		return r.CompletedAt
	case r.CanceledAt != nil:
		return r.CanceledAt
	case r.Deadline != nil:
		return r.Deadline
	case r.StartDate != nil:
		return r.StartDate
	default:
		return nil
	}
}

func decodeSortRows(t *testing.T, s string) []sortRow {
	t.Helper()
	var rows []sortRow
	decodeItems(t, s, &rows)
	return rows
}

// assertDateSorted checks that non-nil dates are ordered per desc and every
// nil-dated row sorts after all dated rows.
func assertDateSorted(t *testing.T, rows []sortRow, desc bool) {
	t.Helper()
	var sawNil bool
	for i := range rows {
		d := rows[i].date()
		if d == nil {
			sawNil = true
			continue
		}
		if sawNil {
			t.Errorf("desc=%v: dated row %d follows a nil-dated row", desc, i)
		}
		if i == 0 {
			continue
		}
		prev := rows[i-1].date()
		if prev == nil {
			continue
		}
		if (!desc && d.Before(*prev)) || (desc && d.After(*prev)) {
			t.Errorf("desc=%v: dates out of order at %d", desc, i)
		}
	}
}

func assertTitleSorted(t *testing.T, rows []sortRow, desc bool) {
	t.Helper()
	for i := 1; i < len(rows); i++ {
		prev, curr := strings.ToLower(rows[i-1].Title), strings.ToLower(rows[i].Title)
		if (!desc && curr < prev) || (desc && curr > prev) {
			t.Errorf("desc=%v: titles out of order at %d: %q before %q", desc, i, rows[i-1].Title, rows[i].Title)
		}
	}
}

func TestSortListFlag(t *testing.T) {
	setupFixtureDB(t)

	t.Run("title ascending", func(t *testing.T) {
		out, _, err := executeCommand(t, "logbook", "--days", "0", "--all", "--sort", "title", "--json")
		if err != nil {
			t.Fatalf("sort title: %v", err)
		}
		assertTitleSorted(t, decodeSortRows(t, out), false)
	})

	t.Run("title descending", func(t *testing.T) {
		out, _, err := executeCommand(t, "logbook", "--days", "0", "--all", "--sort", "title", "--desc", "--json")
		if err != nil {
			t.Fatalf("sort title desc: %v", err)
		}
		assertTitleSorted(t, decodeSortRows(t, out), true)
	})

	t.Run("date sorts nil last both directions", func(t *testing.T) {
		for _, desc := range []bool{false, true} {
			args := []string{"anytime", "--all", "--sort", "date", "--json"}
			if desc {
				args = append(args, "--desc")
			}
			out, _, err := executeCommand(t, args...)
			if err != nil {
				t.Fatalf("sort date desc=%v: %v", desc, err)
			}
			assertDateSorted(t, decodeSortRows(t, out), desc)
		}
	})
}

func TestTagFilterListFlag(t *testing.T) {
	setupFixtureDB(t)

	// Match is case-insensitive: "PENDING" finds the "Pending"-tagged todo.
	out, stderr, err := executeCommand(t, "anytime", "--tag", "PENDING", "--json")
	if err != nil {
		t.Fatalf("tag filter: %v (stderr %s)", err, stderr)
	}
	var rows []struct {
		UUID string   `json:"uuid"`
		Tags []string `json:"tags"`
	}
	decodeItems(t, out, &rows)
	if len(rows) != 1 {
		t.Fatalf("expected exactly one Pending todo, got %d:\n%s", len(rows), out)
	}
	for _, r := range rows {
		var has bool
		for _, tag := range r.Tags {
			if strings.EqualFold(tag, "pending") {
				has = true
			}
		}
		if !has {
			t.Errorf("filtered row %s lacks the Pending tag: %v", r.UUID, r.Tags)
		}
	}

	// The footer total reflects the filtered count (1), not the full anytime view.
	foot, _, err := executeCommand(t, "anytime", "--tag", "pending", "-n", "1", "--page", "2")
	if err != nil {
		t.Fatalf("tag filter footer: %v", err)
	}
	if !strings.Contains(foot, "-- 0-0 of 1 (page 2/1) | all: --all") {
		t.Errorf("footer total should reflect the filtered count:\n%s", foot)
	}
}

func TestListEnvelopePaginates(t *testing.T) {
	setupFixtureDB(t)
	// The fixture has exactly four deadline todos, giving stable page math. With a
	// page size of one, --page must select a different item on every page in json.
	page2 := decodeList(t, runJSON(t, "deadlines", "--json", "-n", "1", "--page", "2"))
	page3 := decodeList(t, runJSON(t, "deadlines", "--json", "-n", "1", "--page", "3"))

	for _, env := range []listEnvelope{page2, page3} {
		if env.Total != thingstest.Deadlines {
			t.Errorf("total = %d, want %d", env.Total, thingstest.Deadlines)
		}
		if env.Pages != thingstest.Deadlines {
			t.Errorf("pages = %d, want %d", env.Pages, thingstest.Deadlines)
		}
		if len(env.Items) != 1 {
			t.Fatalf("a one-item page should hold one item, got %d", len(env.Items))
		}
	}
	if page2.Page != 2 || page3.Page != 3 {
		t.Errorf("page numbers wrong: page2=%d page3=%d", page2.Page, page3.Page)
	}
	if page2.Items[0]["uuid"] == page3.Items[0]["uuid"] {
		t.Errorf("page 2 and page 3 must differ, both returned %v", page2.Items[0]["uuid"])
	}
}

func TestListEnvelopeUnlimited(t *testing.T) {
	setupFixtureDB(t)
	for _, args := range [][]string{
		{"deadlines", "--json", "-n", "0"},
		{"deadlines", "--json", "--all"},
	} {
		env := decodeList(t, runJSON(t, args...))
		if env.Page != 1 || env.Pages != 1 {
			t.Errorf("%v: unlimited output should report page 1/1, got %d/%d", args, env.Page, env.Pages)
		}
		if len(env.Items) != env.Total {
			t.Errorf("%v: unlimited items %d should equal total %d", args, len(env.Items), env.Total)
		}
		if env.Total != thingstest.Deadlines {
			t.Errorf("%v: total = %d, want %d", args, env.Total, thingstest.Deadlines)
		}
	}
}

func TestAreasTagsEnvelope(t *testing.T) {
	setupFixtureDB(t)
	cases := []struct {
		cmd  string
		want int
	}{
		{"areas", thingstest.Areas},
		{"tags", thingstest.Tags},
	}
	for _, tc := range cases {
		t.Run(tc.cmd, func(t *testing.T) {
			env := decodeList(t, runJSON(t, tc.cmd, "--json"))
			if env.Total != tc.want || len(env.Items) != tc.want {
				t.Errorf("%s envelope total/items = %d/%d, want %d", tc.cmd, env.Total, len(env.Items), tc.want)
			}
			if env.Page != 1 || env.Pages != 1 {
				t.Errorf("%s envelope page/pages = %d/%d, want 1/1", tc.cmd, env.Page, env.Pages)
			}
		})
	}
}
