# RFC 017: Repeating Rules and the Complete Edit Surface

Status: Implemented Author: @moonD4rk Date: 2026-11-12

## Summary

Two statements in the project's own documentation turned out to be wrong, and each hid a missing capability. "The recurrence rule is persisted as an opaque binary blob" (RFC 013): it is an XML property list, and reading it tells a user how a task repeats and why a series produces no occurrences. "Checklists are replace-only" (RFC 012, RFC 014): the URL scheme can also prepend and append checklist items, and neither the CLI nor the MCP server could touch a checklist at all after creation. This RFC decodes the rule and closes the gap between what the URL scheme can change and what the tools let a caller change. It also puts the rules of an edit in one place, because the CLI and the MCP server had each grown their own copy and the copies had drifted.

## Design

### 1. Repeating rules

`TMTask.rt1_recurrenceRule` is a property list with two-letter keys:

| Key | Meaning |
| --- | --- |
| `fu` | Unit: 16 daily, 256 weekly, 8 monthly, 4 yearly |
| `fa` | Interval: every `fa` units |
| `tp` | 0 follows a fixed schedule, 1 counts from the day the previous occurrence was completed |
| `of` | Offsets within the period: `dy` day of month (zero-based, -1 for the last), `mo` month (zero-based), `wd` weekday (0 is Sunday), `wdo` weekday ordinal ("the 2nd Friday", -1 for the last) |
| `ed` | End date as Unix seconds; 4001-01-01 means the series never ends |

The read model decodes it into `Repeat` on `Todo` and `Project`: the unit, the interval, `AfterCompletion`, `Until`, and a `Summary` in words ("every 2 weeks on Mon, Fri", "every month on the last day", "every 3 days after completion"). A generated instance reports the rule of the series it came from, so a todo in Today can say how it repeats. A rule that cannot be decoded leaves the item flagged `Repeating` with no `Repeat`, rather than failing the query.

Two states explain a series with no next occurrence, which Upcoming lists last (RFC 016): `Paused`, read from `rt1_instanceCreationPaused`, and `Ended`, a series that is not paused yet has no next occurrence because it passed its end date or used up its count. The CLI marks such rows `repeats (paused)` or `repeats (ended)`, detail views print the rule, and MCP items carry it as a `repeat` string.

Expanding a rule into future dates remains out of scope. The database stores the next occurrence, and that is what the views use.

### 2. What can be edited

The URL scheme's `update` command accepts more than the tools exposed. The surface is now complete, and identical in the CLI and the MCP server:

| Attribute | Set | Add | Clear |
| --- | --- | --- | --- |
| Title | `--title` / `title` |  |  |
| Notes | `--notes` / `notes` | `--prepend-notes`, `--append-notes` / `prepend_notes`, `append_notes` | `--clear-notes` / `clear_notes` |
| Deadline | `--deadline` / `deadline` |  | `--clear-deadline` / `clear_deadline` |
| Reminder | `--reminder` / `reminder` |  |  |
| Tags | `--tags` / `tags` | `--add-tags` / `add_tags` | `--clear-tags` / `clear_tags` |
| Checklist (todos) | `--checklist` / `checklist` | `--prepend-checklist`, `--append-checklist` / `prepend_checklist`, `append_checklist` | `--clear-checklist` / `clear_checklist` |

A field and its clear counterpart are mutually exclusive. The real limit, stated wherever the old "replace-only" wording stood, is that existing checklist items cannot be checked off or changed one by one.

Three smaller gaps close with it. `schedule` accepts a reminder, so "tomorrow at nine" is one call rather than a schedule followed by an edit. `add project` accepts tags and a reminder in the CLI, as the MCP tool already did for tags. `projects --all` no longer doubles as a status filter: `--all` only lifts pagination, as on every other list, and `--status` chooses incomplete, completed, canceled, or any.

### 3. One set of rules

`cmd/things3/internal/change` holds what an edit and a schedule mean. A front end fills in a `change.Edit` or `change.Schedule` from its flags or parameters, and `Apply` builds the update for the resolved item. The rules live there once:

- A checklist change on a project is rejected.
- A reminder travels inside the scheme's `when` parameter, so setting one also sends a day. `Edit` re-sends the day the item already has, keeps a This Evening todo in This Evening, and rejects an item with no scheduled day rather than silently moving it to today.
- A rejected request is an `InputError`: exit 1 in the CLI, `invalid_input` in the MCP server, and no URL is built.

### 4. Batch creation

The library's JSON batch builder, the one path to features the single-item commands lack, covered only part of the `json` command. It gains project headings (`Heading(title, todos...)`, the only way the URL scheme can create a heading; a project's children keep the order they were added in), reminders on todos and projects, checklist items with their own state (`Checklist(ChecklistEntry...)`), and on updates `PrependChecklistItems`, `AppendChecklistItems`, and `HeadingID`.

## Implementation Notes

- **Parsing.** `internal/database/recurrence.go` reads the property-list subset Things writes (dict, array, integer, real, string, booleans) with `encoding/xml`; no dependency is added. The series' rule reaches an instance through a join on `rt1_repeatingTemplate`.
- **Zero-based offsets.** `dy` and `mo` are zero-based in the rule and one-based in the summary. This was confirmed against stored next occurrences: a yearly rule with `mo` 2 and `dy` 13 recurs on March 14.
- **Reminder folding is shared.** URL and JSON builders use one function to fold a reminder into `when`, so both default a missing day to today and both reject someday and anytime.
- **Verified live.** Every write this RFC adds was executed against Things 3.24.1 on a scratch todo and project: the four checklist operations, notes prepend and clear, tags clear, a reminder on a todo and on a project, and `schedule` with `anytime`, which `update` accepts although the documentation lists it only for `add`.
- **A checklist write does not touch the todo.** Things changes the checklist rows and the todo's item counts and leaves its modification date alone, so a verifier that waits for `ModifiedAt` to advance reported every checklist-only edit as unconfirmed. A todo now counts as changed when its modification date advanced or its checklist holds different items. `ResolveTarget` loads the checklist so that the resolved target is the state to compare against, and a verified write returns the todo with its checklist.
- **A write that changes nothing cannot be confirmed.** Things leaves the row untouched when the item is already in the requested state, so such a write reports `verified: false` like one that was dropped. Telling the two apart needs a comparison with the requested state rather than with the previous one; that is left for a later change.
