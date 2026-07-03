package mcpserver

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3"
)

// Tool descriptions. Each states inherited limits rather than working around
// them, so a model reads the constraint instead of discovering it by failure.
const (
	descListTodos = "List todos from a Things sidebar view. 'view' is required: inbox, today, upcoming, " +
		"anytime, someday, logbook, deadlines, or trash. today and upcoming match the app (upcoming includes " +
		"repeating tasks at their next occurrence); logbook is newest-first, deadlines soonest-first. Narrow " +
		"further by project, area, or tag. Paginated: limit defaults to 20, caps at 100."
	descListProjects = "List projects, optionally filtered by area or tag. status selects incomplete " +
		"(default), completed, canceled, or any. Paginated: limit defaults to 20, caps at 100."
	descListAreas = "List all areas. Paginated: limit defaults to 20, caps at 100."
	descListTags  = "List all tags. Paginated: limit defaults to 20, caps at 100."
)

// registerRead registers the six read tools, always present regardless of mode.
func (s *Server) registerRead(r *registrar) {
	regTool(r, "list_todos", descListTodos, s.handleListTodos)
	regTool(r, "list_projects", descListProjects, s.handleListProjects)
	regTool(r, "list_areas", descListAreas, s.handleListAreas)
	regTool(r, "list_tags", descListTags, s.handleListTags)
	regTool(r, "search", descSearch, s.handleSearch)
	regTool(r, "get", descGet, s.handleGet)
}

// ListTodosInput is the list_todos parameter set.
type ListTodosInput struct {
	View    ViewName `json:"view" jsonschema:"the sidebar view to list"`
	Project string   `json:"project,omitempty" jsonschema:"keep only todos in this project (UUID, prefix, or title)"`
	Area    string   `json:"area,omitempty" jsonschema:"keep only todos in this area (UUID, prefix, or title)"`
	Tag     string   `json:"tag,omitempty" jsonschema:"keep only todos carrying this tag, case-insensitive"`
	Limit   int      `json:"limit,omitempty" jsonschema:"page size, default 20, maximum 100"`
	Page    int      `json:"page,omitempty" jsonschema:"1-based page number"`
}

func (s *Server) handleListTodos(
	ctx context.Context, _ *mcp.CallToolRequest, in ListTodosInput,
) (*mcp.CallToolResult, PageResult[Item], error) {
	todos, err := s.fetchView(ctx, string(in.View))
	if err != nil {
		return nil, PageResult[Item]{}, err
	}
	if in.Project != "" {
		p, te, perr := s.resolveProject(ctx, in.Project)
		if perr != nil {
			return nil, PageResult[Item]{}, perr
		}
		if te != nil {
			return nil, pageError[Item](te), nil
		}
		headings, herr := s.projectHeadingSet(ctx, p.UUID)
		if herr != nil {
			return nil, PageResult[Item]{}, herr
		}
		todos = filterTodos(todos, func(t *things3.Todo) bool {
			if t.ProjectUUID == p.UUID {
				return true
			}
			_, ok := headings[t.HeadingUUID]
			return ok
		})
	}
	if in.Area != "" {
		a, te, aerr := s.resolveArea(ctx, in.Area)
		if aerr != nil {
			return nil, PageResult[Item]{}, aerr
		}
		if te != nil {
			return nil, pageError[Item](te), nil
		}
		todos = filterTodos(todos, func(t *things3.Todo) bool { return t.AreaUUID == a.UUID })
	}
	if in.Tag != "" {
		todos = filterByTag(todos, func(t *things3.Todo) []string { return t.Tags }, in.Tag)
	}
	return nil, pageResult(todoItems(todos), in.Page, in.Limit), nil
}

// ListProjectsInput is the list_projects parameter set.
type ListProjectsInput struct {
	Area   string       `json:"area,omitempty" jsonschema:"keep only projects in this area (UUID, prefix, or title)"`
	Tag    string       `json:"tag,omitempty" jsonschema:"keep only projects carrying this tag, case-insensitive"`
	Status StatusFilter `json:"status,omitempty" jsonschema:"incomplete (default), completed, canceled, or any"`
	Limit  int          `json:"limit,omitempty" jsonschema:"page size, default 20, maximum 100"`
	Page   int          `json:"page,omitempty" jsonschema:"1-based page number"`
}

func (s *Server) handleListProjects(
	ctx context.Context, _ *mcp.CallToolRequest, in ListProjectsInput,
) (*mcp.CallToolResult, PageResult[Item], error) {
	q := s.client.Projects()
	switch statusOrDefault(in.Status, statusIncomplete) {
	case statusCompleted:
		q = q.Status().Completed()
	case statusCanceled:
		q = q.Status().Canceled()
	case statusAny:
		q = q.Status().Any()
	default:
		q = q.Status().Incomplete()
	}
	if in.Area != "" {
		a, te, err := s.resolveArea(ctx, in.Area)
		if err != nil {
			return nil, PageResult[Item]{}, err
		}
		if te != nil {
			return nil, pageError[Item](te), nil
		}
		q = q.InArea(a.UUID)
	}
	projects, err := q.All(ctx)
	if err != nil {
		return nil, PageResult[Item]{}, err
	}
	if in.Tag != "" {
		projects = filterByTag(projects, func(p *things3.Project) []string { return p.Tags }, in.Tag)
	}
	return nil, pageResult(projectItems(projects), in.Page, in.Limit), nil
}

// ListAreasInput is the list_areas parameter set.
type ListAreasInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"page size, default 20, maximum 100"`
	Page  int `json:"page,omitempty" jsonschema:"1-based page number"`
}

func (s *Server) handleListAreas(
	ctx context.Context, _ *mcp.CallToolRequest, in ListAreasInput,
) (*mcp.CallToolResult, PageResult[Area], error) {
	areas, err := s.client.Areas().All(ctx)
	if err != nil {
		return nil, PageResult[Area]{}, err
	}
	items := make([]Area, len(areas))
	for i := range areas {
		items[i] = toArea(&areas[i])
	}
	return nil, pageResult(items, in.Page, in.Limit), nil
}

// ListTagsInput is the list_tags parameter set.
type ListTagsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"page size, default 20, maximum 100"`
	Page  int `json:"page,omitempty" jsonschema:"1-based page number"`
}

func (s *Server) handleListTags(
	ctx context.Context, _ *mcp.CallToolRequest, in ListTagsInput,
) (*mcp.CallToolResult, PageResult[Tag], error) {
	tags, err := s.client.Tags().All(ctx)
	if err != nil {
		return nil, PageResult[Tag]{}, err
	}
	items := make([]Tag, len(tags))
	for i := range tags {
		items[i] = toTag(&tags[i])
	}
	return nil, pageResult(items, in.Page, in.Limit), nil
}

// fetchView returns the todos backing a sidebar view, mirroring the CLI's view
// commands verbatim so today/upcoming compose exactly as the app does.
func (s *Server) fetchView(ctx context.Context, view string) ([]things3.Todo, error) {
	c := s.client
	switch view {
	case nameInbox:
		return c.Todos().Start().Inbox().Status().Incomplete().All(ctx)
	case nameToday:
		return c.Today(ctx)
	case nameUpcoming:
		return c.Upcoming(ctx)
	case nameAnytime:
		return c.Todos().Start().Anytime().Status().Incomplete().All(ctx)
	case nameSomeday:
		return c.Todos().StartDate().Exists(false).Start().Someday().Status().Incomplete().All(ctx)
	case nameLogbook:
		return s.logbookTodos(ctx)
	case nameDeadlines:
		return s.deadlineTodos(ctx)
	case nameTrash:
		return c.Todos().Trashed(true).Status().Any().All(ctx)
	default:
		return nil, fmt.Errorf("unknown view %q", view)
	}
}

// logbookTodos returns completed and canceled todos, most recent first.
func (s *Server) logbookTodos(ctx context.Context) ([]things3.Todo, error) {
	todos, err := s.client.Todos().Status().Any().StopDate().Exists(true).All(ctx)
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(todos, func(a, b things3.Todo) int {
		return compareTimePtrDesc(closeTime(&a), closeTime(&b))
	})
	return todos, nil
}

// deadlineTodos returns incomplete todos with a deadline, soonest first.
func (s *Server) deadlineTodos(ctx context.Context) ([]things3.Todo, error) {
	todos, err := s.client.Todos().Deadline().Exists(true).Status().Incomplete().All(ctx)
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(todos, func(a, b things3.Todo) int {
		return compareTimePtrAsc(a.Deadline, b.Deadline)
	})
	return todos, nil
}

// projectHeadingSet returns the UUIDs of a project's headings, so list_todos can
// include todos filed under a heading (whose ProjectUUID is empty) exactly as the
// library's InProject filter does with its TASK.project OR heading-project match.
func (s *Server) projectHeadingSet(ctx context.Context, projectUUID string) (map[string]struct{}, error) {
	headings, err := s.client.Headings().InProject(projectUUID).All(ctx)
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(headings))
	for i := range headings {
		set[headings[i].UUID] = struct{}{}
	}
	return set, nil
}

// filterTodos keeps the todos for which keep returns true.
func filterTodos(todos []things3.Todo, keep func(*things3.Todo) bool) []things3.Todo {
	out := make([]things3.Todo, 0, len(todos))
	for i := range todos {
		if keep(&todos[i]) {
			out = append(out, todos[i])
		}
	}
	return out
}

// filterByTag keeps items carrying want in any of their tags, case-insensitive.
func filterByTag[T any](items []T, tagsOf func(*T) []string, want string) []T {
	out := make([]T, 0, len(items))
	for i := range items {
		if slices.ContainsFunc(tagsOf(&items[i]), func(tag string) bool {
			return strings.EqualFold(tag, want)
		}) {
			out = append(out, items[i])
		}
	}
	return out
}

// statusOrDefault returns the status string, or def when unset.
func statusOrDefault(s StatusFilter, def string) string {
	if s == "" {
		return def
	}
	return string(s)
}

// compareTimePtrAsc orders non-nil times ascending, sorting nil last.
func compareTimePtrAsc(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	default:
		return a.Compare(*b)
	}
}

// compareTimePtrDesc orders non-nil times descending, sorting nil last.
func compareTimePtrDesc(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	default:
		return b.Compare(*a)
	}
}
