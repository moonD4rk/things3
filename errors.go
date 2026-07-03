package things3

import (
	"errors"

	"github.com/moond4rk/things3/internal/database"
	"github.com/moond4rk/things3/internal/scheme"
)

// Database Errors - aliased from internal/database so users can match with errors.Is.
var (
	// ErrDatabaseNotFound is returned when the Things database cannot be located.
	ErrDatabaseNotFound = database.ErrDatabaseNotFound
	// ErrDatabaseVersionTooOld is returned when the database version is not supported.
	ErrDatabaseVersionTooOld = database.ErrDatabaseVersionTooOld
	// ErrAuthTokenNotFound is returned when the URL scheme auth token cannot be read.
	ErrAuthTokenNotFound = database.ErrAuthTokenNotFound
)

// Query Errors
var (
	// ErrTodoNotFound is returned when a todo with the specified UUID does not exist.
	ErrTodoNotFound = errors.New("things3: todo not found")
	// ErrProjectNotFound is returned when a project with the specified UUID does not exist.
	ErrProjectNotFound = errors.New("things3: project not found")
	// ErrHeadingNotFound is returned when a heading with the specified UUID does not exist.
	ErrHeadingNotFound = errors.New("things3: heading not found")
	// ErrAreaNotFound is returned when an area with the specified UUID does not exist.
	ErrAreaNotFound = errors.New("things3: area not found")
	// ErrTagNotFound is returned when a tag with the specified title does not exist.
	ErrTagNotFound = errors.New("things3: tag not found")
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
	// ErrReminderNeedsDate is returned when a reminder is set without a concrete
	// start date (someday and anytime cannot carry one).
	ErrReminderNeedsDate = scheme.ErrReminderNeedsDate
	// ErrTagContainsComma is returned when a tag name contains a comma.
	ErrTagContainsComma = scheme.ErrTagContainsComma
	// ErrTitleContainsNewline is returned when a title contains a newline.
	ErrTitleContainsNewline = scheme.ErrTitleContainsNewline
	// ErrChecklistItemContainsNewline is returned when a checklist item contains a newline.
	ErrChecklistItemContainsNewline = scheme.ErrChecklistItemContainsNewline
)

// URL Scheme Operation Errors - aliased from internal/scheme.
var (
	// ErrEmptyToken is returned when the auth token resolves to empty for an
	// operation that requires one (update commands and batch updates).
	ErrEmptyToken = scheme.ErrEmptyToken
	// ErrIDRequired is returned when id is missing for an update operation.
	ErrIDRequired = scheme.ErrIDRequired
	// ErrNoJSONItems is returned when building a JSON URL with no items.
	ErrNoJSONItems = scheme.ErrNoJSONItems
)
