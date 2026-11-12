package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3/cmd/things3/internal/change"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descEdit = "Edit a todo or project's attributes; set at least one. title sets the title. notes replaces " +
	"the notes, prepend_notes and append_notes add to them, clear_notes empties them. deadline (YYYY-MM-DD) " +
	"sets and clear_deadline removes the deadline. tags replaces all tags, add_tags adds, clear_tags removes " +
	"all. For a todo, checklist replaces the checklist, prepend_checklist and append_checklist add items, and " +
	"clear_checklist removes it; existing items cannot be checked off or changed one by one. reminder (HH:MM) " +
	"keeps the day the item is scheduled for and is rejected on an item with no scheduled day (use schedule " +
	"first). A field and its clear_ counterpart are mutually exclusive. target resolves a UUID, prefix, or " +
	"title. The result reports whether the edit was verified in the database."

// EditInput is the edit parameter set. An empty value leaves the field
// unchanged; the clear_ flags empty it.
type EditInput struct {
	Target           string   `json:"target" jsonschema:"the item to edit (UUID, prefix, or title)"`
	Title            string   `json:"title,omitempty" jsonschema:"set the title"`
	Notes            string   `json:"notes,omitempty" jsonschema:"replace the notes"`
	PrependNotes     string   `json:"prepend_notes,omitempty" jsonschema:"add text before the notes"`
	AppendNotes      string   `json:"append_notes,omitempty" jsonschema:"add text after the notes"`
	ClearNotes       bool     `json:"clear_notes,omitempty" jsonschema:"empty the notes"`
	Deadline         string   `json:"deadline,omitempty" jsonschema:"set the deadline, YYYY-MM-DD"`
	ClearDeadline    bool     `json:"clear_deadline,omitempty" jsonschema:"clear the deadline"`
	Reminder         string   `json:"reminder,omitempty" jsonschema:"set the reminder HH:MM; requires and keeps a scheduled day"`
	Tags             []string `json:"tags,omitempty" jsonschema:"replace all tags"`
	AddTags          []string `json:"add_tags,omitempty" jsonschema:"add tags without removing existing ones"`
	ClearTags        bool     `json:"clear_tags,omitempty" jsonschema:"remove every tag"`
	Checklist        []string `json:"checklist,omitempty" jsonschema:"replace the checklist with these items (todos only)"`
	PrependChecklist []string `json:"prepend_checklist,omitempty" jsonschema:"add checklist items at the top (todos only)"`
	AppendChecklist  []string `json:"append_checklist,omitempty" jsonschema:"add checklist items at the bottom (todos only)"`
	ClearChecklist   bool     `json:"clear_checklist,omitempty" jsonschema:"remove the checklist (todos only)"`
}

// change maps the tool input onto the shared edit rules.
func (in *EditInput) change() (*change.Edit, error) {
	edit := &change.Edit{
		ClearNotes:       in.ClearNotes,
		PrependNotes:     in.PrependNotes,
		AppendNotes:      in.AppendNotes,
		ClearDeadline:    in.ClearDeadline,
		Tags:             in.Tags,
		ClearTags:        in.ClearTags,
		AddTags:          in.AddTags,
		Checklist:        in.Checklist,
		ClearChecklist:   in.ClearChecklist,
		PrependChecklist: in.PrependChecklist,
		AppendChecklist:  in.AppendChecklist,
	}
	if in.Title != "" {
		edit.Title = &in.Title
	}
	if in.Notes != "" {
		edit.Notes = &in.Notes
	}
	if in.Deadline != "" {
		d, err := change.ParseDate("deadline", in.Deadline)
		if err != nil {
			return nil, err
		}
		edit.Deadline = &d
	}
	if in.Reminder != "" {
		clock, err := change.ParseClock(in.Reminder)
		if err != nil {
			return nil, err
		}
		edit.Reminder = &clock
	}
	return edit, nil
}

func (s *Server) handleEdit(ctx context.Context, _ *mcp.CallToolRequest, in EditInput) (*mcp.CallToolResult, WriteResult, error) {
	edit, err := in.change()
	if err != nil {
		return nil, writeError(changeError(err)), nil
	}
	if edit.Empty() {
		return nil, writeError(invalidInput("nothing to edit; set at least one attribute")), nil
	}

	m, te, err := s.resolveTarget(ctx, in.Target, resolve.PreferOpen)
	if err != nil {
		return nil, WriteResult{}, err
	}
	if te != nil {
		return nil, writeError(te), nil
	}

	builder, err := edit.Apply(s.client, m)
	if err != nil {
		return nil, writeError(changeError(err)), nil
	}
	result := s.runWrite(ctx, builder, func(ctx context.Context) WriteResult {
		return s.verifyModified(ctx, m)
	})
	return nil, result, nil
}
