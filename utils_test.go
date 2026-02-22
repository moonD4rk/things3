package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
