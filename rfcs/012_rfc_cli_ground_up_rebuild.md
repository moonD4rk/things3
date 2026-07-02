# RFC 012: CLI Ground-Up Rebuild

Status: Implemented Author: @moond4rk Date: 2026-07-02 Supersedes: RFC 011

## Summary

This RFC records the completed ground-up rebuild of the `things3` command-line tool. The previous surface (`list`, `todo`, `project`, `search` with matching-mode flags) is deleted outright and replaced by a command set modeled directly on the Things 3 macOS app: sidebar views become top-level commands, the app's two mutation shortcuts become two verbs, and every write is fused with database verification so scripts get a confirmable, addressable result that a pure URL-scheme tool cannot provide. RFC 011 proposed this direction; RFC 012 is the design as built, superseding it. Three ideas drive every decision below: the app is the north star, the backend's quirks are absorbed so nothing leaks, and the design carries zero legacy.

## Design

### Principles

1. **The Things 3 app is the design north star.** Every command mirrors a sidebar item or an interaction verb. Someone who knows the app - human or agent - can guess the CLI. This is why views are top-level commands rather than `list <view>`, why scheduling and re-parenting are two separate verbs, and why text output groups items the way the app's list sections do.
2. **Digest the backend's quirks; leak nothing.** The Things SQLite schema and URL scheme have inconsistent corners: packed binary dates, a Today-vs-Evening discriminator, non-cascading trash, fire-and-forget writes, and phantom repeat-template rows. The CLI absorbs all of them. Internal vocabulary (start buckets, `todayIndex`, deadline suppression, sync columns, auth tokens) never appears in help, output, or errors.
3. **Optimal design, zero legacy.** Project policy allows breaking changes, so the old commands are removed without a deprecation window. The only survivors are genuine infrastructure constraints: the `cmd/things3/cmd` package path (goreleaser injects the version there) and the `version`/`completion` commands.

### Mapping the app to the CLI

The command names are not invented; they are read off the app. The mapping and its rationale:

| App concept | CLI surface | Why |
| --- | --- | --- |
| Sidebar views (Today, Inbox, Upcoming, Anytime, Someday, Logbook, Deadlines, Trash) | Top-level commands of the same name | Knowing the sidebar is knowing the commands; no `list` indirection between the user and a view |
| Cmd+S - When (a move in **time**) | `schedule <query> <when>` | The app gives scheduling its own shortcut, so the CLI gives it its own verb; `<when>` is a date or a bucket keyword, never a place |
| Cmd+Shift+M - Move (a move in **space**) | `move <query> --to <dest>` | Re-parenting into a project or area is spatial; a distinct app shortcut earns a distinct verb, keeping time and space from colliding on one flag |
| Quick Find | `show <query>` | One resolver across todos and projects: one hit renders a detail view, many render a list |
| App search | `search <query>` | Full-text listing semantics; an empty result is normal, not an error |
| New To-Do / New Project | `add "title"` / `add project "title"` | Creation verbs mirror the app's add menu |
| Complete / Cancel checkbox | `done` / `cancel` | Two outcomes the app draws differently ([x] vs [-]) map to two verbs |
| Reveal in app | `open [query \| view]` | Hands the item or a built-in list back to the GUI |
| Today / This Evening sections | Grouped text headers | Output grouping echoes the app's own list sections |

The time-verb / space-verb split is the load-bearing decision: because the app treats "when" and "where" as separate gestures, so does the CLI, and neither `schedule` nor `move` accepts the other's argument.

### Absorbing the backend's quirks

Each schema or scheme irregularity is met with a specific countermeasure so it never reaches the user:

| Backend quirk | How it would leak | Countermeasure |
| --- | --- | --- |
| Fire-and-forget URL-scheme writes (no returned UUID, no acknowledgement) | Scripts cannot reference or confirm a write | Post-write verification poll against the database; created items are recovered by title and creation time |
| Packed 27-bit calendar dates and clock times | Raw integers instead of dates | The library decodes them; the CLI prints `YYYY-MM-DD` and `HH:MM` |
| `start` bucket enum (Inbox / Anytime / Someday) | Internal bucket vocabulary in output | Encoded inside the view queries; the words never appear in help or output |
| `startBucket` Today-vs-Evening discriminator and `todayIndex` ordering | Sync and sort columns surfacing | Read internally; Evening appears only as a "This Evening" section; ordering is applied in the client |
| Deadline suppression of overdue items | A "suppressed" concept leaking into Today | Folded into the library's `Today` composition and removed from the public surface |
| Non-cascading `trash` flag (own-row only) | Children of a trashed parent shown or hidden incorrectly | The trash view queries each type's own `trashed` flag; a live child of a trashed parent stays exactly as the database models it |
| Phantom repeat templates and repeat instances | Ghost or duplicate todos | Templates are excluded from every query; instances surface only as a read-only `repeats` marker |
| Auth token stored in an internal settings table | The user forced to locate and pass a token | Read from the database automatically for every write |

### Command surface

Commands are organized into four help groups, presented in this order. Global persistent flags apply everywhere: `-o, --output text|json|yaml` (a parse-time-validated enum, default `text`), `-n, --limit N` (0 = unlimited, applied after fetch), and `--db <path>` (highest precedence over the `THINGSDB` environment variable and auto-discovery). Write commands additionally accept `--dry-run` and `--no-verify`; `open` accepts only `--dry-run`. The global-flags and output sections here are extended by RFC 013, which adds the shared list-pagination flags (`--page`, `--all`, `--sort`, `--desc`, `--tag`) and the container row segment. Amendment: a later change replaced the `-o, --output` enum with three mutually exclusive boolean switches - `--text` (default), `-j, --json`, and `-y, --yaml` - and promoted those list flags to persistent global flags, honored by the list commands and accepted-but-ignored everywhere else.

| Group | Commands |
| --- | --- |
| Views | `today` (Today + This Evening), `inbox`, `upcoming`, `anytime`, `someday`, `logbook` (`--days N`, default 30), `deadlines`, `trash` |
| Collections | `projects` (`--area <q>`, `--all`), `areas`, `tags` |
| Lookup | `show <query>`, `search <query>` |
| Actions | `add "title"` / `add project "title"`, `done <q>`, `cancel <q>`, `schedule <q> <when>`, `move <q> --to <dest>`, `edit <q>`, `open [<q>\|<view>]` |

`add` carries the full creation surface (`--notes`, `--when`, `--deadline`, `--reminder`, mutually exclusive `--project`/`--area`/`--heading` placement, `--tags`, repeatable `--checklist`); `add project` mirrors it for projects (`--area`, repeatable `--todos`). `edit` covers inspector-style attribute changes (`--title`, `--notes`, `--append-notes`, `--deadline`, `--clear-deadline`, `--tags`, `--add-tags`) and deliberately omits `--when`, because scheduling is `schedule`'s job. Every command sets a group and ships two or three realistic examples in its help, since agents learn the tool from `--help`. The retained `version` and `completion` commands are infrastructure, not part of the model.

### Query resolution

One resolver backs `show`, every write's target argument, and the `--project`/`--area`/`--heading`/`--to` destination flags. It matches in four tiers and stops at the first tier that produces any hit:

1. Exact UUID.
2. UUID prefix, requiring at least four characters so a short word cannot collide with the head of a UUID.
3. Exact title, case-insensitive.
4. Title substring.

Hits merge across todos and projects and are ranked stably: open items before closed, then todos before projects. This ranking is what lets the tool print an 8-character short UUID everywhere and accept it back as input, closing the round-trip gap that forced users to copy full UUIDs before. The resolution outcome is contract, not heuristic: a read (`show`) prints all matches and exits 0, while a write demands exactly one match - zero matches is a not-found error, and more than one is an ambiguity that prints candidates and refuses to execute.

### Write-path fusion

Writes are a three-phase pipeline that fuses the read side (database) with the write side (URL scheme):

1. **Resolve** the query to exactly one full UUID and type using the resolver above.
2. **Execute** the corresponding `things:///` URL through the scheme (background `osascript`, or foreground `open`).
3. **Verify** by polling the database within a short budget (roughly two seconds, at a 150 ms interval, cancelable): `done`/`cancel` wait for the status to flip, `schedule`/`move`/`edit` wait for the modification timestamp to advance past the pre-write baseline, and `add` recovers the created item by exact title plus a creation-time guard so the new UUID can be reported back.

The pivotal rule is that **an unverified send is still success**. The URL scheme cannot acknowledge a write, so verification is best-effort: if the budget elapses, or several same-title items make the created one unidentifiable, the command reports "sent to Things, not yet confirmed" and exits 0, because the send itself succeeded. `--dry-run` prints the `things:///` URL and executes nothing; `--no-verify` executes but skips the poll. Scope limits inherited from the scheme are stated plainly rather than faked: no delete, no move-to-Inbox (rejected with a message pointing at the app), checklist replace/append only, repeating rules read-only, and the auth token read from the database automatically.

### Output and exit-code contract

Text is the default, human format; `json` and `yaml` are the machine formats and always carry the same underlying fields. A text list row is a status checkbox (`[ ]`, `[x]`, `[-]`), the 8-character UUID, and the title, with optional ` | date`, ` | #tags`, and ` | repeats` suffixes; cross-type lists (trash, multi-hit `show`, `search`) insert a TYPE column. Grouping and section headers (Today / This Evening, upcoming-by-date, anytime-by-container) are text-only sugar. Machine output never changes shape with grouping: a list is always a flat top-level array - `[]` when empty, never `null` - and a mixed list tags each element with a `type` discriminator instead of splitting into sections. Write commands emit a structured result object. Errors are rendered centrally in `main`: text mode writes `Error: <message>` (plus a candidate table for ambiguity) to stderr, while `-o json` writes `{"error": "...", "candidates": [...]}`.

Exit codes are enforced in tests and documented for scripting:

| Code | Meaning |
| --- | --- |
| 0 | Success, including empty listings and sent-but-unverified writes |
| 1 | Any error: bad flags, database failure, a zero-match lookup or write target, an unparseable `<when>` |
| 2 | Ambiguous write target: the query (or a destination) matched more than one item; candidates are printed and nothing is executed |

## Implementation Notes

- **Module and package layout.** The CLI is a separate Go module under `cmd/things3`, wired to the library through a workspace. The `cmd/things3/cmd` package path is pinned because the release tooling injects the version symbol there; it must never be renamed. The root command wires the four groups, the three global flags, and `SilenceUsage`/`SilenceErrors` so `main` owns all error rendering and the process exit code.
- **Cobra-free cores.** Resolution and verification live in `internal/resolve` and `internal/verify` as pure `(ctx, *things3.Client, ...)` packages with no cobra dependency. This makes them unit-testable against the fixture without a command harness and positions them for later promotion into the library. Verification takes an injectable clock so tests are deterministic and never wait on real time.
- **Shared bootstrap.** A `withClient` wrapper opens the client (honoring `--db` over `THINGSDB` over auto-discovery), runs the command body, and closes the client, so no command manages the database lifecycle itself. A small helper attaches the shared write flags. Command bodies return errors upward and write through the command's output writer, never touching `os.Stdout` directly.
- **Library prerequisites.** The rebuild depended on a handful of library additions that keep the CLI honest: `client.Today` (encapsulating the three-part Today composition and removing deadline suppression from the public surface), `Evening` and `Repeating` fields on the models, project UUID-prefix matching for the resolver, a strict `ParseWhen` that errors on garbage instead of silently ignoring it, escaped substring title matching, and empty query results returning empty slices rather than nil so JSON is always `[]`. Recency-ordered views (logbook, deadlines, upcoming) sort client-side.
- **Testing discipline.** Command tests run table-driven against the tracked fixture database via `--db`/`THINGSDB` and assert stdout, stderr, and exit code. Write paths are exercised only up to URL construction with `--dry-run`; tests never shell out to `osascript`, so the suite is portable off macOS. A help audit test walks the command tree and asserts every non-hidden command has a capitalized `Short`, a non-empty `Example`, and a group.
- **Removals.** `list`, `todo`, `project`, and the `--json`/`--yaml` flags are gone with no compatibility shim, consistent with the project's no-backward-compatibility policy. Their replacements are, respectively, the top-level views, the unified `show`, and the single `-o/--output` enum.
