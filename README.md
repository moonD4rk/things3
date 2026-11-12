# things3

[![Go CI](https://github.com/moond4rk/things3/actions/workflows/ci.yml/badge.svg)](https://github.com/moond4rk/things3/actions/workflows/ci.yml) [![codecov](https://codecov.io/gh/moond4rk/things3/branch/main/graph/badge.svg)](https://codecov.io/gh/moond4rk/things3) [![Go Reference](https://pkg.go.dev/badge/github.com/moond4rk/things3.svg)](https://pkg.go.dev/github.com/moond4rk/things3) [![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](https://github.com/moond4rk/things3/blob/main/LICENSE)

AI-friendly Go library and CLI for [Things 3](https://culturedcode.com/things/) on macOS.

Reads come straight from the Things SQLite database through typed query builders; writes go through the official [Things URL scheme](https://culturedcode.com/things/support/articles/2803573/) and are verified against the database afterwards. Every CLI command speaks the same machine-readable protocol - uniform flags, self-describing JSON pagination, a strict exit-code contract, and stable UUID handles - so it works equally well for a human at a terminal and an AI agent or script driving it. The read API is a full Go port of [things.py](https://github.com/thingsapi/things.py).

## Highlights

- **Complete read API** - typed, chainable query builders for todos, projects, headings, areas, and tags, plus the eight sidebar views (`Inbox`, `Today`, `Upcoming`, `Anytime`, `Someday`, `Logbook`, `Deadlines`, `Trash`), each returning what the app lists there, in the app's order
- **Type-safe writes** - create, update, and batch operations built as fluent URL-scheme builders with automatic auth-token management
- **Verified writes in the CLI** - every action resolves its target in the database, executes the URL scheme, then polls the database to confirm the write actually landed
- **Built for automation** - one global flag surface (`--json` / `--yaml` / `--text`), list output wrapped in a `{items, total, page, pages}` envelope so truncation is never silent, exit codes `0`/`1`/`2`, and every printed 8-char UUID usable as a query
- **App-faithful views** - the CLI mirrors the Things sidebar (`today`, `inbox`, `upcoming`, ...) and its verbs (`schedule`, `move`, `edit`), and every view was verified item for item against the app, so what a list shows is what the app shows

> **Requirements**: macOS with Things 3 installed. CGO enabled when building from source (`github.com/mattn/go-sqlite3`). Write commands need Things running; reads only need the database file.

## CLI

### Install

```bash
brew install moond4rk/tap/things3            # Homebrew (recommended)
go install github.com/moond4rk/things3/cmd/things3@latest
```

Pre-built binaries are on [GitHub Releases](https://github.com/moonD4rk/things3/releases). If macOS complains about verification: `xattr -d com.apple.quarantine /path/to/things3`.

### Commands

The whole surface, straight from `things3 -h`:

```text
Views:
  anytime     List available todos grouped by project or area
  deadlines   List open items with deadlines, soonest first
  inbox       List todos in the Inbox
  logbook     List completed and canceled items, most recent first
  someday     List Someday todos and projects
  today       List the Today list, This Evening last
  trash       List trashed todos and projects
  upcoming    List scheduled items, repeating series, and deadlines by date

Collections:
  areas       List all areas
  projects    List projects
  tags        List all tags

Lookup:
  search      Full-text search across todos and projects
  show        Show an item by UUID, prefix, or title (Quick Find)

Actions:
  add         Add a todo
  cancel      Cancel a todo or project
  done        Complete a todo or project
  edit        Edit a todo or project's attributes
  move        Move a todo or project to a project or area (the app's Move)
  open        Reveal an item, built-in list, or search in Things.app
  schedule    Schedule a todo or project (the app's When)

Flags:
      --all          show all items without pagination (list commands)
      --db string    Things database path (overrides THINGSDB)
      --desc         reverse the --sort order (list commands)
  -h, --help         help for things3
  -j, --json         output as JSON
  -n, --limit int    max items to display (0 = unlimited)
      --page page    page number, 1-based (list commands) (default 1)
      --sort sort    sort by: date, created, modified, title (list commands)
      --tag string   keep only items carrying this tag, case-insensitive (list commands)
      --text         output as plain text (default)
  -y, --yaml         output as YAML
```

The format switches (`--text` / `--json` / `--yaml`) are mutually exclusive, text is the default, and lists show 10 items per page unless `-n` or `--all` says otherwise. Write commands additionally accept `--dry-run` (print the `things:///` URL without executing) and `--no-verify` (skip the post-write confirmation).

### Read examples

```bash
things3 today                        # the Today list, projects included, This Evening sectioned
things3 upcoming                     # by date: scheduled items, repeating series, deadlines
things3 logbook --days 7 -n 5        # five most recent completions this week
things3 anytime --tag work --sort title
things3 search meeting --page 2      # paginated full-text search
things3 show Stzhb3Sc                # any printed 8-char UUID is a valid query, trashed items included
```

Text lists print rows like

```
STATUS   UUID      TITLE
[ ]      52vEKPy3  Buy concert tickets | 2026-07-03 | @Personal | #errand
( )      8QdfrtG2  Plan the tour | 2026-07-03 | @Personal
-- 1-10 of 44 (page 1/5) | next: --page 2 | all: --all
```

where a square mark is a todo and a round one a project, as in the app. Every view lists exactly what the app lists there, in the app's order: `anytime` leaves out the todos of a Someday project, `logbook` folds a closed project's todos into the project's row, and so on. Machine formats wrap every list in a self-describing envelope:

```bash
$ things3 upcoming --json --page 2 | jq '{total, page, pages, first: .items[0].title}'
{ "total": 44, "page": 2, "pages": 5, "first": "Weekly review" }
```

### Write examples

```bash
things3 add "Buy milk" --when today --reminder 18:00 --tags errand
things3 add "Ship v1" --deadline 2026-08-01 --project "Launch"
things3 add project "Launch"
things3 done "Buy milk"              # resolve -> execute -> verify against the DB
things3 schedule "Ship v1" 2026-07-10
things3 move "Ship v1" --to Launch
things3 edit "Ship v1" --title "Ship v1.0" --clear-deadline
```

Targets resolve by exact UUID, UUID prefix (4+ chars), exact title, then title substring. When a title names several items, the single open one wins, so `done "Water plants"` reaches today's todo however many completed copies the Logbook holds. A target that is still ambiguous prints a candidate table with a UUID hint, executes nothing, and exits `2`; success (including a sent-but-unconfirmed write) exits `0`; every other error exits `1`. Trashed items and repeating series resolve by UUID and can be shown, but a write to one is refused with an explanation. Under `--json`, errors arrive on stderr as `{"error", "candidates"}`.

Inherited URL-scheme limits (not worked around): no delete or trash, no move to Inbox, checklist is replace-only, repeating rules are read-only, and unknown tags are silently ignored by Things.

### MCP server

`things3 mcp` runs a local [Model Context Protocol](https://modelcontextprotocol.io) server over stdio, exposing the same verbs as MCP tools so an AI assistant reads and writes Things through the one brew-installed binary. Point an MCP client at it:

```json
{
  "mcpServers": {
    "things3": {
      "command": "things3",
      "args": ["mcp"]
    }
  }
}
```

Add `"--read-only"` after `"mcp"` to register only the read tools, or `"--max-limit", "20"` to cap the list page size. The full tool list (including the `days` window on upcoming/logbook/deadlines and notes truncation) and Claude Code / Claude Desktop setup are in [`cmd/things3/README.md`](cmd/things3/README.md#mcp-server).

## Library

### Install

```bash
go get github.com/moond4rk/things3
```

### Quick start

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

    // Sidebar views: what the app lists, in the app's order
    today, _ := client.Today(ctx)       // []things3.Item: todos and projects
    upcoming, _ := client.Upcoming(ctx) // includes repeating series and deadlines
    fmt.Println(len(today), len(upcoming))

    // Typed query builders
    todos, _ := client.Todos().
        Status().Incomplete().
        StartDate().Future().
        All(ctx)
    fmt.Println(len(todos))

    // Writes via the URL scheme
    _ = client.AddTodo().
        Title("Buy groceries").
        Notes("Milk, eggs, bread").
        When(things3.Today()).
        Tags("errand").
        Execute(ctx)
}
```

### Views

A view returns `[]things3.Item`, where an item is a todo or a project (`item.Todo` or `item.Project`, with accessors such as `UUID()`, `Title()`, and `UpcomingDate()` for what the two share), because that is what the app's lists hold:

```go
client.Inbox(ctx)
client.Today(ctx)                       // This Evening last
client.Upcoming(ctx)                    // by date; series with no next occurrence trail
client.Anytime(ctx)                     // without the todos of Someday projects
client.Someday(ctx)                     // Someday projects and standalone Someday todos
client.Logbook(ctx, time.Time{})        // or pass a day to keep what was closed after it
client.Deadlines(ctx)
client.Trash(ctx)
```

An `Item` marshals to JSON as the todo or project with a leading `"type"` field.

### Query builders

One builder per domain type; filters chain and copy-on-write, terminals execute:

```go
client.Todos().Status().Incomplete().All(ctx)          // []Todo
client.Todos().InProject(uuid).Count(ctx)              // int
client.Todos().WithUUID(uuid).First(ctx)               // *Todo, checklist loaded
client.Todos().Deadline().Before(t).All(ctx)           // date filters: Exists, Future, Past, On, Before, After, ...
client.Todos().InArea(uuid).All(ctx)                   // loose todos and todos in the area's projects
client.Todos().WithUUID(uuid).IncludeTrashed().IncludeTemplates().First(ctx) // any todo that exists
client.Projects().InArea(uuid).All(ctx)
client.Headings().InProject(uuid).All(ctx)
client.Areas().All(ctx)
client.Tags().All(ctx)
```

Relationships are flat: parent references come inline for free (`todo.ProjectTitle`, `todo.AreaTitle` from SQL JOINs); children are separate queries (`Todos().InProject(uuid)`). A todo reports its effective containers: under a heading it carries the heading's project, and inside a project it carries that project's area, although the database stores neither on the todo itself.

Queries return live items by default. `IncludeTrashed()` adds the Trash, and `IncludeTemplates()` adds repeating templates: the hidden items that schedule a repeating series, flagged `Template` and carrying their next occurrence as `StartDate`.

### Writes

```go
client.AddTodo().
    Title("Task").
    When(things3.Tomorrow()).
    Reminder(9, 0).
    Deadline(deadline).
    ChecklistItems("Step 1", "Step 2").
    List("Project Name").              // or ListID(uuid)
    Execute(ctx)

client.AddProject().Title("New Project").Tags("work").Execute(ctx)

client.UpdateTodo(uuid).Completed(true).Execute(ctx)   // auth token managed automatically
client.UpdateProject(uuid).Notes("Updated").Execute(ctx)

client.Show(ctx, uuid)                                 // navigate the app
client.Batch().AddTodo(func(t things3.BatchTodoConfigurator) {
    t.Title("Task 1")
}).Execute(ctx)                                        // multiple items in one URL
```

### Configuration

```go
client, _ := things3.NewClient(
    things3.WithDatabasePath("/path/to/main.sqlite"), // else THINGSDB env, else auto-discovery
    things3.WithPrintSQL(true),                       // log executed SQL
    things3.WithForegroundExecution(),                // writes bring Things to the foreground
    things3.WithBackgroundNavigation(),               // show/navigation without stealing focus
    things3.WithPreloadToken(),                       // read the auth token at construction
    things3.WithClock(clock),                         // pin "today" for reproducible reports and tests
)
```

## License

[Apache License 2.0](LICENSE)
