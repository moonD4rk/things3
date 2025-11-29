package things3

import "errors"

var (
	// ErrDatabaseNotFound is returned when the Things database cannot be located.
	ErrDatabaseNotFound = errors.New("things3: database not found")

	// ErrTaskNotFound is returned when a task with the specified UUID does not exist.
	ErrTaskNotFound = errors.New("things3: task not found")

	// ErrAreaNotFound is returned when an area with the specified UUID does not exist.
	ErrAreaNotFound = errors.New("things3: area not found")

	// ErrTagNotFound is returned when a tag with the specified title does not exist.
	ErrTagNotFound = errors.New("things3: tag not found")

	// ErrInvalidParameter is returned when an invalid parameter value is provided.
	ErrInvalidParameter = errors.New("things3: invalid parameter")

	// ErrDatabaseVersionTooOld is returned when the database version is not supported.
	ErrDatabaseVersionTooOld = errors.New("things3: database version too old (requires version > 21)")

	// ErrAuthTokenNotFound is returned when the URL scheme auth token cannot be read.
	ErrAuthTokenNotFound = errors.New("things3: auth token not found")
)
