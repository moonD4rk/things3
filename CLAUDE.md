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
go test -run TestTodoQuery ./...           # Run single test
golangci-lint run                          # Run linter
gofumpt -l -w .                            # Format (stricter than gofmt)
goimports -w -local github.com/moond4rk/things3 . # Format Import
go build ./...                             # Build
```

## Architecture

### Design Philosophy

- **Single Entry Point**: `NewClient()` is the only public constructor
- **Typed Query Builders**: Separate builders for Todo, Project, Heading (no union Task type)
- **Flat Data Model**: No nested items; parent refs inline, child queries via builders
- **Interface-Based API**: Methods return interfaces, not concrete types
- **Go Idioms**: Small interfaces, generics for type-safe sub-builders

### Design Patterns

- **Client Configuration**: Functional Options pattern (ClientOption)
- **Query Building**: Builder pattern with chainable methods returning typed interfaces
- **URL Building**: Builder pattern with Build() or Execute()
- **Generic Sub-builders**: `StatusFilter[T]`, `StartFilter[T]`, `DateFilter[T]`

### Interface Hierarchy

```
Layer 1: Terminal Operations
├── TodoQueryExecutor    (All, First, Count -> []Todo)
├── ProjectQueryExecutor (All, First, Count -> []Project)
├── HeadingQueryExecutor (All, First, Count -> []Heading)
├── AreaQueryExecutor    (All, First, Count -> []Area)
├── TagQueryExecutor     (All, First -> []Tag)
└── URLBuilder           (Build, Execute)

Layer 2: Generic Sub-builders
├── StatusFilter[T]  (Incomplete, Completed, Canceled, Any)
├── StartFilter[T]   (Inbox, Anytime, Someday)
└── DateFilter[T]    (Exists, Future, Past, On, Before, After, etc.)

Layer 3: Composed Query Builders
├── TodoQueryBuilder    = TodoQueryExecutor + filters + IncludeChecklist
├── ProjectQueryBuilder = ProjectQueryExecutor + filters
├── HeadingQueryBuilder = HeadingQueryExecutor + WithUUID + InProject
├── AreaQueryBuilder    = AreaQueryExecutor + filters
└── TagQueryBuilder     = TagQueryExecutor + filters

Layer 4: URL Scheme Builders (aliased from internal/scheme)
├── TodoAdder, ProjectAdder     (creation)
├── TodoUpdater, ProjectUpdater (update)
└── ShowNavigator               (navigation)

Layer 5: Batch Operations (aliased from internal/scheme)
├── BatchCreator, AuthBatchCreator
└── BatchTodoConfigurator, BatchProjectConfigurator
```

### Core Components

| File | Purpose |
|------|---------|
| `client.go` | Client type, NewClient(), query/URL entry points |
| `client_options.go` | ClientOption functional options |
| `interfaces.go` | All public interface definitions |
| `models.go` | Todo, Project, Heading, Area, Tag, ChecklistItem |
| `types.go` | TaskType, Status, StartBucket enums |
| `db.go` | Internal db type, row-to-model conversion |
| `query.go` | todoQuery, projectQuery, headingQuery builders |
| `query_filter.go` | Generic statusFilter, startFilter, dateFilter |
| `query_area.go` | areaQuery builder |
| `query_tag.go` | tagQuery builder |
| `errors.go` | Error definitions |
| `time_helpers.go` | DaysAgo, WeeksAgo, Today, ApplyWhen |
| `internal/database/` | DB connection, SQL, filters, date conversion |
| `internal/scheme/` | URL scheme building and execution |

### Domain Types

Separate types per domain concept (no union Task type):
- `Todo` - actionable item with checklist, relationships, dates
- `Project` - container for organizing todos
- `Heading` - grouping label within a project (UUID + Title only)
- `Area` - high-level responsibility area
- `Tag` - label for categorizing items
- `ChecklistItem` - sub-item within a todo

### Type System

Enums are integer-based for database mapping:
- Status: 0=incomplete, 2=canceled, 3=completed
- StartBucket: 0=inbox, 1=anytime, 2=someday
- TaskType: 0=todo, 1=project, 2=heading (internal use only)

### Things Date Format

Things uses custom binary date formats:
- Date: YYYYYYYYYYYMMMMDDDDD0000000 (27-bit integer)
- Time: hhhhhmmmmmm00000000000000000000

## API Design

### Client Entry Points

```go
client, _ := things3.NewClient()
defer client.Close()

// Typed query builders
todos, _ := client.Todos().Status().Incomplete().All(ctx)
projects, _ := client.Projects().InArea(uuid).All(ctx)
headings, _ := client.Headings().InProject(uuid).All(ctx)
areas, _ := client.Areas().All(ctx)
tags, _ := client.Tags().All(ctx)

// URL scheme operations
client.AddTodo().Title("Buy milk").Execute(ctx)
client.UpdateTodo(uuid).Completed(true).Execute(ctx)
client.Show(ctx, uuid)
```

### Query Builder Pattern

Filter methods are chainable, terminal methods execute the query:
- `.All(ctx)` - Get all matching results
- `.First(ctx)` - Get first match (auto-loads checklist for todos)
- `.Count(ctx)` - Count matches

### Relationship Model

- **Upward (inline)**: `todo.ProjectUUID`, `todo.AreaTitle` (from SQL JOIN, zero cost)
- **Downward (query)**: `client.Todos().InProject(uuid)` (separate builder call)

## Code Quality Standards

### Naming Conventions

- Exported types: PascalCase (Client, Todo, TodoQueryBuilder)
- Internal types: camelCase (db, taskQuery, todoQuery)
- Interfaces: Verb+er or descriptive (TodoQueryBuilder, URLBuilder)
- Enums: Type prefix (StatusCompleted, StartInbox)
- Query methods: With* for identity, In* for relationships, Has* for existence

### Documentation Requirements

Every exported type and function MUST have Go doc comments starting with the identifier name.

### Testing Requirements

- Table-driven tests for query building
- Integration tests with test database in testdata/
- Never hardcode local paths in tests

## RFC Documentation

RFC documents are stored in `rfcs/` directory with naming format `NNN_snake_case_title.md`.

Active RFCs:
- RFC 008: Domain Model Redesign (Superseded by RFC 009 for query/model sections)
- RFC 009: Query Builder Redesign (current design spec)

### RFC Template

```
# RFC NNN: Title

Status: Draft | Accepted | Implemented | Superseded
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
