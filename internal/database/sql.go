package database

import (
	"fmt"
	"time"
)

// sqlTrue is the default WHERE predicate.
const sqlTrue = "TRUE"

// Expressions shared by the task SELECT list and WHERE clause, so a filter
// always tests exactly the value a row reports.
var (
	// A todo under a heading has no project of its own; it belongs to the
	// heading's project.
	sqlProjectUUID  = "COALESCE(PROJECT.uuid, PROJECT_OF_HEADING.uuid)"
	sqlProjectTitle = "COALESCE(PROJECT.title, PROJECT_OF_HEADING.title)"

	// Only loose todos and projects carry an area themselves; everything inside a
	// project inherits the project's area.
	sqlAreaUUID = "COALESCE(TASK.area, PROJECT.area, PROJECT_OF_HEADING.area)"

	// A repeating template has no start date of its own: its next occurrence
	// plays that role. Old databases store 0001-01-01 there instead of NULL.
	sqlStartDate = fmt.Sprintf(
		"CASE WHEN TASK.%s THEN TASK.%s WHEN TASK.%s >= %d THEN TASK.%s END",
		filterIsNotRecurring, colStartDate, colNextInstanceStartDate, minThingsDate, colNextInstanceStartDate,
	)

	// A task outside any project counts as sitting in an active, open, live one.
	sqlProjectStart   = "COALESCE(PROJECT.start, PROJECT_OF_HEADING.start, 1)"
	sqlProjectStatus  = "COALESCE(PROJECT.status, PROJECT_OF_HEADING.status, 0)"
	sqlProjectTrashed = "COALESCE(PROJECT.trashed, PROJECT_OF_HEADING.trashed, 0)"

	// The day Upcoming files a task under: its start date, else its deadline.
	sqlUpcomingDate = fmt.Sprintf("COALESCE((%s), TASK.%s)", sqlStartDate, colDeadline)

	sqlHasChecklist = fmt.Sprintf("EXISTS (SELECT 1 FROM %s ITEM WHERE ITEM.task = TASK.uuid)", tableChecklistItem)
)

// inTodaySQL is the Today list's membership rule: scheduled into Today, or
// scheduled for a day that has arrived (the app moves those into Today the next
// time it runs), or unscheduled with a deadline that is due and has not been
// dismissed from Today.
func inTodaySQL(today time.Time) string {
	day := timeToThingsDate(today.In(time.Local))
	return fmt.Sprintf(
		"((TASK.start = 1 AND TASK.%[1]s IS NOT NULL)"+
			" OR (TASK.start = 2 AND TASK.%[1]s <= %[3]d)"+
			" OR (TASK.%[1]s IS NULL AND TASK.%[2]s <= %[3]d AND TASK.deadlineSuppressionDate IS NULL))",
		colStartDate, colDeadline, day,
	)
}

// inUpcomingSQL is the Upcoming list's membership rule for dated entries: a
// start date in the future (for a repeating template, its next occurrence), or
// no start date at all and a deadline in the future.
func inUpcomingSQL(today time.Time) string {
	day := timeToThingsDate(today.In(time.Local))
	return fmt.Sprintf(
		"((%[1]s) > %[4]d OR ((%[1]s) IS NULL AND TASK.%[2]s AND TASK.%[3]s > %[4]d))",
		sqlStartDate, filterIsNotRecurring, colDeadline, day,
	)
}

// taskTagsSQL is an EXISTS predicate over a task's tags, narrowed by cond (a
// predicate on TAG) when non-empty. A dangling tag reference never satisfies it.
func taskTagsSQL(cond string) string {
	return tagLinkSQL(tableTaskTag, "tasks", "TASK.uuid", cond)
}

// areaTagsSQL is taskTagsSQL for an area's tags.
func areaTagsSQL(cond string) string {
	return tagLinkSQL(tableAreaTag, "areas", "AREA.uuid", cond)
}

func tagLinkSQL(linkTable, ownerColumn, owner, cond string) string {
	sql := fmt.Sprintf(
		"EXISTS (SELECT 1 FROM %s LINK JOIN %s TAG ON TAG.uuid = LINK.tags WHERE LINK.%s = %s",
		linkTable, tableTag, ownerColumn, owner,
	)
	if cond != "" {
		sql += " AND " + cond
	}
	return sql + ")"
}

// buildTasksSQL builds the SQL query for fetching tasks.
func buildTasksSQL(wherePredicate, orderPredicate string, limit *int) string {
	if wherePredicate == "" {
		wherePredicate = sqlTrue
	}

	sql := fmt.Sprintf(`
		SELECT
			TASK.uuid,
			TASK.type,
			TASK.trashed,
			TASK.title,
			TASK.status,
			AREA.uuid AS area,
			AREA.title AS area_title,
			%s AS project,
			%s AS project_title,
			HEADING.uuid AS heading,
			HEADING.title AS heading_title,
			TASK.notes,
			%s AS has_tags,
			TASK.start,
			%s AS has_checklist,
			%s AS start_date,
			TASK.%s AS deadline,
			TASK.%s AS reminder_time,
			TASK.%s AS stop_date,
			TASK.%s AS created,
			COALESCE(TASK.%s, TASK.%s) AS modified,
			TASK."index",
			TASK.todayIndex AS today_index,
			TASK.startBucket AS start_bucket,
			TASK.rt1_repeatingTemplate IS NOT NULL OR TASK.%s AS repeating,
			TASK.%s AS template,
			COALESCE(TASK.rt1_recurrenceRule, SERIES.rt1_recurrenceRule) AS repeat_rule,
			COALESCE(TASK.rt1_instanceCreationPaused, SERIES.rt1_instanceCreationPaused) AS repeat_paused
		FROM
			%s AS TASK
		LEFT OUTER JOIN
			%s PROJECT ON TASK.project = PROJECT.uuid
		LEFT OUTER JOIN
			%s HEADING ON TASK.heading = HEADING.uuid
		LEFT OUTER JOIN
			%s PROJECT_OF_HEADING ON HEADING.project = PROJECT_OF_HEADING.uuid
		LEFT OUTER JOIN
			%s AREA ON AREA.uuid = %s
		LEFT OUTER JOIN
			%s SERIES ON TASK.rt1_repeatingTemplate = SERIES.uuid
		WHERE
			%s
		ORDER BY
			%s
	`,
		sqlProjectUUID, sqlProjectTitle,
		taskTagsSQL(""), sqlHasChecklist,
		sqlStartDate, colDeadline, colReminderTime,
		colStopDate, colCreationDate, colModificationDate, colCreationDate,
		filterIsRecurring, filterIsRecurring,
		tableTask, tableTask, tableTask, tableTask, tableArea, sqlAreaUUID, tableTask,
		wherePredicate, orderPredicate,
	)

	if limit != nil && *limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", *limit)
	}

	return sql
}

// buildAreasSQL builds the SQL query for fetching areas.
func buildAreasSQL(wherePredicate string) string {
	if wherePredicate == "" {
		wherePredicate = sqlTrue
	}

	return fmt.Sprintf(`
		SELECT
			AREA.uuid,
			AREA.title,
			%s AS has_tags
		FROM
			%s AS AREA
		WHERE
			%s
		ORDER BY AREA."index"
	`, areaTagsSQL(""), tableArea, wherePredicate)
}

// buildTagsSQL builds the SQL query for fetching tags.
func buildTagsSQL(wherePredicate string) string {
	if wherePredicate == "" {
		wherePredicate = sqlTrue
	}

	return fmt.Sprintf(`
		SELECT
			uuid, title, shortcut
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
			uuid, title, status, %s, %s, COALESCE(%s, %s)
		FROM
			%s
		WHERE
			task = ?
		ORDER BY "index"
	`, colStopDate, colCreationDate, colModificationDate, colCreationDate, tableChecklistItem)
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
