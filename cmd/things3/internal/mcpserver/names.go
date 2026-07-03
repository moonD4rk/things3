package mcpserver

// Server identity and item-type discriminators shared across the package.
const (
	// serverName is the MCP server identity reported to clients.
	serverName = "things3"

	// typeTodo and typeProject discriminate the unified Item shape.
	typeTodo    = "todo"
	typeProject = "project"

	// maxCandidates caps how many ambiguity candidates a ToolError carries.
	maxCandidates = 10
)

// Built-in view and when-keyword names, shared by the list, open, move, and schema
// code. The same word can name a sidebar view and a schedule bucket.
const (
	nameInbox     = "inbox"
	nameToday     = "today"
	nameUpcoming  = "upcoming"
	nameAnytime   = "anytime"
	nameSomeday   = "someday"
	nameLogbook   = "logbook"
	nameDeadlines = "deadlines"
	nameTrash     = "trash"
	nameTomorrow  = "tomorrow"
	nameEvening   = "evening"
	nameProjects  = "projects"
)

// instructions is sent to clients to frame how the tools behave.
const instructions = `things3 exposes the Things 3 task manager as verb-shaped tools that mirror its command line.

Reads (list_todos, list_projects, list_areas, list_tags, search, get) query the local database.
Writes (add_todo, add_project, complete, schedule, move, edit) go through the Things URL scheme and are
confirmed against the database; each result reports verified true or false. open reveals an item or list
in the Things app.

Wherever a tool accepts an id, target, or destination it resolves a full UUID, a UUID prefix of four or
more characters, an exact title, or a title substring. An ambiguous match returns candidate UUIDs to retry
with. Dates use YYYY-MM-DD; when also accepts today, tomorrow, evening, anytime, and someday; reminder uses
HH:MM. Inherited URL-scheme limits: no delete or trash, no move to Inbox, checklists replace rather than
append, repeating series are read-only, and unknown tags are ignored by Things.

Lists report total and pages. Start with the default page size, which answers most questions, and fetch
further pages only when the question needs them. When only the count matters, pass limit 1 and read total.
For a date-scoped question on upcoming, logbook, or deadlines, pass days rather than paging the whole view.
List and search items shorten notes; get returns the full note.`
