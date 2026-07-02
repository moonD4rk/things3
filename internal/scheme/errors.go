package scheme

import "errors"

// Operation errors for URL scheme builders.
var (
	// ErrEmptyToken is returned when the auth token resolves to empty for an
	// operation that requires one (update commands and batch updates).
	ErrEmptyToken = errors.New("things3: auth token is empty")
	// ErrIDRequired is returned when id is missing for an update operation.
	ErrIDRequired = errors.New("things3: id required for update operation")
	// ErrNoJSONItems is returned when building a JSON URL with no items.
	ErrNoJSONItems = errors.New("things3: no items provided for JSON operation")
)
