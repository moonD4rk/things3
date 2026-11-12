package database

import "fmt"

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

	sqlHasChecklist = fmt.Sprintf("EXISTS (SELECT 1 FROM %s ITEM WHERE ITEM.task = TASK.uuid)", tableChecklistItem)
)

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
	if orderPredicate == "" {
		orderPredicate = fmt.Sprintf("TASK.%q", IndexDefault)
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
			TASK.%s AS template
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
		tableTask, tableTask, tableTask, tableTask, tableArea, sqlAreaUUID,
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
