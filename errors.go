package things3

import (
	"errors"

	idb "github.com/moond4rk/things3/internal/db"
)

// Database Errors - aliased from internal/db so users can match with errors.Is.
var (
	// ErrDatabaseNotFound is returned when the Things database cannot be located.
	ErrDatabaseNotFound = idb.ErrDatabaseNotFound
	// ErrDatabaseVersionTooOld is returned when the database version is not supported.
	ErrDatabaseVersionTooOld = idb.ErrDatabaseVersionTooOld
	// ErrAuthTokenNotFound is returned when the URL scheme auth token cannot be read.
	ErrAuthTokenNotFound = idb.ErrAuthTokenNotFound
)

// Query Errors
var (
	// ErrTaskNotFound is returned when a task with the specified UUID does not exist.
	ErrTaskNotFound = errors.New("things3: task not found")
	// ErrAreaNotFound is returned when an area with the specified UUID does not exist.
	ErrAreaNotFound = errors.New("things3: area not found")
	// ErrTagNotFound is returned when a tag with the specified title does not exist.
	ErrTagNotFound = errors.New("things3: tag not found")
	// ErrInvalidParameter is returned when an invalid parameter value is provided.
	ErrInvalidParameter = errors.New("things3: invalid parameter")
)

// URL Scheme Errors
var (
	// ErrEmptyToken is returned when an empty token is provided to WithToken.
	ErrEmptyToken = errors.New("things3: empty token provided to WithToken")
	// ErrIDRequired is returned when id is missing for an update operation.
	ErrIDRequired = errors.New("things3: id required for update operation")
	// ErrTitleTooLong is returned when title exceeds 4,000 character limit.
	ErrTitleTooLong = errors.New("things3: title exceeds 4,000 character limit")
	// ErrNotesTooLong is returned when notes exceed 10,000 character limit.
	ErrNotesTooLong = errors.New("things3: notes exceed 10,000 character limit")
	// ErrTooManyChecklistItems is returned when checklist exceeds 100 item limit.
	ErrTooManyChecklistItems = errors.New("things3: checklist exceeds 100 item limit")
	// ErrNoJSONItems is returned when building a JSON URL with no items.
	ErrNoJSONItems = errors.New("things3: no items provided for JSON operation")
	// ErrInvalidReminderTime is returned when reminder hour or minute is out of range.
	ErrInvalidReminderTime = errors.New("things3: invalid reminder time (hour must be 0-23, minute must be 0-59)")
)
