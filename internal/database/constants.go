package database

// Database table names.
const (
	tableTask          = "TMTask"
	tableArea          = "TMArea"
	tableTag           = "TMTag"
	tableChecklistItem = "TMChecklistItem"
	tableTaskTag       = "TMTaskTag"
	tableAreaTag       = "TMAreaTag"
	tableSettings      = "TMSettings"
	tableMeta          = "Meta"
)

// Date column names in the database.
const (
	colCreationDate     = "creationDate"
	colModificationDate = "userModificationDate"
	colStopDate         = "stopDate"
	colDeadline         = "deadline"
	colStartDate        = "startDate"
	colReminderTime     = "reminderTime"
	// colNextInstanceStartDate holds a repeating template's next occurrence in
	// Things date format; for a template it plays the role of startDate.
	colNextInstanceStartDate = "rt1_nextInstanceStartDate"
)

// Filter SQL expressions.
const (
	// Type filters
	filterIsTodo    = "type = 0"
	filterIsProject = "type = 1"
	filterIsHeading = "type = 2"

	// Status filters
	filterIsIncomplete = "status = 0"
	filterIsCanceled   = "status = 2"
	filterIsCompleted  = "status = 3"

	// Start bucket filters
	filterIsInbox   = "start = 0"
	filterIsAnytime = "start = 1"
	filterIsSomeday = "start = 2"

	// Trash filters
	filterIsTrashed    = "trashed = 1"
	filterIsNotTrashed = "trashed = 0"

	// Recurring filters
	filterIsNotRecurring = "rt1_recurrenceRule IS NULL"
	filterIsRecurring    = "rt1_recurrenceRule IS NOT NULL"
)

// startBucketEvening is the TMTask.startBucket value marking a todo as being in
// the This Evening section of Today. Confirmed on a live database (schema v26)
// by dragging a todo into This Evening and observing startBucket flip 0 -> 1.
const startBucketEvening int64 = 1

// Settings UUID for auth token.
const settingsUUID = "RhAzEf6qDxCD5PmnZVtBZR"

// minDatabaseVersion is the minimum supported database version.
const minDatabaseVersion = 21

// EnvDatabasePath is the environment variable name for custom database path.
const EnvDatabasePath = "THINGSDB"

// Index column names for ordering.
const (
	// IndexDefault is the default ordering column.
	IndexDefault = "index"
	// IndexToday is the Today view ordering column.
	IndexToday = "todayIndex"
)
