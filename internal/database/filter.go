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
		return "TRUE"
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

// addOrFilter adds a filter across two columns with value-or-existence fallback.
// A value or exists=true matches if either column matches (OR). For
// exists=false, De Morgan requires both columns to be NULL (AND), otherwise
// "has neither" would wrongly match rows where only one column is set.
func (w *whereBuilder) addOrFilter(col1, col2 string, value *string, exists *bool) {
	if value != nil {
		escaped := escapeString(*value)
		w.addOr(
			fmt.Sprintf("%s = '%s'", col1, escaped),
			fmt.Sprintf("%s = '%s'", col2, escaped),
		)
	} else if exists != nil {
		if *exists {
			w.addOr(existsSQL(col1, true), existsSQL(col2, true))
		} else {
			w.addRawf("(%s AND %s)", existsSQL(col1, false), existsSQL(col2, false))
		}
	}
}

// addIntEqual adds an integer equality condition (skips nil).
func (w *whereBuilder) addIntEqual(column string, value *int) {
	if value != nil {
		w.addRawf("%s = %d", column, *value)
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

// addOr adds an OR combination of conditions (skips empty parts).
func (w *whereBuilder) addOr(parts ...string) {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) > 0 {
		*w = append(*w, "("+strings.Join(nonEmpty, " OR ")+")")
	}
}

// addSearch adds a full-text search condition across multiple columns.
// LIKE metacharacters in the query match literally.
func (w *whereBuilder) addSearch(query string) {
	if query == "" {
		return
	}
	columns := []string{"TASK.title", "TASK.notes", "AREA.title"}
	var searches []string
	for _, col := range columns {
		searches = append(searches, likeSQL(col, "%", query, "%"))
	}
	*w = append(*w, "("+strings.Join(searches, " OR ")+")")
}

// addCreatedAfter adds a time-based filter for creation date.
// The instant is normalized to local time so the same instant yields
// identical SQL regardless of the Location carried by t.
func (w *whereBuilder) addCreatedAfter(column string, t time.Time) {
	if t.IsZero() {
		return
	}
	local := t.In(time.Local).Format("2006-01-02 15:04:05")
	w.addRawf("datetime(%s, 'unixepoch', 'localtime') > '%s'", column, local)
}

// addDateFilter adds a date filter condition.
// isThingsDate indicates whether the column uses Things binary date format (true)
// or Unix timestamp format (false).
func (w *whereBuilder) addDateFilter(column string, v *DateFilterValue, isThingsDate bool) {
	if v == nil {
		return
	}

	// Existence check (format-independent)
	if v.HasDate != nil {
		w.addExists(column, *v.HasDate)
		return
	}

	// Resolve format-specific expressions upfront
	var colExpr, nowExpr string
	if isThingsDate {
		colExpr = column
		nowExpr = todayThingsDateSQL()
	} else {
		colExpr = fmt.Sprintf("date(%s, 'unixepoch', 'localtime')", column)
		nowExpr = "date('now', 'localtime')"
	}

	// Relative date (future/past)
	if v.Relative != "" {
		if v.Relative == DateFuture {
			w.addRawf("%s > %s", colExpr, nowExpr)
		} else {
			w.addRawf("%s <= %s", colExpr, nowExpr)
		}
		return
	}

	// Specific date comparison. Normalize the instant to local time so the
	// same instant yields the same calendar date regardless of its Location.
	if v.Date == nil {
		return
	}
	dateVal, ok := formatDateValue(v.Date.In(time.Local).Format(time.DateOnly), isThingsDate)
	if !ok {
		return
	}
	w.addRawf("%s %s %s", colExpr, v.Operator, dateVal)
}

// formatDateValue converts a date string to the appropriate SQL value.
// Returns the formatted value and true on success.
func formatDateValue(dateStr string, isThingsDate bool) (string, bool) {
	if isThingsDate {
		td, err := stringToThingsDate(dateStr)
		if err != nil || td == 0 {
			return "", false
		}
		return strconv.FormatInt(td, 10), true
	}
	return fmt.Sprintf("date('%s')", dateStr), true
}

// existsSQL returns "column IS [NOT] NULL" as a SQL fragment.
func existsSQL(column string, exists bool) string {
	if exists {
		return column + " IS NOT NULL"
	}
	return column + " IS NULL"
}

// sql returns the combined SQL for all conditions.
func (w *whereBuilder) sql() string {
	return joinConditions(*w)
}
