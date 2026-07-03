# things3

Query and control Things 3 from the terminal. `things3` reads the Things 3 macOS SQLite database directly for fast, offline queries, writes changes through the official Things URL scheme, and then verifies each write against the database so you know whether it actually landed. It mirrors the app's sidebar views (`today`, `inbox`, `upcoming`, ...) and its interaction verbs (`schedule` for When, `move` for Move), so anyone who knows the app can guess the command.

## Install

```bash
# Homebrew (recommended)
brew install moond4rk/tap/things3

# Go toolchain
go install github.com/moond4rk/things3/cmd/things3@latest
```

Build from source:

```bash
git clone https://github.com/moond4rk/things3.git
cd things3/cmd/things3
go build -o things3 .
```

Requirements:

- macOS with Things 3 installed. Reads work anywhere the database file is readable; **write commands require macOS with Things running** (they drive the app through the URL scheme).
- CGO enabled. The SQLite driver (`github.com/mattn/go-sqlite3`) is a cgo package, so `CGO_ENABLED=1` and a C toolchain are needed when building from source. Homebrew and release binaries ship prebuilt.

## Quick start

```bash
things3 today                              # what's on your plate today
things3 add "Buy milk" --when today        # capture a todo, scheduled for today
things3 done "Buy milk"                     # complete it
things3 show "Write report"                # find and inspect an item
things3 logbook --days 7                    # what you finished this past week
things3 search meeting --json              # machine-readable search results
```

## Command reference

Commands are grouped exactly as `things3 --help` shows them: Views, Collections, Lookup, Actions.

### Global flags

Every command accepts these persistent flags:

| Flag | Description |
| --- | --- |
| `--text` | Output as plain text. This is the default when no format flag is set. |
| `-j, --json` | Output as JSON. |
| `-y, --yaml` | Output as YAML. |
| `-n, --limit N` | Maximum items to display (`0` = unlimited). When set, this also becomes the page size. |
| `--db <path>` | Database path. Overrides `THINGSDB` and auto-discovery. |

The three format flags are mutually exclusive; combining any two (for example `--json --yaml`) fails at validation time and exits `1`.

Write commands additionally accept `--dry-run` (print the `things:///` URL, do not execute) and `--no-verify` (skip the post-write database confirmation). `open` accepts `--dry-run` only.

### List flags

The list-shaping flags below are also global persistent flags, giving one uniform flag surface across every command. The list commands (all Views, `projects`, `areas`, `tags`, and `search`) honor them; the other commands (`show` on a single match, all Actions, `version`, `completion`) accept them but ignore them - the accepted trade-off for a uniform surface.

| Flag | Description |
| --- | --- |
| `--page N` | 1-based page to show. Values below `1` fail at parse time. |
| `--all` | Show every item without pagination. On `projects` it additionally includes done/canceled projects. |
| `--sort date\|created\|modified\|title` | Sort key. Omitted keeps the view's natural order. Invalid values fail at parse time. Areas and tags have no dates or timestamps, so those keys leave them in natural order; `title` sorts them. |
| `--desc` | Reverse the `--sort` order. |
| `--tag <name>` | Keep only items carrying this tag (case-insensitive). On `areas` it filters by the area's own tags; on `tags` it matches the tag's own title. |

The pipeline is **fetch, then filter by `--tag`, then `--sort`, then paginate**. The `date` sort uses each row's shown date (completed, else canceled, else deadline, else start) and always ranks undated items last, in both directions.

**Page size.** A list defaults to a page of `10` in every format. `--all` or `-n 0` mean unlimited; an explicit `-n N` sets the page size directly. Pagination is uniform across `text`, `json`, and `yaml`, so `--page`, `--sort`, `--desc`, and `--tag` behave identically. Machine output is never silently truncated: `json` and `yaml` wrap the page in a self-describing envelope (see [JSON and YAML](#json-and-yaml)) whose `total`/`page`/`pages` fields describe the slice within the full result.

**Footer.** In text mode a paginated list ends with a footer that also tells you how to move on:

```
-- 1-3 of 10 (page 1/4) | next: --page 2 | all: --all
```

The `next:` hint disappears on the last page (`-- 10-10 of 10 (page 4/4) | all: --all`). The footer is suppressed for `json`/`yaml`, for unlimited output, and when a single page already shows the whole list.

### Views

| Command | Flags | Description | Example |
| --- | --- | --- | --- |
| `today` | - | Today's todos, with a This Evening section when present | `things3 today` |
| `inbox` | - | Incomplete todos in the Inbox | `things3 inbox` |
| `upcoming` | `--days N` (0 = all) | Scheduled future todos plus repeating tasks at their next occurrence, grouped by date | `things3 upcoming --days 7` |
| `anytime` | - | Anytime todos, grouped by project or area | `things3 anytime` |
| `someday` | - | Someday todos with no scheduled date | `things3 someday` |
| `logbook` | `--days N` (default 30, 0 = all) | Completed and canceled todos, most recent first | `things3 logbook --days 7` |
| `deadlines` | `--days N` (0 = all, keeps overdue) | Incomplete todos with deadlines, soonest first | `things3 deadlines --days 7` |
| `trash` | - | Trashed todos and projects (mixed list) | `things3 trash` |

The **Flags** column lists view-specific flags only; every view also accepts the shared [List flags](#list-flags) (`--page`, `--all`, `--sort`, `--desc`, `--tag`). For example:

```bash
things3 upcoming --tag work           # only repeating/scheduled todos tagged work
things3 logbook -n 5 --page 2         # the second page of five completed items
things3 anytime --sort title          # anytime todos, sorted by title
```

### Collections

| Command | Flags | Description | Example |
| --- | --- | --- | --- |
| `projects` | `--area <q>`, `--all` | Incomplete projects; `--all` includes done/canceled; `--area` filters by area | `things3 projects --area Work` |
| `areas` | - | All areas; honors the shared [List flags](#list-flags) | `things3 areas` |
| `tags` | - | All tags; honors the shared [List flags](#list-flags) | `things3 tags` |

### Lookup

| Command | Args | Description | Example |
| --- | --- | --- | --- |
| `show` | `<query>` | Quick Find across todos and projects. One match prints a detail view; several print a mixed list; none is an error | `things3 show "Write report"` |
| `search` | `<query>` | Full-text search across todos and projects (title, notes, area). Empty results are fine | `things3 search meeting` |

### Actions

Every action resolves the target, executes the write, and verifies it. All accept `--dry-run` and `--no-verify`.

| Command | Args | Key flags | Description | Example |
| --- | --- | --- | --- | --- |
| `add` | `<title>` | `--notes`, `--when`, `--deadline`, `--reminder`, `--project` / `--area` / `--heading`, `--tags`, `--checklist` (repeatable) | Create a todo | `things3 add "Email Bob" --project Work --tags urgent` |
| `add project` | `<title>` | `--notes`, `--when`, `--deadline`, `--area`, `--todos` (repeatable) | Create a project | `things3 add project "Website redesign" --area Work` |
| `done` | `<query>` | - | Complete a todo or project | `things3 done "Buy milk"` |
| `cancel` | `<query>` | - | Cancel a todo or project | `things3 cancel "Old idea"` |
| `schedule` | `<query> <when>` | - | Set When (the app's Cmd+S) | `things3 schedule "Write report" 2026-08-01` |
| `move` | `<query>` | `--to <dest>` (required) | Move to a project or area (the app's Move) | `things3 move "Buy milk" --to Groceries` |
| `edit` | `<query>` | `--title`, `--notes`, `--append-notes`, `--deadline`, `--clear-deadline`, `--tags`, `--add-tags` | Edit attributes (at least one flag) | `things3 edit "Report" --add-tags urgent` |
| `open` | `[<query>\|<view>]` | `--dry-run` | Reveal an item or built-in list in Things.app; no args opens Today | `things3 open today` |

Notes:

- `--when` and `schedule`'s `<when>` accept `today`, `tomorrow`, `evening`, `anytime`, `someday`, or `YYYY-MM-DD`.
- `--deadline` takes `YYYY-MM-DD`; `--reminder` takes `HH:MM`.
- `add`: `--project`, `--area`, and `--heading` are placement targets; `--heading` requires `--project`, and `--project`/`--area` are mutually exclusive.
- `open` view names: `inbox`, `today`, `upcoming`, `anytime`, `someday`, `logbook`, `deadlines`.
- `version` and `completion` (shell completion) are also available.

## How queries resolve

`show` and every write command turn a `<query>` into concrete items through four tiers, tried in order. The first tier that matches anything wins; later tiers are not consulted:

1. **Exact UUID** - the full 22-character Things identifier.
2. **UUID prefix** - four or more characters that begin a UUID (shorter strings are treated as titles, not prefixes).
3. **Exact title** - case-insensitive full-title match.
4. **Title substring** - case-insensitive containment.

Matches are ranked **open items before closed** (completed/canceled), with todos before projects on ties. Trashed items never match.

**Every 8-character UUID the CLI prints is a valid query.** List rows show the first 8 characters of each UUID; because that satisfies the four-character prefix rule and is effectively unique, you can copy any printed short UUID straight back into `show`, `done`, `schedule`, and the rest.

**The write rule:** write commands demand exactly one match.

- Zero matches -> "no item matches ..." error (exit 1).
- One match -> the write proceeds.
- More than one match -> the candidates are printed and **nothing is executed** (exit 2). Disambiguate with a UUID prefix from the candidate list.

`show` is more lenient: several matches simply produce a mixed list (exit 0).

## Output formats

### Text

List rows are compact and columnar:

```
STATUS   UUID      TITLE
[ ]      5pUx6PES  Review pull requests | 2026-07-02 | @Work / Backend | #urgent
[ ]      A1b2C3d4  Call the dentist | due:2026-07-05 | @Errands | #errands
[x]      7F4vqUNi  Draft release notes | 2026-06-30
```

- **STATUS** - `[ ]` incomplete, `[x]` completed, `[-]` canceled.
- **UUID** - first 8 characters (a valid query).
- **TITLE** - followed by optional `|` suffixes, in order: a date (`YYYY-MM-DD` for a scheduled/closed date, `due:YYYY-MM-DD` for a deadline), a container (`@Project` with `/ Heading` when set, else `@Area`), `#tags`, and `repeats` for repeating items.

The container segment makes each row legible on its own. Views that already group by container (`anytime`) omit it to avoid repeating the group header on every row.

Mixed lists (`trash`, `search`, and multi-match `show`) insert a **TYPE** column:

```
STATUS   UUID      TYPE     TITLE
[ ]      A1b2C3d4  todo     Call the dentist | due:2026-07-05
[x]      3x1QqJqf  project  Website redesign | 2026-06-30
```

Grouped views add headers in **text mode only**: `today` splits into `Today` / `This Evening`, `upcoming` groups by date (each repeating task appears under its next occurrence with a `repeats` marker), `anytime` groups by project or area. `areas` and `tags` print as simple `- Title` lists. A single-match `show` prints a vertical detail view (Title, UUID, Status, Start, Project/Area/Heading, Tags, When, Deadline, Reminder, Done/Canceled, Notes, Checklist; a project detail is followed by its incomplete todos).

Paginated lists end with a text-only footer (see [List flags](#list-flags)):

```
2026-09-17 Thursday
[ ]      7F4vqUNi  To-Do in Upcoming | 2026-09-17

2040-01-01 Sunday
[ ]      N1PJHsbj  Weekly review | @Work | repeats
-- 1-3 of 12 (page 1/4) | next: --page 2 | all: --all
```

### JSON and YAML

`--json` and `--yaml` carry the full model objects and **ignore all text-only grouping**. YAML mirrors JSON field-for-field. Timestamps are RFC 3339; date-only and time-only fields are formatted accordingly. Fields marked optional below are omitted when empty or zero.

**List envelope.** Every list command (all Views, `projects`, `search`, multi-match `show`, `areas`, `tags`) wraps its page in a self-describing envelope:

```json
{
  "items": [ ... ],
  "total": 42,
  "page": 1,
  "pages": 5
}
```

`items` is the page slice and always encodes as `[]` when empty (never `null`), so an out-of-range `--page` still returns `{"items": [], ...}`. `total` is the full count after `--tag` filtering; `page` and `pages` are 1-based. Unlimited output (`--all` or `-n 0`) reports `page: 1`, `pages: 1`. `areas` and `tags` run through the same pipeline and envelope, so `--page`, `-n`, `--sort title`, and (for `areas`) `--tag` shape them too. Detail views (single-item `show`, project detail) and write results stay bare objects, not envelopes.

**Todo**

| Field | Type | Notes |
| --- | --- | --- |
| `uuid` | string | Things identifier |
| `title` | string |  |
| `status` | string | `incomplete`, `completed`, or `canceled` |
| `notes` | string | optional |
| `start` | string | `inbox`, `anytime`, or `someday` |
| `area_uuid`, `area_title` | string | parent area, optional |
| `project_uuid`, `project_title` | string | parent project, optional |
| `heading_uuid`, `heading_title` | string | parent heading, optional |
| `tags` | array of string | optional |
| `checklist` | array of item | populated on single-item `show` and verified `add`; optional |
| `start_date` | string (date) | scheduled date, optional |
| `deadline` | string (date) | optional |
| `reminder` | string (time) | optional |
| `created_at`, `modified_at` | string (RFC 3339) |  |
| `completed_at`, `canceled_at` | string (RFC 3339) | optional |
| `trashed` | bool | optional (omitted when false) |
| `evening` | bool | true for This Evening items; optional |
| `repeating` | bool | instance of a repeating template; optional |

**Project** - identical to Todo minus `project_*`, `heading_*`, `checklist`, and `evening`.

**Mixed items** (`trash`, `search`, multi-match `show`) - the full todo or project object with an added `type` discriminator:

```json
{
  "type": "project",
  "uuid": "3x1QqJqf...",
  "title": "Website redesign",
  "status": "incomplete",
  "start": "anytime"
}
```

**writeResult** - every action prints one of these:

| Field | Type | Notes |
| --- | --- | --- |
| `action` | string | `add`, `done`, `cancel`, `schedule`, `move`, `edit`, `open` |
| `verified` | bool | whether the write was confirmed in the database (always present) |
| `dry_run` | bool | present and true only under `--dry-run` |
| `url` | string | the `things:///` URL (dry-run only) |
| `type` | string | `todo` or `project`, optional |
| `todo` / `project` | object | the confirmed item on a verified write, optional |
| `uuid` | string | the affected item's UUID, optional |
| `message` | string | human-readable note (e.g. why a send was not confirmed), optional |

**Errors** go to stderr. In text mode they read `Error: <message>` (plus a candidate table for ambiguity, followed by a hint to rerun with one of the listed UUIDs). Under `--json` the stderr payload is:

```json
{
  "error": "query \"report\" matches 2 items",
  "candidates": [
    { "uuid": "A1b2C3d4", "type": "todo", "title": "Write report" },
    { "uuid": "3x1QqJqf", "type": "project", "title": "Report redesign" }
  ]
}
```

`candidates` is present only for ambiguity errors.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success. Includes empty listings and sent-but-unverified writes (the send itself succeeded). |
| `1` | Any error: bad flags, database failure, a zero-match `show` or write target, an unparseable `<when>`. |
| `2` | Ambiguous write target: the query matched more than one item (or `--to` / `--project` / `--area` was ambiguous). Candidates are printed; nothing was executed. |

Note: an **unverified send still exits 0**. The URL scheme is fire-and-forget, so a send that Things accepted but that the verification poll could not confirm in time is treated as success, not failure.

## Write semantics

Every action runs three phases:

1. **Resolve** - the query is matched against the database (see "How queries resolve"). Exactly one match is required, or the command stops before touching Things.
2. **Execute** - the change is sent to Things through the URL scheme (`osascript` in the background). The auth token required for updates is read automatically from the database; you never pass it.
3. **Verify** - the CLI polls the database for a short budget to confirm the change (a new UUID for `add`, a status flip for `done`/`cancel`, an advanced modification time for `schedule`/`move`/`edit`).

A **verified** write prints the resulting item line (`done: [x] 5pUx6PES Review pull requests`). An **unverified** send - the write was accepted but not confirmed within the poll window - prints `sent to Things (not yet confirmed)` and still exits 0.

- `--dry-run` prints the exact `things:///` URL and executes nothing, e.g. `things:///add?tags=work&title=Draft%20release%20notes&when=2026-07-02`. Ideal for inspecting or piping a command.
- `--no-verify` executes but skips the confirmation poll, reporting the send as unverified.

Limits inherited from the Things URL scheme (the CLI absorbs these; it never pretends to do more):

- **No delete.** There is no delete or move-to-Trash command.
- **No move to Inbox.** `move --to inbox` is rejected; use the Things app.
- **Checklists are replace-only** when set on an existing item.
- **Repeating rules are read-only.** Repeating items are flagged (`repeats`) but their schedules cannot be edited.

## Environment

The database path resolves in this order:

1. `--db <path>` flag (highest precedence)
2. `THINGSDB` environment variable
3. Auto-discovery of the standard Things 3 database in its macOS group container

For safe experimentation, point the CLI at a **copy** of your database instead of the live file:

```bash
cp /path/to/your/things.sqlite ./things-copy.sqlite
things3 --db ./things-copy.sqlite today
```

Read commands only ever read. Write commands, however, round-trip through Things itself, so they always affect the live app regardless of `--db` - use `--dry-run` to see what a write would do without sending it.

## Scripting with `--json`

`--json` plus `jq` makes the CLI composable.

**Filter today's list.** Print today's todos that carry a deadline (list output is an envelope, so iterate `.items[]`; pass `--all` to page through everything):

```bash
things3 today --all --json \
  | jq -r '.items[] | select(.deadline) | "\(.uuid[0:8])  \(.title)  (due \(.deadline[0:10]))"'
```

**Add, then capture the new UUID.** A verified `add` returns the created item's UUID at the top level (write results are bare objects, not envelopes):

```bash
uuid=$(things3 add "Draft release notes" --when today --json | jq -r '.uuid')
echo "created $uuid"
```

**Complete an item by UUID.** Feed that UUID straight into `done` and confirm it verified:

```bash
things3 done "$uuid" --json | jq '.verified'   # -> true
```

Combine with exit codes for control flow: `things3 show "$uuid" --json >/dev/null || echo "gone"`.

## MCP server

`things3 mcp` serves a local [Model Context Protocol](https://modelcontextprotocol.io) server over stdio, exposing the same verbs as MCP tools so an AI assistant drives Things through the same binary. Reads query the database; writes run resolve -> execute -> verify, exactly like the CLI. Logs go to stderr; stdout carries the protocol. The server stops cleanly on stdin EOF or SIGINT/SIGTERM.

### Tools

Thirteen verb-shaped tools mirror the CLI:

| Tool | Kind | Purpose |
| --- | --- | --- |
| `list_todos` | read | a sidebar view (`view`: inbox, today, upcoming, anytime, someday, logbook, deadlines, trash), optionally narrowed by project, area, or tag, or a `days` window on upcoming/logbook/deadlines (logbook defaults to the last 30 days, 0 = all) |
| `list_projects`, `list_areas`, `list_tags` | read | the collections, paginated |
| `search` | read | quick find across todos and projects |
| `get` | read | resolve one item; a project answer nests its incomplete todos and headings |
| `add_todo`, `add_project` | write | create, with when / deadline / reminder / tags / checklist |
| `complete` | write | done, cancel, or reopen via a status enum |
| `schedule`, `move`, `edit` | write | reschedule, refile, or change attributes |
| `open` | nav | reveal an item or list in the app |

Every id, target, and destination resolves by UUID, 4+ char prefix, exact title, then title substring; an ambiguous match returns candidate UUIDs to retry with. Lists paginate with `limit` and 1-based `page` in the `{items, total, page, pages}` envelope; `limit` carries machine-readable schema bounds (default 20, minimum 1, maximum 100), so an omitted `limit` is stamped to 20 and an over-cap value is rejected rather than silently clamped. List and search items shorten notes to 200 characters and set `notes_truncated`; `get` returns the full note. Writes are verified against the database and report `verified: true|false`; an accepted-but-unconfirmed send is still a success. Domain failures ride the envelope as a structured error (`invalid_input`, `not_found`, `ambiguous`, `execution_failed`) so a model can self-correct.

### Flags

- `--read-only` registers only the six read tools; the write tools and `open` are not exposed at all.
- `--max-limit N` caps the list page size for the session (0 uses the built-in maximum of 100), lowering both the advertised schema maximum and the enforced cap.
- `--log-level debug|info|warn|error` (default `info`) sets the stderr log level.
- `--db <path>` selects the database, like every other command.

### Client configuration

**Claude Code** — register the server once:

```bash
claude mcp add things3 -- things3 mcp
```

**Claude Desktop** — add to `claude_desktop_config.json`:

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

Use `"args": ["mcp", "--read-only"]` to expose only the read tools.

## Development

The CLI is a separate Go module under `cmd/things3`, wired to the root library via `go.work`.

```bash
# build
(cd cmd/things3 && go build -o things3 .)

# test - both modules; the CLI suite runs against the tracked fixture
go test ./...
(cd cmd/things3 && go test ./...)

# run a single command against a fixture database
THINGSDB=/path/to/fixture.sqlite go run ./cmd/things3 today

# lint and format
golangci-lint run
gofumpt -l .
```

Tests never execute the URL scheme; write commands are exercised in `--dry-run` mode, and reads run against the fixture supplied through `THINGSDB`. The resolver and verifier live in `cmd/things3/internal/resolve` and `cmd/things3/internal/verify` as cobra-free packages so they can be unit-tested directly.

Design docs: the accepted design decisions are captured in [RFC 012](../../rfcs/012_rfc_cli_ground_up_rebuild.md) and [RFC 013](../../rfcs/013_rfc_cli_list_pagination_and_context.md).
