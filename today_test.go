package things3

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// repeatTemplateUUID is the fixture's repeat-template row (rt1_recurrenceRule
// set); it must never appear in query results.
const repeatTemplateUUID = "N1PJHsbjct4mb1bhcs7aHa"

func TestClientToday(t *testing.T) {
	client := newTestClient(t)
	ctx := t.Context()

	todos, err := client.Today(ctx)
	require.NoError(t, err)
	require.NotNil(t, todos)

	uuids := extractTodoUUIDs(todos)
	// regular (start-date + Anytime) and overdue-deadline branches
	assert.Contains(t, uuids, testUUIDTodoInToday)
	assert.Contains(t, uuids, testUUIDTodoOverdueInToday)
	// every yellow-dot todo (Someday with a past start date) also appears
	someday, err := client.Todos().
		StartDate().Past().Start().Someday().Status().Incomplete().All(ctx)
	require.NoError(t, err)
	for i := range someday {
		assert.Containsf(t, uuids, someday[i].UUID, "yellow-dot todo %s should appear in Today", someday[i].UUID)
	}

	for i := range todos {
		assert.Equalf(t, StatusIncomplete, todos[i].Status, "Today todo %s must be incomplete", todos[i].UUID)
		assert.Falsef(t, todos[i].Trashed, "Today todo %s must not be trashed", todos[i].UUID)
	}
}

func TestClientTodayEmptyIsNonNil(t *testing.T) {
	dbPath := copyWritableFixture(t)
	execFixtureSQL(t, dbPath, "UPDATE TMTask SET status = 3 WHERE status = 0")

	client, err := NewClient(WithDatabasePath(dbPath))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	todos, err := client.Today(t.Context())
	require.NoError(t, err)
	require.NotNil(t, todos)
	assert.Empty(t, todos)
}

func TestClientTodayEveningSortsAfterRegular(t *testing.T) {
	dbPath := copyWritableFixture(t)
	// startBucket == 1 marks This Evening (confirmed live value, schema v26).
	require.Equal(t, int64(1),
		execFixtureSQL(t, dbPath, "UPDATE TMTask SET startBucket = 1 WHERE uuid = ?", testUUIDTodoInToday))

	client, err := NewClient(WithDatabasePath(dbPath))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	ctx := t.Context()

	todos, err := client.Today(ctx)
	require.NoError(t, err)

	// The regular Today group is the head of the result.
	regular, err := client.Todos().
		StartDate().Exists(true).Start().Anytime().Status().Incomplete().All(ctx)
	require.NoError(t, err)
	require.LessOrEqual(t, len(regular), len(todos))
	prefix := todos[:len(regular)]

	// Exactly the injected todo is evening; every other regular member is not;
	// the evening todo sorts to the end of the regular group.
	eveningIdx := -1
	for i := range prefix {
		if prefix[i].UUID == testUUIDTodoInToday {
			assert.True(t, prefix[i].Evening, "injected todo must report Evening")
			eveningIdx = i
		} else {
			assert.Falsef(t, prefix[i].Evening, "non-injected regular todo %s must not be evening", prefix[i].UUID)
		}
	}
	require.GreaterOrEqual(t, eveningIdx, 0, "injected evening todo must be in the regular group")
	assert.Equal(t, len(prefix)-1, eveningIdx, "the evening todo must sort to the end of the regular group")
}

func TestTodoRepeatingField(t *testing.T) {
	client := newTestClient(t)
	ctx := t.Context()

	instance, err := client.Todos().WithUUID(testUUIDTodoRepeating).Status().Any().First(ctx)
	require.NoError(t, err)
	assert.True(t, instance.Repeating, "repeat instance must be flagged repeating")

	normal, err := client.Todos().WithUUID(testUUIDTodoInToday).Status().Any().First(ctx)
	require.NoError(t, err)
	assert.False(t, normal.Repeating, "a normal todo is not repeating")

	all, err := client.Todos().Status().Any().All(ctx)
	require.NoError(t, err)
	assert.NotContains(t, extractTodoUUIDs(all), repeatTemplateUUID,
		"the repeat template row must be filtered from every query")
}

func TestProjectRepeatingField(t *testing.T) {
	dbPath := copyWritableFixture(t)
	// Flag the project as an instance of the repeat template (leave
	// rt1_recurrenceRule NULL so it is not treated as a template itself).
	require.Equal(t, int64(1),
		execFixtureSQL(t, dbPath, "UPDATE TMTask SET rt1_repeatingTemplate = ? WHERE uuid = ? AND type = 1",
			repeatTemplateUUID, testUUIDProjectInArea1))

	client, err := NewClient(WithDatabasePath(dbPath))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	ctx := t.Context()

	repeating, err := client.Projects().WithUUID(testUUIDProjectInArea1).Status().Any().First(ctx)
	require.NoError(t, err)
	assert.True(t, repeating.Repeating, "project flagged as a repeat instance must report Repeating")

	normal, err := newTestClient(t).Projects().WithUUID(testUUIDProjectInArea1).Status().Any().First(ctx)
	require.NoError(t, err)
	assert.False(t, normal.Repeating, "an unmodified project is not repeating")
}

func TestTodoProjectRepeatingJSONTags(t *testing.T) {
	todoJSON, err := json.Marshal(Todo{Evening: true, Repeating: true})
	require.NoError(t, err)
	assert.Contains(t, string(todoJSON), `"evening":true`)
	assert.Contains(t, string(todoJSON), `"repeating":true`)

	zeroTodo, err := json.Marshal(Todo{})
	require.NoError(t, err)
	assert.NotContains(t, string(zeroTodo), `"evening"`)
	assert.NotContains(t, string(zeroTodo), `"repeating"`)

	projJSON, err := json.Marshal(Project{Repeating: true})
	require.NoError(t, err)
	assert.Contains(t, string(projJSON), `"repeating":true`)

	zeroProj, err := json.Marshal(Project{})
	require.NoError(t, err)
	assert.NotContains(t, string(zeroProj), `"repeating"`)
}

// copyWritableFixture copies the tracked read-only fixture (and any WAL/SHM
// sidecars) into a temp dir so a test can mutate it safely.
func copyWritableFixture(t *testing.T) string {
	t.Helper()
	initTestPaths()
	dst := filepath.Join(t.TempDir(), "main.sqlite")
	for _, suffix := range []string{"", "-wal", "-shm"} {
		data, err := os.ReadFile(testDatabasePath + suffix)
		if err != nil {
			if suffix == "" {
				t.Fatalf("read fixture: %v", err)
			}
			continue
		}
		require.NoError(t, os.WriteFile(dst+suffix, data, 0o600)) //nolint:gosec // writes into t.TempDir
	}
	return dst
}

// execFixtureSQL runs a write statement against a fixture copy and returns the
// number of affected rows.
func execFixtureSQL(t *testing.T, dbPath, query string, args ...any) int64 {
	t.Helper()
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	res, err := db.ExecContext(t.Context(), query, args...)
	require.NoError(t, err)
	n, err := res.RowsAffected()
	require.NoError(t, err)
	return n
}
