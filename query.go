package things3

import (
	"context"
	"fmt"
	"time"
)

// taskQuery provides a fluent interface for building task queries.
type taskQuery struct {
	database *db

	// Filters
	uuid               *string
	uuidPrefix         *string
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

// Tasks creates a new taskQuery for querying tasks.
func (d *db) Tasks() *taskQuery {
	return &taskQuery{
		database: d,
		index:    indexDefault,
	}
}

// WithUUID filters tasks by UUID (exact match).
func (q *taskQuery) WithUUID(uuid string) TaskQueryBuilder {
	q.uuid = &uuid
	return q
}

// WithUUIDPrefix filters tasks by UUID prefix (partial match).
func (q *taskQuery) WithUUIDPrefix(prefix string) TaskQueryBuilder {
	q.uuidPrefix = &prefix
	return q
}

// =============================================================================
// Type-Safe Sub-Builder Entry Points
// =============================================================================

// Type returns a typeFilter for type-safe task type filtering.
func (q *taskQuery) Type() TypeFilterBuilder {
	return &typeFilter{query: q}
}

// Status returns a statusFilter for type-safe status filtering.
func (q *taskQuery) Status() StatusFilterBuilder {
	return &statusFilter{query: q}
}

// Start returns a startFilter for type-safe start bucket filtering.
func (q *taskQuery) Start() StartFilterBuilder {
	return &startFilter{query: q}
}

// StartDate returns a dateFilter for start date filtering.
func (q *taskQuery) StartDate() DateFilterBuilder {
	return &dateFilter{query: q, field: dateFieldStartDate}
}

// StopDate returns a dateFilter for stop date filtering.
func (q *taskQuery) StopDate() DateFilterBuilder {
	return &dateFilter{query: q, field: dateFieldStopDate}
}

// Deadline returns a dateFilter for deadline filtering.
func (q *taskQuery) Deadline() DateFilterBuilder {
	return &dateFilter{query: q, field: dateFieldDeadline}
}

// InArea filters tasks by a specific area UUID.
func (q *taskQuery) InArea(uuid string) TaskQueryBuilder {
	q.areaUUID = &uuid
	return q
}

// HasArea filters tasks by whether they have an area.
func (q *taskQuery) HasArea(has bool) TaskQueryBuilder {
	q.hasArea = &has
	return q
}

// InProject filters tasks by a specific project UUID.
func (q *taskQuery) InProject(uuid string) TaskQueryBuilder {
	q.projectUUID = &uuid
	return q
}

// HasProject filters tasks by whether they have a project.
func (q *taskQuery) HasProject(has bool) TaskQueryBuilder {
	q.hasProject = &has
	return q
}

// InHeading filters tasks by a specific heading UUID.
func (q *taskQuery) InHeading(uuid string) TaskQueryBuilder {
	q.headingUUID = &uuid
	return q
}

// HasHeading filters tasks by whether they have a heading.
func (q *taskQuery) HasHeading(has bool) TaskQueryBuilder {
	q.hasHeading = &has
	return q
}

// InTag filters tasks by a specific tag title.
func (q *taskQuery) InTag(title string) TaskQueryBuilder {
	q.tagTitle = &title
	return q
}

// HasTag filters tasks by whether they have any tags.
func (q *taskQuery) HasTag(has bool) TaskQueryBuilder {
	q.hasTags = &has
	return q
}

// WithDeadlineSuppressed filters tasks by deadline suppression status.
func (q *taskQuery) WithDeadlineSuppressed(suppressed bool) TaskQueryBuilder {
	q.deadlineSuppressed = &suppressed
	return q
}

// Trashed filters tasks by trash status.
func (q *taskQuery) Trashed(trashed bool) TaskQueryBuilder {
	q.trashed = &trashed
	return q
}

// ContextTrashed filters tasks by the trash status of their context (project/heading).
func (q *taskQuery) ContextTrashed(trashed bool) TaskQueryBuilder {
	q.contextTrashed = &trashed
	return q
}

// CreatedAfter filters tasks created after the specified time.
func (q *taskQuery) CreatedAfter(t time.Time) TaskQueryBuilder {
	q.createdAfter = &t
	return q
}

// Search filters tasks by a search query.
func (q *taskQuery) Search(query string) TaskQueryBuilder {
	q.searchQuery = &query
	return q
}

// OrderByTodayIndex orders results by today index instead of default index.
func (q *taskQuery) OrderByTodayIndex() TaskQueryBuilder {
	q.index = indexToday
	return q
}

// IncludeItems includes nested items (checklist for to-dos, tasks for projects/headings).
func (q *taskQuery) IncludeItems(include bool) TaskQueryBuilder {
	q.includeItems = include
	return q
}

// buildWhere builds the WHERE clause for the query using filterBuilder.
//
//nolint:gocyclo,funlen // complex but necessary for comprehensive filter handling
func (q *taskQuery) buildWhere() string {
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
	if q.uuidPrefix != nil {
		fb.addLike("TASK.uuid", *q.uuidPrefix+"%")
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
func (q *taskQuery) buildOrder() string {
	return fmt.Sprintf("TASK.%q", q.index)
}

// All executes the query and returns all matching tasks.
func (q *taskQuery) All(ctx context.Context) ([]Task, error) {
	where := q.buildWhere()
	order := q.buildOrder()
	sql := buildTasksSQL(where, order)

	rows, err := q.database.executeQuery(ctx, sql)
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
			tags, err := q.database.getTagsOfTask(ctx, task.UUID)
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
func (q *taskQuery) First(ctx context.Context) (*Task, error) {
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
func (q *taskQuery) Count(ctx context.Context) (int, error) {
	where := q.buildWhere()
	order := q.buildOrder()
	taskSQL := buildTasksSQL(where, order)
	countSQL := buildCountSQL(taskSQL)

	var count int
	if err := q.database.executeQueryRow(ctx, countSQL).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// loadTaskItems loads nested items for a task (checklist for to-dos, tasks for projects/headings).
func (q *taskQuery) loadTaskItems(ctx context.Context, task *Task) error {
	switch task.Type {
	case TaskTypeTodo:
		if task.Checklist != nil {
			items, err := q.database.getChecklistItems(ctx, task.UUID)
			if err != nil {
				return err
			}
			task.Checklist = items
		}
	case TaskTypeProject:
		items, err := q.database.Tasks().
			InProject(task.UUID).
			ContextTrashed(false).
			IncludeItems(true).
			All(ctx)
		if err != nil {
			return err
		}
		task.Items = items
	case TaskTypeHeading:
		items, err := q.database.Tasks().
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
