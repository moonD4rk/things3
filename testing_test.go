package things3

import (
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// Test database paths.
var (
	testDatabasePath    string
	testDatabasePathOld string
	testPathOnce        sync.Once
)

func initTestPaths() {
	testPathOnce.Do(func() {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		testDatabasePath = filepath.Join(dir, "testdata", "main.sqlite")
		testDatabasePathOld = filepath.Join(dir, "testdata", "db2022", "main.sqlite")
	})
}

// Test expected counts from Python test suite.
// Note: Some counts may differ slightly from Python due to database state.
const (
	testHeadings                = 3
	testInbox                   = 2
	testTrashedTodos            = 3 // Includes context-trashed items
	testTrashedProjects         = 1
	testTrashedCanceled         = 1
	testTrashedCompleted        = 1
	testTrashedProjectTodos     = 1
	testTrashed                 = 6 // Total trashed items
	testProjects                = 4
	testProjectsNotTrashed      = 3
	testUpcoming                = 1
	testDeadlinePast            = 3
	testDeadlineFuture          = 1
	testDeadlines               = 4 // testDeadlinePast + testDeadlineFuture
	testTodayProjects           = 1
	testTodayTasks              = 4
	testToday                   = 5  // testTodayProjects + testTodayTasks
	testAnytime                 = 14 // Excludes context-trashed tasks (was 15)
	testLogbook                 = 23
	testCanceled                = 11
	testCompleted               = 12
	testSomeday                 = 1
	testTags                    = 5
	testAreas                   = 3
	testTodosIncomplete         = 15 // Convenience method with ContextTrashed(false)
	testTodosAnytime            = 11 // Builder default (includes context-trashed tasks)
	testTodosAnytimeComplete    = 8
	testTodosComplete           = 12
	testTasksIncomplete         = 20 // Builder default (includes context-trashed)
	testTasksIncompleteFiltered = 19 // Convenience method with ContextTrashed(false)
	testTasksInProjectAll       = 7  // All tasks in test project (including completed/canceled)
	testTasksInProject          = 5  // Incomplete tasks in test project
	testProjectItems            = 7  // Items in first project
	testDatabaseVersion         = 24
	testAuthToken               = "vKkylosuSuGwxrz7qcklOw" //nolint:gosec // Test token, not a real credential
)

// Test UUIDs from the test database.
const (
	testUUIDTag             = "Qt2AY87x2QDdowSn9HKTt1"
	testUUIDTodoChecklist   = "3Eva4XFof6zWb9iSfYy4ej"
	testUUIDTodoNoChecklist = "K9bx7h1xCJdevvyWardZDq"
	testUUIDArea            = "Y3JC4XeyGWxzDocQL4aobo"
	testUUIDProject         = "3x1QqJqfvZyhtw8NSdnZqG"
	testUUIDTodo            = "A2oPvtt4dXoypeoLc8uYzY"
	testUUIDTodoReminder    = "7F4vqUNiTvGKaCUfv5pqYG"
	testUUIDTaskCount       = "5pUx6PESj3ctFYbgth1PXY"
)

// newTestDB creates a new db connected to the test database.
func newTestDB(t *testing.T) *db {
	t.Helper()
	initTestPaths()
	database, err := newDB(withDBPath(testDatabasePath))
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

// newTestDBOld creates a new db connected to the old version test database.
// This should fail with ErrDatabaseVersionTooOld.
func newTestDBOld(t *testing.T) (*db, error) {
	t.Helper()
	initTestPaths()
	return newDB(withDBPath(testDatabasePathOld))
}
