package things3

import (
	"context"
	"time"

	"github.com/moond4rk/things3/internal/database"
)

// =============================================================================
// Shared Internal Query
// =============================================================================

// taskQuery holds the shared query state for todo, project, and heading builders.
//
// Builders are copy-on-write: every chainable method shallow-copies the state
// and returns a new builder, so a builder value can be forked into independent
// queries. The shallow copy is safe because setters always replace whole
// pointer fields instead of mutating through them.
type taskQuery struct {
	database         *db
	filter           database.TaskFilter
	includeChecklist bool
}

// =============================================================================
// TodoQuery Builder
// =============================================================================

// todoQuery provides a fluent interface for building todo queries.
type todoQuery struct {
	inner taskQuery
}

// Todos creates a new todoQuery for querying todos.
func (d *db) Todos() *todoQuery {
	taskType := int(taskTypeTodo)
	return &todoQuery{
		inner: taskQuery{
			database: d,
			filter: database.TaskFilter{
				Index:    database.IndexDefault,
				TaskType: &taskType,
			},
		},
	}
}

// clone returns a shallow copy of the query for copy-on-write chaining.
func (q *todoQuery) clone() *todoQuery {
	c := *q
	return &c
}

// withFilter clones the query, applies mutate to the clone's filter, and
// returns the clone.
func (q *todoQuery) withFilter(mutate func(*database.TaskFilter)) TodoQueryBuilder {
	c := q.clone()
	mutate(&c.inner.filter)
	return c
}

// WithUUID filters todos by UUID (exact match).
func (q *todoQuery) WithUUID(uuid string) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.UUID = &uuid })
}

// WithUUIDPrefix filters todos by UUID prefix (LIKE match).
func (q *todoQuery) WithUUIDPrefix(prefix string) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.UUIDPrefix = &prefix })
}

// WithTitle filters todos by title (keyword match).
func (q *todoQuery) WithTitle(title string) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.Title = &title })
}

// Status returns a StatusFilter for type-safe status filtering.
func (q *todoQuery) Status() StatusFilter[TodoQueryBuilder] {
	return &statusFilter[TodoQueryBuilder]{with: q.withFilter}
}

// Start returns a StartFilter for type-safe start bucket filtering.
func (q *todoQuery) Start() StartFilter[TodoQueryBuilder] {
	return &startFilter[TodoQueryBuilder]{with: q.withFilter}
}

// Trashed filters todos by trash status.
func (q *todoQuery) Trashed(trashed bool) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.Trashed = &trashed })
}

// InArea filters todos by a specific area UUID.
func (q *todoQuery) InArea(uuid string) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.AreaUUID = &uuid })
}

// HasArea filters todos by whether they have an area.
func (q *todoQuery) HasArea(has bool) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.HasArea = &has })
}

// InProject filters todos by a specific project UUID.
func (q *todoQuery) InProject(uuid string) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.ProjectUUID = &uuid })
}

// HasProject filters todos by whether they have a project.
func (q *todoQuery) HasProject(has bool) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.HasProject = &has })
}

// InHeading filters todos by a specific heading UUID.
func (q *todoQuery) InHeading(uuid string) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.HeadingUUID = &uuid })
}

// HasHeading filters todos by whether they have a heading.
func (q *todoQuery) HasHeading(has bool) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.HasHeading = &has })
}

// InTag filters todos by a specific tag title.
func (q *todoQuery) InTag(title string) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.TagTitle = &title })
}

// HasTag filters todos by whether they have any tags.
func (q *todoQuery) HasTag(has bool) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.HasTags = &has })
}

// StartDate returns a DateFilter for start date filtering.
func (q *todoQuery) StartDate() DateFilter[TodoQueryBuilder] {
	return &dateFilter[TodoQueryBuilder]{with: q.withFilter, field: dateFieldStartDate}
}

// StopDate returns a DateFilter for stop date filtering.
func (q *todoQuery) StopDate() DateFilter[TodoQueryBuilder] {
	return &dateFilter[TodoQueryBuilder]{with: q.withFilter, field: dateFieldStopDate}
}

// Deadline returns a DateFilter for deadline filtering.
func (q *todoQuery) Deadline() DateFilter[TodoQueryBuilder] {
	return &dateFilter[TodoQueryBuilder]{with: q.withFilter, field: dateFieldDeadline}
}

// deadlineSuppressed filters todos by whether the deadline has been suppressed.
// It is unexported: deadline suppression is a database internal, and Today is
// its only consumer.
func (q *todoQuery) deadlineSuppressed(suppressed bool) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.DeadlineSuppressed = &suppressed })
}

// repeatingTemplates restricts the query to repeating templates (rows carrying a
// recurrence rule), whose start-date filter targets the next occurrence. It is
// unexported: repeating templates are a database internal, and Upcoming is its
// only consumer.
func (q *todoQuery) repeatingTemplates() TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.RepeatingTemplates = new(true) })
}

// CreatedAfter filters todos created after the specified time.
func (q *todoQuery) CreatedAfter(t time.Time) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.CreatedAfter = &t })
}

// Search filters todos by a search query.
func (q *todoQuery) Search(query string) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.SearchQuery = &query })
}

// OrderByTodayIndex orders results by today index instead of default index.
func (q *todoQuery) OrderByTodayIndex() TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.Index = database.IndexToday })
}

// Limit restricts the maximum number of results returned.
func (q *todoQuery) Limit(n int) TodoQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.Limit = &n })
}

// IncludeChecklist opts in to loading checklist items for each todo.
func (q *todoQuery) IncludeChecklist() TodoQueryBuilder {
	c := q.clone()
	c.inner.includeChecklist = true
	return c
}

// All executes the query and returns all matching todos.
// The result is never nil; an empty result encodes as a JSON array.
func (q *todoQuery) All(ctx context.Context) ([]Todo, error) {
	rows, err := q.inner.database.inner.QueryTasks(ctx, &q.inner.filter)
	if err != nil {
		return nil, err
	}

	todos := make([]Todo, 0, len(rows))
	for i := range rows {
		todo := convertTaskRowToTodo(&rows[i])

		// Load tags if present
		if rows[i].HasTags {
			tags, err := q.inner.database.inner.TagsOfTask(ctx, rows[i].UUID)
			if err != nil {
				return nil, err
			}
			todo.Tags = tags
		}

		// Load checklist if requested
		if q.inner.includeChecklist && rows[i].HasChecklist {
			clRows, err := q.inner.database.inner.QueryChecklistItems(ctx, rows[i].UUID)
			if err != nil {
				return nil, err
			}
			todo.Checklist = convertChecklistItemRows(clRows)
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

// First executes the query and returns the first matching todo.
// Unlike All, First always loads the checklist and fetches at most one row.
// Both adjustments apply to a private copy, leaving the receiver unchanged.
func (q *todoQuery) First(ctx context.Context) (*Todo, error) {
	one := 1
	c := q.clone()
	c.inner.includeChecklist = true
	c.inner.filter.Limit = &one

	todos, err := c.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(todos) == 0 {
		return nil, ErrTodoNotFound
	}
	return &todos[0], nil
}

// Count executes the query and returns the count of matching todos.
func (q *todoQuery) Count(ctx context.Context) (int, error) {
	return q.inner.database.inner.CountTasks(ctx, &q.inner.filter)
}

// =============================================================================
// ProjectQuery Builder
// =============================================================================

// projectQuery provides a fluent interface for building project queries.
type projectQuery struct {
	inner taskQuery
}

// Projects creates a new projectQuery for querying projects.
func (d *db) Projects() *projectQuery {
	taskType := int(taskTypeProject)
	return &projectQuery{
		inner: taskQuery{
			database: d,
			filter: database.TaskFilter{
				Index:    database.IndexDefault,
				TaskType: &taskType,
			},
		},
	}
}

// clone returns a shallow copy of the query for copy-on-write chaining.
func (q *projectQuery) clone() *projectQuery {
	c := *q
	return &c
}

// withFilter clones the query, applies mutate to the clone's filter, and
// returns the clone.
func (q *projectQuery) withFilter(mutate func(*database.TaskFilter)) ProjectQueryBuilder {
	c := q.clone()
	mutate(&c.inner.filter)
	return c
}

// WithUUID filters projects by UUID (exact match).
func (q *projectQuery) WithUUID(uuid string) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.UUID = &uuid })
}

// WithUUIDPrefix filters projects by UUID prefix (LIKE match).
func (q *projectQuery) WithUUIDPrefix(prefix string) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.UUIDPrefix = &prefix })
}

// WithTitle filters projects by title (keyword match).
func (q *projectQuery) WithTitle(title string) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.Title = &title })
}

// Status returns a StatusFilter for type-safe status filtering.
func (q *projectQuery) Status() StatusFilter[ProjectQueryBuilder] {
	return &statusFilter[ProjectQueryBuilder]{with: q.withFilter}
}

// Start returns a StartFilter for type-safe start bucket filtering.
func (q *projectQuery) Start() StartFilter[ProjectQueryBuilder] {
	return &startFilter[ProjectQueryBuilder]{with: q.withFilter}
}

// Trashed filters projects by trash status.
func (q *projectQuery) Trashed(trashed bool) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.Trashed = &trashed })
}

// InArea filters projects by a specific area UUID.
func (q *projectQuery) InArea(uuid string) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.AreaUUID = &uuid })
}

// HasArea filters projects by whether they have an area.
func (q *projectQuery) HasArea(has bool) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.HasArea = &has })
}

// InTag filters projects by a specific tag title.
func (q *projectQuery) InTag(title string) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.TagTitle = &title })
}

// HasTag filters projects by whether they have any tags.
func (q *projectQuery) HasTag(has bool) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.HasTags = &has })
}

// StartDate returns a DateFilter for start date filtering.
func (q *projectQuery) StartDate() DateFilter[ProjectQueryBuilder] {
	return &dateFilter[ProjectQueryBuilder]{with: q.withFilter, field: dateFieldStartDate}
}

// StopDate returns a DateFilter for stop date filtering.
func (q *projectQuery) StopDate() DateFilter[ProjectQueryBuilder] {
	return &dateFilter[ProjectQueryBuilder]{with: q.withFilter, field: dateFieldStopDate}
}

// Deadline returns a DateFilter for deadline filtering.
func (q *projectQuery) Deadline() DateFilter[ProjectQueryBuilder] {
	return &dateFilter[ProjectQueryBuilder]{with: q.withFilter, field: dateFieldDeadline}
}

// CreatedAfter filters projects created after the specified time.
func (q *projectQuery) CreatedAfter(t time.Time) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.CreatedAfter = &t })
}

// Search filters projects by a search query.
func (q *projectQuery) Search(query string) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.SearchQuery = &query })
}

// Limit restricts the maximum number of results returned.
func (q *projectQuery) Limit(n int) ProjectQueryBuilder {
	return q.withFilter(func(f *database.TaskFilter) { f.Limit = &n })
}

// All executes the query and returns all matching projects.
// The result is never nil; an empty result encodes as a JSON array.
func (q *projectQuery) All(ctx context.Context) ([]Project, error) {
	rows, err := q.inner.database.inner.QueryTasks(ctx, &q.inner.filter)
	if err != nil {
		return nil, err
	}

	projects := make([]Project, 0, len(rows))
	for i := range rows {
		project := convertTaskRowToProject(&rows[i])

		// Load tags if present
		if rows[i].HasTags {
			tags, err := q.inner.database.inner.TagsOfTask(ctx, rows[i].UUID)
			if err != nil {
				return nil, err
			}
			project.Tags = tags
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// First executes the query and returns the first matching project.
// It fetches at most one row via a private copy, leaving the receiver unchanged.
func (q *projectQuery) First(ctx context.Context) (*Project, error) {
	one := 1
	c := q.clone()
	c.inner.filter.Limit = &one

	projects, err := c.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, ErrProjectNotFound
	}
	return &projects[0], nil
}

// Count executes the query and returns the count of matching projects.
func (q *projectQuery) Count(ctx context.Context) (int, error) {
	return q.inner.database.inner.CountTasks(ctx, &q.inner.filter)
}

// =============================================================================
// HeadingQuery Builder
// =============================================================================

// headingQuery provides a fluent interface for building heading queries.
type headingQuery struct {
	inner taskQuery
}

// Headings creates a new headingQuery for querying headings.
func (d *db) Headings() *headingQuery {
	taskType := int(taskTypeHeading)
	return &headingQuery{
		inner: taskQuery{
			database: d,
			filter: database.TaskFilter{
				Index:    database.IndexDefault,
				TaskType: &taskType,
			},
		},
	}
}

// clone returns a shallow copy of the query for copy-on-write chaining.
func (q *headingQuery) clone() *headingQuery {
	c := *q
	return &c
}

// WithUUID filters headings by UUID (exact match).
func (q *headingQuery) WithUUID(uuid string) HeadingQueryBuilder {
	c := q.clone()
	c.inner.filter.UUID = &uuid
	return c
}

// WithUUIDPrefix filters headings by UUID prefix (LIKE match).
func (q *headingQuery) WithUUIDPrefix(prefix string) HeadingQueryBuilder {
	c := q.clone()
	c.inner.filter.UUIDPrefix = &prefix
	return c
}

// InProject filters headings by a specific project UUID.
func (q *headingQuery) InProject(uuid string) HeadingQueryBuilder {
	c := q.clone()
	c.inner.filter.ProjectUUID = &uuid
	return c
}

// Limit restricts the maximum number of results returned.
func (q *headingQuery) Limit(n int) HeadingQueryBuilder {
	c := q.clone()
	c.inner.filter.Limit = &n
	return c
}

// All executes the query and returns all matching headings.
// The result is never nil; an empty result encodes as a JSON array.
func (q *headingQuery) All(ctx context.Context) ([]Heading, error) {
	rows, err := q.inner.database.inner.QueryTasks(ctx, &q.inner.filter)
	if err != nil {
		return nil, err
	}

	headings := make([]Heading, len(rows))
	for i := range rows {
		headings[i] = convertTaskRowToHeading(&rows[i])
	}

	return headings, nil
}

// First executes the query and returns the first matching heading.
// It fetches at most one row via a private copy, leaving the receiver unchanged.
func (q *headingQuery) First(ctx context.Context) (*Heading, error) {
	one := 1
	c := q.clone()
	c.inner.filter.Limit = &one

	headings, err := c.All(ctx)
	if err != nil {
		return nil, err
	}
	if len(headings) == 0 {
		return nil, ErrHeadingNotFound
	}
	return &headings[0], nil
}

// Count executes the query and returns the count of matching headings.
func (q *headingQuery) Count(ctx context.Context) (int, error) {
	return q.inner.database.inner.CountTasks(ctx, &q.inner.filter)
}
