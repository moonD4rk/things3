# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Mandatory Rules

- **English Only**: All code, comments, documentation, and commit messages MUST be in English
- **No Emoji**: Never use emoji in any file (code, docs, comments, commits)
- **No Local Paths**: Never expose local machine paths in code, tests, or documentation
- **Design Focus**: RFC documents focus on design decisions, avoid large code blocks
- **No Backward Compatibility**: Breaking changes are acceptable. Prioritize optimal design and elegant code over backward compatibility. Do not deprecate, just remove or redesign

## Project Overview

**things3** is a Go library providing read-only access to the Things 3 macOS application's SQLite database and type-safe URL Scheme building and execution. It is a Go port of the Python things.py library with full API parity.

## Build and Development Commands

```bash
go test ./...                              # Run all tests
go test -cover ./...                       # Run tests with coverage
go test -run TestTaskQuery ./...           # Run single test
golangci-lint run                          # Run linter
gofumpt -l -w .                            # Format (stricter than gofmt)
goimports -w -local github.com/moond4rk/things3 . # Format Import
go build ./...                             # Build
```

## Architecture

### Design Philosophy

- **Single Entry Point**: `NewClient()` is the only public constructor
- **Interface-Based API**: Methods return interfaces, not concrete types
- **Go Idioms**: Small interfaces, composition over inheritance

### Design Patterns

- **Client Configuration**: Functional Options pattern (ClientOption)
- **Query Building**: Builder pattern with chainable methods returning interfaces
- **URL Building**: Builder pattern with Build() or Execute()
- **Convenience Methods**: Direct access for common queries (Inbox, Today, etc.)

### Interface Hierarchy

```
Layer 1: Reusable Small Interfaces (Terminal Operations)
├── TaskQueryExecutor   (All, First, Count)
├── AreaQueryExecutor   (All, First, Count)
├── TagQueryExecutor    (All, First)
└── URLBuilder          (Build, Execute)

Layer 2: Sub-builder Interfaces (Small, Type-Safe)
├── TypeFilterBuilder   (Todo, Project, Heading)
├── StatusFilterBuilder (Incomplete, Completed, Canceled, Any)
├── StartFilterBuilder  (Inbox, Anytime, Someday)
└── DateFilterBuilder   (Exists, Future, Past, On, Before, After, etc.)

Layer 3: Functional Group Interfaces (For Composition)
├── TaskRelationFilter  (InArea, InProject, InTag, etc.)
├── TaskStateFilter     (Type, Status, Start, Trashed)
└── TaskTimeFilter      (CreatedAfter, StartDate, StopDate, Deadline)

Layer 4: Composed Query Builders
├── TaskQueryBuilder = TaskQueryExecutor + TaskRelationFilter + TaskStateFilter + TaskTimeFilter
├── AreaQueryBuilder = AreaQueryExecutor + filters
└── TagQueryBuilder  = TagQueryExecutor + filters

Layer 5: URL Scheme Builders
├── TodoAdder, ProjectAdder     (URLBuilder + creation methods)
├── TodoUpdater, ProjectUpdater (URLBuilder + update methods)
└── ShowNavigator               (navigation methods)

Layer 6: Batch Operations
├── BatchCreator, AuthBatchCreator
└── BatchTodoConfigurator, BatchProjectConfigurator
```

### Core Components

| File | Purpose |
|------|---------|
| `client.go` | Client type, NewClient(), unified API entry point |
| `client_options.go` | ClientOption functional options |
| `interfaces.go` | All public interface definitions (6 layers) |
| `db.go` | Internal db type, database operations |
| `db_options.go` | Internal dbOption functional options |
| `scheme.go` | Internal scheme type, URL building and execution |
| `scheme_options.go` | Internal schemeOption |
| `scheme_builder.go` | AddTodoBuilder, AddProjectBuilder |
| `scheme_update.go` | UpdateTodoBuilder, UpdateProjectBuilder |
| `scheme_show.go` | ShowBuilder for navigation |
| `scheme_json.go` | BatchBuilder for batch operations |
| `query.go` | taskQuery builder with filter methods |
| `query_filter.go` | typeFilter, statusFilter, startFilter, dateFilter |
| `query_area.go` | areaQuery builder |
| `query_tag.go` | tagQuery builder |
| `convenience.go` | Inbox(), Today(), Todos(), etc. |
| `models.go` | Task, Area, Tag, ChecklistItem structs |
| `types.go` | TaskType, Status, StartBucket enums |
| `date.go` | Things date format conversion |
| `sql.go` | SQL query building and execution |
| `database.go` | Database connection and path discovery |
| `errors.go` | Error definitions |
| `constants.go` | Table names, column mappings |

### Type System

Enums are integer-based for database mapping:
- TaskType: 0=to-do, 1=project, 2=heading
- Status: 0=incomplete, 2=canceled, 3=completed
- StartBucket: 0=Inbox, 1=Anytime, 2=Someday

### Things Date Format

Things uses custom binary date formats:
- Date: YYYYYYYYYYYMMMMDDDDD0000000 (27-bit integer)
- Time: hhhhhmmmmmm00000000000000000000

## API Design

### Unified Client Pattern

All operations go through a single Client:

```go
client, _ := things3.NewClient()
defer client.Close()

// Query operations
tasks, _ := client.Today(ctx)
tasks, _ := client.Tasks().Status().Incomplete().All(ctx)

// Add operations
client.AddTodo().Title("Buy milk").Execute(ctx)

// Update operations (auto-manages auth token)
client.UpdateTodo(uuid).Completed(true).Execute(ctx)

// Show operations
client.Show(ctx, uuid)
```

### Query Builder Pattern

Filter methods are chainable, terminal methods execute the query:
- `.All(ctx)` - Get all matching results
- `.First(ctx)` - Get first match
- `.Count(ctx)` - Count matches

### Python to Go API Mapping

| Python | Go |
|--------|-----|
| `tasks(uuid=X)` | `client.Tasks().WithUUID(X).First(ctx)` |
| `tasks(**kwargs)` | `client.Tasks().<filters>.All(ctx)` |
| `todos()` | `client.Todos(ctx)` |
| `inbox()` | `client.Inbox(ctx)` |
| `today()` | `client.Today(ctx)` |

## Code Quality Standards

### Naming Conventions

- Exported types: PascalCase (Client, Task, TaskQueryBuilder)
- Internal types: camelCase (db, scheme, taskQuery)
- Interfaces: Verb+er or descriptive (TaskQueryBuilder, URLBuilder)
- Enums: Type prefix (TaskTypeTodo, StatusCompleted)
- Query methods: With* for filters, In* for relationships

### Documentation Requirements

Every exported type and function MUST have Go doc comments starting with the identifier name.

### Testing Requirements

- Table-driven tests for query building
- Integration tests with test database in testdata/
- Never hardcode local paths in tests

## RFC Documentation

RFC documents are stored in `rfcs/` directory with naming format `NNN_snake_case_title.md`.

### RFC Template

```
# RFC NNN: Title

Status: Draft | Accepted | Implemented
Author: @username
Date: YYYY-MM-DD

## Summary
One paragraph describing the core content.

## Design
Detailed design decisions with rationale.

## Implementation Notes
Key implementation details and considerations.
```

## Release Note Template

Release notes use this format with Library and CLI sections:

```
# things3 vX.Y.Z - Short Title

One-line summary.

## Breaking Changes (if any)
Brief description of what changed and migration path.

## Library
- **Feature name** (#PR): One-line description
- **Bug fix**: Description (#PR)

## CLI
- **Feature name** (#PR): One-line description

**Full Changelog**: https://github.com/moonD4rk/things3/compare/vPREV...vX.Y.Z
```

## CLI Development (cmd/things3)

CLI uses Cobra with factory function pattern (no `init()`).

### Key Rules

- Use `NewXxxCmd()` factory functions, avoid `init()`
- Use `cmd.OutOrStdout()` for testability
- Return errors, let `main()` handle `os.Exit()`
- Register subcommands explicitly in `NewRootCmd()`

## Dependencies

- `github.com/mattn/go-sqlite3` - SQLite driver (CGO, optimal for macOS-only)
- `github.com/stretchr/testify` - Testing (dev only)

## Reference

- Python Source: https://github.com/thingsapi/things.py
- Database path discoverable via THINGSDB environment variable
