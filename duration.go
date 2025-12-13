package things3

import "fmt"

// Duration represents a time duration for filtering tasks by creation date.
// Use the helper functions Days(), Weeks(), Months(), Years() to create durations.
type Duration struct {
	days   int
	weeks  int
	months int
	years  int
}

// Days creates a Duration of n days.
func Days(n int) Duration {
	return Duration{days: n}
}

// Weeks creates a Duration of n weeks.
func Weeks(n int) Duration {
	return Duration{weeks: n}
}

// Months creates a Duration of n months.
func Months(n int) Duration {
	return Duration{months: n}
}

// Years creates a Duration of n years.
func Years(n int) Duration {
	return Duration{years: n}
}

// toSQLModifier converts Duration to SQLite datetime modifier string.
// Returns modifier like "-7 days", "-2 months", "-1 years".
func (d Duration) toSQLModifier() string {
	if d.years > 0 {
		return fmt.Sprintf("-%d years", d.years)
	}
	if d.months > 0 {
		return fmt.Sprintf("-%d months", d.months)
	}
	if d.weeks > 0 {
		return fmt.Sprintf("-%d days", d.weeks*7)
	}
	if d.days > 0 {
		return fmt.Sprintf("-%d days", d.days)
	}
	return ""
}

// IsZero returns true if the duration is zero.
func (d Duration) IsZero() bool {
	return d.days == 0 && d.weeks == 0 && d.months == 0 && d.years == 0
}
