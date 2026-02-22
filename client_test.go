package things3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient creates a new Client connected to the test database.
func newTestClient(t *testing.T) *Client {
	t.Helper()
	initTestPaths()
	client, err := NewClient(WithDatabasePath(testDatabasePath))
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	return client
}

func TestNewClient(t *testing.T) {
	initTestPaths()

	t.Run("success", func(t *testing.T) {
		client, err := NewClient(WithDatabasePath(testDatabasePath))
		require.NoError(t, err)
		require.NotNil(t, client)
		assert.NoError(t, client.Close())
	})

	t.Run("nonexistent database", func(t *testing.T) {
		_, err := NewClient(WithDatabasePath("/nonexistent/path/main.sqlite"))
		assert.Error(t, err)
	})

	t.Run("close nil database", func(t *testing.T) {
		client := &Client{}
		assert.NoError(t, client.Close())
	})
}

func TestClientConvenienceMethods(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	t.Run("Inbox", func(t *testing.T) {
		tasks, err := client.Inbox(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testInboxUUIDs, extractUUIDs(tasks))
	})

	t.Run("Today", func(t *testing.T) {
		tasks, err := client.Today(ctx)
		require.NoError(t, err)
		assert.Len(t, tasks, testToday)
	})

	t.Run("Todos", func(t *testing.T) {
		tasks, err := client.Todos(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testTodosIncompleteUUIDs, extractUUIDs(tasks))
	})

	t.Run("Projects", func(t *testing.T) {
		tasks, err := client.Projects(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testProjectsNotTrashedUUIDs, extractUUIDs(tasks))
	})

	t.Run("Upcoming", func(t *testing.T) {
		tasks, err := client.Upcoming(ctx)
		require.NoError(t, err)
		assert.Len(t, tasks, testUpcoming)
	})

	t.Run("Anytime", func(t *testing.T) {
		tasks, err := client.Anytime(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testAnytimeUUIDs, extractUUIDs(tasks))
	})

	t.Run("Someday", func(t *testing.T) {
		tasks, err := client.Someday(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testSomedayUUIDs, extractUUIDs(tasks))
	})

	t.Run("Logbook", func(t *testing.T) {
		tasks, err := client.Logbook(ctx)
		require.NoError(t, err)
		assert.Len(t, tasks, testLogbook)
	})

	t.Run("Trash", func(t *testing.T) {
		tasks, err := client.Trash(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testTrashedUUIDs, extractUUIDs(tasks))
	})

	t.Run("Completed", func(t *testing.T) {
		tasks, err := client.Completed(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testCompletedUUIDs, extractUUIDs(tasks))
	})

	t.Run("Canceled", func(t *testing.T) {
		tasks, err := client.Canceled(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testCanceledUUIDs, extractUUIDs(tasks))
	})

	t.Run("Deadlines", func(t *testing.T) {
		tasks, err := client.Deadlines(ctx)
		require.NoError(t, err)
		assert.Len(t, tasks, testDeadlines)
	})
}

func TestClientQueryBuilders(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	t.Run("Tasks builder", func(t *testing.T) {
		count, err := client.Tasks().
			Status().Incomplete().
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, testTasksIncomplete, count)
	})

	t.Run("Areas builder", func(t *testing.T) {
		areas, err := client.Areas().All(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testAreaUUIDs, extractAreaUUIDs(areas))
	})

	t.Run("Tags builder", func(t *testing.T) {
		tags, err := client.Tags().All(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, testTagUUIDs, extractTagUUIDs(tags))
	})
}

func TestClientGet(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	t.Run("task", func(t *testing.T) {
		result, err := client.Get(ctx, testUUIDTodoInbox)
		require.NoError(t, err)
		require.NotNil(t, result)
		task, ok := result.(*Task)
		require.True(t, ok, "expected *Task, got %T", result)
		assert.Equal(t, testUUIDTodoInbox, task.UUID)
	})

	t.Run("area", func(t *testing.T) {
		result, err := client.Get(ctx, testUUIDArea1)
		require.NoError(t, err)
		require.NotNil(t, result)
		area, ok := result.(*Area)
		require.True(t, ok, "expected *Area, got %T", result)
		assert.Equal(t, testUUIDArea1, area.UUID)
	})

	t.Run("tag", func(t *testing.T) {
		result, err := client.Get(ctx, testUUIDTagOffice)
		require.NoError(t, err)
		require.NotNil(t, result)
		tag, ok := result.(*Tag)
		require.True(t, ok, "expected *Tag, got %T", result)
		assert.Equal(t, testUUIDTagOffice, tag.UUID)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := client.Get(ctx, "nonexistent-uuid")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestClientSearch(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Search(ctx, "To-Do in Today")
	require.NoError(t, err)
	assert.NotEmpty(t, tasks)
}

func TestClientChecklistItems(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	items, err := client.ChecklistItems(ctx, testUUIDTodoInboxChecklist)
	require.NoError(t, err)
	assert.ElementsMatch(t, testChecklistUUIDs, extractChecklistUUIDs(items))
}

func TestClientToken(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	token, err := client.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAuthToken, token)

	// Second call should return cached token
	token2, err := client.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, token, token2)
}

func TestClientURLSchemeBuilders(t *testing.T) {
	client := newTestClient(t)

	t.Run("AddTodo", func(t *testing.T) {
		uri, err := client.AddTodo().Title("test task").Build()
		require.NoError(t, err)
		cmd, params := parseThingsURL(t, uri)
		require.Equal(t, "add", cmd)
		require.Equal(t, "test task", params.Get("title"))
	})

	t.Run("AddProject", func(t *testing.T) {
		uri, err := client.AddProject().Title("test project").Build()
		require.NoError(t, err)
		cmd, params := parseThingsURL(t, uri)
		require.Equal(t, "add-project", cmd)
		require.Equal(t, "test project", params.Get("title"))
	})

	t.Run("ShowBuilder", func(t *testing.T) {
		uri, err := client.ShowBuilder().ID("some-uuid").Build()
		require.NoError(t, err)
		cmd, params := parseThingsURL(t, uri)
		require.Equal(t, "show", cmd)
		require.Equal(t, "some-uuid", params.Get("id"))
	})

	t.Run("Batch", func(t *testing.T) {
		uri, err := client.Batch().
			AddTodo(func(b BatchTodoConfigurator) {
				b.Title("batch task")
			}).
			Build()
		require.NoError(t, err)
		cmd, _ := parseThingsURL(t, uri)
		require.Equal(t, "json", cmd)
		items := parseJSONItems(t, uri)
		require.Len(t, items, 1)
		require.Equal(t, JSONItemTypeTodo, items[0].Type)
		require.Equal(t, "batch task", items[0].Attributes["title"])
	})

	t.Run("AuthBatch with UpdateTodo", func(t *testing.T) {
		uri, err := client.AuthBatch().
			UpdateTodo("some-uuid", func(b BatchTodoConfigurator) {
				b.Title("updated")
			}).
			Build()
		require.NoError(t, err)
		cmd, params := parseThingsURL(t, uri)
		require.Equal(t, "json", cmd)
		require.Equal(t, testAuthToken, params.Get("auth-token"))
		items := parseJSONItems(t, uri)
		require.Len(t, items, 1)
		require.Equal(t, JSONItemTypeTodo, items[0].Type)
		require.Equal(t, JSONOperationUpdate, items[0].Operation)
		require.Equal(t, "some-uuid", items[0].ID)
		require.Equal(t, "updated", items[0].Attributes["title"])
	})

	t.Run("UpdateTodo", func(t *testing.T) {
		uri, err := client.UpdateTodo("some-uuid").Title("updated").Build()
		require.NoError(t, err)
		cmd, params := parseThingsURL(t, uri)
		require.Equal(t, "update", cmd)
		require.Equal(t, testAuthToken, params.Get("auth-token"))
		require.Equal(t, "some-uuid", params.Get("id"))
		require.Equal(t, "updated", params.Get("title"))
	})

	t.Run("UpdateProject", func(t *testing.T) {
		uri, err := client.UpdateProject("some-uuid").Title("updated").Build()
		require.NoError(t, err)
		cmd, params := parseThingsURL(t, uri)
		require.Equal(t, "update-project", cmd)
		require.Equal(t, testAuthToken, params.Get("auth-token"))
		require.Equal(t, "some-uuid", params.Get("id"))
		require.Equal(t, "updated", params.Get("title"))
	})
}
