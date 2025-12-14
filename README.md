# things3

[![Go CI](https://github.com/moond4rk/things3/actions/workflows/ci.yml/badge.svg)](https://github.com/moond4rk/things3/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/moond4rk/things3/branch/main/graph/badge.svg)](https://codecov.io/gh/moond4rk/things3)
[![Go Reference](https://pkg.go.dev/badge/github.com/moond4rk/things3.svg)](https://pkg.go.dev/github.com/moond4rk/things3)
[![Go Report Card](https://goreportcard.com/badge/github.com/moond4rk/things3)](https://goreportcard.com/report/github.com/moond4rk/things3)
[![License](https://img.shields.io/github/license/moond4rk/things3)](https://github.com/moond4rk/things3/blob/main/LICENSE)

Go library for [Things 3](https://culturedcode.com/things/) on macOS. Provides read-only database access and full URL Scheme support for creating and updating tasks.

## Features

- Read-only access to Things 3 SQLite database
- Fluent query builder with type-safe filters
- Full [Things URL Scheme](https://culturedcode.com/things/support/articles/2803573/) support
- Create todos, projects, and batch operations via URL
- Update existing items with authentication

## Installation

```bash
go get github.com/moond4rk/things3
```

> **Note**: Requires CGO enabled (uses `go-sqlite3`). macOS only.

## Quick Start

### Reading from Database

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/moond4rk/things3"
)

func main() {
    db, err := things3.NewDB()
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    ctx := context.Background()

    // Get today's tasks
    today, _ := db.Today(ctx)
    for _, task := range today {
        fmt.Printf("- %s\n", task.Title)
    }
}
```

### Creating Tasks via URL Scheme

```go
scheme := things3.NewScheme()

// Create a new todo
url, _ := scheme.Todo().
    Title("Buy groceries").
    Notes("Milk, eggs, bread").
    When(things3.WhenToday).
    Tags("shopping").
    Build()
// Output: things:///add?title=Buy+groceries&notes=Milk,+eggs,+bread&when=today&tags=shopping
```

## Database API

### Convenience Methods

```go
db.Inbox(ctx)                    // Tasks in Inbox
db.Today(ctx)                    // Today's tasks
db.Upcoming(ctx)                 // Scheduled future tasks
db.Anytime(ctx)                  // Anytime tasks
db.Someday(ctx)                  // Someday tasks
db.Logbook(ctx)                  // Completed/canceled tasks
db.Trash(ctx)                    // Trashed tasks
db.Todos(ctx)                    // All incomplete to-dos
db.Projects(ctx)                 // All incomplete projects
db.Deadlines(ctx)                // Tasks with deadlines
db.Search(ctx, "query")          // Search tasks
db.CreatedWithin(ctx, Days(7))   // Tasks from last 7 days
```

### Fluent Query Builder

```go
// Type-safe status filtering
tasks, _ := db.Tasks().
    Type().Todo().
    Status().Incomplete().
    All(ctx)

// Date filtering
tasks, _ := db.Tasks().
    StartDate().Future().
    Deadline().OnOrBefore(time.Now().AddDate(0, 0, 7)).
    All(ctx)

// Filter by area, project, or tag
tasks, _ := db.Tasks().
    InArea("area-uuid").
    InTag("work").
    All(ctx)

// Get single task or count
task, _ := db.Tasks().WithUUID("task-uuid").First(ctx)
count, _ := db.Tasks().Status().Completed().Count(ctx)
```

### Areas and Tags

```go
// Get all areas with their tasks
areas, _ := db.Areas().IncludeItems(true).All(ctx)

// Get all tags
tags, _ := db.Tags().All(ctx)
```

## URL Scheme API

### Create Todo

```go
scheme := things3.NewScheme()

url, _ := scheme.Todo().
    Title("Task title").
    Notes("Task notes").
    When(things3.WhenToday).              // or WhenTomorrow, WhenEvening, WhenAnytime, WhenSomeday
    WhenDate(2024, time.December, 25).    // specific date
    Deadline("2024-12-31").
    Tags("work", "urgent").
    ChecklistItems("Step 1", "Step 2").
    List("Project Name").                  // or ListID("project-uuid")
    Reveal(true).
    Build()
```

### Create Project

```go
url, _ := scheme.Project().
    Title("New Project").
    Notes("Project description").
    Area("Work").                          // or AreaID("area-uuid")
    Tags("important").
    Deadline("2024-12-31").
    Todos("Task 1", "Task 2", "Task 3").   // child todos
    Build()
```

### Navigate to Views

```go
scheme.Show().List(things3.ListToday).Build()    // things:///show?id=today
scheme.Show().List(things3.ListInbox).Build()    // things:///show?id=inbox
scheme.Show().ID("project-uuid").Build()          // things:///show?id=project-uuid
scheme.Search("urgent tasks")                     // things:///search?query=urgent+tasks
```

### Update Existing Items (Requires Auth)

```go
// Get auth token from database
token, _ := db.Token(ctx)
auth := scheme.WithToken(token)

// Update a todo
url, _ := auth.UpdateTodo("todo-uuid").
    Title("Updated title").
    Completed(true).
    AddTags("done").
    Build()

// Update a project
url, _ := auth.UpdateProject("project-uuid").
    Notes("Updated notes").
    Canceled(true).
    Build()
```

### Batch Operations with JSON

```go
// Create multiple items at once
url, _ := scheme.JSON().
    AddTodo(func(t *things3.JSONTodoBuilder) {
        t.Title("Task 1").Tags("work")
    }).
    AddTodo(func(t *things3.JSONTodoBuilder) {
        t.Title("Task 2").When(things3.WhenToday)
    }).
    AddProject(func(p *things3.JSONProjectBuilder) {
        p.Title("New Project").Notes("Description")
    }).
    Reveal(true).
    Build()

// Batch updates (requires auth)
url, _ := auth.JSON().
    UpdateTodo("uuid-1", func(t *things3.JSONTodoBuilder) {
        t.Completed(true)
    }).
    UpdateTodo("uuid-2", func(t *things3.JSONTodoBuilder) {
        t.Canceled(true)
    }).
    Build()
```

## Configuration

```go
// Use custom database path
db, _ := things3.NewDB(
    things3.WithDatabasePath("/path/to/main.sqlite"),
)

// Enable SQL logging for debugging
db, _ := things3.NewDB(
    things3.WithPrintSQL(true),
)
```

### Database Discovery

The database path is resolved in order:
1. Custom path via `WithDatabasePath()`
2. `THINGSDB` environment variable
3. Default: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`

## Types

```go
// Task types
things3.TaskTypeTodo     // 0 - To-do item
things3.TaskTypeProject  // 1 - Project
things3.TaskTypeHeading  // 2 - Heading

// Status
things3.StatusIncomplete // 0
things3.StatusCanceled   // 2
things3.StatusCompleted  // 3

// Start bucket
things3.StartInbox    // 0
things3.StartAnytime  // 1
things3.StartSomeday  // 2

// When values for URL Scheme
things3.WhenToday
things3.WhenTomorrow
things3.WhenEvening
things3.WhenAnytime
things3.WhenSomeday

// List IDs for navigation
things3.ListInbox
things3.ListToday
things3.ListUpcoming
things3.ListAnytime
things3.ListSomeday
things3.ListLogbook
things3.ListTrash
```

## References

- [Things URL Scheme Documentation](https://culturedcode.com/things/support/articles/2803573/)
- [things.py](https://github.com/thingsapi/things.py) - Python library (inspiration for database access patterns)

## License

Apache License 2.0
