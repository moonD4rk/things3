package things3

// String constant for unknown enum values.
const unknownString = "unknown"

// TaskType represents the type of a task in Things 3.
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

// DateOp represents comparison operators for date-based queries.
type DateOp int

const (
	// DateOpExists checks if a date value exists (is not null).
	DateOpExists DateOp = iota
	// DateOpNotExists checks if a date value does not exist (is null).
	DateOpNotExists
	// DateOpBefore checks if a date is before a given value.
	DateOpBefore
	// DateOpAfter checks if a date is after a given value.
	DateOpAfter
	// DateOpOn checks if a date equals a given value.
	DateOpOn
	// DateOpBetween checks if a date is between two values.
	DateOpBetween
)

// String returns the string representation of the DateOp.
func (d DateOp) String() string {
	switch d {
	case DateOpExists:
		return "exists"
	case DateOpNotExists:
		return "not_exists"
	case DateOpBefore:
		return "before"
	case DateOpAfter:
		return "after"
	case DateOpOn:
		return "on"
	case DateOpBetween:
		return "between"
	default:
		return unknownString
	}
}

// URLCommand represents Things URL scheme commands.
type URLCommand string

const (
	// URLCommandShow opens and shows an item.
	URLCommandShow URLCommand = "show"
	// URLCommandAdd creates a new to-do.
	URLCommandAdd URLCommand = "add"
	// URLCommandAddProject creates a new project.
	URLCommandAddProject URLCommand = "add-project"
	// URLCommandUpdate updates an existing item.
	URLCommandUpdate URLCommand = "update"
	// URLCommandUpdateProject updates an existing project.
	URLCommandUpdateProject URLCommand = "update-project"
	// URLCommandSearch performs a search.
	URLCommandSearch URLCommand = "search"
)

// String returns the string representation of the URLCommand.
func (u URLCommand) String() string {
	return string(u)
}
