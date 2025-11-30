package things3

import (
	"fmt"
	"strings"
)

// SQL constant for default WHERE predicate.
const sqlTrue = "TRUE"

// escapeString escapes a string for safe use in SQL queries.
// In SQLite, single quotes within strings are escaped by doubling them.
func escapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
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
		orderPredicate = fmt.Sprintf("TASK.%q", indexDefault)
	}

	startDateExpr := thingsDateExpressionToISODate(fmt.Sprintf("TASK.%s", colStartDate))
	deadlineExpr := thingsDateExpressionToISODate(fmt.Sprintf("TASK.%s", colDeadline))
	reminderTimeExpr := thingsTimeExpressionToISOTime(fmt.Sprintf("TASK.%s", colReminderTime))

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
		filterIsTodo, filterIsProject, filterIsHeading,
		filterIsTrashed,
		filterIsIncomplete, filterIsCanceled, filterIsCompleted,
		filterIsInbox, filterIsAnytime, filterIsSomeday,
		startDateExpr, deadlineExpr, reminderTimeExpr,
		colStopDate, colCreationDate, colModificationDate,
		tableTask, tableTask, tableArea, tableTask, tableTask,
		tableTaskTag, tableTag, tableChecklistItem,
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
	`, tableArea, tableAreaTag, tableTag, wherePredicate)
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
	`, tableTag, wherePredicate)
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
	`, filterIsIncomplete, filterIsCanceled, filterIsCompleted,
		colModificationDate, colModificationDate, tableChecklistItem)
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
	`, tableTaskTag, tableTag)
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
	`, tableAreaTag, tableTag)
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
	`, tableSettings, settingsUUID)
}
