# RFC 015: Read Model and Target Resolution Corrections

Status: Implemented Author: @moonD4rk Date: 2026-11-12

## Summary

An audit against a live Things 3.24 database (schema version 29) and against the app itself, read through AppleScript, found that the storage layer the library sits on has not drifted, but that several things the library and CLI report were wrong. This RFC records the corrections that do not change the shape of the sidebar views (those follow in RFC 016): what a todo says about its project and area, how packed dates decode, what "today" means, how a query picks one item out of several, and what happens when the item it picks cannot be modified. It amends the read model of RFC 009, the resolver contract of RFC 012, and the tool contracts of RFC 014. Two principles drive every change below: a row reports where the item lives in the app rather than what its columns happen to store, and every UUID the tool prints must be a UUID the tool accepts.

## Design

### 1. A todo reports its effective containers

Things stores a container reference only on the item directly inside it. A todo under a heading has a `heading` and nothing else; a todo in a project has a `project` and no `area`. The old read model copied those columns, so a todo under a heading had no project (it rendered without its `@Project`, and `anytime` filed it under "No Project") and an area filter matched only the loose todos of an area, which for anyone who keeps their work in projects is nothing at all.

The read model now resolves both. A todo's project is its own or its heading's; its area is its own, its project's, or its heading's project's. `InProject`, `HasProject`, `InArea`, and `HasArea` filter on those same expressions, so a filter always tests the value a row reports. `Project.AreaUUID` is unchanged, since a project carries its area itself.

Text search no longer matches the area title. It was inherited from things.py, and with inherited areas it would return every todo of an area whose name merely contains the query. A search matches an item's own title and notes.

### 2. Packed dates

A Things date is `year<<16 | month<<12 | day<<7`. The decoder masked the year to 11 bits, on the assumption that the format was 27 bits wide. It is not: Things writes the year 4001 into the deadline of every repeating template that has one, meaning "each occurrence gets its own deadline", and 4001 masked to 11 bits is 1953. Every such template appeared in `upcoming` with `due:1953-01-01`. Older databases also store 0001-01-01, rather than NULL, for "no next occurrence".

The year is now read from every bit above the month field, and a date outside `[0002-01-01, 4001-01-01)` is a sentinel, decoded as no date. Decoding moved out of SQL: the query selects the raw integers and `internal/database/date.go` is the single place that interprets them.

Timestamps are fractional Unix seconds. They were truncated to whole seconds, which made write verification (which waits for `ModifiedAt` to advance past a baseline) report a second write within the same second as unconfirmed. They now decode at microsecond precision. An item whose modification date is NULL reports its creation date instead of the zero time.

### 3. One clock

Relative date filters (`Past`, `Future`) and the composed views asked SQLite for `now`, while the CLI and MCP server computed their `--days` windows from `time.Now()`. Nothing could pin "today", so the fixture's absolute dates turned the test suite into a time bomb: it began failing the day a fixture todo scheduled for 2026-09-17 stopped being in the future.

The client now owns a clock. `WithClock` replaces it, `Client.Now` reads it, and every relative filter is rendered from it in Go rather than delegated to SQLite. `thingstest.Now` is the fixture's reference instant, and every test client is pinned to it.

### 4. Scopes: Trash and repeating templates

Queries return live items by default: not trashed, not inside a trashed project, and not a repeating template. That default is right for lists and wrong for identity. `IncludeTrashed()` and `IncludeTemplates()` lift the two exclusions, and `Trashed(true)` keeps its meaning of "only the Trash". A new `Template` field on `Todo` and `Project` flags a template; its `StartDate` is always its next occurrence, whichever query returned it.

### 5. Picking one item

The resolver's tiers are unchanged: exact UUID, UUID prefix of four or more characters, exact title, title substring; the first tier with a hit wins. What changed is what happens to the hits.

- **The UUID tiers reach everything that exists.** `trash` prints the UUIDs of trashed items and `upcoming` prints the UUIDs of repeating templates. Both used to answer "no item matches" when pasted back into `show`, which broke the round-trip guarantee RFC 012 is built on. The title tiers still match live items only.
- **Several hits narrow before they are called ambiguous**: first to the live ones, then to the single item with the preferred status. A repeating task leaves one open todo and a trail of completed copies, all with the same title; `done "Water plants"` used to stop with exit 2 and over a hundred candidates. The preferred status is open for every lookup and write except reopening, which prefers closed. A narrowing that would leave nothing is skipped, so a lone completed match is still returned. When the narrowing leaves several, only those are reported as candidates.
- **A write needs an item Things lets it change.** The URL scheme cannot modify a trashed item or a repeating template. A write that resolves to one stops with an explanation (exit 1 in the CLI, `invalid_input` in the MCP server) instead of sending a URL that Things would ignore. Reads and `open` are unaffected.

### 6. Opening the app's search

The MCP `open` tool documented its `query` parameter as full-text search but sent `show?query=`, which is Quick Find: it jumps to the area, project, or tag with that name and cannot find a todo at all. It now sends `search?query=`. The library exposes this as `Client.SearchBuilder`, and the CLI gains `open --search <text>` so the two surfaces stay aligned.

## Implementation Notes

- **Tags and checklists are EXISTS predicates.** The task query used to LEFT JOIN the tag and checklist tables and collapse the product with `SELECT DISTINCT`. A dangling tag reference next to a real tag produced two distinct rows for one task. The joins are gone: `has_tags`, `has_checklist`, and the tag filters are correlated `EXISTS` subqueries, each row is produced once, and a reference to a deleted tag is not a tag.
- **The row model carries raw values.** `TaskRow` holds the integer type, status, and start bucket rather than strings the root package parsed back into enums.
- **CI.** The workflow recreated `go.work` with `go work init`, which stamps the running toolchain's version; once Go 1.27 became `stable`, the pinned linter (built with Go 1.26) refused to load the workspace. The workspace is now pinned to the `go` directive of `go.mod`, and the linter pin moved to a release built with the current Go.
- **Verification.** Each correction was checked read-only against a live database, comparing the CLI's output with the lists the app itself reports through AppleScript: heading children regained their project, an area filter reached the todos inside the area's projects, repeating templates lost their 1953 deadline, and titles that were ambiguous only because of completed namesakes resolved to the open item.
