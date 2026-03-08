# RFC 008: Domain Model Redesign

Status: Draft
Author: @moond4rk
Date: 2026-02-24

## Summary

Introduce a clean domain model layer between the SQLite database and the public API, with physical separation via `internal/` sub-packages. The core changes are: (1) move database and SQL implementation into `internal/db`, (2) separate domain types for Todo, Project, Heading, Area, Tag, and ChecklistItem with consistent typing, (3) a conversion layer that transforms raw database rows into domain types, (4) move all type conversion from SQL to Go, and (5) a unified naming convention across the entire codebase.

## Motivation

### Current Issues

1. **Union type confusion**: `Task` represents three distinct domain concepts (todo, project, heading) with fields that are only valid for certain types. Users must check `.IsTodo()` before accessing `.Checklist`, and `.IsProject()` before accessing `.Items`.

2. **Inconsistent typing**: The same concept uses different types across structs:
   - `Task.Status` is `Status` (enum), but `ChecklistItem.Status` is `string`
   - `Task.Start` is `string`, but `StartBucket` enum exists and is unused in the model
   - `Area.Type` / `Tag.Type` / `ChecklistItem.Type` are hardcoded strings that duplicate Go type information

3. **Split conversion responsibility**: Type conversion is scattered between SQL (CASE expressions, printf bit manipulation) and Go (scanTask string-to-enum). Neither layer does a complete job.

4. **Flat file structure**: All 29 source files live in the root package. Internal implementation details (SQL building, filter primitives, date conversion, database connection) are mixed with public API types and interfaces.

5. **Type-unsafe containers**: `Tag.Items` is `[]any`, requiring type assertions at every use site.

6. **Inconsistent naming**: The same concept appears as "to-do", "todo", "Todo", and "Task" across SQL strings, Go types, JSON output, and comments. No canonical naming convention is defined.

### Goals

- Each domain concept has its own Go type with only relevant fields
- All type conversion happens in one place (Go conversion layer)
- SQL only does data retrieval, not formatting
- Internal implementation is physically separated from public API via `internal/`
- Zero use of `any` in public model types

## Naming Convention

A unified naming convention across all layers of the codebase. The canonical term is **"todo"** (not "to-do", not "task" for the todo concept).

### Naming Taxonomy

| Layer | Naming | Examples |
|-------|--------|----------|
| Database | Raw integers | `type = 0`, `status = 0`, `start = 0` |
| SQL constants | Internal, descriptive | `filterIsTodo = "type = 0"` |
| Internal types | Unexported Go | `taskType`, `taskTypeTodo` |
| Internal strings | `"todo"` | `taskTypeStringTodo = "todo"` (not `"to-do"`) |
| Domain structs | Exported Go | `Todo`, `Project`, `Heading`, `Area`, `Tag`, `ChecklistItem` |
| Query interfaces | Exported Go | `TodoQueryBuilder`, `ProjectQueryBuilder` |
| URL scheme interfaces | Exported Go | `TodoAdder`, `TodoUpdater`, `ProjectAdder`, `ProjectUpdater` |
| Comments and docs | `todo` everywhere | `// todo item`, `a todo in Things 3` |

### Key Decisions

1. **`"todo"` not `"to-do"`**: All string representations use `"todo"` without hyphen. The database stores integer `0`, not a string. The previous `"to-do"` string existed only in our SQL CASE expressions and will be eliminated.

2. **`taskType` is unexported**: After the type split, `TaskType` becomes `taskType` (unexported). Users distinguish types through Go's type system (`Todo`, `Project`, `Heading`), not through an enum. The enum remains internally for scan-layer routing.

3. **`client.Tasks()` is removed**: No generic "task" query entry point. Users choose `client.Todos()` or `client.Projects()` explicitly. Mixed-type results come from convenience methods with appropriate return types (TBD).

4. **`Todos()` is a builder entry point**: `client.Todos()` returns `TodoQueryBuilder` (no ctx parameter). The convenience shorthand `client.Todos(ctx)` (returning `[]Todo` directly) is removed. Use `client.Todos().All(ctx)` instead.

5. **Consistent `"todo"` in all contexts**: Comments, doc strings, error messages, CLI output, and JSON serialization all use `"todo"`, never `"to-do"` or `"task"` (when referring to the todo concept specifically).

### Before / After

```
Before                              After
------                              -----
taskTypeStringTodo = "to-do"        taskTypeStringTodo = "todo"
TaskType (exported)                 taskType (unexported)
TaskTypeTodo (exported)             taskTypeTodo (unexported)
Task struct (union type)            Todo, Project, Heading structs
Task.Type TaskType                  (removed, Go type is the type)
client.Tasks() TaskQueryBuilder     (removed)
client.Todos(ctx) []Task            client.Todos() TodoQueryBuilder
"a to-do item" (in comments)       "a todo item"
29 files in root package            root (~18) + internal/ (~11)
```

## Design

### File Structure

Physical separation of public API and internal implementation:

```
things3/
|
| -- Public API (user-facing) --------------------------
|-- doc.go                # package documentation
|-- client.go             # Client, NewClient
|-- client_options.go     # ClientOption
|-- models.go             # Todo, Project, Heading, Area, Tag, ChecklistItem
|-- types.go              # Status, StartBucket, Command, ListID
|-- interfaces.go         # all exported interfaces
|-- errors.go             # ErrTaskNotFound...
|-- time_helpers.go       # DaysAgo() etc.
|
| -- Conversion Layer ------------------------------------
|-- convert.go            # RawTask -> Todo/Project/Heading, RawArea -> Area...
|
| -- Query Builders (implement exported interfaces) ------
|-- query.go              # todoQuery, projectQuery
|-- query_area.go         # areaQuery
|-- query_tag.go          # tagQuery
|-- query_filter.go       # typeFilter, statusFilter, startFilter, dateFilter
|-- convenience.go        # Inbox(), Today(), Logbook()...
|
| -- URL Scheme Builders (implement exported interfaces) -
|-- scheme_builder.go     # addTodoBuilder, addProjectBuilder
|-- scheme_update.go      # updateTodoBuilder, updateProjectBuilder
|-- scheme_show.go        # showBuilder
|-- scheme_json.go        # batchBuilder
|
| -- Internal: Database Layer ----------------------------
|-- internal/
|   |-- db/
|   |   |-- db.go         # DB connection, execute, scan raw rows
|   |   |-- raw.go        # RawTask, RawArea, RawTag, RawChecklistItem
|   |   |-- sql.go        # SQL building (buildTasksSQL...)
|   |   |-- filter.go     # FilterBuilder, all filter primitives
|   |   |-- constants.go  # table names, column names, filter expressions
|   |   |-- date.go       # Things binary date <-> time.Time
|   |   +-- options.go    # DB options
|   |
|   +-- scheme/
|       |-- scheme.go     # URL execution (open command)
|       |-- attrs.go      # URL parameter building
|       |-- constants.go  # URL scheme constants
|       +-- options.go    # scheme options
|
| -- Test Support ----------------------------------------
|-- thingstest/           # public test utilities
|   +-- thingstest.go
|-- testdata/
|   +-- main.sqlite
+-- *_test.go
```

Files moved from root to `internal/`:

| Current file | Moved to | Reason |
|-------------|----------|--------|
| `database.go` | `internal/db/db.go` | connection, path discovery |
| `db.go` | `internal/db/db.go` | execute, scan (merged) |
| `db_options.go` | `internal/db/options.go` | internal options |
| `sql.go` | `internal/db/sql.go` | SQL building |
| `filter.go` | `internal/db/filter.go` | filter primitives |
| `constants.go` | `internal/db/constants.go` | table/column names |
| `date.go` | `internal/db/date.go` | binary date conversion |
| `scheme.go` | `internal/scheme/scheme.go` | URL execution |
| `scheme_attrs.go` | `internal/scheme/attrs.go` | URL params |
| `scheme_constants.go` | `internal/scheme/constants.go` | URL constants |
| `scheme_options.go` | `internal/scheme/options.go` | scheme options |

Root package: **29 files -> ~18 files**. Internal implementation hidden in `internal/`.

### Layer Architecture

```
SQLite Database (TMTask, TMArea, TMTag, TMChecklistItem)
        |
        v
    internal/db             SQL building + execution (raw columns, no CASE/printf)
        |
        v
    internal/db.RawTask     raw struct, maps 1:1 to database row
        |
        v
    things3/convert.go      rawTaskToTodo(), rawTaskToProject(), rawTaskToHeading()
        |                   int -> enum, binary date -> time.Time, flat -> grouped
        v
    things3.Todo/Project    exported domain types (user-facing)
```

### Domain Types

Relationship fields use flat `string` values (empty string = no relationship, `omitempty` hides in JSON). No separate reference type is needed since users only read these fields for display or lookup.

```go
// Todo represents an actionable todo item.
type Todo struct {
    UUID   string       `json:"uuid"`
    Title  string       `json:"title"`
    Status Status       `json:"status"`
    Notes  string       `json:"notes,omitempty"`
    Start  *StartBucket `json:"start,omitempty"`

    // Relationships (empty string = no relationship)
    AreaUUID     string `json:"area_uuid,omitempty"`
    AreaTitle    string `json:"area_title,omitempty"`
    ProjectUUID  string `json:"project_uuid,omitempty"`
    ProjectTitle string `json:"project_title,omitempty"`
    HeadingUUID  string `json:"heading_uuid,omitempty"`
    HeadingTitle string `json:"heading_title,omitempty"`

    // Dates
    StartDate    *time.Time `json:"start_date,omitempty"`
    Deadline     *time.Time `json:"deadline,omitempty"`
    ReminderTime *time.Time `json:"reminder_time,omitempty"`
    StopDate     *time.Time `json:"stop_date,omitempty"`
    Created      time.Time  `json:"created"`
    Modified     time.Time  `json:"modified"`

    // Nested items
    Tags      []string        `json:"tags,omitempty"`
    Checklist []ChecklistItem `json:"checklist,omitempty"`

    Trashed bool `json:"trashed,omitempty"`
}

// Project represents a container for organizing todos.
type Project struct {
    UUID   string       `json:"uuid"`
    Title  string       `json:"title"`
    Status Status       `json:"status"`
    Notes  string       `json:"notes,omitempty"`
    Start  *StartBucket `json:"start,omitempty"`

    // Relationships
    AreaUUID  string `json:"area_uuid,omitempty"`
    AreaTitle string `json:"area_title,omitempty"`

    // Dates
    StartDate *time.Time `json:"start_date,omitempty"`
    Deadline  *time.Time `json:"deadline,omitempty"`
    StopDate  *time.Time `json:"stop_date,omitempty"`
    Created   time.Time  `json:"created"`
    Modified  time.Time  `json:"modified"`

    // Nested items
    Tags  []string `json:"tags,omitempty"`
    Items []Todo   `json:"items,omitempty"`

    Trashed bool `json:"trashed,omitempty"`
}

// Heading represents a grouping label within a project.
// Headings are organizational only and have no status or dates.
type Heading struct {
    UUID  string `json:"uuid"`
    Title string `json:"title"`

    // Parent
    ProjectUUID  string `json:"project_uuid,omitempty"`
    ProjectTitle string `json:"project_title,omitempty"`

    // Nested items
    Todos []Todo `json:"todos,omitempty"`
}

// Area represents a high-level responsibility area.
type Area struct {
    UUID  string `json:"uuid"`
    Title string `json:"title"`

    Tags     []string  `json:"tags,omitempty"`
    Todos    []Todo    `json:"todos,omitempty"`
    Projects []Project `json:"projects,omitempty"`
}

// Tag represents a label for categorizing items.
type Tag struct {
    UUID     string `json:"uuid"`
    Title    string `json:"title"`
    Shortcut string `json:"shortcut,omitempty"`

    Todos    []Todo    `json:"todos,omitempty"`
    Projects []Project `json:"projects,omitempty"`
    Areas    []Area    `json:"areas,omitempty"`
}

// ChecklistItem represents a sub-item within a todo.
type ChecklistItem struct {
    UUID     string     `json:"uuid"`
    Title    string     `json:"title"`
    Status   Status     `json:"status"`
    StopDate *time.Time `json:"stop_date,omitempty"`
    Created  time.Time  `json:"created"`
    Modified time.Time  `json:"modified"`
}
```

Key differences from current `Task`:

| Field | Current Task | Todo | Project | Heading |
|-------|-------------|------|---------|---------|
| Checklist | always present | yes | no | no |
| Items/Todos | always present | no | yes | yes |
| ReminderTime | always present | yes | no | no |
| ProjectUUID | always present | yes | no | no |
| HeadingUUID | always present | yes | no | no |
| Status | always present | yes | yes | no |
| Index/TodayIndex | exposed | internal only | internal only | internal only |
| Type field | `TaskType` enum | not needed | not needed | not needed |
| Relationship fields | `*string` (nullable) | `string` (omitempty) | `string` (omitempty) | `string` (omitempty) |

### Internal Raw Types

Located in `internal/db/raw.go`. Map 1:1 to database rows with raw values:

```go
package db

// RawTask maps 1:1 to a TMTask row. All values are raw database types.
type RawTask struct {
    UUID             string
    Type             int       // 0=todo, 1=project, 2=heading
    Status           int       // 0=incomplete, 2=canceled, 3=completed
    Title            string
    Notes            string
    Start            int       // 0=inbox, 1=anytime, 2=someday
    Trashed          bool

    // Relationships (raw UUIDs and titles from JOINs)
    AreaUUID         *string
    AreaTitle        *string
    ProjectUUID      *string
    ProjectTitle     *string
    HeadingUUID      *string
    HeadingTitle     *string

    // Dates (raw database values, no conversion)
    StartDate        *int64    // Things binary date format
    Deadline         *int64    // Things binary date format
    ReminderTime     *int64    // Things binary time format
    StopDate         *int64    // Unix timestamp
    CreationDate     int64     // Unix timestamp
    ModificationDate int64     // Unix timestamp

    // Ordering (internal use only, not exposed in domain types)
    Index            int
    TodayIndex       int

    // Presence flags (from SQL JOINs, for lazy loading)
    HasTags          bool
    HasChecklist     bool
}

// RawArea maps to a TMArea row.
type RawArea struct {
    UUID    string
    Title   string
    HasTags bool
}

// RawTag maps to a TMTag row.
type RawTag struct {
    UUID     string
    Title    string
    Shortcut string
}

// RawChecklistItem maps to a TMChecklistItem row.
type RawChecklistItem struct {
    UUID             string
    Title            string
    Status           int       // raw integer
    StopDate         *int64    // Unix timestamp
    CreationDate     int64
    ModificationDate int64
}
```

### Conversion Layer

Located in `things3/convert.go`. Bridges `internal/db` raw types and public domain types:

```go
package things3

import "github.com/moond4rk/things3/internal/db"

func rawTaskToTodo(raw db.RawTask) Todo {
    return Todo{
        UUID:         raw.UUID,
        Title:        raw.Title,
        Status:       Status(raw.Status),
        Notes:        raw.Notes,
        Start:        toStartBucketPtr(raw.Start),
        AreaUUID:     ptrToString(raw.AreaUUID),
        AreaTitle:    ptrToString(raw.AreaTitle),
        ProjectUUID:  ptrToString(raw.ProjectUUID),
        ProjectTitle: ptrToString(raw.ProjectTitle),
        HeadingUUID:  ptrToString(raw.HeadingUUID),
        HeadingTitle: ptrToString(raw.HeadingTitle),
        StartDate:    db.ThingsDateToTime(raw.StartDate),
        Deadline:     db.ThingsDateToTime(raw.Deadline),
        ReminderTime: db.ThingsTimeToTime(raw.ReminderTime),
        StopDate:     unixToTimePtr(raw.StopDate),
        Created:      time.Unix(raw.CreationDate, 0),
        Modified:     time.Unix(raw.ModificationDate, 0),
        Trashed:      raw.Trashed,
    }
}

func rawTaskToProject(raw db.RawTask) Project {
    return Project{
        UUID:      raw.UUID,
        Title:     raw.Title,
        Status:    Status(raw.Status),
        Notes:     raw.Notes,
        Start:     toStartBucketPtr(raw.Start),
        AreaUUID:  ptrToString(raw.AreaUUID),
        AreaTitle: ptrToString(raw.AreaTitle),
        StartDate: db.ThingsDateToTime(raw.StartDate),
        Deadline:  db.ThingsDateToTime(raw.Deadline),
        StopDate:  unixToTimePtr(raw.StopDate),
        Created:   time.Unix(raw.CreationDate, 0),
        Modified:  time.Unix(raw.ModificationDate, 0),
        Trashed:   raw.Trashed,
    }
}

func rawTaskToHeading(raw db.RawTask) Heading {
    return Heading{
        UUID:         raw.UUID,
        Title:        raw.Title,
        ProjectUUID:  ptrToString(raw.ProjectUUID),
        ProjectTitle: ptrToString(raw.ProjectTitle),
    }
}

func rawAreaToArea(raw db.RawArea) Area {
    return Area{UUID: raw.UUID, Title: raw.Title}
}

func rawTagToTag(raw db.RawTag) Tag {
    return Tag{UUID: raw.UUID, Title: raw.Title, Shortcut: raw.Shortcut}
}

func rawChecklistToItem(raw db.RawChecklistItem) ChecklistItem {
    return ChecklistItem{
        UUID:     raw.UUID,
        Title:    raw.Title,
        Status:   Status(raw.Status),
        StopDate: unixToTimePtr(raw.StopDate),
        Created:  time.Unix(raw.CreationDate, 0),
        Modified: time.Unix(raw.ModificationDate, 0),
    }
}

// Helper functions
func ptrToString(p *string) string {
    if p == nil {
        return ""
    }
    return *p
}

func toStartBucketPtr(raw int) *StartBucket {
    s := StartBucket(raw)
    return &s
}

func unixToTimePtr(ts *int64) *time.Time {
    if ts == nil {
        return nil
    }
    t := time.Unix(*ts, 0)
    return &t
}
```

### SQL Simplification

Remove all CASE expressions and format conversions from SQL. Return raw database values.

Before:
```sql
CASE
    WHEN TASK.type = 0 THEN 'to-do'
    WHEN TASK.type = 1 THEN 'project'
    WHEN TASK.type = 2 THEN 'heading'
END AS type,
CASE
    WHEN TASK.status = 0 THEN 'incomplete'
    WHEN TASK.status = 2 THEN 'canceled'
    WHEN TASK.status = 3 THEN 'completed'
END AS status,
CASE
    WHEN TASK.start = 0 THEN 'Inbox'
    WHEN TASK.start = 1 THEN 'Anytime'
    WHEN TASK.start = 2 THEN 'Someday'
END AS start,
CASE WHEN TASK.startDate THEN
    printf('%d-%02d-%02d',
        (TASK.startDate & 134152192) >> 16,
        (TASK.startDate & 61440) >> 12,
        (TASK.startDate & 3968) >> 7)
ELSE TASK.startDate END AS start_date,
datetime(TASK.creationDate, 'unixepoch', 'localtime') AS created,
```

After:
```sql
TASK.type,
TASK.status,
TASK.start,
TASK.startDate AS start_date,
TASK.creationDate AS created,
```

All conversion moves to `convert.go` using existing functions (`ThingsDateToTime()`, `time.Unix()`, etc.).

### API Design

```go
client, _ := things3.NewClient()
defer client.Close()

// --- Builder queries (primary API) ---
// Todos() and Projects() are builder entry points.
// No client.Tasks() -- users choose the type explicitly.

todos, _ := client.Todos().All(ctx)                        // []Todo
todos, _ := client.Todos().InArea(uuid).All(ctx)           // []Todo
todos, _ := client.Todos().Status().Completed().All(ctx)   // []Todo
projects, _ := client.Projects().All(ctx)                  // []Project
projects, _ := client.Projects().InArea(uuid).All(ctx)     // []Project

// Area and Tag queries
areas, _ := client.Areas().All(ctx)                        // []Area
tags, _ := client.Tags().All(ctx)                          // []Tag

// --- Convenience methods (common use cases) ---
// Return types for mixed-type convenience methods are TBD.
// The exact container type will be decided in a later phase.
items, _ := client.Inbox(ctx)
items, _ := client.Today(ctx)
items, _ := client.Logbook(ctx)
```

### Type Consistency Fixes

| Type | Before | After |
|------|--------|-------|
| `TaskType` | exported enum | `taskType` unexported (Go types replace enum) |
| `taskTypeStringTodo` | `"to-do"` | `"todo"` |
| `ChecklistItem.Status` | `string` | `Status` enum |
| `Task.Start` -> `Todo.Start` | `string` | `*StartBucket` enum |
| `Area.Type` | `string` (always "area") | removed (Go type is the type) |
| `Tag.Type` | `string` (always "tag") | removed |
| `ChecklistItem.Type` | `string` (always "checklist-item") | removed |
| `Tag.Items` | `[]any` | `Todos []Todo` + `Projects []Project` + `Areas []Area` |
| `StartBucket` JSON | no marshaler | `MarshalJSON` / `UnmarshalJSON` (like `Status`) |
| `Task.Index` / `Task.TodayIndex` | exposed in public model | internal only (in `RawTask`) |
| Relationship fields | `*string` (nullable pointer) | `string` with `omitempty` |

## Decided Questions

1. **Heading as a type vs inline**: Heading is a standalone domain type. In flat queries, each Todo carries `HeadingUUID`/`HeadingTitle` to indicate its grouping. In Project items, Heading grouping is expressed through the Todo's heading fields (flat), not through nested `Heading` containers. No independent `client.Headings()` builder since headings only exist within projects. **Decision: standalone type, flat items in Project.**

2. **Builder separation**: `Todos()` and `Projects()` have separate public interfaces (`TodoQueryBuilder`, `ProjectQueryBuilder`) with typed return values. Internally they share the same `taskQuery` implementation -- the builder auto-sets the type filter and converts results via `rawTaskToTodo()`/`rawTaskToProject()` in terminal methods (`All`, `First`, `Count`). **Decision: separate interfaces, shared implementation.**

3. **JSON serialization**: The `Type` fields (`Area.Type`, `Tag.Type`, `ChecklistItem.Type`) are removed. No JSON compatibility with Python things.py is maintained. Go types themselves express the type information. **Decision: drop Type fields.**

4. **File structure**: Internal implementation (database, SQL, filters, date conversion, URL execution) moves to `internal/` sub-packages. Public API stays in root. **Decision: `internal/db` + `internal/scheme`.**

5. **Relationship fields**: Use flat `string` fields with `omitempty` instead of a separate reference type. Users access fields directly (`todo.AreaTitle`), no nil check needed (empty string = no relationship). **Decision: flat string fields, no Ref type.**

6. **Index/TodayIndex**: Internal ordering fields used for SQL `ORDER BY` and convenience method sorting. Not exposed in public domain types. Stay in `RawTask` only. **Decision: internal only.**

7. **Trashed**: Kept in public domain types (at the end of struct, with `omitempty`). Useful for `Get()` results where query context is not available. **Decision: keep, low priority placement.**

8. **Start field**: `*StartBucket` exposed in public domain types. Useful for display (e.g., CLI showing `[Inbox]` / `[Someday]` labels). **Decision: keep.**

## Open Questions

1. **Mixed-type convenience method returns**: Convenience methods like `Inbox()`, `Today()`, `Logbook()`, `Trash()` currently return `[]Task` (mixed todos and projects). After the type split, they need a container type. A candidate is `Items { Todos []Todo; Projects []Project }`, but the exact naming and structure are TBD. This will be decided when implementing the public API layer. **Deferred.**

2. **Structured hierarchical results**: `TodayResult`, `SearchResult`, `AreaGroup`, `ProjectGroup` for hierarchical views (Area > Project > Todo). Not in scope for the initial implementation. Can be added as a separate RFC when there is concrete demand. **Deferred to future RFC.**

## Implementation Notes

### Phase 1: Internal Restructure (no API changes)

Move internal implementation to `internal/` sub-packages:

- Create `internal/db/`: move database connection, SQL building, filter primitives, date conversion, constants, scan functions
- Create `internal/scheme/`: move URL execution, parameter building, scheme constants
- Define `RawTask`, `RawArea`, `RawTag`, `RawChecklistItem` in `internal/db/raw.go`
- Simplify SQL to return raw values (remove all CASE expressions and printf)
- Root package wraps `internal/db` for query execution
- Public API stays the same (still returns `Task`) -- this is a pure internal refactor

### Phase 2: Type Split and Conversion Layer

- Add `convert.go` with `rawTaskToTodo()`, `rawTaskToProject()`, `rawTaskToHeading()` etc.
- Add `Todo`, `Project`, `Heading` as exported types in `models.go`
- Fix `ChecklistItem.Status` to use `Status` enum
- Add `MarshalJSON`/`UnmarshalJSON` to `StartBucket`
- Introduce unexported `taskType` (demote from `TaskType`)
- Change `taskTypeStringTodo` from `"to-do"` to `"todo"`
- Remove `client.Tasks()` entry point
- Add typed builder entry points: `client.Todos()` -> `TodoQueryBuilder`, `client.Projects()` -> `ProjectQueryBuilder`
- Update all comments and docs to use "todo" (not "to-do")

### Phase 3: Cleanup

- Remove exported `Task` type (replaced by Todo/Project/Heading)
- Remove exported `TaskType` enum (replaced by unexported `taskType`)
- Remove redundant `Type` fields from Area, Tag, ChecklistItem
- Remove `Index`/`TodayIndex` from public model
- Split `Tag.Items []any` into typed fields (`Todos`, `Projects`, `Areas`)
- Change relationship fields from `*string` to `string` with `omitempty`
- Update all tests
