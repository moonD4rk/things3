package things3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moond4rk/things3/internal/database"
)

func TestNewDB_WithValidPath(t *testing.T) {
	initTestPaths()

	d, err := newDB(database.WithPath(testDatabasePath))
	require.NoError(t, err)
	require.NotNil(t, d)
	defer d.Close()

	assert.Equal(t, testDatabasePath, d.Filepath())
}

func TestNewDB_WithInvalidPath(t *testing.T) {
	d, err := newDB(database.WithPath("/nonexistent/path.sqlite"))
	require.ErrorIs(t, err, ErrDatabaseNotFound)
	assert.Nil(t, d)
}

func TestNewDB_Close(t *testing.T) {
	initTestPaths()

	d, err := newDB(database.WithPath(testDatabasePath))
	require.NoError(t, err)
	require.NotNil(t, d)

	// Close should not error
	err = d.Close()
	require.NoError(t, err)

	// Double close should not panic (SQLite handles this)
	err = d.Close()
	require.NoError(t, err)
}

func TestDB_Filepath(t *testing.T) {
	initTestPaths()

	d, err := newDB(database.WithPath(testDatabasePath))
	require.NoError(t, err)
	defer d.Close()

	dbPath := d.Filepath()
	assert.Equal(t, testDatabasePath, dbPath)
}

func TestValidateDatabaseVersion_TooOld(t *testing.T) {
	initTestPaths()

	_, err := newTestDBOld(t)
	require.Error(t, err, "expected error for old database version")
	require.ErrorIs(t, err, ErrDatabaseVersionTooOld)
}

func TestExecuteQuery_ContextCancellation(t *testing.T) {
	db := newTestDB(t)

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Query with canceled context should fail
	_, err := db.Todos().
		Status().Incomplete().
		All(ctx)
	assert.Error(t, err, "query with canceled context should fail")
}

func TestExecuteQuery_EmptyResult(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Query with UUID that doesn't exist should return empty, not error
	todos, err := db.Todos().WithUUID("non-existent-uuid-12345").All(ctx)
	require.NoError(t, err)
	assert.Empty(t, todos, "non-existent UUID should return empty result")
}
