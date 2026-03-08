package db

import "fmt"

// sqlTrue is the default WHERE predicate.
const sqlTrue = "TRUE"

// BuildTasksSQL builds the SQL query for fetching tasks.
func BuildTasksSQL(wherePredicate, orderPredicate string) string {
	if wherePredicate == "" {
		wherePredicate = sqlTrue
	}
	if orderPredicate == "" {
		orderPredicate = fmt.Sprintf("TASK.%q", IndexDefault)
	}

	startDateExpr := thingsDateExpressionToISODate("TASK." + colStartDate)
	deadlineExpr := thingsDateExpressionToISODate("TASK." + colDeadline)
	reminderTimeExpr := thingsTimeExpressionToISOTime("TASK." + colReminderTime)

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

// BuildAreasSQL builds the SQL query for fetching areas.
func BuildAreasSQL(wherePredicate string) string {
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

// BuildTagsSQL builds the SQL query for fetching tags.
func BuildTagsSQL(wherePredicate string) string {
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

// BuildChecklistItemsSQL builds the SQL query for fetching checklist items.
func BuildChecklistItemsSQL() string {
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
		colCreationDate, colModificationDate, tableChecklistItem)
}

// BuildTagsOfTaskSQL builds the SQL query for fetching tags of a task.
func BuildTagsOfTaskSQL() string {
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

// BuildTagsOfAreaSQL builds the SQL query for fetching tags of an area.
func BuildTagsOfAreaSQL() string {
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

// BuildCountSQL wraps a SQL query to return only the count.
func BuildCountSQL(sql string) string {
	return fmt.Sprintf("SELECT COUNT(uuid) FROM (\n%s\n)", sql)
}

// BuildAuthTokenSQL builds the SQL query for fetching the auth token.
func BuildAuthTokenSQL() string {
	return fmt.Sprintf(`
		SELECT uriSchemeAuthenticationToken
		FROM %s
		WHERE uuid = '%s'
	`, tableSettings, settingsUUID)
}
