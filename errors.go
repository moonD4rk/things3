package things3

import (
	"errors"

	idb "github.com/moond4rk/things3/internal/db"
	"github.com/moond4rk/things3/internal/scheme"
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

// URL Scheme Validation Errors - aliased from internal/scheme.
var (
	// ErrTitleTooLong is returned when title exceeds the character limit.
	ErrTitleTooLong = scheme.ErrTitleTooLong
	// ErrNotesTooLong is returned when notes exceed the character limit.
	ErrNotesTooLong = scheme.ErrNotesTooLong
	// ErrTooManyChecklistItems is returned when checklist exceeds the item limit.
	ErrTooManyChecklistItems = scheme.ErrTooManyChecklistItems
	// ErrInvalidReminderTime is returned when reminder hour or minute is out of range.
	ErrInvalidReminderTime = scheme.ErrInvalidReminderTime
)

// URL Scheme Operation Errors - aliased from internal/scheme.
var (
	// ErrEmptyToken is returned when an empty token is provided to WithToken.
	ErrEmptyToken = scheme.ErrEmptyToken
	// ErrIDRequired is returned when id is missing for an update operation.
	ErrIDRequired = scheme.ErrIDRequired
	// ErrNoJSONItems is returned when building a JSON URL with no items.
	ErrNoJSONItems = scheme.ErrNoJSONItems
)
