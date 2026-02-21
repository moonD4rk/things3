# things3

[![Go CI](https://github.com/moond4rk/things3/actions/workflows/ci.yml/badge.svg)](https://github.com/moond4rk/things3/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/moond4rk/things3/branch/main/graph/badge.svg)](https://codecov.io/gh/moond4rk/things3)
[![Go Reference](https://pkg.go.dev/badge/github.com/moond4rk/things3.svg)](https://pkg.go.dev/github.com/moond4rk/things3)
[![Go Report Card](https://goreportcard.com/badge/github.com/moond4rk/things3)](https://goreportcard.com/report/github.com/moond4rk/things3)
[![License](https://img.shields.io/github/license/moond4rk/things3)](https://github.com/moond4rk/things3/blob/main/LICENSE)

Go library and CLI for [Things 3](https://culturedcode.com/things/) on macOS. Read tasks from the Things 3 SQLite database, create and update items via [Things URL Scheme](https://culturedcode.com/things/support/articles/2803573/), and query your task list from the terminal.

## Features

- **Unified client** - Single `NewClient()` entry point for all operations
- **Database queries** - Read-only access to the Things 3 SQLite database with fluent query builder and type-safe filters
- **URL Scheme** - Create todos, projects, and batch operations; update existing items with automatic authentication token management
- **CLI** - Query tasks, projects, areas, and tags from the terminal with JSON/YAML output
- **Interface-based API** - All public methods return interfaces for clean, testable code

## CLI

A command-line tool for querying your Things 3 tasks from the terminal.

### Installation

**Homebrew** (recommended):

```bash
brew install moond4rk/tap/things3
```

**Go**:

```bash
go install github.com/moond4rk/things3/cmd/things3@latest
```

> **Note**: Requires CGO enabled (uses `go-sqlite3`). macOS only.

**Manual**: Download pre-built binaries from [GitHub Releases](https://github.com/moonD4rk/things3/releases). If macOS shows "Apple could not verify", remove the quarantine attribute:

```bash
xattr -d com.apple.quarantine /path/to/things3
```

### Commands

```bash
things3 list <view>            # List tasks from a view
things3 search <query>         # Search tasks by title
things3 search --uuid <prefix> # Search tasks by UUID prefix
things3 version                # Print version information
```

**Available views**: `inbox`, `today`, `upcoming`, `anytime`, `someday`, `logbook`, `deadlines`, `projects`, `areas`, `tags`

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON |
| `--yaml` | `-y` | Output as YAML |
| `--limit` | `-n` | Max items to display (0 for unlimited) |

The `logbook` view also supports `--days` (`-d`) to limit results to recent N days (default: 30, 0 for all).

### Examples

```bash
things3 list today              # Today's tasks
things3 list inbox --json       # Inbox as JSON
things3 list logbook --days 7   # Recent 7 days
things3 search meeting          # Search by title
things3 search --uuid 4fthuhgF  # Search by UUID prefix
things3 list areas --yaml       # Areas as YAML
things3 list projects -n 5      # First 5 projects
```

### Output Formats

**Default (table)**:
```
STATUS   UUID      TYPE     TITLE
[x]      4fthuhgF  project  Task title | 2024-01-15 | #tag1 #tag2
[ ]      WZR4hDw5  todo     Another task | due:2024-02-01
[-]      gjUph7Jz  todo     Canceled task | 2024-01-10
```

Status indicators: `[ ]` incomplete, `[x]` completed, `[-]` canceled.

**JSON** (`--json`): Full task objects as a JSON array.

**YAML** (`--yaml`): Full task objects in YAML format.

## Library

### Installation

```bash
go get github.com/moond4rk/things3
```

> **Note**: Requires CGO enabled (uses `go-sqlite3`). macOS only.

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/moond4rk/things3"
)

func main() {
    client, err := things3.NewClient()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Get today's tasks
    today, _ := client.Today(ctx)
    for _, task := range today {
        fmt.Printf("- %s\n", task.Title)
    }

    // Create a new todo
    client.AddTodo().
        Title("Buy groceries").
        Notes("Milk, eggs, bread").
        When(things3.Today()).
        Execute(ctx)
}
```

### API Overview

#### Query Operations

##### Convenience Methods

```go
client.Inbox(ctx)                      // Tasks in Inbox
client.Today(ctx)                      // Today's tasks
client.Upcoming(ctx)                   // Scheduled future tasks
client.Anytime(ctx)                    // Anytime tasks
client.Someday(ctx)                    // Someday tasks
client.Logbook(ctx)                    // Completed/canceled tasks
client.Trash(ctx)                      // Trashed tasks
client.Todos(ctx)                      // All incomplete to-dos
client.Projects(ctx)                   // All incomplete projects
client.Deadlines(ctx)                  // Tasks with deadlines
client.Search(ctx, "query")            // Search tasks
client.CreatedWithin(ctx, DaysAgo(7))  // Tasks from last 7 days
```

##### Fluent Query Builder

```go
// Type-safe status filtering
tasks, _ := client.Tasks().
    Type().Todo().
    Status().Incomplete().
    All(ctx)

// Date filtering
tasks, _ := client.Tasks().
    StartDate().Future().
    Deadline().OnOrBefore(time.Now().AddDate(0, 0, 7)).
    All(ctx)

// Filter by area, project, or tag
tasks, _ := client.Tasks().
    InArea("area-uuid").
    InTag("work").
    All(ctx)

// Get single task or count
task, _ := client.Tasks().WithUUID("task-uuid").First(ctx)
count, _ := client.Tasks().Status().Completed().Count(ctx)
```

##### Areas and Tags

```go
// Get all areas with their tasks
areas, _ := client.Areas().IncludeItems(true).All(ctx)

// Get all tags
tags, _ := client.Tags().All(ctx)
```

#### Add Operations

##### Create Todo

```go
client.AddTodo().
    Title("Task title").
    Notes("Task notes").
    When(things3.Today()).              // today's date
    Deadline(time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local)).
    Tags("work", "urgent").
    ChecklistItems("Step 1", "Step 2").
    List("Project Name").               // or ListID("project-uuid")
    Reveal(true).
    Execute(ctx)
```

##### Create Project

```go
client.AddProject().
    Title("New Project").
    Notes("Project description").
    Area("Work").                       // or AreaID("area-uuid")
    Tags("important").
    Deadline(time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local)).
    Todos("Task 1", "Task 2", "Task 3"). // child todos
    Execute(ctx)
```

#### Update Operations

Update operations automatically manage authentication tokens.

```go
// Update a todo
client.UpdateTodo("todo-uuid").
    Title("Updated title").
    Completed(true).
    AddTags("done").
    Execute(ctx)

// Update a project
client.UpdateProject("project-uuid").
    Notes("Updated notes").
    Canceled(true).
    Execute(ctx)
```

#### Show Operations

```go
client.Show(ctx, "item-uuid")                    // Show specific item
client.ShowList(ctx, things3.ListToday)          // Show Today view
client.ShowSearch(ctx, "urgent tasks")           // Show search results

// Complex navigation
client.ShowBuilder().
    List(things3.ListInbox).
    Filter("work", "urgent").
    Execute(ctx)
```

#### Batch Operations

```go
// Create multiple items at once
client.Batch().
    AddTodo(func(t things3.BatchTodoConfigurator) {
        t.Title("Task 1").Tags("work")
    }).
    AddTodo(func(t things3.BatchTodoConfigurator) {
        t.Title("Task 2").When(things3.Today())
    }).
    AddProject(func(p things3.BatchProjectConfigurator) {
        p.Title("New Project").Notes("Description")
    }).
    Reveal(true).
    Execute(ctx)
```

### Configuration

```go
// Use custom database path
client, _ := things3.NewClient(
    things3.WithDatabasePath("/path/to/main.sqlite"),
)

// Enable SQL logging for debugging
client, _ := things3.NewClient(
    things3.WithPrintSQL(true),
)

// Control Things app focus behavior
client, _ := things3.NewClient(
    things3.WithForeground(),        // Bring Things to foreground (default for show)
    things3.WithBackground(),        // Run in background without stealing focus
)

// Preload authentication token
client, _ := things3.NewClient(
    things3.WithPreloadToken(),
)
```

#### Database Discovery

The database path is resolved in order:
1. Custom path via `WithDatabasePath()`
2. `THINGSDB` environment variable
3. Default: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`

### Types

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

// Date helper functions
things3.Today()       // returns today's date at midnight
things3.Tomorrow()    // returns tomorrow's date at midnight

// Scheduling methods (called on builders)
.When(time.Time)      // schedule for specific date
.WhenEvening()        // schedule for this evening
.WhenAnytime()        // schedule for anytime (no specific date)
.WhenSomeday()        // schedule for someday (indefinite future)

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
