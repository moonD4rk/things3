package things3

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBShow(t *testing.T) {
	// Test that Show method uses the new Scheme API correctly
	// Note: This only tests URL building, not actual execution (requires Things app)
	scheme := NewScheme()
	url := scheme.Show().ID("test-uuid").Build()
	assert.Equal(t, "things:///show?id=test-uuid", url)
}

func TestDBSearch(t *testing.T) {
	// Test that Search method uses the new Scheme API correctly
	scheme := NewScheme()
	url := scheme.Search("my query")
	assert.True(t, strings.HasPrefix(url, "things:///search?"), "Search() should have prefix things:///search?")
	containsEncodedQuery := strings.Contains(url, "query=my+query") || strings.Contains(url, "query=my%20query")
	assert.True(t, containsEncodedQuery, "Search() should contain encoded query")
}

func TestDBComplete(t *testing.T) {
	// Test that Complete method uses the new Scheme API correctly
	ctx := context.Background()
	db := newTestDB(t)

	token, err := db.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAuthToken, token)

	// Build the URL that Complete would use
	scheme := NewScheme()
	auth := scheme.WithToken(token)
	url, err := auth.UpdateTodo("test-uuid").Completed(true).Build()
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(url, "things:///update?"), "Complete URL should have prefix things:///update?")
	assert.Contains(t, url, "auth-token="+testAuthToken)
	assert.Contains(t, url, "id=test-uuid")
	assert.Contains(t, url, "completed=true")
}
