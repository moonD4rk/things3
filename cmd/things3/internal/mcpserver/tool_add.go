package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3"
)

const (
	descAddTodo = "Create a todo. title is required. when accepts today, tomorrow, evening, anytime, someday, " +
		"or YYYY-MM-DD; deadline is YYYY-MM-DD; reminder is HH:MM. Place it with project, area (mutually " +
		"exclusive), or heading (a heading within the destination project, requires project). tags and " +
		"checklist take lists. Unknown tags are ignored by Things. The result reports whether the write was " +
		"verified in the database."
	descAddProject = "Create a project. title is required. when accepts today, tomorrow, evening, anytime, " +
		"someday, or YYYY-MM-DD; deadline is YYYY-MM-DD. area places it; todos seeds initial child todos. The " +
		"result reports whether the write was verified in the database."
)

// AddTodoInput is the add_todo parameter set.
type AddTodoInput struct {
	Title     string   `json:"title" jsonschema:"the todo title"`
	Notes     string   `json:"notes,omitempty" jsonschema:"the todo notes"`
	When      string   `json:"when,omitempty" jsonschema:"schedule: today, tomorrow, evening, anytime, someday, or YYYY-MM-DD"`
	Deadline  string   `json:"deadline,omitempty" jsonschema:"deadline date, YYYY-MM-DD"`
	Reminder  string   `json:"reminder,omitempty" jsonschema:"reminder time, HH:MM"`
	Tags      []string `json:"tags,omitempty" jsonschema:"tag names; unknown tags are ignored by Things"`
	Checklist []string `json:"checklist,omitempty" jsonschema:"checklist item titles"`
	Project   string   `json:"project,omitempty" jsonschema:"destination project (UUID, prefix, or title)"`
	Area      string   `json:"area,omitempty" jsonschema:"destination area (UUID, prefix, or title); mutually exclusive with project"`
	Heading   string   `json:"heading,omitempty" jsonschema:"a heading within the destination project; requires project"`
}

func (s *Server) handleAddTodo(ctx context.Context, _ *mcp.CallToolRequest, in AddTodoInput) (*mcp.CallToolResult, WriteResult, error) {
	if in.Project != "" && in.Area != "" {
		return nil, writeError(invalidInput("project and area are mutually exclusive")), nil
	}

	builder := s.client.AddTodo().Title(in.Title)
	if in.Notes != "" {
		builder = builder.Notes(in.Notes)
	}
	if in.When != "" {
		b, err := things3.ParseWhen(builder, in.When)
		if err != nil {
			return nil, writeError(invalidInput(err.Error())), nil
		}
		builder = b
	}
	if in.Deadline != "" {
		d, te := parseDate(in.Deadline)
		if te != nil {
			return nil, writeError(te), nil
		}
		builder = builder.Deadline(d)
	}
	if in.Reminder != "" {
		hour, minute, te := parseReminder(in.Reminder)
		if te != nil {
			return nil, writeError(te), nil
		}
		builder = builder.Reminder(hour, minute)
	}
	if len(in.Tags) > 0 {
		builder = builder.Tags(in.Tags...)
	}
	if len(in.Checklist) > 0 {
		builder = builder.ChecklistItems(in.Checklist...)
	}

	builder, te, err := s.placeTodo(ctx, builder, in)
	if err != nil {
		return nil, WriteResult{}, err
	}
	if te != nil {
		return nil, writeError(te), nil
	}

	t0 := now().Add(-guardBand)
	result := s.runWrite(ctx, builder, func(ctx context.Context) WriteResult {
		return s.verifyAddedTodo(ctx, in.Title, t0)
	})
	return nil, result, nil
}

// placeTodo resolves the destination flags onto the adder. heading requires
// project; project and area are mutually exclusive (checked by the caller).
func (s *Server) placeTodo(ctx context.Context, builder things3.TodoAdder, in AddTodoInput) (things3.TodoAdder, *ToolError, error) {
	switch {
	case in.Heading != "":
		if in.Project == "" {
			return builder, invalidInput("heading requires project"), nil
		}
		p, te, err := s.resolveProject(ctx, in.Project)
		if err != nil || te != nil {
			return builder, te, err
		}
		h, te, err := s.resolveHeading(ctx, p.UUID, in.Heading)
		if err != nil || te != nil {
			return builder, te, err
		}
		return builder.ListID(p.UUID).HeadingID(h.UUID), nil, nil
	case in.Project != "":
		p, te, err := s.resolveProject(ctx, in.Project)
		if err != nil || te != nil {
			return builder, te, err
		}
		return builder.ListID(p.UUID), nil, nil
	case in.Area != "":
		a, te, err := s.resolveArea(ctx, in.Area)
		if err != nil || te != nil {
			return builder, te, err
		}
		return builder.ListID(a.UUID), nil, nil
	default:
		return builder, nil, nil
	}
}

// AddProjectInput is the add_project parameter set.
type AddProjectInput struct {
	Title    string   `json:"title" jsonschema:"the project title"`
	Notes    string   `json:"notes,omitempty" jsonschema:"the project notes"`
	When     string   `json:"when,omitempty" jsonschema:"schedule: today, tomorrow, evening, anytime, someday, or YYYY-MM-DD"`
	Deadline string   `json:"deadline,omitempty" jsonschema:"deadline date, YYYY-MM-DD"`
	Tags     []string `json:"tags,omitempty" jsonschema:"tag names; unknown tags are ignored by Things"`
	Area     string   `json:"area,omitempty" jsonschema:"destination area (UUID, prefix, or title)"`
	Todos    []string `json:"todos,omitempty" jsonschema:"initial child todo titles"`
}

func (s *Server) handleAddProject(
	ctx context.Context, _ *mcp.CallToolRequest, in AddProjectInput,
) (*mcp.CallToolResult, WriteResult, error) {
	builder := s.client.AddProject().Title(in.Title)
	if in.Notes != "" {
		builder = builder.Notes(in.Notes)
	}
	if in.When != "" {
		b, err := things3.ParseWhen(builder, in.When)
		if err != nil {
			return nil, writeError(invalidInput(err.Error())), nil
		}
		builder = b
	}
	if in.Deadline != "" {
		d, te := parseDate(in.Deadline)
		if te != nil {
			return nil, writeError(te), nil
		}
		builder = builder.Deadline(d)
	}
	if len(in.Tags) > 0 {
		builder = builder.Tags(in.Tags...)
	}
	if in.Area != "" {
		a, te, err := s.resolveArea(ctx, in.Area)
		if err != nil {
			return nil, WriteResult{}, err
		}
		if te != nil {
			return nil, writeError(te), nil
		}
		builder = builder.AreaID(a.UUID)
	}
	if len(in.Todos) > 0 {
		builder = builder.Todos(in.Todos...)
	}

	t0 := now().Add(-guardBand)
	result := s.runWrite(ctx, builder, func(ctx context.Context) WriteResult {
		return s.verifyAddedProject(ctx, in.Title, t0)
	})
	return nil, result, nil
}
