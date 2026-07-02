// Package thingstest provides test utilities for projects that depend on the
// things3 library. It exposes a fixture database and well-known constants so
// downstream tests can reuse the same test data without copying files.
//
// The fixture database is shipped inside the things3 module's testdata/ directory.
// [DatabasePath] copies it to a temporary directory owned by the test, so SQLite
// can open it without write-permission issues in the read-only module cache.
//
// Example:
//
//	func TestMyServer(t *testing.T) {
//	    client, err := things3.NewClient(things3.WithDatabasePath(thingstest.DatabasePath(t)))
//	    require.NoError(t, err)
//	    defer client.Close()
//
//	    todos, err := client.Todos().Start().Inbox().Status().Incomplete().All(ctx)
//	    require.NoError(t, err)
//	    assert.Len(t, todos, thingstest.Inbox)
//	}
package thingstest

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// SQLite WAL-mode sidecar file suffixes.
const (
	walSuffix = "-wal"
	shmSuffix = "-shm"
)

var (
	sourcePath     string
	initSourcePath = sync.OnceFunc(func() {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		sourcePath = filepath.Join(dir, "..", "testdata", "main.sqlite")
	})
)

// DatabasePath copies the things3 test fixture database to a temporary
// directory and returns the path to the copy. The temporary files are
// automatically cleaned up when the test finishes.
//
// Copying is necessary because the module cache is read-only and SQLite
// requires write access to the directory for lock files, even in read-only
// mode. Each test invocation gets its own copy, so parallel tests do not
// conflict.
//
// The fixture is in WAL mode, so the -wal and -shm sidecar files are copied
// alongside the database when they exist; otherwise uncheckpointed data would
// be silently lost.
func DatabasePath(t *testing.T) string {
	t.Helper()
	initSourcePath()

	return copyDatabaseWithSidecars(t, sourcePath, t.TempDir())
}

// copyDatabaseWithSidecars copies the database at src plus any -wal/-shm
// sidecar files into dstDir and returns the path of the copied database.
func copyDatabaseWithSidecars(t *testing.T, src, dstDir string) string {
	t.Helper()

	dst := filepath.Join(dstDir, filepath.Base(src))
	copyFixtureFile(t, src, dst)

	for _, suffix := range []string{walSuffix, shmSuffix} {
		sidecar := src + suffix
		if _, err := os.Stat(sidecar); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("thingstest: stat fixture sidecar: %v", err)
		}
		copyFixtureFile(t, sidecar, dst+suffix)
	}

	return dst
}

// copyFixtureFile copies a fixture file from src to dst, failing the test on
// any error.
func copyFixtureFile(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("thingstest: read fixture file: %v", err)
	}
	if err := os.WriteFile(dst, data, 0o600); err != nil { //nolint:gosec // dst is from t.TempDir(), not user input
		t.Fatalf("thingstest: write fixture file: %v", err)
	}
}

// SourceDatabasePath returns the absolute path to the original fixture
// database inside the module source tree. This path may reside in the
// read-only module cache. Use [DatabasePath] instead when a writable copy
// is needed (which is the common case for SQLite).
func SourceDatabasePath() string {
	initSourcePath()
	return sourcePath
}

// Expected item counts from the test fixture database.
// These mirror the internal test constants and can be used to verify query results.
//
// Counts are specific to each domain type (Todo, Project, Heading).
// There is no unified "Task" count -- use the typed constants instead.
const (
	// Todos (type=0)
	TodosIncomplete      = 15
	TodosAnytime         = 10
	TodosAnytimeComplete = 8
	TodosComplete        = 12
	TodosInProject       = 4
	Inbox                = 2

	// Projects (type=1)
	Projects           = 5
	ProjectsNotTrashed = 4

	// Headings (type=2)
	Headings = 3

	// Areas
	Areas = 3

	// Tags
	Tags = 5

	// Dates
	DeadlinePast   = 3
	DeadlineFuture = 1
	Deadlines      = 4

	// Trashed
	TrashedTodos        = 3
	TrashedProjects     = 1
	TrashedCanceled     = 1
	TrashedCompleted    = 1
	TrashedProjectTodos = 1
	Trashed             = 6

	// Database
	DatabaseVersion = 24
)

// Well-known UUIDs from the test fixture database.
const (
	UUIDTag              = "Qt2AY87x2QDdowSn9HKTt1"
	UUIDTodoChecklist    = "3Eva4XFof6zWb9iSfYy4ej"
	UUIDTodoNoChecklist  = "K9bx7h1xCJdevvyWardZDq"
	UUIDArea             = "Y3JC4XeyGWxzDocQL4aobo"
	UUIDProject          = "3x1QqJqfvZyhtw8NSdnZqG"
	UUIDTodo             = "A2oPvtt4dXoypeoLc8uYzY"
	UUIDTodoReminder     = "7F4vqUNiTvGKaCUfv5pqYG"
	UUIDTodoInToday      = "5pUx6PESj3ctFYbgth1PXY"
	UUIDCompletedProject = "CmpltdProjTestFixture01"
	UUIDSomedayProject   = "SmdyProjTestFixture001"
	UUIDCompletedHeading = "CmpltdHdngTestFixture1"
	UUIDEmptyHeading     = "AddtnlHdngTestFixture1"
)

// AuthToken is a fake authentication token for testing URL scheme operations.
const AuthToken = "vKkylosuSuGwxrz7qcklOw" //nolint:gosec // Test token, not a real credential
