package mcpserver

import (
	"context"
	"slices"
	"testing"

	"github.com/moond4rk/things3/thingstest"
)

func listTodos(t *testing.T, srv *Server, in ListTodosInput) PageResult[Item] {
	t.Helper()
	_, page, err := srv.handleListTodos(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("list_todos %+v: %v", in, err)
	}
	return page
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
	cases := []struct {
		view string
		want int
	}{
		{"inbox", thingstest.Inbox},
		{"deadlines", thingstest.Deadlines},
	}
	for _, tc := range cases {
		t.Run(tc.view, func(t *testing.T) {
			page := listTodos(t, srv, ListTodosInput{View: ViewName(tc.view)})
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
		if got := clampLimit(tc.in); got != tc.want {
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
	page := listTodos(t, srv, ListTodosInput{View: "logbook", Limit: 100})
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
	page := listTodos(t, srv, ListTodosInput{View: "anytime", Project: thingstest.UUIDProject})
	if !slices.ContainsFunc(page.Items, func(it Item) bool { return it.UUID == headingTodo }) {
		t.Errorf("project filter dropped heading-todo %s; got %d items", headingTodo, len(page.Items))
	}
}
