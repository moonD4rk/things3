package things3

import (
	"context"
	"errors"
	"sort"
)

// Todos returns all incomplete to-do items.
func (c *Client) Todos(ctx context.Context) ([]Task, error) {
	return c.Tasks().
		WithType(TaskTypeTodo).
		WithStatus(StatusIncomplete).
		All(ctx)
}

// Projects returns all incomplete projects.
func (c *Client) Projects(ctx context.Context) ([]Task, error) {
	return c.Tasks().
		WithType(TaskTypeProject).
		WithStatus(StatusIncomplete).
		All(ctx)
}

// Inbox returns all tasks in the Inbox.
func (c *Client) Inbox(ctx context.Context) ([]Task, error) {
	return c.Tasks().
		WithStart(StartInbox).
		WithStatus(StatusIncomplete).
		All(ctx)
}

// Today returns tasks that would appear in Today view.
// This includes:
// - Tasks with a start date set to today or earlier and in Anytime
// - Scheduled tasks from Someday with past start dates (yellow dot tasks)
// - Overdue tasks with deadlines that haven't been suppressed
func (c *Client) Today(ctx context.Context) ([]Task, error) {
	// Regular Today tasks
	regularTasks, err := c.Tasks().
		WithStartDate(true).
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		OrderByTodayIndex().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Unconfirmed scheduled tasks (yellow dot)
	scheduledTasks, err := c.Tasks().
		WithStartDate("past").
		WithStart(StartSomeday).
		WithStatus(StatusIncomplete).
		OrderByTodayIndex().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Unconfirmed overdue tasks
	overdueTasks, err := c.Tasks().
		WithStartDate(false).
		WithDeadline("past").
		WithDeadlineSuppressed(false).
		WithStatus(StatusIncomplete).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Combine all tasks
	result := make([]Task, 0, len(regularTasks)+len(scheduledTasks)+len(overdueTasks))
	result = append(result, regularTasks...)
	result = append(result, scheduledTasks...)
	result = append(result, overdueTasks...)

	// Sort by today_index and start_date
	sort.Slice(result, func(i, j int) bool {
		if result[i].TodayIndex != result[j].TodayIndex {
			return result[i].TodayIndex < result[j].TodayIndex
		}
		// If today_index is the same, sort by start_date
		if result[i].StartDate == nil && result[j].StartDate == nil {
			return false
		}
		if result[i].StartDate == nil {
			return false
		}
		if result[j].StartDate == nil {
			return true
		}
		return *result[i].StartDate < *result[j].StartDate
	})

	return result, nil
}

// Upcoming returns tasks scheduled for future dates.
func (c *Client) Upcoming(ctx context.Context) ([]Task, error) {
	return c.Tasks().
		WithStartDate("future").
		WithStart(StartSomeday).
		WithStatus(StatusIncomplete).
		All(ctx)
}

// Anytime returns tasks in the Anytime list.
func (c *Client) Anytime(ctx context.Context) ([]Task, error) {
	return c.Tasks().
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		All(ctx)
}

// Someday returns tasks in the Someday list (without a start date).
func (c *Client) Someday(ctx context.Context) ([]Task, error) {
	return c.Tasks().
		WithStartDate(false).
		WithStart(StartSomeday).
		WithStatus(StatusIncomplete).
		All(ctx)
}

// Logbook returns completed and canceled tasks, sorted by stop date.
func (c *Client) Logbook(ctx context.Context) ([]Task, error) {
	completed, err := c.Completed(ctx)
	if err != nil {
		return nil, err
	}

	canceled, err := c.Canceled(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Task, 0, len(completed)+len(canceled))
	result = append(result, completed...)
	result = append(result, canceled...)

	// Sort by stop_date (newest first)
	sort.Slice(result, func(i, j int) bool {
		if result[i].StopDate == nil && result[j].StopDate == nil {
			return false
		}
		if result[i].StopDate == nil {
			return false
		}
		if result[j].StopDate == nil {
			return true
		}
		return result[i].StopDate.After(*result[j].StopDate)
	})

	return result, nil
}

// Trash returns trashed tasks.
func (c *Client) Trash(ctx context.Context) ([]Task, error) {
	q := c.Tasks().Trashed(true)
	// Remove default status filter for trash
	q.status = nil
	return q.All(ctx)
}

// Completed returns completed tasks.
func (c *Client) Completed(ctx context.Context) ([]Task, error) {
	return c.Tasks().
		WithStatus(StatusCompleted).
		All(ctx)
}

// Canceled returns canceled tasks.
func (c *Client) Canceled(ctx context.Context) ([]Task, error) {
	return c.Tasks().
		WithStatus(StatusCanceled).
		All(ctx)
}

// Deadlines returns tasks with deadlines, sorted by deadline.
func (c *Client) Deadlines(ctx context.Context) ([]Task, error) {
	tasks, err := c.Tasks().
		WithDeadline(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Sort by deadline
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Deadline == nil && tasks[j].Deadline == nil {
			return false
		}
		if tasks[i].Deadline == nil {
			return false
		}
		if tasks[j].Deadline == nil {
			return true
		}
		return *tasks[i].Deadline < *tasks[j].Deadline
	})

	return tasks, nil
}

// Last returns tasks created within the last X days/weeks/years.
// Format: "3d" (3 days), "2w" (2 weeks), "1y" (1 year).
func (c *Client) Last(ctx context.Context, offset string) ([]Task, error) {
	if offset == "" {
		return nil, ErrInvalidParameter
	}

	tasks, err := c.Tasks().
		Last(offset).
		WithStatus(StatusIncomplete).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Sort by created date (newest first)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Created.After(tasks[j].Created)
	})

	return tasks, nil
}

// Search searches for tasks matching the query.
// Searches in task title, notes, and area title.
func (c *Client) Search(ctx context.Context, query string) ([]Task, error) {
	return c.Tasks().
		Search(query).
		WithStatus(StatusIncomplete).
		All(ctx)
}

// Get retrieves an object by UUID.
// Returns a Task, Area, or Tag depending on what is found.
// Returns nil if not found.
func (c *Client) Get(ctx context.Context, uuid string) (any, error) {
	// Try to find as task
	task, err := c.Tasks().WithUUID(uuid).First(ctx)
	if err == nil {
		return task, nil
	}
	if !errors.Is(err, ErrTaskNotFound) {
		return nil, err
	}

	// Try to find as area
	area, err := c.Areas().WithUUID(uuid).First(ctx)
	if err == nil {
		return area, nil
	}
	if !errors.Is(err, ErrAreaNotFound) {
		return nil, err
	}

	// Try to find as tag
	tags, err := c.Tags().All(ctx)
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		if tag.UUID == uuid {
			return &tag, nil
		}
	}

	return nil, nil
}

// ChecklistItems returns the checklist items for a to-do.
func (c *Client) ChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error) {
	return c.getChecklistItems(ctx, todoUUID)
}

// AreaQuery provides a fluent interface for building area queries.
type AreaQuery struct {
	client *Client

	uuid         *string
	tagTitle     any // string, bool, or nil
	includeItems bool
}

// Areas creates a new AreaQuery for querying areas.
func (c *Client) Areas() *AreaQuery {
	return &AreaQuery{
		client: c,
	}
}

// WithUUID filters areas by UUID.
func (q *AreaQuery) WithUUID(uuid string) *AreaQuery {
	q.uuid = &uuid
	return q
}

// WithTag filters areas by tag.
func (q *AreaQuery) WithTag(tag any) *AreaQuery {
	q.tagTitle = tag
	return q
}

// IncludeItems includes tasks in each area.
func (q *AreaQuery) IncludeItems(include bool) *AreaQuery {
	q.includeItems = include
	return q
}

// All executes the query and returns all matching areas.
func (q *AreaQuery) All(ctx context.Context) ([]Area, error) {
	var conditions []string
	conditions = append(conditions, "TRUE")

	if q.uuid != nil {
		conditions = append(conditions, makeFilter("AREA.uuid", *q.uuid)[4:]) // Remove "AND "
	}
	if filter := makeFilter("TAG.title", q.tagTitle); filter != "" {
		conditions = append(conditions, filter[4:]) // Remove "AND "
	}

	where := ""
	for i, c := range conditions {
		if i == 0 {
			where = c
		} else {
			where += " AND " + c
		}
	}

	sql := buildAreasSQL(where)
	rows, err := q.client.executeQuery(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []Area
	for rows.Next() {
		area, err := scanArea(rows)
		if err != nil {
			return nil, err
		}

		// Load tags if present
		if area.Tags != nil {
			tags, err := q.client.getTagsOfArea(ctx, area.UUID)
			if err != nil {
				return nil, err
			}
			area.Tags = tags
		}

		// Load items if requested
		if q.includeItems {
			items, err := q.client.Tasks().
				InArea(area.UUID).
				IncludeItems(true).
				All(ctx)
			if err != nil {
				return nil, err
			}
			area.Items = items
		}

		areas = append(areas, *area)
	}

	return areas, rows.Err()
}

// First executes the query and returns the first matching area.
func (q *AreaQuery) First(ctx context.Context) (*Area, error) {
	areas, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(areas) == 0 {
		return nil, ErrAreaNotFound
	}
	return &areas[0], nil
}

// Count executes the query and returns the count of matching areas.
func (q *AreaQuery) Count(ctx context.Context) (int, error) {
	var conditions []string
	conditions = append(conditions, "TRUE")

	if q.uuid != nil {
		conditions = append(conditions, makeFilter("AREA.uuid", *q.uuid)[4:])
	}
	if filter := makeFilter("TAG.title", q.tagTitle); filter != "" {
		conditions = append(conditions, filter[4:])
	}

	where := ""
	for i, c := range conditions {
		if i == 0 {
			where = c
		} else {
			where += " AND " + c
		}
	}

	areaSQL := buildAreasSQL(where)
	countSQL := buildCountSQL(areaSQL)

	var count int
	if err := q.client.executeQueryRow(ctx, countSQL).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// TagQuery provides a fluent interface for building tag queries.
type TagQuery struct {
	client *Client

	title        *string
	includeItems bool
}

// Tags creates a new TagQuery for querying tags.
func (c *Client) Tags() *TagQuery {
	return &TagQuery{
		client: c,
	}
}

// WithTitle filters tags by title.
func (q *TagQuery) WithTitle(title string) *TagQuery {
	q.title = &title
	return q
}

// IncludeItems includes areas and tasks for each tag.
func (q *TagQuery) IncludeItems(include bool) *TagQuery {
	q.includeItems = include
	return q
}

// All executes the query and returns all matching tags.
func (q *TagQuery) All(ctx context.Context) ([]Tag, error) {
	var conditions []string
	conditions = append(conditions, "TRUE")

	if q.title != nil {
		conditions = append(conditions, makeFilter("title", *q.title)[4:])
	}

	where := ""
	for i, c := range conditions {
		if i == 0 {
			where = c
		} else {
			where += " AND " + c
		}
	}

	sql := buildTagsSQL(where)
	rows, err := q.client.executeQuery(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		tag, err := scanTag(rows)
		if err != nil {
			return nil, err
		}

		// Load items if requested
		if q.includeItems {
			areas, err := q.client.Areas().WithTag(tag.Title).All(ctx)
			if err != nil {
				return nil, err
			}
			tasks, err := q.client.Tasks().WithTag(tag.Title).All(ctx)
			if err != nil {
				return nil, err
			}

			items := make([]any, 0, len(areas)+len(tasks))
			for i := range areas {
				items = append(items, &areas[i])
			}
			for i := range tasks {
				items = append(items, &tasks[i])
			}
			tag.Items = items
		}

		tags = append(tags, *tag)
	}

	return tags, rows.Err()
}

// First executes the query and returns the first matching tag.
func (q *TagQuery) First(ctx context.Context) (*Tag, error) {
	tags, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, ErrTagNotFound
	}
	return &tags[0], nil
}
