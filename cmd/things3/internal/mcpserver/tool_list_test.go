package mcpserver

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

// serverOnPath builds a server over an explicit fixture copy, so a test can both
// query it and mutate the same file with execFixture.
func serverOnPath(t *testing.T, path string, cfg Config) *Server {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(path))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	srv, err := New(client, cfg)
	if err != nil {
		t.Fatalf("server: %v", err)
	}
	return srv
}

func listTodos(t *testing.T, srv *Server, in ListTodosInput) PageResult[Item] {
	t.Helper()
	_, page, err := srv.handleListTodos(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("list_todos %+v: %v", in, err)
	}
	return page
}

// todosWithDays lists a date-ordered view with an explicit days window, building
// the pointer the input expects from a plain value so call sites stay readable.
func todosWithDays(t *testing.T, srv *Server, view string, days, limit int) PageResult[Item] {
	t.Helper()
	d := days
	return listTodos(t, srv, ListTodosInput{View: ViewName(view), Days: &d, Limit: limit})
}

func listProjects(t *testing.T, srv *Server, in ListProjectsInput) PageResult[Item] {
	t.Helper()
	_, page, err := srv.handleListProjects(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("list_projects %+v: %v", in, err)
	}
	return page
}

func TestListTodosViewCounts(t *testing.T) {
	srv := newTestServer(t, Config{})
	// The six single-builder views are status/bucket/existence-scoped, so their
	// totals are stable regardless of the current date (unlike today/upcoming,
	// which are checked by membership). Pinning every one is the WS1 parity guard
	// that the push-of-scoping-into-SQL refactor preserved each view's result.
	zero := 0
	cases := []struct {
		view string
		days *int
		want int
	}{
		{"inbox", nil, thingstest.Inbox},          // 2
		{"anytime", nil, thingstest.TodosAnytime}, // 10
		{"someday", nil, 1},
		{"logbook", &zero, 22},                   // days:0 = all history; the 30-day default returns 0 on the aged fixture
		{"deadlines", nil, thingstest.Deadlines}, // 4
		{"trash", nil, 5},                        // trashed todos of any status; projects excluded
	}
	for _, tc := range cases {
		t.Run(tc.view, func(t *testing.T) {
			page := listTodos(t, srv, ListTodosInput{View: ViewName(tc.view), Days: tc.days})
			if page.Total != tc.want {
				t.Errorf("%s total = %d, want %d", tc.view, page.Total, tc.want)
			}
			for i := range page.Items {
				if page.Items[i].Type != typeTodo {
					t.Errorf("%s item %d type = %q", tc.view, i, page.Items[i].Type)
				}
			}
		})
	}
}

func TestListTodosTodayContainsInToday(t *testing.T) {
	srv := newTestServer(t, Config{})
	page := listTodos(t, srv, ListTodosInput{View: "today", Limit: 100})
	found := false
	for i := range page.Items {
		if page.Items[i].UUID == thingstest.UUIDTodoInToday {
			found = true
		}
	}
	if !found {
		t.Errorf("today should contain %s", thingstest.UUIDTodoInToday)
	}
}

func TestListTodosPagination(t *testing.T) {
	srv := newTestServer(t, Config{})

	p1 := listTodos(t, srv, ListTodosInput{View: "inbox", Limit: 1, Page: 1})
	if p1.Total != thingstest.Inbox || len(p1.Items) != 1 || p1.Page != 1 || p1.Pages != 2 {
		t.Fatalf("page 1 = %+v", p1)
	}
	p2 := listTodos(t, srv, ListTodosInput{View: "inbox", Limit: 1, Page: 2})
	if len(p2.Items) != 1 || p2.Page != 2 {
		t.Fatalf("page 2 = %+v", p2)
	}
	if p1.Items[0].UUID == p2.Items[0].UUID {
		t.Errorf("pages 1 and 2 must differ, both %s", p1.Items[0].UUID)
	}

	p3 := listTodos(t, srv, ListTodosInput{View: "inbox", Limit: 1, Page: 9})
	if len(p3.Items) != 0 || p3.Total != thingstest.Inbox || p3.Pages != 2 {
		t.Errorf("out-of-range page = %+v", p3)
	}
	if p3.Items == nil {
		t.Errorf("items must be a non-nil empty array")
	}
}

func TestClampLimit(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, DefaultLimit},
		{-5, DefaultLimit},
		{7, 7},
		{MaxLimit, MaxLimit},
		{MaxLimit + 50, MaxLimit},
	}
	for _, tc := range cases {
		if got := clampLimit(tc.in, DefaultLimit, MaxLimit); got != tc.want {
			t.Errorf("clampLimit(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestListTodosFilters(t *testing.T) {
	srv := newTestServer(t, Config{})

	t.Run("bogus tag empties the list", func(t *testing.T) {
		page := listTodos(t, srv, ListTodosInput{View: "anytime", Tag: "zzznope", Limit: 100})
		if page.Total != 0 {
			t.Errorf("bogus tag total = %d, want 0", page.Total)
		}
	})

	t.Run("ambiguous project rides the envelope", func(t *testing.T) {
		_, page, err := srv.handleListTodos(context.Background(), nil, ListTodosInput{View: "anytime", Project: "Project in"})
		if err != nil {
			t.Fatalf("ambiguity is not a transport error: %v", err)
		}
		if page.Success || page.Error == nil || page.Error.Code != codeAmbiguous {
			t.Fatalf("want ambiguous envelope, got %+v", page)
		}
		if len(page.Error.Candidates) < 2 {
			t.Errorf("ambiguous error should carry candidates, got %+v", page.Error.Candidates)
		}
	})

	t.Run("area filter never grows the list", func(t *testing.T) {
		all := listTodos(t, srv, ListTodosInput{View: "anytime", Limit: 100})
		filtered := listTodos(t, srv, ListTodosInput{View: "anytime", Area: "Area 1", Limit: 100})
		if filtered.Total > all.Total {
			t.Errorf("area filter grew the list: %d > %d", filtered.Total, all.Total)
		}
		for i := range filtered.Items {
			if a := filtered.Items[i].Area; a != nil && a.Title != "Area 1" {
				t.Errorf("area filter leaked a %q item", a.Title)
			}
		}
	})
}

func TestLogbookDescending(t *testing.T) {
	srv := newTestServer(t, Config{})
	// Days:0 asks for all history; the default 30-day window would return 0 rows
	// against the aged fixture and silently skip this order check.
	page := todosWithDays(t, srv, "logbook", 0, 100)
	if len(page.Items) < 2 {
		t.Skipf("need >= 2 logbook rows, got %d", len(page.Items))
	}
	for i := 1; i < len(page.Items); i++ {
		prev, curr := page.Items[i-1].CompletedAt, page.Items[i].CompletedAt
		if prev != "" && curr != "" && curr > prev {
			t.Errorf("logbook not descending: row %d (%s) newer than row %d (%s)", i, curr, i-1, prev)
		}
	}
}

func TestDeadlinesAscending(t *testing.T) {
	srv := newTestServer(t, Config{})
	page := listTodos(t, srv, ListTodosInput{View: "deadlines", Limit: 100})
	if len(page.Items) < 2 {
		t.Skipf("need >= 2 deadline rows, got %d", len(page.Items))
	}
	for i := 1; i < len(page.Items); i++ {
		prev, curr := page.Items[i-1].Deadline, page.Items[i].Deadline
		if prev != "" && curr != "" && curr < prev {
			t.Errorf("deadlines not ascending: row %d (%s) precedes %d (%s)", i, curr, i-1, prev)
		}
	}
}

func TestListProjectsStatus(t *testing.T) {
	srv := newTestServer(t, Config{})

	inc := listProjects(t, srv, ListProjectsInput{})
	for i := range inc.Items {
		if inc.Items[i].Status != statusIncomplete {
			t.Errorf("default status should be incomplete, got %q", inc.Items[i].Status)
		}
		if inc.Items[i].Type != typeProject {
			t.Errorf("item type = %q, want project", inc.Items[i].Type)
		}
	}

	all := listProjects(t, srv, ListProjectsInput{Status: "any"})
	if all.Total <= inc.Total {
		t.Errorf("status=any (%d) should exceed incomplete (%d)", all.Total, inc.Total)
	}
}

func TestListAreasAndTags(t *testing.T) {
	srv := newTestServer(t, Config{})

	_, areas, err := srv.handleListAreas(context.Background(), nil, ListAreasInput{})
	if err != nil {
		t.Fatalf("list_areas: %v", err)
	}
	if areas.Total != thingstest.Areas {
		t.Errorf("areas total = %d, want %d", areas.Total, thingstest.Areas)
	}

	_, tags, err := srv.handleListTags(context.Background(), nil, ListTagsInput{})
	if err != nil {
		t.Fatalf("list_tags: %v", err)
	}
	if tags.Total != thingstest.Tags {
		t.Errorf("tags total = %d, want %d", tags.Total, thingstest.Tags)
	}
}

// TestListTodosProjectIncludesHeadingTodos proves the project filter keeps todos
// filed under one of the project's headings, whose own ProjectUUID is empty. This
// mirrors the library's InProject OR-semantics; a bare ProjectUUID compare drops
// them and diverges from get(project).
func TestListTodosProjectIncludesHeadingTodos(t *testing.T) {
	srv := newTestServer(t, Config{Version: "test"})
	// "To-Do in Heading" is an anytime, incomplete todo whose heading belongs to
	// UUIDProject; its own project column is empty.
	const headingTodo = "HbKGAeZKFDkWH5osSBNHvz"
	all := listTodos(t, srv, ListTodosInput{View: "anytime", Limit: 100})
	page := listTodos(t, srv, ListTodosInput{View: "anytime", Project: thingstest.UUIDProject, Limit: 100})
	if !slices.ContainsFunc(page.Items, func(it Item) bool { return it.UUID == headingTodo }) {
		t.Errorf("project filter dropped heading-todo %s; got %d items", headingTodo, len(page.Items))
	}
	// Pushing InProject into SQL must stay a bounded, non-leaking subset: every
	// item belongs to the project directly or via one of its headings.
	if page.Total == 0 || page.Total > all.Total {
		t.Fatalf("project scope total = %d, want 1..%d", page.Total, all.Total)
	}
	for i := range page.Items {
		it := page.Items[i]
		direct := it.Project != nil && it.Project.UUID == thingstest.UUIDProject
		if !direct && it.Heading == nil {
			t.Errorf("project filter leaked %s (project=%v, heading=%v)", it.UUID, it.Project, it.Heading)
		}
	}
}

func TestDaysInvalidView(t *testing.T) {
	srv := newTestServer(t, Config{})
	three, negOne := 3, -1
	// days is meaningless on the non-date-ordered views and rides the envelope as
	// invalid_input, never a transport error, before any fetch runs.
	for _, view := range []string{"inbox", "today", "anytime", "someday", "trash"} {
		_, page, err := srv.handleListTodos(context.Background(), nil, ListTodosInput{View: ViewName(view), Days: &three})
		if err != nil {
			t.Fatalf("%s: days rejection must not be a transport error: %v", view, err)
		}
		if page.Success || page.Error == nil || page.Error.Code != codeInvalidInput {
			t.Errorf("%s: want invalid_input envelope, got %+v", view, page)
		}
	}
	// Positive control: the three date views accept a window.
	for _, view := range []string{"upcoming", "logbook", "deadlines"} {
		if page := todosWithDays(t, srv, view, 7, 0); !page.Success {
			t.Errorf("%s: days should be accepted, got %+v", view, page.Error)
		}
	}
	// A negative window is rejected the same way.
	_, neg, err := srv.handleListTodos(context.Background(), nil, ListTodosInput{View: "logbook", Days: &negOne})
	if err != nil {
		t.Fatalf("negative days must not be a transport error: %v", err)
	}
	if neg.Success || neg.Error == nil || neg.Error.Code != codeInvalidInput {
		t.Errorf("negative days: want invalid_input, got %+v", neg)
	}
}

func TestDeadlinesWindow(t *testing.T) {
	srv := newTestServer(t, Config{})
	const farFuture = "HbKGAeZKFDkWH5osSBNHvz" // deadline 2040-11-04, well beyond any near window

	all := listTodos(t, srv, ListTodosInput{View: "deadlines", Limit: 100})
	if all.Total != thingstest.Deadlines {
		t.Fatalf("deadlines (no window) = %d, want %d", all.Total, thingstest.Deadlines)
	}
	// A tight forward window keeps the overdue (2021) deadlines and drops the 2040 one.
	near := todosWithDays(t, srv, "deadlines", 1, 100)
	if slices.ContainsFunc(near.Items, func(it Item) bool { return it.UUID == farFuture }) {
		t.Errorf("days:1 must exclude the far-future deadline %s", farFuture)
	}
	if near.Total != thingstest.Deadlines-1 {
		t.Errorf("days:1 deadlines = %d, want %d (overdue kept, far-future dropped)", near.Total, thingstest.Deadlines-1)
	}
	// days:0 is the same as omitting the window.
	if wide := todosWithDays(t, srv, "deadlines", 0, 100); wide.Total != thingstest.Deadlines {
		t.Errorf("days:0 deadlines = %d, want all %d", wide.Total, thingstest.Deadlines)
	}
	// A huge window includes the far-future deadline again.
	huge := todosWithDays(t, srv, "deadlines", 100000, 100)
	if !slices.ContainsFunc(huge.Items, func(it Item) bool { return it.UUID == farFuture }) {
		t.Errorf("days:100000 must include the far-future deadline %s", farFuture)
	}
}

func TestUpcomingWindow(t *testing.T) {
	srv := newTestServer(t, Config{})
	// Upcoming holds only future-scheduled todos and the fixture dates are
	// absolute, so the window bounds are derived from a target's actual start date
	// read at runtime; hardcoding a days offset would rot as the wall clock moves.
	all := listTodos(t, srv, ListTodosInput{View: "upcoming", Limit: 100})
	target, daysOut := furthestUpcoming(t, all.Items)

	contains := func(p PageResult[Item]) bool {
		return slices.ContainsFunc(p.Items, func(it Item) bool { return it.UUID == target })
	}
	// A window ending before the target's date excludes it; one ending past it
	// includes it.
	near := todosWithDays(t, srv, "upcoming", daysOut/2, 100)
	if contains(near) {
		t.Errorf("a %d-day window must exclude the %d-day-out todo %s", daysOut/2, daysOut, target)
	}
	wide := todosWithDays(t, srv, "upcoming", daysOut+30, 100)
	if !contains(wide) {
		t.Errorf("a %d-day window must include the %d-day-out todo %s", daysOut+30, daysOut, target)
	}
	// Total is monotonic in the window width; never assert exact counts, since
	// repeating next-occurrences float with now.
	if near.Total > wide.Total {
		t.Errorf("window totals not monotonic: near=%d wide=%d", near.Total, wide.Total)
	}
}

// furthestUpcoming returns the UUID and whole days from today of the upcoming
// item scheduled furthest out, giving the window math maximal headroom so the
// test stays clock-independent.
func furthestUpcoming(t *testing.T, items []Item) (uuid string, daysOut int) {
	t.Helper()
	today := time.Now()
	best := -1
	for i := range items {
		if items[i].When == "" {
			continue
		}
		d, err := time.Parse("2006-01-02", items[i].When)
		if err != nil {
			continue
		}
		if out := int(d.Sub(today).Hours() / 24); out > best {
			best, uuid = out, items[i].UUID
		}
	}
	if best < 2 {
		t.Skipf("no upcoming todo far enough out to window (furthest = %d days)", best)
	}
	return uuid, best
}

func TestLogbookDefaultWindow(t *testing.T) {
	path := thingstest.DatabasePath(t)
	srv := serverOnPath(t, path, Config{})

	full := todosWithDays(t, srv, "logbook", 0, 100)
	if full.Total == 0 {
		t.Fatal("fixture has no logbook rows")
	}
	// The fixture's stop dates are all years old, so the default 30-day window is
	// empty until a row is bumped into it.
	if base := listTodos(t, srv, ListTodosInput{View: "logbook"}); base.Total != 0 {
		t.Fatalf("default 30-day window on the aged fixture = %d, want 0", base.Total)
	}

	// stopDate is a float Unix epoch; bump one row to now so the default window sees it.
	target := full.Items[0].UUID
	execFixture(t, path, "UPDATE TMTask SET stopDate = ? WHERE uuid = ?", time.Now().Unix(), target)

	windowed := listTodos(t, srv, ListTodosInput{View: "logbook"})
	if windowed.Total != 1 || windowed.Items[0].UUID != target {
		t.Fatalf("default-window logbook after bump = %+v, want just %s", windowed, target)
	}
	if again := todosWithDays(t, srv, "logbook", 0, 100); again.Total != full.Total {
		t.Errorf("all-history total = %d, want unchanged %d", again.Total, full.Total)
	}
}
