package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3"
)

const descSearch = "Full-text search across todos and projects. type narrows to todo, project, or any " +
	"(default). status selects incomplete, completed, canceled, or any (default). tag keeps only items " +
	"carrying that tag. Results are paginated: read total and pages and fetch more only when needed. Notes " +
	"are shortened here (notes_truncated); use get for full text."

// SearchInput is the search parameter set.
type SearchInput struct {
	Query  string       `json:"query" jsonschema:"the text to search for in titles and notes"`
	Type   SearchType   `json:"type,omitempty" jsonschema:"todo, project, or any (default)"`
	Status StatusFilter `json:"status,omitempty" jsonschema:"incomplete, completed, canceled, or any (default)"`
	Tag    string       `json:"tag,omitempty" jsonschema:"keep only items carrying this tag, case-insensitive"`
	Limit  int          `json:"limit,omitempty" jsonschema:"page size"`
	Page   int          `json:"page,omitempty" jsonschema:"1-based page number"`
}

func (s *Server) handleSearch(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, PageResult[Item], error) {
	status := statusOrDefault(in.Status, statusAny)
	kind := string(in.Type)
	if kind == "" {
		kind = searchAny
	}

	var items []Item
	if kind != searchProject {
		q := applyStatus(s.client.Todos().Search(in.Query).Status(), status)
		todos, err := q.All(ctx)
		if err != nil {
			return nil, PageResult[Item]{}, err
		}
		if in.Tag != "" {
			todos = filterByTag(todos, func(t *things3.Todo) []string { return t.Tags }, in.Tag)
		}
		items = append(items, todoItems(todos)...)
	}
	if kind != searchTodo {
		q := applyStatus(s.client.Projects().Search(in.Query).Status(), status)
		projects, err := q.All(ctx)
		if err != nil {
			return nil, PageResult[Item]{}, err
		}
		if in.Tag != "" {
			projects = filterByTag(projects, func(p *things3.Project) []string { return p.Tags }, in.Tag)
		}
		items = append(items, projectItems(projects)...)
	}
	res := pageResult(items, in.Page, in.Limit, s.defaultLimit, s.maxLimit)
	truncateNotes(res.Items)
	return nil, res, nil
}

// applyStatus applies a status string to any typed status sub-builder, defaulting
// to Any for an unrecognized value.
func applyStatus[T any](sf things3.StatusFilter[T], status string) T {
	switch status {
	case statusIncomplete:
		return sf.Incomplete()
	case statusCompleted:
		return sf.Completed()
	case statusCanceled:
		return sf.Canceled()
	default:
		return sf.Any()
	}
}
