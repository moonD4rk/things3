# RFC 008: Domain Model Redesign

Status: Draft
Author: @moond4rk
Date: 2026-02-24

## Summary

Introduce a clean domain model layer between the SQLite database and the public API. The core changes are: (1) separate domain types for Todo, Project, Area, Tag, and ChecklistItem with consistent typing, (2) move all type conversion from SQL to Go, (3) structured hierarchical results for queries that return mixed types, and (4) a unified naming convention across the entire codebase.

## Motivation

### Current Issues

1. **Union type confusion**: `Task` represents three distinct domain concepts (todo, project, heading) with fields that are only valid for certain types. Users must check `.IsTodo()` before accessing `.Checklist`, and `.IsProject()` before accessing `.Items`.

2. **Inconsistent typing**: The same concept uses different types across structs:
   - `Task.Status` is `Status` (enum), but `ChecklistItem.Status` is `string`
   - `Task.Start` is `string`, but `StartBucket` enum exists and is unused in the model
   - `Area.Type` / `Tag.Type` / `ChecklistItem.Type` are hardcoded strings that duplicate Go type information

3. **Split conversion responsibility**: Type conversion is scattered between SQL (CASE expressions, printf bit manipulation) and Go (scanTask string-to-enum). Neither layer does a complete job.

4. **Flat results hide structure**: Things 3 data is naturally hierarchical (Area > Project > Todo), but all queries return flat `[]Task` lists, forcing users to reconstruct the hierarchy themselves.

5. **Type-unsafe containers**: `Tag.Items` is `[]any`, requiring type assertions at every use site.

6. **Inconsistent naming**: The same concept appears as "to-do", "todo", "Todo", and "Task" across SQL strings, Go types, JSON output, and comments. No canonical naming convention is defined.

### Goals

- Each domain concept has its own Go type with only relevant fields
- All type conversion happens in one place (Go scan layer)
- SQL only does data retrieval, not formatting
- Queries can return both flat typed results and structured hierarchical results
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

3. **`client.Tasks()` is removed**: No generic "task" query entry point. Users choose `client.Todos()` or `client.Projects()` explicitly. Mixed-type results come from structured result types like `SearchResult`.

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
```

## Design

### Layer Architecture

```
SQLite Database (TMTask, TMArea, TMTag, TMChecklistItem)
        |
        v
    SQL Layer          SELECT raw columns (no CASE, no printf)
        |
        v
    Scan Layer         scanTask() reads raw values, converts to Go types
        |
        v
    Internal task      unexported struct, maps 1:1 to TMTask row
        |
        v
    Conversion         taskToTodo(), taskToProject(), taskToHeading()
        |
        v
    Domain Types       Todo, Project, Heading, Area, Tag, ChecklistItem (exported)
```

### Domain Types

```go
// Todo represents an actionable todo item.
type Todo struct {
    UUID   string
    Title  string
    Status Status
    Notes  string
    Start  *StartBucket

    // Relationships
    AreaUUID     *string
    AreaTitle    *string
    ProjectUUID  *string
    ProjectTitle *string
    HeadingUUID  *string
    HeadingTitle *string

    // Dates
    StartDate    *time.Time
    Deadline     *time.Time
    ReminderTime *time.Time   // Only Todo has reminders
    StopDate     *time.Time
    Created      time.Time
    Modified     time.Time

    // Nested items
    Tags      []string
    Checklist []ChecklistItem  // Only Todo has checklist

    // Ordering
    Index      int
    TodayIndex int

    Trashed bool
}

// Project represents a container for organizing todos.
type Project struct {
    UUID   string
    Title  string
    Status Status
    Notes  string
    Start  *StartBucket

    // Relationships
    AreaUUID  *string
    AreaTitle *string

    // Dates
    StartDate *time.Time
    Deadline  *time.Time
    StopDate  *time.Time
    Created   time.Time
    Modified  time.Time

    // Nested items
    Tags  []string
    Todos []Todo      // Only Project has child todos

    // Ordering
    Index int

    Trashed bool
}

// Heading represents a grouping label within a project.
// Headings are organizational only and have no status or dates.
type Heading struct {
    UUID  string
    Title string

    // Parent
    ProjectUUID  *string
    ProjectTitle *string

    // Nested items
    Todos []Todo

    // Ordering
    Index int
}

// Area represents a high-level responsibility area.
type Area struct {
    UUID  string
    Title string
    Tags  []string
}

// Tag represents a label for categorizing items.
type Tag struct {
    UUID     string
    Title    string
    Shortcut string
    Todos    []Todo     // Tagged todos
    Projects []Project  // Tagged projects
    Areas    []Area     // Tagged areas
}

// ChecklistItem represents a sub-item within a todo.
type ChecklistItem struct {
    UUID     string
    Title    string
    Status   Status       // Same enum as Todo/Project, not string
    StopDate *time.Time
    Created  time.Time
    Modified time.Time
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
| Type field | `TaskType` enum (exported) | not needed (Go type is the type) | not needed | not needed |

### Structured Results

For queries that naturally return mixed types, provide hierarchical result types
that reflect the real Things 3 data hierarchy: Area > Project > Todo.

#### Data Relationships

A todo's area comes from two possible paths:

```
Path 1: Todo directly in Area (no project)
    Todo.AreaUUID = "area-1"
    Todo.ProjectUUID = nil

Path 2: Todo inherits Area from Project
    Todo.ProjectUUID = "proj-1"
    Project.AreaUUID = "area-1"    (area inherited through project)

Path 3: Todo with no Area and no Project (standalone)
    Todo.AreaUUID = nil
    Todo.ProjectUUID = nil
```

#### Hierarchical Result Types

```go
// AreaGroup contains all items within one area.
// Each area appears exactly once. Items are further grouped by project.
type AreaGroup struct {
    Area     *Area          // nil for items without an area
    Projects []ProjectGroup // projects that contain matching todos
    Todos    []Todo         // standalone todos directly in this area (no project)
}

// ProjectGroup contains a project and its matching todos.
type ProjectGroup struct {
    Project Project
    Todos   []Todo
}

// TodayResult contains today's items organized as Area > Project > Todo.
type TodayResult struct {
    Groups []AreaGroup
}

// SearchResult contains search matches organized by type.
type SearchResult struct {
    Todos    []Todo
    Projects []Project
    Areas    []Area
    Tags     []Tag
}
```

#### Concrete Example

Given today's items:

```
Area "Work"
+-- Project "Release v1.0"
|   +-- Todo "Write tests"
|   +-- Todo "Fix bug #123"
+-- Todo "Reply to email"          (standalone, directly in area)

Area "Personal"
+-- Todo "Buy groceries"           (standalone)

(No Area)
+-- Project "Side Project"
|   +-- Todo "Design logo"
+-- Todo "Random idea"             (standalone, no area, no project)
```

The `TodayResult` structure:

```go
TodayResult{
    Groups: []AreaGroup{
        {
            Area: &Area{Title: "Work"},
            Projects: []ProjectGroup{
                {
                    Project: Project{Title: "Release v1.0"},
                    Todos:   []Todo{
                        {Title: "Write tests"},
                        {Title: "Fix bug #123"},
                    },
                },
            },
            Todos: []Todo{
                {Title: "Reply to email"},
            },
        },
        {
            Area: &Area{Title: "Personal"},
            Projects: nil,
            Todos: []Todo{
                {Title: "Buy groceries"},
            },
        },
        {
            Area: nil,  // no area
            Projects: []ProjectGroup{
                {
                    Project: Project{Title: "Side Project"},
                    Todos:   []Todo{
                        {Title: "Design logo"},
                    },
                },
            },
            Todos: []Todo{
                {Title: "Random idea"},
            },
        },
    },
}
```

#### Traversal

```go
result, _ := client.Today(ctx)

for _, group := range result.Groups {
    if group.Area != nil {
        fmt.Printf("Area: %s\n", group.Area.Title)
    } else {
        fmt.Println("No Area")
    }

    for _, pg := range group.Projects {
        fmt.Printf("  Project: %s\n", pg.Project.Title)
        for _, todo := range pg.Todos {
            fmt.Printf("    - %s\n", todo.Title)
        }
    }

    for _, todo := range group.Todos {
        fmt.Printf("  - %s (standalone)\n", todo.Title)
    }
}

// Output:
// Area: Work
//   Project: Release v1.0
//     - Write tests
//     - Fix bug #123
//   - Reply to email (standalone)
// Area: Personal
//   - Buy groceries (standalone)
// No Area
//   Project: Side Project
//     - Design logo
//   - Random idea (standalone)
```

### API Design

```go
client, _ := things3.NewClient()
defer client.Close()

// --- Builder queries (primary API) ---
// Todos() and Projects() are builder entry points, not convenience methods.
// No client.Tasks() — users choose the type explicitly.

todos, _ := client.Todos().All(ctx)                        // []Todo - all incomplete
todos, _ := client.Todos().InArea(uuid).All(ctx)           // []Todo
todos, _ := client.Todos().Status().Completed().All(ctx)   // []Todo
projects, _ := client.Projects().All(ctx)                  // []Project
projects, _ := client.Projects().InArea(uuid).All(ctx)     // []Project

// Area and Tag queries
areas, _ := client.Areas().All(ctx)                        // []Area
tags, _ := client.Tags().All(ctx)                          // []Tag

// Search through builders
todos, _ := client.Todos().Search("keyword").All(ctx)      // []Todo
projects, _ := client.Projects().Search("keyword").All(ctx) // []Project
areas, _ := client.Areas().Search("keyword").All(ctx)       // []Area

// --- Convenience methods (common use cases) ---
// TBD: exact list of convenience methods and their return types.

result, _ := client.Today(ctx)                     // TodayResult (Area > Project > Todo)
for _, group := range result.Groups {
    // group.Area      - which area (nil = no area)
    // group.Projects  - projects with their todos
    // group.Todos     - standalone todos in this area
}

results, _ := client.Search(ctx, "keyword")        // SearchResult
// results.Todos, results.Projects, results.Areas, results.Tags
```

### Internal Implementation

The internal `task` type (unexported) maps directly to TMTask:

```go
// task is the internal database row representation.
// Not exported; converted to Todo/Project/Heading before returning to users.
type task struct {
    uuid     string
    typ      taskType     // unexported enum: taskTypeTodo, taskTypeProject, taskTypeHeading
    status   Status
    title    string
    notes    string
    start    *StartBucket
    // ... all fields from TMTask
}
```

Conversion functions:

```go
func taskToTodo(t task) Todo { ... }
func taskToProject(t task, todos []Todo) Project { ... }
func taskToHeading(t task, todos []Todo) Heading { ... }
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

All conversion moves to Go scan layer using existing functions (`thingsDateToTime()`, `time.Unix()`, etc.).

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

## Open Questions

These items require further discussion before implementation:

1. **Heading as a type vs inline**: Heading is a standalone domain type. Project has `Todos []Todo` (standalone) and `Headings []Heading` (each with its own Todos). Todo retains `HeadingUUID`/`HeadingTitle` for context in flat queries. No independent `client.Headings()` builder needed since headings only exist within projects. **Decision: standalone type.**

2. **Builder separation**: `Todos()` and `Projects()` have separate public interfaces (`TodoQueryBuilder`, `ProjectQueryBuilder`) with typed return values. Internally they share the same `taskQuery` implementation — the builder auto-sets the type filter and converts results via `taskToTodo()`/`taskToProject()` in terminal methods (`All`, `First`, `Count`). **Decision: separate interfaces, shared implementation.**

3. **Structured result scope**: Which convenience methods return structured results vs flat lists? Candidates for structured: `Today()`, `Search()`. Candidates for flat: `Inbox()`, `Logbook()`, `Trash()`.

4. **JSON serialization**: The `Type` fields (`Area.Type`, `Tag.Type`, `ChecklistItem.Type`) are removed. No JSON compatibility with Python things.py is maintained — things.py is an inspiration source, not an upstream dependency. Go types themselves express the type information. If CLI or other consumers need a `"type"` field in JSON output, it should be handled in the serialization layer, not in the domain model. **Decision: drop Type fields, no things.py compatibility.**

5. **Migration path**: Should this be done in one large refactor or incrementally? Incremental approach: (1) SQL simplification, (2) type consistency, (3) type split, (4) structured results.

## Implementation Notes

### Phase 1: Foundation (no API changes)
- Simplify SQL to return raw values (remove all CASE expressions)
- Unify scan layer to do all type conversion in Go
- Fix `ChecklistItem.Status` to use `Status` enum
- Add `MarshalJSON`/`UnmarshalJSON` to `StartBucket`
- Change `taskTypeStringTodo` from `"to-do"` to `"todo"`
- These changes are internal; public API stays the same temporarily

### Phase 2: Type Split and Naming
- Introduce unexported `task` and `taskType` as internal types
- Demote `TaskType` to unexported `taskType`
- Add `Todo`, `Project`, `Heading` as exported types
- Add conversion functions (`taskToTodo`, `taskToProject`, `taskToHeading`)
- Remove `client.Tasks()` entry point
- Add typed builder entry points: `client.Todos()` -> `TodoQueryBuilder`, `client.Projects()` -> `ProjectQueryBuilder`
- Update all comments and docs to use "todo" (not "to-do")

### Phase 3: Structured Results
- Add `AreaGroup`, `ProjectGroup`, `TodayResult`, `SearchResult` types
- Implement grouping logic (group todos by area, then by project within each area)
- Handle area inheritance: todos inherit area from their project when not set directly
- Update `Today()`, `Search()` to return structured results
- Add `Areas().Search()`, `Tags().Search()` methods

### Phase 4: Cleanup
- Remove exported `Task` type (replaced by Todo/Project/Heading)
- Remove exported `TaskType` enum (replaced by unexported `taskType`)
- Remove redundant `Type` fields from Area, Tag, ChecklistItem
- Split `Tag.Items []any` into typed fields (`Todos`, `Projects`, `Areas`)
- Update all tests
