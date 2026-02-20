# RFC 001: Project Overview

Status: Implemented
Author: @moond4rk

## Summary

things3 is a Go library providing read-only access to the Things 3 macOS application's SQLite database and type-safe URL Scheme building and execution. It is a complete port of the Python things.py library with full API parity, plus a modern URL builder API with background execution support.

## Architecture

All operations go through a single unified `NewClient()` entry point:

```
things3 library
|
+-- NewClient(opts...)  -> *Client  (Unified entry point)
    |
    +-- Options: WithDatabasePath(), WithPrintSQL(),
    |            WithForeground(), WithBackground(), WithPreloadToken()
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

### Design Principles

| Principle | Implementation |
|-----------|----------------|
| Single Entry Point | `NewClient()` for all operations |
| Interface-Based API | Methods return interfaces, not concrete types |
| Automatic Token Management | Lazy loading for update operations |
| Functional Options | `ClientOption` for configurable behavior |
| Builder Pattern | Chainable methods with terminal `.Build()` or `.Execute()` |
| Context-First | All queries and execution accept `context.Context` |

## RFC Index

| RFC | Title | Status | Description |
|-----|-------|--------|-------------|
| 001 | Project Overview | Implemented | Goals, architecture, dependencies (this document) |
| 002 | Database Schema | Implemented | Things 3 SQLite schema and SQL patterns |
| 003 | Database API | Implemented | Internal `db` type, query builders, convenience methods |
| 004 | URL Scheme | Implemented | Internal `scheme` type, URL builders, official reference |
| 005 | Unified Client | Implemented | `NewClient()`, single entry point, token management |
| 006 | Interface Abstraction | Implemented | Public interfaces, hide implementation details |

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

**Entry Point**: `NewClient(opts...)` - Unified access to all operations
- Combines internal `db` and `scheme` types behind a single API
- All database queries and URL scheme operations through one entry point
- Functional options: `WithDatabasePath()`, `WithPrintSQL()`, `WithForeground()`, `WithBackground()`, `WithPreloadToken()`

**Token Handling**: Automatic lazy loading
- Client fetches auth token from database on first update operation
- Cached with `sync.Mutex` for thread safety (allows retry on transient failures)
- Optional eager loading via `WithPreloadToken()`

## Database Compatibility

The library reads from the Things 3 SQLite database located at:
- Default: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`
- Override via `THINGSDB` environment variable or `WithDatabasePath()` option

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
