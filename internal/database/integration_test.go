package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Well-known rows and counts from the fixture database.
const (
	fixtureTodoInProject = "E18tg5qepzrQk9J6jQtb5C" // creationDate epoch 1616958920
	fixtureTodoInToday   = "5pUx6PESj3ctFYbgth1PXY" // startDate 2021-03-28, tag "Office"
	fixtureTodoInHeading = "HbKGAeZKFDkWH5osSBNHvz" // deadline 2040-11-04
	fixtureAreaWithTags  = "DciSFacytdrNG1nRaMJPgY" // tags "Errand" and "Important"

	fixtureHeading              = "6QpDLSHZMRAUSAeZ9mNvgt" // in fixtureProjectInArea
	fixtureProjectInArea        = "3x1QqJqfvZyhtw8NSdnZqG" // "Project in Area 1", in fixtureAreaWithTags
	fixtureTodoInAreaProject    = "W5JYfjY2xtLdmedQKU6caM" // directly in fixtureProjectInArea
	fixtureTodoDirectlyInArea   = "Q7uN9y3jp5ChZAGjZJhMfY" // loose in fixtureAreaWithTags
	fixtureTodoTrashed          = "Tmt3omxtu61wAoTaSLweQy"
	fixtureTodoInTrashedProject = "NoQLFamrMMooAELuBznao8" // not trashed itself
	fixtureTemplate             = "N1PJHsbjct4mb1bhcs7aHa" // next occurrence 2040-01-01, sentinel deadline
	fixtureRepeatingInstance    = "K9bx7h1xCJdevvyWardZDq"
	fixtureTodoScheduled        = "7F4vqUNiTvGKaCUfv5pqYG" // startDate 2026-09-17

	fixtureTodoInProjectCreationEpoch = int64(1616958920)
	fixtureIncompleteTodos            = 15
	fixtureAreas                      = 3
)

// Task type and status values used by fixture queries.
const (
	typeTodo         = 0
	statusIncomplete = 0
)

// fixtureDatabasePath copies the shared fixture database into the test's
// temporary directory so each test can open (and mutate) its own copy.
func fixtureDatabasePath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	src := filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "main.sqlite")

	data, err := os.ReadFile(src)
	require.NoError(t, err)

	dst := filepath.Join(t.TempDir(), "main.sqlite")
	require.NoError(t, os.WriteFile(dst, data, 0o600)) //nolint:gosec // dst is from t.TempDir(), not user input
	return dst
}

// openFixtureDB opens a private copy of the fixture database.
func openFixtureDB(t *testing.T) *DB {
	t.Helper()
	return openDBAt(t, fixtureDatabasePath(t))
}

// openDBAt opens the database at path and closes it when the test finishes.
func openDBAt(t *testing.T, path string) *DB {
	t.Helper()
	d, err := Open(WithPath(path))
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return d
}

// mutateFixture executes write statements against a fixture database copy.
func mutateFixture(t *testing.T, path string, statements ...string) {
	t.Helper()
	raw, err := sql.Open("sqlite3", path)
	require.NoError(t, err)
	defer raw.Close()

	for _, stmt := range statements {
		_, err := raw.ExecContext(t.Context(), stmt)
		require.NoError(t, err)
	}
}

// queryTaskByUUID fetches a single task row by UUID.
func queryTaskByUUID(t *testing.T, d *DB, uuid string) TaskRow {
	t.Helper()
	rows, err := d.QueryTasks(t.Context(), &TaskFilter{UUID: &uuid})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	return rows[0]
}

// taskUUIDs extracts UUIDs from task rows.
func taskUUIDs(rows []TaskRow) []string {
	uuids := make([]string, len(rows))
	for i := range rows {
		uuids[i] = rows[i].UUID
	}
	return uuids
}

// =============================================================================
// Timezone Model (instants come from raw epochs, packed dates are local days)
// =============================================================================

func TestIntegration_CreatedAtMatchesRawEpoch(t *testing.T) {
	d := openFixtureDB(t)

	row := queryTaskByUUID(t, d, fixtureTodoInProject)

	assert.Equal(t, fixtureTodoInProjectCreationEpoch, row.Created.Unix(),
		"Created must be the raw database epoch regardless of machine timezone")
	assert.Equal(t, time.Local, row.Created.Location())
	assert.False(t, row.Modified.IsZero())
}

func TestIntegration_PackedDatesParseAsLocalMidnight(t *testing.T) {
	d := openFixtureDB(t)

	todayTask := queryTaskByUUID(t, d, fixtureTodoInToday)
	require.NotNil(t, todayTask.StartDate)
	assert.True(t, todayTask.StartDate.Equal(time.Date(2021, 3, 28, 0, 0, 0, 0, time.Local)),
		"StartDate %v must equal local midnight 2021-03-28", todayTask.StartDate)

	headingTask := queryTaskByUUID(t, d, fixtureTodoInHeading)
	require.NotNil(t, headingTask.Deadline)
	assert.True(t, headingTask.Deadline.Equal(time.Date(2040, 11, 4, 0, 0, 0, 0, time.Local)),
		"Deadline %v must equal local midnight 2040-11-04", headingTask.Deadline)
}

func TestIntegration_StopDateMatchesLocaltimeCalendarDay(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()

	// Ground truth straight from SQLite: raw epoch and localtime calendar day.
	type stopInfo struct {
		epoch int64
		day   string
	}
	expected := make(map[string]stopInfo)
	rawRows, err := d.ExecuteQuery(ctx,
		"SELECT uuid, CAST(stopDate AS INTEGER), date(stopDate, 'unixepoch', 'localtime') FROM TMTask WHERE stopDate IS NOT NULL")
	require.NoError(t, err)
	defer rawRows.Close()
	for rawRows.Next() {
		var uuid, day string
		var epoch int64
		require.NoError(t, rawRows.Scan(&uuid, &epoch, &day))
		expected[uuid] = stopInfo{epoch: epoch, day: day}
	}
	require.NoError(t, rawRows.Err())

	taskType := typeTodo
	rows, err := d.QueryTasks(ctx, &TaskFilter{
		Types:          []int{taskType},
		StopDateFilter: &DateFilterValue{HasDate: new(true)},
	})
	require.NoError(t, err)
	require.NotEmpty(t, rows)

	for _, row := range rows {
		require.NotNil(t, row.StopDate, "todo %s should have a stop date", row.UUID)
		want, ok := expected[row.UUID]
		require.True(t, ok, "todo %s missing from raw stopDate map", row.UUID)
		assert.Equal(t, want.epoch, row.StopDate.Unix(),
			"todo %s StopDate instant must match the raw epoch", row.UUID)
		assert.Equal(t, want.day, row.StopDate.In(time.Local).Format(time.DateOnly),
			"todo %s StopDate calendar day must match SQLite localtime day", row.UUID)
	}
}

// =============================================================================
// CreatedAfter Location Insensitivity
// =============================================================================

func TestIntegration_CreatedAfterLocationInsensitive(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()

	taskType := typeTodo
	pivot := time.Date(2021, 3, 29, 6, 0, 0, 0, time.Local)

	local, err := d.QueryTasks(ctx, &TaskFilter{Types: []int{taskType}, CreatedAfter: &pivot})
	require.NoError(t, err)

	pivotUTC := pivot.UTC()
	utc, err := d.QueryTasks(ctx, &TaskFilter{Types: []int{taskType}, CreatedAfter: &pivotUTC})
	require.NoError(t, err)

	assert.ElementsMatch(t, taskUUIDs(local), taskUUIDs(utc),
		"the same instant in different Locations must select the same rows")
	assert.NotEmpty(t, local)
	for _, row := range local {
		assert.True(t, row.Created.After(pivot),
			"todo %s Created %v must be after pivot %v", row.UUID, row.Created, pivot)
	}
}

// =============================================================================
// HasProject Partition
// =============================================================================

func TestIntegration_HasProjectPartition(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()

	incompleteTodos := func(hasProject *bool) []TaskRow {
		taskType, status := typeTodo, statusIncomplete
		rows, err := d.QueryTasks(ctx, &TaskFilter{
			Types:      []int{taskType},
			Status:     &status,
			HasProject: hasProject,
		})
		require.NoError(t, err)
		return rows
	}

	all := incompleteTodos(nil)
	withProject := incompleteTodos(new(true))
	withoutProject := incompleteTodos(new(false))

	require.Len(t, all, fixtureIncompleteTodos)
	assert.Len(t, withProject, len(all)-len(withoutProject), "sets must be disjoint")
	assert.ElementsMatch(t, taskUUIDs(all),
		append(taskUUIDs(withProject), taskUUIDs(withoutProject)...),
		"HasProject(true) and HasProject(false) must partition all incomplete todos")

	for _, row := range withoutProject {
		assert.Empty(t, row.ProjectUUID,
			"HasProject(false) returned todo %s with project", row.UUID)
		assert.Empty(t, row.HeadingUUID,
			"HasProject(false) returned todo %s inside a heading", row.UUID)
	}
	for _, row := range withProject {
		assert.NotEmpty(t, row.ProjectUUID,
			"HasProject(true) returned todo %s without its project", row.UUID)
	}
}

// =============================================================================
// Effective Containers
// =============================================================================

// A todo under a heading has no project or area column of its own; a todo in a
// project has no area column. Both must still report where they live.
func TestIntegration_EffectiveProjectAndArea(t *testing.T) {
	d := openFixtureDB(t)

	underHeading := queryTaskByUUID(t, d, fixtureTodoInHeading)
	assert.Equal(t, fixtureHeading, underHeading.HeadingUUID)
	assert.Equal(t, fixtureProjectInArea, underHeading.ProjectUUID, "the heading's project")
	assert.Equal(t, "Project in Area 1", underHeading.ProjectTitle)
	assert.Equal(t, fixtureAreaWithTags, underHeading.AreaUUID, "the project's area")
	assert.Equal(t, "Area 1", underHeading.AreaTitle)

	inProject := queryTaskByUUID(t, d, fixtureTodoInAreaProject)
	assert.Equal(t, fixtureProjectInArea, inProject.ProjectUUID)
	assert.Equal(t, fixtureAreaWithTags, inProject.AreaUUID, "inherited from the project")
}

func TestIntegration_AreaFilterReachesProjectChildren(t *testing.T) {
	d := openFixtureDB(t)
	taskType, status := typeTodo, statusIncomplete
	area := fixtureAreaWithTags

	rows, err := d.QueryTasks(t.Context(), &TaskFilter{Types: []int{taskType}, Status: &status, AreaUUID: &area})
	require.NoError(t, err)

	uuids := taskUUIDs(rows)
	assert.Contains(t, uuids, fixtureTodoDirectlyInArea)
	assert.Contains(t, uuids, fixtureTodoInAreaProject, "a todo in one of the area's projects")
	assert.Contains(t, uuids, fixtureTodoInHeading, "a todo under a heading of one of the area's projects")
	for _, row := range rows {
		assert.Equal(t, area, row.AreaUUID, "todo %s", row.UUID)
	}
}

// =============================================================================
// Trash and Template Scopes
// =============================================================================

func TestIntegration_TrashScopes(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()
	taskType := typeTodo

	scoped := func(scope TrashScope) []string {
		rows, err := d.QueryTasks(ctx, &TaskFilter{Types: []int{taskType}, Trash: scope})
		require.NoError(t, err)
		return taskUUIDs(rows)
	}
	live, trashed, all := scoped(TrashExcluded), scoped(TrashOnly), scoped(TrashAny)

	assert.NotContains(t, live, fixtureTodoTrashed)
	assert.NotContains(t, live, fixtureTodoInTrashedProject, "a todo inside a trashed project is not live")
	assert.Contains(t, trashed, fixtureTodoTrashed)
	assert.NotContains(t, trashed, fixtureTodoInTrashedProject, "it is not trashed itself either")
	assert.Contains(t, all, fixtureTodoTrashed)
	assert.Contains(t, all, fixtureTodoInTrashedProject, "only the any scope reaches it")
	assert.Subset(t, all, live)
}

// A repeating template carries its next occurrence as its start date and the
// 4001-01-01 sentinel as its deadline, which is not a date at all.
func TestIntegration_TemplateScopes(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()

	ordinary, err := d.QueryTasks(ctx, &TaskFilter{})
	require.NoError(t, err)
	assert.NotContains(t, taskUUIDs(ordinary), fixtureTemplate)

	only, err := d.QueryTasks(ctx, &TaskFilter{Templates: TemplatesOnly})
	require.NoError(t, err)
	require.Len(t, only, 1)
	template := only[0]
	assert.Equal(t, fixtureTemplate, template.UUID)
	assert.True(t, template.Template)
	assert.True(t, template.Repeating)
	require.NotNil(t, template.StartDate)
	assert.True(t, template.StartDate.Equal(time.Date(2040, 1, 1, 0, 0, 0, 0, time.Local)), "its next occurrence")
	assert.Nil(t, template.Deadline, "the per-occurrence sentinel must not decode as a date")

	both, err := d.QueryTasks(ctx, &TaskFilter{Templates: TemplatesAny})
	require.NoError(t, err)
	assert.Len(t, both, len(ordinary)+1)

	instance := queryTaskByUUID(t, d, fixtureRepeatingInstance)
	assert.True(t, instance.Repeating)
	assert.False(t, instance.Template)
}

// =============================================================================
// Clock
// =============================================================================

func TestIntegration_RelativeFiltersFollowTheClock(t *testing.T) {
	path := fixtureDatabasePath(t)
	scheduled := time.Date(2026, 9, 17, 0, 0, 0, 0, time.Local)

	futureAt := func(now time.Time) []string {
		d, err := Open(WithPath(path), WithNow(func() time.Time { return now }))
		require.NoError(t, err)
		defer d.Close()
		rows, err := d.QueryTasks(t.Context(), &TaskFilter{StartDateFilter: &DateFilterValue{Relative: DateFuture}})
		require.NoError(t, err)
		return taskUUIDs(rows)
	}

	assert.Contains(t, futureAt(scheduled.AddDate(0, 0, -1)), fixtureTodoScheduled)
	assert.NotContains(t, futureAt(scheduled), fixtureTodoScheduled, "today is not the future")
	assert.NotContains(t, futureAt(scheduled.AddDate(0, 0, 1)), fixtureTodoScheduled)
}

// =============================================================================
// Row Shape
// =============================================================================

func TestIntegration_ModifiedFallsBackToCreated(t *testing.T) {
	path := fixtureDatabasePath(t)
	mutateFixture(t, path,
		"UPDATE TMTask SET userModificationDate = NULL WHERE uuid = '"+fixtureTodoInToday+"'")
	d := openDBAt(t, path)

	row := queryTaskByUUID(t, d, fixtureTodoInToday)

	assert.False(t, row.Created.IsZero())
	assert.Equal(t, row.Created, row.Modified, "an item never touched since creation was last modified then")
}

func TestIntegration_TimestampsKeepSubSecondPrecision(t *testing.T) {
	path := fixtureDatabasePath(t)
	mutateFixture(t, path,
		"UPDATE TMTask SET userModificationDate = 1700000000.25 WHERE uuid = '"+fixtureTodoInToday+"'")
	d := openDBAt(t, path)

	row := queryTaskByUUID(t, d, fixtureTodoInToday)

	assert.Equal(t, time.UnixMilli(1700000000250), row.Modified.UTC().In(time.Local))
}

// =============================================================================
// LIKE Metacharacter Escaping
// =============================================================================

func TestIntegration_SearchTreatsWildcardsLiterally(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()
	taskType := typeTodo

	literalPercent, err := d.QueryTasks(ctx, &TaskFilter{
		Types:       []int{taskType},
		SearchQuery: new("%"),
	})
	require.NoError(t, err)
	assert.Empty(t, literalPercent,
		"search for literal %% must only match rows containing a percent sign")

	sanity, err := d.QueryTasks(ctx, &TaskFilter{
		Types:       []int{taskType},
		SearchQuery: new("To-Do"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, sanity, "plain text search must still match")
}

func TestIntegration_TitleFilterTreatsUnderscoreLiterally(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()
	taskType := typeTodo

	underscore, err := d.QueryTasks(ctx, &TaskFilter{
		Types: []int{taskType},
		Title: new("To_Do"),
	})
	require.NoError(t, err)
	assert.Empty(t, underscore,
		"underscore in a title filter must not act as a single-character wildcard")

	sanity, err := d.QueryTasks(ctx, &TaskFilter{
		Types: []int{taskType},
		Title: new("To-Do"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, sanity, "plain title filter must still match")
}

// =============================================================================
// Area Visibility NULL Handling
// =============================================================================

func TestIntegration_AreaVisibleTreatsNullAsVisible(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()

	visible, err := d.QueryAreas(ctx, AreaFilter{Visible: new(true)})
	require.NoError(t, err)
	assert.Len(t, visible, fixtureAreas,
		"areas with NULL visible were never hidden and must count as visible")

	hidden, err := d.QueryAreas(ctx, AreaFilter{Visible: new(false)})
	require.NoError(t, err)
	assert.Empty(t, hidden, "no fixture area has been explicitly hidden")
}

// =============================================================================
// Midnight Reminder Round-Trip
// =============================================================================

func TestIntegration_MidnightReminderRoundTrip(t *testing.T) {
	path := fixtureDatabasePath(t)
	mutateFixture(t, path,
		"UPDATE TMTask SET reminderTime = 0 WHERE uuid = '"+fixtureTodoInToday+"'")
	d := openDBAt(t, path)

	row := queryTaskByUUID(t, d, fixtureTodoInToday)

	require.NotNil(t, row.ReminderTime, "a 00:00 reminder must not be dropped")
	assert.Equal(t, "00:00", row.ReminderTime.Format("15:04"))
}

// =============================================================================
// Dangling Tag References
// =============================================================================

func TestIntegration_TagsOfTaskSkipsDanglingTagRef(t *testing.T) {
	path := fixtureDatabasePath(t)
	mutateFixture(t, path,
		"INSERT INTO TMTaskTag (tasks, tags) VALUES ('"+fixtureTodoInToday+"', 'DanglingTagRef00000001')")
	d := openDBAt(t, path)

	tags, err := d.TagsOfTask(t.Context(), fixtureTodoInToday)

	require.NoError(t, err, "a dangling tag reference must not fail the query")
	assert.Equal(t, []string{"Office"}, tags)
}

func TestIntegration_TagsOfAreaSkipsDanglingTagRef(t *testing.T) {
	path := fixtureDatabasePath(t)
	mutateFixture(t, path,
		"INSERT INTO TMAreaTag (areas, tags) VALUES ('"+fixtureAreaWithTags+"', 'DanglingTagRef00000002')")
	d := openDBAt(t, path)

	tags, err := d.TagsOfArea(t.Context(), fixtureAreaWithTags)

	require.NoError(t, err, "a dangling tag reference must not fail the query")
	assert.ElementsMatch(t, []string{"Errand", "Important"}, tags)
}

// A dangling reference next to a real tag used to split the task into two rows.
func TestIntegration_DanglingTagRefNeitherDuplicatesNorCounts(t *testing.T) {
	path := fixtureDatabasePath(t)
	mutateFixture(t, path,
		"INSERT INTO TMTaskTag (tasks, tags) VALUES ('"+fixtureTodoInToday+"', 'DanglingTagRef00000003')",
		"INSERT INTO TMTaskTag (tasks, tags) VALUES ('"+fixtureTodoInProject+"', 'DanglingTagRef00000004')",
		"DELETE FROM TMTaskTag WHERE tasks = '"+fixtureTodoInProject+"' AND tags != 'DanglingTagRef00000004'")
	d := openDBAt(t, path)

	tagged := queryTaskByUUID(t, d, fixtureTodoInToday)
	assert.True(t, tagged.HasTags)

	onlyDangling := queryTaskByUUID(t, d, fixtureTodoInProject)
	assert.False(t, onlyDangling.HasTags, "a reference to a deleted tag is not a tag")
}
