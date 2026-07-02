package thingstest

import (
	"os"
	"path/filepath"
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

// TestCopyDatabaseWithSidecars verifies that the -wal and -shm sidecar files
// are copied along with the main database. The fixture is in WAL mode, so
// dropping the sidecars would silently lose any uncheckpointed data. The test
// uses its own source files because the shared fixture's sidecars are live
// SQLite artifacts whose size changes while other tests hold the database open.
func TestCopyDatabaseWithSidecars(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "main.sqlite")
	require.NoError(t, os.WriteFile(src, []byte("database bytes"), 0o600))
	require.NoError(t, os.WriteFile(src+walSuffix, []byte("wal bytes"), 0o600))
	require.NoError(t, os.WriteFile(src+shmSuffix, []byte("shm bytes"), 0o600))

	dst := copyDatabaseWithSidecars(t, src, t.TempDir())

	for _, suffix := range []string{"", walSuffix, shmSuffix} {
		want, err := os.ReadFile(src + suffix)
		require.NoError(t, err)
		got, err := os.ReadFile(dst + suffix)
		require.NoError(t, err, "sidecar %q must be copied next to the database", suffix)
		assert.Equal(t, want, got, "sidecar %q content must match the source", suffix)
	}
}

// TestCopyDatabaseWithSidecarsWithoutSidecars verifies that missing sidecar
// files are tolerated rather than treated as an error.
func TestCopyDatabaseWithSidecarsWithoutSidecars(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "main.sqlite")
	require.NoError(t, os.WriteFile(src, []byte("database bytes"), 0o600))

	dst := copyDatabaseWithSidecars(t, src, t.TempDir())

	_, err := os.Stat(dst)
	require.NoError(t, err)
	for _, suffix := range []string{walSuffix, shmSuffix} {
		_, err := os.Stat(dst + suffix)
		assert.True(t, os.IsNotExist(err), "sidecar %q must not be fabricated", suffix)
	}
}
