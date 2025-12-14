package things3

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

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
//
//nolint:unparam // key varies by test case, current tests happen to use same value
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
// Date Assertion Helpers
// =============================================================================

// assertDateParam asserts that a URL date parameter matches the expected date.
// Parses RFC3339 format and compares year, month, day.
//
//nolint:unparam // year varies by test case, current tests happen to use 2024
func assertDateParam(t *testing.T, params url.Values, key string, year int, month time.Month, day int) {
	t.Helper()
	value := params.Get(key)
	require.NotEmpty(t, value, "parameter %q should exist", key)

	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err, "parameter %q should be valid RFC3339: %s", key, value)

	assert.Equal(t, year, parsed.Year(), "parameter %q year mismatch", key)
	assert.Equal(t, month, parsed.Month(), "parameter %q month mismatch", key)
	assert.Equal(t, day, parsed.Day(), "parameter %q day mismatch", key)
}

// assertJSONDateAttr asserts that a JSON attribute matches the expected date.
// Parses RFC3339 format and compares year, month, day.
//
//nolint:unparam // year varies by test case, current tests happen to use 2024
func assertJSONDateAttr(t *testing.T, attrs map[string]any, key string, year int, month time.Month, day int) {
	t.Helper()
	value, ok := attrs[key].(string)
	require.True(t, ok, "attribute %q should be a string", key)

	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err, "attribute %q should be valid RFC3339: %s", key, value)

	assert.Equal(t, year, parsed.Year(), "attribute %q year mismatch", key)
	assert.Equal(t, month, parsed.Month(), "attribute %q month mismatch", key)
	assert.Equal(t, day, parsed.Day(), "attribute %q day mismatch", key)
}

// =============================================================================
// JSON Parsing Helpers
// =============================================================================

// parseJSONItems parses JSON data into strongly-typed JSONItem slice.
// Reuses the existing JSONItem struct from scheme_json.go.
func parseJSONItems(t *testing.T, urlStr string) []JSONItem {
	t.Helper()
	_, params := parseThingsURL(t, urlStr)
	data := params.Get("data")
	require.NotEmpty(t, data, "data parameter should not be empty")

	var items []JSONItem
	err := json.Unmarshal([]byte(data), &items)
	require.NoError(t, err, "failed to parse JSON data")
	return items
}
