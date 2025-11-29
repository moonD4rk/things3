# things3

[![Go CI](https://github.com/moond4rk/things3/actions/workflows/ci.yml/badge.svg)](https://github.com/moond4rk/things3/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/moond4rk/things3/branch/main/graph/badge.svg)](https://codecov.io/gh/moond4rk/things3)
[![Go Reference](https://pkg.go.dev/badge/github.com/moond4rk/things3.svg)](https://pkg.go.dev/github.com/moond4rk/things3)

Go library for read-only access to the [Things 3](https://culturedcode.com/things/) macOS app database. A Go port of [things.py](https://github.com/thingsapi/things.py) with full API parity.

## Installation

```bash
go get github.com/moond4rk/things3
```

> **Note**: Requires CGO enabled (uses `go-sqlite3`). macOS only.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/moond4rk/things3"
)

func main() {
    client, err := things3.New()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Get inbox tasks
    inbox, _ := client.Inbox(ctx)
    fmt.Printf("Inbox: %d tasks\n", len(inbox))

    // Get today's tasks
    today, _ := client.Today(ctx)
    for _, task := range today {
        fmt.Printf("- %s\n", task.Title)
    }
}
```

## Core API

### Convenience Methods

```go
client.Inbox(ctx)      // Tasks in Inbox
client.Today(ctx)      // Today's tasks
client.Upcoming(ctx)   // Scheduled future tasks
client.Anytime(ctx)    // Anytime tasks
client.Someday(ctx)    // Someday tasks
client.Logbook(ctx)    // Completed/canceled tasks
client.Trash(ctx)      // Trashed tasks
client.Todos(ctx)      // All incomplete to-dos
client.Projects(ctx)   // All incomplete projects
client.Deadlines(ctx)  // Tasks with deadlines
client.Search(ctx, "query")     // Search tasks
client.Last(ctx, "7d")          // Tasks from last 7 days
```

### Query Builder

For complex queries, use the fluent query builder:

```go
// Find incomplete to-dos in a specific project
tasks, err := client.Tasks().
    WithType(things3.TaskTypeTodo).
    WithStatus(things3.StatusIncomplete).
    InProject("project-uuid").
    All(ctx)

// Find tasks with a specific tag
tasks, err := client.Tasks().
    WithTag("work").
    All(ctx)

// Get a single task by UUID
task, err := client.Tasks().
    WithUUID("task-uuid").
    First(ctx)

// Count matching tasks
count, err := client.Tasks().
    WithStatus(things3.StatusCompleted).
    Count(ctx)
```

### Areas and Tags

```go
// Get all areas
areas, err := client.Areas().All(ctx)

// Get area with its tasks
areas, err := client.Areas().
    IncludeItems(true).
    All(ctx)

// Get all tags
tags, err := client.Tags().All(ctx)
```

## Configuration

```go
// Use custom database path
client, err := things3.New(
    things3.WithDatabasePath("/path/to/main.sqlite"),
)

// Enable SQL logging for debugging
client, err := things3.New(
    things3.WithPrintSQL(true),
)
```

### Database Discovery

The database path is resolved in order:
1. Custom path via `WithDatabasePath()`
2. `THINGSDB` environment variable
3. Default Things 3 location: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite`

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
```

## License

Apache License 2.0
