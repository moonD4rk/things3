# RFC 006: Interface Abstraction

Status: Draft
Author: @moond4rk

## Summary

This RFC defines interface abstractions for the things3 library to provide a cleaner public API. Users interact solely with `Client` and well-defined interfaces, hiding all internal implementation details.

## Motivation

### Current Issues

1. **Multiple Entry Points**: `NewClient()`, `NewDB()`, `NewScheme()` all public
2. **Concrete Type Exposure**: `Tasks()` returns `*TaskQuery`, exposing internal fields
3. **IDE Clutter**: Autocomplete shows internal struct fields and methods
4. **Unclear API Boundary**: Users unsure which types are part of public API

### Goals

- Single entry point: `NewClient()` only
- Return interfaces instead of concrete types
- Hide implementation details from IDE autocomplete
- Clean, minimal public API surface

### Non-Goals

- Backward compatibility (breaking changes are acceptable)
- Maintaining `NewDB()` or `NewScheme()` as public APIs

## Design

### Architecture Overview

```
Public API (Exported)
==================================================
NewClient(opts...) -> *Client

Client
  +-- [Query - Convenience Methods]
  |   Inbox(ctx), Today(ctx), Todos(ctx), Projects(ctx), ...
  |
  +-- [Query - Builders]
  |   Tasks()   -> TaskQuerier
  |   Areas()   -> AreaQuerier
  |   Tags()    -> TagQuerier
  |
  +-- [Add Operations]
  |   AddTodo()    -> TodoAdder
  |   AddProject() -> ProjectAdder
  |   Batch()      -> BatchAdder
  |
  +-- [Update Operations]
  |   UpdateTodo(id)    -> TodoUpdater
  |   UpdateProject(id) -> ProjectUpdater
  |
  +-- [Show Operations]
  |   Show(ctx, uuid), ShowList(ctx, list), ShowSearch(ctx, query)
  |   ShowBuilder() -> Navigator
  |
  +-- [Utilities]
      Get(ctx, uuid), Search(ctx, query), ChecklistItems(ctx, uuid)
      Token(ctx), Close()

Internal (Unexported)
==================================================
*db, *scheme, *taskQuery, *areaQuery, *tagQuery
*addTodoBuilder, *addProjectBuilder, *batchBuilder
*updateTodoBuilder, *updateProjectBuilder, *showBuilder
```

### Interface Definitions

#### Query Interfaces

```go
// TaskQuerier provides methods for querying tasks.
type TaskQuerier interface {
    // Filters
    WithUUID(uuid string) TaskQuerier
    WithUUIDs(uuids ...string) TaskQuerier
    Type() TaskTypeFilter
    Status() StatusFilter
    StartBucket() StartBucketFilter
    Trashed(trashed bool) TaskQuerier
    InProject(uuid string) TaskQuerier
    InArea(uuid string) TaskQuerier
    HasDeadline(has bool) TaskQuerier
    HasTag(tag string) TaskQuerier

    // Terminal operations
    All(ctx context.Context) ([]Task, error)
    First(ctx context.Context) (*Task, error)
    Count(ctx context.Context) (int, error)
}

// TaskTypeFilter provides task type filtering.
type TaskTypeFilter interface {
    Todo() TaskQuerier
    Project() TaskQuerier
    Heading() TaskQuerier
}

// StatusFilter provides status filtering.
type StatusFilter interface {
    Incomplete() TaskQuerier
    Completed() TaskQuerier
    Canceled() TaskQuerier
}

// StartBucketFilter provides start bucket filtering.
type StartBucketFilter interface {
    Inbox() TaskQuerier
    Anytime() TaskQuerier
    Someday() TaskQuerier
}

// AreaQuerier provides methods for querying areas.
type AreaQuerier interface {
    WithUUID(uuid string) AreaQuerier
    All(ctx context.Context) ([]Area, error)
    First(ctx context.Context) (*Area, error)
    Count(ctx context.Context) (int, error)
}

// TagQuerier provides methods for querying tags.
type TagQuerier interface {
    WithUUID(uuid string) TagQuerier
    WithTitle(title string) TagQuerier
    All(ctx context.Context) ([]Tag, error)
    First(ctx context.Context) (*Tag, error)
    Count(ctx context.Context) (int, error)
}
```

#### Add Interfaces

```go
// TodoAdder provides methods for creating a new to-do.
type TodoAdder interface {
    Title(title string) TodoAdder
    Notes(notes string) TodoAdder
    When(when string) TodoAdder
    Deadline(deadline time.Time) TodoAdder
    DeadlineString(deadline string) TodoAdder
    Tags(tags ...string) TodoAdder
    ChecklistItems(items ...string) TodoAdder
    ListID(listID string) TodoAdder
    List(list string) TodoAdder
    Heading(heading string) TodoAdder
    Completed(completed bool) TodoAdder
    Canceled(canceled bool) TodoAdder
    Reveal(reveal bool) TodoAdder

    // Terminal operations
    Build() (string, error)
    Execute(ctx context.Context) error
}

// ProjectAdder provides methods for creating a new project.
type ProjectAdder interface {
    Title(title string) ProjectAdder
    Notes(notes string) ProjectAdder
    When(when string) ProjectAdder
    Deadline(deadline time.Time) ProjectAdder
    DeadlineString(deadline string) ProjectAdder
    Tags(tags ...string) ProjectAdder
    AreaID(areaID string) ProjectAdder
    Area(area string) ProjectAdder
    Todos(todos ...string) ProjectAdder
    Completed(completed bool) ProjectAdder
    Canceled(canceled bool) ProjectAdder
    Reveal(reveal bool) ProjectAdder

    // Terminal operations
    Build() (string, error)
    Execute(ctx context.Context) error
}

// BatchAdder provides methods for batch create operations.
type BatchAdder interface {
    AddTodo(fn func(TodoItemBuilder)) BatchAdder
    AddProject(fn func(ProjectItemBuilder)) BatchAdder
    Reveal(reveal bool) BatchAdder

    // Terminal operations
    Build() (string, error)
    Execute(ctx context.Context) error
}

// TodoItemBuilder provides methods for building a to-do in batch operations.
type TodoItemBuilder interface {
    Title(title string) TodoItemBuilder
    Notes(notes string) TodoItemBuilder
    When(when string) TodoItemBuilder
    Deadline(deadline time.Time) TodoItemBuilder
    Tags(tags ...string) TodoItemBuilder
    ChecklistItems(items ...string) TodoItemBuilder
    Completed(completed bool) TodoItemBuilder
    Canceled(canceled bool) TodoItemBuilder
}

// ProjectItemBuilder provides methods for building a project in batch operations.
type ProjectItemBuilder interface {
    Title(title string) ProjectItemBuilder
    Notes(notes string) ProjectItemBuilder
    When(when string) ProjectItemBuilder
    Deadline(deadline time.Time) ProjectItemBuilder
    Tags(tags ...string) ProjectItemBuilder
    AreaID(areaID string) ProjectItemBuilder
    Area(area string) ProjectItemBuilder
    Items(items ...ProjectItem) ProjectItemBuilder
    Completed(completed bool) ProjectItemBuilder
    Canceled(canceled bool) ProjectItemBuilder
}
```

#### Update Interfaces

```go
// TodoUpdater provides methods for updating an existing to-do.
type TodoUpdater interface {
    Title(title string) TodoUpdater
    Notes(notes string) TodoUpdater
    Prepend(notes string) TodoUpdater
    Append(notes string) TodoUpdater
    When(when string) TodoUpdater
    Deadline(deadline time.Time) TodoUpdater
    DeadlineString(deadline string) TodoUpdater
    Tags(tags ...string) TodoUpdater
    AddTags(tags ...string) TodoUpdater
    ChecklistItems(items ...string) TodoUpdater
    AddChecklistItems(items ...string) TodoUpdater
    ListID(listID string) TodoUpdater
    List(list string) TodoUpdater
    Heading(heading string) TodoUpdater
    Completed(completed bool) TodoUpdater
    Canceled(canceled bool) TodoUpdater
    Reveal(reveal bool) TodoUpdater
    DuplicateID() TodoUpdater

    // Terminal operations
    Build() (string, error)
    Execute(ctx context.Context) error
}

// ProjectUpdater provides methods for updating an existing project.
type ProjectUpdater interface {
    Title(title string) ProjectUpdater
    Notes(notes string) ProjectUpdater
    Prepend(notes string) ProjectUpdater
    Append(notes string) ProjectUpdater
    When(when string) ProjectUpdater
    Deadline(deadline time.Time) ProjectUpdater
    DeadlineString(deadline string) ProjectUpdater
    Tags(tags ...string) ProjectUpdater
    AddTags(tags ...string) ProjectUpdater
    AreaID(areaID string) ProjectUpdater
    Area(area string) ProjectUpdater
    Completed(completed bool) ProjectUpdater
    Canceled(canceled bool) ProjectUpdater
    Reveal(reveal bool) ProjectUpdater

    // Terminal operations
    Build() (string, error)
    Execute(ctx context.Context) error
}
```

#### Navigation Interface

```go
// Navigator provides methods for navigating to items or lists in Things.
type Navigator interface {
    ID(uuid string) Navigator
    List(list ListID) Navigator
    Query(query string) Navigator
    FilterByTag(tags ...string) Navigator

    // Terminal operations
    Build() string
    Execute(ctx context.Context) error
}
```

### Client API

```go
// Client provides unified access to Things 3.
// This is the only public entry point to the library.
type Client struct {
    // unexported fields only
    db     *db
    scheme *scheme
    // ...
}

// NewClient creates a new Things 3 client.
func NewClient(opts ...ClientOption) (*Client, error)

// Close closes the database connection.
func (c *Client) Close() error

// Token returns the authentication token for URL scheme operations.
func (c *Client) Token(ctx context.Context) (string, error)

// Query Operations - Convenience Methods
func (c *Client) Inbox(ctx context.Context) ([]Task, error)
func (c *Client) Today(ctx context.Context) ([]Task, error)
func (c *Client) Todos(ctx context.Context) ([]Task, error)
func (c *Client) Projects(ctx context.Context) ([]Task, error)
func (c *Client) Upcoming(ctx context.Context) ([]Task, error)
func (c *Client) Anytime(ctx context.Context) ([]Task, error)
func (c *Client) Someday(ctx context.Context) ([]Task, error)
func (c *Client) Logbook(ctx context.Context) ([]Task, error)
func (c *Client) Trash(ctx context.Context) ([]Task, error)
func (c *Client) Completed(ctx context.Context) ([]Task, error)
func (c *Client) Canceled(ctx context.Context) ([]Task, error)
func (c *Client) Deadlines(ctx context.Context) ([]Task, error)
func (c *Client) CreatedWithin(ctx context.Context, since time.Time) ([]Task, error)

// Query Operations - Builders (return interfaces)
func (c *Client) Tasks() TaskQuerier
func (c *Client) Areas() AreaQuerier
func (c *Client) Tags() TagQuerier

// Query Operations - Utilities
func (c *Client) Get(ctx context.Context, uuid string) (any, error)
func (c *Client) Search(ctx context.Context, query string) ([]Task, error)
func (c *Client) ChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error)

// Add Operations (return interfaces)
func (c *Client) AddTodo() TodoAdder
func (c *Client) AddProject() ProjectAdder
func (c *Client) Batch() BatchAdder

// Update Operations (return interfaces)
func (c *Client) UpdateTodo(id string) TodoUpdater
func (c *Client) UpdateProject(id string) ProjectUpdater

// Show Operations
func (c *Client) Show(ctx context.Context, uuid string) error
func (c *Client) ShowList(ctx context.Context, list ListID) error
func (c *Client) ShowSearch(ctx context.Context, query string) error
func (c *Client) ShowBuilder() Navigator
```

### Implementation Strategy

#### Step 1: Define Interfaces

Create `interfaces.go` with all interface definitions:

```go
// interfaces.go
package things3

// All interface definitions here
type TaskQuerier interface { ... }
type TodoAdder interface { ... }
// ...
```

#### Step 2: Rename Concrete Types

Rename exported types to unexported:

| Current (Exported) | New (Unexported) |
|-------------------|------------------|
| `DB` | `db` |
| `Scheme` | `scheme` |
| `TaskQuery` | `taskQuery` |
| `AreaQuery` | `areaQuery` |
| `TagQuery` | `tagQuery` |
| `AddTodoBuilder` | `addTodoBuilder` |
| `AddProjectBuilder` | `addProjectBuilder` |
| `UpdateTodoBuilder` | `updateTodoBuilder` |
| `UpdateProjectBuilder` | `updateProjectBuilder` |
| `BatchBuilder` | `batchBuilder` |
| `ShowBuilder` | `showBuilder` |

#### Step 3: Update Client Methods

Change return types from concrete to interface:

```go
// Before
func (c *Client) Tasks() *TaskQuery

// After
func (c *Client) Tasks() TaskQuerier
```

#### Step 4: Remove Public Constructors

Remove or unexport:
- `NewDB()` -> `newDB()` (internal only)
- `NewScheme()` -> `newScheme()` (internal only)

### File Organization

```
things3/
+-- interfaces.go       # All public interfaces (NEW)
+-- client.go           # Client type, NewClient(), all public methods
+-- client_options.go   # ClientOption (unchanged)
+-- db.go               # db type (unexported), newDB()
+-- db_options.go       # dbOptions (internal)
+-- scheme.go           # scheme type (unexported), newScheme()
+-- scheme_options.go   # schemeOptions (internal)
+-- query.go            # taskQuery implements TaskQuerier
+-- query_area.go       # areaQuery implements AreaQuerier
+-- query_tag.go        # tagQuery implements TagQuerier
+-- scheme_builder.go   # addTodoBuilder, addProjectBuilder
+-- scheme_update.go    # updateTodoBuilder, updateProjectBuilder
+-- scheme_show.go      # showBuilder implements Navigator
+-- scheme_json.go      # batchBuilder implements BatchAdder
+-- models.go           # Task, Area, Tag, ChecklistItem (unchanged)
+-- types.go            # TaskType, Status, ListID, etc. (unchanged)
```

## Usage Examples

### Basic Usage

```go
client, err := things3.NewClient()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Query with interface - IDE shows only TaskQuerier methods
tasks, _ := client.Tasks().
    Status().Incomplete().
    HasDeadline(true).
    All(ctx)

// Add with interface - IDE shows only TodoAdder methods
client.AddTodo().
    Title("Buy milk").
    When("today").
    Execute(ctx)

// Update with interface - IDE shows only TodoUpdater methods
client.UpdateTodo(uuid).
    Completed(true).
    Execute(ctx)
```

### IDE Experience

Before (concrete types):
```
client.Tasks().  // Shows: db, printSQL, conditions, joins, orderBy, ...
```

After (interfaces):
```
client.Tasks().  // Shows: WithUUID, Type, Status, All, First, Count
```

## Design Principles

| Principle | Implementation |
|-----------|----------------|
| Single Entry Point | Only `NewClient()` is public |
| Interface Segregation | Small, focused interfaces for each operation type |
| Information Hiding | Concrete types unexported, only interfaces visible |
| Compile-Time Safety | Interface methods enforce valid operation chains |
| Clean IDE Experience | Autocomplete shows only relevant methods |

## References

- RFC 005: Unified Client API - Current client design
- Issue #27: Interface abstraction for cleaner public API
