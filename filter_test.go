package things3

import (
	"testing"

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
	// Verify all filters were added by checking the SQL output
	sql := fb.sql()
	assert.Contains(t, sql, "a")
	assert.Contains(t, sql, "b")
	assert.Contains(t, sql, "c = 'd'")
	assert.Contains(t, sql, "e")
	assert.Contains(t, sql, "(f OR g)")
	assert.Contains(t, sql, "LIKE '%h%'")
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

func TestDurationFilter(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		duration Duration
		want     string
		isEmpty  bool
	}{
		// Days
		{
			"7 days", "creationDate", Days(7),
			"datetime(creationDate, 'unixepoch', 'localtime') > datetime('now', '-7 days')", false,
		},
		{
			"30 days", "creationDate", Days(30),
			"datetime(creationDate, 'unixepoch', 'localtime') > datetime('now', '-30 days')", false,
		},

		// Weeks (converted to days)
		{
			"2 weeks", "creationDate", Weeks(2),
			"datetime(creationDate, 'unixepoch', 'localtime') > datetime('now', '-14 days')", false,
		},
		{
			"4 weeks", "creationDate", Weeks(4),
			"datetime(creationDate, 'unixepoch', 'localtime') > datetime('now', '-28 days')", false,
		},

		// Months
		{
			"1 month", "creationDate", Months(1),
			"datetime(creationDate, 'unixepoch', 'localtime') > datetime('now', '-1 months')", false,
		},
		{
			"6 months", "creationDate", Months(6),
			"datetime(creationDate, 'unixepoch', 'localtime') > datetime('now', '-6 months')", false,
		},

		// Years
		{
			"1 year", "creationDate", Years(1),
			"datetime(creationDate, 'unixepoch', 'localtime') > datetime('now', '-1 years')", false,
		},
		{
			"2 years", "creationDate", Years(2),
			"datetime(creationDate, 'unixepoch', 'localtime') > datetime('now', '-2 years')", false,
		},

		// Empty/zero cases
		{"zero duration", "creationDate", Duration{}, "", true},
		{"zero days", "creationDate", Days(0), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := duration(tt.column, tt.duration)
			assert.Equal(t, tt.want, f.SQL())
			assert.Equal(t, tt.isEmpty, f.IsEmpty())
		})
	}
}

func TestFilterBuilderWithDateFilters(t *testing.T) {
	t.Run("mixed date filters", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic(filterIsTodo).
			add(thingsDate("TASK.startDate", dateOpAfter, "2024-01-01")).
			add(unixTime("TASK.stopDate", dateOpPast, "")).
			addDurationFilter("TASK.creationDate", Days(30))

		sql := fb.sql()
		assert.Contains(t, sql, "type = 0")
		assert.Contains(t, sql, "TASK.startDate >")
		assert.Contains(t, sql, "date(TASK.stopDate, 'unixepoch', 'localtime') <= date('now', 'localtime')")
		assert.Contains(t, sql, "datetime(TASK.creationDate, 'unixepoch', 'localtime') > datetime('now', '-30 days')")
	})

	t.Run("deadline before specific date", func(t *testing.T) {
		fb := newFilterBuilder().
			addStatic(filterIsTodo).
			addStatic(filterIsIncomplete).
			add(thingsDate("TASK.deadline", dateOpBeforeEq, "2024-12-31"))

		sql := fb.sql()
		assert.Contains(t, sql, "type = 0")
		assert.Contains(t, sql, "status = 0")
		assert.Contains(t, sql, "TASK.deadline <=")
	})
}

// =============================================================================
// SQL Injection Prevention Tests
// =============================================================================

func TestSQLInjection_EqualFilter(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		value    string
		expected string // Expected output to verify escaping
	}{
		{
			name:     "basic injection attempt",
			column:   "title",
			value:    "'; DROP TABLE TMTask; --",
			expected: "title = '''; DROP TABLE TMTask; --'", // ' escaped to ''
		},
		{
			name:     "union injection",
			column:   "title",
			value:    "' UNION SELECT * FROM TMTag --",
			expected: "title = ''' UNION SELECT * FROM TMTag --'", // ' escaped to ''
		},
		{
			name:     "comment injection",
			column:   "title",
			value:    "test' -- comment",
			expected: "title = 'test'' -- comment'", // ' escaped to ''
		},
		{
			name:     "nested quotes",
			column:   "title",
			value:    "test'''; DROP TABLE --",
			expected: "title = 'test''''''; DROP TABLE --'", // ''' becomes ''''''
		},
		{
			name:     "null byte injection",
			column:   "title",
			value:    "test\x00' DROP TABLE",
			expected: "title = 'test\x00'' DROP TABLE'", // Null preserved, ' escaped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := equal(tt.column, tt.value)
			sql := f.SQL()

			// Verify exact expected output
			assert.Equal(t, tt.expected, sql,
				"SQL injection should be properly escaped")
		})
	}
}

func TestSQLInjection_SearchFilter(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		verify func(t *testing.T, sql string)
	}{
		{
			name:  "basic injection in search",
			query: "'; DROP TABLE TMTask; --",
			verify: func(t *testing.T, sql string) {
				// All quotes should be escaped to ''
				assert.Contains(t, sql, "''")
				// The value should be safely wrapped in the LIKE clause
				assert.Contains(t, sql, "LIKE '%''")
			},
		},
		{
			name:  "LIKE wildcard injection",
			query: "%' OR 1=1 --",
			verify: func(t *testing.T, sql string) {
				// The % is legitimate for LIKE, but ' should be escaped
				assert.Contains(t, sql, "''")
				// Should still be within LIKE quotes
				assert.Contains(t, sql, "LIKE '%%''")
			},
		},
		{
			name:  "unicode injection attempt",
			query: "test\u0027 OR 1=1",
			verify: func(t *testing.T, sql string) {
				// Unicode single quote (U+0027) should be escaped
				assert.Contains(t, sql, "''")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := search(tt.query)
			sql := f.SQL()
			tt.verify(t, sql)
		})
	}
}

// =============================================================================
// String Escaping Edge Case Tests
// =============================================================================

func TestStringEscaping_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected escaped result in the SQL
	}{
		{
			name:     "single quote",
			input:    "it's",
			expected: "title = 'it''s'",
		},
		{
			name:     "multiple single quotes",
			input:    "it's John's",
			expected: "title = 'it''s John''s'",
		},
		{
			name:     "backslash",
			input:    "path\\to\\file",
			expected: "title = 'path\\to\\file'", // SQLite doesn't treat \ specially
		},
		{
			name:     "double quote",
			input:    `say "hello"`,
			expected: `title = 'say "hello"'`, // Double quotes don't need escaping in SQLite strings
		},
		{
			name:     "newline",
			input:    "line1\nline2",
			expected: "title = 'line1\nline2'", // Newlines are valid in SQLite strings
		},
		{
			name:     "tab",
			input:    "col1\tcol2",
			expected: "title = 'col1\tcol2'",
		},
		{
			name:     "carriage return",
			input:    "line1\r\nline2",
			expected: "title = 'line1\r\nline2'",
		},
		{
			name:     "percent sign",
			input:    "100% complete",
			expected: "title = '100% complete'",
		},
		{
			name:     "underscore",
			input:    "task_name",
			expected: "title = 'task_name'",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "title = ''",
		},
		{
			name:     "only quotes",
			input:    "'''",
			expected: "title = ''''''''", // Three quotes ' become '' each = 6 quotes, plus wrapping = 8
		},
		{
			name:     "unicode characters",
			input:    "Hello",
			expected: "title = 'Hello'",
		},
		{
			name:     "emoji",
			input:    "Task done",
			expected: "title = 'Task done'",
		},
		{
			name:     "chinese characters",
			input:    "Test",
			expected: "title = 'Test'",
		},
		{
			name:     "mixed special chars",
			input:    "John's \"task\" @ 100%",
			expected: `title = 'John''s "task" @ 100%'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := equal("title", tt.input)
			assert.Equal(t, tt.expected, f.SQL())
		})
	}
}

func TestSearchEscaping_LIKEWildcards(t *testing.T) {
	// LIKE wildcards (% and _) in SQLite have special meaning
	// They should be passed through as-is since they're valid search terms
	// but quotes must be escaped
	tests := []struct {
		name   string
		query  string
		verify func(t *testing.T, sql string)
	}{
		{
			name:  "percent wildcard",
			query: "test%",
			verify: func(t *testing.T, sql string) {
				// % should be present (it's valid)
				assert.Contains(t, sql, "%test%%")
			},
		},
		{
			name:  "underscore wildcard",
			query: "test_name",
			verify: func(t *testing.T, sql string) {
				// _ should be present (it's valid in LIKE)
				assert.Contains(t, sql, "%test_name%")
			},
		},
		{
			name:  "quote with wildcard",
			query: "it's%",
			verify: func(t *testing.T, sql string) {
				// Quote should be escaped, % preserved
				assert.Contains(t, sql, "%it''s%%")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := search(tt.query)
			sql := f.SQL()
			tt.verify(t, sql)
		})
	}
}

// =============================================================================
// Filter Combination Tests
// =============================================================================

func TestFilterCombination_ANDCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		filters  []filter
		contains []string
	}{
		{
			name: "two conditions",
			filters: []filter{
				static("type = 0"),
				static("status = 0"),
			},
			contains: []string{"type = 0", "AND", "status = 0"},
		},
		{
			name: "three conditions",
			filters: []filter{
				static("type = 0"),
				static("status = 0"),
				static("trashed = 0"),
			},
			contains: []string{"type = 0", "AND", "status = 0", "AND", "trashed = 0"},
		},
		{
			name: "mixed filter types",
			filters: []filter{
				static("type = 0"),
				equal("uuid", "test-uuid"),
				truthy("recurring", ptrBool(true)),
			},
			contains: []string{"type = 0", "uuid = 'test-uuid'", "recurring"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := filters(tt.filters)
			sql := fs.SQL()

			for _, c := range tt.contains {
				assert.Contains(t, sql, c)
			}
		})
	}
}

func TestFilterCombination_NestedORInAND(t *testing.T) {
	// Test that OR filters work correctly within AND combinations
	fb := newFilterBuilder().
		addStatic("type = 0").
		addOr(
			static("area = 'area1'"),
			static("area = 'area2'"),
			static("area = 'area3'"),
		).
		addStatic("status = 0")

	sql := fb.sql()

	// Verify structure
	assert.Contains(t, sql, "type = 0")
	assert.Contains(t, sql, "AND (area = 'area1' OR area = 'area2' OR area = 'area3')")
	assert.Contains(t, sql, "AND status = 0")
}

// =============================================================================
// NULL Handling in Filters
// =============================================================================

func TestNullHandling_InFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   filter
		expected string
		isEmpty  bool
	}{
		{
			name:     "bool true means IS NOT NULL",
			filter:   equal("deadline", true),
			expected: "deadline IS NOT NULL",
			isEmpty:  false,
		},
		{
			name:     "bool false means IS NULL",
			filter:   equal("deadline", false),
			expected: "deadline IS NULL",
			isEmpty:  false,
		},
		{
			name:     "nil value is empty",
			filter:   equal("deadline", nil),
			expected: "",
			isEmpty:  true,
		},
		{
			name:     "truthy nil pointer is empty",
			filter:   truthy("recurring", nil),
			expected: "",
			isEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.filter.SQL())
			assert.Equal(t, tt.isEmpty, tt.filter.IsEmpty())
		})
	}
}

// =============================================================================
// Date Filter Boundary Tests
// =============================================================================

func TestDateFilter_YearBoundary(t *testing.T) {
	tests := []struct {
		name  string
		date  string
		valid bool // Whether it should produce valid SQL
	}{
		{"year 2024", "2024-06-15", true},
		{"year 2000", "2000-01-01", true},
		{"year 1999", "1999-12-31", true},
		{"year 2099", "2099-12-31", true},
		{"invalid year", "99-06-15", false},
		{"invalid month", "2024-13-01", false},
		{"invalid day", "2024-06-32", false},
		{"leap year", "2024-02-29", true},
		{"non-leap year feb 29", "2023-02-29", false}, // Invalid date
		{"year boundary jan 1", "2024-01-01", true},
		{"year boundary dec 31", "2024-12-31", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := thingsDate("startDate", dateOpEqual, tt.date)
			sql := f.SQL()

			if tt.valid {
				// Valid dates should produce non-empty SQL with the date value
				assert.NotEmpty(t, sql, "valid date %s should produce SQL", tt.date)
				assert.Contains(t, sql, "startDate =")
			} else {
				// Invalid dates should produce empty SQL
				assert.Empty(t, sql, "invalid date %s should produce empty SQL", tt.date)
			}
		})
	}
}

// =============================================================================
// Helper Functions for Tests
// =============================================================================

func ptrBool(b bool) *bool {
	return &b
}
