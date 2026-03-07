package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestComparePtrTimeCmp(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)
	later := now.Add(time.Hour)

	tests := []struct {
		name     string
		a        *time.Time
		b        *time.Time
		expected int
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: 0,
		},
		{
			name:     "a nil b not nil",
			a:        nil,
			b:        &now,
			expected: 1,
		},
		{
			name:     "a not nil b nil",
			a:        &now,
			b:        nil,
			expected: -1,
		},
		{
			name:     "a before b",
			a:        &earlier,
			b:        &later,
			expected: -1,
		},
		{
			name:     "a after b",
			a:        &later,
			b:        &earlier,
			expected: 1,
		},
		{
			name:     "a equals b",
			a:        &now,
			b:        &now,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := comparePtrTimeCmp(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
