package things3

import (
	"fmt"
	"strings"
	"time"
)

// Date filter value constants.
const (
	dateFuture = "future"
	datePast   = "past"
)

// filter represents a SQL WHERE clause condition.
// filter implementations return SQL fragments without "AND " prefix.
type filter interface {
	// SQL returns the SQL fragment for this filter condition.
	SQL() string

	// IsEmpty returns true if this filter has no effect and should be skipped.
	IsEmpty() bool
}

// filters is a collection of filters that combines them with AND.
type filters []filter

// SQL returns all non-empty filters joined with "AND".
// Returns "TRUE" if no filters are present.
func (f filters) SQL() string {
	var parts []string
	for _, fltr := range f {
		if !fltr.IsEmpty() {
			parts = append(parts, fltr.SQL())
		}
	}
	if len(parts) == 0 {
		return "TRUE"
	}
	return strings.Join(parts, "\n            AND ")
}

// IsEmpty returns true if all filters are empty.
func (f filters) IsEmpty() bool {
	for _, fltr := range f {
		if !fltr.IsEmpty() {
			return false
		}
	}
	return true
}

// staticFilter wraps a constant SQL expression.
type staticFilter string

func (f staticFilter) SQL() string   { return string(f) }
func (f staticFilter) IsEmpty() bool { return f == "" }

// static creates a filter from a raw SQL expression.
// Use this for constant filter expressions like "type = 0" or "trashed = 1".
func static(expr string) filter {
	return staticFilter(expr)
}

// equalFilter handles column = value comparisons with type-aware formatting.
type equalFilter struct {
	column string
	value  any
}

func (f equalFilter) SQL() string {
	if f.value == nil {
		return ""
	}
	switch v := f.value.(type) {
	case bool:
		if v {
			return fmt.Sprintf("%s IS NOT NULL", f.column)
		}
		return fmt.Sprintf("%s IS NULL", f.column)
	case string:
		return fmt.Sprintf("%s = '%s'", f.column, escapeString(v))
	default:
		return fmt.Sprintf("%s = '%v'", f.column, v)
	}
}

func (f equalFilter) IsEmpty() bool {
	return f.value == nil
}

// equal creates an equality filter for a column.
// - nil value: returns empty filter (no condition)
// - bool true: returns "column IS NOT NULL"
// - bool false: returns "column IS NULL"
// - string: returns "column = 'escaped_value'"
// - other: returns "column = 'value'"
func equal(column string, value any) filter {
	return equalFilter{column: column, value: value}
}

// likeFilter handles LIKE pattern matching.
type likeFilter struct {
	column  string
	pattern string
}

func (f likeFilter) SQL() string {
	if f.pattern == "" {
		return ""
	}
	return fmt.Sprintf("%s LIKE '%s'", f.column, escapeString(f.pattern))
}

func (f likeFilter) IsEmpty() bool {
	return f.pattern == ""
}

// like creates a LIKE pattern filter for a column.
// The pattern should include wildcards (% or _) as needed.
func like(column, pattern string) filter {
	return likeFilter{column: column, pattern: pattern}
}

// truthyFilter handles boolean column checks with NULL handling.
type truthyFilter struct {
	column string
	value  *bool
}

func (f truthyFilter) SQL() string {
	if f.value == nil {
		return ""
	}
	if *f.value {
		return f.column
	}
	return fmt.Sprintf("NOT IFNULL(%s, 0)", f.column)
}

func (f truthyFilter) IsEmpty() bool {
	return f.value == nil
}

// truthy creates a filter that checks if a column is truthy or falsy.
// - nil: returns empty filter
// - true: returns "column" (truthy check)
// - false: returns "NOT IFNULL(column, 0)" (falsy check, treating NULL as false)
func truthy(column string, value *bool) filter {
	return truthyFilter{column: column, value: value}
}

// orFilter combines multiple filters with OR.
type orFilter struct {
	filters []filter
}

func (f orFilter) SQL() string {
	var parts []string
	for _, fltr := range f.filters {
		if !fltr.IsEmpty() {
			parts = append(parts, fltr.SQL())
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, " OR "))
}

func (f orFilter) IsEmpty() bool {
	for _, fltr := range f.filters {
		if !fltr.IsEmpty() {
			return false
		}
	}
	return true
}

// or combines multiple filters with OR logic.
// Empty filters are skipped. Returns empty if all filters are empty.
func or(fltrs ...filter) filter {
	return orFilter{filters: fltrs}
}

// searchFilter handles full-text search across multiple columns.
type searchFilter struct {
	query   string
	columns []string
}

func (f searchFilter) SQL() string {
	if f.query == "" {
		return ""
	}
	escaped := escapeString(f.query)
	var searches []string
	for _, col := range f.columns {
		searches = append(searches, fmt.Sprintf("%s LIKE '%%%s%%'", col, escaped))
	}
	return fmt.Sprintf("(%s)", strings.Join(searches, " OR "))
}

func (f searchFilter) IsEmpty() bool {
	return f.query == ""
}

// search creates a filter that searches for text across multiple columns.
// The query is escaped and wrapped with % wildcards for LIKE matching.
func search(query string, columns ...string) filter {
	if len(columns) == 0 {
		columns = []string{"TASK.title", "TASK.notes", "AREA.title"}
	}
	return searchFilter{query: query, columns: columns}
}

// filterBuilder provides a fluent interface for building filter collections.
type filterBuilder struct {
	fltrs filters
}

// newFilterBuilder creates a new filter builder.
func newFilterBuilder() *filterBuilder {
	return &filterBuilder{}
}

// add appends a filter to the builder.
func (b *filterBuilder) add(f filter) *filterBuilder {
	b.fltrs = append(b.fltrs, f)
	return b
}

// addStatic adds a static SQL expression filter.
func (b *filterBuilder) addStatic(expr string) *filterBuilder {
	return b.add(static(expr))
}

// addEqual adds an equality filter.
func (b *filterBuilder) addEqual(column string, value any) *filterBuilder {
	return b.add(equal(column, value))
}

// addLike adds a LIKE pattern filter.
func (b *filterBuilder) addLike(column, pattern string) *filterBuilder {
	return b.add(like(column, pattern))
}

// addTruthy adds a truthy/falsy filter.
func (b *filterBuilder) addTruthy(column string, value *bool) *filterBuilder {
	return b.add(truthy(column, value))
}

// addOr adds an OR combination of filters.
func (b *filterBuilder) addOr(fltrs ...filter) *filterBuilder {
	return b.add(or(fltrs...))
}

// addSearch adds a full-text search filter.
//
//nolint:unparam
func (b *filterBuilder) addSearch(query string, columns ...string) *filterBuilder {
	return b.add(search(query, columns...))
}

// build returns the collected filters.
func (b *filterBuilder) build() filters {
	return b.fltrs
}

// sql returns the combined SQL for all filters.
func (b *filterBuilder) sql() string {
	return b.fltrs.SQL()
}

// thingsDateFilter handles Things binary date format comparisons.
type thingsDateFilter struct {
	column string
	op     dateOp
	value  string // ISO date for comparison ops, empty for exists/future/past
}

func (f thingsDateFilter) SQL() string {
	switch f.op {
	case dateOpExists:
		return fmt.Sprintf("%s IS NOT NULL", f.column)
	case dateOpNotExists:
		return fmt.Sprintf("%s IS NULL", f.column)
	case dateOpFuture:
		return fmt.Sprintf("%s > %s", f.column, todayThingsDateSQL())
	case dateOpPast:
		return fmt.Sprintf("%s <= %s", f.column, todayThingsDateSQL())
	case dateOpEqual, dateOpBefore, dateOpBeforeEq, dateOpAfter, dateOpAfterEq:
		if f.value == "" {
			return ""
		}
		thingsDate, err := stringToThingsDate(f.value)
		if err != nil || thingsDate == 0 {
			return ""
		}
		return fmt.Sprintf("%s %s %d", f.column, f.op.SQLOperator(), thingsDate)
	}
	return ""
}

func (f thingsDateFilter) IsEmpty() bool {
	switch f.op {
	case dateOpExists, dateOpNotExists, dateOpFuture, dateOpPast:
		return false
	default:
		return f.value == ""
	}
}

// thingsDate creates a filter for Things binary date format columns.
// Use dateOpExists/dateOpNotExists to check for presence.
// Use dateOpFuture/dateOpPast for relative comparisons to today.
// Use dateOpEqual/Before/After with an ISO date string (YYYY-MM-DD) for specific dates.
func thingsDate(column string, op dateOp, value string) filter {
	return thingsDateFilter{column: column, op: op, value: value}
}

// unixTimeFilter handles Unix timestamp column comparisons.
type unixTimeFilter struct {
	column string
	op     dateOp
	value  string // ISO date for comparison ops, empty for exists/future/past
}

func (f unixTimeFilter) SQL() string {
	dateExpr := fmt.Sprintf("date(%s, 'unixepoch', 'localtime')", f.column)

	switch f.op {
	case dateOpExists:
		return fmt.Sprintf("%s IS NOT NULL", f.column)
	case dateOpNotExists:
		return fmt.Sprintf("%s IS NULL", f.column)
	case dateOpFuture:
		return fmt.Sprintf("%s > date('now', 'localtime')", dateExpr)
	case dateOpPast:
		return fmt.Sprintf("%s <= date('now', 'localtime')", dateExpr)
	case dateOpEqual, dateOpBefore, dateOpBeforeEq, dateOpAfter, dateOpAfterEq:
		if f.value == "" {
			return ""
		}
		return fmt.Sprintf("%s %s date('%s')", dateExpr, f.op.SQLOperator(), f.value)
	}
	return ""
}

func (f unixTimeFilter) IsEmpty() bool {
	switch f.op {
	case dateOpExists, dateOpNotExists, dateOpFuture, dateOpPast:
		return false
	default:
		return f.value == ""
	}
}

// unixTime creates a filter for Unix timestamp columns.
// Use dateOpExists/dateOpNotExists to check for presence.
// Use dateOpFuture/dateOpPast for relative comparisons to today.
// Use dateOpEqual/Before/After with an ISO date string (YYYY-MM-DD) for specific dates.
func unixTime(column string, op dateOp, value string) filter {
	return unixTimeFilter{column: column, op: op, value: value}
}

// createdAfterFilter handles time-based filters for creation date.
type createdAfterFilter struct {
	column string
	after  time.Time
}

func (f createdAfterFilter) SQL() string {
	if f.after.IsZero() {
		return ""
	}
	columnDatetime := fmt.Sprintf("datetime(%s, 'unixepoch', 'localtime')", f.column)
	// Format as ISO 8601 for SQLite datetime comparison
	afterStr := f.after.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s > '%s'", columnDatetime, afterStr)
}

func (f createdAfterFilter) IsEmpty() bool {
	return f.after.IsZero()
}

// createdAfter creates a filter for items created after the specified time.
func createdAfter(column string, t time.Time) filter {
	return createdAfterFilter{column: column, after: t}
}

// addThingsDateValue adds a Things date filter to the builder.
func (b *filterBuilder) addThingsDateValue(column string, op dateOp, value string) *filterBuilder {
	return b.add(thingsDate(column, op, value))
}

// addUnixTimeValue adds a Unix time filter to the builder.
func (b *filterBuilder) addUnixTimeValue(column string, op dateOp, value string) *filterBuilder {
	return b.add(unixTime(column, op, value))
}

// addCreatedAfterFilter adds a time-based filter to the builder.
func (b *filterBuilder) addCreatedAfterFilter(column string, t time.Time) *filterBuilder {
	return b.add(createdAfter(column, t))
}

// addDateFilterValue adds a date filter using the new type-safe dateFilterValue.
// isThingsDate indicates whether the column uses Things binary date format (true)
// or Unix timestamp format (false).
//
//nolint:unparam // return value unused but kept for fluent builder pattern consistency
func (b *filterBuilder) addDateFilterValue(column string, v *dateFilterValue, isThingsDate bool) *filterBuilder {
	if v == nil {
		return b
	}

	// Handle existence check
	if v.hasDate != nil {
		if *v.hasDate {
			return b.add(static(fmt.Sprintf("%s IS NOT NULL", column)))
		}
		return b.add(static(fmt.Sprintf("%s IS NULL", column)))
	}

	// Handle relative date (future/past)
	if v.relative != "" {
		if isThingsDate {
			if v.relative == dateFuture {
				return b.add(static(fmt.Sprintf("%s > %s", column, todayThingsDateSQL())))
			}
			return b.add(static(fmt.Sprintf("%s <= %s", column, todayThingsDateSQL())))
		}
		// Unix timestamp format
		dateExpr := fmt.Sprintf("date(%s, 'unixepoch', 'localtime')", column)
		if v.relative == dateFuture {
			return b.add(static(fmt.Sprintf("%s > date('now', 'localtime')", dateExpr)))
		}
		return b.add(static(fmt.Sprintf("%s <= date('now', 'localtime')", dateExpr)))
	}

	// Handle specific date comparison
	if v.date != nil {
		// Convert time.Time to ISO date string (YYYY-MM-DD)
		dateStr := v.date.Format(time.DateOnly)
		if isThingsDate {
			thingsDate, err := stringToThingsDate(dateStr)
			if err != nil || thingsDate == 0 {
				return b
			}
			return b.add(static(fmt.Sprintf("%s %s %d", column, v.operator, thingsDate)))
		}
		// Unix timestamp format
		dateExpr := fmt.Sprintf("date(%s, 'unixepoch', 'localtime')", column)
		return b.add(static(fmt.Sprintf("%s %s date('%s')", dateExpr, v.operator, dateStr)))
	}

	return b
}
