package cmd

// Built-in view and when-keyword names shared across commands.
const (
	nameInbox     = "inbox"
	nameToday     = "today"
	nameUpcoming  = "upcoming"
	nameAnytime   = "anytime"
	nameSomeday   = "someday"
	nameLogbook   = "logbook"
	nameDeadlines = "deadlines"
	nameTomorrow  = "tomorrow"
	nameEvening   = "evening"
)

// Action names and item type discriminators.
const (
	actionAdd  = "add"
	actionOpen = "open"

	typeTodo    = "todo"
	typeProject = "project"
)
