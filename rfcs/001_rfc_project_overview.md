# RFC 001: Project Overview

Status: Accepted
Author: @moond4rk

## Summary

things3 is a Go library providing read-only access to the Things 3 macOS application's SQLite database and type-safe URL Scheme building and execution. It is a complete port of the Python things.py library with full API parity, plus a modern URL builder API with background execution support.

## Architecture

The library provides two independent entry points with clear separation of concerns:

```
things3 library
|
+-- NewDB()      -> *DB      (Database operations, stateful)
|   |
|   +-- Connection: Close(), Filepath()
|   +-- Convenience: Inbox(), Today(), Upcoming(), Anytime(), Someday(), Logbook(), Trash()
|   +-- Query Builders: Todos(), Projects(), Areas(), Tags()
|   +-- Auth Token: Token() -> for URL Scheme update operations
|
+-- NewScheme(opts...)  -> *Scheme  (URL building + execution)
    |
    +-- Options: WithForeground()
    |
    +-- [Execution Methods]
    |   +-- Show(ctx, uuid)      -> error
    |   +-- Search(ctx, query)   -> error
    |
    +-- [URL Building - No Auth Required]
    |   +-- Todo()        -> *TodoBuilder       -> Build() string
    |   +-- Project()     -> *ProjectBuilder    -> Build() string
    |   +-- ShowBuilder() -> *ShowBuilder       -> Build() string
    |   +-- JSON()        -> *JSONBuilder       -> Build() string
    |   +-- SearchURL(query) -> string
    |   +-- Version()     -> string
    |
    +-- WithToken(token)  -> *AuthScheme  (Authenticated operations)
        +-- UpdateTodo(id)    -> *UpdateTodoBuilder    -> Build() | Execute(ctx)
        +-- UpdateProject(id) -> *UpdateProjectBuilder -> Build() | Execute(ctx)
        +-- JSON()            -> *AuthJSONBuilder      -> Build() string
```

### Design Principles

| Principle | Implementation |
|-----------|----------------|
| Separation of Concerns | `NewDB()` for data, `NewScheme()` for URLs and execution |
| Compile-time Safety | `WithToken()` enforces auth requirements at compile time |
| Functional Options | `SchemeOption` for configurable behavior |
| Background by Default | URL execution uses `osascript` to avoid stealing focus |
| Builder Pattern | Chainable methods with terminal `.Build()` or `.Execute()` |
| Context-First | DB queries and execution accept `context.Context` |

## RFC Index

| RFC | Title | Status | Description |
|-----|-------|--------|-------------|
| 001 | Project Overview | Accepted | Goals, architecture, dependencies (this document) |
| 002 | Database Schema | Accepted | Things 3 SQLite schema and SQL patterns |
| 003 | Database API | Draft | `NewDB()`, query builders, convenience methods |
| 004 | URL Scheme | Draft | `NewScheme()`, URL builders, official reference |
| 005 | Unified Client | Draft | `NewClient()`, single entry point, token management |
| 006 | Interface Abstraction | Draft | Public interfaces, hide implementation details |

### RFC Dependencies

```
001 Project Overview
 |
 +-- 002 Database Schema
 |    |
 |    +-- 003 Database API (uses schema)
 |
 +-- 004 URL Scheme (uses Token from 003)
 |
 +-- 005 Unified Client (combines 003 + 004)
      |
      +-- 006 Interface Abstraction (evolves 005)
```

## Goals

1. **Full API Parity**: Replicate all 30+ APIs from Python things.py
2. **Optimal Performance**: Use native SQLite driver for best query performance
3. **Idiomatic Go API**: Follow Go conventions while maintaining familiar semantics
4. **Read-Only Database**: Safe database queries without modification risk
5. **Type-Safe URLs**: Builder pattern with compile-time safety for URL Scheme
6. **Minimal Dependencies**: Only essential dependencies (SQLite driver, testing)

## Non-Goals

1. Write operations to the Things 3 database
2. HTTP/gRPC service wrapper (library only)
3. Cross-platform support (macOS only, matching Things 3 availability)
4. Real-time sync or change detection
5. x-callback-url response handling

## Core Decisions

**Module Path**: `github.com/moond4rk/things3`

**SQLite Driver**: `github.com/mattn/go-sqlite3`
- CGO-based native SQLite implementation
- Optimal performance for database operations
- Most widely used and battle-tested Go SQLite driver
- Appropriate choice since Things 3 is macOS-only

**Entry Points**:
- `NewDB()` - Database operations (stateful, requires connection)
- `NewScheme(opts...)` - URL building and execution (configurable via `SchemeOption`)

**Token Handling**: `WithToken()` pattern
- Token required upfront for update operations
- Compile-time enforcement via separate `AuthScheme` type
- IDE autocomplete shows only valid methods

## Database Compatibility

The library reads from the Things 3 SQLite database located at:
- Default: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`
- Override via `THINGSDB` environment variable or `WithDBPath()` option

## Type System

### Database Enums

Integer-based enums for direct database mapping:

| Type | Values |
|------|--------|
| TaskType | 0=to-do, 1=project, 2=heading |
| Status | 0=incomplete, 2=canceled, 3=completed |
| StartBucket | 0=Inbox, 1=Anytime, 2=Someday |

### URL Scheme Enums

String-based enums for URL parameters:

| Type | Values |
|------|--------|
| When | today, tomorrow, evening, anytime, someday |
| ListID | inbox, today, anytime, upcoming, someday, logbook, etc. |
| Command | add, add-project, update, update-project, show, search, version, json |

## Dependencies

| Package | Purpose |
|---------|---------|
| github.com/mattn/go-sqlite3 | SQLite database driver (CGO) |
| github.com/stretchr/testify | Testing assertions (dev only) |

## Quality Requirements

- All exported types and functions must have Go doc comments
- Table-driven tests for query building and URL generation
- Integration tests using test database in testdata/
- golangci-lint compliance with project configuration
- English only for code, comments, documentation

## References

- Python Source: https://github.com/thingsapi/things.py
- Things URL Scheme: https://culturedcode.com/things/support/articles/2803573/
