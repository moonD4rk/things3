package things3

// Database table names.
const (
	TableTask          = "TMTask"
	TableArea          = "TMArea"
	TableTag           = "TMTag"
	TableChecklistItem = "TMChecklistItem"
	TableTaskTag       = "TMTaskTag"
	TableAreaTag       = "TMAreaTag"
	TableSettings      = "TMSettings"
	TableMeta          = "Meta"
)

// Date column names in the database.
const (
	// ColCreationDate stores Unix timestamp (UTC) for creation date.
	ColCreationDate = "creationDate"
	// ColModificationDate stores Unix timestamp (UTC) for modification date.
	ColModificationDate = "userModificationDate"
	// ColStopDate stores Unix timestamp (UTC) for completion/cancellation date.
	ColStopDate = "stopDate"
	// ColDeadline stores Things date format (YYYYYYYYYYYMMMMDDDDD0000000).
	ColDeadline = "deadline"
	// ColStartDate stores Things date format (YYYYYYYYYYYMMMMDDDDD0000000).
	ColStartDate = "startDate"
	// ColReminderTime stores Things time format (hhhhhmmmmmm00000000000000000000).
	ColReminderTime = "reminderTime"
)

// Filter SQL expressions.
const (
	// Type filters
	FilterIsTodo    = "type = 0"
	FilterIsProject = "type = 1"
	FilterIsHeading = "type = 2"

	// Status filters
	FilterIsIncomplete = "status = 0"
	FilterIsCanceled   = "status = 2"
	FilterIsCompleted  = "status = 3"

	// Start bucket filters
	FilterIsInbox   = "start = 0"
	FilterIsAnytime = "start = 1"
	FilterIsSomeday = "start = 2"

	// Trash filters
	FilterIsTrashed    = "trashed = 1"
	FilterIsNotTrashed = "trashed = 0"

	// Recurring filters
	FilterIsNotRecurring = "rt1_recurrenceRule IS NULL"
)

// Settings UUID for auth token.
const SettingsUUID = "RhAzEf6qDxCD5PmnZVtBZR"

// Minimum supported database version.
const MinDatabaseVersion = 21

// Environment variable name for custom database path.
const EnvDatabasePath = "THINGSDB"

// Index column names for ordering.
const (
	IndexDefault = "index"
	IndexToday   = "todayIndex"
)
