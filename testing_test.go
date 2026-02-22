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
	testTrashedTodos            = 3 // Includes context-trashed items
	testTrashedProjects         = 1
	testTrashedCanceled         = 1
	testTrashedCompleted        = 1
	testTrashedProjectTodos     = 1
	testTrashed                 = 6 // Total trashed items
	testProjectsNotTrashed      = 4
	testUpcoming                = 1
	testDeadlinePast            = 3
	testDeadlineFuture          = 1
	testDeadlines               = 4 // testDeadlinePast + testDeadlineFuture
	testTodayProjects           = 1
	testTodayTasks              = 4
	testToday                   = 5  // testTodayProjects + testTodayTasks
	testAnytime                 = 15 // Excludes context-trashed tasks
	testLogbook                 = 25
	testCanceled                = 11
	testCompleted               = 14
	testSomeday                 = 2
	testTags                    = 5
	testAreas                   = 3
	testTodosIncomplete         = 15 // Convenience method with ContextTrashed(false)
	testTodosAnytime            = 11 // Builder default (includes context-trashed tasks)
	testTodosAnytimeComplete    = 8
	testTodosComplete           = 12
	testTasksIncomplete         = 22 // Builder default (includes context-trashed)
	testTasksIncompleteFiltered = 21 // Convenience method with ContextTrashed(false)
	testTasksInProjectAll       = 7  // All tasks in test project (including completed/canceled)
	testTasksInProject          = 5  // Incomplete tasks in test project
	testProjectItems            = 7  // Items in first project
	testDatabaseVersion         = 24
	testAuthToken               = "vKkylosuSuGwxrz7qcklOw" //nolint:gosec // Test token, not a real credential
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
	testUUIDTodoDeleted          = "A2oPvtt4dXoypeoLc8uYzY" // trashed
	testUUIDTodoDeletedInProject = "LeMGbV4SZyMgSRTCWq72Yg" // trashed
	testUUIDTodoAnotherDeleted   = "Tmt3omxtu61wAoTaSLweQy" // trashed
)

// Incomplete Todos - Someday (start=2).
const (
	testUUIDTodoUpcoming      = "7F4vqUNiTvGKaCUfv5pqYG"
	testUUIDTodoSomeday       = "JLYSEPFkLfBC5rhGJRa5S1"
	testUUIDTodoRepeatingTmpl = "N1PJHsbjct4mb1bhcs7aHa" // template, excluded from normal queries
	testUUIDTodoUpcomingToday = "6Hf2qWBjWhq7B1xszwdo34" // yellow dot
)

// Canceled Todos (status=2).
const (
	testUUIDCanceledInbox      = "9DyzgLkZf1cBDbJ2dYFGBR"
	testUUIDCanceledToday      = "SzgXfYgNV4kWp5anvjsdJT"
	testUUIDCanceledAnytime    = "Ak7cN3VDSnpW6MQt7tf4cd"
	testUUIDCanceledInProject  = "5HLnvorXMbqcbjUuPN6ywi"
	testUUIDCanceledInArea1    = "BWzcy7ZSQ6T48AX8vsaPC8"
	testUUIDCanceledInHeading  = "RqRi38gMxTFyhPh2X1vH1i"
	testUUIDCanceledDeleted    = "8xADVLww2JjHu3BWV4iK2U" // trashed
	testUUIDCanceledInCancProj = "NsEyVWNres9441aCBtz9bF"
	testUUIDCanceledUpcoming   = "ADLex1EmJzLpu2GHxFvLvc"
	testUUIDCanceledLogbook    = "SuSafUtGHGKatpo3rqUdsh"
	testUUIDCanceledSomeday    = "DkVUPkCVM9mNq8yQuLrDo"
)

// Completed Todos (status=3).
const (
	testUUIDCompletedInbox      = "LgqUAQAdNsS3CGHok4EjLa"
	testUUIDCompletedToday      = "56dtXSk3A373M6n4eqGyr3"
	testUUIDCompletedAnytime    = "NSzDo18ibpJ1H8xStXLvto"
	testUUIDCompletedDeleted    = "Eb85bTZUJWjJe8ppmaRzjN" // trashed
	testUUIDCompletedInProject  = "5u2yGhP4rMQUmPQYEpGYDd"
	testUUIDCompletedInArea1    = "UwNEL2WdQTd92ZLa2HkHnc"
	testUUIDCompletedInHeading  = "2qBNNhNuDUBEGcB2tVRH9W"
	testUUIDCompletedInCancProj = "S8QU6gEvQec7XRMkN5Vjwg"
	testUUIDCompletedBeforeUTC  = "LnGwkFDZw78ydwp98jqo3z"
	testUUIDCompletedAfterUTC   = "JM91cry5BMFP7R3vXDns9z"
	testUUIDCompletedUpcoming   = "LE2WEGxANmtHWD3c9g5iWA"
	testUUIDCompletedLogbook    = "6gM3LexGhMGawEjGmKm3Z4"
	testUUIDCompletedSomeday    = "WQ8p2mhuHWd7g9tMJfed2W"
)

// Projects (type=1).
const (
	testUUIDProjectNoArea    = "TCozQqXVbB2TJkXXXQj2H9"
	testUUIDProjectInArea1   = "3x1QqJqfvZyhtw8NSdnZqG"
	testUUIDProjectInToday   = "PgsWnDkzXRz6zvofTqtHqn"
	testUUIDProjectDeleted   = "Tc7DABDNNMZvV4ZGB8tLDh" // trashed
	testUUIDProjectCanceled  = "SkLdfSe1MXR5vMV1gMYkHE"
	testUUIDCompletedProject = "CmpltdProjTestFixture01"
	testUUIDSomedayProject   = "SmdyProjTestFixture001"
)

// Headings (type=2).
const (
	testUUIDHeading          = "6QpDLSHZMRAUSAeZ9mNvgt"
	testUUIDCompletedHeading = "CmpltdHdngTestFixture1"
	testUUIDEmptyHeading     = "AddtnlHdngTestFixture1"
)

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

// Checklist items.
const (
	testUUIDChecklistItem1 = "Ka8uwUstDgQWkugYyVHB1a"
	testUUIDChecklistItem2 = "UR9qjvuykBsv2dp8yPzWGT"
	testUUIDChecklistItem3 = "XufyKEcAa9vAUxiJuwChK" // completed
)

// =============================================================================
// Expected UUID Sets
// =============================================================================

var (
	testInboxUUIDs = []string{
		testUUIDTodoInbox, testUUIDTodoInboxChecklist,
	}

	testSomedayUUIDs = []string{
		testUUIDTodoSomeday, testUUIDSomedayProject,
	}

	testCompletedUUIDs = []string{
		testUUIDCompletedInbox, testUUIDCompletedToday, testUUIDCompletedAnytime,
		testUUIDCompletedInProject, testUUIDCompletedInArea1, testUUIDCompletedInHeading,
		testUUIDCompletedInCancProj, testUUIDCompletedBeforeUTC, testUUIDCompletedAfterUTC,
		testUUIDCompletedUpcoming, testUUIDCompletedLogbook, testUUIDCompletedSomeday,
		testUUIDCompletedProject, testUUIDCompletedHeading,
	}

	testCanceledUUIDs = []string{
		testUUIDCanceledInbox, testUUIDCanceledToday, testUUIDCanceledAnytime,
		testUUIDCanceledInProject, testUUIDCanceledInArea1, testUUIDCanceledInHeading,
		testUUIDCanceledInCancProj, testUUIDCanceledUpcoming, testUUIDCanceledLogbook,
		testUUIDCanceledSomeday, testUUIDProjectCanceled,
	}

	testAnytimeUUIDs = []string{
		testUUIDTodoInToday, testUUIDTodoAnytime, testUUIDTodoInProject,
		testUUIDTodoInArea1, testUUIDTodoInHeading, testUUIDTodoRepeating,
		testUUIDTodoInArea3, testUUIDTodoInArea1Tags, testUUIDTodoOverdueInToday,
		testUUIDTodoOverdueNotToday, testUUIDProjectInArea1, testUUIDProjectInToday,
		testUUIDProjectNoArea, testUUIDHeading, testUUIDEmptyHeading,
	}

	testTrashedUUIDs = []string{
		testUUIDTodoDeleted, testUUIDTodoDeletedInProject, testUUIDTodoAnotherDeleted,
		testUUIDCanceledDeleted, testUUIDCompletedDeleted, testUUIDProjectDeleted,
	}

	testTrashedTodoUUIDs = []string{
		testUUIDTodoDeleted, testUUIDTodoAnotherDeleted, testUUIDTodoDeletedInProject,
	}

	testTrashedProjectUUIDs = []string{testUUIDProjectDeleted}

	testTodosIncompleteUUIDs = []string{
		testUUIDTodoInbox, testUUIDTodoInboxChecklist, testUUIDTodoInToday,
		testUUIDTodoAnytime, testUUIDTodoInProject, testUUIDTodoInArea1,
		testUUIDTodoInHeading, testUUIDTodoRepeating, testUUIDTodoInArea3,
		testUUIDTodoInArea1Tags, testUUIDTodoOverdueInToday, testUUIDTodoOverdueNotToday,
		testUUIDTodoUpcoming, testUUIDTodoSomeday, testUUIDTodoUpcomingToday,
	}

	testTodosAnytimeUUIDs = []string{
		testUUIDTodoInToday, testUUIDTodoAnytime, testUUIDTodoInProject,
		testUUIDTodoInArea1, testUUIDTodoInHeading, testUUIDTodoRepeating,
		testUUIDTodoInArea3, testUUIDTodoInArea1Tags, testUUIDTodoInDeletedProject,
		testUUIDTodoOverdueInToday, testUUIDTodoOverdueNotToday,
	}

	testTodosAnytimeCompleteUUIDs = []string{
		testUUIDCompletedToday, testUUIDCompletedAnytime, testUUIDCompletedInProject,
		testUUIDCompletedInArea1, testUUIDCompletedInHeading, testUUIDCompletedInCancProj,
		testUUIDCompletedBeforeUTC, testUUIDCompletedAfterUTC,
	}

	testTodosCompleteUUIDs = []string{
		testUUIDCompletedInbox, testUUIDCompletedToday, testUUIDCompletedAnytime,
		testUUIDCompletedInProject, testUUIDCompletedInArea1, testUUIDCompletedInHeading,
		testUUIDCompletedInCancProj, testUUIDCompletedBeforeUTC, testUUIDCompletedAfterUTC,
		testUUIDCompletedUpcoming, testUUIDCompletedLogbook, testUUIDCompletedSomeday,
	}

	testProjectsNotTrashedUUIDs = []string{
		testUUIDProjectNoArea, testUUIDProjectInArea1,
		testUUIDProjectInToday, testUUIDSomedayProject,
	}

	testAreaUUIDs = []string{testUUIDArea1, testUUIDArea2, testUUIDArea3}

	testTagUUIDs = []string{
		testUUIDTagErrand, testUUIDTagHome, testUUIDTagOffice,
		testUUIDTagImportant, testUUIDTagPending,
	}

	testChecklistUUIDs = []string{
		testUUIDChecklistItem1, testUUIDChecklistItem2, testUUIDChecklistItem3,
	}

	testUpcomingUUIDs = []string{testUUIDTodoUpcoming}

	testDeadlineUUIDs = []string{
		testUUIDTodoRepeating, testUUIDTodoOverdueInToday,
		testUUIDTodoOverdueNotToday, testUUIDTodoInHeading,
	}

	testDeadlinePastUUIDs = []string{
		testUUIDTodoRepeating, testUUIDTodoOverdueInToday, testUUIDTodoOverdueNotToday,
	}

	testDeadlineFutureUUIDs = []string{testUUIDTodoInHeading}
)

// =============================================================================
// UUID Extraction Helpers
// =============================================================================

func extractUUIDs(tasks []Task) []string {
	uuids := make([]string, len(tasks))
	for i := range tasks {
		uuids[i] = tasks[i].UUID
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

func extractChecklistUUIDs(items []ChecklistItem) []string {
	uuids := make([]string, len(items))
	for i, item := range items {
		uuids[i] = item.UUID
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
