package database

import "errors"

// Database errors used by the internal db package.
var (
	// ErrDatabaseNotFound is returned when the Things database cannot be located.
	ErrDatabaseNotFound = errors.New("things3: database not found")
	// ErrDatabaseVersionTooOld is returned when the database version is not supported.
	ErrDatabaseVersionTooOld = errors.New("things3: database version too old (requires things3 version > 21)")
	// ErrAuthTokenNotFound is returned when the URL scheme auth token cannot be read.
	ErrAuthTokenNotFound = errors.New("things3: auth token not found")
)
