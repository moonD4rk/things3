package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descOpen = "Reveal an item or list in the Things app. Provide exactly one of: target (an item to reveal, " +
	"resolved by UUID, prefix, or title), view (a built-in list: inbox, today, upcoming, anytime, someday, " +
	"logbook, deadlines, or projects), or query (text to open the app's search). This is the only tool that " +
	"touches the app UI."

// openViewLists maps the open.view enum to built-in Things list ids.
var openViewLists = map[string]things3.ListID{
	nameInbox:     things3.ListInbox,
	nameToday:     things3.ListToday,
	nameUpcoming:  things3.ListUpcoming,
	nameAnytime:   things3.ListAnytime,
	nameSomeday:   things3.ListSomeday,
	nameLogbook:   things3.ListLogbook,
	nameDeadlines: things3.ListDeadlines,
	nameProjects:  things3.ListAllProjects,
}

// OpenInput is the open parameter set; exactly one field must be set.
type OpenInput struct {
	Target string   `json:"target,omitempty" jsonschema:"an item to reveal (UUID, prefix, or title)"`
	View   OpenView `json:"view,omitempty" jsonschema:"a built-in list to open"`
	Query  string   `json:"query,omitempty" jsonschema:"text to open the app's search for"`
}

func (s *Server) handleOpen(ctx context.Context, _ *mcp.CallToolRequest, in OpenInput) (*mcp.CallToolResult, WriteResult, error) {
	set := 0
	for _, v := range []string{in.Target, string(in.View), in.Query} {
		if v != "" {
			set++
		}
	}
	if set != 1 {
		return nil, writeError(invalidInput("provide exactly one of target, view, or query")), nil
	}

	nav := s.client.ShowBuilder()
	message := ""
	switch {
	case in.View != "":
		nav = nav.List(openViewLists[string(in.View)])
		message = string(in.View)
	case in.Query != "":
		nav = nav.Query(in.Query)
		message = "search: " + in.Query
	default:
		m, te, err := s.resolveTarget(ctx, in.Target)
		if err != nil {
			return nil, WriteResult{}, err
		}
		if te != nil {
			return nil, writeError(te), nil
		}
		nav = nav.ID(m.UUID())
		item := matchItem(m)
		message = m.Title()
		result := s.runWrite(ctx, nav, nil)
		if result.Success {
			result.Message = message
			result.Item = &item
		}
		return nil, result, nil
	}

	result := s.runWrite(ctx, nav, nil)
	if result.Success {
		result.Message = message
	}
	return nil, result, nil
}

// matchItem converts a resolved match into an Item for an open result.
func matchItem(m resolve.Match) Item {
	if m.Kind == resolve.KindProject {
		return projectItem(m.Project)
	}
	return todoItem(m.Todo)
}
