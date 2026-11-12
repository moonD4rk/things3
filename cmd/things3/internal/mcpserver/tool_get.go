package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descGet = "Look up a single todo or project by id (Quick Find): a full UUID, a 4+ character prefix, an " +
	"exact title, or a title substring. When a title names several items, the single open one wins. A todo " +
	"answer includes its checklist; a project answer nests its incomplete todos and headings. A UUID also " +
	"finds trashed items and repeating series (trashed or template is then true). This reads only and never " +
	"opens the app; use open to reveal an item."

// GetInput is the get parameter set.
type GetInput struct {
	ID string `json:"id" jsonschema:"the item to look up: full UUID, 4+ character prefix, exact title, or title substring"`
}

func (s *Server) handleGet(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, GetResult, error) {
	m, te, err := s.resolveItem(ctx, in.ID)
	if err != nil {
		return nil, GetResult{}, err
	}
	if te != nil {
		return nil, GetResult{Success: false, Error: te}, nil
	}

	if m.Kind == resolve.KindProject {
		// The todos of a trashed project are hidden from every live query.
		q := s.client.Todos().InProject(m.UUID()).Status().Incomplete()
		if m.Trashed() {
			q = q.IncludeTrashed()
		}
		todos, terr := q.All(ctx)
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
	todo, terr := s.client.Todos().WithUUID(m.UUID()).Status().Any().IncludeTrashed().IncludeTemplates().First(ctx)
	if terr != nil {
		return nil, GetResult{}, terr
	}
	item := todoItem(todo)
	return nil, GetResult{Success: true, Item: &item}, nil
}
