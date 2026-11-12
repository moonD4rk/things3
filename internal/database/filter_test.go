package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWhereBuilder_add(t *testing.T) {
	var w whereBuilder
	w.add("") // skipped
	w.add("x = 1")
	w.add("") // skipped
	w.add("y = 2")
	assert.Equal(t, "x = 1\n            AND y = 2", w.sql())
}

func TestWhereBuilder_addRawf(t *testing.T) {
	var w whereBuilder
	w.addRawf("type = %d", 0)
	w.addRawf("status = %d", 3)
	assert.Equal(t, "type = 0\n            AND status = 3", w.sql())
}

func TestWhereBuilder_addStringEqual(t *testing.T) {
	var w whereBuilder
	w.addStringEqual("col", new("test"))
	assert.Equal(t, "col = 'test'", w.sql())

	var w2 whereBuilder
	w2.addStringEqual("col", new("it's"))
	assert.Equal(t, "col = 'it''s'", w2.sql())

	var w3 whereBuilder
	w3.addStringEqual("col", nil)
	assert.Equal(t, sqlTrue, w3.sql())
}

func TestWhereBuilder_addIntEqual(t *testing.T) {
	var w whereBuilder
	w.addIntEqual("col", new(42))
	assert.Equal(t, "col = 42", w.sql())

	var w2 whereBuilder
	w2.addIntEqual("col", nil)
	assert.Equal(t, sqlTrue, w2.sql())
}

func TestWhereBuilder_addExists(t *testing.T) {
	var w whereBuilder
	w.addExists("col", true)
	assert.Equal(t, "col IS NOT NULL", w.sql())

	var w2 whereBuilder
	w2.addExists("col", false)
	assert.Equal(t, "col IS NULL", w2.sql())
}

func TestWhereBuilder_addFilter(t *testing.T) {
	t.Run("value takes precedence", func(t *testing.T) {
		var w whereBuilder
		w.addFilter("col", new("test"), new(true))
		assert.Equal(t, "col = 'test'", w.sql())
	})

	t.Run("exists fallback", func(t *testing.T) {
		var w whereBuilder
		w.addFilter("col", nil, new(true))
		assert.Equal(t, "col IS NOT NULL", w.sql())
	})

	t.Run("exists false", func(t *testing.T) {
		var w whereBuilder
		w.addFilter("col", nil, new(false))
		assert.Equal(t, "col IS NULL", w.sql())
	})

	t.Run("both nil", func(t *testing.T) {
		var w whereBuilder
		w.addFilter("col", nil, nil)
		assert.Equal(t, sqlTrue, w.sql())
	})
}

func TestWhereBuilder_addTagFilter(t *testing.T) {
	exists := func(cond string) string { return "EXISTS(" + cond + ")" }

	var byTitle whereBuilder
	byTitle.addTagFilter(exists, new("it's"), new(false))
	assert.Equal(t, "EXISTS(TAG.title = 'it''s')", byTitle.sql(), "a title wins over has and is escaped")

	var has whereBuilder
	has.addTagFilter(exists, nil, new(true))
	assert.Equal(t, "EXISTS()", has.sql())

	var hasNot whereBuilder
	hasNot.addTagFilter(exists, nil, new(false))
	assert.Equal(t, "NOT EXISTS()", hasNot.sql())

	var unset whereBuilder
	unset.addTagFilter(exists, nil, nil)
	assert.Equal(t, sqlTrue, unset.sql())
}

func TestWhereBuilder_addLikePrefix(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"no metacharacters", "ABC", `col LIKE 'ABC%' ESCAPE '\'`},
		{"empty skipped", "", sqlTrue},
		{"percent escaped", "50%", `col LIKE '50\%%' ESCAPE '\'`},
		{"underscore escaped", "a_b", `col LIKE 'a\_b%' ESCAPE '\'`},
		{"backslash escaped", `a\b`, `col LIKE 'a\\b%' ESCAPE '\'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w whereBuilder
			w.addLikePrefix("col", tt.value)
			assert.Equal(t, tt.want, w.sql())
		})
	}
}

func TestWhereBuilder_addLikeContains(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"no metacharacters", "milk", `col LIKE '%milk%' ESCAPE '\'`},
		{"empty skipped", "", sqlTrue},
		{"percent escaped", "%", `col LIKE '%\%%' ESCAPE '\'`},
		{"underscore escaped", "To_Do", `col LIKE '%To\_Do%' ESCAPE '\'`},
		{"quote escaped", "it's", `col LIKE '%it''s%' ESCAPE '\'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w whereBuilder
			w.addLikeContains("col", tt.value)
			assert.Equal(t, tt.want, w.sql())
		})
	}
}

func Test_escapeLikePattern(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc", "abc"},
		{"%", `\%`},
		{"_", `\_`},
		{`\`, `\\`},
		{`100%_done\ok`, `100\%\_done\\ok`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, escapeLikePattern(tt.input))
		})
	}
}

func TestWhereBuilder_addTruthy(t *testing.T) {
	tests := []struct {
		name        string
		value       *bool
		nullDefault int
		want        string
	}{
		{"nil", nil, 0, sqlTrue},
		{"true null default 0", new(true), 0, "IFNULL(col, 0)"},
		{"false null default 0", new(false), 0, "NOT IFNULL(col, 0)"},
		{"true null default 1", new(true), 1, "IFNULL(col, 1)"},
		{"false null default 1", new(false), 1, "NOT IFNULL(col, 1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w whereBuilder
			w.addTruthy("col", tt.value, tt.nullDefault)
			assert.Equal(t, tt.want, w.sql())
		})
	}
}

func TestWhereBuilder_addSearch(t *testing.T) {
	var w whereBuilder
	w.addSearch("buy milk")
	assert.Equal(t,
		`(TASK.title LIKE '%buy milk%' ESCAPE '\' OR TASK.notes LIKE '%buy milk%' ESCAPE '\')`,
		w.sql())

	var w2 whereBuilder
	w2.addSearch("")
	assert.Equal(t, sqlTrue, w2.sql())
}

func TestWhereBuilder_addSearch_escapesLikeMetacharacters(t *testing.T) {
	var w whereBuilder
	w.addSearch("%")
	assert.Equal(t,
		`(TASK.title LIKE '%\%%' ESCAPE '\' OR TASK.notes LIKE '%\%%' ESCAPE '\')`,
		w.sql())
}

func TestWhereBuilder_addCreatedAfter(t *testing.T) {
	var w whereBuilder
	w.addCreatedAfter("creationDate", time.Unix(1718418600, 250_000_000))
	assert.Equal(t, "creationDate > 1718418600.250000", w.sql(), "sub-second precision survives")

	var w2 whereBuilder
	w2.addCreatedAfter("creationDate", time.Time{})
	assert.Equal(t, sqlTrue, w2.sql())
}

// The same instant must yield identical SQL regardless of the Location
// carried by the time.Time value.
func TestWhereBuilder_addCreatedAfter_locationInsensitive(t *testing.T) {
	instant := time.Date(2024, 6, 15, 10, 30, 0, 0, time.FixedZone("EAST", 14*3600))

	var east, west, local whereBuilder
	east.addCreatedAfter("creationDate", instant)
	west.addCreatedAfter("creationDate", instant.In(time.FixedZone("WEST", -12*3600)))
	local.addCreatedAfter("creationDate", instant.In(time.Local))

	assert.Equal(t, local.sql(), east.sql())
	assert.Equal(t, local.sql(), west.sql())
}

func TestWhereBuilder_dateFilters(t *testing.T) {
	today := time.Date(2024, 6, 15, 23, 59, 0, 0, time.Local)
	const todayPacked = "132671360"

	things := func(v *DateFilterValue) string {
		var w whereBuilder
		w.addThingsDateFilter("startDate", v, today)
		return w.sql()
	}
	unix := func(v *DateFilterValue) string {
		var w whereBuilder
		w.addUnixDateFilter("stopDate", v, today)
		return w.sql()
	}

	t.Run("nil value adds nothing", func(t *testing.T) {
		assert.Equal(t, sqlTrue, things(nil))
		assert.Equal(t, sqlTrue, unix(nil))
	})

	t.Run("existence", func(t *testing.T) {
		assert.Equal(t, "startDate IS NOT NULL", things(&DateFilterValue{HasDate: new(true)}))
		assert.Equal(t, "startDate IS NULL", things(&DateFilterValue{HasDate: new(false)}))
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') IS NOT NULL", unix(&DateFilterValue{HasDate: new(true)}))
	})

	t.Run("relative filters use the injected today, not the wall clock", func(t *testing.T) {
		assert.Equal(t, "startDate > "+todayPacked, things(&DateFilterValue{Relative: DateFuture}))
		assert.Equal(t, "startDate <= "+todayPacked, things(&DateFilterValue{Relative: DatePast}))
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') > '2024-06-15'", unix(&DateFilterValue{Relative: DateFuture}))
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') <= '2024-06-15'", unix(&DateFilterValue{Relative: DatePast}))
	})

	t.Run("specific date", func(t *testing.T) {
		day := time.Date(2024, 6, 1, 0, 0, 0, 0, time.Local)
		assert.Equal(t, "startDate >= 132669568", things(&DateFilterValue{Operator: ">=", Date: &day}))
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') = '2024-06-01'", unix(&DateFilterValue{Operator: "=", Date: &day}))
	})

	t.Run("specific date is location insensitive", func(t *testing.T) {
		instant := time.Date(2024, 6, 15, 12, 0, 0, 0, time.FixedZone("EAST", 14*3600))
		elsewhere := instant.In(time.FixedZone("WEST", -12*3600))
		for _, build := range []func(*DateFilterValue) string{things, unix} {
			assert.Equal(t,
				build(&DateFilterValue{Operator: "=", Date: &instant}),
				build(&DateFilterValue{Operator: "=", Date: &elsewhere}))
		}
	})

	// A caller can hand any time.Time to a comparison filter. Out-of-range years
	// must still produce a meaningful condition: dropping it would widen the query
	// instead of narrowing it.
	t.Run("out-of-range date clamps instead of vanishing", func(t *testing.T) {
		farFuture := time.Date(12026, 6, 15, 0, 0, 0, 0, time.Local)
		negative := time.Date(-8215, 6, 15, 0, 0, 0, 0, time.Local)

		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') <= '9999-12-31'", unix(&DateFilterValue{Operator: "<=", Date: &farFuture}))
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') <= '0001-01-02'", unix(&DateFilterValue{Operator: "<=", Date: &negative}))
		assert.Equal(t, "startDate <= 655347584", things(&DateFilterValue{Operator: "<=", Date: &farFuture}))
		assert.Equal(t, "startDate <= 69888", things(&DateFilterValue{Operator: "<=", Date: &negative}))
	})
}

func TestWhereBuilder_empty(t *testing.T) {
	var w whereBuilder
	assert.Equal(t, sqlTrue, w.sql())
}
