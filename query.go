package things3

import (
	"context"
	"time"

	idb "github.com/moond4rk/things3/internal/db"
)

// taskQuery provides a fluent interface for building task queries.
type taskQuery struct {
	database     *db
	filter       idb.TaskFilter
	includeItems bool
}

// Tasks creates a new taskQuery for querying tasks.
func (d *db) Tasks() *taskQuery {
	return &taskQuery{
		database: d,
		filter:   idb.TaskFilter{Index: idb.IndexDefault},
	}
}

// WithUUID filters tasks by UUID (exact match).
func (q *taskQuery) WithUUID(uuid string) TaskQueryBuilder {
	q.filter.UUID = &uuid
	return q
}

// WithUUIDPrefix filters tasks by UUID prefix (partial match).
func (q *taskQuery) WithUUIDPrefix(prefix string) TaskQueryBuilder {
	q.filter.UUIDPrefix = &prefix
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
	q.filter.AreaUUID = &uuid
	return q
}

// HasArea filters tasks by whether they have an area.
func (q *taskQuery) HasArea(has bool) TaskQueryBuilder {
	q.filter.HasArea = &has
	return q
}

// InProject filters tasks by a specific project UUID.
func (q *taskQuery) InProject(uuid string) TaskQueryBuilder {
	q.filter.ProjectUUID = &uuid
	return q
}

// HasProject filters tasks by whether they have a project.
func (q *taskQuery) HasProject(has bool) TaskQueryBuilder {
	q.filter.HasProject = &has
	return q
}

// InHeading filters tasks by a specific heading UUID.
func (q *taskQuery) InHeading(uuid string) TaskQueryBuilder {
	q.filter.HeadingUUID = &uuid
	return q
}

// HasHeading filters tasks by whether they have a heading.
func (q *taskQuery) HasHeading(has bool) TaskQueryBuilder {
	q.filter.HasHeading = &has
	return q
}

// InTag filters tasks by a specific tag title.
func (q *taskQuery) InTag(title string) TaskQueryBuilder {
	q.filter.TagTitle = &title
	return q
}

// HasTag filters tasks by whether they have any tags.
func (q *taskQuery) HasTag(has bool) TaskQueryBuilder {
	q.filter.HasTags = &has
	return q
}

// WithDeadlineSuppressed filters tasks by deadline suppression status.
func (q *taskQuery) WithDeadlineSuppressed(suppressed bool) TaskQueryBuilder {
	q.filter.DeadlineSuppressed = &suppressed
	return q
}

// Trashed filters tasks by trash status.
func (q *taskQuery) Trashed(trashed bool) TaskQueryBuilder {
	q.filter.Trashed = &trashed
	return q
}

// ContextTrashed filters tasks by the trash status of their context (project/heading).
func (q *taskQuery) ContextTrashed(trashed bool) TaskQueryBuilder {
	q.filter.ContextTrashed = &trashed
	return q
}

// CreatedAfter filters tasks created after the specified time.
func (q *taskQuery) CreatedAfter(t time.Time) TaskQueryBuilder {
	q.filter.CreatedAfter = &t
	return q
}

// Search filters tasks by a search query.
func (q *taskQuery) Search(query string) TaskQueryBuilder {
	q.filter.SearchQuery = &query
	return q
}

// OrderByTodayIndex orders results by today index instead of default index.
func (q *taskQuery) OrderByTodayIndex() TaskQueryBuilder {
	q.filter.Index = idb.IndexToday
	return q
}

// IncludeItems includes nested items (checklist for to-dos, tasks for projects/headings).
func (q *taskQuery) IncludeItems(include bool) TaskQueryBuilder {
	q.includeItems = include
	return q
}

// All executes the query and returns all matching tasks.
func (q *taskQuery) All(ctx context.Context) ([]Task, error) {
	rows, err := q.database.inner.QueryTasks(ctx, &q.filter)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for i := range rows {
		task := convertTaskRow(&rows[i])

		// Load tags if present
		if rows[i].HasTags {
			tags, err := q.database.inner.TagsOfTask(ctx, rows[i].UUID)
			if err != nil {
				return nil, err
			}
			task.Tags = tags
		}

		// Load nested items if requested
		if q.includeItems {
			if err := q.loadTaskItems(ctx, &task); err != nil {
				return nil, err
			}
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
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
	return q.database.inner.CountTasks(ctx, &q.filter)
}

// loadTaskItems loads nested items for a task (checklist for to-dos, tasks for projects/headings).
func (q *taskQuery) loadTaskItems(ctx context.Context, task *Task) error {
	switch task.Type {
	case TaskTypeTodo:
		if task.Checklist != nil {
			rows, err := q.database.inner.QueryChecklistItems(ctx, task.UUID)
			if err != nil {
				return err
			}
			task.Checklist = convertChecklistItemRows(rows)
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
