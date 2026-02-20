# RFC 005: Unified Client API

Status: Implemented
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

### Non-Goals

- Changing the underlying URL scheme format

## Design

### Architecture Overview

```
NewClient(opts...)  -> *Client  (Unified entry point)
    |
    +-- Options: WithDatabasePath(), WithPrintSQL(),
    |            WithForeground(), WithBackground(),
    |            WithPreloadToken()
    |
    +-- [Query Operations - Read from DB]
    |   +-- Convenience: Inbox(), Today(), Upcoming(), Todos(), Projects(), ...
    |   +-- Builders: Tasks(), Areas(), Tags()
    |   +-- Utilities: Get(), Search(), ChecklistItems()
    |
    +-- [Add Operations - URL Scheme, No Auth]
    |   +-- AddTodo()    -> TodoAdder       -> Build() | Execute(ctx)
    |   +-- AddProject() -> ProjectAdder    -> Build() | Execute(ctx)
    |   +-- Batch()      -> BatchCreator    -> Build() | Execute(ctx)
    |
    +-- [Update Operations - URL Scheme, Auto Auth]
    |   +-- UpdateTodo(id)    -> TodoUpdater    -> Build() | Execute(ctx)
    |   +-- UpdateProject(id) -> ProjectUpdater -> Build() | Execute(ctx)
    |   +-- AuthBatch()       -> AuthBatchCreator -> Build() | Execute(ctx)
    |
    +-- [Show Operations - Navigation]
        +-- Show(ctx, uuid), ShowList(ctx, list), ShowSearch(ctx, query)
        +-- ShowBuilder() -> ShowNavigator -> Build() | Execute(ctx)
```

### Entry Point

```go
// Client provides unified access to Things 3 database and URL scheme operations.
type Client struct {
    database *db
    scheme   *scheme

    // Token management with mutex (allows retry on transient failures)
    tokenMu    sync.Mutex
    tokenCache string
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
func WithForeground() ClientOption  // Bring Things to foreground for create/update
func WithBackground() ClientOption  // Keep Things in background for show operations

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

// Query builders - for complex queries (return interfaces, see RFC 006)
func (c *Client) Tasks() TaskQueryBuilder
func (c *Client) Areas() AreaQueryBuilder
func (c *Client) Tags() TagQueryBuilder

// Utilities
func (c *Client) Get(ctx context.Context, uuid string) (any, error)
func (c *Client) Search(ctx context.Context, query string) ([]Task, error)
func (c *Client) ChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error)
```

### Add Operations

```go
// AddTodo returns a TodoAdder for creating a new to-do.
func (c *Client) AddTodo() TodoAdder

// AddProject returns a ProjectAdder for creating a new project.
func (c *Client) AddProject() ProjectAdder

// Batch returns a BatchCreator for batch create operations.
func (c *Client) Batch() BatchCreator
```

### Update Operations

```go
// UpdateTodo returns a TodoUpdater for modifying an existing to-do.
// The authentication token is fetched automatically on first use.
func (c *Client) UpdateTodo(id string) TodoUpdater

// UpdateProject returns a ProjectUpdater for modifying an existing project.
// The authentication token is fetched automatically on first use.
func (c *Client) UpdateProject(id string) ProjectUpdater
```

### Show Operations

```go
// Show opens Things and shows the item with the given UUID.
func (c *Client) Show(ctx context.Context, uuid string) error

// ShowList opens Things and shows the specified list.
func (c *Client) ShowList(ctx context.Context, list ListID) error

// ShowSearch opens Things and performs a search for the given query.
func (c *Client) ShowSearch(ctx context.Context, query string) error

// ShowBuilder returns a ShowNavigator for complex navigation operations.
func (c *Client) ShowBuilder() ShowNavigator
```

## Token Management

The Client automatically manages authentication tokens for update operations using `sync.Mutex` (not `sync.Once`, to allow retry on transient failures):

```go
// Internal token management
func (c *Client) ensureToken(ctx context.Context) (string, error) {
    c.tokenMu.Lock()
    defer c.tokenMu.Unlock()
    if c.tokenCache != "" {
        return c.tokenCache, nil
    }
    token, err := c.database.Token(ctx)
    if err != nil {
        return "", err
    }
    c.tokenCache = token
    return token, nil
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

## File Organization

```text
things3/
+-- client.go           # Client type, NewClient(), query/add/update/show methods
+-- client_options.go   # ClientOption, With*() functions
+-- db.go               # Internal db type, newDB(), database operations
+-- db_options.go       # Internal dbOption
+-- scheme.go           # Internal scheme type, URL building and execution
+-- scheme_options.go   # Internal schemeOption
+-- scheme_builder.go   # addTodoBuilder, addProjectBuilder
+-- scheme_update.go    # updateTodoBuilder, updateProjectBuilder (with tokenFunc)
+-- scheme_show.go      # showBuilder
+-- scheme_json.go      # batchBuilder, authBatchBuilder
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
