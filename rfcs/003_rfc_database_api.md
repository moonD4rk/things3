# RFC 003: Database API

Status: Draft
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

Direct access methods for common queries. All methods accept `context.Context` as the first parameter.

```go
// Inbox returns all tasks in the Inbox.
func (db *DB) Inbox(ctx context.Context) ([]*Task, error)

// Today returns all tasks scheduled for today.
func (db *DB) Today(ctx context.Context) ([]*Task, error)

// Upcoming returns all tasks with future start dates.
func (db *DB) Upcoming(ctx context.Context) ([]*Task, error)

// Anytime returns all tasks in the Anytime list.
func (db *DB) Anytime(ctx context.Context) ([]*Task, error)

// Someday returns all tasks in the Someday list.
func (db *DB) Someday(ctx context.Context) ([]*Task, error)

// Logbook returns all completed tasks.
func (db *DB) Logbook(ctx context.Context) ([]*Task, error)

// Trash returns all trashed tasks.
func (db *DB) Trash(ctx context.Context) ([]*Task, error)
```

### Query Builders

Builder pattern for complex queries with chainable filter methods and terminal execution methods.

```go
// Query builder entry points
func (db *DB) Todos() *TaskQuery
func (db *DB) Projects() *TaskQuery
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
func (q *TaskQuery) WithTag(tag string) *TaskQuery
func (q *TaskQuery) WithStatus(status Status) *TaskQuery
func (q *TaskQuery) InProject(projectUUID string) *TaskQuery
func (q *TaskQuery) InArea(areaUUID string) *TaskQuery
func (q *TaskQuery) InHeading(headingUUID string) *TaskQuery
func (q *TaskQuery) StartDate(date string) *TaskQuery
func (q *TaskQuery) Deadline(date string) *TaskQuery
func (q *TaskQuery) Trashed(trashed bool) *TaskQuery

// Terminal methods (execute query)
func (q *TaskQuery) All(ctx context.Context) ([]*Task, error)
func (q *TaskQuery) First(ctx context.Context) (*Task, error)
func (q *TaskQuery) Count(ctx context.Context) (int, error)
```

#### AreaQuery

```go
type AreaQuery struct {
    // internal fields
}

// Filter methods (chainable)
func (q *AreaQuery) WithUUID(uuid string) *AreaQuery
func (q *AreaQuery) WithTag(tag string) *AreaQuery
func (q *AreaQuery) Visible(visible bool) *AreaQuery

// Terminal methods (execute query)
func (q *AreaQuery) All(ctx context.Context) ([]*Area, error)
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

// Terminal methods (execute query)
func (q *TagQuery) All(ctx context.Context) ([]*Tag, error)
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
```

### Query Builder Usage

```go
// Find all incomplete todos with a specific tag
tasks, _ := db.Todos().
    WithTag("work").
    WithStatus(things3.StatusIncomplete).
    All(ctx)

// Find a specific task by UUID
task, _ := db.Todos().
    WithUUID("task-uuid").
    First(ctx)

// Find all tasks in a project
tasks, _ := db.Todos().
    InProject("project-uuid").
    All(ctx)

// Count tasks with a deadline
count, _ := db.Todos().
    Deadline("2024-12-31").
    Count(ctx)
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

    // ErrDatabaseReadOnly is returned when attempting write operations.
    ErrDatabaseReadOnly = errors.New("things3: database is read-only")

    // ErrTokenNotFound is returned when the auth token is not configured.
    ErrTokenNotFound = errors.New("things3: auth token not found in settings")
)
```

## File Organization

```text
things3/
├── db.go               # DB type, NewDB(), Close(), Filepath()
├── db_options.go       # DBOption, WithDBPath(), WithPrintSQL()
├── db_convenience.go   # Inbox(), Today(), Upcoming(), etc.
├── db_query.go         # TaskQuery builder
├── db_query_area.go    # AreaQuery builder
├── db_query_tag.go     # TagQuery builder
├── db_token.go         # Token() method
└── db_test.go          # Integration tests
```

| File | Responsibility |
|------|----------------|
| `db.go` | DB type definition, NewDB constructor, connection management |
| `db_options.go` | Functional options pattern for configuration |
| `db_convenience.go` | Pre-built queries for common operations |
| `db_query.go` | TaskQuery builder with filter and terminal methods |
| `db_query_area.go` | AreaQuery builder |
| `db_query_tag.go` | TagQuery builder |
| `db_token.go` | Auth token retrieval from TMSettings |
| `db_test.go` | Integration tests using testdata database |

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
func TestTaskQuery_Filters(t *testing.T) {
    db, _ := things3.NewDB(things3.WithDBPath("testdata/test.sqlite"))
    defer db.Close()

    tests := []struct {
        name    string
        query   func() *things3.TaskQuery
        wantMin int
    }{
        {
            name:    "all todos",
            query:   func() *things3.TaskQuery { return db.Todos() },
            wantMin: 1,
        },
        {
            name:    "with tag",
            query:   func() *things3.TaskQuery { return db.Todos().WithTag("work") },
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
| Read-Only | No write operations to the database |
| Type Safety | Strongly typed query results, no raw SQL exposure |
| Functional Options | Configuration via `DBOption` functions |

## References

- RFC 002: Database Schema - Schema definitions used by this API
- RFC 004: URL Scheme - Uses `Token()` for authenticated operations
- [things.py](https://github.com/thingsapi/things.py) - Python reference implementation
