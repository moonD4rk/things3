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
	inner *taskQuery
}

// Todos creates a new todoQuery for querying todos.
func (d *db) Todos() *todoQuery {
	taskType := int(taskTypeTodo)
	return &todoQuery{
		inner: &taskQuery{
			database: d,
			filter: database.TaskFilter{
				Index:    database.IndexDefault,
				TaskType: &taskType,
			},
		},
	}
}

// WithUUID filters todos by UUID (exact match).
func (q *todoQuery) WithUUID(uuid string) TodoQueryBuilder {
	q.inner.filter.UUID = &uuid
	return q
}

// Status returns a StatusFilter for type-safe status filtering.
func (q *todoQuery) Status() StatusFilter[TodoQueryBuilder] {
	return &statusFilter[TodoQueryBuilder]{query: q.inner, parent: q}
}

// Start returns a StartFilter for type-safe start bucket filtering.
func (q *todoQuery) Start() StartFilter[TodoQueryBuilder] {
	return &startFilter[TodoQueryBuilder]{query: q.inner, parent: q}
}

// Trashed filters todos by trash status.
func (q *todoQuery) Trashed(trashed bool) TodoQueryBuilder {
	q.inner.filter.Trashed = &trashed
	return q
}

// ContextTrashed filters todos by the trash status of their context (project/heading).
func (q *todoQuery) ContextTrashed(trashed bool) TodoQueryBuilder {
	q.inner.filter.ContextTrashed = &trashed
	return q
}

// InArea filters todos by a specific area UUID.
func (q *todoQuery) InArea(uuid string) TodoQueryBuilder {
	q.inner.filter.AreaUUID = &uuid
	return q
}

// HasArea filters todos by whether they have an area.
func (q *todoQuery) HasArea(has bool) TodoQueryBuilder {
	q.inner.filter.HasArea = &has
	return q
}

// InProject filters todos by a specific project UUID.
func (q *todoQuery) InProject(uuid string) TodoQueryBuilder {
	q.inner.filter.ProjectUUID = &uuid
	return q
}

// HasProject filters todos by whether they have a project.
func (q *todoQuery) HasProject(has bool) TodoQueryBuilder {
	q.inner.filter.HasProject = &has
	return q
}

// InHeading filters todos by a specific heading UUID.
func (q *todoQuery) InHeading(uuid string) TodoQueryBuilder {
	q.inner.filter.HeadingUUID = &uuid
	return q
}

// HasHeading filters todos by whether they have a heading.
func (q *todoQuery) HasHeading(has bool) TodoQueryBuilder {
	q.inner.filter.HasHeading = &has
	return q
}

// InTag filters todos by a specific tag title.
func (q *todoQuery) InTag(title string) TodoQueryBuilder {
	q.inner.filter.TagTitle = &title
	return q
}

// HasTag filters todos by whether they have any tags.
func (q *todoQuery) HasTag(has bool) TodoQueryBuilder {
	q.inner.filter.HasTags = &has
	return q
}

// StartDate returns a DateFilter for start date filtering.
func (q *todoQuery) StartDate() DateFilter[TodoQueryBuilder] {
	return &dateFilter[TodoQueryBuilder]{query: q.inner, parent: q, field: dateFieldStartDate}
}

// StopDate returns a DateFilter for stop date filtering.
func (q *todoQuery) StopDate() DateFilter[TodoQueryBuilder] {
	return &dateFilter[TodoQueryBuilder]{query: q.inner, parent: q, field: dateFieldStopDate}
}

// Deadline returns a DateFilter for deadline filtering.
func (q *todoQuery) Deadline() DateFilter[TodoQueryBuilder] {
	return &dateFilter[TodoQueryBuilder]{query: q.inner, parent: q, field: dateFieldDeadline}
}

// CreatedAfter filters todos created after the specified time.
func (q *todoQuery) CreatedAfter(t time.Time) TodoQueryBuilder {
	q.inner.filter.CreatedAfter = &t
	return q
}

// Search filters todos by a search query.
func (q *todoQuery) Search(query string) TodoQueryBuilder {
	q.inner.filter.SearchQuery = &query
	return q
}

// OrderByTodayIndex orders results by today index instead of default index.
func (q *todoQuery) OrderByTodayIndex() TodoQueryBuilder {
	q.inner.filter.Index = database.IndexToday
	return q
}

// Limit restricts the maximum number of results returned.
func (q *todoQuery) Limit(n int) TodoQueryBuilder {
	q.inner.filter.Limit = &n
	return q
}

// IncludeChecklist opts in to loading checklist items for each todo.
func (q *todoQuery) IncludeChecklist() TodoQueryBuilder {
	q.inner.includeChecklist = true
	return q
}

// All executes the query and returns all matching todos.
func (q *todoQuery) All(ctx context.Context) ([]Todo, error) {
	rows, err := q.inner.database.inner.QueryTasks(ctx, &q.inner.filter)
	if err != nil {
		return nil, err
	}

	var todos []Todo
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
func (q *todoQuery) First(ctx context.Context) (*Todo, error) {
	// For single todo fetch, always include checklist
	q.inner.includeChecklist = true

	todos, err := q.All(ctx)
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
	inner *taskQuery
}

// Projects creates a new projectQuery for querying projects.
func (d *db) Projects() *projectQuery {
	taskType := int(taskTypeProject)
	return &projectQuery{
		inner: &taskQuery{
			database: d,
			filter: database.TaskFilter{
				Index:    database.IndexDefault,
				TaskType: &taskType,
			},
		},
	}
}

// WithUUID filters projects by UUID (exact match).
func (q *projectQuery) WithUUID(uuid string) ProjectQueryBuilder {
	q.inner.filter.UUID = &uuid
	return q
}

// Status returns a StatusFilter for type-safe status filtering.
func (q *projectQuery) Status() StatusFilter[ProjectQueryBuilder] {
	return &statusFilter[ProjectQueryBuilder]{query: q.inner, parent: q}
}

// Start returns a StartFilter for type-safe start bucket filtering.
func (q *projectQuery) Start() StartFilter[ProjectQueryBuilder] {
	return &startFilter[ProjectQueryBuilder]{query: q.inner, parent: q}
}

// Trashed filters projects by trash status.
func (q *projectQuery) Trashed(trashed bool) ProjectQueryBuilder {
	q.inner.filter.Trashed = &trashed
	return q
}

// InArea filters projects by a specific area UUID.
func (q *projectQuery) InArea(uuid string) ProjectQueryBuilder {
	q.inner.filter.AreaUUID = &uuid
	return q
}

// HasArea filters projects by whether they have an area.
func (q *projectQuery) HasArea(has bool) ProjectQueryBuilder {
	q.inner.filter.HasArea = &has
	return q
}

// InTag filters projects by a specific tag title.
func (q *projectQuery) InTag(title string) ProjectQueryBuilder {
	q.inner.filter.TagTitle = &title
	return q
}

// HasTag filters projects by whether they have any tags.
func (q *projectQuery) HasTag(has bool) ProjectQueryBuilder {
	q.inner.filter.HasTags = &has
	return q
}

// StartDate returns a DateFilter for start date filtering.
func (q *projectQuery) StartDate() DateFilter[ProjectQueryBuilder] {
	return &dateFilter[ProjectQueryBuilder]{query: q.inner, parent: q, field: dateFieldStartDate}
}

// StopDate returns a DateFilter for stop date filtering.
func (q *projectQuery) StopDate() DateFilter[ProjectQueryBuilder] {
	return &dateFilter[ProjectQueryBuilder]{query: q.inner, parent: q, field: dateFieldStopDate}
}

// Deadline returns a DateFilter for deadline filtering.
func (q *projectQuery) Deadline() DateFilter[ProjectQueryBuilder] {
	return &dateFilter[ProjectQueryBuilder]{query: q.inner, parent: q, field: dateFieldDeadline}
}

// CreatedAfter filters projects created after the specified time.
func (q *projectQuery) CreatedAfter(t time.Time) ProjectQueryBuilder {
	q.inner.filter.CreatedAfter = &t
	return q
}

// Search filters projects by a search query.
func (q *projectQuery) Search(query string) ProjectQueryBuilder {
	q.inner.filter.SearchQuery = &query
	return q
}

// Limit restricts the maximum number of results returned.
func (q *projectQuery) Limit(n int) ProjectQueryBuilder {
	q.inner.filter.Limit = &n
	return q
}

// All executes the query and returns all matching projects.
func (q *projectQuery) All(ctx context.Context) ([]Project, error) {
	rows, err := q.inner.database.inner.QueryTasks(ctx, &q.inner.filter)
	if err != nil {
		return nil, err
	}

	var projects []Project
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
func (q *projectQuery) First(ctx context.Context) (*Project, error) {
	projects, err := q.All(ctx)
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
	inner *taskQuery
}

// Headings creates a new headingQuery for querying headings.
func (d *db) Headings() *headingQuery {
	taskType := int(taskTypeHeading)
	return &headingQuery{
		inner: &taskQuery{
			database: d,
			filter: database.TaskFilter{
				Index:    database.IndexDefault,
				TaskType: &taskType,
			},
		},
	}
}

// WithUUID filters headings by UUID (exact match).
func (q *headingQuery) WithUUID(uuid string) HeadingQueryBuilder {
	q.inner.filter.UUID = &uuid
	return q
}

// InProject filters headings by a specific project UUID.
func (q *headingQuery) InProject(uuid string) HeadingQueryBuilder {
	q.inner.filter.ProjectUUID = &uuid
	return q
}

// Limit restricts the maximum number of results returned.
func (q *headingQuery) Limit(n int) HeadingQueryBuilder {
	q.inner.filter.Limit = &n
	return q
}

// All executes the query and returns all matching headings.
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
func (q *headingQuery) First(ctx context.Context) (*Heading, error) {
	headings, err := q.All(ctx)
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
