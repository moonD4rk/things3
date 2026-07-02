package things3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moond4rk/things3/thingstest"
)

// newThingstestClient creates a Client backed by a writable copy of the
// thingstest fixture database.
func newThingstestClient(t *testing.T) *Client {
	t.Helper()
	client, err := NewClient(WithDatabasePath(thingstest.DatabasePath(t)))
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	return client
}

// =============================================================================
// Copy-on-Write Fork Isolation Tests
// =============================================================================

// TestTodoBuilderForkIsolation verifies that forking a base builder produces
// independent queries: filters applied on one fork must not leak into the
// other fork or back into the base.
func TestTodoBuilderForkIsolation(t *testing.T) {
	client := newThingstestClient(t)
	ctx := t.Context()

	base := client.Todos().Status().Incomplete()

	fork1, err := base.StartDate().Past().All(ctx)
	require.NoError(t, err)
	fork2, err := base.Deadline().Exists(true).All(ctx)
	require.NoError(t, err)

	assert.Len(t, fork1, 3, "fork1 must only carry incomplete + StartDate().Past()")
	assert.Len(t, fork2, 4, "fork2 must only carry incomplete + Deadline().Exists(true)")

	baseCount, err := base.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, thingstest.TodosIncomplete, baseCount,
		"base builder must stay unaffected by forks")
}

// TestTodoBuilderChainAccumulates verifies that chaining within a single
// chain still accumulates all filters.
func TestTodoBuilderChainAccumulates(t *testing.T) {
	client := newThingstestClient(t)
	ctx := t.Context()

	todos, err := client.Todos().
		Status().Incomplete().
		StartDate().Past().
		Deadline().Exists(true).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, 1, "single chain must accumulate all three filters")
}

// TestSubFilterFactoryCopies verifies that terminal calls on sub-filter
// factories (Status(), Start(), date filters) copy the parent snapshot
// instead of mutating it.
func TestSubFilterFactoryCopies(t *testing.T) {
	client := newThingstestClient(t)
	ctx := t.Context()

	base := client.Todos()

	incompleteCount, err := base.Status().Incomplete().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, thingstest.TodosIncomplete, incompleteCount)

	baseCount, err := base.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 37, baseCount, "Status().Incomplete() must not mutate the base builder")

	statusFactory := base.Status()
	completedCount, err := statusFactory.Completed().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, thingstest.TodosComplete, completedCount)

	incompleteAgain, err := statusFactory.Incomplete().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, thingstest.TodosIncomplete, incompleteAgain,
		"reusing a sub-filter factory must fork, not overwrite")
}

// TestBuilderForkIsolationAllTypes verifies copy-on-write across project,
// heading, area, and tag builders by comparing forked results against fresh
// single-chain baselines.
func TestBuilderForkIsolationAllTypes(t *testing.T) {
	client := newThingstestClient(t)
	ctx := t.Context()

	t.Run("project builder", func(t *testing.T) {
		defaultCount, err := client.Projects().Count(ctx)
		require.NoError(t, err)
		trashedCount, err := client.Projects().Trashed(true).Count(ctx)
		require.NoError(t, err)
		require.NotEqual(t, defaultCount, trashedCount, "fixture must distinguish the two queries")

		base := client.Projects()
		forkCount, err := base.Trashed(true).Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, trashedCount, forkCount)

		baseCount, err := base.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, defaultCount, baseCount, "fork must not mutate the base project builder")
	})

	t.Run("heading builder", func(t *testing.T) {
		defaultCount, err := client.Headings().Count(ctx)
		require.NoError(t, err)
		require.Positive(t, defaultCount)

		base := client.Headings()
		forkCount, err := base.InProject("nonexistent-project").Count(ctx)
		require.NoError(t, err)
		assert.Zero(t, forkCount)

		baseCount, err := base.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, defaultCount, baseCount, "fork must not mutate the base heading builder")
	})

	t.Run("area builder", func(t *testing.T) {
		base := client.Areas()
		forkAreas, err := base.WithTitle("Area 1").All(ctx)
		require.NoError(t, err)
		assert.Len(t, forkAreas, 1)

		baseAreas, err := base.All(ctx)
		require.NoError(t, err)
		assert.Len(t, baseAreas, thingstest.Areas, "fork must not mutate the base area builder")
	})

	t.Run("tag builder", func(t *testing.T) {
		base := client.Tags()
		forkTags, err := base.WithTitle("Home").All(ctx)
		require.NoError(t, err)
		assert.Len(t, forkTags, 1)

		baseTags, err := base.All(ctx)
		require.NoError(t, err)
		assert.Len(t, baseTags, thingstest.Tags, "fork must not mutate the base tag builder")
	})
}

// =============================================================================
// First() Side Effect Tests
// =============================================================================

// TestTodoFirstDoesNotMutateBuilder verifies that First() applies its
// checklist auto-include and limit only to its own copy: All() on the
// original builder afterwards must not return checklists.
func TestTodoFirstDoesNotMutateBuilder(t *testing.T) {
	client := newThingstestClient(t)
	ctx := t.Context()

	q := client.Todos().WithUUID(thingstest.UUIDTodoChecklist)

	first, err := q.First(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, first.Checklist, "First() keeps auto-including the checklist")

	todos, err := q.All(ctx)
	require.NoError(t, err)
	require.Len(t, todos, 1)
	assert.Empty(t, todos[0].Checklist,
		"All() on the original builder must not inherit First()'s checklist opt-in")
}

// TestFirstDoesNotLeakLimit verifies that the Limit(1) pushed by First()
// stays on First()'s private copy.
func TestFirstDoesNotLeakLimit(t *testing.T) {
	client := newThingstestClient(t)
	ctx := t.Context()

	q := client.Todos().Status().Incomplete()

	_, err := q.First(ctx)
	require.NoError(t, err)

	todos, err := q.All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, thingstest.TodosIncomplete,
		"All() after First() must return the full result set")
}

// TestFirstStateIsolationWhitebox asserts directly on internal state that
// First() leaves the receiver untouched.
func TestFirstStateIsolationWhitebox(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	tq := db.Todos()
	_, err := tq.First(ctx)
	require.NoError(t, err)
	assert.Nil(t, tq.inner.filter.Limit, "First() must not set Limit on the receiver")
	assert.False(t, tq.inner.includeChecklist, "First() must not set includeChecklist on the receiver")

	pq := db.Projects()
	_, err = pq.First(ctx)
	require.NoError(t, err)
	assert.Nil(t, pq.inner.filter.Limit, "project First() must not set Limit on the receiver")

	hq := db.Headings()
	_, err = hq.First(ctx)
	require.NoError(t, err)
	assert.Nil(t, hq.inner.filter.Limit, "heading First() must not set Limit on the receiver")
}
