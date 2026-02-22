package things3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteQuery_ContextCancellation(t *testing.T) {
	db := newTestDB(t)

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Query with canceled context should fail
	_, err := db.Todos(ctx)
	assert.Error(t, err, "query with canceled context should fail")
}

func TestExecuteQuery_EmptyResult(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Query with UUID that doesn't exist should return empty, not error
	tasks, err := db.Tasks().WithUUID("non-existent-uuid-12345").All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks, "non-existent UUID should return empty result")
}
