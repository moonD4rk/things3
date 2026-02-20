package things3

import (
	"encoding/json"
	"fmt"
)

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
		return taskTypeStringTodo
	case TaskTypeProject:
		return taskTypeStringProject
	case TaskTypeHeading:
		return taskTypeStringHeading
	default:
		return unknownString
	}
}

// MarshalJSON implements json.Marshaler for TaskType.
func (t TaskType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// MarshalYAML implements yaml.Marshaler for TaskType.
func (t TaskType) MarshalYAML() (any, error) {
	return t.String(), nil
}

// UnmarshalJSON implements json.Unmarshaler for TaskType.
func (t *TaskType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := parseTaskType(s)
	if err != nil {
		return err
	}
	*t = v
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler for TaskType.
func (t *TaskType) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	v, err := parseTaskType(s)
	if err != nil {
		return err
	}
	*t = v
	return nil
}

// parseTaskType converts a string to TaskType.
func parseTaskType(s string) (TaskType, error) {
	switch s {
	case taskTypeStringTodo:
		return TaskTypeTodo, nil
	case taskTypeStringProject:
		return TaskTypeProject, nil
	case taskTypeStringHeading:
		return TaskTypeHeading, nil
	default:
		return 0, fmt.Errorf("things3: unknown task type %q", s)
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
		return statusStringIncomplete
	case StatusCanceled:
		return statusStringCanceled
	case StatusCompleted:
		return statusStringCompleted
	default:
		return unknownString
	}
}

// MarshalJSON implements json.Marshaler for Status.
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// MarshalYAML implements yaml.Marshaler for Status.
func (s Status) MarshalYAML() (any, error) {
	return s.String(), nil
}

// UnmarshalJSON implements json.Unmarshaler for Status.
func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	v, err := parseStatus(str)
	if err != nil {
		return err
	}
	*s = v
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler for Status.
func (s *Status) UnmarshalYAML(unmarshal func(any) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	v, err := parseStatus(str)
	if err != nil {
		return err
	}
	*s = v
	return nil
}

// parseStatus converts a string to Status.
func parseStatus(s string) (Status, error) {
	switch s {
	case statusStringIncomplete:
		return StatusIncomplete, nil
	case statusStringCanceled:
		return StatusCanceled, nil
	case statusStringCompleted:
		return StatusCompleted, nil
	default:
		return 0, fmt.Errorf("things3: unknown status %q", s)
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

// when represents scheduling values for the "when" parameter in URL scheme.
// This is a private type; use When(time.Time) for dates or WhenEvening(),
// WhenAnytime(), WhenSomeday() methods for Things 3-specific concepts.
type when string

const (
	// whenEvening schedules for this evening.
	whenEvening when = "evening"
	// whenAnytime schedules for anytime.
	whenAnytime when = "anytime"
	// whenSomeday schedules for someday.
	whenSomeday when = "someday"
)

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
