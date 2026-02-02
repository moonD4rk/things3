# RFC 006: Interface Abstraction

Status: Implemented
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
  |   Tasks()   -> TaskQueryBuilder
  |   Areas()   -> AreaQueryBuilder
  |   Tags()    -> TagQueryBuilder
  |
  +-- [Add Operations]
  |   AddTodo()    -> TodoAdder
  |   AddProject() -> ProjectAdder
  |   Batch()      -> BatchCreator
  |
  +-- [Update Operations]
  |   UpdateTodo(id)    -> TodoUpdater
  |   UpdateProject(id) -> ProjectUpdater
  |
  +-- [Show Operations]
  |   Show(ctx, uuid), ShowList(ctx, list), ShowSearch(ctx, query)
  |   ShowBuilder() -> ShowNavigator
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

The interface design follows Go's philosophy of "small interfaces, composition over inheritance". See `interfaces.go` for the complete 6-layer interface hierarchy:

1. **Layer 1 - Reusable Small Interfaces**: Terminal operations (`TaskQueryExecutor`, `AreaQueryExecutor`, `TagQueryExecutor`, `URLBuilder`)
2. **Layer 2 - Sub-builder Interfaces**: Filter builders (`TypeFilterBuilder`, `StatusFilterBuilder`, `StartFilterBuilder`, `DateFilterBuilder`)
3. **Layer 3 - Functional Group Interfaces**: Composed filters (`TaskRelationFilter`, `TaskStateFilter`, `TaskTimeFilter`)
4. **Layer 4 - Composed Query Builders**: Full query interfaces (`TaskQueryBuilder`, `AreaQueryBuilder`, `TagQueryBuilder`)
5. **Layer 5 - URL Scheme Builders**: Create/update/show operations (`TodoAdder`, `ProjectAdder`, `TodoUpdater`, `ProjectUpdater`, `ShowNavigator`)
6. **Layer 6 - Batch Operations**: Batch create/update interfaces (`BatchCreator`, `AuthBatchCreator`, `BatchTodoConfigurator`, `BatchProjectConfigurator`)

#### Query Interfaces

```go
// TaskQueryExecutor executes task queries and returns results.
type TaskQueryExecutor interface {
    All(ctx context.Context) ([]Task, error)
    First(ctx context.Context) (*Task, error)
    Count(ctx context.Context) (int, error)
}

// TaskQueryBuilder provides a fluent interface for building task queries.
// Composed of: TaskQueryExecutor + TaskRelationFilter + TaskStateFilter + TaskTimeFilter
type TaskQueryBuilder interface {
    TaskQueryExecutor
    TaskRelationFilter
    TaskStateFilter
    TaskTimeFilter

    WithUUID(uuid string) TaskQueryBuilder
    WithDeadlineSuppressed(suppressed bool) TaskQueryBuilder
    Search(query string) TaskQueryBuilder
    OrderByTodayIndex() TaskQueryBuilder
    IncludeItems(include bool) TaskQueryBuilder
}
```

#### Add Interfaces

```go
// TodoAdder builds URLs for creating new to-dos.
type TodoAdder interface {
    URLBuilder

    Title(title string) TodoAdder
    Notes(notes string) TodoAdder
    When(t time.Time) TodoAdder
    WhenEvening() TodoAdder
    WhenAnytime() TodoAdder
    WhenSomeday() TodoAdder
    Deadline(t time.Time) TodoAdder
    Tags(tags ...string) TodoAdder
    // ... (see interfaces.go for complete definition)
}
```

#### Update Interfaces

```go
// TodoUpdater builds URLs for updating existing to-dos.
type TodoUpdater interface {
    URLBuilder

    Title(title string) TodoUpdater
    Notes(notes string) TodoUpdater
    PrependNotes(notes string) TodoUpdater
    AppendNotes(notes string) TodoUpdater
    // ... (see interfaces.go for complete definition)
}
```

#### Navigation Interface

```go
// ShowNavigator builds URLs for navigating to items or lists.
type ShowNavigator interface {
    ID(id string) ShowNavigator
    List(list ListID) ShowNavigator
    Query(query string) ShowNavigator
    Filter(tags ...string) ShowNavigator

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
    database *db
    scheme   *scheme
    tokenMu  sync.Mutex
    tokenCache string
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
func (c *Client) Tasks() TaskQueryBuilder
func (c *Client) Areas() AreaQueryBuilder
func (c *Client) Tags() TagQueryBuilder

// Query Operations - Utilities
func (c *Client) Get(ctx context.Context, uuid string) (any, error)
func (c *Client) Search(ctx context.Context, query string) ([]Task, error)
func (c *Client) ChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error)

// Add Operations (return interfaces)
func (c *Client) AddTodo() TodoAdder
func (c *Client) AddProject() ProjectAdder
func (c *Client) Batch() BatchCreator

// Update Operations (return interfaces)
func (c *Client) UpdateTodo(id string) TodoUpdater
func (c *Client) UpdateProject(id string) ProjectUpdater

// Show Operations
func (c *Client) Show(ctx context.Context, uuid string) error
func (c *Client) ShowList(ctx context.Context, list ListID) error
func (c *Client) ShowSearch(ctx context.Context, query string) error
func (c *Client) ShowBuilder() ShowNavigator
```

### Implementation Strategy

#### Step 1: Define Interfaces

Create `interfaces.go` with all interface definitions:

```go
// interfaces.go
package things3

// All interface definitions here (see interfaces.go for complete definitions)
type TaskQueryBuilder interface { ... }
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
func (c *Client) Tasks() TaskQueryBuilder
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
+-- query.go            # taskQuery implements TaskQueryBuilder
+-- query_area.go       # areaQuery implements AreaQueryBuilder
+-- query_tag.go        # tagQuery implements TagQueryBuilder
+-- scheme_builder.go   # addTodoBuilder, addProjectBuilder
+-- scheme_update.go    # updateTodoBuilder, updateProjectBuilder
+-- scheme_show.go      # showBuilder implements ShowNavigator
+-- scheme_json.go      # batchBuilder implements BatchCreator
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

// Query with interface - IDE shows only TaskQueryBuilder methods
tasks, _ := client.Tasks().
    Status().Incomplete().
    Deadline().Exists(true).
    All(ctx)

// Add with interface - IDE shows only TodoAdder methods
client.AddTodo().
    Title("Buy milk").
    When(things3.Today()).
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
client.Tasks().  // Shows: WithUUID, Type, Status, StartDate, Deadline, All, First, Count
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
