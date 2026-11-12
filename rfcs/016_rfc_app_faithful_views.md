# RFC 016: App-Faithful Views

Status: Implemented Author: @moonD4rk Date: 2026-11-12

## Summary

RFC 012 made the app the north star and named the CLI's views after its sidebar, but the views themselves were still the queries inherited from things.py: todo-only, and built from the start bucket alone. Compared item for item with the lists the Things app reports through AppleScript, only the Inbox agreed. `anytime` returned five times what the app shows, `someday` and the app's Someday list had not a single item in common, and `today`, `upcoming`, and `logbook` each missed or added whole categories. This RFC redefines every view as the list the app shows: same members, same order. The eight views move into the library as `Client` methods returning a mixed list of todos and projects, and the CLI and the MCP server render those lists instead of each carrying its own copy of the recipes. It supersedes the view sections of RFC 012 and RFC 013 and the `list_todos` recipes of RFC 014.

## Design

### 1. A view is the app's list

The rules below were derived by diffing against the app and are the contract. After the change, on a database of several thousand items, every view the app exposes through AppleScript (Inbox, Today, Anytime, Upcoming, Someday, Logbook, Trash) matches the app in both membership and order.

| View | Members | Order |
| --- | --- | --- |
| Inbox | Open todos in the Inbox bucket. | Sidebar |
| Today | Open todos and projects that are scheduled into Today, or scheduled for a day that has arrived (the app moves those into Today when it next runs), or unscheduled with a due deadline that was not dismissed from Today. | Today position; This Evening last |
| Upcoming | Open todos and projects with a start date in the future; each repeating series at its next occurrence; open items with no start date filed under their future deadline; then the series that have no next occurrence (paused or ended). | Filed day, then Today position; the dateless series last |
| Anytime | Open todos in the Anytime bucket whose project is not in Someday. | Sidebar |
| Someday | Open, unscheduled Someday projects, and open, unscheduled Someday todos that are not inside a project. | Sidebar |
| Logbook | Completed and canceled projects, and completed and canceled todos whose project is not itself closed. | Most recently closed first |
| Deadlines | Open todos and projects with a deadline. | Deadline, soonest first |
| Trash | Trashed projects, and trashed todos whose project is not itself trashed. | Most recently trashed first |

Three ideas recur. **Projects are list members**: a project scheduled for today is a row of Today, a Someday project is a row of Someday. **A container can hide or absorb its todos**: a todo in a Someday project keeps `start = 1` but is not available, so Anytime leaves it out; the todos of a completed project are part of the project's Logbook row, and likewise in the Trash. **A todo inside a project belongs to the project**: a Someday todo in an active project shows in that project, not in the Someday list.

The sidebar order is by area, then project, then heading, then the item's own position, with whatever has no area, project, or heading leading its level; items that share a position follow most recently modified first. It is now the default order of every query, so `projects` lists projects grouped by area as the sidebar does, and a project's todos come out heading by heading.

### 2. Items

A view returns `[]things3.Item`. An `Item` holds a `Todo` or a `Project` (exactly one), with accessors for what the two share (`UUID`, `Title`, `Status`, `Tags`, `StartDate`, `Deadline`, `StoppedAt`, `UpcomingDate`, ...). It is a pairing at the view layer only; the typed query builders of RFC 009 are untouched and there is still no union model. In JSON and YAML an item is the todo or project object led by a `type` field, so every list a tool emits has one item shape.

`UpcomingDate` is the day Upcoming files an item under (start date, else deadline); it is what the CLI groups by and what a `days` window measures.

### 3. One recipe, three surfaces

The views are `Client` methods: `Inbox`, `Today`, `Upcoming`, `Anytime`, `Someday`, `Logbook(since)`, `Deadlines`, `Trash`. Only the Logbook takes a parameter, because it is the one view that grows without bound.

The CLI's view commands and the MCP `list_todos` tool fetch from those methods and apply only presentation: tag filtering, `--days` windows, sorting, pagination, and (for the MCP server) project and area scoping in memory, which is exact now that an item reports its effective containers. The recipes that the CLI and the MCP server each used to spell out are gone.

### 4. Text rows

With projects in most lists, a row needs to say what it is. Rather than give every list a type column, the status mark takes the item's shape, as it does in the app: square for a todo (`[ ]`, `[x]`, `[-]`), round for a project (`( )`, `(x)`, `(-)`). The separate "mixed list" format of RFC 013 and its `TYPE` column are removed: `trash`, `search`, `projects`, multi-match `show`, and every view print the same row, and a project row carries its `@Area`. Multi-match `show` now paginates like any other list.

The MCP tool keeps the name `list_todos`, so existing client configurations and permission allowlists keep working; its description states that a view lists projects where the app does.

## Implementation Notes

- **Membership rules live in the database layer.** Today and Upcoming are disjunctions, which the AND-only filter builder could not express, so each used to be several queries stitched together in Go, with an order that was only right within each part. `TaskFilter` gains `InToday` and `InUpcoming` (one SQL predicate each, anchored to the client's clock) and the container predicates `ContainerNotSomeday`, `ContainerOpen`, and `ContainerNotTrashed`. A view is now one filter plus one order, and the order is the app's across the whole list.
- **Orders are explicit.** `TaskOrder` replaces the free-form index column: `OrderSidebar` (default), `OrderToday`, `OrderUpcoming`, `OrderDeadline`, `OrderStopDateDesc`, `OrderModifiedDesc`. Each ends in the item's own position so equal keys never fall back to SQLite's row order.
- **Known gap.** A repeating series stores only its next occurrence, so it appears once in Upcoming. A series that is paused or past its end date has none and trails the list without a day, which is where the app's list puts it as well.
- **Verification.** Fixture tests pin each rule, mutating a fixture copy for the cases the fixture lacks (a todo in a Someday project, a scheduled project, a dormant series). The comparison against the live app is manual and is the check to repeat before changing a rule: for each list, `tell application "Things3" to get id of to dos of list "<Name>"` against the view's UUIDs.
