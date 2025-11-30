package things3

import (
	"fmt"
	"time"
)

// Things date format bit masks.
// Date format: YYYYYYYYYYYMMMMDDDDD0000000 (27-bit)
const (
	yearMask  = 0b111111111110000000000000000 // bits 16-26 for year
	monthMask = 0b000000000001111000000000000 // bits 12-15 for month
	dayMask   = 0b000000000000000111110000000 // bits 7-11 for day
)

// Things time format bit masks.
// Time format: hhhhhmmmmmm00000000000000000000 (31-bit)
const (
	hourMask   = 0b1111100000000000000000000000000 // bits 26-30 for hour
	minuteMask = 0b0000011111100000000000000000000 // bits 20-25 for minute
)

// thingsDateToTime converts a Things date integer to time.Time.
// Things date format: YYYYYYYYYYYMMMMDDDDD0000000 (27-bit binary)
// Returns zero time if thingsDate is 0 or negative.
func thingsDateToTime(thingsDate int64) time.Time {
	if thingsDate <= 0 {
		return time.Time{}
	}

	year := int((thingsDate & yearMask) >> 16)
	month := time.Month((thingsDate & monthMask) >> 12)
	day := int((thingsDate & dayMask) >> 7)

	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

// timeToThingsDate converts a time.Time to Things date integer.
// Things date format: YYYYYYYYYYYMMMMDDDDD0000000 (27-bit binary)
// Returns 0 if t is zero.
func timeToThingsDate(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	year := int64(t.Year())
	month := int64(t.Month())
	day := int64(t.Day())

	return (year << 16) | (month << 12) | (day << 7)
}

// thingsDateToString converts a Things date integer to ISO 8601 date string (YYYY-MM-DD).
// Returns empty string if thingsDate is 0 or negative.
func thingsDateToString(thingsDate int64) string {
	if thingsDate <= 0 {
		return ""
	}

	year := (thingsDate & yearMask) >> 16
	month := (thingsDate & monthMask) >> 12
	day := (thingsDate & dayMask) >> 7

	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

// stringToThingsDate converts an ISO 8601 date string (YYYY-MM-DD) to Things date integer.
// Returns 0 and error if the string is invalid.
func stringToThingsDate(isoDate string) (int64, error) {
	if isoDate == "" {
		return 0, nil
	}

	t, err := time.Parse("2006-01-02", isoDate)
	if err != nil {
		return 0, fmt.Errorf("invalid date format %q: %w", isoDate, err)
	}

	return timeToThingsDate(t), nil
}

// thingsTimeToString converts a Things time integer to time string (HH:MM).
// Things time format: hhhhhmmmmmm00000000000000000000 (31-bit binary)
// Returns empty string if thingsTime is 0 or negative.
func thingsTimeToString(thingsTime int64) string {
	if thingsTime <= 0 {
		return ""
	}

	hours := (thingsTime & hourMask) >> 26
	minutes := (thingsTime & minuteMask) >> 20

	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

// unixToTime converts Unix timestamp (seconds since epoch) to time.Time in local timezone.
// Returns zero time if unixTime is 0.
func unixToTime(unixTime float64) time.Time {
	if unixTime == 0 {
		return time.Time{}
	}
	return time.Unix(int64(unixTime), 0).Local()
}

// nowThingsDate returns the current date as a Things date integer.
func nowThingsDate() int64 {
	return timeToThingsDate(time.Now())
}

// todayThingsDateSQL returns a SQL expression that evaluates to today's Things date.
func todayThingsDateSQL() string {
	return "((strftime('%Y', date('now', 'localtime')) << 16) | " +
		"(strftime('%m', date('now', 'localtime')) << 12) | " +
		"(strftime('%d', date('now', 'localtime')) << 7))"
}
