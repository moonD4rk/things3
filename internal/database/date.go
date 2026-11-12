package database

import (
	"math"
	"time"
)

// A Things date packs a calendar day as year<<16 | month<<12 | day<<7. The year
// is open-ended rather than 11 bits wide: Things stores year 4001 as a sentinel.
const (
	yearShift  = 16
	monthShift = 12
	dayShift   = 7
	monthMask  = 0xF
	dayMask    = 0x1F
)

// A Things time packs a wall-clock time as hour<<26 | minute<<20.
const (
	hourShift   = 26
	minuteShift = 20
	hourMask    = 0x1F
	minuteMask  = 0x3F
)

// Real dates lie in [minThingsDate, maxThingsDate). Things writes 0001-01-01 for
// "unset" and 4001-01-01 for "decided per occurrence" on repeating templates.
const (
	minThingsDate int64 = 2<<yearShift | 1<<monthShift | 1<<dayShift
	maxThingsDate int64 = 4001<<yearShift | 1<<monthShift | 1<<dayShift
)

func packThingsDate(year int, month time.Month, day int) int64 {
	return int64(year)<<yearShift | int64(month)<<monthShift | int64(day)<<dayShift
}

// timeToThingsDate packs the calendar day of t, or 0 for the zero time.
func timeToThingsDate(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return packThingsDate(t.Year(), t.Month(), t.Day())
}

// isRealThingsDate reports whether d is an actual calendar day, not a sentinel.
func isRealThingsDate(d int64) bool {
	return d >= minThingsDate && d < maxThingsDate
}

// thingsDateToTime unpacks d as local midnight; a sentinel yields the zero time.
func thingsDateToTime(d int64) time.Time {
	if !isRealThingsDate(d) {
		return time.Time{}
	}
	year := int(d >> yearShift)
	month := time.Month((d >> monthShift) & monthMask)
	day := int((d >> dayShift) & dayMask)
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

// thingsTimeToClock unpacks a reminder as a wall clock on the zero date: it
// names a time of day, not an instant.
func thingsTimeToClock(v int64) time.Time {
	hour := int((v >> hourShift) & hourMask)
	minute := int((v >> minuteShift) & minuteMask)
	return time.Date(0, time.January, 1, hour, minute, 0, 0, time.UTC)
}

// unixToTime converts fractional Unix seconds to a local instant, keeping the
// microsecond precision Things records. Zero yields the zero time.
func unixToTime(seconds float64) time.Time {
	if seconds == 0 {
		return time.Time{}
	}
	return time.UnixMicro(int64(math.Round(seconds * 1e6))).Local()
}
