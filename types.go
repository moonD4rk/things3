package things3

import (
	"encoding/json"
	"fmt"

	"github.com/moond4rk/things3/internal/scheme"
)

// String constant for unknown enum values.
const unknownString = "unknown"

// Status string constants.
const (
	statusStringIncomplete = "incomplete"
	statusStringCompleted  = "completed"
	statusStringCanceled   = "canceled"
)

// taskType represents the kind of task in the Things 3 database.
// Users distinguish types through Go's type system (Todo, Project, Heading),
// not through this enum. It remains internal for query filter routing.
type taskType int

const (
	taskTypeTodo    taskType = 0
	taskTypeProject taskType = 1
	taskTypeHeading taskType = 2
)

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

// Start bucket string constants.
const (
	startBucketStringInbox   = "inbox"
	startBucketStringAnytime = "anytime"
	startBucketStringSomeday = "someday"
)

// String returns the string representation of the StartBucket.
func (s StartBucket) String() string {
	switch s {
	case StartInbox:
		return startBucketStringInbox
	case StartAnytime:
		return startBucketStringAnytime
	case StartSomeday:
		return startBucketStringSomeday
	default:
		return unknownString
	}
}

// MarshalJSON implements json.Marshaler for StartBucket.
func (s StartBucket) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// MarshalYAML implements yaml.Marshaler for StartBucket.
func (s StartBucket) MarshalYAML() (any, error) {
	return s.String(), nil
}

// UnmarshalJSON implements json.Unmarshaler for StartBucket.
func (s *StartBucket) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	v, err := parseStartBucket(str)
	if err != nil {
		return err
	}
	*s = v
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler for StartBucket.
func (s *StartBucket) UnmarshalYAML(unmarshal func(any) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	v, err := parseStartBucket(str)
	if err != nil {
		return err
	}
	*s = v
	return nil
}

// parseStartBucket converts a string to StartBucket.
func parseStartBucket(s string) (StartBucket, error) {
	switch s {
	case startBucketStringInbox:
		return StartInbox, nil
	case startBucketStringAnytime:
		return StartAnytime, nil
	case startBucketStringSomeday:
		return StartSomeday, nil
	default:
		return 0, fmt.Errorf("things3: unknown start bucket %q", s)
	}
}

// Command represents Things URL scheme commands (aliased from internal/scheme).
type Command = scheme.Command

// Command constants for Things URL scheme.
const (
	CommandShow          = scheme.CommandShow
	CommandAdd           = scheme.CommandAdd
	CommandAddProject    = scheme.CommandAddProject
	CommandUpdate        = scheme.CommandUpdate
	CommandUpdateProject = scheme.CommandUpdateProject
	CommandSearch        = scheme.CommandSearch
	CommandVersion       = scheme.CommandVersion
	CommandJSON          = scheme.CommandJSON
)

// ListID represents built-in Things list identifiers (aliased from internal/scheme).
type ListID = scheme.ListID

// ListID constants for built-in Things lists.
const (
	ListInbox          = scheme.ListInbox
	ListToday          = scheme.ListToday
	ListAnytime        = scheme.ListAnytime
	ListUpcoming       = scheme.ListUpcoming
	ListSomeday        = scheme.ListSomeday
	ListLogbook        = scheme.ListLogbook
	ListTomorrow       = scheme.ListTomorrow
	ListDeadlines      = scheme.ListDeadlines
	ListRepeating      = scheme.ListRepeating
	ListAllProjects    = scheme.ListAllProjects
	ListLoggedProjects = scheme.ListLoggedProjects
)

// JSON batch operation types (aliased from internal/scheme).
type (
	JSONOperation = scheme.JSONOperation
	JSONItemType  = scheme.JSONItemType
	JSONItem      = scheme.JSONItem
)

// JSON operation constants.
const (
	JSONOperationCreate = scheme.JSONOperationCreate
	JSONOperationUpdate = scheme.JSONOperationUpdate
	JSONItemTypeTodo    = scheme.JSONItemTypeTodo
	JSONItemTypeProject = scheme.JSONItemTypeProject
)

// Headings creates heading entries for a project's items.
// Used within BatchProjectConfigurator.Todos to organize todos under headings.
func Headings(headings ...string) string {
	return scheme.Headings(headings...)
}
