package database

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Date filter value constants.
const (
	DateFuture = "future"
	DatePast   = "past"
)

// DateFilterValue holds a parsed date filter configuration.
// Only one field should be set at a time.
type DateFilterValue struct {
	HasDate  *bool      // true/false for existence check
	Relative string     // "future" or "past"
	Operator string     // "=", "<", "<=", ">", ">="
	Date     *time.Time // specific date for comparison
}

// escapeString escapes a string for safe use in SQL queries.
func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// likeEscapeChar is the escape character for LIKE patterns, declared once so
// escaped patterns and their ESCAPE clauses stay in sync.
const likeEscapeChar = `\`

// escapeLikePattern escapes LIKE metacharacters (%, _) and the escape
// character itself so user input matches literally inside a LIKE pattern.
// The pattern must be used with an ESCAPE clause (see likeSQL).
func escapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, likeEscapeChar, likeEscapeChar+likeEscapeChar)
	s = strings.ReplaceAll(s, "%", likeEscapeChar+"%")
	s = strings.ReplaceAll(s, "_", likeEscapeChar+"_")
	return s
}

// likeSQL returns "column LIKE 'pattern' ESCAPE '\'" where value is matched
// literally and prefix/suffix hold the intended wildcards ("%" or "").
func likeSQL(column, prefix, value, suffix string) string {
	escaped := escapeString(escapeLikePattern(value))
	return fmt.Sprintf("%s LIKE '%s%s%s' ESCAPE '%s'", column, prefix, escaped, suffix, likeEscapeChar)
}

// joinConditions joins SQL conditions with AND, returns "TRUE" if empty.
func joinConditions(conditions []string) string {
	if len(conditions) == 0 {
		return sqlTrue
	}
	return strings.Join(conditions, "\n            AND ")
}

// whereBuilder collects SQL WHERE conditions.
type whereBuilder []string

// add appends a raw SQL condition (skips empty strings).
func (w *whereBuilder) add(sql string) {
	if sql != "" {
		*w = append(*w, sql)
	}
}

// addRawf appends a formatted SQL condition.
func (w *whereBuilder) addRawf(format string, args ...any) {
	*w = append(*w, fmt.Sprintf(format, args...))
}

// addStringEqual adds a string equality condition (skips nil).
func (w *whereBuilder) addStringEqual(column string, value *string) {
	if value != nil {
		w.addRawf("%s = '%s'", column, escapeString(*value))
	}
}

// addExists adds "column IS NOT NULL" (true) or "column IS NULL" (false).
func (w *whereBuilder) addExists(column string, exists bool) {
	if exists {
		*w = append(*w, column+" IS NOT NULL")
	} else {
		*w = append(*w, column+" IS NULL")
	}
}

// addFilter adds a column filter: matches value if set, otherwise checks existence.
// Handles the common "specific value takes precedence over has/no" pattern.
func (w *whereBuilder) addFilter(column string, value *string, exists *bool) {
	if value != nil {
		w.addStringEqual(column, value)
	} else if exists != nil {
		w.addExists(column, *exists)
	}
}

// addTagFilter adds a tag condition built from exists, an EXISTS predicate over
// the owner's tags: a specific title takes precedence over has/has-not.
func (w *whereBuilder) addTagFilter(exists func(cond string) string, title *string, has *bool) {
	switch {
	case title != nil:
		w.add(exists(fmt.Sprintf("TAG.title = '%s'", escapeString(*title))))
	case has == nil:
	case *has:
		w.add(exists(""))
	default:
		w.add("NOT " + exists(""))
	}
}

// addIntEqual adds an integer equality condition (skips nil).
func (w *whereBuilder) addIntEqual(column string, value *int) {
	if value != nil {
		w.addRawf("%s = %d", column, *value)
	}
}

// addIntIn adds an integer membership condition (skips an empty set).
func (w *whereBuilder) addIntIn(column string, values []int) {
	switch len(values) {
	case 0:
	case 1:
		w.addRawf("%s = %d", column, values[0])
	default:
		parts := make([]string, len(values))
		for i, v := range values {
			parts[i] = strconv.Itoa(v)
		}
		w.addRawf("%s IN (%s)", column, strings.Join(parts, ", "))
	}
}

// addLikePrefix adds a prefix-match condition; LIKE metacharacters in value
// match literally.
func (w *whereBuilder) addLikePrefix(column, value string) {
	if value != "" {
		w.add(likeSQL(column, "", value, "%"))
	}
}

// addLikeContains adds a substring-match condition; LIKE metacharacters in
// value match literally.
func (w *whereBuilder) addLikeContains(column, value string) {
	if value != "" {
		w.add(likeSQL(column, "%", value, "%"))
	}
}

// addTruthy adds a boolean column check, treating NULL as nullDefault.
// Pass 0 when NULL means false (e.g. trashed) and 1 when NULL means true
// (e.g. AREA.visible, which is NULL until the user hides the area).
//   - true:  "IFNULL(column, nullDefault)"
//   - false: "NOT IFNULL(column, nullDefault)"
func (w *whereBuilder) addTruthy(column string, value *bool, nullDefault int) {
	if value == nil {
		return
	}
	if *value {
		w.addRawf("IFNULL(%s, %d)", column, nullDefault)
	} else {
		w.addRawf("NOT IFNULL(%s, %d)", column, nullDefault)
	}
}

// addSearch matches the query against an item's own title and notes. LIKE
// metacharacters in the query match literally.
func (w *whereBuilder) addSearch(query string) {
	if query == "" {
		return
	}
	columns := []string{"TASK.title", "TASK.notes"}
	var searches []string
	for _, col := range columns {
		searches = append(searches, likeSQL(col, "%", query, "%"))
	}
	*w = append(*w, "("+strings.Join(searches, " OR ")+")")
}

// addCreatedAfter keeps rows created strictly after t. Creation dates are Unix
// seconds, so the comparison is an instant comparison with no time zone in it.
func (w *whereBuilder) addCreatedAfter(column string, t time.Time) {
	if t.IsZero() {
		return
	}
	w.addRawf("%s > %s", column, strconv.FormatFloat(float64(t.UnixMicro())/1e6, 'f', 6, 64))
}

// addThingsDateFilter adds a filter on a packed Things date expression.
func (w *whereBuilder) addThingsDateFilter(expr string, v *DateFilterValue, today time.Time) {
	w.addDateFilter(expr, v,
		strconv.FormatInt(timeToThingsDate(today.In(time.Local)), 10),
		func(day time.Time) string { return strconv.FormatInt(timeToThingsDate(day), 10) },
	)
}

// addUnixDateFilter adds a filter on the local calendar day of a Unix-seconds
// column.
func (w *whereBuilder) addUnixDateFilter(column string, v *DateFilterValue, today time.Time) {
	quoted := func(day time.Time) string { return "'" + day.Format(time.DateOnly) + "'" }
	w.addDateFilter(fmt.Sprintf("date(%s, 'unixepoch', 'localtime')", column), v, quoted(today.In(time.Local)), quoted)
}

// addDateFilter adds a date condition on expr. todaySQL is today's value and
// literal renders any other day, both in expr's own encoding.
func (w *whereBuilder) addDateFilter(expr string, v *DateFilterValue, todaySQL string, literal func(time.Time) string) {
	switch {
	case v == nil:
	case v.HasDate != nil:
		w.addExists(expr, *v.HasDate)
	case v.Relative == DateFuture:
		w.addRawf("%s > %s", expr, todaySQL)
	case v.Relative == DatePast:
		w.addRawf("%s <= %s", expr, todaySQL)
	case v.Date != nil:
		w.addRawf("%s %s %s", expr, v.Operator, literal(clampDate(v.Date.In(time.Local))))
	}
}

// Year bounds a comparison date can take: both date encodings are only
// meaningful for four-digit years.
const (
	minDateYear = 1
	maxDateYear = 9999
)

// clampDate bounds a comparison date to the representable year range, so a
// bound past either end still compares (admitting every dated row) instead of
// producing a literal SQLite reads as NULL. The floor is the second of January
// because the first is Go's zero time, which packs to "no date".
func clampDate(t time.Time) time.Time {
	switch {
	case t.Year() < minDateYear:
		return time.Date(minDateYear, time.January, 2, 0, 0, 0, 0, t.Location())
	case t.Year() > maxDateYear:
		return time.Date(maxDateYear, time.December, 31, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

// sql returns the combined SQL for all conditions.
func (w *whereBuilder) sql() string {
	return joinConditions(*w)
}
