package things3

import (
	"context"
	"slices"
	"time"

	"github.com/moond4rk/things3/internal/database"
)

// The sidebar views. Each returns what the Things app lists there, as verified
// against the app itself, so a todo's presence in a view means the same thing
// it means on screen. Every view returns a non-nil slice.

var (
	todosOnly        = []int{int(taskTypeTodo)}
	projectsOnly     = []int{int(taskTypeProject)}
	todosAndProjects = []int{int(taskTypeTodo), int(taskTypeProject)}
)

// viewFilter returns the filter every view of open items starts from.
func viewFilter(types []int) database.TaskFilter {
	return database.TaskFilter{Types: types, Status: new(int(StatusIncomplete))}
}

func dateExists(exists bool) *database.DateFilterValue {
	return &database.DateFilterValue{HasDate: &exists}
}

// Inbox returns the todos in the Inbox, in sidebar order.
func (c *Client) Inbox(ctx context.Context) ([]Item, error) {
	f := viewFilter(todosOnly)
	f.Start = new(int(StartInbox))
	return c.database.items(ctx, &f)
}

// Today returns the Today list: todos and projects scheduled into Today, those
// whose scheduled day has arrived, and unscheduled ones whose deadline is due.
// Items follow the list's own order, with This Evening last.
func (c *Client) Today(ctx context.Context) ([]Item, error) {
	f := viewFilter(todosAndProjects)
	f.InToday = true
	f.Order = database.OrderToday

	items, err := c.database.items(ctx, &f)
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(items, func(a, b Item) int {
		return compareBool(isEvening(a), isEvening(b))
	})
	return items, nil
}

// Upcoming returns the Upcoming list in the app's order: todos and projects
// scheduled for a future day, each repeating series at its next occurrence, and
// unscheduled items filed under their future deadline, followed by the series
// that have no next occurrence (paused or ended). [Item.UpcomingDate] gives the
// day an item is filed under; it is nil for that trailing group.
//
// The database stores only the next occurrence of a repeating series, so a
// series appears once.
func (c *Client) Upcoming(ctx context.Context) ([]Item, error) {
	dated := viewFilter(todosAndProjects)
	dated.Templates = database.TemplatesAny
	dated.InUpcoming = true
	dated.Order = database.OrderUpcoming

	dormant := viewFilter(todosAndProjects)
	dormant.Templates = database.TemplatesOnly
	dormant.StartDateFilter = dateExists(false)
	dormant.Order = database.OrderToday

	return c.database.itemsOf(ctx, &dated, &dormant)
}

// Anytime returns the Anytime list: the todos available to work on now. A todo
// inside a project that sits in Someday is not listed, exactly as the app hides
// a shelved project's todos.
func (c *Client) Anytime(ctx context.Context) ([]Item, error) {
	f := viewFilter(todosOnly)
	f.Start = new(int(StartAnytime))
	f.ContainerNotSomeday = true
	return c.database.items(ctx, &f)
}

// Someday returns the Someday list: unscheduled Someday todos that stand on
// their own, then Someday projects. A Someday todo inside a project stays in
// its project and is not listed here.
func (c *Client) Someday(ctx context.Context) ([]Item, error) {
	todos := viewFilter(todosOnly)
	todos.Start = new(int(StartSomeday))
	todos.StartDateFilter = dateExists(false)
	todos.HasProject = new(false)

	projects := todos
	projects.Types = projectsOnly

	return c.database.itemsOf(ctx, &todos, &projects)
}

// Logbook returns the Logbook, most recently closed first: completed and
// canceled projects, and the completed and canceled todos that are not part of
// such a project (those are folded into the project's own row). A non-zero
// since keeps only items closed after that day.
func (c *Client) Logbook(ctx context.Context, since time.Time) ([]Item, error) {
	f := database.TaskFilter{
		Types:          todosAndProjects,
		Closed:         true,
		ContainerOpen:  true,
		StopDateFilter: dateExists(true),
		Order:          database.OrderStopDateDesc,
	}
	if !since.IsZero() {
		f.StopDateFilter = &database.DateFilterValue{Operator: ">", Date: &since}
	}
	return c.database.items(ctx, &f)
}

// Deadlines returns every open todo and project that has a deadline, soonest
// first; overdue ones lead the list.
func (c *Client) Deadlines(ctx context.Context) ([]Item, error) {
	f := viewFilter(todosAndProjects)
	f.DeadlineFilter = dateExists(true)
	f.Order = database.OrderDeadline
	return c.database.items(ctx, &f)
}

// Trash returns the Trash, most recently trashed first: trashed todos and
// projects, whatever their status. A trashed todo inside a trashed project is
// part of that project's row and is not listed on its own.
func (c *Client) Trash(ctx context.Context) ([]Item, error) {
	return c.database.items(ctx, &database.TaskFilter{
		Types:               todosAndProjects,
		Trash:               database.TrashOnly,
		ContainerNotTrashed: true,
		Order:               database.OrderModifiedDesc,
	})
}

func isEvening(it Item) bool {
	return it.Todo != nil && it.Todo.Evening
}

// compareBool orders false before true.
func compareBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	default:
		return -1
	}
}
