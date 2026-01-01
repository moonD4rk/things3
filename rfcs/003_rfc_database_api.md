# RFC 003: Database API

Status: Implemented
Author: @moond4rk

## Summary

This RFC defines the Database API layer of the things3 Go library. The `NewDB()` entry point provides read-only access to the Things 3 SQLite database with a clean, type-safe interface following Go idioms.

## Design

### Entry Point

```go
// DB provides read-only access to the Things 3 database.
type DB struct {
    // internal fields
}

// NewDB creates a new database connection.
func NewDB(opts ...DBOption) (*DB, error)

// DBOption configures the database connection.
type DBOption func(*dbConfig)

func WithDBPath(path string) DBOption
func WithPrintSQL(enabled bool) DBOption
```

### Connection Methods

```go
// Close closes the database connection.
func (db *DB) Close() error

// Filepath returns the path to the database file.
func (db *DB) Filepath() string
```

### Convenience Query Methods

Direct access methods for common queries. All methods accept `context.Context` as the first parameter and return value slices (not pointer slices).

```go
// Inbox returns all tasks in the Inbox.
func (db *DB) Inbox(ctx context.Context) ([]Task, error)

// Today returns all tasks scheduled for today.
func (db *DB) Today(ctx context.Context) ([]Task, error)

// Upcoming returns all tasks with future start dates.
func (db *DB) Upcoming(ctx context.Context) ([]Task, error)

// Anytime returns all tasks in the Anytime list.
func (db *DB) Anytime(ctx context.Context) ([]Task, error)

// Someday returns all tasks in the Someday list.
func (db *DB) Someday(ctx context.Context) ([]Task, error)

// Logbook returns all completed tasks.
func (db *DB) Logbook(ctx context.Context) ([]Task, error)

// Trash returns all trashed tasks.
func (db *DB) Trash(ctx context.Context) ([]Task, error)

// Todos returns all incomplete to-do items.
func (db *DB) Todos(ctx context.Context) ([]Task, error)

// Projects returns all projects.
func (db *DB) Projects(ctx context.Context) ([]Task, error)

// Completed returns all completed tasks.
func (db *DB) Completed(ctx context.Context) ([]Task, error)

// Canceled returns all canceled tasks.
func (db *DB) Canceled(ctx context.Context) ([]Task, error)

// Deadlines returns all tasks with deadlines.
func (db *DB) Deadlines(ctx context.Context) ([]Task, error)

// CreatedWithin returns tasks created after the specified time.
func (db *DB) CreatedWithin(ctx context.Context, since time.Time) ([]Task, error)

// Search returns tasks matching the search query.
func (db *DB) Search(ctx context.Context, query string) ([]Task, error)

// Get returns a task, area, or tag by UUID.
func (db *DB) Get(ctx context.Context, uuid string) (any, error)

// ChecklistItems returns checklist items for a task.
func (db *DB) ChecklistItems(ctx context.Context, taskUUID string) ([]ChecklistItem, error)
```

### Query Builders

Builder pattern for complex queries with chainable filter methods and terminal execution methods.

```go
// Query builder entry points
func (db *DB) Tasks() *TaskQuery
func (db *DB) Areas() *AreaQuery
func (db *DB) Tags() *TagQuery
```

#### TaskQuery

```go
type TaskQuery struct {
    // internal fields
}

// Filter methods (chainable)
func (q *TaskQuery) WithUUID(uuid string) *TaskQuery
func (q *TaskQuery) InTag(title string) *TaskQuery
func (q *TaskQuery) InProject(projectUUID string) *TaskQuery
func (q *TaskQuery) InArea(areaUUID string) *TaskQuery
func (q *TaskQuery) InHeading(headingUUID string) *TaskQuery
func (q *TaskQuery) Trashed(trashed bool) *TaskQuery
func (q *TaskQuery) HasArea(has bool) *TaskQuery
func (q *TaskQuery) HasProject(has bool) *TaskQuery
func (q *TaskQuery) HasHeading(has bool) *TaskQuery
func (q *TaskQuery) HasTag(has bool) *TaskQuery
func (q *TaskQuery) WithDeadlineSuppressed(suppressed bool) *TaskQuery
func (q *TaskQuery) ContextTrashed(trashed bool) *TaskQuery
func (q *TaskQuery) IncludeItems(include bool) *TaskQuery
func (q *TaskQuery) OrderByTodayIndex() *TaskQuery
func (q *TaskQuery) CreatedAfter(t time.Time) *TaskQuery
func (q *TaskQuery) Search(query string) *TaskQuery

// Type-safe sub-builders (chainable, return *TaskQuery)
func (q *TaskQuery) Type() *TypeFilter
func (q *TaskQuery) Status() *StatusFilter
func (q *TaskQuery) Start() *StartFilter
func (q *TaskQuery) StartDate() *DateFilter
func (q *TaskQuery) StopDate() *DateFilter
func (q *TaskQuery) Deadline() *DateFilter

// Terminal methods (execute query)
func (q *TaskQuery) All(ctx context.Context) ([]Task, error)
func (q *TaskQuery) First(ctx context.Context) (*Task, error)
func (q *TaskQuery) Count(ctx context.Context) (int, error)
```

#### Type-Safe Sub-Builders

These sub-builders provide compile-time type safety for filter values:

```go
// TypeFilter for task type filtering
type TypeFilter struct { /* internal */ }
func (f *TypeFilter) Todo() *TaskQuery
func (f *TypeFilter) Project() *TaskQuery
func (f *TypeFilter) Heading() *TaskQuery

// StatusFilter for task status filtering
type StatusFilter struct { /* internal */ }
func (f *StatusFilter) Incomplete() *TaskQuery
func (f *StatusFilter) Completed() *TaskQuery
func (f *StatusFilter) Canceled() *TaskQuery
func (f *StatusFilter) Any() *TaskQuery

// StartFilter for start bucket filtering
type StartFilter struct { /* internal */ }
func (f *StartFilter) Inbox() *TaskQuery
func (f *StartFilter) Anytime() *TaskQuery
func (f *StartFilter) Someday() *TaskQuery

// DateFilter for date-based filtering
type DateFilter struct { /* internal */ }
func (f *DateFilter) Exists(exists bool) *TaskQuery
func (f *DateFilter) Future() *TaskQuery
func (f *DateFilter) Past() *TaskQuery
func (f *DateFilter) On(date time.Time) *TaskQuery
func (f *DateFilter) Before(date time.Time) *TaskQuery
func (f *DateFilter) OnOrBefore(date time.Time) *TaskQuery
func (f *DateFilter) After(date time.Time) *TaskQuery
func (f *DateFilter) OnOrAfter(date time.Time) *TaskQuery
```

#### AreaQuery

```go
type AreaQuery struct {
    // internal fields
}

// Filter methods (chainable)
func (q *AreaQuery) WithUUID(uuid string) *AreaQuery
func (q *AreaQuery) WithTitle(title string) *AreaQuery
func (q *AreaQuery) Visible(visible bool) *AreaQuery
func (q *AreaQuery) InTag(tag any) *AreaQuery
func (q *AreaQuery) IncludeItems(include bool) *AreaQuery

// Terminal methods (execute query)
func (q *AreaQuery) All(ctx context.Context) ([]Area, error)
func (q *AreaQuery) First(ctx context.Context) (*Area, error)
func (q *AreaQuery) Count(ctx context.Context) (int, error)
```

#### TagQuery

```go
type TagQuery struct {
    // internal fields
}

// Filter methods (chainable)
func (q *TagQuery) WithUUID(uuid string) *TagQuery
func (q *TagQuery) WithTitle(title string) *TagQuery
func (q *TagQuery) WithParent(parentUUID string) *TagQuery
func (q *TagQuery) IncludeItems(include bool) *TagQuery

// Terminal methods (execute query)
func (q *TagQuery) All(ctx context.Context) ([]Tag, error)
func (q *TagQuery) First(ctx context.Context) (*Tag, error)
func (q *TagQuery) Count(ctx context.Context) (int, error)
```

### Token Retrieval

The `Token()` method retrieves the authentication token required for URL Scheme update operations. This bridges the DB layer with the URL Scheme layer.

```go
// Token returns the authentication token for URL Scheme update operations.
// The token is stored in the TMSettings table.
func (db *DB) Token(ctx context.Context) (string, error)
```

## Usage Examples

### Basic Usage

```go
db, err := things3.NewDB()
if err != nil {
    return err
}
defer db.Close()

// Convenience methods
tasks, _ := db.Inbox(ctx)
tasks, _ := db.Today(ctx)
tasks, _ := db.Upcoming(ctx)
tasks, _ := db.Todos(ctx)
tasks, _ := db.Projects(ctx)
```

### Query Builder with Type-Safe Sub-Builders

```go
// Find all incomplete todos with a specific tag
tasks, _ := db.Tasks().
    Type().Todo().
    Status().Incomplete().
    InTag("work").
    All(ctx)

// Find a specific task by UUID
task, _ := db.Tasks().
    WithUUID("task-uuid").
    First(ctx)

// Find all tasks in a project
tasks, _ := db.Tasks().
    InProject("project-uuid").
    All(ctx)

// Find tasks with deadlines in the past
tasks, _ := db.Tasks().
    Deadline().Past().
    All(ctx)

// Find tasks starting in the future
tasks, _ := db.Tasks().
    StartDate().Future().
    All(ctx)

// Find tasks with deadlines before a specific date
tasks, _ := db.Tasks().
    Deadline().Before(time.Now().AddDate(0, 0, 7)).
    All(ctx)

// Count tasks with a deadline
count, _ := db.Tasks().
    Deadline().Exists(true).
    Count(ctx)

// Find tasks created within last 7 days
tasks, _ := db.Tasks().
    CreatedAfter(things3.DaysAgo(7)).
    All(ctx)

// Search for tasks
tasks, _ := db.Tasks().
    Search("meeting").
    All(ctx)
```

### With URL Scheme Integration

```go
// Get token for URL Scheme update operations
db, _ := things3.NewDB()
token, _ := db.Token(ctx)

// Use token with URL Scheme (see RFC 004)
scheme := things3.NewScheme()
auth := scheme.WithToken(token)
url, _ := auth.UpdateTodo("uuid").Completed(true).Build()
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

```text
things3/
    client.go           # DB type, NewDB(), Close(), Filepath()
    client_options.go   # DBOption, WithDBPath(), WithPrintSQL()
    convenience.go      # Inbox(), Today(), Todos(), Projects(), etc.
    query.go            # TaskQuery builder
    query_filter.go     # Type-safe sub-builders (StatusFilter, TypeFilter, etc.)
    query_area.go       # AreaQuery builder
    query_tag.go        # TagQuery builder
    filter.go           # Internal SQL filter building
    sql.go              # SQL query construction
    models.go           # Task, Area, Tag, ChecklistItem structs
    types.go            # TaskType, Status, StartBucket enums
    date.go             # Things date format conversion
    time_helpers.go     # Time helper functions (DaysAgo, WeeksAgo, etc.)
    database.go         # Database connection and path discovery
    url.go              # Things URL scheme support
    errors.go           # Error definitions
    constants.go        # Table names, column mappings
```

| File | Responsibility |
|------|----------------|
| `client.go` | DB type definition, NewDB constructor, connection management |
| `client_options.go` | Functional options pattern for configuration |
| `convenience.go` | Pre-built queries for common operations |
| `query.go` | TaskQuery builder with filter and terminal methods |
| `query_filter.go` | Type-safe sub-builders (StatusFilter, TypeFilter, DateFilter, etc.) |
| `query_area.go` | AreaQuery builder |
| `query_tag.go` | TagQuery builder |
| `filter.go` | Internal filter interface and implementations |
| `sql.go` | SQL query construction and execution |

## Testing Strategy

### Integration Tests

```go
func TestDB_Inbox(t *testing.T) {
    db, err := things3.NewDB(things3.WithDBPath("testdata/test.sqlite"))
    require.NoError(t, err)
    defer db.Close()

    tasks, err := db.Inbox(context.Background())
    require.NoError(t, err)
    assert.NotEmpty(t, tasks)
}

func TestDB_Token(t *testing.T) {
    db, err := things3.NewDB(things3.WithDBPath("testdata/test.sqlite"))
    require.NoError(t, err)
    defer db.Close()

    token, err := db.Token(context.Background())
    require.NoError(t, err)
    assert.NotEmpty(t, token)
}
```

### Query Builder Tests

```go
func TestTaskQuery_TypeSafeFilters(t *testing.T) {
    db, _ := things3.NewDB(things3.WithDBPath("testdata/test.sqlite"))
    defer db.Close()

    tests := []struct {
        name    string
        query   func() *things3.TaskQuery
        wantMin int
    }{
        {
            name:    "all todos",
            query:   func() *things3.TaskQuery { return db.Tasks().Type().Todo() },
            wantMin: 1,
        },
        {
            name:    "incomplete tasks",
            query:   func() *things3.TaskQuery { return db.Tasks().Status().Incomplete() },
            wantMin: 0,
        },
        {
            name:    "tasks with deadlines",
            query:   func() *things3.TaskQuery { return db.Tasks().Deadline().Exists(true) },
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
| Type-Safe Sub-Builders | StatusFilter, TypeFilter, DateFilter provide compile-time safety |
| Read-Only | No write operations to the database |
| Type Safety | Strongly typed query results, no raw SQL exposure |
| Functional Options | Configuration via `DBOption` functions |
| Value Semantics | Return value slices (`[]Task`) not pointer slices (`[]*Task`) |

## References

- RFC 002: Database Schema - Schema definitions used by this API
- RFC 004: URL Scheme - Uses `Token()` for authenticated operations
- [things.py](https://github.com/thingsapi/things.py) - Python reference implementation
