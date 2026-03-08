package things3

import (
	"context"
	"time"

	"github.com/moond4rk/things3/internal/scheme"
)

// ============================================================================
// Layer 1: Reusable Small Interfaces (Terminal Operations)
// ============================================================================

// TaskQueryExecutor executes task queries and returns results.
type TaskQueryExecutor interface {
	All(ctx context.Context) ([]Task, error)
	First(ctx context.Context) (*Task, error)
	Count(ctx context.Context) (int, error)
}

// AreaQueryExecutor executes area queries and returns results.
type AreaQueryExecutor interface {
	All(ctx context.Context) ([]Area, error)
	First(ctx context.Context) (*Area, error)
	Count(ctx context.Context) (int, error)
}

// TagQueryExecutor executes tag queries and returns results.
type TagQueryExecutor interface {
	All(ctx context.Context) ([]Tag, error)
	First(ctx context.Context) (*Tag, error)
}

// ============================================================================
// Layer 2: Sub-builder Interfaces (Already Small Interfaces)
// ============================================================================

// TypeFilterBuilder provides type-safe task type filtering.
type TypeFilterBuilder interface {
	Todo() TaskQueryBuilder
	Project() TaskQueryBuilder
	Heading() TaskQueryBuilder
}

// StatusFilterBuilder provides type-safe status filtering.
type StatusFilterBuilder interface {
	Incomplete() TaskQueryBuilder
	Completed() TaskQueryBuilder
	Canceled() TaskQueryBuilder
	Any() TaskQueryBuilder
}

// StartFilterBuilder provides type-safe start bucket filtering.
type StartFilterBuilder interface {
	Inbox() TaskQueryBuilder
	Anytime() TaskQueryBuilder
	Someday() TaskQueryBuilder
}

// DateFilterBuilder provides type-safe date filtering.
type DateFilterBuilder interface {
	Exists(has bool) TaskQueryBuilder
	Future() TaskQueryBuilder
	Past() TaskQueryBuilder
	On(date time.Time) TaskQueryBuilder
	Before(date time.Time) TaskQueryBuilder
	OnOrBefore(date time.Time) TaskQueryBuilder
	After(date time.Time) TaskQueryBuilder
	OnOrAfter(date time.Time) TaskQueryBuilder
}

// ============================================================================
// Layer 3: Functional Group Interfaces (For Composition)
// ============================================================================

// TaskRelationFilter provides relation-based filtering for tasks.
type TaskRelationFilter interface {
	InArea(uuid string) TaskQueryBuilder
	HasArea(has bool) TaskQueryBuilder
	InProject(uuid string) TaskQueryBuilder
	HasProject(has bool) TaskQueryBuilder
	InHeading(uuid string) TaskQueryBuilder
	HasHeading(has bool) TaskQueryBuilder
	InTag(title string) TaskQueryBuilder
	HasTag(has bool) TaskQueryBuilder
}

// TaskStateFilter provides state and type filtering for tasks.
type TaskStateFilter interface {
	Type() TypeFilterBuilder
	Status() StatusFilterBuilder
	Start() StartFilterBuilder
	Trashed(trashed bool) TaskQueryBuilder
	ContextTrashed(trashed bool) TaskQueryBuilder
}

// TaskTimeFilter provides time-based filtering for tasks.
type TaskTimeFilter interface {
	CreatedAfter(t time.Time) TaskQueryBuilder
	StartDate() DateFilterBuilder
	StopDate() DateFilterBuilder
	Deadline() DateFilterBuilder
}

// ============================================================================
// Layer 4: Composed Query Builder Interfaces
// ============================================================================

// TaskQueryBuilder provides a fluent interface for building task queries.
// Composed of: TaskQueryExecutor + TaskRelationFilter + TaskStateFilter + TaskTimeFilter
type TaskQueryBuilder interface {
	TaskQueryExecutor
	TaskRelationFilter
	TaskStateFilter
	TaskTimeFilter

	WithUUID(uuid string) TaskQueryBuilder
	WithUUIDPrefix(prefix string) TaskQueryBuilder
	WithDeadlineSuppressed(suppressed bool) TaskQueryBuilder
	Search(query string) TaskQueryBuilder
	OrderByTodayIndex() TaskQueryBuilder
	IncludeItems(include bool) TaskQueryBuilder
}

// AreaQueryBuilder provides a fluent interface for building area queries.
type AreaQueryBuilder interface {
	AreaQueryExecutor

	WithUUID(uuid string) AreaQueryBuilder
	WithTitle(title string) AreaQueryBuilder
	Visible(visible bool) AreaQueryBuilder
	InTag(title string) AreaQueryBuilder
	HasTag(has bool) AreaQueryBuilder
	IncludeItems(include bool) AreaQueryBuilder
}

// TagQueryBuilder provides a fluent interface for building tag queries.
type TagQueryBuilder interface {
	TagQueryExecutor

	WithUUID(uuid string) TagQueryBuilder
	WithTitle(title string) TagQueryBuilder
	WithParent(parentUUID string) TagQueryBuilder
	IncludeItems(include bool) TagQueryBuilder
}

// ============================================================================
// Layer 5: URL Scheme Builder Interfaces (aliased from internal/scheme)
// ============================================================================

// URLBuilder builds and executes Things URL schemes.
type URLBuilder = scheme.URLBuilder

// TodoAdder builds URLs for creating new to-dos.
type TodoAdder = scheme.TodoAdder

// ProjectAdder builds URLs for creating new projects.
type ProjectAdder = scheme.ProjectAdder

// TodoUpdater builds URLs for updating existing to-dos.
type TodoUpdater = scheme.TodoUpdater

// ProjectUpdater builds URLs for updating existing projects.
type ProjectUpdater = scheme.ProjectUpdater

// ShowNavigator builds URLs for navigating to items or lists.
type ShowNavigator = scheme.ShowNavigator

// ============================================================================
// Layer 6: Batch Operation Interfaces (aliased from internal/scheme)
// ============================================================================

// BatchCreator builds URLs for batch create operations.
type BatchCreator = scheme.BatchCreator

// AuthBatchCreator builds URLs for batch operations including updates.
type AuthBatchCreator = scheme.AuthBatchCreator

// BatchTodoConfigurator configures a to-do entry for batch operations.
type BatchTodoConfigurator = scheme.BatchTodoConfigurator

// BatchProjectConfigurator configures a project entry for batch operations.
type BatchProjectConfigurator = scheme.BatchProjectConfigurator
