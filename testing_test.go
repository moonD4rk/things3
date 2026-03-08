package things3

import (
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/moond4rk/things3/internal/database"
)

// Test database paths.
var (
	testDatabasePath    string
	testDatabasePathOld string
	initTestPaths       = sync.OnceFunc(func() {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		testDatabasePath = filepath.Join(dir, "testdata", "main.sqlite")
		testDatabasePathOld = filepath.Join(dir, "testdata", "db2022", "main.sqlite")
	})
)

// Test expected counts from the fixture database.
const (
	testDeadlinePast    = 3
	testDeadlineFuture  = 1
	testDeadlines       = 4 // testDeadlinePast + testDeadlineFuture
	testAreas           = 3
	testTodosIncomplete = 15                       // Incomplete todos (type=0) excluding parent-trashed
	testTodosInProject  = 4                        // Incomplete todos (type=0) in test project
	testAuthToken       = "vKkylosuSuGwxrz7qcklOw" //nolint:gosec // Test token, not a real credential
)

// =============================================================================
// Test UUIDs from the fixture database
// =============================================================================

// Incomplete Todos - Inbox (start=0).
const (
	testUUIDTodoInbox          = "DfYoiXcNLQssk9DkSoJV3Y"
	testUUIDTodoInboxChecklist = "3Eva4XFof6zWb9iSfYy4ej"
)

// Incomplete Todos - Anytime (start=1).
const (
	testUUIDTodoInToday          = "5pUx6PESj3ctFYbgth1PXY"
	testUUIDTodoAnytime          = "QqhVksfbsAVaNnwB1x3CuD"
	testUUIDTodoInProject        = "E18tg5qepzrQk9J6jQtb5C"
	testUUIDTodoInArea1          = "Q7uN9y3jp5ChZAGjZJhMfY"
	testUUIDTodoInHeading        = "HbKGAeZKFDkWH5osSBNHvz"
	testUUIDTodoRepeating        = "K9bx7h1xCJdevvyWardZDq"
	testUUIDTodoInArea3          = "EJxkdyCLyjJx6wucDPUcvu"
	testUUIDTodoInArea1Tags      = "W5JYfjY2xtLdmedQKU6caM"
	testUUIDTodoInDeletedProject = "NoQLFamrMMooAELuBznao8" // context-trashed
	testUUIDTodoOverdueInToday   = "KisAmSsnzCcRRumjY4TkVV"
	testUUIDTodoOverdueNotToday  = "Cc73oaq1C2mDMpZZUJaBxe"
)

// Projects (type=1).
const testUUIDProjectInArea1 = "3x1QqJqfvZyhtw8NSdnZqG"

// Areas.
const (
	testUUIDArea1 = "DciSFacytdrNG1nRaMJPgY"
	testUUIDArea2 = "3UXZmXt9qNMTWL5iZNyrxj"
	testUUIDArea3 = "Y3JC4XeyGWxzDocQL4aobo"
)

// Tags.
const (
	testUUIDTagErrand    = "H96sVJwE7VJveAnv7itmux"
	testUUIDTagHome      = "CK9dARrf2ezbFvrVUUxkHE"
	testUUIDTagOffice    = "Qt2AY87x2QDdowSn9HKTt1"
	testUUIDTagImportant = "XdDBCjmEXEhjZy9A2wFFKP"
	testUUIDTagPending   = "BULfa35PCAn1LtsmBA6A2u"
)

// =============================================================================
// Expected UUID Sets
// =============================================================================

var (
	testTodosAnytimeUUIDs = []string{
		testUUIDTodoInToday, testUUIDTodoAnytime, testUUIDTodoInProject,
		testUUIDTodoInArea1, testUUIDTodoInHeading, testUUIDTodoRepeating,
		testUUIDTodoInArea3, testUUIDTodoInArea1Tags,
		testUUIDTodoOverdueInToday, testUUIDTodoOverdueNotToday,
	}

	testAreaUUIDs = []string{testUUIDArea1, testUUIDArea2, testUUIDArea3}

	testTagUUIDs = []string{
		testUUIDTagErrand, testUUIDTagHome, testUUIDTagOffice,
		testUUIDTagImportant, testUUIDTagPending,
	}
)

// =============================================================================
// UUID Extraction Helpers
// =============================================================================

func extractTodoUUIDs(todos []Todo) []string {
	uuids := make([]string, len(todos))
	for i := range todos {
		uuids[i] = todos[i].UUID
	}
	return uuids
}

func extractAreaUUIDs(areas []Area) []string {
	uuids := make([]string, len(areas))
	for i, a := range areas {
		uuids[i] = a.UUID
	}
	return uuids
}

func extractTagUUIDs(tags []Tag) []string {
	uuids := make([]string, len(tags))
	for i, t := range tags {
		uuids[i] = t.UUID
	}
	return uuids
}

// =============================================================================
// Test DB Helpers
// =============================================================================

// newTestDB creates a new db connected to the test database.
func newTestDB(t *testing.T) *db {
	t.Helper()
	initTestPaths()
	d, err := newDB(database.WithPath(testDatabasePath))
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

// newTestDBOld creates a new db connected to the old version test database.
// This should fail with ErrDatabaseVersionTooOld.
func newTestDBOld(t *testing.T) (*db, error) {
	t.Helper()
	initTestPaths()
	return newDB(database.WithPath(testDatabasePathOld))
}
