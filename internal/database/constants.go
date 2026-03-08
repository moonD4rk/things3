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
)

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
