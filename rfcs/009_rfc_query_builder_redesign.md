# RFC 009: Query Builder Redesign

Status: Draft
Author: @moond4rk
Date: 2026-03-08

## Summary

Redesign the query builder interfaces and domain model to replace the union `Task` type with separate typed builders and domain types. This RFC supersedes the query-related sections of RFC 008 and captures all finalized design decisions for domain types, query interfaces, and data loading strategy.

## Motivation

The current design uses a single `TaskQueryBuilder` that returns `[]Task` -- a union type representing todos, projects, and headings. Users must check `task.Type` before accessing type-specific fields. The new design provides:

- **Type-safe queries**: `client.Todos()` returns `[]Todo`, `client.Projects()` returns `[]Project`
- **Separate domain types**: Each concept has its own struct with only relevant fields
- **Flat data model**: No nested items; all relationships queried via builders
- **Generic sub-builders**: Shared filter logic with type-safe return types

## Domain Types

### Design Principles

1. **Flat structs**: Each struct contains only its own data. No nested `[]Todo` on Project, no `[]Project` on Area
2. **Parent references inline**: Lightweight `UUID + Title` for parent relationships (zero-cost, from SQL JOIN)
3. **Child queries via builder**: `client.Todos().InProject(uuid)` instead of `project.Todos`
4. **Value-type StartBucket**: Not a pointer; zero value (`StartInbox`) is meaningful
5. **Time naming**: Timestamps use `At` suffix (`CreatedAt`, `CompletedAt`); dates keep semantic names (`StartDate`, `Deadline`)

### Todo

```go
type Todo struct {
    UUID   string      `json:"uuid"`
    Title  string      `json:"title"`
    Status Status      `json:"status"`
    Notes  string      `json:"notes,omitempty"`
    Start  StartBucket `json:"start"`

    // Relationships (empty string = no relationship)
    AreaUUID     string `json:"area_uuid,omitempty"`
    AreaTitle    string `json:"area_title,omitempty"`
    ProjectUUID  string `json:"project_uuid,omitempty"`
    ProjectTitle string `json:"project_title,omitempty"`
    HeadingUUID  string `json:"heading_uuid,omitempty"`
    HeadingTitle string `json:"heading_title,omitempty"`

    // Attributes
    Tags      []string        `json:"tags,omitempty"`
    Checklist []ChecklistItem `json:"checklist,omitempty"`

    // Dates (date only, no time component)
    StartDate *time.Time `json:"start_date,omitempty"`
    Deadline  *time.Time `json:"deadline,omitempty"`

    // Time (time only, date component is zero value)
    Reminder *time.Time `json:"reminder,omitempty"`

    // Timestamps
    CreatedAt   time.Time  `json:"created_at"`
    ModifiedAt  time.Time  `json:"modified_at"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    CanceledAt  *time.Time `json:"canceled_at,omitempty"`

    Trashed bool `json:"trashed,omitempty"`
}
```

### Project

```go
type Project struct {
    UUID   string      `json:"uuid"`
    Title  string      `json:"title"`
    Status Status      `json:"status"`
    Notes  string      `json:"notes,omitempty"`
    Start  StartBucket `json:"start"`

    // Relationships
    AreaUUID  string `json:"area_uuid,omitempty"`
    AreaTitle string `json:"area_title,omitempty"`

    // Attributes
    Tags []string `json:"tags,omitempty"`

    // Dates (date only, no time component)
    StartDate *time.Time `json:"start_date,omitempty"`
    Deadline  *time.Time `json:"deadline,omitempty"`

    // Time (time only, date component is zero value)
    Reminder *time.Time `json:"reminder,omitempty"`

    // Timestamps
    CreatedAt   time.Time  `json:"created_at"`
    ModifiedAt  time.Time  `json:"modified_at"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    CanceledAt  *time.Time `json:"canceled_at,omitempty"`

    Trashed bool `json:"trashed,omitempty"`
}
```

### Heading

Headings are organizational only -- no status, dates, or notes.

```go
type Heading struct {
    UUID  string `json:"uuid"`
    Title string `json:"title"`

    // Parent project
    ProjectUUID  string `json:"project_uuid,omitempty"`
    ProjectTitle string `json:"project_title,omitempty"`
}
```

### Area

```go
type Area struct {
    UUID  string   `json:"uuid"`
    Title string   `json:"title"`
    Tags  []string `json:"tags,omitempty"`
}
```

### Tag

```go
type Tag struct {
    UUID     string `json:"uuid"`
    Title    string `json:"title"`
    Shortcut string `json:"shortcut,omitempty"`
}
```

### ChecklistItem

```go
type ChecklistItem struct {
    UUID   string `json:"uuid"`
    Title  string `json:"title"`
    Status Status `json:"status"`

    // Timestamps
    CreatedAt   time.Time  `json:"created_at"`
    ModifiedAt  time.Time  `json:"modified_at"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    CanceledAt  *time.Time `json:"canceled_at,omitempty"`
}
```

### Time Field Summary

All time-related fields across the domain types:

| Field | Type | Meaning | Source |
|-------|------|---------|--------|
| `StartDate` | `*time.Time` | Scheduled date (date only) | Things binary date |
| `Deadline` | `*time.Time` | Due date (date only) | Things binary date |
| `Reminder` | `*time.Time` | Reminder time (time only) | Things binary time |
| `CreatedAt` | `time.Time` | Creation timestamp | Unix timestamp |
| `ModifiedAt` | `time.Time` | Last modification timestamp | Unix timestamp |
| `CompletedAt` | `*time.Time` | Completion timestamp | Unix timestamp (when Status=completed) |
| `CanceledAt` | `*time.Time` | Cancellation timestamp | Unix timestamp (when Status=canceled) |

The database stores a single `stopDate` column. The conversion layer splits it into `CompletedAt` or `CanceledAt` based on the task's `Status` value.

### Relationship Model

All relationships follow the same pattern: parent references inline, child queries via builder.

| Relationship | Direction | Access Method |
|-------------|-----------|---------------|
| Todo -> Project | upward (parent) | `todo.ProjectUUID`, `todo.ProjectTitle` |
| Todo -> Area | upward (parent) | `todo.AreaUUID`, `todo.AreaTitle` |
| Todo -> Heading | upward (parent) | `todo.HeadingUUID`, `todo.HeadingTitle` |
| Project -> Area | upward (parent) | `project.AreaUUID`, `project.AreaTitle` |
| Heading -> Project | upward (parent) | `heading.ProjectUUID`, `heading.ProjectTitle` |
| Project -> Todos | downward (children) | `client.Todos().InProject(uuid)` |
| Area -> Todos | downward (children) | `client.Todos().InArea(uuid)` |
| Area -> Projects | downward (children) | `client.Projects().InArea(uuid)` |
| Heading -> Todos | downward (children) | `client.Todos().InHeading(uuid)` |
| Tag -> Todos | association | `client.Todos().InTag(title)` |
| Tag -> Projects | association | `client.Projects().InTag(title)` |

Upward references are populated from SQL JOINs at zero additional query cost. Downward and association queries require a separate builder call (local SQLite, microsecond latency).

## Query Builder Interfaces

### Client Entry Points

```go
client.Todos()    // -> TodoQueryBuilder
client.Projects() // -> ProjectQueryBuilder
client.Headings() // -> HeadingQueryBuilder
client.Areas()    // -> AreaQueryBuilder
client.Tags()     // -> TagQueryBuilder
```

No `client.Tasks()` -- users choose the type explicitly.

### Generic Sub-builder Interfaces

Sub-builders use Go generics to return the correct parent builder type, avoiding interface duplication.

```go
// StatusFilter provides type-safe status filtering.
type StatusFilter[T any] interface {
    Incomplete() T
    Completed() T
    Canceled() T
    Any() T
}

// StartFilter provides type-safe start bucket filtering.
type StartFilter[T any] interface {
    Inbox() T
    Anytime() T
    Someday() T
}

// DateFilter provides type-safe date filtering.
type DateFilter[T any] interface {
    Exists(has bool) T
    Future() T
    Past() T
    On(date time.Time) T
    Before(date time.Time) T
    OnOrBefore(date time.Time) T
    After(date time.Time) T
    OnOrAfter(date time.Time) T
}
```

### TodoQueryBuilder

```go
type TodoQueryBuilder interface {
    TodoQueryExecutor

    // Identity
    WithUUID(uuid string) TodoQueryBuilder

    // State filters
    Status() StatusFilter[TodoQueryBuilder]
    Start() StartFilter[TodoQueryBuilder]
    Trashed(trashed bool) TodoQueryBuilder
    ContextTrashed(trashed bool) TodoQueryBuilder

    // Relation filters
    InArea(uuid string) TodoQueryBuilder
    HasArea(has bool) TodoQueryBuilder
    InProject(uuid string) TodoQueryBuilder
    HasProject(has bool) TodoQueryBuilder
    InHeading(uuid string) TodoQueryBuilder
    HasHeading(has bool) TodoQueryBuilder
    InTag(title string) TodoQueryBuilder
    HasTag(has bool) TodoQueryBuilder

    // Date filters
    StartDate() DateFilter[TodoQueryBuilder]
    StopDate() DateFilter[TodoQueryBuilder]
    Deadline() DateFilter[TodoQueryBuilder]
    CreatedAfter(t time.Time) TodoQueryBuilder

    // Ordering & search
    Search(query string) TodoQueryBuilder
    OrderByTodayIndex() TodoQueryBuilder

    // Optional loading
    IncludeChecklist() TodoQueryBuilder
}
```

### ProjectQueryBuilder

```go
type ProjectQueryBuilder interface {
    ProjectQueryExecutor

    // Identity
    WithUUID(uuid string) ProjectQueryBuilder

    // State filters
    Status() StatusFilter[ProjectQueryBuilder]
    Start() StartFilter[ProjectQueryBuilder]
    Trashed(trashed bool) ProjectQueryBuilder

    // Relation filters
    InArea(uuid string) ProjectQueryBuilder
    HasArea(has bool) ProjectQueryBuilder
    InTag(title string) ProjectQueryBuilder
    HasTag(has bool) ProjectQueryBuilder

    // Date filters
    StartDate() DateFilter[ProjectQueryBuilder]
    StopDate() DateFilter[ProjectQueryBuilder]
    Deadline() DateFilter[ProjectQueryBuilder]
    CreatedAfter(t time.Time) ProjectQueryBuilder

    // Search
    Search(query string) ProjectQueryBuilder
}
```

### HeadingQueryBuilder

Minimal interface -- headings have no status, dates, or tags.

```go
type HeadingQueryBuilder interface {
    HeadingQueryExecutor

    // Identity
    WithUUID(uuid string) HeadingQueryBuilder

    // Relation filters
    InProject(uuid string) HeadingQueryBuilder
}
```

### AreaQueryBuilder

```go
type AreaQueryBuilder interface {
    AreaQueryExecutor

    WithUUID(uuid string) AreaQueryBuilder
    WithTitle(title string) AreaQueryBuilder
    Visible(visible bool) AreaQueryBuilder
    InTag(title string) AreaQueryBuilder
    HasTag(has bool) AreaQueryBuilder
}
```

### TagQueryBuilder

```go
type TagQueryBuilder interface {
    TagQueryExecutor

    WithUUID(uuid string) TagQueryBuilder
    WithTitle(title string) TagQueryBuilder
    WithParent(parentUUID string) TagQueryBuilder
}
```

## Data Loading Strategy

### Checklist Items

- `Todo.Checklist` field exists on the struct
- `All()` does NOT load checklist items (nil) for list-view performance
- `First()` automatically loads checklist items
- `IncludeChecklist()` explicitly opts in to loading for `All()` queries

```go
// List view -- no checklist
todos, _ := client.Todos().Status().Incomplete().All(ctx)

// Detail view -- checklist auto-loaded
todo, _ := client.Todos().WithUUID(uuid).First(ctx)

// List with checklist -- explicit opt-in
todos, _ := client.Todos().IncludeChecklist().All(ctx)
```

### Tags

Tags are always loaded when present (single additional query per task with tags, using the `HasTags` presence flag from the SQL JOIN).

### Parent References

Always populated from SQL JOINs. Zero additional query cost.

## Internal Implementation

### Shared taskQuery

All three typed builders (`todoQuery`, `projectQuery`, `headingQuery`) share a common `taskQuery` internally. The builder sets the type filter automatically and converts results via the appropriate conversion function.

```go
type taskQuery struct {
    database         *db
    filter           idb.TaskFilter
    includeChecklist bool
}

type todoQuery struct {
    inner *taskQuery
}

type projectQuery struct {
    inner *taskQuery
}

type headingQuery struct {
    inner *taskQuery
}
```

### Generic Sub-builder Implementation

Sub-builders use a shared generic struct with a pointer to the underlying `taskQuery` and a typed parent reference:

```go
type statusFilter[T any] struct {
    query  *taskQuery
    parent T
}

func (f *statusFilter[T]) Incomplete() T {
    v := int(StatusIncomplete)
    f.query.filter.Status = &v
    return f.parent
}

// Usage in todoQuery:
func (q *todoQuery) Status() StatusFilter[TodoQueryBuilder] {
    return &statusFilter[TodoQueryBuilder]{query: q.inner, parent: q}
}

// Usage in projectQuery:
func (q *projectQuery) Status() StatusFilter[ProjectQueryBuilder] {
    return &statusFilter[ProjectQueryBuilder]{query: q.inner, parent: q}
}
```

### Conversion Layer

The `stopDate` split happens in the conversion layer:

```go
func rawTaskToTodo(raw *idb.TaskRow) Todo {
    todo := Todo{
        UUID:       raw.UUID,
        Status:     Status(raw.Status),
        CreatedAt:  time.Unix(raw.CreationDate, 0),
        ModifiedAt: time.Unix(raw.ModificationDate, 0),
        // ... other fields
    }

    // Split stopDate into CompletedAt or CanceledAt based on Status
    if raw.StopDate != nil {
        t := time.Unix(*raw.StopDate, 0)
        switch Status(raw.Status) {
        case StatusCompleted:
            todo.CompletedAt = &t
        case StatusCanceled:
            todo.CanceledAt = &t
        }
    }

    return todo
}
```

## Usage Examples

```go
client, _ := things3.NewClient()
defer client.Close()

// Query todos
todos, _ := client.Todos().Status().Incomplete().All(ctx)
todo, _ := client.Todos().WithUUID(uuid).First(ctx)

// Query projects
projects, _ := client.Projects().InArea(areaUUID).All(ctx)

// Query todos in a project (two queries, both local SQLite)
project, _ := client.Projects().WithUUID(uuid).First(ctx)
todos, _ := client.Todos().InProject(uuid).All(ctx)

// Query headings in a project
headings, _ := client.Headings().InProject(uuid).All(ctx)

// Query with date filters
overdue, _ := client.Todos().Deadline().Past().All(ctx)
upcoming, _ := client.Todos().StartDate().Future().All(ctx)

// Query with multiple filters
todos, _ := client.Todos().
    Status().Incomplete().
    InArea(areaUUID).
    Deadline().Exists(true).
    All(ctx)

// Areas and tags
areas, _ := client.Areas().All(ctx)
tags, _ := client.Tags().All(ctx)
```

## Removed from Previous Design

| Removed | Reason |
|---------|--------|
| `TaskQueryBuilder` | Replaced by `TodoQueryBuilder` + `ProjectQueryBuilder` |
| `TypeFilterBuilder` | Type determined by which builder is used |
| `TaskRelationFilter` / `TaskStateFilter` / `TaskTimeFilter` | Intermediate group interfaces removed; methods directly on builders |
| `IncludeItems(bool)` | Flat design; no nested items |
| `WithUUIDPrefix(string)` | Rarely used; can be re-added if needed |
| `WithDeadlineSuppressed(bool)` | Rarely used; can be re-added if needed |
| Convenience methods (`Inbox`, `Today`, etc.) | Removed per RFC 008 Phase 2 |
| `Project.Items []Todo` | Flat design; use `client.Todos().InProject(uuid)` |
| `Heading.Todos []Todo` | Flat design; use `client.Todos().InHeading(uuid)` |
| `Area.Todos` / `Area.Projects` | Flat design; use builders |
| `Tag.Todos` / `Tag.Projects` / `Tag.Areas` | Flat design; use builders |
