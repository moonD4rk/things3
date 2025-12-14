package things3

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// URL Parsing Helpers
// =============================================================================

// parseThingsURL parses a Things URL and returns the command and query parameters.
func parseThingsURL(t *testing.T, urlStr string) (command string, params url.Values) {
	t.Helper()
	parsed, err := url.Parse(urlStr)
	require.NoError(t, err, "failed to parse URL: %s", urlStr)

	// Extract command from path (e.g., "things:///add" -> "add")
	command = strings.TrimPrefix(parsed.Path, "/")
	params = parsed.Query()
	return command, params
}

// assertURLParam asserts that a URL parameter has the expected value.
func assertURLParam(t *testing.T, params url.Values, key, expected string) {
	t.Helper()
	actual := params.Get(key)
	assert.Equal(t, expected, actual, "parameter %q mismatch", key)
}

// assertURLParamExists asserts that a URL parameter exists.
func assertURLParamExists(t *testing.T, params url.Values, key string) {
	t.Helper()
	assert.True(t, params.Has(key), "parameter %q should exist", key)
}

// assertURLParamNotExists asserts that a URL parameter does not exist.
func assertURLParamNotExists(t *testing.T, params url.Values, key string) {
	t.Helper()
	assert.False(t, params.Has(key), "parameter %q should not exist", key)
}

// assertNoExtraParams asserts that only the expected parameters exist.
func assertNoExtraParams(t *testing.T, params url.Values, expected ...string) {
	t.Helper()
	expectedSet := make(map[string]bool)
	for _, k := range expected {
		expectedSet[k] = true
	}
	for k := range params {
		assert.True(t, expectedSet[k], "unexpected parameter %q", k)
	}
}

// =============================================================================
// JSON Parsing Helpers
// =============================================================================

// parseJSONData extracts and parses the JSON data parameter from a Things URL.
func parseJSONData(t *testing.T, urlStr string) []map[string]any {
	t.Helper()
	_, params := parseThingsURL(t, urlStr)
	data := params.Get("data")
	require.NotEmpty(t, data, "data parameter should not be empty")

	var items []map[string]any
	err := json.Unmarshal([]byte(data), &items)
	require.NoError(t, err, "failed to parse JSON data")
	return items
}

// assertJSONItemType asserts the type of a JSON item.
func assertJSONItemType(t *testing.T, item map[string]any, expectedType string) {
	t.Helper()
	assert.Equal(t, expectedType, item["type"], "item type mismatch")
}

// getJSONAttrs extracts the attributes map from a JSON item.
func getJSONAttrs(t *testing.T, item map[string]any) map[string]any {
	t.Helper()
	attrs, ok := item["attributes"].(map[string]any)
	require.True(t, ok, "attributes should be a map")
	return attrs
}

// assertJSONAttr asserts a specific attribute value in a JSON item.
func assertJSONAttr(t *testing.T, attrs map[string]any, key string, expected any) {
	t.Helper()
	actual := attrs[key]
	assert.Equal(t, expected, actual, "attribute %q mismatch", key)
}

// assertJSONAttrExists asserts that an attribute exists in a JSON item.
func assertJSONAttrExists(t *testing.T, attrs map[string]any, key string) {
	t.Helper()
	_, ok := attrs[key]
	assert.True(t, ok, "attribute %q should exist", key)
}
