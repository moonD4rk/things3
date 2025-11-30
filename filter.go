package things3

import (
	"fmt"
	"regexp"
	"strings"
)

// Date filter value constants.
const (
	dateFuture          = "future"
	datePast            = "past"
	defaultDateOperator = "=="
)

// matchDateWithOperator matches an ISO 8601 date string with optional comparison operator.
// Returns [fullMatch, operator, date] or nil if no match.
func matchDateWithOperator(value string) []string {
	re := regexp.MustCompile(`^(=|==|<|<=|>|>=)?(\d{4}-\d{2}-\d{2})$`)
	matches := re.FindStringSubmatch(value)
	if len(matches) != 3 {
		return nil
	}
	return matches
}

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
		return fmt.Sprintf("%s > %s", f.column, TodayThingsDateSQL())
	case dateOpPast:
		return fmt.Sprintf("%s <= %s", f.column, TodayThingsDateSQL())
	case dateOpEqual, dateOpBefore, dateOpBeforeEq, dateOpAfter, dateOpAfterEq:
		if f.value == "" {
			return ""
		}
		thingsDate, err := StringToThingsDate(f.value)
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

// unixTimeRangeFilter handles "last X days/weeks/years" range filters.
type unixTimeRangeFilter struct {
	column string
	offset string // e.g., "7d", "2w", "1y"
}

func (f unixTimeRangeFilter) SQL() string {
	if f.offset == "" || len(f.offset) < 2 {
		return ""
	}

	number := f.offset[:len(f.offset)-1]
	suffix := f.offset[len(f.offset)-1]

	var modifier string
	switch suffix {
	case 'd':
		modifier = fmt.Sprintf("-%s days", number)
	case 'w':
		var weeks int
		if _, err := fmt.Sscanf(number, "%d", &weeks); err != nil {
			return ""
		}
		modifier = fmt.Sprintf("-%d days", weeks*7)
	case 'y':
		modifier = fmt.Sprintf("-%s years", number)
	default:
		return ""
	}

	columnDatetime := fmt.Sprintf("datetime(%s, 'unixepoch', 'localtime')", f.column)
	offsetDatetime := fmt.Sprintf("datetime('now', '%s')", modifier)

	return fmt.Sprintf("%s > %s", columnDatetime, offsetDatetime)
}

func (f unixTimeRangeFilter) IsEmpty() bool {
	if f.offset == "" || len(f.offset) < 2 {
		return true
	}
	suffix := f.offset[len(f.offset)-1]
	return suffix != 'd' && suffix != 'w' && suffix != 'y'
}

// unixTimeRange creates a filter for items within the last X days/weeks/years.
// offset format: "7d" for 7 days, "2w" for 2 weeks, "1y" for 1 year.
func unixTimeRange(column, offset string) filter {
	return unixTimeRangeFilter{column: column, offset: offset}
}

// addThingsDateValue adds a Things date filter to the builder.
func (b *filterBuilder) addThingsDateValue(column string, op dateOp, value string) *filterBuilder {
	return b.add(thingsDate(column, op, value))
}

// addUnixTimeValue adds a Unix time filter to the builder.
func (b *filterBuilder) addUnixTimeValue(column string, op dateOp, value string) *filterBuilder {
	return b.add(unixTime(column, op, value))
}

// addUnixTimeRangeValue adds a Unix time range filter to the builder.
func (b *filterBuilder) addUnixTimeRangeValue(column, offset string) *filterBuilder {
	return b.add(unixTimeRange(column, offset))
}

// parsedDateFilter creates a filter from any-typed date value (bool, string).
// This bridges the old API (accepting any) with the new Filter abstraction.
type parsedDateFilter struct {
	column   string
	value    any
	isThings bool // true for Things date format, false for Unix timestamp
}

func (f parsedDateFilter) SQL() string {
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
		if v == "" {
			return ""
		}
		if f.isThings {
			return f.parseThingsDateString(v)
		}
		return f.parseUnixTimeString(v)
	default:
		return ""
	}
}

func (f parsedDateFilter) parseThingsDateString(value string) string {
	switch value {
	case dateFuture:
		return fmt.Sprintf("%s > %s", f.column, TodayThingsDateSQL())
	case datePast:
		return fmt.Sprintf("%s <= %s", f.column, TodayThingsDateSQL())
	}

	match := matchDateWithOperator(value)
	if match == nil {
		return ""
	}

	operator, isoDate := match[1], match[2]
	if operator == "" {
		operator = defaultDateOperator
	}

	thingsDate, err := StringToThingsDate(isoDate)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s %s %d", f.column, operator, thingsDate)
}

func (f parsedDateFilter) parseUnixTimeString(value string) string {
	dateExpr := fmt.Sprintf("date(%s, 'unixepoch', 'localtime')", f.column)

	switch value {
	case dateFuture:
		return fmt.Sprintf("%s > date('now', 'localtime')", dateExpr)
	case datePast:
		return fmt.Sprintf("%s <= date('now', 'localtime')", dateExpr)
	}

	match := matchDateWithOperator(value)
	if match == nil {
		return ""
	}

	operator, isoDate := match[1], match[2]
	if operator == "" {
		operator = defaultDateOperator
	}

	return fmt.Sprintf("%s %s date('%s')", dateExpr, operator, isoDate)
}

func (f parsedDateFilter) IsEmpty() bool {
	if f.value == nil {
		return true
	}
	if s, ok := f.value.(string); ok && s == "" {
		return true
	}
	return false
}

// parsedThingsDate creates a filter from an any-typed Things date value.
// Accepts: bool (IS NULL/IS NOT NULL), "future", "past", or ISO date with optional operator.
func parsedThingsDate(column string, value any) filter {
	return parsedDateFilter{column: column, value: value, isThings: true}
}

// parsedUnixTime creates a filter from an any-typed Unix timestamp value.
// Accepts: bool (IS NULL/IS NOT NULL), "future", "past", or ISO date with optional operator.
func parsedUnixTime(column string, value any) filter {
	return parsedDateFilter{column: column, value: value, isThings: false}
}

// addParsedThingsDateValue adds a parsed Things date filter to the builder.
//
//nolint:unparam
func (b *filterBuilder) addParsedThingsDateValue(column string, value any) *filterBuilder {
	return b.add(parsedThingsDate(column, value))
}

// addParsedUnixTimeValue adds a parsed Unix time filter to the builder.
func (b *filterBuilder) addParsedUnixTimeValue(column string, value any) *filterBuilder {
	return b.add(parsedUnixTime(column, value))
}
