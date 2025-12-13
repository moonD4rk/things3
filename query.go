package things3

import (
	"context"
	"fmt"
)

// TaskQuery provides a fluent interface for building task queries.
type TaskQuery struct {
	db *DB

	// Filters
	uuid               *string
	taskType           *TaskType
	status             *Status
	start              *StartBucket
	areaUUID           any // string, bool, or nil
	projectUUID        any // string, bool, or nil
	headingUUID        any // string, bool, or nil
	tagTitle           any // string, bool, or nil
	startDate          any // string, bool, or nil
	stopDate           any // string, bool, or nil
	deadline           any // string, bool, or nil
	deadlineSuppressed *bool
	trashed            *bool
	contextTrashed     *bool
	last               *string
	searchQuery        *string
	index              string

	// Options
	includeItems bool
}

// Tasks creates a new TaskQuery for querying tasks.
func (d *DB) Tasks() *TaskQuery {
	return &TaskQuery{
		db:    d,
		index: indexDefault,
	}
}

// WithUUID filters tasks by UUID.
func (q *TaskQuery) WithUUID(uuid string) *TaskQuery {
	q.uuid = &uuid
	return q
}

// WithType filters tasks by type (to-do, project, or heading).
func (q *TaskQuery) WithType(t TaskType) *TaskQuery {
	q.taskType = &t
	return q
}

// WithStatus filters tasks by status.
func (q *TaskQuery) WithStatus(s Status) *TaskQuery {
	q.status = &s
	return q
}

// WithStart filters tasks by start bucket (Inbox, Anytime, Someday).
func (q *TaskQuery) WithStart(s StartBucket) *TaskQuery {
	q.start = &s
	return q
}

// InArea filters tasks by area.
// Pass a UUID string to filter by specific area.
// Pass true to include only tasks with an area.
// Pass false to include only tasks without an area.
func (q *TaskQuery) InArea(area any) *TaskQuery {
	q.areaUUID = area
	return q
}

// InProject filters tasks by project.
// Pass a UUID string to filter by specific project.
// Pass true to include only tasks with a project.
// Pass false to include only tasks without a project.
func (q *TaskQuery) InProject(project any) *TaskQuery {
	q.projectUUID = project
	return q
}

// InHeading filters tasks by heading.
// Pass a UUID string to filter by specific heading.
// Pass true to include only tasks with a heading.
// Pass false to include only tasks without a heading.
func (q *TaskQuery) InHeading(heading any) *TaskQuery {
	q.headingUUID = heading
	return q
}

// WithTag filters tasks by tag.
// Pass a tag title to filter by specific tag.
// Pass true to include only tasks with tags.
// Pass false to include only tasks without tags.
func (q *TaskQuery) WithTag(tag any) *TaskQuery {
	q.tagTitle = tag
	return q
}

// WithStartDate filters tasks by start date.
// Accepts: bool (has/doesn't have), "future", "past", or ISO date with optional operator.
func (q *TaskQuery) WithStartDate(date any) *TaskQuery {
	q.startDate = date
	return q
}

// WithStopDate filters tasks by stop date (completion/cancellation date).
// Accepts: bool (has/doesn't have), "future", "past", or ISO date with optional operator.
func (q *TaskQuery) WithStopDate(date any) *TaskQuery {
	q.stopDate = date
	return q
}

// WithDeadline filters tasks by deadline.
// Accepts: bool (has/doesn't have), "future", "past", or ISO date with optional operator.
func (q *TaskQuery) WithDeadline(deadline any) *TaskQuery {
	q.deadline = deadline
	return q
}

// WithDeadlineSuppressed filters tasks by deadline suppression status.
func (q *TaskQuery) WithDeadlineSuppressed(suppressed bool) *TaskQuery {
	q.deadlineSuppressed = &suppressed
	return q
}

// Trashed filters tasks by trash status.
// Pass true to include only trashed tasks.
// Pass false to include only non-trashed tasks.
func (q *TaskQuery) Trashed(trashed bool) *TaskQuery {
	q.trashed = &trashed
	return q
}

// ContextTrashed filters tasks by the trash status of their context (project/heading).
func (q *TaskQuery) ContextTrashed(trashed bool) *TaskQuery {
	q.contextTrashed = &trashed
	return q
}

// Last filters tasks created within the last X days/weeks/years.
// Format: "3d" (3 days), "2w" (2 weeks), "1y" (1 year).
func (q *TaskQuery) Last(offset string) *TaskQuery {
	q.last = &offset
	return q
}

// Search filters tasks by a search query.
// Searches in task title, notes, and area title.
func (q *TaskQuery) Search(query string) *TaskQuery {
	q.searchQuery = &query
	return q
}

// OrderByTodayIndex orders results by today index instead of default index.
func (q *TaskQuery) OrderByTodayIndex() *TaskQuery {
	q.index = indexToday
	return q
}

// IncludeItems includes nested items (checklist for to-dos, tasks for projects/headings).
func (q *TaskQuery) IncludeItems(include bool) *TaskQuery {
	q.includeItems = include
	return q
}

// buildWhere builds the WHERE clause for the query using filterBuilder.

//nolint:gocyclo
func (q *TaskQuery) buildWhere() string {
	fb := newFilterBuilder()

	// Always exclude recurring tasks
	fb.addStatic(fmt.Sprintf("TASK.%s", filterIsNotRecurring))

	// Trashed filter (default: not trashed)
	if q.trashed != nil && *q.trashed {
		fb.addStatic(fmt.Sprintf("TASK.%s", filterIsTrashed))
	} else {
		fb.addStatic(fmt.Sprintf("TASK.%s", filterIsNotTrashed))
	}

	// Context trashed filter
	fb.addTruthy("PROJECT.trashed", q.contextTrashed)
	fb.addTruthy("PROJECT_OF_HEADING.trashed", q.contextTrashed)

	// Type filter
	if q.taskType != nil {
		switch *q.taskType {
		case TaskTypeTodo:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsTodo))
		case TaskTypeProject:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsProject))
		case TaskTypeHeading:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsHeading))
		}
	}

	// Status filter
	if q.status != nil {
		switch *q.status {
		case StatusIncomplete:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsIncomplete))
		case StatusCanceled:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsCanceled))
		case StatusCompleted:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsCompleted))
		}
	}

	// Start bucket filter
	if q.start != nil {
		switch *q.start {
		case StartInbox:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsInbox))
		case StartAnytime:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsAnytime))
		case StartSomeday:
			fb.addStatic(fmt.Sprintf("TASK.%s", filterIsSomeday))
		}
	}

	// UUID filter
	if q.uuid != nil {
		fb.addEqual("TASK.uuid", *q.uuid)
	}

	// Area filter
	fb.addEqual("TASK.area", q.areaUUID)

	// Project filter (also check PROJECT_OF_HEADING for tasks in headings)
	if q.projectUUID != nil {
		fb.addOr(
			equal("TASK.project", q.projectUUID),
			equal("PROJECT_OF_HEADING.uuid", q.projectUUID),
		)
	}

	// Heading filter
	fb.addEqual("TASK.heading", q.headingUUID)

	// Tag filter
	fb.addEqual("TAG.title", q.tagTitle)

	// Deadline suppressed filter
	if q.deadlineSuppressed != nil {
		fb.addEqual("TASK.deadlineSuppressionDate", *q.deadlineSuppressed)
	}

	// Date filters
	fb.addParsedThingsDateValue(fmt.Sprintf("TASK.%s", colStartDate), q.startDate)
	fb.addParsedUnixTimeValue(fmt.Sprintf("TASK.%s", colStopDate), q.stopDate)
	fb.addParsedThingsDateValue(fmt.Sprintf("TASK.%s", colDeadline), q.deadline)

	// Last filter
	if q.last != nil {
		fb.addUnixTimeRangeValue(fmt.Sprintf("TASK.%s", colCreationDate), *q.last)
	}

	// Search filter
	if q.searchQuery != nil {
		fb.addSearch(*q.searchQuery)
	}

	return fb.sql()
}

// buildOrder builds the ORDER BY clause.
func (q *TaskQuery) buildOrder() string {
	return fmt.Sprintf("TASK.%q", q.index)
}

// All executes the query and returns all matching tasks.
func (q *TaskQuery) All(ctx context.Context) ([]Task, error) {
	where := q.buildWhere()
	order := q.buildOrder()
	sql := buildTasksSQL(where, order)

	rows, err := q.db.executeQuery(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}

		// Load tags if present
		if task.Tags != nil {
			tags, err := q.db.getTagsOfTask(ctx, task.UUID)
			if err != nil {
				return nil, err
			}
			task.Tags = tags
		}

		// Load nested items if requested
		if q.includeItems {
			if err := q.loadTaskItems(ctx, task); err != nil {
				return nil, err
			}
		}

		tasks = append(tasks, *task)
	}

	return tasks, rows.Err()
}

// First executes the query and returns the first matching task.
func (q *TaskQuery) First(ctx context.Context) (*Task, error) {
	// For single task fetch, always include items
	q.includeItems = true

	tasks, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, ErrTaskNotFound
	}
	return &tasks[0], nil
}

// Count executes the query and returns the count of matching tasks.
func (q *TaskQuery) Count(ctx context.Context) (int, error) {
	where := q.buildWhere()
	order := q.buildOrder()
	taskSQL := buildTasksSQL(where, order)
	countSQL := buildCountSQL(taskSQL)

	var count int
	if err := q.db.executeQueryRow(ctx, countSQL).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// loadTaskItems loads nested items for a task (checklist for to-dos, tasks for projects/headings).
func (q *TaskQuery) loadTaskItems(ctx context.Context, task *Task) error {
	switch task.Type {
	case TaskTypeTodo:
		if task.Checklist != nil {
			items, err := q.db.getChecklistItems(ctx, task.UUID)
			if err != nil {
				return err
			}
			task.Checklist = items
		}
	case TaskTypeProject:
		items, err := q.db.Tasks().
			InProject(task.UUID).
			ContextTrashed(false).
			IncludeItems(true).
			All(ctx)
		if err != nil {
			return err
		}
		task.Items = items
	case TaskTypeHeading:
		items, err := q.db.Tasks().
			WithType(TaskTypeTodo).
			InHeading(task.UUID).
			ContextTrashed(false).
			IncludeItems(true).
			All(ctx)
		if err != nil {
			return err
		}
		task.Items = items
	}
	return nil
}
