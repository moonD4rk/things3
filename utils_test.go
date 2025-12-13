package things3

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected *time.Time
	}{
		{
			name:     "valid date",
			input:    sql.NullString{String: "2024-01-15", Valid: true},
			expected: func() *time.Time { t, _ := time.Parse("2006-01-02", "2024-01-15"); return &t }(),
		},
		{
			name:     "invalid null string",
			input:    sql.NullString{String: "", Valid: false},
			expected: nil,
		},
		{
			name:     "empty string with valid true",
			input:    sql.NullString{String: "", Valid: true},
			expected: nil,
		},
		{
			name:     "invalid date format",
			input:    sql.NullString{String: "15-01-2024", Valid: true},
			expected: nil,
		},
		{
			name:     "invalid date value",
			input:    sql.NullString{String: "not-a-date", Valid: true},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDate(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected *time.Time
	}{
		{
			name:     "valid datetime",
			input:    sql.NullString{String: "2024-01-15 10:30:45", Valid: true},
			expected: func() *time.Time { t, _ := time.Parse("2006-01-02 15:04:05", "2024-01-15 10:30:45"); return &t }(),
		},
		{
			name:     "invalid null string",
			input:    sql.NullString{String: "", Valid: false},
			expected: nil,
		},
		{
			name:     "empty string with valid true",
			input:    sql.NullString{String: "", Valid: true},
			expected: nil,
		},
		{
			name:     "invalid datetime format",
			input:    sql.NullString{String: "2024-01-15", Valid: true},
			expected: nil,
		},
		{
			name:     "invalid datetime value",
			input:    sql.NullString{String: "not-a-datetime", Valid: true},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDateTime(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected *time.Time
	}{
		{
			name:     "valid time",
			input:    sql.NullString{String: "14:30", Valid: true},
			expected: func() *time.Time { t, _ := time.Parse("15:04", "14:30"); return &t }(),
		},
		{
			name:     "invalid null string",
			input:    sql.NullString{String: "", Valid: false},
			expected: nil,
		},
		{
			name:     "empty string with valid true",
			input:    sql.NullString{String: "", Valid: true},
			expected: nil,
		},
		{
			name:     "invalid time format",
			input:    sql.NullString{String: "14:30:45", Valid: true},
			expected: nil,
		},
		{
			name:     "invalid time value",
			input:    sql.NullString{String: "not-a-time", Valid: true},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTime(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestNullString(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected *string
	}{
		{
			name:     "valid string",
			input:    sql.NullString{String: "hello", Valid: true},
			expected: func() *string { s := "hello"; return &s }(),
		},
		{
			name:     "invalid null string",
			input:    sql.NullString{String: "", Valid: false},
			expected: nil,
		},
		{
			name:     "valid empty string",
			input:    sql.NullString{String: "", Valid: true},
			expected: func() *string { s := ""; return &s }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullString(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestNullStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    sql.NullString
		expected string
	}{
		{
			name:     "valid string",
			input:    sql.NullString{String: "hello", Valid: true},
			expected: "hello",
		},
		{
			name:     "invalid null string",
			input:    sql.NullString{String: "ignored", Valid: false},
			expected: "",
		},
		{
			name:     "valid empty string",
			input:    sql.NullString{String: "", Valid: true},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nullStringValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComparePtrTime(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		a        *time.Time
		b        *time.Time
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: false,
		},
		{
			name:     "a nil b not nil",
			a:        nil,
			b:        &now,
			expected: false,
		},
		{
			name:     "a not nil b nil",
			a:        &now,
			b:        nil,
			expected: true,
		},
		{
			name:     "a before b",
			a:        &earlier,
			b:        &later,
			expected: true,
		},
		{
			name:     "a after b",
			a:        &later,
			b:        &earlier,
			expected: false,
		},
		{
			name:     "a equals b",
			a:        &now,
			b:        &now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparePtrTime(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComparePtrTimeDesc(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		a        *time.Time
		b        *time.Time
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: false,
		},
		{
			name:     "a nil b not nil",
			a:        nil,
			b:        &now,
			expected: false,
		},
		{
			name:     "a not nil b nil",
			a:        &now,
			b:        nil,
			expected: true,
		},
		{
			name:     "a before b descending",
			a:        &earlier,
			b:        &later,
			expected: false,
		},
		{
			name:     "a after b descending",
			a:        &later,
			b:        &earlier,
			expected: true,
		},
		{
			name:     "a equals b",
			a:        &now,
			b:        &now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparePtrTimeDesc(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
