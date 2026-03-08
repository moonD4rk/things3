package things3

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	ischeme "github.com/moond4rk/things3/internal/scheme"
)

// testScheme wraps scheme.Scheme with factory methods for testing.
type testScheme struct {
	s *ischeme.Scheme
}

func newScheme(opts ...ischeme.Option) *testScheme {
	return &testScheme{s: ischeme.New(opts...)}
}

func withForeground() ischeme.Option { return ischeme.WithForeground() }

func (ts *testScheme) AddTodo() TodoAdder {
	return ischeme.NewTodoAdder(ts.s)
}

func (ts *testScheme) AddProject() ProjectAdder {
	return ischeme.NewProjectAdder(ts.s)
}

func (ts *testScheme) ShowBuilder() ShowNavigator {
	return ischeme.NewShowNavigator(ts.s)
}

func (ts *testScheme) Batch() BatchCreator {
	return ischeme.NewBatch(ts.s)
}

func (ts *testScheme) WithToken(token string) *testAuthScheme {
	return &testAuthScheme{s: ts.s, token: token}
}

func (ts *testScheme) SearchURL(query string) string {
	q := url.Values{}
	q.Set("query", query)
	return fmt.Sprintf("things:///%s?%s", CommandSearch, ischeme.EncodeQuery(q))
}

func (ts *testScheme) Version() string {
	return fmt.Sprintf("things:///%s", CommandVersion)
}

// testAuthScheme wraps authenticated operations for testing.
type testAuthScheme struct {
	s     *ischeme.Scheme
	token string
}

func (a *testAuthScheme) UpdateTodo(id string) TodoUpdater {
	token := a.token
	return ischeme.NewTodoUpdater(a.s, func(_ context.Context) (string, error) {
		return token, nil
	}, id)
}

func (a *testAuthScheme) UpdateProject(id string) ProjectUpdater {
	token := a.token
	return ischeme.NewProjectUpdater(a.s, func(_ context.Context) (string, error) {
		return token, nil
	}, id)
}

func (a *testAuthScheme) Batch() AuthBatchCreator {
	token := a.token
	return ischeme.NewAuthBatch(a.s, func(_ context.Context) (string, error) {
		return token, nil
	})
}

// parseThingsURL parses a Things URL and returns the command and query parameters.
func parseThingsURL(t *testing.T, thingsURL string) (command string, params url.Values) {
	t.Helper()
	parsed, err := url.Parse(thingsURL)
	require.NoError(t, err, "failed to parse URL: %s", thingsURL)

	// Extract command from path (e.g., "things:///add" -> "add")
	command = strings.TrimPrefix(parsed.Path, "/")
	params = parsed.Query()
	return command, params
}

// parseJSONItems parses JSON data into strongly-typed JSONItem slice.
func parseJSONItems(t *testing.T, thingsURL string) []JSONItem {
	t.Helper()
	_, params := parseThingsURL(t, thingsURL)
	data := params.Get("data")
	require.NotEmpty(t, data, "data parameter should not be empty")

	var items []JSONItem
	err := json.Unmarshal([]byte(data), &items)
	require.NoError(t, err, "failed to parse JSON data")
	return items
}
