package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descGet = "Look up a single todo or project by id (Quick Find): a full UUID, a 4+ character prefix, an " +
	"exact title, or a title substring. A todo answer includes its checklist; a project answer nests its " +
	"incomplete todos and headings. This reads only and never opens the app; use open to reveal an item."

// GetInput is the get parameter set.
type GetInput struct {
	ID string `json:"id" jsonschema:"the item to look up: full UUID, 4+ character prefix, exact title, or title substring"`
}

func (s *Server) handleGet(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, GetResult, error) {
	m, err := resolve.ResolveOne(ctx, s.client, in.ID)
	if err != nil {
		te, rerr := resolveError(in.ID, err)
		if rerr != nil {
			return nil, GetResult{}, rerr
		}
		return nil, GetResult{Success: false, Error: te}, nil
	}

	if m.Kind == resolve.KindProject {
		todos, terr := s.client.Todos().InProject(m.UUID()).Status().Incomplete().All(ctx)
		if terr != nil {
			return nil, GetResult{}, terr
		}
		headings, herr := s.client.Headings().InProject(m.UUID()).All(ctx)
		if herr != nil {
			return nil, GetResult{}, herr
		}
		item := projectItem(m.Project)
		// The resolved project keeps its full note; the todos nested under it are a
		// list like any other, so they shorten the same way list_todos does.
		nested := todoItems(todos)
		truncateNotes(nested)
		return nil, GetResult{Success: true, Item: &item, Todos: nested, Headings: headingRefs(headings)}, nil
	}

	// Re-fetch to load the checklist, which Resolve does not populate.
	todo, terr := s.client.Todos().WithUUID(m.UUID()).Status().Any().IncludeChecklist().First(ctx)
	if terr != nil {
		return nil, GetResult{}, terr
	}
	item := todoItem(todo)
	return nil, GetResult{Success: true, Item: &item}, nil
}
