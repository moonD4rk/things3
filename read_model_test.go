package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moond4rk/things3/thingstest"
)

// Fixture rows behind the read-model tests.
const (
	testUUIDHeading          = "6QpDLSHZMRAUSAeZ9mNvgt" // in testUUIDProjectInArea1
	testUUIDTodoTrashed      = "Tmt3omxtu61wAoTaSLweQy"
	testUUIDTemplate         = "N1PJHsbjct4mb1bhcs7aHa" // next occurrence 2040-01-01, sentinel deadline
	testUUIDProjectTrashed   = "Tc7DABDNNMZvV4ZGB8tLDh"
	testUUIDTodoWithReminder = "7F4vqUNiTvGKaCUfv5pqYG" // scheduled 2026-09-17
)

func newFixtureClientAt(t *testing.T, path string, opts ...ClientOption) *Client {
	t.Helper()
	client, err := NewClient(append([]ClientOption{WithDatabasePath(path), WithClock(thingstest.Now)}, opts...)...)
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	return client
}

// A todo under a heading has neither a project nor an area column, and a todo
// in a project has no area column. Both must still say where they live.
func TestTodoReportsEffectiveProjectAndArea(t *testing.T) {
	client := newFixtureClientAt(t, thingstest.DatabasePath(t))
	ctx := t.Context()

	underHeading, err := client.Todos().WithUUID(testUUIDTodoInHeading).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, testUUIDHeading, underHeading.HeadingUUID)
	assert.Equal(t, testUUIDProjectInArea1, underHeading.ProjectUUID)
	assert.Equal(t, "Project in Area 1", underHeading.ProjectTitle)
	assert.Equal(t, testUUIDArea1, underHeading.AreaUUID)
	assert.Equal(t, "Area 1", underHeading.AreaTitle)

	inProject, err := client.Todos().WithUUID(testUUIDTodoInArea1Tags).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, testUUIDProjectInArea1, inProject.ProjectUUID)
	assert.Equal(t, testUUIDArea1, inProject.AreaUUID, "inherited from the project")
}

func TestInAreaReachesProjectChildren(t *testing.T) {
	client := newFixtureClientAt(t, thingstest.DatabasePath(t))
	ctx := t.Context()

	todos, err := client.Todos().InArea(testUUIDArea1).Status().Incomplete().All(ctx)
	require.NoError(t, err)
	uuids := extractTodoUUIDs(todos)

	assert.Contains(t, uuids, testUUIDTodoInArea1, "loose in the area")
	assert.Contains(t, uuids, testUUIDTodoInArea1Tags, "in a project of the area")
	assert.Contains(t, uuids, testUUIDTodoInHeading, "under a heading in that project")

	inArea, err := client.Todos().HasArea(true).Status().Incomplete().Count(ctx)
	require.NoError(t, err)
	outside, err := client.Todos().HasArea(false).Status().Incomplete().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testTodosIncomplete, inArea+outside, "HasArea partitions every incomplete todo")
	assert.GreaterOrEqual(t, inArea, len(todos))
}

// The search text is matched against an item's own title and notes. It used to
// match the area title too, which with inherited areas would return every todo
// of an area whose name merely contains the query.
func TestSearchMatchesOwnTextOnly(t *testing.T) {
	path := thingstest.DatabasePath(t)
	execFixtureSQL(t, path, "UPDATE TMArea SET title = 'Zebra' WHERE uuid = ?", testUUIDArea3)
	client := newFixtureClientAt(t, path)
	ctx := t.Context()

	inArea, err := client.Todos().InArea(testUUIDArea3).All(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, inArea)
	assert.Equal(t, "Zebra", inArea[0].AreaTitle)

	hits, err := client.Todos().Search("Zebra").Status().Any().All(ctx)
	require.NoError(t, err)
	assert.Empty(t, hits)
}

func TestIncludeTrashed(t *testing.T) {
	client := newFixtureClientAt(t, thingstest.DatabasePath(t))
	ctx := t.Context()

	_, err := client.Todos().WithUUID(testUUIDTodoTrashed).First(ctx)
	require.ErrorIs(t, err, ErrTodoNotFound, "the default scope is live items only")

	todo, err := client.Todos().WithUUID(testUUIDTodoTrashed).IncludeTrashed().First(ctx)
	require.NoError(t, err)
	assert.True(t, todo.Trashed)

	inTrashedProject, err := client.Todos().WithUUID(testUUIDTodoInDeletedProject).IncludeTrashed().First(ctx)
	require.NoError(t, err)
	assert.False(t, inTrashedProject.Trashed, "only its project is trashed")

	project, err := client.Projects().WithUUID(testUUIDProjectTrashed).IncludeTrashed().First(ctx)
	require.NoError(t, err)
	assert.True(t, project.Trashed)

	live, err := client.Todos().Status().Any().Count(ctx)
	require.NoError(t, err)
	all, err := client.Todos().Status().Any().IncludeTrashed().Count(ctx)
	require.NoError(t, err)
	onlyTrashed, err := client.Todos().Status().Any().Trashed(true).Count(ctx)
	require.NoError(t, err)
	assert.Greater(t, all, live+onlyTrashed-1, "any = live + trashed + those inside trashed projects")
	assert.Equal(t, thingstest.TrashedProjectTodos, all-live-onlyTrashed)
}

// A repeating template carries its next occurrence as StartDate. Its deadline
// column holds the 4001-01-01 sentinel, which once decoded as a day in 1953.
func TestIncludeTemplates(t *testing.T) {
	client := newFixtureClientAt(t, thingstest.DatabasePath(t))
	ctx := t.Context()

	_, err := client.Todos().WithUUID(testUUIDTemplate).First(ctx)
	require.ErrorIs(t, err, ErrTodoNotFound, "templates stay hidden by default")

	template, err := client.Todos().WithUUID(testUUIDTemplate).IncludeTemplates().First(ctx)
	require.NoError(t, err)
	assert.True(t, template.Template)
	assert.True(t, template.Repeating)
	require.NotNil(t, template.StartDate)
	assert.Equal(t, "2040-01-01", template.StartDate.Format(time.DateOnly))
	assert.Nil(t, template.Deadline)

	instance, err := client.Todos().WithUUID(testUUIDTodoRepeating).First(ctx)
	require.NoError(t, err)
	assert.True(t, instance.Repeating)
	assert.False(t, instance.Template)
}

func TestWithClock(t *testing.T) {
	path := thingstest.DatabasePath(t)
	scheduled := time.Date(2026, time.September, 17, 0, 0, 0, 0, time.Local)

	upcomingAt := func(now time.Time) []string {
		client := newFixtureClientAt(t, path, WithClock(func() time.Time { return now }))
		assert.Equal(t, now, client.Now())
		items, err := client.Upcoming(t.Context())
		require.NoError(t, err)
		return itemUUIDs(items)
	}

	assert.Contains(t, upcomingAt(scheduled.AddDate(0, 0, -1)), testUUIDTodoWithReminder)
	assert.NotContains(t, upcomingAt(scheduled), testUUIDTodoWithReminder, "a todo scheduled for today has arrived")

	systemClock, err := NewClient(WithDatabasePath(path))
	require.NoError(t, err)
	defer systemClock.Close()
	assert.WithinDuration(t, time.Now(), systemClock.Now(), time.Minute)
}

func TestTimestampsKeepSubSecondPrecision(t *testing.T) {
	path := thingstest.DatabasePath(t)
	execFixtureSQL(t, path, "UPDATE TMTask SET userModificationDate = 1700000000.75 WHERE uuid = ?", testUUIDTodoInToday)
	client := newFixtureClientAt(t, path)

	todo, err := client.Todos().WithUUID(testUUIDTodoInToday).First(t.Context())
	require.NoError(t, err)

	// A verification that compares ModifiedAt to a baseline must see two writes
	// landing within the same second as different instants.
	assert.True(t, todo.ModifiedAt.After(time.Unix(1700000000, 0)))
	assert.Equal(t, int64(1700000000750), todo.ModifiedAt.UnixMilli())
}

func TestSearchBuilder(t *testing.T) {
	client := newFixtureClientAt(t, thingstest.DatabasePath(t))

	uri, err := client.SearchBuilder("buy milk").Build()
	require.NoError(t, err)
	assert.Equal(t, "things:///search?query=buy%20milk", uri)

	bare, err := client.SearchBuilder("").Build()
	require.NoError(t, err)
	assert.Equal(t, "things:///search", bare)
}
