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
func parseThingsURL(t *testing.T, thingsURL string) (command string, params url.Values) {
	t.Helper()
	parsed, err := url.Parse(thingsURL)
	require.NoError(t, err, "failed to parse URL: %s", thingsURL)

	// Extract command from path (e.g., "things:///add" -> "add")
	command = strings.TrimPrefix(parsed.Path, "/")
	params = parsed.Query()
	return command, params
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

// =============================================================================
// JSON Parsing Helpers
// =============================================================================

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
