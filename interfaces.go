package things3

import (
	"context"
	"time"
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

// URLBuilder builds and executes Things URL schemes.
type URLBuilder interface {
	Build() (string, error)
	Execute(ctx context.Context) error
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
	InTag(tag any) AreaQueryBuilder
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
// Layer 5: URL Scheme Builder Interfaces
// ============================================================================

// TodoAdder builds URLs for creating new to-dos.
type TodoAdder interface {
	URLBuilder

	Title(title string) TodoAdder
	Titles(titles ...string) TodoAdder
	Notes(notes string) TodoAdder
	When(t time.Time) TodoAdder
	WhenEvening() TodoAdder
	WhenAnytime() TodoAdder
	WhenSomeday() TodoAdder
	Deadline(t time.Time) TodoAdder
	Reminder(hour, minute int) TodoAdder
	Tags(tags ...string) TodoAdder
	ChecklistItems(items ...string) TodoAdder
	List(name string) TodoAdder
	ListID(id string) TodoAdder
	Heading(name string) TodoAdder
	HeadingID(id string) TodoAdder
	Completed(completed bool) TodoAdder
	Canceled(canceled bool) TodoAdder
	ShowQuickEntry(show bool) TodoAdder
	Reveal(reveal bool) TodoAdder
	CreationDate(date time.Time) TodoAdder
	CompletionDate(date time.Time) TodoAdder
}

// ProjectAdder builds URLs for creating new projects.
type ProjectAdder interface {
	URLBuilder

	Title(title string) ProjectAdder
	Notes(notes string) ProjectAdder
	When(t time.Time) ProjectAdder
	WhenEvening() ProjectAdder
	WhenAnytime() ProjectAdder
	WhenSomeday() ProjectAdder
	Deadline(t time.Time) ProjectAdder
	Reminder(hour, minute int) ProjectAdder
	Tags(tags ...string) ProjectAdder
	Area(name string) ProjectAdder
	AreaID(id string) ProjectAdder
	Todos(titles ...string) ProjectAdder
	Completed(completed bool) ProjectAdder
	Canceled(canceled bool) ProjectAdder
	Reveal(reveal bool) ProjectAdder
	CreationDate(date time.Time) ProjectAdder
	CompletionDate(date time.Time) ProjectAdder
}

// TodoUpdater builds URLs for updating existing to-dos.
type TodoUpdater interface {
	URLBuilder

	Title(title string) TodoUpdater
	Notes(notes string) TodoUpdater
	PrependNotes(notes string) TodoUpdater
	AppendNotes(notes string) TodoUpdater
	When(t time.Time) TodoUpdater
	WhenEvening() TodoUpdater
	WhenAnytime() TodoUpdater
	WhenSomeday() TodoUpdater
	Deadline(t time.Time) TodoUpdater
	ClearDeadline() TodoUpdater
	Reminder(hour, minute int) TodoUpdater
	Tags(tags ...string) TodoUpdater
	AddTags(tags ...string) TodoUpdater
	ChecklistItems(items ...string) TodoUpdater
	PrependChecklistItems(items ...string) TodoUpdater
	AppendChecklistItems(items ...string) TodoUpdater
	List(name string) TodoUpdater
	ListID(id string) TodoUpdater
	Heading(name string) TodoUpdater
	HeadingID(id string) TodoUpdater
	Completed(completed bool) TodoUpdater
	Canceled(canceled bool) TodoUpdater
	Duplicate(duplicate bool) TodoUpdater
	Reveal(reveal bool) TodoUpdater
	CreationDate(date time.Time) TodoUpdater
	CompletionDate(date time.Time) TodoUpdater
}

// ProjectUpdater builds URLs for updating existing projects.
type ProjectUpdater interface {
	URLBuilder

	Title(title string) ProjectUpdater
	Notes(notes string) ProjectUpdater
	PrependNotes(notes string) ProjectUpdater
	AppendNotes(notes string) ProjectUpdater
	When(t time.Time) ProjectUpdater
	WhenEvening() ProjectUpdater
	WhenAnytime() ProjectUpdater
	WhenSomeday() ProjectUpdater
	Deadline(t time.Time) ProjectUpdater
	ClearDeadline() ProjectUpdater
	Reminder(hour, minute int) ProjectUpdater
	Tags(tags ...string) ProjectUpdater
	AddTags(tags ...string) ProjectUpdater
	Area(name string) ProjectUpdater
	AreaID(id string) ProjectUpdater
	Completed(completed bool) ProjectUpdater
	Canceled(canceled bool) ProjectUpdater
	Reveal(reveal bool) ProjectUpdater
}

// ShowNavigator builds URLs for navigating to items or lists.
type ShowNavigator interface {
	ID(id string) ShowNavigator
	List(list ListID) ShowNavigator
	Query(query string) ShowNavigator
	Filter(tags ...string) ShowNavigator

	Build() string
	Execute(ctx context.Context) error
}

// ============================================================================
// Layer 6: Batch Operation Interfaces
// ============================================================================

// BatchCreator builds URLs for batch create operations.
type BatchCreator interface {
	AddTodo(configure func(BatchTodoConfigurator)) BatchCreator
	AddProject(configure func(BatchProjectConfigurator)) BatchCreator
	Reveal(reveal bool) BatchCreator
	Build() (string, error)
	Execute(ctx context.Context) error
}

// AuthBatchCreator builds URLs for batch operations including updates.
type AuthBatchCreator interface {
	AddTodo(configure func(BatchTodoConfigurator)) AuthBatchCreator
	AddProject(configure func(BatchProjectConfigurator)) AuthBatchCreator
	UpdateTodo(id string, configure func(BatchTodoConfigurator)) AuthBatchCreator
	UpdateProject(id string, configure func(BatchProjectConfigurator)) AuthBatchCreator
	Reveal(reveal bool) AuthBatchCreator
	Build() (string, error)
	Execute(ctx context.Context) error
}

// BatchTodoConfigurator configures a to-do entry for batch operations.
type BatchTodoConfigurator interface {
	Title(title string) BatchTodoConfigurator
	Notes(notes string) BatchTodoConfigurator
	PrependNotes(notes string) BatchTodoConfigurator
	AppendNotes(notes string) BatchTodoConfigurator
	When(t time.Time) BatchTodoConfigurator
	WhenEvening() BatchTodoConfigurator
	WhenAnytime() BatchTodoConfigurator
	WhenSomeday() BatchTodoConfigurator
	Deadline(t time.Time) BatchTodoConfigurator
	Tags(tags ...string) BatchTodoConfigurator
	AddTags(tags ...string) BatchTodoConfigurator
	ChecklistItems(items ...string) BatchTodoConfigurator
	List(name string) BatchTodoConfigurator
	ListID(id string) BatchTodoConfigurator
	Heading(name string) BatchTodoConfigurator
	Completed(completed bool) BatchTodoConfigurator
	Canceled(canceled bool) BatchTodoConfigurator
	CreationDate(date time.Time) BatchTodoConfigurator
	CompletionDate(date time.Time) BatchTodoConfigurator
}

// BatchProjectConfigurator configures a project entry for batch operations.
type BatchProjectConfigurator interface {
	Title(title string) BatchProjectConfigurator
	Notes(notes string) BatchProjectConfigurator
	PrependNotes(notes string) BatchProjectConfigurator
	AppendNotes(notes string) BatchProjectConfigurator
	When(t time.Time) BatchProjectConfigurator
	WhenEvening() BatchProjectConfigurator
	WhenAnytime() BatchProjectConfigurator
	WhenSomeday() BatchProjectConfigurator
	Deadline(t time.Time) BatchProjectConfigurator
	Tags(tags ...string) BatchProjectConfigurator
	AddTags(tags ...string) BatchProjectConfigurator
	Area(name string) BatchProjectConfigurator
	AreaID(id string) BatchProjectConfigurator
	Todos(configs ...func(BatchTodoConfigurator)) BatchProjectConfigurator
	Completed(completed bool) BatchProjectConfigurator
	Canceled(canceled bool) BatchProjectConfigurator
	CreationDate(date time.Time) BatchProjectConfigurator
	CompletionDate(date time.Time) BatchProjectConfigurator
}
