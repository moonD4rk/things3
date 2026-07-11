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
		"further by project, area, or tag. For a date-scoped question pass days on upcoming, logbook, or " +
		"deadlines; logbook defaults to the last 30 days. Results are paginated: read total and pages and fetch " +
		"more pages only when the question needs them. Notes are shortened here (notes_truncated); use get for full text."
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
	Area    string   `json:"area,omitempty" jsonschema:"keep only todos in this area (UUID, prefix, or title)"`
	Tag     string   `json:"tag,omitempty" jsonschema:"keep only todos carrying this tag, case-insensitive"`
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

	todos, err := s.viewTodos(ctx, string(in.View), projectUUID, areaUUID, in.Days)
	if err != nil {
		return nil, PageResult[Item]{}, err
	}
	if in.Tag != "" {
		todos = filterByTag(todos, func(t *things3.Todo) []string { return t.Tags }, in.Tag)
	}
	res := pageResult(todoItems(todos), in.Page, in.Limit, s.defaultLimit, s.maxLimit)
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

// viewTodos returns the todos backing a sidebar view, scoped to an optional
// project and area. Single-builder views push the scoping into SQL through the
// library builders; the composed views (today, upcoming) merge several queries
// and so are materialized and filtered in memory, exactly as the app composes
// them.
func (s *Server) viewTodos(ctx context.Context, view, projectUUID, areaUUID string, days *int) ([]things3.Todo, error) {
	if viewIsComposed(view) {
		todos, err := s.composedView(ctx, view)
		if err != nil {
			return nil, err
		}
		if view == nameUpcoming && days != nil && *days > 0 {
			todos = withinUpcomingWindow(todos, *days)
		}
		return s.scopeComposed(ctx, todos, projectUUID, areaUUID)
	}

	q, err := s.viewBuilder(view)
	if err != nil {
		return nil, err
	}
	if projectUUID != "" {
		q = q.InProject(projectUUID)
	}
	if areaUUID != "" {
		q = q.InArea(areaUUID)
	}
	q = applyDaysWindow(q, view, days)
	todos, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	viewSort(view, todos)
	return todos, nil
}

// viewIsComposed reports whether a view is assembled in Go from several queries
// (today, upcoming) and so cannot be expressed as a single builder.
func viewIsComposed(view string) bool {
	return view == nameToday || view == nameUpcoming
}

// composedView returns the todos for the composed views, which merge and
// reorder several queries the way the app does.
func (s *Server) composedView(ctx context.Context, view string) ([]things3.Todo, error) {
	switch view {
	case nameToday:
		return s.client.Today(ctx)
	case nameUpcoming:
		return s.client.Upcoming(ctx)
	default:
		return nil, fmt.Errorf("unknown composed view %q", view)
	}
}

// viewBuilder returns the query builder backing a single-builder view, carrying
// only its WHERE shape; project/area scoping and the display sort are applied by
// the caller. The recipes mirror the CLI's view commands verbatim.
func (s *Server) viewBuilder(view string) (things3.TodoQueryBuilder, error) {
	c := s.client
	switch view {
	case nameInbox:
		return c.Todos().Start().Inbox().Status().Incomplete(), nil
	case nameAnytime:
		return c.Todos().Start().Anytime().Status().Incomplete(), nil
	case nameSomeday:
		return c.Todos().StartDate().Exists(false).Start().Someday().Status().Incomplete(), nil
	case nameLogbook:
		return c.Todos().Status().Any().StopDate().Exists(true), nil
	case nameDeadlines:
		return c.Todos().Deadline().Exists(true).Status().Incomplete(), nil
	case nameTrash:
		return c.Todos().Trashed(true).Status().Any(), nil
	default:
		return nil, fmt.Errorf("unknown view %q", view)
	}
}

// viewSort applies the per-view display order the builders cannot express (they
// order by the sidebar index only): logbook newest-first by stop time,
// deadlines soonest-first by deadline.
func viewSort(view string, todos []things3.Todo) {
	switch view {
	case nameLogbook:
		slices.SortStableFunc(todos, func(a, b things3.Todo) int {
			return compareTimePtrDesc(closeTime(&a), closeTime(&b))
		})
	case nameDeadlines:
		slices.SortStableFunc(todos, func(a, b things3.Todo) int {
			return compareTimePtrAsc(a.Deadline, b.Deadline)
		})
	}
}

// applyDaysWindow chains the date-window filter onto a builder view that
// supports days. logbook windows on stop date (defaulting to the last 30 days
// to mirror the CLI; an explicit 0 means all history); deadlines windows forward
// on deadline and deliberately keeps overdue items inside the window, since they
// are still due. now is time.Now, matching the CLI's logbook window.
func applyDaysWindow(q things3.TodoQueryBuilder, view string, days *int) things3.TodoQueryBuilder {
	now := time.Now()
	switch view {
	case nameLogbook:
		if w := logbookWindowDays(days); w > 0 {
			q = q.StopDate().After(now.AddDate(0, 0, -w))
		}
	case nameDeadlines:
		if days != nil && *days > 0 {
			q = q.Deadline().OnOrBefore(now.AddDate(0, 0, *days))
		}
	}
	return q
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

// withinUpcomingWindow keeps composed upcoming todos whose start date is on or
// before now+days. Upcoming is already future-scheduled, so an upper bound
// alone yields the next N days; it cannot be pushed to SQL without breaking the
// scheduled-plus-repeating-template merge.
func withinUpcomingWindow(todos []things3.Todo, days int) []things3.Todo {
	cutoff := time.Now().AddDate(0, 0, days)
	return filterTodos(todos, func(t *things3.Todo) bool {
		return t.StartDate != nil && !t.StartDate.After(cutoff)
	})
}

// scopeComposed filters composed-view todos by project (heading-aware, matching
// the library's InProject OR-semantics) and area in memory, since a composed
// view is already materialized and cannot be re-queried.
func (s *Server) scopeComposed(
	ctx context.Context, todos []things3.Todo, projectUUID, areaUUID string,
) ([]things3.Todo, error) {
	if projectUUID != "" {
		headings, err := s.projectHeadingSet(ctx, projectUUID)
		if err != nil {
			return nil, err
		}
		todos = filterTodos(todos, func(t *things3.Todo) bool {
			if t.ProjectUUID == projectUUID {
				return true
			}
			_, ok := headings[t.HeadingUUID]
			return ok
		})
	}
	if areaUUID != "" {
		todos = filterTodos(todos, func(t *things3.Todo) bool { return t.AreaUUID == areaUUID })
	}
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
