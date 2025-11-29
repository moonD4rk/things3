package things3

import (
	"fmt"
	"regexp"
	"strings"
)

// SQL constant for default WHERE predicate.
const sqlTrue = "TRUE"

// escapeString escapes a string for safe use in SQL queries.
// In SQLite, single quotes within strings are escaped by doubling them.
func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// makeFilter creates a SQL filter clause for a column and value.
// Returns empty string if value is nil.
// Returns "IS NULL" or "IS NOT NULL" for bool values.
func makeFilter(column string, value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case bool:
		if v {
			return fmt.Sprintf("AND %s IS NOT NULL", column)
		}
		return fmt.Sprintf("AND %s IS NULL", column)
	case string:
		return fmt.Sprintf("AND %s = '%s'", column, escapeString(v))
	default:
		return fmt.Sprintf("AND %s = '%v'", column, v)
	}
}

// makeTruthyFilter creates a SQL filter that matches truthy or falsy values.
// Truthy means TRUE. Falsy means FALSE or NULL.
func makeTruthyFilter(column string, value *bool) string {
	if value == nil {
		return ""
	}
	if *value {
		return fmt.Sprintf("AND %s", column)
	}
	return fmt.Sprintf("AND NOT IFNULL(%s, 0)", column)
}

// makeOrFilter joins multiple filters with OR.
func makeOrFilter(filters ...string) string {
	var nonEmpty []string
	for _, f := range filters {
		f = strings.TrimPrefix(f, "AND ")
		if f != "" {
			nonEmpty = append(nonEmpty, f)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	return fmt.Sprintf("AND (%s)", strings.Join(nonEmpty, " OR "))
}

// makeSearchFilter creates a SQL filter for searching tasks by title, notes, and area title.
func makeSearchFilter(query string) string {
	if query == "" {
		return ""
	}

	escaped := escapeString(query)
	columns := []string{"TASK.title", "TASK.notes", "AREA.title"}

	var searches []string
	for _, col := range columns {
		searches = append(searches, fmt.Sprintf("%s LIKE '%%%s%%'", col, escaped))
	}

	return fmt.Sprintf("AND (%s)", strings.Join(searches, " OR "))
}

// makeThingsDateFilter creates a SQL filter for Things date columns.
// Supports: bool (IS NULL/IS NOT NULL), "future", "past", or ISO 8601 date with optional operator.
func makeThingsDateFilter(column string, value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case bool:
		return makeFilter(column, v)
	case string:
		return makeThingsDateStringFilter(column, v)
	default:
		return ""
	}
}

// makeThingsDateStringFilter handles string-based date filters.
func makeThingsDateStringFilter(column, value string) string {
	if value == "" {
		return ""
	}

	// Check for "future" or "past"
	switch value {
	case "future":
		return fmt.Sprintf("AND %s > %s", column, TodayThingsDateSQL())
	case "past":
		return fmt.Sprintf("AND %s <= %s", column, TodayThingsDateSQL())
	}

	// Check for ISO 8601 date with optional operator
	match := matchDateWithOperator(value)
	if match == nil {
		return ""
	}

	operator, isoDate := match[1], match[2]
	if operator == "" {
		operator = "=="
	}

	thingsDate, err := StringToThingsDate(isoDate)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("AND %s %s %d", column, operator, thingsDate)
}

// makeUnixTimeFilter creates a SQL filter for Unix timestamp columns.
func makeUnixTimeFilter(column string, value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case bool:
		return makeFilter(column, v)
	case string:
		return makeUnixTimeStringFilter(column, v)
	default:
		return ""
	}
}

// makeUnixTimeStringFilter handles string-based Unix time filters.
func makeUnixTimeStringFilter(column, value string) string {
	if value == "" {
		return ""
	}

	dateExpr := fmt.Sprintf("date(%s, 'unixepoch', 'localtime')", column)

	switch value {
	case "future":
		return fmt.Sprintf("AND %s > date('now', 'localtime')", dateExpr)
	case "past":
		return fmt.Sprintf("AND %s <= date('now', 'localtime')", dateExpr)
	}

	match := matchDateWithOperator(value)
	if match == nil {
		return ""
	}

	operator, isoDate := match[1], match[2]
	if operator == "" {
		operator = "=="
	}

	return fmt.Sprintf("AND %s %s date('%s')", dateExpr, operator, isoDate)
}

// makeUnixTimeRangeFilter creates a SQL filter to limit results to the last X days/weeks/years.
func makeUnixTimeRangeFilter(column, offset string) string {
	if offset == "" {
		return ""
	}

	if len(offset) < 2 {
		return ""
	}

	number := offset[:len(offset)-1]
	suffix := offset[len(offset)-1]

	var modifier string
	switch suffix {
	case 'd':
		modifier = fmt.Sprintf("-%s days", number)
	case 'w':
		// Convert weeks to days
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

	columnDatetime := fmt.Sprintf("datetime(%s, 'unixepoch', 'localtime')", column)
	offsetDatetime := fmt.Sprintf("datetime('now', '%s')", modifier)

	return fmt.Sprintf("AND %s > %s", columnDatetime, offsetDatetime)
}

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

// thingsDateExpressionToISODate creates a SQL expression to convert Things date to ISO format.
func thingsDateExpressionToISODate(expr string) string {
	year := fmt.Sprintf("(%s & %d) >> 16", expr, yearMask)
	month := fmt.Sprintf("(%s & %d) >> 12", expr, monthMask)
	day := fmt.Sprintf("(%s & %d) >> 7", expr, dayMask)

	isoDate := fmt.Sprintf("printf('%%d-%%02d-%%02d', %s, %s, %s)", year, month, day)
	return fmt.Sprintf("CASE WHEN %s THEN %s ELSE %s END", expr, isoDate, expr)
}

// thingsTimeExpressionToISOTime creates a SQL expression to convert Things time to HH:MM format.
func thingsTimeExpressionToISOTime(expr string) string {
	hours := fmt.Sprintf("(%s & %d) >> 26", expr, hourMask)
	minutes := fmt.Sprintf("(%s & %d) >> 20", expr, minuteMask)

	isoTime := fmt.Sprintf("printf('%%02d:%%02d', %s, %s)", hours, minutes)
	return fmt.Sprintf("CASE WHEN %s THEN %s ELSE %s END", expr, isoTime, expr)
}

// buildTasksSQL builds the SQL query for fetching tasks.
func buildTasksSQL(wherePredicate, orderPredicate string) string {
	if wherePredicate == "" {
		wherePredicate = sqlTrue
	}
	if orderPredicate == "" {
		orderPredicate = fmt.Sprintf("TASK.%q", IndexDefault)
	}

	startDateExpr := thingsDateExpressionToISODate(fmt.Sprintf("TASK.%s", ColStartDate))
	deadlineExpr := thingsDateExpressionToISODate(fmt.Sprintf("TASK.%s", ColDeadline))
	reminderTimeExpr := thingsTimeExpressionToISOTime(fmt.Sprintf("TASK.%s", ColReminderTime))

	return fmt.Sprintf(`
		SELECT DISTINCT
			TASK.uuid,
			CASE
				WHEN TASK.%s THEN 'to-do'
				WHEN TASK.%s THEN 'project'
				WHEN TASK.%s THEN 'heading'
			END AS type,
			CASE
				WHEN TASK.%s THEN 1
			END AS trashed,
			TASK.title,
			CASE
				WHEN TASK.%s THEN 'incomplete'
				WHEN TASK.%s THEN 'canceled'
				WHEN TASK.%s THEN 'completed'
			END AS status,
			CASE
				WHEN AREA.uuid IS NOT NULL THEN AREA.uuid
			END AS area,
			CASE
				WHEN AREA.uuid IS NOT NULL THEN AREA.title
			END AS area_title,
			CASE
				WHEN PROJECT.uuid IS NOT NULL THEN PROJECT.uuid
			END AS project,
			CASE
				WHEN PROJECT.uuid IS NOT NULL THEN PROJECT.title
			END AS project_title,
			CASE
				WHEN HEADING.uuid IS NOT NULL THEN HEADING.uuid
			END AS heading,
			CASE
				WHEN HEADING.uuid IS NOT NULL THEN HEADING.title
			END AS heading_title,
			TASK.notes,
			CASE
				WHEN TAG.uuid IS NOT NULL THEN 1
			END AS tags,
			CASE
				WHEN TASK.%s THEN 'Inbox'
				WHEN TASK.%s THEN 'Anytime'
				WHEN TASK.%s THEN 'Someday'
			END AS start,
			CASE
				WHEN CHECKLIST_ITEM.uuid IS NOT NULL THEN 1
			END AS checklist,
			%s AS start_date,
			%s AS deadline,
			%s AS reminder_time,
			datetime(TASK.%s, "unixepoch", "localtime") AS stop_date,
			datetime(TASK.%s, "unixepoch", "localtime") AS created,
			datetime(TASK.%s, "unixepoch", "localtime") AS modified,
			TASK.'index',
			TASK.todayIndex AS today_index
		FROM
			%s AS TASK
		LEFT OUTER JOIN
			%s PROJECT ON TASK.project = PROJECT.uuid
		LEFT OUTER JOIN
			%s AREA ON TASK.area = AREA.uuid
		LEFT OUTER JOIN
			%s HEADING ON TASK.heading = HEADING.uuid
		LEFT OUTER JOIN
			%s PROJECT_OF_HEADING ON HEADING.project = PROJECT_OF_HEADING.uuid
		LEFT OUTER JOIN
			%s TAGS ON TASK.uuid = TAGS.tasks
		LEFT OUTER JOIN
			%s TAG ON TAGS.tags = TAG.uuid
		LEFT OUTER JOIN
			%s CHECKLIST_ITEM ON TASK.uuid = CHECKLIST_ITEM.task
		WHERE
			%s
		ORDER BY
			%s
	`,
		FilterIsTodo, FilterIsProject, FilterIsHeading,
		FilterIsTrashed,
		FilterIsIncomplete, FilterIsCanceled, FilterIsCompleted,
		FilterIsInbox, FilterIsAnytime, FilterIsSomeday,
		startDateExpr, deadlineExpr, reminderTimeExpr,
		ColStopDate, ColCreationDate, ColModificationDate,
		TableTask, TableTask, TableArea, TableTask, TableTask,
		TableTaskTag, TableTag, TableChecklistItem,
		wherePredicate, orderPredicate,
	)
}

// buildAreasSQL builds the SQL query for fetching areas.
func buildAreasSQL(wherePredicate string) string {
	if wherePredicate == "" {
		wherePredicate = sqlTrue
	}

	return fmt.Sprintf(`
		SELECT DISTINCT
			AREA.uuid,
			'area' as type,
			AREA.title,
			CASE
				WHEN AREA_TAG.areas IS NOT NULL THEN 1
			END AS tags
		FROM
			%s AS AREA
		LEFT OUTER JOIN
			%s AREA_TAG ON AREA_TAG.areas = AREA.uuid
		LEFT OUTER JOIN
			%s TAG ON TAG.uuid = AREA_TAG.tags
		WHERE
			%s
		ORDER BY AREA."index"
	`, TableArea, TableAreaTag, TableTag, wherePredicate)
}

// buildTagsSQL builds the SQL query for fetching tags.
func buildTagsSQL(wherePredicate string) string {
	if wherePredicate == "" {
		wherePredicate = sqlTrue
	}

	return fmt.Sprintf(`
		SELECT
			uuid, 'tag' AS type, title, shortcut
		FROM
			%s
		WHERE
			%s
		ORDER BY "index"
	`, TableTag, wherePredicate)
}

// buildChecklistItemsSQL builds the SQL query for fetching checklist items.
func buildChecklistItemsSQL() string {
	return fmt.Sprintf(`
		SELECT
			CHECKLIST_ITEM.title,
			CASE
				WHEN CHECKLIST_ITEM.%s THEN 'incomplete'
				WHEN CHECKLIST_ITEM.%s THEN 'canceled'
				WHEN CHECKLIST_ITEM.%s THEN 'completed'
			END AS status,
			date(CHECKLIST_ITEM.stopDate, "unixepoch", "localtime") AS stop_date,
			'checklist-item' as type,
			CHECKLIST_ITEM.uuid,
			datetime(CHECKLIST_ITEM.%s, "unixepoch", "localtime") AS created,
			datetime(CHECKLIST_ITEM.%s, "unixepoch", "localtime") AS modified
		FROM
			%s AS CHECKLIST_ITEM
		WHERE
			CHECKLIST_ITEM.task = ?
		ORDER BY CHECKLIST_ITEM."index"
	`, FilterIsIncomplete, FilterIsCanceled, FilterIsCompleted,
		ColModificationDate, ColModificationDate, TableChecklistItem)
}

// buildTagsOfTaskSQL builds the SQL query for fetching tags of a task.
func buildTagsOfTaskSQL() string {
	return fmt.Sprintf(`
		SELECT
			TAG.title
		FROM
			%s AS TASK_TAG
		LEFT OUTER JOIN
			%s TAG ON TAG.uuid = TASK_TAG.tags
		WHERE
			TASK_TAG.tasks = ?
		ORDER BY TAG."index"
	`, TableTaskTag, TableTag)
}

// buildTagsOfAreaSQL builds the SQL query for fetching tags of an area.
func buildTagsOfAreaSQL() string {
	return fmt.Sprintf(`
		SELECT
			TAG.title
		FROM
			%s AS AREA_TAG
		LEFT OUTER JOIN
			%s TAG ON TAG.uuid = AREA_TAG.tags
		WHERE
			AREA_TAG.areas = ?
		ORDER BY TAG."index"
	`, TableAreaTag, TableTag)
}

// buildCountSQL wraps a SQL query to return only the count.
func buildCountSQL(sql string) string {
	return fmt.Sprintf("SELECT COUNT(uuid) FROM (\n%s\n)", sql)
}

// buildAuthTokenSQL builds the SQL query for fetching the auth token.
func buildAuthTokenSQL() string {
	return fmt.Sprintf(`
		SELECT uriSchemeAuthenticationToken
		FROM %s
		WHERE uuid = '%s'
	`, TableSettings, SettingsUUID)
}
