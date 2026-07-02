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

func TestWhereBuilder_addOrFilter(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		var w whereBuilder
		w.addOrFilter("a", "b", new("test"), nil)
		assert.Equal(t, "(a = 'test' OR b = 'test')", w.sql())
	})

	t.Run("exists true", func(t *testing.T) {
		var w whereBuilder
		w.addOrFilter("a", "b", nil, new(true))
		assert.Equal(t, "(a IS NOT NULL OR b IS NOT NULL)", w.sql())
	})

	t.Run("exists false requires both columns NULL", func(t *testing.T) {
		var w whereBuilder
		w.addOrFilter("a", "b", nil, new(false))
		assert.Equal(t, "(a IS NULL AND b IS NULL)", w.sql())
	})

	t.Run("both nil", func(t *testing.T) {
		var w whereBuilder
		w.addOrFilter("a", "b", nil, nil)
		assert.Equal(t, sqlTrue, w.sql())
	})
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

func TestWhereBuilder_addOr(t *testing.T) {
	var w whereBuilder
	w.addOr("a = 1", "b = 2")
	assert.Equal(t, "(a = 1 OR b = 2)", w.sql())

	var w2 whereBuilder
	w2.addOr("", "b = 2", "")
	assert.Equal(t, "(b = 2)", w2.sql())

	var w3 whereBuilder
	w3.addOr("", "")
	assert.Equal(t, sqlTrue, w3.sql())
}

func TestWhereBuilder_addSearch(t *testing.T) {
	var w whereBuilder
	w.addSearch("buy milk")
	assert.Equal(t,
		`(TASK.title LIKE '%buy milk%' ESCAPE '\' OR TASK.notes LIKE '%buy milk%' ESCAPE '\' OR AREA.title LIKE '%buy milk%' ESCAPE '\')`,
		w.sql())

	var w2 whereBuilder
	w2.addSearch("")
	assert.Equal(t, sqlTrue, w2.sql())
}

func TestWhereBuilder_addSearch_escapesLikeMetacharacters(t *testing.T) {
	var w whereBuilder
	w.addSearch("%")
	assert.Equal(t,
		`(TASK.title LIKE '%\%%' ESCAPE '\' OR TASK.notes LIKE '%\%%' ESCAPE '\' OR AREA.title LIKE '%\%%' ESCAPE '\')`,
		w.sql())
}

func TestWhereBuilder_addCreatedAfter(t *testing.T) {
	var w whereBuilder
	w.addCreatedAfter("creationDate", time.Date(2024, 6, 15, 10, 30, 0, 0, time.Local))
	assert.Equal(t, "datetime(creationDate, 'unixepoch', 'localtime') > '2024-06-15 10:30:00'", w.sql())

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

func TestWhereBuilder_addDateFilter(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("col", nil, true)
		assert.Equal(t, sqlTrue, w.sql())
	})

	t.Run("exists true", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("col", &DateFilterValue{HasDate: new(true)}, true)
		assert.Equal(t, "col IS NOT NULL", w.sql())
	})

	t.Run("exists false", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("col", &DateFilterValue{HasDate: new(false)}, true)
		assert.Equal(t, "col IS NULL", w.sql())
	})

	t.Run("things date future", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("startDate", &DateFilterValue{Relative: DateFuture}, true)
		assert.Equal(t, "startDate > "+todayThingsDateSQL(), w.sql())
	})

	t.Run("things date past", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("startDate", &DateFilterValue{Relative: DatePast}, true)
		assert.Equal(t, "startDate <= "+todayThingsDateSQL(), w.sql())
	})

	t.Run("unix time future", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("stopDate", &DateFilterValue{Relative: DateFuture}, false)
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') > date('now', 'localtime')", w.sql())
	})

	t.Run("unix time past", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("stopDate", &DateFilterValue{Relative: DatePast}, false)
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') <= date('now', 'localtime')", w.sql())
	})

	t.Run("unix time specific date", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("stopDate", &DateFilterValue{
			Operator: "=",
			Date:     new(time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)),
		}, false)
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') = date('2024-06-15')", w.sql())
	})

	t.Run("specific date is location insensitive", func(t *testing.T) {
		instant := time.Date(2024, 6, 15, 12, 0, 0, 0, time.FixedZone("EAST", 14*3600))
		for _, isThingsDate := range []bool{true, false} {
			var east, west whereBuilder
			east.addDateFilter("col", &DateFilterValue{Operator: "=", Date: new(instant)}, isThingsDate)
			west.addDateFilter("col", &DateFilterValue{
				Operator: "=",
				Date:     new(instant.In(time.FixedZone("WEST", -12*3600))),
			}, isThingsDate)
			assert.Equal(t, east.sql(), west.sql(), "isThingsDate=%v", isThingsDate)
		}
	})
}

func TestWhereBuilder_empty(t *testing.T) {
	var w whereBuilder
	assert.Equal(t, sqlTrue, w.sql())
}

func TestExistsSQL(t *testing.T) {
	assert.Equal(t, "col IS NOT NULL", existsSQL("col", true))
	assert.Equal(t, "col IS NULL", existsSQL("col", false))
}
