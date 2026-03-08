package things3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	idb "github.com/moond4rk/things3/internal/db"
)

func TestNewDB_WithValidPath(t *testing.T) {
	initTestPaths()

	database, err := newDB(idb.WithPath(testDatabasePath))
	require.NoError(t, err)
	require.NotNil(t, database)
	defer database.Close()

	assert.Equal(t, testDatabasePath, database.Filepath())
}

func TestNewDB_WithInvalidPath(t *testing.T) {
	database, err := newDB(idb.WithPath("/nonexistent/path.sqlite"))
	require.ErrorIs(t, err, ErrDatabaseNotFound)
	assert.Nil(t, database)
}

func TestNewDB_Close(t *testing.T) {
	initTestPaths()

	database, err := newDB(idb.WithPath(testDatabasePath))
	require.NoError(t, err)
	require.NotNil(t, database)

	// Close should not error
	err = database.Close()
	require.NoError(t, err)

	// Double close should not panic (SQLite handles this)
	err = database.Close()
	require.NoError(t, err)
}

func TestDB_Filepath(t *testing.T) {
	initTestPaths()

	database, err := newDB(idb.WithPath(testDatabasePath))
	require.NoError(t, err)
	defer database.Close()

	dbPath := database.Filepath()
	assert.Equal(t, testDatabasePath, dbPath)
}

func TestValidateDatabaseVersion_TooOld(t *testing.T) {
	initTestPaths()

	_, err := newTestDBOld(t)
	require.Error(t, err, "expected error for old database version")
	require.ErrorIs(t, err, ErrDatabaseVersionTooOld)
}
