# RFC 003: Database API

Status: Implemented
Author: @moond4rk

## Summary

This RFC defines the Database API layer of the things3 Go library. The internal `db` type provides read-only access to the Things 3 SQLite database with a clean, type-safe interface following Go idioms. Users access database operations through the unified `NewClient()` entry point (see RFC 005).

## Design

### Entry Point

The `db` type is internal (unexported). Users access it through `NewClient()`:

```go
// db provides read-only access to the Things 3 database (internal).
type db struct {
    // internal fields
}

// newDB creates a new database connection (internal).
func newDB(opts ...dbOption) (*db, error)

// dbOption configures the database connection (internal).
type dbOption func(*dbConfig)
```

Users configure database behavior via `ClientOption`:

```go
things3.WithDatabasePath(path string)  // Custom database path
things3.WithPrintSQL(enabled bool)     // SQL logging for debugging
```

### Convenience Query Methods

Direct access methods for common queries. All methods accept `context.Context` as the first parameter and return value slices (not pointer slices). These are exposed on `Client`:

```go
client.Inbox(ctx)                      // Tasks in Inbox
client.Today(ctx)                      // Today's tasks
client.Upcoming(ctx)                   // Scheduled future tasks
client.Anytime(ctx)                    // Anytime tasks
client.Someday(ctx)                    // Someday tasks
client.Logbook(ctx)                    // Completed/canceled tasks
client.Trash(ctx)                      // Trashed tasks
client.Todos(ctx)                      // All incomplete to-dos
client.Projects(ctx)                   // All incomplete projects
client.Completed(ctx)                  // All completed tasks
client.Canceled(ctx)                   // All canceled tasks
client.Deadlines(ctx)                  // Tasks with deadlines
client.CreatedWithin(ctx, since)       // Tasks created after time
client.Search(ctx, query)             // Search tasks
client.Get(ctx, uuid)                 // Get task, area, or tag by UUID
client.ChecklistItems(ctx, taskUUID)  // Checklist items for a task
```

### Query Builders

Builder pattern for complex queries with chainable filter methods and terminal execution methods. All query builders return interfaces, not concrete types (see RFC 006).

```go
// Query builder entry points (on Client)
client.Tasks() -> TaskQueryBuilder
client.Areas() -> AreaQueryBuilder
client.Tags()  -> TagQueryBuilder
```

#### TaskQueryBuilder

Composed of: `TaskQueryExecutor` + `TaskRelationFilter` + `TaskStateFilter` + `TaskTimeFilter`

```go
// Filter methods (chainable, return TaskQueryBuilder interface)
WithUUID(uuid string) TaskQueryBuilder
WithUUIDPrefix(prefix string) TaskQueryBuilder
InTag(title string) TaskQueryBuilder
InProject(uuid string) TaskQueryBuilder
InArea(uuid string) TaskQueryBuilder
InHeading(uuid string) TaskQueryBuilder
HasArea(has bool) TaskQueryBuilder
HasProject(has bool) TaskQueryBuilder
HasHeading(has bool) TaskQueryBuilder
HasTag(has bool) TaskQueryBuilder
Trashed(trashed bool) TaskQueryBuilder
ContextTrashed(trashed bool) TaskQueryBuilder
WithDeadlineSuppressed(suppressed bool) TaskQueryBuilder
IncludeItems(include bool) TaskQueryBuilder
OrderByTodayIndex() TaskQueryBuilder
CreatedAfter(t time.Time) TaskQueryBuilder
Search(query string) TaskQueryBuilder

// Type-safe sub-builders (return sub-builder interfaces)
Type() TypeFilterBuilder      // Todo(), Project(), Heading()
Status() StatusFilterBuilder  // Incomplete(), Completed(), Canceled(), Any()
Start() StartFilterBuilder    // Inbox(), Anytime(), Someday()
StartDate() DateFilterBuilder // Exists(), Future(), Past(), On(), Before(), etc.
StopDate() DateFilterBuilder
Deadline() DateFilterBuilder

// Terminal methods (execute query)
All(ctx context.Context) ([]Task, error)
First(ctx context.Context) (*Task, error)
Count(ctx context.Context) (int, error)
```

#### AreaQueryBuilder

```go
// Filter methods (chainable, return AreaQueryBuilder interface)
WithUUID(uuid string) AreaQueryBuilder
WithTitle(title string) AreaQueryBuilder
Visible(visible bool) AreaQueryBuilder
InTag(title string) AreaQueryBuilder   // Filter by specific tag title
HasTag(has bool) AreaQueryBuilder      // Filter by whether area has any tags
IncludeItems(include bool) AreaQueryBuilder

// Terminal methods
All(ctx context.Context) ([]Area, error)
First(ctx context.Context) (*Area, error)
Count(ctx context.Context) (int, error)
```

#### TagQueryBuilder

```go
// Filter methods (chainable, return TagQueryBuilder interface)
WithUUID(uuid string) TagQueryBuilder
WithTitle(title string) TagQueryBuilder
WithParent(parentUUID string) TagQueryBuilder
IncludeItems(include bool) TagQueryBuilder

// Terminal methods
All(ctx context.Context) ([]Tag, error)
First(ctx context.Context) (*Tag, error)
```

### Token Retrieval

The internal `db.Token()` method retrieves the authentication token required for URL Scheme update operations. This bridges the database layer with the URL Scheme layer. Token management is handled automatically by the `Client` (see RFC 005).

## Usage Examples

### Basic Usage

```go
client, err := things3.NewClient()
if err != nil {
    return err
}
defer client.Close()

// Convenience methods
tasks, _ := client.Inbox(ctx)
tasks, _ := client.Today(ctx)
tasks, _ := client.Upcoming(ctx)
tasks, _ := client.Todos(ctx)
tasks, _ := client.Projects(ctx)
```

### Query Builder with Type-Safe Sub-Builders

```go
// Find all incomplete todos with a specific tag
tasks, _ := client.Tasks().
    Type().Todo().
    Status().Incomplete().
    InTag("work").
    All(ctx)

// Find a specific task by UUID
task, _ := client.Tasks().
    WithUUID("task-uuid").
    First(ctx)

// Find all tasks in a project
tasks, _ := client.Tasks().
    InProject("project-uuid").
    All(ctx)

// Find tasks with deadlines in the past
tasks, _ := client.Tasks().
    Deadline().Past().
    All(ctx)

// Find tasks starting in the future
tasks, _ := client.Tasks().
    StartDate().Future().
    All(ctx)

// Count tasks with a deadline
count, _ := client.Tasks().
    Deadline().Exists(true).
    Count(ctx)

// Search for tasks
tasks, _ := client.Tasks().
    Search("meeting").
    All(ctx)
```

### With URL Scheme Integration

Token management is automatic via the unified Client:

```go
client, _ := things3.NewClient()
defer client.Close()

// Update operations auto-fetch token from database
client.UpdateTodo("uuid").Completed(true).Execute(ctx)
```

## Error Definitions

```go
var (
    // ErrDatabaseNotFound is returned when the Things database cannot be found.
    ErrDatabaseNotFound = errors.New("things3: database not found")

    // ErrDatabaseVersionTooOld is returned when the database version is not supported.
    ErrDatabaseVersionTooOld = errors.New("things3: database version too old")

    // ErrAuthTokenNotFound is returned when the auth token is not configured.
    ErrAuthTokenNotFound = errors.New("things3: auth token not found")

    // ErrInvalidParameter is returned when a parameter is invalid.
    ErrInvalidParameter = errors.New("things3: invalid parameter")
)
```

## File Organization

| File | Responsibility |
|------|----------------|
| `client.go` | Client type, NewClient(), unified API entry point |
| `client_options.go` | ClientOption functional options |
| `interfaces.go` | All public interface definitions (6 layers) |
| `db.go` | Internal db type, database operations |
| `db_options.go` | Internal dbOption functional options |
| `convenience.go` | Inbox(), Today(), Todos(), Projects(), etc. |
| `query.go` | taskQuery builder with filter methods |
| `query_filter.go` | typeFilter, statusFilter, startFilter, dateFilter |
| `query_area.go` | areaQuery builder |
| `query_tag.go` | tagQuery builder |
| `sql.go` | SQL query building and execution |
| `models.go` | Task, Area, Tag, ChecklistItem structs |
| `types.go` | TaskType, Status, StartBucket enums |
| `date.go` | Things date format conversion |
| `database.go` | Database connection and path discovery |
| `errors.go` | Error definitions |
| `constants.go` | Table names, column mappings |

## Testing Strategy

### Integration Tests

```go
func TestClient_Inbox(t *testing.T) {
    client, err := things3.NewClient(
        things3.WithDatabasePath("testdata/test.sqlite"),
    )
    require.NoError(t, err)
    defer client.Close()

    tasks, err := client.Inbox(context.Background())
    require.NoError(t, err)
    assert.NotEmpty(t, tasks)
}
```

### Query Builder Tests

```go
func TestTaskQuery_TypeSafeFilters(t *testing.T) {
    client, _ := things3.NewClient(
        things3.WithDatabasePath("testdata/test.sqlite"),
    )
    defer client.Close()

    tests := []struct {
        name    string
        query   func() things3.TaskQueryBuilder
        wantMin int
    }{
        {
            name:    "all todos",
            query:   func() things3.TaskQueryBuilder { return client.Tasks().Type().Todo() },
            wantMin: 1,
        },
        {
            name:    "incomplete tasks",
            query:   func() things3.TaskQueryBuilder { return client.Tasks().Status().Incomplete() },
            wantMin: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tasks, err := tt.query().All(context.Background())
            require.NoError(t, err)
            assert.GreaterOrEqual(t, len(tasks), tt.wantMin)
        })
    }
}
```

## Design Principles

| Principle | Implementation |
|-----------|----------------|
| Context-First | All query methods accept `context.Context` as first parameter |
| Builder Pattern | Chainable filter methods with terminal execution methods |
| Interface-Based | Query builders return interfaces, not concrete types (see RFC 006) |
| Type-Safe Sub-Builders | StatusFilterBuilder, TypeFilterBuilder, DateFilterBuilder provide compile-time safety |
| Read-Only | No write operations to the database |
| Type Safety | Strongly typed query results, no raw SQL exposure |
| Functional Options | Configuration via `ClientOption` functions |
| Value Semantics | Return value slices (`[]Task`) not pointer slices (`[]*Task`) |

## References

- RFC 002: Database Schema - Schema definitions used by this API
- RFC 004: URL Scheme - Uses `Token()` for authenticated operations
- [things.py](https://github.com/thingsapi/things.py) - Python reference implementation
