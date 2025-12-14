package things3

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_thingsDateToTime(t *testing.T) {
	tests := []struct {
		name       string
		thingsDate int64
		wantYear   int
		wantMonth  time.Month
		wantDay    int
	}{
		{
			name:       "2021-03-28",
			thingsDate: 132464128, // From Python test
			wantYear:   2021,
			wantMonth:  time.March,
			wantDay:    28,
		},
		{
			name:       "zero",
			thingsDate: 0,
			wantYear:   1,
			wantMonth:  time.January,
			wantDay:    1,
		},
		{
			name:       "negative",
			thingsDate: -1,
			wantYear:   1,
			wantMonth:  time.January,
			wantDay:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thingsDateToTime(tt.thingsDate)
			if tt.thingsDate <= 0 {
				assert.True(t, got.IsZero(), "thingsDateToTime(%d) should return zero time", tt.thingsDate)
				return
			}
			assert.Equal(t, tt.wantYear, got.Year())
			assert.Equal(t, tt.wantMonth, got.Month())
			assert.Equal(t, tt.wantDay, got.Day())
		})
	}
}

func Test_timeToThingsDate(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want int64
	}{
		{
			name: "2021-03-28",
			time: time.Date(2021, time.March, 28, 0, 0, 0, 0, time.Local),
			want: 132464128,
		},
		{
			name: "zero",
			time: time.Time{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeToThingsDate(tt.time)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_thingsDateToString(t *testing.T) {
	tests := []struct {
		name       string
		thingsDate int64
		want       string
	}{
		{
			name:       "2021-03-28",
			thingsDate: 132464128,
			want:       "2021-03-28",
		},
		{
			name:       "zero",
			thingsDate: 0,
			want:       "",
		},
		{
			name:       "negative",
			thingsDate: -1,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thingsDateToString(tt.thingsDate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_stringToThingsDate(t *testing.T) {
	tests := []struct {
		name    string
		isoDate string
		want    int64
		wantErr bool
	}{
		{
			name:    "2021-03-28",
			isoDate: "2021-03-28",
			want:    132464128,
			wantErr: false,
		},
		{
			name:    "empty",
			isoDate: "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid",
			isoDate: "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "wrong format",
			isoDate: "03-28-2021",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringToThingsDate(tt.isoDate)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_thingsTimeToString(t *testing.T) {
	tests := []struct {
		name       string
		thingsTime int64
		want       string
	}{
		{
			name:       "12:34",
			thingsTime: (12 << 26) | (34 << 20), // Correct calculation
			want:       "12:34",
		},
		{
			name:       "00:00",
			thingsTime: 0,
			want:       "",
		},
		{
			name:       "negative",
			thingsTime: -1,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thingsTimeToString(tt.thingsTime)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_nowThingsDate(t *testing.T) {
	now := time.Now()
	thingsDate := nowThingsDate()

	// Convert back and check
	converted := thingsDateToTime(thingsDate)

	assert.Equal(t, now.Year(), converted.Year())
	assert.Equal(t, now.Month(), converted.Month())
	assert.Equal(t, now.Day(), converted.Day())
}

func Test_todayThingsDateSQL(t *testing.T) {
	sql := todayThingsDateSQL()
	assert.NotEmpty(t, sql)
	assert.Greater(t, len(sql), 50, "todayThingsDateSQL() seems too short")
}

func Test_unixToTime(t *testing.T) {
	tests := []struct {
		name     string
		unixTime float64
		wantZero bool
	}{
		{
			name:     "zero",
			unixTime: 0,
			wantZero: true,
		},
		{
			name:     "valid",
			unixTime: 1616946000, // 2021-03-28 15:00:00 UTC
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unixToTime(tt.unixTime)
			if tt.wantZero {
				assert.True(t, got.IsZero())
			} else {
				assert.False(t, got.IsZero())
			}
		})
	}
}

// =============================================================================
// Round-Trip Tests - Verify data integrity through conversions
// =============================================================================

func Test_RoundTrip_TimeToThingsDateAndBack(t *testing.T) {
	// Test that time -> ThingsDate -> time preserves date information
	tests := []struct {
		name  string
		year  int
		month time.Month
		day   int
	}{
		{"today's date", time.Now().Year(), time.Now().Month(), time.Now().Day()},
		{"new year", 2024, time.January, 1},
		{"year end", 2024, time.December, 31},
		{"leap year feb 29", 2024, time.February, 29},
		{"mid year", 2024, time.June, 15},
		{"start of month", 2024, time.March, 1},
		{"end of month", 2024, time.March, 31},
		{"year 2000", 2000, time.June, 15},
		{"year 1999", 1999, time.December, 31},
		{"year 2047 max supported", 2047, time.January, 1}, // 11 bits = max 2047
		{"minimum date", 1, time.January, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := time.Date(tt.year, tt.month, tt.day, 0, 0, 0, 0, time.Local)

			// Convert to ThingsDate and back
			thingsDate := timeToThingsDate(original)
			converted := thingsDateToTime(thingsDate)

			// Verify date components are preserved
			assert.Equal(t, tt.year, converted.Year(), "year mismatch")
			assert.Equal(t, tt.month, converted.Month(), "month mismatch")
			assert.Equal(t, tt.day, converted.Day(), "day mismatch")
		})
	}
}

func Test_RoundTrip_StringToThingsDateAndBack(t *testing.T) {
	// Test that string -> ThingsDate -> string preserves date string
	tests := []struct {
		isoDate string
	}{
		{"2024-01-01"},
		{"2024-12-31"},
		{"2024-06-15"},
		{"2024-02-29"}, // leap year
		{"2000-01-01"},
		{"2047-12-31"}, // max supported year (11 bits = 2047)
		{"1999-06-15"},
	}

	for _, tt := range tests {
		t.Run(tt.isoDate, func(t *testing.T) {
			// Convert string to ThingsDate
			thingsDate, err := stringToThingsDate(tt.isoDate)
			require.NoError(t, err)

			// Convert ThingsDate back to string
			converted := thingsDateToString(thingsDate)

			assert.Equal(t, tt.isoDate, converted, "round-trip failed")
		})
	}
}

func Test_RoundTrip_ThingsDateToTimeAndBack(t *testing.T) {
	// Test that ThingsDate -> time -> ThingsDate preserves value
	tests := []struct {
		name       string
		thingsDate int64
	}{
		{"known date 2021-03-28", 132464128},
		{"year 2024 date", timeToThingsDate(time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local))},
		{"minimum viable", timeToThingsDate(time.Date(1, 1, 1, 0, 0, 0, 0, time.Local))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to time and back
			asTime := thingsDateToTime(tt.thingsDate)
			backToThingsDate := timeToThingsDate(asTime)

			assert.Equal(t, tt.thingsDate, backToThingsDate, "round-trip failed")
		})
	}
}

// =============================================================================
// Boundary Value Tests
// =============================================================================

func Test_ThingsDate_BoundaryValues(t *testing.T) {
	tests := []struct {
		name       string
		thingsDate int64
		expectZero bool
		wantYear   int
		wantMonth  time.Month
		wantDay    int
	}{
		// Zero and negative cases
		{"zero", 0, true, 0, 0, 0},
		{"negative one", -1, true, 0, 0, 0},
		{"max negative", -9999999999, true, 0, 0, 0},

		// Valid edge cases
		{"min valid date", timeToThingsDate(time.Date(1, 1, 1, 0, 0, 0, 0, time.Local)), false, 1, time.January, 1},
		{"year 2000", timeToThingsDate(time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local)), false, 2000, time.January, 1},
		{"large year", timeToThingsDate(time.Date(2047, 12, 31, 0, 0, 0, 0, time.Local)), false, 2047, time.December, 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thingsDateToTime(tt.thingsDate)

			if tt.expectZero {
				assert.True(t, result.IsZero(), "expected zero time for %d", tt.thingsDate)
			} else {
				assert.Equal(t, tt.wantYear, result.Year())
				assert.Equal(t, tt.wantMonth, result.Month())
				assert.Equal(t, tt.wantDay, result.Day())
			}
		})
	}
}

func Test_ThingsTime_BoundaryValues(t *testing.T) {
	tests := []struct {
		name       string
		thingsTime int64
		expected   string
	}{
		// Zero and negative
		{"zero", 0, ""},
		{"negative", -1, ""},

		// Valid times
		{"midnight", (0 << 26) | (0 << 20), ""}, // 00:00 returns empty (treated as no time)
		{"one minute past midnight", (0 << 26) | (1 << 20), "00:01"},
		{"noon", (12 << 26) | (0 << 20), "12:00"},
		{"max time", (23 << 26) | (59 << 20), "23:59"},
		{"mid day", (12 << 26) | (34 << 20), "12:34"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thingsTimeToString(tt.thingsTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Invalid Input Tests
// =============================================================================

func Test_stringToThingsDate_InvalidInputs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Valid inputs
		{"valid date", "2024-06-15", false},
		{"empty string", "", false}, // Empty returns 0, no error

		// Invalid formats
		{"wrong separator", "2024/06/15", true},
		{"US format", "06-15-2024", true},
		{"no separators", "20240615", true},
		{"partial date", "2024-06", true},
		{"too many parts", "2024-06-15-00", true},
		{"letters", "abcd-ef-gh", true},
		{"mixed", "2024-ab-15", true},

		// Invalid dates (format correct but date invalid)
		{"month 13", "2024-13-01", true},
		{"day 32", "2024-01-32", true},
		{"month 0", "2024-00-15", true},
		{"day 0", "2024-06-00", true},
		{"feb 30", "2024-02-30", true},
		{"non-leap feb 29", "2023-02-29", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := stringToThingsDate(tt.input)

			if tt.wantErr {
				assert.Error(t, err, "expected error for input %q", tt.input)
			} else {
				assert.NoError(t, err, "unexpected error for input %q", tt.input)
				if tt.input == "" {
					assert.Equal(t, int64(0), result)
				}
			}
		})
	}
}

// =============================================================================
// Timezone Consistency Tests
// =============================================================================

func Test_ThingsDate_LocalTimezone(t *testing.T) {
	// Verify that ThingsDate uses local timezone consistently
	original := time.Date(2024, time.June, 15, 0, 0, 0, 0, time.Local)
	thingsDate := timeToThingsDate(original)
	converted := thingsDateToTime(thingsDate)

	// Location should be Local
	assert.Equal(t, time.Local, converted.Location(),
		"thingsDateToTime should return Local timezone")

	// Time components should be zero (date only)
	assert.Equal(t, 0, converted.Hour())
	assert.Equal(t, 0, converted.Minute())
	assert.Equal(t, 0, converted.Second())
}

func Test_UnixToTime_LocalTimezone(t *testing.T) {
	// Unix timestamp is UTC-based, result should be in Local
	unixTime := float64(1718438400) // 2024-06-15 00:00:00 UTC
	result := unixToTime(unixTime)

	// Should return Local timezone
	assert.Equal(t, time.Local, result.Location(),
		"unixToTime should return Local timezone")
	assert.False(t, result.IsZero())
}

// =============================================================================
// Special Date Tests
// =============================================================================

func Test_LeapYear_Dates(t *testing.T) {
	tests := []struct {
		name      string
		year      int
		isLeap    bool
		feb29Date string
	}{
		{"2024 is leap year", 2024, true, "2024-02-29"},
		{"2000 is leap year", 2000, true, "2000-02-29"},
		{"2023 is not leap year", 2023, false, ""},
		{"1900 is not leap year", 1900, false, ""},
		{"2100 is not leap year", 2100, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isLeap {
				// Should be able to create and convert Feb 29
				thingsDate, err := stringToThingsDate(tt.feb29Date)
				require.NoError(t, err, "leap year should allow Feb 29")

				result := thingsDateToString(thingsDate)
				assert.Equal(t, tt.feb29Date, result)
			} else {
				// Feb 29 should fail for non-leap years
				invalidDate := fmt.Sprintf("%d-02-29", tt.year)
				_, err := stringToThingsDate(invalidDate)
				assert.Error(t, err, "non-leap year should reject Feb 29")
			}
		})
	}
}

func Test_MonthEnd_Dates(t *testing.T) {
	// Test that month-end dates are correctly handled
	monthEnds := []struct {
		month   time.Month
		lastDay int
	}{
		{time.January, 31},
		{time.February, 29}, // 2024 is leap year
		{time.March, 31},
		{time.April, 30},
		{time.May, 31},
		{time.June, 30},
		{time.July, 31},
		{time.August, 31},
		{time.September, 30},
		{time.October, 31},
		{time.November, 30},
		{time.December, 31},
	}

	for _, me := range monthEnds {
		t.Run(me.month.String(), func(t *testing.T) {
			isoDate := fmt.Sprintf("2024-%02d-%02d", me.month, me.lastDay)
			thingsDate, err := stringToThingsDate(isoDate)
			require.NoError(t, err)

			result := thingsDateToString(thingsDate)
			assert.Equal(t, isoDate, result, "month-end date should round-trip")
		})
	}
}

// =============================================================================
// todayThingsDateSQL Tests
// =============================================================================

func Test_todayThingsDateSQL_Format(t *testing.T) {
	sql := todayThingsDateSQL()

	// Verify SQL structure
	assert.Contains(t, sql, "strftime('%Y'", "should extract year")
	assert.Contains(t, sql, "strftime('%m'", "should extract month")
	assert.Contains(t, sql, "strftime('%d'", "should extract day")
	assert.Contains(t, sql, "<< 16", "year should shift left 16 bits")
	assert.Contains(t, sql, "<< 12", "month should shift left 12 bits")
	assert.Contains(t, sql, "<< 7", "day should shift left 7 bits")
	assert.Contains(t, sql, "localtime", "should use local timezone")
}
