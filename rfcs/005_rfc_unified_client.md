# RFC 005: Unified Client API

Status: Draft
Author: @moond4rk

## Summary

This RFC defines the unified Client API that combines database read operations with URL scheme write operations, providing a single entry point for all Things 3 interactions. It also establishes naming conventions for improved API clarity and consistency.

## Motivation

### Current Issues

1. **Fragmented Entry Points**: Users must manage `NewDB()` and `NewScheme()` separately
2. **Token Management**: Manual token fetching and passing for update operations
3. **Inconsistent Naming**: `TodoBuilder` vs `AddTodoBuilder`, `Reveal` vs `Show`
4. **Unclear Intent**: `scheme.Todo()` doesn't clearly indicate "add" operation
5. **JSON vs Batch**: "JSON" exposes implementation detail instead of user intent

### Goals

- Single `NewClient()` entry point for all operations
- Automatic token management with lazy loading
- Consistent naming conventions across all APIs
- Clear method names that express intent
- Backward compatibility with existing `NewDB()` and `NewScheme()`

### Non-Goals

- Deprecating `NewDB()` or `NewScheme()` (they remain for advanced use cases)
- Changing the underlying URL scheme format

## Design

### Architecture Overview

```
NewClient(opts...)  -> *Client  (Unified entry point)
    |
    +-- Options: WithDatabasePath(), WithPrintSQL(),
    |            WithForegroundExecution(), WithBackgroundNavigation(),
    |            WithPreloadToken()
    |
    +-- [Query Operations - Read from DB]
    |   +-- Inbox(ctx), Today(ctx), Todos(ctx), Projects(ctx), ...
    |   +-- Tasks() -> *TaskQuery, Areas() -> *AreaQuery, Tags() -> *TagQuery
    |   +-- Get(ctx, uuid), Search(ctx, query), ChecklistItems(ctx, uuid)
    |
    +-- [Add Operations - URL Scheme, No Auth]
    |   +-- AddTodo()    -> *AddTodoBuilder    -> Build() | Execute(ctx)
    |   +-- AddProject() -> *AddProjectBuilder -> Build() | Execute(ctx)
    |   +-- Batch()      -> *BatchBuilder      -> Build() | Execute(ctx)
    |
    +-- [Update Operations - URL Scheme, Auto Auth]
    |   +-- UpdateTodo(id)    -> *UpdateTodoBuilder    -> Build() | Execute(ctx)
    |   +-- UpdateProject(id) -> *UpdateProjectBuilder -> Build() | Execute(ctx)
    |
    +-- [Show Operations - Navigation]
        +-- Show(ctx, uuid), ShowList(ctx, list), ShowSearch(ctx, query)
        +-- ShowBuilder() -> *ShowBuilder -> Build()
```

### Entry Point

```go
// Client provides unified access to Things 3 database and URL scheme operations.
type Client struct {
    db     *DB
    scheme *Scheme

    // Token management with lazy loading
    tokenOnce  sync.Once
    tokenCache string
    tokenErr   error
}

// NewClient creates a new unified Things 3 client.
func NewClient(opts ...ClientOption) (*Client, error)

// Close closes the database connection.
func (c *Client) Close() error

// Token returns the cached authentication token, fetching it if needed.
func (c *Client) Token(ctx context.Context) (string, error)
```

### Client Options

```go
type ClientOption func(*clientOptions)

// Database options
func WithDatabasePath(path string) ClientOption
func WithPrintSQL(enabled bool) ClientOption

// Scheme options
func WithForegroundExecution() ClientOption   // Bring Things to foreground for create/update
func WithBackgroundNavigation() ClientOption  // Keep Things in background for show operations

// Token options
func WithPreloadToken() ClientOption  // Fetch token immediately during NewClient()
```

## Naming Conventions

### Principle: Method Names Express Intent

All public methods should clearly communicate what action they perform:
- **Add**: Creating new items (`AddTodo`, `AddProject`)
- **Update**: Modifying existing items (`UpdateTodo`, `UpdateProject`)
- **Show**: Navigation/display operations (`Show`, `ShowList`, `ShowSearch`)
- **Batch**: Multiple operations at once (`Batch`)

### Naming Changes

#### 1. Add Operations - Types and Methods

| Current | New | Reason |
|---------|-----|--------|
| `TodoBuilder` | `AddTodoBuilder` | Clarifies "add" intent |
| `ProjectBuilder` | `AddProjectBuilder` | Clarifies "add" intent |
| `Scheme.Todo()` | `Scheme.AddTodo()` | Consistent with Client API |
| `Scheme.Project()` | `Scheme.AddProject()` | Consistent with Client API |

#### 2. Batch Operations - Types and Methods

| Current | New | Reason |
|---------|-----|--------|
| `JSONBuilder` | `BatchBuilder` | "JSON" is implementation detail |
| `JSONTodoBuilder` | `BatchTodoBuilder` | Consistency |
| `JSONProjectBuilder` | `BatchProjectBuilder` | Consistency |
| `AuthJSONBuilder` | `AuthBatchBuilder` | Consistency |
| `Scheme.JSON()` | `Scheme.Batch()` | Expresses intent |
| `AuthScheme.JSON()` | `AuthScheme.Batch()` | Consistency |
| `Client.AddJSON()` | `Client.Batch()` | Clearer naming |

#### 3. Show Operations - Unified Naming

Things URL Scheme uses `show` command, so we align with official terminology:

| Current | New | Reason |
|---------|-----|--------|
| `Client.Reveal()` | `Client.Show()` | Matches URL Scheme |
| `Client.RevealList()` | `Client.ShowList()` | Consistency |
| `Client.RevealSearch()` | `Client.ShowSearch()` | Consistency |
| `Client.RevealBuilder()` | `Client.ShowBuilder()` | Matches return type |

Note: `ShowBuilder` type name remains unchanged as it already matches.

### Complete Type Mapping

```go
// Add Operations
type AddTodoBuilder struct { ... }      // was: TodoBuilder
type AddProjectBuilder struct { ... }   // was: ProjectBuilder

// Batch Operations
type BatchBuilder struct { ... }        // was: JSONBuilder
type BatchTodoBuilder struct { ... }    // was: JSONTodoBuilder
type BatchProjectBuilder struct { ... } // was: JSONProjectBuilder
type AuthBatchBuilder struct { ... }    // was: AuthJSONBuilder

// Update Operations (unchanged - already clear)
type UpdateTodoBuilder struct { ... }
type UpdateProjectBuilder struct { ... }

// Show Operations (unchanged - already clear)
type ShowBuilder struct { ... }
```

### Complete Method Mapping

```go
// Scheme methods
func (s *Scheme) AddTodo() *AddTodoBuilder       // was: Todo()
func (s *Scheme) AddProject() *AddProjectBuilder // was: Project()
func (s *Scheme) Batch() *BatchBuilder           // was: JSON()
func (s *Scheme) ShowBuilder() *ShowBuilder      // unchanged

// AuthScheme methods
func (a *AuthScheme) Batch() *AuthBatchBuilder   // was: JSON()

// Client methods
func (c *Client) AddTodo() *AddTodoBuilder
func (c *Client) AddProject() *AddProjectBuilder
func (c *Client) Batch() *BatchBuilder           // was: AddJSON()
func (c *Client) Show(ctx, uuid) error           // was: Reveal()
func (c *Client) ShowList(ctx, list) error       // was: RevealList()
func (c *Client) ShowSearch(ctx, query) error    // was: RevealSearch()
func (c *Client) ShowBuilder() *ShowBuilder      // was: RevealBuilder()
```

## Client API

### Query Operations

```go
// Convenience methods - direct access to common queries
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

// Query builders - for complex queries
func (c *Client) Tasks() *TaskQuery
func (c *Client) Areas() *AreaQuery
func (c *Client) Tags() *TagQuery

// Utilities
func (c *Client) Get(ctx context.Context, uuid string) (any, error)
func (c *Client) Search(ctx context.Context, query string) ([]Task, error)
func (c *Client) ChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error)
```

### Add Operations

```go
// AddTodo returns an AddTodoBuilder for creating a new to-do.
func (c *Client) AddTodo() *AddTodoBuilder

// AddProject returns an AddProjectBuilder for creating a new project.
func (c *Client) AddProject() *AddProjectBuilder

// Batch returns a BatchBuilder for batch create operations.
func (c *Client) Batch() *BatchBuilder
```

### Update Operations

```go
// UpdateTodo returns an UpdateTodoBuilder for modifying an existing to-do.
// The authentication token is fetched automatically on first use.
func (c *Client) UpdateTodo(id string) *UpdateTodoBuilder

// UpdateProject returns an UpdateProjectBuilder for modifying an existing project.
// The authentication token is fetched automatically on first use.
func (c *Client) UpdateProject(id string) *UpdateProjectBuilder
```

### Show Operations

```go
// Show opens Things and shows the item with the given UUID.
func (c *Client) Show(ctx context.Context, uuid string) error

// ShowList opens Things and shows the specified list.
func (c *Client) ShowList(ctx context.Context, list ListID) error

// ShowSearch opens Things and performs a search for the given query.
func (c *Client) ShowSearch(ctx context.Context, query string) error

// ShowBuilder returns a ShowBuilder for complex navigation operations.
func (c *Client) ShowBuilder() *ShowBuilder
```

## Token Management

The Client automatically manages authentication tokens for update operations:

```go
// Internal token management
func (c *Client) ensureToken(ctx context.Context) (string, error) {
    c.tokenOnce.Do(func() {
        c.tokenCache, c.tokenErr = c.db.Token(ctx)
    })
    return c.tokenCache, c.tokenErr
}
```

UpdateTodoBuilder and UpdateProjectBuilder use a `tokenFunc` callback for lazy loading:

```go
type UpdateTodoBuilder struct {
    scheme    *Scheme
    token     string
    tokenFunc func(context.Context) (string, error)  // Lazy token loader
    id        string
    attrs     urlAttrs
    err       error
}

func (b *UpdateTodoBuilder) Execute(ctx context.Context) error {
    // Lazy load token if needed
    if b.token == "" && b.tokenFunc != nil {
        token, err := b.tokenFunc(ctx)
        if err != nil {
            return err
        }
        b.token = token
    }
    uri, err := b.Build()
    if err != nil {
        return err
    }
    return b.scheme.execute(ctx, uri)
}
```

## Usage Examples

### Basic Usage

```go
client, err := things3.NewClient()
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Query operations
tasks, _ := client.Inbox(ctx)
tasks, _ := client.Tasks().Status().Incomplete().All(ctx)

// Add operations
client.AddTodo().
    Title("Buy milk").
    When(things3.Today()).
    Execute(ctx)

// Update operations (token managed automatically)
client.UpdateTodo(uuid).
    Completed(true).
    Execute(ctx)

// Show operations
client.Show(ctx, uuid)
client.ShowList(ctx, things3.ListToday)
```

### Batch Operations

```go
client.Batch().
    AddTodo(func(b *things3.BatchTodoBuilder) {
        b.Title("Task 1").When(things3.Today())
    }).
    AddTodo(func(b *things3.BatchTodoBuilder) {
        b.Title("Task 2").When(things3.Tomorrow())
    }).
    Execute(ctx)
```

### Using Scheme Directly

For advanced use cases, `NewScheme()` and `NewDB()` remain available:

```go
// URL building only (no database)
scheme := things3.NewScheme()
url, _ := scheme.AddTodo().Title("Task").Build()

// Database access only (no URL scheme)
db, _ := things3.NewDB()
tasks, _ := db.Inbox(ctx)
```

## File Organization

```text
things3/
+-- client.go           # Client type, NewClient(), query/add/update/show methods
+-- client_options.go   # ClientOption, With*() functions
+-- db.go               # DB type, NewDB() (unchanged)
+-- db_options.go       # DBOption (unchanged)
+-- scheme.go           # Scheme type, NewScheme(), WithToken() (method renames)
+-- scheme_options.go   # SchemeOption (unchanged)
+-- scheme_builder.go   # AddTodoBuilder, AddProjectBuilder (renamed)
+-- scheme_update.go    # UpdateTodoBuilder, UpdateProjectBuilder (tokenFunc added)
+-- scheme_show.go      # ShowBuilder (unchanged)
+-- scheme_json.go      # BatchBuilder, BatchTodoBuilder, etc. (renamed)
```

## Migration Guide

### For Scheme Users

```go
// Before
scheme.Todo().Title("Task").Build()
scheme.Project().Title("Project").Build()
scheme.JSON().AddTodo(...).Build()

// After
scheme.AddTodo().Title("Task").Build()
scheme.AddProject().Title("Project").Build()
scheme.Batch().AddTodo(...).Build()
```

### For Direct Type Usage

```go
// Before
var builder *things3.TodoBuilder
var jsonBuilder *things3.JSONBuilder

// After
var builder *things3.AddTodoBuilder
var batchBuilder *things3.BatchBuilder
```

## Design Principles

| Principle | Implementation |
|-----------|----------------|
| Single Entry Point | `NewClient()` for most use cases |
| Automatic Token Management | Lazy loading with `sync.Once` |
| Clear Intent | Method names express action (Add, Update, Show, Batch) |
| Consistency | Same naming pattern across Client and Scheme |
| Separation of Concerns | Client composes DB and Scheme, doesn't replace them |

## References

- RFC 003: Database API - Provides `Token()` for authenticated operations
- RFC 004: URL Scheme - Defines URL building and execution patterns
- [Things URL Scheme Documentation](https://culturedcode.com/things/support/articles/2803573/)
