package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descEdit = "Edit a todo or project's attributes. Set at least one of title, notes (replaces), " +
	"append_notes, deadline (YYYY-MM-DD), clear_deadline, reminder (HH:MM), tags (replaces), or add_tags. " +
	"deadline and clear_deadline are mutually exclusive. reminder requires the item to already have a " +
	"scheduled date, which it preserves; on an anytime, someday, or inbox item it is rejected (use schedule " +
	"to set a date first). target resolves a UUID, prefix, or title. The result reports whether the edit " +
	"was verified in the database."

// EditInput is the edit parameter set. An empty string means the field is left
// unchanged; clearing the deadline uses clear_deadline.
type EditInput struct {
	Target        string   `json:"target" jsonschema:"the item to edit (UUID, prefix, or title)"`
	Title         string   `json:"title,omitempty" jsonschema:"set the title"`
	Notes         string   `json:"notes,omitempty" jsonschema:"replace the notes"`
	AppendNotes   string   `json:"append_notes,omitempty" jsonschema:"append to the notes"`
	Deadline      string   `json:"deadline,omitempty" jsonschema:"set the deadline, YYYY-MM-DD"`
	ClearDeadline bool     `json:"clear_deadline,omitempty" jsonschema:"clear the deadline"`
	Reminder      string   `json:"reminder,omitempty" jsonschema:"set the reminder HH:MM; requires and keeps an existing scheduled date"`
	Tags          []string `json:"tags,omitempty" jsonschema:"replace all tags"`
	AddTags       []string `json:"add_tags,omitempty" jsonschema:"add tags without removing existing ones"`
}

// hasEdit reports whether any attribute was supplied.
func (in EditInput) hasEdit() bool {
	return in.Title != "" || in.Notes != "" || in.AppendNotes != "" || in.Deadline != "" ||
		in.ClearDeadline || in.Reminder != "" || len(in.Tags) > 0 || len(in.AddTags) > 0
}

func (s *Server) handleEdit(ctx context.Context, _ *mcp.CallToolRequest, in EditInput) (*mcp.CallToolResult, WriteResult, error) {
	if !in.hasEdit() {
		return nil, writeError(invalidInput("nothing to edit; set at least one attribute")), nil
	}
	if in.Deadline != "" && in.ClearDeadline {
		return nil, writeError(invalidInput("deadline and clear_deadline are mutually exclusive")), nil
	}

	m, te, err := s.resolveTarget(ctx, in.Target)
	if err != nil {
		return nil, WriteResult{}, err
	}
	if te != nil {
		return nil, writeError(te), nil
	}

	baseline := baselineOf(m)
	reminderWhen := scheduledDate(m)
	var builder URLBuilder
	if m.Kind == resolve.KindProject {
		u, eerr := applyEdit(s.client.UpdateProject(m.UUID()), in, reminderWhen)
		if eerr != nil {
			return nil, writeError(eerr), nil
		}
		builder = u
	} else {
		u, eerr := applyEdit(s.client.UpdateTodo(m.UUID()), in, reminderWhen)
		if eerr != nil {
			return nil, writeError(eerr), nil
		}
		builder = u
	}
	result := s.runWrite(ctx, builder, func(ctx context.Context) WriteResult {
		return s.verifyModified(ctx, m.UUID(), string(m.Kind), baseline)
	})
	return nil, result, nil
}

// editable is the shared attribute surface of TodoUpdater and ProjectUpdater that
// edit drives, letting one generic apply to both.
type editable[T any] interface {
	Title(string) T
	Notes(string) T
	AppendNotes(string) T
	When(time.Time) T
	Deadline(time.Time) T
	ClearDeadline() T
	Reminder(hour, minute int) T
	Tags(...string) T
	AddTags(...string) T
}

// applyEdit applies the supplied attributes to an updater, returning a structured
// error on an unparseable deadline or reminder. reminderWhen is the item's current
// scheduled date: a reminder folds into when and would otherwise default the item's
// schedule to today, so edit re-sends the existing date to preserve it and rejects
// a reminder on an item that has no scheduled date at all.
func applyEdit[T editable[T]](u T, in EditInput, reminderWhen *time.Time) (T, *ToolError) {
	if in.Title != "" {
		u = u.Title(in.Title)
	}
	if in.Notes != "" {
		u = u.Notes(in.Notes)
	}
	if in.AppendNotes != "" {
		u = u.AppendNotes(in.AppendNotes)
	}
	if in.Deadline != "" {
		d, te := parseDate(in.Deadline)
		if te != nil {
			return u, te
		}
		u = u.Deadline(d)
	}
	if in.ClearDeadline {
		u = u.ClearDeadline()
	}
	if in.Reminder != "" {
		hour, minute, te := parseReminder(in.Reminder)
		if te != nil {
			return u, te
		}
		if reminderWhen == nil {
			return u, invalidInput("reminder needs a scheduled date: schedule the item to a date first, " +
				"otherwise the reminder would silently move it to today")
		}
		u = u.When(*reminderWhen).Reminder(hour, minute)
	}
	if len(in.Tags) > 0 {
		u = u.Tags(in.Tags...)
	}
	if len(in.AddTags) > 0 {
		u = u.AddTags(in.AddTags...)
	}
	return u, nil
}

// scheduledDate returns the resolved item's concrete start date, or nil when it
// sits in anytime, someday, or the inbox. edit preserves this date when applying a
// reminder so the reminder does not silently reschedule the item to today.
func scheduledDate(m resolve.Match) *time.Time {
	if m.Kind == resolve.KindProject {
		return m.Project.StartDate
	}
	return m.Todo.StartDate
}
