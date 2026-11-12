package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moond4rk/things3/internal/database"
	"github.com/moond4rk/things3/thingstest"
)

func TestSummarizeRepeat(t *testing.T) {
	day := func(d int) database.RepeatOffset { return database.RepeatOffset{Day: &d} }
	weekday := func(w time.Weekday) database.RepeatOffset { return database.RepeatOffset{Weekday: &w} }
	nth := func(n int, w time.Weekday) database.RepeatOffset {
		return database.RepeatOffset{Weekday: &w, Ordinal: &n}
	}
	yearDay := func(month, d int) database.RepeatOffset { return database.RepeatOffset{Month: &month, Day: &d} }
	until := time.Date(2027, time.January, 31, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name string
		rule database.RepeatRule
		want string
	}{
		{"daily", database.RepeatRule{Unit: database.RepeatDaily, Interval: 1}, "every day"},
		{"every few days", database.RepeatRule{Unit: database.RepeatDaily, Interval: 3}, "every 3 days"},
		{
			"after completion has no anchor",
			database.RepeatRule{Unit: database.RepeatMonthly, Interval: 11, AfterCompletion: true, Offsets: []database.RepeatOffset{day(0)}},
			"every 11 months after completion",
		},
		{
			"weekly on one day",
			database.RepeatRule{Unit: database.RepeatWeekly, Interval: 1, Offsets: []database.RepeatOffset{weekday(time.Friday)}},
			"every week on Fri",
		},
		{
			"weekly on several days",
			database.RepeatRule{
				Unit: database.RepeatWeekly, Interval: 2,
				Offsets: []database.RepeatOffset{weekday(time.Monday), weekday(time.Thursday)},
			},
			"every 2 weeks on Mon, Thu",
		},
		{
			"monthly days are stored zero-based",
			database.RepeatRule{Unit: database.RepeatMonthly, Interval: 1, Offsets: []database.RepeatOffset{day(1)}},
			"every month on the 2nd",
		},
		{
			"monthly on the last day",
			database.RepeatRule{Unit: database.RepeatMonthly, Interval: 3, Offsets: []database.RepeatOffset{day(-1)}},
			"every 3 months on the last day",
		},
		{
			"monthly on a weekday",
			database.RepeatRule{Unit: database.RepeatMonthly, Interval: 1, Offsets: []database.RepeatOffset{nth(1, time.Monday)}},
			"every month on the 1st Mon",
		},
		{
			"monthly on the last weekday",
			database.RepeatRule{Unit: database.RepeatMonthly, Interval: 1, Offsets: []database.RepeatOffset{nth(-1, time.Friday)}},
			"every month on the last Fri",
		},
		{
			"yearly months are stored zero-based",
			database.RepeatRule{Unit: database.RepeatYearly, Interval: 1, Offsets: []database.RepeatOffset{yearDay(2, 13)}},
			"every year on Mar 14",
		},
		{
			"an end date",
			database.RepeatRule{Unit: database.RepeatWeekly, Interval: 1, Until: &until, Offsets: []database.RepeatOffset{weekday(time.Sunday)}},
			"every week on Sun, until 2027-01-31",
		},
		{"an unknown unit still says it repeats", database.RepeatRule{Unit: 2, Interval: 1}, "repeats"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, summarizeRepeat(&tt.rule))
		})
	}
}

func TestOrdinal(t *testing.T) {
	for n, want := range map[int]string{
		1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 11: "11th", 12: "12th", 13: "13th",
		21: "21st", 22: "22nd", 23: "23rd", 30: "30th", 31: "31st", -1: "last",
	} {
		assert.Equal(t, want, ordinal(n))
	}
}

// Both the series and the copies Things created from it carry the rule.
func TestRepeatOnSeriesAndInstance(t *testing.T) {
	client := newFixtureClientAt(t, thingstest.DatabasePath(t))
	ctx := t.Context()

	series, err := client.Todos().WithUUID(testUUIDTemplate).IncludeTemplates().First(ctx)
	require.NoError(t, err)
	require.NotNil(t, series.Repeat)
	assert.Equal(t, RepeatWeekly, series.Repeat.Unit)
	assert.Equal(t, 1, series.Repeat.Every)
	assert.True(t, series.Repeat.AfterCompletion)
	assert.Equal(t, "every week after completion", series.Repeat.Summary)
	assert.False(t, series.Repeat.Paused)
	assert.False(t, series.Repeat.Ended, "it has a next occurrence")

	instance, err := client.Todos().WithUUID(testUUIDTodoRepeating).First(ctx)
	require.NoError(t, err)
	require.NotNil(t, instance.Repeat, "a copy reports its series' rule")
	assert.Equal(t, series.Repeat.Summary, instance.Repeat.Summary)

	plain, err := client.Todos().WithUUID(testUUIDTodoInToday).First(ctx)
	require.NoError(t, err)
	assert.Nil(t, plain.Repeat)
}

func TestRepeatDormantStates(t *testing.T) {
	path := thingstest.DatabasePath(t)
	execFixtureSQL(t, path, "UPDATE TMTask SET rt1_nextInstanceStartDate = NULL WHERE uuid = ?", testUUIDTemplate)
	ended, err := newFixtureClientAt(t, path).Todos().WithUUID(testUUIDTemplate).IncludeTemplates().First(t.Context())
	require.NoError(t, err)
	assert.True(t, ended.Repeat.Ended, "no next occurrence and not paused: the series has run out")
	assert.Equal(t, "every week after completion (ended)", ended.Repeat.String())

	execFixtureSQL(t, path, "UPDATE TMTask SET rt1_instanceCreationPaused = 1 WHERE uuid = ?", testUUIDTemplate)
	paused, err := newFixtureClientAt(t, path).Todos().WithUUID(testUUIDTemplate).IncludeTemplates().First(t.Context())
	require.NoError(t, err)
	assert.True(t, paused.Repeat.Paused)
	assert.False(t, paused.Repeat.Ended)
	assert.Equal(t, "every week after completion (paused)", paused.Repeat.String())
}

// A rule this version cannot read must not break the query: the item is still
// flagged as repeating, just without the rule in words.
func TestUnreadableRepeatRuleIsTolerated(t *testing.T) {
	path := thingstest.DatabasePath(t)
	execFixtureSQL(t, path, "UPDATE TMTask SET rt1_recurrenceRule = X'00FF' WHERE uuid = ?", testUUIDTemplate)

	series, err := newFixtureClientAt(t, path).Todos().WithUUID(testUUIDTemplate).IncludeTemplates().First(t.Context())
	require.NoError(t, err)
	assert.True(t, series.Repeating)
	assert.True(t, series.Template)
	assert.Nil(t, series.Repeat)
}
