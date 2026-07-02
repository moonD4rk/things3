# CLAUDE.md

Guidance for Claude Code when working in this repository.

## What This Is

**things3** is two things in one repo, wired as separate Go modules via `go.work`:

- A Go **library** (repo root): read-only access to the Things 3 macOS SQLite database plus type-safe `things:///` URL-scheme building and execution. A port of Python `things.py` with full API parity.
- A **CLI** (`cmd/things3`): mirrors the app's sidebar views and interaction verbs on top of the library.

The single most important architectural fact: **reads and writes take different roads**. Reads query the SQLite database directly (`internal/database`). Writes can only go through the fire-and-forget URL scheme (`internal/scheme`) - there is no write access to the database, and the scheme's limits are hard boundaries, never worked around: no delete or trash, no move-to-Inbox, checklist replace-only, repeating items read-only, tags never created, auth token auto-read from the DB.

## Commands

Run in both modules (repo root and `cmd/things3`):

```bash
go build ./... && go test ./...                    # Build and test
golangci-lint run                                  # Lint
gofumpt -l -w .                                    # Format (stricter than gofmt)
goimports -w -local github.com/moond4rk/things3 .  # Format imports
```

## Library Architecture

Core principles, in order of importance:

- **Single entry point**: `NewClient()` is the only public constructor (functional options via `ClientOption`). Everything hangs off `Client`.
- **Typed, copy-on-write query builders**: one builder per domain type (`Todos()`, `Projects()`, `Headings()`, `Areas()`, `Tags()`), no union Task type. Chainable filters clone the builder and return interfaces; terminals are `All`/`First`/`Count`. Generic sub-builders (`StatusFilter[T]`, `StartFilter[T]`, `DateFilter[T]`) keep filter chains type-safe.
- **Flat data model**: no nested items. Upward relationships come inline for free from SQL JOINs (`todo.ProjectTitle`, `todo.AreaTitle`); downward relationships are separate builder queries (`Todos().InProject(uuid)`).
- **View composition lives on Client**: anything beyond a single query is a Client method, not a builder trick - `Today()` (three-part composition with Evening ordering), `Upcoming()` (scheduled todos merged with repeating templates).
- **Naming**: `With*` identity, `In*` relationship, `Has*` existence; enums type-prefixed (`StatusCompleted`, `StartInbox`).

Domain facts that are easy to get wrong:

- Enums map database integers: Status 0/2/3 = incomplete/canceled/completed; StartBucket 0/1/2 = inbox/anytime/someday.
- Things stores dates in custom binary formats (27-bit date, packed time); all conversion is centralized in `internal/database/date.go` - never hand-roll it.
- Repeating tasks: future occurrences are never materialized as rows. Only the template row (`rt1_recurrenceRule` set) exists, carrying the single next occurrence in `rt1_nextInstanceStartDate`. Normal queries exclude templates; `Upcoming()` opts them back in and maps the next occurrence to `StartDate`.

## CLI Architecture (cmd/things3)

Cobra with factory functions (`NewXxxCmd()`, registered explicitly in `NewRootCmd()`; no `init()` except the pinned `version.go` - do not touch it). Four command groups: **Views** (`today`, `inbox`, `upcoming`, `anytime`, `someday`, `logbook`, `deadlines`, `trash`), **Collections** (`projects`, `areas`, `tags`), **Lookup** (`show`, `search`), **Actions** (`add`, `done`, `cancel`, `schedule`, `move`, `edit`, `open`).

### One uniform global flag surface

All flags are persistent on root; every command shows the same surface, non-list commands accept-and-ignore the list flags (deliberate trade-off):

- Format: `--text` (default) / `-j, --json` / `-y, --yaml`, mutually exclusive booleans
- List shaping: `-n` (page size), `--page`, `--all`, `--sort date|created|modified|title`, `--desc`, `--tag`
- `--db <path>` overrides discovery (precedence: `--db` > `THINGSDB` env > auto)

List pipeline is always fetch -> tag-filter -> sort -> paginate, default page size 10 in every format. Text prints a pagination footer; `json`/`yaml` lists are a self-describing envelope `{items, total, page, pages}` where `items` is always `[]`, never `null`. Detail views, write results, and stderr error objects stay bare. Write commands add `--dry-run` (print URL, do not execute) and `--no-verify`.

### Resolve -> execute -> verify

Every action follows `runWrite` (`write.go`): resolve the target, execute the URL scheme, poll the database to confirm. Two cobra-free internal packages carry the logic and are candidates for promotion into the library:

- `internal/resolve`: tiered matching - exact UUID -> UUID prefix (>= 4 chars) -> exact title -> title substring - open-before-closed; every 8-char UUID the CLI prints is a valid query, for targets and `--to` destinations alike.
- `internal/verify`: polls the DB (2s budget, injectable clock) to confirm a fire-and-forget write landed.

### Exit-code contract

- **0** - success, including empty lists and sent-but-unconfirmed writes
- **1** - any error (bad flags, DB failure, not found, unparseable input)
- **2** - ambiguous target: candidates printed with a UUID-targeting hint, nothing executed

Commands only return errors; `main.go` owns rendering (`RenderError`, format-aware: JSON errors go to stderr as `{"error", "candidates"}`) and `os.Exit`. Wrap DB command bodies with `withClient` (never open/close the client inline); write through `cmd.OutOrStdout()`; set `GroupID` and a realistic `Example` on every command.

## Development Workflow

1. **Design first**: significant changes start as an RFC in `rfcs/NNN_snake_case_title.md` (header: Status/Author/Date; sections: Summary, Design, Implementation Notes). Current specs: RFC 009 (library query/model), RFC 012 (CLI) as amended by RFC 013 (list pagination, output envelope).
2. **Implement library before CLI** when a change spans both; keep the CLI a thin consumer.
3. **Verify in both modules**: `go build`, `go test`, `golangci-lint run`, `gofumpt` before finishing.
4. **Sync docs**: `README.md` and `cmd/things3/README.md` must match implemented behavior. Release notes use Library and CLI bullet sections with PR references.

House rules for every artifact: English only, no emoji, no local machine paths. Backward compatibility is a non-goal - remove or redesign, never deprecate.

## Testing

- Table-driven throughout; integration tests run against the fixture in `testdata/` via `thingstest.DatabasePath(t)` with `t.Setenv("THINGSDB", ...)` or `--db`.
- **Action-command tests never execute osascript**: they assert `--dry-run` URLs and error paths only. No test requires Things to be installed; resolve/verify are covered directly against the fixture.
- Every exported identifier gets a doc comment starting with its name. Real end-to-end write verification is manual, against a live Things app.

## Reference

- Python origin: https://github.com/thingsapi/things.py
