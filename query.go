package things3

import (
	"context"
	"fmt"
	"strings"
)

// TaskQuery provides a fluent interface for building task queries.
type TaskQuery struct {
	client *Client

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
func (c *Client) Tasks() *TaskQuery {
	return &TaskQuery{
		client: c,
		index:  IndexDefault,
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
	q.index = IndexToday
	return q
}

// IncludeItems includes nested items (checklist for to-dos, tasks for projects/headings).
func (q *TaskQuery) IncludeItems(include bool) *TaskQuery {
	q.includeItems = include
	return q
}

// buildWhere builds the WHERE clause for the query.
//
//nolint:gocyclo,funlen // Complexity is inherent to handling many filter conditions
func (q *TaskQuery) buildWhere() string {
	var conditions []string

	// Always exclude recurring tasks
	conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsNotRecurring))

	// Trashed filter (default: not trashed)
	if q.trashed != nil {
		if *q.trashed {
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsTrashed))
		} else {
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsNotTrashed))
		}
	} else {
		conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsNotTrashed))
	}

	// Context trashed filter
	if q.contextTrashed != nil {
		projectFilter := makeTruthyFilter("PROJECT.trashed", q.contextTrashed)
		headingFilter := makeTruthyFilter("PROJECT_OF_HEADING.trashed", q.contextTrashed)
		if projectFilter != "" {
			conditions = append(conditions, strings.TrimPrefix(projectFilter, "AND "))
		}
		if headingFilter != "" {
			conditions = append(conditions, strings.TrimPrefix(headingFilter, "AND "))
		}
	}

	// Type filter
	if q.taskType != nil {
		switch *q.taskType {
		case TaskTypeTodo:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsTodo))
		case TaskTypeProject:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsProject))
		case TaskTypeHeading:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsHeading))
		}
	}

	// Status filter (default: incomplete)
	if q.status != nil {
		switch *q.status {
		case StatusIncomplete:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsIncomplete))
		case StatusCanceled:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsCanceled))
		case StatusCompleted:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsCompleted))
		}
	}

	// Start bucket filter
	if q.start != nil {
		switch *q.start {
		case StartInbox:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsInbox))
		case StartAnytime:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsAnytime))
		case StartSomeday:
			conditions = append(conditions, fmt.Sprintf("TASK.%s", FilterIsSomeday))
		}
	}

	// UUID filter
	if q.uuid != nil {
		conditions = append(conditions, fmt.Sprintf("TASK.uuid = '%s'", escapeString(*q.uuid)))
	}

	// Area filter
	if filter := makeFilter("TASK.area", q.areaUUID); filter != "" {
		conditions = append(conditions, strings.TrimPrefix(filter, "AND "))
	}

	// Project filter (also check PROJECT_OF_HEADING for tasks in headings)
	if q.projectUUID != nil {
		projectFilter := makeFilter("TASK.project", q.projectUUID)
		headingProjectFilter := makeFilter("PROJECT_OF_HEADING.uuid", q.projectUUID)
		orFilter := makeOrFilter(projectFilter, headingProjectFilter)
		if orFilter != "" {
			conditions = append(conditions, strings.TrimPrefix(orFilter, "AND "))
		}
	}

	// Heading filter
	if filter := makeFilter("TASK.heading", q.headingUUID); filter != "" {
		conditions = append(conditions, strings.TrimPrefix(filter, "AND "))
	}

	// Tag filter
	if filter := makeFilter("TAG.title", q.tagTitle); filter != "" {
		conditions = append(conditions, strings.TrimPrefix(filter, "AND "))
	}

	// Deadline suppressed filter
	if q.deadlineSuppressed != nil {
		if *q.deadlineSuppressed {
			conditions = append(conditions, "TASK.deadlineSuppressionDate IS NOT NULL")
		} else {
			conditions = append(conditions, "TASK.deadlineSuppressionDate IS NULL")
		}
	}

	// Date filters
	if filter := makeThingsDateFilter(fmt.Sprintf("TASK.%s", ColStartDate), q.startDate); filter != "" {
		conditions = append(conditions, strings.TrimPrefix(filter, "AND "))
	}
	if filter := makeUnixTimeFilter(fmt.Sprintf("TASK.%s", ColStopDate), q.stopDate); filter != "" {
		conditions = append(conditions, strings.TrimPrefix(filter, "AND "))
	}
	if filter := makeThingsDateFilter(fmt.Sprintf("TASK.%s", ColDeadline), q.deadline); filter != "" {
		conditions = append(conditions, strings.TrimPrefix(filter, "AND "))
	}

	// Last filter
	if q.last != nil {
		if filter := makeUnixTimeRangeFilter(fmt.Sprintf("TASK.%s", ColCreationDate), *q.last); filter != "" {
			conditions = append(conditions, strings.TrimPrefix(filter, "AND "))
		}
	}

	// Search filter
	if q.searchQuery != nil {
		if filter := makeSearchFilter(*q.searchQuery); filter != "" {
			conditions = append(conditions, strings.TrimPrefix(filter, "AND "))
		}
	}

	return strings.Join(conditions, "\n            AND ")
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

	rows, err := q.client.executeQuery(ctx, sql)
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
			tags, err := q.client.getTagsOfTask(ctx, task.UUID)
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
	if err := q.client.executeQueryRow(ctx, countSQL).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// loadTaskItems loads nested items for a task (checklist for to-dos, tasks for projects/headings).
func (q *TaskQuery) loadTaskItems(ctx context.Context, task *Task) error {
	switch task.Type {
	case TaskTypeTodo:
		if task.Checklist != nil {
			items, err := q.client.getChecklistItems(ctx, task.UUID)
			if err != nil {
				return err
			}
			task.Checklist = items
		}
	case TaskTypeProject:
		items, err := q.client.Tasks().
			InProject(task.UUID).
			ContextTrashed(false).
			IncludeItems(true).
			All(ctx)
		if err != nil {
			return err
		}
		task.Items = items
	case TaskTypeHeading:
		items, err := q.client.Tasks().
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
