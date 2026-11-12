package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestThingsDateRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		packed int64
		year   int
		month  time.Month
		day    int
	}{
		{"value from a real database", 132464128, 2021, time.March, 28},
		{"leap day", packThingsDate(2024, time.February, 29), 2024, time.February, 29},
		{"first real date", minThingsDate, 2, time.January, 1},
		{"last real date", packThingsDate(4000, time.December, 31), 4000, time.December, 31},
		{"year past the 11-bit boundary", packThingsDate(2048, time.January, 1), 2048, time.January, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thingsDateToTime(tt.packed)
			assert.Equal(t, time.Date(tt.year, tt.month, tt.day, 0, 0, 0, 0, time.Local), got)
			assert.Equal(t, tt.packed, timeToThingsDate(got))
		})
	}
}

func TestThingsDateSentinelsAreNotDates(t *testing.T) {
	for name, packed := range map[string]int64{
		"zero":                        0,
		"negative":                    -1,
		"unset (0001-01-01)":          packThingsDate(1, time.January, 1),
		"unset with low bits set":     69769,
		"per-occurrence (4001-01-01)": maxThingsDate,
	} {
		t.Run(name, func(t *testing.T) {
			assert.False(t, isRealThingsDate(packed))
			assert.True(t, thingsDateToTime(packed).IsZero())
		})
	}
}

func TestTimeToThingsDate(t *testing.T) {
	assert.Zero(t, timeToThingsDate(time.Time{}), "the zero time packs to no date")
	assert.Equal(t,
		timeToThingsDate(time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)),
		timeToThingsDate(time.Date(2024, 6, 15, 23, 59, 59, 0, time.Local)),
		"the time of day is dropped")
}

func TestThingsTimeToClock(t *testing.T) {
	pack := func(hour, minute int64) int64 { return hour<<hourShift | minute<<minuteShift }

	for _, tt := range []struct {
		name         string
		packed       int64
		hour, minute int
	}{
		{"midnight is a real reminder", 0, 0, 0},
		{"value from the fixture", 840957952, 12, 34},
		{"last minute of the day", pack(23, 59), 23, 59},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := thingsTimeToClock(tt.packed)
			assert.Equal(t, tt.hour, got.Hour())
			assert.Equal(t, tt.minute, got.Minute())
			assert.Equal(t, 0, got.Year(), "a reminder is a time of day, not an instant")
		})
	}
}

func TestUnixToTime(t *testing.T) {
	assert.True(t, unixToTime(0).IsZero())

	got := unixToTime(1616958920.4014101)
	assert.Equal(t, int64(1616958920), got.Unix())
	assert.Equal(t, 401410, got.Nanosecond()/1000, "microseconds are kept")
	assert.Equal(t, time.Local, got.Location())
}
