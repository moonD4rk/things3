package db

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

func TestWhereBuilder_addEqual(t *testing.T) {
	tests := []struct {
		name   string
		column string
		value  any
		want   string
	}{
		{"nil", "col", nil, "TRUE"},
		{"bool true", "col", true, "col IS NOT NULL"},
		{"bool false", "col", false, "col IS NULL"},
		{"string", "col", "test", "col = 'test'"},
		{"string with quote", "col", "it's", "col = 'it''s'"},
		{"empty string", "col", "", "col = ''"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w whereBuilder
			w.addEqual(tt.column, tt.value)
			assert.Equal(t, tt.want, w.sql())
		})
	}
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
		{"true", ptr(true), "col"},
		{"false", ptr(false), "NOT IFNULL(col, 0)"},
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
		w.addDateFilter("col", &DateFilterValue{HasDate: ptr(true)}, true)
		assert.Equal(t, "col IS NOT NULL", w.sql())
	})

	t.Run("exists false", func(t *testing.T) {
		var w whereBuilder
		w.addDateFilter("col", &DateFilterValue{HasDate: ptr(false)}, true)
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
			Date:     ptr(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)),
		}, false)
		assert.Equal(t, "date(stopDate, 'unixepoch', 'localtime') = date('2024-06-15')", w.sql())
	})
}

func TestWhereBuilder_empty(t *testing.T) {
	var w whereBuilder
	assert.Equal(t, "TRUE", w.sql())
}

func TestEqualSQL(t *testing.T) {
	assert.Equal(t, "col IS NOT NULL", equalSQL("col", true))
	assert.Equal(t, "col IS NULL", equalSQL("col", false))
	assert.Equal(t, "col = 'test'", equalSQL("col", "test"))
	assert.Empty(t, equalSQL("col", nil))
}
