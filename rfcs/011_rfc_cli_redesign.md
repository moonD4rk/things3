# RFC 011: CLI Redesign - Task-Oriented Command Surface

Status: Draft Author: @moond4rk Date: 2026-07-02 Supersedes: RFC 010

## Summary

Redesign the things3 CLI from an API-mirroring, read-only tool into a task-oriented command surface that fuses database queries and URL scheme execution behind user intentions. Users never choose matching modes, never handle full UUIDs, and gain full write access (add, complete, cancel, edit, move, open) with database-backed verification that pure URL-scheme tools cannot provide.

## Motivation

The current CLI (RFC 010) has three structural problems:

1. **Read-only.** The library's entire write surface (TodoAdder, TodoUpdater, ShowNavigator, batch operations) is unreachable from the command line.
2. **Implementation details leak.** Users must choose a matching mode (`--title` vs `--search` vs positional UUID), know that `project` requires a full UUID while `todo` accepts a prefix, and navigate a `list` subcommand that mixes todo views (inbox, today) with domain collections (projects, areas, tags).
3. **Database and scheme are not fused.** The short UUIDs the CLI displays cannot be fed back into `project`. Write operations, once added naively, would be fire-and-forget with no confirmation and no way to reference the created item.

## Design

### Principles

1. **Task-oriented, not API-oriented.** Commands express what the user wants (see today, finish a task), not which library method runs.
2. **Human identifiers.** Every command accepting a task reference resolves it through one shared algorithm over titles and UUID prefixes. Users are never required to look up a UUID.
3. **Read-write fusion.** Write commands are three-phase: resolve (database) -> execute (URL scheme) -> verify (database). The database makes writes addressable and confirmable; the scheme makes the database actionable.
4. **Script-friendly.** Single `--output` flag, semantic exit codes, empty collections encode as `[]`, created items echo their real UUID.
5. **Progressive disclosure.** Frequent operations are shortest. Internal concepts (start buckets, deadline suppression, todayIndex, trashed-parent cascades, auth tokens) never appear in help or output.

### Command Surface

```
Views (top level, replacing `list <view>`):
  things3 today | inbox | upcoming | anytime | someday
  things3 logbook [--days N] | deadlines | trash

Collections:
  things3 projects [--area <q>] [--all]
  things3 areas
  things3 tags

Lookup:
  things3 show <query>          resolve across todos and projects;
                                1 match -> detail, N -> list, 0 -> exit 1
  things3 search <query>        full-text search across todos and projects

Actions:
  things3 add "title" [--notes s] [--when W] [--deadline D] [--reminder HH:MM]
              [--project <q> | --area <q> | --heading <q>]
              [--tags a,b] [--checklist "x" --checklist "y"]
  things3 add project "title" [--area <q>] [--notes s] [--when W] [--deadline D] [--todos "a" --todos "b"]
  things3 done <query>
  things3 cancel <query>
  things3 edit <query> [--title s] [--notes s] [--append-notes s] [--when W]
              [--deadline D] [--clear-deadline] [--tags a,b] [--add-tags c]
  things3 move <query> --to <project-q | area-q | inbox | today | anytime | someday>
  things3 open [<query> | <view>]   reveal in the Things app

Misc:
  things3 version

Global flags:
  -o, --output text|json|yaml   (replaces --json / --yaml)
  -n, --limit N
      --db <path>               (overrides THINGSDB and auto-discovery)
```

Command groups (cobra groups) keep help readable: Views, Collections, Lookup, Actions.

Removed relative to RFC 010: `list` (views move to top level), `todo` and `project` detail commands (unified into `show`), all matching-mode flags (`--title`, `--search`, `--uuid`). Per project policy there is no backward compatibility or deprecation period.

### Identifier Resolution

One resolver shared by `show`, `done`, `cancel`, `edit`, `move`, `open`, and the `--project`/`--area`/`--heading` flags:

1. Exact UUID match (todos and projects).
2. UUID prefix match.
3. Exact title match, case-insensitive.
4. Title substring match.

First tier with hits wins; hits merge across todos and projects with incomplete items ranked before closed ones. Resolution outcome:

- Exactly one match: proceed.
- Multiple matches: read commands (`show`) print the list normally and exit 0; write commands print candidates (short UUID, type, title) to stderr and exit 2 without executing. Writes require an unambiguous target.
- Zero matches: message to stderr, exit 1.

The 8-character short UUIDs the CLI prints are always valid inputs, closing the current round-trip break.

### Write-Path Fusion

`done <q>`: resolve to full UUID and type -> `UpdateTodo(uuid).Completed(true).Execute()` -> poll the database (short interval, roughly 2s budget) until the status flips -> print the confirmed item. On verification timeout, report "sent to Things, not yet confirmed" and exit 0; the send itself succeeded.

`add "title"`: record t0 -> execute the add -> poll `Todos().WithTitle(title).CreatedAfter(t0)` -> print the created item with its real UUID (JSON output includes the full model). This gives scripts a handle to the new item, which fire-and-forget URL scheme tools cannot provide. On timeout, same unconfirmed semantics as above.

`--when` values (`today`, `tomorrow`, `evening`, `someday`, `anytime`, `yyyy-mm-dd`) parse through a strict variant of ApplyWhen that rejects unknown input with an error instead of silently ignoring it.

Deliberate scope cuts: no `delete` (the URL scheme cannot trash items; `cancel` plus the app covers it), no checklist mutation beyond add/replace (scheme limitation), no repeating-task management (not in the scheme).

### Output and Exit Codes

- `--output text` (default): current compact line and detail formats, plus Reminder in the detail view.
- `--output json|yaml`: full models; empty result sets encode as `[]`, never `null`.
- Exit codes: 0 success (including empty view listings), 1 error or zero-match lookup, 2 ambiguous write target.
- `--json --yaml` style conflicts become impossible by construction (one enum flag).

## Implementation Notes

### Library Prerequisites

Ordered by how hard they block the CLI:

1. `ProjectQueryBuilder.WithUUIDPrefix` - resolver cannot work for projects without it.
2. Timezone model fix (scan parses localtime strings as UTC; `CreatedAfter` formats the caller's wall clock instead of normalizing to local) - `add` verification and `logbook --days` are unreliable until fixed.
3. LIKE metacharacter escaping (`%`, `_`) - resolver title matching must be literal.
4. Strict when-parser (`ParseWhen` returning an error) alongside the lenient `ApplyWhen`.
5. `client.Today(ctx)` encapsulating the three-query Today composition, removing `DeadlineSuppressed` from the public surface.
6. Empty query results return empty slices, not nil (JSON `[]`).
7. Sort support for recency-ordered views (logbook by stop date descending, deadlines by deadline, upcoming by start date); client-side sorting in the CLI is an acceptable interim.
8. Optional, later: promote the resolver into the library (`client.Resolve`) once CLI usage stabilizes its semantics.

### Phasing

1. Library fixes above (items 1-6 minimum).
2. CLI skeleton migration: top-level views, `--output`, `--db`, exit codes, shared client bootstrap helper, command groups.
3. Resolver plus `show`/`search` unification.
4. Write commands: `add`, `done`, `cancel`, `edit`, `move`, `open`, with verification loops.
5. Tests against the fixture database via `--db` (the CLI currently has zero tests; the bootstrap helper plus `--db` makes command-level testing possible), README rewrite, cobra completion.

### Testing

Every command gets table-driven tests running against `testdata/main.sqlite` through `--db`, asserting stdout, stderr, and exit codes. Write commands are tested up to URL construction (scheme execution mocked or skipped off-macOS); resolution and verification logic are testable purely against the fixture.
