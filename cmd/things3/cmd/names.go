package cmd

import "github.com/moond4rk/things3"

// Built-in view and when-keyword names shared across commands.
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
)

// Action names and item type discriminators.
const (
	actionAdd  = "add"
	actionOpen = "open"

	typeTodo    = things3.ItemTypeTodo
	typeProject = things3.ItemTypeProject
)
