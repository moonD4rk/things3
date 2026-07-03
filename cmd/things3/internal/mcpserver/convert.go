package mcpserver

import (
	"time"

	"github.com/moond4rk/things3"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
)

// dateStr formats a date-only pointer as YYYY-MM-DD, or "" when nil.
func dateStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(dateLayout)
}

// timeStr formats a time-only pointer as HH:MM, or "" when nil.
func timeStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(timeLayout)
}

// stampStr formats a timestamp pointer as RFC 3339, or "" when nil.
func stampStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// todoItem converts a library Todo into the unified Item shape.
func todoItem(t *things3.Todo) Item {
	it := Item{
		Type:        typeTodo,
		UUID:        t.UUID,
		Title:       t.Title,
		Status:      t.Status.String(),
		Start:       t.Start.String(),
		Notes:       t.Notes,
		Tags:        t.Tags,
		When:        dateStr(t.StartDate),
		Deadline:    dateStr(t.Deadline),
		Reminder:    timeStr(t.Reminder),
		Evening:     t.Evening,
		Repeating:   t.Repeating,
		CompletedAt: stampStr(closeTime(t)),
	}
	if t.ProjectUUID != "" {
		it.Project = &Ref{UUID: t.ProjectUUID, Title: t.ProjectTitle}
	}
	if t.AreaUUID != "" {
		it.Area = &Ref{UUID: t.AreaUUID, Title: t.AreaTitle}
	}
	if t.HeadingUUID != "" {
		it.Heading = &Ref{UUID: t.HeadingUUID, Title: t.HeadingTitle}
	}
	for i := range t.Checklist {
		c := &t.Checklist[i]
		it.Checklist = append(it.Checklist, ChecklistItem{UUID: c.UUID, Title: c.Title, Status: c.Status.String()})
	}
	return it
}

// projectItem converts a library Project into the unified Item shape.
func projectItem(p *things3.Project) Item {
	it := Item{
		Type:        typeProject,
		UUID:        p.UUID,
		Title:       p.Title,
		Status:      p.Status.String(),
		Start:       p.Start.String(),
		Notes:       p.Notes,
		Tags:        p.Tags,
		When:        dateStr(p.StartDate),
		Deadline:    dateStr(p.Deadline),
		Reminder:    timeStr(p.Reminder),
		Repeating:   p.Repeating,
		CompletedAt: stampStr(closeTimeProject(p)),
	}
	if p.AreaUUID != "" {
		it.Area = &Ref{UUID: p.AreaUUID, Title: p.AreaTitle}
	}
	return it
}

// todoItems converts a slice of Todos into Items.
func todoItems(todos []things3.Todo) []Item {
	items := make([]Item, len(todos))
	for i := range todos {
		items[i] = todoItem(&todos[i])
	}
	return items
}

// projectItems converts a slice of Projects into Items.
func projectItems(projects []things3.Project) []Item {
	items := make([]Item, len(projects))
	for i := range projects {
		items[i] = projectItem(&projects[i])
	}
	return items
}

// notesLimit bounds the notes text carried in a list or search item; the full
// note remains available through get.
const notesLimit = 200

// truncateNotes shortens over-long notes in a list or search page to notesLimit
// runes in place (rune-safe, never splitting a codepoint) and flags each item it
// cut. It runs on the returned page only, so the shared item converters and the
// get path keep full notes.
func truncateNotes(items []Item) {
	for i := range items {
		r := []rune(items[i].Notes)
		if len(r) > notesLimit {
			items[i].Notes = string(r[:notesLimit])
			items[i].NotesTruncated = true
		}
	}
}

// headingRefs converts headings into inline Refs.
func headingRefs(headings []things3.Heading) []Ref {
	refs := make([]Ref, len(headings))
	for i := range headings {
		refs[i] = Ref{UUID: headings[i].UUID, Title: headings[i].Title}
	}
	return refs
}

// toArea converts a library Area into the small Area shape.
func toArea(a *things3.Area) Area {
	return Area{UUID: a.UUID, Title: a.Title, Tags: a.Tags}
}

// toTag converts a library Tag into the small Tag shape.
func toTag(t *things3.Tag) Tag {
	return Tag{UUID: t.UUID, Title: t.Title, Shortcut: t.Shortcut}
}

// closeTime returns a todo's completion time, else cancellation time, else nil.
func closeTime(t *things3.Todo) *time.Time {
	if t.CompletedAt != nil {
		return t.CompletedAt
	}
	return t.CanceledAt
}

// closeTimeProject mirrors closeTime for a project.
func closeTimeProject(p *things3.Project) *time.Time {
	if p.CompletedAt != nil {
		return p.CompletedAt
	}
	return p.CanceledAt
}
