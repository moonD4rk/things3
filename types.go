package things3

// String constant for unknown enum values.
const unknownString = "unknown"

// TaskType represents the kind of task in Things 3.
// Tasks can be to-dos, projects, or headings within projects.
type TaskType int

const (
	// TaskTypeTodo represents a regular to-do item.
	TaskTypeTodo TaskType = 0
	// TaskTypeProject represents a project containing tasks.
	TaskTypeProject TaskType = 1
	// TaskTypeHeading represents a heading within a project.
	TaskTypeHeading TaskType = 2
)

// String returns the string representation of the TaskType.
func (t TaskType) String() string {
	switch t {
	case TaskTypeTodo:
		return "to-do"
	case TaskTypeProject:
		return "project"
	case TaskTypeHeading:
		return "heading"
	default:
		return unknownString
	}
}

// Status represents the completion status of a task.
type Status int

const (
	// StatusIncomplete indicates the task is not yet completed.
	StatusIncomplete Status = 0
	// StatusCanceled indicates the task was canceled.
	StatusCanceled Status = 2
	// StatusCompleted indicates the task was completed.
	StatusCompleted Status = 3
)

// String returns the string representation of the Status.
func (s Status) String() string {
	switch s {
	case StatusIncomplete:
		return "incomplete"
	case StatusCanceled:
		return "canceled"
	case StatusCompleted:
		return "completed"
	default:
		return unknownString
	}
}

// IsOpen returns true if the status indicates an open (incomplete) task.
func (s Status) IsOpen() bool {
	return s == StatusIncomplete
}

// IsClosed returns true if the status indicates a closed (completed or canceled) task.
func (s Status) IsClosed() bool {
	return s == StatusCompleted || s == StatusCanceled
}

// StartBucket represents the scheduling bucket for a task.
type StartBucket int

const (
	// StartInbox indicates the task is in the Inbox.
	StartInbox StartBucket = 0
	// StartAnytime indicates the task is scheduled for Anytime.
	StartAnytime StartBucket = 1
	// StartSomeday indicates the task is scheduled for Someday.
	StartSomeday StartBucket = 2
)

// String returns the string representation of the StartBucket.
func (s StartBucket) String() string {
	switch s {
	case StartInbox:
		return "Inbox"
	case StartAnytime:
		return "Anytime"
	case StartSomeday:
		return "Someday"
	default:
		return unknownString
	}
}

// dateOp represents comparison operators for date-based queries.
type dateOp int

const (
	// dateOpExists checks if a date value exists (is not null).
	dateOpExists dateOp = iota
	// dateOpNotExists checks if a date value does not exist (is null).
	dateOpNotExists
	// dateOpEqual checks if a date equals a given value (=).
	dateOpEqual
	// dateOpBefore checks if a date is before a given value (<).
	dateOpBefore
	// dateOpBeforeEq checks if a date is before or equal to a given value (<=).
	dateOpBeforeEq
	// dateOpAfter checks if a date is after a given value (>).
	dateOpAfter
	// dateOpAfterEq checks if a date is after or equal to a given value (>=).
	dateOpAfterEq
	// dateOpFuture checks if a date is in the future (> today).
	dateOpFuture
	// dateOpPast checks if a date is in the past (<= today).
	dateOpPast
)

// SQLOperator returns the SQL operator for comparison operations.
// Returns empty string for non-comparison operations like Exists/Future/Past.
func (d dateOp) SQLOperator() string {
	switch d {
	case dateOpEqual:
		return "="
	case dateOpBefore:
		return "<"
	case dateOpBeforeEq:
		return "<="
	case dateOpAfter:
		return ">"
	case dateOpAfterEq:
		return ">="
	default:
		return ""
	}
}

// String returns the string representation of the dateOp.
func (d dateOp) String() string {
	switch d {
	case dateOpExists:
		return "exists"
	case dateOpNotExists:
		return "not_exists"
	case dateOpEqual:
		return "equal"
	case dateOpBefore:
		return "before"
	case dateOpBeforeEq:
		return "before_eq"
	case dateOpAfter:
		return "after"
	case dateOpAfterEq:
		return "after_eq"
	case dateOpFuture:
		return "future"
	case dateOpPast:
		return "past"
	default:
		return unknownString
	}
}

// =============================================================================
// URL Scheme Types
// =============================================================================

// Command represents Things URL scheme commands.
type Command string

const (
	// CommandShow opens and shows an item.
	CommandShow Command = "show"
	// CommandAdd creates a new to-do.
	CommandAdd Command = "add"
	// CommandAddProject creates a new project.
	CommandAddProject Command = "add-project"
	// CommandUpdate updates an existing item (requires auth token).
	CommandUpdate Command = "update"
	// CommandUpdateProject updates an existing project (requires auth token).
	CommandUpdateProject Command = "update-project"
	// CommandSearch performs a search.
	CommandSearch Command = "search"
	// CommandVersion returns Things version information.
	CommandVersion Command = "version"
	// CommandJSON enables advanced JSON-based operations.
	CommandJSON Command = "json"
)

// String returns the string representation of the Command.
func (c Command) String() string {
	return string(c)
}

// When represents scheduling values for the "when" parameter in URL scheme.
type When string

const (
	// WhenToday schedules for today.
	WhenToday When = "today"
	// WhenTomorrow schedules for tomorrow.
	WhenTomorrow When = "tomorrow"
	// WhenEvening schedules for this evening.
	WhenEvening When = "evening"
	// WhenAnytime schedules for anytime.
	WhenAnytime When = "anytime"
	// WhenSomeday schedules for someday.
	WhenSomeday When = "someday"
)

// String returns the string representation of the When value.
func (w When) String() string {
	return string(w)
}

// ListID represents built-in Things list identifiers for the show command.
type ListID string

const (
	// ListInbox is the Inbox list.
	ListInbox ListID = "inbox"
	// ListToday is the Today list.
	ListToday ListID = "today"
	// ListAnytime is the Anytime list.
	ListAnytime ListID = "anytime"
	// ListUpcoming is the Upcoming list.
	ListUpcoming ListID = "upcoming"
	// ListSomeday is the Someday list.
	ListSomeday ListID = "someday"
	// ListLogbook is the Logbook list.
	ListLogbook ListID = "logbook"
	// ListTomorrow is the Tomorrow list.
	ListTomorrow ListID = "tomorrow"
	// ListDeadlines is the Deadlines list.
	ListDeadlines ListID = "deadlines"
	// ListRepeating is the Repeating list.
	ListRepeating ListID = "repeating"
	// ListAllProjects is the All Projects list.
	ListAllProjects ListID = "all-projects"
	// ListLoggedProjects is the Logged Projects list.
	ListLoggedProjects ListID = "logged-projects"
)

// String returns the string representation of the ListID.
func (l ListID) String() string {
	return string(l)
}
