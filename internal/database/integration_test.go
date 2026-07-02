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
		TaskType:       &taskType,
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

	local, err := d.QueryTasks(ctx, &TaskFilter{TaskType: &taskType, CreatedAfter: &pivot})
	require.NoError(t, err)

	pivotUTC := pivot.UTC()
	utc, err := d.QueryTasks(ctx, &TaskFilter{TaskType: &taskType, CreatedAfter: &pivotUTC})
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
			TaskType:   &taskType,
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
		assert.Nil(t, row.ProjectUUID,
			"HasProject(false) returned todo %s with project", row.UUID)
		assert.Nil(t, row.HeadingUUID,
			"HasProject(false) returned todo %s inside a heading", row.UUID)
	}
	for _, row := range withProject {
		assert.True(t, row.ProjectUUID != nil || row.HeadingUUID != nil,
			"HasProject(true) returned todo %s without project context", row.UUID)
	}
}

// =============================================================================
// LIKE Metacharacter Escaping
// =============================================================================

func TestIntegration_SearchTreatsWildcardsLiterally(t *testing.T) {
	d := openFixtureDB(t)
	ctx := t.Context()
	taskType := typeTodo

	literalPercent, err := d.QueryTasks(ctx, &TaskFilter{
		TaskType:    &taskType,
		SearchQuery: new("%"),
	})
	require.NoError(t, err)
	assert.Empty(t, literalPercent,
		"search for literal %% must only match rows containing a percent sign")

	sanity, err := d.QueryTasks(ctx, &TaskFilter{
		TaskType:    &taskType,
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
		TaskType: &taskType,
		Title:    new("To_Do"),
	})
	require.NoError(t, err)
	assert.Empty(t, underscore,
		"underscore in a title filter must not act as a single-character wildcard")

	sanity, err := d.QueryTasks(ctx, &TaskFilter{
		TaskType: &taskType,
		Title:    new("To-Do"),
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
	assert.Equal(t, "00:00", row.ReminderTime.Format(timeFormat))
}

func Test_thingsTimeExpressionToISOTime_SQL(t *testing.T) {
	raw, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer raw.Close()

	tests := []struct {
		name    string
		literal string
		want    *string
	}{
		{"null means no reminder", "NULL", nil},
		{"zero is midnight", "0", new("00:00")},
		{"packed afternoon time", "840957952", new("12:34")},
		{"packed end of day", "1605369856", new("23:59")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got sql.NullString
			query := "SELECT " + thingsTimeExpressionToISOTime(tt.literal)
			require.NoError(t, raw.QueryRowContext(t.Context(), query).Scan(&got))
			if tt.want == nil {
				assert.False(t, got.Valid, "expected NULL")
			} else {
				require.True(t, got.Valid)
				assert.Equal(t, *tt.want, got.String)
			}
		})
	}
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
