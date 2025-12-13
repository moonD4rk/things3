package things3

import (
	"database/sql"
	"time"
)

// Time format constants used by Things 3 database.
const (
	dateFormat     = "2006-01-02"
	dateTimeFormat = "2006-01-02 15:04:05"
	timeFormat     = "15:04"
)

// comparePtrTime compares two *time.Time pointers for sorting.
// nil values are sorted to the end.
func comparePtrTime(a, b *time.Time) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return a.Before(*b)
}

// comparePtrTimeDesc compares two *time.Time pointers for descending sort.
// nil values are sorted to the end.
func comparePtrTimeDesc(a, b *time.Time) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return a.After(*b)
}

// parseDate parses a date string in "2006-01-02" format.
// Returns nil if the string is empty or invalid.
func parseDate(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(dateFormat, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// parseDateTime parses a datetime string in "2006-01-02 15:04:05" format.
// Returns nil if the string is empty or invalid.
func parseDateTime(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(dateTimeFormat, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// parseTime parses a time string in "15:04" format.
// Returns nil if the string is empty or invalid.
func parseTime(s sql.NullString) *time.Time {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(timeFormat, s.String)
	if err != nil {
		return nil
	}
	return &t
}

// nullString returns nil if NULL, otherwise returns pointer to string.
func nullString(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

// nullStringValue returns empty string if NULL, otherwise returns the string value.
func nullStringValue(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}
