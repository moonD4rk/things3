package thingstest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabasePath(t *testing.T) {
	path := DatabasePath(t)
	require.NotEmpty(t, path)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.False(t, info.IsDir())
	assert.Positive(t, info.Size())
}

func TestDatabasePathWritable(t *testing.T) {
	path := DatabasePath(t)

	// The copy should be in a writable temp directory.
	f, err := os.OpenFile(path, os.O_RDWR, 0o600)
	require.NoError(t, err)
	f.Close()
}

func TestDatabasePathIsolation(t *testing.T) {
	path1 := DatabasePath(t)
	path2 := DatabasePath(t)
	assert.NotEqual(t, path1, path2, "each call should return an independent copy")
}

func TestSourceDatabasePath(t *testing.T) {
	path := SourceDatabasePath()
	require.NotEmpty(t, path)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Positive(t, info.Size())
}
