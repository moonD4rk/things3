package verify

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

var pastT0 = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func newClient(t *testing.T, dbPath string) *things3.Client {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(dbPath), things3.WithClock(thingstest.Now))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func instantSleep(context.Context, time.Duration) error { return nil }

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

// loadTodo reads a todo the way a write target is resolved: checklist included.
func loadTodo(t *testing.T, c *things3.Client, uuid string) *things3.Todo {
	t.Helper()
	todo, err := c.Todos().WithUUID(uuid).Status().Any().IncludeChecklist().First(context.Background())
	if err != nil {
		t.Fatalf("todo %s: %v", uuid, err)
	}
	return todo
}

func loadProject(t *testing.T, c *things3.Client, uuid string) *things3.Project {
	t.Helper()
	project, err := c.Projects().WithUUID(uuid).Status().Any().First(context.Background())
	if err != nil {
		t.Fatalf("project %s: %v", uuid, err)
	}
	return project
}

func countingSleep(maxCalls int, count *int) func(context.Context, time.Duration) error {
	return func(context.Context, time.Duration) error {
		*count++
		if *count >= maxCalls {
			return context.DeadlineExceeded
		}
		return nil
	}
}

func TestPollImmediate(t *testing.T) {
	count := 0
	confirmed, err := Poll(context.Background(), Options{Sleep: countingSleep(100, &count)},
		func(context.Context) (bool, error) { return true, nil })
	if !confirmed || err != nil {
		t.Fatalf("want confirmed, got confirmed=%v err=%v", confirmed, err)
	}
	if count != 0 {
		t.Errorf("check true on first call should sleep 0 times, got %d", count)
	}
}

func TestPollBudget(t *testing.T) {
	count := 0
	confirmed, _ := Poll(context.Background(), Options{Sleep: countingSleep(5, &count)},
		func(context.Context) (bool, error) { return false, nil })
	if confirmed {
		t.Error("should not confirm when check is never true")
	}
	if count != 5 {
		t.Errorf("expected 5 sleeps before giving up, got %d", count)
	}
}

func TestPollRetainsLastError(t *testing.T) {
	count := 0
	boom := errors.New("boom")
	_, err := Poll(context.Background(), Options{Sleep: countingSleep(3, &count)},
		func(context.Context) (bool, error) { return false, boom })
	if !errors.Is(err, boom) {
		t.Errorf("Poll should return the last check error, got %v", err)
	}
}

func TestPollContextCancelAborts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	count := 0
	sleep := func(context.Context, time.Duration) error {
		count++
		cancel()
		return context.Canceled
	}
	confirmed, _ := Poll(ctx, Options{Sleep: sleep}, func(context.Context) (bool, error) { return false, nil })
	if confirmed {
		t.Error("context cancel should abort the poll")
	}
	if count != 1 {
		t.Errorf("expected abort after 1 sleep, got %d", count)
	}
}

func TestTodoStatusConfirmed(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	todo, outcome, err := TodoStatus(context.Background(), c, thingstest.UUIDTodoInToday,
		things3.StatusIncomplete, Options{Sleep: instantSleep})
	if err != nil || outcome != Confirmed || todo == nil {
		t.Fatalf("want Confirmed, got outcome=%v err=%v todo=%v", outcome, err, todo)
	}
}

func TestAddedTodoConfirmed(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	todo, outcome, err := AddedTodo(context.Background(), c, "To-Do in Today", pastT0, Options{Sleep: instantSleep})
	if err != nil || outcome != Confirmed {
		t.Fatalf("want Confirmed, got outcome=%v err=%v", outcome, err)
	}
	if todo == nil || todo.UUID != thingstest.UUIDTodoInToday {
		t.Errorf("want uuid %s, got %+v", thingstest.UUIDTodoInToday, todo)
	}
}

func TestAddedTodoUnverified(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	count := 0
	_, outcome, _ := AddedTodo(context.Background(), c, "zzz no such title zzz", pastT0,
		Options{Sleep: countingSleep(3, &count)})
	if outcome != Unverified {
		t.Errorf("nonexistent title should be Unverified, got %v", outcome)
	}
}

func TestAddedTodoAmbiguous(t *testing.T) {
	dbPath := thingstest.DatabasePath(t)
	execFixture(t, dbPath, "UPDATE TMTask SET title = 'To-Do in Today' WHERE uuid = ?", thingstest.UUIDTodoChecklist)

	c := newClient(t, dbPath)
	_, outcome, _ := AddedTodo(context.Background(), c, "To-Do in Today", pastT0, Options{Sleep: instantSleep})
	if outcome != AmbiguousMatch {
		t.Errorf("two same-title rows should be AmbiguousMatch, got %v", outcome)
	}
}

func TestTodoStatusUnverified(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	count := 0
	_, outcome, _ := TodoStatus(context.Background(), c, thingstest.UUIDTodoInToday,
		things3.StatusCompleted, Options{Sleep: countingSleep(3, &count)})
	if outcome != Unverified {
		t.Errorf("incomplete todo never reaches Completed -> Unverified, got %v", outcome)
	}
}

func TestProjectStatusConfirmed(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	project, outcome, err := ProjectStatus(context.Background(), c, thingstest.UUIDProject,
		things3.StatusIncomplete, Options{Sleep: instantSleep})
	if err != nil || outcome != Confirmed || project == nil {
		t.Fatalf("want Confirmed, got outcome=%v err=%v project=%v", outcome, err, project)
	}
}

func TestTodoModifiedConfirmed(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	before := loadTodo(t, c, thingstest.UUIDTodoInToday)
	before.ModifiedAt = pastT0
	todo, outcome, err := TodoModified(context.Background(), c, before, Options{Sleep: instantSleep})
	if err != nil || outcome != Confirmed || todo == nil {
		t.Fatalf("want Confirmed, got outcome=%v err=%v", outcome, err)
	}
}

func TestTodoModifiedUnverified(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	count := 0
	_, outcome, _ := TodoModified(context.Background(), c, loadTodo(t, c, thingstest.UUIDTodoChecklist),
		Options{Sleep: countingSleep(3, &count)})
	if outcome != Unverified {
		t.Errorf("a todo that did not change -> Unverified, got %v", outcome)
	}
}

// Things leaves a todo's modification date alone when only its checklist
// changes, so the checklist itself has to confirm the write.
func TestTodoModifiedConfirmsAChecklistOnlyChange(t *testing.T) {
	for name, mutation := range map[string]string{
		"item added": `INSERT INTO TMChecklistItem (uuid, title, status, "index", task, creationDate, userModificationDate)
			VALUES ('AddedChecklistItem0001', 'Added', 0, 1, ?, 1, 1)`,
		"item removed": `DELETE FROM TMChecklistItem WHERE uuid = (SELECT uuid FROM TMChecklistItem WHERE task = ? LIMIT 1)`,
		"list cleared": `DELETE FROM TMChecklistItem WHERE task = ?`,
	} {
		t.Run(name, func(t *testing.T) {
			dbPath := thingstest.DatabasePath(t)
			c := newClient(t, dbPath)
			before := loadTodo(t, c, thingstest.UUIDTodoChecklist)
			execFixture(t, dbPath, mutation, thingstest.UUIDTodoChecklist)

			todo, outcome, err := TodoModified(context.Background(), c, before, Options{Sleep: instantSleep})
			if err != nil || outcome != Confirmed || todo == nil {
				t.Fatalf("want Confirmed, got outcome=%v err=%v", outcome, err)
			}
			if !todo.ModifiedAt.Equal(before.ModifiedAt) {
				t.Fatalf("the fixture mutation must leave the modification date alone")
			}
			if len(todo.Checklist) == len(before.Checklist) {
				t.Errorf("the confirmed todo should carry the new checklist, got %d items", len(todo.Checklist))
			}
		})
	}
}

func TestProjectModifiedConfirmed(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	before := loadProject(t, c, thingstest.UUIDProject)
	before.ModifiedAt = pastT0
	project, outcome, err := ProjectModified(context.Background(), c, before, Options{Sleep: instantSleep})
	if err != nil || outcome != Confirmed || project == nil {
		t.Fatalf("want Confirmed, got outcome=%v err=%v", outcome, err)
	}
}

func TestProjectModifiedUnverified(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	count := 0
	_, outcome, _ := ProjectModified(context.Background(), c, loadProject(t, c, thingstest.UUIDProject),
		Options{Sleep: countingSleep(3, &count)})
	if outcome != Unverified {
		t.Errorf("a project that did not change -> Unverified, got %v", outcome)
	}
}

func TestAddedProjectConfirmed(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	project, outcome, err := AddedProject(context.Background(), c, "Project in Area 1", pastT0, Options{Sleep: instantSleep})
	if err != nil || outcome != Confirmed {
		t.Fatalf("want Confirmed, got outcome=%v err=%v", outcome, err)
	}
	if project == nil || project.UUID != thingstest.UUIDProject {
		t.Errorf("want uuid %s, got %+v", thingstest.UUIDProject, project)
	}
}

func TestAddedProjectUnverified(t *testing.T) {
	c := newClient(t, thingstest.DatabasePath(t))
	count := 0
	_, outcome, _ := AddedProject(context.Background(), c, "zzz no such project zzz", pastT0,
		Options{Sleep: countingSleep(3, &count)})
	if outcome != Unverified {
		t.Errorf("nonexistent project title -> Unverified, got %v", outcome)
	}
}
