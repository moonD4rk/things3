package scheme

import "errors"

// Operation errors for URL scheme builders.
var (
	// ErrEmptyToken is returned when an empty token is provided to WithToken.
	ErrEmptyToken = errors.New("things3: empty token provided to WithToken")
	// ErrIDRequired is returned when id is missing for an update operation.
	ErrIDRequired = errors.New("things3: id required for update operation")
	// ErrNoJSONItems is returned when building a JSON URL with no items.
	ErrNoJSONItems = errors.New("things3: no items provided for JSON operation")
)
