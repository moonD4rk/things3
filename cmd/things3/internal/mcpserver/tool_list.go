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
	descListTodos = "List a Things sidebar view exactly as the app shows it. 'view' is required: inbox, today, " +
		"upcoming, anytime, someday, logbook, deadlines, or trash. Items are todos and, where the app lists " +
		"them, projects (type tells which): today and upcoming include scheduled projects, someday lists " +
		"Someday projects and standalone Someday todos, logbook lists closed projects and the closed todos not " +
		"folded into one. anytime holds only todos you can act on now, so todos of a Someday project are " +
		"absent. upcoming is in date order and includes each repeating series at its next occurrence and " +
		"unscheduled items at their deadline; logbook is newest-first, deadlines soonest-first. Narrow further " +
		"by project, area, or tag. For a date-scoped question pass days on upcoming, logbook, or deadlines; " +
		"logbook defaults to the last 30 days. Results are paginated: read total and pages and fetch more pages " +
		"only when the question needs them. Notes are shortened here (notes_truncated); use get for full text."
	descListProjects = "List projects, optionally filtered by area or tag. status selects incomplete " +
		"(default), completed, canceled, or any. Results are paginated: read total and pages and fetch more " +
		"only when needed. Notes are shortened here (notes_truncated); use get for full text."
	descListAreas = "List all areas. Results are paginated: read total and pages and fetch more only when needed."
	descListTags  = "List all tags. Results are paginated: read total and pages and fetch more only when needed."
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
	Area    string   `json:"area,omitempty" jsonschema:"keep only items in this area (UUID, prefix, or title)"`
	Tag     string   `json:"tag,omitempty" jsonschema:"keep only items carrying this tag, case-insensitive"`
	Days    *int     `json:"days,omitempty" jsonschema:"day window for upcoming/logbook/deadlines; logbook defaults to 30, 0 = all"`
	Limit   int      `json:"limit,omitempty" jsonschema:"page size"`
	Page    int      `json:"page,omitempty" jsonschema:"1-based page number"`
}

func (s *Server) handleListTodos(
	ctx context.Context, _ *mcp.CallToolRequest, in ListTodosInput,
) (*mcp.CallToolResult, PageResult[Item], error) {
	if in.Days != nil {
		if te := validateDays(string(in.View), *in.Days); te != nil {
			return nil, pageError[Item](te), nil
		}
	}

	projectUUID, areaUUID, te, err := s.resolveContainers(ctx, in.Project, in.Area)
	if err != nil {
		return nil, PageResult[Item]{}, err
	}
	if te != nil {
		return nil, pageError[Item](te), nil
	}

	items, err := s.viewItems(ctx, string(in.View), in.Days)
	if err != nil {
		return nil, PageResult[Item]{}, err
	}
	items = scopeItems(items, projectUUID, areaUUID)
	if in.Tag != "" {
		items = filterByTag(items, func(it *things3.Item) []string { return it.Tags() }, in.Tag)
	}
	res := pageItems(items, convertItem, in.Page, in.Limit, s.defaultLimit, s.maxLimit)
	truncateNotes(res.Items)
	return nil, res, nil
}

// validateDays rejects a days window on a view that does not support one, and a
// negative window, as a structured invalid_input that rides the envelope rather
// than a transport error. It runs before any fetch, so a bad request touches no
// database.
func validateDays(view string, days int) *ToolError {
	if !viewSupportsDays(view) {
		return invalidInput("days is only valid for the upcoming, logbook, and deadlines views")
	}
	if days < 0 {
		return invalidInput("days must be zero or positive")
	}
	return nil
}

// viewSupportsDays reports whether a view accepts a days window: the three
// date-ordered views whose default result is unbounded in time.
func viewSupportsDays(view string) bool {
	return view == nameUpcoming || view == nameLogbook || view == nameDeadlines
}

// resolveContainers resolves the optional project and area filters to UUIDs. A
// resolution ambiguity or miss rides the envelope (te != nil); a database
// failure is a transport error.
func (s *Server) resolveContainers(
	ctx context.Context, project, area string,
) (projectUUID, areaUUID string, te *ToolError, err error) {
	if project != "" {
		p, pte, perr := s.resolveProject(ctx, project)
		if perr != nil {
			return "", "", nil, perr
		}
		if pte != nil {
			return "", "", pte, nil
		}
		projectUUID = p.UUID
	}
	if area != "" {
		a, ate, aerr := s.resolveArea(ctx, area)
		if aerr != nil {
			return "", "", nil, aerr
		}
		if ate != nil {
			return "", "", ate, nil
		}
		areaUUID = a.UUID
	}
	return projectUUID, areaUUID, nil, nil
}

// ListProjectsInput is the list_projects parameter set.
type ListProjectsInput struct {
	Area   string       `json:"area,omitempty" jsonschema:"keep only projects in this area (UUID, prefix, or title)"`
	Tag    string       `json:"tag,omitempty" jsonschema:"keep only projects carrying this tag, case-insensitive"`
	Status StatusFilter `json:"status,omitempty" jsonschema:"incomplete (default), completed, canceled, or any"`
	Limit  int          `json:"limit,omitempty" jsonschema:"page size"`
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
	res := pageResult(projectItems(projects), in.Page, in.Limit, s.defaultLimit, s.maxLimit)
	truncateNotes(res.Items)
	return nil, res, nil
}

// ListAreasInput is the list_areas parameter set.
type ListAreasInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"page size"`
	Page  int `json:"page,omitempty" jsonschema:"1-based page number"`
}

func (s *Server) handleListAreas(
	ctx context.Context, _ *mcp.CallToolRequest, in ListAreasInput,
) (*mcp.CallToolResult, PageResult[Area], error) {
	areas, err := s.client.Areas().All(ctx)
	if err != nil {
		return nil, PageResult[Area]{}, err
	}
	return nil, pageItems(areas, toArea, in.Page, in.Limit, s.defaultLimit, s.maxLimit), nil
}

// ListTagsInput is the list_tags parameter set.
type ListTagsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"page size"`
	Page  int `json:"page,omitempty" jsonschema:"1-based page number"`
}

func (s *Server) handleListTags(
	ctx context.Context, _ *mcp.CallToolRequest, in ListTagsInput,
) (*mcp.CallToolResult, PageResult[Tag], error) {
	tags, err := s.client.Tags().All(ctx)
	if err != nil {
		return nil, PageResult[Tag]{}, err
	}
	return nil, pageItems(tags, toTag, in.Page, in.Limit, s.defaultLimit, s.maxLimit), nil
}

// viewItems returns a sidebar view as the library composes it, narrowed to the
// days window on the three date-ordered views.
func (s *Server) viewItems(ctx context.Context, view string, days *int) ([]things3.Item, error) {
	c := s.client
	switch view {
	case nameInbox:
		return c.Inbox(ctx)
	case nameToday:
		return c.Today(ctx)
	case nameAnytime:
		return c.Anytime(ctx)
	case nameSomeday:
		return c.Someday(ctx)
	case nameTrash:
		return c.Trash(ctx)
	case nameUpcoming:
		items, err := c.Upcoming(ctx)
		return withinDays(items, things3.Item.UpcomingDate, c.Now(), days), err
	case nameDeadlines:
		// The window keeps overdue items: they are still due.
		items, err := c.Deadlines(ctx)
		return withinDays(items, things3.Item.Deadline, c.Now(), days), err
	case nameLogbook:
		var since time.Time
		if w := logbookWindowDays(days); w > 0 {
			since = c.Now().AddDate(0, 0, -w)
		}
		return c.Logbook(ctx, since)
	default:
		return nil, fmt.Errorf("unknown view %q", view)
	}
}

// logbookWindowDays resolves the logbook window: an omitted days defaults to 30
// (the CLI's default), while an explicit value passes through, with 0 meaning
// all history.
func logbookWindowDays(days *int) int {
	if days == nil {
		return 30
	}
	return *days
}

// withinDays keeps the items whose date falls on or before now+days; an unset
// or zero window keeps everything.
func withinDays(items []things3.Item, date func(things3.Item) *time.Time, now time.Time, days *int) []things3.Item {
	if days == nil || *days <= 0 {
		return items
	}
	cutoff := now.AddDate(0, 0, *days)
	return slices.DeleteFunc(items, func(it things3.Item) bool {
		d := date(it)
		return d == nil || d.After(cutoff)
	})
}

// scopeItems narrows a view to a project and an area. Items report their
// effective containers, so a todo under a heading matches its project and a
// todo in a project matches that project's area.
func scopeItems(items []things3.Item, projectUUID, areaUUID string) []things3.Item {
	return slices.DeleteFunc(items, func(it things3.Item) bool {
		if projectUUID != "" && (it.Todo == nil || it.Todo.ProjectUUID != projectUUID) {
			return true
		}
		return areaUUID != "" && it.AreaUUID() != areaUUID
	})
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
