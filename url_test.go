package things3

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowURL(t *testing.T) {
	db := newTestDB(t)

	url := db.ShowURL("uuid")
	assert.Equal(t, "things:///show?id=uuid", url)
}

func TestLink(t *testing.T) {
	db := newTestDB(t)

	// Link is alias for ShowURL
	link := db.Link("uuid")
	assert.Equal(t, "things:///show?id=uuid", link)
}

func TestURL(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test show command
	url, err := db.URL(ctx, URLCommandShow, map[string]string{"id": "uuid"})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(url, "things:///show?"), "URL(show) should have prefix things:///show?")
	assert.Contains(t, url, "id=uuid")
}

func TestAddTodoURL(t *testing.T) {
	db := newTestDB(t)

	params := map[string]string{
		"title":   "nice_title",
		"list-id": "6c7e77b4-f4d7-44bc-8480-80c0bea585ea",
	}

	url := db.AddTodoURL(params)

	// Check it starts with correct command
	assert.True(t, strings.HasPrefix(url, "things:///add?"), "AddTodoURL() should have prefix things:///add?")

	// Check it contains the parameters
	assert.Contains(t, url, "title=nice_title")
	assert.Contains(t, url, "list-id=6c7e77b4-f4d7-44bc-8480-80c0bea585ea")
}

func TestAddProjectURL(t *testing.T) {
	db := newTestDB(t)

	params := map[string]string{
		"title": "Test Project",
	}

	url := db.AddProjectURL(params)

	assert.True(t, strings.HasPrefix(url, "things:///add-project?"), "AddProjectURL() should have prefix things:///add-project?")
}

func TestSearchURL(t *testing.T) {
	db := newTestDB(t)

	url := db.SearchURL("my query")

	assert.True(t, strings.HasPrefix(url, "things:///search?"), "SearchURL() should have prefix things:///search?")
	// Query could be encoded with + or %20
	containsEncodedQuery := strings.Contains(url, "query=my+query") || strings.Contains(url, "query=my%20query")
	assert.True(t, containsEncodedQuery, "SearchURL() should contain encoded query")
}

func TestURLEncoding(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	params := map[string]string{
		"title": "test task",
		"notes": "nice notes\nI really like notes",
	}

	url, err := db.URL(ctx, URLCommandAdd, params)
	require.NoError(t, err)

	// Check URL encoding - title could be encoded with + or %20
	titleEncoded := strings.Contains(url, "title=test+task") || strings.Contains(url, "title=test%20task")
	assert.True(t, titleEncoded, "URL() title not properly encoded: %q", url)

	// Newline should be encoded
	newlineEncoded := strings.Contains(url, "%0A") || strings.Contains(url, "%0a")
	assert.True(t, newlineEncoded, "URL() newline not properly encoded: %q", url)
}

func TestURLUpdate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Update command requires auth token
	url, err := db.URL(ctx, URLCommandUpdate, map[string]string{
		"id":        "test-uuid",
		"completed": "true",
	})
	require.NoError(t, err)

	// Should contain auth-token
	assert.Contains(t, url, "auth-token="+testAuthToken)
	assert.Contains(t, url, "id=test-uuid")
	assert.Contains(t, url, "completed=true")
}
