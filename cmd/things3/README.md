# things3 CLI

A command-line interface for querying Things 3 tasks on macOS.

## Installation

```bash
go install github.com/moond4rk/things3/cmd/things3@latest
```

Or build from source:

```bash
go build -o things3 ./cmd/things3
```

## Usage

```
things3 [command] [flags]
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON |
| `--yaml` | `-y` | Output as YAML |
| `--limit` | `-n` | Max items to display (0 for unlimited) |

## Commands

### list

List tasks from various Things 3 views.

```bash
things3 list <view> [flags]
```

Available views:

| View | Description |
|------|-------------|
| `inbox` | Tasks in the Inbox |
| `today` | Tasks scheduled for today |
| `upcoming` | Tasks scheduled for future dates |
| `anytime` | Tasks in the Anytime list |
| `someday` | Tasks in the Someday list |
| `logbook` | Completed and canceled tasks (supports `--days`) |
| `deadlines` | Tasks with deadlines |
| `projects` | All incomplete projects |
| `areas` | All areas |
| `tags` | All tags |

#### logbook flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--days` | `-d` | 30 | Limit to recent N days (0 for all) |

### search

Search for tasks by title or UUID prefix.

```bash
things3 search <query> [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--uuid` | `-u` | Search by UUID prefix instead of title |

### version

Print version information.

```bash
things3 version
```

## Examples

List today's tasks:
```bash
things3 list today
```

List inbox tasks as JSON:
```bash
things3 list inbox --json
```

List first 5 projects:
```bash
things3 list projects -n 5
```

List logbook from last 7 days:
```bash
things3 list logbook --days 7
```

List all logbook entries:
```bash
things3 list logbook --days 0
```

Search for tasks containing "meeting":
```bash
things3 search meeting
```

Search by UUID prefix:
```bash
things3 search --uuid 4fthuhgF
```

List all areas as YAML:
```bash
things3 list areas --yaml
```

## Output Formats

### Default (Table format)

```
STATUS   UUID      TYPE     TITLE
[x]      4fthuhgF  project  Task title | 2024-01-15 | #tag1 #tag2
[ ]      WZR4hDw5  todo     Another task | due:2024-02-01
[-]      gjUph7Jz  todo     Canceled task | 2024-01-10
```

- **STATUS**: `[ ]` incomplete, `[x]` completed, `[-]` canceled
- **UUID**: First 8 characters (use `search --uuid` to query)
- **TYPE**: `todo`, `project`, or `heading`
- **TITLE**: Task title with optional date and tags

### JSON

```json
[
  {
    "uuid": "ABC123",
    "type": "to-do",
    "title": "Task title",
    "status": "incomplete",
    "start": "Today"
  }
]
```

### YAML

```yaml
- uuid: ABC123
  type: to-do
  title: Task title
  status: incomplete
  start: Today
```

## Architecture

```
cmd/things3/
├── main.go           # Entry point
├── cmd/
│   ├── root.go       # Root command with global flags
│   ├── list.go       # List command with view subcommands
│   ├── search.go     # Search command
│   ├── output.go     # Output formatting (text/JSON/YAML)
│   └── version.go    # Version command
└── README.md
```

### Design Principles

- **Factory Functions**: Use `NewXxxCmd()` pattern, no `init()`
- **Testability**: Use `cmd.OutOrStdout()` for output
- **Error Handling**: Return errors, let `main()` handle `os.Exit()`
- **Single Registration**: Subcommands registered explicitly in parent

## Requirements

- macOS (Things 3 is macOS-only)
- Things 3 application installed
- Read access to Things 3 database
