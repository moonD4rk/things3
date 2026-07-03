package mcpserver

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descMove = "Move a todo or project to a project or area (the app's Move). to is required and resolves a " +
	"UUID, prefix, or title. A todo can move to a project or an area; a project can move only to an area. " +
	"The URL scheme cannot move items to the Inbox, and dates belong to schedule, not here. The result " +
	"reports whether the move was verified in the database."

// MoveInput is the move parameter set.
type MoveInput struct {
	Target string `json:"target" jsonschema:"the item to move (UUID, prefix, or title)"`
	To     string `json:"to" jsonschema:"destination project or area (UUID, prefix, or title)"`
}

func (s *Server) handleMove(ctx context.Context, _ *mcp.CallToolRequest, in MoveInput) (*mcp.CallToolResult, WriteResult, error) {
	switch {
	case strings.EqualFold(in.To, nameInbox):
		return nil, writeError(invalidInput("the Things URL scheme cannot move items to the Inbox")), nil
	case isWhenKeyword(in.To):
		return nil, writeError(invalidInput(`"to" takes a project or area; use schedule for dates`)), nil
	}

	m, te, err := s.resolveTarget(ctx, in.Target)
	if err != nil {
		return nil, WriteResult{}, err
	}
	if te != nil {
		return nil, writeError(te), nil
	}

	baseline := baselineOf(m)
	builder, te, err := s.moveBuilder(ctx, m, in.To)
	if err != nil {
		return nil, WriteResult{}, err
	}
	if te != nil {
		return nil, writeError(te), nil
	}
	result := s.runWrite(ctx, builder, func(ctx context.Context) WriteResult {
		return s.verifyModified(ctx, m.UUID(), string(m.Kind), baseline)
	})
	return nil, result, nil
}

// moveBuilder builds the destination updater. A project can only move to an area;
// a todo tries a project destination first, then an area.
func (s *Server) moveBuilder(ctx context.Context, m resolve.Match, dest string) (URLBuilder, *ToolError, error) {
	if m.Kind == resolve.KindProject {
		area, err := resolve.Area(ctx, s.client, dest)
		if err != nil {
			if isNotFound(err) {
				if _, perr := resolve.Project(ctx, s.client, dest); perr == nil {
					return nil, invalidInput("a project can only move to an area, not another project"), nil
				}
			}
			te, rerr := resolveError(dest, err)
			return nil, te, rerr
		}
		return s.client.UpdateProject(m.UUID()).AreaID(area.UUID), nil, nil
	}

	project, err := resolve.Project(ctx, s.client, dest)
	if err == nil {
		return s.client.UpdateTodo(m.UUID()).ListID(project.UUID), nil, nil
	}
	if !isNotFound(err) {
		te, rerr := resolveError(dest, err)
		return nil, te, rerr
	}
	area, aerr := resolve.Area(ctx, s.client, dest)
	if aerr != nil {
		te, rerr := resolveError(dest, aerr)
		return nil, te, rerr
	}
	return s.client.UpdateTodo(m.UUID()).ListID(area.UUID), nil, nil
}

// isWhenKeyword reports whether dest names a schedule bucket rather than a place.
func isWhenKeyword(dest string) bool {
	switch strings.ToLower(dest) {
	case nameToday, nameTomorrow, nameEvening, nameAnytime, nameSomeday:
		return true
	default:
		return false
	}
}
