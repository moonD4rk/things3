# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Mandatory Rules

- **English Only**: All code, comments, documentation, and commit messages MUST be in English
- **No Emoji**: Never use emoji in any file (code, docs, comments, commits)
- **No Local Paths**: Never expose local machine paths in code, tests, or documentation
- **Design Focus**: RFC documents focus on design decisions, avoid large code blocks

## Project Overview

**things3** is a Go library providing read-only access to the Things 3 macOS application's SQLite database. It is a Go port of the Python things.py library with full API parity.

## Build and Development Commands

```bash
go test ./...                              # Run all tests
go test -cover ./...                       # Run tests with coverage
go test -run TestTaskQuery ./...           # Run single test
golangci-lint run                          # Run linter
gofmt -w . && goimports -w -local github.com/moond4rk/things3 .  # Format
go build ./...                             # Build
```

## Architecture

### Design Patterns

- **Client Configuration**: Functional Options pattern
- **Query Building**: Builder pattern with chainable methods
- **Convenience Methods**: Direct access for common queries

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
| `date.go` | Things date format conversion |
| `sql.go` | SQL query building and execution |
| `database.go` | Database connection and path discovery |
| `url.go` | Things URL scheme support |
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

- Exported types: PascalCase (TaskQuery, ChecklistItem)
- Private functions: camelCase (buildWhereClause)
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

### Planned RFCs

- 001_rfc_project_overview.md - Project goals, non-goals, core decisions
- 002_rfc_api_design.md - API patterns and public interface
- 003_rfc_database_schema.md - Things 3 database structure
- 004_rfc_types_and_models.md - Type system design
- 005_rfc_error_handling.md - Error handling strategy

## Dependencies

- `github.com/mattn/go-sqlite3` - SQLite driver (CGO, optimal for macOS-only)
- `github.com/stretchr/testify` - Testing (dev only)

## Reference

- Python Source: https://github.com/thingsapi/things.py
- Database path discoverable via THINGSDB environment variable
