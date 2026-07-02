package things3

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moond4rk/things3/thingstest"
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

func TestClientQueryBuilders(t *testing.T) {
	client := newTestClient(t)
	ctx := t.Context()

	t.Run("Todos builder", func(t *testing.T) {
		count, err := client.Todos().
			Status().Incomplete().
			Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, testTodosIncomplete, count)
	})

	t.Run("Projects builder", func(t *testing.T) {
		projects, err := client.Projects().
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, projects)
	})

	t.Run("Headings builder", func(t *testing.T) {
		headings, err := client.Headings().All(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, headings)
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

func TestClientToken(t *testing.T) {
	client := newTestClient(t)
	ctx := t.Context()

	token, err := client.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAuthToken, token)

	// Second call should return cached token
	token2, err := client.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, token, token2)
}

func TestClientTokenEmptyInDatabase(t *testing.T) {
	// An empty stored token means Things URL scheme authorization was never
	// enabled; Token must fail with a descriptive error instead of returning
	// the empty token as success.
	dbPath := thingstest.DatabasePath(t)

	raw, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	_, err = raw.ExecContext(t.Context(), "UPDATE TMSettings SET uriSchemeAuthenticationToken = ''")
	require.NoError(t, err)
	require.NoError(t, raw.Close())

	client, err := NewClient(WithDatabasePath(dbPath))
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })

	_, err = client.Token(t.Context())
	require.ErrorIs(t, err, ErrAuthTokenNotFound)
	assert.Contains(t, err.Error(), "authorization not set up")
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
