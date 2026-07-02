package things3

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testUUIDTodoInUpcoming is a fixture todo scheduled for a future date (Someday
// bucket with a future start date), so it belongs to the Upcoming view.
const testUUIDTodoInUpcoming = "7F4vqUNiTvGKaCUfv5pqYG"

func TestClientUpcoming(t *testing.T) {
	client := newTestClient(t)
	ctx := t.Context()

	todos, err := client.Upcoming(ctx)
	require.NoError(t, err)
	require.NotNil(t, todos)

	uuids := extractTodoUUIDs(todos)
	// The regular future-scheduled todo and the repeating template both appear.
	assert.Contains(t, uuids, testUUIDTodoInUpcoming)
	assert.Contains(t, uuids, repeatTemplateUUID)

	// Every Upcoming todo is incomplete, not trashed, and future-dated.
	for i := range todos {
		assert.Equalf(t, StatusIncomplete, todos[i].Status, "upcoming todo %s must be incomplete", todos[i].UUID)
		assert.Falsef(t, todos[i].Trashed, "upcoming todo %s must not be trashed", todos[i].UUID)
		assert.NotNilf(t, todos[i].StartDate, "upcoming todo %s must have a start date", todos[i].UUID)
	}

	// The merged result is sorted ascending by start date.
	assert.Truef(t, slices.IsSortedFunc(todos, func(a, b Todo) int {
		return compareStartDateAsc(a.StartDate, b.StartDate)
	}), "upcoming todos must be sorted ascending by start date")

	// The repeating template surfaces its next occurrence as its start date.
	tmpl := findTodo(todos, repeatTemplateUUID)
	require.NotNil(t, tmpl, "repeating template must appear in Upcoming")
	assert.True(t, tmpl.Repeating, "repeating template must be flagged repeating")
	require.NotNil(t, tmpl.StartDate)
	assert.Equal(t, time.Date(2040, 1, 1, 0, 0, 0, 0, time.Local), *tmpl.StartDate)
}

func TestClientUpcomingEmptyIsNonNil(t *testing.T) {
	dbPath := copyWritableFixture(t)
	// Completing every active todo (including the repeating template) empties
	// the Upcoming view.
	execFixtureSQL(t, dbPath, "UPDATE TMTask SET status = 3 WHERE status = 0")

	client, err := NewClient(WithDatabasePath(dbPath))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	todos, err := client.Upcoming(t.Context())
	require.NoError(t, err)
	require.NotNil(t, todos)
	assert.Empty(t, todos)
}

// findTodo returns a pointer to the todo with the given UUID, or nil.
func findTodo(todos []Todo, uuid string) *Todo {
	i := slices.IndexFunc(todos, func(t Todo) bool { return t.UUID == uuid })
	if i < 0 {
		return nil
	}
	return &todos[i]
}
