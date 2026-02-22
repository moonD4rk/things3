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
//	    tasks, err := client.Inbox(ctx)
//	    require.NoError(t, err)
//	    assert.Len(t, tasks, thingstest.Inbox)
//	}
package thingstest

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var (
	sourcePath string
	pathOnce   sync.Once
)

func initSourcePath() {
	pathOnce.Do(func() {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		sourcePath = filepath.Join(dir, "..", "testdata", "main.sqlite")
	})
}

// DatabasePath copies the things3 test fixture database to a temporary
// directory and returns the path to the copy. The temporary file is
// automatically cleaned up when the test finishes.
//
// Copying is necessary because the module cache is read-only and SQLite
// requires write access to the directory for lock files, even in read-only
// mode. Each test invocation gets its own copy, so parallel tests do not
// conflict.
func DatabasePath(t *testing.T) string {
	t.Helper()
	initSourcePath()

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("thingstest: read fixture database: %v", err)
	}

	dst := filepath.Join(t.TempDir(), "main.sqlite")
	if err := os.WriteFile(dst, data, 0o600); err != nil {
		t.Fatalf("thingstest: write fixture database: %v", err)
	}

	return dst
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
const (
	Headings                = 3
	Inbox                   = 2
	TrashedTodos            = 3
	TrashedProjects         = 1
	TrashedCanceled         = 1
	TrashedCompleted        = 1
	TrashedProjectTodos     = 1
	Trashed                 = 6
	Projects                = 5
	ProjectsNotTrashed      = 4
	Upcoming                = 1
	DeadlinePast            = 3
	DeadlineFuture          = 1
	Deadlines               = 4
	TodayProjects           = 1
	TodayTasks              = 4
	Today                   = 5
	Anytime                 = 15
	Logbook                 = 25
	Canceled                = 11
	Completed               = 14
	Someday                 = 2
	Tags                    = 5
	Areas                   = 3
	TodosIncomplete         = 15
	TodosAnytime            = 11
	TodosAnytimeComplete    = 8
	TodosComplete           = 12
	TasksIncomplete         = 22
	TasksIncompleteFiltered = 21
	TasksInProjectAll       = 7
	TasksInProject          = 5
	ProjectItems            = 7
	DatabaseVersion         = 24
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
	UUIDTaskCount        = "5pUx6PESj3ctFYbgth1PXY"
	UUIDCompletedProject = "CmpltdProjTestFixture01"
	UUIDSomedayProject   = "SmdyProjTestFixture001"
	UUIDCompletedHeading = "CmpltdHdngTestFixture1"
	UUIDEmptyHeading     = "AddtnlHdngTestFixture1"
)

// AuthToken is a fake authentication token for testing URL scheme operations.
const AuthToken = "vKkylosuSuGwxrz7qcklOw" //nolint:gosec // Test token, not a real credential
