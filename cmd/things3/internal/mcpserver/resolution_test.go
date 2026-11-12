package mcpserver

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/verify"
	"github.com/moond4rk/things3/thingstest"
)

// Fixture rows that exercise the edges of target resolution.
const (
	uuidTrashedTodo      = "Tmt3omxtu61wAoTaSLweQy"
	uuidTemplate         = "N1PJHsbjct4mb1bhcs7aHa" // "Repeating To-Do", the series itself
	uuidCompletedInToday = "56dtXSk3A373M6n4eqGyr3"
	uuidTodoInHeading    = "HbKGAeZKFDkWH5osSBNHvz" // under a heading of "Project in Area 1"
	uuidTodoInAreaProj   = "W5JYfjY2xtLdmedQKU6caM" // directly in "Project in Area 1"
	uuidTodoInArea       = "Q7uN9y3jp5ChZAGjZJhMfY" // loose in "Area 1"
)

// recordingServerOnPath is newWriteServer over a specific (mutated) fixture copy.
func recordingServerOnPath(t *testing.T, path string) (*Server, *recorder) {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(path), things3.WithClock(thingstest.Now))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	rec := &recorder{}
	srv, err := New(client, Config{Execute: rec.exec, Verify: verify.Options{Sleep: oneShotSleep}})
	if err != nil {
		t.Fatalf("server: %v", err)
	}
	return srv, rec
}

// A UUID from the trash view or from upcoming's repeating series must resolve,
// and the answer must say what kind of item it is.
func TestGetReachesTrashedItemsAndRepeatingSeries(t *testing.T) {
	srv := newTestServer(t, Config{})

	trashed := get(t, srv, uuidTrashedTodo[:8])
	if !trashed.Success || trashed.Item == nil || trashed.Item.UUID != uuidTrashedTodo || !trashed.Item.Trashed {
		t.Errorf("get of a trashed todo = %+v", trashed)
	}

	series := get(t, srv, uuidTemplate)
	if !series.Success || series.Item == nil {
		t.Fatalf("get of a repeating series failed: %+v", series.Error)
	}
	if !series.Item.Template || !series.Item.Repeating {
		t.Errorf("the series should be flagged template and repeating: %+v", series.Item)
	}
	if series.Item.When != "2040-01-01" {
		t.Errorf("when should hold the next occurrence, got %q", series.Item.When)
	}
	if series.Item.Deadline != "" {
		t.Errorf("the sentinel deadline must not surface as a date, got %q", series.Item.Deadline)
	}
}

func TestWriteToolsRefuseTrashedItemsAndRepeatingSeries(t *testing.T) {
	srv, rec := newWriteServer(t)
	ctx := context.Background()

	for name, uuid := range map[string]string{"trashed": uuidTrashedTodo, "repeating series": uuidTemplate} {
		results := map[string]WriteResult{}
		_, results["complete"], _ = srv.handleComplete(ctx, nil, CompleteInput{Target: uuid, Status: statusCompleted})
		_, results["schedule"], _ = srv.handleSchedule(ctx, nil, ScheduleInput{Target: uuid, When: "today"})
		_, results["move"], _ = srv.handleMove(ctx, nil, MoveInput{Target: uuid, To: "Area 1"})
		_, results["edit"], _ = srv.handleEdit(ctx, nil, EditInput{Target: uuid, Title: "x"})

		for tool, res := range results {
			if res.Success || res.Error == nil || res.Error.Code != codeInvalidInput {
				t.Errorf("%s on a %s item should be invalid_input, got %+v", tool, name, res)
			}
		}
	}
	if n := rec.count(); n != 0 {
		t.Errorf("no URL may be executed for a read-only item, got %d", n)
	}

	// Revealing is not a change.
	res := openTool(t, srv, OpenInput{Target: uuidTrashedTodo})
	if !res.Success || rec.last(t).Query().Get("id") != uuidTrashedTodo {
		t.Errorf("open should still reveal a trashed item: %+v", res)
	}
}

// One open item among completed namesakes is what a repeating task looks like:
// a status change names the open one, a reopen names the closed one.
func TestTargetPreferenceFollowsTheVerb(t *testing.T) {
	path := thingstest.DatabasePath(t)
	execFixture(t, path, "UPDATE TMTask SET title = 'To-Do in Today' WHERE uuid = ?", uuidCompletedInToday)
	srv, rec := recordingServerOnPath(t, path)

	for status, want := range map[CompleteStatus]string{
		statusCompleted:  thingstest.UUIDTodoInToday,
		statusCanceled:   thingstest.UUIDTodoInToday,
		statusIncomplete: uuidCompletedInToday,
	} {
		res := complete(t, srv, CompleteInput{Target: "To-Do in Today", Status: status})
		if !res.Success {
			t.Fatalf("complete %s: %+v", status, res.Error)
		}
		if got := rec.last(t).Query().Get("id"); got != want {
			t.Errorf("status %s targeted %s, want %s", status, got, want)
		}
	}

	found := get(t, srv, "To-Do in Today")
	if !found.Success || found.Item.UUID != thingstest.UUIDTodoInToday {
		t.Errorf("get by a shared title should answer with the open item, got %+v", found)
	}
}

// Only loose todos carry an area column; an area filter that stopped there
// returned nothing for anyone who keeps their todos in projects.
func TestAreaFilterReachesProjectChildren(t *testing.T) {
	srv := newTestServer(t, Config{})

	page := listTodos(t, srv, ListTodosInput{View: "anytime", Area: "Area 1", Limit: 100})
	has := func(uuid string) bool {
		return slices.ContainsFunc(page.Items, func(it Item) bool { return it.UUID == uuid })
	}
	for name, uuid := range map[string]string{
		"loose in the area":               uuidTodoInArea,
		"in a project of the area":        uuidTodoInAreaProj,
		"under a heading in that project": uuidTodoInHeading,
	} {
		if !has(uuid) {
			t.Errorf("area filter is missing the todo %s (%s)", name, uuid)
		}
	}
	for _, it := range page.Items {
		if it.Area == nil || it.Area.Title != "Area 1" {
			t.Errorf("%s should report its effective area, got %+v", it.UUID, it.Area)
		}
	}
}

func TestHeadingChildCarriesItsProject(t *testing.T) {
	srv := newTestServer(t, Config{})

	res := get(t, srv, uuidTodoInHeading)
	if !res.Success || res.Item == nil {
		t.Fatalf("get: %+v", res.Error)
	}
	if res.Item.Project == nil || res.Item.Project.Title != "Project in Area 1" {
		t.Errorf("a todo under a heading must report the heading's project, got %+v", res.Item.Project)
	}
	if res.Item.Heading == nil || res.Item.Heading.Title != "Heading" {
		t.Errorf("heading = %+v", res.Item.Heading)
	}

	scoped := listTodos(t, srv, ListTodosInput{View: "anytime", Project: "Project in Area 1", Limit: 100})
	if !slices.ContainsFunc(scoped.Items, func(it Item) bool { return it.UUID == uuidTodoInHeading }) {
		t.Errorf("project filter should include the todo under its heading")
	}
}

// Scoping a view that lists projects: an area keeps its projects along with its
// todos, while a project filter keeps only what is inside that project.
func TestScopingAViewThatListsProjects(t *testing.T) {
	srv := newTestServer(t, Config{})
	all := 0

	inArea := listTodos(t, srv, ListTodosInput{View: "logbook", Days: &all, Area: "Area 1", Limit: 100})
	var projects, todos int
	for _, it := range inArea.Items {
		if it.Area == nil || it.Area.Title != "Area 1" {
			t.Errorf("%s %s is outside Area 1", it.Type, it.UUID)
		}
		if it.Type == typeProject {
			projects++
		} else {
			todos++
		}
	}
	if projects != 2 || todos == 0 {
		t.Errorf("Area 1 logbook = %d projects and %d todos, want its 2 closed projects and its closed todos", projects, todos)
	}

	inProject := listTodos(t, srv, ListTodosInput{View: "today", Project: "Project in Area 1", Limit: 100})
	if inProject.Total == 0 {
		t.Fatalf("fixture drift: no Today todo in the project")
	}
	for _, it := range inProject.Items {
		if it.Type != typeTodo || it.Project == nil || it.Project.Title != "Project in Area 1" {
			t.Errorf("project scope kept %s %s", it.Type, it.UUID)
		}
	}
}

// Upcoming files an unscheduled item under its deadline, so the days window
// must measure that day too.
func TestUpcomingWindowCoversDeadlineOnlyItems(t *testing.T) {
	srv := newTestServer(t, Config{})

	has := func(days int) bool {
		page := todosWithDays(t, srv, "upcoming", days, 100)
		return slices.ContainsFunc(page.Items, func(it Item) bool { return it.UUID == uuidTodoInHeading })
	}
	deadline := time.Date(2040, time.November, 4, 0, 0, 0, 0, time.Local)
	daysOut := int(deadline.Sub(thingstest.Now()).Hours()/24) + 1

	if has(daysOut - 30) {
		t.Errorf("a window ending before the deadline must not include the item")
	}
	if !has(daysOut + 30) {
		t.Errorf("a window reaching the deadline must include the item")
	}

	page := listTodos(t, srv, ListTodosInput{View: "upcoming", Limit: 100})
	i := slices.IndexFunc(page.Items, func(it Item) bool { return it.UUID == uuidTodoInHeading })
	if i < 0 || page.Items[i].When != "" || page.Items[i].Deadline != "2040-11-04" {
		t.Errorf("the item is filed by its deadline, not a start date: %+v", page.Items)
	}
}
