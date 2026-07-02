package things3

import (
	"context"
	"time"

	"github.com/moond4rk/things3/internal/scheme"
)

// ============================================================================
// Layer 1: Terminal Operation Interfaces
// ============================================================================

// TodoQueryExecutor executes todo queries and returns results.
type TodoQueryExecutor interface {
	All(ctx context.Context) ([]Todo, error)
	First(ctx context.Context) (*Todo, error)
	Count(ctx context.Context) (int, error)
}

// ProjectQueryExecutor executes project queries and returns results.
type ProjectQueryExecutor interface {
	All(ctx context.Context) ([]Project, error)
	First(ctx context.Context) (*Project, error)
	Count(ctx context.Context) (int, error)
}

// HeadingQueryExecutor executes heading queries and returns results.
type HeadingQueryExecutor interface {
	All(ctx context.Context) ([]Heading, error)
	First(ctx context.Context) (*Heading, error)
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
// Layer 2: Generic Sub-builder Interfaces
// ============================================================================

// StatusFilter provides type-safe status filtering.
type StatusFilter[T any] interface {
	Incomplete() T
	Completed() T
	Canceled() T
	Any() T
}

// StartFilter provides type-safe start bucket filtering.
type StartFilter[T any] interface {
	Inbox() T
	Anytime() T
	Someday() T
}

// DateFilter provides type-safe date filtering.
type DateFilter[T any] interface {
	Exists(has bool) T
	Future() T
	Past() T
	On(date time.Time) T
	Before(date time.Time) T
	OnOrBefore(date time.Time) T
	After(date time.Time) T
	OnOrAfter(date time.Time) T
}

// ============================================================================
// Layer 3: Composed Query Builder Interfaces
// ============================================================================
//
// Query builders are copy-on-write: every chainable method returns a new
// builder and leaves the receiver unchanged, so a base builder can be forked
// into independent queries.

// TodoQueryBuilder provides a fluent interface for building todo queries.
type TodoQueryBuilder interface {
	TodoQueryExecutor

	WithUUID(uuid string) TodoQueryBuilder
	WithUUIDPrefix(prefix string) TodoQueryBuilder
	WithTitle(title string) TodoQueryBuilder

	Status() StatusFilter[TodoQueryBuilder]
	Start() StartFilter[TodoQueryBuilder]
	Trashed(trashed bool) TodoQueryBuilder

	InArea(uuid string) TodoQueryBuilder
	HasArea(has bool) TodoQueryBuilder
	InProject(uuid string) TodoQueryBuilder
	HasProject(has bool) TodoQueryBuilder
	InHeading(uuid string) TodoQueryBuilder
	HasHeading(has bool) TodoQueryBuilder
	InTag(title string) TodoQueryBuilder
	HasTag(has bool) TodoQueryBuilder

	StartDate() DateFilter[TodoQueryBuilder]
	StopDate() DateFilter[TodoQueryBuilder]
	Deadline() DateFilter[TodoQueryBuilder]
	DeadlineSuppressed(suppressed bool) TodoQueryBuilder
	CreatedAfter(t time.Time) TodoQueryBuilder

	Search(query string) TodoQueryBuilder
	OrderByTodayIndex() TodoQueryBuilder
	Limit(n int) TodoQueryBuilder

	IncludeChecklist() TodoQueryBuilder
}

// ProjectQueryBuilder provides a fluent interface for building project queries.
type ProjectQueryBuilder interface {
	ProjectQueryExecutor

	WithUUID(uuid string) ProjectQueryBuilder
	WithUUIDPrefix(prefix string) ProjectQueryBuilder
	WithTitle(title string) ProjectQueryBuilder

	Status() StatusFilter[ProjectQueryBuilder]
	Start() StartFilter[ProjectQueryBuilder]
	Trashed(trashed bool) ProjectQueryBuilder

	InArea(uuid string) ProjectQueryBuilder
	HasArea(has bool) ProjectQueryBuilder
	InTag(title string) ProjectQueryBuilder
	HasTag(has bool) ProjectQueryBuilder

	StartDate() DateFilter[ProjectQueryBuilder]
	StopDate() DateFilter[ProjectQueryBuilder]
	Deadline() DateFilter[ProjectQueryBuilder]
	CreatedAfter(t time.Time) ProjectQueryBuilder

	Search(query string) ProjectQueryBuilder
	Limit(n int) ProjectQueryBuilder
}

// HeadingQueryBuilder provides a fluent interface for building heading queries.
type HeadingQueryBuilder interface {
	HeadingQueryExecutor

	WithUUID(uuid string) HeadingQueryBuilder
	WithUUIDPrefix(prefix string) HeadingQueryBuilder
	InProject(uuid string) HeadingQueryBuilder
	Limit(n int) HeadingQueryBuilder
}

// AreaQueryBuilder provides a fluent interface for building area queries.
type AreaQueryBuilder interface {
	AreaQueryExecutor

	WithUUID(uuid string) AreaQueryBuilder
	WithTitle(title string) AreaQueryBuilder
	Visible(visible bool) AreaQueryBuilder
	InTag(title string) AreaQueryBuilder
	HasTag(has bool) AreaQueryBuilder
}

// TagQueryBuilder provides a fluent interface for building tag queries.
type TagQueryBuilder interface {
	TagQueryExecutor

	WithUUID(uuid string) TagQueryBuilder
	WithTitle(title string) TagQueryBuilder
	WithParent(parentUUID string) TagQueryBuilder
}

// ============================================================================
// Layer 4: URL Scheme Builder Interfaces (aliased from internal/scheme)
// ============================================================================

// URLBuilder builds and executes Things URL schemes.
type URLBuilder = scheme.URLBuilder

// TodoAdder builds URLs for creating new todos.
type TodoAdder = scheme.TodoAdder

// ProjectAdder builds URLs for creating new projects.
type ProjectAdder = scheme.ProjectAdder

// TodoUpdater builds URLs for updating existing todos.
type TodoUpdater = scheme.TodoUpdater

// ProjectUpdater builds URLs for updating existing projects.
type ProjectUpdater = scheme.ProjectUpdater

// ShowNavigator builds URLs for navigating to items or lists.
type ShowNavigator = scheme.ShowNavigator

// ============================================================================
// Layer 5: Batch Operation Interfaces (aliased from internal/scheme)
// ============================================================================

// BatchCreator builds URLs for batch create operations.
type BatchCreator = scheme.BatchCreator

// AuthBatchCreator builds URLs for batch operations including updates.
type AuthBatchCreator = scheme.AuthBatchCreator

// BatchTodoConfigurator configures a todo entry for batch operations.
type BatchTodoConfigurator = scheme.BatchTodoConfigurator

// BatchProjectConfigurator configures a project entry for batch operations.
type BatchProjectConfigurator = scheme.BatchProjectConfigurator
