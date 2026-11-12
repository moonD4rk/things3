// Package change turns a requested modification of an existing item into the
// URL-scheme builder that performs it. The CLI's action commands and the MCP
// write tools both describe a change with these types and apply it here, so one
// set of rules decides what a write means. Like resolve and verify, it is
// cobra-free.
package change

import (
	"context"
	"fmt"
	"time"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

// Builder is the Build/Execute surface of every scheme builder.
type Builder interface {
	Build() (string, error)
	Execute(ctx context.Context) error
}

// InputError reports a change that cannot be expressed. The request is wrong,
// as opposed to the environment failing.
type InputError struct{ Message string }

func (e *InputError) Error() string { return e.Message }

func inputErrorf(format string, args ...any) error {
	return &InputError{Message: fmt.Sprintf(format, args...)}
}

// Clock is a reminder's time of day.
type Clock struct{ Hour, Minute int }

// ParseClock parses an HH:MM reminder time.
func ParseClock(s string) (Clock, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return Clock{}, inputErrorf("invalid reminder %q: use HH:MM", s)
	}
	return Clock{Hour: t.Hour(), Minute: t.Minute()}, nil
}

// ParseDate parses a YYYY-MM-DD date, naming the field in the error.
func ParseDate(field, s string) (time.Time, error) {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return time.Time{}, inputErrorf("invalid %s %q: use YYYY-MM-DD", field, s)
	}
	return d, nil
}

// Edit is a set of attribute changes to one todo or project. Zero-valued
// fields leave the attribute untouched; each Clear field empties it, and is
// mutually exclusive with setting the same attribute.
type Edit struct {
	Title *string

	Notes        *string
	ClearNotes   bool
	PrependNotes string
	AppendNotes  string

	Deadline      *time.Time
	ClearDeadline bool

	// Reminder sets the reminder time, keeping the day the item is scheduled
	// for. It needs an item that has such a day.
	Reminder *Clock

	Tags      []string
	ClearTags bool
	AddTags   []string

	// The checklist fields apply to todos only.
	Checklist        []string
	ClearChecklist   bool
	PrependChecklist []string
	AppendChecklist  []string
}

// Empty reports whether the edit changes nothing.
func (e *Edit) Empty() bool {
	return e.Title == nil && e.Notes == nil && !e.ClearNotes && e.PrependNotes == "" && e.AppendNotes == "" &&
		e.Deadline == nil && !e.ClearDeadline && e.Reminder == nil &&
		len(e.Tags) == 0 && !e.ClearTags && len(e.AddTags) == 0 && !e.touchesChecklist()
}

func (e *Edit) touchesChecklist() bool {
	return len(e.Checklist) > 0 || e.ClearChecklist || len(e.PrependChecklist) > 0 || len(e.AppendChecklist) > 0
}

func (e *Edit) validate(m resolve.Match) error {
	switch {
	case e.Empty():
		return inputErrorf("nothing to edit: set at least one attribute")
	case e.Notes != nil && e.ClearNotes:
		return inputErrorf("notes and clear notes are mutually exclusive")
	case e.Deadline != nil && e.ClearDeadline:
		return inputErrorf("deadline and clear deadline are mutually exclusive")
	case len(e.Tags) > 0 && e.ClearTags:
		return inputErrorf("tags and clear tags are mutually exclusive")
	case len(e.Checklist) > 0 && e.ClearChecklist:
		return inputErrorf("checklist and clear checklist are mutually exclusive")
	case e.touchesChecklist() && m.Kind == resolve.KindProject:
		return inputErrorf("%q is a project, and only todos have a checklist", m.Title())
	default:
		return nil
	}
}

// Apply builds the update that performs the edit on the resolved item.
func (e *Edit) Apply(c *things3.Client, m resolve.Match) (Builder, error) {
	if err := e.validate(m); err != nil {
		return nil, err
	}
	if m.Kind == resolve.KindProject {
		return applyAttributes(c.UpdateProject(m.UUID()), e, m.Project.StartDate, false)
	}

	u, err := applyAttributes(c.UpdateTodo(m.UUID()), e, m.Todo.StartDate, m.Todo.Evening)
	if err != nil {
		return nil, err
	}
	if e.ClearChecklist || len(e.Checklist) > 0 {
		u = u.ChecklistItems(e.Checklist...)
	}
	if len(e.PrependChecklist) > 0 {
		u = u.PrependChecklistItems(e.PrependChecklist...)
	}
	if len(e.AppendChecklist) > 0 {
		u = u.AppendChecklistItems(e.AppendChecklist...)
	}
	return u, nil
}

// attributes is the surface TodoUpdater and ProjectUpdater share.
type attributes[T any] interface {
	Builder
	Title(string) T
	Notes(string) T
	PrependNotes(string) T
	AppendNotes(string) T
	When(time.Time) T
	WhenEvening() T
	Deadline(time.Time) T
	ClearDeadline() T
	Reminder(hour, minute int) T
	Tags(...string) T
	AddTags(...string) T
}

// applyAttributes applies the attributes todos and projects have in common.
// scheduled and evening describe where the item sits now.
func applyAttributes[T attributes[T]](u T, e *Edit, scheduled *time.Time, evening bool) (T, error) {
	if e.Title != nil {
		u = u.Title(*e.Title)
	}
	switch {
	case e.ClearNotes:
		u = u.Notes("")
	case e.Notes != nil:
		u = u.Notes(*e.Notes)
	}
	if e.PrependNotes != "" {
		u = u.PrependNotes(e.PrependNotes)
	}
	if e.AppendNotes != "" {
		u = u.AppendNotes(e.AppendNotes)
	}
	switch {
	case e.ClearDeadline:
		u = u.ClearDeadline()
	case e.Deadline != nil:
		u = u.Deadline(*e.Deadline)
	}
	if e.ClearTags || len(e.Tags) > 0 {
		u = u.Tags(e.Tags...)
	}
	if len(e.AddTags) > 0 {
		u = u.AddTags(e.AddTags...)
	}
	if e.Reminder != nil {
		// The URL scheme carries a reminder inside "when", so sending one also
		// sends a day. Re-send the day the item already has, or it would be
		// rescheduled to today.
		switch {
		case scheduled == nil:
			return u, inputErrorf("a reminder needs a scheduled day: schedule the item first, " +
				"otherwise setting the reminder would move it to today")
		case evening:
			u = u.WhenEvening()
		default:
			u = u.When(*scheduled)
		}
		u = u.Reminder(e.Reminder.Hour, e.Reminder.Minute)
	}
	return u, nil
}

// Schedule changes when an item is scheduled, optionally with a reminder.
type Schedule struct {
	// When is today, tomorrow, evening, anytime, someday, or YYYY-MM-DD.
	When     string
	Reminder *Clock
}

// Apply builds the update that reschedules the resolved item.
func (s Schedule) Apply(c *things3.Client, m resolve.Match) (Builder, error) {
	if m.Kind == resolve.KindProject {
		return applySchedule(c.UpdateProject(m.UUID()), s)
	}
	return applySchedule(c.UpdateTodo(m.UUID()), s)
}

type schedulable[T any] interface {
	Builder
	things3.WhenScheduler[T]
	Reminder(hour, minute int) T
}

func applySchedule[T schedulable[T]](u T, s Schedule) (Builder, error) {
	u, err := things3.ParseWhen(u, s.When)
	if err != nil {
		return nil, &InputError{Message: err.Error()}
	}
	if s.Reminder != nil {
		u = u.Reminder(s.Reminder.Hour, s.Reminder.Minute)
	}
	return u, nil
}
