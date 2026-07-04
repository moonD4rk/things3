# RFC 014: MCP Server as a CLI Subcommand

Status: Implemented Author: @moonD4rk Date: 2026-07-03

## Summary

things3 gains a local Model Context Protocol server, served over stdio by a new `things3 mcp` subcommand. It replaces the experimental `things3-mcp-server` repository (to be archived), and its tool surface is a clean-slate design that mirrors the rebuilt CLI's verbs instead of inheriting the old server's entity/method taxonomy. One brew-installed binary serves both humans and AI assistants; an MCP client configures it as `{"command": "things3", "args": ["mcp"]}`. Reads run on the library's typed query builders and composed views; writes run resolve -> execute -> verify on the same infrastructure as the CLI.

## Design

### Why in-repo, why from scratch

The old server cannot be ported: it pins library v0.5.4 and calls APIs the rebuild removed (`client.Tasks()`, the convenience view methods). Beyond that, it lacks exactly what this repo has since built: writes were fire-and-forget with no database verification, ids had to be exact UUIDs (no fuzzy resolution), there were zero tests, no configuration surface (`--db`, read-only), two inconsistent error contracts, no enum constraints in tool schemas, and malformed dates were silently dropped. Rebuilding inside this repo lets the MCP layer consume `internal/resolve` and `internal/verify` unchanged and ship in the existing goreleaser/homebrew single-binary pipeline with zero packaging changes.

The old server's tool taxonomy (entity tools with a `method` discriminator: `todo_read`, `todo_write`, ...) is deliberately discarded, not just reimplemented. Only its result vocabulary (`success`/`message`/`hint`) and backlog ideas (`append_notes`, `add_tags`, `clear_deadline`, reminder support) survive, folded into v1.

### Placement

- `cmd/things3/internal/mcpserver`: a cobra-free server package, mirroring how `resolve` and `verify` are structured for direct unit testing. It owns the MCP tool handlers, input/output types, schema registration, pagination math, and conversion from library models.
- `cmd/things3/cmd/mcp.go`: a thin cobra command that reads the ldflags-injected version (private to package `cmd`; `version.go` and goreleaser stay untouched), builds the client via the existing `withClient`, and runs the server.

MCP output shapes are owned by `mcpserver`, not shared with the CLI's private rendering types. The pagination envelope semantics (`items`/`total`/`page`/`pages`) are kept identical to the CLI contract by tests, not by code coupling: the math is trivial and the two layers serve different consumers (display vs. protocol).

### Tool surface: verb-shaped, mirroring the CLI

RFC 012's pivotal rule was "knowing the sidebar is knowing the commands". This RFC extends it: knowing the CLI is knowing the MCP tools. Each tool is one verb with one small, single-purpose schema; views and attributes are parameters. Thirteen tools:

| Tool | Kind | Parameters (all optional unless noted) | CLI counterpart |
| --- | --- | --- | --- |
| `list_todos` | read | `view` (required enum: `inbox`, `today`, `upcoming`, `anytime`, `someday`, `logbook`, `deadlines`, `trash`), `project`, `area`, `tag`, `days` (upcoming/logbook/deadlines only), `limit`, `page` | the eight view commands |
| `list_projects` | read | `area`, `tag`, `status` (enum, default `incomplete`), `limit`, `page` | `projects` |
| `list_areas` | read | `limit`, `page` | `areas` |
| `list_tags` | read | `limit`, `page` | `tags` |
| `search` | read | `query` (required), `type` (enum: `todo`, `project`, `any`; default `any`), `status`, `tag`, `limit`, `page` | `search` |
| `get` | read | `id` (required; UUID, 4+ char prefix, or title) | `show <query>` |
| `add_todo` | write | `title` (required), `notes`, `when`, `deadline`, `reminder`, `tags`, `checklist`, `project`, `area`, `heading` | `add` |
| `add_project` | write | `title` (required), `notes`, `when`, `deadline`, `tags`, `area`, `todos` | `add project` |
| `complete` | write | `target` (required), `status` (required enum: `completed`, `canceled`, `incomplete`) | `done` / `cancel` |
| `schedule` | write | `target` (required), `when` (required) | `schedule` |
| `move` | write | `target` (required), `to` (required; project or area) | `move` |
| `edit` | write | `target` (required), plus at least one of: `title`, `notes`, `append_notes`, `deadline`, `clear_deadline`, `reminder`, `tags`, `add_tags` | `edit` |
| `open` | nav | exactly one of `target`, `view` (enum as `list_todos` plus `projects`), `query` | `open` |

Notes on the shape:

- `view` is a parameter of `list_todos`, not thirteen separate view tools: the view is data, the verb is the tool. This keeps the selection space small without the old design's double dispatch (entity plus method).
- `get` is Quick Find: fuzzy resolution across todos and projects; a todo answer includes its checklist, a project answer nests its incomplete todos and headings. Because project detail carries headings, no dedicated heading tool is needed (`add_todo.heading` names one within the destination project).
- `complete` folds done/cancel/reopen into one status enum; `incomplete` reopens.
- `open` is the only tool that touches the app UI; `get` never does. The old server's `show` conflated the two names; here lookup is `get` and navigation is `open`, exactly like the CLI. Its `view` enum is the navigable `list_todos` views plus `projects`, minus `trash`: the URL scheme exposes no trash list id, so trash is not a navigation target.
- Backends mirror the CLI verbatim: `today` and `upcoming` call `Client.Today`/`Client.Upcoming`, so `upcoming` includes repeating tasks at their next occurrence; `logbook` sorts by stop time descending and defaults to the last 30 days, `deadlines` by deadline ascending.

Cross-cutting behavior:

- Pagination on every list/search tool: `limit` and 1-based `page`, returning the `{items, total, page, pages}` envelope. `limit` carries machine-readable schema bounds (default 20, minimum 1, maximum 100) that the SDK applies and validates, so an omitted `limit` is stamped to 20 and an over-cap value is rejected before the handler runs rather than silently clamped. There is deliberately no unlimited mode: MCP output lands in a model context window. An out-of-range page returns empty `items` with intact metadata. The default is 20 rather than the CLI's 10 because a tool round trip costs more than a terminal keystroke.
- Fuzzy resolution wherever an id is accepted (`id`, `target`, `to`, `project`, `area`, `heading`): exact UUID, then UUID prefix of 4+ characters, then exact title, then title substring, narrowed by the parameter's entity type before the one-match decision (a `move.to` matching one project and one todo is not ambiguous). Ambiguity returns up to 10 full-UUID candidates plus a hint to retry with a UUID.
- Strict input parsing: invalid `when`, `deadline`, or `reminder` values produce a structured `invalid_input` error before anything executes. Nothing is silently dropped.
- Verified writes: after executing the URL scheme, the handler polls the database through `internal/verify` (`AddedTodo`/`AddedProject` for the add tools, `TodoStatus`/`ProjectStatus` for `complete`, `TodoModified`/`ProjectModified` for `schedule`/`move`/`edit`). Results carry `verified: true|false`; an unconfirmed send is still `success: true` with an explanatory message, matching the CLI's exit-0 semantics. A server-level mutex serializes execute-plus-verify pairs so concurrent tool calls cannot cross-confirm same-title creates.
- Schema-level validation: every constrained field (`view`, `status`, `type`) is a real JSON Schema enum, so invalid values are rejected by the SDK before a handler runs. Struct tags alone cannot express enums in the official SDK; the server registers explicit schemas for named string types at construction time.

Item shape: one unified `Item` struct for todos and projects (type-discriminated), dates as `YYYY-MM-DD` strings, reminder as `HH:MM`, full UUIDs, inline `project`/`area`/`heading` titles and UUIDs, `evening` and `repeating` flags; separate small shapes for areas and tags.

Inherited URL-scheme limits apply unchanged and are stated in tool descriptions rather than worked around: no delete or trash, no move to Inbox, checklist is replace-only, repeating rules are read-only (a write targeting a repeating template sends and then reports `verified: false`; documented, not pre-blocked, because the repeating flag also marks writable instances), and unknown tags are silently ignored by Things.

### Token efficiency

MCP output lands in a model context window, and early use showed models defaulting to the maximum `limit` with no date filter, so a plain "what did I finish" pulled the entire logbook. Four measures keep a typical answer small without hiding data:

- Schema-enforced pagination (above): `limit` and `page` carry `default`/`minimum`/`maximum` as real JSON Schema keywords, so the model reads the true cap and the SDK enforces it, replacing a silent server-side clamp the model never saw.
- A `days` window on the three date-ordered views (`upcoming`, `logbook`, `deadlines`), whose default result is otherwise unbounded in time. `logbook` defaults to the last 30 days (mirroring the CLI); `0` means all history; the other two default to no window. `days` on any other view is a structured `invalid_input`, rejected before any fetch touches the database.
- Notes truncation: list and search items cap `notes` at 200 runes (rune-safe, never splitting a codepoint) and set `notes_truncated`; `get` returns the full note, so nothing is lost, only deferred to the detail call.
- `--max-limit` (see configuration): the operator can lower the session's page-size cap further, and can only tighten it.

Two structural changes back these up. Project and area scoping for the single-builder views is pushed into SQL through the library's `InProject`/`InArea` builders, whose heading-aware OR-semantics (a todo directly in the project or under one of its headings) match the in-memory filter the composed views still use, rather than fetching a whole view and filtering in Go, so a scoped query reads only its rows. For the date-ordered views the `days` window is likewise pushed into SQL (`StopDate().After`, `Deadline().OnOrBefore`). The composed views (`today`, `upcoming`) stay materialized-then-filtered, exactly as they compose in the app; `upcoming`'s window is applied in memory because it cannot be pushed without breaking the scheduled-plus-repeating-template merge.

### Error contract

One convention for all tools. Domain failures return the tool's normal output envelope with `success: false` and a structured error:

```go
ToolError{Code, Message, Hint, Candidates}
// Code: invalid_input | not_found | ambiguous | execution_failed
```

`ambiguous` carries `Candidates` (full UUID, type, title). Because the error is part of the declared output schema, every failure is machine-parseable and self-correcting: the model sees candidates and a hint instead of an opaque error string. Go errors - and therefore MCP `isError` - are reserved for transport-level failures only (database I/O errors, canceled context), which the model cannot act on.

### Server command and configuration

- `things3 mcp` registers in the Actions group. stdio transport only in v1; the process exits on stdin EOF or SIGINT/SIGTERM via the existing signal context in `main.go`.
- The global `--db` flag is inherited through `withClient`; the other global list/format flags are accepted and ignored per the CLI's uniform-surface convention.
- `--read-only`: registers only the six read tools. The write tools and `open` are not registered at all (not merely rejected), so clients cannot even list them. Read-only is defined as "never executes the URL scheme", which is why `open` - harmless but scheme-executing - is excluded.
- `--max-limit N`: lowers the list page-size cap for the whole session (the default page size is clamped down with it); `0` uses the built-in maximum of 100. It can only tighten the cap, never raise it, so a model can never request a larger page than the operator allows.
- `--log-level debug|info|warn|error` (default `info`): structured `slog` to stderr. stdout belongs exclusively to the protocol.
- Server identity: name `things3`, version from the CLI's ldflags version variable.

### Client lifecycle and concurrency

One `*things3.Client` lives for the whole server session, opened exactly like every other command and closed when the serve loop returns. This is safe: the library opens SQLite read-only (`mode=ro`) with standard `database/sql` pooling, each query is an independent read transaction under WAL, and the auth token is fetched once and cached. Reads run fully concurrent; only writes serialize (see above). If the database file is replaced while the server runs (app reinstall), reads fail visibly until restart - accepted for a local tool.

### Testability

The house rule extends verbatim: MCP write-tool tests never execute osascript. Handlers reach the URL scheme only through an injected execute function; the production default calls the builder's `Execute`, tests inject a recorder that builds the URL and captures it, asserted with the same parsing style as the CLI's `--dry-run` tests. Verification is deterministic in tests via the injectable `verify.Options` clock (instant unverified path) and fixture SQL timestamp bumps (confirmed path). Coverage runs at two levels: direct handler calls against the `thingstest` fixture for behavior, and in-memory MCP transport round trips for protocol concerns - schema enforcement, structured content, and read-only tool listing.

## Implementation Notes

- SDK: official `github.com/modelcontextprotocol/go-sdk` v1.6.x. Typed tool handlers return output structs that the SDK serializes as structured content with an inferred output schema; enum constraints are injected by registering explicit schemas for named string types. The first implementation phase is a spike test proving structured content, failure-envelope schema conformance, and enum rejection on the pinned SDK version before any tool logic is written.
- Package file plan: `server.go` (Server, Config with Version/ReadOnly/Logger/Verify/Execute seams, registration, write mutex), `schema.go`, `result.go`, `convert.go`, `paginate.go`, `resolve.go` (typed wrappers over `internal/resolve`), `writeops.go` (shared attribute-apply helpers for the write tools), one `tool_*.go` per tool group (`tool_list.go`, `tool_search.go`, `tool_get.go`, `tool_add.go`, `tool_complete.go`, `tool_schedule.go`, `tool_move.go`, `tool_edit.go`, `tool_open.go`) with mirrored test files, plus a protocol-level test suite.
- `when` accepts the `ParseWhen` grammar (`today`, `tomorrow`, `evening`, `anytime`, `someday`, `YYYY-MM-DD`). All parsing happens in the server process's local timezone, which is correct for a local stdio server.
- Implementation order after this RFC: scaffolding and SDK spike, read tools, write tools, `open` plus read-only registration, cobra wiring (`cmd/mcp.go`, one line in `root.go`), docs (CLAUDE.md, README.md, cmd/things3/README.md with Claude Code and Claude Desktop configuration snippets). CI and goreleaser need no changes; the existing workflows already build and test the CLI module.
- The old `things3-mcp-server` repository is archived on GitHub once `things3 mcp` ships; nothing in this repo references it.
- Out of scope for v1, recorded for later: HTTP transports, MCP resources/prompts, an `output_mode` token-thinning parameter, trashed-project access, and AppleScript-backed operations the URL scheme cannot express (delete/trash).
