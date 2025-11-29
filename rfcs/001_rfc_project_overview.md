# RFC 001: Project Overview

Status: Accepted
Author: @moond4rk

## Summary

things3 is a Go library providing read-only access to the Things 3 macOS application's SQLite database. It is a complete port of the Python things.py library with full API parity, enabling Go developers to programmatically query tasks, projects, areas, and tags from the Things 3 app.

## Design

### Goals

1. **Full API Parity**: Replicate all 30+ APIs from Python things.py
2. **Optimal Performance**: Use native SQLite driver for best query performance
3. **Idiomatic Go API**: Follow Go conventions while maintaining familiar semantics
4. **Read-Only Access**: Safe database queries without modification risk
5. **Minimal Dependencies**: Only essential dependencies (SQLite driver, testing)

### Non-Goals

1. Write operations to the Things 3 database
2. HTTP/gRPC service wrapper (library only)
3. Cross-platform support (macOS only, matching Things 3 availability)
4. Real-time sync or change detection

### Core Decisions

**Module Path**: `github.com/moond4rk/things3`

**SQLite Driver**: `github.com/mattn/go-sqlite3`
- CGO-based native SQLite implementation
- Optimal performance for database operations
- Most widely used and battle-tested Go SQLite driver
- Appropriate choice since Things 3 is macOS-only

**API Style**: Hybrid Pattern
- Functional Options for Client configuration
- Query Builder with chainable methods for task queries
- Direct convenience methods for common operations

**Package Structure**: Simplified single-package layout
- All public types and functions in the main `things3` package
- No internal subpackages for this scope
- Clear separation between files by responsibility

### API Design Principles

1. **Context-First**: All query methods accept `context.Context` as first parameter
2. **Nil-Safe**: Optional fields use pointer types with proper nil handling
3. **Builder Pattern**: Query filters are chainable and composable
4. **Terminal Methods**: `All()`, `First()`, `Count()` execute queries

### Database Compatibility

The library reads from the Things 3 SQLite database located at:
- Default: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`
- Override via `THINGSDB` environment variable or `WithDatabasePath()` option

### Type System

Integer-based enums for direct database mapping:

| Type | Values |
|------|--------|
| TaskType | 0=to-do, 1=project, 2=heading |
| Status | 0=incomplete, 2=canceled, 3=completed |
| StartBucket | 0=Inbox, 1=Anytime, 2=Someday |

## Implementation Notes

### File Organization

| File | Responsibility |
|------|----------------|
| client.go | Client type, New(), Close() |
| client_options.go | Functional options (WithDatabasePath, WithPrintSQL) |
| query.go | TaskQuery builder with filter methods |
| query_area.go | AreaQuery builder |
| query_tag.go | TagQuery builder |
| convenience.go | Inbox(), Today(), Todos(), etc. |
| models.go | Task, Area, Tag, ChecklistItem structs |
| types.go | TaskType, Status, StartBucket enums |
| date.go | Things date format conversion |
| sql.go | SQL query building and execution |
| database.go | Connection and path discovery |
| url.go | Things URL scheme support |
| errors.go | Error definitions |
| constants.go | Table names, column mappings |

### Dependencies

| Package | Purpose |
|---------|---------|
| github.com/mattn/go-sqlite3 | SQLite database driver (CGO) |
| github.com/stretchr/testify | Testing assertions (dev only) |

### Quality Requirements

- All exported types and functions must have Go doc comments
- Table-driven tests for query building
- Integration tests using test database in testdata/
- golangci-lint compliance with project configuration

### Reference

Python Source: https://github.com/thingsapi/things.py
