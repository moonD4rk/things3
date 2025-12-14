package things3

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverDatabasePath_CustomValid(t *testing.T) {
	initTestPaths()

	path, err := discoverDatabasePath(testDatabasePath)
	require.NoError(t, err)
	assert.Equal(t, testDatabasePath, path)
}

func TestDiscoverDatabasePath_CustomInvalid(t *testing.T) {
	path, err := discoverDatabasePath("/nonexistent/path/to/database.sqlite")
	require.ErrorIs(t, err, ErrDatabaseNotFound)
	assert.Empty(t, path)
}

func TestDiscoverDatabasePath_EnvVariable(t *testing.T) {
	initTestPaths()

	// Set environment variable
	t.Setenv(envDatabasePath, testDatabasePath)

	path, err := discoverDatabasePath("")
	require.NoError(t, err)
	assert.Equal(t, testDatabasePath, path)
}

func TestDiscoverDatabasePath_EnvVariableInvalid(t *testing.T) {
	// Set invalid environment variable
	t.Setenv(envDatabasePath, "/nonexistent/path.sqlite")

	path, err := discoverDatabasePath("")
	require.ErrorIs(t, err, ErrDatabaseNotFound)
	assert.Empty(t, path)
}

func TestExpandPath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"tilde only", "~", home},
		{"tilde with path", "~/Documents/test.db", home + "/Documents/test.db"},
		{"no tilde", "/absolute/path/test.db", "/absolute/path/test.db"},
		{"empty string", "", ""},
		{"relative path", "relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDatabaseVersion_IntegerFormat(t *testing.T) {
	db := newTestDB(t)

	version, err := getDatabaseVersion(db.db)
	require.NoError(t, err)
	assert.Equal(t, testDatabaseVersion, version)
}

func TestValidateDatabaseVersion_Valid(t *testing.T) {
	db := newTestDB(t)

	err := validateDatabaseVersion(db.db)
	assert.NoError(t, err)
}

func TestValidateDatabaseVersion_TooOld(t *testing.T) {
	initTestPaths()

	// Try to open old database - should fail at NewDB level
	_, err := newTestDBOld(t)
	require.Error(t, err, "expected error for old database version")
	require.ErrorIs(t, err, ErrDatabaseVersionTooOld)
}

func TestOpenDatabase_ReadOnlyMode(t *testing.T) {
	initTestPaths()

	db, err := openDatabase(testDatabasePath)
	require.NoError(t, err)
	defer db.Close()

	// Attempt to write - should fail because database is read-only
	_, err = db.ExecContext(context.Background(), "INSERT INTO TMTag (uuid, title) VALUES ('test-uuid', 'test-tag')")
	assert.Error(t, err, "write operation should fail on read-only database")
}

func TestOpenDatabase_InvalidPath(t *testing.T) {
	db, err := openDatabase("/nonexistent/path/database.sqlite")
	require.Error(t, err)
	assert.Nil(t, db)
}

func TestNewDB_WithValidPath(t *testing.T) {
	initTestPaths()

	db, err := NewDB(WithDBPath(testDatabasePath))
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	assert.Equal(t, testDatabasePath, db.Filepath())
}

func TestNewDB_WithInvalidPath(t *testing.T) {
	db, err := NewDB(WithDBPath("/nonexistent/path.sqlite"))
	require.ErrorIs(t, err, ErrDatabaseNotFound)
	assert.Nil(t, db)
}

func TestNewDB_Close(t *testing.T) {
	initTestPaths()

	db, err := NewDB(WithDBPath(testDatabasePath))
	require.NoError(t, err)
	require.NotNil(t, db)

	// Close should not error
	err = db.Close()
	require.NoError(t, err)

	// Double close should not panic (SQLite handles this)
	err = db.Close()
	require.NoError(t, err)
}

func TestDB_Filepath(t *testing.T) {
	initTestPaths()

	db, err := NewDB(WithDBPath(testDatabasePath))
	require.NoError(t, err)
	defer db.Close()

	dbPath := db.Filepath()
	assert.Equal(t, testDatabasePath, dbPath)
}
