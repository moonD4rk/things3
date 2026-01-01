package things3

import (
	"context"
	"fmt"
	"time"
)

// TaskQuery provides a fluent interface for building task queries.
type TaskQuery struct {
	db *DB

	// Filters
	uuid               *string
	taskType           *TaskType
	status             *Status
	start              *StartBucket
	areaUUID           *string // specific area UUID
	hasArea            *bool   // true: has area, false: no area
	projectUUID        *string // specific project UUID
	hasProject         *bool   // true: has project, false: no project
	headingUUID        *string // specific heading UUID
	hasHeading         *bool   // true: has heading, false: no heading
	tagTitle           *string // specific tag title
	hasTags            *bool   // true: has tags, false: no tags
	deadlineSuppressed *bool
	trashed            *bool
	contextTrashed     *bool
	createdAfter       *time.Time
	searchQuery        *string
	index              string

	// Date filters (new type-safe approach)
	startDateFilter *dateFilterValue
	stopDateFilter  *dateFilterValue
	deadlineFilter  *dateFilterValue

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

// =============================================================================
// Type-Safe Sub-Builder Entry Points
// =============================================================================

// Type returns a TypeFilter for type-safe task type filtering.
// Example: db.Tasks().Type().Todo().All(ctx)
func (q *TaskQuery) Type() *TypeFilter {
	return &TypeFilter{query: q}
}

// Status returns a StatusFilter for type-safe status filtering.
// Example: db.Tasks().Status().Incomplete().All(ctx)
func (q *TaskQuery) Status() *StatusFilter {
	return &StatusFilter{query: q}
}

// Start returns a StartFilter for type-safe start bucket filtering.
// Example: db.Tasks().Start().Inbox().All(ctx)
func (q *TaskQuery) Start() *StartFilter {
	return &StartFilter{query: q}
}

// StartDate returns a DateFilter for start date filtering.
// Example: db.Tasks().StartDate().Future().All(ctx)
func (q *TaskQuery) StartDate() *DateFilter {
	return &DateFilter{query: q, field: dateFieldStartDate}
}

// StopDate returns a DateFilter for stop date filtering.
// Example: db.Tasks().StopDate().Exists(true).All(ctx)
func (q *TaskQuery) StopDate() *DateFilter {
	return &DateFilter{query: q, field: dateFieldStopDate}
}

// Deadline returns a DateFilter for deadline filtering.
// Example: db.Tasks().Deadline().OnOrBefore("2024-12-31").All(ctx)
func (q *TaskQuery) Deadline() *DateFilter {
	return &DateFilter{query: q, field: dateFieldDeadline}
}

// InArea filters tasks by a specific area UUID.
func (q *TaskQuery) InArea(uuid string) *TaskQuery {
	q.areaUUID = &uuid
	return q
}

// HasArea filters tasks by whether they have an area.
// Pass true to include only tasks with an area.
// Pass false to include only tasks without an area.
func (q *TaskQuery) HasArea(has bool) *TaskQuery {
	q.hasArea = &has
	return q
}

// InProject filters tasks by a specific project UUID.
func (q *TaskQuery) InProject(uuid string) *TaskQuery {
	q.projectUUID = &uuid
	return q
}

// HasProject filters tasks by whether they have a project.
// Pass true to include only tasks with a project.
// Pass false to include only tasks without a project.
func (q *TaskQuery) HasProject(has bool) *TaskQuery {
	q.hasProject = &has
	return q
}

// InHeading filters tasks by a specific heading UUID.
func (q *TaskQuery) InHeading(uuid string) *TaskQuery {
	q.headingUUID = &uuid
	return q
}

// HasHeading filters tasks by whether they have a heading.
// Pass true to include only tasks with a heading.
// Pass false to include only tasks without a heading.
func (q *TaskQuery) HasHeading(has bool) *TaskQuery {
	q.hasHeading = &has
	return q
}

// InTag filters tasks by a specific tag title.
func (q *TaskQuery) InTag(title string) *TaskQuery {
	q.tagTitle = &title
	return q
}

// HasTag filters tasks by whether they have any tags.
// Pass true to include only tasks with tags.
// Pass false to include only tasks without tags.
func (q *TaskQuery) HasTag(has bool) *TaskQuery {
	q.hasTags = &has
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

// CreatedAfter filters tasks created after the specified time.
// Example: db.Tasks().CreatedAfter(things3.DaysAgo(7)).All(ctx)
func (q *TaskQuery) CreatedAfter(t time.Time) *TaskQuery {
	q.createdAfter = &t
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
//
//nolint:gocyclo,funlen // complex but necessary for comprehensive filter handling
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

	// Area filter: specific UUID or has/no area
	if q.areaUUID != nil {
		fb.addEqual("TASK.area", *q.areaUUID)
	} else if q.hasArea != nil {
		fb.addEqual("TASK.area", *q.hasArea)
	}

	// Project filter (also check PROJECT_OF_HEADING for tasks in headings)
	if q.projectUUID != nil {
		fb.addOr(
			equal("TASK.project", *q.projectUUID),
			equal("PROJECT_OF_HEADING.uuid", *q.projectUUID),
		)
	} else if q.hasProject != nil {
		fb.addOr(
			equal("TASK.project", *q.hasProject),
			equal("PROJECT_OF_HEADING.uuid", *q.hasProject),
		)
	}

	// Heading filter: specific UUID or has/no heading
	if q.headingUUID != nil {
		fb.addEqual("TASK.heading", *q.headingUUID)
	} else if q.hasHeading != nil {
		fb.addEqual("TASK.heading", *q.hasHeading)
	}

	// Tag filter: specific title or has/no tags
	if q.tagTitle != nil {
		fb.addEqual("TAG.title", *q.tagTitle)
	} else if q.hasTags != nil {
		fb.addEqual("TAG.title", *q.hasTags)
	}

	// Deadline suppressed filter
	if q.deadlineSuppressed != nil {
		fb.addEqual("TASK.deadlineSuppressionDate", *q.deadlineSuppressed)
	}

	// Date filters using new type-safe structure
	if q.startDateFilter != nil {
		fb.addDateFilterValue(fmt.Sprintf("TASK.%s", colStartDate), q.startDateFilter, true)
	}
	if q.stopDateFilter != nil {
		fb.addDateFilterValue(fmt.Sprintf("TASK.%s", colStopDate), q.stopDateFilter, false)
	}
	if q.deadlineFilter != nil {
		fb.addDateFilterValue(fmt.Sprintf("TASK.%s", colDeadline), q.deadlineFilter, true)
	}

	// CreatedAfter filter
	if q.createdAfter != nil {
		fb.addCreatedAfterFilter(fmt.Sprintf("TASK.%s", colCreationDate), *q.createdAfter)
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
			Type().Todo().
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
