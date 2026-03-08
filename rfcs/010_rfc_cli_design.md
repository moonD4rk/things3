# RFC 010: CLI Design

Status: Draft
Author: @moond4rk
Date: 2026-03-09

## Summary

Design specification for the `things3` CLI tool. The CLI provides command-line access to Things 3 data using the typed query builder API from RFC 009. It supports read operations across all domain types with flexible output formats.

## Architecture

The CLI uses Cobra with factory function pattern (no `init()`). All commands return errors and let `main()` handle `os.Exit()`.

```
things3
|
+-- list <view>          (List todos/projects/areas/tags by view)
|   +-- inbox
|   +-- today            (Three-query composition)
|   +-- upcoming
|   +-- anytime
|   +-- someday
|   +-- logbook           --days/-d (default: 30)
|   +-- deadlines
|   +-- projects
|   +-- areas
|   +-- tags
|
+-- todo <identifier>    (View todo by UUID prefix, title keyword, or search)
|   --title/-t           (Match by title keyword)
|   --search/-s          (Search title + notes + area)
|
+-- project <identifier> (View project by UUID, title keyword, or search)
|   --title/-t           (Match by title keyword)
|   --search/-s          (Search title + notes + area)
|
+-- search <query>       (Search todos by title/notes/area)
|   --uuid/-u            (Search by UUID prefix instead)
|
+-- version              (Show version)
|
Global flags:
  --limit/-n             (Max items to display, 0=unlimited)
  --json/-j              (Output as JSON)
  --yaml/-y              (Output as YAML)
```

## Command Design

### List Commands

Each list subcommand maps directly to a query builder chain. Key design decisions:

**Default filters**: All queries automatically exclude trashed items and items in trashed projects (parent-trashed exclusion is a default behavior in the database layer, not an explicit filter).

**Today view** is the most complex, composed of three separate queries:

1. Regular Today tasks: `StartDate().Exists(true)` + `Start().Anytime()` + `Status().Incomplete()`
2. Yellow dot tasks: `StartDate().Past()` + `Start().Someday()` + `Status().Incomplete()`
3. Overdue deadline tasks: `StartDate().Exists(false)` + `Deadline().Past()` + `DeadlineSuppressed(false)` + `Status().Incomplete()`

Results are concatenated in order: regular, scheduled, overdue.

**Upcoming** uses `Start().Someday()` (not Anytime) with `StartDate().Future()`.

**Someday** uses `StartDate().Exists(false)` to exclude tasks with scheduled dates.

**Logbook** defaults to 30 days via `--days` flag, uses `StopDate().Exists(true)` + `Status().Any()`.

### Detail Commands (todo, project)

Identifier resolution strategy:
- Default (todo): UUID prefix match via `WithUUIDPrefix()`
- Default (project): Exact UUID match via `WithUUID()`
- `--title/-t`: Title keyword match via `WithTitle()` (LIKE `%keyword%`)
- `--search/-s`: Full-text search via `Search()` (title + notes + area)

All modes use `Status().Any()` to include completed/canceled items.

Display behavior:
- 1 result: Show detailed view (all fields, checklist for todos, incomplete todos for projects)
- Multiple results: Show compact list view (same format as list commands)
- 0 results: No output

### Output Formats

Three output modes controlled by global flags:

| Format | Flag | Description |
|--------|------|-------------|
| Text | (default) | Compact one-line format with status checkbox and short UUID |
| JSON | `--json` | Full structured output via `json.Encoder` |
| YAML | `--yaml` | Full structured output via `yaml.Marshal` |

Text format columns: `STATUS UUID TITLE [| date] [| #tags]`

Detail text view shows all fields vertically (Title, UUID, Status, Start, Project, Area, Tags, When, Deadline, Notes, Checklist).

## File Structure

| File | Purpose |
|------|---------|
| `cmd/things3/main.go` | Entry point, calls `NewRootCmd().Execute()` |
| `cmd/things3/cmd/root.go` | Root command, global flags, subcommand registration |
| `cmd/things3/cmd/list.go` | List subcommands (inbox, today, upcoming, etc.) |
| `cmd/things3/cmd/todo.go` | Single todo lookup command |
| `cmd/things3/cmd/project.go` | Single project lookup command |
| `cmd/things3/cmd/search.go` | Search command |
| `cmd/things3/cmd/output.go` | Output formatting (text, JSON, YAML) |
| `cmd/things3/cmd/version.go` | Version command |

## Future Considerations

Potential commands not yet implemented:

- **Add**: `things3 add todo --title "..." --when today` via `client.AddTodo()`
- **Update**: `things3 update todo <uuid> --completed` via `client.UpdateTodo()`
- **Show/Open**: `things3 open <uuid>` via `client.Show()` to open in Things app
- **Trash**: `things3 list trash` for viewing trashed items
- **Headings**: `things3 list headings --project <uuid>` for project headings
- **Today convenience method**: `client.Today(ctx)` on Client to encapsulate the three-query Today logic, removing the need to expose `DeadlineSuppressed` on the public API
