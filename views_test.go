package things3

import (
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moond4rk/things3/thingstest"
)

// Fixture rows behind the view tests. Every expectation holds at thingstest.Now.
const (
	testUUIDTodoArrived       = "6Hf2qWBjWhq7B1xszwdo34" // Someday, scheduled 2021-05-04: its day has come
	testUUIDTodoScheduled     = "7F4vqUNiTvGKaCUfv5pqYG" // Someday, scheduled 2026-09-17
	testUUIDTodoSomeday       = "JLYSEPFkLfBC5rhGJRa5S1" // Someday, unscheduled, on its own
	testUUIDProjectInToday    = "PgsWnDkzXRz6zvofTqtHqn"
	testUUIDProjectNoArea     = "TCozQqXVbB2TJkXXXQj2H9" // holds testUUIDTodoInProject
	testUUIDProjectSomeday    = "SmdyProjTestFixture001"
	testUUIDProjectCanceled   = "SkLdfSe1MXR5vMV1gMYkHE"
	testUUIDProjectCompleted  = "CmpltdProjTestFixture01"
	testUUIDTodoInCanceledPrj = "S8QU6gEvQec7XRMkN5Vjwg" // completed, inside testUUIDProjectCanceled
	testUUIDTodoDoneInProject = "5u2yGhP4rMQUmPQYEpGYDd" // completed, inside the open testUUIDProjectNoArea
	testUUIDTodoTrashedInside = "LeMGbV4SZyMgSRTCWq72Yg" // trashed, inside the trashed testUUIDProjectTrashed
)

func itemUUIDs(items []Item) []string {
	uuids := make([]string, len(items))
	for i := range items {
		uuids[i] = items[i].UUID()
	}
	return uuids
}

func findItem(t *testing.T, items []Item, uuid string) Item {
	t.Helper()
	i := slices.IndexFunc(items, func(it Item) bool { return it.UUID() == uuid })
	require.GreaterOrEqualf(t, i, 0, "item %s not in view", uuid)
	return items[i]
}

// mutatedClient opens a client over a fixture copy with the statements applied.
func mutatedClient(t *testing.T, statements ...string) *Client {
	t.Helper()
	path := thingstest.DatabasePath(t)
	for _, stmt := range statements {
		execFixtureSQL(t, path, stmt)
	}
	return newFixtureClientAt(t, path)
}

func TestInbox(t *testing.T) {
	items, err := newTestClient(t).Inbox(t.Context())
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{testUUIDTodoInbox, testUUIDTodoInboxChecklist}, itemUUIDs(items))
}

// Today mixes todos and projects: whatever is scheduled into Today, whatever was
// scheduled for a day that has arrived, and whatever has a due deadline.
func TestToday(t *testing.T) {
	items, err := newTestClient(t).Today(t.Context())
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{
		testUUIDTodoInToday,        // scheduled into Today
		testUUIDTodoRepeating,      // a repeating instance scheduled into Today
		testUUIDProjectInToday,     // a project scheduled into Today
		testUUIDTodoArrived,        // its scheduled day has come
		testUUIDTodoOverdueInToday, // unscheduled, deadline due
	}, itemUUIDs(items))

	assert.NotContains(t, itemUUIDs(items), testUUIDTodoOverdueNotToday, "a deadline dismissed from Today stays out")
	assert.NotContains(t, itemUUIDs(items), testUUIDTodoScheduled, "its day has not come yet")
	assert.True(t, findItem(t, items, testUUIDProjectInToday).IsProject())
	for _, it := range items {
		assert.Equalf(t, StatusIncomplete, it.Status(), "%s", it.UUID())
	}
}

func TestTodayFollowsTheClock(t *testing.T) {
	arrival := time.Date(2026, time.September, 17, 8, 0, 0, 0, time.Local)
	client := newFixtureClientAt(t, thingstest.DatabasePath(t), WithClock(func() time.Time { return arrival }))

	items, err := client.Today(t.Context())
	require.NoError(t, err)
	assert.Contains(t, itemUUIDs(items), testUUIDTodoScheduled, "scheduled for today: the app would move it in")
}

func TestTodayPutsThisEveningLast(t *testing.T) {
	// startBucket 1 marks This Evening.
	client := mutatedClient(t, "UPDATE TMTask SET startBucket = 1 WHERE uuid = '"+testUUIDTodoInToday+"'")

	items, err := client.Today(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, items)

	last := items[len(items)-1]
	assert.Equal(t, testUUIDTodoInToday, last.UUID())
	assert.True(t, last.Todo.Evening)
	for _, it := range items[:len(items)-1] {
		assert.Falsef(t, it.Todo != nil && it.Todo.Evening, "%s is not an evening todo", it.UUID())
	}
}

// Upcoming files each entry under a day: a scheduled item under its start date,
// a repeating series under its next occurrence, an unscheduled item under its
// deadline.
func TestUpcoming(t *testing.T) {
	items, err := newTestClient(t).Upcoming(t.Context())
	require.NoError(t, err)

	require.Equal(t, []string{testUUIDTodoScheduled, testUUIDTemplate, testUUIDTodoInHeading}, itemUUIDs(items))

	days := make([]string, len(items))
	for i, it := range items {
		require.NotNil(t, it.UpcomingDate())
		days[i] = it.UpcomingDate().Format(time.DateOnly)
	}
	assert.Equal(t, []string{"2026-09-17", "2040-01-01", "2040-11-04"}, days)

	series := findItem(t, items, testUUIDTemplate)
	assert.True(t, series.Todo.Template)
	assert.Nil(t, series.Todo.Deadline, "the series carries a sentinel, not a deadline")

	due := findItem(t, items, testUUIDTodoInHeading)
	assert.Nil(t, due.StartDate(), "filed under its deadline because it has no start date")
}

func TestUpcomingIncludesProjectsAndDormantSeries(t *testing.T) {
	client := mutatedClient(t,
		// A project scheduled for a future day.
		"UPDATE TMTask SET startDate = 132911232 WHERE uuid = '"+testUUIDProjectSomeday+"'", // 2028-01-01
		// A series with no next occurrence: paused, or past its end date.
		"UPDATE TMTask SET rt1_nextInstanceStartDate = NULL WHERE uuid = '"+testUUIDTemplate+"'",
	)

	items, err := client.Upcoming(t.Context())
	require.NoError(t, err)

	assert.Equal(t, []string{testUUIDTodoScheduled, testUUIDProjectSomeday, testUUIDTodoInHeading, testUUIDTemplate}, itemUUIDs(items))
	assert.True(t, findItem(t, items, testUUIDProjectSomeday).IsProject())
	assert.Nil(t, items[len(items)-1].UpcomingDate(), "a dormant series trails the list with no day")
}

func TestAnytime(t *testing.T) {
	items, err := newTestClient(t).Anytime(t.Context())
	require.NoError(t, err)

	assert.ElementsMatch(t, testTodosAnytimeUUIDs, itemUUIDs(items))
	for _, it := range items {
		assert.Falsef(t, it.IsProject(), "%s: Anytime lists todos, grouped under their project", it.UUID())
	}
}

// Shelving a project in Someday takes its todos out of Anytime, although each of
// them still sits in the Anytime bucket itself.
func TestAnytimeHidesTodosOfSomedayProjects(t *testing.T) {
	client := mutatedClient(t,
		"UPDATE TMTask SET start = 2 WHERE uuid IN ('"+testUUIDProjectNoArea+"', '"+testUUIDProjectInArea1+"')")

	items, err := client.Anytime(t.Context())
	require.NoError(t, err)
	uuids := itemUUIDs(items)

	assert.NotContains(t, uuids, testUUIDTodoInProject, "directly in a Someday project")
	assert.NotContains(t, uuids, testUUIDTodoInHeading, "under a heading of a Someday project")
	assert.Contains(t, uuids, testUUIDTodoAnytime, "a todo on its own is unaffected")
	assert.Contains(t, uuids, testUUIDTodoInArea1, "so is a todo loose in an area")
}

// The sidebar order: by area, then project, then heading, then position, with
// whatever has no container leading.
func TestAnytimeFollowsSidebarOrder(t *testing.T) {
	items, err := newTestClient(t).Anytime(t.Context())
	require.NoError(t, err)

	var containers []string
	for _, it := range items {
		c := it.Todo.AreaTitle + "/" + it.Todo.ProjectTitle
		if len(containers) == 0 || containers[len(containers)-1] != c {
			containers = append(containers, c)
		}
	}
	unique := slices.Compact(slices.Sorted(slices.Values(containers)))
	assert.Len(t, containers, len(unique), "todos of one container must be contiguous: %v", containers)
	assert.Equal(t, "/", containers[0], "todos outside any area or project lead the list")
}

func TestSomeday(t *testing.T) {
	items, err := newTestClient(t).Someday(t.Context())
	require.NoError(t, err)

	assert.Equal(t, []string{testUUIDTodoSomeday, testUUIDProjectSomeday}, itemUUIDs(items))
	assert.True(t, items[1].IsProject())
}

// A Someday todo inside a project lives in that project, not in the Someday list.
func TestSomedayOmitsTodosInsideProjects(t *testing.T) {
	client := mutatedClient(t,
		"UPDATE TMTask SET project = '"+testUUIDProjectNoArea+"' WHERE uuid = '"+testUUIDTodoSomeday+"'")

	items, err := client.Someday(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []string{testUUIDProjectSomeday}, itemUUIDs(items))
}

// The Logbook lists a closed project as one row and folds its todos into it.
func TestLogbook(t *testing.T) {
	items, err := newTestClient(t).Logbook(t.Context(), time.Time{})
	require.NoError(t, err)
	uuids := itemUUIDs(items)

	assert.Contains(t, uuids, testUUIDProjectCompleted)
	assert.Contains(t, uuids, testUUIDProjectCanceled)
	assert.NotContains(t, uuids, testUUIDTodoInCanceledPrj, "folded into its canceled project")
	assert.Contains(t, uuids, testUUIDTodoDoneInProject, "completed inside a project that is still open")

	for _, it := range items {
		assert.Truef(t, it.Status().IsClosed(), "%s", it.UUID())
		assert.NotNilf(t, it.StoppedAt(), "%s", it.UUID())
	}
	assert.True(t, slices.IsSortedFunc(items, func(a, b Item) int {
		return b.StoppedAt().Compare(*a.StoppedAt())
	}), "most recently closed first")
}

func TestLogbookSince(t *testing.T) {
	client := newTestClient(t)
	ctx := t.Context()

	all, err := client.Logbook(ctx, time.Time{})
	require.NoError(t, err)
	require.NotEmpty(t, all)
	newest := *all[0].StoppedAt()

	recent, err := client.Logbook(ctx, newest.AddDate(0, 0, -1))
	require.NoError(t, err)
	assert.NotEmpty(t, recent)
	assert.Less(t, len(recent), len(all))
	for _, it := range recent {
		assert.Equal(t, newest.Format(time.DateOnly), it.StoppedAt().Format(time.DateOnly))
	}

	none, err := client.Logbook(ctx, newest)
	require.NoError(t, err)
	assert.Empty(t, none, "since is exclusive of its own day")
}

func TestDeadlines(t *testing.T) {
	client := mutatedClient(t,
		"UPDATE TMTask SET deadline = 132911232 WHERE uuid = '"+testUUIDProjectInToday+"'") // 2028-01-01

	items, err := client.Deadlines(t.Context())
	require.NoError(t, err)

	assert.Equal(t, []string{
		testUUIDTodoRepeating,       // 2021-03-28
		testUUIDTodoOverdueInToday,  // 2021-05-21
		testUUIDTodoOverdueNotToday, // 2021-05-21
		testUUIDProjectInToday,      // 2028-01-01
		testUUIDTodoInHeading,       // 2040-11-04
	}, itemUUIDs(items))
	assert.NotContains(t, itemUUIDs(items), testUUIDTemplate, "a sentinel deadline is not a deadline")
}

// The Trash lists a trashed project as one row and folds its trashed todos in.
func TestTrash(t *testing.T) {
	items, err := newTestClient(t).Trash(t.Context())
	require.NoError(t, err)
	uuids := itemUUIDs(items)

	assert.Contains(t, uuids, testUUIDProjectTrashed)
	assert.Contains(t, uuids, testUUIDTodoTrashed)
	assert.NotContains(t, uuids, testUUIDTodoTrashedInside, "folded into its trashed project")
	assert.NotContains(t, uuids, testUUIDTodoInDeletedProject, "not trashed itself")
	assert.Len(t, items, thingstest.TrashedTodos+thingstest.TrashedCanceled+thingstest.TrashedCompleted-1+thingstest.TrashedProjects)
	assert.True(t, slices.IsSortedFunc(items, func(a, b Item) int {
		return b.ModifiedAt().Compare(a.ModifiedAt())
	}), "most recently trashed first")
}

func TestViewsAreNeverNil(t *testing.T) {
	client := mutatedClient(t, "DELETE FROM TMTask")
	ctx := t.Context()

	for name, view := range map[string]func() ([]Item, error){
		"inbox":     func() ([]Item, error) { return client.Inbox(ctx) },
		"today":     func() ([]Item, error) { return client.Today(ctx) },
		"upcoming":  func() ([]Item, error) { return client.Upcoming(ctx) },
		"anytime":   func() ([]Item, error) { return client.Anytime(ctx) },
		"someday":   func() ([]Item, error) { return client.Someday(ctx) },
		"logbook":   func() ([]Item, error) { return client.Logbook(ctx, time.Time{}) },
		"deadlines": func() ([]Item, error) { return client.Deadlines(ctx) },
		"trash":     func() ([]Item, error) { return client.Trash(ctx) },
	} {
		items, err := view()
		require.NoErrorf(t, err, "%s", name)
		assert.NotNilf(t, items, "%s must encode as [] rather than null", name)
		assert.Emptyf(t, items, "%s", name)
	}
}

func TestItemJSONCarriesTypeFirst(t *testing.T) {
	todo, err := json.Marshal(Item{Todo: &Todo{UUID: "t1", Title: "Buy milk"}})
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"todo","uuid":"t1","title":"Buy milk","status":"incomplete","start":"inbox",`+
		`"created_at":"0001-01-01T00:00:00Z","modified_at":"0001-01-01T00:00:00Z"}`, string(todo))
	assert.Regexp(t, `^\{"type":"todo","uuid":"t1"`, string(todo), "type leads, then the model's own field order")

	project, err := json.Marshal(Item{Project: &Project{UUID: "p1"}})
	require.NoError(t, err)
	assert.Regexp(t, `^\{"type":"project","uuid":"p1"`, string(project))
}

func TestItemAccessors(t *testing.T) {
	start := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.Local)
	deadline := start.AddDate(0, 0, 7)
	done := start.AddDate(0, 0, 3)

	scheduled := Item{Todo: &Todo{StartDate: &start, Deadline: &deadline}}
	assert.Equal(t, &start, scheduled.UpcomingDate(), "the start date wins")

	dueOnly := Item{Project: &Project{Deadline: &deadline, CanceledAt: &done, Status: StatusCanceled}}
	assert.Equal(t, &deadline, dueOnly.UpcomingDate())
	assert.Equal(t, &done, dueOnly.StoppedAt())
	assert.Equal(t, ItemTypeProject, dueOnly.Type())

	assert.Nil(t, Item{Todo: &Todo{}}.UpcomingDate())
	assert.Nil(t, Item{Todo: &Todo{}}.StoppedAt())
}
