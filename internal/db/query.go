package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// TaskFilter captures all parameters for a task query.
type TaskFilter struct {
	UUID               *string
	UUIDPrefix         *string
	TaskType           *int
	Status             *int
	Start              *int
	AreaUUID           *string
	HasArea            *bool
	ProjectUUID        *string
	HasProject         *bool
	HeadingUUID        *string
	HasHeading         *bool
	TagTitle           *string
	HasTags            *bool
	DeadlineSuppressed *bool
	Trashed            *bool
	ContextTrashed     *bool
	CreatedAfter       *time.Time
	SearchQuery        *string
	Index              string
	StartDateFilter    *DateFilterValue
	StopDateFilter     *DateFilterValue
	DeadlineFilter     *DateFilterValue
}

// buildWhere builds the WHERE clause for a task query.
func (f *TaskFilter) buildWhere() string {
	var w whereBuilder

	// Always exclude recurring tasks
	w.add("TASK." + filterIsNotRecurring)

	// Trashed filter (default: not trashed)
	if f.Trashed != nil && *f.Trashed {
		w.add("TASK." + filterIsTrashed)
	} else {
		w.add("TASK." + filterIsNotTrashed)
	}

	// Context trashed
	w.addTruthy("PROJECT.trashed", f.ContextTrashed)
	w.addTruthy("PROJECT_OF_HEADING.trashed", f.ContextTrashed)

	// Integer field filters
	w.addIntEqual("TASK.type", f.TaskType)
	w.addIntEqual("TASK.status", f.Status)
	w.addIntEqual("TASK.start", f.Start)

	// UUID filters
	if f.UUID != nil {
		w.addEqual("TASK.uuid", *f.UUID)
	}
	if f.UUIDPrefix != nil {
		w.addLike("TASK.uuid", *f.UUIDPrefix+"%")
	}

	// Relation filters
	w.addEqualOrHas("TASK.area", f.AreaUUID, f.HasArea)
	w.addOrEqualOrHas("TASK.project", "PROJECT_OF_HEADING.uuid", f.ProjectUUID, f.HasProject)
	w.addEqualOrHas("TASK.heading", f.HeadingUUID, f.HasHeading)
	w.addEqualOrHas("TAG.title", f.TagTitle, f.HasTags)

	// Deadline suppressed
	if f.DeadlineSuppressed != nil {
		w.addEqual("TASK.deadlineSuppressionDate", *f.DeadlineSuppressed)
	}

	// Date filters
	w.addDateFilter("TASK."+colStartDate, f.StartDateFilter, true)
	w.addDateFilter("TASK."+colStopDate, f.StopDateFilter, false)
	w.addDateFilter("TASK."+colDeadline, f.DeadlineFilter, true)

	// Time-based filters
	if f.CreatedAfter != nil {
		w.addCreatedAfter("TASK."+colCreationDate, *f.CreatedAfter)
	}
	if f.SearchQuery != nil {
		w.addSearch(*f.SearchQuery)
	}

	return w.sql()
}

// buildOrder builds the ORDER BY clause.
func (f *TaskFilter) buildOrder() string {
	index := f.Index
	if index == "" {
		index = IndexDefault
	}
	return fmt.Sprintf("TASK.%q", index)
}

// AreaFilter captures all parameters for an area query.
type AreaFilter struct {
	UUID     *string
	Title    *string
	Visible  *bool
	TagTitle *string
	HasTag   *bool
}

// buildWhere builds the WHERE clause for an area query.
func (f *AreaFilter) buildWhere() string {
	var w whereBuilder

	if f.UUID != nil {
		w.addEqual("AREA.uuid", *f.UUID)
	}
	if f.Title != nil {
		w.addEqual("AREA.title", *f.Title)
	}
	w.addTruthy("AREA.visible", f.Visible)

	// Tag filter: specific title or has/no tags
	if f.TagTitle != nil {
		w.addEqual("TAG.title", *f.TagTitle)
	} else if f.HasTag != nil {
		w.addEqual("TAG.title", *f.HasTag)
	}

	return w.sql()
}

// TagFilter captures all parameters for a tag query.
type TagFilter struct {
	UUID       *string
	Title      *string
	ParentUUID *string
}

// buildWhere builds the WHERE clause for a tag query.
func (f *TagFilter) buildWhere() string {
	var w whereBuilder

	if f.UUID != nil {
		w.addEqual("uuid", *f.UUID)
	}
	if f.Title != nil {
		w.addEqual("title", *f.Title)
	}
	if f.ParentUUID != nil {
		w.addEqual("parent", *f.ParentUUID)
	}

	return w.sql()
}

// QueryTasks executes a task query and returns matching rows.
func (d *DB) QueryTasks(ctx context.Context, f *TaskFilter) ([]TaskRow, error) {
	where := f.buildWhere()
	order := f.buildOrder()
	query := BuildTasksSQL(where, order)

	rows, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []TaskRow
	for rows.Next() {
		task, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}

	return tasks, rows.Err()
}

// CountTasks returns the count of tasks matching the filter.
func (d *DB) CountTasks(ctx context.Context, f *TaskFilter) (int, error) {
	where := f.buildWhere()
	order := f.buildOrder()
	taskSQL := BuildTasksSQL(where, order)
	countSQL := BuildCountSQL(taskSQL)

	var count int
	if err := d.ExecuteQueryRow(ctx, countSQL).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// QueryAreas executes an area query and returns matching rows.
func (d *DB) QueryAreas(ctx context.Context, f AreaFilter) ([]AreaRow, error) {
	query := BuildAreasSQL(f.buildWhere())
	rows, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []AreaRow
	for rows.Next() {
		area, err := scanAreaRow(rows)
		if err != nil {
			return nil, err
		}
		areas = append(areas, *area)
	}

	return areas, rows.Err()
}

// CountAreas returns the count of areas matching the filter.
func (d *DB) CountAreas(ctx context.Context, f AreaFilter) (int, error) {
	areaSQL := BuildAreasSQL(f.buildWhere())
	countSQL := BuildCountSQL(areaSQL)

	var count int
	if err := d.ExecuteQueryRow(ctx, countSQL).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// QueryTags executes a tag query and returns matching rows.
func (d *DB) QueryTags(ctx context.Context, f TagFilter) ([]TagRow, error) {
	query := BuildTagsSQL(f.buildWhere())
	rows, err := d.ExecuteQuery(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []TagRow
	for rows.Next() {
		tag, err := scanTagRow(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, *tag)
	}

	return tags, rows.Err()
}

// TagsOfTask returns the tag titles for a task.
func (d *DB) TagsOfTask(ctx context.Context, taskUUID string) ([]string, error) {
	query := BuildTagsOfTaskSQL()
	rows, err := d.ExecuteQuery(ctx, query, taskUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, err
		}
		tags = append(tags, title)
	}

	return tags, rows.Err()
}

// TagsOfArea returns the tag titles for an area.
func (d *DB) TagsOfArea(ctx context.Context, areaUUID string) ([]string, error) {
	query := BuildTagsOfAreaSQL()
	rows, err := d.ExecuteQuery(ctx, query, areaUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, err
		}
		tags = append(tags, title)
	}

	return tags, rows.Err()
}

// QueryChecklistItems returns checklist items for a task.
func (d *DB) QueryChecklistItems(ctx context.Context, taskUUID string) ([]ChecklistItemRow, error) {
	query := BuildChecklistItemsSQL()
	rows, err := d.ExecuteQuery(ctx, query, taskUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ChecklistItemRow
	for rows.Next() {
		item, err := scanChecklistItemRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	return items, rows.Err()
}

// AuthToken returns the Things URL scheme authentication token.
func (d *DB) AuthToken(ctx context.Context) (string, error) {
	query := BuildAuthTokenSQL()
	var token sql.NullString
	if err := d.ExecuteQueryRow(ctx, query).Scan(&token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrAuthTokenNotFound
		}
		return "", err
	}
	if !token.Valid {
		return "", ErrAuthTokenNotFound
	}

	return token.String, nil
}
