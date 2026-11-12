package things3

import (
	"encoding/json"
	"testing"

	"github.com/moond4rk/things3/thingstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repeatTemplateUUID is the fixture's repeat-template row (rt1_recurrenceRule
// set); it must never appear in query results.
const repeatTemplateUUID = "N1PJHsbjct4mb1bhcs7aHa"

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

	client, err := NewClient(WithDatabasePath(dbPath), WithClock(thingstest.Now))
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
