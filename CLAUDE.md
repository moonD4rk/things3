# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**things3** is a Go library providing read-only access to the Things 3 macOS application's SQLite database. It is a Go port of the Python [things.py](https://github.com/thingsapi/things.py) library with full API parity.

**Goal**: Provide a clean, idiomatic Go API for querying Things 3 tasks, projects, areas, and tags.

## Build & Development Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a single test
go test -run TestTaskQuery ./...

# Run linter (golangci-lint required)
golangci-lint run

# Format code
gofmt -w .
goimports -w -local github.com/moond4rk/things3 .

# Generate documentation
go doc -all

# Build (library only, no main)
go build ./...
```

## Architecture

### Design Patterns

- **Client Configuration**: Functional Options pattern (`New(WithDatabasePath(...))`)
- **Query Building**: Builder pattern with chainable methods (`client.Tasks().WithTag("home").All(ctx)`)
- **Convenience Methods**: Direct access for common queries (`Inbox()`, `Today()`, etc.)

### Core Components

| File | Purpose |
|------|---------|
| `client.go` | Client type, New(), Close() |
| `client_options.go` | Functional options for client config |
| `query.go` | TaskQuery builder with filter methods |
| `query_area.go` | AreaQuery builder |
| `query_tag.go` | TagQuery builder |
| `convenience.go` | Inbox(), Today(), Todos(), etc. |
| `models.go` | Task, Area, Tag, ChecklistItem structs |
| `types.go` | TaskType, Status, StartBucket enums |
| `date.go` | Things date format conversion (critical) |
| `sql.go` | SQL query building and execution |
| `database.go` | Database connection and path discovery |
| `url.go` | Things URL scheme support |
| `errors.go` | Error definitions |
| `constants.go` | Table names, column mappings |

### Type System

```go
// Enums (integer-based for database mapping)
type TaskType int     // 0=to-do, 1=project, 2=heading
type Status int       // 0=incomplete, 2=canceled, 3=completed
type StartBucket int  // 0=Inbox, 1=Anytime, 2=Someday
```

### Database Path

Things 3 database location:
- Things 3.15.16+: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite`
- Legacy: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`
- Override via `THINGSDB` environment variable

### Things Date Format

Things uses a custom binary date format that must be converted:
- **Date**: `YYYYYYYYYYYMMMMDDDDD0000000` (27-bit integer)
- **Time**: `hhhhhmmmmmm00000000000000000000`

The `date.go` file handles all conversions between Things format and Go `time.Time`.

## API Design

### Query Builder Pattern

```go
// All filter methods are chainable
client.Tasks().
    WithType(TaskTypeTodo).
    WithStatus(StatusIncomplete).
    InProject("project-uuid").
    WithTag("urgent").
    WithDeadline(DateOpExists).
    IncludeItems(true).
    All(ctx)

// Terminal methods
.All(ctx) ([]Task, error)    // Get all matching
.First(ctx) (*Task, error)   // Get first match
.Count(ctx) (int, error)     // Count matches
```

### Python → Go API Mapping

| Python | Go |
|--------|-----|
| `tasks(uuid=X)` | `client.Tasks().WithUUID(X).First(ctx)` |
| `tasks(**kwargs)` | `client.Tasks().<filters>.All(ctx)` |
| `tasks(count_only=True)` | `client.Tasks().<filters>.Count(ctx)` |
| `todos()` | `client.Todos(ctx)` |
| `inbox()` | `client.Inbox(ctx)` |
| `today()` | `client.Today(ctx)` |
| `search(query)` | `client.Search(ctx, query)` |

## Code Quality Standards

### Naming Conventions

- **Exported types**: PascalCase (`TaskQuery`, `ChecklistItem`)
- **Private functions**: camelCase (`buildWhereClause`)
- **Constants**: PascalCase for exported, camelCase for internal
- **Enums**: Type prefix (`TaskTypeTodo`, `StatusCompleted`)
- **Query methods**: `With*` for filters, `In*` for relationships

### Documentation Requirements

Every exported type and function MUST have:
- Package-level doc comment in `doc.go`
- Function doc starting with function name
- Example code for complex APIs

```go
// TaskQuery builds queries for Things 3 tasks.
// Use [Client.Tasks] to create a new TaskQuery.
type TaskQuery struct { ... }

// WithTag filters tasks by tag title.
// Use [TaskQuery.HasTags] to filter by tag existence.
func (q *TaskQuery) WithTag(title string) *TaskQuery { ... }
```

### Testing Requirements

- Table-driven tests for query building
- Integration tests with test database
- Test file naming: `*_test.go`
- Use `testdata/` for test fixtures

## Dependencies

- `modernc.org/sqlite` - Pure Go SQLite driver (no CGO)
- `github.com/stretchr/testify` - Testing (dev only)

## Reference

**Python Source** (for porting logic):
- `/Users/leeroger/Developer/python/opython/things.py/things/api.py`
- `/Users/leeroger/Developer/python/opython/things.py/things/database.py`

**Database Tables**:
| Table | Purpose |
|-------|---------|
| TMTask | Tasks (to-do, project, heading) |
| TMArea | Areas/workspaces |
| TMTag | Tags |
| TMChecklistItem | Checklist items |
| TMTaskTag | Task-Tag relationship |
| TMSettings | App settings (auth token) |
