package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStaticFilter(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    string
		isEmpty bool
	}{
		{"empty expression", "", "", true},
		{"simple expression", "type = 0", "type = 0", false},
		{"complex expression", "status = 0 AND trashed = 0", "status = 0 AND trashed = 0", false},
		{"filter constant", filterIsTodo, "type = 0", false},
		{"recurring filter", filterIsNotRecurring, "rt1_recurrenceRule IS NULL", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := static(tt.expr)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestEqualFilter(t *testing.T) {
	tests := []struct {
		name    string
		column  string
		value   any
		want    string
		isEmpty bool
	}{
		// nil cases
		{"nil value", "col", nil, "", true},

		// bool cases
		{"bool true", "status", true, "status IS NOT NULL", false},
		{"bool false", "status", false, "status IS NULL", false},

		// string cases
		{"simple string", "title", "test", "title = 'test'", false},
		{"string with quote", "title", "it's", "title = 'it''s'", false},
		{"empty string", "title", "", "title = ''", false},
		{"string with spaces", "title", "hello world", "title = 'hello world'", false},

		// other types
		{"integer", "count", 42, "count = '42'", false},
		{"float", "rate", 3.14, "rate = '3.14'", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := equal(tt.column, tt.value)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestTruthyFilter(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name    string
		column  string
		value   *bool
		want    string
		isEmpty bool
	}{
		{"nil pointer", "recurring", nil, "", true},
		{"true value", "recurring", &trueVal, "recurring", false},
		{"false value", "recurring", &falseVal, "NOT IFNULL(recurring, 0)", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := truthy(tt.column, tt.value)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestOrFilter(t *testing.T) {
	tests := []struct {
		name    string
		filters []filter
		want    string
		isEmpty bool
	}{
		// Empty cases
		{"no filters", []filter{}, "", true},
		{"all empty filters", []filter{static(""), static("")}, "", true},

		// Single filter
		{"single filter", []filter{static("x = 1")}, "(x = 1)", false},

		// Multiple filters
		{"two filters", []filter{static("x = 1"), static("y = 2")}, "(x = 1 OR y = 2)", false},
		{"three filters", []filter{static("x = 1"), static("y = 2"), static("z = 3")}, "(x = 1 OR y = 2 OR z = 3)", false},

		// Mixed empty and non-empty
		{"skip empty", []filter{static("x = 1"), static(""), static("y = 2")}, "(x = 1 OR y = 2)", false},
		{"all but one empty", []filter{static(""), static("x = 1"), static("")}, "(x = 1)", false},

		// Nested filters
		{"with equal filter", []filter{equal("col", "val"), static("x = 1")}, "(col = 'val' OR x = 1)", false},
		{"with nil equal", []filter{equal("col", nil), static("x = 1")}, "(x = 1)", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := or(tt.filters...)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestSearchFilter(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		columns []string
		want    string
		isEmpty bool
	}{
		{"empty query", "", nil, "", true},
		{
			"simple query default columns", "test", nil,
			"(TASK.title LIKE '%test%' OR TASK.notes LIKE '%test%' OR AREA.title LIKE '%test%')", false,
		},
		{
			"query with quote", "it's", nil,
			"(TASK.title LIKE '%it''s%' OR TASK.notes LIKE '%it''s%' OR AREA.title LIKE '%it''s%')", false,
		},
		{
			"custom columns", "test",
			[]string{"col1", "col2"},
			"(col1 LIKE '%test%' OR col2 LIKE '%test%')", false,
		},
		{
			"single column", "test",
			[]string{"title"},
			"(title LIKE '%test%')", false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := search(tt.query, tt.columns...)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestFiltersSQL(t *testing.T) {
	tests := []struct {
		name    string
		filters filters
		want    string
	}{
		{"empty filters", filters{}, "TRUE"},
		{"all empty", filters{static(""), static("")}, "TRUE"},
		{"single filter", filters{static("x = 1")}, "x = 1"},
		{"two filters", filters{static("x = 1"), static("y = 2")}, "x = 1\n            AND y = 2"},
		{"skip empty", filters{static("x = 1"), static(""), static("y = 2")}, "x = 1\n            AND y = 2"},
		{"mixed types", filters{
			static("type = 0"),
			equal("uuid", "abc"),
			or(static("a = 1"), static("b = 2")),
		}, "type = 0\n            AND uuid = 'abc'\n            AND (a = 1 OR b = 2)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filters.SQL())
		})
	}
}

func TestFiltersIsEmpty(t *testing.T) {
	tests := []struct {
		name    string
		filters filters
		want    bool
	}{
		{"nil filters", nil, true},
		{"empty slice", filters{}, true},
		{"all empty filters", filters{static(""), equal("col", nil)}, true},
		{"one non-empty", filters{static(""), static("x = 1")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filters.IsEmpty())
		})
	}
}

func TestFilterBuilder(t *testing.T) {
	trueVal := true

	t.Run("empty builder", func(t *testing.T) {
		fb := newFilterBuilder()
		assert.Equal(t, "TRUE", fb.sql())
	})

	t.Run("static filters", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic("type = 0").
			addStatic("status = 0")
		assert.Equal(t, "type = 0\n            AND status = 0", fb.sql())
	})

	t.Run("mixed filters", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic(filterIsNotRecurring).
			addStatic(filterIsNotTrashed).
			addEqual("TASK.uuid", "test-uuid").
			addTruthy("PROJECT.trashed", &trueVal)

		sql := fb.sql()
		assert.Contains(t, sql, "rt1_recurrenceRule IS NULL")
		assert.Contains(t, sql, "trashed = 0")
		assert.Contains(t, sql, "TASK.uuid = 'test-uuid'")
		assert.Contains(t, sql, "PROJECT.trashed")
	})

	t.Run("with Or filter", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic("type = 0").
			addOr(
				equal("TASK.project", "proj-uuid"),
				equal("PROJECT_OF_HEADING.uuid", "proj-uuid"),
			)

		sql := fb.sql()
		assert.Contains(t, sql, "type = 0")
		assert.Contains(t, sql, "(TASK.project = 'proj-uuid' OR PROJECT_OF_HEADING.uuid = 'proj-uuid')")
	})

	t.Run("with search", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic("type = 0").
			addSearch("test query")

		sql := fb.sql()
		assert.Contains(t, sql, "type = 0")
		assert.Contains(t, sql, "TASK.title LIKE '%test query%'")
	})

	t.Run("skip nil values", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic("type = 0").
			addEqual("col", nil).
			addTruthy("col2", nil).
			addSearch("")

		// Should only have "type = 0"
		assert.Equal(t, "type = 0", fb.sql())
	})

	t.Run("Build returns filters", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic("x = 1").
			addStatic("y = 2")

		filters := fb.build()
		assert.Len(t, filters, 2)
	})
}

func TestFilterBuilderChaining(t *testing.T) {
	// Verify that all methods return *filterBuilder for chaining
	fb := newFilterBuilder()
	trueVal := true

	result := fb.
		add(static("a")).
		addStatic("b").
		addEqual("c", "d").
		addTruthy("e", &trueVal).
		addOr(static("f"), static("g")).
		addSearch("h")

	assert.NotNil(t, result)
	assert.Equal(t, fb, result) // Same instance
	assert.Len(t, fb.build(), 6)
}

func TestDateOpSQLOperator(t *testing.T) {
	tests := []struct {
		op   dateOp
		want string
	}{
		{dateOpExists, ""},
		{dateOpNotExists, ""},
		{dateOpEqual, "="},
		{dateOpBefore, "<"},
		{dateOpBeforeEq, "<="},
		{dateOpAfter, ">"},
		{dateOpAfterEq, ">="},
		{dateOpFuture, ""},
		{dateOpPast, ""},
	}
	for _, tt := range tests {
		t.Run(tt.op.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.op.SQLOperator())
		})
	}
}

func TestThingsDateFilter(t *testing.T) {
	tests := []struct {
		name    string
		column  string
		op      dateOp
		value   string
		want    string
		isEmpty bool
	}{
		// Existence checks
		{"exists", "startDate", dateOpExists, "", "startDate IS NOT NULL", false},
		{"not exists", "deadline", dateOpNotExists, "", "deadline IS NULL", false},

		// Relative to today
		{
			"future", "startDate", dateOpFuture, "",
			"startDate > " + todayThingsDateSQL(), false,
		},
		{
			"past", "deadline", dateOpPast, "",
			"deadline <= " + todayThingsDateSQL(), false,
		},

		// Date comparisons (2024-06-15 = year:2024, month:6, day:15)
		// Things date: (2024 << 16) | (6 << 12) | (15 << 7) = 132644864 + 24576 + 1920 = 132671360
		{
			"equal date", "startDate", dateOpEqual, "2024-06-15",
			"startDate = 132671360", false,
		},
		{
			"before date", "deadline", dateOpBefore, "2024-06-15",
			"deadline < 132671360", false,
		},
		{
			"before or equal", "startDate", dateOpBeforeEq, "2024-06-15",
			"startDate <= 132671360", false,
		},
		{
			"after date", "deadline", dateOpAfter, "2024-06-15",
			"deadline > 132671360", false,
		},
		{
			"after or equal", "startDate", dateOpAfterEq, "2024-06-15",
			"startDate >= 132671360", false,
		},

		// Empty cases
		{"empty value for equal", "startDate", dateOpEqual, "", "", true},
		// Invalid date format: value is not empty so IsEmpty=false, but SQL returns empty
		{"invalid date format", "startDate", dateOpEqual, "not-a-date", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := thingsDate(tt.column, tt.op, tt.value)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestUnixTimeFilter(t *testing.T) {
	tests := []struct {
		name    string
		column  string
		op      dateOp
		value   string
		want    string
		isEmpty bool
	}{
		// Existence checks
		{"exists", "stopDate", dateOpExists, "", "stopDate IS NOT NULL", false},
		{"not exists", "stopDate", dateOpNotExists, "", "stopDate IS NULL", false},

		// Relative to today
		{
			"future", "stopDate", dateOpFuture, "",
			"date(stopDate, 'unixepoch', 'localtime') > date('now', 'localtime')", false,
		},
		{
			"past", "stopDate", dateOpPast, "",
			"date(stopDate, 'unixepoch', 'localtime') <= date('now', 'localtime')", false,
		},

		// Date comparisons
		{
			"equal date", "stopDate", dateOpEqual, "2024-06-15",
			"date(stopDate, 'unixepoch', 'localtime') = date('2024-06-15')", false,
		},
		{
			"before date", "stopDate", dateOpBefore, "2024-06-15",
			"date(stopDate, 'unixepoch', 'localtime') < date('2024-06-15')", false,
		},
		{
			"before or equal", "stopDate", dateOpBeforeEq, "2024-06-15",
			"date(stopDate, 'unixepoch', 'localtime') <= date('2024-06-15')", false,
		},
		{
			"after date", "stopDate", dateOpAfter, "2024-06-15",
			"date(stopDate, 'unixepoch', 'localtime') > date('2024-06-15')", false,
		},
		{
			"after or equal", "stopDate", dateOpAfterEq, "2024-06-15",
			"date(stopDate, 'unixepoch', 'localtime') >= date('2024-06-15')", false,
		},

		// Empty cases
		{"empty value for equal", "stopDate", dateOpEqual, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := unixTime(tt.column, tt.op, tt.value)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestCreatedAfterFilter(t *testing.T) {
	// Use a fixed time for testing
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.Local)

	tests := []struct {
		name    string
		column  string
		after   time.Time
		want    string
		isEmpty bool
	}{
		{
			"specific time", "creationDate", testTime,
			"datetime(creationDate, 'unixepoch', 'localtime') > '2024-06-15 10:30:00'", false,
		},
		{
			"zero time", "creationDate",
			time.Time{},
			"", true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := createdAfter(tt.column, tt.after)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestFilterBuilderWithDateFilters(t *testing.T) {
	t.Run("mixed date filters", func(t *testing.T) {
		testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
		fb := newFilterBuilder().
			addStatic(filterIsTodo).
			addThingsDateValue("TASK.startDate", dateOpAfter, "2024-01-01").
			addUnixTimeValue("TASK.stopDate", dateOpPast, "").
			addCreatedAfterFilter("TASK.creationDate", testTime)

		sql := fb.sql()
		assert.Contains(t, sql, "type = 0")
		assert.Contains(t, sql, "TASK.startDate >")
		assert.Contains(t, sql, "date(TASK.stopDate, 'unixepoch', 'localtime') <= date('now', 'localtime')")
		assert.Contains(t, sql, "datetime(TASK.creationDate, 'unixepoch', 'localtime') > '2024-01-01 00:00:00'")
	})

	t.Run("deadline before specific date", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic(filterIsTodo).
			addStatic(filterIsIncomplete).
			addThingsDateValue("TASK.deadline", dateOpBeforeEq, "2024-12-31")

		sql := fb.sql()
		assert.Contains(t, sql, "type = 0")
		assert.Contains(t, sql, "status = 0")
		assert.Contains(t, sql, "TASK.deadline <=")
	})
}
