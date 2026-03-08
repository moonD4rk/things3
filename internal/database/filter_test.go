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
	assert.Equal(t, "TRUE", w3.sql())
}

func TestWhereBuilder_addIntEqual(t *testing.T) {
	var w whereBuilder
	w.addIntEqual("col", new(42))
	assert.Equal(t, "col = 42", w.sql())

	var w2 whereBuilder
	w2.addIntEqual("col", nil)
	assert.Equal(t, "TRUE", w2.sql())
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
		assert.Equal(t, "TRUE", w.sql())
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

	t.Run("exists false", func(t *testing.T) {
		var w whereBuilder
		w.addOrFilter("a", "b", nil, new(false))
		assert.Equal(t, "(a IS NULL OR b IS NULL)", w.sql())
	})

	t.Run("both nil", func(t *testing.T) {
		var w whereBuilder
		w.addOrFilter("a", "b", nil, nil)
		assert.Equal(t, "TRUE", w.sql())
	})
}

func TestWhereBuilder_addLike(t *testing.T) {
	var w whereBuilder
	w.addLike("col", "ABC%")
	assert.Equal(t, "col LIKE 'ABC%'", w.sql())

	var w2 whereBuilder
	w2.addLike("col", "")
	assert.Equal(t, "TRUE", w2.sql())
}

func TestWhereBuilder_addTruthy(t *testing.T) {
	tests := []struct {
		name  string
		value *bool
		want  string
	}{
		{"nil", nil, "TRUE"},
		{"true", new(true), "col"},
		{"false", new(false), "NOT IFNULL(col, 0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w whereBuilder
			w.addTruthy("col", tt.value)
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
	assert.Equal(t, "TRUE", w3.sql())
}

func TestWhereBuilder_addSearch(t *testing.T) {
	var w whereBuilder
	w.addSearch("buy milk")
	assert.Equal(t,
		"(TASK.title LIKE '%buy milk%' OR TASK.notes LIKE '%buy milk%' OR AREA.title LIKE '%buy milk%')",
		w.sql())

	var w2 whereBuilder
	w2.addSearch("")
	assert.Equal(t, "TRUE", w2.sql())
}

func TestWhereBuilder_addCreatedAfter(t *testing.T) {
	var w whereBuilder
	w.addCreatedAfter("creationDate", time.Date(2024, 6, 15, 10, 30, 0, 0, time.Local))
	assert.Equal(t, "datetime(creationDate, 'unixepoch', 'localtime') > '2024-06-15 10:30:00'", w.sql())

	var w2 whereBuilder
	w2.addCreatedAfter("creationDate", time.Time{})
	assert.Equal(t, "TRUE", w2.sql())
}

func TestWhereBuilder_addDateFilter(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("col", nil, true)
		assert.Equal(t, "TRUE", w.sql())
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
			Date:     new(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)),
		}, false)
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') = date('2024-06-15')", w.sql())
	})
}

func TestWhereBuilder_empty(t *testing.T) {
	var w whereBuilder
	assert.Equal(t, "TRUE", w.sql())
}

func TestExistsSQL(t *testing.T) {
	assert.Equal(t, "col IS NOT NULL", existsSQL("col", true))
	assert.Equal(t, "col IS NULL", existsSQL("col", false))
}
