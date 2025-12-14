package things3

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
// Reuses the existing JSONItem struct from scheme_json.go.
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
